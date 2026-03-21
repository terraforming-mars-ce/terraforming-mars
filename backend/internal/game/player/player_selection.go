package player

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// Selection manages player-specific card selection state.
type Selection struct {
	ds       *datastore.DataStore
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

func newSelection(ds *datastore.DataStore, eventBus *events.EventBusImpl, gameID, playerID string) *Selection {
	return &Selection{
		ds:       ds,
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (s *Selection) update(fn func(st *datastore.PlayerState)) {
	if err := s.ds.UpdatePlayer(s.gameID, s.playerID, fn); err != nil {
		logger.Get().Warn("Failed to update player state", zap.String("game_id", s.gameID), zap.String("player_id", s.playerID), zap.Error(err))
	}
}

func (s *Selection) read(fn func(st *datastore.PlayerState)) {
	if err := s.ds.ReadPlayer(s.gameID, s.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", s.gameID), zap.String("player_id", s.playerID), zap.Error(err))
	}
}

func (s *Selection) GetSelectCorporationPhase() *shared.SelectCorporationPhase {
	var phase *shared.SelectCorporationPhase
	s.read(func(st *datastore.PlayerState) {
		phase = st.SelectCorporationPhase
	})
	return phase
}

func (s *Selection) SetSelectCorporationPhase(phase *shared.SelectCorporationPhase) {
	s.update(func(st *datastore.PlayerState) {
		st.SelectCorporationPhase = phase
	})
}

func (s *Selection) GetSelectStartingCardsPhase() *shared.SelectStartingCardsPhase {
	var phase *shared.SelectStartingCardsPhase
	s.read(func(st *datastore.PlayerState) {
		phase = st.SelectStartingCardsPhase
	})
	return phase
}

func (s *Selection) SetSelectStartingCardsPhase(phase *shared.SelectStartingCardsPhase) {
	s.update(func(st *datastore.PlayerState) {
		st.SelectStartingCardsPhase = phase
	})
}

func (s *Selection) GetSelectPreludeCardsPhase() *shared.SelectPreludeCardsPhase {
	var phase *shared.SelectPreludeCardsPhase
	s.read(func(st *datastore.PlayerState) {
		phase = st.SelectPreludeCardsPhase
	})
	return phase
}

func (s *Selection) SetSelectPreludeCardsPhase(phase *shared.SelectPreludeCardsPhase) {
	s.update(func(st *datastore.PlayerState) {
		st.SelectPreludeCardsPhase = phase
	})
}

func (s *Selection) GetPendingCardSelection() *shared.PendingCardSelection {
	var sel *shared.PendingCardSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingCardSelection
	})
	return sel
}

func (s *Selection) SetPendingCardSelection(selection *shared.PendingCardSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingCardSelection = selection
	})
}

func (s *Selection) GetPendingCardDrawSelection() *shared.PendingCardDrawSelection {
	var sel *shared.PendingCardDrawSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingCardDrawSelection
	})
	return sel
}

func (s *Selection) SetPendingCardDrawSelection(selection *shared.PendingCardDrawSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingCardDrawSelection = selection
	})
}

func (s *Selection) GetPendingCardDiscardSelection() *shared.PendingCardDiscardSelection {
	var sel *shared.PendingCardDiscardSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingCardDiscardSelection
	})
	return sel
}

func (s *Selection) SetPendingCardDiscardSelection(selection *shared.PendingCardDiscardSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingCardDiscardSelection = selection
	})
}

func (s *Selection) GetPendingBehaviorChoiceSelection() *shared.PendingBehaviorChoiceSelection {
	var sel *shared.PendingBehaviorChoiceSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingBehaviorChoiceSelection
	})
	return sel
}

func (s *Selection) SetPendingBehaviorChoiceSelection(selection *shared.PendingBehaviorChoiceSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingBehaviorChoiceSelection = selection
	})
}

func (s *Selection) GetPendingStealTargetSelection() *shared.PendingStealTargetSelection {
	var sel *shared.PendingStealTargetSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingStealTargetSelection
	})
	return sel
}

func (s *Selection) SetPendingStealTargetSelection(selection *shared.PendingStealTargetSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingStealTargetSelection = selection
	})
}

func (s *Selection) GetPendingColonyResourceSelection() *shared.PendingColonyResourceSelection {
	var sel *shared.PendingColonyResourceSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingColonyResourceSelection
	})
	return sel
}

func (s *Selection) SetPendingColonyResourceSelection(selection *shared.PendingColonyResourceSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingColonyResourceSelection = selection
	})
}

func (s *Selection) GetPendingAwardFundSelection() *shared.PendingAwardFundSelection {
	var sel *shared.PendingAwardFundSelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingAwardFundSelection
	})
	return sel
}

func (s *Selection) SetPendingAwardFundSelection(selection *shared.PendingAwardFundSelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingAwardFundSelection = selection
	})
}

func (s *Selection) GetPendingColonySelection() *shared.PendingColonySelection {
	var sel *shared.PendingColonySelection
	s.read(func(st *datastore.PlayerState) {
		sel = st.PendingColonySelection
	})
	return sel
}

func (s *Selection) SetPendingColonySelection(selection *shared.PendingColonySelection) {
	s.update(func(st *datastore.PlayerState) {
		st.PendingColonySelection = selection
	})
}
