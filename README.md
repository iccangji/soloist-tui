# Spotify Soloist TUI

Wrapper [Spotify Soloist](https://developer.spotify.com/documentation/soloist) TUI via WebSocket API. 

## Features
- Connects to a locally‑running Soloist instance (`--ws 127.0.0.1:9090`).
- Displays current track, artist, album, playback state and progress bar.
- Keyboard shortcuts for playback control:
  - `Space` – Play / Pause
  - `n` / `p` – Next / Previous track
  - `←` / `→` – Seek backward / forward (10 s)
  - `↑` / `↓` – Volume up / down (step 5 %)
  - `s` – Toggle shuffle
  - `r` – Cycle repeat mode (off → context → track → off)
  - `q` – Quit
- Handles reconnection with exponential back‑off.
- Shows clear error messages when Soloist is unavailable or authentication fails.
- Launches Soloist automatically using the API key from `SOLOIST_API_KEY`.

## Prerequisites
- **Spotify Soloist** binary installed and accessible in `$PATH`.
- A **Soloist API key** (generated from the Spotify Developer Dashboard).

## Quick start
```bash
# Create an .env file (or export variables)
cat > .env <<EOF
SOLOIST_API_KEY=your_api_key_here
SOLOIST_DEVICE_NAME=you_device_name
SOLOIST_WS_URL=ws://127.0.0.1:9090
EOF

# Build the TUI
go build -o bin/spotify-tui cmd/tui/main.go

# Run the program – Soloist will be started automatically
./bin/spotify-tui
```

The program will spawn the Soloist process, connect to it, and render a compact UI in your terminal.

## Development
- Run tests: `go test ./...`
- Lint/format: `go fmt ./... && go vet ./...`
- The code follows the Elm‑architecture pattern using **Bubble Tea**.

## Issues
- Handle commands require authentication
- Handle queue changed event
