package action

import (
	"fmt"
	"strings"
	"time"

	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/milestones"
)

// CalculatePlayerCardState computes playability state for a card.
// This function can access both Game and Player without circular dependencies.
// card parameter must be *gamecards.Card
func CalculatePlayerCardState(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
	colonyBonusLookup ...gamecards.ColonyBonusLookup,
) player.EntityState {
	var errors []player.StateError
	var warnings []player.StateWarning
	metadata := make(map[string]interface{})

	errors = append(errors, validatePhase(g)...)
	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	costMap, discounts := calculateEffectiveCost(card, p, cardRegistry)
	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	errors = append(errors, validateAffordabilityWithSubstitutes(p, card, costMap)...)
	errors = append(errors, validateRequirements(card, p, g, cardRegistry)...)
	errors = append(errors, validateProductionOutputs(card, p)...)
	errors = append(errors, validateCardResourceOutputs(card, p, cardRegistry)...)
	errors = append(errors, validateCardDiscardOutputs(card, p)...)
	errors = append(errors, validateNegativeResourceOutputsForCard(card, p)...)
	errors = append(errors, ValidateTileOutputs(card, p, g)...)

	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}
		warnings = append(warnings, validateGlobalParamWarnings(behavior.Outputs, g)...)
		for _, choice := range behavior.Choices {
			warnings = append(warnings, validateGlobalParamWarnings(choice.Outputs, g)...)
		}
	}

	var lookup gamecards.ColonyBonusLookup
	if len(colonyBonusLookup) > 0 {
		lookup = colonyBonusLookup[0]
	}
	warnings = append(warnings, validateColonyBonusStorageTargets(card, p, g, cardRegistry, lookup)...)
	computedValues := computeBehaviorValues(card.Behaviors, "", p, g, cardRegistry, lookup)

	return player.EntityState{
		Errors:         errors,
		Warnings:       warnings,
		Cost:           costMap,
		Metadata:       metadata,
		ComputedValues: computedValues,
		LastCalculated: time.Now(),
	}
}

// CalculatePendingCardPlayability computes playability state for a pending card
// (during card selection/buying). Skips phase, turn, and tile-selection checks
// since those are irrelevant during card selection.
func CalculatePendingCardPlayability(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	var warnings []player.StateWarning
	metadata := make(map[string]interface{})

	costMap, discounts := calculateEffectiveCost(card, p, cardRegistry)
	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	errors = append(errors, validateAffordabilityWithSubstitutes(p, card, costMap)...)
	errors = append(errors, validateRequirements(card, p, g, cardRegistry)...)
	errors = append(errors, validateProductionOutputs(card, p)...)
	errors = append(errors, validateCardResourceOutputs(card, p, cardRegistry)...)
	errors = append(errors, validateCardDiscardOutputs(card, p)...)
	errors = append(errors, validateNegativeResourceOutputsForCard(card, p)...)
	errors = append(errors, ValidateTileOutputs(card, p, g)...)

	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}
		warnings = append(warnings, validateGlobalParamWarnings(behavior.Outputs, g)...)
		for _, choice := range behavior.Choices {
			warnings = append(warnings, validateGlobalParamWarnings(choice.Outputs, g)...)
		}
	}

	return player.EntityState{
		Errors:         errors,
		Warnings:       warnings,
		Cost:           costMap,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// CalculatePlayerCardActionState computes usability state for a card action.
func CalculatePlayerCardActionState(
	cardID string,
	behavior shared.CardBehavior,
	timesUsedThisGeneration int,
	p *player.Player,
	g *game.Game,
	cardRegistry ...cards.CardRegistry,
) player.EntityState {
	var errors []player.StateError

	currentTurn := g.CurrentTurn()
	if currentTurn != nil && currentTurn.PlayerID() != p.ID() {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeNotYourTurn,
			Category: player.ErrorCategoryTurn,
			Message:  "Not your turn",
		})
	}

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	resources := p.Resources().Get()
	for _, inputBC := range behavior.Inputs {
		// Skip variable-amount inputs — the player selects how much to spend (can be 0)
		if shared.IsVariableAmount(inputBC) {
			continue
		}

		rt := inputBC.GetResourceType()
		amt := inputBC.GetAmount()
		target := inputBC.GetTarget()

		// Storage resource inputs (target: "self-card") check card storage instead of player resources
		if target == "self-card" && gamecards.IsStorageResourceType(rt) {
			storage := p.Resources().GetCardStorage(cardID)
			if storage < amt {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientResources,
					Category: player.ErrorCategoryInput,
					Message:  fmt.Sprintf("Not enough %s on card", rt),
				})
			}
			continue
		}

		// Credit inputs with paymentAllowed consider alternative resources (e.g., titanium)
		paymentAllowed := shared.GetPaymentAllowed(inputBC)
		if rt == shared.ResourceCredit && len(paymentAllowed) > 0 {
			effectiveCredits := resources.Credits
			substitutes := p.Resources().PaymentSubstitutes()
			for _, allowed := range paymentAllowed {
				for _, sub := range substitutes {
					if sub.ResourceType == allowed {
						available := resources.GetAmount(allowed)
						effectiveCredits += available * sub.ConversionRate
					}
				}
			}
			if effectiveCredits < amt {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientResources,
					Category: player.ErrorCategoryInput,
					Message:  fmt.Sprintf("Not enough %s", rt),
				})
			}
			continue
		}

		// Production inputs check player production instead of basic resources
		if shared.IsProductionResourceType(rt) {
			available := p.Resources().Production().GetAmount(rt)
			if available < amt {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientResources,
					Category: player.ErrorCategoryInput,
					Message:  fmt.Sprintf("Not enough %s", rt),
				})
			}
			continue
		}

		available := resources.GetAmount(rt)
		if available < amt {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryInput,
				Message:  fmt.Sprintf("Not enough %s", rt),
			})
		}
	}

	for _, outputBC := range behavior.Outputs {
		if outputBC.GetTarget() == "steal-from-any-card" {
			totalAvailable := 0
			var reg cards.CardRegistry
			if len(cardRegistry) > 0 {
				reg = cardRegistry[0]
			}
			for _, anyPlayer := range g.GetAllPlayers() {
				for _, playerCardID := range anyPlayer.PlayedCards().Cards() {
					if playerCardID == cardID && anyPlayer.ID() == p.ID() {
						continue
					}
					storage := anyPlayer.Resources().GetCardStorage(playerCardID)
					if storage <= 0 {
						continue
					}
					if reg != nil {
						registryCard, err := reg.GetByID(playerCardID)
						if err != nil || registryCard.ResourceStorage == nil {
							continue
						}
						if registryCard.ResourceStorage.Type != outputBC.GetResourceType() {
							continue
						}
					}
					totalAvailable += storage
				}
			}
			if totalAvailable < outputBC.GetAmount() {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientResources,
					Category: player.ErrorCategoryInput,
					Message:  fmt.Sprintf("No %s available on any card", outputBC.GetResourceType()),
				})
			}
		}
	}

	errors = append(errors, validateActionUsageLimit(behavior, timesUsedThisGeneration)...)
	errors = append(errors, validateActionReuseAvailability(cardID, behavior, p)...)
	errors = append(errors, validateBehaviorTileOutputs(behavior, p, g)...)
	errors = append(errors, validateGenerationalEventRequirements(behavior, p)...)
	errors = append(errors, validateNegativeResourceOutputs(behavior, p)...)

	var warnings []player.StateWarning
	warnings = append(warnings, validateGlobalParamWarnings(behavior.Outputs, g)...)
	for _, choice := range behavior.Choices {
		warnings = append(warnings, validateGlobalParamWarnings(choice.Outputs, g)...)
	}

	var reg cards.CardRegistry
	if len(cardRegistry) > 0 {
		reg = cardRegistry[0]
	}
	computedValues := computeBehaviorValues([]shared.CardBehavior{behavior}, cardID, p, g, reg, nil)

	return player.EntityState{
		Errors:         errors,
		Warnings:       warnings,
		Cost:           make(map[string]int),
		Metadata:       make(map[string]any),
		ComputedValues: computedValues,
		LastCalculated: time.Now(),
	}
}

