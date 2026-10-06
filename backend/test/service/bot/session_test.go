package bot_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"terraforming-mars-backend/internal/events"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"
)

type fakeRunner struct {
	mu    sync.Mutex
	calls []bot.Invocation
	fn    func(ctx context.Context, inv bot.Invocation) (bot.Result, error)
}

func (r *fakeRunner) Run(ctx context.Context, inv bot.Invocation) (bot.Result, error) {
	r.mu.Lock()
	r.calls = append(r.calls, inv)
	r.mu.Unlock()
	return r.fn(ctx, inv)
}

func (r *fakeRunner) callsFor(model string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, c := range r.calls {
		if c.Model == model {
			n++
		}
	}
	return n
}

type fakeBroadcaster struct {
	mu       sync.Mutex
	chats    []shared.ChatMessage
	emotes   []string
	thoughts []string
}

func (b *fakeBroadcaster) BroadcastGameState(string, []string) {}

func (b *fakeBroadcaster) BroadcastChatMessage(_ string, msg shared.ChatMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.chats = append(b.chats, msg)
}

func (b *fakeBroadcaster) BroadcastEmote(_, _, emote string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.emotes = append(b.emotes, emote)
}

func (b *fakeBroadcaster) BroadcastBotThought(_, _, text string, _ bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.thoughts = append(b.thoughts, text)
}

func (b *fakeBroadcaster) snapshot() ([]shared.ChatMessage, []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]shared.ChatMessage(nil), b.chats...), append([]string(nil), b.emotes...)
}

const (
	executorModel = "test-executor"
	plannerModel  = "test-planner"
	reactorModel  = "test-reactor"
)

func testConfig() bot.Config {
	cfg := bot.DefaultConfig("strategy")
	cfg.ExecutorModel, cfg.PlannerModel, cfg.ReactorModel = executorModel, plannerModel, reactorModel
	cfg.TurnTimeout, cfg.PlanTimeout, cfg.ReactTimeout = 5*time.Second, 5*time.Second, 5*time.Second
	cfg.ActionGap = 0
	cfg.RetryBackoff = []time.Duration{time.Millisecond}
	cfg.ReactionBatch = 10 * time.Millisecond
	cfg.ReactionCooldown = 0
	cfg.BigEventChance = 1
	return cfg
}

