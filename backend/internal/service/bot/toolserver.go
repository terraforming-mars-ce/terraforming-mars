package bot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	awardAction "terraforming-mars-backend/internal/action/award"
	cardAction "terraforming-mars-backend/internal/action/card"
	colonyAction "terraforming-mars-backend/internal/action/colony"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	pfAction "terraforming-mars-backend/internal/action/projectfunding"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/award"
	"terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/milestone"
	"terraforming-mars-backend/internal/game/standardproject"
)

// Actions are the game actions a bot can take. They are the same actions human handlers call.
type Actions struct {
	PlayCard               *cardAction.PlayCardAction
	UseCardAction          *cardAction.UseCardActionAction
	SkipAction             *turnAction.SkipActionAction
	SelectStartingChoices  *turnAction.SelectStartingChoicesAction
	ConfirmInitAdvance     *turnAction.ConfirmInitAdvanceAction
	SelectTile             *tileAction.SelectTileAction
	ConfirmProductionCards *confirmAction.ConfirmProductionCardsAction
	ConfirmCardDraw        *confirmAction.ConfirmCardDrawAction
	ConfirmCardDiscard     *confirmAction.ConfirmCardDiscardAction
	ConfirmBehaviorChoice  *confirmAction.ConfirmBehaviorChoiceAction
	ConfirmEffectSelection *confirmAction.ConfirmEffectSelectionAction
	ConfirmCardReveal      *confirmAction.ConfirmCardRevealAction
	ConfirmSellPatents     *confirmAction.ConfirmSellPatentsAction
	ConfirmResourceRemoval *confirmAction.ConfirmResourceRemovalAction
	ConfirmColonyPlacement *confirmAction.ConfirmColonyPlacementAction
	ConfirmColonyResource  *confirmAction.ConfirmColonyResourceAction
	ConfirmAwardFund       *confirmAction.ConfirmAwardFundAction
	ConfirmFreeTrade       *confirmAction.ConfirmFreeTradeAction
	ExecuteStandardProject *stdprojAction.ExecuteStandardProjectAction
	ConvertHeat            *resconvAction.ConvertHeatToTemperatureAction
	ConvertPlants          *resconvAction.ConvertPlantsToGreeneryAction
	ClaimMilestone         *milestoneAction.ClaimMilestoneAction
	FundAward              *awardAction.FundAwardAction
	ColonyTrade            *colonyAction.TradeAction
	ColonyBuild            *colonyAction.BuildColonyAction
	FundSeat               *pfAction.FundSeatAction
}

// Registries are the lookups needed to quote payments.
type Registries struct {
	Cards            cards.CardRegistry
	StandardProjects standardproject.StandardProjectRegistry
	Milestones       milestone.MilestoneRegistry
	Awards           award.AwardRegistry
}

// Executor runs functions where game state may be read and changed: the WebSocket hub's
// goroutine in production. Every bot access to game state goes through it.
type Executor interface {
	Do(ctx context.Context, fn func()) error
}

// Role decides which tools an invocation may use.
type Role string

const (
	// RoleExecutor plays the bot's turn and may take game actions.
	RoleExecutor Role = "executor"
	// RolePlanner only reads the game and records a plan.
	RolePlanner Role = "planner"
)

// SessionHooks connect tools to the bot session that owns the invocation. Think, Emote,
// Say, AfterAction and Snapshot touch game state and are only called on the executor.
type SessionHooks interface {
	Plan() Plan
	SetPlan(plan Plan)
	Think(text string)
	Emote(name string) error
	Say(ctx context.Context, text string) error
	// BeforeAction blocks until the bot may take its next game action.
	BeforeAction(ctx context.Context) error
	// AfterAction records a successful game action and publishes the new state.
	AfterAction()
	Snapshot(ctx context.Context) (*Snapshot, error)
	// RecordTool records a finished tool call for the admin inspector.
	RecordTool(name, input, result string, isError bool)
}

type grant struct {
	gameID   string
	playerID string
	role     Role
	hooks    SessionHooks
	server   *mcp.Server
	// actionMu serialises game actions when the model issues parallel tool calls.
	actionMu sync.Mutex
}

// ToolServer serves the game tools to model invocations over MCP on a loopback-only listener.
// Every invocation gets its own bearer token, bound to one game, player and role.
type ToolServer struct {
	actions    Actions
	registries Registries
	gameRepo   game.GameRepository
	exec       Executor
	logger     *slog.Logger

	mu     sync.Mutex
	grants map[string]*grant

	listener net.Listener
	server   *http.Server
}

// NewToolServer creates a tool server. Call Start before issuing grants.
func NewToolServer(actions Actions, registries Registries, gameRepo game.GameRepository, exec Executor, logger *slog.Logger) *ToolServer {
	return &ToolServer{
		actions:    actions,
		registries: registries,
		gameRepo:   gameRepo,
		exec:       exec,
		logger:     logger,
		grants:     make(map[string]*grant),
	}
}

// Start listens on a random loopback port and serves until ctx is cancelled.
func (ts *ToolServer) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen for bot tools: %w", err)
	}
	ts.listener = listener

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		if g := ts.lookup(r); g != nil {
			return g.server
		}
		return nil
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})

	ts.server = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ts.lookup(r) == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			mcpHandler.ServeHTTP(w, r)
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := ts.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ts.logger.Error("Bot tool server stopped", slog.Any("error", err))
		}
	}()
	go func() {
		<-ctx.Done()
		if err := ts.server.Close(); err != nil {
			ts.logger.Warn("Failed to close bot tool server", slog.Any("error", err))
		}
	}()

	ts.logger.Debug("Bot tool server listening", slog.String("addr", listener.Addr().String()))
	return nil
}

// URL is the endpoint invocations connect to.
func (ts *ToolServer) URL() string {
	return "http://" + ts.listener.Addr().String() + "/mcp"
}

// Issue creates a grant for one invocation and returns its endpoint and a revoke function.
func (ts *ToolServer) Issue(gameID, playerID string, role Role, hooks SessionHooks) (*MCPEndpoint, func()) {
	g := &grant{gameID: gameID, playerID: playerID, role: role, hooks: hooks}
	g.server = mcp.NewServer(&mcp.Implementation{Name: "terraforming-mars", Version: "1"}, nil)
	g.server.AddReceivingMiddleware(recordToolCalls(hooks))
	ts.registerTools(g)

	token := newToken()
	ts.mu.Lock()
	ts.grants[token] = g
	ts.mu.Unlock()

	return &MCPEndpoint{URL: ts.URL(), Bearer: token}, func() {
		ts.mu.Lock()
		delete(ts.grants, token)
		ts.mu.Unlock()
	}
}

func (ts *ToolServer) lookup(r *http.Request) *grant {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return nil
	}
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.grants[token]
}

// recordToolCalls reports every tool call and its outcome to the session's trace.
func recordToolCalls(hooks SessionHooks) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			res, err := next(ctx, method, req)
			if method != "tools/call" {
				return res, err
			}
			params, ok := req.GetParams().(*mcp.CallToolParamsRaw)
			if !ok {
				return res, err
			}
			text, isError := "", err != nil
			if err != nil {
				text = err.Error()
			} else if result, ok := res.(*mcp.CallToolResult); ok {
				isError = result.IsError
				for _, c := range result.Content {
					if t, ok := c.(*mcp.TextContent); ok {
						text = t.Text
						break
					}
				}
			}
			hooks.RecordTool(params.Name, string(params.Arguments), text, isError)
			return res, err
		}
	}
}

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return hex.EncodeToString(b)
}
