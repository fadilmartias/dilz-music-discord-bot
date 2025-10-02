package config

import (
	"os"
	"sync"
)

type DiscordConfig struct {
	BotToken string
}

var (
	discordConfig *DiscordConfig
	discordOnce   sync.Once
)

func LoadDiscordConfig() *DiscordConfig {
	discordOnce.Do(func() {
		discordConfig = &DiscordConfig{
			BotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		}
	})
	return discordConfig
}
