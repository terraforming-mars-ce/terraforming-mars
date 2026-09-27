package game

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/google/uuid"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// CreateGameAction handles the business logic for creating new games
type CreateGameAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	mapRegistry  *board.MapRegistry
	logger       *slog.Logger
}

// NewCreateGameAction creates a new create game action
func NewCreateGameAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	mapRegistry *board.MapRegistry,
	logger *slog.Logger,
) *CreateGameAction {
	return &CreateGameAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		mapRegistry:  mapRegistry,
		logger:       logger,
	}
}

// Execute performs the create game action
func (a *CreateGameAction) Execute(
	ctx context.Context,
	settings shared.GameSettings,
) (*game.Game, error) {
	log := a.logger.With(
		slog.Int("max_players", settings.MaxPlayers),
		slog.Any("card_packs", settings.CardPacks),
	)
	log.Debug("Creating new game")

	// 1. Generate game ID
	gameID := uuid.New().String()

	// 2. Apply default settings
	if settings.MaxPlayers == 0 {
		settings.MaxPlayers = game.DefaultMaxPlayers
	}
	if settings.MapID == "" {
		settings.MapID = board.DefaultMapID()
	}
	if len(settings.CardPacks) == 0 {
		settings.CardPacks = shared.DefaultCardPacks()
	}
	if settings.VenusNextEnabled && !slices.Contains(settings.CardPacks, shared.PackVenus) {
		settings.CardPacks = append(settings.CardPacks, shared.PackVenus)
	}

	// 3. Generate board tiles from selected map
	mapDef, ok := a.mapRegistry.GetMap(settings.MapID)
	if !ok {
		return nil, fmt.Errorf("unknown map: %s", settings.MapID)
	}
	initialTiles := board.GenerateBoardFromMap(mapDef, settings.VenusNextEnabled)

	// 4. Create game entity
	newGame := game.NewGame(a.gameRepo.DataStore(), gameID, "", settings, initialTiles)

	// 4. Initialize deck with cards from selected packs
	projectCardIDs, corpIDs, preludeIDs := cards.GetCardIDsByPacks(a.cardRegistry, settings.CardPacks)
	newGame.InitDeck(projectCardIDs, corpIDs, preludeIDs)
	newGame.SetVPCardLookup(cards.NewVPCardLookupAdapter(a.cardRegistry))
	log.Debug("Deck initialized",
		slog.Int("project_cards", len(projectCardIDs)),
		slog.Int("corporations", len(corpIDs)),
		slog.Int("preludes", len(preludeIDs)),
		slog.Any("first_5_corps", getFirst5(corpIDs)))

	// 5. Store game in repository
	err := a.gameRepo.Create(ctx, newGame)
	if err != nil {
		log.Error("Failed to create game", slog.Any("error", err))
		return nil, err
	}

	// Log the master RNG seed so a reported game can be reproduced deterministically
	// (seed + action sequence) via the replay harness. Kept out of the client DTO on
	// purpose: exposing it would let players predict the deck shuffle.
	log.Info("Game created", slog.String("game_id", gameID), slog.Uint64("seed", newGame.Seed()))
	return newGame, nil
}

// getFirst5 returns up to the first 5 elements of a slice (for logging)
func getFirst5(ids []string) []string {
	if len(ids) <= 5 {
		return ids
	}
	return ids[:5]
}

// GameOptions contains domain settings and boards available when creating a game.
type GameOptions struct {
	Defaults      shared.GameSettings
	AvailableMaps []*board.MapDefinition
}

// Options returns the defaults and boards used by game creation.
func (a *CreateGameAction) Options() GameOptions {
	availableMaps := make([]*board.MapDefinition, 0)
	for _, info := range a.mapRegistry.ListMaps() {
		definition, _ := a.mapRegistry.GetMap(info.ID)
		availableMaps = append(availableMaps, definition)
	}
	return GameOptions{Defaults: shared.GameSettings{
		MaxPlayers: game.DefaultMaxPlayers, MapID: board.DefaultMapID(),
		CardPacks: shared.DefaultCardPacks(), DevelopmentMode: true,
	}, AvailableMaps: availableMaps}
}

// ExecuteSetup validates the complete setup before registering a game.
func (a *CreateGameAction) ExecuteSetup(ctx context.Context, setup *shared.GameSettings) (*game.Game, error) {
	settings := a.Options().Defaults
	if setup != nil {
		settings = *setup
	}
	if err := validateSetup(settings, a.mapRegistry); err != nil {
		return nil, err
	}
	return a.Execute(ctx, settings)
}

func validateSetup(settings shared.GameSettings, registry *board.MapRegistry) error {
	if settings.MaxPlayers < 1 || settings.MaxPlayers > 10 {
		return fmt.Errorf("max players must be between 1 and 10")
	}
	if _, ok := registry.GetMap(settings.MapID); !ok {
		return fmt.Errorf("unknown map: %s", settings.MapID)
	}
	if !slices.Contains(settings.CardPacks, shared.PackBaseGame) {
		return fmt.Errorf("base game pack cannot be disabled")
	}
	return nil
}
