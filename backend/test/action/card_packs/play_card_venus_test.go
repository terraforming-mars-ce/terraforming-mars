package card_packs_test

import (
	"context"
	"terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
	"testing"
)

// =============================================================================
// Card 213: Aerial Mappers (active)
// "Action: Add 1 floater to any card, or spend 1 floater here to draw a card."
// Has floater resource storage. 1 VP fixed.
// =============================================================================
func TestAerialMappers_PlayAndStorageCreated(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Aerial Mappers")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Aerial Mappers should play successfully")
	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 0, storage, "Aerial Mappers should start with 0 floaters")
}
func TestAerialMappers_Action_AddFloater(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-aerial-mappers"
	p.PlayedCards().AddCard(cardID, "Aerial Mappers", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 0)
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{
		{ID: cardID, Name: "Aerial Mappers", Type: gamecards.CardTypeActive, Tags: []shared.CardTag{shared.TagVenus}, ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceFloater}},
	})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "any-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Aerial Mappers", BehaviorIndex: 0, Behavior: behavior},
	})
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add floater) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater after adding")
}
func TestAerialMappers_Action_SpendFloaterForCardDraw(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-aerial-mappers"
	p.PlayedCards().AddCard(cardID, "Aerial Mappers", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 2)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "any-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Aerial Mappers", BehaviorIndex: 0, Behavior: behavior},
	})
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 1 (spend floater for card draw) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater after spending 1 from 2")
}

// =============================================================================
// Card 214: Aerosport Tournament (event)
// "Requires that you have 5 floaters. Gain 1 M€ for each city tile in play."
// 1 VP fixed.
// =============================================================================
func TestAerosportTournament_RequiresFloaters(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Aerosport Tournament")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	// Player has no floaters - should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Aerosport Tournament should fail without 5 floaters")
}

// =============================================================================
// Card 216: Atalanta Planitia Lab (automated)
// "Requires 3 science tags. Draw 2 cards." Tags: science, venus. 2 VP.
// =============================================================================
func TestAtalantaPlanitiaLab_RequiresScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Atalanta Planitia Lab")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	// Player has 0 science tags - should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Atalanta Planitia Lab should fail without 3 science tags")
}
func TestAtalantaPlanitiaLab_SucceedsWithScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Atalanta Planitia Lab")
	sciCard1 := gamecards.Card{ID: "sci-1", Name: "Sci 1", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagScience}}
	sciCard2 := gamecards.Card{ID: "sci-2", Name: "Sci 2", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagScience}}
	sciCard3 := gamecards.Card{ID: "sci-3", Name: "Sci 3", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagScience}}
	additionalCards := []gamecards.Card{sciCard1, sciCard2, sciCard3}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.PlayedCards().AddCard("sci-1", "Sci 1", "automated", []string{"science"})
	p.PlayedCards().AddCard("sci-2", "Sci 2", "automated", []string{"science"})
	p.PlayedCards().AddCard("sci-3", "Sci 3", "automated", []string{"science"})
	p.Hand().AddCard(card.ID)
	handBefore := p.Hand().CardCount()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Atalanta Planitia Lab should succeed with 3 science tags")
	handAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handBefore-1+2, handAfter, "Should draw 2 cards (hand: -1 played +2 drawn)")
}

// =============================================================================
// Card 219: Corroder Suits (automated)
// "Increase your M€ production 2 steps. Add 1 resource to any venus card."
// Tags: venus.
// =============================================================================
func TestCorroderSuits_CreditProductionIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Corroder Suits")
	venusTargetCard := gamecards.Card{
		ID:   "venus-target-card",
		Name: "Venus Target",
		Type: gamecards.CardTypeActive,
		Pack: "venus-next",
		Cost: 1,
		Tags: []shared.CardTag{shared.TagVenus},
		ResourceStorage: &gamecards.ResourceStorage{
			Type:     shared.ResourceFloater,
			Starting: 0,
		},
	}
	additionalCards := []gamecards.Card{venusTargetCard}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	p.PlayedCards().AddCard("venus-target-card", "Venus Target", "active", []string{"venus"})
	p.Resources().AddToStorage("venus-target-card", 0)
	prodBefore := p.Resources().Production()
	targetCardID := "venus-target-card"
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "Corroder Suits should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits, "Credit production should increase by 2")
}

