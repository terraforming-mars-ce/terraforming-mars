package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game"
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
	targetCardID      string                // Card ID for any-card targeting (optional, set by caller)
	targetPlayerID    string                // Player ID for any-player targeting (optional, set by caller)
	stealSourceCardID string                // Card ID to steal resources from for steal-from-any-card outputs (optional)
	sourceBehaviorIdx int                   // Behavior index for card draw selection tracking
	cardRegistry      CardRegistryInterface // Card registry for tag counting in per conditions (optional)
	logger            *zap.Logger
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

// WithTargetCardID sets the target card ID for any-card resource placement
func (a *BehaviorApplier) WithTargetCardID(cardID string) *BehaviorApplier {
	a.targetCardID = cardID
	return a
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

	log.Debug("💰 Processing behavior inputs")

	resources := a.player.Resources().Get()

	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredit:
			if resources.Credits < input.Amount {
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, resources.Credits)
			}
		case shared.ResourceSteel:
			if resources.Steel < input.Amount {
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, resources.Steel)
			}
		case shared.ResourceTitanium:
			if resources.Titanium < input.Amount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, resources.Titanium)
			}
		case shared.ResourcePlant:
			if resources.Plants < input.Amount {
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, resources.Plants)
			}
		case shared.ResourceEnergy:
			if resources.Energy < input.Amount {
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, resources.Energy)
			}
		case shared.ResourceHeat:
			if resources.Heat < input.Amount {
				return fmt.Errorf("insufficient heat: need %d, have %d", input.Amount, resources.Heat)
			}
		default:
			log.Warn("⚠️ Unhandled input type", zap.String("type", string(input.ResourceType)))
		}
	}

	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredit:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceCredit: -input.Amount,
			})
			log.Info("💸 Deducted credits", zap.Int("amount", input.Amount))

		case shared.ResourceSteel:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceSteel: -input.Amount,
			})
			log.Info("🔩 Deducted steel", zap.Int("amount", input.Amount))

		case shared.ResourceTitanium:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceTitanium: -input.Amount,
			})
			log.Info("⚙️ Deducted titanium", zap.Int("amount", input.Amount))

		case shared.ResourcePlant:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourcePlant: -input.Amount,
			})
			log.Info("🌱 Deducted plants", zap.Int("amount", input.Amount))

		case shared.ResourceEnergy:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceEnergy: -input.Amount,
			})
			log.Info("⚡ Deducted energy", zap.Int("amount", input.Amount))

		case shared.ResourceHeat:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceHeat: -input.Amount,
			})
			log.Info("🔥 Deducted heat", zap.Int("amount", input.Amount))
		}
	}

	return nil
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
) ([]game.CalculatedOutput, error) {
	if len(outputs) == 0 {
		return nil, nil
	}

	log := a.logger.With(
		zap.String("source", a.source),
		zap.Int("output_count", len(outputs)),
	)

	log.Debug("✨ Processing behavior outputs")

	var calculatedOutputs []game.CalculatedOutput

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
				log.Debug("📊 Calculated scaled output",
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("base_amount", output.Amount),
					zap.Int("count", count),
					zap.Int("per_amount", output.Per.Amount),
					zap.Int("calculated_amount", actualAmount))
			}
		}

		// Create a modified output with the calculated amount
		modifiedOutput := output
		modifiedOutput.Amount = actualAmount

		if err := a.applyOutput(ctx, modifiedOutput, log); err != nil {
			return calculatedOutputs, err
		}

		// Track the calculated output
		if isScaled || actualAmount != 0 {
			calculatedOutputs = append(calculatedOutputs, game.CalculatedOutput{
				ResourceType: string(output.ResourceType),
				Amount:       actualAmount,
				IsScaled:     isScaled,
			})
		}
	}

	if a.game != nil && a.player != nil && len(outputs) > 0 {
		a.game.AddTriggeredEffect(game.TriggeredEffect{
			CardName: a.source,
			PlayerID: a.player.ID(),
			Outputs:  outputs,
		})
	}

	return calculatedOutputs, nil
}

