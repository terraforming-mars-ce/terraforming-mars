package card

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// UseCardActionAction handles the business logic for using a card's manual action
// Card actions are repeatable blue card abilities with inputs and outputs
type UseCardActionAction struct {
	baseaction.BaseAction
}

// NewUseCardActionAction creates a new use card action action
func NewUseCardActionAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *UseCardActionAction {
	return &UseCardActionAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// Execute performs the use card action
func (a *UseCardActionAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	cardID string,
	behaviorIndex int,
	choiceIndex *int,
	cardStorageTargets []string,
	targetPlayerID *string,
	stealSourceCardID *string,
	selectedAmount *int,
	actionPayment *gamecards.CardPayment,
	reuseSourceCardID *string,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex),
		zap.String("action", "use_card_action"),
	)
	if choiceIndex != nil {
		log = log.With(zap.Int("choice_index", *choiceIndex))
	}
	if len(cardStorageTargets) > 0 {
		log = log.With(zap.Strings("card_storage_targets", cardStorageTargets))
	}
	if targetPlayerID != nil {
		log = log.With(zap.String("target_player_id", *targetPlayerID))
	}
	if stealSourceCardID != nil {
		log = log.With(zap.String("source_card_for_input", *stealSourceCardID))
	}
	log.Debug("Player attempting to use card action")

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

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	cardAction, err := a.findCardAction(p, cardID, behaviorIndex, log)
	if err != nil {
		return err
	}

	if reuseSourceCardID != nil {
		return a.executeReuse(ctx, g, p, cardAction, cardID, behaviorIndex, choiceIndex, cardStorageTargets, targetPlayerID, stealSourceCardID, selectedAmount, actionPayment, *reuseSourceCardID, log)
	}

	if a.hasManualTrigger(cardAction.Behavior) && cardAction.TimesUsedThisGeneration >= 1 {
		log.Warn("Action already played this generation",
			zap.Int("times_used", cardAction.TimesUsedThisGeneration))
		return fmt.Errorf("action already played this generation")
	}

	log.Debug("Found card action",
		zap.String("card_name", cardAction.CardName),
		zap.Int("times_used_this_generation", cardAction.TimesUsedThisGeneration))

	if choiceIndex != nil && cardAction.Behavior.ChoicePolicy != nil {
		production := p.Resources().Production()
		if !shared.IsChoiceValidForPolicy(*choiceIndex, cardAction.Behavior.Choices, cardAction.Behavior.ChoicePolicy, production) {
			log.Warn("Choice rejected by policy",
				zap.String("policy_type", string(cardAction.Behavior.ChoicePolicy.Type)),
				zap.Int("choice_index", *choiceIndex))
			return fmt.Errorf("choice not valid: policy %q restricts available options", cardAction.Behavior.ChoicePolicy.Type)
		}
	}

	applier := gamecards.NewBehaviorApplier(p, g, cardAction.CardName, log).
		WithSourceCardID(cardID).
		WithSourceBehaviorIndex(behaviorIndex).
		WithCardRegistry(a.CardRegistry()).
		WithSourceType(shared.SourceTypeCardAction)
	if len(cardStorageTargets) > 0 {
		applier = applier.WithTargetCardIDs(cardStorageTargets)
	}
	if targetPlayerID != nil {
		applier = applier.WithTargetPlayerID(*targetPlayerID)
	}
	if stealSourceCardID != nil {
		applier = applier.WithStealSourceCardID(*stealSourceCardID)
	}
	if selectedAmount != nil {
		applier = applier.WithSelectedAmount(*selectedAmount)
	}
	if actionPayment != nil {
		applier = applier.WithActionPayment(actionPayment)
	}

	inputs, outputs := cardAction.Behavior.ExtractInputsOutputs(choiceIndex)

	if choiceIndex != nil {
		log.Debug("Using choice-specific behavior",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("input_count", len(inputs)),
			zap.Int("output_count", len(outputs)))
	}

	if hasVariableAmount(inputs, outputs) && selectedAmount == nil {
		log.Warn("Variable-amount action requires selectedAmount")
		return fmt.Errorf("must select an amount for this action")
	}

	if err := validateOutputAffordability(p, outputs); err != nil {
		log.Warn("Cannot afford negative resource outputs", zap.Error(err))
		return err
	}

	if err := applier.ApplyInputs(ctx, inputs); err != nil {
		log.Error("Failed to apply inputs", zap.Error(err))
		return err
	}

	// Check for card draw outputs (card-peek/take/buy) - these create pending selection
	hasPending, err := applier.ApplyCardDrawOutputs(ctx, outputs)
	if err != nil {
		log.Error("Failed to apply card draw outputs", zap.Error(err))
		return err
	}
	if hasPending {
		// Pending selection created - action completion deferred to confirmation
		// Don't increment usage counts or consume action here - that happens in ConfirmCardDraw
		log.Debug("Card draw selection pending, awaiting player choice")
		return nil
	}

	calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
	if err != nil {
		log.Error("Failed to apply outputs", zap.Error(err))
		return err
	}

	a.incrementUsageCounts(p, cardID, behaviorIndex, log)

	a.ConsumePlayerAction(g, log)

	description := fmt.Sprintf("Used %s action", cardAction.CardName)
	var displayData *game.LogDisplayData
	if cardFromRegistry, err := a.CardRegistry().GetByID(cardID); err == nil {
		displayData = baseaction.BuildCardDisplayData(cardFromRegistry, shared.SourceTypeCardAction)
	}
	a.WriteStateLogFull(ctx, g, cardAction.CardName, shared.SourceTypeCardAction, playerID, description, choiceIndex, calculatedOutputs, displayData)

	log.Info("Card action executed")
	return nil
}

