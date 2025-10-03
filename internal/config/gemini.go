package config

import (
	"os"
	"sync"
)

type GeminiConfig struct {
	ApiKey string
}

var (
	geminiConfig *GeminiConfig
	geminiOnce   sync.Once
)

func LoadGeminiConfig() *GeminiConfig {
	geminiOnce.Do(func() {
		geminiConfig = &GeminiConfig{
			ApiKey: os.Getenv("GEMINI_API_KEY"),
		}
	})
	return geminiConfig
}