// =============================================================================
// Card 221: Deuterium Export (active)
// "Action: Add 1 floater to this card, or spend 1 floater here to increase
//
//	your energy production 1 step."
//
// Tags: power, space, venus. Has floater storage.
// =============================================================================
func TestDeuteriumExport_PlayAndStorageCreated(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Deuterium Export")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Deuterium Export should play successfully")
	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 0, storage, "Deuterium Export should start with 0 floaters")
}
func TestDeuteriumExport_Action_AddFloater(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-deuterium-export"
	p.PlayedCards().AddCard(cardID, "Deuterium Export", "active", []string{"power", "space", "venus"})
	p.Resources().AddToStorage(cardID, 0)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Deuterium Export", BehaviorIndex: 0, Behavior: behavior},
	})
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add floater) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater after adding")
}
func TestDeuteriumExport_Action_SpendFloaterForEnergyProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-deuterium-export"
	p.PlayedCards().AddCard(cardID, "Deuterium Export", "active", []string{"power", "space", "venus"})
	p.Resources().AddToStorage(cardID, 3)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Deuterium Export", BehaviorIndex: 0, Behavior: behavior},
	})
	prodBefore := p.Resources().Production()
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 1 (spend floater for energy production) should succeed")
	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Card should have 2 floaters after spending 1 from 3")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy+1, prodAfter.Energy, "Energy production should increase by 1")
}

// =============================================================================
// Card 223: Extractor Balloons (active)
// "Action: Add 1 floater to this card, or remove 2 floaters here to raise
//
//	Venus 1 step. Add 3 floaters to this card." (on play)
//
// Tags: venus. Has floater storage.
// =============================================================================
func TestExtractorBalloons_PlayAdds3Floaters(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Extractor Balloons")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Extractor Balloons should play successfully")
	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 3, storage, "Extractor Balloons should have 3 floaters after playing")
}

// =============================================================================
// Card 224: Extremophiles (active)
// "Action: Add 1 microbe to any card. Requires 2 science tags."
// Tags: microbe, venus. Has microbe storage. 1 VP per 3 microbes.
// =============================================================================
func TestExtremophiles_RequiresScienceTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Extremophiles")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	// No science tags - should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Extremophiles should fail without 2 science tags")
}

// =============================================================================
// Card 225: Floating Habs (active)
// "Action: Spend 2 M€ to add 1 floater to any card. Requires 2 science tags."
// Tags: venus. Has floater storage. 1 VP per 2 floaters.
// =============================================================================
func TestFloatingHabs_Action_Spend2CreditsForFloater(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-floating-habs"
	p.PlayedCards().AddCard(cardID, "Floating Habs", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 0)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 50})
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{
		{ID: cardID, Name: "Floating Habs", Type: gamecards.CardTypeActive, Tags: []shared.CardTag{shared.TagVenus}, ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceFloater}},
	})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceFloater, Amount: 1, Target: "any-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Floating Habs", BehaviorIndex: 0, Behavior: behavior},
	})
	creditsBefore := p.Resources().Get().Credits
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Floating Habs action should succeed")
	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-2, creditsAfter, "Should spend 2 credits")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater")
}

// =============================================================================
// Card 226: Forced Precipitation (active)
// "Action: Spend 2 M€ to add a floater to this card, or spend 2 floaters
//
//	here to increase Venus 1 step."
//
// Tags: venus. Has floater storage.
// =============================================================================
func TestForcedPrecipitation_Action_PayCreditsForFloater(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-forced-precipitation"
	p.PlayedCards().AddCard(cardID, "Forced Precipitation", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 0)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 50})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 2, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Forced Precipitation", BehaviorIndex: 0, Behavior: behavior},
	})
	creditsBefore := p.Resources().Get().Credits
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (spend credits for floater) should succeed")
	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-2, creditsAfter, "Should spend 2 credits")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater")
}
func TestForcedPrecipitation_Action_FailsWithoutEnoughFloaters(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-forced-precipitation"
	p.PlayedCards().AddCard(cardID, "Forced Precipitation", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 1) // Only 1 floater, need 2
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 2, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Forced Precipitation", BehaviorIndex: 0, Behavior: behavior},
	})
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Choice 1 should fail with only 1 floater (need 2)")
}

// =============================================================================
// Card 228: GHG Import From Venus (event)
// "Raise Venus 1 step. Increase your heat production 3 steps."
// Tags: space, venus.
// =============================================================================
func TestGHGImportFromVenus_HeatProductionIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("GHG Import From Venus")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "GHG Import From Venus should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+3, prodAfter.Heat, "Heat production should increase by 3")
}

// =============================================================================
// Card 232: Io Sulphur Research (automated)
// "Draw 1 card, or draw 3 cards if you have at least 3 Venus tags."
// Tags: jovian, science. 2 VP.
// =============================================================================
func TestIoSulphurResearch_DrawOneCardWithoutVenusTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Io Sulphur Research")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	handBefore := p.Hand().CardCount()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 17}
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, &choiceIndex, nil, nil, nil)
	testutil.AssertNoError(t, err, "Io Sulphur Research choice 0 should succeed")
	handAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handBefore-1+1, handAfter, "Should draw 1 card (hand: -1 played +1 drawn)")
}
func TestIoSulphurResearch_FailsDraw3WithoutVenusTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Io Sulphur Research")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	// No venus tags - choice 1 (draw 3) should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 17}
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, &choiceIndex, nil, nil, nil)
	testutil.AssertError(t, err, "Io Sulphur Research choice 1 should fail without 3 venus tags")
}

