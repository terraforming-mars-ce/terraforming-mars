package play_card_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestPlayCardAction_DiscountEffectRegistered(t *testing.T) {
	// Setup: Create game with player who has Space Station in hand
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Set game to active status and action phase for playing cards
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Give player enough credits and add Space Station to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-space-station")

	// Also add a space-tagged card to hand for modifier calculation
	player.Hand().AddCard("card-space-mirrors")

	// Verify initial state: no effects
	effectsBefore := player.Effects().List()
	testutil.AssertEqual(t, 0, len(effectsBefore), "Should have no effects initially")

	// Play Space Station
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Verify: effect should be registered
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Should have 1 effect after playing Space Station")
	testutil.AssertEqual(t, "card-space-station", effectsAfter[0].CardID, "Effect should be from Space Station")
	testutil.AssertEqual(t, "Space Station", effectsAfter[0].CardName, "Effect card name should be Space Station")

	// Verify: discounts are calculated on-demand via RequirementModifierCalculator
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	spaceMirrorsCard, err := cardRegistry.GetByID("card-space-mirrors")
	testutil.AssertNoError(t, err, "Space Mirrors card should exist")

	discount := calculator.CalculateCardDiscounts(player, spaceMirrorsCard)
	testutil.AssertEqual(t, 2, discount, "Space Mirrors should have 2 credit discount from Space Station effect")
}

func TestPlayCardAction_ChoiceCardPlantProduction(t *testing.T) {
	// Setup: Create game with player who has Artificial Photosynthesis in hand
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Set game to active status and action phase for playing cards
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Give player enough credits and add Artificial Photosynthesis to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-artificial-photosynthesis")

	// Verify initial production state
	productionBefore := player.Resources().Production()
	testutil.AssertEqual(t, 0, productionBefore.Plants, "Should have 0 plant production initially")
	testutil.AssertEqual(t, 0, productionBefore.Energy, "Should have 0 energy production initially")

	// Play Artificial Photosynthesis with choice index 0 (plant production +1)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-artificial-photosynthesis", payment, &choiceIndex, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Artificial Photosynthesis with choice 0")

	// Verify: plant production increased by 1, energy unchanged
	productionAfter := player.Resources().Production()
	testutil.AssertEqual(t, 1, productionAfter.Plants, "Should have 1 plant production after choice 0")
	testutil.AssertEqual(t, 0, productionAfter.Energy, "Should have 0 energy production after choice 0")

	// Verify card was played
	testutil.AssertFalse(t, player.Hand().HasCard("card-artificial-photosynthesis"), "Card should not be in hand")
}

func TestPlayCardAction_ChoiceCardEnergyProduction(t *testing.T) {
	// Setup: Create game with player who has Artificial Photosynthesis in hand
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Set game to active status and action phase for playing cards
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Give player enough credits and add Artificial Photosynthesis to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-artificial-photosynthesis")

	// Verify initial production state
	productionBefore := player.Resources().Production()
	testutil.AssertEqual(t, 0, productionBefore.Plants, "Should have 0 plant production initially")
	testutil.AssertEqual(t, 0, productionBefore.Energy, "Should have 0 energy production initially")

	// Play Artificial Photosynthesis with choice index 1 (energy production +2)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-artificial-photosynthesis", payment, &choiceIndex, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Artificial Photosynthesis with choice 1")

	// Verify: energy production increased by 2, plants unchanged
	productionAfter := player.Resources().Production()
	testutil.AssertEqual(t, 0, productionAfter.Plants, "Should have 0 plant production after choice 1")
	testutil.AssertEqual(t, 2, productionAfter.Energy, "Should have 2 energy production after choice 1")

	// Verify card was played
	testutil.AssertFalse(t, player.Hand().HasCard("card-artificial-photosynthesis"), "Card should not be in hand")
}

