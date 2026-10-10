package game

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"openmars/internal/action"
	"openmars/internal/game"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
)

// BotLifecycle assigns bot identities and checks a new bot's credential in the background.
type BotLifecycle interface {
	AssignIdentity(seed uint64, takenNames []string) (name, personaID string)
	PrepareBot(gameID, playerID string)
}

// AddBotAction handles adding a bot player to a game lobby
type AddBotAction struct {
	gameRepo          game.GameRepository
	cardRegistry      gamecards.CardRegistry
	colonyBonusLookup gamecards.ColonyBonusLookup
	bots              BotLifecycle
	logger            *slog.Logger
}

// AddBotResult contains the result of adding a bot
type AddBotResult struct {
	PlayerID string
}

// NewAddBotAction creates a new add bot action
func NewAddBotAction(
	gameRepo game.GameRepository,
	cardRegistry gamecards.CardRegistry,
	bots BotLifecycle,
	logger *slog.Logger,
	colonyBonusLookup ...gamecards.ColonyBonusLookup,
) *AddBotAction {
	var lookup gamecards.ColonyBonusLookup
	if len(colonyBonusLookup) > 0 {
		lookup = colonyBonusLookup[0]
	}
	return &AddBotAction{
		gameRepo:          gameRepo,
		cardRegistry:      cardRegistry,
		colonyBonusLookup: lookup,
		bots:              bots,
		logger:            logger,
	}
}

// Execute adds a bot player to the game lobby. Only the host may add bots, since bots
// spend the host's Claude token.
func (a *AddBotAction) Execute(ctx context.Context, gameID, requesterID string) (*AddBotResult, error) {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("action", "add_bot"),
	)
	log.Debug("Adding bot to game")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Warn("Game not found", slog.Any("error", err))
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	if g.HostPlayerID() != requesterID {
		return nil, fmt.Errorf("only the host can add bots")
	}

	if g.Status() != shared.GameStatusLobby {
		return nil, fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	if g.Settings().ClaudeOAuthToken == "" {
		return nil, fmt.Errorf("a Claude token is required to add bots")
	}

	existingPlayers := g.GetAllPlayers()
	maxPlayers := g.Settings().MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = game.DefaultMaxPlayers
	}
	if len(existingPlayers) >= maxPlayers {
		return nil, fmt.Errorf("game is full")
	}

	takenNames := make([]string, 0, len(existingPlayers))
	for _, p := range existingPlayers {
		takenNames = append(takenNames, p.Name())
	}
	botName, personaID := a.bots.AssignIdentity(g.Seed(), takenNames)

	botID := uuid.New().String()
	botPlayer, err := g.AddNewBotPlayer(ctx, botID, botName, personaID)
	if err != nil {
		log.Error("Failed to add bot to game", slog.Any("error", err))
		return nil, fmt.Errorf("failed to add bot to game: %w", err)
	}
	action.SetupPlayerCardStore(botPlayer, g, a.cardRegistry, a.colonyBonusLookup)

	a.bots.PrepareBot(gameID, botID)

	log.Info("Bot added to game", slog.String("bot_id", botID), slog.String("bot_name", botName), slog.String("persona", personaID))
	return &AddBotResult{PlayerID: botID}, nil
}
