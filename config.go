package main

type config struct {
	DBAddr string `env:"DB_ADDR" envDefault:"localhost:6379"`

	TelegramToken string `env:"TG_TOKEN,required"`
	ChannelID     int64  `env:"CHANNEL_ID,required"`

	VKToken  string `env:"VK_TOKEN,required"`
	VKChatID uint   `env:"VK_CHAT_ID,required"`

	TikTokUsername  string `env:"TIKTOK_USERNAME,required"`
	TikTokSecUserID string `env:"TIKTOK_SEC_USER_ID"`
}
