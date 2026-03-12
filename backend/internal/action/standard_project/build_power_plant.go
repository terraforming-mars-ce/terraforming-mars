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
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// BuildPowerPlantCost is the megacredit cost to build a power plant via standard project
	BuildPowerPlantCost = 11
)

// BuildPowerPlantAction handles the build power plant standard project
// New architecture: Uses only GameRepository + logger
type BuildPowerPlantAction struct {
	baseaction.BaseAction
}

// NewBuildPowerPlantAction creates a new build power plant action
func NewBuildPowerPlantAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *BuildPowerPlantAction {
	return &BuildPowerPlantAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// Execute performs the build power plant action
func (a *BuildPowerPlantAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Debug("Building power plant")

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

	effectiveCost := BuildPowerPlantCost
	if a.CardRegistry() != nil {
		calculator := gamecards.NewRequirementModifierCalculator(a.CardRegistry())
		discounts := calculator.CalculateStandardProjectDiscounts(player, shared.StandardProjectPowerPlant)
		creditDiscount := discounts[shared.ResourceCredit]
		effectiveCost = BuildPowerPlantCost - creditDiscount
		if effectiveCost < 0 {
			effectiveCost = 0
		}
		if creditDiscount > 0 {
			log.Debug("Applied power plant discount",
				zap.Int("base_cost", BuildPowerPlantCost),
				zap.Int("discount", creditDiscount),
				zap.Int("effective_cost", effectiveCost))
		}
	}

	resources := player.Resources().Get()
	if resources.Credits < effectiveCost {
		log.Warn("Insufficient credits for power plant",
			zap.Int("cost", effectiveCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", effectiveCost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -effectiveCost,
	})

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: string(shared.StandardProjectPowerPlant),
		ProjectCost: BuildPowerPlantCost,
		Timestamp:   time.Now(),
	})

	resources = player.Resources().Get()
	log.Debug("Deducted power plant cost",
		zap.Int("cost", effectiveCost),
		zap.Int("remaining_credits", resources.Credits))

	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 1,
	})

	production := player.Resources().Production()
	log.Debug("Increased energy production",
		zap.Int("new_energy_production", production.Energy))

	a.ConsumePlayerAction(g, log)

	calculatedOutputs := []shared.CalculatedOutput{
		{ResourceType: string(shared.ResourceEnergyProduction), Amount: 1, IsScaled: false},
	}

	g.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:          "Power Plant",
		PlayerID:          playerID,
		SourceType:        shared.SourceTypeStandardProject,
		CalculatedOutputs: calculatedOutputs,
	})

	displayData := baseaction.GetStandardProjectDisplayData("Standard Project: Power Plant")
	a.WriteStateLogFull(ctx, g, "Standard Project: Power Plant", shared.SourceTypeStandardProject, playerID, "Built power plant", nil, calculatedOutputs, displayData)

	log.Info("Power plant built",
		zap.Int("new_energy_production", production.Energy),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
