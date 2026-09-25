package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iccangji/soloist-tui/internal/tui"
)

func main() {
	apiKey := os.Getenv("SOLOIST_API_KEY")
	deviceName := os.Getenv("SOLOIST_DEVICE_NAME")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: SOLOIST_API_KEY environment variable not set")
		os.Exit(1)
	}
	if deviceName == "" {
		deviceName = "Soloist"
	}

	// Start Soloist process in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	soloistCmd := exec.CommandContext(ctx, "soloist",
		"--device-name", deviceName,
		"--api-key", apiKey,
		"--ws", "127.0.0.1:9090",
	)
	go func() {
		<-time.After(2 * time.Second)
		soloistCmd.Stdout = os.Stdout
		soloistCmd.Stderr = os.Stderr
	}()

	if err := soloistCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start Soloist: %v\n", err)
		os.Exit(1)
	}

	// Ensure Soloist is terminated on exit
	defer func() {
		// Attempt graceful shutdown
		_ = soloistCmd.Process.Signal(os.Interrupt)
		// Wait up to 2 seconds, then kill if still alive
		done := make(chan error, 1)
		go func() { done <- soloistCmd.Wait() }()
		select {
		case <-time.After(2 * time.Second):
			_ = soloistCmd.Process.Kill()
		case <-done:
			// exited cleanly
		}
	}()

	// Small pause to give Soloist time to start listening
	time.Sleep(500 * time.Millisecond)

	model := tui.Model{}
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