// CalculatePlayerStandardProjectState computes availability state for a standard project.
func CalculatePlayerStandardProjectState(
	projectType shared.StandardProject,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	var warnings []player.StateWarning
	metadata := make(map[string]interface{})

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	baseCosts := getStandardProjectBaseCosts(projectType)
	if baseCosts == nil {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInvalidProjectType,
			Category: player.ErrorCategoryConfiguration,
			Message:  fmt.Sprintf("Unknown standard project type: %s", projectType),
		})
		return player.EntityState{
			Errors:         errors,
			Cost:           make(map[string]int),
			Metadata:       metadata,
			LastCalculated: time.Now(),
		}
	}

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	projectDiscounts := calculator.CalculateStandardProjectDiscounts(p, projectType)

	effectiveCosts := make(map[string]int)
	discounts := make(map[string]int)
	for resourceType, amount := range baseCosts {
		discount := projectDiscounts[shared.ResourceType(resourceType)]
		effectiveCost := amount - discount
		if effectiveCost < 0 {
			effectiveCost = 0
		}
		effectiveCosts[resourceType] = effectiveCost
		if discount > 0 {
			discounts[resourceType] = discount
		}
	}

	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	errors = append(errors, validateAffordabilityMap(p, effectiveCosts)...)

	switch projectType {
	case shared.StandardProjectSellPatents:
		cardCount := p.Hand().CardCount()
		if cardCount == 0 {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeNoCardsInHand,
				Category: player.ErrorCategoryAvailability,
				Message:  "No cards in hand to sell",
			})
		}

	case shared.StandardProjectAquifer:
		currentOceans := g.GlobalParameters().Oceans()
		oceansRemaining := 9 - currentOceans
		metadata["oceansRemaining"] = oceansRemaining
		if oceansRemaining <= 0 {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeNoOceanTiles,
				Category: player.ErrorCategoryAvailability,
				Message:  "No ocean tiles remaining",
			})
		}

	case shared.StandardProjectAsteroid:
		if g.GlobalParameters().Temperature() >= global_parameters.MaxTemperature {
			warnings = append(warnings, player.StateWarning{
				Code:    player.WarningCodeGlobalParamMaxed,
				Message: "Temperature is already at maximum",
			})
		}

	case shared.StandardProjectCity:
		cityPlacements := g.CountAvailableHexesForTile("city", p.ID(), nil)
		if cityPlacements == 0 {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeNoCityPlacements,
				Category: player.ErrorCategoryAvailability,
				Message:  "No valid city placements",
			})
		}

	case shared.StandardProjectGreenery:
		greeneryPlacements := g.CountAvailableHexesForTile("greenery", p.ID(), nil)
		if greeneryPlacements == 0 {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeNoGreeneryPlacements,
				Category: player.ErrorCategoryAvailability,
				Message:  "No valid greenery placements",
			})
		}
		if g.GlobalParameters().Oxygen() >= global_parameters.MaxOxygen {
			warnings = append(warnings, player.StateWarning{
				Code:    player.WarningCodeGlobalParamMaxed,
				Message: "Oxygen is already at maximum",
			})
		}

	case shared.StandardProjectPowerPlant:

	case shared.StandardProjectAirScrapping:
		if g.GlobalParameters().Venus() >= global_parameters.MaxVenus {
			warnings = append(warnings, player.StateWarning{
				Code:    player.WarningCodeGlobalParamMaxed,
				Message: "Venus is already at maximum",
			})
		}

	case shared.StandardProjectConvertHeatToTemperature:
		if g.GlobalParameters().Temperature() >= global_parameters.MaxTemperature {
			warnings = append(warnings, player.StateWarning{
				Code:    player.WarningCodeGlobalParamMaxed,
				Message: "Temperature already at maximum",
			})
		}

	case shared.StandardProjectConvertPlantsToGreenery:
		if g.GlobalParameters().Oxygen() >= global_parameters.MaxOxygen {
			warnings = append(warnings, player.StateWarning{
				Code:    player.WarningCodeGlobalParamMaxed,
				Message: "Oxygen already at maximum",
			})
		}

	default:
	}

	return player.EntityState{
		Errors:         errors,
		Warnings:       warnings,
		Cost:           effectiveCosts,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// validateActionsRemaining checks if the player has actions remaining in their turn.
func validateActionsRemaining(p *player.Player, g *game.Game) []player.StateError {
	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		return nil
	}
	if currentTurn.PlayerID() != p.ID() {
		return nil
	}
	remaining := currentTurn.ActionsRemaining()
	if remaining == 0 {
		return []player.StateError{{
			Code:     player.ErrorCodeNoActionsRemaining,
			Category: player.ErrorCategoryTurn,
			Message:  "No actions remaining",
		}}
	}
	return nil
}

// validatePhase checks if action is allowed in current phase.
func validatePhase(g *game.Game) []player.StateError {
	if g.CurrentPhase() != shared.GamePhaseAction {
		return []player.StateError{{
			Code:     player.ErrorCodeWrongPhase,
			Category: player.ErrorCategoryPhase,
			Message:  "Not in action phase",
		}}
	}
	return nil
}

// validateNoPendingSelection checks if player has any pending selection that blocks actions.
func validateNoPendingSelection(p *player.Player, g *game.Game) []player.StateError {
	if g.HasAnyPendingSelection(p.ID()) {
		return []player.StateError{{
			Code:     player.ErrorCodeActiveTileSelection,
			Category: player.ErrorCategoryPhase,
			Message:  "Pending selection",
		}}
	}
	return nil
}

// validateActionUsageLimit checks if the action has already been used this generation.
// Manual trigger actions can only be used once per generation by default.
func validateActionUsageLimit(
	behavior shared.CardBehavior,
	timesUsedThisGeneration int,
) []player.StateError {
	hasManualTrigger := false
	for _, trigger := range behavior.Triggers {
		if trigger.Type == shared.TriggerTypeManual {
			hasManualTrigger = true
			break
		}
	}

	if !hasManualTrigger {
		return nil
	}

	if timesUsedThisGeneration >= 1 {
		return []player.StateError{{
			Code:     player.ErrorCodeActionAlreadyPlayed,
			Category: player.ErrorCategoryAvailability,
			Message:  "Already played",
		}}
	}

	return nil
}

