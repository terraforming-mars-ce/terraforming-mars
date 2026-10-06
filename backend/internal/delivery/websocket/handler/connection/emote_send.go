package connection

import (
	"context"
	"encoding/json"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game/shared"
)

// EmoteBroadcaster shows an emote to everyone in a game.
type EmoteBroadcaster interface {
	BroadcastEmote(gameID, playerID, emote string)
}

// EmoteSendHandler relays a player's emote to the table. Emotes are not stored.
type EmoteSendHandler struct {
	broadcaster EmoteBroadcaster
}

// NewEmoteSendHandler creates an emote handler.
func NewEmoteSendHandler(broadcaster EmoteBroadcaster) *EmoteSendHandler {
	return &EmoteSendHandler{broadcaster: broadcaster}
}

// HandleMessage processes an emote-send from a player.
func (h *EmoteSendHandler) HandleMessage(_ context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	if connection.GameID == "" || connection.PlayerID == "" || connection.IsSpectator() {
		return
	}
	raw, err := json.Marshal(message.Payload)
	if err != nil {
		return
	}
	var payload dto.EmoteSendPayload
	if err := json.Unmarshal(raw, &payload); err != nil || !shared.IsEmote(string(payload.Emote)) {
		return
	}
	h.broadcaster.BroadcastEmote(connection.GameID, connection.PlayerID, string(payload.Emote))
}
