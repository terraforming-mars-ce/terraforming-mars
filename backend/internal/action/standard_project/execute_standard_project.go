package standard_project

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/standardprojects"
)

// ExecuteStandardProjectAction handles all standard projects via a single unified action
type ExecuteStandardProjectAction struct {
	baseaction.BaseAction
	standardProjectRegistry standardprojects.StandardProjectRegistry
}

// NewExecuteStandardProjectAction creates a new unified standard project action
func NewExecuteStandardProjectAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	standardProjectRegistry standardprojects.StandardProjectRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ExecuteStandardProjectAction {
	return &ExecuteStandardProjectAction{
		BaseAction:              baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
		standardProjectRegistry: standardProjectRegistry,
	}
}

// Execute performs the standard project action
func (a *ExecuteStandardProjectAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	projectID string,
) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("project_id", projectID))
	log.Debug("Executing standard project")

	definition, err := a.standardProjectRegistry.GetByID(projectID)
	if err != nil {
		log.Warn("Unknown standard project", zap.String("project_id", projectID))
		return fmt.Errorf("unknown standard project: %s", projectID)
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

	if err := baseaction.ValidateNoPendingSelections(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	effectiveCost := definition.CreditCost()
	if a.CardRegistry() != nil && effectiveCost > 0 {
		calculator := gamecards.NewRequirementModifierCalculator(a.CardRegistry())
		discounts := calculator.CalculateStandardProjectDiscounts(player, shared.StandardProject(projectID))
		creditDiscount := discounts[shared.ResourceCredit]
		effectiveCost = definition.CreditCost() - creditDiscount
		if effectiveCost < 0 {
			effectiveCost = 0
		}
		if creditDiscount > 0 {
			log.Debug("Applied standard project discount",
				zap.Int("base_cost", definition.CreditCost()),
				zap.Int("discount", creditDiscount),
				zap.Int("effective_cost", effectiveCost))
		}
	}

	if effectiveCost > 0 {
		resources := player.Resources().Get()
		if resources.Credits < effectiveCost {
			log.Warn("Insufficient credits",
				zap.Int("cost", effectiveCost),
				zap.Int("player_credits", resources.Credits))
			return fmt.Errorf("insufficient credits: need %d, have %d", effectiveCost, resources.Credits)
		}

		player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: -effectiveCost,
		})
	}

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: projectID,
		ProjectCost: definition.CreditCost(),
		Timestamp:   time.Now(),
	})

	// Check for sell-patents pattern: card-discard input with variableAmount
	if hasSellPatentsBehavior(definition.Behaviors) {
		return a.executeSellPatents(g, player, log)
	}

	// Apply behavior outputs via BehaviorApplier
	applier := gamecards.NewBehaviorApplier(player, g, definition.Name, log).
		WithSourceType(shared.SourceTypeStandardProject)
	if a.CardRegistry() != nil {
		applier = applier.WithCardRegistry(a.CardRegistry())
	}

	var allCalculatedOutputs []shared.CalculatedOutput
	for _, behavior := range definition.Behaviors {
		calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, behavior.Outputs)
		if err != nil {
			return fmt.Errorf("failed to apply standard project outputs: %w", err)
		}
		allCalculatedOutputs = append(allCalculatedOutputs, calculatedOutputs...)
	}

	a.ConsumePlayerAction(g, log)

	displayData := &game.LogDisplayData{}
	source := "Standard Project: " + definition.Name
	a.WriteStateLogFull(ctx, g, source, shared.SourceTypeStandardProject, playerID, definition.Name, nil, allCalculatedOutputs, displayData)

	log.Info("Standard project executed", zap.String("project", definition.Name))
	return nil
}

// executeSellPatents handles the sell patents special case by creating a pending card selection
func (a *ExecuteStandardProjectAction) executeSellPatents(
	g *game.Game,
	p *player.Player,
	log *zap.Logger,
) error {
	_ = g
	playerCards := p.Hand().Cards()
	if len(playerCards) == 0 {
		log.Warn("No cards to sell")
		return fmt.Errorf("no cards available to sell")
	}

	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range playerCards {
		cardCosts[cardID] = 0
		cardRewards[cardID] = 1
	}

	p.Selection().SetPendingCardSelection(&shared.PendingCardSelection{
		Source:         "sell-patents",
		AvailableCards: playerCards,
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		MinCards:       0,
		MaxCards:       len(playerCards),
	})

	log.Debug("Created pending card selection for sell patents",
		zap.Int("available_cards", len(playerCards)))

	log.Info("Sell patents initiated")
	return nil
}

// hasSellPatentsBehavior checks if any behavior has a card-discard input with variableAmount
func hasSellPatentsBehavior(behaviors []shared.CardBehavior) bool {
	for _, b := range behaviors {
		for _, input := range b.Inputs {
			if input.ResourceType == shared.ResourceCardDiscard && input.VariableAmount {
				return true
			}
		}
	}
	return false
}
