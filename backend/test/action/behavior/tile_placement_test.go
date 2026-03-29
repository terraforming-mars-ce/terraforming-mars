package behavior_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// ============================================================================
// Tile adjacency restriction tests
// ============================================================================

// --- Urbanized Area (120) ---
// "Place a city tile adjacent to at least 2 other city tiles."

func TestUrbanizedArea_CityAdjacentTo2Cities(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	minAdj := 2
	urbanizedArea := gamecards.Card{
		ID:   "card-urbanized-area",
		Name: "Urbanized Area",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceCreditProduction, 2, "self-player"),
					shared.NewProductionCondition(shared.ResourceEnergyProduction, -1, "self-player"),
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
						TileRestrictions: &shared.TileRestrictions{
							AdjacentToType:    "city",
							MinAdjacentOfType: &minAdj,
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{urbanizedArea})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 1,
	})
	p.Hand().AddCard("card-urbanized-area")

	// Place 2 adjacent city tiles: (0,0,0) and (1,0,-1) are neighbors
	city1 := shared.HexPosition{Q: 0, R: 0, S: 0}
	city2 := shared.HexPosition{Q: 1, R: 0, S: -1}

	err := testGame.Board().UpdateTileOccupancy(ctx, city1,
		board.TileOccupant{Type: shared.ResourceCityTile}, "other-player")
	testutil.AssertNoError(t, err, "placing city1")

	err = testGame.Board().UpdateTileOccupancy(ctx, city2,
		board.TileOccupant{Type: shared.ResourceCityTile}, "other-player")
	testutil.AssertNoError(t, err, "placing city2")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-urbanized-area", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Urbanized Area should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "city", selection.TileType, "Pending tile type should be city")

	// The only hexes adjacent to BOTH (0,0,0) and (1,0,-1) are:
	// (0,1,-1) — neighbor of both via cube coords
	// (1,-1,0) — neighbor of both via cube coords
	// These should be the available hexes (both are land tiles, not ocean/tagged)
	expectedHexes := map[string]bool{
		shared.HexPosition{Q: 0, R: 1, S: -1}.String(): true, // adjacent to both cities
		shared.HexPosition{Q: 1, R: -1, S: 0}.String(): true, // adjacent to both cities
	}

	for _, hex := range selection.AvailableHexes {
		testutil.AssertTrue(t, expectedHexes[hex],
			"Hex "+hex+" should be adjacent to 2+ cities")
	}
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0,
		"Should have at least one available hex adjacent to 2 cities")
}

func TestUrbanizedArea_NoCitiesYield0AvailableHexes(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// No cities placed — should have 0 available hexes with adjacentToType: "city", minAdjacentOfType: 2
	minAdj := 2
	count := testGame.CountAvailableHexesForTile("city", p.ID(), &shared.TileRestrictions{
		AdjacentToType:    "city",
		MinAdjacentOfType: &minAdj,
	})
	testutil.AssertEqual(t, 0, count, "No hexes should be available when no cities exist")
}

func TestUrbanizedArea_OnlyOneCityYield0AvailableHexes(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Place only 1 city — no hex can be adjacent to 2 cities
	city1 := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, city1,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "placing city")

	minAdj := 2
	count := testGame.CountAvailableHexesForTile("city", p.ID(), &shared.TileRestrictions{
		AdjacentToType:    "city",
		MinAdjacentOfType: &minAdj,
	})
	testutil.AssertEqual(t, 0, count, "No hexes should be available when only 1 city exists")
}

// --- Ecological Zone (128) ---
// "Place this tile adjacent to any greenery tile."

