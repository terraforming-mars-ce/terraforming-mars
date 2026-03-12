package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// CreateDemoLobbyAction handles creating a demo game lobby where player count is set
type CreateDemoLobbyAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// DemoLobbySettings contains settings for creating a demo lobby
type DemoLobbySettings struct {
	PlayerCount int      // 1-5, required
	CardPacks   []string // default: ["base-game"]
	PlayerName  string   // name for the human player
}

// DemoLobbyResult contains the result of creating a demo lobby
type DemoLobbyResult struct {
	PlayerID string
	GameDto  dto.GameDto
}

// NewCreateDemoLobbyAction creates a new demo lobby creation action
func NewCreateDemoLobbyAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *CreateDemoLobbyAction {
	return &CreateDemoLobbyAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute creates a demo game lobby with the specified player count
func (a *CreateDemoLobbyAction) Execute(
	ctx context.Context,
	settings DemoLobbySettings,
) (*DemoLobbyResult, error) {
	log := a.logger.With(zap.String("action", "create_demo_lobby"))
	log.Debug("Creating demo lobby", zap.Int("player_count", settings.PlayerCount))

	// Validate required fields
	if settings.PlayerCount < 1 || settings.PlayerCount > 5 {
		return nil, fmt.Errorf("player count must be between 1 and 5, got %d", settings.PlayerCount)
	}

	if len(settings.CardPacks) == 0 {
		settings.CardPacks = []string{"base-game"}
	}
	if settings.PlayerName == "" {
		settings.PlayerName = "You"
	}

	// Generate game ID and create base game
	gameID := uuid.New().String()
	baseSettings := shared.GameSettings{
		MaxPlayers:      settings.PlayerCount,
		CardPacks:       settings.CardPacks,
		DevelopmentMode: true,
		DemoGame:        true,
	}

	newGame := game.NewGame(a.gameRepo.DataStore(), gameID, "", baseSettings)

	projectCardIDs, corpIDs, preludeIDs := cards.GetCardIDsByPacks(a.cardRegistry, settings.CardPacks)
	newGame.InitDeck(projectCardIDs, corpIDs, preludeIDs)
	newGame.SetVPCardLookup(cards.NewVPCardLookupAdapter(a.cardRegistry))
	log.Debug("Deck initialized",
		zap.Int("project_cards", len(projectCardIDs)),
		zap.Int("corporations", len(corpIDs)))

	// Store game in repository
	if err := a.gameRepo.Create(ctx, newGame); err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Set host before adding player (so auto-broadcast includes hostPlayerID)
	playerID := uuid.New().String()
	if err := newGame.SetHostPlayerID(ctx, playerID); err != nil {
		return nil, fmt.Errorf("failed to set host: %w", err)
	}

	// Create and add human player (other players join via normal lobby system)
	demoPlayer, err := newGame.AddNewPlayer(ctx, playerID, settings.PlayerName)
	if err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}
	action.SetupPlayerCardStore(demoPlayer, newGame, a.cardRegistry)
	log.Debug("Human player added", zap.String("player_id", playerID), zap.String("name", settings.PlayerName))

	// Game stays in lobby status - ready for configuration
	log.Info("Demo lobby created", zap.String("game_id", gameID))

	gameDto := dto.ToGameDto(newGame, a.cardRegistry, playerID)
	return &DemoLobbyResult{
		PlayerID: playerID,
		GameDto:  gameDto,
	}, nil
}
