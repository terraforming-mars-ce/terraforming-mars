package card_packs_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action/admin"
	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/action/confirmation"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestAphrodite_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Aphrodite"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 47, resources.Credits, "Aphrodite should have 47 starting credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Plants, "Aphrodite should start with 1 plant production")
}

func TestAphrodite_Gain2MCWhenVenusRaised(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Aphrodite"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 47, resources.Credits, "Aphrodite should have 47 credits before Venus increase")

	testGame.GlobalParameters().IncreaseVenus(ctx, 1)
	time.Sleep(50 * time.Millisecond)

	resources = p.Resources().Get()
	testutil.AssertEqual(t, 49, resources.Credits, "Aphrodite should have 49 credits after Venus increase (gained 2 M€)")
}

func TestCelestic_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Celestic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 42, resources.Credits, "Celestic should start with 42 credits")
}

func TestCelestic_FloaterStorageRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Celestic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	corpCardID := testutil.CardID("Celestic")
	storageMap := p.Resources().Storage()
	_, hasStorage := storageMap[corpCardID]
	testutil.AssertTrue(t, hasStorage, "Celestic should have floater storage initialized on the corp card")
}

func TestCelestic_FirstActionDrawsFloaterCards(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Set up a controlled deck: 3 non-floater cards followed by 2 floater cards
	// The draw-until-matching logic should skip non-floater cards and find the floater ones
	floaterCardIDs := []string{"213", "222"} // Aerial Mappers (floater storage), Dirigibles (floater storage+output)
	nonFloaterCardIDs := []string{"001", "002", "003"}
	allProjectCards := append(nonFloaterCardIDs, floaterCardIDs...)
	customDeck := deck.NewDeck(testGame.ID(), allProjectCards, nil, nil)
	testGame.SetDeck(customDeck)

	// Clear hand before setting corporation
	p, _ := testGame.GetPlayer(playerID)
	for _, c := range p.Hand().Cards() {
		p.Hand().RemoveCard(c)
	}

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Celestic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	handCards := p.Hand().Cards()
	testutil.AssertEqual(t, 2, len(handCards),
		"Celestic first action should draw exactly 2 cards")

	// Verify both drawn cards are floater cards
	for _, cID := range handCards {
		card, err := cardRegistry.GetByID(cID)
		testutil.AssertNoError(t, err, "Card should exist in registry")
		hasFloater := false
		if card.ResourceStorage != nil && card.ResourceStorage.Type == shared.ResourceFloater {
			hasFloater = true
		}
		for _, b := range card.Behaviors {
			for _, o := range b.Outputs {
				if o.ResourceType == shared.ResourceFloater {
					hasFloater = true
				}
			}
			for _, i := range b.Inputs {
				if i.ResourceType == shared.ResourceFloater {
					hasFloater = true
				}
			}
		}
		testutil.AssertTrue(t, hasFloater,
			"Card "+card.Name+" ("+cID+") should have floater resource")
	}

	// Verify non-floater cards were NOT drawn into hand
	for _, nonFloaterID := range nonFloaterCardIDs {
		found := false
		for _, handCard := range handCards {
			if handCard == nonFloaterID {
				found = true
			}
		}
		testutil.AssertTrue(t, !found,
			"Non-floater card "+nonFloaterID+" should not be in hand")
	}
}

func TestCelestic_AddFloaterToAnyCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Celestic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	corpCardID := testutil.CardID("Celestic")
	p.PlayedCards().AddCard(corpCardID, "Celestic", "corporation", []string{"venus"})
	p.Resources().AddToStorage(corpCardID, 0)

	otherCardID := testutil.CardID("Aerial Mappers")
	p.PlayedCards().AddCard(otherCardID, "Aerial Mappers", "active", []string{"venus", "science"})
	p.Resources().AddToStorage(otherCardID, 0)

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), playerID, corpCardID, 2, nil, []string{otherCardID}, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Celestic add floater to another card should succeed")

	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(corpCardID), "Celestic corp card should still have 0 floaters")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(otherCardID), "Target card should have 1 floater after action")
}

