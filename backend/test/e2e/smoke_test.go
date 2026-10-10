package e2e_test

import (
	"testing"

	"openmars/test/e2e/harness"
)

func TestSmoke_PlayToActionPhase(t *testing.T) {
	t.Parallel()
	for _, players := range [][]string{{"solo"}, {"alice", "bob"}, {"alice", "bob", "carol"}} {
		t.Run(players[0]+"+"+string(rune('0'+len(players))), func(t *testing.T) {
			t.Parallel()
			srv := harness.Start(t)
			game := srv.Play(t, nil, players...)
			game.Current(t)
		})
	}
}
