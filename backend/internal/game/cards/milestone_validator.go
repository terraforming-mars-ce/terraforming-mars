package cards

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/milestone"
	"terraforming-mars-backend/internal/game/player"
)

// CalculateMilestoneProgress returns the current progress for a player towards a milestone
func CalculateMilestoneProgress(def *milestone.MilestoneDefinition, p *player.Player, b *board.Board, cardRegistry CardRegistryInterface) int {
	req := def.Requirement
	switch req.Kind {
	case "countable":
		if req.Countable == nil {
			return 0
		}
		return CountPerCondition(&req.Countable.PerCondition, "", p, b, cardRegistry, nil)
	default:
		return 0
	}
}

// CanClaimMilestone checks if a player meets the requirements for a milestone
func CanClaimMilestone(def *milestone.MilestoneDefinition, p *player.Player, b *board.Board, cardRegistry CardRegistryInterface) bool {
	req := def.Requirement
	switch req.Kind {
	case "countable":
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
	default:
		return false
	}
}
