package card_packs_test

import (
	"context"
	"fmt"
	"time"

	cardAction "terraforming-mars-backend/internal/action/card"
	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
	"testing"
)

// =============================================================================
// Card 006: Inventors' Guild (active, cost 9, tags: [science])
// "Action: look at the top card and either buy it or discard it"
// Behavior: manual trigger, outputs: card-buy 1 + card-peek 1 to self-player
// =============================================================================
func TestInventorsGuild_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Inventors' Guild")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Inventors' Guild should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Inventors' Guild should be in played cards")
}
func TestInventorsGuild_ActionCanBeUsed(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Inventors' Guild")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 4), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	// Play the card first (registers the manual action)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Inventors' Guild should play successfully")
	// Use the card action (behavior index 0 since only one behavior, manual)
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, 0, nil, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Inventors' Guild action should succeed")
}

// =============================================================================
// Card 014: Development Center (active, cost 11, tags: [building, science])
// "Action: Spend 1 energy to draw a card."
// Behavior: manual trigger, inputs: energy 1, outputs: card-draw 1
// =============================================================================
func TestDevelopmentCenter_PlaysAndActionSpendEnergyDrawCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Development Center")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 4), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourceEnergy: 3,
	})
	p.Hand().AddCard(card.ID)
	energyBefore := p.Resources().Get().Energy
	// Play the card first
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Development Center should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Development Center should be in played cards")
	// Use the card action: spend 1 energy to draw a card
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, 0, nil, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Development Center action should succeed")
	energyAfter := p.Resources().Get().Energy
	testutil.AssertEqual(t, energyBefore-1, energyAfter,
		"Should have 1 less energy after using Development Center action")
}
func TestDevelopmentCenter_ActionFailsWithoutEnergy(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Development Center")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 4), "SetCurrentTurn failed")
	// Give credits to play the card but NO energy
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	// Play the card first
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Development Center should play successfully")
	// Try to use the action with 0 energy - should fail
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, 0, nil, nil, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Development Center action should fail without energy")
}

// --- Space Station (025) ---
// active, cost 10, tags: [space]
// "Effect: When you play a space card, you pay 2 M€ less for it."
// Auto trigger, outputs: discount 2 to self-player with selector tags:[space].
func TestSpaceStation_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Space Station")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	creditsBefore := p.Resources().Get().Credits
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Space Station should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Space Station should be in played cards")
	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-10, creditsAfter,
		"Should have paid 10 credits for Space Station")
}

// --- Virus (050) ---
// event, cost 1, tags: [microbe]
// "Remove up to 2 animals or 5 plants from any player."
// Auto trigger with choices:
//
//	choice 0 = animal removal (2) from any-card
//	choice 1 = plant removal (5) from any-player
func TestVirus_Choice1_RemovePlantsFromOpponent(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Virus")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID(testutil.CardID("Tharsis Republic"))
	target.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, attacker.ID(), 2), "SetCurrentTurn failed")
	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 8,
	})
	attacker.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetID := target.ID()
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), card.ID, payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Virus should play successfully with choice 1 (remove plants)")
	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 3, targetResources.Plants, "Target should have 3 plants (8 - 5)")
}
func TestVirus_Choice1_PartialPlantRemoval(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Virus")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID(testutil.CardID("Tharsis Republic"))
	target.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, attacker.ID(), 2), "SetCurrentTurn failed")
	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 2,
	})
	attacker.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetID := target.ID()
	choiceIndex := 1
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), card.ID, payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Virus should play successfully with partial plant removal")
	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 0, targetResources.Plants, "Target should have 0 plants (had 2, Virus removes up to 5)")
}
func TestVirus_Choice0_RemoveAnimalsFromCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Virus")
	animalHost := gamecards.Card{
		ID:              "card-animal-host-virus",
		Name:            "Animal Host",
		Type:            gamecards.CardTypeActive,
		Cost:            0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}
	additionalCards := []gamecards.Card{animalHost}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.PlayedCards().AddCard("card-animal-host-virus", "Animal Host", "active", []string{"animal"})
	p.Resources().AddToStorage("card-animal-host-virus", 5)
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-animal-host-virus"
	choiceIndex := 0
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, &choiceIndex, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "Virus should play successfully with choice 0 (remove animals)")
	animalStorage := p.Resources().GetCardStorage("card-animal-host-virus")
	testutil.AssertEqual(t, 3, animalStorage, "Animal host should have 3 animals (5 - 2)")
}

