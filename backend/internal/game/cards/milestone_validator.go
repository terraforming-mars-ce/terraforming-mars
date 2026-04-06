package cards

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/milestone"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CalculateMilestoneProgress returns the current progress for a player towards a milestone
func CalculateMilestoneProgress(def *milestone.MilestoneDefinition, p *player.Player, b *board.Board, cardRegistry CardRegistryInterface) int {
	req := def.Requirement
	switch req.Kind {
	case milestone.RequirementKindCountable:
		if req.Countable == nil {
			return 0
		}
		return CountPerCondition(&req.Countable.PerCondition, "", p, b, cardRegistry, nil)
	case milestone.RequirementKindState:
		if req.State == nil {
			return 0
		}
		return countProductionsAtOrAbove(p, req.State.Min)
	default:
		return 0
	}
}

// CanClaimMilestone checks if a player meets the requirements for a milestone
func CanClaimMilestone(def *milestone.MilestoneDefinition, p *player.Player, b *board.Board, cardRegistry CardRegistryInterface) bool {
	req := def.Requirement
	switch req.Kind {
	case milestone.RequirementKindCountable:
		if req.Countable == nil {
			return false
		}
		progress := CountPerCondition(&req.Countable.PerCondition, "", p, b, cardRegistry, nil)
		if req.Countable.Min != nil && progress < *req.Countable.Min {
			return false
		}
		if req.Countable.Max != nil && progress > *req.Countable.Max {
			return false
		}
		return true
	case milestone.RequirementKindState:
		if req.State == nil {
			return false
		}
		return countProductionsAtOrAbove(p, req.State.Min) == 6
	default:
		return false
	}
}

func countProductionsAtOrAbove(p *player.Player, min int) int {
	prod := p.Resources().Production()
	count := 0
	productions := []shared.ResourceType{
		shared.ResourceCreditProduction,
		shared.ResourceSteelProduction,
		shared.ResourceTitaniumProduction,
		shared.ResourcePlantProduction,
		shared.ResourceEnergyProduction,
		shared.ResourceHeatProduction,
	}
	for _, rt := range productions {
		if prod.GetAmount(rt) >= min {
			count++
		}
	}
	return count
}
