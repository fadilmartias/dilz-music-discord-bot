package main

import (
	"fmt"
	"log"

	"github.com/fadilmartias/dilz-music-discord-bot/internal/handler"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/service"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/usecase"
	"github.com/joho/godotenv"
)

func main() {

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("could not load env")
	}

	// Create service
	discordService := service.NewDiscordService()
	if discordService == nil {
		fmt.Println("Failed to initialize Discord service")
		return
	}

	// Create handler
	musicService := service.NewMusicService()
	musicUC := usecase.NewMusicUsecase(musicService)
	discordHandler := handler.NewDiscordHandler(musicUC)
	messageHandler := discordHandler.MessageCreate

	// Register handler
	discordService.RegisterHandler(messageHandler)

	// Start bot
	err := discordService.Start()
	if err != nil {
		fmt.Println("Error starting bot:", err)
	}
}
