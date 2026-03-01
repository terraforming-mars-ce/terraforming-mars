package confirmation

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"

	"go.uber.org/zap"
)

// ConfirmBehaviorChoiceAction handles the business logic for confirming a behavior choice selection
type ConfirmBehaviorChoiceAction struct {
	baseaction.BaseAction
}

// NewConfirmBehaviorChoiceAction creates a new confirm behavior choice action
func NewConfirmBehaviorChoiceAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConfirmBehaviorChoiceAction {
	return &ConfirmBehaviorChoiceAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute performs the confirm behavior choice action
func (a *ConfirmBehaviorChoiceAction) Execute(ctx context.Context, gameID string, playerID string, choiceIndex int, cardStorageTargets []string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_behavior_choice"),
		zap.Int("choice_index", choiceIndex),
	)
	log.Info("🔀 Confirming behavior choice selection")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	selection := p.Selection().GetPendingBehaviorChoiceSelection()
	if selection == nil {
		log.Warn("No pending behavior choice selection found")
		return fmt.Errorf("no pending behavior choice selection found")
	}

	if choiceIndex < 0 || choiceIndex >= len(selection.Choices) {
		log.Warn("Invalid choice index",
			zap.Int("choice_index", choiceIndex),
			zap.Int("num_choices", len(selection.Choices)))
		return fmt.Errorf("invalid choice index %d, must be 0-%d", choiceIndex, len(selection.Choices)-1)
	}

	selectedChoice := selection.Choices[choiceIndex]

	// Validate choice requirements before applying
	if choiceErrors := baseaction.CalculateChoiceErrors(selectedChoice, p, g, a.CardRegistry()); len(choiceErrors) > 0 {
		log.Warn("Choice requirements not met",
			zap.Int("choice_index", choiceIndex),
			zap.String("error", choiceErrors[0].Message))
		return fmt.Errorf("choice %d requirements not met: %s", choiceIndex, choiceErrors[0].Message)
	}

	applier := gamecards.NewBehaviorApplier(p, g, selection.Source, log).
		WithSourceCardID(selection.SourceCardID).
		WithCardRegistry(a.CardRegistry()).
		WithSourceType(game.SourceTypePassiveEffect).
		WithOnCardsAddedToHand(baseaction.MakeCardDrawCallback(p, g, a.CardRegistry()))

	if len(cardStorageTargets) > 0 {
		applier = applier.WithTargetCardIDs(cardStorageTargets)
	}

	// Apply inputs (deduct resources)
	if len(selectedChoice.Inputs) > 0 {
		if err := applier.ApplyInputs(ctx, selectedChoice.Inputs); err != nil {
			log.Error("Failed to apply choice inputs", zap.Error(err))
			return fmt.Errorf("failed to apply choice inputs: %w", err)
		}
	}

	// Apply outputs (add resources)
	if len(selectedChoice.Outputs) > 0 {
		if err := applier.ApplyOutputs(ctx, selectedChoice.Outputs); err != nil {
			log.Error("Failed to apply choice outputs", zap.Error(err))
			return fmt.Errorf("failed to apply choice outputs: %w", err)
		}
	}

	// Clear the pending selection
	p.Selection().SetPendingBehaviorChoiceSelection(nil)

	log.Info("✅ Behavior choice confirmation completed",
		zap.String("source", selection.Source),
		zap.Int("choice_selected", choiceIndex))

	return nil
}