func validateActionReuseAvailability(
	cardID string,
	behavior shared.CardBehavior,
	p *player.Player,
) []player.StateError {
	hasActionReuse := false
	for _, output := range behavior.Outputs {
		if output.GetResourceType() == shared.ResourceActionReuse {
			hasActionReuse = true
			break
		}
	}
	if !hasActionReuse {
		return nil
	}

	for _, act := range p.Actions().List() {
		if act.CardID == cardID {
			continue
		}
		hasManual := false
		for _, trigger := range act.Behavior.Triggers {
			if trigger.Type == shared.TriggerTypeManual {
				hasManual = true
				break
			}
		}
		if hasManual && act.TimesUsedThisGeneration >= 1 {
			return nil
		}
	}

	return []player.StateError{{
		Code:     player.ErrorCodeNoUsedActions,
		Category: player.ErrorCategoryAvailability,
		Message:  "No used actions to reuse",
	}}
}

// validateRequirements checks all card requirements.
// Includes global parameter lenience from temporary effects (e.g., Special Design).
func validateRequirements(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) []player.StateError {
	if card.Requirements == nil || len(card.Requirements.Items) == 0 {
		return nil
	}

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	if calculator.HasIgnoreGlobalRequirements(p) {
		return nil
	}

	var errors []player.StateError

	for _, req := range card.Requirements.Items {
		err := checkRequirement(req, p, g, cardRegistry, calculator)
		if err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateProductionOutputs checks that player has enough production for negative production outputs.
// Cards like "Urbanized Area" have negative production outputs (e.g., -1 energy production).
// The player must have at least that much production to play the card.
func validateProductionOutputs(
	card *gamecards.Card,
	p *player.Player,
) []player.StateError {
	if len(card.Behaviors) == 0 {
		return nil
	}

	var errors []player.StateError
	production := p.Resources().Production()

	// Check all behaviors for auto-triggers with negative production outputs
	for _, behavior := range card.Behaviors {
		// Only check auto-trigger behaviors (immediate effects when card is played)
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		// Check outputs for negative production
		for _, outputBC := range behavior.Outputs {
			// Skip variable-amount outputs — the player controls the amount
			if shared.IsVariableAmount(outputBC) {
				continue
			}
			// Only check production resource types with negative amounts
			if outputBC.GetAmount() >= 0 {
				continue
			}

			// Map production resource types to base resource types for checking
			var baseResourceType shared.ResourceType
			switch outputBC.GetResourceType() {
			case shared.ResourceCreditProduction:
				baseResourceType = shared.ResourceCredit
			case shared.ResourceSteelProduction:
				baseResourceType = shared.ResourceSteel
			case shared.ResourceTitaniumProduction:
				baseResourceType = shared.ResourceTitanium
			case shared.ResourcePlantProduction:
				baseResourceType = shared.ResourcePlant
			case shared.ResourceEnergyProduction:
				baseResourceType = shared.ResourceEnergy
			case shared.ResourceHeatProduction:
				baseResourceType = shared.ResourceHeat
			default:
				// Not a production type, skip
				continue
			}

			// Check if player has enough production
			currentProduction := production.GetAmount(baseResourceType)
			resultingProduction := currentProduction + outputBC.GetAmount()

			// MC production can go to -5, others cannot go below 0
			var minProduction int
			if baseResourceType == shared.ResourceCredit {
				minProduction = shared.MinCreditProduction
			} else {
				minProduction = shared.MinOtherProduction
			}

			if resultingProduction < minProduction {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientProduction,
					Category: player.ErrorCategoryRequirement,
					Message:  formatInsufficientProductionMessage(baseResourceType),
				})
			}
		}
	}

	return errors
}

// isBasicPlayerResource returns true for the 6 basic player resource types.
func isBasicPlayerResource(rt shared.ResourceType) bool {
	switch rt {
	case shared.ResourceCredit, shared.ResourceSteel, shared.ResourceTitanium,
		shared.ResourcePlant, shared.ResourceEnergy, shared.ResourceHeat:
		return true
	}
	return false
}

// validateNegativeResourceOutputsForCard checks all auto-trigger behaviors on a card
// for negative resource outputs (e.g., "spend 5 heat" modeled as heat: -5).
func validateNegativeResourceOutputsForCard(
	card *gamecards.Card,
	p *player.Player,
) []player.StateError {
	var errors []player.StateError
	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}
		errors = append(errors, validateNegativeResourceOutputs(behavior, p)...)
	}
	return errors
}

// validateNegativeResourceOutputs checks that the player has enough resources for negative
// resource outputs in a card action behavior's base outputs.
// This is a defense-in-depth check: card costs should normally be modeled as inputs,
// but this catches any negative resource outputs that slip through.
func validateNegativeResourceOutputs(
	behavior shared.CardBehavior,
	p *player.Player,
) []player.StateError {
	var errors []player.StateError
	resources := p.Resources().Get()

	for _, outputBC := range behavior.Outputs {
		if shared.IsVariableAmount(outputBC) || outputBC.GetAmount() >= 0 {
			continue
		}
		if outputBC.GetTarget() != "" && outputBC.GetTarget() != "self-player" {
			continue
		}
		if !isBasicPlayerResource(outputBC.GetResourceType()) {
			continue
		}
		available := resources.GetAmount(outputBC.GetResourceType())
		required := -outputBC.GetAmount()
		if available < required {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryInput,
				Message:  fmt.Sprintf("Not enough %s", outputBC.GetResourceType()),
			})
		}
	}
	return errors
}

// validateCardResourceOutputs checks that the player has at least one valid target card
// with resource storage for card-resource outputs with any-card target.
func validateCardResourceOutputs(
	card *gamecards.Card,
	p *player.Player,
	cardRegistry cards.CardRegistry,
) []player.StateError {
	if len(card.Behaviors) == 0 || cardRegistry == nil {
		return nil
	}

	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		// Check both direct outputs and choice outputs
		allOutputs := behavior.Outputs
		for _, choice := range behavior.Choices {
			allOutputs = append(allOutputs, choice.Outputs...)
		}

		for _, outputBC := range allOutputs {
			if outputBC.GetResourceType() != shared.ResourceCardResource || outputBC.GetTarget() != "any-card" {
				continue
			}

			selectors := shared.GetSelectors(outputBC)

			// Check if player has any played card with resource storage matching selectors
			hasValidTarget := false
			for _, playedCardID := range p.PlayedCards().Cards() {
				playedCard, err := cardRegistry.GetByID(playedCardID)
				if err != nil {
					continue
				}
				if playedCard.ResourceStorage == nil {
					continue
				}
				// If selectors specified, card must match
				if len(selectors) > 0 && !gamecards.MatchesAnySelector(playedCard, selectors) {
					continue
				}
				hasValidTarget = true
				break
			}

			if !hasValidTarget {
				return []player.StateError{{
					Code:     player.ErrorCodeInsufficientResources,
					Category: player.ErrorCategoryAvailability,
					Message:  "No valid card with resource storage",
				}}
			}
		}
	}

	return nil
}