func TestIoSulphurResearch_FailsDraw3WithOnly2VenusTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 200})

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)

	// Play 2 venus-tagged cards: Dirigibles (cost 11) and Jet Stream Microscrappers (cost 12)
	dirigibles := testutil.GetCardByName("Dirigibles")
	p.Hand().AddCard(dirigibles.ID)
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), dirigibles.ID, cardAction.PaymentRequest{Credits: 11}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Dirigibles should play successfully")

	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	jetStream := testutil.GetCardByName("Jet Stream Microscrappers")
	p.Hand().AddCard(jetStream.ID)
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), jetStream.ID, cardAction.PaymentRequest{Credits: 12}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Jet Stream Microscrappers should play successfully")

	// Verify player has exactly 2 venus tags
	venusTagCount := gamecards.CountPlayerTagsByType(p, cardRegistry, shared.TagVenus)
	testutil.AssertEqual(t, 2, venusTagCount, "Player should have exactly 2 venus tags")

	// Now play Io Sulphur Research with choice 1 (draw 3, requires 3+ venus tags) — should fail
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	ioSulphur := testutil.GetCardByName("Io Sulphur Research")
	p.Hand().AddCard(ioSulphur.ID)
	choiceIndex := 1
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), ioSulphur.ID, cardAction.PaymentRequest{Credits: 17}, &choiceIndex, nil, nil, nil)
	testutil.AssertError(t, err, "Io Sulphur Research choice 1 should fail with only 2 venus tags")

	// Verify card is still in hand (play was rejected)
	testutil.AssertTrue(t, p.Hand().HasCard(ioSulphur.ID), "Io Sulphur Research should remain in hand after failed play")

	// Also verify state calculator marks choice 1 as unavailable
	for _, behavior := range ioSulphur.Behaviors {
		if len(behavior.Choices) >= 2 {
			errors := action.CalculateChoiceErrors(behavior.Choices[1], p, testGame, cardRegistry)
			testutil.AssertTrue(t, len(errors) > 0, "State calculator should report errors for choice 1 with only 2 venus tags")
		}
	}
}

// =============================================================================
// Card 233: Ishtar Mining (automated)
// "Requires Venus 8%. Increase your titanium production 1 step."
// Tags: venus.
// =============================================================================
func TestIshtarMining_TitaniumProductionIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Ishtar Mining")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	testGame.GlobalParameters().SetVenus(ctx, 8)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ishtar Mining should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium, "Titanium production should increase by 1")
}

// =============================================================================
// Card 234: Jet Stream Microscrappers (active)
// "Action: Spend 1 titanium to add 2 floaters to this card, or remove 2
//
//	floaters here to raise Venus 1 step."
//
// Tags: venus. Has floater storage.
// =============================================================================
func TestJetStreamMicroscrappers_Action_SpendTitaniumFor2Floaters(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-jet-stream-microscrappers"
	p.PlayedCards().AddCard(cardID, "Jet Stream Microscrappers", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 0)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceTitanium: 5})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitanium, Amount: 1, Target: "self-player"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 2, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 2, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Jet Stream Microscrappers", BehaviorIndex: 0, Behavior: behavior},
	})
	titaniumBefore := p.Resources().Get().Titanium
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (spend titanium for floaters) should succeed")
	titaniumAfter := p.Resources().Get().Titanium
	testutil.AssertEqual(t, titaniumBefore-1, titaniumAfter, "Should spend 1 titanium")
	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Card should have 2 floaters")
}

// =============================================================================
// Card 235: Local Shading (active)
// "Action: Add 1 floater to this card, or spend 1 floater here to raise your
//
//	M€ production 1 step."
//
// Tags: venus. Has floater storage.
// =============================================================================
func TestLocalShading_Action_AddFloater(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-local-shading"
	p.PlayedCards().AddCard(cardID, "Local Shading", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 0)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Local Shading", BehaviorIndex: 0, Behavior: behavior},
	})
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add floater) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater")
}
func TestLocalShading_Action_SpendFloaterForCreditProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-local-shading"
	p.PlayedCards().AddCard(cardID, "Local Shading", "active", []string{"venus"})
	p.Resources().AddToStorage(cardID, 3)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{CardID: cardID, CardName: "Local Shading", BehaviorIndex: 0, Behavior: behavior},
	})
	prodBefore := p.Resources().Production()
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 1 (spend floater for credit production) should succeed")
	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Card should have 2 floaters after spending 1")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+1, prodAfter.Credits, "Credit production should increase by 1")
}