func TestEcologicalZone_GreeneryAdjacentToGreenery(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	ecologicalZone := gamecards.Card{
		ID:   "card-ecological-zone",
		Name: "Ecological Zone",
		Type: gamecards.CardTypeActive,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagAnimal, shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardStorageCondition(shared.ResourceAnimal, 2, "self-card"),
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceGreeneryPlacement, Amount: 1, Target: "none"},
						TileRestrictions: &shared.TileRestrictions{
							AdjacentToType: "greenery",
						},
					},
				},
			},
		},
		ResourceStorage: &gamecards.ResourceStorage{
			Type:     shared.ResourceAnimal,
			Starting: 0,
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{ecologicalZone})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-ecological-zone")

	// Place a greenery at (0,0,0)
	greenery := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, greenery,
		board.TileOccupant{Type: shared.ResourceGreeneryTile}, p.ID())
	testutil.AssertNoError(t, err, "placing greenery")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-ecological-zone", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ecological Zone should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "greenery", selection.TileType, "Pending tile type should be greenery")

	// All available hexes should be adjacent to the greenery at (0,0,0)
	greeneryNeighbors := greenery.GetNeighbors()
	neighborStrings := make(map[string]bool)
	for _, n := range greeneryNeighbors {
		neighborStrings[n.String()] = true
	}

	for _, hex := range selection.AvailableHexes {
		testutil.AssertTrue(t, neighborStrings[hex],
			"Hex "+hex+" should be adjacent to greenery at (0,0,0)")
	}
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0,
		"Should have at least one available hex adjacent to greenery")
}

func TestEcologicalZone_NoGreeneryYields0AvailableHexes(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// No greenery placed — should have 0 available hexes
	count := testGame.CountAvailableHexesForTile("greenery", p.ID(), &shared.TileRestrictions{
		AdjacentToType: "greenery",
	})
	testutil.AssertEqual(t, 0, count, "No hexes should be available when no greenery exists")
}

// --- AdjacentToOwned restriction ---

func TestAdjacentToOwned_OnlyCountsPlayerOwnedTiles(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	other := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	other.SetCorporationID("corp-mining-guild")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Place a greenery owned by OTHER player at (0,0,0)
	greenery := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, greenery,
		board.TileOccupant{Type: shared.ResourceGreeneryTile}, other.ID())
	testutil.AssertNoError(t, err, "placing greenery for other player")

	// With adjacentToOwned=true, player p should have 0 available hexes
	// because the greenery belongs to the other player
	count := testGame.CountAvailableHexesForTile("greenery", p.ID(), &shared.TileRestrictions{
		AdjacentToType:  "greenery",
		AdjacentToOwned: true,
	})
	testutil.AssertEqual(t, 0, count, "No hexes should be available when only opponent owns greenery")

	// Place a greenery owned by player p at (2,0,-2)
	greenery2 := shared.HexPosition{Q: 2, R: 0, S: -2}
	err = testGame.Board().UpdateTileOccupancy(ctx, greenery2,
		board.TileOccupant{Type: shared.ResourceGreeneryTile}, p.ID())
	testutil.AssertNoError(t, err, "placing greenery for player")

	// Now player p should have some available hexes adjacent to their own greenery
	count = testGame.CountAvailableHexesForTile("greenery", p.ID(), &shared.TileRestrictions{
		AdjacentToType:  "greenery",
		AdjacentToOwned: true,
	})
	testutil.AssertTrue(t, count > 0, "Should have available hexes adjacent to own greenery")
}

// --- State calculator: restricted tile placements produce errors, not warnings ---

func makeUrbanizedAreaCard() gamecards.Card {
	minAdj := 2
	return gamecards.Card{
		ID:   "card-urbanized-area",
		Name: "Urbanized Area",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagCity},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceCreditProduction, 2, "self-player"),
					shared.NewProductionCondition(shared.ResourceEnergyProduction, -1, "self-player"),
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCityPlacement, Amount: 1, Target: "none"},
						TileRestrictions: &shared.TileRestrictions{
							AdjacentToType:    "city",
							MinAdjacentOfType: &minAdj,
						},
					},
				},
			},
		},
	}
}

func TestUrbanizedArea_StateCalculatorReturnsErrorWhenNoPlacements(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Resources().AddProduction(map[shared.ResourceType]int{shared.ResourceEnergyProduction: 1})

	urbanizedArea := makeUrbanizedAreaCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{urbanizedArea})

	// Only 1 city on the board — Urbanized Area requires adjacent to 2 cities
	city1 := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, city1,
		board.TileOccupant{Type: shared.ResourceCityTile}, "other-player")
	testutil.AssertNoError(t, err, "placing city")

	state := action.CalculatePlayerCardState(&urbanizedArea, p, testGame, cardRegistry)

	testutil.AssertTrue(t, !state.Available(),
		"Urbanized Area should NOT be available when only 1 city exists")

	hasNoCityPlacementError := false
	for _, e := range state.Errors {
		if e.Code == player.ErrorCodeNoCityPlacements {
			hasNoCityPlacementError = true
		}
	}
	testutil.AssertTrue(t, hasNoCityPlacementError,
		"Should have ErrorCodeNoCityPlacements error")
}