// --- Electro Catapult (069) ---
// Active card, cost 17, tags: [building]. Requirements: oxygen max 8.
// Behavior 0 (auto): outputs energy-production -1 to self-player
// Behavior 1 (manual): choices: [spend 1 plant, spend 1 steel] -> outputs credit 7 to self-player
func TestElectroCatapult_PlayDecreasesEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Electro Catapult")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 17}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Electro Catapult should play successfully at 0% oxygen")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1 after playing Electro Catapult")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Electro Catapult should be in played cards")
}
func TestElectroCatapult_ActionSpendPlantGainCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-electro-catapult"
	p.PlayedCards().AddCard(cardID, "Electro Catapult", "active", []string{"building"})
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourcePlant: 5,
	})
	creditsBefore := p.Resources().Get().Credits
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"}},
		},
		Choices: []shared.Choice{
			{
				Inputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"}},
				},
			},
			{
				Inputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"}},
				},
			},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Electro Catapult",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})
	choiceIndex := 0 // spend 1 plant
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &choiceIndex, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Electro Catapult action (spend plant) should succeed")
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 4, resources.Plants, "Should have 4 plants after spending 1")
	testutil.AssertEqual(t, creditsBefore+7, resources.Credits, "Should gain 7 credits")
}
func TestElectroCatapult_ActionSpendSteelGainCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-electro-catapult-steel"
	p.PlayedCards().AddCard(cardID, "Electro Catapult", "active", []string{"building"})
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 3,
	})
	creditsBefore := p.Resources().Get().Credits
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 7, Target: "self-player"}},
		},
		Choices: []shared.Choice{
			{
				Inputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePlant, Amount: 1, Target: "self-player"}},
				},
			},
			{
				Inputs: []shared.BehaviorCondition{
					&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceSteel, Amount: 1, Target: "self-player"}},
				},
			},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Electro Catapult",
			BehaviorIndex: 1,
			Behavior:      behavior,
		},
	})
	choiceIndex := 1 // spend 1 steel
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, &choiceIndex, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Electro Catapult action (spend steel) should succeed")
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Steel, "Should have 2 steel after spending 1")
	testutil.AssertEqual(t, creditsBefore+7, resources.Credits, "Should gain 7 credits")
}

// --- Earth Catapult (070) ---
// Active card, cost 23, tags: [earth].
// "When you play a card, you pay 2 M€ less."
// Auto trigger, outputs: discount 2 to self-player (no selectors = applies to all cards).
func TestEarthCatapult_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Earth Catapult")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Earth Catapult should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Earth Catapult should be in played cards")
}

// =============================================================================
// Card 071: Advanced Alloys (active, cost 9, tags: [science])
// "Each steel/titanium worth 1 M€ extra."
// Auto trigger, outputs: 2x value-modifier (1 for steel, 1 for titanium) to
// self-player with selectors.
// =============================================================================
func TestAdvancedAlloys_PlaysSuccessfully_ValueModifiersApplied(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Advanced Alloys")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	steelModBefore := p.Resources().GetValueModifier(shared.ResourceSteel)
	titaniumModBefore := p.Resources().GetValueModifier(shared.ResourceTitanium)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Advanced Alloys should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Advanced Alloys should be in played cards")
	steelModAfter := p.Resources().GetValueModifier(shared.ResourceSteel)
	titaniumModAfter := p.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, steelModBefore+1, steelModAfter,
		"Steel value modifier should increase by 1")
	testutil.AssertEqual(t, titaniumModBefore+1, titaniumModAfter,
		"Titanium value modifier should increase by 1")
}

