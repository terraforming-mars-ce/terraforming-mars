package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ColonyBonusLookup provides colony definition lookup for colony-bonus output handling.
type ColonyBonusLookup interface {
	GetByID(colonyID string) (*colony.ColonyDefinition, error)
}

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
	colonyBonusLookup ColonyBonusLookup
	deferredSteal     shared.BehaviorCondition
	logger            *zap.Logger
}

// DeferredSteal returns the deferred steal output, if any (for post-tile-placement processing)
func (a *BehaviorApplier) DeferredSteal() shared.BehaviorCondition {
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

// WithColonyBonusLookup sets the colony definition lookup for colony-bonus outputs
func (a *BehaviorApplier) WithColonyBonusLookup(lookup ColonyBonusLookup) *BehaviorApplier {
	a.colonyBonusLookup = lookup
	return a
}

// ApplyInputs validates player has required resources and deducts them
// Returns error if player is nil or insufficient resources
func (a *BehaviorApplier) ApplyInputs(
	ctx context.Context,
	inputs []shared.BehaviorCondition,
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

	// Pass 1: Validate all inputs before deducting anything
	for _, input := range inputs {
		rt := input.GetResourceType()
		effectiveAmount := input.GetAmount()
		if shared.IsVariableAmount(input) {
			effectiveAmount = input.GetAmount() * a.selectedAmount
		}

		// Storage resource inputs (target: "self-card") deduct from card storage
		if input.GetTarget() == "self-card" && IsStorageResourceType(rt) {
			if a.sourceCardID == "" {
				return fmt.Errorf("cannot deduct from self-card: no source card ID")
			}
			storage := a.player.Resources().GetCardStorage(a.sourceCardID)
			if storage < effectiveAmount {
				return fmt.Errorf("insufficient %s on card: need %d, have %d", rt, effectiveAmount, storage)
			}
			continue
		}

		// Credit inputs with paymentAllowed use CardPayment-style validation
		if paymentAllowed := shared.GetPaymentAllowed(input); rt == shared.ResourceCredit && len(paymentAllowed) > 0 {
			if err := a.validateActionPayment(effectiveAmount, paymentAllowed, log); err != nil {
				return err
			}
			continue
		}

		if err := a.validateInputAmount(rt, effectiveAmount, resources); err != nil {
			return err
		}
	}

	// Pass 2: Deduct resources
	for _, input := range inputs {
		rt := input.GetResourceType()
		effectiveAmount := input.GetAmount()
		if shared.IsVariableAmount(input) {
			effectiveAmount = input.GetAmount() * a.selectedAmount
		}

		if effectiveAmount == 0 {
			continue
		}

		// Storage resource inputs (target: "self-card") deduct from card storage
		if input.GetTarget() == "self-card" && IsStorageResourceType(rt) {
			a.player.Resources().AddToStorage(a.sourceCardID, -effectiveAmount)
			log.Debug("Deducted from card storage",
				zap.String("card_id", a.sourceCardID),
				zap.String("resource_type", string(rt)),
				zap.Int("amount", effectiveAmount))
			continue
		}

		// Credit inputs with paymentAllowed use CardPayment-style deduction
		if paymentAllowed := shared.GetPaymentAllowed(input); rt == shared.ResourceCredit && len(paymentAllowed) > 0 {
			a.applyActionPayment(effectiveAmount, log)
			continue
		}

		if shared.IsProductionResourceType(rt) {
			a.player.Resources().AddProduction(map[shared.ResourceType]int{rt: -effectiveAmount})
			log.Debug("Deducted production", zap.String("type", string(rt)), zap.Int("amount", effectiveAmount))
		} else {
			a.player.Resources().Add(map[shared.ResourceType]int{rt: -effectiveAmount})
			log.Debug("Deducted resource", zap.String("type", string(rt)), zap.Int("amount", effectiveAmount))
		}
	}

	return nil
}

// validateInputAmount checks the player has enough of the given resource type.
func (a *BehaviorApplier) validateInputAmount(rt shared.ResourceType, amount int, resources shared.Resources) error {
	switch rt {
	case shared.ResourceCredit:
		if resources.Credits < amount {
			return fmt.Errorf("insufficient credits: need %d, have %d", amount, resources.Credits)
		}
	case shared.ResourceSteel:
		if resources.Steel < amount {
			return fmt.Errorf("insufficient steel: need %d, have %d", amount, resources.Steel)
		}
	case shared.ResourceTitanium:
		if resources.Titanium < amount {
			return fmt.Errorf("insufficient titanium: need %d, have %d", amount, resources.Titanium)
		}
	case shared.ResourcePlant:
		if resources.Plants < amount {
			return fmt.Errorf("insufficient plants: need %d, have %d", amount, resources.Plants)
		}
	case shared.ResourceEnergy:
		if resources.Energy < amount {
			return fmt.Errorf("insufficient energy: need %d, have %d", amount, resources.Energy)
		}
	case shared.ResourceHeat:
		if resources.Heat < amount {
			return fmt.Errorf("insufficient heat: need %d, have %d", amount, resources.Heat)
		}
	default:
		if shared.IsProductionResourceType(rt) {
			production := a.player.Resources().Production()
			available := production.GetAmount(rt)
			if available < amount {
				return fmt.Errorf("insufficient %s: need %d, have %d", rt, amount, available)
			}
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
func IsStorageResourceType(rt shared.ResourceType) bool {
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
		shared.ResourceGlobalParameterLenience, shared.ResourceIgnoreGlobalRequirements,
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
	outputs []shared.BehaviorCondition,
) error {
	_, err := a.ApplyOutputsAndGetCalculated(ctx, outputs)
	return err
}

// ApplyOutputsAndGetCalculated applies outputs and returns the calculated values
// This is useful for logging scaled outputs (e.g., "+1 MC per 2 plant tags" becomes "+3 MC")
func (a *BehaviorApplier) ApplyOutputsAndGetCalculated(
	ctx context.Context,
	outputs []shared.BehaviorCondition,
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
		rt := output.GetResourceType()
		baseAmount := output.GetAmount()

		// Calculate the actual amount if this output has a Per condition
		actualAmount := baseAmount
		isScaled := false

		if per := shared.GetPerCondition(output); per != nil && a.player != nil && a.game != nil {
			count := a.countPerCondition(per)
			if per.Amount > 0 {
				multiplier := count / per.Amount
				actualAmount = baseAmount * multiplier
				isScaled = true
				log.Debug("Calculated scaled output",
					zap.String("resource_type", string(rt)),
					zap.Int("base_amount", baseAmount),
					zap.Int("count", count),
					zap.Int("per_amount", per.Amount),
					zap.Int("calculated_amount", actualAmount))
			}
		}

		// Apply variable amount multiplier (player-selected amount)
		if shared.IsVariableAmount(output) {
			actualAmount = baseAmount * a.selectedAmount
			isScaled = true
			log.Debug("Applied variable amount",
				zap.String("resource_type", string(rt)),
				zap.Int("base_amount", baseAmount),
				zap.Int("selected_amount", a.selectedAmount),
				zap.Int("calculated_amount", actualAmount))
		}

		if err := a.applyOutput(ctx, output, actualAmount, log); err != nil {
			return calculatedOutputs, err
		}

		// Colony-bonus outputs expand into the actual resources gained
		if rt == shared.ResourceColonyBonus {
			bonusOutputs := a.collectColonyBonusOutputs(log)
			calculatedOutputs = append(calculatedOutputs, bonusOutputs...)
			notificationOutputs = append(notificationOutputs, bonusOutputs...)
			continue
		}

		// Track for state diff log (existing behavior)
		if isScaled || actualAmount != 0 {
			calculatedOutputs = append(calculatedOutputs, shared.CalculatedOutput{
				ResourceType: string(rt),
				Amount:       actualAmount,
				IsScaled:     isScaled,
			})
		}

		// Track non-zero resource outputs for triggered effect notifications
		// Skip effect-type outputs (discount, payment-substitute, etc.) since they get
		// their own "Effect:" notification via SourceTypeEffectAdded
		if actualAmount != 0 && !isEffectOutputType(rt) {
			resourceType := string(rt)
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
		effect := shared.TriggeredEffect{
			CardName:          a.source,
			PlayerID:          a.player.ID(),
			SourceType:        a.sourceType,
			Outputs:           outputs,
			CalculatedOutputs: notificationOutputs,
		}
		if a.sourceType == shared.SourceTypePassiveEffect {
			a.game.AddOrMergeTriggeredEffect(effect)
		} else {
			a.game.AddTriggeredEffect(effect)
		}
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
	if per != nil && per.ResourceType == shared.ResourceColonyCount && a.game != nil {
		return a.game.Colonies().CountAllColonies()
	}
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
	outputs []shared.BehaviorCondition,
) (bool, error) {
	log := a.logger.With(
		zap.String("source", a.source),
		zap.String("method", "ApplyCardDrawOutputs"),
	)

	// Scan outputs for card-peek, card-take, card-buy
	var peekAmount, takeAmount, buyAmount int
	var isPrelude bool
	for _, output := range outputs {
		switch output.GetResourceType() {
		case shared.ResourceCardPeek:
			peekAmount += output.GetAmount()
		case shared.ResourceCardTake:
			takeAmount += output.GetAmount()
		case shared.ResourceCardBuy:
			buyAmount += output.GetAmount()
		}
		if co, ok := output.(*shared.CardOperationCondition); ok && hasPreludeCardType(co.Selectors) {
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

	// Calculate card buy cost (accounts for discounts like Polyphemos)
	cardBuyCost := 3
	if a.player != nil && a.cardRegistry != nil {
		calc := NewRequirementModifierCalculator(a.cardRegistry)
		discounts := calc.CalculateActionDiscounts(a.player, shared.ActionCardBuying)
		cardBuyCost = max(3-discounts[shared.ResourceCredit], 0)
	}

	// Create pending card draw selection
	selection := &shared.PendingCardDrawSelection{
		AvailableCards:      drawnCards,
		FreeTakeCount:       takeAmount,
		MaxBuyCount:         buyAmount,
		CardBuyCost:         cardBuyCost,
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

// applyOutput dispatches a single output to the appropriate category handler.
// The amount parameter is the pre-calculated value (accounting for Per and VariableAmount).
func (a *BehaviorApplier) applyOutput(
	ctx context.Context,
	output shared.BehaviorCondition,
	amount int,
	log *zap.Logger,
) error {
	switch o := output.(type) {
	case *shared.BasicResourceCondition:
		return a.applyBasicResourceOutput(ctx, o, amount, log)
	case *shared.ProductionCondition:
		return a.applyProductionOutput(ctx, o, amount, log)
	case *shared.GlobalParameterCondition:
		return a.applyGlobalParameterOutput(ctx, o, amount, log)
	case *shared.TilePlacementCondition:
		return a.applyTilePlacementOutput(ctx, o, amount, log)
	case *shared.EffectCondition:
		return a.applyEffectOutput(ctx, o, amount, log)
	case *shared.CardStorageCondition:
		return a.applyCardStorageOutput(ctx, o, amount, log)
	case *shared.CardOperationCondition:
		return a.applyCardOperationOutput(ctx, o, amount, log)
	case *shared.ColonyCondition:
		return a.applyColonyOutput(ctx, o, amount, log)
	case *shared.TileModificationCondition:
		return a.applyTileModificationOutput(ctx, o, amount, log)
	case *shared.MiscCondition:
		return a.applyMiscOutput(ctx, o, amount, log)
	default:
		return fmt.Errorf("unknown output condition type: %T", output)
	}
}

// applyColonyBonuses applies all colony bonuses for the player.
// Card-targeted resources (microbe, animal, floater) are queued for player selection.
func (a *BehaviorApplier) applyColonyBonuses(_ context.Context, log *zap.Logger) {
	bonuses := CollectColonyBonuses(a.player.ID(), a.game.Colonies().States(), a.colonyBonusLookup)

	pendingByType := map[string]int{}
	var pendingOrder []string

	for _, b := range bonuses {
		rt := shared.ResourceType(b.ResourceType)
		if IsStorageResourceType(rt) {
			if _, exists := pendingByType[b.ResourceType]; !exists {
				pendingOrder = append(pendingOrder, b.ResourceType)
			}
			pendingByType[b.ResourceType] += b.Amount
		} else {
			a.player.Resources().Add(map[shared.ResourceType]int{rt: b.Amount})
		}
		log.Debug("Applied colony bonus",
			zap.String("type", b.ResourceType),
			zap.Int("amount", b.Amount))
	}

	for _, rt := range pendingOrder {
		amount := pendingByType[rt]
		if !HasEligibleStorageCard(a.player, shared.ResourceType(rt), a.cardRegistry) {
			log.Debug("No eligible storage card for colony bonus, resources lost",
				zap.String("resource_type", rt),
				zap.Int("amount", amount))
			continue
		}
		a.player.Selection().AppendPendingColonyResource(shared.PendingColonyResourceSelection{
			ResourceType: rt,
			Amount:       amount,
			Source:       a.source,
			Reason:       "colony-bonus",
		})
	}
}

func (a *BehaviorApplier) collectColonyBonusOutputs(_ *zap.Logger) []shared.CalculatedOutput {
	if a.game == nil || a.player == nil || a.colonyBonusLookup == nil {
		return nil
	}
	return ColonyBonusesToCalculatedOutputs(
		CollectColonyBonuses(a.player.ID(), a.game.Colonies().States(), a.colonyBonusLookup),
	)
}

// ColonyBonusEntry represents a single colony bonus resource gain.
type ColonyBonusEntry struct {
	ResourceType string
	Amount       int
}

// CollectColonyBonuses iterates colony tile states and returns all bonuses for the given player.
func CollectColonyBonuses(playerID string, tileStates []*colony.ColonyState, lookup ColonyBonusLookup) []ColonyBonusEntry {
	if lookup == nil {
		return nil
	}
	var result []ColonyBonusEntry
	for _, ts := range tileStates {
		colonyCount := 0
		for _, ownerID := range ts.PlayerColonies {
			if ownerID == playerID {
				colonyCount++
			}
		}
		if colonyCount == 0 {
			continue
		}
		def, err := lookup.GetByID(ts.DefinitionID)
		if err != nil {
			continue
		}
		for i := 0; i < colonyCount; i++ {
			for _, bonus := range def.ColonyBonus {
				if bonus.Amount > 0 {
					result = append(result, ColonyBonusEntry{ResourceType: bonus.Type, Amount: bonus.Amount})
				}
			}
		}
	}
	return result
}

// ColonyBonusesToCalculatedOutputs aggregates colony bonus entries into calculated outputs by type.
func ColonyBonusesToCalculatedOutputs(bonuses []ColonyBonusEntry) []shared.CalculatedOutput {
	if len(bonuses) == 0 {
		return []shared.CalculatedOutput{}
	}
	totals := map[string]int{}
	var order []string
	for _, b := range bonuses {
		if _, exists := totals[b.ResourceType]; !exists {
			order = append(order, b.ResourceType)
		}
		totals[b.ResourceType] += b.Amount
	}
	outputs := make([]shared.CalculatedOutput, 0, len(totals))
	for _, rt := range order {
		outputs = append(outputs, shared.CalculatedOutput{ResourceType: rt, Amount: totals[rt]})
	}
	return outputs
}

// HasEligibleStorageCard checks if a player has any played card or corporation
// that can store the given resource type.
func HasEligibleStorageCard(p *player.Player, resourceType shared.ResourceType, cardRegistry CardRegistryInterface) bool {
	if cardRegistry == nil {
		return false
	}
	for _, cardID := range p.PlayedCards().Cards() {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}
		if card.ResourceStorage != nil && card.ResourceStorage.Type == resourceType {
			return true
		}
	}
	if corpID := p.CorporationID(); corpID != "" {
		corp, err := cardRegistry.GetByID(corpID)
		if err == nil && corp.ResourceStorage != nil && corp.ResourceStorage.Type == resourceType {
			return true
		}
	}
	return false
}
