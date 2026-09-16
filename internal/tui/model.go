package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iccangji/soloist-tui/internal/soloist"
)

// Model holds the UI state.
type Model struct {
	client        *soloist.Client
	connected     bool
	playing       bool
	track         soloist.Track
	position      int64 // current interpolated position used for UI
	duration      int64
	volume        int
	shuffle       bool
	repeat        string
	reconnecting  bool
	lastPosMs     int64
	lastTimestamp int64
	speed         float64
	errMsg        string
	quitting      bool
	loggedIn          bool // Spotify Connect authentication status
	reconnectAttempts int
}

// Ensure Model implements tea.Model.
var _ tea.Model = (*Model)(nil)

func (m Model) Init() tea.Cmd {
	return connectCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.client == nil || !m.connected {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}
		ctx := context.Background()
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case " ":
			if m.playing {
				_ = m.client.Pause(ctx)
			} else {
				_ = m.client.Play(ctx)
			}
		case "n":
			_ = m.client.Next(ctx)
		case "p":
			_ = m.client.Previous(ctx)
		case "right":
			newPos := m.position + 10000
			if m.duration > 0 && newPos > m.duration {
				newPos = m.duration
			}
			_ = m.client.Seek(ctx, newPos)
		case "left":
			newPos := m.position - 10000
			if newPos < 0 {
				newPos = 0
			}
			_ = m.client.Seek(ctx, newPos)
		case "up":
			newVol := m.volume + 5
			if newVol > 100 {
				newVol = 100
			}
			_ = m.client.SetVolume(ctx, newVol)
		case "down":
			newVol := m.volume - 5
			if newVol < 0 {
				newVol = 0
			}
			_ = m.client.SetVolume(ctx, newVol)
		case "s":
			_ = m.client.SetShuffle(ctx, !m.shuffle)
		case "r":
			nextRepeat := "off"
			switch m.repeat {
			case "off", "":
				nextRepeat = "context"
			case "context":
				nextRepeat = "track"
			case "track":
				nextRepeat = "off"
			}
			_ = m.client.SetRepeat(ctx, nextRepeat)
		}
		return m, nil

	case soloistMessage:
		ev, typ, err := soloist.ParseEvent(msg.data)
		if err != nil {
			m.errMsg = fmt.Sprintf("parse error: %v", err)
			if m.client != nil {
				return m, readCmd(m.client)
			}
			return m, nil
		}
		switch typ {
	case soloist.EventAuthState:
		if e, ok := ev.(soloist.AuthStateEvent); ok {
			m.loggedIn = e.LoggedIn
			m.shuffle = e.IsActive // use is_active as placeholder for active state, not needed for UI now
		}

		case soloist.EventPlaybackState:
			if e, ok := ev.(soloist.PlaybackStateEvent); ok {
				m.playing = (e.Status == "playing")
				m.track = e.Item
				m.duration = e.Item.GetDuration()
				m.position = e.Position.PositionMs
				m.volume = e.Volume
				m.shuffle = e.Options.Shuffle
				m.repeat = e.Options.Repeat
				m.lastPosMs = e.Position.PositionMs
				m.lastTimestamp = e.Position.TimestampMs
				m.speed = e.Position.Speed
			}
		case soloist.EventTrackChanged:
			if e, ok := ev.(soloist.TrackChangedEvent); ok {
				m.track = e.Item
				m.duration = e.Item.GetDuration()
			}
		case soloist.EventPlaybackChanged:
			if e, ok := ev.(soloist.PlaybackChangedEvent); ok {
				m.playing = (e.Status == "playing")
			}
		case soloist.EventVolumeChanged:
			if e, ok := ev.(soloist.VolumeChangedEvent); ok {
				m.volume = e.Volume
			}
		case soloist.EventOptionsChanged:
			if e, ok := ev.(soloist.OptionsChangedEvent); ok {
				m.shuffle = e.Options.Shuffle
				m.repeat = e.Options.Repeat
			}
		case soloist.EventPositionSync:
			if e, ok := ev.(soloist.PositionSyncEvent); ok {
				m.position = e.Position.PositionMs
				m.lastPosMs = e.Position.PositionMs
				m.lastTimestamp = e.Position.TimestampMs
				m.speed = e.Position.Speed
			}
		case soloist.EventError:
			if e, ok := ev.(soloist.ErrorEvent); ok {
				m.errMsg = e.Message
			}
		}
		if m.client != nil {
			return m, readCmd(m.client)
		}
		return m, nil

	case errMsg:
		m.errMsg = string(msg)
		m.connected = false
		if !m.reconnecting {
			m.reconnecting = true
			m.reconnectAttempts = 0
			return m, reconnectCmd(m.reconnectAttempts)
		}
		return m, nil

	case connectedMsg:
		m.client = msg.client
		m.connected = true
		m.reconnecting = false
		m.reconnectAttempts = 0
		m.errMsg = ""
		return m, tea.Batch(readCmd(m.client), getStateCmd(m.client), tickCmd())

	case disconnectedMsg:
		m.connected = false
		m.client = nil
		if !m.reconnecting {
			m.reconnecting = true
			m.reconnectAttempts = 0
			return m, reconnectCmd(m.reconnectAttempts)
		}
		return m, nil

	case reconnectMsg:
		m.reconnectAttempts = msg.attempts + 1
		return m, connectCmd()

	case tickMsg:
		if m.connected && m.speed > 0 {
			now := time.Now().UnixMilli()
			elapsed := now - m.lastTimestamp
			if elapsed < 0 {
				elapsed = 0
			}
			m.position = m.lastPosMs + int64(float64(elapsed)*m.speed)
		}
		return m, tickCmd()
	}
	return m, nil
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	status := "○ Disconnected"
	if m.connected {
		status = "● Connected"
	} else if m.reconnecting {
		status = "↻ Reconnecting..."
	}
	header := lipgloss.NewStyle().Bold(true).Render("Spotify Soloist " + status)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(64)

	// If not connected, show dedicated connection screen outside the active player box
	if !m.connected {
		errBox := ""
		if m.errMsg != "" {
			errBox = fmt.Sprintf("\n\n  Error: %s\n  Make sure Soloist is running with --ws", m.errMsg)
		}
		content := fmt.Sprintf("\n  Connecting to Spotify Soloist...\n  URL: ws://127.0.0.1:9090%s\n\n  Press q to quit.", errBox)
		return boxStyle.Render(header + "\n" + content)
	}

	artist := m.track.GetArtist()
	if artist == "" {
		artist = "Unknown Artist"
	}
	title := m.track.GetName()
	if title == "" {
		title = "No track playing"
	}
	album := m.track.GetAlbum()

	nowPlaying := fmt.Sprintf("\n  NOW PLAYING: %s\n\n  Artist: %s\n  Album:  %s\n", title, artist, album)

	progress := ""
	if m.duration > 0 {
		pct := float64(m.position) / float64(m.duration) * 100
		if pct > 100 {
			pct = 100
		}
		if pct < 0 {
			pct = 0
		}
		filled := int(pct / 100 * 25)
		if filled > 25 {
			filled = 25
		}
		empty := 25 - filled
		bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
		currMin, currSec := m.position/60000, (m.position/1000)%60
		durMin, durSec := m.duration/60000, (m.duration/1000)%60
		progress = fmt.Sprintf("  %s  %02d:%02d / %02d:%02d\n", bar, currMin, currSec, durMin, durSec)
	} else {
		progress = "  ─────────────────────────  --:-- / --:--\n"
	}

	stateIcon := "▶"
	if m.playing {
		stateIcon = "❚❚"
	}
	controls := fmt.Sprintf("\n                 ◀    %s    ▶\n", stateIcon)

	volBar := strings.Repeat("█", m.volume/5) + strings.Repeat("░", 20-(m.volume/5))
	volumeInfo := fmt.Sprintf("\n  Volume  %s  %d%%\n", volBar, m.volume)

	shufStatus := "OFF"
	if m.shuffle {
		shufStatus = "ON"
	}
	repStatus := strings.ToUpper(m.repeat)
	if repStatus == "" {
		repStatus = "OFF"
	}
	optionsInfo := fmt.Sprintf("\n  Shuffle: %-15s Repeat: %s\n", shufStatus, repStatus)

	errMsgDisplay := ""
	if m.errMsg != "" {
		// errMsgDisplay = fmt.Sprintf("\n  ⚠️  %s", m.errMsg)
	}

	footer := "\n  Space Play/Pause   n Next   p Prev   ←/→ Seek\n  ↑/↓ Volume         s Shuffle   r Repeat   q Quit"

	content := fmt.Sprintf("%s%s%s%s%s%s%s\n%s", nowPlaying, progress, controls, volumeInfo, optionsInfo, errMsgDisplay, footer, "")
	return boxStyle.Render(header + "\n" + content)
}

