package e2e_test

import (
	"net/http"
	"strings"
	"testing"

	"openmars/internal/delivery/dto"
	"openmars/test/e2e/harness"
)

func TestHTTP_ServerInformation(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)

	var health map[string]any
	if status := srv.Request(t, http.MethodGet, "/api/v1/health", nil, &health); status != http.StatusOK {
		t.Fatalf("health: status %d", status)
	}
	var meta dto.MetaResponse
	if status := srv.Request(t, http.MethodGet, "/api/v1/meta", nil, &meta); status != http.StatusOK || meta.Alias != "test" {
		t.Fatalf("meta: status %d, %+v", status, meta)
	}
	options := srv.GameOptions(t)
	if options.Defaults.DevelopmentMode {
		t.Fatalf("new games should not default to development mode")
	}
	if len(options.Defaults.CardPacks) == 0 || len(options.AvailableMaps) == 0 {
		t.Fatalf("game options should offer card packs and maps: %+v", options.Defaults)
	}
	if status := srv.Request(t, http.MethodGet, "/api/v1/no-such-route", nil, nil); status != http.StatusNotFound {
		t.Fatalf("an unknown route should be 404, got %d", status)
	}
	if status := srv.Request(t, http.MethodOptions, "/api/v1/games", nil, nil); status != http.StatusNoContent {
		t.Fatalf("a CORS preflight should be 204, got %d", status)
	}
}

func TestHTTP_CreatingGamesValidatesSettings(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	defaults := srv.GameOptions(t).Defaults

	for name, edit := range map[string]func(*dto.GameSetupDto){
		"no players allowed": func(s *dto.GameSetupDto) { s.MaxPlayers = 0 },
		"an unknown map":     func(s *dto.GameSetupDto) { s.MapID = "no-such-map" },
		"an unknown pack":    func(s *dto.GameSetupDto) { s.CardPacks = []string{"no-such-pack"} },
	} {
		settings := defaults
		settings.CardPacks = append([]string(nil), defaults.CardPacks...)
		edit(&settings)
		var rejected dto.ErrorResponse
		if status := srv.Request(t, http.MethodPost, "/api/v1/games", dto.CreateGameRequest{Settings: &settings}, &rejected); status != http.StatusBadRequest || rejected.Error == "" {
			t.Fatalf("a game with %s should be rejected with a reason, got %d %+v", name, status, rejected)
		}
	}
	if status := srv.Request(t, http.MethodPost, "/api/v1/games", "not an object", nil); status != http.StatusBadRequest {
		t.Fatalf("a malformed body should be 400, got %d", status)
	}

	var created dto.CreateGameResponse
	if status := srv.Request(t, http.MethodPost, "/api/v1/games", dto.CreateGameRequest{}, &created); status != http.StatusCreated || created.Game.ID == "" {
		t.Fatalf("a game without settings should use the defaults: %d", status)
	}
	if created.Game.Settings.MaxPlayers != defaults.MaxPlayers {
		t.Fatalf("defaults were not applied: %+v", created.Game.Settings)
	}
}

func TestHTTP_GamesAndPlayersCanBeLookedUp(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, nil, "alice", "bob")
	alice := game.Players[0]

	var list dto.ListGamesResponse
	srv.Request(t, http.MethodGet, "/api/v1/games", nil, &list)
	if len(list.Games) != 1 || list.Games[0].ID != game.ID {
		t.Fatalf("the game should be listed: %+v", list.Games)
	}

	var got dto.GetGameResponse
	if status := srv.Request(t, http.MethodGet, "/api/v1/games/"+game.ID+"?playerId="+alice.PlayerID(), nil, &got); status != http.StatusOK {
		t.Fatalf("get game: status %d", status)
	}
	if got.Game.CurrentPlayer.ID != alice.PlayerID() {
		t.Fatalf("asking as alice should show alice's view")
	}
	for path, want := range map[string]int{
		"/api/v1/games/no-such-game":                                  http.StatusNotFound,
		"/api/v1/games/" + game.ID + "?playerId=no-such-player":       http.StatusNotFound,
		"/api/v1/games/" + game.ID + "/players/" + alice.PlayerID():   http.StatusOK,
		"/api/v1/games/" + game.ID + "/players/no-such-player":        http.StatusNotFound,
		"/api/v1/games/no-such-game/players/" + alice.PlayerID():      http.StatusNotFound,
		"/api/v1/games/" + game.ID + "/logs":                          http.StatusOK,
		"/api/v1/games/no-such-game/logs":                             http.StatusNotFound,
		"/api/v1/games/" + game.ID + "/history":                       http.StatusOK,
		"/api/v1/games/" + game.ID + "/history?policy=no-such-policy": http.StatusBadRequest,
		"/api/v1/games/" + game.ID + "/history?phases=no-such-phase":  http.StatusBadRequest,
		"/api/v1/games/" + game.ID + "/history?policy=everyTurn":      http.StatusOK,
		"/api/v1/games/no-such-game/history":                          http.StatusNotFound,
	} {
		if status := srv.Request(t, http.MethodGet, path, nil, nil); status != want {
			t.Errorf("GET %s: want %d, got %d", path, want, status)
		}
	}
}

