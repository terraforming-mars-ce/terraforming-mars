package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// BehaviorApplier handles applying card behavior inputs and outputs
// This is the single source of truth for all input/output application
type BehaviorApplier struct {
	player            *player.Player        // Player affected by the behavior (may be nil for game-only effects)
	game              *game.Game            // Game context for global params/tiles (may be nil for player-only effects)
	source            string                // Source identifier for logging (card name, action name, etc.)
	sourceCardID      string                // Card ID for self-card targeting (optional)
	targetCardIDs     []string              // Card IDs for any-card targeting (positional, one per any-card output)
	anyCardTargetIdx  int                   // Index into targetCardIDs, incremented each time an any-card output is processed
	targetPlayerID    string                // Player ID for any-player targeting (optional, set by caller)
	stealSourceCardID string                // Card ID to steal resources from for steal-from-any-card outputs (optional)
	sourceBehaviorIdx int                   // Behavior index for card draw selection tracking
	selectedAmount    int                   // Player-selected amount for variable-amount behaviors (0 = not applicable)
	actionPayment     *CardPayment          // Optional payment for action inputs with paymentAllowed (e.g., titanium for Water Import From Europa)
	cardRegistry      CardRegistryInterface // Card registry for tag counting in per conditions (optional)
	sourceType        shared.SourceType     // Source type for triggered effect classification
	deferredSteal     *shared.ResourceCondition
	logger            *zap.Logger
}

// DeferredSteal returns the deferred steal output, if any (for post-tile-placement processing)
func (a *BehaviorApplier) DeferredSteal() *shared.ResourceCondition {
	return a.deferredSteal
}

// NewBehaviorApplier creates a new behavior applier
// player and game can be nil if not needed for the specific operations
func NewBehaviorApplier(
	p *player.Player,
	g *game.Game,
	source string,
	logger *zap.Logger,
) *BehaviorApplier {
	return &BehaviorApplier{
		player: p,
		game:   g,
		source: source,
		logger: logger,
	}
}

// WithSourceCardID sets the source card ID for self-card targeting
func (a *BehaviorApplier) WithSourceCardID(cardID string) *BehaviorApplier {
	a.sourceCardID = cardID
	return a
}

// WithTargetCardIDs sets the target card IDs for any-card resource placement (positional, one per any-card output)
func (a *BehaviorApplier) WithTargetCardIDs(cardIDs []string) *BehaviorApplier {
	a.targetCardIDs = cardIDs
	a.anyCardTargetIdx = 0
	return a
}

// nextTargetCardID returns the next target card ID for any-card outputs and advances the index
func (a *BehaviorApplier) nextTargetCardID() string {
	if a.anyCardTargetIdx >= len(a.targetCardIDs) {
		return ""
	}
	id := a.targetCardIDs[a.anyCardTargetIdx]
	a.anyCardTargetIdx++
	return id
}

// WithCardRegistry sets the card registry for tag counting in scaled outputs
func (a *BehaviorApplier) WithCardRegistry(registry CardRegistryInterface) *BehaviorApplier {
	a.cardRegistry = registry
	return a
}

// WithTargetPlayerID sets the target player ID for any-player resource/production removal
func (a *BehaviorApplier) WithTargetPlayerID(playerID string) *BehaviorApplier {
	a.targetPlayerID = playerID
	return a
}

// WithStealSourceCardID sets the source card ID for steal-from-any-card outputs
func (a *BehaviorApplier) WithStealSourceCardID(cardID string) *BehaviorApplier {
	a.stealSourceCardID = cardID
	return a
}

// WithSourceBehaviorIndex sets the source behavior index for card draw selection tracking
func (a *BehaviorApplier) WithSourceBehaviorIndex(behaviorIndex int) *BehaviorApplier {
	a.sourceBehaviorIdx = behaviorIndex
	return a
}

// WithSelectedAmount sets the player-selected amount for variable-amount behaviors
func (a *BehaviorApplier) WithSelectedAmount(amount int) *BehaviorApplier {
	a.selectedAmount = amount
	return a
}

// WithActionPayment sets the payment for action inputs that have paymentAllowed
// (e.g., Water Import From Europa allows titanium as payment for the 12 M€ action cost)
func (a *BehaviorApplier) WithActionPayment(payment *CardPayment) *BehaviorApplier {
	a.actionPayment = payment
	return a
}

// WithSourceType sets the source type for triggered effect classification
func (a *BehaviorApplier) WithSourceType(sourceType shared.SourceType) *BehaviorApplier {
	a.sourceType = sourceType
	return a
}

