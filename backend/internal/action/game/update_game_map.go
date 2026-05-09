package game

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/maps"
)

// UpdateGameMapAction handles changing the map during lobby phase
type UpdateGameMapAction struct {
	gameRepo    game.GameRepository
	mapRegistry *maps.MapRegistry
	logger      *zap.Logger
}

// NewUpdateGameMapAction creates a new update game map action
func NewUpdateGameMapAction(
	gameRepo game.GameRepository,
	mapRegistry *maps.MapRegistry,
	logger *zap.Logger,
) *UpdateGameMapAction {
	return &UpdateGameMapAction{
		gameRepo:    gameRepo,
		mapRegistry: mapRegistry,
		logger:      logger,
	}
}

// Execute changes the map for a game in lobby phase
func (a *UpdateGameMapAction) Execute(ctx context.Context, gameID string, playerID string, mapID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("map_id", mapID),
	)

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	if g.Status() != shared.GameStatusLobby {
		return fmt.Errorf("map can only be changed during lobby phase")
	}

	if g.HostPlayerID() != playerID {
		return fmt.Errorf("only the host can change the map")
	}

	mapDef, ok := a.mapRegistry.GetMap(mapID)
	if !ok {
		return fmt.Errorf("unknown map: %s", mapID)
	}

	settings := g.Settings()
	settings.MapID = mapID
	g.UpdateSettings(ctx, settings)

	initialTiles := maps.GenerateBoardFromMap(mapDef, settings.VenusNextEnabled)
	g.ReplaceBoard(ctx, initialTiles)

	log.Debug("Game map updated", zap.String("map_name", mapDef.Name))
	return nil
}
