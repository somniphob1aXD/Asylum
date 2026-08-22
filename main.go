package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/mattn/go-runewidth"
)

const (
	PROXY_ENV = "SC_PROXY"
	CACHE_DIR = ".sc-tui-cache"
	FAV_FILE  = "favorites.json"
)

var (
	currentVolume float64 = 1.0
	volumeMu      sync.Mutex
	proxyURL      string
	httpClient    *http.Client
	program       *tea.Program
)

//go:embed gengar.txt
var gengarArt string

type Track struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	URL        string `json:"url"`
	ArtworkURL string `json:"artwork_url"`
	Duration   int    `json:"duration"`
	IsFavorite bool   `json:"is_favorite"`
}

type AudioPlayer struct {
	streamer beep.StreamSeeker
	ctrl     *beep.Ctrl
	vol      *VolumeStreamer
	format   beep.Format
	done     chan bool
	tmpFile  string
}

type VolumeStreamer struct {
	streamer beep.Streamer
	volume   float64
	err      error
}

func (v *VolumeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0] *= v.volume
		samples[i][1] *= v.volume
	}
	return n, ok
}

func (v *VolumeStreamer) Err() error {
	return v.err
}

func (v *VolumeStreamer) SetVolume(vol float64) {
	v.volume = vol
}

var player *AudioPlayer

func checkDependencies() error {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp not found. Install: sudo pacman -S yt-dlp")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found. Install: sudo pacman -S ffmpeg")
	}
	return nil
}

func initHTTPClient() {
	proxyURL = os.Getenv(PROXY_ENV)
	if proxyURL != "" {
		p, err := url.Parse(proxyURL)
		if err == nil {
			httpClient = &http.Client{
				Transport: &http.Transport{Proxy: http.ProxyURL(p)},
				Timeout:   60 * time.Second,
			}
			return
		}
	}
	httpClient = &http.Client{Timeout: 60 * time.Second}
}

func getYtDlpArgs() []string {
	args := []string{
		"--no-warnings",
		"--no-check-certificates",
	}
	if proxyURL != "" {
		args = append(args, "--proxy", proxyURL)
	}
	return args
}

func favoritesPath() string {
	return filepath.Join(CACHE_DIR, FAV_FILE)
}

func loadFavorites() []Track {
	os.MkdirAll(CACHE_DIR, 0755)
	data, err := os.ReadFile(favoritesPath())
	if err != nil {
		return []Track{}
	}
	var favs []Track
	if err := json.Unmarshal(data, &favs); err != nil {
		return []Track{}
	}
	return favs
}

func saveFavorites(favs []Track) {
	os.MkdirAll(CACHE_DIR, 0755)
	data, err := json.MarshalIndent(favs, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(favoritesPath(), data, 0644)
}

func isFavorite(trackURL string, favs []Track) bool {
	for _, t := range favs {
		if t.URL == trackURL {
			return true
		}
	}
	return false
}

func toggleFavorite(track Track, favs []Track) []Track {
	for i, t := range favs {
		if t.URL == track.URL {
			return append(favs[:i], favs[i+1:]...)
		}
	}
	track.IsFavorite = true
	return append(favs, track)
}

func searchTracks(query string) ([]Track, error) {
	args := getYtDlpArgs()
	args = append(args, "--flat-playlist", "--dump-json")

	if strings.Contains(query, "soundcloud.com") {
		args = append(args, query)
	} else {
		args = append(args, fmt.Sprintf("scsearch20:%s", query))
	}

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp: %s", string(output))
	}

	return parseYtDlpJSON(output)
}

type YtDlpEntry struct {
	ID         string       `json:"id"`
	Title      string       `json:"title"`
	Uploader   string       `json:"uploader"`
	Channel    string       `json:"channel"`
	Thumbnail  string       `json:"thumbnail"`
	Duration   float64      `json:"duration"`
	URL        string       `json:"url"`
	WebpageURL string       `json:"webpage_url"`
	Extractor  string       `json:"extractor"`
	Entries    []YtDlpEntry `json:"entries"`
}

