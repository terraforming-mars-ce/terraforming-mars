package play_card_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Miranda Resort (051) ---
// "Increase your M€ production 1 step for each Earth tag you have."

func TestMirandaResort_CreditProductionPerEarthTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	earthTag := shared.TagEarth

	mirandaResort := gamecards.Card{
		ID:   "card-miranda-resort",
		Name: "Miranda Resort",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.ProductionCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagEarth),
							Amount:       1,
							Target:       &selfPlayerTarget,
							Tag:          &earthTag,
						},
					},
				},
			},
		},
	}

	earthCard1 := gamecards.Card{ID: "card-earth-1", Name: "Earth Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagEarth}, Cost: 0}
	earthCard2 := gamecards.Card{ID: "card-earth-2", Name: "Earth Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagEarth}, Cost: 0}
	earthCard3 := gamecards.Card{ID: "card-earth-3", Name: "Earth Card 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagEarth}, Cost: 0}
	nonEarthCard := gamecards.Card{ID: "card-non-earth", Name: "Non-Earth Card", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagBuilding}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mirandaResort, earthCard1, earthCard2, earthCard3, nonEarthCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Play 3 earth-tagged cards and 1 non-earth card
	p.PlayedCards().AddCard("card-earth-1", "Earth Card 1", "automated", []string{"earth"})
	p.PlayedCards().AddCard("card-earth-2", "Earth Card 2", "automated", []string{"earth"})
	p.PlayedCards().AddCard("card-earth-3", "Earth Card 3", "automated", []string{"earth"})
	p.PlayedCards().AddCard("card-non-earth", "Non-Earth Card", "automated", []string{"building"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-miranda-resort")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-miranda-resort", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Miranda Resort should play successfully")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits+3, productionAfter.Credits,
		"Should gain +3 credit production (1 per each of 3 earth tags)")
}

func TestMirandaResort_ZeroEarthTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	earthTag := shared.TagEarth

	mirandaResort := gamecards.Card{
		ID:   "card-miranda-resort",
		Name: "Miranda Resort",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.ProductionCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagEarth),
							Amount:       1,
							Target:       &selfPlayerTarget,
							Tag:          &earthTag,
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mirandaResort})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-miranda-resort")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-miranda-resort", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Miranda Resort should play successfully with 0 earth tags")

	productionAfter := p.Resources().Production()
	testutil.AssertEqual(t, productionBefore.Credits, productionAfter.Credits,
		"Credit production should not change with 0 earth tags")
}

// --- Terraforming Ganymede (197) ---
// "Raise your TR 1 step for each Jovian tag you have, including this."

func TestTerraformingGanymede_TRPerJovianTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	jovianTag := shared.TagJovian

	terraformingGanymede := gamecards.Card{
		ID:   "card-terraforming-ganymede",
		Name: "Terraforming Ganymede",
		Type: gamecards.CardTypeAutomated,
		Cost: 33,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.GlobalParameterCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTR, Amount: 1, Target: "self-player"},
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagJovian),
							Amount:       1,
							Target:       &selfPlayerTarget,
							Tag:          &jovianTag,
						},
					},
				},
			},
		},
	}

	jovianCard1 := gamecards.Card{ID: "card-jovian-1", Name: "Jovian Card 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}, Cost: 0}
	jovianCard2 := gamecards.Card{ID: "card-jovian-2", Name: "Jovian Card 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{terraformingGanymede, jovianCard1, jovianCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Play 2 jovian-tagged cards first
	p.PlayedCards().AddCard("card-jovian-1", "Jovian Card 1", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("card-jovian-2", "Jovian Card 2", "automated", []string{"jovian"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-terraforming-ganymede")

	trBefore := p.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 33}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-terraforming-ganymede", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Terraforming Ganymede should play successfully")

	trAfter := p.Resources().TerraformRating()
	// 2 existing jovian tags + 1 from Terraforming Ganymede itself = 3 TR
	testutil.AssertEqual(t, trBefore+3, trAfter,
		"Should gain +3 TR (1 per each of 2 existing + 1 from self jovian tag)")
}

func TestTerraformingGanymede_OnlyCountsSelfPlayerTags(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	selfPlayerTarget := "self-player"
	jovianTag := shared.TagJovian

	terraformingGanymede := gamecards.Card{
		ID:   "card-terraforming-ganymede",
		Name: "Terraforming Ganymede",
		Type: gamecards.CardTypeAutomated,
		Cost: 33,
		Tags: []shared.CardTag{shared.TagJovian, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.GlobalParameterCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTR, Amount: 1, Target: "self-player"},
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagJovian),
							Amount:       1,
							Target:       &selfPlayerTarget,
							Tag:          &jovianTag,
						},
					},
				},
			},
		},
	}

	jovianCard := gamecards.Card{ID: "card-jovian-other", Name: "Jovian Other", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{terraformingGanymede, jovianCard})

	players := testGame.GetAllPlayers()
	attacker := players[0]
	other := players[1]
	attacker.SetCorporationID("corp-tharsis-republic")
	other.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, attacker.ID(), 2), "set current turn")

	// Other player has 3 jovian tags - should NOT count for attacker
	other.PlayedCards().AddCard("card-jovian-other", "Jovian Other", "automated", []string{"jovian"})

	attacker.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	attacker.Hand().AddCard("card-terraforming-ganymede")

	trBefore := attacker.Resources().TerraformRating()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 33}
	err := playCardAction.Execute(ctx, testGame.ID(), attacker.ID(), "card-terraforming-ganymede", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Terraforming Ganymede should play successfully")

	trAfter := attacker.Resources().TerraformRating()
	// Only 1 jovian tag from Terraforming Ganymede itself (other player's tags don't count)
	testutil.AssertEqual(t, trBefore+1, trAfter,
		"Should gain +1 TR (only self jovian tag, not other player's)")
}