// countPerCondition counts items matching a Per condition
func (a *BehaviorApplier) countPerCondition(per *shared.PerCondition) int {
	if per == nil {
		return 0
	}

	// Handle resource storage on card (e.g., animals on this card)
	if per.Target != nil && *per.Target == string(TargetSelfCard) {
		if a.sourceCardID != "" {
			return a.player.Resources().GetCardStorage(a.sourceCardID)
		}
		return 0
	}

	// Handle tag counting
	if per.Tag != nil && a.cardRegistry != nil {
		if per.Target != nil && *per.Target == "any-player" && a.game != nil {
			return CountAllPlayersTagsByType(a.game.GetAllPlayers(), a.cardRegistry, *per.Tag)
		}
		return CountPlayerTagsByType(a.player, a.cardRegistry, *per.Tag)
	}

	// Handle tile counting
	if a.game != nil {
		cityTileType := shared.ResourceCityTile
		greeneryTileType := shared.ResourceGreeneryTile

		switch per.ResourceType {
		case shared.ResourceOceanTile:
			return CountAllTilesOfType(a.game.Board(), shared.ResourceOceanTile)
		case shared.ResourceCityTile:
			if per.Target != nil && *per.Target == string(TargetSelfPlayer) {
				return CountPlayerTiles(a.player.ID(), a.game.Board(), &cityTileType)
			}
			return CountAllTilesOfType(a.game.Board(), shared.ResourceCityTile)
		case shared.ResourceGreeneryTile:
			if per.Target != nil && *per.Target == string(TargetSelfPlayer) {
				return CountPlayerTiles(a.player.ID(), a.game.Board(), &greeneryTileType)
			}
			return CountAllTilesOfType(a.game.Board(), shared.ResourceGreeneryTile)
		}
	}

	// Try to count as a tag type if cardRegistry is available
	if a.cardRegistry != nil {
		return CountPlayerTagsByType(a.player, a.cardRegistry, shared.CardTag(per.ResourceType))
	}

	return 0
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
	for _, output := range outputs {
		switch output.ResourceType {
		case shared.ResourceCardPeek:
			peekAmount += output.Amount
		case shared.ResourceCardTake:
			takeAmount += output.Amount
		case shared.ResourceCardBuy:
			buyAmount += output.Amount
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

	// Draw cards from deck
	drawnCards, err := a.game.Deck().DrawProjectCards(ctx, peekAmount)
	if err != nil {
		return false, fmt.Errorf("failed to draw cards: %w", err)
	}

	log.Info("🃏 Drew cards for peek selection",
		zap.Int("peek_amount", peekAmount),
		zap.Int("take_amount", takeAmount),
		zap.Int("buy_amount", buyAmount),
		zap.Strings("drawn_cards", drawnCards))

	// Create pending card draw selection
	selection := &player.PendingCardDrawSelection{
		AvailableCards:      drawnCards,
		FreeTakeCount:       takeAmount,
		MaxBuyCount:         buyAmount,
		CardBuyCost:         3, // Standard card cost
		Source:              a.source,
		SourceCardID:        a.sourceCardID,
		SourceBehaviorIndex: a.sourceBehaviorIdx,
	}

	// Set on player
	a.player.Selection().SetPendingCardDrawSelection(selection)

	log.Info("🃏 Created pending card draw selection",
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
		log.Debug("⏭️ Skipping steal: no target player (solo mode)",
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

	log.Info("🎯 Stole resource from target player",
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
		log.Debug("⏭️ Skipping any-player resource removal: no target player (solo mode)",
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

	removeAmount := min(amount, current)

	if removeAmount > 0 {
		targetPlayer.Resources().Add(map[shared.ResourceType]int{
			resourceType: -removeAmount,
		})
	}

	log.Info("🎯 Removed resource from target player",
		zap.String("target_player_id", a.targetPlayerID),
		zap.String("resource_type", string(resourceType)),
		zap.Int("requested", amount),
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
		log.Debug("⏭️ Skipping any-player production change: no target player (solo mode)",
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

	log.Info("🎯 Applied production change to target player",
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
		log.Info("💰 Added credits", zap.Int("amount", output.Amount))

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
		log.Info("🔩 Added steel", zap.Int("amount", output.Amount))

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
		log.Info("⚙️ Added titanium", zap.Int("amount", output.Amount))

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
		log.Info("🌱 Added plants", zap.Int("amount", output.Amount))

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
		log.Info("⚡ Added energy", zap.Int("amount", output.Amount))

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
		log.Info("🔥 Added heat", zap.Int("amount", output.Amount))

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
		log.Info("💰 Added credits production", zap.Int("amount", output.Amount))

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
		log.Info("🔩 Added steel production", zap.Int("amount", output.Amount))

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
		log.Info("⚙️ Added titanium production", zap.Int("amount", output.Amount))

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
		log.Info("🌱 Added plants production", zap.Int("amount", output.Amount))

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
		log.Info("⚡ Added energy production", zap.Int("amount", output.Amount))

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
		log.Info("🔥 Added heat production", zap.Int("amount", output.Amount))

	case shared.ResourceTR:
		if a.player == nil {
			return fmt.Errorf("cannot apply terraform rating: no player context")
		}
		a.player.Resources().UpdateTerraformRating(output.Amount)
		log.Info("🌍 Added terraform rating", zap.Int("amount", output.Amount))

	case shared.ResourceOxygen:
		if a.game == nil {
			return fmt.Errorf("cannot apply oxygen: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseOxygen(ctx, output.Amount)
		if err != nil {
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}
		log.Info("🌊 Increased oxygen", zap.Int("steps", actualSteps))

	case shared.ResourceTemperature:
		if a.game == nil {
			return fmt.Errorf("cannot apply temperature: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseTemperature(ctx, output.Amount)
		if err != nil {
			return fmt.Errorf("failed to increase temperature: %w", err)
		}
		log.Info("🌡️ Increased temperature", zap.Int("steps", actualSteps))

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
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, nil); err != nil {
			return fmt.Errorf("failed to append land claim to pending tile selection queue: %w", err)
		}

		log.Info("🏴 Added land claim tile selection to queue",
			zap.Int("count", output.Amount))

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
				BoardTags:  output.TileRestrictions.BoardTags,
				Adjacency:  output.TileRestrictions.Adjacency,
				OnTileType: output.TileRestrictions.OnTileType,
			}
		}

		// Atomically append to queue (thread-safe)
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source, tileRestrictions); err != nil {
			return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
		}

		log.Info("🏗️ Added tile placements to queue",
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
			log.Info("💰 Added payment substitute",
				zap.String("resource_type", string(resourceType)),
				zap.Int("conversion_rate", output.Amount))
		} else {
			log.Warn("⚠️ payment-substitute output missing selectors with resources")
		}

	case shared.ResourceDiscount:
		// Discounts are registered as effects and handled by RequirementModifierCalculator
		// The effect registration happens when the card is played (auto effects are added to player.Effects)
		// The calculator then computes the actual modifiers from all effects
		log.Info("🏷️ Discount effect registered",
			zap.Int("amount", output.Amount),
			zap.Any("selectors", output.Selectors))

	case shared.ResourceValueModifier:
		if a.player == nil {
			return fmt.Errorf("cannot apply value modifier: no player context")
		}
		// Apply value modifier to each affected resource (e.g., titanium +1, steel +1)
		for _, resourceStr := range GetResourcesFromSelectors(output.Selectors) {
			resourceType := shared.ResourceType(resourceStr)
			a.player.Resources().AddValueModifier(resourceType, output.Amount)
			log.Info("💎 Added resource value modifier",
				zap.String("resource_type", string(resourceType)),
				zap.Int("modifier_amount", output.Amount))
		}

	case shared.ResourceAnimal, shared.ResourceMicrobe, shared.ResourceFloater:
		if a.player == nil {
			return fmt.Errorf("cannot apply card resource: no player context")
		}

		target := output.Target
		switch target {
		case "self-card":
			if a.sourceCardID == "" {
				log.Warn("⚠️ Cannot place resource on self-card: no source card ID",
					zap.String("resource_type", string(output.ResourceType)))
				return nil
			}
			a.player.Resources().AddToStorage(a.sourceCardID, output.Amount)
			log.Info("🐾 Added resource to card storage",
				zap.String("card_id", a.sourceCardID),
				zap.String("resource_type", string(output.ResourceType)),
				zap.Int("amount", output.Amount))

		case "steal-from-any-card":
			if a.stealSourceCardID == "" {
				log.Debug("⏭️ Skipping steal-from-any-card: no source card specified")
				return nil
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
					log.Info("🎯 Stole resource from card",
						zap.String("source_card_id", a.stealSourceCardID),
						zap.String("owner_player_id", p.ID()),
						zap.String("resource_type", string(output.ResourceType)),
						zap.Int("amount", stolenAmount))
					break
				}
			}
			if stolenAmount > 0 && a.sourceCardID != "" {
				a.player.Resources().AddToStorage(a.sourceCardID, stolenAmount)
				log.Info("🐾 Added stolen resource to self card",
					zap.String("card_id", a.sourceCardID),
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("amount", stolenAmount))
			}

		case "any-card":
			// Add resources to the specified target card
			if a.targetCardID == "" {
				log.Warn("⚠️ No target card specified for any-card resource placement",
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("amount", output.Amount))
				return fmt.Errorf("no target card specified for any-card resource placement")
			}
			a.player.Resources().AddToStorage(a.targetCardID, output.Amount)
			log.Info("🐾 Added resource to target card storage",
				zap.String("card_id", a.targetCardID),
				zap.String("resource_type", string(output.ResourceType)),
				zap.Int("amount", output.Amount))

		default:
			// Default to self-card if target is empty and sourceCardID is set
			if a.sourceCardID != "" {
				a.player.Resources().AddToStorage(a.sourceCardID, output.Amount)
				log.Info("🐾 Added resource to card storage (default to self)",
					zap.String("card_id", a.sourceCardID),
					zap.String("resource_type", string(output.ResourceType)),
					zap.Int("amount", output.Amount))
			} else {
				log.Warn("⚠️ Unhandled target for card resource",
					zap.String("target", target),
					zap.String("resource_type", string(output.ResourceType)))
			}
		}

	case shared.ResourceCardPeek, shared.ResourceCardTake, shared.ResourceCardBuy:
		// Handled by ApplyCardDrawOutputs - skip here
		log.Debug("🃏 Skipping card draw output (handled by ApplyCardDrawOutputs)",
			zap.String("type", string(output.ResourceType)),
			zap.Int("amount", output.Amount))

	default:
		log.Warn("⚠️ Unhandled output type",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
}
