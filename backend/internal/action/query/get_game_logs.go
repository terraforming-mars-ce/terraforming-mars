package query

import (
	"context"

	"terraforming-mars-backend/internal/game"

	"go.uber.org/zap"
)

// GetGameLogsAction handles querying game state diff logs
type GetGameLogsAction struct {
	stateRepo game.GameStateRepository
	logger    *zap.Logger
}

// NewGetGameLogsAction creates a new get game logs query action
func NewGetGameLogsAction(
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *GetGameLogsAction {
	return &GetGameLogsAction{
		stateRepo: stateRepo,
		logger:    logger,
	}
}

// Execute retrieves all state diffs for a game
func (a *GetGameLogsAction) Execute(ctx context.Context, gameID string, since int64) ([]game.StateDiff, error) {
	log := a.logger.With(zap.String("game_id", gameID), zap.Int64("since", since))
	log.Debug("Querying game logs")

	diffs, err := a.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game logs", zap.Error(err))
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

	log.Debug("Game logs query completed", zap.Int("count", len(diffs)))
	return diffs, nil
}
