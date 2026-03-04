package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ValidateCardCanBePlayed checks if a card can be played in the current game context
// Returns nil if valid, or an error describing why the card cannot be played
func ValidateCardCanBePlayed(
	card *Card,
	pl *player.Player,
	globalParams *global_parameters.GlobalParameters,
	playedCards []*Card, // All cards played by the player (for tag counting)
) error {
	if card.Requirements == nil {
		return nil
	}
	for _, req := range card.Requirements.Items {
		if err := validateRequirementMet(req, pl, globalParams, playedCards); err != nil {
			return fmt.Errorf("requirement not met: %w", err)
		}
	}

	return nil
}

// CanAffordCard checks if a player can afford to pay for a card
func CanAffordCard(card *Card, pl *player.Player, discounts map[shared.CardTag]int) bool {
	totalCost := card.Cost

	// Apply tag-based discounts
	for _, tag := range card.Tags {
		if discount, ok := discounts[tag]; ok {
			totalCost -= discount
		}
	}

	// Cost cannot go below 0
	if totalCost < 0 {
		totalCost = 0
	}

	resources := pl.Resources().Get()
	return resources.Credits >= totalCost
}

// validateRequirementMet checks if a single requirement is met
func validateRequirementMet(
	req Requirement,
	pl *player.Player,
	globalParams *global_parameters.GlobalParameters,
	playedCards []*Card,
) error {
	switch req.Type {
	case RequirementTemperature:
		return validateTemperatureRequirement(req, globalParams)
	case RequirementOxygen:
		return validateOxygenRequirement(req, globalParams)
	case RequirementOceans:
		return validateOceansRequirement(req, globalParams)
	case RequirementTags:
		return validateTagsRequirement(req, playedCards)
	case RequirementProduction:
		return validateProductionRequirement(req, pl)
	case RequirementTR:
		return validateTRRequirement(req, pl)
	case RequirementResource:
		return validateResourceRequirement(req, pl)
	case RequirementCities:
		return validateCitiesRequirement(req)
	case RequirementGreeneries:
		return validateGreeeneriesRequirement(req)
	default:
		return fmt.Errorf("unknown requirement type: %s", req.Type)
	}
}

// validateTemperatureRequirement checks if temperature requirement is met
func validateTemperatureRequirement(req Requirement, globalParams *global_parameters.GlobalParameters) error {
	temp := globalParams.Temperature()

	if req.Min != nil && temp < *req.Min {
		return fmt.Errorf("temperature %d°C is below required minimum %d°C", temp, *req.Min)
	}

	if req.Max != nil && temp > *req.Max {
		return fmt.Errorf("temperature %d°C is above required maximum %d°C", temp, *req.Max)
	}

	return nil
}

// validateOxygenRequirement checks if oxygen requirement is met
func validateOxygenRequirement(req Requirement, globalParams *global_parameters.GlobalParameters) error {
	oxygen := globalParams.Oxygen()

	if req.Min != nil && oxygen < *req.Min {
		return fmt.Errorf("oxygen %d%% is below required minimum %d%%", oxygen, *req.Min)
	}

	if req.Max != nil && oxygen > *req.Max {
		return fmt.Errorf("oxygen %d%% is above required maximum %d%%", oxygen, *req.Max)
	}

	return nil
}

// validateOceansRequirement checks if oceans requirement is met
func validateOceansRequirement(req Requirement, globalParams *global_parameters.GlobalParameters) error {
	oceans := globalParams.Oceans()

	if req.Min != nil && oceans < *req.Min {
		return fmt.Errorf("ocean count %d is below required minimum %d", oceans, *req.Min)
	}

	if req.Max != nil && oceans > *req.Max {
		return fmt.Errorf("ocean count %d is above required maximum %d", oceans, *req.Max)
	}

	return nil
}

// validateTagsRequirement checks if tag count requirement is met
func validateTagsRequirement(req Requirement, playedCards []*Card) error {
	if req.Tag == nil {
		return fmt.Errorf("tags requirement missing tag specification")
	}

	tagCount := countTagsInPlayedCards(*req.Tag, playedCards)

	if req.Min != nil && tagCount < *req.Min {
		return fmt.Errorf("tag %s count %d is below required minimum %d", *req.Tag, tagCount, *req.Min)
	}

	if req.Max != nil && tagCount > *req.Max {
		return fmt.Errorf("tag %s count %d is above required maximum %d", *req.Tag, tagCount, *req.Max)
	}

	return nil
}