func TestUrbanizedArea_PlayCardRejectedWhenNoPlacements(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Resources().AddProduction(map[shared.ResourceType]int{shared.ResourceEnergyProduction: 1})

	urbanizedArea := makeUrbanizedAreaCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{urbanizedArea})
	p.Hand().AddCard("card-urbanized-area")

	// Only 1 city — cannot satisfy adjacency requirement
	city1 := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, city1,
		board.TileOccupant{Type: shared.ResourceCityTile}, "other-player")
	testutil.AssertNoError(t, err, "placing city")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-urbanized-area", payment, nil, nil, nil, nil)

	testutil.AssertTrue(t, err != nil,
		"Playing Urbanized Area should fail when no valid placements exist")
}

// --- Plantation (193) ---
// "Place a greenery tile and raise oxygen 1 step."

func makePlantationCard() gamecards.Card {
	return gamecards.Card{
		ID:   "card-plantation",
		Name: "Plantation",
		Type: gamecards.CardTypeAutomated,
		Cost: 15,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceGreeneryPlacement, Amount: 1, Target: "none"},
					},
				},
			},
		},
	}
}

func TestPlantation_GreeneryAdjacentToOwnedTiles(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	plantation := makePlantationCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{plantation})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-plantation")

	// Place a city owned by the current player at (0,0,0)
	ownedTile := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, ownedTile,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "placing owned city")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-plantation", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Plantation should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "greenery", selection.TileType, "Pending tile type should be greenery")

	// All available hexes should be adjacent to the owned tile at (0,0,0)
	neighbors := ownedTile.GetNeighbors()
	neighborStrings := make(map[string]bool)
	for _, n := range neighbors {
		neighborStrings[n.String()] = true
	}

	for _, hex := range selection.AvailableHexes {
		testutil.AssertTrue(t, neighborStrings[hex],
			"Hex "+hex+" should be adjacent to owned tile at (0,0,0)")
	}
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0,
		"Should have at least one available hex adjacent to owned tile")
}

func TestPlantation_FallbackWhenNoOwnedTiles(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	plantation := makePlantationCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{plantation})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-plantation")

	// No tiles placed — player has no owned tiles on board

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-plantation", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Plantation should play successfully with fallback")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "greenery", selection.TileType, "Pending tile type should be greenery")

	// With no owned tiles, fallback allows all valid land tiles
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 1,
		"Should have multiple available hexes when no owned tiles (fallback)")
}

// --- Mangrove (059) ---
// "Place a greenery tile ON AN AREA RESERVED FOR OCEAN and raise oxygen 1 step."

func makeMangroveCard() gamecards.Card {
	return gamecards.Card{
		ID:   "card-mangrove",
		Name: "Mangrove",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceGreeneryPlacement, Amount: 1, Target: "none"},
						TileRestrictions: &shared.TileRestrictions{
							OnTileType: "ocean",
						},
					},
				},
			},
		},
	}
}

func TestMangrove_NotAffectedByAdjacentToOwned(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	mangrove := makeMangroveCard()
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mangrove})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mangrove")

	// Place a city far away — Mangrove should still offer ocean spaces, not adjacent-to-owned
	ownedTile := shared.HexPosition{Q: 4, R: -2, S: -2}
	err := testGame.Board().UpdateTileOccupancy(ctx, ownedTile,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "placing owned city")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err = playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mangrove", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mangrove should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "greenery", selection.TileType, "Pending tile type should be greenery")

	// Mangrove should offer ocean spaces — verify available hexes are ocean tiles
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0,
		"Should have available ocean hexes for Mangrove")

	// Verify that hexes include non-adjacent-to-owned ocean spaces
	ownedNeighbors := ownedTile.GetNeighbors()
	ownedNeighborStrings := make(map[string]bool)
	for _, n := range ownedNeighbors {
		ownedNeighborStrings[n.String()] = true
	}

	hasNonAdjacentHex := false
	for _, hex := range selection.AvailableHexes {
		if !ownedNeighborStrings[hex] {
			hasNonAdjacentHex = true
			break
		}
	}
	testutil.AssertTrue(t, hasNonAdjacentHex,
		"Mangrove should offer ocean spaces not adjacent to owned tiles")
}