// findCardAction finds a card action in the player's available actions
func (a *UseCardActionAction) findCardAction(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) (*shared.CardAction, error) {
	actions := p.Actions().List()

	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			return &actions[i], nil
		}
	}

	log.Error("Card action not found in player's available actions",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))
	return nil, fmt.Errorf("card action not found: %s[%d]", cardID, behaviorIndex)
}

// incrementUsageCounts increments the usage counts for a card action
func (a *UseCardActionAction) incrementUsageCounts(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) {
	actions := p.Actions().List()

	// Find and increment both turn and generation counts
	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			actions[i].TimesUsedThisTurn++
			actions[i].TimesUsedThisGeneration++
			log.Debug("Incremented action usage counts",
				zap.Int("times_used_this_turn", actions[i].TimesUsedThisTurn),
				zap.Int("times_used_this_generation", actions[i].TimesUsedThisGeneration))
			break
		}
	}

	// Update player actions
	p.Actions().SetActions(actions)
}

func (a *UseCardActionAction) executeReuse(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	targetAction *shared.CardAction,
	targetCardID string,
	targetBehaviorIndex int,
	choiceIndex *int,
	cardStorageTargets []string,
	targetPlayerID *string,
	stealSourceCardID *string,
	selectedAmount *int,
	actionPayment *gamecards.CardPayment,
	reuseSourceCardID string,
	log *zap.Logger,
) error {
	log = log.With(zap.String("reuse_source_card_id", reuseSourceCardID))
	log.Debug("Executing action reuse")

	reuseAction, err := a.findActionReuseAction(p, reuseSourceCardID, log)
	if err != nil {
		return err
	}

	if reuseAction.TimesUsedThisGeneration >= 1 {
		log.Warn("Reuse action already played this generation")
		return fmt.Errorf("reuse action already played this generation")
	}

	if targetCardID == reuseSourceCardID {
		log.Warn("Cannot reuse own action-reuse ability")
		return fmt.Errorf("cannot reuse own action-reuse ability")
	}

	if !a.hasManualTrigger(targetAction.Behavior) {
		log.Warn("Target action is not a manual action")
		return fmt.Errorf("target action is not a manual action")
	}

	if targetAction.TimesUsedThisGeneration < 1 {
		log.Warn("Target action has not been used this generation")
		return fmt.Errorf("target action has not been used this generation")
	}

	if choiceIndex != nil && targetAction.Behavior.ChoicePolicy != nil {
		production := p.Resources().Production()
		if !shared.IsChoiceValidForPolicy(*choiceIndex, targetAction.Behavior.Choices, targetAction.Behavior.ChoicePolicy, production) {
			log.Warn("Choice rejected by policy",
				zap.String("policy_type", string(targetAction.Behavior.ChoicePolicy.Type)),
				zap.Int("choice_index", *choiceIndex))
			return fmt.Errorf("choice not valid: policy %q restricts available options", targetAction.Behavior.ChoicePolicy.Type)
		}
	}

	applier := gamecards.NewBehaviorApplier(p, g, targetAction.CardName, log).
		WithSourceCardID(targetCardID).
		WithSourceBehaviorIndex(targetBehaviorIndex).
		WithCardRegistry(a.CardRegistry()).
		WithSourceType(shared.SourceTypeCardAction)
	if len(cardStorageTargets) > 0 {
		applier = applier.WithTargetCardIDs(cardStorageTargets)
	}
	if targetPlayerID != nil {
		applier = applier.WithTargetPlayerID(*targetPlayerID)
	}
	if stealSourceCardID != nil {
		applier = applier.WithStealSourceCardID(*stealSourceCardID)
	}
	if selectedAmount != nil {
		applier = applier.WithSelectedAmount(*selectedAmount)
	}
	if actionPayment != nil {
		applier = applier.WithActionPayment(actionPayment)
	}

	inputs, outputs := targetAction.Behavior.ExtractInputsOutputs(choiceIndex)

	if hasVariableAmount(inputs, outputs) && selectedAmount == nil {
		log.Warn("Variable-amount action requires selectedAmount")
		return fmt.Errorf("must select an amount for this action")
	}

	if err := validateOutputAffordability(p, outputs); err != nil {
		log.Warn("Cannot afford negative resource outputs", zap.Error(err))
		return err
	}

	if err := applier.ApplyInputs(ctx, inputs); err != nil {
		log.Error("Failed to apply inputs for reused action", zap.Error(err))
		return err
	}

	hasPending, err := applier.ApplyCardDrawOutputs(ctx, outputs)
	if err != nil {
		log.Error("Failed to apply card draw outputs for reused action", zap.Error(err))
		return err
	}
	if hasPending {
		log.Debug("Card draw selection pending for reused action")
		return nil
	}

	calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
	if err != nil {
		log.Error("Failed to apply outputs for reused action", zap.Error(err))
		return err
	}

	a.incrementUsageCounts(p, reuseSourceCardID, reuseAction.BehaviorIndex, log)

	a.ConsumePlayerAction(g, log)

	description := fmt.Sprintf("Used %s to reuse %s action", reuseAction.CardName, targetAction.CardName)
	var displayData *game.LogDisplayData
	if cardFromRegistry, err := a.CardRegistry().GetByID(targetCardID); err == nil {
		displayData = baseaction.BuildCardDisplayData(cardFromRegistry, shared.SourceTypeCardAction)
	}
	a.WriteStateLogFull(ctx, g, targetAction.CardName, shared.SourceTypeCardAction, p.ID(), description, choiceIndex, calculatedOutputs, displayData)

	log.Info("Action reused",
		zap.String("reuse_source", reuseAction.CardName),
		zap.String("target_action", targetAction.CardName))
	return nil
}

