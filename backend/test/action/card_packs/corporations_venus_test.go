package card_packs_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action/admin"
	cardAction "terraforming-mars-backend/internal/action/card"
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
	testutil.AssertEqual(t, 49, resources.Credits, "Aphrodite should have 49 credits (47 starting + 2 from auto effect)")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Plants, "Aphrodite should start with 1 plant production")
}

func TestAphrodite_Gain2MCWhenVenusTerraformed(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Aphrodite"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 49, resources.Credits, "Aphrodite should have 49 credits before Venus increase")

	testGame.GlobalParameters().IncreaseVenus(ctx, 1)
	time.Sleep(50 * time.Millisecond)

	resources = p.Resources().Get()
	testutil.AssertEqual(t, 51, resources.Credits, "Aphrodite should have 51 credits after Venus increase (gained 2 M€)")
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

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Celestic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	handSize := p.Hand().CardCount()

	testutil.AssertTrue(t, handSize >= 2,
		"Celestic first action should draw 2 cards with floater icons into hand")

	handCards := p.Hand().Cards()
	floaterCardCount := 0
	for _, cID := range handCards {
		card, err := cardRegistry.GetByID(cID)
		if err != nil || card == nil {
			continue
		}
		if card.ResourceStorage != nil && card.ResourceStorage.Type == shared.ResourceFloater {
			floaterCardCount++
			continue
		}
		for _, b := range card.Behaviors {
			for _, o := range b.Outputs {
				if o.ResourceType == shared.ResourceFloater {
					floaterCardCount++
					break
				}
			}
		}
	}
	testutil.AssertEqual(t, 2, floaterCardCount,
		"Celestic first action should draw exactly 2 cards with floater icons")
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
	err = useAction.Execute(ctx, testGame.ID(), playerID, corpCardID, 2, nil, []string{otherCardID}, nil, nil, nil, nil)
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
	testutil.AssertEqual(t, 36, resources.Credits, "Manutech should start with 36 credits (35 + 1 from auto)")
	testutil.AssertEqual(t, 1, resources.Steel, "Manutech should start with 1 steel from auto effect")
	testutil.AssertEqual(t, 1, resources.Titanium, "Manutech should start with 1 titanium from auto effect")
	testutil.AssertEqual(t, 1, resources.Plants, "Manutech should start with 1 plant from auto effect")
	testutil.AssertEqual(t, 1, resources.Energy, "Manutech should start with 1 energy from auto effect")
	testutil.AssertEqual(t, 1, resources.Heat, "Manutech should start with 1 heat from auto effect")

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
	testutil.AssertEqual(t, 1, steelCount, "Manutech should start with 1 steel from auto effect")

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
			testutil.AssertEqual(t, shared.ResourceVenusLenience, effect.Behavior.Outputs[0].ResourceType,
				"Morning Star Inc. effect output should be venus lenience")
			testutil.AssertEqual(t, 2, effect.Behavior.Outputs[0].Amount,
				"Morning Star Inc. venus lenience amount should be 2")
			break
		}
	}
	testutil.AssertTrue(t, found, "Morning Star Inc. should have registered its venus lenience effect at behavior index 2")
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

func TestViron_ReuseBlueCardAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Viron"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	p.PlayedCards().AddCard("test-blue-card", "Test Blue Card", "active", []string{"microbe"})
	p.Resources().AddToStorage("test-blue-card", 0)

	testAction := player.CardAction{
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
	p.Actions().SetActions([]player.CardAction{testAction})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), playerID, "test-blue-card", 0, nil, []string{"test-blue-card"}, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Blue card action should fail because already used this generation")

	vironCardID := testutil.CardID("Viron")
	actions := p.Actions().List()
	hasVironAction := false
	for _, a := range actions {
		if a.CardID == vironCardID {
			hasVironAction = true
			break
		}
	}
	testutil.AssertTrue(t, hasVironAction, "Viron should have a manual action to reuse blue card actions (NOT YET IMPLEMENTED)")
}
