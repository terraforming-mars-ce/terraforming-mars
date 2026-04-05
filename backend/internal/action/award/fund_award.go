package award

import (
	"context"
	"fmt"
	"slices"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// FundAwardAction handles the business logic for funding an award
type FundAwardAction struct {
	baseaction.BaseAction
	awardRegistry awards.AwardRegistry
}

// NewFundAwardAction creates a new fund award action
func NewFundAwardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	awardRegistry awards.AwardRegistry,
	logger *zap.Logger,
) *FundAwardAction {
	return &FundAwardAction{
		BaseAction:    baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
		awardRegistry: awardRegistry,
	}
}

// Execute funds an award for the player
func (a *FundAwardAction) Execute(ctx context.Context, gameID string, playerID string, awardType string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "fund_award"), zap.String("award", awardType))
	log.Debug("Funding award")

	def, err := a.awardRegistry.GetByID(awardType)
	if err != nil {
		log.Warn("Invalid award type", zap.String("award_type", awardType))
		return fmt.Errorf("invalid award type: %s", awardType)
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

	// Validate award is in the selected set for this game
	if selected := g.SelectedAwards(); len(selected) > 0 && !slices.Contains(selected, awardType) {
		log.Warn("Award not available in this game", zap.String("award", awardType))
		return fmt.Errorf("award %s is not available in this game", awardType)
	}

	awardState := g.Awards()
	at := shared.AwardType(awardType)
	if awardState.IsFunded(at) {
		log.Warn("Award already funded", zap.String("award", awardType))
		return fmt.Errorf("award %s is already funded", awardType)
	}

	if !awardState.CanFundMore() {
		log.Warn("Maximum awards already funded", zap.Int("max", game.MaxFundedAwards))
		return fmt.Errorf("maximum awards (%d) already funded", game.MaxFundedAwards)
	}

	fundingCost := def.GetCostForFundedCount(awardState.FundedCount())
	resources := player.Resources().Get()
	if resources.Credits < fundingCost {
		log.Warn("Insufficient credits for award",
			zap.Int("cost", fundingCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", fundingCost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -fundingCost,
	})
	log.Debug("Deducted award funding cost",
		zap.Int("cost", fundingCost),
		zap.Int("remaining_credits", player.Resources().Get().Credits))

	if err := awardState.FundAward(ctx, at, playerID, fundingCost); err != nil {
		log.Error("Failed to fund award", zap.Error(err))
		return fmt.Errorf("failed to fund award: %w", err)
	}

	a.ConsumePlayerAction(g, log)

	a.WriteStateLog(ctx, g, def.Name, shared.SourceTypeAward, playerID, fmt.Sprintf("Funded %s award", def.Name))

	log.Info("Award funded",
		zap.String("award", awardType),
		zap.Int("total_funded", awardState.FundedCount()))

	return nil
}
