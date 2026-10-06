package game

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// SettingsPatch is the domain-level, transport-agnostic patch of editable lobby
// settings. The delivery layer maps its request DTO onto this so the action does not
// depend on delivery/dto.
type SettingsPatch struct {
	MaxPlayers       *int
	MapID            *string
	VenusNextEnabled *bool
	DevelopmentMode  *bool
	DemoGame         *bool
	AllowRandomBuy   *bool
	CardPacks        *[]string
	ClaudeOAuthToken *string
	BotSpendCapUSD   *float64
}

// UpdateGameSettingsAction handles editing game settings during the lobby phase.
type UpdateGameSettingsAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	mapRegistry  *board.MapRegistry
	logger       *slog.Logger
}

// NewUpdateGameSettingsAction creates a new update game settings action.
func NewUpdateGameSettingsAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	mapRegistry *board.MapRegistry,
	logger *slog.Logger,
) *UpdateGameSettingsAction {
	return &UpdateGameSettingsAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		mapRegistry:  mapRegistry,
		logger:       logger,
	}
}

// Execute applies the patch to the game's settings. Host-only, lobby-only.
func (a *UpdateGameSettingsAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	patch *SettingsPatch,
) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	if g.Status() != shared.GameStatusLobby {
		return fmt.Errorf("settings can only be changed during lobby phase")
	}

	if g.HostPlayerID() != playerID {
		return fmt.Errorf("only the host can change settings")
	}

	settings := g.Settings()
	cardPacksChanged := false
	venusChanged := false
	mapChanged := false
	demoTurnedOff := false

	if patch.MaxPlayers != nil {
		newMax := *patch.MaxPlayers
		if newMax < 1 || newMax > 10 {
			return fmt.Errorf("max players must be between 1 and 10")
		}
		joined := len(g.GetAllPlayers())
		if newMax < joined {
			return fmt.Errorf("cannot reduce max players below current player count (%d)", joined)
		}
		settings.MaxPlayers = newMax
	}

	if patch.MapID != nil {
		if _, ok := a.mapRegistry.GetMap(*patch.MapID); !ok {
			return fmt.Errorf("unknown map: %s", *patch.MapID)
		}
		if settings.MapID != *patch.MapID {
			settings.MapID = *patch.MapID
			mapChanged = true
		}
	}

	if patch.VenusNextEnabled != nil && settings.VenusNextEnabled != *patch.VenusNextEnabled {
		settings.VenusNextEnabled = *patch.VenusNextEnabled
		venusChanged = true
	}

	if patch.DevelopmentMode != nil {
		settings.DevelopmentMode = *patch.DevelopmentMode
	}

	if patch.DemoGame != nil {
		if settings.DemoGame && !*patch.DemoGame {
			demoTurnedOff = true
		}
		settings.DemoGame = *patch.DemoGame
	}

	if patch.AllowRandomBuy != nil {
		settings.AllowRandomBuy = *patch.AllowRandomBuy
	}

	if patch.CardPacks != nil {
		packs := *patch.CardPacks
		if len(packs) == 0 {
			return fmt.Errorf("at least one card pack must be enabled")
		}
		if !slices.Contains(packs, shared.PackBaseGame) {
			return fmt.Errorf("base game pack cannot be disabled")
		}
		if !equalStringSlices(settings.CardPacks, packs) {
			settings.CardPacks = packs
			cardPacksChanged = true
		}
	}

	if patch.ClaudeOAuthToken != nil {
		settings.ClaudeOAuthToken = *patch.ClaudeOAuthToken
	}
	if patch.BotSpendCapUSD != nil {
		if *patch.BotSpendCapUSD <= 0 {
			return fmt.Errorf("bot spend cap must be positive")
		}
		settings.BotSpendCapUSD = *patch.BotSpendCapUSD
	}

	if demoTurnedOff {
		settings.Temperature = nil
		settings.Oxygen = nil
		settings.Oceans = nil
		settings.Venus = nil
		settings.Generation = nil
		for _, p := range g.GetAllPlayers() {
			p.SetPendingDemoChoices(nil)
		}
	}

	if err := validateSetup(settings, a.mapRegistry); err != nil {
		return err
	}
	g.UpdateSettings(ctx, settings)

	if cardPacksChanged || venusChanged {
		effectivePacks := settings.CardPacks
		if settings.VenusNextEnabled && !slices.Contains(effectivePacks, shared.PackVenus) {
			effectivePacks = append(effectivePacks, shared.PackVenus)
		}
		projectIDs, corpIDs, preludeIDs := cards.GetCardIDsByPacks(a.cardRegistry, effectivePacks)
		g.InitDeck(projectIDs, corpIDs, preludeIDs)
		log.Debug("Deck rebuilt",
			slog.Int("project_cards", len(projectIDs)),
			slog.Int("corporations", len(corpIDs)),
			slog.Int("preludes", len(preludeIDs)))
	}

	if mapChanged || venusChanged {
		mapDef, _ := a.mapRegistry.GetMap(settings.MapID)
		initialTiles := board.GenerateBoardFromMap(mapDef, settings.VenusNextEnabled)
		g.ReplaceBoard(ctx, initialTiles)
		log.Debug("Board rebuilt", slog.String("map_id", settings.MapID))
	}

	log.Info("Game settings updated")
	return nil
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