// ApplyInputs validates player has required resources and deducts them
// Returns error if player is nil or insufficient resources
func (a *BehaviorApplier) ApplyInputs(
	ctx context.Context,
	inputs []shared.ResourceCondition,
) error {
	if len(inputs) == 0 {
		return nil
	}

	if a.player == nil {
		return fmt.Errorf("cannot apply inputs: no player context")
	}

	log := a.logger.With(
		zap.String("source", a.source),
		zap.Int("input_count", len(inputs)),
	)

	log.Debug("Processing behavior inputs")

	resources := a.player.Resources().Get()

	for _, input := range inputs {
		effectiveAmount := input.Amount
		if input.VariableAmount {
			effectiveAmount = input.Amount * a.selectedAmount
		}

		// Storage resource inputs (target: "self-card") deduct from card storage
		if input.Target == "self-card" && isStorageResourceType(input.ResourceType) {
			if a.sourceCardID == "" {
				return fmt.Errorf("cannot deduct from self-card: no source card ID")
			}
			storage := a.player.Resources().GetCardStorage(a.sourceCardID)
			if storage < effectiveAmount {
				return fmt.Errorf("insufficient %s on card: need %d, have %d", input.ResourceType, effectiveAmount, storage)
			}
			continue
		}

		// Credit inputs with paymentAllowed use CardPayment-style validation
		if input.ResourceType == shared.ResourceCredit && len(input.PaymentAllowed) > 0 {
			if err := a.validateActionPayment(effectiveAmount, input.PaymentAllowed, log); err != nil {
				return err
			}
			continue
		}

		switch input.ResourceType {
		case shared.ResourceCredit:
			if resources.Credits < effectiveAmount {
				return fmt.Errorf("insufficient credits: need %d, have %d", effectiveAmount, resources.Credits)
			}
		case shared.ResourceSteel:
			if resources.Steel < effectiveAmount {
				return fmt.Errorf("insufficient steel: need %d, have %d", effectiveAmount, resources.Steel)
			}
		case shared.ResourceTitanium:
			if resources.Titanium < effectiveAmount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", effectiveAmount, resources.Titanium)
			}
		case shared.ResourcePlant:
			if resources.Plants < effectiveAmount {
				return fmt.Errorf("insufficient plants: need %d, have %d", effectiveAmount, resources.Plants)
			}
		case shared.ResourceEnergy:
			if resources.Energy < effectiveAmount {
				return fmt.Errorf("insufficient energy: need %d, have %d", effectiveAmount, resources.Energy)
			}
		case shared.ResourceHeat:
			if resources.Heat < effectiveAmount {
				return fmt.Errorf("insufficient heat: need %d, have %d", effectiveAmount, resources.Heat)
			}
		case shared.ResourceEnergyProduction:
			production := a.player.Resources().Production()
			if production.Energy < effectiveAmount {
				return fmt.Errorf("insufficient energy production: need %d, have %d", effectiveAmount, production.Energy)
			}
		case shared.ResourceCreditProduction:
			production := a.player.Resources().Production()
			if production.Credits < effectiveAmount {
				return fmt.Errorf("insufficient credit production: need %d, have %d", effectiveAmount, production.Credits)
			}
		case shared.ResourceSteelProduction:
			production := a.player.Resources().Production()
			if production.Steel < effectiveAmount {
				return fmt.Errorf("insufficient steel production: need %d, have %d", effectiveAmount, production.Steel)
			}
		case shared.ResourceTitaniumProduction:
			production := a.player.Resources().Production()
			if production.Titanium < effectiveAmount {
				return fmt.Errorf("insufficient titanium production: need %d, have %d", effectiveAmount, production.Titanium)
			}
		case shared.ResourcePlantProduction:
			production := a.player.Resources().Production()
			if production.Plants < effectiveAmount {
				return fmt.Errorf("insufficient plant production: need %d, have %d", effectiveAmount, production.Plants)
			}
		case shared.ResourceHeatProduction:
			production := a.player.Resources().Production()
			if production.Heat < effectiveAmount {
				return fmt.Errorf("insufficient heat production: need %d, have %d", effectiveAmount, production.Heat)
			}
		default:
			log.Warn("Unhandled input type", zap.String("type", string(input.ResourceType)))
		}
	}

	for _, input := range inputs {
		effectiveAmount := input.Amount
		if input.VariableAmount {
			effectiveAmount = input.Amount * a.selectedAmount
		}

		if effectiveAmount == 0 {
			continue
		}

		// Storage resource inputs (target: "self-card") deduct from card storage
		if input.Target == "self-card" && isStorageResourceType(input.ResourceType) {
			a.player.Resources().AddToStorage(a.sourceCardID, -effectiveAmount)
			log.Debug("Deducted from card storage",
				zap.String("card_id", a.sourceCardID),
				zap.String("resource_type", string(input.ResourceType)),
				zap.Int("amount", effectiveAmount))
			continue
		}

		// Credit inputs with paymentAllowed use CardPayment-style deduction
		if input.ResourceType == shared.ResourceCredit && len(input.PaymentAllowed) > 0 {
			a.applyActionPayment(effectiveAmount, log)
			continue
		}

		switch input.ResourceType {
		case shared.ResourceCredit:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceCredit: -effectiveAmount,
			})
			log.Debug("Deducted credits", zap.Int("amount", effectiveAmount))

		case shared.ResourceSteel:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceSteel: -effectiveAmount,
			})
			log.Debug("Deducted steel", zap.Int("amount", effectiveAmount))

		case shared.ResourceTitanium:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceTitanium: -effectiveAmount,
			})
			log.Debug("Deducted titanium", zap.Int("amount", effectiveAmount))

		case shared.ResourcePlant:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourcePlant: -effectiveAmount,
			})
			log.Debug("Deducted plants", zap.Int("amount", effectiveAmount))

		case shared.ResourceEnergy:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceEnergy: -effectiveAmount,
			})
			log.Debug("Deducted energy", zap.Int("amount", effectiveAmount))

		case shared.ResourceHeat:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceHeat: -effectiveAmount,
			})
			log.Debug("Deducted heat", zap.Int("amount", effectiveAmount))

		case shared.ResourceEnergyProduction:
			a.player.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceEnergyProduction: -effectiveAmount,
			})
			log.Debug("Deducted energy production", zap.Int("amount", effectiveAmount))

		case shared.ResourceCreditProduction:
			a.player.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceCreditProduction: -effectiveAmount,
			})
			log.Debug("Deducted credit production", zap.Int("amount", effectiveAmount))

		case shared.ResourceSteelProduction:
			a.player.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceSteelProduction: -effectiveAmount,
			})
			log.Debug("Deducted steel production", zap.Int("amount", effectiveAmount))

		case shared.ResourceTitaniumProduction:
			a.player.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceTitaniumProduction: -effectiveAmount,
			})
			log.Debug("Deducted titanium production", zap.Int("amount", effectiveAmount))

		case shared.ResourcePlantProduction:
			a.player.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourcePlantProduction: -effectiveAmount,
			})
			log.Debug("Deducted plant production", zap.Int("amount", effectiveAmount))

		case shared.ResourceHeatProduction:
			a.player.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceHeatProduction: -effectiveAmount,
			})
			log.Debug("Deducted heat production", zap.Int("amount", effectiveAmount))
		}
	}

	return nil
}

// validateActionPayment validates that the action payment covers the required cost
func (a *BehaviorApplier) validateActionPayment(
	requiredAmount int,
	paymentAllowed []shared.ResourceType,
	log *zap.Logger,
) error {
	if a.actionPayment == nil {
		// No payment provided — fall back to checking if player has enough credits
		resources := a.player.Resources().Get()
		if resources.Credits < requiredAmount {
			return fmt.Errorf("insufficient credits: need %d, have %d", requiredAmount, resources.Credits)
		}
		return nil
	}

	payment := a.actionPayment

	if err := payment.Validate(); err != nil {
		return fmt.Errorf("invalid action payment: %w", err)
	}

	// Verify player has the resources
	resources := a.player.Resources().Get()
	if resources.Credits < payment.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", payment.Credits, resources.Credits)
	}

	// Build allowed resource set
	allowed := make(map[shared.ResourceType]bool)
	for _, rt := range paymentAllowed {
		allowed[rt] = true
	}

	// Validate titanium usage
	if payment.Titanium > 0 {
		if !allowed[shared.ResourceTitanium] {
			return fmt.Errorf("titanium is not allowed as payment for this action")
		}
		if resources.Titanium < payment.Titanium {
			return fmt.Errorf("insufficient titanium: need %d, have %d", payment.Titanium, resources.Titanium)
		}
	}

	// Validate steel usage
	if payment.Steel > 0 {
		if !allowed[shared.ResourceSteel] {
			return fmt.Errorf("steel is not allowed as payment for this action")
		}
		if resources.Steel < payment.Steel {
			return fmt.Errorf("insufficient steel: need %d, have %d", payment.Steel, resources.Steel)
		}
	}

	// Calculate total payment value using player's substitution rates
	playerSubstitutes := a.player.Resources().PaymentSubstitutes()
	totalValue := payment.TotalValue(playerSubstitutes, nil)

	if totalValue < requiredAmount {
		return fmt.Errorf("payment insufficient: action costs %d MC, payment provides %d MC", requiredAmount, totalValue)
	}

	log.Debug("Validated action payment",
		zap.Int("required", requiredAmount),
		zap.Int("credits", payment.Credits),
		zap.Int("titanium", payment.Titanium),
		zap.Int("steel", payment.Steel),
		zap.Int("total_value", totalValue))

	return nil
}

