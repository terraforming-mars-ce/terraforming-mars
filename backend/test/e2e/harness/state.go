package harness

import (
	"testing"

	"openmars/internal/delivery/dto"
)

// Other returns the public view of another player, failing the test when the player is
// not in the game.
func Other(t testing.TB, g dto.GameDto, playerID string) dto.OtherPlayerDto {
	t.Helper()
	for _, p := range g.OtherPlayers {
		if p.ID == playerID {
			return p
		}
	}
	t.Fatalf("player %s is not among the other players of game %s", playerID, g.ID)
	return dto.OtherPlayerDto{}
}

// HasOther reports whether another player is in the game.
func HasOther(g dto.GameDto, playerID string) bool {
	for _, p := range g.OtherPlayers {
		if p.ID == playerID {
			return true
		}
	}
	return false
}

// Connected reports whether another player is in the game and connected.
func Connected(g dto.GameDto, playerID string) bool {
	for _, p := range g.OtherPlayers {
		if p.ID == playerID {
			return p.IsConnected
		}
	}
	return false
}
