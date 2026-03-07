package game

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	gamePkg "terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
)

// BotStarter starts a bot session for a player.
type BotStarter interface {
	StartBot(gameID, playerID, botName, difficulty, speed string, settings gamePkg.GameSettings) error
}

// ConvertToBotAction converts a human player to a bot in an active game.
type ConvertToBotAction struct {
	gameRepo   gamePkg.GameRepository
	botStarter BotStarter
	logger     *zap.Logger
}

func NewConvertToBotAction(
	gameRepo gamePkg.GameRepository,
	botStarter BotStarter,
	logger *zap.Logger,
) *ConvertToBotAction {
	return &ConvertToBotAction{
		gameRepo:   gameRepo,
		botStarter: botStarter,
		logger:     logger,
	}
}

func (a *ConvertToBotAction) Execute(ctx context.Context, gameID string, requesterID string, targetPlayerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("requester_id", requesterID),
		zap.String("target_player_id", targetPlayerID),
		zap.String("action", "convert_to_bot"),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() != gamePkg.GameStatusActive {
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

	if g.Settings().ClaudeAPIKey == "" {
		return fmt.Errorf("claude API key is required to convert players to bots")
	}

	log.Debug("Converting player to bot", zap.String("player_name", target.Name()))

	target.SetPlayerType(playerPkg.PlayerTypeBot)
	target.SetBotDifficulty(playerPkg.BotDifficultyNormal)
	target.SetBotSpeed(playerPkg.BotSpeedFast)
	target.SetBotStatus(playerPkg.BotStatusLoading)
	target.SetConnected(true)

	if a.botStarter != nil {
		settings := g.Settings()
		if err := a.botStarter.StartBot(gameID, targetPlayerID, target.Name(), string(playerPkg.BotDifficultyNormal), string(playerPkg.BotSpeedFast), settings); err != nil {
			log.Error("Failed to start bot session", zap.Error(err))
			target.SetBotStatus(playerPkg.BotStatusFailed)
		}
	}

	log.Info("Player converted to bot", zap.String("player_name", target.Name()))
	return nil
}