func TestManutech_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Manutech"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 35, resources.Credits, "Manutech should start with 35 credits")
	testutil.AssertEqual(t, 1, resources.Steel, "Manutech should start with 1 steel from production-increased trigger")
	testutil.AssertEqual(t, 0, resources.Titanium, "Manutech should start with 0 titanium")
	testutil.AssertEqual(t, 0, resources.Plants, "Manutech should start with 0 plants")
	testutil.AssertEqual(t, 0, resources.Energy, "Manutech should start with 0 energy")
	testutil.AssertEqual(t, 0, resources.Heat, "Manutech should start with 0 heat")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Steel, "Manutech should start with 1 steel production")
}

func TestManutech_GainResourceWhenProductionIncreased(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Manutech"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	steelCount := p.Resources().Get().Steel
	testutil.AssertEqual(t, 1, steelCount, "Manutech should start with 1 steel from production-increased trigger")

	energyBefore := p.Resources().Get().Energy

	deepWellHeatingID := testutil.CardID("Deep Well Heating")
	p.Hand().AddCard(deepWellHeatingID)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 13}
	err = playCard.Execute(ctx, testGame.ID(), playerID, deepWellHeatingID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Deep Well Heating should play successfully")

	time.Sleep(50 * time.Millisecond)

	energyAfter := p.Resources().Get().Energy
	testutil.AssertTrue(t, energyAfter >= energyBefore+1, "Manutech should gain energy when energy production is increased")
}

func TestMorningStarInc_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Morning Star Inc."))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 50, resources.Credits, "Morning Star Inc. should start with 50 credits")
}

func TestMorningStarInc_VenusLenienceRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Morning Star Inc."))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	effects := p.Effects().List()
	found := false
	for _, effect := range effects {
		if effect.CardName == "Morning Star Inc." && effect.BehaviorIndex == 2 {
			found = true
			testutil.AssertEqual(t, shared.ResourceGlobalParameterLenience, effect.Behavior.Outputs[0].ResourceType,
				"Morning Star Inc. effect output should be global parameter lenience")
			testutil.AssertEqual(t, 2, effect.Behavior.Outputs[0].Amount,
				"Morning Star Inc. venus lenience amount should be 2")
			break
		}
	}
	testutil.AssertTrue(t, found, "Morning Star Inc. should have registered its venus lenience effect at behavior index 2")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	venusLenience := calculator.CalculateGlobalParameterLenience(p, "venus")
	testutil.AssertEqual(t, 2, venusLenience, "Morning Star Inc. should provide venus lenience of 2")
	tempLenience := calculator.CalculateGlobalParameterLenience(p, "temperature")
	testutil.AssertEqual(t, 0, tempLenience, "Morning Star Inc. should NOT provide temperature lenience")
}

func TestViron_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 48, resources.Credits, "Viron should start with 48 credits")
}

func TestViron_HasActionReuseAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	vironCardID := testutil.CardID("Viron")
	actions := p.Actions().List()
	hasVironAction := false
	for _, a := range actions {
		if a.CardID == vironCardID {
			for _, output := range a.Behavior.Outputs {
				if output.ResourceType == shared.ResourceActionReuse {
					hasVironAction = true
					break
				}
			}
		}
	}
	testutil.AssertTrue(t, hasVironAction, "Viron should have a manual action-reuse action")
}

