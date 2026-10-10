package e2e_test

import (
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/internal/game/shared"
	"openmars/test/e2e/harness"
)

func TestRobustness_AStalledClientDoesNotFreezeOtherGames(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, nil, "stalled", "other")
	other := srv.Lobby(t, nil, "elsewhere").Host()

	stalled := srv.DialStalled(t, "frozen-tab")
	stalled.SendRaw(t, []byte(`{"type":"player-connect","payload":{"gameId":"`+game.ID+`","playerName":"stalled","playerId":"`+game.Players[0].PlayerID()+`"}}`))
	for i := range 50000 {
		if !stalled.TrySend(t, dto.MessageTypeRequestLogs, map[string]any{}) {
			break
		}
		if i%250 == 0 {
			other.Sync(t)
		}
	}
	other.Sync(t)
	game.Players[1].Sync(t)
}

func TestRobustness_AStalledPlayerIsDisconnectedInsteadOfMissingUpdates(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Lobby(t, nil, "host")
	host := game.Host()

	stalled := srv.DialStalled(t, "frozen-tab")
	stalled.SendRaw(t, []byte(`{"type":"player-connect","payload":{"gameId":"`+game.ID+`","playerName":"frozen"}}`))
	state := host.AwaitState(t, "the frozen player joined", func(g dto.GameDto) bool { return len(g.OtherPlayers) == 1 })
	frozenID := state.OtherPlayers[0].ID

	colors := []string{shared.PlayerColors[len(shared.PlayerColors)-1], shared.PlayerColors[len(shared.PlayerColors)-2]}
	for i := 0; ; i++ {
		host.Send(t, dto.MessageTypeSetPlayerColor, map[string]any{"color": colors[i%2]})
		host.Sync(t)
		if !harness.HasOther(host.State(t), frozenID) {
			return
		}
		if i == 5000 {
			t.Fatalf("a player who reads nothing was never disconnected")
		}
	}
}

func TestRobustness_ShutdownWithConnectedClientsLeavesNothingRunning(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, nil, "alice", "bob")
	watcher := srv.Dial(t, "watcher")
	watcher.Spectate(t, game.ID, "watcher")
	idle := srv.Dial(t, "idle")
	idle.Sync(t)

	srv.Close()
	for _, c := range append(game.Players, watcher, idle) {
		c.AwaitClosed(t)
	}
}
