package award

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// FundAwardAction handles the business logic for funding an award
type FundAwardAction struct {
	baseaction.BaseAction
}

// NewFundAwardAction creates a new fund award action
func NewFundAwardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *FundAwardAction {
	return &FundAwardAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// Execute funds an award for the player
func (a *FundAwardAction) Execute(ctx context.Context, gameID string, playerID string, awardType string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "fund_award"), zap.String("award", awardType))
	log.Debug("Funding award")

	if !shared.ValidAwardType(awardType) {
		log.Warn("Invalid award type", zap.String("award_type", awardType))
		return fmt.Errorf("invalid award type: %s", awardType)
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

	awards := g.Awards()
	at := shared.AwardType(awardType)
	if awards.IsFunded(at) {
		log.Warn("Award already funded", zap.String("award", awardType))
		return fmt.Errorf("award %s is already funded", awardType)
	}

	if !awards.CanFundMore() {
		log.Warn("Maximum awards already funded", zap.Int("max", game.MaxFundedAwards))
		return fmt.Errorf("maximum awards (%d) already funded", game.MaxFundedAwards)
	}

	fundingCost := awards.GetCurrentFundingCost()
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

	if err := awards.FundAward(ctx, at, playerID); err != nil {
		log.Error("Failed to fund award", zap.Error(err))
		return fmt.Errorf("failed to fund award: %w", err)
	}

	a.ConsumePlayerAction(g, log)

	awardName := awardType
	if info, ok := game.GetAwardInfo(shared.AwardType(awardType)); ok {
		awardName = info.Name
	}
	a.WriteStateLog(ctx, g, awardName, game.SourceTypeAward, playerID, fmt.Sprintf("Funded %s award", awardName))

	log.Info("Award funded",
		zap.String("award", awardType),
		zap.Int("total_funded", awards.FundedCount()))

	return nil
}
