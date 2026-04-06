package action

import (
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// standardProjectDisplayData contains pre-built display data for resource conversion projects
var standardProjectDisplayData = map[string]*game.LogDisplayData{
	"Convert Heat": {
		Behaviors: []shared.CardBehavior{{
			Outputs: []shared.BehaviorCondition{&shared.GlobalParameterCondition{
				ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTemperature, Amount: 1, Target: "global"},
			}},
		}},
	},
	"Convert Plants": {
		Behaviors: []shared.CardBehavior{{
			Outputs: []shared.BehaviorCondition{&shared.TilePlacementCondition{
				ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceGreeneryPlacement, Amount: 1, Target: "global"},
			}},
		}},
	},
}

// GetStandardProjectDisplayData returns display data for a standard project
func GetStandardProjectDisplayData(source string) *game.LogDisplayData {
	return standardProjectDisplayData[source]
}

// BuildCardDisplayData builds display data for a card log entry
func BuildCardDisplayData(card *gamecards.Card, sourceType shared.SourceType) *game.LogDisplayData {
	if card == nil {
		return nil
	}

	data := &game.LogDisplayData{
		Tags: card.Tags,
	}

	// Convert VP conditions
	for _, vp := range card.VPConditions {
		vpForLog := shared.VPConditionForLog{
			Amount:    vp.Amount,
			Condition: string(vp.Condition),
		}
		if vp.MaxTrigger != nil {
			vpForLog.MaxTrigger = vp.MaxTrigger
		}
		if vp.Per != nil {
			vpForLog.Per = &shared.PerCondition{
				ResourceType: vp.Per.ResourceType,
				Amount:       vp.Per.Amount,
			}
			if vp.Per.Location != nil {
				loc := string(*vp.Per.Location)
				vpForLog.Per.Location = &loc
			}
			if vp.Per.Target != nil {
				target := string(*vp.Per.Target)
				vpForLog.Per.Target = &target
			}
			if vp.Per.Tag != nil {
				vpForLog.Per.Tag = vp.Per.Tag
			}
		}
		data.VPConditions = append(data.VPConditions, vpForLog)
	}

	// Select appropriate behaviors based on source type
	switch sourceType {
	case shared.SourceTypeCardPlay:
		data.Behaviors = card.Behaviors
	case shared.SourceTypeCardAction:
		// Only include manual trigger behaviors for card actions
		for _, b := range card.Behaviors {
			if gamecards.HasManualTrigger(b) {
				data.Behaviors = append(data.Behaviors, b)
			}
		}
	}

	return data
}
