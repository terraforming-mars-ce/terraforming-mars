package bot_test

import (
	"context"
	"testing"

	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

func TestAutopilot_ResolvesPendingTileThenPasses(t *testing.T) {
	fx := newBotFixture(t)
	ctx := context.Background()
	hex := testutil.FindUnoccupiedLandHex(t, fx.game)
	testutil.AssertNoError(t, fx.game.SetPendingTileSelection(ctx, fx.botID, &shared.PendingTileSelection{
		TileType: "city", AvailableHexes: []string{hex}, Source: "test",
	}), "pending tile should be set")
	hooks := &fakeHooks{fx: fx}

	snap, err := hooks.Snapshot(ctx)
	testutil.AssertNoError(t, err, "snapshot")
	testutil.AssertNoError(t, fx.tools.Autopilot(ctx, snap), "autopilot should place the tile")
	testutil.AssertTrue(t, testutil.GetTileAtHex(t, fx.game, hex).OccupiedBy != nil, "the tile should be placed")
	testutil.AssertEqual(t, fx.botID, fx.currentTurn(), "placing the tile does not end the turn")

	snap, err = hooks.Snapshot(ctx)
	testutil.AssertNoError(t, err, "snapshot")
	testutil.AssertNoError(t, fx.tools.Autopilot(ctx, snap), "autopilot should pass")
	testutil.AssertEqual(t, fx.humanID, fx.currentTurn(), "passing hands the turn over")
}
