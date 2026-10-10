package core

import (
	"log/slog"
	"net/http"

	"openmars/internal/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// The gateway in front of the server decides which origins may connect.
		return true
	},
}

// Handler upgrades HTTP requests to WebSocket connections owned by a hub.
type Handler struct {
	hub    *Hub
	logger *slog.Logger
}

// NewHandler creates a WebSocket endpoint for the hub.
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub, logger: logger.Get()}
}

// ServeWS upgrades the request and starts the connection's read and write loops.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("Failed to upgrade connection to WebSocket", slog.Any("error", err))
		return
	}

	c := newConnection(uuid.New().String(), conn, h.hub)
	if !h.hub.add(c) {
		c.Close()
		return
	}
	h.logger.Debug("WebSocket connection opened",
		slog.String("connection_id", c.ID),
		slog.String("remote_addr", r.RemoteAddr))

	go c.writePump()
	go c.readPump()
}
