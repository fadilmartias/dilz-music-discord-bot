package usecase

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/service"
)

type MusicUsecase struct {
	musicService   service.MusicServiceInterface
	spotifyService service.SpotifyServiceInterface
}

func NewMusicUsecase(ms service.MusicServiceInterface, ss service.SpotifyServiceInterface) *MusicUsecase {
	return &MusicUsecase{musicService: ms, spotifyService: ss}
}

func (u *MusicUsecase) PlayMusic(session *discordgo.Session, guildID, userID, username, input string) error {
	// cek user ada di voice channel mana
	vs, _ := session.State.VoiceState(guildID, userID)
	if vs == nil {
		return fmt.Errorf("user belum join voice channel")
	}

	// join dulu
	if err := u.musicService.Join(session, guildID, vs.ChannelID); err != nil {
		if err != nil {
			if _, ok := session.VoiceConnections[guildID]; ok {
				return nil
			} else {
				return err
			}
		}
	}

	source := input

	// detect spotify url
	re := regexp.MustCompile(`open\.spotify\.com/track/([a-zA-Z0-9]+)`)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		trackID := match[1]
		trackName, err := u.spotifyService.GetTrackInfo(trackID)
		if err != nil {
			return fmt.Errorf("gagal ambil metadata dari Spotify: %w", err)
		}
		source = trackName
		fmt.Println(source)
	}

	// kalau bukan URL → default ke YouTube search
	if !strings.HasPrefix(source, "http") {
		source = "ytsearch:" + source
		fmt.Println(source)
	}

	return u.musicService.Play(source, username)
}

func (u *MusicUsecase) StopMusic() error {
	return u.musicService.Leave()
}
