package game

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
)

// MemDBGameRepository implements GameRepository using go-memdb for storage
// and a local cache for returning the same *Game pointers with runtime state.
type MemDBGameRepository struct {
	ds    *datastore.DataStore
	rm    *datastore.RuntimeManager
	mu    sync.RWMutex
	cache map[string]*Game
}

func NewMemDBGameRepository(ds *datastore.DataStore, rm *datastore.RuntimeManager) *MemDBGameRepository {
	return &MemDBGameRepository{
		ds:    ds,
		rm:    rm,
		cache: make(map[string]*Game),
	}
}

// DataStore returns the underlying DataStore for use by actions that create games.
func (r *MemDBGameRepository) DataStore() *datastore.DataStore {
	return r.ds
}

func (r *MemDBGameRepository) Get(ctx context.Context, gameID string) (*Game, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	g, exists := r.cache[gameID]
	if !exists {
		return nil, fmt.Errorf("game %s not found", gameID)
	}
	return g, nil
}

// Create registers an already-constructed Game (whose state is already in the DataStore).
func (r *MemDBGameRepository) Create(ctx context.Context, g *Game) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if g == nil {
		return fmt.Errorf("game cannot be nil")
	}

	r.rm.Register(g.ID(), g.EventBus())

	r.mu.Lock()
	r.cache[g.ID()] = g
	r.mu.Unlock()

	return nil
}

func (r *MemDBGameRepository) Delete(ctx context.Context, gameID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	txn := r.ds.BeginTxn()
	defer txn.Abort()

	if err := txn.DeleteGame(gameID); err != nil {
		return err
	}
	txn.Commit()

	r.rm.Delete(gameID)

	r.mu.Lock()
	delete(r.cache, gameID)
	r.mu.Unlock()

	return nil
}

func (r *MemDBGameRepository) List(ctx context.Context, status *shared.GameStatus) ([]*Game, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	games := make([]*Game, 0, len(r.cache))
	for _, g := range r.cache {
		if status != nil && g.Status() != *status {
			continue
		}
		games = append(games, g)
	}
	return games, nil
}

func (r *MemDBGameRepository) Exists(ctx context.Context, gameID string) bool {
	return r.ds.GameExists(gameID)
}
