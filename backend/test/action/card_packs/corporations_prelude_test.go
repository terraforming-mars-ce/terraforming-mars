package card_packs_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action/admin"
	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestCheungShingMars_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Cheung Shing Mars"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Cheung Shing Mars")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 44, resources.Credits, "Cheung Shing Mars should start with 44 credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 3, production.Credits, "Cheung Shing Mars should start with 3 credit production")
}

func TestCheungShingMars_DiscountEffectRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Cheung Shing Mars"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Cheung Shing Mars")

	p, _ := testGame.GetPlayer(playerID)
	effects := p.Effects().List()

	found := false
	for _, effect := range effects {
		if effect.CardName == "Cheung Shing Mars" && effect.BehaviorIndex == 1 {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "Cheung Shing Mars should have a registered discount effect at behavior index 1")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	buildingCard := &gamecards.Card{
		ID:   "test-building-card",
		Name: "Test Building Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding},
	}
	discount := calculator.CalculateCardDiscounts(p, buildingCard)
	testutil.AssertEqual(t, 2, discount, "Cheung Shing Mars should provide 2 discount for building-tagged cards")

	nonBuildingCard := &gamecards.Card{
		ID:   "test-space-card",
		Name: "Test Space Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagSpace},
	}
	nonBuildingDiscount := calculator.CalculateCardDiscounts(p, nonBuildingCard)
	testutil.AssertEqual(t, 0, nonBuildingDiscount, "Cheung Shing Mars should provide no discount for non-building cards")
}

func TestPointLuna_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Point Luna"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Point Luna")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 38, resources.Credits, "Point Luna should start with 38 credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Titanium, "Point Luna should start with 1 titanium production")

	if len(p.Hand().Cards()) < 1 {
		t.Fatalf("Point Luna should draw at least 1 card at startup, but hand has %d cards", len(p.Hand().Cards()))
	}
}

func TestPointLuna_DrawCardWhenPlayingEarthTag(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Point Luna"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Point Luna")

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	sponsorsID := testutil.CardID("Sponsors")
	p.Hand().AddCard(sponsorsID)

	handBefore := len(p.Hand().Cards())

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 6}
	err = playCard.Execute(ctx, testGame.ID(), playerID, sponsorsID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Sponsors should succeed")

	time.Sleep(50 * time.Millisecond)

	handAfter := len(p.Hand().Cards())
	testutil.AssertTrue(t, handAfter >= handBefore, "Point Luna should draw a card when an Earth tag is played")
}

func TestValleyTrust_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Valley Trust first action draws from prelude deck, so we need one
	preludeIDs := []string{"P01", "P02", "P03", "P04", "P05"}
	testGame.InitDeck(nil, nil, preludeIDs)

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Valley Trust"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Valley Trust")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 37, resources.Credits, "Valley Trust should start with 37 credits")
}

func TestValleyTrust_DiscountEffectRegistered(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	preludeIDs := []string{"P01", "P02", "P03", "P04", "P05"}
	testGame.InitDeck(nil, nil, preludeIDs)

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Valley Trust"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Valley Trust")

	p, _ := testGame.GetPlayer(playerID)
	effects := p.Effects().List()

	found := false
	for _, effect := range effects {
		if effect.CardName == "Valley Trust" && effect.BehaviorIndex == 1 {
			found = true
			break
		}
	}
	testutil.AssertTrue(t, found, "Valley Trust should have a registered discount effect at behavior index 1")

	calculator := gamecards.NewRequirementModifierCalculator(cardRegistry)

	scienceCard := &gamecards.Card{
		ID:   "test-science-card",
		Name: "Test Science Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagScience},
	}
	discount := calculator.CalculateCardDiscounts(p, scienceCard)
	testutil.AssertEqual(t, 2, discount, "Valley Trust should provide 2 discount for science-tagged cards")

	nonScienceCard := &gamecards.Card{
		ID:   "test-building-card",
		Name: "Test Building Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding},
	}
	nonScienceDiscount := calculator.CalculateCardDiscounts(p, nonScienceCard)
	testutil.AssertEqual(t, 0, nonScienceDiscount, "Valley Trust should provide no discount for non-science cards")
}

func TestVitor_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Vitor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Vitor")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 45, resources.Credits, "Vitor should start with 45 credits")
}

func TestVitor_Gain3MCWhenPlayingCardWithVP(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Vitor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Vitor")

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	vpCardID := testutil.CardID("Colonizer Training Camp")
	p.Hand().AddCard(vpCardID)

	creditsBefore := p.Resources().Get().Credits

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err = playCard.Execute(ctx, testGame.ID(), playerID, vpCardID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Colonizer Training Camp should succeed")

	time.Sleep(50 * time.Millisecond)

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-8+3, creditsAfter, "Vitor should grant 3 MC when playing a card with VP")
}
