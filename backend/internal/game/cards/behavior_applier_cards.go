package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game/shared"
)

func (a *BehaviorApplier) applyCardStorageOutput(ctx context.Context, o *shared.CardStorageCondition, amount int, log *zap.Logger) error {
	if a.player == nil {
		return fmt.Errorf("cannot apply card resource: no player context")
	}
	rt := o.ResourceType

	// Generic card-resource: add resources of whatever type the target card stores
	if rt == shared.ResourceCardResource {
		targetID := a.nextTargetCardID()
		if targetID == "" {
			return fmt.Errorf("no target card specified for card-resource output")
		}
		if a.cardRegistry == nil {
			return fmt.Errorf("cannot apply card-resource: no card registry")
		}
		targetCard, err := a.cardRegistry.GetByID(targetID)
		if err != nil {
			return fmt.Errorf("target card not found in registry: %w", err)
		}
		if targetCard.ResourceStorage == nil {
			return fmt.Errorf("target card %s has no resource storage", targetID)
		}
		a.player.Resources().AddToStorage(targetID, amount)
		log.Debug("Added card-resource to target card storage",
			zap.String("card_id", targetID), zap.String("storage_type", string(targetCard.ResourceStorage.Type)), zap.Int("amount", amount))
		return nil
	}

	// Specific storage types (animal, microbe, floater, etc.)
	switch o.Target {
	case "self-card":
		if a.sourceCardID == "" {
			log.Warn("Cannot place resource on self-card: no source card ID", zap.String("resource_type", string(rt)))
			return nil
		}
		a.player.Resources().AddToStorage(a.sourceCardID, amount)
		log.Debug("Added resource to card storage",
			zap.String("card_id", a.sourceCardID), zap.String("resource_type", string(rt)), zap.Int("amount", amount))

	case "steal-from-any-card":
		if a.stealSourceCardID == "" {
			return fmt.Errorf("steal-from-any-card requires a source card ID")
		}
		if a.game == nil {
			return fmt.Errorf("cannot steal from card: no game context")
		}
		stolenAmount := 0
		for _, p := range a.game.GetAllPlayers() {
			storage := p.Resources().GetCardStorage(a.stealSourceCardID)
			if storage > 0 {
				stolenAmount = min(amount, storage)
				p.Resources().AddToStorage(a.stealSourceCardID, -stolenAmount)
				log.Debug("Stole resource from card",
					zap.String("source_card_id", a.stealSourceCardID), zap.String("owner_player_id", p.ID()),
					zap.String("resource_type", string(rt)), zap.Int("amount", stolenAmount))
				break
			}
		}
		if stolenAmount > 0 && a.sourceCardID != "" {
			a.player.Resources().AddToStorage(a.sourceCardID, stolenAmount)
			log.Debug("Added stolen resource to self card",
				zap.String("card_id", a.sourceCardID), zap.String("resource_type", string(rt)), zap.Int("amount", stolenAmount))
		}

	case "any-card":
		targetID := a.nextTargetCardID()
		if targetID == "" {
			log.Warn("No target card for any-card resource placement — resources lost",
				zap.String("resource_type", string(rt)), zap.Int("amount", amount))
			return nil
		}
		if a.cardRegistry != nil {
			targetCard, err := a.cardRegistry.GetByID(targetID)
			if err != nil {
				return fmt.Errorf("target card %s not found in registry: %w", targetID, err)
			}
			if targetCard.ResourceStorage == nil {
				return fmt.Errorf("target card %s has no resource storage", targetID)
			}
			if targetCard.ResourceStorage.Type != rt {
				return fmt.Errorf("target card %s stores %s, cannot add %s", targetID, targetCard.ResourceStorage.Type, rt)
			}
		}
		a.player.Resources().AddToStorage(targetID, amount)
		log.Debug("Added resource to target card storage",
			zap.String("card_id", targetID), zap.String("resource_type", string(rt)), zap.Int("amount", amount))

	default:
		if a.sourceCardID != "" {
			a.player.Resources().AddToStorage(a.sourceCardID, amount)
			log.Debug("Added resource to card storage (default to self)",
				zap.String("card_id", a.sourceCardID), zap.String("resource_type", string(rt)), zap.Int("amount", amount))
		} else {
			log.Warn("Unhandled target for card resource", zap.String("target", o.Target), zap.String("resource_type", string(rt)))
		}
	}
	return nil
}

func (a *BehaviorApplier) applyCardOperationOutput(ctx context.Context, o *shared.CardOperationCondition, amount int, log *zap.Logger) error {
	switch o.ResourceType {
	case shared.ResourceCardDraw:
		if a.game == nil || a.player == nil {
			return fmt.Errorf("cannot apply card-draw: missing game or player context")
		}
		if o.Target == "all-opponents" {
			for _, opponent := range a.game.GetAllPlayers() {
				if opponent.ID() == a.player.ID() {
					continue
				}
				drawnCards, err := a.game.Deck().DrawProjectCards(ctx, amount)
				if err != nil {
					log.Warn("Failed to draw cards for opponent", zap.String("opponent_id", opponent.ID()), zap.Error(err))
					continue
				}
				for _, cardID := range drawnCards {
					opponent.Hand().AddCard(cardID)
				}
				log.Debug("Opponent drew cards", zap.String("opponent_id", opponent.ID()), zap.Int("amount", len(drawnCards)))
			}
		} else if HasCardSelectors(o.Selectors) && a.cardRegistry != nil {
			matcher := func(cardID string) bool {
				card, err := a.cardRegistry.GetByID(cardID)
				if err != nil || card == nil {
					return false
				}
				return MatchesAnySelector(card, o.Selectors)
			}
			matched, discarded, err := a.game.Deck().DrawProjectCardsUntilMatching(ctx, amount, matcher)
			if err != nil {
				log.Warn("Failed to draw matching cards", zap.Error(err))
				return nil
			}
			for _, cardID := range matched {
				a.player.Hand().AddCard(cardID)
			}
			if len(discarded) > 0 {
				_ = a.game.Deck().Discard(ctx, discarded)
			}
			log.Debug("Drew matching cards (draw-until)", zap.Int("matched", len(matched)), zap.Int("discarded", len(discarded)))
		} else {
			drawnCards, err := a.game.Deck().DrawProjectCards(ctx, amount)
			if err != nil {
				log.Warn("Failed to draw cards", zap.Error(err))
				return nil
			}
			for _, cardID := range drawnCards {
				a.player.Hand().AddCard(cardID)
			}
			log.Debug("Drew cards and added to hand", zap.Int("amount", len(drawnCards)))
		}

	case shared.ResourceCardDiscard:
		log.Debug("Skipping card-discard output (handled at action layer)")

	case shared.ResourceCardPeek, shared.ResourceCardTake, shared.ResourceCardBuy:
		log.Debug("Skipping card draw output (handled by ApplyCardDrawOutputs)",
			zap.String("type", string(o.ResourceType)), zap.Int("amount", amount))

	default:
		log.Warn("Unhandled card operation type", zap.String("type", string(o.ResourceType)))
	}
	return nil
}
