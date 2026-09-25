package soloist

import (
	"encoding/json"
	"fmt"
)

// EventType represents the category of event received from Soloist.
type EventType string

const (
	EventAuthState       EventType = "auth_state"
	EventPlaybackState   EventType = "playback_state"
	EventTrackChanged    EventType = "track_changed"
	EventPlaybackChanged EventType = "playback_changed"
	EventVolumeChanged   EventType = "volume_changed"
	EventOptionsChanged  EventType = "options_changed"
	EventPositionSync    EventType = "position_sync"
	EventError           EventType = "error"
)

type BaseEvent struct {
	Type EventType `json:"type"`
}

// Entity and decoration models matching the Soloist API spec
type Identity struct {
	Name string `json:"name"`
}

type ParentEntity struct {
	Entity struct {
		Decorations struct {
			Identity Identity `json:"identity"`
		} `json:"decorations"`
	} `json:"entity"`
}

type CreatorEntity struct {
	Entity struct {
		Decorations struct {
			Identity Identity `json:"identity"`
		} `json:"decorations"`
	} `json:"entity"`
}

type PlaybackMeta struct {
	DurationMs int64 `json:"duration_ms"`
}

type Decorations struct {
	Identity Identity      `json:"identity"`
	Parent   ParentEntity    `json:"parent"`
	Creators []CreatorEntity `json:"creators"`
	Playback PlaybackMeta    `json:"playback"`
}

type Track struct {
	URI         string      `json:"uri"`
	EntityType  string      `json:"entity_type"`
	Decorations Decorations `json:"decorations"`
}

// Helper methods to extract display info cleanly
func (t Track) GetName() string {
	return t.Decorations.Identity.Name
}

func (t Track) GetArtist() string {
	if len(t.Decorations.Creators) > 0 {
		name := t.Decorations.Creators[0].Entity.Decorations.Identity.Name
		if name != "" {
			return name
		}
	}
	return ""
}

func (t Track) GetAlbum() string {
	return t.Decorations.Parent.Entity.Decorations.Identity.Name
}

func (t Track) GetDuration() int64 {
	return t.Decorations.Playback.DurationMs
}

type PositionInfo struct {
	PositionMs  int64   `json:"position_ms"`
	TimestampMs int64   `json:"timestamp_ms"`
	Speed       float64 `json:"speed"`
}

type OptionsInfo struct {
	Shuffle bool   `json:"shuffle"`
	Repeat  string `json:"repeat"`
}

// Event structs
type AuthStateEvent struct {
	Type       EventType `json:"type"`
	LoggedIn   bool      `json:"logged_in"`
	IsActive   bool      `json:"is_active"`
	DeviceName string    `json:"device_name"`
}

type PlaybackStateEvent struct {
	Type     EventType    `json:"type"`
	Status   string       `json:"status"`
	Item     Track        `json:"item"`
	Position PositionInfo `json:"position"`
	Volume   int          `json:"volume"`
	Options  OptionsInfo  `json:"options"`
	IsActive bool         `json:"is_active"`
}

type TrackChangedEvent struct {
	Type EventType `json:"type"`
	Item Track     `json:"item"`
}

type PlaybackChangedEvent struct {
	Type   EventType `json:"type"`
	Status string    `json:"status"`
}

type VolumeChangedEvent struct {
	Type   EventType `json:"type"`
	Volume int       `json:"volume"`
}

type OptionsChangedEvent struct {
	Type    EventType   `json:"type"`
	Options OptionsInfo `json:"options"`
}

type PositionSyncEvent struct {
	Type     EventType    `json:"type"`
	Position PositionInfo `json:"position"`
}

type ErrorEvent struct {
	Type    EventType `json:"type"`
	Message string    `json:"message"`
}

// ParseEvent inspects raw JSON and parses it into the corresponding event struct.
func ParseEvent(data []byte) (interface{}, EventType, error) {
	var base BaseEvent
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal base event: %w", err)
	}

	switch base.Type {
	case EventAuthState:
		var ev AuthStateEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventPlaybackState:
		var ev PlaybackStateEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventTrackChanged:
		var ev TrackChangedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventPlaybackChanged:
		var ev PlaybackChangedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventVolumeChanged:
		var ev VolumeChangedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventOptionsChanged:
		var ev OptionsChangedEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventPositionSync:
		var ev PositionSyncEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	case EventError:
		var ev ErrorEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, base.Type, err
		}
		return ev, base.Type, nil

	default:
		// Unknown or unimplemented events are ignored.
		return nil, base.Type, nil
	}
}