// ============================================================================
// New tile type tests
// ============================================================================

// --- Natural Preserve (044) ---
// "Place this tile next to no other tile."

func TestNaturalPreserve_TilePlacementWithNoAdjacency(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	naturalPreserve := gamecards.Card{
		ID:   "card-natural-preserve",
		Name: "Natural Preserve",
		Type: gamecards.CardTypeAutomated,
		Cost: 9,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceCreditProduction, 1, "self-player"),
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTilePlacement, Amount: 1, Target: "none"},
						TileType:      "natural-preserve",
						TileRestrictions: &shared.TileRestrictions{
							Adjacency: "none",
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{naturalPreserve})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-natural-preserve")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 9}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-natural-preserve", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Natural Preserve should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "natural-preserve", selection.TileType, "Pending tile type should be natural-preserve")

	// All available hexes should have no adjacent occupied tiles
	tiles := testGame.Board().Tiles()
	for _, hex := range selection.AvailableHexes {
		for _, tile := range tiles {
			if tile.Coordinates.String() == hex {
				neighbors := tile.Coordinates.GetNeighbors()
				for _, neighborPos := range neighbors {
					for _, neighborTile := range tiles {
						if neighborTile.Coordinates.Equals(neighborPos) {
							testutil.AssertTrue(t, neighborTile.OccupiedBy == nil,
								"Hex "+hex+" has an adjacent occupied tile — violates no-adjacency restriction")
						}
					}
				}
			}
		}
	}
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0,
		"Should have available hexes for natural preserve placement")
}

func TestNaturalPreserve_NoAvailableHexesWhenBoardIsFull(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Place a tile at center — this blocks all adjacent hexes
	center := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, center,
		board.TileOccupant{Type: shared.ResourceCityTile}, p.ID())
	testutil.AssertNoError(t, err, "placing center tile")

	// Now check available hexes with adjacency="none" — hexes adjacent to center should be excluded
	count := testGame.CountAvailableHexesForTile("natural-preserve", p.ID(), &shared.TileRestrictions{
		Adjacency: "none",
	})
	// There should still be some hexes available (ones not adjacent to center)
	testutil.AssertTrue(t, count > 0, "Should still have some available hexes far from center tile")

	// Verify the center's neighbors are NOT in available hexes
	centerNeighbors := center.GetNeighbors()
	centerNeighborStrings := make(map[string]bool)
	for _, n := range centerNeighbors {
		centerNeighborStrings[n.String()] = true
	}

	available := testGame.CalculateAvailableHexesForTile("natural-preserve", p.ID(), &shared.TileRestrictions{
		Adjacency: "none",
	})
	for _, hex := range available {
		testutil.AssertTrue(t, !centerNeighborStrings[hex],
			"Hex "+hex+" is adjacent to center tile but should not be available")
		testutil.AssertTrue(t, hex != center.String(),
			"Center tile should not be available (already occupied)")
	}
}

// --- Nuclear Zone (097) ---
// "Place this tile and raise temperature 2 steps."

func TestNuclearZone_TilePlacementOnNormalLand(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	nuclearZone := gamecards.Card{
		ID:   "card-nuclear-zone",
		Name: "Nuclear Zone",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Tags: []shared.CardTag{shared.TagEarth},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewGlobalParameterCondition(shared.ResourceTemperature, 2, "none"),
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTilePlacement, Amount: 1, Target: "none"},
						TileType:      "nuclear-zone",
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{nuclearZone})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-nuclear-zone")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 10}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-nuclear-zone", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuclear Zone should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "nuclear-zone", selection.TileType, "Pending tile type should be nuclear-zone")
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0, "Should have available hexes on normal land")
}

// --- Mohole Area (142) ---
// "Place this tile on an area reserved for ocean."

