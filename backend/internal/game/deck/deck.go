package deck

import (
	"context"
	"fmt"
	"sync"
)

// Deck represents the card deck state for a game with encapsulated state
type Deck struct {
	mu             sync.RWMutex
	gameID         string
	projectCards   []string // Available project card IDs (draw pile)
	corporations   []string // Available corporation card IDs
	discardPile    []string // Discarded card IDs
	removedCards   []string // Cards removed from game permanently
	preludeCards   []string // Available prelude card IDs
	drawnCardCount int      // Total cards drawn (for statistics)
	shuffleCount   int      // Number of times deck was shuffled
}

// NewDeck creates a new game deck with all cards available
func NewDeck(gameID string, projectCardIDs, corpIDs, preludeIDs []string) *Deck {
	projectCopy := make([]string, len(projectCardIDs))
	copy(projectCopy, projectCardIDs)

	corpCopy := make([]string, len(corpIDs))
	copy(corpCopy, corpIDs)

	preludeCopy := make([]string, len(preludeIDs))
	copy(preludeCopy, preludeIDs)

	return &Deck{
		gameID:         gameID,
		projectCards:   projectCopy,
		corporations:   corpCopy,
		preludeCards:   preludeCopy,
		discardPile:    make([]string, 0),
		removedCards:   make([]string, 0),
		drawnCardCount: 0,
		shuffleCount:   0,
	}
}

// GameID returns the game ID this deck belongs to
func (d *Deck) GameID() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.gameID
}

// ProjectCards returns a copy of available project cards
func (d *Deck) ProjectCards() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cardsCopy := make([]string, len(d.projectCards))
	copy(cardsCopy, d.projectCards)
	return cardsCopy
}

// Corporations returns a copy of available corporation cards
func (d *Deck) Corporations() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	corpsCopy := make([]string, len(d.corporations))
	copy(corpsCopy, d.corporations)
	return corpsCopy
}

// DiscardPile returns a copy of the discard pile
func (d *Deck) DiscardPile() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	discardCopy := make([]string, len(d.discardPile))
	copy(discardCopy, d.discardPile)
	return discardCopy
}

// RemovedCards returns a copy of permanently removed cards
func (d *Deck) RemovedCards() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	removedCopy := make([]string, len(d.removedCards))
	copy(removedCopy, d.removedCards)
	return removedCopy
}

// PreludeCards returns a copy of available prelude cards
func (d *Deck) PreludeCards() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	preludeCopy := make([]string, len(d.preludeCards))
	copy(preludeCopy, d.preludeCards)
	return preludeCopy
}

// DrawnCardCount returns the total number of cards drawn
func (d *Deck) DrawnCardCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.drawnCardCount
}

// ShuffleCount returns the number of times the deck was shuffled
func (d *Deck) ShuffleCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.shuffleCount
}

// GetAvailableCardCount returns the number of available project cards
func (d *Deck) GetAvailableCardCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.projectCards)
}

// DrawProjectCards draws N project cards from the deck
// Returns the drawn card IDs or error if not enough cards available
func (d *Deck) DrawProjectCards(ctx context.Context, count int) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	available := len(d.projectCards)
	if count > available {
		d.shuffleLocked()
		available = len(d.projectCards)
		if count > available {
			return nil, fmt.Errorf("not enough cards available: requested %d, have %d", count, available)
		}
	}

	drawnCards := make([]string, count)
	copy(drawnCards, d.projectCards[:count])
	d.projectCards = d.projectCards[count:]
	d.drawnCardCount += count

	return drawnCards, nil
}

// DrawCorporations draws N corporation cards
// Returns the drawn corporation IDs or error if not enough available
func (d *Deck) DrawCorporations(ctx context.Context, count int) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	available := len(d.corporations)
	if count > available {
		return nil, fmt.Errorf("not enough corporations available: requested %d, have %d", count, available)
	}

	drawnCorps := make([]string, count)
	copy(drawnCorps, d.corporations[:count])
	d.corporations = d.corporations[count:]

	return drawnCorps, nil
}

// Discard adds cards to the discard pile
func (d *Deck) Discard(ctx context.Context, cardIDs []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.discardPile = append(d.discardPile, cardIDs...)
	return nil
}

// Remove permanently removes cards from the game
func (d *Deck) Remove(ctx context.Context, cardIDs []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.removedCards = append(d.removedCards, cardIDs...)
	return nil
}

// shuffleLocked reshuffles the discard pile back into the project cards.
// Must be called while d.mu is already held.
func (d *Deck) shuffleLocked() {
	d.projectCards = append(d.projectCards, d.discardPile...)
	d.discardPile = make([]string, 0)
	d.shuffleCount++
}

// Shuffle reshuffles the discard pile back into the project cards
func (d *Deck) Shuffle(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.shuffleLocked()
	return nil
}
