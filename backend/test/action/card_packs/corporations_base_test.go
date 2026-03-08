package card_packs_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action/admin"
	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/events"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestCrediCor_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("CrediCor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for CrediCor")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 57, resources.Credits, "CrediCor should start with 57 credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 0, production.Credits, "Credit production should be 0")
	testutil.AssertEqual(t, 0, production.Steel, "Steel production should be 0")
	testutil.AssertEqual(t, 0, production.Titanium, "Titanium production should be 0")
	testutil.AssertEqual(t, 0, production.Plants, "Plant production should be 0")
	testutil.AssertEqual(t, 0, production.Energy, "Energy production should be 0")
	testutil.AssertEqual(t, 0, production.Heat, "Heat production should be 0")
}

func TestCrediCor_Gain4MCWhenPlayingExpensiveCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("CrediCor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for CrediCor")

	p, _ := testGame.GetPlayer(playerID)
	res := p.Resources().Get()
	res.Credits = 100
	p.Resources().Set(res)

	cardID := testutil.CardID("Comet")
	p.Hand().AddCard(cardID)

	creditsBefore := p.Resources().Get().Credits

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 21}
	err = playCard.Execute(ctx, testGame.ID(), playerID, cardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "PlayCard should succeed for Comet")

	time.Sleep(50 * time.Millisecond)

	expected := creditsBefore - 21 + 4
	actual := p.Resources().Get().Credits
	testutil.AssertEqual(t, expected, actual, "CrediCor should gain 4 MC back after playing a card costing 20+ MC")
}

func TestEcoline_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Ecoline"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Ecoline")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 36, resources.Credits, "Ecoline should start with 36 credits")
	testutil.AssertEqual(t, 3, resources.Plants, "Ecoline should start with 3 plants")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 2, production.Plants, "Ecoline should start with 2 plant production")
}

func TestEcoline_DiscountEffectRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Ecoline"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Ecoline")

	p, _ := testGame.GetPlayer(playerID)
	effects := p.Effects().List()

	found := false
	for _, effect := range effects {
		if effect.CardName == "Ecoline" && effect.BehaviorIndex == 1 {
			found = true
			testutil.AssertEqual(t, shared.ResourceDiscount, effect.Behavior.Outputs[0].ResourceType,
				"Ecoline effect output should be a discount")
			testutil.AssertEqual(t, 1, effect.Behavior.Outputs[0].Amount,
				"Ecoline discount amount should be 1")
			break
		}
	}
	testutil.AssertTrue(t, found, "Ecoline should have a registered discount effect at behavior index 1")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	discounts := calculator.CalculateStandardProjectDiscounts(p, shared.StandardProjectConvertPlantsToGreenery)
	testutil.AssertEqual(t, 1, discounts[shared.ResourcePlant],
		"Ecoline should provide 1 plant discount for convert-plants-to-greenery")
}

func TestHelion_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Helion"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 42, resources.Credits, "Should have 42 credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 3, production.Heat, "Should have 3 heat production")
	testutil.AssertEqual(t, 0, production.Credits, "Credit production should be 0")
	testutil.AssertEqual(t, 0, production.Steel, "Steel production should be 0")
	testutil.AssertEqual(t, 0, production.Titanium, "Titanium production should be 0")
	testutil.AssertEqual(t, 0, production.Plants, "Plant production should be 0")
	testutil.AssertEqual(t, 0, production.Energy, "Energy production should be 0")
}

func TestHelion_CanPayWithHeat(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Helion"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceHeat: 10})

	p.Hand().AddCard(testutil.CardID("Virus"))

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{
		Credits: 0,
		Substitutes: map[shared.ResourceType]int{
			shared.ResourceHeat: 1,
		},
	}
	err = playCardAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Virus"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Virus with heat payment should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 9, resources.Heat, "Heat should decrease by 1")
}

func TestInterplanetaryCinematics_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Interplanetary Cinematics"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 30, resources.Credits, "Should have 30 credits")
	testutil.AssertEqual(t, 20, resources.Steel, "Should have 20 steel")
}

func TestInterplanetaryCinematics_Gain2MCWhenPlayingEvent(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Interplanetary Cinematics"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	p.Hand().AddCard(testutil.CardID("Virus"))

	creditsBefore := p.Resources().Get().Credits

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err = playCardAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Virus"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Virus should succeed")

	time.Sleep(50 * time.Millisecond)

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-1+2, creditsAfter, "Should gain 2 MC from IC effect after paying 1 for Virus")
}

func TestInventrix_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Inventrix"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 45, resources.Credits, "Inventrix should start with 45 credits")
}

func TestInventrix_GlobalParameterLenienceRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Inventrix"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	lenience := calculator.CalculateGlobalParameterLenience(p, "temperature")
	testutil.AssertEqual(t, 2, lenience, "Inventrix should provide global parameter lenience of 2")
}

func TestInventrix_LenienceStacksWithSpecialDesign(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Inventrix"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)
	lenienceBefore := calculator.CalculateGlobalParameterLenience(p, "temperature")
	testutil.AssertEqual(t, 2, lenienceBefore, "Inventrix alone should provide lenience of 2")

	specialDesignID := testutil.CardID("Special Design")
	p.Hand().AddCard(specialDesignID)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 4}
	err = playCard.Execute(ctx, testGame.ID(), playerID, specialDesignID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Special Design should play successfully")

	lenienceAfter := calculator.CalculateGlobalParameterLenience(p, "temperature")
	testutil.AssertEqual(t, 4, lenienceAfter, "Inventrix (2) + Special Design (2) should stack to lenience of 4")
}

