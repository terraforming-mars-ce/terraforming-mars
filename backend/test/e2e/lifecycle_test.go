package e2e_test

import (
	"fmt"
	"slices"
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/test/e2e/harness"
	"openmars/test/testutil"
)

func TestLifecycle_GamesRunFromLobbyToFinalScores(t *testing.T) {
	t.Parallel()
	for _, players := range [][]string{{"solo"}, {"alice", "bob"}, {"alice", "bob", "carol"}} {
		t.Run(fmt.Sprintf("%d players", len(players)), func(t *testing.T) {
			t.Parallel()
			srv := harness.Start(t)
			game := srv.Play(t, harness.DevMode, players...)
			game.ExpectConverged(t)

			current := game.Ready(t)
			current.StandardProject(t, "power-plant")
			game.ExpectConverged(t)
			if got := current.State(t).CurrentPlayer.Production.Energy; got < 1 {
				t.Fatalf("a power plant should raise energy production, got %d", got)
			}

			game.PassAll(t)
			game.ConfirmProduction(t)
			for _, c := range game.Players {
				c.AwaitState(t, "generation 2", func(g dto.GameDto) bool {
					return g.Generation == 2 && g.CurrentPhase == dto.GamePhaseAction
				})
			}
			game.ExpectConverged(t)

			game.Host().Admin(t, dto.AdminCommandTypeSetGlobalParams, map[string]any{"globalParameters": map[string]any{
				"temperature": 8, "oxygen": 14, "oceans": 9, "venus": 0,
			}})
			game.PassAll(t)
			game.ConfirmProduction(t)
			game.PassAll(t)

			var scores []dto.FinalScoreDto
			for _, c := range game.Players {
				state := c.AwaitState(t, "the game is over", func(g dto.GameDto) bool { return g.Status == dto.GameStatusCompleted })
				if len(state.FinalScores) != len(players) {
					t.Fatalf("%s: expected %d final scores, got %d", c.Name, len(players), len(state.FinalScores))
				}
				if scores == nil {
					scores = state.FinalScores
				} else if fmt.Sprint(scores) != fmt.Sprint(state.FinalScores) {
					t.Fatalf("%s sees different final scores", c.Name)
				}
			}

			for _, c := range game.Players {
				c.Send(t, dto.MessageTypeActionSkipAction, map[string]any{})
				c.AwaitError(t)
			}
		})
	}
}

func TestLifecycle_CorporationFirstActionsResolve(t *testing.T) {
	t.Parallel()
	for _, corp := range []string{"Tharsis Republic", "Inventrix", "Valley Trust", "Vitor"} {
		t.Run(corp, func(t *testing.T) {
			t.Parallel()
			srv := harness.Start(t)
			game := srv.Play(t, harness.DevMode, "alice", "bob")
			player := game.Ready(t)
			before := player.State(t)

			player.Admin(t, dto.AdminCommandTypeSetCorporation, map[string]any{"playerId": player.PlayerID(), "corporationId": testutil.CardID(corp)})
			if game.Ready(t) != player && player.State(t).CurrentPlayer.ForcedFirstAction != nil {
				t.Fatalf("%s's first action is still pending after the turn moved on", corp)
			}
			after := player.State(t)
			if after.CurrentPlayer.ForcedFirstAction != nil {
				t.Fatalf("%s's first action was not resolved", corp)
			}
			if after.CurrentPlayer.Corporation == nil || after.CurrentPlayer.Corporation.ID != testutil.CardID(corp) {
				t.Fatalf("the corporation should now be %s", corp)
			}
			switch corp {
			case "Tharsis Republic":
				if countOwnedCities(after, player.PlayerID()) != countOwnedCities(before, player.PlayerID())+1 {
					t.Fatalf("Tharsis Republic's first action should place a city")
				}
			case "Inventrix":
				if len(after.CurrentPlayer.Cards) <= len(before.CurrentPlayer.Cards) {
					t.Fatalf("Inventrix's first action should draw cards")
				}
			case "Valley Trust":
				if len(after.CurrentPlayer.PlayedCards) <= len(before.CurrentPlayer.PlayedCards) {
					t.Fatalf("Valley Trust's first action should play a prelude")
				}
			case "Vitor":
				if !slices.ContainsFunc(after.Awards, func(a dto.AwardDto) bool { return a.FundedBy != nil && *a.FundedBy == player.PlayerID() }) {
					t.Fatalf("Vitor's first action should fund an award")
				}
			}
			game.ExpectConverged(t)
		})
	}
}

func countOwnedCities(g dto.GameDto, playerID string) int {
	n := 0
	for _, tile := range g.Board.Tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == "city-tile" && tile.OwnerID != nil && *tile.OwnerID == playerID {
			n++
		}
	}
	return n
}
