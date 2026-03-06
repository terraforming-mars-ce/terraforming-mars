package standard_project

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// BuildCityCost is the megacredit cost to build a city via standard project
	BuildCityCost = 25
)

// BuildCityAction handles the business logic for building a city standard project
type BuildCityAction struct {
	baseaction.BaseAction
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *BuildCityAction {
	return &BuildCityAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "build_city"))
	log.Debug("Building city")

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

	resources := player.Resources().Get()
	if resources.Credits < BuildCityCost {
		log.Warn("Insufficient credits for city",
			zap.Int("cost", BuildCityCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildCityCost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -BuildCityCost,
	})

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: string(shared.StandardProjectCity),
		ProjectCost: BuildCityCost,
		Timestamp:   time.Now(),
	})

	resources = player.Resources().Get()
	log.Debug("Deducted city cost",
		zap.Int("cost", BuildCityCost),
		zap.Int("remaining_credits", resources.Credits))

	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 1,
	})

	production := player.Resources().Production()
	log.Debug("Increased credit production",
		zap.Int("new_credit_production", production.Credits))

	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"city"},
		Source: "standard-project-city",
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Debug("Created tile queue for city placement (auto-processed by SetPendingTileSelectionQueue)")

	a.ConsumePlayerAction(g, log)

	calculatedOutputs := []game.CalculatedOutput{
		{ResourceType: string(shared.ResourceCreditProduction), Amount: 1, IsScaled: false},
		{ResourceType: string(shared.ResourceCityPlacement), Amount: 1, IsScaled: false},
	}

	g.AddTriggeredEffect(game.TriggeredEffect{
		CardName:          "City",
		PlayerID:          playerID,
		SourceType:        game.SourceTypeStandardProject,
		CalculatedOutputs: calculatedOutputs,
	})

	displayData := baseaction.GetStandardProjectDisplayData("Standard Project: City")
	a.WriteStateLogFull(ctx, g, "Standard Project: City", game.SourceTypeStandardProject, playerID, "Built city", nil, calculatedOutputs, displayData)

	log.Info("City built",
		zap.Int("new_credit_production", production.Credits),
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
