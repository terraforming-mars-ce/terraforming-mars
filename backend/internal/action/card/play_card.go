package card

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
)

// PlayCardAction handles the business logic for playing a project card from hand
// Card playing involves: validating requirements, calculating costs (with discounts),
// moving card to played cards, applying immediate effects, and deducting payment
type PlayCardAction struct {
	baseaction.BaseAction
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *PlayCardAction {
	return &PlayCardAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// PaymentRequest represents the payment resources provided by the player
type PaymentRequest struct {
	Credits            int                         `json:"credits"`
	Steel              int                         `json:"steel"`
	Titanium           int                         `json:"titanium"`
	Substitutes        map[shared.ResourceType]int `json:"substitutes"`
	StorageSubstitutes map[string]int              `json:"storageSubstitutes"` // cardID -> amount of storage resources to use as payment
}

// Execute performs the play card action
func (a *PlayCardAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	cardID string,
	payment PaymentRequest,
	choiceIndex *int,
	cardStorageTargets []string,
	targetPlayerID *string,
	selectedAmount *int,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.String("action", "play_card"),
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
	log.Debug("Player attempting to play card")

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

	// Collect temporary "next-card" effect card IDs BEFORE playing
	// so we can clean them up after this card is played (but not any new ones created by this card)
	prePlayTemporaryCardIDs := collectTemporaryEffectCardIDs(player, shared.TemporaryNextCard)

	if !player.Hand().HasCard(cardID) {
		log.Error("Card not in player's hand")
		return fmt.Errorf("card %s not in hand", cardID)
	}

	card, err := a.CardRegistry().GetByID(cardID)
	if err != nil {
		log.Error("Card not found in registry", zap.Error(err))
		return fmt.Errorf("card not found: %w", err)
	}

	log.Debug("Card data retrieved",
		zap.String("card_name", card.Name),
		zap.Int("base_cost", card.Cost))

	calculator := gamecards.NewRequirementModifierCalculator(a.CardRegistry())

	if err := validateCardRequirements(card, g, player, calculator, a.CardRegistry()); err != nil {
		log.Error("Card requirements not met", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	log.Debug("Card requirements validated")

	if tileErrors := baseaction.ValidateTileOutputs(card, player, g); len(tileErrors) > 0 {
		log.Error("Tile placement not available", zap.String("error", tileErrors[0].Message))
		return fmt.Errorf("cannot play card: %s", tileErrors[0].Message)
	}

	if err := validateProductionOutputs(card, player); err != nil {
		log.Error("Production output validation failed", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	if err := validateNegativeResourceOutputs(card, player); err != nil {
		log.Error("Negative resource output validation failed", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	// Validate choice requirements early (before any state mutations)
	if choiceIndex != nil {
		for _, behavior := range card.Behaviors {
			if gamecards.HasAutoTrigger(behavior) && *choiceIndex >= 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				if selectedChoice.Requirements != nil {
					if err := validateChoiceRequirements(selectedChoice.Requirements, player, g, a.CardRegistry()); err != nil {
						log.Error("Choice requirements not met", zap.Error(err))
						return fmt.Errorf("choice %d requirements not met: %w", *choiceIndex, err)
					}
				}
			}
		}
		log.Debug("Choice requirements validated")
	}

	discountAmount := calculator.CalculateCardDiscounts(player, card)
	effectiveCost := card.Cost - discountAmount
	if effectiveCost < 0 {
		effectiveCost = 0
	}

	if discountAmount > 0 {
		log.Debug("Discount applied",
			zap.Int("base_cost", card.Cost),
			zap.Int("discount", discountAmount),
			zap.Int("effective_cost", effectiveCost))
	}

	playerSubstitutes := player.Resources().PaymentSubstitutes()

	// Get storage payment substitutes applicable to this card (filtered by selectors)
	allStorageSubs := player.Resources().StoragePaymentSubstitutes()
	var applicableStorageSubs []shared.StoragePaymentSubstitute
	for _, sub := range allStorageSubs {
		if len(sub.Selectors) == 0 || gamecards.MatchesAnySelector(card, sub.Selectors) {
			applicableStorageSubs = append(applicableStorageSubs, sub)
		}
	}

	allowSteel := gamecards.HasTag(card, shared.TagBuilding)
	allowTitanium := gamecards.HasTag(card, shared.TagSpace)

	adjustedPayment := adjustPaymentToEffectiveCost(payment, effectiveCost, allowSteel, allowTitanium, playerSubstitutes, applicableStorageSubs, player)

	cardPayment := gamecards.CardPayment{
		Credits:            adjustedPayment.Credits,
		Steel:              adjustedPayment.Steel,
		Titanium:           adjustedPayment.Titanium,
		Substitutes:        adjustedPayment.Substitutes,
		StorageSubstitutes: adjustedPayment.StorageSubstitutes,
	}

	if err := cardPayment.CoversCardCost(effectiveCost, allowSteel, allowTitanium, playerSubstitutes, applicableStorageSubs); err != nil {
		log.Error("Payment validation failed", zap.Error(err))
		return err
	}

	totalValue := cardPayment.TotalValue(playerSubstitutes, applicableStorageSubs)
	log.Debug("Payment validated",
		zap.Int("effective_cost", effectiveCost),
		zap.Int("payment_value", totalValue),
		zap.Int("credits", adjustedPayment.Credits),
		zap.Int("steel", adjustedPayment.Steel),
		zap.Int("titanium", adjustedPayment.Titanium),
		zap.Any("substitutes", adjustedPayment.Substitutes),
		zap.Any("storageSubstitutes", adjustedPayment.StorageSubstitutes))

	resources := player.Resources().Get()
	storageGetter := func(cardID string) int {
		return player.Resources().GetCardStorage(cardID)
	}
	if err := cardPayment.CanAfford(resources, storageGetter); err != nil {
		log.Error("Player can't afford payment", zap.Error(err))
		return err
	}

	if !player.Hand().RemoveCard(cardID) {
		log.Error("Failed to remove card from hand - card not found")
		return fmt.Errorf("failed to remove card from hand: card not found")
	}

	log.Debug("Card removed from hand")

	cardTags := make([]string, len(card.Tags))
	for i, tag := range card.Tags {
		cardTags[i] = string(tag)
	}

	player.PlayedCards().AddCard(cardID, card.Name, string(card.Type), cardTags)

	log.Debug("Card added to played cards")

	if card.ResourceStorage != nil {
		player.Resources().AddToStorage(cardID, card.ResourceStorage.Starting)
		log.Debug("Initialized resource storage",
			zap.String("card_id", cardID),
			zap.String("resource_type", string(card.ResourceStorage.Type)),
			zap.Int("starting_amount", card.ResourceStorage.Starting))
	}

	deductions := map[shared.ResourceType]int{
		shared.ResourceCredit:   -adjustedPayment.Credits,
		shared.ResourceSteel:    -adjustedPayment.Steel,
		shared.ResourceTitanium: -adjustedPayment.Titanium,
	}

	for resourceType, amount := range adjustedPayment.Substitutes {
		deductions[resourceType] = -amount
	}

	player.Resources().Add(deductions)

	// Deduct storage payment substitutes (e.g., Dirigibles floaters)
	for cardID, amount := range adjustedPayment.StorageSubstitutes {
		if amount > 0 {
			player.Resources().AddToStorage(cardID, -amount)
			log.Debug("Deducted storage payment",
				zap.String("card_id", cardID),
				zap.Int("amount", amount))
		}
	}

	log.Debug("Payment deducted",
		zap.Int("credits", adjustedPayment.Credits),
		zap.Int("steel", adjustedPayment.Steel),
		zap.Int("titanium", adjustedPayment.Titanium),
		zap.Any("substitutes", adjustedPayment.Substitutes),
		zap.Any("storageSubstitutes", adjustedPayment.StorageSubstitutes))

	calculatedOutputs, err := a.applyCardBehaviors(ctx, g, card, player, choiceIndex, cardStorageTargets, targetPlayerID, selectedAmount, log)
	if err != nil {
		log.Error("Failed to apply card behaviors", zap.Error(err))
		return fmt.Errorf("failed to apply card behaviors: %w", err)
	}

	// Clean up temporary "next-card" effects that existed before this card was played
	removePrePlayTemporaryEffects(player, prePlayTemporaryCardIDs, log)

	a.ConsumePlayerAction(g, log)

	description := fmt.Sprintf("Played %s for %d credits", card.Name, totalValue)
	displayData := baseaction.BuildCardDisplayData(card, shared.SourceTypeCardPlay)
	a.WriteStateLogFull(ctx, g, card.Name, shared.SourceTypeCardPlay, playerID, description, choiceIndex, calculatedOutputs, displayData)

	log.Info("Card played",
		zap.String("card_name", card.Name),
		zap.Int("card_cost", card.Cost),
		zap.Int("payment_value", totalValue))

	return nil
}

// validateCardRequirements validates that the player and game state meet all card requirements.
// Uses RequirementModifierCalculator to include global parameter lenience from temporary effects.
func validateCardRequirements(card *gamecards.Card, g *game.Game, player *player.Player, calculator *gamecards.RequirementModifierCalculator, cardRegistry cards.CardRegistry) error {
	if card.Requirements == nil || len(card.Requirements.Items) == 0 {
		return nil // No requirements to validate
	}

	for _, req := range card.Requirements.Items {
		switch req.Type {
		case gamecards.RequirementTemperature:
			lenience := calculator.CalculateGlobalParameterLenience(player, "temperature")
			temp := g.GlobalParameters().Temperature()
			if req.Min != nil && temp < *req.Min-lenience {
				return fmt.Errorf("temperature requirement not met: need %d°C, current %d°C", *req.Min, temp)
			}
			if req.Max != nil && temp > *req.Max+lenience {
				return fmt.Errorf("temperature requirement not met: max %d°C, current %d°C", *req.Max, temp)
			}

		case gamecards.RequirementOxygen:
			lenience := calculator.CalculateGlobalParameterLenience(player, "oxygen")
			oxygen := g.GlobalParameters().Oxygen()
			if req.Min != nil && oxygen < *req.Min-lenience {
				return fmt.Errorf("oxygen requirement not met: need %d%%, current %d%%", *req.Min, oxygen)
			}
			if req.Max != nil && oxygen > *req.Max+lenience {
				return fmt.Errorf("oxygen requirement not met: max %d%%, current %d%%", *req.Max, oxygen)
			}

		case gamecards.RequirementOceans:
			lenience := calculator.CalculateGlobalParameterLenience(player, "ocean")
			oceans := g.GlobalParameters().Oceans()
			if req.Min != nil && oceans < *req.Min-lenience {
				return fmt.Errorf("ocean requirement not met: need %d, current %d", *req.Min, oceans)
			}
			if req.Max != nil && oceans > *req.Max+lenience {
				return fmt.Errorf("ocean requirement not met: max %d, current %d", *req.Max, oceans)
			}

		case gamecards.RequirementTR:
			tr := player.Resources().TerraformRating()
			if req.Min != nil && tr < *req.Min {
				return fmt.Errorf("terraform rating requirement not met: need %d, current %d", *req.Min, tr)
			}
			if req.Max != nil && tr > *req.Max {
				return fmt.Errorf("terraform rating requirement not met: max %d, current %d", *req.Max, tr)
			}

		case gamecards.RequirementTags:
			if req.Tag == nil {
				return fmt.Errorf("tag requirement missing tag specification")
			}

			// Count the card's own tags toward requirements (per TM rules, the card being played counts)
			tagCount := gamecards.CountPlayerTagsByType(player, cardRegistry, *req.Tag, card.Tags)

			if req.Min != nil && tagCount < *req.Min {
				return fmt.Errorf("tag requirement not met: need %d %s tags, have %d", *req.Min, *req.Tag, tagCount)
			}
			if req.Max != nil && tagCount > *req.Max {
				return fmt.Errorf("tag requirement not met: max %d %s tags, have %d", *req.Max, *req.Tag, tagCount)
			}

		case gamecards.RequirementProduction:
			if req.Resource == nil {
				return fmt.Errorf("production requirement missing resource specification")
			}
			production := player.Resources().Production()
			var currentProd int
			switch *req.Resource {
			case shared.ResourceCredit, shared.ResourceCreditProduction:
				currentProd = production.Credits
			case shared.ResourceSteel, shared.ResourceSteelProduction:
				currentProd = production.Steel
			case shared.ResourceTitanium, shared.ResourceTitaniumProduction:
				currentProd = production.Titanium
			case shared.ResourcePlant, shared.ResourcePlantProduction:
				currentProd = production.Plants
			case shared.ResourceEnergy, shared.ResourceEnergyProduction:
				currentProd = production.Energy
			case shared.ResourceHeat, shared.ResourceHeatProduction:
				currentProd = production.Heat
			}
			if req.Min != nil && currentProd < *req.Min {
				return fmt.Errorf("production requirement not met: need %d %s production, have %d", *req.Min, *req.Resource, currentProd)
			}
			if req.Max != nil && currentProd > *req.Max {
				return fmt.Errorf("production requirement not met: max %d %s production, have %d", *req.Max, *req.Resource, currentProd)
			}

		case gamecards.RequirementResource:
			if req.Resource == nil {
				return fmt.Errorf("resource requirement missing resource specification")
			}
			resources := player.Resources().Get()
			var currentAmount int

			switch *req.Resource {
			case shared.ResourceCredit:
				currentAmount = resources.Credits
			case shared.ResourceSteel:
				currentAmount = resources.Steel
			case shared.ResourceTitanium:
				currentAmount = resources.Titanium
			case shared.ResourcePlant:
				currentAmount = resources.Plants
			case shared.ResourceEnergy:
				currentAmount = resources.Energy
			case shared.ResourceHeat:
				currentAmount = resources.Heat
			}

			if req.Min != nil && currentAmount < *req.Min {
				return fmt.Errorf("resource requirement not met: need %d %s, have %d", *req.Min, *req.Resource, currentAmount)
			}
			if req.Max != nil && currentAmount > *req.Max {
				return fmt.Errorf("resource requirement not met: max %d %s, have %d", *req.Max, *req.Resource, currentAmount)
			}

		case gamecards.RequirementCities, gamecards.RequirementGreeneries:
			// TODO: Implement tile-based requirements when Board tile counting is ready
			// For now, skip these validations

		case gamecards.RequirementVenus:
			lenience := calculator.CalculateGlobalParameterLenience(player, "venus")
			venus := g.GlobalParameters().Venus()
			if req.Min != nil && venus < *req.Min-lenience {
				return fmt.Errorf("venus requirement not met: need %d%%, current %d%%", *req.Min, venus)
			}
			if req.Max != nil && venus > *req.Max+lenience {
				return fmt.Errorf("venus requirement not met: max %d%%, current %d%%", *req.Max, venus)
			}
		}
	}

	return nil
}

// applyCardBehaviors processes all card behaviors and applies immediate effects or registers actions/effects
// Returns calculated outputs for logging purposes
func (a *PlayCardAction) applyCardBehaviors(
	ctx context.Context,
	g *game.Game,
	card *gamecards.Card,
	p *player.Player,
	choiceIndex *int,
	cardStorageTargets []string,
	targetPlayerID *string,
	selectedAmount *int,
	log *zap.Logger,
) ([]shared.CalculatedOutput, error) {
	if len(card.Behaviors) == 0 {
		log.Debug("No card behaviors to apply")
		return nil, nil
	}

	log.Debug("Processing card behaviors",
		zap.String("card_id", card.ID),
		zap.Int("behavior_count", len(card.Behaviors)))

	var allCalculatedOutputs []shared.CalculatedOutput

	for behaviorIndex, behavior := range card.Behaviors {
		log.Debug("Processing behavior",
			zap.Int("index", behaviorIndex),
			zap.Int("trigger_count", len(behavior.Triggers)))

		// Apply auto-trigger behaviors immediately
		if gamecards.HasAutoTrigger(behavior) {
			// Auto-select choice if behavior has an auto-selection policy
			effectiveChoiceIndex := choiceIndex
			if behavior.ChoicePolicy != nil && len(behavior.Choices) > 0 {
				count := resolveChoicePolicyCount(behavior.ChoicePolicy, p, a.CardRegistry())
				autoIdx := shared.AutoSelectChoiceIndex(behavior.ChoicePolicy, count)
				if autoIdx >= 0 {
					effectiveChoiceIndex = &autoIdx
					log.Debug("Auto-selected choice by policy",
						zap.String("policy_type", string(behavior.ChoicePolicy.Type)),
						zap.Int("choice_index", autoIdx))
				}
			}

			// Extract inputs and outputs, incorporating choice if present
			inputs, outputs := behavior.ExtractInputsOutputs(effectiveChoiceIndex)

			// Check for card-discard inputs — these defer output application
			if gamecards.HasCardDiscardInput(behavior) {
				a.createPendingCardDiscard(p, card, inputs, outputs, log)
				continue
			}

			// Check for card-discard outputs — player must choose cards to discard first
			if gamecards.HasCardDiscardOutput(behavior) {
				a.createPendingCardDiscardFromOutputs(p, card, outputs, log)
				continue
			}

			log.Debug("Found auto-trigger behavior, applying outputs immediately",
				zap.Int("output_count", len(outputs)))

			// Use BehaviorApplier for consistent output handling
			applier := gamecards.NewBehaviorApplier(p, g, card.Name, log).
				WithSourceCardID(card.ID).
				WithCardRegistry(a.CardRegistry()).
				WithSourceType(shared.SourceTypeCardPlay)
			if len(cardStorageTargets) > 0 {
				applier = applier.WithTargetCardIDs(cardStorageTargets)
			}
			if targetPlayerID != nil {
				applier = applier.WithTargetPlayerID(*targetPlayerID)
			}
			if selectedAmount != nil {
				applier = applier.WithSelectedAmount(*selectedAmount)
			}

			calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
			if err != nil {
				return nil, fmt.Errorf("failed to apply auto behavior %d outputs: %w", behaviorIndex, err)
			}

			allCalculatedOutputs = append(allCalculatedOutputs, calculatedOutputs...)

			if deferred := applier.DeferredSteal(); deferred != nil {
				callback := &shared.TileCompletionCallback{
					Type: "adjacent-steal",
					Data: map[string]interface{}{
						"resourceType": string(deferred.ResourceType),
						"amount":       deferred.Amount,
						"sourceCardID": card.ID,
						"source":       card.Name,
					},
				}
				g.SetTileQueueOnComplete(ctx, p.ID(), callback)
			}

			// Also register as effect if it has persistent outputs (discount, payment-substitute)
			// These need to show in the effects list for display and for modifier calculations
			if gamecards.HasPersistentEffects(behavior) {
				log.Debug("Registering auto-trigger behavior with persistent effects",
					zap.String("card_name", card.Name))

				effect := shared.CardEffect{
					CardID:        card.ID,
					CardName:      card.Name,
					BehaviorIndex: behaviorIndex,
					Behavior:      behavior,
				}
				p.Effects().AddEffect(effect)

				events.Publish(g.EventBus(), events.PlayerEffectsChangedEvent{
					GameID:    g.ID(),
					PlayerID:  p.ID(),
					Timestamp: time.Now(),
				})

				g.AddTriggeredEffect(shared.TriggeredEffect{
					CardName:   card.Name,
					PlayerID:   p.ID(),
					SourceType: shared.SourceTypeEffectAdded,
					Behaviors:  []shared.CardBehavior{behavior},
				})
			}
		}

		// Register manual-trigger behaviors as player actions
		if gamecards.HasManualTrigger(behavior) {
			log.Debug("Found manual-trigger behavior, registering as player action")

			p.Actions().AddAction(shared.CardAction{
				CardID:                  card.ID,
				CardName:                card.Name,
				BehaviorIndex:           behaviorIndex,
				Behavior:                behavior,
				TimesUsedThisTurn:       0,
				TimesUsedThisGeneration: 0,
			})

			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   card.Name,
				PlayerID:   p.ID(),
				SourceType: shared.SourceTypeActionAdded,
				Behaviors:  []shared.CardBehavior{behavior},
			})
		}

		// Register conditional-trigger behaviors as passive effects
		if gamecards.HasConditionalTrigger(behavior) {
			log.Debug("Found conditional-trigger behavior, registering as passive effect",
				zap.Int("trigger_count", len(behavior.Triggers)))

			effect := shared.CardEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			p.Effects().AddEffect(effect)

			events.Publish(g.EventBus(), events.PlayerEffectsChangedEvent{
				GameID:    g.ID(),
				PlayerID:  p.ID(),
				Timestamp: time.Now(),
			})

			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   card.Name,
				PlayerID:   p.ID(),
				SourceType: shared.SourceTypeEffectAdded,
				Behaviors:  []shared.CardBehavior{behavior},
			})

			// Subscribe passive effects to relevant events
			baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, a.CardRegistry())
		}
	}

	// Add VP notification if card has VP conditions
	if len(card.VPConditions) > 0 {
		var vpForLog []shared.VPConditionForLog
		for _, vp := range card.VPConditions {
			vpLog := shared.VPConditionForLog{
				Amount:    vp.Amount,
				Condition: string(vp.Condition),
			}
			if vp.MaxTrigger != nil {
				vpLog.MaxTrigger = vp.MaxTrigger
			}
			if vp.Per != nil {
				vpLog.Per = &shared.PerCondition{
					ResourceType: vp.Per.ResourceType,
					Amount:       vp.Per.Amount,
				}
				if vp.Per.Location != nil {
					loc := string(*vp.Per.Location)
					vpLog.Per.Location = &loc
				}
				if vp.Per.Target != nil {
					target := string(*vp.Per.Target)
					vpLog.Per.Target = &target
				}
				if vp.Per.Tag != nil {
					vpLog.Per.Tag = vp.Per.Tag
				}
			}
			vpForLog = append(vpForLog, vpLog)
		}
		g.AddTriggeredEffect(shared.TriggeredEffect{
			CardName:     card.Name,
			PlayerID:     p.ID(),
			SourceType:   shared.SourceTypeCardPlay,
			VPConditions: vpForLog,
		})
	}

	log.Debug("Card behaviors processed")
	return allCalculatedOutputs, nil
}

// createPendingCardDiscard creates a PendingCardDiscardSelection for behaviors with card-discard inputs.
// The player must select cards to discard before outputs are applied.
func (a *PlayCardAction) createPendingCardDiscard(
	p *player.Player,
	card *gamecards.Card,
	inputs []shared.ResourceCondition,
	outputs []shared.ResourceCondition,
	log *zap.Logger,
) {
	minCards := 0
	maxCards := 0
	isOptional := false

	for _, input := range inputs {
		if input.ResourceType == shared.ResourceCardDiscard {
			maxCards += input.Amount
			if !input.Optional {
				minCards += input.Amount
			} else {
				isOptional = true
			}
		}
	}

	// If optional but player has no cards in hand, skip entirely
	if isOptional && len(p.Hand().Cards()) == 0 {
		log.Debug("Skipping card discard: optional and player has no cards in hand")
		return
	}

	selection := &shared.PendingCardDiscardSelection{
		MinCards:       minCards,
		MaxCards:       maxCards,
		Source:         card.Name,
		SourceCardID:   card.ID,
		PendingOutputs: outputs,
	}

	p.Selection().SetPendingCardDiscardSelection(selection)

	log.Debug("Created pending card discard selection",
		zap.String("card_name", card.Name),
		zap.Int("min_cards", minCards),
		zap.Int("max_cards", maxCards),
		zap.Bool("optional", isOptional),
		zap.Int("pending_outputs", len(outputs)))
}

// createPendingCardDiscardFromOutputs creates a PendingCardDiscardSelection for behaviors with card-discard outputs.
// The player must select cards to discard before remaining outputs (draws, etc.) are applied.
func (a *PlayCardAction) createPendingCardDiscardFromOutputs(
	p *player.Player,
	card *gamecards.Card,
	outputs []shared.ResourceCondition,
	log *zap.Logger,
) {
	minCards := 0
	maxCards := 0
	var pendingOutputs []shared.ResourceCondition

	for _, output := range outputs {
		if output.ResourceType == shared.ResourceCardDiscard {
			minCards += output.Amount
			maxCards += output.Amount
		} else {
			pendingOutputs = append(pendingOutputs, output)
		}
	}

	selection := &shared.PendingCardDiscardSelection{
		MinCards:       minCards,
		MaxCards:       maxCards,
		Source:         card.Name,
		SourceCardID:   card.ID,
		PendingOutputs: pendingOutputs,
	}

	p.Selection().SetPendingCardDiscardSelection(selection)

	log.Debug("Created pending card discard selection from outputs",
		zap.String("card_name", card.Name),
		zap.Int("min_cards", minCards),
		zap.Int("max_cards", maxCards),
		zap.Int("pending_outputs", len(pendingOutputs)))
}

func adjustPaymentToEffectiveCost(
	payment PaymentRequest,
	effectiveCost int,
	allowSteel bool,
	allowTitanium bool,
	playerSubstitutes []shared.PaymentSubstitute,
	storageSubstitutes []shared.StoragePaymentSubstitute,
	p *player.Player,
) PaymentRequest {
	if effectiveCost <= 0 {
		return PaymentRequest{}
	}

	steelRate := 2
	titaniumRate := 3
	for _, sub := range playerSubstitutes {
		if sub.ResourceType == shared.ResourceSteel {
			steelRate = sub.ConversionRate
		}
		if sub.ResourceType == shared.ResourceTitanium {
			titaniumRate = sub.ConversionRate
		}
	}

	nonCreditValue := 0
	if allowSteel {
		nonCreditValue += payment.Steel * steelRate
	}
	if allowTitanium {
		nonCreditValue += payment.Titanium * titaniumRate
	}

	for resourceType, amount := range payment.Substitutes {
		for _, sub := range playerSubstitutes {
			if sub.ResourceType == resourceType {
				nonCreditValue += amount * sub.ConversionRate
				break
			}
		}
	}

	// Clamp storage substitute amounts to what's actually available on the card
	clampedStorageSubs := make(map[string]int)
	for cardID, amount := range payment.StorageSubstitutes {
		available := p.Resources().GetCardStorage(cardID)
		clamped := amount
		if clamped > available {
			clamped = available
		}
		if clamped > 0 {
			clampedStorageSubs[cardID] = clamped
		}
	}

	// Add storage substitute values
	storageSubValues := make(map[string]int)
	for _, sub := range storageSubstitutes {
		storageSubValues[sub.CardID] = sub.ConversionRate
	}
	for cardID, amount := range clampedStorageSubs {
		if rate, ok := storageSubValues[cardID]; ok {
			nonCreditValue += amount * rate
		}
	}

	if nonCreditValue >= effectiveCost {
		return PaymentRequest{
			Credits:            0,
			Steel:              payment.Steel,
			Titanium:           payment.Titanium,
			Substitutes:        payment.Substitutes,
			StorageSubstitutes: clampedStorageSubs,
		}
	}

	creditsNeeded := effectiveCost - nonCreditValue
	if creditsNeeded > payment.Credits {
		creditsNeeded = payment.Credits
	}

	return PaymentRequest{
		Credits:            creditsNeeded,
		Steel:              payment.Steel,
		Titanium:           payment.Titanium,
		Substitutes:        payment.Substitutes,
		StorageSubstitutes: clampedStorageSubs,
	}
}

// collectTemporaryEffectCardIDs returns the card IDs of all effects with the given temporary type.
func collectTemporaryEffectCardIDs(p *player.Player, temporaryType string) []string {
	var cardIDs []string
	for _, effect := range p.Effects().List() {
		for _, output := range effect.Behavior.Outputs {
			if output.Temporary == temporaryType {
				cardIDs = append(cardIDs, effect.CardID)
				break
			}
		}
	}
	return cardIDs
}

// removePrePlayTemporaryEffects removes temporary effects by their card IDs (collected before card play).
func removePrePlayTemporaryEffects(p *player.Player, cardIDs []string, log *zap.Logger) {
	for _, cardID := range cardIDs {
		p.Effects().RemoveEffectsByCardID(cardID)
		log.Debug("Removed temporary next-card effect",
			zap.String("effect_card_id", cardID))
	}
}

// validateChoiceRequirements checks if a choice's requirements are met by the player.
func validateChoiceRequirements(reqs *shared.ChoiceRequirements, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) error {
	if reqs == nil || len(reqs.Items) == 0 {
		return nil
	}

	for _, req := range reqs.Items {
		if err := checkChoiceRequirement(req, p, g, cardRegistry); err != nil {
			return err
		}
	}
	return nil
}

// checkChoiceRequirement validates a single choice requirement.
func checkChoiceRequirement(req shared.ChoiceRequirement, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) error {
	switch req.Type {
	case "tags":
		if req.Tag == nil {
			return fmt.Errorf("tag requirement missing tag specification")
		}
		tagCount := gamecards.CountPlayerTagsByType(p, cardRegistry, *req.Tag)
		if req.Min != nil && tagCount < *req.Min {
			return fmt.Errorf("need %d %s tags, have %d", *req.Min, *req.Tag, tagCount)
		}
		if req.Max != nil && tagCount > *req.Max {
			return fmt.Errorf("max %d %s tags, have %d", *req.Max, *req.Tag, tagCount)
		}

	case "temperature":
		temp := g.GlobalParameters().Temperature()
		if req.Min != nil && temp < *req.Min {
			return fmt.Errorf("temperature too low: need %d, have %d", *req.Min, temp)
		}
		if req.Max != nil && temp > *req.Max {
			return fmt.Errorf("temperature too high: max %d, have %d", *req.Max, temp)
		}

	case "oxygen":
		oxygen := g.GlobalParameters().Oxygen()
		if req.Min != nil && oxygen < *req.Min {
			return fmt.Errorf("oxygen too low: need %d, have %d", *req.Min, oxygen)
		}
		if req.Max != nil && oxygen > *req.Max {
			return fmt.Errorf("oxygen too high: max %d, have %d", *req.Max, oxygen)
		}

	case "ocean":
		oceans := g.GlobalParameters().Oceans()
		if req.Min != nil && oceans < *req.Min {
			return fmt.Errorf("too few oceans: need %d, have %d", *req.Min, oceans)
		}
		if req.Max != nil && oceans > *req.Max {
			return fmt.Errorf("too many oceans: max %d, have %d", *req.Max, oceans)
		}

	case "venus":
		venus := g.GlobalParameters().Venus()
		if req.Min != nil && venus < *req.Min {
			return fmt.Errorf("venus too low: need %d, have %d", *req.Min, venus)
		}
		if req.Max != nil && venus > *req.Max {
			return fmt.Errorf("venus too high: max %d, have %d", *req.Max, venus)
		}

	case "tr":
		tr := p.Resources().TerraformRating()
		if req.Min != nil && tr < *req.Min {
			return fmt.Errorf("TR too low: need %d, have %d", *req.Min, tr)
		}
		if req.Max != nil && tr > *req.Max {
			return fmt.Errorf("TR too high: max %d, have %d", *req.Max, tr)
		}

	case "production":
		if req.Resource == nil {
			return fmt.Errorf("production requirement missing resource type")
		}
		production := p.Resources().Production()
		amount := production.GetAmount(*req.Resource)
		if req.Min != nil && amount < *req.Min {
			return fmt.Errorf("%s production too low: need %d, have %d", *req.Resource, *req.Min, amount)
		}

	case "resource":
		if req.Resource == nil {
			return fmt.Errorf("resource requirement missing resource type")
		}
		resources := p.Resources().Get()
		amount := resources.GetAmount(*req.Resource)
		if req.Min != nil && amount < *req.Min {
			return fmt.Errorf("%s too low: need %d, have %d", *req.Resource, *req.Min, amount)
		}
	}

	return nil
}

// validateProductionOutputs checks that playing the card won't bring production below the minimum.
func validateProductionOutputs(card *gamecards.Card, p *player.Player) error {
	if len(card.Behaviors) == 0 {
		return nil
	}

	production := p.Resources().Production()

	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}
		for _, output := range behavior.Outputs {
			if output.VariableAmount || output.Amount >= 0 {
				continue
			}
			if output.Target != "self-player" {
				continue
			}
			var current, minProd int
			switch output.ResourceType {
			case shared.ResourceCreditProduction:
				current, minProd = production.Credits, shared.MinCreditProduction
			case shared.ResourceSteelProduction:
				current, minProd = production.Steel, shared.MinOtherProduction
			case shared.ResourceTitaniumProduction:
				current, minProd = production.Titanium, shared.MinOtherProduction
			case shared.ResourcePlantProduction:
				current, minProd = production.Plants, shared.MinOtherProduction
			case shared.ResourceEnergyProduction:
				current, minProd = production.Energy, shared.MinOtherProduction
			case shared.ResourceHeatProduction:
				current, minProd = production.Heat, shared.MinOtherProduction
			default:
				continue
			}
			if current+output.Amount < minProd {
				return fmt.Errorf("insufficient %s: have %d, need at least %d to decrease by %d", output.ResourceType, current, -output.Amount, -output.Amount)
			}
		}
	}
	return nil
}

func validateNegativeResourceOutputs(card *gamecards.Card, p *player.Player) error {
	if len(card.Behaviors) == 0 {
		return nil
	}

	resources := p.Resources().Get()

	for _, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}
		for _, output := range behavior.Outputs {
			if output.VariableAmount || output.Amount >= 0 {
				continue
			}
			if output.Target != "self-player" {
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
			required := -output.Amount
			if available < required {
				return fmt.Errorf("not enough %s: have %d, need %d", output.ResourceType, available, required)
			}
		}
	}
	return nil
}

func resolveChoicePolicyCount(policy *shared.ChoicePolicy, p *player.Player, registry gamecards.CardRegistryInterface) int {
	if policy == nil || policy.Select == nil {
		return 0
	}
	sel := policy.Select
	if sel.ResourceType == "tag" && sel.Tag != nil {
		return gamecards.CountPlayerTagsByType(p, registry, *sel.Tag)
	}
	return 0
}
