package action

import (
	"fmt"
	"openmars/internal/game/colony"
	"slices"
	"time"

	"openmars/internal/game"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// Canonical base costs for a single colony trade, before any action discounts.
// A trade is paid with exactly one of these resource types.
const (
	TradeCreditsCost  = 9
	TradeEnergyCost   = 3
	TradeTitaniumCost = 3
)

// tradePaymentCost pairs a payment resource with its base cost so affordability
// and effective-cost computation share a single iteration.
type tradePaymentCost struct {
	resource shared.ResourceType
	baseCost int
}

var tradePaymentCosts = []tradePaymentCost{
	{shared.ResourceCredit, TradeCreditsCost},
	{shared.ResourceEnergy, TradeEnergyCost},
	{shared.ResourceTitanium, TradeTitaniumCost},
}

// CalculateColonyTradeState is the single source of truth for colony-trade
// affordability and availability. The trade is available when the player's
// trade fleet is free, at least one colony has not yet been traded this
// generation, and the player can afford at least one payment type after
// applying action discounts (e.g. Rim Freighters).
//
// The returned EntityState carries the effective per-resource trade cost in
// Cost and any discounts in Metadata, so both the DTO mapper and the
// available-actions aggregator consume identical numbers.
func CalculateColonyTradeState(
	p *player.Player,
	g *game.Game,
	cardRegistry gamecards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]any)

	errors = append(errors, validatePhase(g)...)
	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	effectiveCosts, discounts := CalculateEffectiveTradeCosts(p, cardRegistry)
	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	if g.Colonies().TradeFleet(p.ID()).Available() == 0 {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeNoActionsRemaining,
			Category: player.ErrorCategoryAvailability,
			Message:  "Trade fleet is not available",
		})
	}

	if len(g.Colonies().GetTradeableIDs()) == 0 {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeNoTilePlacements,
			Category: player.ErrorCategoryAvailability,
			Message:  "No tradeable colonies",
		})
	}

	if !canAffordAnyTrade(p, g, cardRegistry, effectiveCosts) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInsufficientResources,
			Category: player.ErrorCategoryCost,
			Message:  "Cannot afford any trade payment type",
		})
	}

	return player.EntityState{
		Errors:         errors,
		Cost:           effectiveCosts,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// CalculateEffectiveTradeCosts returns the per-resource effective trade cost
// after applying the player's colony-trade action discounts, alongside the
// non-zero discounts that were applied. It is the shared cost computation used
// by trade execution, the DTO mapper, and CalculateColonyTradeState.
func CalculateEffectiveTradeCosts(
	p *player.Player,
	cardRegistry gamecards.CardRegistry,
) (effectiveCosts map[string]int, discounts map[string]int) {
	calc := gamecards.NewRequirementModifierCalculator(cardRegistry)
	tradeDiscounts := calc.CalculateActionDiscounts(p, shared.ActionColonyTrade)

	effectiveCosts = make(map[string]int, len(tradePaymentCosts))
	discounts = make(map[string]int)
	for _, pc := range tradePaymentCosts {
		discount := tradeDiscounts[pc.resource]
		effective := pc.baseCost - discount
		if effective < 0 {
			effective = 0
		}
		effectiveCosts[string(pc.resource)] = effective
		if discount > 0 {
			discounts[string(pc.resource)] = discount
		}
	}
	return effectiveCosts, discounts
}

// canAffordAnyTrade reports whether the player can pay at least one payment type
// at its effective cost.
func canAffordAnyTrade(p *player.Player, g *game.Game, registry gamecards.CardRegistry, costs map[string]int) bool {
	for rt, cost := range costs {
		if gamecards.PaymentCapacity(p, g, registry, shared.ActionColonyTrade, shared.ResourceType(rt)) >= cost {
			return true
		}
	}
	return false
}

// ColonyTradeOption is one legal track increase and the resulting gains for the trader.
type ColonyTradeOption struct {
	TrackSteps     int
	MarkerPosition int
	Outputs        []colony.Output
}

// CalculateColonyTradeOptions resolves contextual track effects without mutating state.
func CalculateColonyTradeOptions(p *player.Player, state *colony.ColonyState, definition *colony.ColonyDefinition, registry gamecards.CardRegistryInterface) []ColonyTradeOption {
	if p == nil || state == nil || definition == nil || state.AwaitingResource != "" || state.MarkerPosition < 0 || state.MarkerPosition >= len(definition.Steps) {
		return nil
	}
	advances := map[int]bool{0: true}
	headroom := len(definition.Steps) - 1 - state.MarkerPosition
	ids := append([]string(nil), p.PlayedCards().Cards()...)
	if corp := p.CorporationID(); corp != "" && !slices.Contains(ids, corp) {
		ids = append(ids, corp)
	}
	if registry != nil {
		for _, id := range ids {
			card, err := registry.GetByID(id)
			if err != nil {
				continue
			}
			for _, behavior := range card.Behaviors {
				matched := false
				for _, trigger := range behavior.Triggers {
					if trigger.Type == shared.TriggerTypeAuto && trigger.Condition != nil && trigger.Condition.Type == "before-colony-trade" && trigger.Condition.Target != nil && *trigger.Condition.Target == "self-player" {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
				for _, output := range behavior.Outputs {
					step, ok := output.(*shared.ColonyCondition)
					if !ok || step.ResourceType != shared.ResourceColonyTrackStep || step.Target != "trigger-colony" || step.Amount <= 0 {
						continue
					}
					next := map[int]bool{}
					for amount := range advances {
						if step.Optional {
							next[amount] = true
						}
						next[min(headroom, amount+step.Amount)] = true
					}
					advances = next
				}
			}
		}
	}
	amounts := make([]int, 0, len(advances))
	for amount := range advances {
		amounts = append(amounts, amount)
	}
	slices.Sort(amounts)
	options := make([]ColonyTradeOption, 0, len(amounts))
	for _, amount := range amounts {
		position := state.MarkerPosition + amount
		outputs := append([]colony.Output(nil), definition.Steps[position].Outputs...)
		for _, owner := range state.PlayerColonies {
			if owner == p.ID() {
				outputs = append(outputs, definition.ColonyBonus...)
			}
		}
		combined := make([]colony.Output, 0, len(outputs))
		for _, output := range outputs {
			index := slices.IndexFunc(combined, func(o colony.Output) bool { return o.Type == output.Type })
			if index < 0 {
				combined = append(combined, output)
			} else {
				combined[index].Amount += output.Amount
			}
		}
		options = append(options, ColonyTradeOption{TrackSteps: amount, MarkerPosition: position, Outputs: combined})
	}
	return options
}

// ValidateColonyTradeOption rejects stale or forged increases before any trade cost is spent.
func ValidateColonyTradeOption(p *player.Player, state *colony.ColonyState, definition *colony.ColonyDefinition, registry gamecards.CardRegistryInterface, trackSteps int) (ColonyTradeOption, error) {
	for _, option := range CalculateColonyTradeOptions(p, state, definition, registry) {
		if option.TrackSteps == trackSteps {
			return option, nil
		}
	}
	return ColonyTradeOption{}, fmt.Errorf("invalid trade track increase: %d", trackSteps)
}
