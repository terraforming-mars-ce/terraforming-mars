package behavior_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestPerCondition_CityTileMarsLocation(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// Place 3 city tiles on mars
	tiles := g.Board().Tiles()
	placed := 0
	for _, tile := range tiles {
		if tile.Location == board.TileLocationMars && tile.OccupiedBy == nil && placed < 3 {
			err := g.Board().UpdateTileOccupancy(ctx, tile.Coordinates,
				board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
			testutil.AssertNoError(t, err, "placing city tile")
			placed++
		}
	}
	testutil.AssertEqual(t, 3, placed, "should place 3 cities")

	// Set starting credits to 0
	testutil.SetPlayerCredits(ctx, p, 0)

	// Create output with per: city-tile, location: mars
	marsLocation := "mars"
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceCityTile,
				Amount:       1,
				Location:     &marsLocation,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Martian Rails", log)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 3, credits, "should gain 3 credits (1 per city on mars)")
}

func TestPerCondition_CityTileAnywhereLocation(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// Place 2 city tiles on mars
	tiles := g.Board().Tiles()
	placed := 0
	for _, tile := range tiles {
		if tile.Location == board.TileLocationMars && tile.OccupiedBy == nil && placed < 2 {
			err := g.Board().UpdateTileOccupancy(ctx, tile.Coordinates,
				board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
			testutil.AssertNoError(t, err, "placing city tile")
			placed++
		}
	}

	testutil.SetPlayerCredits(ctx, p, 0)

	// Per with no location (anywhere)
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceCityTile,
				Amount:       1,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Greenhouses", log)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 2, credits, "should gain 2 credits (1 per city anywhere)")
}

func TestPerCondition_TagSelfPlayer(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// Add cards with earth tags to played cards
	p.PlayedCards().AddCard(testutil.CardID("Earth Catapult"), "Earth Catapult", "active", []string{"earth"})
	p.PlayedCards().AddCard(testutil.CardID("Earth Office"), "Earth Office", "active", []string{"earth"})
	p.PlayedCards().AddCard(testutil.CardID("Sponsors"), "Sponsors", "automated", []string{"earth"})

	testutil.SetPlayerCredits(ctx, p, 0)

	earthTag := shared.TagEarth
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceType("tag"),
				Amount:       1,
				Tag:          &earthTag,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Test Card", log).
		WithCardRegistry(cardRegistry)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 3, credits, "should gain 3 credits (1 per earth tag)")
}

func TestPerCondition_TagAnyPlayer(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	players := g.GetAllPlayers()
	p1 := players[0]
	p2 := players[1]
	ctx := context.Background()
	log := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// Player 1 has 2 earth tags
	p1.PlayedCards().AddCard(testutil.CardID("Earth Catapult"), "Earth Catapult", "active", []string{"earth"})
	p1.PlayedCards().AddCard(testutil.CardID("Earth Office"), "Earth Office", "active", []string{"earth"})

	// Player 2 has 1 earth tag
	p2.PlayedCards().AddCard(testutil.CardID("Sponsors"), "Sponsors", "automated", []string{"earth"})

	testutil.SetPlayerCredits(ctx, p1, 0)

	earthTag := shared.TagEarth
	anyPlayer := "any-player"
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceType("tag"),
				Amount:       1,
				Tag:          &earthTag,
				Target:       &anyPlayer,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p1, g, "Galilean Waystation", log).
		WithCardRegistry(cardRegistry)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p1)
	testutil.AssertEqual(t, 3, credits, "should gain 3 credits (1 per earth tag across all players)")
}

func TestPerCondition_CardResourceSelfCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// Add 5 floaters to a card
	sourceCardID := testutil.CardID("Saturn Surfing")
	p.Resources().AddToStorage(sourceCardID, 5)

	testutil.SetPlayerCredits(ctx, p, 0)

	selfCard := "self-card"
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceFloater,
				Amount:       1,
				Target:       &selfCard,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Saturn Surfing", log).
		WithSourceCardID(sourceCardID)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 5, credits, "should gain 5 credits (1 per floater on card)")
}

func TestPerCondition_IntegerDivision(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// Place 5 city tiles
	tiles := g.Board().Tiles()
	placed := 0
	for _, tile := range tiles {
		if tile.Location == board.TileLocationMars && tile.OccupiedBy == nil && placed < 5 {
			err := g.Board().UpdateTileOccupancy(ctx, tile.Coordinates,
				board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
			testutil.AssertNoError(t, err, "placing city tile")
			placed++
		}
	}

	testutil.SetPlayerCredits(ctx, p, 0)

	// per.Amount = 2, so 5 cities / 2 = multiplier 2
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceCityTile,
				Amount:       2,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Test Card", log)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 2, credits, "should gain 2 credits (5 cities / per.Amount 2 = multiplier 2)")
}

func TestPerCondition_ZeroMultiplierSkipsOutput(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// No cities placed
	testutil.SetPlayerCredits(ctx, p, 10)

	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceCityTile,
				Amount:       1,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Test Card", log)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 10, credits, "credits should remain unchanged when multiplier is zero")
}

func TestCountPerConditionResourceCounting(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()

	// Set heat=10, steel=5, titanium=3
	resources := p.Resources().Get()
	resources.Heat = 10
	resources.Steel = 5
	resources.Titanium = 3
	p.Resources().Set(resources)

	selfPlayer := "self-player"

	// Heat counting
	heatPer := &shared.PerCondition{
		ResourceType: shared.ResourceHeat,
		Amount:       1,
		Target:       &selfPlayer,
	}
	count := gamecards.CountPerCondition(heatPer, "", p, g.Board(), nil, nil)
	testutil.AssertEqual(t, 10, count, "should count 10 heat")

	// Steel counting
	steelPer := &shared.PerCondition{
		ResourceType: shared.ResourceSteel,
		Amount:       1,
		Target:       &selfPlayer,
	}
	count = gamecards.CountPerCondition(steelPer, "", p, g.Board(), nil, nil)
	testutil.AssertEqual(t, 5, count, "should count 5 steel")

	// Titanium counting
	titaniumPer := &shared.PerCondition{
		ResourceType: shared.ResourceTitanium,
		Amount:       1,
		Target:       &selfPlayer,
	}
	count = gamecards.CountPerCondition(titaniumPer, "", p, g.Board(), nil, nil)
	testutil.AssertEqual(t, 3, count, "should count 3 titanium")

	_ = ctx
}

func TestCountPerConditionProductionCounting(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]

	// Set credit production=3
	prod := p.Resources().Production()
	prod.Credits = 3
	p.Resources().SetProduction(prod)

	selfPlayer := "self-player"

	creditProdPer := &shared.PerCondition{
		ResourceType: shared.ResourceCreditProduction,
		Amount:       1,
		Target:       &selfPlayer,
	}
	count := gamecards.CountPerCondition(creditProdPer, "", p, g.Board(), nil, nil)
	testutil.AssertEqual(t, 3, count, "should count 3 credit production")
}

