package core

import (
	"net/http"
	"time"

	"terraforming-mars-backend/internal/logger"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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
	logger *zap.Logger
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
	h.logger.Debug("WebSocket connection request received", zap.String("remote_addr", r.RemoteAddr))

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	// Create connection ID and connection object
	connectionID := uuid.New().String()
	connection := NewConnection(connectionID, conn,
		h.hub.GetManager(), // Direct manager reference
		func(msg HubMessage) { h.hub.Messages <- msg },      // onMessage callback
		func(conn *Connection) { h.hub.Unregister <- conn }) // onDisconnect callback

	h.logger.Debug("New WebSocket connection established",
		zap.String("connection_id", connectionID),
		zap.String("remote_addr", r.RemoteAddr))

	h.hub.Register <- connection

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go connection.WritePump()
	go connection.ReadPump()

	h.logger.Debug("WebSocket connection fully initialized", zap.String("connection_id", connectionID))
}
