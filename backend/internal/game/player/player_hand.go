package player

import (
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/logger"
)

// Hand manages player card hand as a pure DataStore adapter.
type Hand struct {
	ds       *datastore.DataStore
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

func newHand(ds *datastore.DataStore, eventBus *events.EventBusImpl, gameID, playerID string) *Hand {
	return &Hand{
		ds:       ds,
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (h *Hand) update(fn func(s *datastore.PlayerState)) {
	if err := h.ds.UpdatePlayer(h.gameID, h.playerID, fn); err != nil {
		logger.Get().Warn("Failed to update player state", zap.String("game_id", h.gameID), zap.String("player_id", h.playerID), zap.Error(err))
	}
}

func (h *Hand) read(fn func(s *datastore.PlayerState)) {
	if err := h.ds.ReadPlayer(h.gameID, h.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", h.gameID), zap.String("player_id", h.playerID), zap.Error(err))
	}
}

func (h *Hand) Cards() []string {
	var cardsCopy []string
	h.read(func(s *datastore.PlayerState) {
		cardsCopy = make([]string, len(s.HandCardIDs))
		copy(cardsCopy, s.HandCardIDs)
	})
	return cardsCopy
}

func (h *Hand) CardCount() int {
	var count int
	h.read(func(s *datastore.PlayerState) {
		count = len(s.HandCardIDs)
	})
	return count
}

func (h *Hand) HasCard(cardID string) bool {
	var found bool
	h.read(func(s *datastore.PlayerState) {
		for _, id := range s.HandCardIDs {
			if id == cardID {
				found = true
				return
			}
		}
	})
	return found
}

func (h *Hand) SetCards(cards []string) {
	var cardsCopy []string
	h.update(func(s *datastore.PlayerState) {
		if cards == nil {
			s.HandCardIDs = []string{}
		} else {
			s.HandCardIDs = make([]string, len(cards))
			copy(s.HandCardIDs, cards)
		}
		cardsCopy = make([]string, len(s.HandCardIDs))
		copy(cardsCopy, s.HandCardIDs)
	})

	if h.eventBus != nil {
		events.Publish(h.eventBus, events.CardHandUpdatedEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardIDs:   cardsCopy,
			Timestamp: time.Now(),
		})
	}
}

func (h *Hand) AddCard(cardID string) {
	var cardsCopy []string
	h.update(func(s *datastore.PlayerState) {
		s.HandCardIDs = append(s.HandCardIDs, cardID)
		cardsCopy = make([]string, len(s.HandCardIDs))
		copy(cardsCopy, s.HandCardIDs)
	})

	if h.eventBus != nil {
		events.Publish(h.eventBus, events.CardAddedToHandEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardID:    cardID,
			Timestamp: time.Now(),
		})

		events.Publish(h.eventBus, events.CardHandUpdatedEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardIDs:   cardsCopy,
			Timestamp: time.Now(),
		})
	}
}

func (h *Hand) RemoveCard(cardID string) bool {
	var removed bool
	var cardsCopy []string
	h.update(func(s *datastore.PlayerState) {
		for i, id := range s.HandCardIDs {
			if id == cardID {
				s.HandCardIDs = append(s.HandCardIDs[:i], s.HandCardIDs[i+1:]...)
				removed = true
				break
			}
		}
		cardsCopy = make([]string, len(s.HandCardIDs))
		copy(cardsCopy, s.HandCardIDs)
	})

	if removed && h.eventBus != nil {
		events.Publish(h.eventBus, events.CardHandUpdatedEvent{
			GameID:    h.gameID,
			PlayerID:  h.playerID,
			CardIDs:   cardsCopy,
			Timestamp: time.Now(),
		})
	}

	return removed
}
