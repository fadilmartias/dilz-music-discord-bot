package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/encoder"
	"github.com/fadilmartias/dilz-music-discord-bot/internal/util"
)

type Track struct {
	Title     string
	URL       string
	Requester string
}

type LoopMode int

const (
	LoopOff LoopMode = iota
	LoopTrack
	LoopQueue
)

type MusicServiceInterface interface {
	Join(session *discordgo.Session, guildID, channelID string) error
	Leave() error
	Play(source string, requester string) error
	AddToQueue(source, requester string) error
	AddPlaylist(playlistURL, requester string) error
	Skip() error
	Pause()
	Resume()
	SetLoop(mode LoopMode)
	GetQueue() []*Track
	GetNowPlaying() *Track
	IsPlaying() bool
	IsPaused() bool
}

type MusicService struct {
	voice        *discordgo.VoiceConnection
	session      *discordgo.Session
	handlerAdded bool
	queue        []*Track
	nowPlaying   *Track
	isPlaying    bool
	isPaused     bool
	loopMode     LoopMode
	mu           sync.Mutex
	ffmpegCmd    *exec.Cmd
	shouldStop   bool
	shouldSkip   bool
	ds           DiscordServiceInterface
}

func NewMusicService() *MusicService {
	return &MusicService{
		queue:    make([]*Track, 0),
		loopMode: LoopOff,
	}
}

func (m *MusicService) SetDiscordService(ds DiscordServiceInterface) {
	m.ds = ds
}

// Join - connect ke voice channel
func (m *MusicService) Join(session *discordgo.Session, guildID, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Disconnect dulu kalau udah connect
	if m.voice != nil {
		m.voice.Disconnect()
		time.Sleep(500 * time.Millisecond)
	}

	vc, err := session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	m.voice = vc
	m.session = session

	// Register handler cuma sekali
	if !m.handlerAdded {
		session.AddHandler(m.HandleVoiceStateUpdate)
		m.handlerAdded = true
	}

	fmt.Println("✅ Connected to voice channel")
	return nil
}

// Leave - disconnect dari voice channel
func (m *MusicService) Leave() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.voice != nil {
		m.shouldStop = true

		// Kill ffmpeg kalau lagi jalan
		m.killFfmpeg()

		// Disconnect dari voice
		err := m.voice.Disconnect()
		m.voice = nil

		fmt.Println("👋 Disconnected from voice channel")

		return err
	}
	return nil
}

// Play - langsung play tanpa queue
func (m *MusicService) Play(source, requester string) error {
	m.mu.Lock()
	m.queue = make([]*Track, 0)
	m.mu.Unlock()

	return m.AddToQueue(source, requester)
}

// AddToQueue - tambah ke queue
func (m *MusicService) AddToQueue(source, requester string) error {
	track, err := m.getTrackInfo(source)
	if err != nil {
		return err
	}
	track.Requester = requester

	m.mu.Lock()
	m.queue = append(m.queue, track)
	shouldStart := !m.isPlaying
	m.mu.Unlock()

	fmt.Printf("✅ Added to queue: %s\n", track.Title)

	if shouldStart {
		go m.playQueue()
	}

	return nil
}

// AddPlaylist - tambah playlist ke queue
func (m *MusicService) AddPlaylist(playlistURL, requester string) error {
	tracks, err := m.getPlaylistTracks(playlistURL)
	if err != nil {
		return err
	}

	m.mu.Lock()
	for _, track := range tracks {
		track.Requester = requester
		m.queue = append(m.queue, track)
	}
	shouldStart := !m.isPlaying
	m.mu.Unlock()

	fmt.Printf("✅ Added %d tracks to queue\n", len(tracks))

	if shouldStart {
		go m.playQueue()
	}

	return nil
}

func (m *MusicService) getTrackInfo(source string) (*Track, error) {
	if !strings.HasPrefix(source, "http") {
		source = "ytsearch:" + source
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Ambil title DAN webpage_url sekaligus
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-check-certificate",
		"--no-playlist",
		"--get-title",
		"--get-id", // Ambil video ID
		source,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp error: %v - %s", err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid yt-dlp output")
	}

	// Simpan URL proper, bukan ytsearch
	videoURL := "https://www.youtube.com/watch?v=" + lines[1]

	return &Track{
		Title: lines[0],
		URL:   videoURL, // Simpan URL proper, bukan ytsearch
	}, nil
}

