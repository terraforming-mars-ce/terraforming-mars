package e2e_test

import (
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/test/e2e/harness"
)

func logUpdates(t *testing.T, received []harness.Message) []dto.LogUpdatePayload {
	t.Helper()
	var updates []dto.LogUpdatePayload
	for _, msg := range received {
		if msg.Type == dto.MessageTypeLogUpdate {
			var payload dto.LogUpdatePayload
			msg.Decode(t, &payload)
			updates = append(updates, payload)
		}
	}
	return updates
}

func TestLogs_JoiningAndAskingSendHistoryActionsSendLiveEntries(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, nil, "alice", "bob")
	game.SyncAll(t)

	late := srv.Dial(t, "alice-phone")
	late.Send(t, dto.MessageTypePlayerConnect, map[string]any{"gameId": game.ID, "playerName": "alice", "playerId": game.Players[0].PlayerID()})
	history := logUpdates(t, late.Sync(t))
	if len(history) != 1 || !history[0].IsHistory || len(history[0].Logs) == 0 {
		t.Fatalf("a reconnecting player should get the whole log once, as history: %+v", history)
	}

	watcher := srv.Dial(t, "watcher")
	watcher.Send(t, dto.MessageTypeSpectatorConnect, map[string]any{"gameId": game.ID, "spectatorName": "watcher"})
	if got := logUpdates(t, watcher.Sync(t)); len(got) != 1 || !got[0].IsHistory {
		t.Fatalf("a new spectator should get the log as history: %+v", got)
	}
	watcher.Send(t, dto.MessageTypeRequestLogs, map[string]any{})
	if got := logUpdates(t, watcher.Sync(t)); len(got) != 1 || !got[0].IsHistory || len(got[0].Logs) != len(history[0].Logs) {
		t.Fatalf("asking for the log should send all of it as history: %+v", got)
	}
	game.SyncAll(t)

	current := game.Current(t)
	current.Send(t, dto.MessageTypeActionStandardProject, map[string]any{"projectId": "power-plant", "payment": harness.Pay(powerPlantCost(t, current))})
	for _, c := range append([]*harness.Client{current}, append(game.Others(current), watcher)...) {
		received := c.Sync(t)
		for _, msg := range received {
			if msg.Type == dto.MessageTypeError {
				t.Fatalf("%s: building a power plant failed: %s", c.Name, msg.Payload)
			}
		}
		live := logUpdates(t, received)
		if len(live) != 1 || live[0].IsHistory || len(live[0].Logs) == 0 {
			t.Fatalf("%s: an action should send its new log entries live: %+v", c.Name, live)
		}
	}
}

func powerPlantCost(t *testing.T, c *harness.Client) map[string]int {
	t.Helper()
	for _, p := range c.State(t).CurrentPlayer.StandardProjects {
		if p.ProjectType == "power-plant" {
			return p.EffectiveCost
		}
	}
	t.Fatalf("no power plant project")
	return nil
}
