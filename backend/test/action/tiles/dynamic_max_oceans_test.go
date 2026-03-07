package tiles_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func getOceanSpaceCoords() []shared.HexPosition {
	return []shared.HexPosition{
		{Q: -4, R: 0, S: 4},
		{Q: -3, R: -1, S: 4},
		{Q: -1, R: -2, S: 3},
		{Q: 1, R: 1, S: -2},
		{Q: 2, R: -1, S: -1},
		{Q: 3, R: -2, S: -1},
		{Q: 0, R: 3, S: -3},
		{Q: -2, R: 4, S: -2},
		{Q: 1, R: 3, S: -4},
	}
}

func getLandCoords() shared.HexPosition {
	return shared.HexPosition{Q: 0, R: 0, S: 0}
}

func TestMaxOceans_DefaultIs9(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	testutil.AssertEqual(t, global_parameters.MaxOceans, testGame.GlobalParameters().GetMaxOceans(), "default maxOceans should be 9")
}

func TestMaxOceans_ReducesWhenNonOceanTileOnOceanSpace(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	p := testGame.GetAllPlayers()[0]
	oceanSpaces := getOceanSpaceCoords()

	occupant := board.TileOccupant{
		Type: shared.ResourceType("mohole-tile"),
		Tags: []string{},
	}
	err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[0], occupant, p.ID())
	testutil.AssertNoError(t, err, "placing mohole on ocean space")

	time.Sleep(20 * time.Millisecond)

	testutil.AssertEqual(t, 8, testGame.GlobalParameters().GetMaxOceans(), "maxOceans should reduce to 8")
}

func TestMaxOceans_ReducesAfterOceansPlaced(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	p := testGame.GetAllPlayers()[0]
	oceanSpaces := getOceanSpaceCoords()

	// Place 5 ocean tiles
	for i := 0; i < 5; i++ {
		oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		testutil.AssertNoError(t, err, "placing ocean tile")
		testGame.GlobalParameters().PlaceOcean(ctx)
	}
	time.Sleep(20 * time.Millisecond)

	testutil.AssertEqual(t, 5, testGame.GlobalParameters().Oceans(), "should have 5 oceans")
	testutil.AssertEqual(t, 9, testGame.GlobalParameters().GetMaxOceans(), "maxOceans should still be 9")
	testutil.AssertEqual(t, 4, testGame.Board().FreeOceanSpaces(), "should have 4 free ocean spaces")

	// Place mohole on ocean space: 3 free, need 9-5=4 → 3 < 4 → reduce to 5+3=8
	moholeOccupant := board.TileOccupant{Type: shared.ResourceType("mohole-tile"), Tags: []string{}}
	err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[5], moholeOccupant, p.ID())
	testutil.AssertNoError(t, err, "placing mohole on ocean space")
	time.Sleep(20 * time.Millisecond)

	testutil.AssertEqual(t, 8, testGame.GlobalParameters().GetMaxOceans(), "maxOceans should reduce to 8")
	testutil.AssertEqual(t, 3, testGame.Board().FreeOceanSpaces(), "should have 3 free ocean spaces")
}

func TestMaxOceans_OceanTileOnOceanSpace_NoReduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	p := testGame.GetAllPlayers()[0]
	oceanSpaces := getOceanSpaceCoords()

	oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
	err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[0], oceanOccupant, p.ID())
	testutil.AssertNoError(t, err, "placing ocean tile")
	testGame.GlobalParameters().PlaceOcean(ctx)
	time.Sleep(20 * time.Millisecond)

	testutil.AssertEqual(t, 9, testGame.GlobalParameters().GetMaxOceans(), "maxOceans should remain 9")
	testutil.AssertEqual(t, 1, testGame.GlobalParameters().Oceans(), "should have 1 ocean")
}