func TestViron_ReuseBlueCardAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	vironCardID := testutil.CardID("Viron")

	p.PlayedCards().AddCard("test-blue-card", "Test Blue Card", "active", []string{"microbe"})
	p.Resources().AddToStorage("test-blue-card", 0)

	testBlueAction := player.CardAction{
		CardID:        "test-blue-card",
		CardName:      "Test Blue Card",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
			},
		},
		TimesUsedThisGeneration: 1,
	}

	existingActions := p.Actions().List()
	existingActions = append(existingActions, testBlueAction)
	p.Actions().SetActions(existingActions)

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)

	err = useAction.Execute(ctx, testGame.ID(), playerID, "test-blue-card", 0, nil, []string{"test-blue-card"}, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Blue card action should fail because already used this generation")

	reuseSource := vironCardID
	err = useAction.Execute(ctx, testGame.ID(), playerID, "test-blue-card", 0, nil, []string{"test-blue-card"}, nil, nil, nil, nil, &reuseSource)
	testutil.AssertNoError(t, err, "Viron should allow reusing already-used blue card action")

	storage := p.Resources().GetCardStorage("test-blue-card")
	testutil.AssertEqual(t, 1, storage, "Test blue card should have gained 1 microbe from reuse")

	actions := p.Actions().List()
	for _, a := range actions {
		if a.CardID == vironCardID {
			for _, output := range a.Behavior.Outputs {
				if output.ResourceType == shared.ResourceActionReuse {
					testutil.AssertEqual(t, 1, a.TimesUsedThisGeneration, "Viron action should be marked as used")
				}
			}
		}
	}
}

func TestViron_CannotReuseUnusedAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	vironCardID := testutil.CardID("Viron")

	p.PlayedCards().AddCard("test-blue-card", "Test Blue Card", "active", []string{"microbe"})
	p.Resources().AddToStorage("test-blue-card", 0)

	testBlueAction := player.CardAction{
		CardID:        "test-blue-card",
		CardName:      "Test Blue Card",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
			},
		},
		TimesUsedThisGeneration: 0,
	}

	existingActions := p.Actions().List()
	existingActions = append(existingActions, testBlueAction)
	p.Actions().SetActions(existingActions)

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	reuseSource := vironCardID
	err = useAction.Execute(ctx, testGame.ID(), playerID, "test-blue-card", 0, nil, []string{"test-blue-card"}, nil, nil, nil, nil, &reuseSource)
	testutil.AssertError(t, err, "Should not be able to reuse an action that has not been used this generation")
}

func TestViron_CannotReuseSelf(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	vironCardID := testutil.CardID("Viron")

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	reuseSource := vironCardID
	err = useAction.Execute(ctx, testGame.ID(), playerID, vironCardID, 1, nil, nil, nil, nil, nil, nil, &reuseSource)
	testutil.AssertError(t, err, "Should not be able to reuse own action-reuse ability")
}

func TestViron_CannotReuseAfterAlreadyUsedThisGen(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	vironCardID := testutil.CardID("Viron")

	p.PlayedCards().AddCard("test-blue-card", "Test Blue Card", "active", []string{"microbe"})
	p.Resources().AddToStorage("test-blue-card", 0)

	testBlueAction := player.CardAction{
		CardID:        "test-blue-card",
		CardName:      "Test Blue Card",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "self-card"},
			},
		},
		TimesUsedThisGeneration: 1,
	}

	existingActions := p.Actions().List()
	existingActions = append(existingActions, testBlueAction)
	p.Actions().SetActions(existingActions)

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	reuseSource := vironCardID

	err = useAction.Execute(ctx, testGame.ID(), playerID, "test-blue-card", 0, nil, []string{"test-blue-card"}, nil, nil, nil, nil, &reuseSource)
	testutil.AssertNoError(t, err, "First Viron reuse should succeed")

	p.Resources().AddToStorage("test-blue-card", 0)

	testBlueAction2 := player.CardAction{
		CardID:        "test-blue-card-2",
		CardName:      "Test Blue Card 2",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCredit, Amount: 2, Target: "self-player"},
			},
		},
		TimesUsedThisGeneration: 1,
	}
	currentActions := p.Actions().List()
	currentActions = append(currentActions, testBlueAction2)
	p.Actions().SetActions(currentActions)

	err = useAction.Execute(ctx, testGame.ID(), playerID, "test-blue-card-2", 0, nil, nil, nil, nil, nil, nil, &reuseSource)
	testutil.AssertError(t, err, "Viron should not be able to reuse again this generation")
}

