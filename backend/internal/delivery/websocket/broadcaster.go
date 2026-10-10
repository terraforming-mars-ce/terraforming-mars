package websocket

import (
	"context"
	"log/slog"
	"sync"

	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game"
	"openmars/internal/game/award"
	"openmars/internal/game/cards"
	"openmars/internal/game/colony"
	"openmars/internal/game/milestone"
	pfRegistry "openmars/internal/game/projectfunding"
	"openmars/internal/game/shared"
	"openmars/internal/game/standardproject"
	"openmars/internal/logger"
)

// BotNotifier lets the bot controller observe broadcasts it reacts to.
type BotNotifier interface {
	OnGameBroadcast(gameID string)
	OnChatMessage(gameID string, chatMsg shared.ChatMessage)
}

// Broadcaster handles game state broadcasting to WebSocket clients
// Called explicitly by WebSocket handlers after actions complete
type Broadcaster struct {
	gameRepo                game.GameRepository
	stateRepo               game.GameStateRepository
	hub                     *core.Hub
	cardRegistry            cards.CardRegistry
	colonyRegistry          colony.ColonyRegistry
	projectFundingRegistry  pfRegistry.ProjectFundingRegistry
	standardProjectRegistry standardproject.StandardProjectRegistry
	awardRegistry           award.AwardRegistry
	milestoneRegistry       milestone.MilestoneRegistry
	availableMaps           []dto.MapInfoDto
	botNotifier             BotNotifier
	logger                  *slog.Logger
	lastBroadcastedSeq      map[string]int64 // gameID -> last broadcasted log sequence
	lastBroadcastedLock     sync.RWMutex
}

// NewBroadcaster creates a broadcaster for explicit broadcasting
func NewBroadcaster(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	hub *core.Hub,
	cardRegistry cards.CardRegistry,
	colonyRegistry colony.ColonyRegistry,
	pfReg pfRegistry.ProjectFundingRegistry,
	stdProjReg standardproject.StandardProjectRegistry,
	awardReg award.AwardRegistry,
	msReg milestone.MilestoneRegistry,
	availableMaps []dto.MapInfoDto,
) *Broadcaster {
	broadcaster := &Broadcaster{
		gameRepo:                gameRepo,
		stateRepo:               stateRepo,
		hub:                     hub,
		cardRegistry:            cardRegistry,
		colonyRegistry:          colonyRegistry,
		projectFundingRegistry:  pfReg,
		standardProjectRegistry: stdProjReg,
		awardRegistry:           awardReg,
		milestoneRegistry:       msReg,
		availableMaps:           availableMaps,
		logger:                  logger.Get(),
		lastBroadcastedSeq:      make(map[string]int64),
	}

	broadcaster.logger.Debug("Broadcaster initialized")

	return broadcaster
}

// SetBotNotifier registers a bot notifier to be called after every broadcast.
func (b *Broadcaster) SetBotNotifier(notifier BotNotifier) {
	b.botNotifier = notifier
}