// --- Imported Nitrogen (163) ---
// "Raise your TR 1 step and gain 4 plants. Add 3 microbes to another card and 2 animals to another card."

func TestImportedNitrogen_MultipleAnyCardTargets(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	importedNitrogen := gamecards.Card{
		ID:   "card-imported-nitrogen",
		Name: "Imported Nitrogen",
		Type: gamecards.CardTypeEvent,
		Cost: 23,
		Tags: []shared.CardTag{shared.TagEarth, shared.TagSpace},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewBasicResourceCondition(shared.ResourcePlant, 4, "self-player"),
					shared.NewCardStorageCondition(shared.ResourceMicrobe, 3, "any-card"),
					shared.NewGlobalParameterCondition(shared.ResourceTR, 1, "self-player"),
					shared.NewCardStorageCondition(shared.ResourceAnimal, 2, "any-card"),
				},
			},
		},
	}

	microbeCard := gamecards.Card{
		ID: "card-microbe-host", Name: "Microbe Host", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagMicrobe},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceMicrobe, Starting: 0},
	}
	animalCard := gamecards.Card{
		ID: "card-animal-host", Name: "Animal Host", Type: gamecards.CardTypeActive, Cost: 0,
		Tags:            []shared.CardTag{shared.TagAnimal},
		ResourceStorage: &gamecards.ResourceStorage{Type: shared.ResourceAnimal, Starting: 0},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{importedNitrogen, microbeCard, animalCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.PlayedCards().AddCard("card-microbe-host", "Microbe Host", "active", []string{"microbe"})
	p.Resources().AddToStorage("card-microbe-host", 0)
	p.PlayedCards().AddCard("card-animal-host", "Animal Host", "active", []string{"animal"})
	p.Resources().AddToStorage("card-animal-host", 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-imported-nitrogen")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 23}

	// Test with two correct targets (microbe host for microbes, animal host for animals)
	initialPlants := p.Resources().Get().Plants
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-imported-nitrogen", payment, nil, []string{"card-microbe-host", "card-animal-host"}, nil, nil)
	testutil.AssertNoError(t, err, "Should succeed with two correct targets")

	microbeStorage := p.Resources().GetCardStorage("card-microbe-host")
	testutil.AssertEqual(t, 3, microbeStorage, "Microbe host should have 3 microbes")

	animalStorage := p.Resources().GetCardStorage("card-animal-host")
	testutil.AssertEqual(t, 2, animalStorage, "Animal host should have 2 animals")

	finalPlants := p.Resources().Get().Plants
	testutil.AssertEqual(t, initialPlants+4, finalPlants, "Player should gain 4 plants")
}
