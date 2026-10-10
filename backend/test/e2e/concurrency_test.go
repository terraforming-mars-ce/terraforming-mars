package e2e_test

import (
	"math/rand/v2"
	"strings"
	"sync"
	"testing"
	"time"

	"openmars/internal/delivery/dto"
	"openmars/test/e2e/harness"
)

// TestConcurrency_ManyClientsAtOnce has every client send a stream of random requests,
// valid and invalid, at the same time. Run with -race. Afterwards the server must still
// answer, no request may have crashed its handler, and every client must agree on the
// state of the game.
func TestConcurrency_ManyClientsAtOnce(t *testing.T) {
	t.Parallel()
	seed := uint64(time.Now().UnixNano())
	t.Logf("seed %d", seed)

	srv := harness.Start(t)
	game := srv.Play(t, harness.DevMode, "alice", "bob", "carol", "dave")
	watchers := []*harness.Client{srv.Dial(t, "watcher-1"), srv.Dial(t, "watcher-2")}
	for _, w := range watchers {
		w.Spectate(t, game.ID, w.Name)
	}
	other := srv.Play(t, nil, "elsewhere")

	types := make([]dto.MessageType, 0, len(contracts))
	for messageType := range contracts {
		switch messageType {
		case dto.MessageTypeEndGame, dto.MessageTypeKickPlayer, dto.MessageTypeKickSpectator, dto.MessageTypeConvertToBot, dto.MessageTypeAdminCommand:
			continue
		}
		types = append(types, messageType)
	}

	perClient := 300
	if harness.Race || testing.CoverMode() != "" {
		perClient = 60
	}
	clients := append(append([]*harness.Client{}, game.Players...), watchers...)
	var wg sync.WaitGroup
	for i, c := range clients {
		wg.Add(1)
		go func(c *harness.Client, rng *rand.Rand) {
			defer wg.Done()
			for range perClient {
				messageType := types[rng.IntN(len(types))]
				payloads := append(append([]any{}, garbage...), contracts[messageType]...)
				var payload any = payloads[rng.IntN(len(payloads))]
				switch rng.IntN(4) {
				case 0:
					messageType, payload = dto.MessageTypeChatMessage, map[string]any{"message": "hello"}
				case 1:
					messageType, payload = dto.MessageTypeActionSkipAction, map[string]any{}
				}
				if !c.TrySend(t, messageType, payload) {
					return
				}
			}
		}(c, rand.New(rand.NewPCG(seed, uint64(i))))
	}
	wg.Wait()

	for _, c := range append(clients, other.Players...) {
		for _, msg := range c.Sync(t) {
			if msg.Type == dto.MessageTypeError && strings.Contains(string(msg.Payload), "Internal server error") {
				t.Fatalf("%s: a request crashed its handler: %s", c.Name, msg.Payload)
			}
		}
	}
	game.ExpectConverged(t)
	other.ExpectConverged(t)
}