// =============================================================================
// Card 236: Luna Metropolis (automated)
// "Increase your M€ production 1 step for each Earth tag you have, including
//
//	this. Place a city tile on the reserved area."
//
// Tags: city, earth, space.  2 VP.
// =============================================================================
func TestLunaMetropolis_CreditProductionPerEarthTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Luna Metropolis")
	earthCard1 := gamecards.Card{ID: "earth-1", Name: "Earth Card 1", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	earthCard2 := gamecards.Card{ID: "earth-2", Name: "Earth Card 2", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	additionalCards := []gamecards.Card{earthCard1, earthCard2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	// Add 2 earth-tagged played cards
	p.PlayedCards().AddCard("earth-1", "Earth Card 1", "automated", []string{"earth"})
	p.PlayedCards().AddCard("earth-2", "Earth Card 2", "automated", []string{"earth"})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Luna Metropolis should play successfully")
	prodAfter := p.Resources().Production()
	// 2 existing earth tags + 1 from this card (earth tag) = 3 earth tags
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3 (1 per earth tag: 2 existing + 1 from this card)")
}

// =============================================================================
// Card 237: Luxury Foods (automated)
// "Requires Venus, Earth and Jovian tags." No behaviors, just VP.
// 2 VP fixed.
// =============================================================================
func TestLuxuryFoods_FailsWithoutRequiredTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Luxury Foods")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	// No tags at all - should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Luxury Foods should fail without required tags")
}
func TestLuxuryFoods_SucceedsWithAllTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Luxury Foods")
	earthCard := gamecards.Card{ID: "earth-tag-card", Name: "Earth Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	jovianCard := gamecards.Card{ID: "jovian-tag-card", Name: "Jovian Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagJovian}}
	venusCard := gamecards.Card{ID: "venus-tag-card", Name: "Venus Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	additionalCards := []gamecards.Card{earthCard, jovianCard, venusCard}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.PlayedCards().AddCard("earth-tag-card", "Earth Card", "automated", []string{"earth"})
	p.PlayedCards().AddCard("jovian-tag-card", "Jovian Card", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("venus-tag-card", "Venus Card", "automated", []string{"venus"})
	p.Hand().AddCard(card.ID)
	creditsBefore := p.Resources().Get().Credits
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Luxury Foods should succeed with earth, jovian, and venus tags")
	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-8, creditsAfter, "Should only pay 8 credits (no other effects)")
}

// =============================================================================
// Card 230: Gyropolis (automated)
// "Decrease your energy production 2 steps. Increase your M€ production 1
//
//	step for each Venus and Earth tag you have. Place a city tile."
//
// Tags: building, city.
// =============================================================================
func TestGyropolis_ProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Gyropolis")
	earthCard := gamecards.Card{ID: "earth-g", Name: "Earth Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	venusCard1 := gamecards.Card{ID: "venus-g1", Name: "Venus Card 1", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	venusCard2 := gamecards.Card{ID: "venus-g2", Name: "Venus Card 2", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	additionalCards := []gamecards.Card{earthCard, venusCard1, venusCard2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 3,
	})
	p.PlayedCards().AddCard("earth-g", "Earth Card", "automated", []string{"earth"})
	p.PlayedCards().AddCard("venus-g1", "Venus Card 1", "automated", []string{"venus"})
	p.PlayedCards().AddCard("venus-g2", "Venus Card 2", "automated", []string{"venus"})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 20}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Gyropolis should play successfully")
	prodAfter := p.Resources().Production()
	// 1 earth tag + 2 venus tags = 3 credit production increase
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3 (1 earth + 2 venus tags)")
	testutil.AssertEqual(t, prodBefore.Energy-2, prodAfter.Energy,
		"Energy production should decrease by 2")
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

