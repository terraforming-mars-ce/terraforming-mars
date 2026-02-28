package websocket

import (
	"context"
	"sync"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster handles game state broadcasting to WebSocket clients
// Called explicitly by WebSocket handlers after actions complete
type Broadcaster struct {
	gameRepo            game.GameRepository
	stateRepo           game.GameStateRepository
	hub                 *core.Hub
	cardRegistry        cards.CardRegistry
	logger              *zap.Logger
	lastBroadcastedSeq  map[string]int64 // gameID -> last broadcasted log sequence
	lastBroadcastedLock sync.RWMutex
}

// NewBroadcaster creates a broadcaster for explicit broadcasting
func NewBroadcaster(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	hub *core.Hub,
	cardRegistry cards.CardRegistry,
) *Broadcaster {
	broadcaster := &Broadcaster{
		gameRepo:           gameRepo,
		stateRepo:          stateRepo,
		hub:                hub,
		cardRegistry:       cardRegistry,
		logger:             logger.Get(),
		lastBroadcastedSeq: make(map[string]int64),
	}

	broadcaster.logger.Info("📡 Broadcaster initialized")

	return broadcaster
}

// BroadcastGameState broadcasts game state to specified players (nil = all players)
// Called explicitly by WebSocket handlers after action execution completes
// Also broadcasts any new log entries since the last broadcast
func (b *Broadcaster) BroadcastGameState(gameID string, playerIDs []string) {
	ctx := context.Background()
	log := b.logger.With(zap.String("game_id", gameID))

	g, err := b.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for broadcast", zap.Error(err))
		return
	}

	if playerIDs == nil {
		// Broadcast to all players in the game
		players := g.GetAllPlayers()
		playerIDs = make([]string, len(players))
		for i, player := range players {
			playerIDs[i] = player.ID()
		}
		log.Debug("📢 Broadcasting to all players", zap.Int("player_count", len(playerIDs)))
	} else {
		// Broadcast to specific players
		log.Debug("📢 Broadcasting to specific players", zap.Strings("player_ids", playerIDs))
	}

	for _, playerID := range playerIDs {
		if err := b.sendToPlayer(ctx, g, playerID); err != nil {
			log.Error("Failed to send game state to player",
				zap.String("player_id", playerID),
				zap.Error(err))
			// Continue with other players even if one fails
		}
	}

	// Clear triggered effects after all players have received them
	g.ClearTriggeredEffects()

	// Broadcast any new log entries since the last broadcast
	b.broadcastNewLogs(gameID, playerIDs)

	log.Debug("✅ Broadcast completed", zap.Int("player_count", len(playerIDs)))
}

// broadcastNewLogs sends any new log entries to the specified players
func (b *Broadcaster) broadcastNewLogs(gameID string, playerIDs []string) {
	ctx := context.Background()
	log := b.logger.With(zap.String("game_id", gameID))

	// Get the last broadcasted sequence for this game
	b.lastBroadcastedLock.RLock()
	lastSeq := b.lastBroadcastedSeq[gameID]
	b.lastBroadcastedLock.RUnlock()

	// Fetch all logs
	diffs, err := b.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Debug("No logs to broadcast", zap.Error(err))
		return
	}

	// Filter to only new logs
	var newLogs []game.StateDiff
	var maxSeq int64
	for _, diff := range diffs {
		if diff.SequenceNumber > lastSeq {
			newLogs = append(newLogs, diff)
		}
		if diff.SequenceNumber > maxSeq {
			maxSeq = diff.SequenceNumber
		}
	}

	if len(newLogs) == 0 {
		return
	}

	// Update the last broadcasted sequence
	b.lastBroadcastedLock.Lock()
	b.lastBroadcastedSeq[gameID] = maxSeq
	b.lastBroadcastedLock.Unlock()

	// Convert to DTOs and broadcast
	logDtos := dto.ToStateDiffDtos(newLogs)
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeLogUpdate,
		GameID: gameID,
		Payload: dto.LogUpdatePayload{
			Logs: logDtos,
		},
	}

	for _, playerID := range playerIDs {
		if err := b.hub.SendToPlayer(gameID, playerID, message); err != nil {
			log.Error("Failed to send log update to player",
				zap.String("player_id", playerID),
				zap.Error(err))
		}
	}

	log.Debug("📜 Broadcasted new logs", zap.Int("log_count", len(newLogs)))
}

// sendToPlayer creates a personalized DTO for a player and sends it via WebSocket
func (b *Broadcaster) sendToPlayer(ctx context.Context, game *game.Game, playerID string) error {
	log := b.logger.With(
		zap.String("game_id", game.ID()),
		zap.String("player_id", playerID),
	)

	gameDto := dto.ToGameDto(game, b.cardRegistry, playerID)

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeGameUpdated,
		GameID: game.ID(),
		Payload: dto.GameUpdatedPayload{
			Game: gameDto,
		},
	}

	if err := b.hub.SendToPlayer(game.ID(), playerID, message); err != nil {
		return err
	}

	log.Debug("✅ Sent personalized game state to player")
	return nil
}

// SendInitialLogs sends all game logs to a specific player (used on connect/reconnect)
func (b *Broadcaster) SendInitialLogs(gameID string, playerID string) {
	ctx := context.Background()
	log := b.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
	)

	diffs, err := b.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Debug("No logs to send (game may be new)", zap.Error(err))
		return
	}

	if len(diffs) == 0 {
		log.Debug("No logs to send")
		return
	}

	logDtos := dto.ToStateDiffDtos(diffs)

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeLogUpdate,
		GameID: gameID,
		Payload: dto.LogUpdatePayload{
			Logs: logDtos,
		},
	}

	if err := b.hub.SendToPlayer(gameID, playerID, message); err != nil {
		log.Error("Failed to send initial logs", zap.Error(err))
		return
	}

	log.Debug("📜 Sent initial logs to player", zap.Int("log_count", len(logDtos)))
}

// BroadcastLogUpdate broadcasts a single log entry to all players in a game
func (b *Broadcaster) BroadcastLogUpdate(gameID string, logEntry *game.StateDiff) {
	ctx := context.Background()
	log := b.logger.With(zap.String("game_id", gameID))

	g, err := b.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for log broadcast", zap.Error(err))
		return
	}

	logDto := dto.ToStateDiffDto(logEntry)
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeLogUpdate,
		GameID: gameID,
		Payload: dto.LogUpdatePayload{
			Logs: []dto.StateDiffDto{logDto},
		},
	}

	players := g.GetAllPlayers()
	for _, player := range players {
		if err := b.hub.SendToPlayer(gameID, player.ID(), message); err != nil {
			log.Error("Failed to send log update to player",
				zap.String("player_id", player.ID()),
				zap.Error(err))
		}
	}

	log.Debug("📜 Broadcasted log update", zap.Int64("sequence", logEntry.SequenceNumber))
}
