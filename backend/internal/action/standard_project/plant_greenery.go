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
	// PlantGreeneryStandardProjectCost is the megacredit cost to plant greenery via standard project
	PlantGreeneryStandardProjectCost = 23
)

// PlantGreeneryAction handles the business logic for the plant greenery standard project
type PlantGreeneryAction struct {
	baseaction.BaseAction
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "plant_greenery"))
	log.Info("🌱 Planting greenery (standard project)")

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
	if resources.Credits < PlantGreeneryStandardProjectCost {
		log.Warn("Insufficient credits for greenery",
			zap.Int("cost", PlantGreeneryStandardProjectCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", PlantGreeneryStandardProjectCost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -PlantGreeneryStandardProjectCost,
	})

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: string(shared.StandardProjectGreenery),
		ProjectCost: PlantGreeneryStandardProjectCost,
		Timestamp:   time.Now(),
	})

	resources = player.Resources().Get()
	log.Info("💰 Deducted greenery cost",
		zap.Int("cost", PlantGreeneryStandardProjectCost),
		zap.Int("remaining_credits", resources.Credits))

	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"greenery"},
		Source: "standard-project-greenery",
		OnComplete: &playerPkg.TileCompletionCallback{
			Type: "standard-project-greenery",
		},
		TileRestrictions: &shared.TileRestrictions{
			AdjacentToOwned: true,
		},
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for greenery placement (auto-processed by SetPendingTileSelectionQueue)")

	a.ConsumePlayerAction(g, log)

	log.Info("✅ Greenery tile selection ready",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
