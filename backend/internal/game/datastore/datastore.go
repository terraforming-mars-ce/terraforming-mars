package datastore

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-memdb"

	"terraforming-mars-backend/internal/game/shared"
)

// SnapshotEnricher computes per-player VP breakdowns for a snapshot, using the live
// game state at the moment the snapshot is recorded. The map is stored on the history
// entry so the history mapper doesn't need to reimplement scoring logic.
type SnapshotEnricher func(state *GameState) map[string]shared.VPBreakdown

// DataStore is the single source of truth for all game data.
type DataStore struct {
	db              *memdb.MemDB
	historySeqMu    sync.Mutex
	historySequence map[string]int64 // per-game sequence counter
	enricher        SnapshotEnricher
}

func NewDataStore() (*DataStore, error) {
	db, err := memdb.NewMemDB(createSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to create memdb: %w", err)
	}
	return &DataStore{
		db:              db,
		historySequence: make(map[string]int64),
	}, nil
}

// SetSnapshotEnricher registers a callback invoked for each history append.
// Pass nil to disable enrichment.
func (ds *DataStore) SetSnapshotEnricher(fn SnapshotEnricher) {
	ds.enricher = fn
}

func (ds *DataStore) GetGame(gameID string) (*GameState, error) {
	txn := ds.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("games", "id", gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game %s: %w", gameID, err)
	}
	if raw == nil {
		return nil, fmt.Errorf("game %s not found", gameID)
	}
	return raw.(*GameState), nil
}

func (ds *DataStore) ListGames(status *shared.GameStatus) ([]*GameState, error) {
	txn := ds.db.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	var err error

	if status != nil {
		it, err = txn.Get("games", "status", *status)
	} else {
		it, err = txn.Get("games", "id")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}

	var games []*GameState
	for obj := it.Next(); obj != nil; obj = it.Next() {
		games = append(games, obj.(*GameState))
	}
	return games, nil
}

// UpdateGame fetches a game inside a write transaction, passes it to fn for mutation,
// then re-inserts and commits. If the game is not found, returns an error without calling fn.
func (ds *DataStore) UpdateGame(gameID string, fn func(state *GameState)) error {
	txn := ds.db.Txn(true)
	defer txn.Abort()

	raw, err := txn.First("games", "id", gameID)
	if err != nil {
		return fmt.Errorf("failed to get game %s: %w", gameID, err)
	}
	if raw == nil {
		return fmt.Errorf("game %s not found", gameID)
	}
	state := raw.(*GameState)
	fn(state)
	if err := txn.Insert("games", state); err != nil {
		return fmt.Errorf("failed to insert game %s: %w", state.ID, err)
	}
	txn.Commit()

	ds.appendHistory(state)

	return nil
}

func (ds *DataStore) appendHistory(state *GameState) {
	copied := deepCopyGameState(state)
	if copied == nil {
		return
	}

	ds.historySeqMu.Lock()
	ds.historySequence[state.ID]++
	seq := ds.historySequence[state.ID]
	ds.historySeqMu.Unlock()

	var vpBreakdowns map[string]shared.VPBreakdown
	if ds.enricher != nil {
		vpBreakdowns = ds.enricher(state)
	}

	entry := &GameStateHistoryEntry{
		GameID:       state.ID,
		Sequence:     seq,
		Timestamp:    time.Now(),
		State:        copied,
		VPBreakdowns: vpBreakdowns,
	}

	htxn := ds.db.Txn(true)
	_ = htxn.Insert("game_history", entry)
	htxn.Commit()
}

func deepCopyGameState(src *GameState) *GameState {
	data, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	dst := &GameState{}
	if err := json.Unmarshal(data, dst); err != nil {
		return nil
	}
	return dst
}

// RecordInitialHistory records the initial game state as the first history entry.
func (ds *DataStore) RecordInitialHistory(state *GameState) {
	ds.appendHistory(state)
}

