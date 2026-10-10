package game

import (
	"context"
	"fmt"
	"log/slog"

	gamePkg "openmars/internal/game"
	playerPkg "openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// BotStarter starts a bot session for a player.
type BotStarter interface {
	StartBot(gameID, playerID string) error
}

// ConvertToBotAction converts a human player to a bot in an active game.
type ConvertToBotAction struct {
	gameRepo   gamePkg.GameRepository
	botStarter BotStarter
	logger     *slog.Logger
}

func NewConvertToBotAction(
	gameRepo gamePkg.GameRepository,
	botStarter BotStarter,
	logger *slog.Logger,
) *ConvertToBotAction {
	return &ConvertToBotAction{
		gameRepo:   gameRepo,
		botStarter: botStarter,
		logger:     logger,
	}
}

func (a *ConvertToBotAction) Execute(ctx context.Context, gameID string, requesterID string, targetPlayerID string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("requester_id", requesterID),
		slog.String("target_player_id", targetPlayerID),
		slog.String("action", "convert_to_bot"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() != shared.GameStatusActive {
		return fmt.Errorf("game is not active")
	}

	if g.HostPlayerID() != requesterID {
		return fmt.Errorf("only host can convert players to bots")
	}

	if requesterID == targetPlayerID {
		return fmt.Errorf("cannot convert yourself to a bot")
	}

	target, err := g.GetPlayer(targetPlayerID)
	if err != nil {
		return fmt.Errorf("player not found: %s", targetPlayerID)
	}

	if target.IsBot() {
		return fmt.Errorf("player is already a bot")
	}

	if target.HasExited() {
		return fmt.Errorf("cannot convert exited player to bot")
	}

	if g.Settings().ClaudeOAuthToken == "" {
		return fmt.Errorf("a Claude token is required to convert players to bots")
	}

	log.Debug("Converting player to bot", slog.String("player_name", target.Name()))

	target.SetPlayerType(playerPkg.PlayerTypeBot)
	target.SetBotStatus(playerPkg.BotStatusReady)
	target.SetBotError("")
	target.SetConnected(true)

	if a.botStarter != nil {
		if err := a.botStarter.StartBot(gameID, targetPlayerID); err != nil {
			log.Error("Failed to start bot session", slog.Any("error", err))
			target.SetBotStatus(playerPkg.BotStatusFailed)
			target.SetBotError("the bot could not start")
		}
	}

	log.Info("Player converted to bot", slog.String("player_name", target.Name()))
	return nil
}