// --- message types ---

type soloistMessage struct{ data []byte }
type errMsg string
type connectedMsg struct{ client *soloist.Client }
type disconnectedMsg struct{}

type reconnectMsg struct{ attempts int }

// --- commands ---

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		wsURL := "ws://127.0.0.1:9090"
		if env := os.Getenv("SOLOIST_WS_URL"); env != "" {
			wsURL = env
		}
		client, err := soloist.NewClient(wsURL)
		if err != nil {
			return errMsg(err.Error())
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := client.Connect(ctx); err != nil {
			return errMsg(err.Error())
		}
		return connectedMsg{client: client}
	}
}

func getStateCmd(c *soloist.Client) tea.Cmd {
	return func() tea.Msg {
		_ = c.Send(context.Background(), map[string]interface{}{
			"type":    "command",
			"command": "get_state",
		})
		return nil
	}
}

func reconnectCmd(attempts int) tea.Cmd {
	return func() tea.Msg {
		delay := time.Duration(1<<uint(attempts)) * time.Second
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		time.Sleep(delay)
		return reconnectMsg{attempts: attempts}
	}
}

func readCmd(c *soloist.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := c.Read(context.Background())
		if err != nil {
			return errMsg(err.Error())
		}
		return soloistMessage{data: data}
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/10, func(t time.Time) tea.Msg { return tickMsg(t) })
}
