package card_packs_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Nuke (EXP004) ---
// "Replace any non-ocean tile with a Nuclear Zone tile."
// Uses tile-replacement:nuclear-zone, -2 VP, targets non-ocean occupied tiles only.

func TestNuke_ReplacesOccupiedTileWithNuclearZone(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 200)

	// Place a city for p2 on a land tile
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	// Verify city is placed
	tile := testutil.GetTileAtHex(t, testGame, cityHex)
	testutil.AssertTrue(t, tile.OccupiedBy != nil, "City should be placed")
	testutil.AssertEqual(t, shared.ResourceCityTile, tile.OccupiedBy.Type, "Should be a city tile")

	// Play Nuke card
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: nukeCard.Cost}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke card should play successfully")

	// Should have pending tile selection for tile-replacement
	pending := testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "tile-replacement:nuclear-zone", pending.TileType, "Pending type should be tile-replacement:nuclear-zone")

	// The city hex should be in available hexes
	testutil.AssertTrue(t, testutil.ContainsHex(pending.AvailableHexes, cityHex),
		"City hex should be available for replacement")

	// Select the city hex for replacement
	selectTile := tileAction.NewSelectTileAction(repo, cardRegistry, stateRepo, logger)
	_, err = selectTile.Execute(ctx, testGame.ID(), p1ID, cityHex)
	testutil.AssertNoError(t, err, "Tile replacement should succeed")

	// Verify the tile is now a nuclear zone
	tile = testutil.GetTileAtHex(t, testGame, cityHex)
	testutil.AssertTrue(t, tile.OccupiedBy != nil, "Tile should still be occupied")
	testutil.AssertEqual(t, shared.ResourceNuclearZoneTile, tile.OccupiedBy.Type, "Should be a nuclear zone tile")

	// Pending tile selection should be cleared
	pending = testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending == nil, "Pending tile selection should be cleared")
}

func TestNuke_DoesNotTargetOceanTiles(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()
	stateRepo := game.NewInMemoryGameStateRepository()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 200)

	// Place an ocean tile
	oceanHex := testutil.FindUnoccupiedOceanHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "ocean", oceanHex)

	// Place a city on a land tile
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	// Play Nuke card
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: nukeCard.Cost}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke card should play successfully")

	pending := testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending != nil, "Should have pending tile selection")

	// Ocean hex should NOT be in available hexes
	testutil.AssertTrue(t, !testutil.ContainsHex(pending.AvailableHexes, oceanHex),
		"Ocean hex should NOT be available for replacement")

	// City hex should be in available hexes
	testutil.AssertTrue(t, testutil.ContainsHex(pending.AvailableHexes, cityHex),
		"City hex should be available for replacement")
}

func TestNuke_HasBuildingAndSpaceTags(t *testing.T) {
	nukeCard := testutil.GetCardByName("Nuke")
	testutil.AssertTrue(t, nukeCard.Tags != nil, "Nuke should have tags")

	hasBuilding := false
	hasSpace := false
	hasEvent := false
	for _, tag := range nukeCard.Tags {
		switch tag {
		case "building":
			hasBuilding = true
		case "space":
			hasSpace = true
		case "event":
			hasEvent = true
		}
	}
	testutil.AssertTrue(t, hasBuilding, "Nuke should have building tag")
	testutil.AssertTrue(t, hasSpace, "Nuke should have space tag")
	testutil.AssertTrue(t, !hasEvent, "Nuke should not have event tag (card type is already event)")
}

func TestNuke_PayWithSteel(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)

	// Give player steel and some credits
	p1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
		shared.ResourceSteel:  5,
	})

	// Place a city so tile-replacement has a target
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	// Pay with 5 steel (5*2=10) + 20 credits = 30
	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: 20, Steel: 5}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke should accept steel payment (building tag)")

	resources := p1.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Steel, "Steel should be spent")
}

func TestNuke_PayWithTitanium(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)

	// Give player titanium and some credits
	p1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   21,
		shared.ResourceTitanium: 3,
	})

	// Place a city so tile-replacement has a target
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	// Pay with 3 titanium (3*3=9) + 21 credits = 30
	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: 21, Titanium: 3}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke should accept titanium payment (space tag)")

	resources := p1.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Titanium, "Titanium should be spent")
}

func TestNuke_PayWithBothSteelAndTitanium(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p1, _ := testGame.GetPlayer(p1ID)

	// Give player both steel and titanium
	p1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   11,
		shared.ResourceSteel:    3,
		shared.ResourceTitanium: 4,
	})

	// Place a city so tile-replacement has a target
	cityHex := testutil.FindUnoccupiedLandHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p2ID, "city", cityHex)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "set turn")

	nukeCard := testutil.GetCardByName("Nuke")
	p1.Hand().AddCard(nukeCard.ID)

	// Pay with 3 steel (3*2=6) + 4 titanium (4*3=12) + 12 credits = 30
	// Wait: 6 + 12 + 11 = 29, need 30. Adjust: 12 credits
	// Actually: 3*2=6 + 4*3=12 + 12 = 30. Give 12 credits.
	p1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 1, // now 12 total
	})

	playCard := cardAction.NewPlayCardAction(repo, cardRegistry, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: 12, Steel: 3, Titanium: 4}
	err = playCard.Execute(ctx, testGame.ID(), p1ID, nukeCard.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Nuke should accept both steel and titanium payment (building + space tags)")

	resources := p1.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Steel, "Steel should be spent")
	testutil.AssertEqual(t, 0, resources.Titanium, "Titanium should be spent")
	testutil.AssertEqual(t, 0, resources.Credits, "Credits should be spent")
}

