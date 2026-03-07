package milestone

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// ClaimMilestoneAction handles the business logic for claiming a milestone
type ClaimMilestoneAction struct {
	baseaction.BaseAction
}

// NewClaimMilestoneAction creates a new claim milestone action
func NewClaimMilestoneAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ClaimMilestoneAction {
	return &ClaimMilestoneAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// Execute claims a milestone for the player
func (a *ClaimMilestoneAction) Execute(ctx context.Context, gameID string, playerID string, milestoneType string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "claim_milestone"), zap.String("milestone", milestoneType))
	log.Debug("Claiming milestone")

	if !shared.ValidMilestoneType(milestoneType) {
		log.Warn("Invalid milestone type", zap.String("milestone_type", milestoneType))
		return fmt.Errorf("invalid milestone type: %s", milestoneType)
	}

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateActionsRemaining(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	milestones := g.Milestones()
	mt := shared.MilestoneType(milestoneType)
	if milestones.IsClaimed(mt) {
		log.Warn("Milestone already claimed", zap.String("milestone", milestoneType))
		return fmt.Errorf("milestone %s is already claimed", milestoneType)
	}

	if !milestones.CanClaimMore() {
		log.Warn("Maximum milestones already claimed", zap.Int("max", game.MaxClaimedMilestones))
		return fmt.Errorf("maximum milestones (%d) already claimed", game.MaxClaimedMilestones)
	}

	resources := player.Resources().Get()
	if resources.Credits < game.MilestoneClaimCost {
		log.Warn("Insufficient credits for milestone",
			zap.Int("cost", game.MilestoneClaimCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", game.MilestoneClaimCost, resources.Credits)
	}

	if !gamecards.CanClaimMilestone(mt, player, g.Board(), a.CardRegistry()) {
		requirement := gamecards.GetMilestoneRequirement(mt)
		progress := gamecards.GetPlayerMilestoneProgress(mt, player, g.Board(), a.CardRegistry())
		log.Warn("Player does not meet milestone requirements",
			zap.String("requirement", requirement.Description),
			zap.Int("required", requirement.Required),
			zap.Int("current", progress))
		return fmt.Errorf("requirements not met: %s (have %d, need %d)", requirement.Description, progress, requirement.Required)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -game.MilestoneClaimCost,
	})
	log.Debug("Deducted milestone cost",
		zap.Int("cost", game.MilestoneClaimCost),
		zap.Int("remaining_credits", player.Resources().Get().Credits))

	if err := milestones.ClaimMilestone(ctx, mt, playerID, g.Generation()); err != nil {
		log.Error("Failed to claim milestone", zap.Error(err))
		return fmt.Errorf("failed to claim milestone: %w", err)
	}

	a.ConsumePlayerAction(g, log)

	a.WriteStateLog(ctx, g, milestoneType, game.SourceTypeMilestone, playerID, fmt.Sprintf("Claimed %s milestone", milestoneType))

	log.Info("Milestone claimed",
		zap.String("milestone", milestoneType),
		zap.Int("total_claimed", milestones.ClaimedCount()))

	return nil
}
