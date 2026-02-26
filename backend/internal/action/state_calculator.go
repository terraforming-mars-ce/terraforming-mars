package action

import (
	"fmt"
	"strings"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CalculatePlayerCardState computes playability state for a card.
// This function can access both Game and Player without circular dependencies.
// card parameter must be *gamecards.Card
func CalculatePlayerCardState(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	var warnings []player.StateWarning
	metadata := make(map[string]interface{})

	errors = append(errors, validatePhase(g)...)
	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoActiveTileSelection(p, g)...)

	costMap, discounts := calculateEffectiveCost(card, p, cardRegistry)
	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	errors = append(errors, validateAffordabilityWithSubstitutes(p, costMap)...)
	errors = append(errors, validateRequirements(card, p, g, cardRegistry)...)
	errors = append(errors, validateProductionOutputs(card, p)...)

	tileErrors, tileWarnings := validateTileOutputs(card, p, g)
	errors = append(errors, tileErrors...)
	warnings = append(warnings, tileWarnings...)

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
	errors = append(errors, validateNoActiveTileSelection(p, g)...)

	resources := p.Resources().Get()
	for _, input := range behavior.Inputs {
		available := getResourceAmount(resources, input.ResourceType)
		if available < input.Amount {
			errors = append(errors, player.StateError{
				Code:     player.ErrorCodeInsufficientResources,
				Category: player.ErrorCategoryInput,
				Message:  fmt.Sprintf("Not enough %s", input.ResourceType),
			})
		}
	}

	for _, output := range behavior.Outputs {
		if output.Target == "steal-from-any-card" {
			totalAvailable := 0
			for _, anyPlayer := range g.GetAllPlayers() {
				for _, cardID := range anyPlayer.PlayedCards().Cards() {
					totalAvailable += anyPlayer.Resources().GetCardStorage(cardID)
				}
			}
			if totalAvailable < output.Amount {
				errors = append(errors, player.StateError{
					Code:     player.ErrorCodeInsufficientResources,
					Category: player.ErrorCategoryInput,
					Message:  fmt.Sprintf("No %s available on any card", output.ResourceType),
				})
			}
		}
	}

	errors = append(errors, validateActionUsageLimit(behavior, timesUsedThisGeneration)...)
	errors = append(errors, validateBehaviorTileOutputs(behavior, p, g)...)
	errors = append(errors, validateGenerationalEventRequirements(behavior, p)...)

	return player.EntityState{
		Errors:         errors,
		Cost:           make(map[string]int), // Actions typically don't have credit costs (empty map)
		Metadata:       make(map[string]interface{}),
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
	metadata := make(map[string]interface{})

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoActiveTileSelection(p, g)...)

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

	case shared.StandardProjectPowerPlant:

	default:
	}

	return player.EntityState{
		Errors:         errors,
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
	if g.CurrentPhase() != game.GamePhaseAction {
		return []player.StateError{{
			Code:     player.ErrorCodeWrongPhase,
			Category: player.ErrorCategoryPhase,
			Message:  fmt.Sprintf("Can only play cards during action phase, current phase: %s", g.CurrentPhase()),
		}}
	}
	return nil
}