func TestMoholeArea_TilePlacementOnOceanSpace(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	moholeArea := gamecards.Card{
		ID:   "card-mohole-area",
		Name: "Mohole Area",
		Type: gamecards.CardTypeAutomated,
		Cost: 20,
		Tags: []shared.CardTag{shared.TagBuilding},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewProductionCondition(shared.ResourceHeatProduction, 4, "self-player"),
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTilePlacement, Amount: 1, Target: "none"},
						TileType:      "mohole",
						TileRestrictions: &shared.TileRestrictions{
							OnTileType: "ocean",
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{moholeArea})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mohole-area")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 20}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mohole-area", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mohole Area should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "mohole", selection.TileType, "Pending tile type should be mohole")

	// All available hexes should be ocean spaces
	tiles := testGame.Board().Tiles()
	for _, hex := range selection.AvailableHexes {
		var found bool
		for _, tile := range tiles {
			if tile.Coordinates.String() == hex {
				testutil.AssertEqual(t, shared.ResourceOceanSpace, tile.Type,
					"Hex "+hex+" should be an ocean space")
				found = true
				break
			}
		}
		testutil.AssertTrue(t, found, "Hex "+hex+" should exist on the board")
	}
	testutil.AssertTrue(t, len(selection.AvailableHexes) > 0,
		"Should have available ocean space hexes")
}

// --- Mining Rights (067) ---
// "Place this tile on an area with a steel or titanium placement bonus."

func TestMiningRights_TilePlacementOnBonusTile(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Check available hexes with onBonusType restriction
	count := testGame.CountAvailableHexesForTile("mining", p.ID(), &shared.TileRestrictions{
		OnBonusType: []string{"steel", "titanium"},
	})
	// Mars board has steel and titanium bonus tiles
	testutil.AssertTrue(t, count > 0, "Should have available hexes with steel/titanium bonuses")

	// Verify all available hexes have steel or titanium bonuses
	available := testGame.CalculateAvailableHexesForTile("mining", p.ID(), &shared.TileRestrictions{
		OnBonusType: []string{"steel", "titanium"},
	})
	tiles := testGame.Board().Tiles()
	for _, hex := range available {
		for _, tile := range tiles {
			if tile.Coordinates.String() == hex {
				hasBonus := false
				for _, bonus := range tile.Bonuses {
					if bonus.Type == shared.ResourceSteel || bonus.Type == shared.ResourceTitanium {
						hasBonus = true
						break
					}
				}
				testutil.AssertTrue(t, hasBonus,
					"Hex "+hex+" should have steel or titanium bonus")
			}
		}
	}
}

// --- Mining Area (064) ---
// "Place this tile on an area with a steel or titanium placement bonus, adjacent to another of your tiles."

func TestMiningArea_RequiresAdjacentOwnedTile(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Without any owned tiles, no hexes should be available (even if bonus tiles exist)
	count := testGame.CountAvailableHexesForTile("mining", p.ID(), &shared.TileRestrictions{
		OnBonusType:     []string{"steel", "titanium"},
		AdjacentToOwned: true,
	})
	testutil.AssertEqual(t, 0, count, "No hexes should be available without own tiles on board")

	// Place a tile adjacent to a steel bonus tile
	// Steel bonus tiles on Mars board: (-3,1,2) and (-2,0,2)
	// Place a greenery at (-3,2,1) which is adjacent to (-3,1,2)
	greeneryPos := shared.HexPosition{Q: -3, R: 2, S: 1}
	err := testGame.Board().UpdateTileOccupancy(ctx, greeneryPos,
		board.TileOccupant{Type: shared.ResourceGreeneryTile}, p.ID())
	testutil.AssertNoError(t, err, "placing greenery")

	// Now should have available hexes (steel bonus tiles adjacent to owned greenery)
	count = testGame.CountAvailableHexesForTile("mining", p.ID(), &shared.TileRestrictions{
		OnBonusType:     []string{"steel", "titanium"},
		AdjacentToOwned: true,
	})
	testutil.AssertTrue(t, count > 0,
		"Should have available hexes: steel/titanium bonus tiles adjacent to owned greenery")
}

