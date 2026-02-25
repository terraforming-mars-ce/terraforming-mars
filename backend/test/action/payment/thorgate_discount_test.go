package payment_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestThorGate_CardDiscount tests that ThorGate's discount applies to cards with power tag
func TestThorGate_CardDiscount(t *testing.T) {
	// Setup: Create game with player
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	p := players[0]

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)

	// Manually register ThorGate discount effect on player (simulating corporation selection)
	thorGateCard, err := cardRegistry.GetByID("corp-thorgate")
	testutil.AssertNoError(t, err, "ThorGate card should exist")

	// Register the discount effect
	for behaviorIndex, behavior := range thorGateCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(player.CardEffect{
				CardID:        thorGateCard.ID,
				CardName:      thorGateCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Verify effect is registered
	effects := p.Effects().List()
	testutil.AssertEqual(t, 1, len(effects), "Should have 1 effect registered")

	// Test card discount calculation
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	// Deep Well Heating has power tag - should get 3 credit discount
	deepWellHeatingCard, err := cardRegistry.GetByID("card-deep-well-heating")
	testutil.AssertNoError(t, err, "Deep Well Heating card should exist")

	discount := calculator.CalculateCardDiscounts(p, deepWellHeatingCard)
	testutil.AssertEqual(t, 3, discount, "Deep Well Heating should have 3 credit discount from ThorGate")

	// Space Mirrors has both power and space tags - should get 3 credit discount
	spaceMirrorsCard, err := cardRegistry.GetByID("card-space-mirrors")
	testutil.AssertNoError(t, err, "Space Mirrors card should exist")

	spaceMirrorsDiscount := calculator.CalculateCardDiscounts(p, spaceMirrorsCard)
	testutil.AssertEqual(t, 3, spaceMirrorsDiscount, "Space Mirrors should have 3 credit discount from ThorGate")

	// Arctic Algae has plant tag (no power tag) - should NOT get discount
	arcticAlgaeCard, err := cardRegistry.GetByID("card-arctic-algae")
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")

	noDiscount := calculator.CalculateCardDiscounts(p, arcticAlgaeCard)
	testutil.AssertEqual(t, 0, noDiscount, "Arctic Algae should have no discount from ThorGate")
}

// TestThorGate_StandardProjectDiscount tests that ThorGate's discount applies to power-plant standard project
func TestThorGate_StandardProjectDiscount(t *testing.T) {
	// Setup: Create game with player
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	p := players[0]

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)

	// Manually register ThorGate discount effect on player
	thorGateCard, err := cardRegistry.GetByID("corp-thorgate")
	testutil.AssertNoError(t, err, "ThorGate card should exist")

	// Register the discount effect
	for behaviorIndex, behavior := range thorGateCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(player.CardEffect{
				CardID:        thorGateCard.ID,
				CardName:      thorGateCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Test standard project discount calculation
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	// Power plant standard project should get 3 credit discount
	powerPlantDiscounts := calculator.CalculateStandardProjectDiscounts(p, shared.StandardProjectPowerPlant)
	creditDiscount := powerPlantDiscounts[shared.ResourceCredit]
	testutil.AssertEqual(t, 3, creditDiscount, "Power plant should have 3 credit discount from ThorGate")

	// Asteroid standard project should NOT get discount
	asteroidDiscounts := calculator.CalculateStandardProjectDiscounts(p, shared.StandardProjectAsteroid)
	asteroidCreditDiscount := asteroidDiscounts[shared.ResourceCredit]
	testutil.AssertEqual(t, 0, asteroidCreditDiscount, "Asteroid should have no discount from ThorGate")

	// Aquifer standard project should NOT get discount
	aquiferDiscounts := calculator.CalculateStandardProjectDiscounts(p, shared.StandardProjectAquifer)
	aquiferCreditDiscount := aquiferDiscounts[shared.ResourceCredit]
	testutil.AssertEqual(t, 0, aquiferCreditDiscount, "Aquifer should have no discount from ThorGate")
}

// TestThorGate_CombinedDiscount tests that ThorGate's single output works for both cards and standard projects
func TestThorGate_CombinedDiscount(t *testing.T) {
	// Setup: Create game with player
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	p := players[0]

	// Set game to active status
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)

	// Manually register ThorGate discount effect
	thorGateCard, err := cardRegistry.GetByID("corp-thorgate")
	testutil.AssertNoError(t, err, "ThorGate card should exist")

	// Verify ThorGate card has correct structure
	testutil.AssertEqual(t, 1, len(thorGateCard.Behaviors), "ThorGate should have exactly 1 behavior")

	// Check the behavior has selectors for power tag and power-plant standard project
	behavior := thorGateCard.Behaviors[0]
	testutil.AssertEqual(t, 1, len(behavior.Outputs), "Behavior should have 1 output")

	output := behavior.Outputs[0]
	testutil.AssertEqual(t, shared.ResourceDiscount, output.ResourceType, "Output should be discount type")
	testutil.AssertEqual(t, 3, output.Amount, "Discount amount should be 3")
	testutil.AssertEqual(t, 2, len(output.Selectors), "Should have 2 selectors (one for power tag, one for power-plant SP)")

	// Register the effect
	p.Effects().AddEffect(player.CardEffect{
		CardID:        thorGateCard.ID,
		CardName:      thorGateCard.Name,
		BehaviorIndex: 0,
		Behavior:      behavior,
	})

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	// Test both card and standard project discounts work from same effect
	deepWellHeatingCard, err := cardRegistry.GetByID("card-deep-well-heating")
	testutil.AssertNoError(t, err, "Deep Well Heating card should exist")

	cardDiscount := calculator.CalculateCardDiscounts(p, deepWellHeatingCard)
	testutil.AssertEqual(t, 3, cardDiscount, "Card discount should be 3")

	projectDiscounts := calculator.CalculateStandardProjectDiscounts(p, shared.StandardProjectPowerPlant)
	projectDiscount := projectDiscounts[shared.ResourceCredit]
	testutil.AssertEqual(t, 3, projectDiscount, "Standard project discount should be 3")
}

// TestDiscountORLogic tests that discounts with multiple targeting criteria use OR logic
func TestDiscountORLogic(t *testing.T) {
	// Setup: Create a custom discount effect that targets both tags AND card types
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)

	// Create a custom discount effect that targets space tag OR event card type (OR logic between selectors)
	customBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: "auto"}},
		Outputs: []shared.ResourceCondition{
			{
				ResourceType: shared.ResourceDiscount,
				Amount:       5,
				Selectors: []shared.Selector{
					{Tags: []shared.CardTag{shared.TagSpace}},
					{CardTypes: []string{"event"}},
				},
			},
		},
	}

	p.Effects().AddEffect(player.CardEffect{
		CardID:        "test-custom-discount",
		CardName:      "Test Custom Discount",
		BehaviorIndex: 0,
		Behavior:      customBehavior,
	})

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	// Asteroid card: space tag + event type - should match on EITHER (OR logic)
	asteroidCard, err := cardRegistry.GetByID("card-asteroid")
	testutil.AssertNoError(t, err, "Asteroid card should exist")
	// Note: Asteroid has space tag AND event type, so it matches both criteria
	asteroidDiscount := calculator.CalculateCardDiscounts(p, asteroidCard)
	testutil.AssertEqual(t, 5, asteroidDiscount, "Asteroid should get discount (matches space tag AND event type)")

	// Space Mirrors: space tag but NOT event type - should still get discount (OR logic)
	spaceMirrorsCard, err := cardRegistry.GetByID("card-space-mirrors")
	testutil.AssertNoError(t, err, "Space Mirrors card should exist")
	spaceMirrorsDiscount := calculator.CalculateCardDiscounts(p, spaceMirrorsCard)
	testutil.AssertEqual(t, 5, spaceMirrorsDiscount, "Space Mirrors should get discount (matches space tag)")

	// Arctic Algae: plant tag, not space, not event - should NOT get discount
	arcticAlgaeCard, err := cardRegistry.GetByID("card-arctic-algae")
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")
	arcticAlgaeDiscount := calculator.CalculateCardDiscounts(p, arcticAlgaeCard)
	testutil.AssertEqual(t, 0, arcticAlgaeDiscount, "Arctic Algae should NOT get discount (no match)")
}