// GetGameHistory returns all history entries for a game in chronological order.
func (ds *DataStore) GetGameHistory(gameID string) ([]*GameStateHistoryEntry, error) {
	txn := ds.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("game_history", "game_id", gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history for %s: %w", gameID, err)
	}

	var entries []*GameStateHistoryEntry
	for obj := it.Next(); obj != nil; obj = it.Next() {
		entries = append(entries, obj.(*GameStateHistoryEntry))
	}
	return entries, nil
}

// ReadGame fetches a game inside a read-only transaction and passes it to fn.
func (ds *DataStore) ReadGame(gameID string, fn func(state *GameState)) error {
	txn := ds.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("games", "id", gameID)
	if err != nil {
		return fmt.Errorf("failed to get game %s: %w", gameID, err)
	}
	if raw == nil {
		return fmt.Errorf("game %s not found", gameID)
	}
	fn(raw.(*GameState))
	return nil
}

func (ds *DataStore) UpdatePlayer(gameID, playerID string, fn func(state *PlayerState)) error {
	return ds.UpdateGame(gameID, func(s *GameState) {
		ps, ok := s.Players[playerID]
		if !ok {
			return
		}
		fn(ps)
	})
}

func (ds *DataStore) ReadPlayer(gameID, playerID string, fn func(state *PlayerState)) error {
	return ds.ReadGame(gameID, func(s *GameState) {
		ps, ok := s.Players[playerID]
		if !ok {
			return
		}
		fn(ps)
	})
}

func (ds *DataStore) GameExists(gameID string) bool {
	txn := ds.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("games", "id", gameID)
	return err == nil && raw != nil
}

// Txn wraps a go-memdb write transaction.
// Usage: txn := ds.BeginTxn(); defer txn.Abort(); ...; txn.Commit()
type Txn struct {
	txn *memdb.Txn
}

func (ds *DataStore) BeginTxn() *Txn {
	return &Txn{txn: ds.db.Txn(true)}
}

func (t *Txn) GetGame(gameID string) (*GameState, error) {
	raw, err := t.txn.First("games", "id", gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game %s: %w", gameID, err)
	}
	if raw == nil {
		return nil, fmt.Errorf("game %s not found", gameID)
	}
	return raw.(*GameState), nil
}

func (t *Txn) InsertGame(state *GameState) error {
	if state == nil {
		return fmt.Errorf("game state cannot be nil")
	}
	if err := t.txn.Insert("games", state); err != nil {
		return fmt.Errorf("failed to insert game %s: %w", state.ID, err)
	}
	return nil
}

func (t *Txn) DeleteGame(gameID string) error {
	raw, err := t.txn.First("games", "id", gameID)
	if err != nil {
		return fmt.Errorf("failed to find game %s for deletion: %w", gameID, err)
	}
	if raw == nil {
		return fmt.Errorf("game %s not found", gameID)
	}
	if err := t.txn.Delete("games", raw); err != nil {
		return fmt.Errorf("failed to delete game %s: %w", gameID, err)
	}
	return nil
}

func (t *Txn) GetSnapshot(gameID string) (*GameSnapshot, error) {
	raw, err := t.txn.First("snapshots", "id", gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot for game %s: %w", gameID, err)
	}
	if raw == nil {
		return nil, nil // No snapshot yet is not an error
	}
	return raw.(*GameSnapshot), nil
}

func (t *Txn) InsertSnapshot(snapshot *GameSnapshot) error {
	if err := t.txn.Insert("snapshots", snapshot); err != nil {
		return fmt.Errorf("failed to insert snapshot: %w", err)
	}
	return nil
}

func (t *Txn) GetDiffLog(gameID string) (*DiffLog, error) {
	raw, err := t.txn.First("difflogs", "id", gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff log for game %s: %w", gameID, err)
	}
	if raw == nil {
		return nil, nil // No diff log yet is not an error
	}
	return raw.(*DiffLog), nil
}

func (t *Txn) InsertDiffLog(diffLog *DiffLog) error {
	if err := t.txn.Insert("difflogs", diffLog); err != nil {
		return fmt.Errorf("failed to insert diff log: %w", err)
	}
	return nil
}

func (t *Txn) Commit() {
	t.txn.Commit()
}

// Abort discards the transaction. Safe to call after Commit (no-op).
func (t *Txn) Abort() {
	t.txn.Abort()
}
