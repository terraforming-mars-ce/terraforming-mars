package card_packs_test

import (
	"context"
	"fmt"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	tileAction "terraforming-mars-backend/internal/action/tile"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func formatHex(pos shared.HexPosition) string {
	return fmt.Sprintf("%d,%d,%d", pos.Q, pos.R, pos.S)
}

// Known board positions from GenerateMarsBoard (official Tharsis layout):
// Ocean at (4,-1,-3) is adjacent to land (3,-1,-2) and land (3,0,-3)
var (
	// An ocean tile position that is adjacent to landHexA and landHexB
	oceanHexA = formatHex(shared.HexPosition{Q: 4, R: -1, S: -3})
	// A land tile adjacent to oceanHexA
	landHexA = formatHex(shared.HexPosition{Q: 3, R: -1, S: -2})
	// Another land tile adjacent to oceanHexA
	landHexB = formatHex(shared.HexPosition{Q: 3, R: 0, S: -3})

	// An ocean tile at the bottom of the board for isolated test
	oceanIsolated = formatHex(shared.HexPosition{Q: 0, R: 4, S: -4})
)

func placeTileForPlayer(ctx context.Context, t *testing.T, g *game.Game, repo game.GameRepository, playerID string, tileType string, hexStr string) {
	t.Helper()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	cr := testutil.CreateTestCardRegistry()

	// Temporarily set turn to this player so tile placement is allowed
	err := g.SetCurrentTurn(ctx, playerID, 2)
	if err != nil {
		t.Fatalf("Failed to set turn for tile placement: %v", err)
	}

	g.SetPendingTileSelection(ctx, playerID, &player.PendingTileSelection{
		TileType:       tileType,
		AvailableHexes: []string{hexStr},
		Source:         "test",
	})

	selectTile := tileAction.NewSelectTileAction(repo, cr, stateRepo, logger)
	_, err = selectTile.Execute(ctx, g.ID(), playerID, hexStr)
	if err != nil {
		t.Fatalf("Failed to place %s tile on %s: %v", tileType, hexStr, err)
	}
}

func playFloodingCard(ctx context.Context, t *testing.T, g *game.Game, repo game.GameRepository, playerID string) {
	t.Helper()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	cr := testutil.CreateTestCardRegistry()

	p, _ := g.GetPlayer(playerID)
	card := testutil.GetCardByName("Flooding")
	p.Hand().AddCard(card.ID)

	playCard := cardAction.NewPlayCardAction(repo, cr, stateRepo, logger)
	payment := cardAction.PaymentRequest{Credits: card.Cost}
	err := playCard.Execute(ctx, g.ID(), playerID, card.ID, payment, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("Flooding card should play: %v", err)
	}
}

func selectOceanTile(ctx context.Context, t *testing.T, g *game.Game, repo game.GameRepository, playerID string, oceanHex string) {
	t.Helper()
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	cr := testutil.CreateTestCardRegistry()

	selectTile := tileAction.NewSelectTileAction(repo, cr, stateRepo, logger)
	_, err := selectTile.Execute(ctx, g.ID(), playerID, oceanHex)
	if err != nil {
		t.Fatalf("Ocean tile placement should succeed: %v", err)
	}
}

func TestFlooding_AdjacentOpponentTile_StealOffered(t *testing.T) {
	ctx := context.Background()
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	stateRepo := game.NewInMemoryGameStateRepository()
	_ = stateRepo
	_ = cardRegistry
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]

	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 100)
	p2, _ := testGame.GetPlayer(p2ID)
	testutil.SetPlayerCredits(ctx, p2, 100)

	// Place a city for p2 on the land tile adjacent to the ocean spot
	placeTileForPlayer(ctx, t, testGame, repo, p2ID, "city", landHexA)

	// Play Flooding card for p1
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)

	// Should have pending tile selection for the ocean
	pending := testGame.GetPendingTileSelection(p1ID)
	testutil.AssertTrue(t, pending != nil, "Player should have pending ocean tile selection")

	// Place the ocean on the hex adjacent to p2's city
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	// Player should now have pending steal target selection
	stealSelection := p1.Selection().GetPendingStealTargetSelection()
	testutil.AssertTrue(t, stealSelection != nil, "Player should have pending steal target selection")
	testutil.AssertEqual(t, 4, stealSelection.Amount, "Steal amount should be 4")
	testutil.AssertTrue(t, contains(stealSelection.EligiblePlayerIDs, p2ID),
		"P2 should be in eligible steal targets")
	testutil.AssertTrue(t, !contains(stealSelection.EligiblePlayerIDs, p1ID),
		"P1 (self) should not be in eligible steal targets")
}

