package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fadilmartias/dilz-music-discord-bot/internal/config"
)

type SpotifyServiceInterface interface {
	GetTrackInfo(trackID string) (string, error)
}

type SpotifyService struct {
	ClientID     string
	ClientSecret string
	token        string
	expiresAt    time.Time
}

func NewSpotifyService() *SpotifyService {
	config := config.LoadSpotifyConfig()
	return &SpotifyService{ClientID: config.ClientID, ClientSecret: config.ClientSecret}
}

// ambil access token pakai client credentials
func (s *SpotifyService) getToken() error {
	if time.Now().Before(s.expiresAt) && s.token != "" {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, _ := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	req.SetBasicAuth(s.ClientID, s.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		log.Println(err)
		return err
	}

	s.token = res.AccessToken
	s.expiresAt = time.Now().Add(time.Duration(res.ExpiresIn-60) * time.Second) // refresh 1m sebelum expired
	return nil
}

// ambil track metadata
func (s *SpotifyService) GetTrackInfo(trackID string) (string, error) {
	if err := s.getToken(); err != nil {
		return "", err
	}
	log.Println(trackID)

	url := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var res struct {
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		log.Println(err)
		return "", err
	}

	if len(res.Artists) == 0 {
		return res.Name, nil
	}

	// contoh: "Shivers Ed Sheeran"
	return fmt.Sprintf("%s %s", res.Name, res.Artists[0].Name), nil
}
