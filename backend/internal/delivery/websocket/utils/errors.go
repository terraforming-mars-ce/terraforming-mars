package utils

import (
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
)

// ErrorHandler provides standardized error handling for WebSocket messages
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// SendError sends an error message to a connection
func (eh *ErrorHandler) SendError(connection *core.Connection, message string) {
	_, gameID := connection.GetPlayer()

	errorMessage := dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: dto.ErrorPayload{
			Message: message,
		},
		GameID: gameID,
	}

	connection.SendMessage(errorMessage)
}

// Standard error messages
const (
	ErrInvalidPayload     = "Invalid message payload"
	ErrUnknownMessageType = "Unknown message type"
	ErrMustConnectFirst   = "You must connect to a game first"
	ErrInvalidActionType  = "Invalid action type"
	ErrGameNotFound       = "Game not found"
	ErrPlayerNotFound     = "Player not found"
	ErrActionFailed       = "Action failed"
	ErrConnectionFailed   = "Connection failed"
)
