package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"openmars/internal/game/save"

	gameaction "openmars/internal/action/game"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game"
)

type saveHandler struct {
	action      *gameaction.SaveGameAction
	repo        game.GameRepository
	hub         *core.Hub
	broadcaster *Broadcaster
}

// RegisterSaveHandlers installs resume operations and a server-side gameplay gate.
func RegisterSaveHandlers(hub *core.Hub, broadcaster *Broadcaster, repo game.GameRepository, action *gameaction.SaveGameAction) {
	h := &saveHandler{action: action, repo: repo, hub: hub, broadcaster: broadcaster}
	for _, kind := range []dto.MessageType{dto.MessageTypeWatchResumeGame, dto.MessageTypeClaimResumeSeat, dto.MessageTypeReleaseResumeSeat, dto.MessageTypeResumeGame, dto.MessageTypeResumeBotToken} {
		hub.RegisterHandler(kind, h)
	}
	hub.SetMessageGuard(h.guard)
}

func (h *saveHandler) guard(ctx context.Context, c *core.Connection, m dto.WebSocketMessage) error {
	currentID := c.GameID()
	ids := []string{currentID, m.GameID}
	if payload, ok := m.Payload.(map[string]any); ok {
		if id, ok := payload["gameId"].(string); ok {
			ids = append(ids, id)
		}
	}
	paused := false
	for _, id := range ids {
		if id != "" {
			g, err := h.repo.Get(ctx, id)
			if err == nil && g.ResumeLobby() != nil {
				paused = true
			}
		}
	}
	if !paused {
		return nil
	}
	switch m.Type {
	case dto.MessageTypeWatchResumeGame, dto.MessageTypeClaimResumeSeat, dto.MessageTypeReleaseResumeSeat, dto.MessageTypeResumeGame, dto.MessageTypeResumeBotToken, dto.MessageTypePlayerConnect, dto.MessageTypeChatMessage:
		return nil
	default:
		return fmt.Errorf("the game is paused while players join the resume lobby")
	}
}

func (h *saveHandler) HandleMessage(ctx context.Context, c *core.Connection, m dto.WebSocketMessage) {
	var request dto.ResumeGameRequest
	data, err := json.Marshal(m.Payload)
	if err == nil {
		err = json.Unmarshal(data, &request)
	}
	if err != nil {
		h.fail(c, m.Type, save.Failure("invalid_request", "This request couldn’t be read. Try again.", err))
		return
	}
	identity := c.Identity()
	currentID, currentGame := identity.PlayerID, identity.GameID
	if request.GameID == "" {
		request.GameID = currentGame
	}
	if currentGame != request.GameID {
		currentID = ""
	}
	g, err := h.repo.Get(ctx, request.GameID)
	if err != nil {
		h.fail(c, m.Type, err)
		return
	}
	switch m.Type {
	case dto.MessageTypeWatchResumeGame:
		if g.ResumeLobby() == nil {
			h.fail(c, m.Type, save.Failure("not_ready", "This game has already resumed.", nil))
			return
		}
		if currentGame != request.GameID {
			c.BindPlayer(request.GameID, "")
		}
	case dto.MessageTypeClaimResumeSeat:
		err = h.action.ClaimSeat(ctx, request.GameID, currentID, request.SeatID, request.PlayerName)
		if err == nil {
			c.BindPlayer(request.GameID, request.SeatID)
		}
	case dto.MessageTypeReleaseResumeSeat:
		err = h.action.ReleaseSeat(ctx, request.GameID, currentID, request.SeatID)
		if err == nil {
			for _, connection := range h.hub.Manager().GameConnections(request.GameID) {
				id := connection.PlayerID()
				if id == request.SeatID {
					connection.BindPlayer(request.GameID, "")
				}
			}
		}
	case dto.MessageTypeResumeGame:
		err = h.action.Resume(ctx, request.GameID, currentID)
	case dto.MessageTypeResumeBotToken:
		err = h.action.SetBotToken(ctx, request.GameID, currentID, request.BotToken)
	}
	if err != nil {
		h.fail(c, m.Type, err)
		return
	}
	h.broadcaster.BroadcastGameState(request.GameID, nil)
}

func (h *saveHandler) fail(c *core.Connection, requestType dto.MessageType, err error) {
	public := save.PublicError(err, "The resume request couldn’t be completed. Try again.")
	if public.Code == "internal_error" {
		slog.Error("Resume request failed", "error", err)
	}
	c.SendErrorPayload(dto.ErrorPayload{Message: public.Message, Code: public.Code, RequestType: requestType})
}
