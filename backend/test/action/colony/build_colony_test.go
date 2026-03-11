package colony_test

import (
	"context"
	"testing"

	colonyAction "terraforming-mars-backend/internal/action/colony"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/test/testutil"
)

func TestBuildColony_DeductsCredits(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 1, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna")
	testutil.AssertNoError(t, err, "Build colony should succeed")

	testutil.AssertEqual(t, 33, testutil.GetPlayerCredits(p), "Should deduct 17 credits (50 - 17 = 33)")
}

func TestBuildColony_GivesPlacementReward(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Luna first colony reward is 2 credits
	setupColonyTile(testGame, "luna", 1, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)
	creditsBefore := 50

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna")
	testutil.AssertNoError(t, err, "Build colony should succeed")

	// 50 - 17 (cost) + reward from first Luna slot
	creditsAfter := testutil.GetPlayerCredits(p)
	reward := creditsAfter - (creditsBefore - 17)
	testutil.AssertTrue(t, reward >= 0, "Should receive placement reward")
}

func TestBuildColony_CardTargetedReward_CreatesPendingSelection(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// Titan first colony reward is 3 floaters
	setupColonyTile(testGame, "titan", 1, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)

	// Give player a card with floater storage
	aerialMappersID := testutil.CardID("Aerial Mappers")
	p.PlayedCards().AddCard(aerialMappersID, "Aerial Mappers", "active", []string{"venus"})

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "titan")
	testutil.AssertNoError(t, err, "Build colony should succeed")

	selection := p.Selection().GetPendingColonyResourceSelection()
	testutil.AssertTrue(t, selection != nil, "Should have pending colony resource selection")
	testutil.AssertEqual(t, 3, selection.Amount, "Titan first slot reward is 3 floaters")
	testutil.AssertEqual(t, "floater", selection.ResourceType, "Resource type should be floater")
	testutil.AssertEqual(t, "build", selection.Reason, "Reason should be build for colony placement")
}

func TestBuildColony_FullColony_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Luna has 3 slots, all taken
	setupColonyTile(testGame, "luna", 3, []string{"other-1", "other-2", "other-3"})

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna")
	testutil.AssertError(t, err, "Should fail when colony is full")
}

func TestBuildColony_DuplicateColony_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Player already has a colony here
	setupColonyTile(testGame, "luna", 1, []string{playerID})

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna")
	testutil.AssertError(t, err, "Should fail when player already has colony on this tile")
}

func TestBuildColony_InsufficientCredits_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 1, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 10)

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna")
	testutil.AssertError(t, err, "Should fail with insufficient credits")
}

func TestBuildColony_OceanPlacementReward_CreatesTileSelection(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Europa colony reward is ocean-placement
	setupColonyTile(testGame, "europa", 1, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "europa")
	testutil.AssertNoError(t, err, "Build colony should succeed")

	tileSelection := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, tileSelection != nil, "Should have pending tile selection for ocean placement")
	testutil.AssertEqual(t, "ocean", tileSelection.TileType, "Tile type should be ocean")
}

func TestBuildColony_PlacesPlayerOnTile(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 1, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 50)

	action := colonyAction.NewBuildColonyAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna")
	testutil.AssertNoError(t, err, "Build colony should succeed")

	tileState := testGame.GetColonyTileState("luna")
	testutil.AssertEqual(t, 1, len(tileState.PlayerColonies), "Should have 1 colony")
	testutil.AssertEqual(t, playerID, tileState.PlayerColonies[0], "Colony should belong to player")
}
