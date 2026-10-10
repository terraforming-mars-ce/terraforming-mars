package query

import (
	"context"
	"log/slog"

	"openmars/internal/game"
)

// GetGameLogsAction handles querying game state diff logs
type GetGameLogsAction struct {
	stateRepo game.GameStateRepository
	logger    *slog.Logger
}

// NewGetGameLogsAction creates a new get game logs query action
func NewGetGameLogsAction(
	stateRepo game.GameStateRepository,
	logger *slog.Logger,
) *GetGameLogsAction {
	return &GetGameLogsAction{
		stateRepo: stateRepo,
		logger:    logger,
	}
}

// Execute retrieves all state diffs for a game
func (a *GetGameLogsAction) Execute(ctx context.Context, gameID string, since int64) ([]game.StateDiff, error) {
	log := a.logger.With(slog.String("game_id", gameID), slog.Int64("since", since))
	log.Debug("Querying game logs")

	diffs, err := a.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game logs", slog.Any("error", err))
		return nil, err
	}

	if since > 0 {
		filtered := make([]game.StateDiff, 0)
		for _, diff := range diffs {
			if diff.SequenceNumber > since {
				filtered = append(filtered, diff)
			}
		}
		diffs = filtered
	}

	log.Debug("Game logs query completed", slog.Int("count", len(diffs)))
	return diffs, nil
}
