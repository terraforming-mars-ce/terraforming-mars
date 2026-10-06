package bot_test

import (
	"context"
	"sync"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"
)

type inspectorConn struct {
	mu       sync.Mutex
	messages []dto.WebSocketMessage
}

func (c *inspectorConn) send(m dto.WebSocketMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, m)
}

func (c *inspectorConn) events() []dto.BotTraceEventDto {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []dto.BotTraceEventDto
	for _, m := range c.messages {
		if e, ok := m.Payload.(dto.BotTraceEventDto); ok {
			out = append(out, e)
		}
	}
	return out
}

func enableDevMode(fx *botFixture) {
	fx.read(func() {
		s := fx.game.Settings()
		s.DevelopmentMode = true
		fx.game.UpdateSettings(context.Background(), s)
	})
}

func inspect(fx *botFixture, bc *bot.BotController, requester string, conn *inspectorConn) error {
	var err error
	fx.read(func() {
		err = bc.InspectBot(context.Background(), fx.game.ID(), requester, "conn-1", fx.botID, conn.send, make(chan struct{}))
	})
	return err
}

func TestInspectBot_RequiresHostAndDevelopmentMode(t *testing.T) {
	fx := newBotFixture(t)
	testutil.AssertNoError(t, fx.game.SetCurrentTurn(context.Background(), fx.humanID, 2), "human should have the turn")
	bc, _ := newController(t, fx, &fakeRunner{fn: func(context.Context, bot.Invocation) (bot.Result, error) { return bot.Result{}, nil }})
	startBot(t, fx, bc)
	conn := &inspectorConn{}

	testutil.AssertError(t, inspect(fx, bc, fx.game.HostPlayerID(), conn), "development mode is required")
	enableDevMode(fx)
	testutil.AssertError(t, inspect(fx, bc, "someone-else", conn), "only the host may inspect")
	testutil.AssertNoError(t, inspect(fx, bc, fx.game.HostPlayerID(), conn), "host in development mode may inspect")
	testutil.AssertEqual(t, 1, len(conn.messages), "a snapshot is sent")
	testutil.AssertEqual(t, dto.MessageTypeBotTraceSnapshot, conn.messages[0].Type, "first message is the snapshot")
}

func TestInspectBot_StreamsCallsToolsAndPlan(t *testing.T) {
	fx := newBotFixture(t)
	enableDevMode(fx)
	testutil.AssertNoError(t, fx.game.SetCurrentTurn(context.Background(), fx.humanID, 2), "human should have the turn first")
	runner := &fakeRunner{fn: func(ctx context.Context, inv bot.Invocation) (bot.Result, error) {
		switch inv.Model {
		case executorModel:
			inv.OnText("Passing for now.")
			if _, err := callTool(ctx, inv.MCP, "skip_action", map[string]any{}); err != nil {
				return bot.Result{}, err
			}
		case plannerModel:
			if _, err := callTool(ctx, inv.MCP, "set_plan", map[string]any{"summary": "Rush Mayor", "wantedHexes": []string{"0,1,-1"}}); err != nil {
				return bot.Result{}, err
			}
		}
		return bot.Result{CostUSD: 0.01}, nil
	}}
	bc, _ := newController(t, fx, runner)
	startBot(t, fx, bc)
	conn := &inspectorConn{}
	testutil.AssertNoError(t, inspect(fx, bc, fx.game.HostPlayerID(), conn), "inspect should start")

	fx.read(func() { _ = fx.game.SetCurrentTurn(context.Background(), fx.botID, 2) })
	bc.OnGameBroadcast(fx.game.ID())

	eventually(t, "the plan to be streamed", func() bool {
		for _, e := range conn.events() {
			if e.Kind == "plan" && e.Plan.Summary == "Rush Mayor" {
				return true
			}
		}
		return false
	})
	kinds := map[string]bool{}
	tools := map[string]bool{}
	for _, e := range conn.events() {
		kinds[e.Kind] = true
		if e.Kind == "call-step" && e.Step.Kind == "tool" {
			tools[e.Step.Tool] = true
		}
		if e.Kind == "call-step" && e.Step.Kind == "text" {
			testutil.AssertEqual(t, "Passing for now.", e.Step.Text, "reasoning text is recorded")
		}
	}
	for _, k := range []string{"call-start", "call-step", "call-end", "plan"} {
		testutil.AssertTrue(t, kinds[k], "event kind "+k+" should be streamed")
	}
	testutil.AssertTrue(t, tools["skip_action"], "the executor's tool call is recorded")
	testutil.AssertTrue(t, tools["set_plan"], "the planner's tool call is recorded")
}
