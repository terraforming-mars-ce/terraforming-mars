package standard_project

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// LaunchAsteroidCost is the megacredit cost to launch an asteroid via standard project
	LaunchAsteroidCost = 14
)

// LaunchAsteroidAction handles the business logic for the launch asteroid standard project
type LaunchAsteroidAction struct {
	baseaction.BaseAction
}

// NewLaunchAsteroidAction creates a new launch asteroid action
func NewLaunchAsteroidAction(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *LaunchAsteroidAction {
	return &LaunchAsteroidAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
	}
}

// Execute performs the launch asteroid action
func (a *LaunchAsteroidAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "launch_asteroid"))
	log.Debug("Launching asteroid")

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

	resources := player.Resources().Get()
	if resources.Credits < LaunchAsteroidCost {
		log.Warn("Insufficient credits for asteroid",
			zap.Int("cost", LaunchAsteroidCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", LaunchAsteroidCost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -LaunchAsteroidCost,
	})

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: string(shared.StandardProjectAsteroid),
		ProjectCost: LaunchAsteroidCost,
		Timestamp:   time.Now(),
	})

	resources = player.Resources().Get()
	log.Debug("Deducted asteroid cost",
		zap.Int("cost", LaunchAsteroidCost),
		zap.Int("remaining_credits", resources.Credits))

	oldTemp := g.GlobalParameters().Temperature()
	stepsRaised, err := g.GlobalParameters().IncreaseTemperature(ctx, 1, playerID)
	if err != nil {
		log.Error("Failed to increase temperature", zap.Error(err))
		return fmt.Errorf("failed to increase temperature: %w", err)
	}
	newTemp := g.GlobalParameters().Temperature()

	if stepsRaised > 0 {
		log.Debug("Increased temperature",
			zap.Int("old_temperature", oldTemp),
			zap.Int("new_temperature", newTemp),
			zap.Int("steps_raised", stepsRaised))
	}

	if stepsRaised > 0 {
		oldTR := player.Resources().TerraformRating()
		player.Resources().UpdateTerraformRating(1)
		newTR := player.Resources().TerraformRating()

		log.Debug("Increased terraform rating",
			zap.Int("old_tr", oldTR),
			zap.Int("new_tr", newTR))
	}

	a.ConsumePlayerAction(g, log)

	calculatedOutputs := []shared.CalculatedOutput{
		{ResourceType: string(shared.ResourceTemperature), Amount: stepsRaised, IsScaled: false},
	}
	if stepsRaised > 0 {
		calculatedOutputs = append(calculatedOutputs, shared.CalculatedOutput{
			ResourceType: string(shared.ResourceTR), Amount: 1, IsScaled: false,
		})
	}

	g.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:          "Asteroid",
		PlayerID:          playerID,
		SourceType:        shared.SourceTypeStandardProject,
		CalculatedOutputs: calculatedOutputs,
	})

	displayData := baseaction.GetStandardProjectDisplayData("Standard Project: Asteroid")
	a.WriteStateLogFull(ctx, g, "Standard Project: Asteroid", shared.SourceTypeStandardProject, playerID, "Launched asteroid", nil, calculatedOutputs, displayData)

	log.Info("Asteroid launched",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