func TestMiningGuild_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Mining Guild"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 30, resources.Credits, "Mining Guild should start with 30 credits")
	testutil.AssertEqual(t, 5, resources.Steel, "Mining Guild should start with 5 steel")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Steel, "Mining Guild should start with 1 steel production")
}

func TestMiningGuild_SteelProductionOnPlacementBonus(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Mining Guild"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	productionBefore := p.Resources().Production().Steel
	testutil.AssertEqual(t, 1, productionBefore, "Mining Guild should start with 1 steel production")

	events.Publish(testGame.EventBus(), events.PlacementBonusGainedEvent{
		GameID:   testGame.ID(),
		PlayerID: playerID,
		Resources: map[string]int{
			"steel": 1,
		},
	})

	time.Sleep(50 * time.Millisecond)

	productionAfter := p.Resources().Production().Steel
	testutil.AssertEqual(t, 2, productionAfter, "Steel production should be 2 after placement bonus trigger")
}

func TestPhoboLog_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("PhoboLog"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 23, resources.Credits, "PhoboLog should start with 23 credits")
	testutil.AssertEqual(t, 10, resources.Titanium, "PhoboLog should start with 10 titanium")
}

func TestPhoboLog_TitaniumWorthExtra(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("PhoboLog"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	transNeptuneProbeID := testutil.CardID("Trans-Neptune Probe")
	p.Hand().AddCard(transNeptuneProbeID)

	titaniumBefore := p.Resources().Get().Titanium

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Titanium: 2}
	err = playCard.Execute(ctx, testGame.ID(), playerID, transNeptuneProbeID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Trans-Neptune Probe with 2 titanium should succeed (4 M€ each = 8 M€ covers 6-cost card)")

	titaniumAfter := p.Resources().Get().Titanium
	testutil.AssertEqual(t, titaniumBefore-2, titaniumAfter, "Titanium should decrease by 2")
}

func TestTharsisRepublic_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 40, resources.Credits, "Tharsis Republic should start with 40 credits")
}

func TestTharsisRepublic_ForcedFirstActionSetup(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	forcedAction := testGame.GetForcedFirstAction(playerID)
	testutil.AssertTrue(t, forcedAction != nil, "Tharsis Republic should create a forced first action")
	testutil.AssertEqual(t, "city-placement", forcedAction.ActionType, "Forced first action should be city-placement")
	testutil.AssertEqual(t, testutil.CardID("Tharsis Republic"), forcedAction.CorporationID, "Forced first action should reference Tharsis Republic")
}

func TestTharsisRepublic_GainCreditsAndProductionOnCityPlacement(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Tharsis Republic"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	creditsBefore := p.Resources().Get().Credits
	creditProductionBefore := p.Resources().Production().Credits

	events.Publish(testGame.EventBus(), events.TilePlacedEvent{
		GameID:   testGame.ID(),
		PlayerID: playerID,
		TileType: "city",
	})

	time.Sleep(50 * time.Millisecond)

	testutil.AssertEqual(t, creditsBefore+3, p.Resources().Get().Credits, "Self city placement should gain 3 M€")
	testutil.AssertEqual(t, creditProductionBefore+1, p.Resources().Production().Credits, "City on mars should increase M€ production by 1")
}

func TestThorGate_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("ThorGate"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 48, resources.Credits, "ThorGate should start with 48 credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Energy, "ThorGate should start with 1 energy production")
}

func TestThorGate_DiscountEffectRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("ThorGate"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)

	effects := p.Effects().List()
	found := false
	for _, effect := range effects {
		if effect.CardName == "ThorGate" && effect.BehaviorIndex == 1 {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "ThorGate should have registered its discount effect at behavior index 1")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	powerCard := &gamecards.Card{
		ID:   "test-power-card",
		Name: "Test Power Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagPower},
	}
	discount := calculator.CalculateCardDiscounts(p, powerCard)
	testutil.AssertEqual(t, 3, discount, "ThorGate should provide 3 M€ discount for power cards")

	nonPowerCard := &gamecards.Card{
		ID:   "test-building-card",
		Name: "Test Building Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding},
	}
	nonPowerDiscount := calculator.CalculateCardDiscounts(p, nonPowerCard)
	testutil.AssertEqual(t, 0, nonPowerDiscount, "ThorGate should provide no discount for non-power cards")
}

func TestUNMI_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("United Nations Mars Initiative"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 40, resources.Credits, "UNMI should start with 40 credits")
}

func TestUNMI_Pay3MCToRaiseTR(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("United Nations Mars Initiative"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.GenerationalEvents().Increment(shared.GenerationalEventTRRaise)

	trBefore := p.Resources().TerraformRating()
	creditsBefore := p.Resources().Get().Credits

	cardID := testutil.CardID("United Nations Mars Initiative")
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), playerID, cardID, 1, nil, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "UNMI action should succeed")

	testutil.AssertEqual(t, creditsBefore-3, p.Resources().Get().Credits, "UNMI action should cost 3 credits")
	testutil.AssertEqual(t, trBefore+1, p.Resources().TerraformRating(), "UNMI action should raise TR by 1")
}
