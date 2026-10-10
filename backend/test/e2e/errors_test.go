package e2e_test

import (
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/test/e2e/harness"
)

func TestErrors_RejectedCardPlayNamesTheCardAndKeepsItInHand(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, harness.DevMode, "alice", "bob")
	player := game.Current(t)
	card := cardNobodyHolds(t, game, "Giant Ice Asteroid", "Comet", "Big Asteroid")
	player.Admin(t, dto.AdminCommandTypeGiveCard, map[string]any{"playerId": player.PlayerID(), "cardId": card})
	player.Admin(t, dto.AdminCommandTypeSetResources, map[string]any{"playerId": player.PlayerID(), "resources": map[string]any{"credits": 0}})
	game.SyncAll(t)

	player.Send(t, dto.MessageTypeActionPlayCard, map[string]any{"cardId": card, "payment": map[string]any{"credits": 0}})
	rejected := player.AwaitError(t)
	if rejected.CardID != card || rejected.RequestType != dto.MessageTypeActionPlayCard {
		t.Fatalf("the error should name the rejected card play, got %+v", rejected)
	}
	for _, other := range game.Others(player) {
		other.ExpectQuiet(t)
	}
	inHand := false
	for _, c := range player.State(t).CurrentPlayer.Cards {
		inHand = inHand || c.ID == card
	}
	if !inHand {
		t.Fatalf("a card that could not be played must stay in hand")
	}
}

func TestErrors_RejectedResourceRemovalNamesTheSelection(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, nil, "alice", "bob")
	player := game.Current(t)
	for _, amount := range []any{"x", 1.5, -1, nil} {
		player.Send(t, dto.MessageTypeActionConfirmResourceRemoval, map[string]any{"selectionId": "pending-7", "targetPlayerId": "target", "amount": amount})
		if got := player.AwaitError(t); got.SelectionID != "pending-7" {
			t.Fatalf("amount %v: the error should name the selection, got %+v", amount, got)
		}
	}
}

func TestErrors_RejectedTradeChoicesNameTheProblem(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, nil, "alice", "bob")
	player := game.Current(t)
	for _, steps := range []any{"two", -1, 1.5, nil} {
		player.Send(t, dto.MessageTypeActionColonyTrade, map[string]any{"colonyId": "luna", "paymentType": "energy", "trackSteps": steps})
		player.AwaitError(t)
	}
	for _, other := range game.Others(player) {
		other.ExpectQuiet(t)
	}
}
