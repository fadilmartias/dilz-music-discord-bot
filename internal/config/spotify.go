package config

import (
	"os"
	"sync"
)

type SpotifyConfig struct {
	ClientID     string
	ClientSecret string
}

var (
	spotifyConfig *SpotifyConfig
	spotifyOnce   sync.Once
)

func LoadSpotifyConfig() *SpotifyConfig {
	spotifyOnce.Do(func() {
		spotifyConfig = &SpotifyConfig{
			ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
			ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		}
	})
	return spotifyConfig
}
