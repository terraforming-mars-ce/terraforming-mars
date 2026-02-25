package card_effects_test

import (
	"context"
	"testing"

	baseaction "terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestIndenturedWorkers_DiscountAppliedToNextCard verifies that playing Indentured Workers
// creates a temporary 8 M€ discount that applies to the next card played.
func TestIndenturedWorkers_DiscountAppliedToNextCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-indentured-workers")
	player.Hand().AddCard("card-power-plant") // Cost 4

	// Play Indentured Workers (cost 0)
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playAction.Execute(ctx, testGame.ID(), player.ID(), "card-indentured-workers", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Indentured Workers")

	// Verify: temporary discount effect is registered
	effects := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effects), "Should have 1 effect after playing Indentured Workers")
	testutil.AssertEqual(t, "card-indentured-workers", effects[0].CardID, "Effect should be from Indentured Workers")

	// Verify: discount is calculated for next card
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	powerPlantCard, err := cardRegistry.GetByID("card-power-plant")
	testutil.AssertNoError(t, err, "Power Plant card should exist")

	discount := calculator.CalculateCardDiscounts(player, powerPlantCard)
	testutil.AssertEqual(t, 8, discount, "Next card should have 8 credit discount from Indentured Workers")

	// Give player another action
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	// Play Power Plant (cost 4, but with 8 discount → effective cost 0)
	payment = cardAction.PaymentRequest{Credits: 0}
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-power-plant", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Power Plant with Indentured Workers discount")

	// Verify: temporary effect is removed after playing the next card
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 0, len(effectsAfter), "Temporary effect should be removed after next card is played")
}

// TestIndenturedWorkers_DiscountRemovedAfterOneCard verifies that the discount only
// applies to one card and is then removed.
func TestIndenturedWorkers_DiscountRemovedAfterOneCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 200,
	})
	player.Hand().AddCard("card-indentured-workers")
	player.Hand().AddCard("card-power-plant") // Cost 4
	player.Hand().AddCard("card-dust-seals")  // Cost 2

	// Play Indentured Workers
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	err := playAction.Execute(ctx, testGame.ID(), player.ID(), "card-indentured-workers",
		cardAction.PaymentRequest{Credits: 0}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Indentured Workers")

	// Play Power Plant → discount should apply
	testGame.SetCurrentTurn(ctx, player.ID(), 2)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-power-plant",
		cardAction.PaymentRequest{Credits: 0}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Power Plant with discount")

	// Verify: discount is gone for the third card
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	dustSealsCard, err := cardRegistry.GetByID("card-dust-seals")
	testutil.AssertNoError(t, err, "Dust Seals card should exist")

	discount := calculator.CalculateCardDiscounts(player, dustSealsCard)
	testutil.AssertEqual(t, 0, discount, "No discount should remain for second card after Indentured Workers")
}

// TestSpecialDesign_LenienceAppliedToNextCard verifies that playing Special Design
// allows the next card to be played with ±2 global parameter lenience.
func TestSpecialDesign_LenienceAppliedToNextCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-special-design")
	player.Hand().AddCard("card-temp-req-test") // Requires temp >= -24°C

	// Temperature is at -30°C (default). Card requires -24°C.
	// Without lenience: can't play (need -24, have -30)
	// With +2 lenience: effective requirement is -26, still can't play (-30 < -26)
	// But wait, lenience of 2 means min is lowered by 2: -24 - 2 = -26, still -30 < -26

	// Let's increase temperature to -26
	testGame.GlobalParameters().IncreaseTemperature(ctx, 2) // -30 + 2*2 = -26

	// Without Special Design, temp is -26 and card requires -24 → can't play
	tempReqCard, err := cardRegistry.GetByID("card-temp-req-test")
	testutil.AssertNoError(t, err, "Temp req card should exist")

	state := baseaction.CalculatePlayerCardState(tempReqCard, player, testGame, cardRegistry)
	hasErrors := len(state.Errors) > 0
	testutil.AssertTrue(t, hasErrors, "Card should NOT be playable without lenience (temp -26, need -24)")

	// Play Special Design
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-special-design",
		cardAction.PaymentRequest{Credits: 4}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Special Design")

	// Verify: lenience effect is registered
	effects := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effects), "Should have 1 effect after playing Special Design")

	// Verify: lenience is calculated
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	lenience := calculator.CalculateGlobalParameterLenience(player)
	testutil.AssertEqual(t, 2, lenience, "Should have 2 lenience from Special Design")

	// Now the card should be playable: temp -26, requirement -24 with lenience 2 → effective min -26
	state = baseaction.CalculatePlayerCardState(tempReqCard, player, testGame, cardRegistry)
	playable := len(state.Errors) == 0
	testutil.AssertTrue(t, playable, "Card should be playable with Special Design lenience (temp -26, need -24, lenience 2 → effective -26)")

	// Play the temperature requirement card
	testGame.SetCurrentTurn(ctx, player.ID(), 2)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-temp-req-test",
		cardAction.PaymentRequest{Credits: 5}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should be able to play with Special Design lenience")

	// Verify: temporary effect is removed after playing the next card
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 0, len(effectsAfter), "Temporary effect should be removed after next card is played")
}

