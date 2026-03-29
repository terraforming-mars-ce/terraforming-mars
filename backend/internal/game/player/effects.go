package player

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// Effects manages passive effects from played cards.
type Effects struct {
	ds            *datastore.DataStore
	gameID        string
	playerID      string
	subscriptions map[string][]events.SubscriptionID
	eventBus      *events.EventBusImpl
}

// NewEffects creates a new Effects view backed by the DataStore.
func NewEffects(ds *datastore.DataStore, eventBus *events.EventBusImpl, gameID, playerID string) *Effects {
	return &Effects{
		ds:            ds,
		gameID:        gameID,
		playerID:      playerID,
		subscriptions: make(map[string][]events.SubscriptionID),
		eventBus:      eventBus,
	}
}

func (e *Effects) update(fn func(s *datastore.PlayerState)) {
	if err := e.ds.UpdatePlayer(e.gameID, e.playerID, fn); err != nil {
		logger.Get().Warn("Failed to update player state", zap.String("game_id", e.gameID), zap.String("player_id", e.playerID), zap.Error(err))
	}
}

func (e *Effects) read(fn func(s *datastore.PlayerState)) {
	if err := e.ds.ReadPlayer(e.gameID, e.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", e.gameID), zap.String("player_id", e.playerID), zap.Error(err))
	}
}

func (e *Effects) List() []shared.CardEffect {
	var effectsCopy []shared.CardEffect
	e.read(func(s *datastore.PlayerState) {
		effectsCopy = make([]shared.CardEffect, len(s.Effects))
		copy(effectsCopy, s.Effects)
	})
	return effectsCopy
}

func (e *Effects) SetEffects(effects []shared.CardEffect) {
	e.update(func(s *datastore.PlayerState) {
		if effects == nil {
			s.Effects = []shared.CardEffect{}
		} else {
			s.Effects = make([]shared.CardEffect, len(effects))
			copy(s.Effects, effects)
		}
	})
}

func (e *Effects) AddEffect(effect shared.CardEffect) {
	e.update(func(s *datastore.PlayerState) {
		s.Effects = append(s.Effects, effect)
	})
}

// RegisterSubscription tracks an event subscription for a card so it can be unsubscribed later
func (e *Effects) RegisterSubscription(cardID string, subID events.SubscriptionID) {
	e.subscriptions[cardID] = append(e.subscriptions[cardID], subID)
}

// RemoveEffectsByCardID removes all effects from a specific card and unsubscribes from events
func (e *Effects) RemoveEffectsByCardID(cardID string) {
	e.update(func(s *datastore.PlayerState) {
		filtered := make([]shared.CardEffect, 0, len(s.Effects))
		for _, effect := range s.Effects {
			if effect.CardID != cardID {
				filtered = append(filtered, effect)
			}
		}
		s.Effects = filtered
	})

	if subs, exists := e.subscriptions[cardID]; exists {
		for _, subID := range subs {
			e.eventBus.Unsubscribe(subID)
		}
		delete(e.subscriptions, cardID)
	}
}

// RemoveTemporaryEffects removes all effects that have outputs with the given temporary type.
// Returns the card IDs of removed effects.
func (e *Effects) RemoveTemporaryEffects(temporaryType string) []string {
	var removedCardIDs []string
	e.update(func(s *datastore.PlayerState) {
		filtered := make([]shared.CardEffect, 0, len(s.Effects))

		for _, effect := range s.Effects {
			hasTemporary := false
			for _, output := range effect.Behavior.Outputs {
				if shared.GetTemporary(output) == temporaryType {
					hasTemporary = true
					break
				}
			}

			if hasTemporary {
				removedCardIDs = append(removedCardIDs, effect.CardID)
			} else {
				filtered = append(filtered, effect)
			}
		}

		s.Effects = filtered
	})

	for _, cardID := range removedCardIDs {
		if subs, exists := e.subscriptions[cardID]; exists {
			for _, subID := range subs {
				e.eventBus.Unsubscribe(subID)
			}
			delete(e.subscriptions, cardID)
		}
	}

	return removedCardIDs
}
