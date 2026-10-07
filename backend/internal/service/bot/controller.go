package bot

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/cards"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Broadcaster publishes bot activity to the players.
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
	BroadcastChatMessage(gameID string, chatMsg shared.ChatMessage)
	BroadcastEmote(gameID, playerID, emote string)
	BroadcastBotThought(gameID, playerID, text string, typing bool)
}

// Config tunes the bot runtime.
type Config struct {
	ExecutorModel string
	PlannerModel  string
	ReactorModel  string

	TurnTimeout  time.Duration
	PlanTimeout  time.Duration
	ReactTimeout time.Duration

	// ActionGap is the minimum time between two game actions of one bot, so humans can follow them.
	ActionGap time.Duration
	// RetryBackoff holds the wait before each retry; when it runs out the bot falls back to autopilot.
	RetryBackoff []time.Duration
	// MaxStepsPerTurn bounds model calls and autopilot steps in one turn.
	MaxStepsPerTurn int

	ReactionBatch    time.Duration
	ReactionCooldown time.Duration
	BigEventChance   float64
	// BigCardCost is the card cost from which an opponent's card counts as a big moment.
	BigCardCost int
	// ThoughtInterval is the minimum time between two thought bubbles of one bot.
	ThoughtInterval time.Duration
	// IntentThoughtChance is how often a turn opens with the plan's intent as a thought bubble.
	IntentThoughtChance float64
	// RemarkThoughtChance is how often a turn ends with the bot's closing remark as a thought bubble.
	RemarkThoughtChance float64

	ExecutorBudgetUSD float64
	PlannerBudgetUSD  float64
	ReactorBudgetUSD  float64

	Strategy string
}

// DefaultConfig returns production settings.
func DefaultConfig(strategy string) Config {
	return Config{
		ExecutorModel:       "sonnet",
		PlannerModel:        "opus",
		ReactorModel:        "haiku",
		TurnTimeout:         3 * time.Minute,
		PlanTimeout:         4 * time.Minute,
		ReactTimeout:        45 * time.Second,
		ActionGap:           1500 * time.Millisecond,
		RetryBackoff:        []time.Duration{5 * time.Second, 15 * time.Second},
		MaxStepsPerTurn:     40,
		ReactionBatch:       2 * time.Second,
		ReactionCooldown:    30 * time.Second,
		BigEventChance:      0.25,
		BigCardCost:         20,
		ThoughtInterval:     10 * time.Second,
		IntentThoughtChance: 0.5,
		RemarkThoughtChance: 0.5,
		ExecutorBudgetUSD:   1.0,
		PlannerBudgetUSD:    1.5,
		ReactorBudgetUSD:    0.05,
		Strategy:            strategy,
	}
}

// BotController owns every running bot session.
type BotController struct {
	gameRepo     game.GameRepository
	stateRepo    game.GameStateRepository
	cardRegistry cards.CardRegistry
	broadcaster  Broadcaster
	runner       Runner
	tools        *ToolServer
	personas     *PersonaCatalog
	cfg          Config
	logger       *slog.Logger

	mu       sync.Mutex
	sessions map[string]map[string]*botSession

	inspectMu sync.Mutex
	inspect   map[string]inspector

	// leaving holds bots that finished their recap, per game, until every bot of the game has.
	leaving map[string][]string
}

// inspector is an admin connection watching one bot's trace.
type inspector struct {
	gameID   string
	playerID string
	send     func(dto.WebSocketMessage)
	done     <-chan struct{}
}

// NewBotController creates a bot controller.
func NewBotController(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	cardRegistry cards.CardRegistry,
	broadcaster Broadcaster,
	runner Runner,
	tools *ToolServer,
	personas *PersonaCatalog,
	cfg Config,
	logger *slog.Logger,
) *BotController {
	return &BotController{
		gameRepo:     gameRepo,
		stateRepo:    stateRepo,
		cardRegistry: cardRegistry,
		broadcaster:  broadcaster,
		runner:       runner,
		tools:        tools,
		personas:     personas,
		cfg:          cfg,
		logger:       logger,
		sessions:     make(map[string]map[string]*botSession),
		inspect:      make(map[string]inspector),
		leaving:      make(map[string][]string),
	}
}

