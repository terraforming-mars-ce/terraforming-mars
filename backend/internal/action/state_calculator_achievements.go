package action

import (
	"fmt"
	"openmars/internal/game"
	"openmars/internal/game/award"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/milestone"
	"openmars/internal/game/player"
	"openmars/internal/game/shared"
	"time"
)

// CalculateMilestoneState computes eligibility state for claiming a milestone.
// Returns EntityState with errors indicating why the milestone cannot be claimed.
func CalculateMilestoneState(
	milestoneType shared.MilestoneType,
	p *player.Player,
	g *game.Game,
	cardRegistry gamecards.CardRegistry,
	milestoneRegistry milestone.MilestoneRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]interface{})

	def, err := milestoneRegistry.GetByID(string(milestoneType))
	if err != nil {
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

	ms := g.Milestones()

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	if ms.IsClaimed(milestoneType) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMilestoneAlreadyClaimed,
			Category: player.ErrorCategoryAchievement,
			Message:  "Already claimed",
		})
	}

	if ms.ClaimedCount() >= game.MaxClaimedMilestones {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMaxMilestonesClaimed,
			Category: player.ErrorCategoryAchievement,
			Message:  "Maximum milestones claimed",
		})
	}

	progress := gamecards.CalculateMilestoneProgress(def, p, g.Board(), cardRegistry)
	required := def.GetRequired()
	metadata["progress"] = progress
	metadata["required"] = required

	if progress < required {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMilestoneRequirementNotMet,
			Category: player.ErrorCategoryRequirement,
			Message:  "Requirement not met",
		})
	}

	cost := def.ClaimCost
	errors = append(errors, paymentAffordability(p, g, cardRegistry, gamecards.PaymentContext{Costs: map[shared.ResourceType]int{shared.ResourceCredit: cost}, Action: "claim-milestone"})...)

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
	awardRegistry award.AwardRegistry,
	cardRegistry gamecards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]interface{})

	def, err := awardRegistry.GetByID(string(awardType))
	if err != nil {
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

	gameAwards := g.Awards()

	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	if gameAwards.IsFunded(awardType) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeAwardAlreadyFunded,
			Category: player.ErrorCategoryAchievement,
			Message:  "Already funded",
		})
	}

	if gameAwards.FundedCount() >= game.MaxFundedAwards {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeMaxAwardsFunded,
			Category: player.ErrorCategoryAchievement,
			Message:  "Maximum awards funded",
		})
	}

	cost := def.GetCostForFundedCount(gameAwards.FundedCount())
	metadata["fundingCost"] = cost

	errors = append(errors, paymentAffordability(p, g, cardRegistry, gamecards.PaymentContext{Costs: map[shared.ResourceType]int{shared.ResourceCredit: cost}, Action: "fund-award"})...)

	costMap := map[string]int{string(shared.ResourceCredit): cost}

	return player.EntityState{
		Errors:         errors,
		Cost:           costMap,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}
