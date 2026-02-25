package payment_test

import (
	"context"
	"fmt"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestValueModifier_StoredOnPlayerResources tests that value modifiers are stored correctly
func TestValueModifier_StoredOnPlayerResources(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Initial state: no value modifiers
	titaniumModifier := player.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 0, titaniumModifier, "Initial titanium value modifier should be 0")

	// Add value modifier (simulating playing Phobolog)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Verify modifier was stored
	titaniumModifier = player.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 1, titaniumModifier, "Titanium value modifier should be 1 after adding")

	// Add another modifier (simulating playing Advanced Alloys)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Verify modifiers stack
	titaniumModifier = player.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 2, titaniumModifier, "Titanium value modifier should stack to 2")

	_ = ctx // context available for future use
}

// TestValueModifier_PaymentSubstitutesIncludeSteelTitanium tests that PaymentSubstitutes always includes steel/titanium
func TestValueModifier_PaymentSubstitutesIncludeSteelTitanium(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Get payment substitutes
	substitutes := player.Resources().PaymentSubstitutes()

	// Find steel and titanium in substitutes
	var steelSub, titaniumSub *shared.PaymentSubstitute
	for i := range substitutes {
		if substitutes[i].ResourceType == shared.ResourceSteel {
			steelSub = &substitutes[i]
		}
		if substitutes[i].ResourceType == shared.ResourceTitanium {
			titaniumSub = &substitutes[i]
		}
	}

	// Verify both are present with base values
	testutil.AssertTrue(t, steelSub != nil, "Steel should be in payment substitutes")
	testutil.AssertTrue(t, titaniumSub != nil, "Titanium should be in payment substitutes")
	testutil.AssertEqual(t, 2, steelSub.ConversionRate, "Steel base value should be 2")
	testutil.AssertEqual(t, 3, titaniumSub.ConversionRate, "Titanium base value should be 3")
}

// TestValueModifier_TitaniumValueInPaymentSubstitutes tests titanium value modifier affects PaymentSubstitutes
func TestValueModifier_TitaniumValueInPaymentSubstitutes(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add titanium value modifier (Phobolog effect)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Get payment substitutes
	substitutes := player.Resources().PaymentSubstitutes()

	// Find titanium
	var titaniumValue int
	for _, sub := range substitutes {
		if sub.ResourceType == shared.ResourceTitanium {
			titaniumValue = sub.ConversionRate
			break
		}
	}

	// Verify titanium value is base (3) + modifier (1) = 4
	testutil.AssertEqual(t, 4, titaniumValue, "Titanium value should be 4 with Phobolog modifier")
}

// TestValueModifier_SteelValueInPaymentSubstitutes tests steel value modifier affects PaymentSubstitutes
func TestValueModifier_SteelValueInPaymentSubstitutes(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add steel value modifier (Advanced Alloys effect)
	player.Resources().AddValueModifier(shared.ResourceSteel, 1)

	// Get payment substitutes
	substitutes := player.Resources().PaymentSubstitutes()

	// Find steel
	var steelValue int
	for _, sub := range substitutes {
		if sub.ResourceType == shared.ResourceSteel {
			steelValue = sub.ConversionRate
			break
		}
	}

	// Verify steel value is base (2) + modifier (1) = 3
	testutil.AssertEqual(t, 3, steelValue, "Steel value should be 3 with Advanced Alloys modifier")
}

// TestValueModifier_CombinedModifiersStack tests that multiple value modifiers stack
func TestValueModifier_CombinedModifiersStack(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add both steel and titanium modifiers (Advanced Alloys) then extra titanium (Phobolog)
	player.Resources().AddValueModifier(shared.ResourceSteel, 1)    // Advanced Alloys
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // Advanced Alloys
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // Phobolog

	// Get payment substitutes
	substitutes := player.Resources().PaymentSubstitutes()

	// Find values
	var steelValue, titaniumValue int
	for _, sub := range substitutes {
		if sub.ResourceType == shared.ResourceSteel {
			steelValue = sub.ConversionRate
		}
		if sub.ResourceType == shared.ResourceTitanium {
			titaniumValue = sub.ConversionRate
		}
	}

	// Verify values
	testutil.AssertEqual(t, 3, steelValue, "Steel value should be 3 (2 base + 1 from Advanced Alloys)")
	testutil.AssertEqual(t, 5, titaniumValue, "Titanium value should be 5 (3 base + 1 Advanced Alloys + 1 Phobolog)")
}

