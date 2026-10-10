package dto

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
	GameID  string      `json:"gameId,omitempty"`
}

// PlayerConnectPayload contains player connection data
type PlayerConnectPayload struct {
	PlayerName string `json:"playerName"`
	GameID     string `json:"gameId"`
	PlayerID   string `json:"playerId,omitempty"` // Optional: used for reconnection
}

// GameUpdatedPayload contains updated game state
type GameUpdatedPayload struct {
	Game GameDto `json:"game"`
}

// PlayerConnectedPayload confirms which player a connection now acts as.
type PlayerConnectedPayload struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
}

// ErrorPayload tells a client that one of its requests failed. It is only ever sent to
// the client that made the request.
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	// RequestType is the client message type that failed.
	RequestType MessageType `json:"requestType,omitempty"`
	// CardID, ResolutionID and SelectionID identify what the failed request was about,
	// so the client can restore that part of its UI.
	CardID       string `json:"cardId,omitempty"`
	ResolutionID string `json:"resolutionId,omitempty"`
	SelectionID  string `json:"selectionId,omitempty"`
}

// LogUpdatePayload contains game log entries sent via WebSocket
type LogUpdatePayload struct {
	Logs      []StateDiffDto `json:"logs"`
	IsHistory bool           `json:"isHistory"`
}

// PlayerTakeoverPayload contains data for player takeover requests
type PlayerTakeoverPayload struct {
	GameID         string `json:"gameId"`
	TargetPlayerID string `json:"targetPlayerId"`
}

// SpectatorConnectPayload contains spectator connection data.
type SpectatorConnectPayload struct {
	SpectatorName string `json:"spectatorName"`
	GameID        string `json:"gameId"`
}

// SpectatorConnectedPayload confirms which spectator a connection now acts as.
type SpectatorConnectedPayload struct {
	SpectatorID string `json:"spectatorId"`
}

// ChatMessagePayload contains a chat message from a client.
type ChatMessagePayload struct {
	Message string `json:"message"`
}

// ChatUpdatePayload contains a new chat message broadcast to all clients.
type ChatUpdatePayload struct {
	ChatMessage ChatMessageDto `json:"chatMessage"`
}

// EmoteName is one of the emotes a player or bot can show over their player card.
type EmoteName string

const (
	EmoteAngry     EmoteName = "angry"
	EmoteCelebrate EmoteName = "celebrate"
	EmoteThinking  EmoteName = "thinking"
	EmoteApplause  EmoteName = "applause"
	EmoteShock     EmoteName = "shock"
	EmoteLaugh     EmoteName = "laugh"
	EmoteSad       EmoteName = "sad"
	EmoteCool      EmoteName = "cool"
)

// EmoteSendPayload is sent by a client to show an emote.
type EmoteSendPayload struct {
	Emote EmoteName `json:"emote" tstype:"'angry' | 'celebrate' | 'thinking' | 'applause' | 'shock' | 'laugh' | 'sad' | 'cool'"`
}

// EmotePayload is broadcast when a player or bot shows an emote.
type EmotePayload struct {
	PlayerID string    `json:"playerId"`
	Emote    EmoteName `json:"emote" tstype:"'angry' | 'celebrate' | 'thinking' | 'applause' | 'shock' | 'laugh' | 'sad' | 'cool'"`
}

// BotThoughtPayload is broadcast when a bot shows a thought bubble or starts or stops typing.
type BotThoughtPayload struct {
	PlayerID string `json:"playerId"`
	Text     string `json:"text"`
	Typing   bool   `json:"typing"`
}

// BotRetryPayload is sent by the host to retry a failed bot.
type BotRetryPayload struct {
	PlayerID string `json:"playerId"`
}