func (a *UseCardActionAction) findActionReuseAction(
	p *player.Player,
	reuseSourceCardID string,
	log *zap.Logger,
) (*shared.CardAction, error) {
	actions := p.Actions().List()
	for i := range actions {
		if actions[i].CardID != reuseSourceCardID {
			continue
		}
		for _, output := range actions[i].Behavior.Outputs {
			if output.ResourceType == shared.ResourceActionReuse {
				return &actions[i], nil
			}
		}
	}
	log.Error("Action-reuse action not found", zap.String("card_id", reuseSourceCardID))
	return nil, fmt.Errorf("action-reuse action not found on card %s", reuseSourceCardID)
}

func (a *UseCardActionAction) hasManualTrigger(behavior shared.CardBehavior) bool {
	for _, trigger := range behavior.Triggers {
		if trigger.Type == shared.TriggerTypeManual {
			return true
		}
	}
	return false
}

func hasVariableAmount(inputs, outputs []shared.ResourceCondition) bool {
	for _, input := range inputs {
		if input.VariableAmount {
			return true
		}
	}
	for _, output := range outputs {
		if output.VariableAmount {
			return true
		}
	}
	return false
}

// validateOutputAffordability checks that the player can afford all negative resource outputs
// before they are applied. This is a defense-in-depth check; card costs should be modeled as
// inputs, but this catches any negative resource outputs that slip through.
func validateOutputAffordability(p *player.Player, outputs []shared.ResourceCondition) error {
	resources := p.Resources().Get()
	for _, output := range outputs {
		if output.VariableAmount || output.Amount >= 0 {
			continue
		}
		var available int
		switch output.ResourceType {
		case shared.ResourceCredit:
			available = resources.Credits
		case shared.ResourceSteel:
			available = resources.Steel
		case shared.ResourceTitanium:
			available = resources.Titanium
		case shared.ResourcePlant:
			available = resources.Plants
		case shared.ResourceEnergy:
			available = resources.Energy
		case shared.ResourceHeat:
			available = resources.Heat
		default:
			continue
		}
		if available < -output.Amount {
			return fmt.Errorf("insufficient %s: need %d, have %d", output.ResourceType, -output.Amount, available)
		}
	}
	return nil
}