// countTagsInPlayedCards counts occurrences of a specific tag in played cards (excluding events).
// Wild tags count toward any tag type.
func countTagsInPlayedCards(tag shared.CardTag, playedCards []*Card) int {
	count := 0
	for _, card := range playedCards {
		if card.Type == CardTypeEvent {
			continue
		}
		count += countTagsInList(card.Tags, tag)
	}
	return count
}

// validateProductionRequirement checks if production requirement is met
func validateProductionRequirement(req Requirement, pl *player.Player) error {
	if req.Resource == nil {
		return fmt.Errorf("production requirement missing resource type specification")
	}

	production := pl.Resources().Production()
	var productionAmount int

	switch *req.Resource {
	case shared.ResourceCredit:
		productionAmount = production.Credits
	case shared.ResourceSteel:
		productionAmount = production.Steel
	case shared.ResourceTitanium:
		productionAmount = production.Titanium
	case shared.ResourcePlant:
		productionAmount = production.Plants
	case shared.ResourceEnergy:
		productionAmount = production.Energy
	case shared.ResourceHeat:
		productionAmount = production.Heat
	default:
		return fmt.Errorf("invalid resource type for production requirement: %s", *req.Resource)
	}

	if req.Min != nil && productionAmount < *req.Min {
		return fmt.Errorf("%s production %d is below required minimum %d", *req.Resource, productionAmount, *req.Min)
	}

	if req.Max != nil && productionAmount > *req.Max {
		return fmt.Errorf("%s production %d is above required maximum %d", *req.Resource, productionAmount, *req.Max)
	}

	return nil
}

// validateTRRequirement checks if terraform rating requirement is met
func validateTRRequirement(req Requirement, pl *player.Player) error {
	tr := pl.Resources().TerraformRating()

	if req.Min != nil && tr < *req.Min {
		return fmt.Errorf("terraform rating %d is below required minimum %d", tr, *req.Min)
	}

	if req.Max != nil && tr > *req.Max {
		return fmt.Errorf("terraform rating %d is above required maximum %d", tr, *req.Max)
	}

	return nil
}

// validateResourceRequirement checks if resource amount requirement is met
func validateResourceRequirement(req Requirement, pl *player.Player) error {
	if req.Resource == nil {
		return fmt.Errorf("resource requirement missing resource type specification")
	}

	resources := pl.Resources().Get()
	var resourceAmount int

	switch *req.Resource {
	case shared.ResourceCredit:
		resourceAmount = resources.Credits
	case shared.ResourceSteel:
		resourceAmount = resources.Steel
	case shared.ResourceTitanium:
		resourceAmount = resources.Titanium
	case shared.ResourcePlant:
		resourceAmount = resources.Plants
	case shared.ResourceEnergy:
		resourceAmount = resources.Energy
	case shared.ResourceHeat:
		resourceAmount = resources.Heat
	default:
		storage := pl.Resources().Storage()
		if amount, ok := storage[string(*req.Resource)]; ok {
			resourceAmount = amount
		} else {
			resourceAmount = 0
		}
	}

	if req.Min != nil && resourceAmount < *req.Min {
		return fmt.Errorf("%s amount %d is below required minimum %d", *req.Resource, resourceAmount, *req.Min)
	}

	if req.Max != nil && resourceAmount > *req.Max {
		return fmt.Errorf("%s amount %d is above required maximum %d", *req.Resource, resourceAmount, *req.Max)
	}

	return nil
}

// validateCitiesRequirement checks if cities requirement is met
// TODO: Implement when city tracking is available
func validateCitiesRequirement(req Requirement) error {
	// Placeholder - requires board state to count cities
	return nil
}

// validateGreeeneriesRequirement checks if greeneries requirement is met
// TODO: Implement when greenery tracking is available
func validateGreeeneriesRequirement(req Requirement) error {
	// Placeholder - requires board state to count greeneries
	return nil
}
