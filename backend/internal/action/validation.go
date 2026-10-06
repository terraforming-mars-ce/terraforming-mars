package action

import (
	"context"
	"fmt"
	"log/slog"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// ValidateGameExists validates that a game exists (any status)
// Returns the game if valid, or an error if not found
func ValidateGameExists(
	ctx context.Context,
	gameRepo game.GameRepository,
	gameID string,
	log *slog.Logger,
) (*game.Game, error) {
	game, err := gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", slog.Any("error", err))
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
	log *slog.Logger,
) (*game.Game, error) {
	return ValidateGameStatus(ctx, gameRepo, gameID, shared.GameStatusActive, log)
}

// ValidateLobbyGame validates that a game exists and is in lobby status
// Returns the game if valid, or an error if not found or wrong status
func ValidateLobbyGame(
	ctx context.Context,
	gameRepo game.GameRepository,
	gameID string,
	log *slog.Logger,
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
	log *slog.Logger,
) (*game.Game, error) {
	gameResult, err := gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", slog.Any("error", err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	if gameResult.Status() != expectedStatus {
		log.Error("Game not in expected status",
			slog.String("expected", string(expectedStatus)),
			slog.String("actual", string(gameResult.Status())))
		return nil, fmt.Errorf("game not in %s status", expectedStatus)
	}

	return gameResult, nil
}

// ValidateGamePhase validates that a game is in the expected phase
// Returns error if game is not in the expected phase
func ValidateGamePhase(
	gameInstance *game.Game,
	expectedPhase shared.GamePhase,
	log *slog.Logger,
) error {
	if gameInstance.CurrentPhase() != expectedPhase {
		log.Error("Game not in expected phase",
			slog.String("expected", string(expectedPhase)),
			slog.String("actual", string(gameInstance.CurrentPhase())))
		return fmt.Errorf("game not in %s phase", expectedPhase)
	}
	return nil
}

// ValidateHostPermission validates that the specified player is the game host
// Returns error if player is not the host
func ValidateHostPermission(
	gameInstance *game.Game,
	playerID string,
	log *slog.Logger,
) error {
	if gameInstance.HostPlayerID() != playerID {
		log.Error("Non-host attempted privileged action",
			slog.String("player_id", playerID),
			slog.String("host_id", gameInstance.HostPlayerID()))
		return fmt.Errorf("only the host can perform this action")
	}
	return nil
}

// ValidateCurrentTurn validates that it's the specified player's turn
// Returns error if it's not their turn or no current turn is set
func ValidateCurrentTurn(
	gameInstance *game.Game,
	playerID string,
	log *slog.Logger,
) error {
	currentTurn := gameInstance.CurrentTurn()
	if currentTurn == nil {
		log.Error("No current turn set")
		return fmt.Errorf("no current turn set")
	}

	if currentTurn.PlayerID() != playerID {
		log.Error("Not player's turn",
			slog.String("player_id", playerID),
			slog.String("current_turn", currentTurn.PlayerID()))
		return fmt.Errorf("not your turn")
	}

	return nil
}

// ValidateCurrentTurnOrInitPlayer validates that the player may act now: during the
// init_apply phases that is the player whose corp or preludes are being applied,
// otherwise it is the current turn's player.
func ValidateCurrentTurnOrInitPlayer(
	gameInstance *game.Game,
	playerID string,
	log *slog.Logger,
) error {
	phase := gameInstance.CurrentPhase()
	if phase != shared.GamePhaseInitApplyCorp && phase != shared.GamePhaseInitApplyPrelude {
		return ValidateCurrentTurn(gameInstance, playerID, log)
	}
	turnOrder := gameInstance.TurnOrder()
	idx := gameInstance.InitPhasePlayerIndex()
	if idx >= len(turnOrder) || turnOrder[idx] != playerID {
		log.Warn("Not the init player", slog.String("player_id", playerID))
		return fmt.Errorf("not your turn")
	}
	return nil
}

// ValidateActionsRemaining validates that the current player has actions remaining
// Returns error if actionsRemaining == 0; allows -1 (unlimited) and >0
func ValidateActionsRemaining(
	gameInstance *game.Game,
	playerID string,
	log *slog.Logger,
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
			slog.String("player_id", playerID),
			slog.Int("actions_remaining", remaining))
		return fmt.Errorf("no actions remaining")
	}

	return nil
}

// ValidateNoPendingSelections validates that the player has no pending selections.
// Actions should be blocked while a player is resolving a pending selection.
func ValidateNoPendingSelections(
	gameInstance *game.Game,
	playerID string,
	log *slog.Logger,
) error {
	if gameInstance.CurrentPhase() == shared.GamePhaseAction && gameInstance.GetForcedFirstAction(playerID) != nil {
		return fmt.Errorf("corporation first action must finish first")
	}
	if gameInstance.HasAnyPendingSelection(playerID) {
		return fmt.Errorf("pending selection")
	}
	return nil
}
