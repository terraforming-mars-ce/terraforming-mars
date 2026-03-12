package deck

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/logger"
)

type Deck struct {
	ds     *datastore.DataStore
	gameID string
}

func NewDeck(ds *datastore.DataStore, gameID string, projectCardIDs, corpIDs, preludeIDs []string) *Deck {
	projectCopy := make([]string, len(projectCardIDs))
	copy(projectCopy, projectCardIDs)
	rand.Shuffle(len(projectCopy), func(i, j int) { projectCopy[i], projectCopy[j] = projectCopy[j], projectCopy[i] })

	corpCopy := make([]string, len(corpIDs))
	copy(corpCopy, corpIDs)
	rand.Shuffle(len(corpCopy), func(i, j int) { corpCopy[i], corpCopy[j] = corpCopy[j], corpCopy[i] })

	preludeCopy := make([]string, len(preludeIDs))
	copy(preludeCopy, preludeIDs)
	rand.Shuffle(len(preludeCopy), func(i, j int) { preludeCopy[i], preludeCopy[j] = preludeCopy[j], preludeCopy[i] })

	if err := ds.UpdateGame(gameID, func(s *datastore.GameState) {
		s.ProjectCards = projectCopy
		s.Corporations = corpCopy
		s.PreludeCards = preludeCopy
		s.DiscardPile = make([]string, 0)
		s.RemovedCards = make([]string, 0)
		s.DrawnCardCount = 0
		s.ShuffleCount = 0
	}); err != nil {
		logger.Get().Error("Failed to initialize deck state", zap.String("game_id", gameID), zap.Error(err))
	}

	return &Deck{ds: ds, gameID: gameID}
}

func NewDeckView(ds *datastore.DataStore, gameID string) *Deck {
	return &Deck{ds: ds, gameID: gameID}
}

func (d *Deck) update(fn func(s *datastore.GameState)) {
	if err := d.ds.UpdateGame(d.gameID, fn); err != nil {
		logger.Get().Warn("Failed to update game state", zap.String("game_id", d.gameID), zap.Error(err))
	}
}

func (d *Deck) read(fn func(s *datastore.GameState)) {
	if err := d.ds.ReadGame(d.gameID, fn); err != nil {
		logger.Get().Warn("Failed to read game state", zap.String("game_id", d.gameID), zap.Error(err))
	}
}

func (d *Deck) GameID() string {
	return d.gameID
}

func (d *Deck) ProjectCards() []string {
	var result []string
	d.read(func(s *datastore.GameState) {
		result = make([]string, len(s.ProjectCards))
		copy(result, s.ProjectCards)
	})
	return result
}

func (d *Deck) Corporations() []string {
	var result []string
	d.read(func(s *datastore.GameState) {
		result = make([]string, len(s.Corporations))
		copy(result, s.Corporations)
	})
	return result
}

func (d *Deck) DiscardPile() []string {
	var result []string
	d.read(func(s *datastore.GameState) {
		result = make([]string, len(s.DiscardPile))
		copy(result, s.DiscardPile)
	})
	return result
}

func (d *Deck) RemovedCards() []string {
	var result []string
	d.read(func(s *datastore.GameState) {
		result = make([]string, len(s.RemovedCards))
		copy(result, s.RemovedCards)
	})
	return result
}

func (d *Deck) PreludeCards() []string {
	var result []string
	d.read(func(s *datastore.GameState) {
		result = make([]string, len(s.PreludeCards))
		copy(result, s.PreludeCards)
	})
	return result
}

func (d *Deck) DrawnCardCount() int {
	var v int
	d.read(func(s *datastore.GameState) { v = s.DrawnCardCount })
	return v
}

func (d *Deck) ShuffleCount() int {
	var v int
	d.read(func(s *datastore.GameState) { v = s.ShuffleCount })
	return v
}

func (d *Deck) GetAvailableCardCount() int {
	var v int
	d.read(func(s *datastore.GameState) { v = len(s.ProjectCards) })
	return v
}

func (d *Deck) DrawProjectCards(ctx context.Context, count int) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var drawn []string
	var drawErr error
	d.update(func(s *datastore.GameState) {
		available := len(s.ProjectCards)
		if count > available {
			shuffleDeck(s)
			available = len(s.ProjectCards)
			if count > available {
				drawErr = fmt.Errorf("not enough cards available: requested %d, have %d", count, available)
				return
			}
		}

		drawn = make([]string, count)
		copy(drawn, s.ProjectCards[:count])
		s.ProjectCards = s.ProjectCards[count:]
		s.DrawnCardCount += count
	})
	return drawn, drawErr
}

func (d *Deck) DrawProjectCardsUntilMatching(ctx context.Context, count int, matcher func(cardID string) bool) (matched []string, discarded []string, err error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	d.update(func(s *datastore.GameState) {
		for len(matched) < count {
			if len(s.ProjectCards) == 0 {
				shuffleDeck(s)
				if len(s.ProjectCards) == 0 {
					break
				}
			}

			cardID := s.ProjectCards[0]
			s.ProjectCards = s.ProjectCards[1:]
			s.DrawnCardCount++

			if matcher(cardID) {
				matched = append(matched, cardID)
			} else {
				discarded = append(discarded, cardID)
			}
		}
	})

	return matched, discarded, nil
}

func (d *Deck) DrawCorporations(ctx context.Context, count int) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var drawn []string
	var drawErr error
	d.update(func(s *datastore.GameState) {
		available := len(s.Corporations)
		if count > available {
			drawErr = fmt.Errorf("not enough corporations available: requested %d, have %d", count, available)
			return
		}

		drawn = make([]string, count)
		copy(drawn, s.Corporations[:count])
		s.Corporations = s.Corporations[count:]
	})
	return drawn, drawErr
}

func (d *Deck) DrawPreludeCards(ctx context.Context, count int) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var drawn []string
	var drawErr error
	d.update(func(s *datastore.GameState) {
		available := len(s.PreludeCards)
		if count > available {
			drawErr = fmt.Errorf("not enough prelude cards available: requested %d, have %d", count, available)
			return
		}

		drawn = make([]string, count)
		copy(drawn, s.PreludeCards[:count])
		s.PreludeCards = s.PreludeCards[count:]
	})
	return drawn, drawErr
}

func (d *Deck) Discard(ctx context.Context, cardIDs []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.update(func(s *datastore.GameState) {
		s.DiscardPile = append(s.DiscardPile, cardIDs...)
	})
	return nil
}

func (d *Deck) Remove(ctx context.Context, cardIDs []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.update(func(s *datastore.GameState) {
		s.RemovedCards = append(s.RemovedCards, cardIDs...)
	})
	return nil
}

func shuffleDeck(s *datastore.GameState) {
	s.ProjectCards = append(s.ProjectCards, s.DiscardPile...)
	rand.Shuffle(len(s.ProjectCards), func(i, j int) { s.ProjectCards[i], s.ProjectCards[j] = s.ProjectCards[j], s.ProjectCards[i] })
	s.DiscardPile = make([]string, 0)
	s.ShuffleCount++
}

func (d *Deck) Shuffle(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.update(func(s *datastore.GameState) {
		shuffleDeck(s)
	})
	return nil
}