// =============================================================================
// Card 064: Mining Area (automated, cost 4, tags: [building])
// "Place mining tile on steel/titanium bonus area adjacent to your tile.
//
//	Increase production of that resource."
//
// Auto trigger with choices AND tile-placement output.
// Mining Area no longer uses choices — production is determined by which bonus tile is placed on.
// The tile-placed trigger handles this automatically (not yet implemented).
// These tests verify the card plays without a choice index and queues a tile placement.
// =============================================================================
func TestMiningArea_PlaceOnSteelBonus_IncreaseSteelProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Mining Area")
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	p := testGame.GetAllPlayers()[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")

	// Place a player-owned tile adjacent to the steel bonus tile at (-4,3,1)
	adjacentToSteelBonus := shared.HexPosition{Q: -3, R: 3, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, adjacentToSteelBonus,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "Should place player tile adjacent to steel bonus")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: card.Cost}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mining Area should play without requiring a choice index")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending mining tile selection")

	// Select the steel bonus hex at (-4,3,1) — land tile with Steel×2
	steelBonusHex := fmt.Sprintf("%d,%d,%d", -4, 3, 1)
	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTileAction.Execute(ctx, testGame.ID(), p.ID(), steelBonusHex)
	testutil.AssertNoError(t, err, "Should be able to select steel bonus hex")

	time.Sleep(50 * time.Millisecond)

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Steel+1, prodAfter.Steel,
		"Steel production should increase by 1 when mining tile placed on steel bonus")
	testutil.AssertEqual(t, prodBefore.Titanium, prodAfter.Titanium,
		"Titanium production should not change")
}

func TestMiningArea_PlaceOnTitaniumBonus_IncreaseTitaniumProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Mining Area")
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	p := testGame.GetAllPlayers()[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")

	// Place a player-owned tile adjacent to the titanium bonus tile at (1,3,-4)
	adjacentToTitaniumBonus := shared.HexPosition{Q: 0, R: 3, S: -3}
	err := testGame.Board().UpdateTileOccupancy(ctx, adjacentToTitaniumBonus,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "Should place player tile adjacent to titanium bonus")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)

	prodBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: card.Cost}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mining Area should play without requiring a choice index")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending mining tile selection")

	// Select the titanium bonus hex at (1,3,-4) — land tile with Titanium×1
	titaniumBonusHex := fmt.Sprintf("%d,%d,%d", 1, 3, -4)
	selectTileAction := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTileAction.Execute(ctx, testGame.ID(), p.ID(), titaniumBonusHex)
	testutil.AssertNoError(t, err, "Should be able to select titanium bonus hex")

	time.Sleep(50 * time.Millisecond)

	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Titanium+1, prodAfter.Titanium,
		"Titanium production should increase by 1 when mining tile placed on titanium bonus")
	testutil.AssertEqual(t, prodBefore.Steel, prodAfter.Steel,
		"Steel production should not change")
}

// --- Mars University (073) ---
// "When you play a science tag, including this, you may discard a card from hand to draw a card."
// Passive triggered effect: auto trigger with condition type:"tag-played" for science tags.
// Optional card-discard input, card-draw output. Just test that the card plays successfully.
func TestMarsUniversity_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Mars University")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mars University should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Mars University should be in played cards")
}

// --- Viral Enhancers (074) ---
// "When you play a plant, microbe, or an animal tag, including this, gain 1 plant or add 1 resource to that card."
// Passive triggered effect with choices. Just test that the card plays successfully.
func TestViralEnhancers_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Viral Enhancers")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Viral Enhancers should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Viral Enhancers should be in played cards")
}

// --- Robotic Workforce (086) ---
// "Duplicate only the production box of one of your building cards."
// Has NO behaviors in JSON (null). Just test it plays successfully.
func TestRoboticWorkforce_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Robotic Workforce")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Robotic Workforce should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Robotic Workforce should be in played cards")
}

// --- Earth Office (105) ---
// Active card, cost 1, tags: [earth].
// "When you play an Earth tag, you pay 3 M€ less."
// Auto trigger, outputs: discount 3 to self-player with selector tags:[earth].
func TestEarthOffice_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Earth Office")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Earth Office should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Earth Office should be in played cards")
}

// --- Business Contacts (111) ---
// Event card, cost 7, tags: [earth].
// "Look at top 4 cards, take 2, discard 2."
// Auto trigger, outputs: card-take 2 + card-peek 4 to self-player.
func TestBusinessContacts_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Business Contacts")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 7}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Business Contacts should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Business Contacts should be in played cards")
}