// applyActionPayment deducts resources according to the action payment
func (a *BehaviorApplier) applyActionPayment(
	requiredAmount int,
	log *zap.Logger,
) {
	if a.actionPayment == nil {
		// No payment struct — just deduct credits
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: -requiredAmount,
		})
		log.Debug("Deducted credits (no action payment)", zap.Int("amount", requiredAmount))
		return
	}

	payment := a.actionPayment

	if payment.Credits > 0 {
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: -payment.Credits,
		})
		log.Debug("Deducted credits from action payment", zap.Int("amount", payment.Credits))
	}
	if payment.Titanium > 0 {
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceTitanium: -payment.Titanium,
		})
		log.Debug("Deducted titanium from action payment", zap.Int("amount", payment.Titanium))
	}
	if payment.Steel > 0 {
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceSteel: -payment.Steel,
		})
		log.Debug("Deducted steel from action payment", zap.Int("amount", payment.Steel))
	}
}

// isStorageResourceType returns true for resource types that are stored on cards
func isStorageResourceType(rt shared.ResourceType) bool {
	switch rt {
	case shared.ResourceMicrobe, shared.ResourceAnimal, shared.ResourceFloater,
		shared.ResourceScience, shared.ResourceAsteroid, shared.ResourceFighter, shared.ResourceDisease:
		return true
	}
	return false
}

// isEffectOutputType returns true for output types that represent persistent effects
// rather than immediate resource gains (these get their own "Effect:" notification)
func isEffectOutputType(rt shared.ResourceType) bool {
	switch rt {
	case shared.ResourceDiscount, shared.ResourcePaymentSubstitute, shared.ResourceValueModifier,
		shared.ResourceGlobalParameterLenience,
		shared.ResourceStoragePaymentSubstitute, shared.ResourceOceanAdjacencyBonus,
		shared.ResourceDefense, shared.ResourceActionReuse:
		return true
	}
	return false
}

// ApplyOutputs applies resource gains, production changes, global params, tile placements
// Returns error if required context (player/game) is missing for the operation
func (a *BehaviorApplier) ApplyOutputs(
	ctx context.Context,
	outputs []shared.ResourceCondition,
) error {
	_, err := a.ApplyOutputsAndGetCalculated(ctx, outputs)
	return err
}

// ApplyOutputsAndGetCalculated applies outputs and returns the calculated values
// This is useful for logging scaled outputs (e.g., "+1 MC per 2 plant tags" becomes "+3 MC")
func (a *BehaviorApplier) ApplyOutputsAndGetCalculated(
	ctx context.Context,
	outputs []shared.ResourceCondition,
) ([]shared.CalculatedOutput, error) {
	if len(outputs) == 0 {
		return nil, nil
	}

	log := a.logger.With(
		zap.String("source", a.source),
		zap.Int("output_count", len(outputs)),
	)

	log.Debug("Processing behavior outputs")

	var calculatedOutputs []shared.CalculatedOutput
	var notificationOutputs []shared.CalculatedOutput

	for _, output := range outputs {
		// Calculate the actual amount if this output has a Per condition
		actualAmount := output.Amount
		isScaled := false

		if output.Per != nil && a.player != nil && a.game != nil {
			// Calculate scaled amount based on Per condition
			count := a.countPerCondition(output.Per)
			if output.Per.Amount > 0 {
				multiplier := count / output.Per.Amount
				actualAmount = output.Amount * multiplier
				isScaled = true
				log.Debug("Calculated scaled output",
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("base_amount", output.Amount),
					zap.Int("count", count),
					zap.Int("per_amount", output.Per.Amount),
					zap.Int("calculated_amount", actualAmount))
			}
		}

		// Apply variable amount multiplier (player-selected amount)
		if output.VariableAmount {
			actualAmount = output.Amount * a.selectedAmount
			isScaled = true
			log.Debug("Applied variable amount",
				zap.String("resource_type", string(output.ResourceType)),
				zap.Int("base_amount", output.Amount),
				zap.Int("selected_amount", a.selectedAmount),
				zap.Int("calculated_amount", actualAmount))
		}

		// Create a modified output with the calculated amount
		modifiedOutput := output
		modifiedOutput.Amount = actualAmount

		if err := a.applyOutput(ctx, modifiedOutput, log); err != nil {
			return calculatedOutputs, err
		}

		// Track for state diff log (existing behavior)
		if isScaled || actualAmount != 0 {
			calculatedOutputs = append(calculatedOutputs, shared.CalculatedOutput{
				ResourceType: string(output.ResourceType),
				Amount:       actualAmount,
				IsScaled:     isScaled,
			})
		}

		// Track non-zero resource outputs for triggered effect notifications
		// Skip effect-type outputs (discount, payment-substitute, etc.) since they get
		// their own "Effect:" notification via SourceTypeEffectAdded
		if actualAmount != 0 && !isEffectOutputType(output.ResourceType) {
			resourceType := string(output.ResourceType)
			if resourceType == string(shared.ResourceCardResource) {
				resourceType = a.resolveCardResourceType()
			}
			notificationOutputs = append(notificationOutputs, shared.CalculatedOutput{
				ResourceType: resourceType,
				Amount:       actualAmount,
				IsScaled:     isScaled,
			})
		}
	}

	if a.game != nil && a.player != nil && len(outputs) > 0 {
		a.game.AddTriggeredEffect(shared.TriggeredEffect{
			CardName:          a.source,
			PlayerID:          a.player.ID(),
			SourceType:        a.sourceType,
			Outputs:           outputs,
			CalculatedOutputs: notificationOutputs,
		})
	}

	return calculatedOutputs, nil
}

