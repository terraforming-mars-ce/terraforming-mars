package confirmation

import (
	"context"
	"fmt"
	"slices"
	"time"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	colonyaction "terraforming-mars-backend/internal/action/colony"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmFreeTradeAction handles confirming a free trade from a card effect
type ConfirmFreeTradeAction struct {
	baseaction.BaseAction
	colonyRegistry colonies.ColonyRegistry
}

// NewConfirmFreeTradeAction creates a new confirm free trade action
func NewConfirmFreeTradeAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	colonyRegistry colonies.ColonyRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ConfirmFreeTradeAction {
	return &ConfirmFreeTradeAction{
		BaseAction:     baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
		colonyRegistry: colonyRegistry,
	}
}

// Execute performs the free trade with the selected colony
func (a *ConfirmFreeTradeAction) Execute(ctx context.Context, gameID string, playerID string, colonyID string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_free_trade"),
		zap.String("colony_id", colonyID),
	)
	log.Debug("Confirming free trade")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	pendingSelection := p.Selection().GetPendingFreeTradeSelection()
	if pendingSelection == nil {
		return fmt.Errorf("no pending free trade selection")
	}

	if !slices.Contains(pendingSelection.AvailableColonyIDs, colonyID) {
		return fmt.Errorf("colony %s is not in the available list", colonyID)
	}

	tileState := g.GetColonyTileState(colonyID)
	if tileState == nil {
		return fmt.Errorf("colony tile not found: %s", colonyID)
	}

	if tileState.TradedThisGen {
		return fmt.Errorf("colony tile already traded this generation")
	}

	if !g.GetTradeFleetAvailable(playerID) {
		return fmt.Errorf("trade fleet is not available")
	}

	definition, err := a.colonyRegistry.GetByID(colonyID)
	if err != nil {
		return fmt.Errorf("colony definition not found: %w", err)
	}

	// Apply trade step bonus from cards like Trade Envoys
	tradeStepBonus := colonyaction.CountTradeStepBonus(p, a.CardRegistry())
	if tradeStepBonus > 0 {
		maxStep := len(definition.Steps) - 1
		newPosition := tileState.MarkerPosition + tradeStepBonus
		if newPosition > maxStep {
			newPosition = maxStep
		}
		tileState.MarkerPosition = newPosition
	}

	// Apply trade income based on marker position (no payment needed - it's free!)
	if tileState.MarkerPosition >= 0 && tileState.MarkerPosition < len(definition.Steps) {
		step := definition.Steps[tileState.MarkerPosition]
		for _, output := range step.Outputs {
			if output.Amount > 0 {
				colonyaction.ApplyTradeOutput(ctx, g, p, output.Type, output.Amount, a.CardRegistry(), log)
			}
		}
	}

	// Apply colony bonus to all players with colonies
	for _, colonyOwnerID := range tileState.PlayerColonies {
		colonyOwner, ownerErr := g.GetPlayer(colonyOwnerID)
		if ownerErr != nil {
			continue
		}
		for _, bonus := range definition.ColonyBonus {
			if bonus.Amount > 0 {
				colonyaction.ApplyTradeOutput(ctx, g, colonyOwner, bonus.Type, bonus.Amount, a.CardRegistry(), log)
			}
		}
	}

	// Reset marker and mark as traded
	tileState.MarkerPosition = len(tileState.PlayerColonies)
	tileState.TradedThisGen = true
	tileState.TraderID = playerID

	g.SetTradeFleetAvailable(playerID, false)

	events.Publish(g.EventBus(), events.ColonyTradedEvent{
		GameID:    g.ID(),
		PlayerID:  playerID,
		ColonyID:  colonyID,
		Timestamp: time.Now(),
	})

	// Clear the pending selection
	p.Selection().SetPendingFreeTradeSelection(nil)

	a.WriteStateLogFull(ctx, g, "Free Trade: "+definition.Name, shared.SourceTypeColonyTrade,
		playerID, fmt.Sprintf("Free traded with %s", definition.Name), nil, nil, nil)

	log.Info("Free trade confirmed",
		zap.String("colony_id", colonyID))

	return nil
}
