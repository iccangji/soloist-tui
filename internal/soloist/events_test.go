package soloist

import (
	"testing"
)

func TestParseEventPlaybackState(t *testing.T) {
	raw := []byte(`{
		"type": "playback_state",
		"status": "playing",
		"position": {
			"position_ms": 12345,
			"timestamp_ms": 1747654321000,
			"speed": 1.0
		},
		"item": {
			"uri": "spotify:track:123",
			"entity_type": "track",
			"decorations": {
				"identity": {
					"name": "Test Track"
				},
				"playback": {
					"duration_ms": 60000
				}
			}
		}
	}`)

	ev, typ, err := ParseEvent(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing playback_state: %v", err)
	}
	if typ != EventPlaybackState {
		t.Errorf("expected type playback_state, got %s", typ)
	}

	state, ok := ev.(PlaybackStateEvent)
	if !ok {
		t.Fatalf("expected PlaybackStateEvent, got %T", ev)
	}
	if state.Status != "playing" {
		t.Errorf("expected status 'playing', got '%s'", state.Status)
	}
	if state.Position.PositionMs != 12345 {
		t.Errorf("expected position_ms 12345, got %d", state.Position.PositionMs)
	}
	if state.Item.GetName() != "Test Track" {
		t.Errorf("expected track name 'Test Track', got '%s'", state.Item.GetName())
	}
}

func TestParseEventVolumeChanged(t *testing.T) {
	raw := []byte(`{
		"type": "volume_changed",
		"volume": 75
	}`)

	ev, typ, err := ParseEvent(raw)
	if err != nil {
		t.Fatalf("unexpected error parsing volume_changed: %v", err)
	}
	if typ != EventVolumeChanged {
		t.Errorf("expected type volume_changed, got %s", typ)
	}

	state, ok := ev.(VolumeChangedEvent)
	if !ok {
		t.Fatalf("expected VolumeChangedEvent, got %T", ev)
	}
	if state.Volume != 75 {
		t.Errorf("expected volume 75, got %d", state.Volume)
	}
}