func (m *MusicService) getPlaylistTracks(playlistURL string) ([]*Track, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-check-certificate",
		"--flat-playlist",
		"--get-title",
		"--get-id",
		playlistURL,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	tracks := make([]*Track, 0)

	for i := 0; i < len(lines); i += 2 {
		if i+1 >= len(lines) {
			break
		}
		track := &Track{
			Title: lines[i],
			URL:   "https://www.youtube.com/watch?v=" + lines[i+1],
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (m *MusicService) playQueue() {
	for {
		m.mu.Lock()

		if len(m.queue) == 0 {
			m.isPlaying = false
			m.nowPlaying = nil
			m.mu.Unlock()
			fmt.Println("Queue selesai")
			return
		}

		track := m.queue[0]
		m.queue = m.queue[1:]
		m.nowPlaying = track
		m.isPlaying = true
		m.shouldSkip = false
		m.shouldStop = false
		m.mu.Unlock()

		fmt.Printf("🎵 Now playing: %s\n", track.Title)

		err := m.playTrack(track)
		if err != nil {
			fmt.Printf("❌ Error playing: %v\n", err)
		}

		m.mu.Lock()
		// Kalau di-stop, clear queue
		if m.shouldStop {
			m.queue = make([]*Track, 0)
			m.isPlaying = false
			m.nowPlaying = nil
			m.mu.Unlock()
			return
		}

		// Handle loop mode (kalau ga di-skip)
		if !m.shouldSkip {
			switch m.loopMode {
			case LoopTrack:
				m.queue = append([]*Track{track}, m.queue...)
			case LoopQueue:
				m.queue = append(m.queue, track)
			}
		}
		m.mu.Unlock()
	}
}

func (m *MusicService) playTrack(track *Track) error {
	if m.voice == nil {
		return fmt.Errorf("not connected to voice channel")
	}

	// Retry mechanism untuk get stream URL
	var url string
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("🔄 Retry %d/%d...\n", i+1, maxRetries)
			time.Sleep(2 * time.Second)
		}

		fmt.Printf("⏳ Getting stream URL for: %s\n", track.Title)
		url, err = m.getStreamURL(track.URL)
		if err == nil {
			break
		}
		fmt.Printf("⚠️ Attempt %d failed: %v\n", i+1, err)
	}

	if err != nil {
		return fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
	}

	fmt.Printf("✅ Got stream URL, starting playback...\n")

	m.ffmpegCmd = exec.Command("ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", url,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"-loglevel", "warning",
		"pipe:1")

	ffmpegout, err := m.ffmpegCmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Capture stderr untuk debug
	ffmpegErr, _ := m.ffmpegCmd.StderrPipe()
	go func() {
		scanner := bufio.NewScanner(ffmpegErr)
		for scanner.Scan() {
			fmt.Println("ffmpeg:", scanner.Text())
		}
	}()

	err = m.ffmpegCmd.Start()
	if err != nil {
		return err
	}

	return m.streamAudio(ffmpegout)
}

func (m *MusicService) getStreamURL(source string) (string, error) {
	// Sekarang source udah proper URL, ga perlu timeout lama
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--no-check-certificate",
		"--no-playlist",
		"-f", "bestaudio[ext=webm]/bestaudio/best",
		"--get-url",
		source, // Source udah proper URL, bukan ytsearch
	)

	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("yt-dlp timeout after 20s")
	}

	if err != nil {
		return "", fmt.Errorf("yt-dlp error: %v - %s", err, string(out))
	}

	url := strings.TrimSpace(string(out))
	if url == "" {
		return "", fmt.Errorf("empty stream URL")
	}

	return url, nil
}

func (m *MusicService) streamAudio(stream io.Reader) error {
	reader := bufio.NewReader(stream)

	m.voice.Speaking(true)
	defer m.voice.Speaking(false)

	opusEncoder, err := encoder.NewOpusEncoder()
	if err != nil {
		return err
	}

	buffer := make([]int16, 960*2)

	for {
		// Check stop/skip flags
		m.mu.Lock()
		if m.shouldStop || m.shouldSkip {
			m.mu.Unlock()
			m.killFfmpeg()
			return nil
		}
		paused := m.isPaused
		m.mu.Unlock()

		// Kalau pause, sleep bentar
		if paused {
			m.voice.Speaking(false)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		m.voice.Speaking(true)

		// Baca PCM data
		err := util.ReadPCM16(reader, buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Encode ke opus
		opus, err := opusEncoder.Encode(buffer, 960, 960*2*2)
		if err != nil {
			continue
		}

		m.voice.OpusSend <- opus
	}

	return m.ffmpegCmd.Wait()
}

func (m *MusicService) killFfmpeg() {
	if m.ffmpegCmd != nil && m.ffmpegCmd.Process != nil {
		m.ffmpegCmd.Process.Kill()
	}
}

// HandleVoiceStateUpdate - auto pause saat bot dimute
// HandleVoiceStateUpdate - auto pause saat bot dimute
func (m *MusicService) HandleVoiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.voice == nil || m.session == nil {
		return
	}

	// Check kalau yang update adalah bot sendiri
	if vsu.UserID != s.State.User.ID {
		return
	}

	// Check kalau di guild yang sama
	if vsu.GuildID != m.voice.GuildID {
		return
	}

	// Auto pause kalau dimute
	if vsu.Mute || vsu.SelfMute {
		if !m.isPaused {
			fmt.Println("🔇 Bot dimute, auto pause")
			m.isPaused = true
		}
	} else if m.isPaused && m.isPlaying {
		// Auto resume kalau unmute
		fmt.Println("🔊 Bot unmute, auto resume")
		m.isPaused = false
	}
}

// Playback controls
func (m *MusicService) Skip() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isPlaying {
		return fmt.Errorf("nothing playing")
	}

	m.shouldSkip = true
	return nil
}

func (m *MusicService) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isPlaying {
		return fmt.Errorf("nothing playing")
	}

	m.shouldStop = true
	return nil
}

func (m *MusicService) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isPaused = true
}

func (m *MusicService) Resume() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isPaused = false
}

func (m *MusicService) SetLoop(mode LoopMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loopMode = mode

	modeStr := "off"
	if mode == LoopTrack {
		modeStr = "track"
	} else if mode == LoopQueue {
		modeStr = "queue"
	}
	fmt.Printf("🔁 Loop mode: %s\n", modeStr)
}

// Getters
func (m *MusicService) GetQueue() []*Track {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.queue
}

func (m *MusicService) GetNowPlaying() *Track {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nowPlaying
}

func (m *MusicService) IsPlaying() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isPlaying
}

func (m *MusicService) IsPaused() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isPaused
}