// validateColonyBonusStorageTargets warns when colony bonuses include card-targeted resources
// but the player has no valid card to store them on.
func validateColonyBonusStorageTargets(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
	colonyBonusLookup gamecards.ColonyBonusLookup,
) []player.StateWarning {
	if colonyBonusLookup == nil || g == nil || !g.HasColonies() {
		return nil
	}

	hasColonyBonusOutput := false
	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}
		for _, output := range behavior.Outputs {
			if output.GetResourceType() == shared.ResourceColonyBonus {
				hasColonyBonusOutput = true
				break
			}
		}
	}
	if !hasColonyBonusOutput {
		return nil
	}

	bonuses := gamecards.CollectColonyBonuses(p.ID(), g.Colonies().States(), colonyBonusLookup)

	hasReward := false
	for _, b := range bonuses {
		if b.Amount > 0 {
			hasReward = true
			break
		}
	}
	if !hasReward {
		return []player.StateWarning{{
			Code:    "no-colony-bonus-reward",
			Message: "No reward",
		}}
	}

	var warnings []player.StateWarning
	seen := map[string]bool{}
	for _, b := range bonuses {
		rt := shared.ResourceType(b.ResourceType)
		if !gamecards.IsStorageResourceType(rt) || seen[b.ResourceType] {
			continue
		}
		seen[b.ResourceType] = true
		if !gamecards.HasEligibleStorageCard(p, rt, cardRegistry) {
			warnings = append(warnings, player.StateWarning{
				Code:    "no-storage-for-colony-bonus",
				Message: fmt.Sprintf("No %s storage", b.ResourceType),
			})
		}
	}
	return warnings
}

// validateCardDiscardOutputs checks that the player has enough cards in hand to satisfy
// card-discard outputs when playing a card.
func validateCardDiscardOutputs(
	card *gamecards.Card,
	p *player.Player,
) []player.StateError {
	if len(card.Behaviors) == 0 {
		return nil
	}

	totalDiscard := 0
	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		for _, output := range behavior.Outputs {
			if output.GetResourceType() == shared.ResourceCardDiscard && output.GetTarget() == "self-player" {
				totalDiscard += output.GetAmount()
			}
		}

		for _, choice := range behavior.Choices {
			for _, output := range choice.Outputs {
				if output.GetResourceType() == shared.ResourceCardDiscard && output.GetTarget() == "self-player" {
					totalDiscard += output.GetAmount()
				}
			}
		}
	}

	if totalDiscard == 0 {
		return nil
	}

	// The card being played leaves the hand first, so available cards = hand - 1
	availableCards := p.Hand().CardCount() - 1
	if availableCards < totalDiscard {
		return []player.StateError{{
			Code:     player.ErrorCodeNoCardsInHand,
			Category: player.ErrorCategoryAvailability,
			Message:  "Not enough cards to discard",
		}}
	}

	return nil
}

// ValidateTileOutputs checks that the board has available placements for any tile outputs.
// If a card outputs city/greenery/ocean tiles, the player must have valid placement locations.
func ValidateTileOutputs(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
) []player.StateError {
	if len(card.Behaviors) == 0 || g == nil {
		return nil
	}

	var errors []player.StateError

	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		for _, outputBC := range behavior.Outputs {
			switch outputBC.GetResourceType() {
			case shared.ResourceCityPlacement:
				cityPlacements := g.CountAvailableHexesForTile("city", p.ID(), shared.GetTileRestrictions(outputBC))
				if cityPlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoCityPlacements,
						Category: player.ErrorCategoryAvailability,
						Message:  "No valid city placements",
					})
				}

			case shared.ResourceGreeneryPlacement:
				greeneryPlacements := g.CountAvailableHexesForTile("greenery", p.ID(), shared.GetTileRestrictions(outputBC))
				if greeneryPlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoGreeneryPlacements,
						Category: player.ErrorCategoryAvailability,
						Message:  "No valid greenery placements",
					})
				}

			case shared.ResourceOceanPlacement:
				oceanPlacements := g.CountAvailableHexesForTile("ocean", p.ID(), nil)
				if oceanPlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoOceanTiles,
						Category: player.ErrorCategoryAvailability,
						Message:  "No ocean tiles remaining",
					})
				}

			case shared.ResourceVolcanoPlacement:
				volcanoPlacements := g.CountAvailableHexesForTile("volcano", p.ID(), nil)
				if volcanoPlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoTilePlacements,
						Category: player.ErrorCategoryAvailability,
						Message:  "No valid volcanic placements",
					})
				}

			case shared.ResourceTileReplacement:
				var tileType string
				if tc, ok := outputBC.(*shared.TileModificationCondition); ok {
					tileType = tc.TileType
				}
				if tileType == "" {
					continue
				}
				replacementPlacements := g.CountAvailableHexesForTile("tile-replacement:"+tileType, p.ID(), nil)
				if replacementPlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoTilePlacements,
						Category: player.ErrorCategoryAvailability,
						Message:  "No valid tile placements",
					})
				}
			case shared.ResourceTilePlacement:
				var tileType string
				var tileRestrictions *shared.TileRestrictions
				if tc, ok := outputBC.(*shared.TilePlacementCondition); ok {
					tileType = tc.TileType
					tileRestrictions = tc.TileRestrictions
				}
				if tileType == "" {
					continue
				}
				tilePlacements := g.CountAvailableHexesForTile(tileType, p.ID(), tileRestrictions)
				if tilePlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoTilePlacements,
						Category: player.ErrorCategoryAvailability,
						Message:  "No valid tile placements",
					})
				}
			}
		}
	}

	return errors
}

func validateGenerationalEventRequirements(
	behavior shared.CardBehavior,
	p *player.Player,
) []player.StateError {
	if len(behavior.GenerationalEventRequirements) == 0 {
		return nil
	}

	var errors []player.StateError
	playerEvents := p.GenerationalEvents()

	for _, req := range behavior.GenerationalEventRequirements {
		count := playerEvents.GetCount(req.Event)

		if req.Count != nil {
			if req.Count.Min != nil && count < *req.Count.Min {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeGenerationalEventNotMet,
					Category: player.ErrorCategoryRequirement,
					Message:  formatGenerationalEventError(req.Event),
				})
			}
			if req.Count.Max != nil && count > *req.Count.Max {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeGenerationalEventNotMet,
					Category: player.ErrorCategoryRequirement,
					Message:  formatGenerationalEventError(req.Event),
				})
			}
		}
	}

	return errors
}

func formatGenerationalEventError(event shared.GenerationalEvent) string {
	switch event {
	case shared.GenerationalEventTRRaise:
		return "TR not raised this generation"
	case shared.GenerationalEventOceanPlacement:
		return "Ocean not placed this generation"
	case shared.GenerationalEventCityPlacement:
		return "City not placed this generation"
	case shared.GenerationalEventGreeneryPlacement:
		return "Greenery not placed this generation"
	default:
		return "Generational event requirement not met"
	}
}