func TestFlooding_NoAdjacentOwnedTiles_NoPrompt(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	p1ID := playerIDs[0]

	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 100)

	// Play Flooding card without any tiles placed by anyone
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)

	// Place ocean in an area with no adjacent owned tiles
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanIsolated)

	// No pending steal target selection
	stealSelection := p1.Selection().GetPendingStealTargetSelection()
	testutil.AssertTrue(t, stealSelection == nil, "No steal target selection expected when no adjacent owned tiles")
}

func TestFlooding_MultipleAdjacentOpponents_AllEligible(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 3)
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	p3ID := playerIDs[2]

	for _, pid := range playerIDs {
		p, _ := testGame.GetPlayer(pid)
		testutil.SetPlayerCredits(ctx, p, 100)
	}

	// Place cities for p2 and p3 on two different land tiles adjacent to same ocean
	placeTileForPlayer(ctx, t, testGame, repo, p2ID, "city", landHexA)
	placeTileForPlayer(ctx, t, testGame, repo, p3ID, "city", landHexB)

	// Play Flooding for p1
	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	p1, _ := testGame.GetPlayer(p1ID)
	stealSelection := p1.Selection().GetPendingStealTargetSelection()
	testutil.AssertTrue(t, stealSelection != nil, "Player should have pending steal target selection")
	testutil.AssertTrue(t, contains(stealSelection.EligiblePlayerIDs, p2ID), "P2 should be eligible")
	testutil.AssertTrue(t, contains(stealSelection.EligiblePlayerIDs, p3ID), "P3 should be eligible")
	testutil.AssertTrue(t, !contains(stealSelection.EligiblePlayerIDs, p1ID), "P1 should not be eligible")
}

func TestFlooding_ConfirmSteal_CreditsTransferred(t *testing.T) {
	ctx := context.Background()
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]

	p1, _ := testGame.GetPlayer(p1ID)
	p2, _ := testGame.GetPlayer(p2ID)
	testutil.SetPlayerCredits(ctx, p1, 100)
	testutil.SetPlayerCredits(ctx, p2, 50)

	placeTileForPlayer(ctx, t, testGame, repo, p2ID, "city", landHexA)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	p1CreditsBefore := testutil.GetPlayerCredits(p1)
	p2CreditsBefore := testutil.GetPlayerCredits(p2)

	confirmSteal := confirmAction.NewConfirmStealTargetAction(repo, cardRegistry, stateRepo, logger)
	err = confirmSteal.Execute(ctx, testGame.ID(), p1ID, p2ID)
	testutil.AssertNoError(t, err, "Confirm steal should succeed")

	testutil.AssertEqual(t, p1CreditsBefore+4, testutil.GetPlayerCredits(p1), "P1 should gain 4 credits")
	testutil.AssertEqual(t, p2CreditsBefore-4, testutil.GetPlayerCredits(p2), "P2 should lose 4 credits")

	testutil.AssertTrue(t, p1.Selection().GetPendingStealTargetSelection() == nil,
		"Pending steal selection should be cleared after confirm")
}

