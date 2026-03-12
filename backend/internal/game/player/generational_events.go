package player

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// GenerationalEvents tracks per-generation player events.
type GenerationalEvents struct {
	ds       *datastore.DataStore
	gameID   string
	playerID string
}

func newGenerationalEvents(ds *datastore.DataStore, gameID, playerID string) *GenerationalEvents {
	return &GenerationalEvents{ds: ds, gameID: gameID, playerID: playerID}
}

func (ge *GenerationalEvents) update(fn func(s *datastore.PlayerState)) {
	if err := ge.ds.UpdatePlayer(ge.gameID, ge.playerID, fn); err != nil {
		logger.Get().Warn("Failed to update player state", zap.String("game_id", ge.gameID), zap.String("player_id", ge.playerID), zap.Error(err))
	}
}

func (ge *GenerationalEvents) read(fn func(s *datastore.PlayerState)) {
	if err := ge.ds.ReadPlayer(ge.gameID, ge.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", ge.gameID), zap.String("player_id", ge.playerID), zap.Error(err))
	}
}

func (ge *GenerationalEvents) Increment(event shared.GenerationalEvent) {
	ge.update(func(s *datastore.PlayerState) {
		if s.GenerationalEvents == nil {
			s.GenerationalEvents = make(map[shared.GenerationalEvent]int)
		}
		s.GenerationalEvents[event]++
	})
}

func (ge *GenerationalEvents) GetCount(event shared.GenerationalEvent) int {
	var count int
	ge.read(func(s *datastore.PlayerState) {
		if s.GenerationalEvents == nil {
			return
		}
		count = s.GenerationalEvents[event]
	})
	return count
}

func (ge *GenerationalEvents) GetAll() []shared.PlayerGenerationalEventEntry {
	var entries []shared.PlayerGenerationalEventEntry
	ge.read(func(s *datastore.PlayerState) {
		entries = make([]shared.PlayerGenerationalEventEntry, 0, len(s.GenerationalEvents))
		for event, count := range s.GenerationalEvents {
			entries = append(entries, shared.PlayerGenerationalEventEntry{
				Event: event,
				Count: count,
			})
		}
	})
	return entries
}

func (ge *GenerationalEvents) Clear() {
	ge.update(func(s *datastore.PlayerState) {
		s.GenerationalEvents = make(map[shared.GenerationalEvent]int)
	})
}