// validateNoActiveTileSelection checks if player has an active tile selection pending.
func validateNoActiveTileSelection(p *player.Player, g *game.Game) []player.StateError {
	if g.GetPendingTileSelection(p.ID()) != nil {
		return []player.StateError{{
			Code:     player.ErrorCodeActiveTileSelection,
			Category: player.ErrorCategoryPhase,
			Message:  "Active tile selection",
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

// validateRequirements checks all card requirements.
// Extracted from PlayCardAction.validateCardRequirements() lines 209-321.
func validateRequirements(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) []player.StateError {
	if card.Requirements == nil || len(card.Requirements.Items) == 0 {
		return nil
	}

	var errors []player.StateError

	for _, req := range card.Requirements.Items {
		err := checkRequirement(req, p, g, cardRegistry)
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
		for _, output := range behavior.Outputs {
			// Only check production resource types with negative amounts
			if output.Amount >= 0 {
				continue
			}

			// Map production resource types to base resource types for checking
			var baseResourceType shared.ResourceType
			switch output.ResourceType {
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
			currentProduction := getProductionAmount(production, baseResourceType)
			resultingProduction := currentProduction + output.Amount // output.Amount is negative

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

// validateTileOutputs checks that the board has available placements for any tile outputs.
// If a card outputs city/greenery/ocean tiles, the player must have valid placement locations.
// Returns both errors (blocking) and warnings (non-blocking).
// If tile restrictions are specified (e.g., adjacency="none"), missing placements become warnings.
func validateTileOutputs(
	card *gamecards.Card,
	p *player.Player,
	g *game.Game,
) ([]player.StateError, []player.StateWarning) {
	if len(card.Behaviors) == 0 || g == nil {
		return nil, nil
	}

	var errors []player.StateError
	var warnings []player.StateWarning

	// Check all behaviors for auto-triggers with tile placement outputs
	for _, behavior := range card.Behaviors {
		// Only check auto-trigger behaviors (immediate effects when card is played)
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		// Check outputs for tile placements
		for _, output := range behavior.Outputs {
			switch output.ResourceType {
			case shared.ResourceCityPlacement:
				// Extract tile restrictions
				var tileRestrictions *shared.TileRestrictions
				if output.TileRestrictions != nil {
					tileRestrictions = output.TileRestrictions
				}

				cityPlacements := g.CountAvailableHexesForTile("city", p.ID(), tileRestrictions)
				if cityPlacements == 0 {
					// If restrictions exist, it's a warning (card still playable)
					// If no restrictions, it's an error (normal city rules blocked)
					if tileRestrictions != nil && (len(tileRestrictions.BoardTags) > 0 || tileRestrictions.Adjacency != "") {
						warnings = append(warnings, player.StateWarning{
							Code:    player.WarningCodeNoValidTilePlacements,
							Message: "No valid city placements available",
						})
					} else {
						errors = append(errors, player.StateError{
							Code:     player.ErrorCodeNoCityPlacements,
							Category: player.ErrorCategoryAvailability,
							Message:  "No valid city placements",
						})
					}
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

			case shared.ResourceVolcanoPlacement:
				volcanoPlacements := g.CountAvailableHexesForTile("volcano", p.ID(), nil)
				if volcanoPlacements == 0 {
					warnings = append(warnings, player.StateWarning{
						Code:    player.WarningCodeNoValidTilePlacements,
						Message: "No valid volcanic placements available",
					})
				}
			}
		}
	}

	return errors, warnings
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
	for _, output := range behavior.Outputs {
		switch output.ResourceType {
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
		}
	}

	return errors
}

// checkRequirement validates a single requirement.
// Extracted from PlayCardAction - contains the switch statement for all requirement types.
func checkRequirement(
	req gamecards.Requirement,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) *player.StateError {
	switch req.Type {
	case gamecards.RequirementTemperature:
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

	case gamecards.RequirementOxygen:
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

	case gamecards.RequirementOceans:
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
		for _, playedCardID := range p.PlayedCards().Cards() {
			if cardRegistry == nil {
				continue
			}
			card, err := cardRegistry.GetByID(playedCardID)
			if err != nil {
				continue
			}
			for _, tag := range card.Tags {
				if tag == *req.Tag {
					tagCount++
				}
			}
		}

		if corpID := p.CorporationID(); corpID != "" && cardRegistry != nil {
			if corp, err := cardRegistry.GetByID(corpID); err == nil {
				for _, tag := range corp.Tags {
					if tag == *req.Tag {
						tagCount++
					}
				}
			}
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
		currentProduction := getProductionAmount(production, *req.Resource)

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
		currentAmount := getResourceAmount(resources, *req.Resource)

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

	case gamecards.RequirementVenus:
		// TODO: Implement Venus track when expansion is supported
		// For now, skip Venus validation (same as PlayCardAction line 314-316)
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
// considering payment substitutes like Helion's heat-to-credit conversion.
func validateAffordabilityWithSubstitutes(p *player.Player, costMap map[string]int) []player.StateError {
	var errors []player.StateError
	resources := p.Resources().Get()
	substitutes := p.Resources().PaymentSubstitutes()

	for resourceType, cost := range costMap {
		if shared.ResourceType(resourceType) == shared.ResourceCredit {
			// For credit costs, calculate effective purchasing power including substitutes
			effectiveCredits := resources.Credits

			// Add substitute resources at their conversion rates
			for _, sub := range substitutes {
				switch sub.ResourceType {
				case shared.ResourceHeat:
					effectiveCredits += resources.Heat * sub.ConversionRate
				case shared.ResourceEnergy:
					effectiveCredits += resources.Energy * sub.ConversionRate
				case shared.ResourcePlant:
					effectiveCredits += resources.Plants * sub.ConversionRate
				case shared.ResourceSteel:
					effectiveCredits += resources.Steel * sub.ConversionRate
				case shared.ResourceTitanium:
					effectiveCredits += resources.Titanium * sub.ConversionRate
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
			available := getResourceAmount(resources, shared.ResourceType(resourceType))
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

// getResourceAmount extracts the amount of a specific resource from Resources.
func getResourceAmount(resources shared.Resources, resourceType shared.ResourceType) int {
	switch resourceType {
	case shared.ResourceCredit:
		return resources.Credits
	case shared.ResourceSteel:
		return resources.Steel
	case shared.ResourceTitanium:
		return resources.Titanium
	case shared.ResourcePlant:
		return resources.Plants
	case shared.ResourceEnergy:
		return resources.Energy
	case shared.ResourceHeat:
		return resources.Heat
	default:
		return 0
	}
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

// getProductionAmount extracts the production value for a specific resource type.
// Accepts both base resources (e.g., "titanium") and production types (e.g., "titanium-production").
// Returns 0 for non-producible resource types.
func getProductionAmount(production shared.Production, resourceType shared.ResourceType) int {
	switch resourceType {
	case shared.ResourceCredit, shared.ResourceCreditProduction:
		return production.Credits
	case shared.ResourceSteel, shared.ResourceSteelProduction:
		return production.Steel
	case shared.ResourceTitanium, shared.ResourceTitaniumProduction:
		return production.Titanium
	case shared.ResourcePlant, shared.ResourcePlantProduction:
		return production.Plants
	case shared.ResourceEnergy, shared.ResourceEnergyProduction:
		return production.Energy
	case shared.ResourceHeat, shared.ResourceHeatProduction:
		return production.Heat
	default:
		return 0
	}
}

// resourceDisplayNames maps ResourceType values to human-readable names for error messages
var resourceDisplayNames = map[shared.ResourceType]string{
	// Base resources
	shared.ResourceCredit:   "credit",
	shared.ResourceSteel:    "steel",
	shared.ResourceTitanium: "titanium",
	shared.ResourcePlant:    "plant",
	shared.ResourceEnergy:   "energy",
	shared.ResourceHeat:     "heat",
	shared.ResourceMicrobe:  "microbe",
	shared.ResourceAnimal:   "animal",
	shared.ResourceFloater:  "floater",
	shared.ResourceScience:  "science",
	shared.ResourceAsteroid: "asteroid",
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

// formatInsufficientTagsMessage returns "Not enough {tag} tags" with the tag name lowercased.
func formatInsufficientTagsMessage(tag string) string {
	return fmt.Sprintf("Not enough %s tags", strings.ToLower(tag))
}

// formatTooManyTagsMessage returns "Too many {tag} tags" with the tag name lowercased.
func formatTooManyTagsMessage(tag string) string {
	return fmt.Sprintf("Too many %s tags", strings.ToLower(tag))
}

// CalculateMilestoneState computes eligibility state for claiming a milestone.
// Returns EntityState with errors indicating why the milestone cannot be claimed.
func CalculateMilestoneState(
	milestoneType shared.MilestoneType,
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]interface{})

	milestoneInfo, found := game.GetMilestoneInfo(milestoneType)
	if !found {
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

	milestones := g.Milestones()

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoActiveTileSelection(p, g)...)

	if milestones.IsClaimed(milestoneType) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMilestoneAlreadyClaimed,
			Category: player.ErrorCategoryAchievement,
			Message:  "Already claimed",
		})
	}

	if milestones.ClaimedCount() >= game.MaxClaimedMilestones {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMaxMilestonesClaimed,
			Category: player.ErrorCategoryAchievement,
			Message:  "Maximum milestones claimed",
		})
	}

	progress := gamecards.GetPlayerMilestoneProgress(milestoneType, p, g.Board(), cardRegistry)
	required := milestoneInfo.Requirement
	metadata["progress"] = progress
	metadata["required"] = required

	if progress < required {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMilestoneRequirementNotMet,
			Category: player.ErrorCategoryRequirement,
			Message:  formatMilestoneRequirementError(milestoneType),
		})
	}

	cost := game.MilestoneClaimCost
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
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]interface{})

	if !shared.ValidAwardType(string(awardType)) {
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

	awards := g.Awards()

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoActiveTileSelection(p, g)...)

	if awards.IsFunded(awardType) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeAwardAlreadyFunded,
			Category: player.ErrorCategoryAchievement,
			Message:  "Already funded",
		})
	}

	if awards.FundedCount() >= game.MaxFundedAwards {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMaxAwardsFunded,
			Category: player.ErrorCategoryAchievement,
			Message:  "Maximum awards funded",
		})
	}

	cost := awards.GetCurrentFundingCost()
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

// formatMilestoneRequirementError returns a short error message for milestone requirements.
func formatMilestoneRequirementError(milestoneType shared.MilestoneType) string {
	switch milestoneType {
	case shared.MilestoneTerraformer:
		return "Not enough TR"
	case shared.MilestoneMayor:
		return "Not enough cities"
	case shared.MilestoneGardener:
		return "Not enough greeneries"
	case shared.MilestoneBuilder:
		return "Not enough building tags"
	case shared.MilestonePlanner:
		return "Not enough cards"
	default:
		return "Requirement not met"
	}
}
