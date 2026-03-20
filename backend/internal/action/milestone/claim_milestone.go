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
	"terraforming-mars-backend/internal/milestones"
)

// ClaimMilestoneAction handles the business logic for claiming a milestone
type ClaimMilestoneAction struct {
	baseaction.BaseAction
	milestoneRegistry milestones.MilestoneRegistry
}

// NewClaimMilestoneAction creates a new claim milestone action
func NewClaimMilestoneAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	milestoneRegistry milestones.MilestoneRegistry,
	logger *zap.Logger,
) *ClaimMilestoneAction {
	return &ClaimMilestoneAction{
		BaseAction:        baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
		milestoneRegistry: milestoneRegistry,
	}
}

// Execute claims a milestone for the player
func (a *ClaimMilestoneAction) Execute(ctx context.Context, gameID string, playerID string, milestoneType string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "claim_milestone"), zap.String("milestone", milestoneType))
	log.Debug("Claiming milestone")

	def, err := a.milestoneRegistry.GetByID(milestoneType)
	if err != nil {
		log.Warn("Invalid milestone type", zap.String("milestone_type", milestoneType))
		return fmt.Errorf("invalid milestone type: %s", milestoneType)
	}

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, shared.GamePhaseAction, log); err != nil {
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

	ms := g.Milestones()
	mt := shared.MilestoneType(milestoneType)
	if ms.IsClaimed(mt) {
		log.Warn("Milestone already claimed", zap.String("milestone", milestoneType))
		return fmt.Errorf("milestone %s is already claimed", milestoneType)
	}

	if !ms.CanClaimMore() {
		log.Warn("Maximum milestones already claimed", zap.Int("max", game.MaxClaimedMilestones))
		return fmt.Errorf("maximum milestones (%d) already claimed", game.MaxClaimedMilestones)
	}

	resources := player.Resources().Get()
	if resources.Credits < def.ClaimCost {
		log.Warn("Insufficient credits for milestone",
			zap.Int("cost", def.ClaimCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", def.ClaimCost, resources.Credits)
	}

	if !gamecards.CanClaimMilestone(def, player, g.Board(), a.CardRegistry()) {
		progress := gamecards.CalculateMilestoneProgress(def, player, g.Board(), a.CardRegistry())
		required := def.GetRequired()
		log.Warn("Player does not meet milestone requirements",
			zap.String("requirement", def.Description),
			zap.Int("required", required),
			zap.Int("current", progress))
		return fmt.Errorf("requirements not met: %s (have %d, need %d)", def.Description, progress, required)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -def.ClaimCost,
	})
	log.Debug("Deducted milestone cost",
		zap.Int("cost", def.ClaimCost),
		zap.Int("remaining_credits", player.Resources().Get().Credits))

	if err := ms.ClaimMilestone(ctx, mt, playerID, g.Generation()); err != nil {
		log.Error("Failed to claim milestone", zap.Error(err))
		return fmt.Errorf("failed to claim milestone: %w", err)
	}

	a.ConsumePlayerAction(g, log)

	a.WriteStateLog(ctx, g, def.Name, shared.SourceTypeMilestone, playerID, fmt.Sprintf("Claimed %s milestone", def.Name))

	log.Info("Milestone claimed",
		zap.String("milestone", milestoneType),
		zap.Int("total_claimed", ms.ClaimedCount()))

	return nil
}