// --- Sabotage (121) ---
// Event card, cost 1, tags: none.
// "Remove up to 3 titanium from any player, or 4 steel, or 7 M€."
// Auto trigger with 3 choices targeting any-player.
func TestSabotage_RemoveTitaniumFromOpponent(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Sabotage")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID(testutil.CardID("Tharsis Republic"))
	target.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, attacker.ID(), 2), "SetCurrentTurn failed")
	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard(card.ID)
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 5,
	})
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	choiceIndex := 0
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), card.ID, payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Sabotage should play successfully with choice 0 (remove titanium)")
	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Titanium, "Target should have 2 titanium after 3 removed (5 - 3)")
}
func TestSabotage_RemoveSteelFromOpponent(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Sabotage")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	attacker := players[0]
	target := players[1]
	attacker.SetCorporationID(testutil.CardID("Tharsis Republic"))
	target.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, attacker.ID(), 2), "SetCurrentTurn failed")
	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard(card.ID)
	target.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 6,
	})
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	choiceIndex := 1
	targetID := target.ID()
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), card.ID, payment, &choiceIndex, nil, &targetID, nil)
	testutil.AssertNoError(t, err, "Sabotage should play successfully with choice 1 (remove steel)")
	targetResources := target.Resources().Get()
	testutil.AssertEqual(t, 2, targetResources.Steel, "Target should have 2 steel after 4 removed (6 - 4)")
}

// --- CEO's Favorite Project (149) ---
// "Add 1 resource to a card with at least 1 resource on it."
// Event, cost 1, no tags. Auto trigger, outputs: card-resource 1 to any-card.
func TestCEOsFavoriteProject_AddsMicrobeToTargetCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	ceosFavorite := testutil.GetCardByName("CEO's Favorite Project")
	targetCard := gamecards.Card{
		ID:              "card-target-microbes",
		Name:            "Target Microbe Card",
		Type:            gamecards.CardTypeActive,
		Cost:            0,
		Tags:            []shared.CardTag{shared.TagMicrobe},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, targetCard})
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	// Set up a played card with microbe storage containing 2 microbes
	p.PlayedCards().AddCard("card-target-microbes", "Target Microbe Card", "active", []string{"microbe"})
	p.Resources().AddToStorage("card-target-microbes", 2)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(ceosFavorite.ID)
	storageBefore := p.Resources().GetCardStorage("card-target-microbes")
	testutil.AssertEqual(t, 2, storageBefore, "Target card should start with 2 microbes")
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	targetCardID := "card-target-microbes"
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), ceosFavorite.ID, payment, nil, []string{targetCardID}, nil, nil)
	testutil.AssertNoError(t, err, "CEO's Favorite Project should play successfully")
	storageAfter := p.Resources().GetCardStorage("card-target-microbes")
	testutil.AssertEqual(t, 3, storageAfter, "Target card should have 3 microbes (2 existing + 1 added)")
	testutil.AssertTrue(t, p.PlayedCards().Contains(ceosFavorite.ID),
		"CEO's Favorite Project should be in played cards")
	testutil.AssertEqual(t, false, p.Hand().HasCard(ceosFavorite.ID),
		"CEO's Favorite Project should be removed from hand")
}
func TestCEOsFavoriteProject_FailsWithoutTargetCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	ceosFavorite := testutil.GetCardByName("CEO's Favorite Project")
	targetCard := gamecards.Card{
		ID:              "card-target-animals",
		Name:            "Target Animal Card",
		Type:            gamecards.CardTypeActive,
		Cost:            0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ceosFavorite, targetCard})
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.PlayedCards().AddCard("card-target-animals", "Target Animal Card", "active", []string{"animal"})
	p.Resources().AddToStorage("card-target-animals", 1)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(ceosFavorite.ID)
	// Play without specifying a target card - should fail
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), ceosFavorite.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should fail without target card for card-resource output")
}