// BroadcastGameState broadcasts game state to specified players (nil = all players)
// Called explicitly by WebSocket handlers after action execution completes
// Also broadcasts any new log entries since the last broadcast
func (b *Broadcaster) BroadcastGameState(gameID string, playerIDs []string) {
	ctx := context.Background()
	log := b.logger.With(slog.String("game_id", gameID))

	g, err := b.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for broadcast", slog.Any("error", err))
		return
	}

	if g.ResumeLobby() != nil {
		for _, connection := range b.hub.Manager().GameConnections(gameID) {
			playerID := connection.PlayerID()
			message := dto.WebSocketMessage{Type: dto.MessageTypeGameUpdated, GameID: gameID, Payload: dto.GameUpdatedPayload{Game: dto.ToResumeGameDto(g, playerID, b.cardRegistry)}}
			connection.Send(message)
		}
		return
	}
	if playerIDs == nil {
		// Broadcast to all players in the game (skip exited - no active connection)
		players := g.GetAllPlayers()
		playerIDs = make([]string, 0, len(players))
		for _, player := range players {
			if !player.HasExited() {
				playerIDs = append(playerIDs, player.ID())
			}
		}
		log.Debug("Broadcasting to all players", slog.Int("player_count", len(playerIDs)))
	} else {
		// Broadcast to specific players
		log.Debug("Broadcasting to specific players", slog.Any("player_ids", playerIDs))
	}

	for _, playerID := range playerIDs {
		b.sendToPlayer(g, playerID)
	}

	// Broadcast to all spectators (before clearing triggered effects so spectators see them)
	b.broadcastToSpectators(g)

	// Clear triggered effects after all players and spectators have received them
	g.ClearTriggeredEffects()

	// Broadcast any new log entries since the last broadcast
	b.broadcastNewLogs(gameID, playerIDs)

	log.Debug("Broadcast completed", slog.Int("player_count", len(playerIDs)))

	if b.botNotifier != nil {
		b.botNotifier.OnGameBroadcast(gameID)
	}
}

// broadcastNewLogs sends any new log entries to the specified players
func (b *Broadcaster) broadcastNewLogs(gameID string, playerIDs []string) {
	ctx := context.Background()
	log := b.logger.With(slog.String("game_id", gameID))

	// Get the last broadcasted sequence for this game
	b.lastBroadcastedLock.RLock()
	lastSeq := b.lastBroadcastedSeq[gameID]
	b.lastBroadcastedLock.RUnlock()

	// Fetch all logs
	diffs, err := b.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Debug("No logs to broadcast", slog.Any("error", err))
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
		b.hub.SendToPlayer(gameID, playerID, message)
	}

	spectatorConns := b.hub.Manager().SpectatorConnections(gameID)
	for _, conn := range spectatorConns {
		conn.Send(message)
	}

	log.Debug("Broadcasted new logs", slog.Int("log_count", len(newLogs)))
}

// sendToPlayer creates a personalized DTO for a player and sends it via WebSocket
func (b *Broadcaster) sendToPlayer(game *game.Game, playerID string) {
	log := b.logger.With(
		slog.String("game_id", game.ID()),
		slog.String("player_id", playerID),
	)

	gameDto := dto.ToGameDtoFull(game, b.cardRegistry, playerID, dto.Registries{
		ColonyRegistry:          b.colonyRegistry,
		ProjectFundingRegistry:  b.projectFundingRegistry,
		StandardProjectRegistry: b.standardProjectRegistry,
		AwardRegistry:           b.awardRegistry,
		MilestoneRegistry:       b.milestoneRegistry,
		AvailableMaps:           b.availableMaps,
	})

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeGameUpdated,
		GameID: game.ID(),
		Payload: dto.GameUpdatedPayload{
			Game: gameDto,
		},
	}

	b.hub.SendToPlayer(game.ID(), playerID, message)
	log.Debug("Sent personalized game state to player")
}

// SendInitialLogs sends all game logs to a specific player (used on connect/reconnect)
func (b *Broadcaster) SendInitialLogs(gameID string, playerID string) {
	ctx := context.Background()
	log := b.logger.With(
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
	)

	diffs, err := b.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Debug("No logs to send (game may be new)", slog.Any("error", err))
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
			Logs:      logDtos,
			IsHistory: true,
		},
	}

	b.hub.SendToPlayer(gameID, playerID, message)

	log.Debug("Sent initial logs to player", slog.Int("log_count", len(logDtos)))
}

