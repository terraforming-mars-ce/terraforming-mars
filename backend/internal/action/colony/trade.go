package colony

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// TradePaymentType represents the resource used to pay for a colony trade
type TradePaymentType string

const (
	TradePaymentCredits  TradePaymentType = "credits"
	TradePaymentEnergy   TradePaymentType = "energy"
	TradePaymentTitanium TradePaymentType = "titanium"
)

const (
	TradeCreditsCost  = 9
	TradeEnergyCost   = 3
	TradeTitaniumCost = 3
)

// TradeAction handles the business logic for trading with a colony tile
type TradeAction struct {
	baseaction.BaseAction
	colonyRegistry colonies.ColonyRegistry
	cardRegistry   cards.CardRegistry
}

// NewTradeAction creates a new trade action
func NewTradeAction(
	gameRepo game.GameRepository,
	colonyRegistry colonies.ColonyRegistry,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *TradeAction {
	return &TradeAction{
		BaseAction:     baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
		colonyRegistry: colonyRegistry,
		cardRegistry:   cardRegistry,
	}
}

// Execute performs the trade action
func (a *TradeAction) Execute(ctx context.Context, gameID string, playerID string, colonyID string, paymentType TradePaymentType) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "colony_trade"),
		zap.String("colony_id", colonyID),
	)
	log.Debug("Trading with colony")

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

	if !g.HasColonies() {
		return fmt.Errorf("colonies expansion is not enabled")
	}

	traderPlayer, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	if !g.GetTradeFleetAvailable(playerID) {
		return fmt.Errorf("trade fleet is not available")
	}

	tileState := g.GetColonyTileState(colonyID)
	if tileState == nil {
		return fmt.Errorf("colony tile not found: %s", colonyID)
	}

	if tileState.TradedThisGen {
		return fmt.Errorf("colony tile already traded this generation")
	}

	resources := traderPlayer.Resources().Get()
	switch paymentType {
	case TradePaymentCredits:
		if resources.Credits < TradeCreditsCost {
			return fmt.Errorf("insufficient credits: need %d, have %d", TradeCreditsCost, resources.Credits)
		}
	case TradePaymentEnergy:
		if resources.Energy < TradeEnergyCost {
			return fmt.Errorf("insufficient energy: need %d, have %d", TradeEnergyCost, resources.Energy)
		}
	case TradePaymentTitanium:
		if resources.Titanium < TradeTitaniumCost {
			return fmt.Errorf("insufficient titanium: need %d, have %d", TradeTitaniumCost, resources.Titanium)
		}
	default:
		return fmt.Errorf("invalid trade payment type: %s", paymentType)
	}

	definition, err := a.colonyRegistry.GetByID(colonyID)
	if err != nil {
		return fmt.Errorf("colony definition not found: %w", err)
	}

	// Deduct trade cost
	switch paymentType {
	case TradePaymentCredits:
		traderPlayer.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: -TradeCreditsCost,
		})
	case TradePaymentEnergy:
		traderPlayer.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceEnergy: -TradeEnergyCost,
		})
	case TradePaymentTitanium:
		traderPlayer.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceTitanium: -TradeTitaniumCost,
		})
	}

	// Collect pending card-targeted resources per player, so same-type resources
	// from trade income + colony bonus are combined into a single selection.
	pendingByPlayer := map[string][]*PendingResource{}
	outputsByPlayer := map[string][]game.CalculatedOutput{}

	// Give trade income based on marker position
	if tileState.MarkerPosition >= 0 && tileState.MarkerPosition < len(definition.Steps) {
		step := definition.Steps[tileState.MarkerPosition]
		for _, output := range step.Outputs {
			if output.Amount > 0 {
				pending := applyOutput(ctx, g, traderPlayer, output.Type, output.Amount, a.cardRegistry, log)
				if pending != nil {
					pendingByPlayer[playerID] = append(pendingByPlayer[playerID], pending)
				}
				outputsByPlayer[playerID] = append(outputsByPlayer[playerID], game.CalculatedOutput{
					ResourceType: output.Type,
					Amount:       output.Amount,
				})
			}
		}
	}

	// Give colony bonus to all players with colonies on this tile
	for _, colonyOwnerID := range tileState.PlayerColonies {
		colonyOwner, ownerErr := g.GetPlayer(colonyOwnerID)
		if ownerErr != nil {
			continue
		}
		for _, bonus := range definition.ColonyBonus {
			if bonus.Amount > 0 {
				pending := applyOutput(ctx, g, colonyOwner, bonus.Type, bonus.Amount, a.cardRegistry, log)
				if pending != nil {
					pendingByPlayer[colonyOwnerID] = append(pendingByPlayer[colonyOwnerID], pending)
				}
				outputsByPlayer[colonyOwnerID] = append(outputsByPlayer[colonyOwnerID], game.CalculatedOutput{
					ResourceType: bonus.Type,
					Amount:       bonus.Amount,
				})
			}
		}
	}

	// Resolve pending resources — combine same-type for each player
	for pid, pendings := range pendingByPlayer {
		p, pErr := g.GetPlayer(pid)
		if pErr != nil {
			continue
		}
		reason := "colony-tax"
		if pid == playerID {
			reason = "trade"
		}
		for _, combined := range combinePendingResources(pendings) {
			setPendingColonyResource(p, combined, definition.Name, colonyID, reason, a.cardRegistry, log)
		}
	}

	// Add triggered effects for trader and colony bonus recipients
	for pid, outputs := range outputsByPlayer {
		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:          "Trade: " + definition.Name,
			PlayerID:          pid,
			SourceType:        game.SourceTypeColonyTrade,
			CalculatedOutputs: combineCalculatedOutputs(outputs),
		})
	}

	a.WriteStateLogFull(ctx, g, "Trade: "+definition.Name, game.SourceTypeColonyTrade,
		playerID, fmt.Sprintf("Traded with %s", definition.Name), nil, combineCalculatedOutputs(outputsByPlayer[playerID]), nil)

	// Reset marker to position after last colony
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

	a.ConsumePlayerAction(g, log)

	log.Info("Colony traded",
		zap.String("colony_id", colonyID),
		zap.Int("marker_position", tileState.MarkerPosition))

	return nil
}

