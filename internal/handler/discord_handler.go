package handler

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/usecase"
)

type DiscordHandler struct {
	musicUC *usecase.MusicUsecase
}

func NewDiscordHandler(musicUC *usecase.MusicUsecase) *DiscordHandler {
	return &DiscordHandler{
		musicUC: musicUC,
	}
}

func (h *DiscordHandler) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, ";play") {
		args := strings.SplitN(m.Content, " ", 2)
		if len(args) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: ;play <url>")
			return
		}
		err := h.musicUC.PlayMusic(s, m.GuildID, m.Author.ID, m.Author.Username, args[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, "🎶 Playing: "+args[1])
		}
	}

	if strings.HasPrefix(m.Content, ";stop") {
		err := h.musicUC.StopMusic()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, "⏹️ Stopped music")
		}
	}

	if strings.HasPrefix(m.Content, ";jokes") {
		s.ChannelMessageSend(m.ChannelID, "HIDUP JOKOWI")
	}
}