// validateBehaviorTileOutputs checks tile availability for a single behavior's outputs.
// Used by card action state calculation.
func validateBehaviorTileOutputs(
	behavior shared.CardBehavior,
	p *player.Player,
	g *game.Game,
) []player.StateError {
	if g == nil {
		return nil
	}

	var errors []player.StateError

	// Check outputs for tile placements
	for _, outputBC := range behavior.Outputs {
		switch outputBC.GetResourceType() {
		case shared.ResourceCityPlacement:
			cityPlacements := g.CountAvailableHexesForTile("city", p.ID(), nil)
			if cityPlacements == 0 {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeNoCityPlacements,
					Category: player.ErrorCategoryAvailability,
					Message:  "No valid city placements",
				})
			}
		case shared.ResourceGreeneryPlacement:
			greeneryPlacements := g.CountAvailableHexesForTile("greenery", p.ID(), nil)
			if greeneryPlacements == 0 {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeNoGreeneryPlacements,
					Category: player.ErrorCategoryAvailability,
					Message:  "No valid greenery placements",
				})
			}
		case shared.ResourceOceanPlacement:
			oceanPlacements := g.CountAvailableHexesForTile("ocean", p.ID(), nil)
			if oceanPlacements == 0 {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeNoOceanTiles,
					Category: player.ErrorCategoryAvailability,
					Message:  "No ocean tiles remaining",
				})
			}
		case shared.ResourceTilePlacement:
			if tc, ok := outputBC.(*shared.TilePlacementCondition); ok {
				if tc.TileType == "" {
					continue
				}
				tilePlacements := g.CountAvailableHexesForTile(tc.TileType, p.ID(), tc.TileRestrictions)
				if tilePlacements == 0 {
					errors = append(errors, player.StateError{
						Code:     player.ErrorCodeNoTilePlacements,
						Category: player.ErrorCategoryAvailability,
						Message:  "No valid tile placements",
					})
				}
			}
		}
	}

	return errors
}

// checkRequirement validates a single requirement.
// Uses calculator to compute per-parameter lenience from player effects.
func checkRequirement(
	req gamecards.Requirement,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
	calculator *gamecards.RequirementModifierCalculator,
) *player.StateError {
	switch req.Type {
	case gamecards.RequirementTemperature:
		lenience := calculator.CalculateGlobalParameterLenience(p, "temperature")
		temp := g.GlobalParameters().Temperature()
		if req.Min != nil && temp < *req.Min-lenience {
			return &player.StateError{
				Code:     player.ErrorCodeTemperatureTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Temperature too low",
			}
		}
		if req.Max != nil && temp > *req.Max+lenience {
			return &player.StateError{
				Code:     player.ErrorCodeTemperatureTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Temperature too high",
			}
		}

	case gamecards.RequirementOxygen:
		lenience := calculator.CalculateGlobalParameterLenience(p, "oxygen")
		oxygen := g.GlobalParameters().Oxygen()
		if req.Min != nil && oxygen < *req.Min-lenience {
			return &player.StateError{
				Code:     player.ErrorCodeOxygenTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Oxygen too low",
			}
		}
		if req.Max != nil && oxygen > *req.Max+lenience {
			return &player.StateError{
				Code:     player.ErrorCodeOxygenTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Oxygen too high",
			}
		}

	case gamecards.RequirementOceans:
		lenience := calculator.CalculateGlobalParameterLenience(p, "ocean")
		oceans := g.GlobalParameters().Oceans()
		if req.Min != nil && oceans < *req.Min-lenience {
			return &player.StateError{
				Code:     player.ErrorCodeOceansTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Too few oceans",
			}
		}
		if req.Max != nil && oceans > *req.Max+lenience {
			return &player.StateError{
				Code:     player.ErrorCodeOceansTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Too many oceans",
			}
		}

	case gamecards.RequirementTR:
		tr := p.Resources().TerraformRating()
		if req.Min != nil && tr < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeTRTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "TR too low",
			}
		}
		if req.Max != nil && tr > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeTRTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "TR too high",
			}
		}

	case gamecards.RequirementTags:
		if req.Tag == nil {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid tag requirement",
			}
		}

		tagCount := 0
		if cardRegistry != nil {
			tagCount = gamecards.CountPlayerTagsByType(p, cardRegistry, *req.Tag)
		}

		if req.Min != nil && tagCount < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeInsufficientTags,
				Category: player.ErrorCategoryRequirement,
				Message:  formatInsufficientTagsMessage(string(*req.Tag)),
			}
		}
		if req.Max != nil && tagCount > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeTooManyTags,
				Category: player.ErrorCategoryRequirement,
				Message:  formatTooManyTagsMessage(string(*req.Tag)),
			}
		}

	case gamecards.RequirementProduction:
		if req.Resource == nil {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid production requirement",
			}
		}

		// Validate resource type is producible (only 6 basic resources have production)
		if !isProducibleResource(*req.Resource) {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid production type",
			}
		}

		production := p.Resources().Production()
		currentProduction := production.GetAmount(*req.Resource)

		if req.Min != nil && currentProduction < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeInsufficientProduction,
				Category: player.ErrorCategoryRequirement,
				Message:  formatInsufficientProductionMessage(*req.Resource),
			}
		}
		// Note: No max check needed for production requirements in base game

	case gamecards.RequirementResource:
		if req.Resource == nil {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid resource requirement",
			}
		}
		resources := p.Resources().Get()
		currentAmount := resources.GetAmount(*req.Resource)

		if req.Min != nil && currentAmount < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryRequirement,
				Message:  formatInsufficientResourceMessage(*req.Resource),
			}
		}
		if req.Max != nil && currentAmount > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeTooManyResources,
				Category: player.ErrorCategoryRequirement,
				Message:  formatTooMuchResourceMessage(*req.Resource),
			}
		}

	case gamecards.RequirementCities, gamecards.RequirementGreeneries:
		// TODO: Implement tile-based requirements when Board tile counting is ready
		// For now, skip these validations (same as PlayCardAction line 310-312)

	case gamecards.RequirementColony:
		colonyCount := g.Colonies().CountPlayerColonies(p.ID())
		if req.Min != nil && colonyCount < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeColoniesTooFew,
				Category: player.ErrorCategoryRequirement,
				Message:  "Not enough colonies",
			}
		}
		if req.Max != nil && colonyCount > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeColoniesTooMany,
				Category: player.ErrorCategoryRequirement,
				Message:  "Too many colonies",
			}
		}

	case gamecards.RequirementVenus:
		lenience := calculator.CalculateGlobalParameterLenience(p, "venus")
		venus := g.GlobalParameters().Venus()
		if req.Min != nil && venus < *req.Min-lenience {
			return &player.StateError{
				Code:     player.ErrorCodeVenusTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Venus too low",
			}
		}
		if req.Max != nil && venus > *req.Max+lenience {
			return &player.StateError{
				Code:     player.ErrorCodeVenusTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Venus too high",
			}
		}
	}

	return nil
}

// calculateEffectiveCost computes cost with discounts using RequirementModifierCalculator.
// Returns the effective cost map (resource type -> amount) and discounts map (resource type -> discount amount).
// Cards typically only cost credits, so the map will usually just have {"credits": X}.
func calculateEffectiveCost(card *gamecards.Card, p *player.Player, cardRegistry cards.CardRegistry) (map[string]int, map[string]int) {
	baseCost := card.Cost

	// Calculate discounts using RequirementModifierCalculator
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	discountAmount := calculator.CalculateCardDiscounts(p, card)

	effectiveCost := baseCost - discountAmount
	if effectiveCost < 0 {
		effectiveCost = 0
	}

	// Build cost map (cards cost credits only)
	costMap := make(map[string]int)
	if effectiveCost > 0 {
		costMap[string(shared.ResourceCredit)] = effectiveCost
	}

	// Build discounts map for metadata
	discounts := make(map[string]int)
	if discountAmount > 0 {
		discounts[string(shared.ResourceCredit)] = discountAmount
	}

	return costMap, discounts
}