func TestNuke_HasNegativeVP(t *testing.T) {
	nukeCard := testutil.GetCardByName("Nuke")
	testutil.AssertTrue(t, len(nukeCard.VPConditions) > 0, "Nuke should have VP conditions")
	testutil.AssertEqual(t, -2, nukeCard.VPConditions[0].Amount, "Nuke should have -2 VP")
	testutil.AssertEqual(t, "fixed", nukeCard.VPConditions[0].Condition, "VP condition should be fixed")
}

func TestNuke_UnavailableWhenNoOccupiedNonOceanTiles(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)

	p1ID := playerIDs[0]
	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 200)

	// No tiles placed — tile-replacement should find 0 available hexes
	availableHexes := testGame.CountAvailableHexesForTile("tile-replacement:nuclear-zone", p1ID, nil)
	testutil.AssertEqual(t, 0, availableHexes, "Should have no available hexes when no occupied non-ocean tiles")

	// Place only an ocean tile
	oceanHex := testutil.FindUnoccupiedOceanHex(t, testGame)
	testutil.PlaceTileForPlayer(ctx, t, testGame, repo, p1ID, "ocean", oceanHex)

	availableHexes = testGame.CountAvailableHexesForTile("tile-replacement:nuclear-zone", p1ID, nil)
	testutil.AssertEqual(t, 0, availableHexes, "Should have no available hexes when only ocean tiles exist")
}

// =============================================================================
// Front of the Line (EXP001, event, experimental)
// Grants 2 extra actions this turn.
// =============================================================================

func TestFrontOfTheLine_GrantsExtraActions(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Hand().AddCard(testutil.CardID("Front of the Line"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Should start with 2 actions")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().TotalActions(), "Should start with 2 total actions")

	err := playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play successfully")

	testutil.AssertEqual(t, 3, testGame.CurrentTurn().ActionsRemaining(), "Should have 3 actions remaining (2 - 1 + 2)")
	testutil.AssertEqual(t, 4, testGame.CurrentTurn().TotalActions(), "Should have 4 total actions (2 + 2)")
}

func TestFrontOfTheLine_AsLastAction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	err := testGame.SetCurrentTurn(ctx, playerID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Hand().AddCard(testutil.CardID("Front of the Line"))

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play as last action")

	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Should have 2 actions remaining (1 - 1 + 2)")
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().TotalActions(), "Should have 3 total actions (1 + 2)")
	testutil.AssertEqual(t, playerID, testGame.CurrentTurn().PlayerID(), "Turn should still be current player's")
}

func TestFrontOfTheLine_UnlimitedActionsUnchanged(t *testing.T) {
	testGame, repo, cardRegistry, playerID := testutil.SetupSoloGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Hand().AddCard(testutil.CardID("Front of the Line"))

	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Solo should have unlimited actions")

	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	err := playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play in solo mode")

	testutil.AssertEqual(t, -1, testGame.CurrentTurn().ActionsRemaining(), "Solo should still have unlimited actions")
	testutil.AssertEqual(t, -1, testGame.CurrentTurn().TotalActions(), "Solo total actions should still be unlimited")
}

func TestFrontOfTheLine_ExtraActionsCanBeUsed(t *testing.T) {
	testGame, repo, cardRegistry, playerID, player2ID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 200)

	err := testGame.SetCurrentTurn(ctx, playerID, 1)
	testutil.AssertNoError(t, err, "Setting turn should succeed")

	p.Hand().AddCard(testutil.CardID("Front of the Line"))
	playAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 3}

	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Front of the Line"), payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Front of the Line should play successfully")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Should have 2 actions after Front of the Line")
	testutil.AssertEqual(t, 3, testGame.CurrentTurn().TotalActions(), "Should have 3 total actions (1 + 2)")

	p.Hand().AddCard(testutil.CardID("Power Plant"))
	payment2 := cardAction.PaymentRequest{Credits: 4}
	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Power Plant"), payment2, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "First extra action should succeed")
	testutil.AssertEqual(t, 1, testGame.CurrentTurn().ActionsRemaining(), "Should have 1 action remaining")
	testutil.AssertEqual(t, playerID, testGame.CurrentTurn().PlayerID(), "Should still be current player's turn")

	p.Hand().AddCard(testutil.CardID("Asteroid"))
	payment3 := cardAction.PaymentRequest{Credits: 14}
	err = playAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Asteroid"), payment3, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Second extra action should succeed")

	testutil.AssertEqual(t, player2ID, testGame.CurrentTurn().PlayerID(), "Turn should advance to player 2")
	testutil.AssertEqual(t, 2, testGame.CurrentTurn().ActionsRemaining(), "Player 2 should have 2 actions")
}

func TestFrontOfTheLine_CardIsEvent(t *testing.T) {
	card := testutil.GetCardByName("Front of the Line")
	testutil.AssertEqual(t, "event", string(card.Type), "Front of the Line should be an event card")
	testutil.AssertEqual(t, 3, card.Cost, "Front of the Line should cost 3")

	found := false
	for _, behavior := range card.Behaviors {
		for _, output := range behavior.Outputs {
			if output.ResourceType == shared.ResourceExtraActions && output.Amount == 2 {
				found = true
			}
		}
	}
	testutil.AssertTrue(t, found, "Front of the Line should have extra-actions output of 2")
}
