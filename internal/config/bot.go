package config

import (
	"log"
	"os"
	"sync"
)

type BotConfig struct {
	Name string
	Env  string
}

var (
	botConfig *BotConfig
	botOnce   sync.Once
)

func LoadBotConfig() *BotConfig {
	botOnce.Do(func() {
		env := os.Getenv("BOT_ENV")
		if env == "" {
			env = "development"
			log.Printf("Warning: BOT_ENV not set, defaulting to %s", env)
		}
		botConfig = &BotConfig{
			Name: os.Getenv("BOT_NAME"),
			Env:  env,
		}
	})
	return botConfig
}