// broadcastToSpectators sends spectator-specific game state to all spectator connections.
func (b *Broadcaster) broadcastToSpectators(g *game.Game) {
	spectators := g.GetAllSpectators()
	if len(spectators) == 0 {
		return
	}

	gameDto := dto.ToSpectatorGameDto(g, b.cardRegistry, b.awardRegistry, b.milestoneRegistry)
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeGameUpdated,
		GameID: g.ID(),
		Payload: dto.GameUpdatedPayload{
			Game: gameDto,
		},
	}

	for _, s := range spectators {
		b.hub.SendToSpectator(g.ID(), s.ID(), message)
	}

	b.logger.Debug("Broadcasted to spectators",
		slog.String("game_id", g.ID()),
		slog.Int("spectator_count", len(spectators)))
}

// BroadcastChatMessage broadcasts a chat message to all players and spectators in a game.
func (b *Broadcaster) BroadcastChatMessage(gameID string, chatMsg shared.ChatMessage) {
	ctx := context.Background()
	log := b.logger.With(slog.String("game_id", gameID))

	g, err := b.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for chat broadcast", slog.Any("error", err))
		return
	}

	b.sendToEveryone(g, dto.WebSocketMessage{
		Type:    dto.MessageTypeChatUpdate,
		GameID:  gameID,
		Payload: dto.ChatUpdatePayload{ChatMessage: dto.ToChatMessageDto(chatMsg)},
	})
	log.Debug("Broadcasted chat message")

	if b.botNotifier != nil {
		b.botNotifier.OnChatMessage(gameID, chatMsg)
	}
}

// BroadcastEmote shows an emote over a player's card for everyone in the game.
func (b *Broadcaster) BroadcastEmote(gameID, playerID, emote string) {
	g, err := b.gameRepo.Get(context.Background(), gameID)
	if err != nil {
		b.logger.Warn("Failed to get game for emote broadcast", slog.String("game_id", gameID), slog.Any("error", err))
		return
	}
	b.sendToEveryone(g, dto.WebSocketMessage{
		Type:    dto.MessageTypeEmote,
		GameID:  gameID,
		Payload: dto.EmotePayload{PlayerID: playerID, Emote: dto.EmoteName(emote)},
	})
}

// BroadcastBotThought shows a bot's thought bubble, or its typing state, for everyone in the game.
func (b *Broadcaster) BroadcastBotThought(gameID, playerID, text string, typing bool) {
	g, err := b.gameRepo.Get(context.Background(), gameID)
	if err != nil {
		b.logger.Warn("Failed to get game for bot thought broadcast", slog.String("game_id", gameID), slog.Any("error", err))
		return
	}
	b.sendToEveryone(g, dto.WebSocketMessage{
		Type:    dto.MessageTypeBotThought,
		GameID:  gameID,
		Payload: dto.BotThoughtPayload{PlayerID: playerID, Text: text, Typing: typing},
	})
}

func (b *Broadcaster) sendToEveryone(g *game.Game, message dto.WebSocketMessage) {
	for _, p := range g.GetAllPlayers() {
		if p.HasExited() {
			continue
		}
		b.hub.SendToPlayer(g.ID(), p.ID(), message)
	}
	for _, conn := range b.hub.Manager().SpectatorConnections(g.ID()) {
		conn.Send(message)
	}
}

// SendInitialLogsToSpectator sends all game logs to a spectator (used on connect).
func (b *Broadcaster) SendInitialLogsToSpectator(gameID string, spectatorID string) {
	ctx := context.Background()
	log := b.logger.With(
		slog.String("game_id", gameID),
		slog.String("spectator_id", spectatorID),
	)

	diffs, err := b.stateRepo.GetDiff(ctx, gameID)
	if err != nil {
		log.Debug("No logs to send to spectator", slog.Any("error", err))
		return
	}

	if len(diffs) == 0 {
		return
	}

	logDtos := dto.ToStateDiffDtos(diffs)
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeLogUpdate,
		GameID: gameID,
		Payload: dto.LogUpdatePayload{
			Logs:      logDtos,
			IsHistory: true,
		},
	}

	b.hub.SendToSpectator(gameID, spectatorID, message)

	log.Debug("Sent initial logs to spectator", slog.Int("log_count", len(logDtos)))
}