// announceLeft records that a bot left after the game ended. When the last bot of the game
// has left, one "left the game" line per bot is posted, so they come after every recap.
func (bc *BotController) announceLeft(gameID, playerID string) {
	bc.mu.Lock()
	bc.leaving[gameID] = append(bc.leaving[gameID], playerID)
	if len(bc.sessions[gameID]) > 0 {
		bc.mu.Unlock()
		return
	}
	left := bc.leaving[gameID]
	delete(bc.leaving, gameID)
	bc.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = bc.onGame(ctx, func() {
			g, err := bc.gameRepo.Get(ctx, gameID)
			if err != nil {
				return
			}
			for _, id := range left {
				if bot, err := g.GetPlayer(id); err == nil {
					bc.say(ctx, g, bot, bot.Name()+" left the game", shared.ChatMessageKindSystem)
				}
			}
		})
	}()
}

// InspectBot subscribes an admin connection to one bot's trace and sends the current
// snapshot. An empty playerID unsubscribes. Only the host may inspect, and only in
// development mode, because the trace shows the bot's hand and plans. It runs on the
// executor, like the WebSocket handler that calls it.
func (bc *BotController) InspectBot(ctx context.Context, gameID, requesterID, connectionID, playerID string, send func(dto.WebSocketMessage), done <-chan struct{}) error {
	if playerID == "" {
		bc.inspectMu.Lock()
		delete(bc.inspect, connectionID)
		bc.inspectMu.Unlock()
		return nil
	}
	g, err := bc.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}
	if g.HostPlayerID() != requesterID {
		return fmt.Errorf("only the host can inspect bots")
	}
	if !g.Settings().DevelopmentMode {
		return fmt.Errorf("bot inspection is only available in development mode")
	}
	bc.mu.Lock()
	s := bc.sessions[gameID][playerID]
	bc.mu.Unlock()
	if s == nil {
		return fmt.Errorf("no running bot for player %s", playerID)
	}

	bc.inspectMu.Lock()
	bc.inspect[connectionID] = inspector{gameID: gameID, playerID: playerID, send: send, done: done}
	bc.inspectMu.Unlock()
	send(dto.WebSocketMessage{Type: dto.MessageTypeBotTraceSnapshot, GameID: gameID, Payload: s.trace.snapshot()})
	return nil
}

// publishTrace sends a trace event to every connection inspecting that bot.
func (bc *BotController) publishTrace(gameID string, event dto.BotTraceEventDto) {
	bc.inspectMu.Lock()
	var targets []func(dto.WebSocketMessage)
	for id, in := range bc.inspect {
		select {
		case <-in.done:
			delete(bc.inspect, id)
			continue
		default:
		}
		if in.gameID == gameID && in.playerID == event.PlayerID {
			targets = append(targets, in.send)
		}
	}
	bc.inspectMu.Unlock()
	message := dto.WebSocketMessage{Type: dto.MessageTypeBotTraceEvent, GameID: gameID, Payload: event}
	for _, send := range targets {
		send(message)
	}
}

// AssignIdentity picks a free bot name and its persona for a new bot.
func (bc *BotController) AssignIdentity(seed uint64, takenNames []string) (name, personaID string) {
	return bc.personas.AssignIdentity(seed, takenNames)
}

// PrepareBot checks the bot's credential with a short in-character greeting, then marks the
// bot ready (posting the greeting) or failed. It runs in the background.
func (bc *BotController) PrepareBot(gameID, playerID string) {
	go bc.prepareBot(gameID, playerID)
}