func TestFlooding_PartialSteal_TargetHasLessThan4Credits(t *testing.T) {
	ctx := context.Background()
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]

	p1, _ := testGame.GetPlayer(p1ID)
	p2, _ := testGame.GetPlayer(p2ID)
	testutil.SetPlayerCredits(ctx, p1, 100)
	testutil.SetPlayerCredits(ctx, p2, 2)

	placeTileForPlayer(ctx, t, testGame, repo, p2ID, "city", landHexA)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	p1CreditsBefore := testutil.GetPlayerCredits(p1)

	confirmSteal := confirmAction.NewConfirmStealTargetAction(repo, cardRegistry, stateRepo, logger)
	err = confirmSteal.Execute(ctx, testGame.ID(), p1ID, p2ID)
	testutil.AssertNoError(t, err, "Confirm steal should succeed")

	testutil.AssertEqual(t, p1CreditsBefore+2, testutil.GetPlayerCredits(p1), "P1 should gain only 2 credits")
	testutil.AssertEqual(t, 0, testutil.GetPlayerCredits(p2), "P2 should have 0 credits")
}

func TestFlooding_SkipSteal_NoCreditsChange(t *testing.T) {
	ctx := context.Background()
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]

	p1, _ := testGame.GetPlayer(p1ID)
	p2, _ := testGame.GetPlayer(p2ID)
	testutil.SetPlayerCredits(ctx, p1, 100)
	testutil.SetPlayerCredits(ctx, p2, 50)

	placeTileForPlayer(ctx, t, testGame, repo, p2ID, "city", landHexA)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	p1CreditsBefore := testutil.GetPlayerCredits(p1)
	p2CreditsBefore := testutil.GetPlayerCredits(p2)

	confirmSteal := confirmAction.NewConfirmStealTargetAction(repo, cardRegistry, stateRepo, logger)
	err = confirmSteal.Execute(ctx, testGame.ID(), p1ID, "")
	testutil.AssertNoError(t, err, "Skip steal should succeed")

	testutil.AssertEqual(t, p1CreditsBefore, testutil.GetPlayerCredits(p1), "P1 credits should be unchanged")
	testutil.AssertEqual(t, p2CreditsBefore, testutil.GetPlayerCredits(p2), "P2 credits should be unchanged")

	testutil.AssertTrue(t, p1.Selection().GetPendingStealTargetSelection() == nil,
		"Pending steal selection should be cleared after skip")
}

func TestFlooding_OwnTilesNotEligible(t *testing.T) {
	ctx := context.Background()
	testGame, repo, _, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	p1ID := playerIDs[0]

	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 100)

	// Place a city for p1 (self) on the adjacent land tile
	placeTileForPlayer(ctx, t, testGame, repo, p1ID, "city", landHexA)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	stealSelection := p1.Selection().GetPendingStealTargetSelection()
	testutil.AssertTrue(t, stealSelection == nil,
		"No steal target selection expected when only own tiles are adjacent")
}

func TestFlooding_InvalidTargetPlayer_Rejected(t *testing.T) {
	ctx := context.Background()
	testGame, repo, cardRegistry, playerIDs := testutil.SetupMultiPlayerGame(t, 2)
	logger := testutil.TestLogger()
	stateRepo := game.NewInMemoryGameStateRepository()
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]

	p1, _ := testGame.GetPlayer(p1ID)
	testutil.SetPlayerCredits(ctx, p1, 100)
	p2, _ := testGame.GetPlayer(p2ID)
	testutil.SetPlayerCredits(ctx, p2, 100)

	placeTileForPlayer(ctx, t, testGame, repo, p2ID, "city", landHexA)

	err := testGame.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Failed to set turn")
	playFloodingCard(ctx, t, testGame, repo, p1ID)
	selectOceanTile(ctx, t, testGame, repo, p1ID, oceanHexA)

	confirmSteal := confirmAction.NewConfirmStealTargetAction(repo, cardRegistry, stateRepo, logger)
	err = confirmSteal.Execute(ctx, testGame.ID(), p1ID, "nonexistent-player")
	testutil.AssertTrue(t, err != nil, "Should fail when targeting ineligible player")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