// validateAffordability checks if player can afford the cost (single int, for credits).
func validateAffordability(p *player.Player, cost int) []player.StateError {
	credits := p.Resources().Get().Credits
	if credits < cost {
		return []player.StateError{{
			Code:     player.ErrorCodeInsufficientCredits,
			Category: player.ErrorCategoryCost,
			Message:  "Cannot afford",
		}}
	}
	return nil
}

// validateAffordabilityMap checks if player can afford a multi-resource cost.
// Note: This function does NOT consider payment substitutes. Use validateAffordabilityWithSubstitutes for card costs.
func validateAffordabilityMap(p *player.Player, costMap map[string]int) []player.StateError {
	var errors []player.StateError
	resources := p.Resources().Get()

	for resourceType, cost := range costMap {
		var available int
		var errorCode player.StateErrorCode
		var errorMessage string

		switch shared.ResourceType(resourceType) {
		case shared.ResourceCredit:
			available = resources.Credits
			errorCode = player.ErrorCodeInsufficientCredits
			errorMessage = "Cannot afford"
		case shared.ResourceSteel:
			available = resources.Steel
			errorCode = player.ErrorCodeInsufficientResources
			errorMessage = "Not enough steel"
		case shared.ResourceTitanium:
			available = resources.Titanium
			errorCode = player.ErrorCodeInsufficientResources
			errorMessage = "Not enough titanium"
		case shared.ResourcePlant:
			available = resources.Plants
			errorCode = player.ErrorCodeInsufficientResources
			errorMessage = "Not enough plants"
		case shared.ResourceEnergy:
			available = resources.Energy
			errorCode = player.ErrorCodeInsufficientResources
			errorMessage = "Not enough energy"
		case shared.ResourceHeat:
			available = resources.Heat
			errorCode = player.ErrorCodeInsufficientResources
			errorMessage = "Not enough heat"
		default:
			available = 0
			errorCode = player.ErrorCodeInsufficientResources
			errorMessage = fmt.Sprintf("Not enough %s", resourceType)
		}

		if available < cost {
			errors = append(errors, player.StateError{
				Code:     errorCode,
				Category: player.ErrorCategoryCost,
				Message:  errorMessage,
			})
		}
	}
	return errors
}

// validateAffordabilityWithSubstitutes checks if player can afford a multi-resource cost,
// considering payment substitutes like Helion's heat-to-credit conversion and
// storage payment substitutes like Dirigibles' floaters.
// Steel is only counted for cards with the Building tag, titanium only for Space tag.
func validateAffordabilityWithSubstitutes(p *player.Player, card *gamecards.Card, costMap map[string]int) []player.StateError {
	var errors []player.StateError
	resources := p.Resources().Get()
	substitutes := p.Resources().PaymentSubstitutes()

	allowSteel := hasCardTag(card.Tags, shared.TagBuilding)
	allowTitanium := hasCardTag(card.Tags, shared.TagSpace)

	for resourceType, cost := range costMap {
		if shared.ResourceType(resourceType) == shared.ResourceCredit {
			// For credit costs, calculate effective purchasing power including substitutes
			effectiveCredits := resources.Credits

			// Add substitute resources at their conversion rates
			for _, sub := range substitutes {
				switch sub.ResourceType {
				case shared.ResourceSteel:
					if allowSteel {
						effectiveCredits += resources.Steel * sub.ConversionRate
					}
				case shared.ResourceTitanium:
					if allowTitanium {
						effectiveCredits += resources.Titanium * sub.ConversionRate
					}
				case shared.ResourceHeat:
					effectiveCredits += resources.Heat * sub.ConversionRate
				case shared.ResourceEnergy:
					effectiveCredits += resources.Energy * sub.ConversionRate
				case shared.ResourcePlant:
					effectiveCredits += resources.Plants * sub.ConversionRate
				}
			}

			// Add storage payment substitutes (e.g., Dirigibles floaters for Venus cards)
			for _, storageSub := range p.Resources().StoragePaymentSubstitutes() {
				if len(storageSub.Selectors) == 0 || gamecards.MatchesAnySelector(card, storageSub.Selectors) {
					stored := p.Resources().GetCardStorage(storageSub.CardID)
					effectiveCredits += stored * storageSub.ConversionRate
				}
			}

			if effectiveCredits < cost {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientCredits,
					Category: player.ErrorCategoryCost,
					Message:  "Cannot afford",
				})
			}
		} else {
			// Non-credit costs checked directly (no substitutes apply)
			available := resources.GetAmount(shared.ResourceType(resourceType))
			if available < cost {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientCredits,
					Category: player.ErrorCategoryCost,
					Message:  "Cannot afford",
				})
			}
		}
	}
	return errors
}

// getStandardProjectBaseCosts returns the base cost map for a standard project.
// Most projects cost credits only, but convert-plants-to-greenery costs plants
// and convert-heat-to-temperature costs heat.
func getStandardProjectBaseCosts(projectType shared.StandardProject) map[string]int {
	switch projectType {
	case shared.StandardProjectConvertPlantsToGreenery:
		return map[string]int{string(shared.ResourcePlant): 8}
	case shared.StandardProjectConvertHeatToTemperature:
		return map[string]int{string(shared.ResourceHeat): 8}
	default:
		// Credit-based projects
		cost, exists := shared.StandardProjectCost[projectType]
		if !exists {
			return nil
		}
		if cost > 0 {
			return map[string]int{string(shared.ResourceCredit): cost}
		}
		// For sell-patents (cost = 0), return empty map
		return map[string]int{}
	}
}

// GetStandardProjectBaseCosts is an exported version for use by mappers.
func GetStandardProjectBaseCosts(projectType shared.StandardProject) map[string]int {
	return getStandardProjectBaseCosts(projectType)
}

// hasCardTag checks if a tag list contains a specific tag.
func hasCardTag(tags []shared.CardTag, tag shared.CardTag) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// isProducibleResource checks if the resource type is one of the 6 basic producible resources
// or their production variants (e.g., "titanium" or "titanium-production").
func isProducibleResource(resourceType shared.ResourceType) bool {
	switch resourceType {
	case shared.ResourceCredit, shared.ResourceSteel, shared.ResourceTitanium,
		shared.ResourcePlant, shared.ResourceEnergy, shared.ResourceHeat,
		shared.ResourceCreditProduction, shared.ResourceSteelProduction, shared.ResourceTitaniumProduction,
		shared.ResourcePlantProduction, shared.ResourceEnergyProduction, shared.ResourceHeatProduction:
		return true
	default:
		return false
	}
}

