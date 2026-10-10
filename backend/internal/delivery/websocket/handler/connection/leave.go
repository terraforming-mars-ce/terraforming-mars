package connection

import (
	"context"
	"log/slog"

	connaction "openmars/internal/action/connection"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/logger"
)

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
	SendInitialLogs(gameID string, playerID string)
	SendInitialLogsToSpectator(gameID string, spectatorID string)
}

// LeaveHandlers returns what the hub runs when the last connection of a player or a
// spectator goes away.
func LeaveHandlers(playerAction *connaction.PlayerDisconnectedAction, spectatorAction *connaction.SpectatorDisconnectedAction, broadcaster Broadcaster) (player, spectator core.LeaveFunc) {
	log := logger.Get()
	player = func(ctx context.Context, gameID, playerID string) {
		if err := playerAction.Execute(ctx, gameID, playerID); err != nil {
			log.Debug("Player left a game that no longer has them", slog.String("game_id", gameID), slog.String("player_id", playerID), slog.Any("error", err))
			return
		}
		broadcaster.BroadcastGameState(gameID, nil)
	}
	spectator = func(ctx context.Context, gameID, spectatorID string) {
		if err := spectatorAction.Execute(ctx, gameID, spectatorID); err != nil {
			log.Debug("Spectator left a game that no longer has them", slog.String("game_id", gameID), slog.String("spectator_id", spectatorID), slog.Any("error", err))
			return
		}
		broadcaster.BroadcastGameState(gameID, nil)
	}
	return player, spectator
}