func parseYtDlpJSON(output []byte) ([]Track, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tracks []Track

	favs := loadFavorites()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry YtDlpEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Extractor != "soundcloud" && !strings.Contains(entry.WebpageURL, "soundcloud.com") {
			continue
		}

		if len(entry.Entries) > 0 {
			for _, e := range entry.Entries {
				track := entryToTrack(&e)
				track.IsFavorite = isFavorite(track.URL, favs)
				if track.URL != "" {
					tracks = append(tracks, track)
				}
			}
		} else {
			track := entryToTrack(&entry)
			track.IsFavorite = isFavorite(track.URL, favs)
			if track.URL != "" {
				tracks = append(tracks, track)
			}
		}
	}

	return tracks, nil
}

func entryToTrack(entry *YtDlpEntry) Track {
	artist := entry.Uploader
	if artist == "" {
		artist = entry.Channel
	}

	return Track{
		ID:         entry.ID,
		Title:      entry.Title,
		Artist:     artist,
		URL:        entry.WebpageURL,
		ArtworkURL: entry.Thumbnail,
		Duration:   int(entry.Duration),
	}
}

func downloadTrack(trackURL string) (string, error) {
	os.MkdirAll(CACHE_DIR, 0755)

	tmpFile := filepath.Join(CACHE_DIR, fmt.Sprintf("track-%d.mp3", time.Now().UnixNano()))

	args := getYtDlpArgs()
	args = append(args,
		"-x",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", tmpFile,
		"--no-playlist",
		"--force-overwrites",
		trackURL)

	cmd := exec.Command("yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp error: %s", string(output))
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		return "", fmt.Errorf("file not created: %w", err)
	}

	if info.Size() < 1000 {
		os.Remove(tmpFile)
		return "", fmt.Errorf("file too small")
	}

	return tmpFile, nil
}

