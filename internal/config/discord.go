package config

import (
	"os"
	"sync"
)

type DiscordConfig struct {
	BotToken  string
	BotToken2 string
}

var (
	discordConfig *DiscordConfig
	discordOnce   sync.Once
)

func LoadDiscordConfig() *DiscordConfig {
	discordOnce.Do(func() {
		discordConfig = &DiscordConfig{
			BotToken:  os.Getenv("DISCORD_BOT_TOKEN"),
			BotToken2: os.Getenv("DISCORD_BOT_TOKEN2"),
		}
	})
	return discordConfig
}