// TestValueModifier_PaymentTotalValue tests that CardPayment.TotalValue uses modified values
func TestValueModifier_PaymentTotalValue(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add titanium modifier (Phobolog)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Create payment: 3 titanium
	payment := gamecards.CardPayment{
		Credits:  0,
		Steel:    0,
		Titanium: 3,
	}

	// Calculate total value with player's substitutes
	substitutes := player.Resources().PaymentSubstitutes()
	totalValue := payment.TotalValue(substitutes, nil)

	// 3 titanium * 4 MC each = 12 MC
	testutil.AssertEqual(t, 12, totalValue, "3 titanium at value 4 should equal 12 MC")
}

// TestValueModifier_PaymentWithModifiedSteel tests payment with modified steel value
func TestValueModifier_PaymentWithModifiedSteel(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add steel modifier
	player.Resources().AddValueModifier(shared.ResourceSteel, 1)

	// Create payment: 3 steel
	payment := gamecards.CardPayment{
		Credits: 0,
		Steel:   3,
	}

	// Calculate total value
	substitutes := player.Resources().PaymentSubstitutes()
	totalValue := payment.TotalValue(substitutes, nil)

	// 3 steel * 3 MC each = 9 MC
	testutil.AssertEqual(t, 9, totalValue, "3 steel at value 3 should equal 9 MC")
}

// TestValueModifier_PaymentCoversCardCost tests that modified payment values correctly cover costs
func TestValueModifier_PaymentCoversCardCost(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add titanium modifier (Phobolog)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Create payment: 3 titanium (should cover 12 MC with modifier)
	payment := gamecards.CardPayment{
		Titanium: 3,
	}

	substitutes := player.Resources().PaymentSubstitutes()

	// Test: Can cover 12 MC cost with titanium (3 * 4 = 12)
	err := payment.CoversCardCost(12, false, true, substitutes, nil)
	testutil.AssertNoError(t, err, "Payment of 3 titanium at value 4 should cover 12 MC cost")

	// Test: Cannot cover 13 MC cost with same payment
	err = payment.CoversCardCost(13, false, true, substitutes, nil)
	testutil.AssertError(t, err, "Payment of 3 titanium at value 4 should NOT cover 13 MC cost")
}

// TestValueModifier_PaymentInsufficientWithModifier tests that insufficient payment is rejected
func TestValueModifier_PaymentInsufficientWithModifier(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Even with modifier, 2 steel at value 3 = 6 MC cannot pay 8 cost
	player.Resources().AddValueModifier(shared.ResourceSteel, 1)

	payment := gamecards.CardPayment{
		Steel: 2,
	}

	substitutes := player.Resources().PaymentSubstitutes()

	// 2 steel * 3 = 6 MC, cannot cover 8 MC
	err := payment.CoversCardCost(8, true, false, substitutes, nil)
	testutil.AssertError(t, err, "Payment of 2 steel at value 3 (6 MC) should NOT cover 8 MC cost")
}

// TestValueModifier_SteelRestrictedToBuildingTags tests that steel tag restriction is independent of value modifier
func TestValueModifier_SteelRestrictedToBuildingTags(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add steel modifier - should not change tag restriction
	player.Resources().AddValueModifier(shared.ResourceSteel, 1)

	payment := gamecards.CardPayment{
		Steel: 3,
	}

	substitutes := player.Resources().PaymentSubstitutes()

	// Steel not allowed (no building tag) - should fail regardless of modifier
	err := payment.CoversCardCost(9, false, false, substitutes, nil)
	testutil.AssertError(t, err, "Steel should not be allowed for non-building cards even with value modifier")

	// Steel allowed (building tag) - should pass
	err = payment.CoversCardCost(9, true, false, substitutes, nil)
	testutil.AssertNoError(t, err, "Steel should be allowed for building cards with value modifier")
}

// TestValueModifier_TitaniumRestrictedToSpaceTags tests that titanium tag restriction is independent of value modifier
func TestValueModifier_TitaniumRestrictedToSpaceTags(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Add titanium modifier - should not change tag restriction
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	payment := gamecards.CardPayment{
		Titanium: 3,
	}

	substitutes := player.Resources().PaymentSubstitutes()

	// Titanium not allowed (no space tag) - should fail regardless of modifier
	err := payment.CoversCardCost(12, false, false, substitutes, nil)
	testutil.AssertError(t, err, "Titanium should not be allowed for non-space cards even with value modifier")

	// Titanium allowed (space tag) - should pass
	err = payment.CoversCardCost(12, false, true, substitutes, nil)
	testutil.AssertNoError(t, err, "Titanium should be allowed for space cards with value modifier")
}