// resolveCardResourceType resolves "card-resource" to the actual storage type of the last consumed target card
func (a *BehaviorApplier) resolveCardResourceType() string {
	if a.anyCardTargetIdx == 0 || a.cardRegistry == nil {
		return string(shared.ResourceCardResource)
	}
	lastTargetID := a.targetCardIDs[a.anyCardTargetIdx-1]
	targetCard, err := a.cardRegistry.GetByID(lastTargetID)
	if err != nil || targetCard.ResourceStorage == nil {
		return string(shared.ResourceCardResource)
	}
	return string(targetCard.ResourceStorage.Type)
}

func (a *BehaviorApplier) countPerCondition(per *shared.PerCondition) int {
	var b *board.Board
	var allPlayers []*player.Player
	if a.game != nil {
		b = a.game.Board()
		allPlayers = a.game.GetAllPlayers()
	}
	return CountPerCondition(per, a.sourceCardID, a.player, b, a.cardRegistry, allPlayers)
}

// ApplyCardDrawOutputs processes card-peek/take/buy outputs together
// Returns true if a pending selection was created (caller should defer action consumption)
func (a *BehaviorApplier) ApplyCardDrawOutputs(
	ctx context.Context,
	outputs []shared.ResourceCondition,
) (bool, error) {
	log := a.logger.With(
		zap.String("source", a.source),
		zap.String("method", "ApplyCardDrawOutputs"),
	)

	// Scan outputs for card-peek, card-take, card-buy
	var peekAmount, takeAmount, buyAmount int
	var isPrelude bool
	for _, output := range outputs {
		switch output.ResourceType {
		case shared.ResourceCardPeek:
			peekAmount += output.Amount
		case shared.ResourceCardTake:
			takeAmount += output.Amount
		case shared.ResourceCardBuy:
			buyAmount += output.Amount
		}
		if hasPreludeCardType(output.Selectors) {
			isPrelude = true
		}
	}

	// If no card-peek found, nothing to do
	if peekAmount == 0 {
		return false, nil
	}

	if a.player == nil {
		return false, fmt.Errorf("cannot apply card draw outputs: no player context")
	}
	if a.game == nil {
		return false, fmt.Errorf("cannot apply card draw outputs: no game context")
	}

	// Draw cards from the appropriate deck
	var drawnCards []string
	var err error
	if isPrelude {
		drawnCards, err = a.game.Deck().DrawPreludeCards(ctx, peekAmount)
	} else {
		drawnCards, err = a.game.Deck().DrawProjectCards(ctx, peekAmount)
	}
	if err != nil {
		return false, fmt.Errorf("failed to draw cards: %w", err)
	}

	log.Debug("Drew cards for peek selection",
		zap.Int("peek_amount", peekAmount),
		zap.Int("take_amount", takeAmount),
		zap.Int("buy_amount", buyAmount),
		zap.Bool("is_prelude", isPrelude),
		zap.Strings("drawn_cards", drawnCards))

	// Create pending card draw selection
	selection := &shared.PendingCardDrawSelection{
		AvailableCards:      drawnCards,
		FreeTakeCount:       takeAmount,
		MaxBuyCount:         buyAmount,
		CardBuyCost:         3, // Standard card cost
		Source:              a.source,
		SourceCardID:        a.sourceCardID,
		SourceBehaviorIndex: a.sourceBehaviorIdx,
		PlayAsPrelude:       isPrelude,
	}

	// Set on player
	a.player.Selection().SetPendingCardDrawSelection(selection)

	log.Debug("Created pending card draw selection",
		zap.String("source", a.source),
		zap.String("source_card_id", a.sourceCardID),
		zap.Int("source_behavior_index", a.sourceBehaviorIdx),
		zap.Int("available_cards", len(drawnCards)),
		zap.Int("free_take", takeAmount),
		zap.Int("max_buy", buyAmount))

	return true, nil
}

// stealAnyPlayerResource removes resources from the target player and adds them to self
func (a *BehaviorApplier) stealAnyPlayerResource(
	resourceType shared.ResourceType,
	amount int,
	log *zap.Logger,
) error {
	if a.targetPlayerID == "" {
		log.Debug("Skipping steal: no target player (solo mode)",
			zap.String("resource_type", string(resourceType)))
		return nil
	}
	if a.game == nil {
		return fmt.Errorf("cannot steal resource: no game context")
	}
	if a.player == nil {
		return fmt.Errorf("cannot steal resource: no player context")
	}
	targetPlayer, err := a.game.GetPlayer(a.targetPlayerID)
	if err != nil {
		return fmt.Errorf("target player not found: %w", err)
	}

	resources := targetPlayer.Resources().Get()
	var current int
	switch resourceType {
	case shared.ResourceCredit:
		current = resources.Credits
	case shared.ResourceSteel:
		current = resources.Steel
	case shared.ResourceTitanium:
		current = resources.Titanium
	case shared.ResourcePlant:
		current = resources.Plants
	case shared.ResourceEnergy:
		current = resources.Energy
	case shared.ResourceHeat:
		current = resources.Heat
	}

	stolenAmount := min(amount, current)

	if stolenAmount > 0 {
		targetPlayer.Resources().Add(map[shared.ResourceType]int{
			resourceType: -stolenAmount,
		})
		a.player.Resources().Add(map[shared.ResourceType]int{
			resourceType: stolenAmount,
		})
	}

	log.Debug("Stole resource from target player",
		zap.String("target_player_id", a.targetPlayerID),
		zap.String("resource_type", string(resourceType)),
		zap.Int("requested", amount),
		zap.Int("stolen", stolenAmount))
	return nil
}

