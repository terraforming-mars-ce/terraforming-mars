package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"

	"go.uber.org/zap"
)

// Broadcaster is used by the controller to broadcast game state after dispatched commands.
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// BotController manages all bot sessions and coordinates the turn-play loop.
type BotController struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	dispatcher   *CommandDispatcher
	broadcaster  Broadcaster
	logger       *zap.Logger
	mu           sync.Mutex
	sessions     map[string]map[string]*BotSession // gameID -> playerID -> session
}

// BotSession holds the state for a single bot player in a game.
type BotSession struct {
	gameID     string
	playerID   string
	botName    string
	model      string
	apiKey     string
	difficulty string
	runDir     string

	invoker       *Invoker
	stateWriter   *StateWriter
	commandReader *CommandReader
	historyWriter *HistoryWriter

	turnCh chan struct{}
	cancel context.CancelFunc
	done   chan struct{}
}

// NewBotController creates a new bot controller.
func NewBotController(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	dispatcher *CommandDispatcher,
	broadcaster Broadcaster,
	logger *zap.Logger,
) *BotController {
	return &BotController{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		dispatcher:   dispatcher,
		broadcaster:  broadcaster,
		logger:       logger,
		sessions:     make(map[string]map[string]*BotSession),
	}
}

// StartBot initializes and starts a bot session for the given player.
func (bc *BotController) StartBot(gameID, playerID, botName, difficulty, speed string, settings game.GameSettings) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if _, exists := bc.sessions[gameID]; !exists {
		bc.sessions[gameID] = make(map[string]*BotSession)
	}
	if _, exists := bc.sessions[gameID][playerID]; exists {
		return fmt.Errorf("bot session already exists for player %s in game %s", playerID, gameID)
	}

	var model string
	switch speed {
	case "fast":
		model = "haiku"
	case "thinker":
		model = "opus"
	default:
		model = "sonnet"
	}

	runDir, err := os.MkdirTemp("", fmt.Sprintf("tm-bot-%s-", playerID[:8]))
	if err != nil {
		return fmt.Errorf("create bot temp dir: %w", err)
	}

	statePath := filepath.Join(runDir, "state.txt")
	commandPath := filepath.Join(runDir, "commands.jsonl")
	historyPath := filepath.Join(runDir, "history.log")

	botLogger := bc.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("bot_name", botName),
	)

	historyWriter, err := NewHistoryWriter(historyPath, botLogger)
	if err != nil {
		os.RemoveAll(runDir)
		return fmt.Errorf("create history writer: %w", err)
	}

	commandReader := NewCommandReader(commandPath, botLogger)
	if err := commandReader.Start(); err != nil {
		historyWriter.Close()
		os.RemoveAll(runDir)
		return fmt.Errorf("start command reader: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if difficulty == "" {
		difficulty = "normal"
	}

	session := &BotSession{
		gameID:        gameID,
		playerID:      playerID,
		botName:       botName,
		model:         model,
		apiKey:        settings.ClaudeAPIKey,
		difficulty:    difficulty,
		runDir:        runDir,
		invoker:       NewInvoker(historyPath, statePath, commandPath, model, settings.ClaudeAPIKey, difficulty, botLogger),
		stateWriter:   NewStateWriter(statePath),
		commandReader: commandReader,
		historyWriter: historyWriter,
		turnCh:        make(chan struct{}, 1),
		cancel:        cancel,
		done:          make(chan struct{}),
	}

	bc.sessions[gameID][playerID] = session

	go bc.runBotLoop(ctx, session)

	bc.logger.Info("🤖 Bot session started",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("bot_name", botName),
		zap.String("model", model))

	return nil
}

// OnGameBroadcast is called by the Broadcaster after every game state broadcast.
// It checks if any bot in the game should take a turn.
func (bc *BotController) OnGameBroadcast(gameID string) {
	bc.mu.Lock()
	gameSessions, exists := bc.sessions[gameID]
	if !exists {
		bc.mu.Unlock()
		return
	}
	sessions := make([]*BotSession, 0, len(gameSessions))
	for _, s := range gameSessions {
		sessions = append(sessions, s)
	}
	bc.mu.Unlock()

	ctx := context.Background()
	g, err := bc.gameRepo.Get(ctx, gameID)
	if err != nil {
		return
	}

	for _, session := range sessions {
		// Skip exited bot players
		p, err := g.GetPlayer(session.playerID)
		if err != nil || p.HasExited() {
			continue
		}

		gameDto := dto.ToGameDto(g, bc.cardRegistry, session.playerID)
		if IsMyTurn(&gameDto, session.playerID) {
			select {
			case session.turnCh <- struct{}{}:
			default:
			}
		}
	}
}

// BotStopper can stop individual bot sessions.
type BotStopper interface {
	StopBot(gameID, playerID string)
}

// BotGameStopper can stop all bot sessions for a game.
type BotGameStopper interface {
	StopAllBotsForGame(gameID string)
}

