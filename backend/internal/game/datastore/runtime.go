package datastore

import (
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/events"
)

// Runtime holds per-game non-serializable state outside memdb.
type Runtime struct {
	EventBus *events.EventBusImpl
}

// RuntimeManager manages per-game runtime state outside of memdb.
type RuntimeManager struct {
	mu       sync.RWMutex
	runtimes map[string]*Runtime
}

func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		runtimes: make(map[string]*Runtime),
	}
}

func (rm *RuntimeManager) Get(gameID string) *Runtime {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.runtimes[gameID]
}

// GetOrCreate returns the runtime for a game, creating one if it doesn't exist.
func (rm *RuntimeManager) GetOrCreate(gameID string) *Runtime {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if r, ok := rm.runtimes[gameID]; ok {
		return r
	}
	r := &Runtime{
		EventBus: events.NewEventBus(),
	}
	rm.runtimes[gameID] = r
	return r
}

// Create creates a new runtime for a game. Returns error if one already exists.
func (rm *RuntimeManager) Create(gameID string) (*Runtime, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, ok := rm.runtimes[gameID]; ok {
		return nil, fmt.Errorf("runtime already exists for game %s", gameID)
	}
	r := &Runtime{
		EventBus: events.NewEventBus(),
	}
	rm.runtimes[gameID] = r
	return r, nil
}

// Register stores an externally-created EventBus for a game, replacing any existing one.
func (rm *RuntimeManager) Register(gameID string, eventBus *events.EventBusImpl) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.runtimes[gameID] = &Runtime{EventBus: eventBus}
}

func (rm *RuntimeManager) Delete(gameID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.runtimes, gameID)
}
