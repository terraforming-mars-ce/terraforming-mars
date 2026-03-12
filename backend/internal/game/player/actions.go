package player

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// Actions manages available manual actions.
type Actions struct {
	ds       *datastore.DataStore
	gameID   string
	playerID string
}

// NewActions creates a new Actions view backed by the DataStore.
func NewActions(ds *datastore.DataStore, gameID, playerID string) *Actions {
	return &Actions{ds: ds, gameID: gameID, playerID: playerID}
}

func (a *Actions) update(fn func(s *datastore.PlayerState)) {
	if err := a.ds.UpdatePlayer(a.gameID, a.playerID, fn); err != nil {
		logger.Get().Warn("Failed to update player state", zap.String("game_id", a.gameID), zap.String("player_id", a.playerID), zap.Error(err))
	}
}

func (a *Actions) read(fn func(s *datastore.PlayerState)) {
	if err := a.ds.ReadPlayer(a.gameID, a.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", a.gameID), zap.String("player_id", a.playerID), zap.Error(err))
	}
}

func (a *Actions) List() []shared.CardAction {
	var actionsCopy []shared.CardAction
	a.read(func(s *datastore.PlayerState) {
		actionsCopy = make([]shared.CardAction, len(s.Actions))
		copy(actionsCopy, s.Actions)
	})
	return actionsCopy
}

func (a *Actions) SetActions(actions []shared.CardAction) {
	a.update(func(s *datastore.PlayerState) {
		if actions == nil {
			s.Actions = []shared.CardAction{}
		} else {
			s.Actions = make([]shared.CardAction, len(actions))
			copy(s.Actions, actions)
		}
	})
}

func (a *Actions) AddAction(action shared.CardAction) {
	a.update(func(s *datastore.PlayerState) {
		s.Actions = append(s.Actions, action)
	})
}

// ResetGenerationCounts resets the generation counts for all actions to 0
func (a *Actions) ResetGenerationCounts() {
	a.update(func(s *datastore.PlayerState) {
		for i := range s.Actions {
			s.Actions[i].TimesUsedThisGeneration = 0
		}
	})
}

// ResetTurnCounts resets the turn counts for all actions to 0
func (a *Actions) ResetTurnCounts() {
	a.update(func(s *datastore.PlayerState) {
		for i := range s.Actions {
			s.Actions[i].TimesUsedThisTurn = 0
		}
	})
}

// RemoveActionsByCardID removes all actions from a specific card
func (a *Actions) RemoveActionsByCardID(cardID string) {
	a.update(func(s *datastore.PlayerState) {
		filtered := make([]shared.CardAction, 0, len(s.Actions))
		for _, action := range s.Actions {
			if action.CardID != cardID {
				filtered = append(filtered, action)
			}
		}
		s.Actions = filtered
	})
}
