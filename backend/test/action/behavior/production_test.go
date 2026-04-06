package behavior_test

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

// --- Worms (130) ---
// "Requires 4% oxygen. Increase your plant production 1 step for every 2 microbe tags you have, including this."

// Helper to build the Worms card definition used across all tests
func makeWormsCard() gamecards.Card {
	selfPlayerTarget := "self-player"
	microbeTag := shared.TagMicrobe
	return gamecards.Card{
		ID:   "130",
		Name: "Worms",
		Type: gamecards.CardTypeAutomated,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagMicrobe},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.ProductionCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourcePlantProduction, Amount: 1, Target: "self-player"},
						Per: &shared.PerCondition{
							ResourceType: shared.ResourceType(shared.TagMicrobe),
							Amount:       2,
							Target:       &selfPlayerTarget,
							Tag:          &microbeTag,
						},
					},
				},
			},
		},
	}
}

// 1 microbe tag before Worms → total 2 → floor(2/2)=1 → +1 plant production
func TestWorms_OneMicrobeTagBefore_GainsOnePlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	microbeCard := gamecards.Card{
		ID: "card-microbe-1", Name: "Microbe Card", Type: gamecards.CardTypeAutomated,
		Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0,
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms, microbeCard})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// 1 microbe tag already in play
	p.PlayedCards().AddCard("card-microbe-1", "Microbe Card", "automated", []string{"microbe"})

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// 1 existing + 1 from Worms = 2 microbe tags; floor(2/2) * 1 = 1
	testutil.AssertEqual(t, productionBefore.Plants+1, productionAfter.Plants,
		"Should gain +1 plant production (1 existing + 1 self = 2 microbe tags, floor(2/2)=1)")
}

// 0 microbe tags before Worms → total 1 (just Worms) → floor(1/2)=0 → +0 plant production
func TestWorms_ZeroMicrobeTagsBefore_GainsZeroPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// No microbe tags in play — Worms is the only microbe card
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// Only Worms itself = 1 microbe tag; floor(1/2) * 1 = 0
	testutil.AssertEqual(t, productionBefore.Plants, productionAfter.Plants,
		"Should gain +0 plant production (only 1 microbe tag from Worms, floor(1/2)=0)")
}

// 3 microbe tags before Worms → total 4 → floor(4/2)=2 → +2 plant production
func TestWorms_ThreeMicrobeTagsBefore_GainsTwoPlantProduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	microbeCard1 := gamecards.Card{ID: "card-m1", Name: "Microbe 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}
	microbeCard2 := gamecards.Card{ID: "card-m2", Name: "Microbe 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}
	microbeCard3 := gamecards.Card{ID: "card-m3", Name: "Microbe 3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms, microbeCard1, microbeCard2, microbeCard3})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.PlayedCards().AddCard("card-m1", "Microbe 1", "automated", []string{"microbe"})
	p.PlayedCards().AddCard("card-m2", "Microbe 2", "automated", []string{"microbe"})
	p.PlayedCards().AddCard("card-m3", "Microbe 3", "automated", []string{"microbe"})

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// 3 existing + 1 from Worms = 4; floor(4/2) * 1 = 2
	testutil.AssertEqual(t, productionBefore.Plants+2, productionAfter.Plants,
		"Should gain +2 plant production (3 existing + 1 self = 4 microbe tags, floor(4/2)=2)")
}

// 2 microbe tags before Worms → total 3 → floor(3/2)=1 → +1 plant production (rounds down)
func TestWorms_TwoMicrobeTagsBefore_RoundsDown(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	worms := makeWormsCard()
	microbeCard1 := gamecards.Card{ID: "card-m1", Name: "Microbe 1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}
	microbeCard2 := gamecards.Card{ID: "card-m2", Name: "Microbe 2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagMicrobe}, Cost: 0}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{worms, microbeCard1, microbeCard2})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.PlayedCards().AddCard("card-m1", "Microbe 1", "automated", []string{"microbe"})
	p.PlayedCards().AddCard("card-m2", "Microbe 2", "automated", []string{"microbe"})

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard("130")

	productionBefore := p.Resources().Production()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 8}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "130", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Worms should play successfully")

	productionAfter := p.Resources().Production()
	// 2 existing + 1 from Worms = 3; floor(3/2) * 1 = 1
	testutil.AssertEqual(t, productionBefore.Plants+1, productionAfter.Plants,
		"Should gain +1 plant production (2 existing + 1 self = 3 microbe tags, floor(3/2)=1)")
}

func TestProduction_AnyPlayerReduction(t *testing.T) {
	testGame, _, cardRegistry, playerID, targetPlayerID := testutil.SetupTwoPlayerGame(t)

	target, _ := testGame.GetPlayer(targetPlayerID)
	p, _ := testGame.GetPlayer(playerID)

	target.Resources().AddProduction(map[shared.ResourceType]int{shared.ResourceEnergyProduction: 3})
	energyProdBefore := target.Resources().Production().GetAmount(shared.ResourceEnergyProduction)

	reduceOutput := shared.NewProductionCondition(shared.ResourceEnergyProduction, -1, "any-player")

	applyOutputsWithOptions(t, p, testGame, applyOptions{targetPlayerID: targetPlayerID, cardRegistry: cardRegistry}, reduceOutput)

	energyProdAfter := target.Resources().Production().GetAmount(shared.ResourceEnergyProduction)
	testutil.AssertEqual(t, energyProdBefore-1, energyProdAfter, "Target energy production should decrease by 1")
}

func TestProduction_AllProductionTypes(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)

	outputs := []shared.BehaviorCondition{
		shared.NewProductionCondition(shared.ResourceCreditProduction, 1, "self-player"),
		shared.NewProductionCondition(shared.ResourceSteelProduction, 1, "self-player"),
		shared.NewProductionCondition(shared.ResourceTitaniumProduction, 1, "self-player"),
		shared.NewProductionCondition(shared.ResourcePlantProduction, 1, "self-player"),
		shared.NewProductionCondition(shared.ResourceEnergyProduction, 1, "self-player"),
		shared.NewProductionCondition(shared.ResourceHeatProduction, 1, "self-player"),
	}

	prodBefore := p.Resources().Production()

	applyOutputs(t, p, testGame, cardRegistry, outputs...)

	assertProduction(t, p, map[shared.ResourceType]int{
		shared.ResourceCreditProduction:   prodBefore.Credits + 1,
		shared.ResourceSteelProduction:    prodBefore.Steel + 1,
		shared.ResourceTitaniumProduction: prodBefore.Titanium + 1,
		shared.ResourcePlantProduction:    prodBefore.Plants + 1,
		shared.ResourceEnergyProduction:   prodBefore.Energy + 1,
		shared.ResourceHeatProduction:     prodBefore.Heat + 1,
	})
}

func TestProduction_ReduceToZero(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().AddProduction(map[shared.ResourceType]int{shared.ResourceEnergyProduction: 3})

	output := shared.NewProductionCondition(shared.ResourceEnergyProduction, -3, "self-player")
	applyOutputs(t, p, testGame, cardRegistry, output)

	assertProduction(t, p, map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 0,
	})
}

func TestProduction_ReduceBelowMinimumClamps(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().AddProduction(map[shared.ResourceType]int{shared.ResourceSteelProduction: 1})

	output := shared.NewProductionCondition(shared.ResourceSteelProduction, -5, "self-player")
	applyOutputs(t, p, testGame, cardRegistry, output)

	assertProduction(t, p, map[shared.ResourceType]int{
		shared.ResourceSteelProduction: 0,
	})
}
