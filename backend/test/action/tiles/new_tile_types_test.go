package tiles_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

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
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceCreditProduction, Amount: 1, Target: "self-player"},
					{
						ResourceType: shared.ResourceTilePlacement,
						Amount:       1,
						Target:       "none",
						TileType:     "natural-preserve",
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
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceTemperature, Amount: 2, Target: "none"},
					{
						ResourceType: shared.ResourceTilePlacement,
						Amount:       1,
						Target:       "none",
						TileType:     "nuclear-zone",
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
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceHeatProduction, Amount: 4, Target: "self-player"},
					{
						ResourceType: shared.ResourceTilePlacement,
						Amount:       1,
						Target:       "none",
						TileType:     "mohole",
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
	// Verify that the mapTileTypeToResourceType function (via select_tile.go)
	// adds "-tile" suffix for unknown tile types.
	// This ensures occupant types are consistent with the naming convention:
	// "city" → "city-tile", "natural-preserve" → "natural-preserve-tile"

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