func playTrack(track Track) error {
	if player != nil && player.ctrl != nil {
		speaker.Lock()
		player.ctrl.Paused = true
		speaker.Unlock()
	}

	tmpFile, err := downloadTrack(track.URL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("open file: %w", err)
	}

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		file.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("decode audio: %w", err)
	}

	if player == nil {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	}

	volStreamer := &VolumeStreamer{streamer: streamer, volume: getCurrentVolume()}
	ctrl := &beep.Ctrl{Streamer: volStreamer}
	speaker.Play(ctrl)

	player = &AudioPlayer{
		streamer: streamer,
		ctrl:     ctrl,
		vol:      volStreamer,
		format:   format,
		done:     make(chan bool, 1),
		tmpFile:  tmpFile,
	}

	go func() {
		for {
			speaker.Lock()
			pos := streamer.Position()
			speaker.Unlock()

			if pos >= streamer.Len() {
				if program != nil {
					program.Send(nextTrackMsg{})
				}
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	return nil
}

func getCurrentVolume() float64 {
	volumeMu.Lock()
	defer volumeMu.Unlock()
	return currentVolume
}

func setVolume(vol float64) {
	volumeMu.Lock()
	currentVolume = vol
	volumeMu.Unlock()

	if player != nil && player.vol != nil {
		if vol < 0 {
			vol = 0
		}
		if vol > 2.0 {
			vol = 2.0
		}
		player.vol.SetVolume(vol)
	}
}

func togglePause() {
	if player != nil && player.ctrl != nil {
		speaker.Lock()
		player.ctrl.Paused = !player.ctrl.Paused
		speaker.Unlock()
	}
}

func seek(seconds int) {
	if player != nil && player.streamer != nil {
		speaker.Lock()
		newPos := player.streamer.Position() + player.format.SampleRate.N(time.Duration(seconds)*time.Second)
		if newPos < 0 {
			newPos = 0
		}
		player.streamer.Seek(newPos)
		speaker.Unlock()
	}
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "??:??"
	}
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func getMaxArtworkURL(artworkURL string) string {
	if artworkURL == "" {
		return ""
	}
	if strings.Contains(artworkURL, "ytimg.com") {
		re := regexp.MustCompile(`(maxresdefault|sddefault|hqdefault|mqdefault|default)\.jpg`)
		return re.ReplaceAllString(artworkURL, "maxresdefault.jpg")
	}
	re := regexp.MustCompile(`-(large|t\d+x\d+|original)\.(jpg|png)$`)
	return re.ReplaceAllString(artworkURL, "-t500x500.$2")
}

type ViewMode int

const (
	ModeSearch ViewMode = iota
	ModeFavorites
)

type model struct {
	tracks      []Track
	cursor      int
	currentIdx  int
	isPlaying   bool
	volume      float64
	searchInput textinput.Model
	progress    progress.Model
	searching   bool
	loading     bool
	err         string
	coverANSI   string
	width       int
	height      int
	viewMode    ViewMode
	favorites   []Track
}

type tickMsg time.Time
type tracksLoadedMsg []Track
type errMsg string
type nextTrackMsg struct{}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search SoundCloud or paste URL..."
	ti.Focus()
	ti.CharLimit = 150
	ti.Width = 50

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	proxyStatus := ""
	if proxyURL != "" {
		proxyStatus = fmt.Sprintf(" (proxy: %s)", proxyURL)
	}

	return model{
		searchInput: ti,
		progress:    p,
		volume:      1.0,
		coverANSI:   generateGengarASCII(),
		loading:     true,
		err:         proxyStatus,
		viewMode:    ModeSearch,
		favorites:   loadFavorites(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	}))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 4
		return m, nil

	case tickMsg:
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}))
		return m, tea.Batch(cmds...)

	case tracksLoadedMsg:
		m.tracks = msg
		m.cursor = 0
		m.searching = false
		m.loading = false
		m.err = ""
		return m, nil

	case errMsg:
		m.err = string(msg)
		m.searching = false
		m.loading = false
		return m, nil

	case nextTrackMsg:
		if m.currentIdx < len(m.tracks)-1 {
			m.currentIdx++
			m.cursor = m.currentIdx
			m.isPlaying = true
			go func() {
				if m.currentIdx < len(m.tracks) {
					track := m.tracks[m.currentIdx]
					playTrack(track)
				}
			}()
		} else {
			m.isPlaying = false
		}
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter":
				query := m.searchInput.Value()
				if query != "" {
					m.loading = true
					return m, m.search(query)
				}
			case "esc":
				m.searching = false
			case "ctrl+c":
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "/":
			m.viewMode = ModeSearch
			m.searching = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, textinput.Blink

		case "F":
			m.viewMode = ModeFavorites
			m.tracks = m.favorites
			m.cursor = 0
			m.err = ""
			if len(m.tracks) == 0 {
				m.err = "No favorites yet. Press 'f' on a track to add it."
			}
			return m, nil

		case "esc":
			if m.viewMode == ModeFavorites {
				m.viewMode = ModeSearch
				m.err = ""
			}
			return m, nil

		case "f":
			if len(m.tracks) > 0 && m.cursor < len(m.tracks) {
				track := m.tracks[m.cursor]
				m.favorites = toggleFavorite(track, m.favorites)
				saveFavorites(m.favorites)

				for i := range m.tracks {
					m.tracks[i].IsFavorite = isFavorite(m.tracks[i].URL, m.favorites)
				}

				if m.viewMode == ModeFavorites {
					m.tracks = m.favorites
					if m.cursor >= len(m.tracks) && len(m.tracks) > 0 {
						m.cursor = len(m.tracks) - 1
					}
				}
			}
			return m, nil

		case "j", "down":
			if m.cursor < len(m.tracks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if len(m.tracks) > 0 {
				m.currentIdx = m.cursor
				m.isPlaying = true
				go func() {
					track := m.tracks[m.currentIdx]
					playTrack(track)
				}()
			}
			return m, nil

		case " ", "p":
			togglePause()
			m.isPlaying = !m.isPlaying
		case "l", "right":
			seek(5)
		case "h", "left":
			seek(-5)
		case "n":
			if m.currentIdx < len(m.tracks)-1 {
				m.currentIdx++
				m.cursor = m.currentIdx
				m.isPlaying = true
				go func() {
					track := m.tracks[m.currentIdx]
					playTrack(track)
				}()
			}
			return m, nil

		case "N":
			if m.currentIdx > 0 {
				m.currentIdx--
				m.cursor = m.currentIdx
				m.isPlaying = true
				go func() {
					track := m.tracks[m.currentIdx]
					playTrack(track)
				}()
			}
			return m, nil

		case "+", "=":
			m.volume += 0.1
			if m.volume > 2.0 {
				m.volume = 2.0
			}
			setVolume(m.volume)
		case "-", "_":
			m.volume -= 0.1
			if m.volume < 0 {
				m.volume = 0
			}
			setVolume(m.volume)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) search(query string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := searchTracks(query)
		if err != nil {
			return errMsg(err.Error())
		}
		if len(tracks) == 0 {
			return errMsg("Nothing found")
		}
		return tracksLoadedMsg(tracks)
	}
}

