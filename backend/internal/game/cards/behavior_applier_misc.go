package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyColonyOutput(ctx context.Context, o *shared.ColonyCondition, amount int, log *zap.Logger) error {
	switch o.ResourceType {
	case shared.ResourceColonyTile:
		if a.game == nil || a.player == nil {
			return fmt.Errorf("cannot apply colony tile: missing game or player context")
		}
		if !a.game.HasColonies() {
			log.Warn("Colony tile output ignored: colonies expansion not enabled")
			return nil
		}
		colonyIDs := a.game.GetPlaceableColonyIDs(a.player.ID(), o.AllowDuplicatePlayerColony)
		if len(colonyIDs) == 0 {
			log.Warn("No colony tiles available for placement")
			return nil
		}
		a.player.Selection().SetPendingColonySelection(&shared.PendingColonySelection{
			AvailableColonyIDs:         colonyIDs,
			AllowDuplicatePlayerColony: o.AllowDuplicatePlayerColony,
			Source:                     a.source,
			SourceCardID:               a.sourceCardID,
		})
		log.Debug("Set pending colony selection",
			zap.Int("available_colonies", len(colonyIDs)), zap.Bool("allow_duplicate", o.AllowDuplicatePlayerColony))

	case shared.ResourceColonyBonus:
		if a.game == nil || a.player == nil {
			return fmt.Errorf("cannot apply colony bonus: missing game or player context")
		}
		if !a.game.HasColonies() {
			log.Warn("Colony bonus output ignored: colonies expansion not enabled")
			return nil
		}
		if a.colonyBonusLookup == nil {
			return fmt.Errorf("cannot apply colony bonus: no colony bonus lookup configured")
		}
		a.applyColonyBonuses(ctx, log)

	case shared.ResourceColonyCount, shared.ResourceColonyTrackStep:
		log.Debug("Colony count/track output (informational)", zap.String("type", string(o.ResourceType)), zap.Int("amount", amount))

	default:
		log.Warn("Unhandled colony type", zap.String("type", string(o.ResourceType)))
	}
	return nil
}

func (a *BehaviorApplier) applyMiscOutput(ctx context.Context, o *shared.MiscCondition, amount int, log *zap.Logger) error {
	switch o.ResourceType {
	case shared.ResourceExtraActions:
		if a.game == nil {
			return fmt.Errorf("cannot apply extra actions: no game context")
		}
		currentTurn := a.game.CurrentTurn()
		if currentTurn != nil {
			currentTurn.AddExtraActions(amount)
		}
		log.Debug("Granted extra actions", zap.Int("amount", amount))

	case shared.ResourceBonusTags:
		if a.player == nil {
			return fmt.Errorf("cannot apply bonus tags: no player context")
		}
		if o.Per != nil && o.Per.Tag != nil {
			tagToCount := *o.Per.Tag
			tagToGrant := shared.CardTag(o.ResourceType)
			if len(o.Selectors) > 0 && len(o.Selectors[0].Tags) > 0 {
				tagToGrant = o.Selectors[0].Tags[0]
			}
			var tagCount int
			if a.cardRegistry != nil {
				tagCount = CountPlayerTagsByType(a.player, a.cardRegistry, tagToCount)
			}
			bonusCount := tagCount * amount
			if bonusCount > 0 {
				a.player.AddBonusTags(tagToGrant, bonusCount)
			}
			log.Debug("Added bonus tags",
				zap.String("tag_type", string(tagToGrant)), zap.Int("count", bonusCount),
				zap.String("per_tag", string(tagToCount)), zap.Int("tag_count", tagCount))
		}

	case shared.ResourceFreeTrade:
		if a.game == nil || a.player == nil {
			return fmt.Errorf("cannot apply free trade: missing game or player context")
		}
		if !a.game.HasColonies() {
			log.Warn("Free trade output ignored: colonies expansion not enabled")
			return nil
		}
		if !a.game.GetTradeFleetAvailable(a.player.ID()) {
			log.Warn("Free trade output ignored: no trade fleet available")
			return nil
		}
		tradeableColonyIDs := a.game.GetTradeableColonyIDs()
		if len(tradeableColonyIDs) == 0 {
			log.Warn("Free trade output ignored: no colonies available for trading")
			return nil
		}
		a.player.Selection().SetPendingFreeTradeSelection(&shared.PendingFreeTradeSelection{
			AvailableColonyIDs: tradeableColonyIDs,
			Source:             a.source,
			SourceCardID:       a.sourceCardID,
		})
		log.Debug("Set pending free trade selection", zap.Int("available_colonies", len(tradeableColonyIDs)))

	case shared.ResourceWorldTreeTile:
		log.Debug("World tree tile output", zap.Int("amount", amount))

	case shared.ResourceAwardFund:
		log.Debug("Award fund output", zap.Int("amount", amount))

	default:
		log.Warn("Unhandled misc output type", zap.String("type", string(o.ResourceType)))
	}
	return nil
}