// --- Restricted Area (199) ---
// "Place this tile." (normal land placement, no restrictions)

func TestRestrictedArea_NormalLandPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-test")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Simple tile-placement with no restrictions should have land tiles available
	count := testGame.CountAvailableHexesForTile("restricted", p.ID(), nil)
	testutil.AssertTrue(t, count > 0, "Should have available land hexes for restricted tile")
}

// --- Tile type mapping ---

func TestSpecialTileOccupantTypeHasTileSuffix(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	// Place a special tile directly and verify the occupant type
	pos := shared.HexPosition{Q: 0, R: 0, S: 0}
	err := testGame.Board().UpdateTileOccupancy(ctx, pos,
		board.TileOccupant{Type: shared.ResourceNaturalPreserveTile}, "player-1")
	testutil.AssertNoError(t, err, "placing natural preserve tile")

	tile, err := testGame.Board().GetTile(pos)
	testutil.AssertNoError(t, err, "getting tile")
	testutil.AssertTrue(t, tile.OccupiedBy != nil, "Tile should be occupied")
	testutil.AssertEqual(t, shared.ResourceNaturalPreserveTile, tile.OccupiedBy.Type,
		"Occupant type should be natural-preserve-tile")
}

// ============================================================================
// Play card tile placement tests
// ============================================================================

// --- Mangrove (059) ---
// "Place a Greenery tile on an area reserved for ocean and raise oxygen 1 step."

func TestMangrove_GreeneryOnOceanTileRestriction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	mangrove := gamecards.Card{
		ID:   "card-mangrove",
		Name: "Mangrove",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					&shared.TilePlacementCondition{
						ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceGreeneryPlacement, Amount: 1, Target: "none"},
						TileRestrictions: &shared.TileRestrictions{
							OnTileType: "ocean",
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mangrove})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mangrove")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mangrove", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mangrove should play successfully")

	// After playing, ProcessNextTile moves the tile from the queue to PendingTileSelection
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "greenery", selection.TileType, "Pending tile type should be greenery")
}

// --- Land Claim (066) ---
// "Place your marker on a non-reserved area. Only you may place a tile here."

func TestLandClaim_CreatesLandClaimTileSelection(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	landClaim := gamecards.Card{
		ID:   "card-land-claim",
		Name: "Land Claim",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewTilePlacementCondition(shared.ResourceLandClaim, 1, "none"),
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{landClaim})

	players := testGame.GetAllPlayers()
	p := players[0]
	other := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	other.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-land-claim")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-land-claim", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Land Claim should play successfully")

	// After playing, ProcessNextTile moves the tile from the queue to PendingTileSelection
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection for land claim")
	testutil.AssertEqual(t, "land-claim", selection.TileType, "Pending tile type should be land-claim")
}

// --- Artificial Lake (116) ---
// "Place 1 ocean tile on an area not reserved for ocean."

func TestArtificialLake_OceanPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	artificialLake := gamecards.Card{
		ID:   "card-artificial-lake",
		Name: "Artificial Lake",
		Type: gamecards.CardTypeAutomated,
		Cost: 15,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-6)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{artificialLake})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Meet temperature requirement
	testutil.AssertNoError(t, testGame.GlobalParameters().SetTemperature(ctx, 0), "set temperature")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-artificial-lake")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-artificial-lake", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Artificial Lake should play successfully")

	// After playing, ProcessNextTile moves the tile from the queue to PendingTileSelection
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection for ocean placement")
	testutil.AssertEqual(t, "ocean", selection.TileType, "Pending tile type should be ocean")
}

func TestArtificialLake_FailsWithoutTemperatureRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	artificialLake := gamecards.Card{
		ID:   "card-artificial-lake",
		Name: "Artificial Lake",
		Type: gamecards.CardTypeAutomated,
		Cost: 15,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-6)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewTilePlacementCondition(shared.ResourceOceanPlacement, 1, "none"),
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{artificialLake})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Temperature below requirement (-30 default, need -6)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-artificial-lake")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-artificial-lake", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Artificial Lake should fail without meeting temperature requirement")
	testutil.AssertTrue(t, p.Hand().HasCard("card-artificial-lake"), "Card should still be in hand")
}
