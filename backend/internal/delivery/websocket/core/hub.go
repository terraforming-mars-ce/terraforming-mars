package core

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"

	"openmars/internal/delivery/dto"
	"openmars/internal/logger"
)

// MessageHandler handles one client message type.
type MessageHandler interface {
	HandleMessage(ctx context.Context, connection *Connection, message dto.WebSocketMessage)
}

// HubMessage is a client message waiting to be handled.
type HubMessage struct {
	Connection *Connection
	Message    dto.WebSocketMessage
}

// LeaveFunc is called when the last connection of a player or spectator goes away.
type LeaveFunc func(ctx context.Context, gameID, id string)

// hubJob is work from outside the hub, such as a bot, that must run on the hub goroutine.
type hubJob struct {
	ctx  context.Context
	fn   func()
	done chan struct{}
}

// Hub owns every connection and runs all message handling on one goroutine, so actions
// never run concurrently.
type Hub struct {
	register    chan *Connection
	unregisters chan *Connection
	messages    chan HubMessage
	jobs        chan hubJob
	stopped     chan struct{}

	// ctx is the context of the running hub. It is only read on the hub goroutine.
	ctx context.Context

	manager         *Manager
	logger          *slog.Logger
	handlers        map[dto.MessageType]MessageHandler
	onPlayerLeft    LeaveFunc
	onSpectatorLeft LeaveFunc
	guard           func(context.Context, *Connection, dto.WebSocketMessage) error
}

// NewHub creates a hub. Register handlers before calling Run.
func NewHub() *Hub {
	return &Hub{
		register:    make(chan *Connection),
		unregisters: make(chan *Connection),
		messages:    make(chan HubMessage),
		jobs:        make(chan hubJob),
		stopped:     make(chan struct{}),
		manager:     NewManager(),
		logger:      logger.Get(),
		handlers:    make(map[dto.MessageType]MessageHandler),
	}
}

// RegisterHandler routes a client message type to a handler.
func (h *Hub) RegisterHandler(messageType dto.MessageType, handler MessageHandler) {
	h.handlers[messageType] = handler
}

// MessageTypes lists every client message type the hub handles.
func (h *Hub) MessageTypes() []dto.MessageType {
	types := make([]dto.MessageType, 0, len(h.handlers))
	for t := range h.handlers {
		types = append(types, t)
	}
	slices.Sort(types)
	return types
}

// OnLeave sets what happens when the last connection of a player or a spectator goes
// away, whether it closed or moved to another identity.
func (h *Hub) OnLeave(player, spectator LeaveFunc) {
	h.onPlayerLeft = player
	h.onSpectatorLeft = spectator
}

// Manager returns the connection index.
func (h *Hub) Manager() *Manager {
	return h.manager
}

// Run handles connections and messages until ctx ends, then closes every connection.
func (h *Hub) Run(ctx context.Context) {
	h.ctx = ctx
	defer func() {
		close(h.stopped)
		h.manager.closeAll()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case c := <-h.register:
			h.manager.add(c)
		case c := <-h.unregisters:
			if previous, ok := h.manager.remove(c); ok {
				h.released(previous)
			}
		case msg := <-h.messages:
			h.route(msg)
		case job := <-h.jobs:
			h.runJob(job)
		}
	}
}

func (h *Hub) add(c *Connection) bool {
	select {
	case h.register <- c:
		return true
	case <-h.stopped:
		return false
	}
}

func (h *Hub) submit(msg HubMessage) bool {
	select {
	case h.messages <- msg:
		return true
	case <-h.stopped:
		return false
	}
}

func (h *Hub) unregister(c *Connection) {
	select {
	case h.unregisters <- c:
	case <-h.stopped:
	}
}

// bind runs on the hub goroutine, from a handler.
func (h *Hub) bind(c *Connection, id Identity) {
	previous, ok := h.manager.bind(c, id)
	if ok && previous != id {
		h.released(previous)
	}
}

