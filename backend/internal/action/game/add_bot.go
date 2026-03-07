package game

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
)

var botNames = []string{
	"HAL 9000", "GLaDOS", "SHODAN", "Cortana", "JARVIS",
	"Deep Thought", "WOPR", "MU-TH-UR", "Skynet", "Data",
	"Bishop", "Ash", "CASE", "TARS", "Marvin",
	"R2-D2", "C-3PO", "Wall-E", "Bender", "Sonny",
}

// BotHealthChecker verifies that a Claude API key is valid and generates a greeting.
type BotHealthChecker interface {
	CheckHealth(ctx context.Context, apiKey, model, botName, difficulty string) (string, error)
}

// BotBroadcaster broadcasts game state and chat updates.
type BotBroadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
	BroadcastChatMessage(gameID string, chatMsg dto.ChatMessageDto)
}

// AddBotAction handles adding a bot player to a game lobby
type AddBotAction struct {
	gameRepo      game.GameRepository
	cardRegistry  cards.CardRegistry
	healthChecker BotHealthChecker
	broadcaster   BotBroadcaster
	logger        *zap.Logger
}

// AddBotResult contains the result of adding a bot
type AddBotResult struct {
	PlayerID string
	GameDto  dto.GameDto
}

// NewAddBotAction creates a new add bot action
func NewAddBotAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	healthChecker BotHealthChecker,
	broadcaster BotBroadcaster,
	logger *zap.Logger,
) *AddBotAction {
	return &AddBotAction{
		gameRepo:      gameRepo,
		cardRegistry:  cardRegistry,
		healthChecker: healthChecker,
		broadcaster:   broadcaster,
		logger:        logger,
	}
}

// Execute adds a bot player to the game lobby
func (a *AddBotAction) Execute(ctx context.Context, gameID string, botName string, difficulty string, speed string) (*AddBotResult, error) {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("bot_name", botName),
		zap.String("action", "add_bot"),
	)
	log.Debug("Adding bot to game")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	if g.Status() != game.GameStatusLobby {
		log.Warn("Game is not in lobby", zap.String("status", string(g.Status())))
		return nil, fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	if g.Settings().ClaudeAPIKey == "" {
		return nil, fmt.Errorf("claude API key is required to add bots (set claudeApiKey in game settings)")
	}

	existingPlayers := g.GetAllPlayers()
	maxPlayers := g.Settings().MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = game.DefaultMaxPlayers
	}
	if len(existingPlayers) >= maxPlayers {
		return nil, fmt.Errorf("game is full")
	}

	if botName == "" {
		botName = a.generateBotName(existingPlayers)
	}

	botDifficulty := playerPkg.BotDifficulty(difficulty)
	if botDifficulty != playerPkg.BotDifficultyNormal && botDifficulty != playerPkg.BotDifficultyHard && botDifficulty != playerPkg.BotDifficultyExtreme {
		botDifficulty = playerPkg.BotDifficultyNormal
	}

	botSpeed := playerPkg.BotSpeed(speed)
	if botSpeed != playerPkg.BotSpeedFast && botSpeed != playerPkg.BotSpeedNormal && botSpeed != playerPkg.BotSpeedThinker {
		botSpeed = playerPkg.BotSpeedNormal
	}

	botID := uuid.New().String()
	bot := playerPkg.NewBotPlayer(g.EventBus(), gameID, botID, botName, botDifficulty, botSpeed)

	if err := g.AddPlayer(ctx, bot); err != nil {
		log.Error("Failed to add bot to game", zap.Error(err))
		return nil, fmt.Errorf("failed to add bot to game: %w", err)
	}

	log.Info("Bot added to game", zap.String("bot_id", botID), zap.String("bot_name", botName))

	if a.healthChecker != nil && a.broadcaster != nil {
		settings := g.Settings()
		go a.runHealthCheck(gameID, botID, botName, difficulty, settings.ClaudeAPIKey, settings.ClaudeModel, log)
	}

	gameDto := dto.ToGameDto(g, a.cardRegistry, botID)
	return &AddBotResult{
		PlayerID: botID,
		GameDto:  gameDto,
	}, nil
}

func (a *AddBotAction) runHealthCheck(gameID, botID, botName, difficulty, apiKey, model string, log *zap.Logger) {
	log.Debug("Running Claude health check for bot", zap.String("bot_id", botID))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	greeting, err := a.healthChecker.CheckHealth(ctx, apiKey, model, botName, difficulty)

	g, gErr := a.gameRepo.Get(ctx, gameID)
	if gErr != nil {
		log.Error("Failed to get game after health check", zap.Error(gErr))
		return
	}

	bot, bErr := g.GetPlayer(botID)
	if bErr != nil {
		log.Error("Bot not found after health check", zap.Error(bErr))
		return
	}

	if err != nil {
		log.Error("Health check failed for bot", zap.String("bot_id", botID), zap.Error(err))
		bot.SetBotStatus(playerPkg.BotStatusFailed)
		a.broadcaster.BroadcastGameState(gameID, nil)
		return
	}

	log.Debug("Bot health check passed", zap.String("bot_id", botID))
	bot.SetBotStatus(playerPkg.BotStatusReady)
	a.broadcaster.BroadcastGameState(gameID, nil)

	if greeting != "" {
		chatMsg := game.ChatMessage{
			SenderID:    botID,
			SenderName:  botName,
			SenderColor: bot.Color(),
			Message:     greeting,
			Timestamp:   time.Now(),
		}
		g.AddChatMessage(ctx, chatMsg)
		a.broadcaster.BroadcastChatMessage(gameID, dto.ChatMessageDto{
			SenderID:    chatMsg.SenderID,
			SenderName:  chatMsg.SenderName,
			SenderColor: chatMsg.SenderColor,
			Message:     chatMsg.Message,
			Timestamp:   chatMsg.Timestamp.Format(time.RFC3339),
		})
	}
}

func (a *AddBotAction) generateBotName(existingPlayers []*playerPkg.Player) string {
	taken := make(map[string]bool, len(existingPlayers))
	for _, p := range existingPlayers {
		taken[p.Name()] = true
	}

	// Shuffle and pick the first available name
	perm := rand.Perm(len(botNames))
	for _, i := range perm {
		if !taken[botNames[i]] {
			return botNames[i]
		}
	}

	// Fallback if all names taken
	for i := 1; ; i++ {
		name := fmt.Sprintf("Claude Bot %d", i)
		if !taken[name] {
			return name
		}
	}
}
