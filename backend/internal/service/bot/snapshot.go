package bot

import (
	"fmt"
	"strings"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Snapshot is the game as one bot sees it at one moment.
type Snapshot struct {
	Game     *game.Game
	Player   *player.Player
	View     dto.GameDto
	PlayerID string
	Log      []game.StateDiff
	Chat     []shared.ChatMessage
}

// Describe renders the full state for a model.
func (s *Snapshot) Describe() string {
	parts := []string{SummarizeGameState(&s.View, s.PlayerID)}
	nameOf := func(playerID string) string {
		if playerID == s.PlayerID {
			return "you (" + s.View.CurrentPlayer.Name + ")"
		}
		return findPlayerName(&s.View, playerID)
	}
	if log := formatRecentLog(s.Log, 20, nameOf); log != "" {
		parts = append(parts, log)
	}
	if chat := formatRecentChat(s.Chat, 10); chat != "" {
		parts = append(parts, chat)
	}
	return strings.Join(parts, "\n\n")
}

// Brief is the short status returned after each action.
func (s *Snapshot) Brief() string {
	p := &s.View.CurrentPlayer
	turn := "not your turn"
	if IsMyTurn(&s.View, s.PlayerID) {
		turn = "still your turn"
	}
	pending := GetPendingActionType(&s.View)
	if pending == "" {
		pending = "none"
	}
	return fmt.Sprintf("%s | actions remaining: %d | pending: %s | credits %d, steel %d, titanium %d, plants %d, energy %d, heat %d",
		turn, p.AvailableActions, pending,
		p.Resources.Credits, p.Resources.Steel, p.Resources.Titanium, p.Resources.Plants, p.Resources.Energy, p.Resources.Heat)
}
