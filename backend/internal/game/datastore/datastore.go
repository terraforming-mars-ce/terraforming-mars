package datastore

import (
	"fmt"

	"github.com/hashicorp/go-memdb"

	"terraforming-mars-backend/internal/game/shared"
)

// DataStore is the single source of truth for all game data.
type DataStore struct {
	db *memdb.MemDB
}

func NewDataStore() (*DataStore, error) {
	db, err := memdb.NewMemDB(createSchema())
	if err != nil {
		return nil, fmt.Errorf("failed to create memdb: %w", err)
	}
	return &DataStore{db: db}, nil
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
	return nil
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
