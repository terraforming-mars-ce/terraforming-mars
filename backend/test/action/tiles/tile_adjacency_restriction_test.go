package tiles_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

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
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{
						ResourceType: shared.ResourceCityPlacement,
						Amount:       1,
						Target:       "none",
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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAnimal, Amount: 2, Target: "self-card"},
					{
						ResourceType: shared.ResourceGreeneryPlacement,
						Amount:       1,
						Target:       "none",
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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 2, Target: "self-player"},
					{ResourceType: shared.ResourceEnergyProduction, Amount: -1, Target: "self-player"},
					{
						ResourceType: shared.ResourceCityPlacement,
						Amount:       1,
						Target:       "none",
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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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

	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

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
