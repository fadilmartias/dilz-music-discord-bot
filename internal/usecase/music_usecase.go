package usecase

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/service"
)

type MusicUsecase struct {
	musicService service.MusicServiceInterface
}

func NewMusicUsecase(ms service.MusicServiceInterface) *MusicUsecase {
	return &MusicUsecase{musicService: ms}
}

func (u *MusicUsecase) PlayMusic(session *discordgo.Session, guildID, userID, url string) error {
	// cek user ada di voice channel mana
	vs, _ := session.State.VoiceState(guildID, userID)
	if vs == nil {
		return fmt.Errorf("user belum join voice channel")
	}

	// join channel dulu kalau belum
	err := u.musicService.Join(session, guildID, vs.ChannelID)
	if err != nil {
		return fmt.Errorf("gagal join: %w", err)
	}

	// mainkan musik
	return u.musicService.Play(url)
}

func (u *MusicUsecase) StopMusic() error {
	return u.musicService.Leave()
}
