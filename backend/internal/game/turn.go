package game

import (
	"log/slog"
	"openmars/internal/game/datastore"
)

type Turn struct {
	ds     *datastore.DataStore
	gameID string
}

func NewTurn(ds *datastore.DataStore, gameID string) *Turn {
	return &Turn{ds: ds, gameID: gameID}
}

func (t *Turn) update(fn func(s *datastore.GameState)) {
	if err := t.ds.UpdateGame(t.gameID, fn); err != nil {
		slog.Default().Warn("Failed to update game state", slog.String("game_id", t.gameID), slog.Any("error", err))
	}
}

func (t *Turn) read(fn func(s *datastore.GameState)) {
	if err := t.ds.ReadGame(t.gameID, fn); err != nil {
		slog.Default().Warn("Failed to read game state", slog.String("game_id", t.gameID), slog.Any("error", err))
	}
}

func (t *Turn) PlayerID() string {
	var v string
	t.read(func(s *datastore.GameState) { v = s.CurrentTurnPlayerID })
	return v
}

func (t *Turn) ActionsRemaining() int {
	var v int
	t.read(func(s *datastore.GameState) { v = s.CurrentTurnActions })
	return v
}

func (t *Turn) TotalActions() int {
	var v int
	t.read(func(s *datastore.GameState) { v = s.CurrentTurnTotalActions })
	return v
}

func (t *Turn) SetPlayerID(playerID string) {
	t.update(func(s *datastore.GameState) { s.CurrentTurnPlayerID = playerID })
}

func (t *Turn) SetActionsRemaining(actions int) {
	t.update(func(s *datastore.GameState) { s.CurrentTurnActions = actions })
}

func (t *Turn) SetTotalActions(totalActions int) {
	t.update(func(s *datastore.GameState) { s.CurrentTurnTotalActions = totalActions })
}

func (t *Turn) AddExtraActions(amount int) {
	t.update(func(s *datastore.GameState) {
		if s.CurrentTurnActions >= 0 {
			s.CurrentTurnActions += amount
			s.CurrentTurnTotalActions += amount
		}
	})
}

func (t *Turn) GlobalActionCounter() int {
	var v int
	t.read(func(s *datastore.GameState) { v = s.GlobalActionCounter })
	return v
}

func (t *Turn) IncrementGlobalActionCounter() {
	t.update(func(s *datastore.GameState) {
		s.GlobalActionCounter++
	})
}

func (t *Turn) ConsumeAction() bool {
	var consumed bool
	t.update(func(s *datastore.GameState) {
		if s.CurrentTurnActions > 0 {
			s.CurrentTurnActions--
			consumed = true
		}
	})
	return consumed
}
