package service

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/config"
)

type DiscordServiceInterface interface {
	RegisterHandler(handler any)
	Start() error
	Stop()
	GetSession() *discordgo.Session
	GetToken() string
}

type DiscordService struct {
	Client *discordgo.Session
	Token  string
}

// NewDiscordService membuat service baru dengan inisialisasi discordgo.Session
func NewDiscordService(ms *MusicService) *DiscordService {
	cfg := config.LoadDiscordConfig()

	dg, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return nil
	}

	return &DiscordService{
		Client: dg,
		Token:  cfg.BotToken,
	}
}

// RegisterHandler menambahkan handler ke client
func (s *DiscordService) RegisterHandler(handler any) {
	s.Client.AddHandler(handler)
}

// Start menjalankan koneksi bot
func (s *DiscordService) Start() error {
	err := s.Client.Open()
	if err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")

	// Tunggu signal exit
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	return nil
}

// Stop menutup koneksi bot
func (s *DiscordService) Stop() {
	s.Client.Close()
	fmt.Println("Bot stopped gracefully")
}

func (s *DiscordService) GetSession() *discordgo.Session {
	return s.Client
}

func (s *DiscordService) GetToken() string {
	return s.Token
}