// validateGlobalParamWarnings checks behavior outputs for global parameter raises
// that would have no effect because the parameter is already at maximum.
func validateGlobalParamWarnings(outputs []shared.BehaviorCondition, g *game.Game) []player.StateWarning {
	var warnings []player.StateWarning
	seen := make(map[player.StateWarningCode]bool)

	addWarning := func(code player.StateWarningCode, msg string) {
		if !seen[code] {
			seen[code] = true
			warnings = append(warnings, player.StateWarning{Code: code, Message: msg})
		}
	}

	for _, output := range outputs {
		if output.GetAmount() <= 0 && !shared.IsVariableAmount(output) {
			continue
		}
		switch output.GetResourceType() {
		case shared.ResourceTemperature:
			if g.GlobalParameters().Temperature() >= global_parameters.MaxTemperature {
				addWarning(player.WarningCodeGlobalParamMaxed, "Temperature already at maximum")
			}
		case shared.ResourceOxygen:
			if g.GlobalParameters().Oxygen() >= global_parameters.MaxOxygen {
				addWarning(player.WarningCodeGlobalParamMaxed, "Oxygen already at maximum")
			}
		case shared.ResourceOcean, shared.ResourceOceanTile, shared.ResourceOceanPlacement:
			if g.GlobalParameters().Oceans() >= g.GlobalParameters().GetMaxOceans() {
				addWarning(player.WarningCodeGlobalParamMaxed, "All oceans already placed")
			}
		case shared.ResourceVenus:
			if g.GlobalParameters().Venus() >= global_parameters.MaxVenus {
				addWarning(player.WarningCodeGlobalParamMaxed, "Venus already at maximum")
			}
		}

		// Greenery tiles raise oxygen
		if output.GetResourceType() == shared.ResourceGreeneryTile {
			if g.GlobalParameters().Oxygen() >= global_parameters.MaxOxygen {
				addWarning(player.WarningCodeNoTRGain, "Oxygen already at maximum")
			}
		}
	}

	return warnings
}

// resourceDisplayNames maps ResourceType values to human-readable names for error messages
var resourceDisplayNames = map[shared.ResourceType]string{
	// Base resources
	shared.ResourceCredit:   "credits",
	shared.ResourceSteel:    "steel",
	shared.ResourceTitanium: "titanium",
	shared.ResourcePlant:    "plants",
	shared.ResourceEnergy:   "energy",
	shared.ResourceHeat:     "heat",
	shared.ResourceMicrobe:  "microbes",
	shared.ResourceAnimal:   "animals",
	shared.ResourceFloater:  "floaters",
	shared.ResourceScience:  "science",
	shared.ResourceAsteroid: "asteroids",
	shared.ResourceDisease:  "disease",

	// Production types (map to base resource name)
	shared.ResourceCreditProduction:   "credit",
	shared.ResourceSteelProduction:    "steel",
	shared.ResourceTitaniumProduction: "titanium",
	shared.ResourcePlantProduction:    "plant",
	shared.ResourceEnergyProduction:   "energy",
	shared.ResourceHeatProduction:     "heat",

	// Global parameters
	shared.ResourceTemperature: "temperature",
	shared.ResourceOxygen:      "oxygen",
	shared.ResourceOcean:       "ocean",
	shared.ResourceVenus:       "venus",
	shared.ResourceTR:          "terraform rating",
}

// getResourceDisplayName returns human-readable name for a ResourceType, falling back to the raw value
func getResourceDisplayName(resourceType shared.ResourceType) string {
	if name, ok := resourceDisplayNames[resourceType]; ok {
		return name
	}
	return string(resourceType)
}

func formatInsufficientResourceMessage(resourceType shared.ResourceType) string {
	return fmt.Sprintf("Not enough %s", getResourceDisplayName(resourceType))
}

func formatTooMuchResourceMessage(resourceType shared.ResourceType) string {
	return fmt.Sprintf("Too much %s", getResourceDisplayName(resourceType))
}

func formatInsufficientProductionMessage(resourceType shared.ResourceType) string {
	return fmt.Sprintf("Not enough %s production", getResourceDisplayName(resourceType))
}

// formatTagDisplayName returns a human-readable tag name.
// Proper nouns (Venus, Earth, Jovian) get capitalized; other tags stay lowercase.
func formatTagDisplayName(tag string) string {
	switch strings.ToLower(tag) {
	case "venus", "earth", "jovian":
		return strings.ToUpper(tag[:1]) + strings.ToLower(tag[1:])
	default:
		return strings.ToLower(tag)
	}
}

func formatInsufficientTagsMessage(tag string) string {
	return fmt.Sprintf("Not enough %s tags", formatTagDisplayName(tag))
}

func formatTooManyTagsMessage(tag string) string {
	return fmt.Sprintf("Too many %s tags", formatTagDisplayName(tag))
}

// CalculateMilestoneState computes eligibility state for claiming a milestone.
// Returns EntityState with errors indicating why the milestone cannot be claimed.
func CalculateMilestoneState(
	milestoneType shared.MilestoneType,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
	milestoneRegistry milestones.MilestoneRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]interface{})

	def, err := milestoneRegistry.GetByID(string(milestoneType))
	if err != nil {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInvalidRequirement,
			Category: player.ErrorCategoryConfiguration,
			Message:  fmt.Sprintf("Unknown milestone type: %s", milestoneType),
		})
		return player.EntityState{
			Errors:         errors,
			Cost:           make(map[string]int),
			Metadata:       metadata,
			LastCalculated: time.Now(),
		}
	}

	ms := g.Milestones()

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	if ms.IsClaimed(milestoneType) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMilestoneAlreadyClaimed,
			Category: player.ErrorCategoryAchievement,
			Message:  "Already claimed",
		})
	}

	if ms.ClaimedCount() >= game.MaxClaimedMilestones {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMaxMilestonesClaimed,
			Category: player.ErrorCategoryAchievement,
			Message:  "Maximum milestones claimed",
		})
	}

	progress := gamecards.CalculateMilestoneProgress(def, p, g.Board(), cardRegistry)
	required := def.GetRequired()
	metadata["progress"] = progress
	metadata["required"] = required

	if progress < required {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMilestoneRequirementNotMet,
			Category: player.ErrorCategoryRequirement,
			Message:  "Requirement not met",
		})
	}

	cost := def.ClaimCost
	if p.Resources().Get().Credits < cost {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInsufficientCredits,
			Category: player.ErrorCategoryCost,
			Message:  "Cannot afford",
		})
	}

	costMap := map[string]int{string(shared.ResourceCredit): cost}

	return player.EntityState{
		Errors:         errors,
		Cost:           costMap,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// CalculateAwardState computes eligibility state for funding an award.
// Returns EntityState with errors indicating why the award cannot be funded.
func CalculateAwardState(
	awardType shared.AwardType,
	p *player.Player,
	g *game.Game,
	awardRegistry awards.AwardRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]interface{})

	def, err := awardRegistry.GetByID(string(awardType))
	if err != nil {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInvalidRequirement,
			Category: player.ErrorCategoryConfiguration,
			Message:  fmt.Sprintf("Unknown award type: %s", awardType),
		})
		return player.EntityState{
			Errors:         errors,
			Cost:           make(map[string]int),
			Metadata:       metadata,
			LastCalculated: time.Now(),
		}
	}

	gameAwards := g.Awards()

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	if gameAwards.IsFunded(awardType) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeAwardAlreadyFunded,
			Category: player.ErrorCategoryAchievement,
			Message:  "Already funded",
		})
	}

	if gameAwards.FundedCount() >= game.MaxFundedAwards {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMaxAwardsFunded,
			Category: player.ErrorCategoryAchievement,
			Message:  "Maximum awards funded",
		})
	}

	cost := def.GetCostForFundedCount(gameAwards.FundedCount())
	metadata["fundingCost"] = cost

	if p.Resources().Get().Credits < cost {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInsufficientCredits,
			Category: player.ErrorCategoryCost,
			Message:  "Cannot afford",
		})
	}

	costMap := map[string]int{string(shared.ResourceCredit): cost}

	return player.EntityState{
		Errors:         errors,
		Cost:           costMap,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// CalculateChoiceErrors validates a single choice's requirements against the player/game state.
// Returns a list of errors explaining why the choice is unavailable (empty if available).
func CalculateChoiceErrors(choice shared.Choice, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []player.StateError {
	var errors []player.StateError

	if choice.Requirements != nil && len(choice.Requirements.Items) > 0 {
		for _, req := range choice.Requirements.Items {
			if err := checkChoiceRequirement(req, p, g, cardRegistry); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	resources := p.Resources().Get()
	for _, outputBC := range choice.Outputs {
		if shared.IsVariableAmount(outputBC) || outputBC.GetAmount() >= 0 {
			continue
		}
		if !isBasicPlayerResource(outputBC.GetResourceType()) {
			continue
		}
		available := resources.GetAmount(outputBC.GetResourceType())
		if available < -outputBC.GetAmount() {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryInput,
				Message:  fmt.Sprintf("Not enough %s", outputBC.GetResourceType()),
			})
		}
	}

	for _, inputBC := range choice.Inputs {
		if shared.IsVariableAmount(inputBC) || inputBC.GetAmount() <= 0 {
			continue
		}
		if !isBasicPlayerResource(inputBC.GetResourceType()) {
			continue
		}
		available := resources.GetAmount(inputBC.GetResourceType())
		if available < inputBC.GetAmount() {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryInput,
				Message:  fmt.Sprintf("Not enough %s", inputBC.GetResourceType()),
			})
		}
	}

	return errors
}