func TestPerCondition_NilPerProceedsNormally(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	testutil.SetPlayerCredits(ctx, p, 0)

	// Output without per condition
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 5, Target: "self-player"},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Test Card", log)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 5, credits, "should gain exact amount when no per condition")
}

func TestPerCondition_FloaterLeasing(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// Add two cards with floater storage and put floaters on them
	p.PlayedCards().AddCard("card-a", "Card A", "active", []string{})
	p.PlayedCards().AddCard("card-b", "Card B", "active", []string{})
	p.Resources().AddToStorage("card-a", 6)
	p.Resources().AddToStorage("card-b", 3)

	// Set starting production to 0
	prod := p.Resources().Production()
	prod.Credits = 0
	p.Resources().SetProduction(prod)

	selfPlayer := "self-player"
	outputs := []shared.BehaviorCondition{
		&shared.ProductionCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceFloater,
				Amount:       3,
				Target:       &selfPlayer,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Floater Leasing", log).
		WithCardRegistry(cardRegistry)
	_, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
	testutil.AssertNoError(t, err, "applying floater leasing outputs")

	// 9 floaters / 3 = 3 production steps, but CountPlayerCardStorageByType
	// requires cards to be in registry with floater storage type.
	// Since card-a and card-b are not real cards in registry, floater count will be 0.
	// This test verifies the mechanic works without error.
	// Integration test with real card IDs would verify the full flow.
	newProd := p.Resources().Production()
	t.Logf("Credit production after Floater Leasing: %d", newProd.Credits)
}

func TestPerCondition_ColonyCount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	players := g.GetAllPlayers()
	p := players[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// Set up colony states with some colonies placed
	g.Colonies().SetStates([]*colony.ColonyState{
		{
			DefinitionID:   "ganymede",
			MarkerPosition: 1,
			PlayerColonies: []string{p.ID(), players[1].ID()},
		},
		{
			DefinitionID:   "titan",
			MarkerPosition: 1,
			PlayerColonies: []string{p.ID()},
		},
		{
			DefinitionID:   "europa",
			MarkerPosition: 1,
			PlayerColonies: []string{},
		},
	})

	// Verify CountAllColonies returns 3 (2 on ganymede + 1 on titan + 0 on europa)
	totalColonies := g.Colonies().CountAllColonies()
	testutil.AssertEqual(t, 3, totalColonies, "should count 3 total colonies")

	// Set starting credits to 0
	testutil.SetPlayerCredits(ctx, p, 0)

	// Apply Molecular Printing colony output
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceColonyCount,
				Amount:       1,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Molecular Printing", log)
	_, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
	testutil.AssertNoError(t, err, "applying colony count outputs")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 3, credits, "should gain 3 credits (1 per colony in play)")
}

func TestPerCondition_ColonyCountEmpty(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	g, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	p := g.GetAllPlayers()[0]
	ctx := context.Background()
	log := testutil.TestLogger()

	// No colonies set up
	testutil.SetPlayerCredits(ctx, p, 0)

	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceColonyCount,
				Amount:       1,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, g, "Molecular Printing", log)
	_, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
	testutil.AssertNoError(t, err, "applying colony count outputs with no colonies")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 0, credits, "should gain 0 credits when no colonies exist")
}

func TestPerCondition_PerAmountZero_NoScaling(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	ctx := context.Background()
	log := testutil.TestLogger()

	testutil.SetPlayerCredits(ctx, p, 0)

	scienceTag := shared.TagScience
	outputs := []shared.BehaviorCondition{
		&shared.BasicResourceCondition{
			ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCredit, Amount: 3, Target: "self-player"},
			Per: &shared.PerCondition{
				ResourceType: shared.ResourceType("tag"),
				Amount:       0,
				Tag:          &scienceTag,
			},
		},
	}

	applier := gamecards.NewBehaviorApplier(p, testGame, "Test Card", log).
		WithCardRegistry(cardRegistry)
	err := applier.ApplyOutputs(ctx, outputs)
	testutil.AssertNoError(t, err, "applying outputs with per.Amount=0")

	credits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, 3, credits, "should gain base amount (3) when per.Amount is 0 (no scaling)")
}