// TestValueModifier_PlayCardWithModifiedTitanium tests full play card flow with value modifier
func TestValueModifier_PlayCardWithModifiedTitanium(t *testing.T) {
	// Setup: Create game with player who has value modifier
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Start the game
	testutil.StartTestGame(t, testGame)

	// Give player titanium and add value modifier
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 5,
		shared.ResourceCredit:   50, // Extra credits for buying other cards if needed
	})
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // +1 titanium value

	// Add a space-tagged card to hand that costs 14 (asteroid)
	player.Hand().AddCard("card-asteroid")

	// Play card with 4 titanium (4 * 4 MC = 16 MC covers 14 cost)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{
		Credits:  0,
		Titanium: 4, // 4 titanium at 4 MC each = 16 MC
	}

	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-asteroid", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should be able to play 14-cost card with 4 titanium at value 4")

	// Verify titanium was deducted
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Titanium, "Should have 3 titanium remaining (5 - 4 + 2 from Asteroid behavior)")

	// Verify card is no longer in hand
	testutil.AssertFalse(t, player.Hand().HasCard("card-asteroid"), "Card should be removed from hand")
}

// TestValueModifier_MixedPaymentWithModifier tests mixed payment (credits + modified titanium)
func TestValueModifier_MixedPaymentWithModifier(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Start the game
	testutil.StartTestGame(t, testGame)

	// Give player resources and value modifier
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 2,
		shared.ResourceCredit:   50,
	})
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // +1 titanium value

	// Add a space-tagged card to hand that costs 14
	player.Hand().AddCard("card-asteroid")

	// Play card with 2 titanium (8 MC) + 6 credits = 14 MC exactly
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{
		Credits:  6,
		Titanium: 2, // 2 titanium at 4 MC each = 8 MC
	}

	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-asteroid", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should be able to play 14-cost card with 2 titanium (8) + 6 credits")

	// Verify resources were deducted
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Titanium, "Should have 2 titanium remaining (0 + 2 from Asteroid behavior)")
	testutil.AssertEqual(t, 44, resources.Credits, "Should have 44 credits remaining (50 - 6)")
}

// TestValueModifier_InsufficientPaymentRejected tests that insufficient payment with modifier is rejected
func TestValueModifier_InsufficientPaymentRejected(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	// Start the game
	testutil.StartTestGame(t, testGame)

	// Give player resources and value modifier
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 3,
		shared.ResourceCredit:   0,
	})
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // +1 titanium value

	// Add a space-tagged card to hand that costs 14
	player.Hand().AddCard("card-asteroid")

	// Try to play card with only 3 titanium (12 MC) - not enough for 14 cost
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{
		Titanium: 3, // 3 titanium at 4 MC each = 12 MC (not enough for 14)
	}

	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), "card-asteroid", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should reject payment of 12 MC for 14 cost card")

	// Verify titanium was NOT deducted (action failed)
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Titanium, "Titanium should not be deducted on failed payment")
}

// TestValueModifier_NoModifierUsesBaseValue tests that base values work when no modifier
func TestValueModifier_NoModifierUsesBaseValue(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	// Get player - no modifiers added
	players := testGame.GetAllPlayers()
	player := players[0]

	// Create payment: 3 titanium without modifier
	payment := gamecards.CardPayment{
		Titanium: 3,
	}

	// Calculate total value with base values
	substitutes := player.Resources().PaymentSubstitutes()
	totalValue := payment.TotalValue(substitutes, nil)

	// 3 titanium * 3 MC each (base) = 9 MC
	testutil.AssertEqual(t, 9, totalValue, "3 titanium at base value 3 should equal 9 MC")
}

// TestValueModifier_DefaultValuesForAllPlayers tests that all players get steel/titanium in substitutes
func TestValueModifier_DefaultValuesForAllPlayers(t *testing.T) {
	// Setup with 4 players
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 4, broadcaster)

	// Check each player
	players := testGame.GetAllPlayers()
	for i, player := range players {
		substitutes := player.Resources().PaymentSubstitutes()

		var hasSteelSub, hasTitaniumSub bool
		for _, sub := range substitutes {
			if sub.ResourceType == shared.ResourceSteel && sub.ConversionRate == 2 {
				hasSteelSub = true
			}
			if sub.ResourceType == shared.ResourceTitanium && sub.ConversionRate == 3 {
				hasTitaniumSub = true
			}
		}

		testutil.AssertTrue(t, hasSteelSub, fmt.Sprintf("Player %d should have steel in substitutes with value 2", i))
		testutil.AssertTrue(t, hasTitaniumSub, fmt.Sprintf("Player %d should have titanium in substitutes with value 3", i))
	}
}