// =============================================================================
// Card 239: Mining Quota
// "Requires Venus, Earth and Jovian tags. Increase your steel production 2 steps."
// =============================================================================
func TestMiningQuota_IncreaseSteelProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Mining Quota")
	earthHelper := gamecards.Card{ID: "earth-1", Name: "Earth Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	jovianHelper := gamecards.Card{ID: "jovian-1", Name: "Jovian Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagJovian}}
	venusHelper := gamecards.Card{ID: "venus-1", Name: "Venus Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	additionalCards := []gamecards.Card{earthHelper, jovianHelper, venusHelper}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.PlayedCards().AddCard("earth-1", "Earth Card", "automated", []string{"earth"})
	p.PlayedCards().AddCard("jovian-1", "Jovian Card", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("venus-1", "Venus Card", "automated", []string{"venus"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mining Quota should play successfully with required tags")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+2, prodAfter.Steel, "Steel production should increase by 2")
}
func TestMiningQuota_FailsWithoutRequiredTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Mining Quota")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	// Only earth tag, missing jovian and venus
	p.PlayedCards().AddCard("earth-1", "Earth Card", "automated", []string{"earth"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Mining Quota should fail without jovian and venus tags")
}

// =============================================================================
// Card 240: Neutralizer Factory
// "Requires Venus 10%. Increase Venus 1 step."
// =============================================================================
func TestNeutralizerFactory_IncreaseVenus(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Neutralizer Factory")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	testGame.GlobalParameters().SetVenus(ctx, 10)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Neutralizer Factory should play successfully")
}

// =============================================================================
// Card 241: Omnicourt
// "Requires Venus, Earth, and Jovian tags. Increase your TR 2 steps."
// =============================================================================
func TestOmnicourt_IncreaseTR(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Omnicourt")
	earthHelper := gamecards.Card{ID: "earth-1", Name: "Earth Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	jovianHelper := gamecards.Card{ID: "jovian-1", Name: "Jovian Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagJovian}}
	venusHelper := gamecards.Card{ID: "venus-1", Name: "Venus Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	additionalCards := []gamecards.Card{earthHelper, jovianHelper, venusHelper}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.PlayedCards().AddCard("earth-1", "Earth Card", "automated", []string{"earth"})
	p.PlayedCards().AddCard("jovian-1", "Jovian Card", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("venus-1", "Venus Card", "automated", []string{"venus"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	trBefore := p.Resources().TerraformRating()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Omnicourt should play successfully with required tags")
	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+2, trAfter, "TR should increase by 2")
}

// =============================================================================
// Card 242: Orbital Reflectors
// "Raise Venus 2 steps. Increase your heat production 2 steps."
// =============================================================================
func TestOrbitalReflectors_HeatProductionAndVenus(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Orbital Reflectors")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 26}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Orbital Reflectors should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Heat+2, prodAfter.Heat, "Heat production should increase by 2")
}

// =============================================================================
// Card 243: Rotator Impacts
// "Action: Spend 6 M€ to add a floater to this card (titanium may be used), or
// spend 1 floater here to increase Venus 1 step."
// =============================================================================
func TestRotatorImpacts_AddFloater(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-rotator-impacts"
	p.PlayedCards().AddCard(cardID, "Rotator Impacts", "active", []string{"space"})
	p.Resources().AddToStorage(cardID, 0)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 6, Target: "self-player", PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium}},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceFloater, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Rotator Impacts",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add floater) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 floater")
}

// =============================================================================
// Card 244: Sister Planet Support
// "Requires Venus and Earth tags. Increase your M€ production 3 steps."
// =============================================================================
func TestSisterPlanetSupport_CreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Sister Planet Support")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	// Self-tags satisfy the requirement since the card itself has venus+earth
	p.PlayedCards().AddCard("venus-1", "Venus Card", "automated", []string{"venus"})
	p.PlayedCards().AddCard("earth-1", "Earth Card", "automated", []string{"earth"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Sister Planet Support should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits, "Credit production should increase by 3")
}

// =============================================================================
// Card 245: Solarnet
// "Requires Venus, Earth, and Jovian tags. Draw 2 cards."
// =============================================================================
func TestSolarnet_Draw2Cards(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Solarnet")
	venusHelper := gamecards.Card{ID: "venus-1", Name: "Venus Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	earthHelper := gamecards.Card{ID: "earth-1", Name: "Earth Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagEarth}}
	jovianHelper := gamecards.Card{ID: "jovian-1", Name: "Jovian Card", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagJovian}}
	additionalCards := []gamecards.Card{venusHelper, earthHelper, jovianHelper}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.PlayedCards().AddCard("venus-1", "Venus Card", "automated", []string{"venus"})
	p.PlayedCards().AddCard("earth-1", "Earth Card", "automated", []string{"earth"})
	p.PlayedCards().AddCard("jovian-1", "Jovian Card", "automated", []string{"jovian"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	handSizeBefore := len(p.Hand().Cards())
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Solarnet should play successfully")
	handSizeAfter := len(p.Hand().Cards())
	// -1 from playing the card, +2 from drawing = net +1
	testutil.AssertEqual(t, handSizeBefore+1, handSizeAfter,
		"Hand should have net +1 card (played 1, drew 2)")
}

// =============================================================================
// Card 246: Spin-Inducing Asteroid
// "Raise Venus 2 steps."
// =============================================================================
func TestSpinInducingAsteroid_RaiseVenus(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Spin-Inducing Asteroid")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 16}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Spin-Inducing Asteroid should play successfully")
}

// =============================================================================
// Card 247: Sponsored Academies
// "Discard 1 card from hand and **then** draw 3 cards. All **opponents** draw 1 card."
// =============================================================================
func TestSponsoredAcademies_DiscardDrawAndOpponentDraw(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Sponsored Academies")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	opponent := players[1]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	opponent.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	// Player has: Sponsored Academies + 2 fodder cards = 3 cards in hand
	p.Hand().AddCard(card.ID)
	p.Hand().AddCard("card-fodder-1")
	p.Hand().AddCard("card-fodder-2")

	playerHandBefore := p.Hand().CardCount()             // 3
	opponentHandBefore := opponent.Hand().CardCount()     // 0

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Sponsored Academies should play successfully")

	playerHandAfter := p.Hand().CardCount()
	opponentHandAfter := opponent.Hand().CardCount()

	// Player: started with 3, played 1 (-1), discarded 1 (-1), drew 3 (+3) = net +1 = 4
	testutil.AssertEqual(t, playerHandBefore+1, playerHandAfter,
		"Player should have net +1 cards (played 1, discarded 1, drew 3)")

	// Opponent: drew 1 card (+1)
	testutil.AssertEqual(t, opponentHandBefore+1, opponentHandAfter,
		"Opponent should have drawn 1 card from Sponsored Academies")
}

// =============================================================================
// Card 249: Stratospheric Birds
// "Action: Add 1 animal to this card. Requires Venus 12%.
// Remove 1 floater from any card. 1 VP per animal on this card."
// =============================================================================
func TestStratosphericBirds_ActionAddAnimal(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-stratospheric-birds"
	p.PlayedCards().AddCard(cardID, "Stratospheric Birds", "active", []string{"animal", "venus"})
	p.Resources().AddToStorage(cardID, 0)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Stratospheric Birds",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Adding animal via action should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 animal")
}

