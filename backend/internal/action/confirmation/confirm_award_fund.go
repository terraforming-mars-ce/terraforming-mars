package confirmation

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

// ConfirmAwardFundAction handles confirming a free award fund selection (e.g., Vitor)
type ConfirmAwardFundAction struct {
	baseaction.BaseAction
	awardRegistry awards.AwardRegistry
}

// NewConfirmAwardFundAction creates a new confirm award fund action
func NewConfirmAwardFundAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	awardRegistry awards.AwardRegistry,
	logger *zap.Logger,
) *ConfirmAwardFundAction {
	return &ConfirmAwardFundAction{
		BaseAction:    baseaction.NewBaseAction(gameRepo, cardRegistry),
		awardRegistry: awardRegistry,
	}
}

// Execute funds the selected award for free and clears the pending selection
func (a *ConfirmAwardFundAction) Execute(ctx context.Context, gameID string, playerID string, awardType string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_award_fund"),
		zap.String("award_type", awardType),
	)
	log.Debug("Confirming award fund selection")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	pending := p.Selection().GetPendingAwardFundSelection()
	if pending == nil {
		return fmt.Errorf("no pending award fund selection")
	}

	if !slices.Contains(pending.AvailableAwards, awardType) {
		return fmt.Errorf("award %s is not available for selection", awardType)
	}

	def, err := a.awardRegistry.GetByID(awardType)
	if err != nil {
		return fmt.Errorf("invalid award type: %s", awardType)
	}

	if err := g.Awards().FundAward(ctx, shared.AwardType(awardType), playerID, 0); err != nil {
		return fmt.Errorf("failed to fund award: %w", err)
	}

	p.Selection().SetPendingAwardFundSelection(nil)

	if err := g.SetForcedFirstAction(ctx, playerID, nil); err != nil {
		return fmt.Errorf("failed to clear forced first action: %w", err)
	}

	log.Info("Award funded for free",
		zap.String("award", def.Name))

	return nil
}
