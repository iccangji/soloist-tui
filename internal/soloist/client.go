package soloist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Client manages the WebSocket connection to Spotify Soloist.
type Client struct {
	connURL *url.URL
	conn    *websocket.Conn
	mu      sync.Mutex
	closed  bool
}

// NewClient creates a new Soloist WebSocket client.
func NewClient(wsURL string) (*Client, error) {
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid soloist ws url: %w", err)
	}
	return &Client{connURL: u}, nil
}

// Connect dials the WebSocket endpoint.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, _, err := websocket.Dial(ctx, c.connURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to dial soloist ws: %w", err)
	}
	c.conn = conn
	c.closed = false
	return nil
}

// Read reads the next raw JSON message from the WebSocket connection.
func (c *Client) Read(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	typ, p, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	if typ != websocket.MessageText {
		return p, nil
	}
	return p, nil
}

// Send marshals and sends a command payload over the WebSocket.
func (c *Client) Send(ctx context.Context, v interface{}) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	return wsjson.Write(ctx, conn, v)
}

// Close gracefully closes the WebSocket connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.conn == nil {
		return nil
	}
	c.closed = true
	return c.conn.Close(websocket.StatusNormalClosure, "TUI shutting down")
}

func (c *Client) sendCommand(ctx context.Context, payload map[string]interface{}) error {
	payload["type"] = "command"
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("not connected")
	}
	return c.conn.Write(ctx, websocket.MessageText, data)
}

// Playback and option commands matching Soloist WebSocket API reference
func (c *Client) Play(ctx context.Context) error {
	return c.sendCommand(ctx, map[string]interface{}{"command": "play"})
}

func (c *Client) Pause(ctx context.Context) error {
	return c.sendCommand(ctx, map[string]interface{}{"command": "pause"})
}

func (c *Client) Next(ctx context.Context) error {
	return c.sendCommand(ctx, map[string]interface{}{"command": "skip_next"})
}

func (c *Client) Previous(ctx context.Context) error {
	return c.sendCommand(ctx, map[string]interface{}{"command": "skip_prev"})
}

func (c *Client) Seek(ctx context.Context, positionMs int64) error {
	return c.sendCommand(ctx, map[string]interface{}{
		"command":     "seek",
		"position_ms": positionMs,
	})
}

func (c *Client) SetVolume(ctx context.Context, volume int) error {
	return c.sendCommand(ctx, map[string]interface{}{
		"command": "set_volume",
		"volume":  volume,
	})
}

func (c *Client) SetShuffle(ctx context.Context, enabled bool) error {
	return c.sendCommand(ctx, map[string]interface{}{
		"command": "set_shuffle",
		"enabled": enabled,
	})
}

func (c *Client) SetRepeat(ctx context.Context, mode string) error {
	// According to spec:
	// off: set_repeat_track(false), then set_repeat_context(false)
	// context: set_repeat_track(false), then set_repeat_context(true)
	// track: set_repeat_context(false), then set_repeat_track(true)
	var err error
	switch mode {
	case "off":
		if err = c.sendCommand(ctx, map[string]interface{}{"command": "set_repeat_track", "enabled": false}); err != nil {
			return err
		}
		return c.sendCommand(ctx, map[string]interface{}{"command": "set_repeat_context", "enabled": false})
	case "context":
		if err = c.sendCommand(ctx, map[string]interface{}{"command": "set_repeat_track", "enabled": false}); err != nil {
			return err
		}
		return c.sendCommand(ctx, map[string]interface{}{"command": "set_repeat_context", "enabled": true})
	case "track":
		if err = c.sendCommand(ctx, map[string]interface{}{"command": "set_repeat_context", "enabled": false}); err != nil {
			return err
		}
		return c.sendCommand(ctx, map[string]interface{}{"command": "set_repeat_track", "enabled": true})
	}
	return fmt.Errorf("unknown repeat mode: %s", mode)
}
