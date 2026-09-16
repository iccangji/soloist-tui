package tui

import "github.com/charmbracelet/bubbletea"

// KeyMap defines keybindings for the TUI.
type KeyMap struct {
	PlayPause      tea.KeyMsg
	NextTrack      tea.KeyMsg
	PrevTrack      tea.KeyMsg
	SeekForward    tea.KeyMsg
	SeekBackward   tea.KeyMsg
	VolumeUp       tea.KeyMsg
	VolumeDown     tea.KeyMsg
	ToggleShuffle  tea.KeyMsg
	CycleRepeat    tea.KeyMsg
	Quit           tea.KeyMsg
}
