package main

import (
	"log"
	"os"
	"time"

	"github.com/XRain777/vkapi"
	"github.com/caarlos0/env/v6"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

var cfg config
var r *redis.Client
var tg *tb.Bot
var vk *vkapi.API

func main() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("Config", err)
	}

	r = redis.NewClient(&redis.Options{
		Addr: cfg.DBAddr,
	})

	var err error

	tg, err = tb.NewBot(tb.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatalln("Telegram", err)
	}

	vk = vkapi.NewClient(cfg.VKToken)

	if cfg.TikTokSecUserID == "" {
		cfg.TikTokSecUserID, err = getSecUserID(cfg.TikTokUsername)
		if err != nil {
			log.Fatalln("SecUID", err)
		}
	}

	for {
		log.Println("Polling...")
		checkNewVideos()
		time.Sleep(time.Minute)
	}
}

func checkNewVideos() {
	likes, err := getLikedVideos(cfg.TikTokSecUserID, 20)
	if err != nil {
		log.Println("Likes", err)
		return
	}

	for _, v := range likes {
		if wasAlreadyPosted(v.ID) {
			continue
		}

		log.Println("Posting to Telegram channel", v.ID)

		menu := &tb.ReplyMarkup{}
		menu.Inline(
			menu.Row(menu.URL("Оригинал", v.ShareURL)),
		)

		_, err = tg.Send(tb.ChatID(cfg.ChannelID), &tb.Video{
			File: tb.File{FileURL: v.DownloadURL},
		}, menu)
		if err != nil {
			log.Println("Send video Telegram", err, v.DownloadURL)
		}

		log.Println("Sending to VK chat", v.ID)
		videoSaveParams := vkapi.VideoSaveParams{
			Name:      "Tiktok " + v.ID,
			IsPrivate: true,
			WallPost:  false,
			Repeat:    true,
		}
		videoSaveResponse, err := vk.VideoSave(videoSaveParams)
		if err != nil {
			log.Println("Video save VK", err)
			continue
		}
		err = downloadFile("video.mp4", v.DownloadURL)
		if err != nil {
			log.Println("Video download VK", err)
			continue
		}
		videoUploadResponse, err := vkapi.UploadVideoFromFile(videoSaveResponse.UploadURL, "video.mp4")
		if err != nil {
			log.Println("Video upload VK", err)
			continue
		}
		err = os.Remove("video.mp4")
		if err != nil {
			log.Fatal(err)
		}

		messageSendParams := vkapi.MessagesSendParams{
			ChatID:     cfg.VKChatID,
			Attachment: vkapi.MakeAttachment("video", videoSaveResponse.OwnerID, videoUploadResponse.VideoID),
		}
		_, err = vk.MessagesSend(messageSendParams)
		if err != nil {
			log.Println("Message send VK", err)
		}

		time.Sleep(time.Second * 3)
	}
}