// --- Protected Habitats (173) ---
// "Opponents may not remove your plants/animals/microbes."
// Active, cost 5, no tags. Auto trigger, outputs: defense 1 to self-card with selectors for plant, microbe, animal.
func TestProtectedHabitats_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	protectedHabitats := testutil.GetCardByName("Protected Habitats")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(protectedHabitats.ID)
	creditsBefore := p.Resources().Get().Credits
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), protectedHabitats.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Protected Habitats should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(protectedHabitats.ID),
		"Protected Habitats should be in played cards")
	testutil.AssertEqual(t, false, p.Hand().HasCard(protectedHabitats.ID),
		"Protected Habitats should be removed from hand")
	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-5, creditsAfter,
		"Should have paid 5 credits for Protected Habitats")
}

// --- Corporate Stronghold (182) ---
// "Decrease your energy production 1 step and increase your M€ production 3 steps. Place a city tile."
func TestCorporateStronghold_ProductionAndCityPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Corporate Stronghold")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard(card.ID)
	prodBefore := p.Resources().Production()
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Corporate Stronghold should play successfully")
	prodAfter := p.Resources().Production()
	testutil.AssertEqual(t, prodBefore.Credits+3, prodAfter.Credits,
		"Credit production should increase by 3")
	testutil.AssertEqual(t, prodBefore.Energy-1, prodAfter.Energy,
		"Energy production should decrease by 1")
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")
}
func TestCorporateStronghold_FailsWithoutEnergyProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Corporate Stronghold")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 11}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Corporate Stronghold should fail without energy production")
}

// --- Olympus Conference (185) ---
// "When you play a science tag, including this, either add a science resource to this card,
// or remove a science resource from this card to draw a card."
// Passive triggered effect with choices and resource storage. Just test that the card plays successfully.
func TestOlympusConference_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Olympus Conference")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Olympus Conference should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Olympus Conference should be in played cards")
}

// --- Invention Contest (192) ---
// "Look at top 3 cards, take 1, discard 2."
// Auto trigger, outputs: card-take 1 + card-peek 3.
func TestInventionContest_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Invention Contest")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 2}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Invention Contest should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Invention Contest should be in played cards")
}

// --- Power Infrastructure (194) ---
// "Action: Spend any number of energy to gain that amount of M€."
// Manual trigger, variableAmount inputs (energy) and outputs (credit).
func TestPowerInfrastructure_PlayAndUseAction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Power Infrastructure")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
		shared.ResourceEnergy: 10,
	})
	p.Hand().AddCard(card.ID)
	// Play the card first
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Power Infrastructure should be in played cards")
	creditsBefore := p.Resources().Get().Credits
	energyBefore := p.Resources().Get().Energy
	// Give player another action since playing the card consumed one
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")
	// Use the action: spend 3 energy to gain 3 credits
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 3
	err = useAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, 0, nil, nil, nil, nil, &selectedAmount, nil, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure action should succeed")
	resources := p.Resources().Get()
	testutil.AssertEqual(t, energyBefore-3, resources.Energy, "Energy should decrease by 3")
	testutil.AssertEqual(t, creditsBefore+3, resources.Credits, "Credits should increase by 3")
}
func TestPowerInfrastructure_UseActionSpendAllEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	p, _ := testGame.GetPlayer(playerID)
	cardID := "test-power-infra-ce9"
	p.PlayedCards().AddCard(cardID, "Power Infrastructure", "active", []string{"building", "power"})
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 7,
		shared.ResourceCredit: 5,
	})
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player"}, VariableAmount: true},
		},
		Outputs: []shared.BehaviorCondition{
			&shared.BasicResourceCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"}, VariableAmount: true},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        cardID,
			CardName:      "Power Infrastructure",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})
	// Spend all 7 energy
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 7
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, &selectedAmount, nil, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure should succeed spending all energy")
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Energy, "Should have 0 energy after spending all 7")
	testutil.AssertEqual(t, 12, resources.Credits, "Should have 12 credits (5 + 7)")
}

// --- Indentured Workers (195) ---
// "The next card you play this generation costs 8 M€ less."
// Auto trigger, outputs: discount 8 with temporary:"next-card".
func TestIndenturedWorkers_PlaysSuccessfully(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Indentured Workers")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "UpdateStatus failed")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "UpdatePhase failed")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "SetCurrentTurn failed")
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard(card.ID)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Indentured Workers should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID),
		"Indentured Workers should be in played cards")
}
