package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/shared"
)

// ValidateCardJSON validates the JSON structure of a card at load time
// This ensures all enum values are valid and structure is correct
func ValidateCardJSON(card *Card) []error {
	var errors []error

	if !isValidCardType(card.Type) {
		errors = append(errors, fmt.Errorf("card %s: invalid card type: %s", card.ID, card.Type))
	}

	for _, tag := range card.Tags {
		if !isValidCardTag(tag) {
			errors = append(errors, fmt.Errorf("card %s: invalid tag: %s", card.ID, tag))
		}
	}

	if card.Requirements != nil {
		for i, req := range card.Requirements.Items {
			if reqErr := validateRequirement(card.ID, i, req); reqErr != nil {
				errors = append(errors, reqErr)
			}
		}
	}

	for i, behavior := range card.Behaviors {
		behaviorErrors := validateBehavior(card.ID, i, behavior)
		errors = append(errors, behaviorErrors...)
	}

	if card.ResourceStorage != nil {
		if !isValidResourceType(card.ResourceStorage.Type) {
			errors = append(errors, fmt.Errorf("card %s: invalid resource storage type: %s", card.ID, card.ResourceStorage.Type))
		}
	}

	for i, vp := range card.VPConditions {
		if vpErr := validateVictoryPointCondition(card.ID, i, vp); vpErr != nil {
			errors = append(errors, vpErr)
		}
	}

	if card.StartingResources != nil {
		if rsErr := validateResourceSet(card.ID, "starting resources", *card.StartingResources); rsErr != nil {
			errors = append(errors, rsErr)
		}
	}

	if card.StartingProduction != nil {
		if prodErr := validateResourceSet(card.ID, "starting production", *card.StartingProduction); prodErr != nil {
			errors = append(errors, prodErr)
		}
	}

	return errors
}

func validateRequirement(cardID string, index int, req Requirement) error {
	if !isValidRequirementType(req.Type) {
		return fmt.Errorf("card %s: requirement[%d] has invalid type: %s", cardID, index, req.Type)
	}

	if req.Type == RequirementTags && req.Tag != nil {
		if !isValidCardTag(*req.Tag) {
			return fmt.Errorf("card %s: requirement[%d] has invalid tag: %s", cardID, index, *req.Tag)
		}
	}

	return nil
}

func validateBehavior(cardID string, index int, behavior shared.CardBehavior) []error {
	var errors []error

	for i, trigger := range behavior.Triggers {
		if trigger.Type == "" {
			errors = append(errors, fmt.Errorf("card %s: behavior[%d].trigger[%d] has empty type", cardID, index, i))
		}

		if trigger.Condition != nil {
			for _, rt := range trigger.Condition.ResourceTypes {
				if !isValidResourceType(rt) {
					errors = append(errors, fmt.Errorf("card %s: behavior[%d].trigger[%d] has invalid resource type: %s", cardID, index, i, rt))
				}
			}
		}
	}

	for i, input := range behavior.Inputs {
		if inputErr := validateBehaviorCondition(cardID, index, "input", i, input); inputErr != nil {
			errors = append(errors, inputErr)
		}
	}

	for i, output := range behavior.Outputs {
		if outputErr := validateBehaviorCondition(cardID, index, "output", i, output); outputErr != nil {
			errors = append(errors, outputErr)
		}
	}

	for i, choice := range behavior.Choices {
		for k, input := range choice.Inputs {
			if inputErr := validateBehaviorCondition(cardID, index, fmt.Sprintf("choice[%d].input", i), k, input); inputErr != nil {
				errors = append(errors, inputErr)
			}
		}

		for k, output := range choice.Outputs {
			if outputErr := validateBehaviorCondition(cardID, index, fmt.Sprintf("choice[%d].output", i), k, output); outputErr != nil {
				errors = append(errors, outputErr)
			}
		}
	}

	return errors
}

func validateBehaviorCondition(cardID string, behaviorIndex int, condType string, index int, cond shared.BehaviorCondition) error {
	if cond.GetTarget() == "" {
		return fmt.Errorf("card %s: behavior[%d].%s[%d] has empty target", cardID, behaviorIndex, condType, index)
	}

	if !isValidResourceType(cond.GetResourceType()) {
		return fmt.Errorf("card %s: behavior[%d].%s[%d] has invalid resource type: %s", cardID, behaviorIndex, condType, index, cond.GetResourceType())
	}

	if per := shared.GetPerCondition(cond); per != nil {
		if !isValidResourceType(per.ResourceType) {
			return fmt.Errorf("card %s: behavior[%d].%s[%d].per has invalid resource type: %s", cardID, behaviorIndex, condType, index, per.ResourceType)
		}
	}

	return nil
}

func validateVictoryPointCondition(cardID string, index int, vp VictoryPointCondition) error {
	if vp.Per != nil {
		if !isValidResourceType(vp.Per.ResourceType) {
			return fmt.Errorf("card %s: victory_point_condition[%d].per has invalid resource type: %s", cardID, index, vp.Per.ResourceType)
		}
	}

	return nil
}

func validateResourceSet(cardID, fieldName string, rs shared.ResourceSet) error {
	return nil
}

func isValidCardType(ct CardType) bool {
	switch ct {
	case CardTypeCorporation, CardTypeAutomated, CardTypeActive, CardTypeEvent, CardTypePrelude:
		return true
	default:
		return false
	}
}

func isValidCardTag(tag shared.CardTag) bool {
	switch tag {
	case shared.TagBuilding, shared.TagSpace, shared.TagScience,
		shared.TagPower, shared.TagEarth, shared.TagJovian,
		shared.TagVenus, shared.TagPlant, shared.TagMicrobe,
		shared.TagAnimal, shared.TagCity, shared.TagEvent,
		shared.TagWildlife, shared.TagWild:
		return true
	default:
		return false
	}
}

func isValidRequirementType(rt RequirementType) bool {
	switch rt {
	case RequirementTemperature, RequirementOxygen, RequirementOceans,
		RequirementTags, RequirementProduction, RequirementTR,
		RequirementResource, RequirementVenus, RequirementCities,
		RequirementGreeneries:
		return true
	default:
		return false
	}
}

func isValidResourceType(rt shared.ResourceType) bool {
	return rt != ""
}