// applyAnyPlayerResource removes resources from the target player (clamped to what they have)
func (a *BehaviorApplier) applyAnyPlayerResource(
	resourceType shared.ResourceType,
	amount int,
	log *zap.Logger,
) error {
	if a.targetPlayerID == "" {
		log.Debug("Skipping any-player resource removal: no target player (solo mode)",
			zap.String("resource_type", string(resourceType)))
		return nil
	}
	if a.game == nil {
		return fmt.Errorf("cannot apply any-player resource: no game context")
	}
	targetPlayer, err := a.game.GetPlayer(a.targetPlayerID)
	if err != nil {
		return fmt.Errorf("target player not found: %w", err)
	}

	resources := targetPlayer.Resources().Get()
	var current int
	switch resourceType {
	case shared.ResourceCredit:
		current = resources.Credits
	case shared.ResourceSteel:
		current = resources.Steel
	case shared.ResourceTitanium:
		current = resources.Titanium
	case shared.ResourcePlant:
		current = resources.Plants
	case shared.ResourceEnergy:
		current = resources.Energy
	case shared.ResourceHeat:
		current = resources.Heat
	}

	// Card data uses negative amounts for removal (e.g., Deimos Down: amount=-8).
	// Normalize to positive for clamping.
	absAmount := amount
	if absAmount < 0 {
		absAmount = -absAmount
	}

	removeAmount := min(absAmount, current)

	if removeAmount > 0 {
		targetPlayer.Resources().Add(map[shared.ResourceType]int{
			resourceType: -removeAmount,
		})
	}

	log.Debug("Removed resource from target player",
		zap.String("target_player_id", a.targetPlayerID),
		zap.String("resource_type", string(resourceType)),
		zap.Int("requested", absAmount),
		zap.Int("removed", removeAmount))
	return nil
}

// applyAnyPlayerProduction applies production changes to the target player.
// Card data uses negative amounts for decreases (e.g., Asteroid Mining Consortium: amount=-1).
// The amount is applied directly via AddProduction (which handles clamping to minimums).
func (a *BehaviorApplier) applyAnyPlayerProduction(
	productionType shared.ResourceType,
	amount int,
	log *zap.Logger,
) error {
	if a.targetPlayerID == "" {
		log.Debug("Skipping any-player production change: no target player (solo mode)",
			zap.String("production_type", string(productionType)))
		return nil
	}
	if a.game == nil {
		return fmt.Errorf("cannot apply any-player production: no game context")
	}
	targetPlayer, err := a.game.GetPlayer(a.targetPlayerID)
	if err != nil {
		return fmt.Errorf("target player not found: %w", err)
	}

	targetPlayer.Resources().AddProduction(map[shared.ResourceType]int{
		productionType: amount,
	})

	log.Debug("Applied production change to target player",
		zap.String("target_player_id", a.targetPlayerID),
		zap.String("production_type", string(productionType)),
		zap.Int("amount", amount))
	return nil
}

