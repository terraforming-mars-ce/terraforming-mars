package core_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// =============================================================================
// Pending selection lock: actions blocked while any selection is pending
// =============================================================================

func TestPendingSelectionLock_TileSelection_BlocksCardPlay(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	testutil.AssertNoError(t, testGame.SetPendingTileSelection(ctx, playerID, &shared.PendingTileSelection{
		TileType: "city",
		Source:   "test",
	}), "set pending tile selection")

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, !state.Available(), "Card should not be available during pending tile selection")
}

func TestPendingSelectionLock_StealTarget_BlocksCardPlay(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	p.Selection().SetPendingStealTargetSelection(&shared.PendingStealTargetSelection{
		Amount:            4,
		EligiblePlayerIDs: []string{"other-player"},
	})

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, !state.Available(), "Card should not be available during pending steal target selection")
}

func TestPendingSelectionLock_ColonyResourceQueue_BlocksCardPlay(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	p.Selection().AppendPendingColonyResource(shared.PendingColonyResourceSelection{
		ResourceType: "floater",
		Amount:       2,
		Source:       "Titan",
		Reason:       "trade",
	})

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, !state.Available(), "Card should not be available during pending colony resource selection")
}

func TestPendingSelectionLock_FreeTradeSelection_BlocksCardPlay(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	p.Selection().SetPendingFreeTradeSelection(&shared.PendingFreeTradeSelection{
		AvailableColonyIDs: []string{"luna"},
		Source:             "test",
	})

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, !state.Available(), "Card should not be available during pending free trade selection")
}

func TestPendingSelectionLock_ColonyPlacement_BlocksCardPlay(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	p.Selection().SetPendingColonySelection(&shared.PendingColonySelection{
		AvailableColonyIDs: []string{"luna"},
		Source:             "test",
	})

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, !state.Available(), "Card should not be available during pending colony placement selection")
}

func TestPendingSelectionLock_CardDrawSelection_BlocksCardPlay(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	p.Selection().SetPendingCardDrawSelection(&shared.PendingCardDrawSelection{
		AvailableCards: []string{"some-card"},
		FreeTakeCount:  1,
		Source:         "test",
	})

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, !state.Available(), "Card should not be available during pending card draw selection")
}

func TestPendingSelectionLock_NoPending_CardPlayable(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	c := testutil.GetCardByName("Power Plant")
	card := &c
	p.Hand().AddCard(card.ID)
	testutil.SetPlayerCredits(ctx, p, 100)

	state := action.CalculatePlayerCardState(card, p, testGame, cardRegistry)
	testutil.AssertTrue(t, state.Available(), "Card should be available when no pending selections")
}