func TestPlayCardAction_DiscountCalculatedOnDemand(t *testing.T) {
	// Setup: Create game with Space Station already played (effect registered)
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Give player credits and add Space Station to hand
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-space-station")

	// Play Space Station to register the discount effect
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Verify: effect registered
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Should have 1 effect")

	// Verify: discounts are calculated on-demand for any space card
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	// Get Space Mirrors card (has space tag)
	spaceMirrorsCard, err := cardRegistry.GetByID("card-space-mirrors")
	testutil.AssertNoError(t, err, "Space Mirrors card should exist")

	// Discount should apply regardless of whether card is in hand
	discount := calculator.CalculateCardDiscounts(player, spaceMirrorsCard)
	testutil.AssertEqual(t, 2, discount, "Space Mirrors should have 2 credit discount from Space Station effect")

	// Non-space card should not get discount
	nonSpaceCard, err := cardRegistry.GetByID("card-arctic-algae")
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")

	nonSpaceDiscount := calculator.CalculateCardDiscounts(player, nonSpaceCard)
	testutil.AssertEqual(t, 0, nonSpaceDiscount, "Non-space card should have no discount from Space Station")
}

// TestPlayCardAction_WithSingleDiscount tests that payment validation uses discounted cost
// Bug fix: Previously, payment was validated against base cost even with discounts active
func TestPlayCardAction_WithSingleDiscount(t *testing.T) {
	// Setup: Create game with player who has Teractor's Earth discount
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-teractor")

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Register Teractor's discount effect (3 credit discount on Earth cards)
	teractorCard, err := cardRegistry.GetByID("corp-teractor")
	testutil.AssertNoError(t, err, "Teractor card should exist")

	for behaviorIndex, behavior := range teractorCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(player.CardEffect{
				CardID:        teractorCard.ID,
				CardName:      teractorCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Give player exactly the discounted cost (Sponsors costs 6, discount is 3 = 3 credits needed)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 3,
	})

	// Add Earth-tagged card (Sponsors, cost 6) to hand
	p.Hand().AddCard("card-sponsors")

	// Verify discount calculation
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	sponsorsCard, err := cardRegistry.GetByID("card-sponsors")
	testutil.AssertNoError(t, err, "Sponsors card should exist")

	discount := calculator.CalculateCardDiscounts(p, sponsorsCard)
	testutil.AssertEqual(t, 3, discount, "Sponsors should have 3 credit discount from Teractor")

	// Play Sponsors with only 3 credits (effective cost after discount)
	// This should SUCCEED because the fix applies discounts during payment validation
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-sponsors", payment, nil, nil, nil, nil)

	testutil.AssertNoError(t, err, "Should be able to play Sponsors for 3 credits with Teractor discount")

	// Verify card was played
	testutil.AssertFalse(t, p.Hand().HasCard("card-sponsors"), "Sponsors should no longer be in hand")
	testutil.AssertEqual(t, 0, p.Resources().Get().Credits, "All credits should be spent")
}

