package game

import (
	"context"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// CreateGameHandler handles create game requests.
type CreateGameHandler struct {
	createGameAction *gameaction.CreateGameAction
	broadcaster      Broadcaster
	logger           *zap.Logger
}

// Broadcaster interface for explicit broadcasting
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
	SendInitialLogs(gameID string, playerID string)
}

// NewCreateGameHandler creates a new create game handler.
func NewCreateGameHandler(createGameAction *gameaction.CreateGameAction, broadcaster Broadcaster) *CreateGameHandler {
	return &CreateGameHandler{
		createGameAction: createGameAction,
		broadcaster:      broadcaster,
		logger:           logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *CreateGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🎮 Processing create game request")

	settings := game.GameSettings{
		MaxPlayers: game.DefaultMaxPlayers,
		CardPacks:  game.DefaultCardPacks(),
	}

	if payloadMap, ok := message.Payload.(map[string]interface{}); ok {
		if maxPlayers, ok := payloadMap["maxPlayers"].(float64); ok {
			settings.MaxPlayers = int(maxPlayers)
		}
		if cardPacks, ok := payloadMap["cardPacks"].([]interface{}); ok {
			packs := make([]string, len(cardPacks))
			for i, pack := range cardPacks {
				if packStr, ok := pack.(string); ok {
					packs[i] = packStr
				}
			}
			if len(packs) > 0 {
				settings.CardPacks = packs
			}
		}
	}

	log.Debug("Parsed create game settings",
		zap.Int("max_players", settings.MaxPlayers),
		zap.Strings("card_packs", settings.CardPacks))

	game, err := h.createGameAction.Execute(ctx, settings)
	if err != nil {
		log.Error("Failed to execute create game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Create game action completed successfully",
		zap.String("game_id", game.ID()))

	h.broadcaster.BroadcastGameState(game.ID(), nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: map[string]interface{}{
			"gameId":  game.ID(),
			"success": true,
			"message": "Game created successfully. Join with playerConnect.",
		},
	}

	connection.Send <- response

	log.Info("📤 Sent game created response to client")
}

// sendError sends an error message to the client
func (h *CreateGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
