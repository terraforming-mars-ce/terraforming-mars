package resource_conversion

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

const (
	// BaseHeatForTemperature is the base cost in heat to raise temperature (before card discounts)
	BaseHeatForTemperature = 8
)

// ConvertHeatToTemperatureAction handles converting heat to raise temperature
// New architecture: Uses only GameRepository + logger, events handle broadcasting
type ConvertHeatToTemperatureAction struct {
	baseaction.BaseAction
	cardRegistry cards.CardRegistry
}

// NewConvertHeatToTemperatureAction creates a new convert heat action
func NewConvertHeatToTemperatureAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		BaseAction:   baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
		cardRegistry: cardRegistry,
	}
}

// Execute performs the convert heat to temperature action
func (a *ConvertHeatToTemperatureAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔥 Converting heat to temperature")

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
	discounts := calculator.CalculateStandardProjectDiscounts(player, shared.StandardProjectConvertHeatToTemperature)
	heatDiscount := discounts[shared.ResourceHeat]
	requiredHeat := BaseHeatForTemperature - heatDiscount
	if requiredHeat < 1 {
		requiredHeat = 1
	}
	log.Debug("💰 Calculated heat cost",
		zap.Int("base_cost", BaseHeatForTemperature),
		zap.Int("discount", heatDiscount),
		zap.Int("final_cost", requiredHeat))

	resources := player.Resources().Get()
	if resources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", resources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, resources.Heat)
	}

	resources.Heat -= requiredHeat
	player.Resources().Set(resources)

	log.Info("🔥 Deducted heat",
		zap.Int("heat_spent", requiredHeat),
		zap.Int("remaining_heat", resources.Heat))

	var stepsRaised int
	currentTemp := g.GlobalParameters().Temperature()
	if currentTemp < global_parameters.MaxTemperature {
		var err error
		stepsRaised, err = g.GlobalParameters().IncreaseTemperature(ctx, 1)
		if err != nil {
			log.Error("Failed to raise temperature", zap.Error(err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		if stepsRaised > 0 {
			newTemp := g.GlobalParameters().Temperature()
			log.Info("🌡️ Temperature raised",
				zap.Int("old_temperature", currentTemp),
				zap.Int("new_temperature", newTemp),
				zap.Int("steps_raised", stepsRaised))

			oldTR := player.Resources().TerraformRating()
			player.Resources().UpdateTerraformRating(1)
			newTR := player.Resources().TerraformRating()

			log.Info("🏆 Increased terraform rating",
				zap.Int("old_tr", oldTR),
				zap.Int("new_tr", newTR))
		}
	} else {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")
	}

	a.ConsumePlayerAction(g, log)

	calculatedOutputs := []game.CalculatedOutput{
		{ResourceType: string(shared.ResourceTemperature), Amount: stepsRaised, IsScaled: false},
	}
	if stepsRaised > 0 {
		calculatedOutputs = append(calculatedOutputs, game.CalculatedOutput{
			ResourceType: string(shared.ResourceTR), Amount: 1, IsScaled: false,
		})
	}

	g.AddTriggeredEffect(game.TriggeredEffect{
		CardName:          "Convert Heat",
		PlayerID:          playerID,
		SourceType:        game.SourceTypeResourceConvert,
		CalculatedOutputs: calculatedOutputs,
	})

	displayData := baseaction.GetStandardProjectDisplayData("Convert Heat")
	a.WriteStateLogFull(ctx, g, "Convert Heat", game.SourceTypeResourceConvert, playerID, "Converted heat to raise temperature", nil, calculatedOutputs, displayData)

	log.Info("✅ Heat converted successfully",
		zap.Int("heat_spent", requiredHeat))
	return nil
}
