package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// ValidateGameExists validates that a game exists (any status)
// Returns the game if valid, or an error if not found
func ValidateGameExists(
	ctx context.Context,
	gameRepo game.GameRepository,
	gameID string,
	log *zap.Logger,
) (*game.Game, error) {
	game, err := gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}
	return game, nil
}

// ValidateActiveGame validates that a game exists and is in active status
// Returns the game if valid, or an error if not found or wrong status
func ValidateActiveGame(
	ctx context.Context,
	gameRepo game.GameRepository,
	gameID string,
	log *zap.Logger,
) (*game.Game, error) {
	return ValidateGameStatus(ctx, gameRepo, gameID, shared.GameStatusActive, log)
}

// ValidateLobbyGame validates that a game exists and is in lobby status
// Returns the game if valid, or an error if not found or wrong status
func ValidateLobbyGame(
	ctx context.Context,
	gameRepo game.GameRepository,
	gameID string,
	log *zap.Logger,
) (*game.Game, error) {
	return ValidateGameStatus(ctx, gameRepo, gameID, shared.GameStatusLobby, log)
}

// ValidateGameStatus validates that a game exists and has the expected status
// Returns the game if valid, or an error if not found or wrong status
func ValidateGameStatus(
	ctx context.Context,
	gameRepo game.GameRepository,
	gameID string,
	expectedStatus shared.GameStatus,
	log *zap.Logger,
) (*game.Game, error) {
	gameResult, err := gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	if gameResult.Status() != expectedStatus {
		log.Error("Game not in expected status",
			zap.String("expected", string(expectedStatus)),
			zap.String("actual", string(gameResult.Status())))
		return nil, fmt.Errorf("game not in %s status", expectedStatus)
	}

	return gameResult, nil
}

// ValidateGamePhase validates that a game is in the expected phase
// Returns error if game is not in the expected phase
func ValidateGamePhase(
	gameInstance *game.Game,
	expectedPhase shared.GamePhase,
	log *zap.Logger,
) error {
	if gameInstance.CurrentPhase() != expectedPhase {
		log.Error("Game not in expected phase",
			zap.String("expected", string(expectedPhase)),
			zap.String("actual", string(gameInstance.CurrentPhase())))
		return fmt.Errorf("game not in %s phase", expectedPhase)
	}
	return nil
}

// ValidateHostPermission validates that the specified player is the game host
// Returns error if player is not the host
func ValidateHostPermission(
	gameInstance *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	if gameInstance.HostPlayerID() != playerID {
		log.Error("Non-host attempted privileged action",
			zap.String("player_id", playerID),
			zap.String("host_id", gameInstance.HostPlayerID()))
		return fmt.Errorf("only the host can perform this action")
	}
	return nil
}

// ValidateCurrentTurn validates that it's the specified player's turn
// Returns error if it's not their turn or no current turn is set
func ValidateCurrentTurn(
	gameInstance *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	currentTurn := gameInstance.CurrentTurn()
	if currentTurn == nil {
		log.Error("No current turn set")
		return fmt.Errorf("no current turn set")
	}

	if currentTurn.PlayerID() != playerID {
		log.Error("Not player's turn",
			zap.String("player_id", playerID),
			zap.String("current_turn", currentTurn.PlayerID()))
		return fmt.Errorf("not your turn")
	}

	return nil
}

// ValidateActionsRemaining validates that the current player has actions remaining
// Returns error if actionsRemaining == 0; allows -1 (unlimited) and >0
func ValidateActionsRemaining(
	gameInstance *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	currentTurn := gameInstance.CurrentTurn()
	if currentTurn == nil {
		return nil
	}

	if currentTurn.PlayerID() != playerID {
		return nil
	}

	remaining := currentTurn.ActionsRemaining()
	if remaining == 0 {
		log.Warn("No actions remaining",
			zap.String("player_id", playerID),
			zap.Int("actions_remaining", remaining))
		return fmt.Errorf("no actions remaining")
	}

	return nil
}

// ValidateNoPendingSelections validates that the player has no pending selections.
// Actions should be blocked while a player is resolving a pending selection.
func ValidateNoPendingSelections(
	gameInstance *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	if gameInstance.HasAnyPendingSelection(playerID) {
		return fmt.Errorf("pending selection")
	}
	return nil
}
