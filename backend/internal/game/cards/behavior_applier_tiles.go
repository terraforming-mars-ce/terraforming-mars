package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyTilePlacementOutput(ctx context.Context, o *shared.TilePlacementCondition, amount int, log *zap.Logger) error {
	if a.game == nil {
		return fmt.Errorf("cannot apply tile placement: no game context")
	}
	if a.player == nil {
		return fmt.Errorf("cannot apply tile placement: no player context")
	}

	rt := o.ResourceType

	// Land claim is a special tile type
	if rt == shared.ResourceLandClaim {
		tileTypes := make([]string, amount)
		for i := range tileTypes {
			tileTypes[i] = "land-claim"
		}
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, nil); err != nil {
			return fmt.Errorf("failed to append land claim to pending tile selection queue: %w", err)
		}
		log.Debug("Added land claim tile selection to queue", zap.Int("count", amount))
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
		restrictions := copyTileRestrictions(o.TileRestrictions)
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, restrictions); err != nil {
			return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
		}
		log.Debug("Added special tile placements to queue",
			zap.String("tile_type", o.TileType), zap.Int("count", amount), zap.Any("tile_restrictions", restrictions))
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

	restrictions := copyTileRestrictions(o.TileRestrictions)

	// For greenery, enforce adjacency to owned tiles unless card overrides
	if rt == shared.ResourceGreeneryPlacement {
		if restrictions == nil {
			restrictions = &shared.TileRestrictions{AdjacentToOwned: true}
		} else if restrictions.OnTileType == "" {
			restrictions.AdjacentToOwned = true
		}
	}

	if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, restrictions); err != nil {
		return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
	}
	log.Debug("Added tile placements to queue",
		zap.String("tile_type", tileType), zap.Int("count", amount), zap.Any("tile_restrictions", restrictions))
	return nil
}

// copyTileRestrictions creates a deep copy of TileRestrictions, or returns nil.
func copyTileRestrictions(tr *shared.TileRestrictions) *shared.TileRestrictions {
	if tr == nil {
		return nil
	}
	result := *tr
	if tr.BoardTags != nil {
		bt := make([]string, len(tr.BoardTags))
		copy(bt, tr.BoardTags)
		result.BoardTags = bt
	}
	if tr.OnBonusType != nil {
		ob := make([]string, len(tr.OnBonusType))
		copy(ob, tr.OnBonusType)
		result.OnBonusType = ob
	}
	if tr.MinAdjacentOfType != nil {
		v := *tr.MinAdjacentOfType
		result.MinAdjacentOfType = &v
	}
	return &result
}

func (a *BehaviorApplier) applyTileModificationOutput(ctx context.Context, o *shared.TileModificationCondition, amount int, log *zap.Logger) error {
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
		log.Debug("Added tile replacement selection to queue", zap.String("replacement_tile", o.TileType))
	}
	return nil
}