func TestValleyTrust_FirstActionCreatesPreludeSelection(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Set up controlled prelude deck with exactly 5 preludes
	preludeIDs := []string{"P01", "P02", "P03", "P04", "P05"}
	customDeck := deck.NewDeck(testGame.ID(), nil, nil, preludeIDs)
	testGame.SetDeck(customDeck)

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Valley Trust"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	selection := p.Selection().GetPendingCardDrawSelection()
	testutil.AssertTrue(t, selection != nil, "Valley Trust should create a pending card draw selection")
	testutil.AssertEqual(t, 3, len(selection.AvailableCards), "Should have 3 prelude cards available")
	testutil.AssertEqual(t, 1, selection.FreeTakeCount, "Should be able to take 1 card for free")
	testutil.AssertEqual(t, 0, selection.MaxBuyCount, "Should not be able to buy cards")
	testutil.AssertTrue(t, selection.PlayAsPrelude, "Selection should be marked as prelude")

	// Verify all available cards are prelude cards
	for _, cardID := range selection.AvailableCards {
		card, err := cardRegistry.GetByID(cardID)
		testutil.AssertNoError(t, err, "Card should exist in registry")
		testutil.AssertEqual(t, string(gamecards.CardTypePrelude), string(card.Type),
			"Available card "+card.Name+" should be a prelude card")
	}

	// Verify forced first action was set
	forcedAction := testGame.GetForcedFirstAction(playerID)
	testutil.AssertTrue(t, forcedAction != nil, "Should create forced first action")
	testutil.AssertEqual(t, "card-draw-selection", forcedAction.ActionType, "Action type should be card-draw-selection")
}

func TestValleyTrust_ConfirmPreludeSelection(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Set up controlled prelude deck
	preludeIDs := []string{"P01", "P02", "P03", "P04", "P05"}
	customDeck := deck.NewDeck(testGame.ID(), nil, nil, preludeIDs)
	testGame.SetDeck(customDeck)

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Valley Trust"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	selection := p.Selection().GetPendingCardDrawSelection()
	testutil.AssertTrue(t, selection != nil, "Should have pending selection")

	// Record credits before confirming (Valley Trust starts with 37)
	creditsBefore := p.Resources().Get().Credits

	// Take the first available prelude card (Allied Bank - P01 gives 3 credits + 4 credit production)
	selectedPreludeID := selection.AvailableCards[0]

	confirmAction := confirmation.NewConfirmCardDrawAction(repo, cardRegistry, logger)
	err = confirmAction.Execute(ctx, testGame.ID(), playerID, []string{selectedPreludeID}, nil)
	testutil.AssertNoError(t, err, "Confirm card draw should succeed")

	// Verify prelude was played (added to played cards)
	playedCards := p.PlayedCards().Cards()
	found := false
	for _, pc := range playedCards {
		if pc == selectedPreludeID {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "Selected prelude should be in played cards")

	// Verify prelude effects were applied (Allied Bank gives 3 credits)
	if selectedPreludeID == "P01" {
		creditsAfter := p.Resources().Get().Credits
		testutil.AssertEqual(t, creditsBefore+3, creditsAfter,
			"Allied Bank prelude should have given 3 credits")

		production := p.Resources().Production()
		testutil.AssertEqual(t, 4, production.Credits,
			"Allied Bank prelude should have given 4 credit production")
	}

	// Verify selection was cleared
	testutil.AssertTrue(t, p.Selection().GetPendingCardDrawSelection() == nil,
		"Pending card draw selection should be cleared")

	// Verify forced first action was cleared
	testutil.AssertTrue(t, testGame.GetForcedFirstAction(playerID) == nil,
		"Forced first action should be cleared")

	// Verify unselected preludes were removed permanently (not discarded)
	removedCards := testGame.Deck().RemovedCards()
	testutil.AssertEqual(t, 2, len(removedCards), "2 unselected preludes should be permanently removed")
}
