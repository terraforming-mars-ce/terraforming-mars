package query

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// HistoryPolicy controls how history entries are grouped/reduced.
type HistoryPolicy string

const (
	// HistoryPolicyLatestInAction keeps only the last entry per action turn.
	HistoryPolicyLatestInAction HistoryPolicy = "latestInAction"
	// HistoryPolicyEveryTurn keeps the last entry before each turn-player change.
	HistoryPolicyEveryTurn HistoryPolicy = "everyTurn"
	// HistoryPolicyEveryAction keeps the last entry per global action counter value.
	HistoryPolicyEveryAction HistoryPolicy = "everyAction"
)

// HistoryFilter specifies which history entries to return.
type HistoryFilter struct {
	Phases []shared.GamePhase
	Policy HistoryPolicy
}

// GetGameHistoryAction handles querying game state history
type GetGameHistoryAction struct {
	ds     *datastore.DataStore
	logger *zap.Logger
}

// NewGetGameHistoryAction creates a new get game history query action
func NewGetGameHistoryAction(
	ds *datastore.DataStore,
	logger *zap.Logger,
) *GetGameHistoryAction {
	return &GetGameHistoryAction{
		ds:     ds,
		logger: logger,
	}
}

// Execute retrieves history entries for a game, optionally filtered.
func (a *GetGameHistoryAction) Execute(ctx context.Context, gameID string, filter *HistoryFilter) ([]*datastore.GameStateHistoryEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	log := a.logger.With(zap.String("game_id", gameID))
	log.Debug("Querying game history")

	entries, err := a.ds.GetGameHistory(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}

	if filter != nil {
		entries = applyFilter(entries, filter)
	}

	log.Debug("Game history query completed", zap.Int("count", len(entries)))
	return entries, nil
}

func applyFilter(entries []*datastore.GameStateHistoryEntry, filter *HistoryFilter) []*datastore.GameStateHistoryEntry {
	if len(filter.Phases) > 0 {
		allowed := make(map[shared.GamePhase]bool, len(filter.Phases))
		for _, p := range filter.Phases {
			allowed[p] = true
		}
		filtered := make([]*datastore.GameStateHistoryEntry, 0, len(entries))
		for _, e := range entries {
			if allowed[e.State.CurrentPhase] {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	switch filter.Policy {
	case HistoryPolicyLatestInAction:
		entries = applyLatestInAction(entries)
	case HistoryPolicyEveryTurn:
		entries = applyEveryTurn(entries)
	case HistoryPolicyEveryAction:
		entries = applyEveryAction(entries)
	}

	return entries
}

// applyEveryAction keeps the last entry per global action counter value,
// giving one snapshot per action consumed across all players.
func applyEveryAction(entries []*datastore.GameStateHistoryEntry) []*datastore.GameStateHistoryEntry {
	if len(entries) == 0 {
		return entries
	}
	result := make([]*datastore.GameStateHistoryEntry, 0, len(entries))
	for i, e := range entries {
		isLast := i == len(entries)-1
		if isLast {
			result = append(result, e)
			break
		}
		next := entries[i+1]
		if e.State.GlobalActionCounter != next.State.GlobalActionCounter {
			result = append(result, e)
		}
	}
	return result
}

// applyEveryTurn keeps the last entry before each turn-player change,
// giving one snapshot per player turn regardless of action count.
func applyEveryTurn(entries []*datastore.GameStateHistoryEntry) []*datastore.GameStateHistoryEntry {
	if len(entries) == 0 {
		return entries
	}
	result := make([]*datastore.GameStateHistoryEntry, 0, len(entries))
	for i, e := range entries {
		isLast := i == len(entries)-1
		if isLast {
			result = append(result, e)
			break
		}
		next := entries[i+1]
		sameTurn := e.State.CurrentTurnPlayerID == next.State.CurrentTurnPlayerID &&
			e.State.CurrentPhase == next.State.CurrentPhase
		if !sameTurn {
			result = append(result, e)
		}
	}
	return result
}

// applyLatestInAction keeps only the last entry in each run of consecutive entries
// with the same player turn and remaining actions count.
func applyLatestInAction(entries []*datastore.GameStateHistoryEntry) []*datastore.GameStateHistoryEntry {
	if len(entries) == 0 {
		return entries
	}

	result := make([]*datastore.GameStateHistoryEntry, 0, len(entries))
	for i, e := range entries {
		isLast := i == len(entries)-1
		if isLast {
			result = append(result, e)
			break
		}
		next := entries[i+1]
		sameAction := e.State.CurrentTurnPlayerID == next.State.CurrentTurnPlayerID &&
			e.State.CurrentTurnActions == next.State.CurrentTurnActions &&
			e.State.CurrentPhase == next.State.CurrentPhase
		if !sameAction {
			result = append(result, e)
		}
	}
	return result
}