// =============================================================================
// Card 250: Sulphur Exports
// "Increase your M€ production 1 step per Venus tag you have, including this."
// =============================================================================
func TestSulphurExports_CreditProductionPerVenusTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Sulphur Exports")
	venusHelper1 := gamecards.Card{ID: "venus-1", Name: "Venus Card 1", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	venusHelper2 := gamecards.Card{ID: "venus-2", Name: "Venus Card 2", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagVenus}}
	additionalCards := []gamecards.Card{venusHelper1, venusHelper2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	// 2 existing venus tags + 1 from card itself = 3 total
	p.PlayedCards().AddCard("venus-1", "Venus Card 1", "automated", []string{"venus"})
	p.PlayedCards().AddCard("venus-2", "Venus Card 2", "automated", []string{"venus"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Sulphur Exports should play successfully")
	prodAfter := p.Resources().Production()
	// 2 existing + 1 from card = 3 venus tags -> +3 credit production
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3 (1 per venus tag, 3 total)")
}

// =============================================================================
// Card 252: Terraforming Contract
// "Requires TR 25 or higher. Increase your M€ production 4 steps."
// =============================================================================
func TestTerraformingContract_CreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Terraforming Contract")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	// Raise TR to 25
	p.Resources().UpdateTerraformRating(5) // starts at 20, add 5 to reach 25
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Terraforming Contract should play successfully at TR 25")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+4, prodAfter.Credits,
		"Credit production should increase by 4")
}
func TestTerraformingContract_FailsBelowTR25(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Terraforming Contract")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	// TR stays at default (20), below 25
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Terraforming Contract should fail below TR 25")
}

// =============================================================================
// Card 253: Thermophiles
// "Action: Add 1 microbe to this card, or spend 2 microbes here to raise Venus 1 step."
// Requires Venus 6%.
// =============================================================================
func TestThermophiles_AddMicrobe(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-thermophiles"
	p.PlayedCards().AddCard(cardID, "Thermophiles", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 0)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 2, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Thermophiles",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add microbe) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 microbe")
}
func TestThermophiles_SpendMicrobesForVenus(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-thermophiles"
	p.PlayedCards().AddCard(cardID, "Thermophiles", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 3)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 2, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Thermophiles",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 1 (spend 2 microbes for Venus) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 microbe after spending 2 from 3")
}

// =============================================================================
// Card 254: Water To Venus
// "Raise Venus 1 step."
// =============================================================================
func TestWaterToVenus_RaiseVenus(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Water To Venus")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Water To Venus should play successfully")
}

// =============================================================================
// Card 255: Venus Governor
// "Requires 2 Venus tags. Increase your M€ production 2 steps."
// =============================================================================
func TestVenusGovernor_CreditProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Venus Governor")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	testGame.GlobalParameters().SetVenus(ctx, 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Venus Governor should play with Venus >= 2%")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+2, prodAfter.Credits,
		"Credit production should increase by 2")
}

// =============================================================================
// Card 256: Venus Magnetizer
// "Action: Decrease your energy production 1 step to raise Venus 1 step."
// Requires Venus 10%.
// =============================================================================
func TestVenusMagnetizer_ActionDecraseEnergyForVenus(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	cardID := "test-venus-magnetizer"
	p.PlayedCards().AddCard(cardID, "Venus Magnetizer", "active", []string{"venus"})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergyProduction, Amount: 1, Target: "self-player"},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Venus Magnetizer",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	prodBefore := p.Resources().Production()
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Venus Magnetizer action should succeed")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
}

// =============================================================================
// Card 257: Venus Soils
// "Raise Venus 1 step. Increase your plant production 1 step. Add 2 microbes to another card."
// =============================================================================
func TestVenusSoils_PlantProductionAndVenus(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Venus Soils")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 20}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Venus Soils should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Plants+1, prodAfter.Plants,
		"Plant production should increase by 1")
}

// =============================================================================
// Card 258: Venus Waystation
// "Effect: When you play a Venus tag, you pay 2 M€ less for it."
// =============================================================================
func TestVenusWaystation_DiscountEffect(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Venus Waystation")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Venus Waystation should play successfully and register discount effect")
}