func TestHTTP_AnonymousGameViewsRevealNoPrivateCards(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	game := srv.Play(t, harness.DevMode, "alice", "bob")
	alice := game.Players[0]
	secret := cardNobodyHolds(t, game, "Giant Ice Asteroid", "Comet", "Big Asteroid")
	alice.Admin(t, dto.AdminCommandTypeGiveCard, map[string]any{"playerId": alice.PlayerID(), "cardId": secret})

	for _, path := range []string{"/api/v1/games", "/api/v1/games/" + game.ID} {
		var raw map[string]any
		srv.Request(t, http.MethodGet, path, nil, &raw)
		if body := mustJSON(raw); strings.Contains(body, `"`+secret+`"`) {
			t.Fatalf("GET %s without a player reveals alice's hand", path)
		}
	}

	var got dto.GetGameResponse
	srv.Request(t, http.MethodGet, "/api/v1/games/"+game.ID, nil, &got)
	if got.Game.CurrentPlayer.ID != "" || len(got.Game.OtherPlayers) != 2 {
		t.Fatalf("an anonymous view should list every player publicly, got viewer %q and %d others", got.Game.CurrentPlayer.ID, len(got.Game.OtherPlayers))
	}
}

func TestHTTP_CardsAndAchievementsAreListed(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	var page dto.ListCardsResponse
	if status := srv.Request(t, http.MethodGet, "/api/v1/cards?offset=0&limit=10", nil, &page); status != http.StatusOK {
		t.Fatalf("list cards: status %d", status)
	}
	if len(page.Cards) != 10 || page.TotalCount <= 10 {
		t.Fatalf("expected a page of 10 out of all cards, got %d of %d", len(page.Cards), page.TotalCount)
	}
	var next dto.ListCardsResponse
	srv.Request(t, http.MethodGet, "/api/v1/cards?offset=10&limit=10", nil, &next)
	if len(next.Cards) == 0 || next.Cards[0].ID == page.Cards[0].ID {
		t.Fatalf("the second page should continue where the first ended")
	}
	var achievements dto.ListMilestonesAwardsResponse
	if status := srv.Request(t, http.MethodGet, "/api/v1/milestones-awards", nil, &achievements); status != http.StatusOK || len(achievements.Milestones) == 0 || len(achievements.Awards) == 0 {
		t.Fatalf("milestones and awards should be listed: %d", status)
	}
}

func TestHTTP_BugReportsWithoutGitHubAreUnavailable(t *testing.T) {
	t.Parallel()
	srv := harness.Start(t)
	var status dto.FeedbackStatusResponse
	srv.Request(t, http.MethodGet, "/api/v1/bugs/status", nil, &status)
	if status.Available {
		t.Fatalf("bug reports need GitHub credentials, which the test server has none of")
	}
	if code := srv.Request(t, http.MethodPost, "/api/v1/bugs", map[string]any{"description": "broken"}, nil); code != http.StatusServiceUnavailable {
		t.Fatalf("submitting without GitHub should be 503, got %d", code)
	}
	if code := srv.Request(t, http.MethodGet, "/api/v1/bugs/no-such-report", nil, nil); code != http.StatusNotFound {
		t.Fatalf("an unknown report should be 404, got %d", code)
	}
}
