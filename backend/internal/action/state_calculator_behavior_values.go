package action

import (
	"fmt"
	"openmars/internal/game"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// computeBehaviorValues computes per-condition output values for card behaviors.
// Returns a slice of ComputedBehaviorValue with target format "behaviors::N".
func computeBehaviorValues(
	behaviors []shared.CardBehavior,
	sourceCardID string,
	p *player.Player,
	g *game.Game,
	cardRegistry gamecards.CardRegistry,
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
			count := gamecards.CountPerCondition(per, sourceCardID, p, board, cardRegistry, allPlayers, g.Colonies(), gamecards.TagCountContext{ActorID: p.ID()})
			if per.Amount > 0 {
				actualAmount := shared.CalculateScaledAmount(outputBC, count)
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
				count := gamecards.CountPerCondition(per, sourceCardID, p, board, cardRegistry, allPlayers, g.Colonies(), gamecards.TagCountContext{ActorID: p.ID()})
				if per.Amount > 0 {
					actualAmount := shared.CalculateScaledAmount(outputBC, count)
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