// =============================================================================
// Card 260: Venusian Insects
// "Action: Add 1 microbe to this card. Requires Venus 12%.
// 1 VP per 2 microbes on this card."
// =============================================================================
func TestVenusianInsects_ActionAddMicrobe(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-venusian-insects"
	p.PlayedCards().AddCard(cardID, "Venusian Insects", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 0)
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Venusian Insects",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Adding microbe via action should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 microbe")
}

// =============================================================================
// Card 238: Maxwell Base
// "Decrease your energy production 1 step. Place a city tile on the reserved area."
// "Action: Add 1 resource to another venus card."
// Requires Venus 12%.
// =============================================================================
func TestMaxwellBase_DecreaseEnergyProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Maxwell Base")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard(card.ID)
	testGame.GlobalParameters().SetVenus(ctx, 12)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Maxwell Base should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
}

// =============================================================================
// Card 248: Stratopolis
// "Requires 2 science tags. Place a city tile.
// Action: Add 1 floater to this card, or add 2 floaters to this card.
// 1 VP per 3 floaters on this card."
// =============================================================================
func TestStratopolis_CityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Stratopolis")
	sciHelper1 := gamecards.Card{ID: "science-1", Name: "Science Card 1", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagScience}}
	sciHelper2 := gamecards.Card{ID: "science-2", Name: "Science Card 2", Type: gamecards.CardTypeAutomated, Pack: "base", Cost: 1, Tags: []shared.CardTag{shared.TagScience}}
	additionalCards := []gamecards.Card{sciHelper1, sciHelper2}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.PlayedCards().AddCard("science-1", "Science Card 1", "automated", []string{"science"})
	p.PlayedCards().AddCard("science-2", "Science Card 2", "automated", []string{"science"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 22}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Stratopolis should play successfully with 2 science tags")
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}

// =============================================================================
// Card 259: Venusian Animals
// "Effect: When you play a science tag, including this, add 1 animal to this card.
// Requires Venus 18%. 1 VP for each animal on this card."
// =============================================================================
func TestVenusianAnimals_PlaysAndRegistersEffect(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Venusian Animals")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	testGame.GlobalParameters().SetVenus(ctx, 18)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Venusian Animals should play successfully")
}

// =============================================================================
// Card 261: Venusian Plants
// "Requires Venus 16%. Raise Venus 1 step. Add 1 microbe or 1 animal to another venus card."
// Uses choices for the microbe/animal selection.
// =============================================================================
func TestVenusianPlants_RaiseVenusWithAnimalChoice(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Venusian Plants")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	testGame.GlobalParameters().SetVenus(ctx, 16)
	choiceIndex := 0 // choose animal
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, &choiceIndex, nil, nil, nil)
	testutil.AssertNoError(t, err, "Venusian Plants should play with animal choice")
}

// =============================================================================
// Card 251: Sulphur-Eating Bacteria (active)
// "Action: Add 1 microbe to this card, or spend any number of microbes here
//
//	to gain triple that amount of M€."
//
// Requires Venus 6%. Tags: microbe, venus. Has microbe storage.
// =============================================================================
func sulphurEatingBacteriaBehavior() shared.CardBehavior {
	return shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card", VariableAmount: true},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCredit, Amount: 3, Target: "self-player", VariableAmount: true},
				},
			},
		},
	}
}
func TestSulphurEatingBacteria_Choice0_AddMicrobe(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-sulphur-eating-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 0)
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      sulphurEatingBacteriaBehavior(),
		},
	})
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Choice 0 (add microbe) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 microbe after adding")
}
func TestSulphurEatingBacteria_Choice1_SpendMicrobesForCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-sulphur-eating-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 3)
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      sulphurEatingBacteriaBehavior(),
		},
	})
	creditsBefore := p.Resources().Get().Credits
	choiceIndex := 1
	selectedAmount := 2
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Choice 1 (spend microbes for credits) should succeed")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(cardID), "Card should have 1 microbe after spending 2 from 3")
	testutil.AssertEqual(t, creditsBefore+6, p.Resources().Get().Credits, "Should gain 6 credits (2 microbes * 3)")
}
func TestSulphurEatingBacteria_Choice1_FailsWithoutSelectedAmount(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-sulphur-eating-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 3)
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      sulphurEatingBacteriaBehavior(),
		},
	})
	creditsBefore := p.Resources().Get().Credits
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Choice 1 without selectedAmount should fail")
	testutil.AssertEqual(t, 3, p.Resources().GetCardStorage(cardID), "Microbes should be unchanged")
	testutil.AssertEqual(t, creditsBefore, p.Resources().Get().Credits, "Credits should be unchanged")
	actions := p.Actions().List()
	testutil.AssertEqual(t, 0, actions[0].TimesUsedThisGeneration, "Action should not be marked as used")
}
func TestSulphurEatingBacteria_Choice1_SpendAllMicrobes(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-sulphur-eating-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 5)
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      sulphurEatingBacteriaBehavior(),
		},
	})
	creditsBefore := p.Resources().Get().Credits
	choiceIndex := 1
	selectedAmount := 5
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Spending all microbes should succeed")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(cardID), "Card should have 0 microbes")
	testutil.AssertEqual(t, creditsBefore+15, p.Resources().Get().Credits, "Should gain 15 credits (5 * 3)")
}
func TestSulphurEatingBacteria_Choice1_FailsWhenInsufficientMicrobes(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-sulphur-eating-bacteria"
	p.PlayedCards().AddCard(cardID, "Sulphur-Eating Bacteria", "active", []string{"microbe", "venus"})
	p.Resources().AddToStorage(cardID, 2)
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Sulphur-Eating Bacteria",
			BehaviorIndex: 0,
			Behavior:      sulphurEatingBacteriaBehavior(),
		},
	})
	choiceIndex := 1
	selectedAmount := 5
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, []string{cardID}, nil, nil, &selectedAmount, nil)
	testutil.AssertError(t, err, "Should fail when trying to spend more microbes than available")
	testutil.AssertEqual(t, 2, p.Resources().GetCardStorage(cardID), "Microbes should be unchanged")
}

