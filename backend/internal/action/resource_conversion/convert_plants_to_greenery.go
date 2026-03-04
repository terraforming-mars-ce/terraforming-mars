package resource_conversion

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// BasePlantsForGreenery is the base cost in plants to convert to greenery (before card discounts)
	BasePlantsForGreenery = 8
)

// ConvertPlantsToGreeneryAction handles the business logic for converting plants to greenery tile
// Uses RequirementModifierCalculator to apply card discounts (e.g., Ecoline: 7 plants instead of 8)
type ConvertPlantsToGreeneryAction struct {
	baseaction.BaseAction
	cardRegistry cards.CardRegistry
}

// NewConvertPlantsToGreeneryAction creates a new convert plants to greenery action
func NewConvertPlantsToGreeneryAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ConvertPlantsToGreeneryAction {
	return &ConvertPlantsToGreeneryAction{
		BaseAction:   baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
		cardRegistry: cardRegistry,
	}
}

// Execute performs the convert plants to greenery action
func (a *ConvertPlantsToGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "convert_plants_to_greenery"))
	log.Info("🌱 Converting plants to greenery")

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

	calculator := gamecards.NewRequirementModifierCalculator(a.cardRegistry)
	discounts := calculator.CalculateStandardProjectDiscounts(player, shared.StandardProjectConvertPlantsToGreenery)
	plantDiscount := discounts[shared.ResourcePlant]
	requiredPlants := BasePlantsForGreenery - plantDiscount
	if requiredPlants < 1 {
		requiredPlants = 1
	}
	log.Debug("💰 Calculated plants cost",
		zap.Int("base_cost", BasePlantsForGreenery),
		zap.Int("discount", plantDiscount),
		zap.Int("final_cost", requiredPlants))

	resources := player.Resources().Get()
	if resources.Plants < requiredPlants {
		log.Warn("Player cannot afford plants conversion",
			zap.Int("required", requiredPlants),
			zap.Int("available", resources.Plants))
		return fmt.Errorf("insufficient plants: need %d, have %d", requiredPlants, resources.Plants)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: -requiredPlants,
	})

	resources = player.Resources().Get()
	log.Info("🌿 Deducted plants",
		zap.Int("plants_spent", requiredPlants),
		zap.Int("remaining_plants", resources.Plants))

	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"greenery"},
		Source: "convert-plants-to-greenery",
		OnComplete: &playerPkg.TileCompletionCallback{
			Type: "convert-plants-to-greenery",
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

	log.Info("✅ Plants converted successfully, greenery tile queued for placement",
		zap.Int("plants_spent", requiredPlants))
	return nil
}