// applyOutput applies a single output
func (a *BehaviorApplier) applyOutput(
	ctx context.Context,
	output shared.ResourceCondition,
	log *zap.Logger,
) error {
	switch output.ResourceType {
	case shared.ResourceCredit:
		if output.Target == "steal-any-player" && output.TargetRestriction != nil && output.TargetRestriction.Adjacent == "self-card" {
			a.deferredSteal = &output
			log.Debug("Deferred adjacent steal for post-tile-placement",
				zap.Int("amount", output.Amount))
			return nil
		}
		if output.Target == "steal-any-player" {
			return a.stealAnyPlayerResource(shared.ResourceCredit, output.Amount, log)
		}
		if output.Target == "any-player" {
			return a.applyAnyPlayerResource(shared.ResourceCredit, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply credits: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: output.Amount,
		})
		log.Debug("Added credits", zap.Int("amount", output.Amount))

	case shared.ResourceSteel:
		if output.Target == "steal-any-player" {
			return a.stealAnyPlayerResource(shared.ResourceSteel, output.Amount, log)
		}
		if output.Target == "any-player" {
			return a.applyAnyPlayerResource(shared.ResourceSteel, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply steel: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceSteel: output.Amount,
		})
		log.Debug("Added steel", zap.Int("amount", output.Amount))

	case shared.ResourceTitanium:
		if output.Target == "steal-any-player" {
			return a.stealAnyPlayerResource(shared.ResourceTitanium, output.Amount, log)
		}
		if output.Target == "any-player" {
			return a.applyAnyPlayerResource(shared.ResourceTitanium, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply titanium: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceTitanium: output.Amount,
		})
		log.Debug("Added titanium", zap.Int("amount", output.Amount))

	case shared.ResourcePlant:
		if output.Target == "steal-any-player" {
			return a.stealAnyPlayerResource(shared.ResourcePlant, output.Amount, log)
		}
		if output.Target == "any-player" {
			return a.applyAnyPlayerResource(shared.ResourcePlant, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply plants: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourcePlant: output.Amount,
		})
		log.Debug("Added plants", zap.Int("amount", output.Amount))

	case shared.ResourceEnergy:
		if output.Target == "steal-any-player" {
			return a.stealAnyPlayerResource(shared.ResourceEnergy, output.Amount, log)
		}
		if output.Target == "any-player" {
			return a.applyAnyPlayerResource(shared.ResourceEnergy, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply energy: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceEnergy: output.Amount,
		})
		log.Debug("Added energy", zap.Int("amount", output.Amount))

	case shared.ResourceHeat:
		if output.Target == "steal-any-player" {
			return a.stealAnyPlayerResource(shared.ResourceHeat, output.Amount, log)
		}
		if output.Target == "any-player" {
			return a.applyAnyPlayerResource(shared.ResourceHeat, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply heat: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceHeat: output.Amount,
		})
		log.Debug("Added heat", zap.Int("amount", output.Amount))

	case shared.ResourceCreditProduction:
		if output.Target == "any-player" {
			return a.applyAnyPlayerProduction(shared.ResourceCreditProduction, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply credits production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceCreditProduction: output.Amount,
		})
		log.Debug("Added credits production", zap.Int("amount", output.Amount))

	case shared.ResourceSteelProduction:
		if output.Target == "any-player" {
			return a.applyAnyPlayerProduction(shared.ResourceSteelProduction, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply steel production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceSteelProduction: output.Amount,
		})
		log.Debug("Added steel production", zap.Int("amount", output.Amount))

	case shared.ResourceTitaniumProduction:
		if output.Target == "any-player" {
			return a.applyAnyPlayerProduction(shared.ResourceTitaniumProduction, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply titanium production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceTitaniumProduction: output.Amount,
		})
		log.Debug("Added titanium production", zap.Int("amount", output.Amount))

	case shared.ResourcePlantProduction:
		if output.Target == "any-player" {
			return a.applyAnyPlayerProduction(shared.ResourcePlantProduction, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply plants production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourcePlantProduction: output.Amount,
		})
		log.Debug("Added plants production", zap.Int("amount", output.Amount))

	case shared.ResourceEnergyProduction:
		if output.Target == "any-player" {
			return a.applyAnyPlayerProduction(shared.ResourceEnergyProduction, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply energy production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceEnergyProduction: output.Amount,
		})
		log.Debug("Added energy production", zap.Int("amount", output.Amount))

	case shared.ResourceHeatProduction:
		if output.Target == "any-player" {
			return a.applyAnyPlayerProduction(shared.ResourceHeatProduction, output.Amount, log)
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply heat production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceHeatProduction: output.Amount,
		})
		log.Debug("Added heat production", zap.Int("amount", output.Amount))

	case shared.ResourceTR:
		if a.player == nil {
			return fmt.Errorf("cannot apply terraform rating: no player context")
		}
		a.player.Resources().UpdateTerraformRating(output.Amount)
		log.Debug("Added terraform rating", zap.Int("amount", output.Amount))

	case shared.ResourceOxygen:
		if a.game == nil {
			return fmt.Errorf("cannot apply oxygen: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseOxygen(ctx, output.Amount, a.player.ID())
		if err != nil {
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}
		if actualSteps > 0 && a.player != nil {
			a.player.Resources().UpdateTerraformRating(actualSteps)
		}
		log.Debug("Increased oxygen", zap.Int("steps", actualSteps), zap.Int("tr_gained", actualSteps))

	case shared.ResourceTemperature:
		if a.game == nil {
			return fmt.Errorf("cannot apply temperature: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseTemperature(ctx, output.Amount, a.player.ID())
		if err != nil {
			return fmt.Errorf("failed to increase temperature: %w", err)
		}
		if actualSteps > 0 && a.player != nil {
			a.player.Resources().UpdateTerraformRating(actualSteps)
		}
		log.Debug("Increased temperature", zap.Int("steps", actualSteps), zap.Int("tr_gained", actualSteps))

	case shared.ResourceVenus:
		if a.game == nil {
			return fmt.Errorf("cannot apply venus: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseVenus(ctx, output.Amount, a.player.ID())
		if err != nil {
			return fmt.Errorf("failed to increase venus: %w", err)
		}
		if actualSteps > 0 && a.player != nil {
			a.player.Resources().UpdateTerraformRating(actualSteps)
		}
		log.Debug("Increased venus", zap.Int("steps", actualSteps), zap.Int("tr_gained", actualSteps))

	case shared.ResourceLandClaim:
		if a.game == nil {
			return fmt.Errorf("cannot apply land claim: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply land claim: no player context")
		}

		// Build array of land-claim tile types to append (for multiple placements)
		tileTypes := make([]string, output.Amount)
		for i := 0; i < output.Amount; i++ {
			tileTypes[i] = "land-claim"
		}

		// Atomically append to queue (thread-safe)
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, nil); err != nil {
			return fmt.Errorf("failed to append land claim to pending tile selection queue: %w", err)
		}

		log.Debug("Added land claim tile selection to queue",
			zap.Int("count", output.Amount))

	case shared.ResourceTilePlacement:
		if a.game == nil {
			return fmt.Errorf("cannot apply tile placement: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply tile placement: no player context")
		}

		tileType := output.TileType
		if tileType == "" {
			return fmt.Errorf("tile-placement output missing tileType field")
		}

		// Build array of tile types to append (for multiple placements)
		tileTypes := make([]string, output.Amount)
		for i := 0; i < output.Amount; i++ {
			tileTypes[i] = tileType
		}

		// Convert tile restrictions to shared type if present
		var tileRestrictions *shared.TileRestrictions
		if output.TileRestrictions != nil {
			tileRestrictions = &shared.TileRestrictions{
				BoardTags:         output.TileRestrictions.BoardTags,
				Adjacency:         output.TileRestrictions.Adjacency,
				OnTileType:        output.TileRestrictions.OnTileType,
				AdjacentToType:    output.TileRestrictions.AdjacentToType,
				MinAdjacentOfType: output.TileRestrictions.MinAdjacentOfType,
				AdjacentToOwned:   output.TileRestrictions.AdjacentToOwned,
				OnBonusType:       output.TileRestrictions.OnBonusType,
			}
		}

		// Atomically append to queue (thread-safe)
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, tileRestrictions); err != nil {
			return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
		}

		log.Debug("Added special tile placements to queue",
			zap.String("tile_type", tileType),
			zap.Int("count", output.Amount),
			zap.Any("tile_restrictions", tileRestrictions))

	case shared.ResourceCityPlacement, shared.ResourceGreeneryPlacement, shared.ResourceOceanPlacement, shared.ResourceVolcanoPlacement:
		if a.game == nil {
			return fmt.Errorf("cannot apply tile placement: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply tile placement: no player context")
		}

		// Map resource type to tile type string
		var tileType string
		switch output.ResourceType {
		case shared.ResourceCityPlacement:
			tileType = "city"
		case shared.ResourceGreeneryPlacement:
			tileType = "greenery"
		case shared.ResourceOceanPlacement:
			tileType = "ocean"
		case shared.ResourceVolcanoPlacement:
			tileType = "volcano"
		}

		// Build array of tile types to append (for multiple placements)
		tileTypes := make([]string, output.Amount)
		for i := 0; i < output.Amount; i++ {
			tileTypes[i] = tileType
		}

		// Convert tile restrictions to shared type if present
		var tileRestrictions *shared.TileRestrictions
		if output.TileRestrictions != nil {
			tileRestrictions = &shared.TileRestrictions{
				BoardTags:         output.TileRestrictions.BoardTags,
				Adjacency:         output.TileRestrictions.Adjacency,
				OnTileType:        output.TileRestrictions.OnTileType,
				AdjacentToType:    output.TileRestrictions.AdjacentToType,
				MinAdjacentOfType: output.TileRestrictions.MinAdjacentOfType,
				AdjacentToOwned:   output.TileRestrictions.AdjacentToOwned,
				OnBonusType:       output.TileRestrictions.OnBonusType,
			}
		}

		// For greenery placements, enforce adjacency to owned tiles (TM rules)
		// unless the card explicitly overrides placement (e.g., Mangrove with onTileType: "ocean")
		if output.ResourceType == shared.ResourceGreeneryPlacement {
			if tileRestrictions == nil {
				tileRestrictions = &shared.TileRestrictions{
					AdjacentToOwned: true,
				}
			} else if tileRestrictions.OnTileType == "" {
				tileRestrictions.AdjacentToOwned = true
			}
		}

		// Atomically append to queue (thread-safe)
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, tileRestrictions); err != nil {
			return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
		}

		log.Debug("Added tile placements to queue",
			zap.String("tile_type", tileType),
			zap.Int("count", output.Amount),
			zap.Any("tile_restrictions", tileRestrictions))

	case shared.ResourcePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply payment substitute: no player context")
		}
		// Extract resource type from selectors (e.g., "heat" for Helion)
		resources := GetResourcesFromSelectors(output.Selectors)
		if len(resources) > 0 {
			resourceType := shared.ResourceType(resources[0])
			a.player.Resources().AddPaymentSubstitute(resourceType, output.Amount)
			log.Debug("Added payment substitute",
				zap.String("resource_type", string(resourceType)),
				zap.Int("conversion_rate", output.Amount))
		} else {
			log.Warn("payment-substitute output missing selectors with resources")
		}

	case shared.ResourceDiscount:
		// Discounts are registered as effects and handled by RequirementModifierCalculator
		// The effect registration happens when the card is played (auto effects are added to player.Effects)
		// The calculator then computes the actual modifiers from all effects
		log.Debug("Discount effect registered",
			zap.Int("amount", output.Amount),
			zap.Any("selectors", output.Selectors))

	case shared.ResourceGlobalParameterLenience:
		// Lenience is registered as an effect and handled by RequirementModifierCalculator
		// during requirement validation. It widens the min/max window for global parameter checks.
		log.Debug("Global parameter lenience effect registered",
			zap.Int("amount", output.Amount),
			zap.String("temporary", output.Temporary))

	case shared.ResourceValueModifier:
		if a.player == nil {
			return fmt.Errorf("cannot apply value modifier: no player context")
		}
		// Apply value modifier to each affected resource (e.g., titanium +1, steel +1)
		for _, resourceStr := range GetResourcesFromSelectors(output.Selectors) {
			resourceType := shared.ResourceType(resourceStr)
			a.player.Resources().AddValueModifier(resourceType, output.Amount)
			log.Debug("Added resource value modifier",
				zap.String("resource_type", string(resourceType)),
				zap.Int("modifier_amount", output.Amount))
		}

	case shared.ResourceStoragePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply storage payment substitute: no player context")
		}
		if a.sourceCardID == "" {
			log.Warn("storage-payment-substitute output missing source card ID")
			return nil
		}
		a.player.Resources().AddStoragePaymentSubstitute(shared.StoragePaymentSubstitute{
			CardID:         a.sourceCardID,
			ResourceType:   shared.ResourceFloater, // Default; could be extended via selectors
			ConversionRate: output.Amount,
			Selectors:      output.Selectors,
		})
		log.Debug("Added storage payment substitute",
			zap.String("card_id", a.sourceCardID),
			zap.Int("conversion_rate", output.Amount),
			zap.Any("selectors", output.Selectors))

	case shared.ResourceCardResource:
		// Generic "card-resource" — add resources of whatever type the target card stores
		if a.player == nil {
			return fmt.Errorf("cannot apply card-resource: no player context")
		}
		targetID := a.nextTargetCardID()
		if targetID == "" {
			log.Warn("No target card specified for card-resource output")
			return fmt.Errorf("no target card specified for card-resource output")
		}
		if a.cardRegistry == nil {
			return fmt.Errorf("cannot apply card-resource: no card registry")
		}
		targetCard, err := a.cardRegistry.GetByID(targetID)
		if err != nil {
			return fmt.Errorf("target card not found in registry: %w", err)
		}
		if targetCard.ResourceStorage == nil {
			return fmt.Errorf("target card %s has no resource storage", targetID)
		}
		a.player.Resources().AddToStorage(targetID, output.Amount)
		log.Debug("Added card-resource to target card storage",
			zap.String("card_id", targetID),
			zap.String("storage_type", string(targetCard.ResourceStorage.Type)),
			zap.Int("amount", output.Amount))

	case shared.ResourceAnimal, shared.ResourceMicrobe, shared.ResourceFloater,
		shared.ResourceFighter, shared.ResourceScience, shared.ResourceAsteroid, shared.ResourceDisease:
		if a.player == nil {
			return fmt.Errorf("cannot apply card resource: no player context")
		}

		target := output.Target
		switch target {
		case "self-card":
			if a.sourceCardID == "" {
				log.Warn("Cannot place resource on self-card: no source card ID",
					zap.String("resource_type", string(output.ResourceType)))
				return nil
			}
			a.player.Resources().AddToStorage(a.sourceCardID, output.Amount)
			log.Debug("Added resource to card storage",
				zap.String("card_id", a.sourceCardID),
				zap.String("resource_type", string(output.ResourceType)),
				zap.Int("amount", output.Amount))

		case "steal-from-any-card":
			if a.stealSourceCardID == "" {
				return fmt.Errorf("steal-from-any-card requires a source card ID")
			}
			if a.game == nil {
				return fmt.Errorf("cannot steal from card: no game context")
			}
			stolenAmount := 0
			for _, p := range a.game.GetAllPlayers() {
				storage := p.Resources().GetCardStorage(a.stealSourceCardID)
				if storage > 0 {
					stolenAmount = min(output.Amount, storage)
					p.Resources().AddToStorage(a.stealSourceCardID, -stolenAmount)
					log.Debug("Stole resource from card",
						zap.String("source_card_id", a.stealSourceCardID),
						zap.String("owner_player_id", p.ID()),
						zap.String("resource_type", string(output.ResourceType)),
						zap.Int("amount", stolenAmount))
					break
				}
			}
			if stolenAmount > 0 && a.sourceCardID != "" {
				a.player.Resources().AddToStorage(a.sourceCardID, stolenAmount)
				log.Debug("Added stolen resource to self card",
					zap.String("card_id", a.sourceCardID),
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("amount", stolenAmount))
			}

		case "any-card":
			targetID := a.nextTargetCardID()
			if targetID == "" {
				log.Warn("No target card for any-card resource placement — resources lost",
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("amount", output.Amount))
				return nil
			}
			if a.cardRegistry != nil {
				targetCard, err := a.cardRegistry.GetByID(targetID)
				if err != nil {
					return fmt.Errorf("target card %s not found in registry: %w", targetID, err)
				}
				if targetCard.ResourceStorage == nil {
					return fmt.Errorf("target card %s has no resource storage", targetID)
				}
				if targetCard.ResourceStorage.Type != output.ResourceType {
					return fmt.Errorf("target card %s stores %s, cannot add %s", targetID, targetCard.ResourceStorage.Type, output.ResourceType)
				}
			}
			a.player.Resources().AddToStorage(targetID, output.Amount)
			log.Debug("Added resource to target card storage",
				zap.String("card_id", targetID),
				zap.String("resource_type", string(output.ResourceType)),
				zap.Int("amount", output.Amount))

		default:
			// Default to self-card if target is empty and sourceCardID is set
			if a.sourceCardID != "" {
				a.player.Resources().AddToStorage(a.sourceCardID, output.Amount)
				log.Debug("Added resource to card storage (default to self)",
					zap.String("card_id", a.sourceCardID),
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("amount", output.Amount))
			} else {
				log.Warn("Unhandled target for card resource",
					zap.String("target", target),
					zap.String("resource_type", string(output.ResourceType)))
			}
		}

	case shared.ResourceCardDraw:
		if a.game == nil || a.player == nil {
			return fmt.Errorf("cannot apply card-draw: missing game or player context")
		}

		if output.Target == "all-opponents" {
			for _, opponent := range a.game.GetAllPlayers() {
				if opponent.ID() == a.player.ID() {
					continue
				}
				drawnCards, err := a.game.Deck().DrawProjectCards(ctx, output.Amount)
				if err != nil {
					log.Warn("Failed to draw cards for opponent",
						zap.String("opponent_id", opponent.ID()),
						zap.Error(err))
					continue
				}
				for _, cardID := range drawnCards {
					opponent.Hand().AddCard(cardID)
				}
				log.Debug("Opponent drew cards",
					zap.String("opponent_id", opponent.ID()),
					zap.Int("amount", len(drawnCards)))
			}
		} else if HasCardSelectors(output.Selectors) && a.cardRegistry != nil {
			matcher := func(cardID string) bool {
				card, err := a.cardRegistry.GetByID(cardID)
				if err != nil || card == nil {
					return false
				}
				return MatchesAnySelector(card, output.Selectors)
			}
			matched, discarded, err := a.game.Deck().DrawProjectCardsUntilMatching(ctx, output.Amount, matcher)
			if err != nil {
				log.Warn("Failed to draw matching cards", zap.Error(err))
				return nil
			}
			for _, cardID := range matched {
				a.player.Hand().AddCard(cardID)
			}
			if len(discarded) > 0 {
				_ = a.game.Deck().Discard(ctx, discarded)
			}
			log.Debug("Drew matching cards (draw-until)",
				zap.Int("matched", len(matched)),
				zap.Int("discarded", len(discarded)))
		} else {
			drawnCards, err := a.game.Deck().DrawProjectCards(ctx, output.Amount)
			if err != nil {
				log.Warn("Failed to draw cards", zap.Error(err))
				return nil
			}
			for _, cardID := range drawnCards {
				a.player.Hand().AddCard(cardID)
			}
			log.Debug("Drew cards and added to hand",
				zap.Int("amount", len(drawnCards)))
		}

	case shared.ResourceCardDiscard:
		log.Debug("Skipping card-discard output (handled at action layer)")

	case shared.ResourceCardPeek, shared.ResourceCardTake, shared.ResourceCardBuy:
		// Handled by ApplyCardDrawOutputs - skip here
		log.Debug("Skipping card draw output (handled by ApplyCardDrawOutputs)",
			zap.String("type", string(output.ResourceType)),
			zap.Int("amount", output.Amount))

	case shared.ResourceExtraActions:
		if a.game == nil {
			return fmt.Errorf("cannot apply extra actions: no game context")
		}
		currentTurn := a.game.CurrentTurn()
		if currentTurn != nil {
			currentTurn.AddExtraActions(output.Amount)
		}
		log.Debug("Granted extra actions", zap.Int("amount", output.Amount))

	case shared.ResourceBonusTags:
		if a.player == nil {
			return fmt.Errorf("cannot apply bonus tags: no player context")
		}
		if output.Per != nil && output.Per.Tag != nil {
			tagToCount := *output.Per.Tag
			tagToGrant := shared.CardTag(output.ResourceType)
			if len(output.Selectors) > 0 && len(output.Selectors[0].Tags) > 0 {
				tagToGrant = output.Selectors[0].Tags[0]
			}
			var tagCount int
			if a.cardRegistry != nil {
				tagCount = CountPlayerTagsByType(a.player, a.cardRegistry, tagToCount)
			}
			bonusCount := tagCount * output.Amount
			if bonusCount > 0 {
				a.player.AddBonusTags(tagToGrant, bonusCount)
			}
			log.Debug("Added bonus tags",
				zap.String("tag_type", string(tagToGrant)),
				zap.Int("count", bonusCount),
				zap.String("per_tag", string(tagToCount)),
				zap.Int("tag_count", tagCount))
		}

	case shared.ResourceTileDestruction:
		if a.game == nil {
			return fmt.Errorf("cannot apply tile destruction: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply tile destruction: no player context")
		}
		tileTypes := make([]string, output.Amount)
		for i := 0; i < output.Amount; i++ {
			tileTypes[i] = "tile-destruction"
		}
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, nil); err != nil {
			return fmt.Errorf("failed to append tile destruction to pending tile selection queue: %w", err)
		}
		log.Debug("Added tile destruction selection to queue")

	case shared.ResourceTileReplacement:
		if a.game == nil {
			return fmt.Errorf("cannot apply tile replacement: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply tile replacement: no player context")
		}
		tileTypes := make([]string, output.Amount)
		for i := 0; i < output.Amount; i++ {
			tileTypes[i] = "tile-replacement:" + output.TileType
		}
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, a.sourceCardID, nil); err != nil {
			return fmt.Errorf("failed to append tile replacement to pending tile selection queue: %w", err)
		}
		log.Debug("Added tile replacement selection to queue",
			zap.String("replacement_tile", output.TileType))

	case shared.ResourceActionReuse:
		log.Debug("Skipping action-reuse output (handled at action layer)")

	case shared.ResourceColonyTile:
		if a.game == nil {
			return fmt.Errorf("cannot apply colony tile: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply colony tile: no player context")
		}
		if !a.game.HasColonies() {
			log.Warn("Colony tile output ignored: colonies expansion not enabled")
			return nil
		}
		colonyIDs := a.game.GetAvailableColonyIDs()
		if len(colonyIDs) == 0 {
			log.Warn("No colony tiles available for placement")
			return nil
		}
		a.player.Selection().SetPendingColonySelection(&shared.PendingColonySelection{
			AvailableColonyIDs:         colonyIDs,
			AllowDuplicatePlayerColony: output.AllowDuplicatePlayerColony,
			Source:                     a.source,
			SourceCardID:               a.sourceCardID,
		})
		log.Debug("Set pending colony selection",
			zap.Int("available_colonies", len(colonyIDs)),
			zap.Bool("allow_duplicate", output.AllowDuplicatePlayerColony))

	default:
		log.Warn("Unhandled output type",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
}
