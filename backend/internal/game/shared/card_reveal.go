package shared

import (
	"encoding/json"
	"fmt"
)

// CardRevealCondition publicly reveals project cards and resolves matching rewards.
type CardRevealCondition struct {
	ConditionBase
	Destination string       `json:"destination"`
	OnMatch     *RevealMatch `json:"onMatch,omitempty"`
}

// RevealMatch applies Outputs once for each revealed card matching Selectors.
type RevealMatch struct {
	Selectors []Selector          `json:"selectors"`
	Outputs   []BehaviorCondition `json:"outputs"`
}

func (m *RevealMatch) UnmarshalJSON(data []byte) error {
	var raw struct {
		Selectors []Selector              `json:"selectors"`
		Outputs   []resourceConditionJSON `json:"outputs"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.Selectors = raw.Selectors
	m.Outputs = resourceConditionsToInterface(raw.Outputs)
	return nil
}
func (c *CardRevealCondition) isBehaviorCondition() {}
func (c *CardRevealCondition) deepCopyCondition() BehaviorCondition {
	result := *c
	if c.OnMatch != nil {
		result.OnMatch = &RevealMatch{Selectors: CloneSelectors(c.OnMatch.Selectors)}
		for _, output := range c.OnMatch.Outputs {
			result.OnMatch.Outputs = append(result.OnMatch.Outputs, CloneCondition(output))
		}
	}
	return &result
}

// RevealedCard records public information, independently of the private draw pile.
type RevealedCard struct {
	CardID  string `json:"cardId"`
	Name    string `json:"name"`
	Matched bool   `json:"matched"`
}

// PendingCardReveal acknowledges an already resolved reveal; it never reapplies outputs.
type PendingCardReveal struct {
	Source       string
	SourceCardID string
	Cards        []RevealedCard
	Rewards      []CalculatedOutput
}

// IsImmediateRevealReward excludes outcomes requiring a target, choice, or another deferred action.
func IsImmediateRevealReward(output BehaviorCondition) bool {
	if per := GetPerCondition(output); per != nil && per.Amount <= 0 {
		return false
	}
	if output.GetAmount() < 0 || IsVariableAmount(output) || IsOptional(output) {
		return false
	}
	switch c := output.(type) {
	case *BasicResourceCondition:
		return c.Target == "self-player"
	case *ProductionCondition:
		return c.Target == "self-player" && c.ResourceType != ResourceAnyProduction
	case *CardStorageCondition:
		return c.Target == "self-card" && c.ResourceType != ResourceCardResource
	default:
		return false
	}
}

func (rc *resourceConditionJSON) UnmarshalJSON(data []byte) error {
	type plain resourceConditionJSON
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.ResourceType == ResourceType("storage-payment-substitute") {
		return fmt.Errorf("unknown resource type %q", decoded.ResourceType)
	}
	if decoded.ResourceType != ResourcePaymentSubstitute && (decoded.Source != nil || decoded.TargetResource != "") {
		return fmt.Errorf("source and targetResource require payment-substitute")
	}
	if decoded.ResourceType == ResourcePaymentSubstitute {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		for key := range fields {
			switch key {
			case "type", "target", "amount", "source", "targetResource", "selectors":
			default:
				return fmt.Errorf("field %s is not valid on payment-substitute", key)
			}
		}
		if decoded.Source == nil {
			return fmt.Errorf("payment-substitute requires source")
		}
		var sourceFields map[string]json.RawMessage
		if err := json.Unmarshal(fields["source"], &sourceFields); err != nil {
			return err
		}
		for key := range sourceFields {
			if key != "target" && key != "resource" {
				return fmt.Errorf("field %s is not valid on payment source", key)
			}
		}
		c := PaymentSubstituteCondition{ConditionBase: ConditionBase{ResourceType: decoded.ResourceType, Amount: decoded.Amount, Target: decoded.Target}, Source: *decoded.Source, TargetResource: decoded.TargetResource, Selectors: decoded.Selectors}
		if errors := ValidateResourceCondition(&c, false); len(errors) > 0 {
			return fmt.Errorf("invalid payment-substitute: %v", errors)
		}
	}
	if decoded.ResourceType != ResourceDefense && decoded.Against != "" {
		return fmt.Errorf("against is only valid on defense")
	}
	if decoded.ResourceType == ResourceDefense && decoded.Against != "any-player" && decoded.Against != "opponents" {
		return fmt.Errorf("defense requires against: any-player or opponents")
	}
	if decoded.ResourceType != ResourceCardReveal && (decoded.Destination != "" || decoded.OnMatch != nil) {
		return fmt.Errorf("destination and onMatch are only valid on card-reveal")
	}
	if decoded.ResourceType == ResourceCardReveal {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		for key := range fields {
			switch key {
			case "type", "amount", "target", "destination", "onMatch":
			default:
				return fmt.Errorf("field %s is not valid on card-reveal", key)
			}
		}
	}
	*rc = resourceConditionJSON(decoded)
	return nil
}
