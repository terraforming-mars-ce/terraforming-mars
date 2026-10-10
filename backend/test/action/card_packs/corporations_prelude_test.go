package card_packs_test

import (
	"context"
	"slices"
	"testing"

	"openmars/internal/action/admin"
	cardAction "openmars/internal/action/card"
	confirmAction "openmars/internal/action/confirmation"
	tileAction "openmars/internal/action/tile"
	turnAction "openmars/internal/action/turn_management"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

func TestCheungShingMars_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
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

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
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

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Point Luna"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Point Luna")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 38, resources.Credits, "Point Luna should start with 38 credits")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 1, production.Titanium, "Point Luna should start with 1 titanium production")

	testutil.AssertEqual(t, 1, len(p.Hand().Cards()), "Point Luna draws exactly one card for its own Earth tag")
}

func TestPointLuna_DrawCardWhenPlayingEarthTag(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Point Luna"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Point Luna")

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})

	sponsorsID := testutil.CardID("Sponsors")
	p.Hand().AddCard(sponsorsID)

	handBefore := len(p.Hand().Cards())

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := shared.NativePayment(shared.
		ResourceCredit, 6)
	err = playCard.Execute(ctx, testGame.ID(), playerID, sponsorsID, payment, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Sponsors should succeed")

	handAfter := len(p.Hand().Cards())
	testutil.AssertEqual(t, handBefore, handAfter, "Point Luna replaces the played Earth card with exactly one draw")
}

func TestValleyTrust_StartingResources(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Valley Trust first action draws from prelude deck, so we need one
	preludeIDs := []string{"P01", "P02", "P03", "P04", "P05"}
	testGame.InitDeck(nil, nil, preludeIDs)

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
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

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
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

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Vitor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Vitor")

	p, _ := testGame.GetPlayer(playerID)
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 45, resources.Credits, "Vitor should start with 45 credits")
}

func TestVitor_FundAwardForFree(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Vitor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Vitor")

	p, _ := testGame.GetPlayer(playerID)

	// Should have pending award fund selection
	pending := p.Selection().GetPendingAwardFundSelection()
	testutil.AssertTrue(t, pending != nil, "Vitor should create pending award fund selection")
	testutil.AssertTrue(t, len(pending.AvailableAwards) > 0, "Should have available awards")

	// Should have a forced first action
	forcedAction := testGame.GetForcedFirstAction(playerID)
	testutil.AssertTrue(t, forcedAction != nil, "Vitor should create forced first action")
	testutil.AssertEqual(t, "resolving", forcedAction.State, "First action should be resolving")

	// Record credits before
	creditsBefore := p.Resources().Get().Credits

	// Confirm the award fund
	confirmAction := confirmAction.NewConfirmAwardFundAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
	err = confirmAction.Execute(ctx, testGame.ID(), playerID, pending.AvailableAwards[0])
	testutil.AssertNoError(t, err, "ConfirmAwardFund should succeed")

	// Credits should be unchanged (free award)
	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore, creditsAfter, "Vitor should fund award for free (no credit deduction)")

	// Award should be funded
	testutil.AssertTrue(t, testGame.Awards().IsFunded(shared.AwardType(pending.AvailableAwards[0])), "Award should be funded")
	testutil.AssertTrue(t, testGame.Awards().IsFundedBy(shared.AwardType(pending.AvailableAwards[0]), playerID), "Award should be funded by the player")

	// Pending selection should be cleared
	testutil.AssertTrue(t, p.Selection().GetPendingAwardFundSelection() == nil, "Pending award fund selection should be cleared")

	// Forced first action should be cleared
	testutil.AssertTrue(t, testGame.GetForcedFirstAction(playerID) == nil, "Forced first action should be cleared")
}