func newController(t *testing.T, fx *botFixture, runner bot.Runner) (*bot.BotController, *fakeBroadcaster) {
	t.Helper()
	catalog, err := bot.NewPersonaCatalog([]bot.Persona{{ID: "rival", Label: "Rival", Voice: "competitive", Names: []string{"SHODAN"}}})
	testutil.AssertNoError(t, err, "catalog should be valid")
	broadcaster := &fakeBroadcaster{}
	bc := bot.NewBotController(fx.repo, fx.stateRepo, fx.registry.Cards, broadcaster, runner, fx.tools, catalog, testConfig(), testutil.TestLogger())
	t.Cleanup(func() { bc.StopAllBotsForGame(fx.game.ID()) })
	return bc, broadcaster
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func startBot(t *testing.T, fx *botFixture, bc *bot.BotController) {
	t.Helper()
	var err error
	fx.read(func() { err = bc.StartBot(fx.game.ID(), fx.botID) })
	testutil.AssertNoError(t, err, "bot should start")
}

type botView struct {
	status playerPkg.BotStatus
	err    string
	spend  float64
}

func readBot(t *testing.T, fx *botFixture) botView {
	t.Helper()
	var v botView
	fx.read(func() {
		p, err := fx.game.GetPlayer(fx.botID)
		testutil.AssertNoError(t, err, "bot should exist")
		v = botView{status: p.BotStatus(), err: p.BotError(), spend: fx.game.BotSpendUSD()}
	})
	return v
}

func TestSession_FailingModelFallsBackToAutopilot(t *testing.T) {
	fx := newBotFixture(t)
	runner := &fakeRunner{fn: func(context.Context, bot.Invocation) (bot.Result, error) {
		return bot.Result{CostUSD: 0.01}, errors.New("rate limited")
	}}
	bc, _ := newController(t, fx, runner)

	startBot(t, fx, bc)

	eventually(t, "the turn to pass to the human", func() bool { return fx.currentTurn() == fx.humanID })
	v := readBot(t, fx)
	testutil.AssertEqual(t, playerPkg.BotStatusFailed, v.status, "bot should be marked failed")
	testutil.AssertTrue(t, v.err != "", "the failure reason should be visible")
	testutil.AssertEqual(t, 2, runner.callsFor(executorModel), "one try plus one retry before autopilot")
	testutil.AssertEqual(t, 0, runner.callsFor(plannerModel), "a failed bot does not plan")
}

func TestSession_ModelActsThroughTools(t *testing.T) {
	fx := newBotFixture(t)
	runner := &fakeRunner{fn: func(ctx context.Context, inv bot.Invocation) (bot.Result, error) {
		if inv.Model == executorModel {
			if _, err := callTool(ctx, inv.MCP, "skip_action", map[string]any{}); err != nil {
				return bot.Result{}, err
			}
		}
		return bot.Result{Text: "done", CostUSD: 0.02}, nil
	}}
	bc, _ := newController(t, fx, runner)

	startBot(t, fx, bc)

	eventually(t, "the turn to pass to the human", func() bool { return fx.currentTurn() == fx.humanID })
	eventually(t, "the planner to run after the turn", func() bool { return runner.callsFor(plannerModel) == 1 })
	eventually(t, "the bot to be ready with spend recorded", func() bool {
		v := readBot(t, fx)
		return v.status == playerPkg.BotStatusReady && v.spend >= 0.04
	})
}

func TestSession_SpendCapUsesAutopilotWithoutTheModel(t *testing.T) {
	fx := newBotFixture(t)
	fx.game.AddBotSpend(10)
	runner := &fakeRunner{fn: func(context.Context, bot.Invocation) (bot.Result, error) {
		return bot.Result{}, nil
	}}
	bc, _ := newController(t, fx, runner)

	startBot(t, fx, bc)

	eventually(t, "the turn to pass to the human", func() bool { return fx.currentTurn() == fx.humanID })
	testutil.AssertEqual(t, "spend cap reached", readBot(t, fx).err, "the cap is the reason")
	testutil.AssertEqual(t, 0, len(runner.calls), "no model calls over the cap")
}

func TestSession_RetryClearsTheFailure(t *testing.T) {
	fx := newBotFixture(t)
	runner := &fakeRunner{fn: func(context.Context, bot.Invocation) (bot.Result, error) {
		return bot.Result{}, errors.New("bad token")
	}}
	bc, _ := newController(t, fx, runner)
	startBot(t, fx, bc)
	eventually(t, "the bot to fail", func() bool { return readBot(t, fx).status == playerPkg.BotStatusFailed })
	eventually(t, "the turn to pass", func() bool { return fx.currentTurn() == fx.humanID })

	var err error
	fx.read(func() { err = bc.RetryBot(context.Background(), fx.game.ID(), "someone-else", fx.botID) })
	testutil.AssertError(t, err, "only the host can retry")

	fx.read(func() { err = bc.RetryBot(context.Background(), fx.game.ID(), fx.game.HostPlayerID(), fx.botID) })
	testutil.AssertNoError(t, err, "host can retry")
	eventually(t, "the bot to be ready again", func() bool {
		v := readBot(t, fx)
		return v.status == playerPkg.BotStatusReady && v.err == ""
	})
}

func TestSession_GameEndStopsTheBotAfterARecap(t *testing.T) {
	fx := newBotFixture(t)
	testutil.AssertNoError(t, fx.game.SetCurrentTurn(context.Background(), fx.humanID, 2), "human should have the turn")
	runner := &fakeRunner{fn: func(_ context.Context, inv bot.Invocation) (bot.Result, error) {
		return bot.Result{Text: "Good game. Rematch?"}, nil
	}}
	bc, broadcaster := newController(t, fx, runner)
	startBot(t, fx, bc)
	testutil.AssertEqual(t, 1, bc.RunningBots(), "bot should be running")

	fx.read(func() {
		events.Publish(fx.game.EventBus(), events.GameEndedEvent{GameID: fx.game.ID(), WinnerID: fx.humanID})
	})

	eventually(t, "the bot to stop", func() bool { return bc.RunningBots() == 0 })
	eventually(t, "the recap and the leave notice", func() bool {
		chats, _ := broadcaster.snapshot()
		return len(chats) == 2
	})
	chats, _ := broadcaster.snapshot()
	testutil.AssertEqual(t, shared.ChatMessageKindRecap, chats[0].Kind, "the recap comes first")
	testutil.AssertEqual(t, shared.ChatMessageKindSystem, chats[1].Kind, "then the bot leaves")
	testutil.AssertTrue(t, strings.HasSuffix(chats[1].Message, "left the game"), "leave notice text")
}

func TestSession_ReactsToChatBetweenTurns(t *testing.T) {
	fx := newBotFixture(t)
	testutil.AssertNoError(t, fx.game.SetCurrentTurn(context.Background(), fx.humanID, 2), "human should have the turn")
	runner := &fakeRunner{fn: func(_ context.Context, inv bot.Invocation) (bot.Result, error) {
		if inv.Model == reactorModel {
			return bot.Result{Text: `{"kind":"emote","emote":"laugh"}`}, nil
		}
		return bot.Result{}, nil
	}}
	bc, broadcaster := newController(t, fx, runner)
	startBot(t, fx, bc)

	bc.OnChatMessage(fx.game.ID(), shared.ChatMessage{SenderID: fx.humanID, SenderName: "Alice", Message: "you are going down", Timestamp: time.Now()})

	eventually(t, "an emote", func() bool {
		_, emotes := broadcaster.snapshot()
		return len(emotes) == 1 && emotes[0] == "laugh"
	})
}

func TestSession_StaysQuietDuringTheShowcase(t *testing.T) {
	fx := newBotFixture(t)
	ctx := context.Background()
	testutil.AssertNoError(t, fx.game.SetCurrentTurn(ctx, fx.humanID, 2), "human should have the turn")
	testutil.AssertNoError(t, fx.game.UpdatePhase(ctx, shared.GamePhaseInitApplyCorp), "showcase phase")
	runner := &fakeRunner{fn: func(context.Context, bot.Invocation) (bot.Result, error) {
		return bot.Result{Text: `{"kind":"emote","emote":"laugh"}`}, nil
	}}
	bc, broadcaster := newController(t, fx, runner)
	startBot(t, fx, bc)

	bc.OnChatMessage(fx.game.ID(), shared.ChatMessage{SenderID: fx.humanID, SenderName: "Alice", Message: "hello bot", Timestamp: time.Now()})
	time.Sleep(200 * time.Millisecond)

	_, emotes := broadcaster.snapshot()
	testutil.AssertEqual(t, 0, len(emotes), "no reactions during the showcase")
	testutil.AssertEqual(t, 0, runner.callsFor(reactorModel), "no reaction model call during the showcase")
}

func TestSession_ReactsWhenAnotherPlayerDestroysItsPlants(t *testing.T) {
	fx := newBotFixture(t)
	ctx := context.Background()
	testutil.AssertNoError(t, fx.game.SetCurrentTurn(ctx, fx.humanID, 2), "human should have the turn")
	p, _ := fx.game.GetPlayer(fx.botID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourcePlant: 8})
	_, err := fx.stateRepo.Write(ctx, fx.game.ID(), fx.game, "setup", shared.SourceTypeGameEvent, "", "")
	testutil.AssertNoError(t, err, "baseline log entry")

	var mu sync.Mutex
	var reactorPrompt string
	runner := &fakeRunner{fn: func(_ context.Context, inv bot.Invocation) (bot.Result, error) {
		if inv.Model == reactorModel {
			mu.Lock()
			reactorPrompt = inv.Prompt
			mu.Unlock()
			return bot.Result{Text: `{"kind":"emote","emote":"angry"}`}, nil
		}
		return bot.Result{}, nil
	}}
	bc, broadcaster := newController(t, fx, runner)
	startBot(t, fx, bc)

	fx.read(func() {
		p.Resources().Add(map[shared.ResourceType]int{shared.ResourcePlant: -6})
		_, _ = fx.stateRepo.Write(ctx, fx.game.ID(), fx.game, "Giant Ice Asteroid", shared.SourceTypeCardPlay, fx.humanID, "Played Giant Ice Asteroid for 36 credits")
	})

	eventually(t, "an angry emote", func() bool {
		_, emotes := broadcaster.snapshot()
		return len(emotes) == 1 && emotes[0] == "angry"
	})
	mu.Lock()
	defer mu.Unlock()
	humanName := ""
	fx.read(func() {
		h, _ := fx.game.GetPlayer(fx.humanID)
		humanName = h.Name()
	})
	testutil.AssertTrue(t, strings.Contains(reactorPrompt, humanName+" took away 6 of your plants"), "the reactor is told who destroyed the plants: "+reactorPrompt)
}