func (bc *BotController) prepareBot(gameID, playerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), bc.cfg.ReactTimeout)
	defer cancel()
	log := bc.logger.With(slog.String("game_id", gameID), slog.String("player_id", playerID))

	var name, personaID, token string
	found := false
	if err := bc.onGame(ctx, func() {
		g, err := bc.gameRepo.Get(ctx, gameID)
		if err != nil {
			return
		}
		bot, err := g.GetPlayer(playerID)
		if err != nil {
			return
		}
		name, personaID, token, found = bot.Name(), bot.BotPersona(), g.Settings().ClaudeOAuthToken, true
	}); err != nil || !found {
		return
	}

	result, runErr := bc.runner.Run(ctx, Invocation{
		Model:        bc.cfg.ReactorModel,
		SystemPrompt: identityPreamble(name, bc.personas.Get(personaID)) + "\n\n" + untrustedTextRule,
		Prompt:       greetingPrompt(),
		Token:        token,
		MaxBudgetUSD: bc.cfg.ReactorBudgetUSD,
	})

	_ = bc.onGame(ctx, func() {
		g, err := bc.gameRepo.Get(ctx, gameID)
		if err != nil {
			return
		}
		bot, err := g.GetPlayer(playerID)
		if err != nil {
			return
		}
		g.AddBotSpend(result.CostUSD)
		if runErr != nil {
			log.Warn("Bot credential check failed", slog.Any("error", runErr))
			bot.SetBotStatus(playerPkg.BotStatusFailed)
			bot.SetBotError(credentialError(runErr))
			bc.broadcaster.BroadcastGameState(gameID, nil)
			return
		}
		bot.SetBotStatus(playerPkg.BotStatusReady)
		bot.SetBotError("")
		bc.broadcaster.BroadcastGameState(gameID, nil)
		if greeting := cleanLine(result.Text, shared.MaxChatMessageLength); greeting != "" {
			bc.say(ctx, g, bot, greeting, shared.ChatMessageKindChat)
		}
	})
}

// onGame runs fn on the executor and reports whether it ran.
func (bc *BotController) onGame(ctx context.Context, fn func()) error {
	return bc.tools.exec.Do(ctx, fn)
}

func credentialError(err error) string {
	if errors.Is(err, ErrAuthentication) {
		return "The Claude token was rejected. Create one with `claude setup-token`."
	}
	return "Claude token check failed: " + truncate(err.Error(), 200)
}

// StartBot starts the session that plays for a bot player. It runs on the executor,
// like the game actions that call it.
func (bc *BotController) StartBot(gameID, playerID string) error {
	ctx := context.Background()
	g, err := bc.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("get game: %w", err)
	}
	bot, err := g.GetPlayer(playerID)
	if err != nil {
		return fmt.Errorf("get bot player: %w", err)
	}
	if bot.BotPersona() == "" {
		bot.SetBotPersona(bc.personaFor(bot.Name()))
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()
	if bc.sessions[gameID] == nil {
		bc.sessions[gameID] = make(map[string]*botSession)
	}
	if _, exists := bc.sessions[gameID][playerID]; exists {
		return fmt.Errorf("bot session already exists for player %s in game %s", playerID, gameID)
	}
	s := newBotSession(bc, g, bot)
	if diffs, err := bc.stateRepo.GetDiff(ctx, gameID); err == nil && len(diffs) > 0 {
		s.seenLogSeq = diffs[len(diffs)-1].SequenceNumber
	}
	bc.sessions[gameID][playerID] = s
	s.start()

	bc.logger.Debug("Bot session started", slog.String("game_id", gameID), slog.String("player_id", playerID), slog.String("persona", bot.BotPersona()))
	return nil
}

func (bc *BotController) personaFor(name string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return bc.personas.AssignPersona(h.Sum64())
}

// OnGameBroadcast wakes every bot of the game so it can check whether it must act.
func (bc *BotController) OnGameBroadcast(gameID string) {
	for _, s := range bc.gameSessions(gameID) {
		s.wake()
	}
}

// OnChatMessage forwards chat to the game's bots so they can answer between turns.
// It runs on the executor, inside the chat broadcast.
func (bc *BotController) OnChatMessage(gameID string, chatMsg shared.ChatMessage) {
	sessions := bc.gameSessions(gameID)
	if len(sessions) == 0 || chatMsg.IsSpectator {
		return
	}
	fromBot := false
	if g, err := bc.gameRepo.Get(context.Background(), gameID); err == nil {
		if sender, err := g.GetPlayer(chatMsg.SenderID); err == nil {
			fromBot = sender.IsBot()
		}
	}
	for _, s := range sessions {
		if chatMsg.SenderID == s.playerID {
			continue
		}
		s.observe(Happening{Kind: HappeningChat, ActorID: chatMsg.SenderID, Detail: chatMsg.Message, FromBot: fromBot, At: chatMsg.Timestamp})
	}
}