// TestSpecialDesign_LenienceWithMaxRequirement verifies lenience also works for max requirements.
func TestSpecialDesign_LenienceWithMaxRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-special-design")
	player.Hand().AddCard("card-oxygen-max-req-test") // Requires oxygen <= 5%

	// Increase oxygen to 7% (3.5 steps from default 0)
	testGame.GlobalParameters().IncreaseOxygen(ctx, 7)

	// Without lenience: oxygen 7, max 5 → can't play
	oxygenCard, err := cardRegistry.GetByID("card-oxygen-max-req-test")
	testutil.AssertNoError(t, err, "Oxygen card should exist")

	state := baseaction.CalculatePlayerCardState(oxygenCard, player, testGame, cardRegistry)
	hasErrors := len(state.Errors) > 0
	testutil.AssertTrue(t, hasErrors, "Card should NOT be playable without lenience (oxygen 7, max 5)")

	// Play Special Design
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-special-design",
		cardAction.PaymentRequest{Credits: 4}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Special Design")

	// With lenience 2: effective max = 5 + 2 = 7. Oxygen 7 <= 7 → playable
	state = baseaction.CalculatePlayerCardState(oxygenCard, player, testGame, cardRegistry)
	playable := len(state.Errors) == 0
	testutil.AssertTrue(t, playable, "Card should be playable with lenience (oxygen 7, max 5+2=7)")

	// Play the card
	testGame.SetCurrentTurn(ctx, player.ID(), 2)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-oxygen-max-req-test",
		cardAction.PaymentRequest{Credits: 5}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should be able to play with Special Design lenience")
}

// TestTemporaryEffects_ClearedOnGenerationAdvance verifies that temporary effects
// are cleared when the generation advances.
func TestTemporaryEffects_ClearedOnGenerationAdvance(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player.Hand().AddCard("card-indentured-workers")

	// Play Indentured Workers
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	err := playAction.Execute(ctx, testGame.ID(), player.ID(), "card-indentured-workers",
		cardAction.PaymentRequest{Credits: 0}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Indentured Workers")

	// Verify effect is registered
	effects := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effects), "Should have 1 effect")

	// Advance generation → should clear temporary effects
	testGame.AdvanceGeneration(ctx)

	// Verify effects are cleared
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 0, len(effectsAfter), "Temporary effects should be cleared on generation advance")
}

// TestIndenturedWorkers_WithExistingPermanentDiscount verifies that playing
// Indentured Workers doesn't affect existing permanent discounts.
func TestIndenturedWorkers_WithExistingPermanentDiscount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 200,
	})
	player.Hand().AddCard("card-space-station") // Permanent -2 for space tags
	player.Hand().AddCard("card-indentured-workers")
	player.Hand().AddCard("card-space-mirrors") // Space tag, cost 3

	// Play Space Station (permanent space discount of 2)
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	err := playAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-station",
		cardAction.PaymentRequest{Credits: 10}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Space Station")

	// Play Indentured Workers (temporary discount of 8)
	testGame.SetCurrentTurn(ctx, player.ID(), 2)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-indentured-workers",
		cardAction.PaymentRequest{Credits: 0}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Indentured Workers")

	// Verify: both effects exist (1 permanent + 1 temporary)
	effects := player.Effects().List()
	testutil.AssertEqual(t, 2, len(effects), "Should have 2 effects (permanent + temporary)")

	// Verify: combined discount for Space Mirrors = 2 (space) + 8 (indentured) = 10
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	spaceMirrors, _ := cardRegistry.GetByID("card-space-mirrors")
	discount := calculator.CalculateCardDiscounts(player, spaceMirrors)
	testutil.AssertEqual(t, 10, discount, "Space Mirrors should have combined discount of 10")

	// Play Space Mirrors → temporary effect consumed
	testGame.SetCurrentTurn(ctx, player.ID(), 2)
	err = playAction.Execute(ctx, testGame.ID(), player.ID(), "card-space-mirrors",
		cardAction.PaymentRequest{Credits: 0}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Failed to play Space Mirrors")

	// Verify: only permanent effect remains
	effectsAfter := player.Effects().List()
	testutil.AssertEqual(t, 1, len(effectsAfter), "Only permanent effect should remain after temporary is consumed")
	testutil.AssertEqual(t, "card-space-station", effectsAfter[0].CardID, "Remaining effect should be Space Station")
}
