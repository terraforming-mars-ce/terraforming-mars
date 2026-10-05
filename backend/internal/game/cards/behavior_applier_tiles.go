package cards

import (
	"context"
	"fmt"
	"log/slog"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyTilePlacementOutput(ctx context.Context, o *shared.TilePlacementCondition, amount int, log *slog.Logger) error {
	if a.game == nil {
		return fmt.Errorf("cannot apply tile placement: no game context")
	}
	if a.player == nil {
		return fmt.Errorf("cannot apply tile placement: no player context")
	}

	rt := o.ResourceType
	if rt == shared.ResourceOceanPlacement && a.game.GlobalParameters().Oceans() >= a.game.GlobalParameters().GetMaxOceans() {
		return nil
	}

	// Land claim is a special tile type
	if rt == shared.ResourceLandClaim {
		tileTypes := make([]string, amount)
		for i := range tileTypes {
			tileTypes[i] = "land-claim"
		}
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, o.TileRestrictions.Clone()); err != nil {
			return fmt.Errorf("failed to append land claim to pending tile selection queue: %w", err)
		}
		log.Debug("Added land claim tile selection to queue", slog.Int("count", amount))
		return nil
	}

	// Generic tile-placement uses the TileType field
	if rt == shared.ResourceTilePlacement {
		if o.TileType == "" {
			return fmt.Errorf("tile-placement output missing tileType field")
		}
		tileTypes := make([]string, amount)
		for i := range tileTypes {
			tileTypes[i] = o.TileType
		}
		restrictions := o.TileRestrictions.Clone()
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, restrictions); err != nil {
			return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
		}
		log.Debug("Added special tile placements to queue",
			slog.String("tile_type", o.TileType), slog.Int("count", amount), slog.Any("tile_restrictions", restrictions))
		return nil
	}

	// Standard placements: city, greenery, ocean, volcano
	var tileType string
	switch rt {
	case shared.ResourceCityPlacement:
		tileType = "city"
	case shared.ResourceGreeneryPlacement:
		tileType = "greenery"
	case shared.ResourceOceanPlacement:
		tileType = "ocean"
	case shared.ResourceVolcanoPlacement:
		tileType = "volcano"
	default:
		return fmt.Errorf("unknown tile placement type: %s", rt)
	}

	tileTypes := make([]string, amount)
	for i := range tileTypes {
		tileTypes[i] = tileType
	}

	restrictions := o.TileRestrictions.Clone()

	// For greenery, enforce adjacency to owned tiles unless card overrides
	if rt == shared.ResourceGreeneryPlacement {
		if restrictions == nil {
			restrictions = &shared.TileRestrictions{AdjacentToOwned: true}
		} else if restrictions.Area == "" {
			restrictions.AdjacentToOwned = true
		}
	}

	if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, restrictions); err != nil {
		return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
	}
	log.Debug("Added tile placements to queue",
		slog.String("tile_type", tileType), slog.Int("count", amount), slog.Any("tile_restrictions", restrictions))
	return nil
}

func (a *BehaviorApplier) applyTileModificationOutput(ctx context.Context, o *shared.TileModificationCondition, amount int, log *slog.Logger) error {
	if a.game == nil {
		return fmt.Errorf("cannot apply tile modification: no game context")
	}
	if a.player == nil {
		return fmt.Errorf("cannot apply tile modification: no player context")
	}

	switch o.ResourceType {
	case shared.ResourceTileDestruction:
		tileTypes := make([]string, amount)
		for i := range tileTypes {
			tileTypes[i] = "tile-destruction"
		}
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, nil); err != nil {
			return fmt.Errorf("failed to append tile destruction to queue: %w", err)
		}
		log.Debug("Added tile destruction selection to queue")

	case shared.ResourceTileReplacement:
		tileTypes := make([]string, amount)
		for i := range tileTypes {
			tileTypes[i] = "tile-replacement:" + o.TileType
		}
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, nil); err != nil {
			return fmt.Errorf("failed to append tile replacement to queue: %w", err)
		}
		log.Debug("Added tile replacement selection to queue", slog.String("replacement_tile", o.TileType))
	}
	return nil
}
