package game

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/deck"
)

// CreateGameAction handles the business logic for creating new games
type CreateGameAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewCreateGameAction creates a new create game action
func NewCreateGameAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *CreateGameAction {
	return &CreateGameAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the create game action
func (a *CreateGameAction) Execute(
	ctx context.Context,
	settings game.GameSettings,
) (*game.Game, error) {
	log := a.logger.With(
		zap.Int("max_players", settings.MaxPlayers),
		zap.Strings("card_packs", settings.CardPacks),
	)
	log.Debug("Creating new game")

	// 1. Generate game ID
	gameID := uuid.New().String()

	// 2. Apply default settings
	if settings.MaxPlayers == 0 {
		settings.MaxPlayers = game.DefaultMaxPlayers
	}
	if len(settings.CardPacks) == 0 {
		settings.CardPacks = game.DefaultCardPacks()
	}
	if settings.VenusNextEnabled && !slices.Contains(settings.CardPacks, game.PackVenus) {
		settings.CardPacks = append(settings.CardPacks, game.PackVenus)
	}

	// 3. Create game entity
	// Note: hostPlayerID is empty initially, will be set when first player joins
	// Board is automatically created by NewGame
	// EventBus is created per-game for synchronous event handling
	newGame := game.NewGame(gameID, "", settings)

	// 4. Initialize deck with cards from selected packs
	projectCardIDs, corpIDs, preludeIDs := cards.GetCardIDsByPacks(a.cardRegistry, settings.CardPacks)
	gameDeck := deck.NewDeck(gameID, projectCardIDs, corpIDs, preludeIDs)
	newGame.SetDeck(gameDeck)
	newGame.SetVPCardLookup(cards.NewVPCardLookupAdapter(a.cardRegistry))
	log.Debug("Deck initialized",
		zap.Int("project_cards", len(projectCardIDs)),
		zap.Int("corporations", len(corpIDs)),
		zap.Int("preludes", len(preludeIDs)),
		zap.Strings("first_5_corps", getFirst5(corpIDs)))

	// 5. Store game in repository
	err := a.gameRepo.Create(ctx, newGame)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return nil, err
	}

	log.Info("Game created", zap.String("game_id", gameID))
	return newGame, nil
}

// getFirst5 returns up to the first 5 elements of a slice (for logging)
func getFirst5(ids []string) []string {
	if len(ids) <= 5 {
		return ids
	}
	return ids[:5]
}