// =============================================================================
// Venus Global Parameter Tests
// =============================================================================
func TestVenusRequirement_BlocksWhenTooLow(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := gamecards.Card{
		ID:   "card-venus-req-low-test",
		Name: "Venus Req Low Test",
		Type: gamecards.CardTypeAutomated,
		Pack: "venus-next",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagVenus},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementVenus, Min: testutil.IntPtr(8)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("card-venus-req-low-test")
	testGame.GlobalParameters().SetVenus(ctx, 6)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-venus-req-low-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail when Venus is too low")
}
func TestVenusRequirement_BlocksWhenTooHigh(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := gamecards.Card{
		ID:   "card-venus-req-high-test",
		Name: "Venus Req High Test",
		Type: gamecards.CardTypeAutomated,
		Pack: "venus-next",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagVenus},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementVenus, Max: testutil.IntPtr(14)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("card-venus-req-high-test")
	testGame.GlobalParameters().SetVenus(ctx, 16)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-venus-req-high-test", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail when Venus is too high")
}
func TestVenusIncrease_RaisesGlobalParameter(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := gamecards.Card{
		ID:   "card-venus-raise-test",
		Name: "Venus Raise Test",
		Type: gamecards.CardTypeAutomated,
		Pack: "venus-next",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagVenus},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("card-venus-raise-test")
	venusBefore := testGame.GlobalParameters().Venus()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-venus-raise-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing card with Venus output should succeed")
	venusAfter := testGame.GlobalParameters().Venus()
	testutil.AssertEqual(t, venusBefore+2, venusAfter, "Venus should increase by 2 (1 step = 2%)")
}
func TestVenusIncrease_CappedAtMax(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	testGame.GlobalParameters().SetVenus(ctx, 28)
	actualSteps, err := testGame.GlobalParameters().IncreaseVenus(ctx, 2)
	testutil.AssertNoError(t, err, "IncreaseVenus should not error")
	testutil.AssertEqual(t, 1, actualSteps, "Should only increase 1 step (capped at 30)")
	testutil.AssertEqual(t, 30, testGame.GlobalParameters().Venus(), "Venus should be capped at 30")
}
func TestVenusStateCalculator_RequirementValidation(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	card := gamecards.Card{
		ID:   "card-venus-state-calc-test",
		Name: "Venus State Calc Test",
		Type: gamecards.CardTypeAutomated,
		Pack: "venus-next",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagVenus},
		Requirements: &gamecards.CardRequirements{
			Items: []gamecards.Requirement{
				{Type: gamecards.RequirementVenus, Min: testutil.IntPtr(10)},
			},
		},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTitaniumProduction, Amount: 1, Target: "self-player"},
				},
			},
		},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards([]gamecards.Card{card})
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("card-venus-state-calc-test")
	g, _ := repo.Get(ctx, testGame.ID())
	testGame.GlobalParameters().SetVenus(ctx, 8)
	state := action.CalculatePlayerCardState(&card, p, g, cardRegistry)
	hasVenusError := false
	for _, e := range state.Errors {
		if e.Code == player.ErrorCodeVenusTooLow {
			hasVenusError = true
		}
	}
	testutil.AssertTrue(t, hasVenusError, "State calculator should report venus-too-low when Venus is 8 and requirement is 10")
	testGame.GlobalParameters().SetVenus(ctx, 10)
	state = action.CalculatePlayerCardState(&card, p, g, cardRegistry)
	hasVenusError = false
	for _, e := range state.Errors {
		if e.Code == player.ErrorCodeVenusTooLow {
			hasVenusError = true
		}
	}
	testutil.AssertTrue(t, !hasVenusError, "State calculator should NOT report venus-too-low when Venus meets requirement")
}
