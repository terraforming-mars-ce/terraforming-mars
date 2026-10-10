package shared

import "encoding/json"

// UnmarshalJSON restores the concrete condition types in a standalone choice.
func (c *Choice) UnmarshalJSON(data []byte) error {
	var raw choiceJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.Inputs = resourceConditionsToInterface(raw.Inputs)
	c.Outputs = resourceConditionsToInterface(raw.Outputs)
	c.Requirements = raw.Requirements
	return nil
}

// UnmarshalJSON restores deferred outputs without applying them.
func (p *PendingBehaviorResolution) UnmarshalJSON(data []byte) error {
	type plain PendingBehaviorResolution
	var raw struct {
		*plain
		PendingOutputs []resourceConditionJSON
	}
	raw.plain = (*plain)(p)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.PendingOutputs = resourceConditionsToInterface(raw.PendingOutputs)
	return nil
}

// UnmarshalJSON restores deferred effect outputs.
func (p *PendingEffectSelection) UnmarshalJSON(data []byte) error {
	type plain PendingEffectSelection
	var raw struct {
		*plain
		Outputs []resourceConditionJSON
	}
	raw.plain = (*plain)(p)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Outputs = resourceConditionsToInterface(raw.Outputs)
	return nil
}

// UnmarshalJSON restores the concrete outputs of a target assignment.
func (p *EffectSelectionOption) UnmarshalJSON(data []byte) error {
	type plain EffectSelectionOption
	var raw struct {
		*plain
		Outputs []resourceConditionJSON
	}
	raw.plain = (*plain)(p)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Outputs = resourceConditionsToInterface(raw.Outputs)
	return nil
}
