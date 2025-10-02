package service

import "github.com/bwmarrin/discordgo"

// kontrak: MusicServiceInterface
type MusicServiceInterface interface {
	Join(session *discordgo.Session, guildID, channelID string) error
	Leave() error
	Play(source string) error
}

// implementasi
type MusicService struct {
	voice *discordgo.VoiceConnection
}

func NewMusicService() *MusicService {
	return &MusicService{}
}

func (m *MusicService) Join(session *discordgo.Session, guildID, channelID string) error {
	vc, err := session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	m.voice = vc
	return nil
}

func (m *MusicService) Leave() error {
	if m.voice != nil {
		m.voice.Disconnect()
		m.voice = nil
	}
	return nil
}

func (m *MusicService) Play(source string) error {
	// nanti implementasi stream audio (pakai ffmpeg)
	return nil
}