// TestPlayCardAction_WithDoubleDiscount tests that multiple discounts stack correctly
func TestPlayCardAction_WithDoubleDiscount(t *testing.T) {
	// Setup: Create game with player
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-teractor")

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Register Teractor's discount effect (3 credit discount on Earth cards)
	teractorCard, err := cardRegistry.GetByID("corp-teractor")
	testutil.AssertNoError(t, err, "Teractor card should exist")

	for behaviorIndex, behavior := range teractorCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(player.CardEffect{
				CardID:        teractorCard.ID,
				CardName:      teractorCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Give player credits for Earth Office (1 credit with 3 discount = free, but minimum 0)
	// and Earth Catapult (23 credits with 6 discount = 17 credits)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100, // Enough for everything
	})

	// Add Earth Office to hand
	p.Hand().AddCard("card-earth-office")

	// Play Earth Office (cost 1, but free with 3 credit discount from Teractor)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-earth-office", payment, nil, nil, nil, nil)

	testutil.AssertNoError(t, err, "Should be able to play Earth Office for free with Teractor discount")
	testutil.AssertFalse(t, p.Hand().HasCard("card-earth-office"), "Earth Office should no longer be in hand")

	// Verify Earth Office's discount effect was registered
	effects := p.Effects().List()
	testutil.AssertEqual(t, 2, len(effects), "Should have 2 discount effects (Teractor + Earth Office)")

	// Now verify double discount is calculated correctly
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	earthCatapultCard, err := cardRegistry.GetByID("card-earth-catapult")
	testutil.AssertNoError(t, err, "Earth Catapult card should exist")

	// Earth Catapult (23 cost, Earth tag) should get 6 credit discount (3 from Teractor + 3 from Earth Office)
	totalDiscount := calculator.CalculateCardDiscounts(p, earthCatapultCard)
	testutil.AssertEqual(t, 6, totalDiscount, "Earth Catapult should have 6 credit discount (3+3)")

	// Add Earth Catapult to hand and play it
	p.Hand().AddCard("card-earth-catapult")

	// Reset credits to exactly the discounted cost (23 - 6 = 17)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -p.Resources().Get().Credits + 17,
	})
	testutil.AssertEqual(t, 17, p.Resources().Get().Credits, "Should have exactly 17 credits")

	// Play Earth Catapult with 17 credits (effective cost after double discount)
	payment = cardAction.PaymentRequest{Credits: 17}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-earth-catapult", payment, nil, nil, nil, nil)

	testutil.AssertNoError(t, err, "Should be able to play Earth Catapult for 17 credits with double discount")
	testutil.AssertFalse(t, p.Hand().HasCard("card-earth-catapult"), "Earth Catapult should no longer be in hand")
	testutil.AssertEqual(t, 0, p.Resources().Get().Credits, "All credits should be spent")
}

// TestPlayCardAction_DiscountDoesNotApplyToNonMatchingCard tests that discounts only apply to matching tags
func TestPlayCardAction_DiscountDoesNotApplyToNonMatchingCard(t *testing.T) {
	// Setup: Create game with player who has Teractor's Earth discount
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-teractor")

	// Set game to active status and action phase
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Register Teractor's discount effect (3 credit discount on Earth cards)
	teractorCard, err := cardRegistry.GetByID("corp-teractor")
	testutil.AssertNoError(t, err, "Teractor card should exist")

	for behaviorIndex, behavior := range teractorCard.Behaviors {
		if gamecards.HasAutoTrigger(behavior) && gamecards.HasPersistentEffects(behavior) {
			p.Effects().AddEffect(player.CardEffect{
				CardID:        teractorCard.ID,
				CardName:      teractorCard.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			})
		}
	}

	// Give player only 9 credits (not enough for full cost of Arctic Algae which costs 12)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 9,
	})

	// Add Arctic Algae (Plant tag, NOT Earth) to hand
	p.Hand().AddCard("card-arctic-algae")

	// Verify NO discount applies to non-Earth card
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	arcticAlgaeCard, err := cardRegistry.GetByID("card-arctic-algae")
	testutil.AssertNoError(t, err, "Arctic Algae card should exist")

	discount := calculator.CalculateCardDiscounts(p, arcticAlgaeCard)
	testutil.AssertEqual(t, 0, discount, "Arctic Algae (Plant tag) should have no discount from Teractor (Earth discount)")

	// Try to play Arctic Algae with insufficient credits (9 credits, but card costs 12)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-arctic-algae", payment, nil, nil, nil, nil)

	// Should FAIL because discount doesn't apply and player doesn't have enough credits
	testutil.AssertError(t, err, "Should NOT be able to play Arctic Algae with only 9 credits (no discount applies)")
	testutil.AssertTrue(t, p.Hand().HasCard("card-arctic-algae"), "Arctic Algae should still be in hand")
}