func TestVitor_OnlyOffersAwardsInThisGame(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()
	awardRegistry := testutil.CreateTestAwardRegistry()

	all := awardRegistry.GetAll()
	testutil.AssertTrue(t, len(all) > 3, "registry has more awards than a game uses")
	selected := []string{all[0].ID, all[1].ID}
	notInGame := all[2].ID
	testGame.SetSelectedAwards(selected)

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, awardRegistry, logger)
	testutil.AssertNoError(t, setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Vitor")), "set Vitor")

	p, _ := testGame.GetPlayer(playerID)
	pending := p.Selection().GetPendingAwardFundSelection()
	testutil.AssertTrue(t, pending != nil, "Vitor should create pending award fund selection")
	testutil.AssertEqual(t, len(selected), len(pending.AvailableAwards), "only this game's awards are offered")
	for _, id := range pending.AvailableAwards {
		testutil.AssertTrue(t, slices.Contains(selected, id), "offered award is in this game: "+id)
	}

	confirm := confirmAction.NewConfirmAwardFundAction(repo, cardRegistry, awardRegistry, logger)
	testutil.AssertErrorContains(t, confirm.Execute(ctx, testGame.ID(), playerID, notInGame), "award scientist is not available for selection", "award outside this game is rejected")
	testutil.AssertFalse(t, testGame.Awards().IsFunded(shared.AwardType(notInGame)), "nothing funded")
}

func TestVitor_Gain3MCWhenPlayingCardWithVP(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, testutil.CreateTestAwardRegistry(), logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Vitor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed for Vitor")

	p, _ := testGame.GetPlayer(playerID)

	// Clear Vitor's forced award fund selection so card play is not blocked
	p.Selection().SetPendingAwardFundSelection(nil)
	testutil.AssertNoError(t, testGame.SetForcedFirstAction(ctx, playerID, nil), "clear forced first action")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})

	vpCardID := testutil.CardID("Colonizer Training Camp")
	p.Hand().AddCard(vpCardID)

	creditsBefore := p.Resources().Get().Credits

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := shared.NativePayment(shared.
		ResourceCredit, 8)
	err = playCard.Execute(ctx, testGame.ID(), playerID, vpCardID, payment, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing Colonizer Training Camp should succeed")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore-8+3, creditsAfter, "Vitor should grant 3 MC when playing a card with VP")
}

func TestCorporationEffects_NormalStartingChoices(t *testing.T) {
	for _, name := range []string{"Point Luna", "Arklight", "Saturn Systems", "Manutech"} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			g, _, registry, id, _ := testutil.SetupTwoPlayerGame(t)
			p, _ := g.GetPlayer(id)
			corpID := testutil.CardID(name)
			p.SetCorporationID(corpID)
			testutil.AssertNoError(t, g.SetDeferredStartingChoices(ctx, id, &shared.DeferredStartingChoices{CorporationID: corpID}), "starting choices")
			log := testutil.TestLogger()
			processor := gamecards.NewCorporationProcessor(registry, nil, log)
			testutil.AssertNoError(t, turnAction.ApplyCorpForPlayer(ctx, g, id, registry, processor, log), "apply normal corporation selection")
			switch name {
			case "Point Luna":
				testutil.AssertEqual(t, 1, len(p.Hand().Cards()), "one draw including own Earth tag")
				testutil.AssertEqual(t, 38, p.Resources().Get().Credits, "starting credits")
				testutil.AssertEqual(t, 1, p.Resources().Production().Titanium, "starting titanium production")
			case "Arklight":
				testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(corpID), "one animal including own animal tag")
				testutil.AssertEqual(t, 45, p.Resources().Get().Credits, "starting credits")
				testutil.AssertEqual(t, 2, p.Resources().Production().Credits, "starting credit production")
			case "Saturn Systems":
				testutil.AssertEqual(t, 42, p.Resources().Get().Credits, "starting credits")
				testutil.AssertEqual(t, 1, p.Resources().Production().Titanium, "starting titanium production")
				testutil.AssertEqual(t, 1, p.Resources().Production().Credits, "own Jovian tag rewards exactly once")
			case "Manutech":
				testutil.AssertEqual(t, 35, p.Resources().Get().Credits, "starting credits")
				testutil.AssertEqual(t, 1, p.Resources().Production().Steel, "starting steel production")
				testutil.AssertEqual(t, 1, p.Resources().Get().Steel, "starting production rewards exactly once")

			}
		})
	}
}

