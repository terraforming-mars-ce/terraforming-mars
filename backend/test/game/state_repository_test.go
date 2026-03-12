package game_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestStateRepository_WriteInitialState(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	repo := game.NewInMemoryGameStateRepository()

	diff, err := repo.Write(context.Background(), testGame.ID(), testGame, "Game Setup", shared.SourceTypeInitial, "", "Game created")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	testutil.AssertEqual(t, int64(1), diff.SequenceNumber, "First write should have sequence 1")
	testutil.AssertEqual(t, testGame.ID(), diff.GameID, "GameID should match")
	testutil.AssertEqual(t, "Game Setup", diff.Source, "Source should match")
	testutil.AssertEqual(t, shared.SourceTypeInitial, diff.SourceType, "SourceType should match")
	testutil.AssertEqual(t, "Game created", diff.Description, "Description should match")

	if diff.Changes == nil {
		t.Fatal("Changes should not be nil")
	}
	if diff.Changes.Status == nil {
		t.Fatal("Status change should be captured for initial state")
	}
	testutil.AssertEqual(t, "", diff.Changes.Status.Old, "Old status should be empty for initial write")
}

func TestStateRepository_WriteIncrementalChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	repo := game.NewInMemoryGameStateRepository()

	_, err := repo.Write(context.Background(), testGame.ID(), testGame, "Game Setup", shared.SourceTypeInitial, "", "Game created")
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	players := testGame.GetAllPlayers()
	player := players[0]
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 10,
	})

	diff, err := repo.Write(context.Background(), testGame.ID(), testGame, "Test Card", shared.SourceTypeCardPlay, player.ID(), "Played Test Card")
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	testutil.AssertEqual(t, int64(2), diff.SequenceNumber, "Second write should have sequence 2")

	if diff.Changes.PlayerChanges == nil {
		t.Fatal("PlayerChanges should not be nil after resource change")
	}

	pc := diff.Changes.PlayerChanges[player.ID()]
	if pc == nil {
		t.Fatal("Player changes should exist for player")
	}
	if pc.Credits == nil {
		t.Fatal("Credits change should be captured")
	}
	testutil.AssertEqual(t, 10, pc.Credits.New-pc.Credits.Old, "Credits should have increased by 10")
}

func TestStateRepository_GetDiff(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	repo := game.NewInMemoryGameStateRepository()

	_, err := repo.Write(context.Background(), testGame.ID(), testGame, "Game Setup", shared.SourceTypeInitial, "", "Game created")
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	players := testGame.GetAllPlayers()
	player := players[0]
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceSteel: 5,
	})
	_, err = repo.Write(context.Background(), testGame.ID(), testGame, "Card A", shared.SourceTypeCardPlay, player.ID(), "Played Card A")
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourcePlantProduction: 2,
	})
	_, err = repo.Write(context.Background(), testGame.ID(), testGame, "Card B", shared.SourceTypeCardPlay, player.ID(), "Played Card B")
	if err != nil {
		t.Fatalf("Third write failed: %v", err)
	}

	diffs, err := repo.GetDiff(context.Background(), testGame.ID())
	if err != nil {
		t.Fatalf("GetDiff failed: %v", err)
	}

	testutil.AssertEqual(t, 3, len(diffs), "Should have 3 diffs")
	testutil.AssertEqual(t, int64(1), diffs[0].SequenceNumber, "First diff should have sequence 1")
	testutil.AssertEqual(t, int64(2), diffs[1].SequenceNumber, "Second diff should have sequence 2")
	testutil.AssertEqual(t, int64(3), diffs[2].SequenceNumber, "Third diff should have sequence 3")
}

func TestStateRepository_GetDiffGameNotFound(t *testing.T) {
	repo := game.NewInMemoryGameStateRepository()

	_, err := repo.GetDiff(context.Background(), "non-existent-game")
	if err == nil {
		t.Fatal("GetDiff should return error for non-existent game")
	}
}

func TestStateRepository_WriteNilGameReturnsError(t *testing.T) {
	repo := game.NewInMemoryGameStateRepository()

	_, err := repo.Write(context.Background(), "test-game", nil, "Test", shared.SourceTypeInitial, "", "Test")
	if err == nil {
		t.Fatal("Write should return error for nil game")
	}
}

func TestStateRepository_WriteContextCancelled(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	repo := game.NewInMemoryGameStateRepository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Write(ctx, testGame.ID(), testGame, "Test", shared.SourceTypeInitial, "", "Test")
	if err == nil {
		t.Fatal("Write should return error when context is cancelled")
	}
}

func TestStateRepository_GlobalParameterChanges(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	repo := game.NewInMemoryGameStateRepository()
	players := testGame.GetAllPlayers()
	player := players[0]

	_, err := repo.Write(context.Background(), testGame.ID(), testGame, "Game Setup", shared.SourceTypeInitial, "", "Game created")
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	_, err = testGame.GlobalParameters().IncreaseTemperature(context.Background(), 2, "")
	if err != nil {
		t.Fatalf("IncreaseTemperature failed: %v", err)
	}

	diff, err := repo.Write(context.Background(), testGame.ID(), testGame, "Heat Conversion", shared.SourceTypeResourceConvert, player.ID(), "Converted heat to temperature")
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	if diff.Changes.Temperature == nil {
		t.Fatal("Temperature change should be captured")
	}
	testutil.AssertEqual(t, 4, diff.Changes.Temperature.New-diff.Changes.Temperature.Old, "Temperature should have increased by 4 degrees")
}

func TestStateRepository_PhaseChange(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	repo := game.NewInMemoryGameStateRepository()

	_, err := repo.Write(context.Background(), testGame.ID(), testGame, "Game Setup", shared.SourceTypeInitial, "", "Game created")
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	err = testGame.UpdatePhase(context.Background(), shared.GamePhaseAction)
	if err != nil {
		t.Fatalf("UpdatePhase failed: %v", err)
	}

	diff, err := repo.Write(context.Background(), testGame.ID(), testGame, "Phase Transition", shared.SourceTypeGameEvent, "", "Phase changed to action")
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	if diff.Changes.Phase == nil {
		t.Fatal("Phase change should be captured")
	}
	testutil.AssertEqual(t, string(shared.GamePhaseAction), diff.Changes.Phase.New, "New phase should be action")
}

func TestStateRepository_NoChangesProducesEmptyDiff(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	repo := game.NewInMemoryGameStateRepository()

	_, err := repo.Write(context.Background(), testGame.ID(), testGame, "Game Setup", shared.SourceTypeInitial, "", "Game created")
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	diff, err := repo.Write(context.Background(), testGame.ID(), testGame, "No Op", shared.SourceTypeGameEvent, "", "No changes")
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	testutil.AssertEqual(t, int64(2), diff.SequenceNumber, "Should still increment sequence")

	if diff.Changes.Status != nil {
		t.Error("Status should be nil when unchanged")
	}
	if diff.Changes.Phase != nil {
		t.Error("Phase should be nil when unchanged")
	}
	if diff.Changes.Generation != nil {
		t.Error("Generation should be nil when unchanged")
	}
	if diff.Changes.Temperature != nil {
		t.Error("Temperature should be nil when unchanged")
	}
	if diff.Changes.Oxygen != nil {
		t.Error("Oxygen should be nil when unchanged")
	}
	if diff.Changes.Oceans != nil {
		t.Error("Oceans should be nil when unchanged")
	}
}
