package game

import "time"

// MaxChatMessages is the maximum number of chat messages stored per game.
const MaxChatMessages = 200

// MaxChatMessageLength is the maximum allowed length of a single chat message.
const MaxChatMessageLength = 500

// ChatMessage represents a message sent by a player or spectator.
type ChatMessage struct {
	SenderID    string
	SenderName  string
	SenderColor string
	Message     string
	Timestamp   time.Time
	IsSpectator bool
}
