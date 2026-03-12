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
		{Q: 1, R: -4, S: 3},
		{Q: 3, R: -4, S: 1},
		{Q: 4, R: -4, S: 0},
		{Q: 4, R: -3, S: -1},
		{Q: 4, R: -1, S: -3},
		{Q: -1, R: 0, S: 1},
		{Q: 0, R: 0, S: 0},
		{Q: 1, R: 0, S: -1},
		{Q: 1, R: 1, S: -2},
		{Q: 2, R: 1, S: -3},
		{Q: 3, R: 1, S: -4},
		{Q: 0, R: 4, S: -4},
	}
}

func getLandCoords() shared.HexPosition {
	return shared.HexPosition{Q: 2, R: -4, S: 2}
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

	// With 12 ocean spaces and 9 max oceans, we need to place 4 non-ocean tiles
	// on ocean spaces before maxOceans reduces (12-4=8 free < 9 remaining)
	occupant := board.TileOccupant{
		Type: shared.ResourceType("mohole-tile"),
		Tags: []string{},
	}
	for i := 0; i < 4; i++ {
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], occupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing mohole %d on ocean space", i))
	}
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
		_, err = testGame.GlobalParameters().PlaceOcean(ctx, "")
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean %d", i))
	}
	time.Sleep(20 * time.Millisecond)

	testutil.AssertEqual(t, 5, testGame.GlobalParameters().Oceans(), "should have 5 oceans")
	testutil.AssertEqual(t, 9, testGame.GlobalParameters().GetMaxOceans(), "maxOceans should still be 9")
	testutil.AssertEqual(t, 7, testGame.Board().FreeOceanSpaces(), "should have 7 free ocean spaces")

	// Place 4 moholes on ocean spaces: 3 free, need 9-5=4 → 3 < 4 → reduce to 5+3=8
	moholeOccupant := board.TileOccupant{Type: shared.ResourceType("mohole-tile"), Tags: []string{}}
	for i := 5; i < 9; i++ {
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], moholeOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing mohole on ocean space %d", i))
	}
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
	_, err = testGame.GlobalParameters().PlaceOcean(ctx, "")
	testutil.AssertNoError(t, err, "placing ocean")
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

	// Place 4 moholes on ocean spaces → maxOceans reduces to 8
	moholeOccupant := board.TileOccupant{Type: shared.ResourceType("mohole-tile"), Tags: []string{}}
	for i := 0; i < 4; i++ {
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], moholeOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing mohole %d", i))
	}
	time.Sleep(20 * time.Millisecond)
	testutil.AssertEqual(t, 8, gp.GetMaxOceans(), "maxOceans should be 8")

	// Max temperature and oxygen
	err := gp.SetTemperature(ctx, global_parameters.MaxTemperature)
	testutil.AssertNoError(t, err, "set max temperature")
	err = gp.SetOxygen(ctx, global_parameters.MaxOxygen)
	testutil.AssertNoError(t, err, "set max oxygen")

	// Place 8 oceans (the new max)
	for i := 4; i < 12; i++ {
		oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean %d", i-3))
		_, err = gp.PlaceOcean(ctx, "")
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean %d", i-3))
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
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean tile %d", i))
		_, err = gp.PlaceOcean(ctx, "")
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean %d", i))
		time.Sleep(20 * time.Millisecond)
		checkInvariant(fmt.Sprintf("after ocean %d", i+1))
	}

	// Place moholes on ocean spaces 3-6
	moholeOccupant := board.TileOccupant{Type: shared.ResourceType("mohole-tile"), Tags: []string{}}
	for i := 3; i < 7; i++ {
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], moholeOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing mohole on space %d", i))
		time.Sleep(20 * time.Millisecond)
		checkInvariant(fmt.Sprintf("after mohole on space %d", i))
	}

	// Place 2 more ocean tiles
	for i := 7; i < 9; i++ {
		oceanOccupant := board.TileOccupant{Type: shared.ResourceOceanTile, Tags: []string{}}
		err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[i], oceanOccupant, p.ID())
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean tile on space %d", i))
		_, err = gp.PlaceOcean(ctx, "")
		testutil.AssertNoError(t, err, fmt.Sprintf("placing ocean on space %d", i))
		time.Sleep(20 * time.Millisecond)
		checkInvariant(fmt.Sprintf("after ocean on space %d", i))
	}

	// Place mohole on ocean space 9
	err := testGame.Board().UpdateTileOccupancy(ctx, oceanSpaces[9], moholeOccupant, p.ID())
	testutil.AssertNoError(t, err, "placing mohole on space 9")
	time.Sleep(20 * time.Millisecond)
	checkInvariant("after mohole on space 9")

	// Final: 5 oceans, 5 moholes, 2 free ocean spaces
	testutil.AssertEqual(t, 5, gp.Oceans(), "should have 5 oceans")
	testutil.AssertEqual(t, 2, testGame.Board().FreeOceanSpaces(), "should have 2 free ocean spaces")
	testutil.AssertEqual(t, 7, gp.GetMaxOceans(), "maxOceans should be 7")
}

func TestFreeOceanSpaces_InitialCount(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)

	testutil.AssertEqual(t, 12, testGame.Board().FreeOceanSpaces(), "should start with 12 free ocean spaces")
}
