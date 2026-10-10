package connection

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	"openmars/internal/game/shared"
)

// SetPlayerColorAction handles changing a player's color during the lobby phase.
type SetPlayerColorAction struct {
	gameRepo game.GameRepository
	logger   *slog.Logger
}

// NewSetPlayerColorAction creates a new SetPlayerColorAction.
func NewSetPlayerColorAction(
	gameRepo game.GameRepository,
	logger *slog.Logger,
) *SetPlayerColorAction {
	return &SetPlayerColorAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute changes a player's color if the game is in lobby and the color is available.
// requesterID is the player making the request, targetPlayerID is the player whose color is being changed.
// A player can change their own color, or the host can change a bot's color.
func (a *SetPlayerColorAction) Execute(ctx context.Context, gameID, requesterID, targetPlayerID, color string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("requester_id", requesterID),
		slog.String("target_player_id", targetPlayerID),
		slog.String("color", color),
		slog.String("action", "set_player_color"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", slog.Any("error", err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() != shared.GameStatusLobby {
		return fmt.Errorf("can only change color during lobby phase")
	}

	if requesterID != targetPlayerID {
		if g.HostPlayerID() != requesterID {
			return fmt.Errorf("only the host can change another player's color")
		}
		target, err := g.GetPlayer(targetPlayerID)
		if err != nil {
			return fmt.Errorf("player not found: %s", targetPlayerID)
		}
		if !target.IsBot() {
			return fmt.Errorf("can only change bot colors, not other human players")
		}
	}

	if !g.IsPlayerColorAvailable(color, targetPlayerID) {
		return fmt.Errorf("color %s is not available", color)
	}

	p, err := g.GetPlayer(targetPlayerID)
	if err != nil {
		log.Error("Player not found", slog.Any("error", err))
		return fmt.Errorf("player not found: %s", targetPlayerID)
	}

	p.SetColor(color)

	log.Debug("Player color changed")
	return nil
}