func TestValleyTrust_FirstActionWaitsForChosenPreludePlacement(t *testing.T) {
	ctx := context.Background()
	g, repo, registry, first, id := testutil.SetupTwoPlayerGame(t)
	g.InitDeck([]string{"001", "002", "003"}, nil, []string{"P12", "P01", "P03"})
	p, _ := g.GetPlayer(id)
	testutil.AssertNoError(t, g.UpdatePhase(ctx, shared.GamePhaseInitApplyCorp), "setup")
	setter := admin.NewSetCorporationAction(repo, registry, nil, testutil.TestLogger())
	testutil.AssertNoError(t, setter.Execute(ctx, g.ID(), id, testutil.CardID("Valley Trust")), "queue first action")
	testutil.AssertTrue(t, p.Selection().GetPendingCardDrawSelection() == nil, "no early prelude peek")
	testutil.AssertNoError(t, g.UpdatePhase(ctx, shared.GamePhaseAction), "action phase")
	testutil.AssertNoError(t, g.SetCurrentTurn(ctx, first, 2), "first player goes first")
	testutil.AssertTrue(t, p.Selection().GetPendingCardDrawSelection() == nil, "wait for owner's turn")
	testutil.AssertNoError(t, g.SetCurrentTurn(ctx, id, 2), "owner turn")
	confirm := confirmAction.NewConfirmCardDrawAction(repo, registry, testutil.TestLogger())
	testutil.AssertNoError(t, confirm.Execute(ctx, g.ID(), id, []string{"P12"}, nil, shared.Payment{}), "choose Experimental Forest")
	testutil.AssertTrue(t, p.PlayedCards().Contains("P12"), "chosen prelude is played")
	testutil.AssertTrue(t, g.GetForcedFirstAction(id) != nil, "placement still belongs to first action")
	testutil.AssertEqual(t, 2, g.CurrentTurn().ActionsRemaining(), "action completes after placement")
	pending := g.GetPendingTileSelection(id)
	testutil.AssertTrue(t, pending != nil, "prelude queues greenery")
	_, err := tileAction.NewSelectTileAction(repo, registry, nil, testutil.TestLogger()).Execute(ctx, g.ID(), id, pending.AvailableHexes[0])
	testutil.AssertNoError(t, err, "place greenery")
	testutil.AssertTrue(t, g.GetForcedFirstAction(id) == nil, "all first-action effects complete")
	testutil.AssertEqual(t, 1, g.CurrentTurn().ActionsRemaining(), "one normal action remains")
	testutil.AssertErrorContains(t, confirm.Execute(ctx, g.ID(), id, []string{"P12"}, nil, shared.Payment{}), "no pending card draw selection", "cannot confirm again")
	testutil.AssertEqual(t, 1, g.CurrentTurn().ActionsRemaining(), "no duplicate consumption")
	testutil.AssertEqual(t, 0, len(p.Selection().CardReceipts()), "selection does not create a redundant receipt")
}

func TestVitor_FirstActionWaitsForOwnerTurn(t *testing.T) {
	ctx := context.Background()
	g, repo, registry, first, id := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(id)
	awards := testutil.CreateTestAwardRegistry()
	testutil.AssertNoError(t, g.UpdatePhase(ctx, shared.GamePhaseInitApplyCorp), "setup")
	testutil.AssertNoError(t, admin.NewSetCorporationAction(repo, registry, awards, testutil.TestLogger()).Execute(ctx, g.ID(), id, testutil.CardID("Vitor")), "queue Vitor")
	testutil.AssertTrue(t, p.Selection().GetPendingAwardFundSelection() == nil, "no award selection during setup")
	testutil.AssertNoError(t, g.UpdatePhase(ctx, shared.GamePhaseAction), "action phase")
	testutil.AssertNoError(t, g.SetCurrentTurn(ctx, first, 2), "first player's turn")
	testutil.AssertTrue(t, p.Selection().GetPendingAwardFundSelection() == nil, "not owner's turn")
	testutil.AssertNoError(t, g.SetCurrentTurn(ctx, id, 2), "Vitor turn")
	selected := p.Selection().GetPendingAwardFundSelection().AvailableAwards[0]
	testutil.AssertNoError(t, confirmAction.NewConfirmAwardFundAction(repo, registry, awards, testutil.TestLogger()).Execute(ctx, g.ID(), id, selected), "fund award")
	testutil.AssertEqual(t, 1, g.CurrentTurn().ActionsRemaining(), "first action consumes one action")
}