// released tells the game that an identity lost a connection, once no other connection
// still acts as it.
func (h *Hub) released(id Identity) {
	if h.ctx.Err() != nil {
		return
	}
	switch {
	case id.IsPlayer():
		if len(h.manager.PlayerConnections(id.GameID, id.PlayerID)) == 0 && h.onPlayerLeft != nil {
			h.onPlayerLeft(h.ctx, id.GameID, id.PlayerID)
		}
	case id.IsSpectator():
		if h.onSpectatorLeft != nil {
			h.onSpectatorLeft(h.ctx, id.GameID, id.SpectatorID)
		}
	}
}

func (h *Hub) route(msg HubMessage) {
	if h.guard != nil {
		if err := h.guard(h.ctx, msg.Connection, msg.Message); err != nil {
			msg.Connection.SendError(msg.Message.Type, err.Error())
			return
		}
	}
	handler, ok := h.handlers[msg.Message.Type]
	if !ok {
		h.logger.Warn("Unknown message type", slog.String("message_type", string(msg.Message.Type)))
		msg.Connection.SendError(msg.Message.Type, fmt.Sprintf("Unknown message type %q", msg.Message.Type))
		return
	}
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Message handler panicked",
				slog.String("message_type", string(msg.Message.Type)),
				slog.Any("panic", r),
				slog.String("stack", string(debug.Stack())))
			msg.Connection.SendError(msg.Message.Type, "Internal server error")
		}
	}()
	handler.HandleMessage(h.ctx, msg.Connection, msg.Message)
}

// SendToPlayer sends a message to every connection of a player.
func (h *Hub) SendToPlayer(gameID, playerID string, message dto.WebSocketMessage) {
	for _, c := range h.manager.PlayerConnections(gameID, playerID) {
		c.Send(message)
	}
}

// SendToSpectator sends a message to a spectator.
func (h *Hub) SendToSpectator(gameID, spectatorID string, message dto.WebSocketMessage) {
	if c := h.manager.SpectatorConnection(gameID, spectatorID); c != nil {
		c.Send(message)
	}
}

// DisconnectPlayer sends a final notice to every connection of a player and closes them.
// The game is not told the player left: the caller has already removed them.
func (h *Hub) DisconnectPlayer(gameID, playerID string, notice dto.WebSocketMessage) {
	h.disconnect(h.manager.PlayerConnections(gameID, playerID), notice)
}

// DisconnectSpectator sends a final notice to a spectator and closes the connection.
func (h *Hub) DisconnectSpectator(gameID, spectatorID string, notice dto.WebSocketMessage) {
	if c := h.manager.SpectatorConnection(gameID, spectatorID); c != nil {
		h.disconnect([]*Connection{c}, notice)
	}
}

// DisconnectGame sends a final notice to every connection in a game and closes them.
func (h *Hub) DisconnectGame(gameID string, notice dto.WebSocketMessage) {
	h.disconnect(h.manager.GameConnections(gameID), notice)
}

func (h *Hub) disconnect(conns []*Connection, notice dto.WebSocketMessage) {
	for _, c := range conns {
		h.manager.bind(c, Identity{})
		c.Send(notice)
		c.CloseAfterFlush()
	}
}

// Do runs fn on the hub goroutine and waits for it, so callers outside the hub can read
// and change game state without racing player actions. It must not be called from the
// hub goroutine itself. fn is skipped if ctx ends before the hub reaches it.
func (h *Hub) Do(ctx context.Context, fn func()) error {
	job := hubJob{ctx: ctx, fn: fn, done: make(chan struct{})}
	select {
	case h.jobs <- job:
	case <-ctx.Done():
		return ctx.Err()
	case <-h.stopped:
		return context.Canceled
	}
	select {
	case <-job.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) runJob(job hubJob) {
	defer close(job.done)
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Hub job panicked", slog.Any("panic", r), slog.String("stack", string(debug.Stack())))
		}
	}()
	if job.ctx.Err() == nil {
		job.fn()
	}
}

// SetMessageGuard checks session restrictions before routing a client command.
func (h *Hub) SetMessageGuard(guard func(context.Context, *Connection, dto.WebSocketMessage) error) {
	h.guard = guard
}
