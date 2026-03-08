package cards

import (
	"terraforming-mars-backend/internal/game/shared"
)

// HasAutoTrigger checks if a behavior has an auto trigger without conditions
func HasAutoTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition == nil {
			return true
		}
	}
	return false
}

// HasManualTrigger checks if a behavior has a manual trigger
func HasManualTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerManual) {
			return true
		}
	}
	return false
}

// HasConditionalTrigger checks if a behavior has an auto trigger with a condition (passive effect)
func HasConditionalTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition != nil {
			return true
		}
	}
	return false
}

// HasCorporationStartTrigger checks if a behavior has a corporation start trigger
func HasCorporationStartTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAutoCorporationStart) {
			return true
		}
	}
	return false
}

// HasCorporationFirstActionTrigger checks if a behavior has a corporation first action trigger
func HasCorporationFirstActionTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == string(ResourceTriggerAutoCorporationFirstAction) {
			return true
		}
	}
	return false
}

// GetImmediateBehaviors returns all behaviors with auto triggers (no conditions)
func GetImmediateBehaviors(card *Card) []shared.CardBehavior {
	var immediate []shared.CardBehavior
	for _, behavior := range card.Behaviors {
		if HasAutoTrigger(behavior) {
			immediate = append(immediate, behavior)
		}
	}
	return immediate
}

// GetManualBehaviors returns all behaviors with manual triggers
func GetManualBehaviors(card *Card) []shared.CardBehavior {
	var manual []shared.CardBehavior
	for _, behavior := range card.Behaviors {
		if HasManualTrigger(behavior) {
			manual = append(manual, behavior)
		}
	}
	return manual
}

// GetPassiveBehaviors returns all behaviors with conditional triggers
func GetPassiveBehaviors(card *Card) []shared.CardBehavior {
	var passive []shared.CardBehavior
	for _, behavior := range card.Behaviors {
		if HasConditionalTrigger(behavior) {
			passive = append(passive, behavior)
		}
	}
	return passive
}

// HasPersistentEffects checks if a behavior has persistent outputs that should be
// registered as effects (e.g., discount, payment-substitute, global-parameter-lenience)
// These are different from immediate resource gains - they modify future actions
func HasPersistentEffects(behavior shared.CardBehavior) bool {
	for _, output := range behavior.Outputs {
		switch output.ResourceType {
		case shared.ResourceDiscount, shared.ResourcePaymentSubstitute, shared.ResourceGlobalParameterLenience, shared.ResourceStoragePaymentSubstitute:
			return true
		}
	}
	return false
}

// HasTemporaryOutputs checks if a behavior has any outputs marked as temporary
func HasTemporaryOutputs(behavior shared.CardBehavior) bool {
	for _, output := range behavior.Outputs {
		if output.Temporary != "" {
			return true
		}
	}
	return false
}

// HasChoices checks if a behavior has player choices
func HasChoices(behavior shared.CardBehavior) bool {
	return len(behavior.Choices) > 0
}

// HasCardDiscardInput checks if a behavior has card-discard type inputs
// These require a pending selection before outputs can be applied
func HasCardDiscardInput(behavior shared.CardBehavior) bool {
	for _, input := range behavior.Inputs {
		if input.ResourceType == shared.ResourceCardDiscard {
			return true
		}
	}
	return false
}

// HasCardDiscardOutput checks if a behavior has card-discard type outputs
// These require a pending selection before remaining outputs can be applied
func HasCardDiscardOutput(behavior shared.CardBehavior) bool {
	for _, output := range behavior.Outputs {
		if output.ResourceType == shared.ResourceCardDiscard {
			return true
		}
	}
	return false
}