// StopBot stops a single bot session for a player.
func (bc *BotController) StopBot(gameID, playerID string) {
	bc.mu.Lock()
	gameSessions, exists := bc.sessions[gameID]
	if !exists {
		bc.mu.Unlock()
		return
	}
	session, exists := gameSessions[playerID]
	if !exists {
		bc.mu.Unlock()
		return
	}
	delete(gameSessions, playerID)
	bc.mu.Unlock()

	session.cancel()
	<-session.done
	bc.cleanupSession(session)

	bc.logger.Info("🤖 Bot stopped for player",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))
}

// StopAllBotsForGame stops all bot sessions for a game.
func (bc *BotController) StopAllBotsForGame(gameID string) {
	bc.mu.Lock()
	gameSessions, exists := bc.sessions[gameID]
	if !exists {
		bc.mu.Unlock()
		return
	}
	delete(bc.sessions, gameID)
	bc.mu.Unlock()

	for _, session := range gameSessions {
		session.cancel()
		<-session.done
		bc.cleanupSession(session)
	}

	bc.logger.Info("🤖 All bots stopped for game", zap.String("game_id", gameID))
}

func (bc *BotController) runBotLoop(ctx context.Context, session *BotSession) {
	defer close(session.done)

	log := bc.logger.With(
		zap.String("game_id", session.gameID),
		zap.String("player_id", session.playerID),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info("🤖 Bot loop exiting")
			return
		case <-session.turnCh:
			bc.handleTurn(ctx, session, log)
		}
	}
}

func (bc *BotController) handleTurn(ctx context.Context, session *BotSession, log *zap.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}

		g, err := bc.gameRepo.Get(ctx, session.gameID)
		if err != nil {
			log.Error("Failed to get game", zap.Error(err))
			return
		}

		gameDto := dto.ToGameDto(g, bc.cardRegistry, session.playerID)
		if !IsMyTurn(&gameDto, session.playerID) {
			log.Debug("🤖 Not my turn anymore, stopping")
			return
		}

		log.Info("🤖 Starting turn invocation")

		if bot, err := g.GetPlayer(session.playerID); err == nil {
			bot.SetBotStatus(playerPkg.BotStatusThinking)
			bc.broadcaster.BroadcastGameState(session.gameID, nil)
		}

		summary := SummarizeGameState(&gameDto, session.playerID)
		if err := session.stateWriter.WriteState(summary); err != nil {
			log.Error("Failed to write state", zap.Error(err))
			return
		}

		if err := session.commandReader.Reset(); err != nil {
			log.Error("Failed to reset command reader", zap.Error(err))
			return
		}

		cmdCtx, cmdCancel := context.WithCancel(ctx)
		cmdDone := make(chan struct{})
		go bc.processCommands(cmdCtx, session, cmdDone, log)

		invokeCtx, invokeCancel := context.WithTimeout(ctx, 5*time.Minute)
		err = session.invoker.PlayTurn(invokeCtx, &gameDto, session.playerID)
		invokeCancel()

		if err != nil {
			log.Error("🤖 Claude CLI invocation failed", zap.Error(err))
		}

		// Give time for remaining commands to be processed
		time.Sleep(3 * time.Second)
		cmdCancel()
		<-cmdDone

		log.Info("🤖 Turn invocation complete")

		if g, err := bc.gameRepo.Get(ctx, session.gameID); err == nil {
			if bot, err := g.GetPlayer(session.playerID); err == nil {
				bot.SetBotStatus(playerPkg.BotStatusReady)
				bc.broadcaster.BroadcastGameState(session.gameID, nil)
			}
		}
	}
}

func (bc *BotController) processCommands(ctx context.Context, session *BotSession, done chan struct{}, log *zap.Logger) {
	defer close(done)

	for {
		select {
		case <-ctx.Done():
			return
		case rawCmd, ok := <-session.commandReader.Commands():
			if !ok {
				return
			}
			bc.executeCommand(ctx, session, rawCmd, log)
		}
	}
}

func (bc *BotController) executeCommand(ctx context.Context, session *BotSession, rawCmd json.RawMessage, log *zap.Logger) {
	session.historyWriter.WriteSent("command", rawCmd)

	err := bc.dispatcher.Dispatch(ctx, session.gameID, session.playerID, rawCmd)
	if err != nil {
		log.Error("🤖 Command dispatch failed", zap.Error(err))
		errPayload, _ := json.Marshal(map[string]string{
			"type":  "error",
			"error": err.Error(),
		})
		session.historyWriter.WriteReceived("error", errPayload)
		return
	}

	successPayload, _ := json.Marshal(map[string]string{
		"type": "action-success",
	})
	session.historyWriter.WriteReceived("action-success", successPayload)

	bc.broadcaster.BroadcastGameState(session.gameID, nil)

	if g, err := bc.gameRepo.Get(ctx, session.gameID); err == nil {
		gameDto := dto.ToGameDto(g, bc.cardRegistry, session.playerID)
		summary := SummarizeGameState(&gameDto, session.playerID)
		if err := session.stateWriter.WriteState(summary); err != nil {
			log.Error("Failed to update state file after command", zap.Error(err))
		}
	}
}

func (bc *BotController) cleanupSession(session *BotSession) {
	session.commandReader.Stop()
	session.historyWriter.Close()
	os.RemoveAll(session.runDir)
}