func getCoverWidth(art string) int {
	maxW := 0
	for _, line := range strings.Split(art, "\n") {
		w := runewidth.StringWidth(line)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

func generateGengarASCII() string {
	return strings.TrimRight(gengarArt, "\n")
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var b strings.Builder

	coverWidth := getCoverWidth(m.coverANSI)
	coverWidth += 4

	infoStyle := lipgloss.NewStyle().
		Padding(1, 0).
		Width(m.width - coverWidth - 2)

	title := "No Track"
	artist := ""
	if len(m.tracks) > 0 && m.currentIdx < len(m.tracks) {
		title = m.tracks[m.currentIdx].Title
		artist = m.tracks[m.currentIdx].Artist
	}

	modeLabel := "[soundcloud]"
	if m.viewMode == ModeFavorites {
		modeLabel = "[favorites]"
	}

	info := fmt.Sprintf("%s\n\n%s\n\n%s\n\nVolume: %s",
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render(title),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(artist),
		lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(modeLabel),
		getVolumeBar(m.volume))

	coverStyled := lipgloss.NewStyle().
		Width(coverWidth).
		Padding(1, 2).
		Render(m.coverANSI)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, coverStyled, infoStyle.Render(info))
	b.WriteString(layout + "\n\n")

	if player != nil && player.streamer != nil {
		speaker.Lock()
		pos := player.streamer.Position()
		total := player.streamer.Len()
		speaker.Unlock()

		percent := float64(pos) / float64(total)
		if total == 0 {
			percent = 0
		}

		curTime := player.format.SampleRate.D(pos).Round(time.Second)
		totalTime := player.format.SampleRate.D(total).Round(time.Second)

		timeStr := fmt.Sprintf("%s / %s", curTime, totalTime)
		b.WriteString(m.progress.ViewAs(percent) + " " + timeStr + "\n\n")
	} else {
		b.WriteString(m.progress.ViewAs(0) + " 0:00 / 0:00\n\n")
	}

	if m.err != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⚠ "+m.err) + "\n\n")
	}

	if m.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("⏳ Loading...") + "\n\n")
	}

	if m.searching {
		b.WriteString("Search: " + m.searchInput.View() + "\n\n")
	} else if m.viewMode == ModeFavorites {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Favorites mode. Press 'F' to search, 'f' to toggle, 'Esc' to exit") + "\n\n")
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press / to search, F for favorites, f to toggle favorite, j/k navigate, enter to play, q to quit") + "\n\n")
	}

	listStyle := lipgloss.NewStyle().Padding(0, 2)
	var list strings.Builder
	for i, t := range m.tracks {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = "▶ "
			style = style.Foreground(lipgloss.Color("212")).Bold(true)
		}
		if i == m.currentIdx && m.isPlaying {
			cursor = "♫ "
			style = style.Foreground(lipgloss.Color("82"))
		}

		favIcon := "  "
		if t.IsFavorite {
			favIcon = "★ "
		}

		duration := formatDuration(t.Duration)
		list.WriteString(fmt.Sprintf("%s%s%s - %s %s\n", cursor, favIcon, t.Artist, t.Title, duration))
	}
	b.WriteString(listStyle.Render(list.String()))

	return b.String()
}

func getVolumeBar(vol float64) string {
	bars := int(vol * 10)
	if bars > 10 {
		bars = 10
	}
	if bars < 0 {
		bars = 0
	}
	return "[" + strings.Repeat("█", bars) + strings.Repeat("░", 10-bars) + "]"
}

func main() {
	if err := checkDependencies(); err != nil {
		fmt.Println("❌", err)
		os.Exit(1)
	}

	initHTTPClient()

	fmt.Println("🎵 SoundCloud TUI Player (yt-dlp)")
	if proxyURL != "" {
		fmt.Printf(" Using proxy: %s\n", proxyURL)
	}

	os.MkdirAll(CACHE_DIR, 0755)

	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	program = p

	go func() {
		tracks, err := searchTracks("phonk")
		if err != nil {
			p.Send(errMsg(err.Error()))
			return
		}
		if len(tracks) > 0 {
			p.Send(tracksLoadedMsg(tracks))
		} else {
			p.Send(errMsg("Nothing found"))
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
