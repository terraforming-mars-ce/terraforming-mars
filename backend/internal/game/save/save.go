// Package save defines the portable game contract independently of storage and transport.
package save

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"openmars/internal/game"
	"openmars/internal/game/datastore"
)

const (
	FormatVersion  = 1
	RulesVersion   = 1
	MaxBytes       = 128 << 20
	MaxEntries     = 100000
	MaxStringBytes = 65536
)

// Document contains one playable position and its immutable history.
type Document struct {
	FormatVersion      int                                `json:"formatVersion" save:"state"`
	RulesVersion       int                                `json:"rulesVersion" save:"state"`
	ContentFingerprint string                             `json:"contentFingerprint" save:"state"`
	ApplicationVersion string                             `json:"applicationVersion" save:"state"`
	SavedAt            time.Time                          `json:"savedAt" save:"state"`
	State              *datastore.GameState               `json:"state" save:"state"`
	History            []*datastore.GameStateHistoryEntry `json:"history" save:"state"`
	Log                []game.StateDiff                   `json:"log" save:"state"`
	LogBaseline        *game.GameSnapshot                 `json:"logBaseline" save:"state"`
}

// Fingerprint identifies the exact catalog used by a rules version.
func Fingerprint(catalog any) (string, error) {
	data, err := json.Marshal(catalog)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// Capture must run on the game's executor, between commands. History and log
// records are immutable; only the current position needs a detached copy here.
func Capture(g *game.Game, ds *datastore.DataStore, logs *game.InMemoryGameStateRepository, fingerprint, version string) (*Document, error) {
	stateData, err := json.Marshal(g.State())
	if err != nil {
		return nil, err
	}
	var state datastore.GameState
	if err := json.Unmarshal(stateData, &state); err != nil {
		return nil, err
	}
	history, err := ds.GetGameHistory(g.ID())
	if err != nil {
		return nil, err
	}
	entries, baseline := logs.CaptureLog(g.ID())
	return &Document{FormatVersion: FormatVersion, RulesVersion: RulesVersion, ContentFingerprint: fingerprint, ApplicationVersion: version, SavedAt: time.Now().UTC(), State: &state, History: history, Log: entries, LogBaseline: baseline}, nil
}

// Encode writes a bounded canonical representation. It never truncates history.
func Encode(doc *Document) ([]byte, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	if len(data) > MaxBytes {
		return nil, Failure("save_too_large", "This save exceeds the 128 MiB limit.", nil)
	}
	return data, nil
}

// Decode rejects ambiguous or lossy JSON, including fields ignored by custom decoders.
func Decode(data []byte) (*Document, error) {
	if len(data) > MaxBytes {
		return nil, Failure("save_too_large", "This save exceeds the 128 MiB limit.", nil)
	}
	if err := checkJSON(data); err != nil {
		return nil, Failure("invalid_json", "This text isn’t valid save JSON. Use the complete exported game.", err)
	}
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, Failure("invalid_save", "This isn’t an exported Open Mars game. Select a file downloaded with Save game.", nil)
	}
	var doc Document
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return nil, Failure("invalid_save", "This isn’t a valid exported Open Mars game. Select a file downloaded with Save game.", err)
	}
	if err := checkFields(data, &doc); err != nil {
		return nil, Failure("invalid_save", "This save is incomplete or contains invalid game data. Select another export.", err)
	}
	return &doc, nil
}
