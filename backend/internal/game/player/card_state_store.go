package player

import "sync"

// CardStateStore manages computed EntityState for cards in a player's hand.
// This is a runtime cache driven by events — the DataStore owns card IDs,
// and this store holds the calculated playability state for each card.
type CardStateStore struct {
	states map[string]EntityState
	mu     sync.RWMutex
}

// NewCardStateStore creates a new empty CardStateStore.
func NewCardStateStore() *CardStateStore {
	return &CardStateStore{
		states: make(map[string]EntityState),
	}
}

// GetState returns the cached EntityState for a card.
func (s *CardStateStore) GetState(cardID string) (EntityState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[cardID]
	return state, ok
}

// SetState stores the EntityState for a card.
func (s *CardStateStore) SetState(cardID string, state EntityState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[cardID] = state
}

// RemoveState removes a card's state from the store.
func (s *CardStateStore) RemoveState(cardID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, cardID)
}

// SyncWithHand removes entries not present in the given card ID list.
// Called when the hand changes to clean up stale entries.
func (s *CardStateStore) SyncWithHand(cardIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	active := make(map[string]struct{}, len(cardIDs))
	for _, id := range cardIDs {
		active[id] = struct{}{}
	}
	for id := range s.states {
		if _, ok := active[id]; !ok {
			delete(s.states, id)
		}
	}
}

// RecalculateAll recomputes state for every card in the store using the provided function.
func (s *CardStateStore) RecalculateAll(fn func(cardID string) EntityState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for cardID := range s.states {
		s.states[cardID] = fn(cardID)
	}
}

// AllCardIDs returns all card IDs tracked by the store.
func (s *CardStateStore) AllCardIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.states))
	for id := range s.states {
		ids = append(ids, id)
	}
	return ids
}
