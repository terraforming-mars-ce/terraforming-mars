package core

import (
	"log/slog"
	"net/http"
	"time"

	"openmars/internal/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development - should be restricted in production
		return true
	},
}

// Handler handles WebSocket HTTP upgrade requests
type Handler struct {
	hub    *Hub
	logger *slog.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub:    hub,
		logger: logger.Get(),
	}
}

// ServeWS handles WebSocket upgrade requests from clients
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("WebSocket connection request received", slog.String("remote_addr", r.RemoteAddr))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection to WebSocket", slog.Any("error", err))
		return
	}

	// Create connection ID and connection object
	connectionID := uuid.New().String()
	connection := NewConnection(connectionID, conn,
		h.hub.GetManager(), // Direct manager reference
		func(msg HubMessage) { h.hub.Messages <- msg },      // onMessage callback
		func(conn *Connection) { h.hub.Unregister <- conn }) // onDisconnect callback

	h.logger.Debug("New WebSocket connection established",
		slog.String("connection_id", connectionID),
		slog.String("remote_addr", r.RemoteAddr))

	h.hub.Register <- connection

	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		h.logger.Warn("Failed to set initial read deadline", slog.Any("error", err), slog.String("connection_id", connectionID))
	}
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		h.logger.Warn("Failed to set initial write deadline", slog.Any("error", err), slog.String("connection_id", connectionID))
	}

	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			h.logger.Warn("Failed to set read deadline in pong handler", slog.Any("error", err), slog.String("connection_id", connectionID))
		}
		return nil
	})

	go connection.WritePump()
	go connection.ReadPump()

	h.logger.Debug("WebSocket connection fully initialized", slog.String("connection_id", connectionID))
}