// setPendingColonyResource sets a pending colony resource selection on a player
// if they have at least one card that can store the resource type.
func setPendingColonyResource(p *player.Player, pending *PendingResource, colonyName string, colonyID string, reason string, cardRegistry cards.CardRegistry, log *zap.Logger) {
	if !hasEligibleStorageCard(p, pending.ResourceType, cardRegistry) {
		log.Debug("No eligible storage card, resources lost",
			zap.String("player_id", p.ID()),
			zap.String("resource_type", pending.ResourceType),
			zap.Int("amount", pending.Amount))
		return
	}

	p.Selection().SetPendingColonyResourceSelection(&player.PendingColonyResourceSelection{
		ResourceType: pending.ResourceType,
		Amount:       pending.Amount,
		Source:       colonyName,
		ColonyID:     colonyID,
		Reason:       reason,
	})

	log.Debug("Set pending colony resource selection",
		zap.String("player_id", p.ID()),
		zap.String("resource_type", pending.ResourceType),
		zap.Int("amount", pending.Amount))
}

// hasEligibleStorageCard checks if a player has any played card that can store the given resource type.
func hasEligibleStorageCard(p *player.Player, resourceType string, cardRegistry cards.CardRegistry) bool {
	if cardRegistry == nil {
		return false
	}
	rt := shared.ResourceType(resourceType)

	// Check played cards
	for _, cardID := range p.PlayedCards().Cards() {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}
		if card.ResourceStorage != nil && card.ResourceStorage.Type == rt {
			return true
		}
	}

	// Check corporation
	if corpID := p.CorporationID(); corpID != "" {
		corp, err := cardRegistry.GetByID(corpID)
		if err == nil {
			if corp.ResourceStorage != nil && corp.ResourceStorage.Type == rt {
				return true
			}
		}
	}

	return false
}