// checkChoiceRequirement validates a single choice requirement and returns a StateError if not met.
func checkChoiceRequirement(req shared.ChoiceRequirement, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) *player.StateError {
	switch req.Type {
	case "tags":
		if req.Tag == nil {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid tag requirement",
			}
		}
		tagCount := gamecards.CountPlayerTagsByType(p, cardRegistry, *req.Tag)
		if req.Min != nil && tagCount < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeInsufficientTags,
				Category: player.ErrorCategoryRequirement,
				Message:  formatInsufficientTagsMessage(string(*req.Tag)),
			}
		}
		if req.Max != nil && tagCount > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeTooManyTags,
				Category: player.ErrorCategoryRequirement,
				Message:  formatTooManyTagsMessage(string(*req.Tag)),
			}
		}

	case "temperature":
		temp := g.GlobalParameters().Temperature()
		if req.Min != nil && temp < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeTemperatureTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Temperature too low",
			}
		}
		if req.Max != nil && temp > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeTemperatureTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Temperature too high",
			}
		}

	case "oxygen":
		oxygen := g.GlobalParameters().Oxygen()
		if req.Min != nil && oxygen < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeOxygenTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Oxygen too low",
			}
		}
		if req.Max != nil && oxygen > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeOxygenTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Oxygen too high",
			}
		}

	case "ocean":
		oceans := g.GlobalParameters().Oceans()
		if req.Min != nil && oceans < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeOceansTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "Too few oceans",
			}
		}
		if req.Max != nil && oceans > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeOceansTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "Too many oceans",
			}
		}

	case "tr":
		tr := p.Resources().TerraformRating()
		if req.Min != nil && tr < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeTRTooLow,
				Category: player.ErrorCategoryRequirement,
				Message:  "TR too low",
			}
		}
		if req.Max != nil && tr > *req.Max {
			return &player.StateError{
				Code:     player.ErrorCodeTRTooHigh,
				Category: player.ErrorCategoryRequirement,
				Message:  "TR too high",
			}
		}

	case "production":
		if req.Resource == nil {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid production requirement",
			}
		}
		production := p.Resources().Production()
		amount := production.GetAmount(*req.Resource)
		if req.Min != nil && amount < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeInsufficientProduction,
				Category: player.ErrorCategoryRequirement,
				Message:  formatInsufficientProductionMessage(*req.Resource),
			}
		}

	case "resource":
		if req.Resource == nil {
			return &player.StateError{
				Code:     player.ErrorCodeInvalidRequirement,
				Category: player.ErrorCategoryRequirement,
				Message:  "Invalid resource requirement",
			}
		}
		resources := p.Resources().Get()
		amount := resources.GetAmount(*req.Resource)
		if req.Min != nil && amount < *req.Min {
			return &player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryRequirement,
				Message:  formatInsufficientResourceMessage(*req.Resource),
			}
		}
	}

	return nil
}

// computeBehaviorValues computes per-condition output values for card behaviors.
// Returns a slice of ComputedBehaviorValue with target format "behaviors::N".
func computeBehaviorValues(
	behaviors []shared.CardBehavior,
	sourceCardID string,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
	colonyBonusLookup gamecards.ColonyBonusLookup,
) []player.ComputedBehaviorValue {
	var result []player.ComputedBehaviorValue

	board := g.Board()
	allPlayers := g.GetAllPlayers()

	for i, behavior := range behaviors {
		var outputs []shared.CalculatedOutput
		for _, outputBC := range behavior.Outputs {
			if outputBC.GetResourceType() == shared.ResourceColonyBonus {
				bonusOutputs := computeColonyBonusValues(p, g, colonyBonusLookup)
				if outputs == nil {
					outputs = bonusOutputs
				} else {
					outputs = append(outputs, bonusOutputs...)
				}
				continue
			}
			per := shared.GetPerCondition(outputBC)
			if per == nil {
				continue
			}
			count := gamecards.CountPerCondition(per, sourceCardID, p, board, cardRegistry, allPlayers)
			if per.Amount > 0 {
				multiplier := count / per.Amount
				actualAmount := outputBC.GetAmount() * multiplier
				outputs = append(outputs, shared.CalculatedOutput{
					ResourceType: string(outputBC.GetResourceType()),
					Amount:       actualAmount,
					IsScaled:     true,
				})
			}
		}
		for _, choice := range behavior.Choices {
			for _, outputBC := range choice.Outputs {
				per := shared.GetPerCondition(outputBC)
				if per == nil {
					continue
				}
				count := gamecards.CountPerCondition(per, sourceCardID, p, board, cardRegistry, allPlayers)
				if per.Amount > 0 {
					multiplier := count / per.Amount
					actualAmount := outputBC.GetAmount() * multiplier
					outputs = append(outputs, shared.CalculatedOutput{
						ResourceType: string(outputBC.GetResourceType()),
						Amount:       actualAmount,
						IsScaled:     true,
					})
				}
			}
		}
		if outputs != nil {
			result = append(result, player.ComputedBehaviorValue{
				Target:  fmt.Sprintf("behaviors::%d", i),
				Outputs: outputs,
			})
		}
	}

	return result
}

func computeColonyBonusValues(
	p *player.Player,
	g *game.Game,
	colonyBonusLookup gamecards.ColonyBonusLookup,
) []shared.CalculatedOutput {
	if p == nil || g == nil || colonyBonusLookup == nil || !g.HasColonies() {
		return []shared.CalculatedOutput{}
	}
	return gamecards.ColonyBonusesToCalculatedOutputs(
		gamecards.CollectColonyBonuses(p.ID(), g.Colonies().States(), colonyBonusLookup),
	)
}
