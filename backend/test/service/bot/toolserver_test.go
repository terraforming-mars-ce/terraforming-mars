package bot_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"
)

func toolNames(t *testing.T, ctx context.Context, endpoint *bot.MCPEndpoint) map[string]bool {
	t.Helper()
	session, err := connectMCP(ctx, endpoint)
	testutil.AssertNoError(t, err, "client should connect")
	defer func() { _ = session.Close() }()
	list, err := session.ListTools(ctx, nil)
	testutil.AssertNoError(t, err, "tools should list")
	names := map[string]bool{}
	for _, tool := range list.Tools {
		names[tool.Name] = true
	}
	return names
}

func TestToolServer_ExecutorGetsGameAndExpressionTools(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RoleExecutor, &fakeHooks{fx: fx})
	defer revoke()

	names := toolNames(t, ctx, endpoint)
	for _, want := range []string{"get_state", "set_plan", "thought", "chat", "emote", "play_card", "skip_action",
		"colony_trade", "colony_build", "project_fund_seat", "confirm_free_trade", "select_tile"} {
		testutil.AssertTrue(t, names[want], "executor should have tool "+want)
	}
}

func TestToolServer_PlannerCannotTakeActions(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RolePlanner, &fakeHooks{fx: fx})
	defer revoke()

	names := toolNames(t, ctx, endpoint)
	testutil.AssertTrue(t, names["set_plan"], "planner should be able to set its plan")
	testutil.AssertFalse(t, names["skip_action"], "planner must not take game actions")
	testutil.AssertFalse(t, names["chat"], "planner must not chat")
}

func TestToolServer_RejectsUnknownAndRevokedTokens(t *testing.T) {
	fx := newBotFixture(t)
	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RoleExecutor, &fakeHooks{fx: fx})

	for _, token := range []string{"", "not-a-token"} {
		req, _ := http.NewRequest(http.MethodPost, endpoint.URL, strings.NewReader(`{}`))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		testutil.AssertNoError(t, err, "request should complete")
		_ = resp.Body.Close()
		testutil.AssertEqual(t, http.StatusUnauthorized, resp.StatusCode, "unknown token should be rejected")
	}

	revoke()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := connectMCP(ctx, endpoint)
	testutil.AssertError(t, err, "revoked token should be rejected")
}

func TestToolServer_SkipActionRunsTheGameAction(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hooks := &fakeHooks{fx: fx}
	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RoleExecutor, hooks)
	defer revoke()

	result, err := callTool(ctx, endpoint, "skip_action", map[string]any{})
	testutil.AssertNoError(t, err, "tool call should complete")
	testutil.AssertFalse(t, result.IsError, "skip should be accepted: "+resultText(result))
	testutil.AssertTrue(t, strings.HasPrefix(resultText(result), "Done."), "result should report the new state")
	testutil.AssertEqual(t, 1, hooks.actions, "action should be recorded")
	testutil.AssertEqual(t, fx.humanID, fx.game.CurrentTurn().PlayerID(), "turn should pass to the human")
}

func TestToolServer_StandardProjectPaysAutomatically(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p, _ := fx.game.GetPlayer(fx.botID)
	testutil.SetPlayerCredits(ctx, p, 30)
	before := fx.game.GlobalParameters().Temperature()

	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RoleExecutor, &fakeHooks{fx: fx})
	defer revoke()

	result, err := callTool(ctx, endpoint, "standard_project", map[string]any{"project": "asteroid"})
	testutil.AssertNoError(t, err, "tool call should complete")
	testutil.AssertFalse(t, result.IsError, "asteroid should be accepted: "+resultText(result))
	testutil.AssertEqual(t, 16, testutil.GetPlayerCredits(p), "asteroid should cost 14 credits")
	testutil.AssertTrue(t, fx.game.GlobalParameters().Temperature() > before, "temperature should rise")
}

func TestToolServer_RejectedActionIsReportedToTheModel(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p, _ := fx.game.GetPlayer(fx.botID)
	testutil.SetPlayerCredits(ctx, p, 0)

	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RoleExecutor, &fakeHooks{fx: fx})
	defer revoke()

	result, err := callTool(ctx, endpoint, "standard_project", map[string]any{"project": "asteroid"})
	testutil.AssertNoError(t, err, "tool call should complete")
	testutil.AssertTrue(t, result.IsError, "unaffordable project should be rejected")
}

func TestToolServer_SetPlanAndThought(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hooks := &fakeHooks{fx: fx}
	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RolePlanner, hooks)
	defer revoke()

	_, err := callTool(ctx, endpoint, "set_plan", map[string]any{
		"summary":          "Rush Mayor",
		"wantedHexes":      []string{"0,1,-1"},
		"targetMilestones": []string{"mayor"},
		"nextMoves":        []map[string]any{{"move": "city on 0,1,-1", "why": "Cities, cities, cities."}},
	})
	testutil.AssertNoError(t, err, "set_plan should complete")
	plan := hooks.Plan()
	testutil.AssertEqual(t, "Rush Mayor", plan.Summary, "plan summary should be stored")
	testutil.AssertEqual(t, "Cities, cities, cities.", plan.Intent(), "intent comes from the first move")

	_, err = callTool(ctx, endpoint, "thought", map[string]any{"text": "Hmm."})
	testutil.AssertNoError(t, err, "thought should complete")
	testutil.AssertEqual(t, 1, len(hooks.thought), "thought should be shown")
}

func TestToolServer_ClaimMilestoneAcceptsTheDisplayName(t *testing.T) {
	fx := newBotFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	p, _ := fx.game.GetPlayer(fx.botID)
	p.Resources().SetTerraformRating(38)
	testutil.SetPlayerCredits(ctx, p, 20)

	endpoint, revoke := fx.tools.Issue(fx.game.ID(), fx.botID, bot.RoleExecutor, &fakeHooks{fx: fx})
	defer revoke()

	result, err := callTool(ctx, endpoint, "claim_milestone", map[string]any{"milestoneType": "Terraformer"})
	testutil.AssertNoError(t, err, "tool call should complete")
	testutil.AssertFalse(t, result.IsError, "claim should be accepted: "+resultText(result))
	testutil.AssertEqual(t, 12, testutil.GetPlayerCredits(p), "the milestone costs 8")
}