// RetryBot clears a failed bot and lets it use the model again. Host only. It runs on
// the executor, like the WebSocket handler that calls it.
func (bc *BotController) RetryBot(ctx context.Context, gameID, requesterID, playerID string) error {
	g, err := bc.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}
	if g.HostPlayerID() != requesterID {
		return fmt.Errorf("only the host can retry a bot")
	}
	bc.mu.Lock()
	s := bc.sessions[gameID][playerID]
	bc.mu.Unlock()
	if s == nil {
		return fmt.Errorf("no running bot for player %s", playerID)
	}
	s.requestRetry()
	return nil
}

// BotStopper can stop individual bot sessions.
type BotStopper interface {
	StopBot(gameID, playerID string)
}

// BotGameStopper can stop all bot sessions for a game.
type BotGameStopper interface {
	StopAllBotsForGame(gameID string)
}

// StopBot stops a single bot session and waits for it to finish.
func (bc *BotController) StopBot(gameID, playerID string) {
	bc.mu.Lock()
	s := bc.sessions[gameID][playerID]
	if s != nil {
		delete(bc.sessions[gameID], playerID)
	}
	bc.mu.Unlock()
	if s != nil {
		s.stop()
	}
}

// StopAllBotsForGame stops every bot session of a game and waits for them to finish.
func (bc *BotController) StopAllBotsForGame(gameID string) {
	bc.mu.Lock()
	sessions := bc.sessions[gameID]
	delete(bc.sessions, gameID)
	bc.mu.Unlock()
	for _, s := range sessions {
		s.stop()
	}
	bc.inspectMu.Lock()
	for id, in := range bc.inspect {
		if in.gameID == gameID {
			delete(bc.inspect, id)
		}
	}
	bc.inspectMu.Unlock()
	bc.logger.Debug("All bots stopped for game", slog.String("game_id", gameID))
}

// RunningBots returns how many bot sessions are running, for tests and diagnostics.
func (bc *BotController) RunningBots() int {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	n := 0
	for _, sessions := range bc.sessions {
		n += len(sessions)
	}
	return n
}

func (bc *BotController) gameSessions(gameID string) []*botSession {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	sessions := make([]*botSession, 0, len(bc.sessions[gameID]))
	for _, s := range bc.sessions[gameID] {
		sessions = append(sessions, s)
	}
	return sessions
}

// forget removes a session that ended on its own, such as after the game ended.
func (bc *BotController) forget(s *botSession) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if bc.sessions[s.gameID][s.playerID] == s {
		delete(bc.sessions[s.gameID], s.playerID)
		if len(bc.sessions[s.gameID]) == 0 {
			delete(bc.sessions, s.gameID)
		}
	}
}

// snapshot reads the game as one bot sees it. It must run on the executor.
func (bc *BotController) snapshot(ctx context.Context, gameID, playerID string) (*Snapshot, error) {
	g, err := bc.gameRepo.Get(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("get game: %w", err)
	}
	p, err := g.GetPlayer(playerID)
	if err != nil {
		return nil, fmt.Errorf("get bot player: %w", err)
	}
	snap := &Snapshot{
		Game:     g,
		Player:   p,
		View:     dto.ToGameDto(g, bc.cardRegistry, playerID),
		PlayerID: playerID,
		Chat:     g.GetChatMessages(),
	}
	if diffs, err := bc.stateRepo.GetDiff(ctx, gameID); err == nil {
		snap.Log = diffs
	}
	return snap, nil
}

// say posts a chat message from a bot. It must run on the executor.
func (bc *BotController) say(ctx context.Context, g *game.Game, bot *playerPkg.Player, text string, kind shared.ChatMessageKind) {
	msg := shared.ChatMessage{
		SenderID:    bot.ID(),
		SenderName:  bot.Name(),
		SenderColor: bot.Color(),
		Message:     text,
		Timestamp:   time.Now(),
		Kind:        kind,
	}
	g.AddChatMessage(ctx, msg)
	bc.broadcaster.BroadcastChatMessage(g.ID(), msg)
}
