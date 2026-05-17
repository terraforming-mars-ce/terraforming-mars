package game_lifecycle_test

import (
	"context"
	"testing"

	gameAction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	gamePkg "terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

type updateSettingsTestRig struct {
	action    *gameAction.UpdateGameSettingsAction
	repo      gamePkg.GameRepository
	gameID    string
	playerIDs []string
}

func newUpdateSettingsRig(t *testing.T, players int) *updateSettingsTestRig {
	t.Helper()
	repo := testutil.NewTestGameRepository(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	mapRegistry := testutil.CreateTestMapRegistry()
	logger := testutil.TestLogger()

	createAction := gameAction.NewCreateGameAction(repo, cardRegistry, mapRegistry, logger)
	g, err := createAction.Execute(context.Background(), shared.GameSettings{})
	testutil.AssertNoError(t, err, "create game")

	playerIDs := make([]string, 0, players)
	for i := 0; i < players; i++ {
		pid := "player-" + string(rune('a'+i))
		p, err := g.AddNewPlayer(context.Background(), pid, pid)
		testutil.AssertNoError(t, err, "add player")
		playerIDs = append(playerIDs, p.ID())
	}
	if len(playerIDs) > 0 {
		err := g.SetHostPlayerID(context.Background(), playerIDs[0])
		testutil.AssertNoError(t, err, "set host")
	}

	action := gameAction.NewUpdateGameSettingsAction(repo, cardRegistry, mapRegistry, logger)
	return &updateSettingsTestRig{
		action:    action,
		repo:      repo,
		gameID:    g.ID(),
		playerIDs: playerIDs,
	}
}

func (r *updateSettingsTestRig) host() string {
	return r.playerIDs[0]
}

func (r *updateSettingsTestRig) game(t *testing.T) *gamePkg.Game {
	t.Helper()
	g, err := r.repo.Get(context.Background(), r.gameID)
	testutil.AssertNoError(t, err, "get game")
	return g
}

func boolPtr(b bool) *bool          { return &b }
func intPtr(v int) *int             { return &v }
func strPtr(s string) *string       { return &s }
func packsPtr(p []string) *[]string { return &p }

func TestUpdateGameSettings_HostOnly(t *testing.T) {
	rig := newUpdateSettingsRig(t, 2)
	nonHost := rig.playerIDs[1]

	err := rig.action.Execute(context.Background(), rig.gameID, nonHost, &dto.UpdateGameSettingsRequest{
		DevelopmentMode: boolPtr(true),
	})
	testutil.AssertError(t, err, "non-host should be rejected")
}

func TestUpdateGameSettings_LobbyOnly(t *testing.T) {
	rig := newUpdateSettingsRig(t, 1)
	g := rig.game(t)
	err := g.UpdateStatus(context.Background(), shared.GameStatusActive)
	testutil.AssertNoError(t, err, "update status")

	err = rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		DevelopmentMode: boolPtr(true),
	})
	testutil.AssertError(t, err, "active game should reject settings change")
}

func TestUpdateGameSettings_MaxPlayersBelowJoinedRejected(t *testing.T) {
	rig := newUpdateSettingsRig(t, 3)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		MaxPlayers: intPtr(2),
	})
	testutil.AssertError(t, err, "lowering max below joined count should reject")
}

func TestUpdateGameSettings_MaxPlayersValidRange(t *testing.T) {
	rig := newUpdateSettingsRig(t, 1)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		MaxPlayers: intPtr(0),
	})
	testutil.AssertError(t, err, "max=0 should reject")

	err = rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		MaxPlayers: intPtr(11),
	})
	testutil.AssertError(t, err, "max=11 should reject")

	err = rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		MaxPlayers: intPtr(6),
	})
	testutil.AssertNoError(t, err, "max=6 should succeed")
	testutil.AssertEqual(t, 6, rig.game(t).Settings().MaxPlayers, "max players should update")
}

func TestUpdateGameSettings_CardPacksRequiresBaseGame(t *testing.T) {
	rig := newUpdateSettingsRig(t, 1)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		CardPacks: packsPtr([]string{}),
	})
	testutil.AssertError(t, err, "empty card packs should reject")

	err = rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		CardPacks: packsPtr([]string{"prelude"}),
	})
	testutil.AssertError(t, err, "missing base-game should reject")
}

func TestUpdateGameSettings_DemoOffClearsChoices(t *testing.T) {
	rig := newUpdateSettingsRig(t, 2)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		DemoGame: boolPtr(true),
	})
	testutil.AssertNoError(t, err, "demo on")

	g := rig.game(t)
	p, err := g.GetPlayer(rig.playerIDs[1])
	testutil.AssertNoError(t, err, "get player")
	p.SetPendingDemoChoices(&shared.PendingDemoChoices{CorporationID: "credicor"})
	testutil.AssertTrue(t, p.HasPendingDemoChoices(), "should have choices before toggle")

	err = rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		DemoGame: boolPtr(false),
	})
	testutil.AssertNoError(t, err, "demo off")

	g = rig.game(t)
	p, err = g.GetPlayer(rig.playerIDs[1])
	testutil.AssertNoError(t, err, "get player after toggle")
	testutil.AssertFalse(t, p.HasPendingDemoChoices(), "demo off should clear pending demo choices")
	testutil.AssertEqual(t, (*int)(nil), g.Settings().Temperature, "Temperature override should be cleared")
}

func TestUpdateGameSettings_AllowRandomBuyToggles(t *testing.T) {
	rig := newUpdateSettingsRig(t, 1)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		AllowRandomBuy: boolPtr(true),
	})
	testutil.AssertNoError(t, err, "set allow random buy")
	testutil.AssertTrue(t, rig.game(t).Settings().AllowRandomBuy, "allowRandomBuy should be set")

	err = rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		AllowRandomBuy: boolPtr(false),
	})
	testutil.AssertNoError(t, err, "unset allow random buy")
	testutil.AssertFalse(t, rig.game(t).Settings().AllowRandomBuy, "allowRandomBuy should be cleared")
}

func TestUpdateGameSettings_UnknownMapRejected(t *testing.T) {
	rig := newUpdateSettingsRig(t, 1)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		MapID: strPtr("not-a-real-map"),
	})
	testutil.AssertError(t, err, "unknown map should reject")
}

func TestUpdateGameSettings_DevelopmentModeAndClaudeTokenSet(t *testing.T) {
	rig := newUpdateSettingsRig(t, 1)

	err := rig.action.Execute(context.Background(), rig.gameID, rig.host(), &dto.UpdateGameSettingsRequest{
		DevelopmentMode: boolPtr(false),
		ClaudeAPIKey:    strPtr("sk-ant-test"),
	})
	testutil.AssertNoError(t, err, "update dev + token")

	s := rig.game(t).Settings()
	testutil.AssertFalse(t, s.DevelopmentMode, "dev mode should be off")
	testutil.AssertEqual(t, "sk-ant-test", s.ClaudeAPIKey, "claude token should be set")
}