func TestMaxOceans_NonOceanOnLandSpace_NoReduction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	p := testGame.GetAllPlayers()[0]

	cityOccupant := board.TileOccupant{Type: shared.ResourceCityTile, Tags: []string{}}
	err := testGame.Board().UpdateTileOccupancy(ctx, getLandCoords(), cityOccupant, p.ID())
	testutil.AssertNoError(t, err, "placing city on land")
	time.Sleep(20 * time.Millisecond)

	testutil.AssertEqual(t, 9, testGame.GlobalParameters().GetMaxOceans(), "maxOceans should remain 9")
}

func TestMaxOceans_IsMaxedRespectsReducedMax(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	p := testGame.GetAllPlayers()[0]
	oceanSpaces := getOceanSpaceCoords()
	gp := testGame.GlobalParameters()

	// Place mohole on ocean space → maxOceans reduces to 8
	moholeOccupant := board.TileOccupant{Type: shared.ResourceType("mohole-tile"), Tags: []string{}}
	err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[0], moholeOccupant, p.ID())
	testutil.AssertNoError(t, err, "placing mohole")
	time.Sleep(20 * time.Millisecond)
	testutil.AssertEqual(t, 8, gp.GetMaxOceans(), "maxOceans should be 8")

	// Max temperature and oxygen
	gp.SetTemperature(ctx, global_parameters.MaxTemperature)
	gp.SetOxygen(ctx, global_parameters.MaxOxygen)

	// Place 8 oceans (the new max)
	for i := 1; i <= 8; i++ {
		oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean %d", i))
		gp.PlaceOcean(ctx)
	}

	testutil.AssertEqual(t, 8, gp.Oceans(), "should have 8 oceans")
	testutil.AssertTrue(t, gp.IsMaxed(), "IsMaxed should be true when oceans equals reduced max")
}

func TestMaxOceans_FreeOceanSpaces_AlwaysGTE_Remaining(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()
	p := testGame.GetAllPlayers()[0]
	oceanSpaces := getOceanSpaceCoords()
	gp := testGame.GlobalParameters()

	checkInvariant := func(step string) {
		t.Helper()
		freeSpaces := testGame.Board().FreeOceanSpaces()
		remaining := gp.GetMaxOceans() - gp.Oceans()
		if freeSpaces < remaining {
			t.Errorf("[%s] Invariant violated: freeOceanSpaces(%d) < oceansRemaining(%d)", step, freeSpaces, remaining)
		}
	}

	checkInvariant("initial")

	// Place 3 ocean tiles
	for i := 0; i < 3; i++ {
		oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
		testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		gp.PlaceOcean(ctx)
		time.Sleep(20 * time.Millisecond)
		checkInvariant(fmt.Sprintf("after ocean %d", i+1))
	}

	// Place mohole on ocean space 3
	moholeOccupant := board.TileOccupant{Type: shared.ResourceType("mohole-tile"), Tags: []string{}}
	testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[3], moholeOccupant, p.ID())
	time.Sleep(20 * time.Millisecond)
	checkInvariant("after mohole 1")

	// Place another mohole on ocean space 4
	testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[4], moholeOccupant, p.ID())
	time.Sleep(20 * time.Millisecond)
	checkInvariant("after mohole 2")

	// Place 2 more ocean tiles
	for i := 5; i < 7; i++ {
		oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
		testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		gp.PlaceOcean(ctx)
		time.Sleep(20 * time.Millisecond)
		checkInvariant(fmt.Sprintf("after ocean on space %d", i))
	}

	// Place mohole on ocean space 7
	testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[7], moholeOccupant, p.ID())
	time.Sleep(20 * time.Millisecond)
	checkInvariant("after mohole 3")

	// Final: 5 oceans, 3 moholes, 1 free ocean space
	testutil.AssertEqual(t, 5, gp.Oceans(), "should have 5 oceans")
	testutil.AssertEqual(t, 1, testGame.Board().FreeOceanSpaces(), "should have 1 free ocean space")
	testutil.AssertEqual(t, 6, gp.GetMaxOceans(), "maxOceans should be 6")
}

func TestFreeOceanSpaces_InitialCount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	testutil.AssertEqual(t, 9, testGame.Board().FreeOceanSpaces(), "should start with 9 free ocean spaces")
}
