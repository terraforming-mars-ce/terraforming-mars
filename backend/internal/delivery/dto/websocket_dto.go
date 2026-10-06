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

// PlayerConnectedPayload contains data about a newly connected player
type PlayerConnectedPayload struct {
	PlayerID   string  `json:"playerId"`
	PlayerName string  `json:"playerName"`
	Game       GameDto `json:"game"`
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// PlayCardErrorPayload identifies the rejected card play for client presentation recovery.
type PlayCardErrorPayload struct {
	Action string `json:"action" tstype:"'play-card'"`
	CardID string `json:"cardId"`
	Error  string `json:"error"`
}

// FullStatePayload contains the complete game state
type FullStatePayload struct {
	Game     GameDto `json:"game"`
	PlayerID string  `json:"playerId"`
}

// PlayerReconnectedPayload contains data about a reconnected player
type PlayerReconnectedPayload struct {
	PlayerID   string  `json:"playerId"`
	PlayerName string  `json:"playerName"`
	Game       GameDto `json:"game"`
}

// PlayerDisconnectedPayload contains data about a disconnected player (for internal handler use)
type PlayerDisconnectedPayload struct {
	PlayerID string `json:"playerId"`
	GameID   string `json:"gameId"`
}

// PlayerProductionData contains production data for a single player
type PlayerProductionData struct {
	PlayerID   string        `json:"playerId"`
	PlayerName string        `json:"playerName"`
	Production ProductionDto `json:"production"`
}

// ProductionPhaseStartedPayload contains data when production phase begins
type ProductionPhaseStartedPayload struct {
	Generation  int                    `json:"generation"`
	PlayersData []PlayerProductionData `json:"playersData"`
	Game        GameDto                `json:"game"`
}

// LogUpdatePayload contains game log entries sent via WebSocket
type LogUpdatePayload struct {
	Logs      []StateDiffDto `json:"logs"`
	IsHistory bool           `json:"isHistory"`
}

// ConfirmStartingCardSelectionMessage represents confirm starting card selection message
type ConfirmStartingCardSelectionMessage struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
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

// SpectatorConnectedPayload contains data about a newly connected spectator.
type SpectatorConnectedPayload struct {
	SpectatorID string  `json:"spectatorId"`
	Game        GameDto `json:"game"`
}

// SpectatorDisconnectedPayload contains data about a disconnected spectator.
type SpectatorDisconnectedPayload struct {
	SpectatorID string `json:"spectatorId"`
	GameID      string `json:"gameId"`
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
