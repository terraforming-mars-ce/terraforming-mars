package dto

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type    MessageType `json:"type" ts:"MessageType"`
	Payload interface{} `json:"payload" ts:"any"`
	GameID  string      `json:"gameId,omitempty" ts:"string"`
}

// PlayerConnectPayload contains player connection data
type PlayerConnectPayload struct {
	PlayerName string `json:"playerName" ts:"string"`
	GameID     string `json:"gameId" ts:"string"`
	PlayerID   string `json:"playerId,omitempty" ts:"string | undefined"` // Optional: used for reconnection
}

// GameUpdatedPayload contains updated game state
type GameUpdatedPayload struct {
	Game GameDto `json:"game" ts:"GameDto"`
}

// PlayerConnectedPayload contains data about a newly connected player
type PlayerConnectedPayload struct {
	PlayerID   string  `json:"playerId" ts:"string"`
	PlayerName string  `json:"playerName" ts:"string"`
	Game       GameDto `json:"game" ts:"GameDto"`
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Message string `json:"message" ts:"string"`
	Code    string `json:"code,omitempty" ts:"string"`
}

// FullStatePayload contains the complete game state
type FullStatePayload struct {
	Game     GameDto `json:"game" ts:"GameDto"`
	PlayerID string  `json:"playerId" ts:"string"`
}

// PlayerReconnectedPayload contains data about a reconnected player
type PlayerReconnectedPayload struct {
	PlayerID   string  `json:"playerId" ts:"string"`
	PlayerName string  `json:"playerName" ts:"string"`
	Game       GameDto `json:"game" ts:"GameDto"`
}

// PlayerDisconnectedPayload contains data about a disconnected player (for internal handler use)
type PlayerDisconnectedPayload struct {
	PlayerID string `json:"playerId" ts:"string"`
	GameID   string `json:"gameId" ts:"string"`
}

// PlayerProductionData contains production data for a single player
type PlayerProductionData struct {
	PlayerID   string        `json:"playerId" ts:"string"`
	PlayerName string        `json:"playerName" ts:"string"`
	Production ProductionDto `json:"production" ts:"ProductionDto"`
}

// ProductionPhaseStartedPayload contains data when production phase begins
type ProductionPhaseStartedPayload struct {
	Generation  int                    `json:"generation" ts:"number"`
	PlayersData []PlayerProductionData `json:"playersData" ts:"PlayerProductionData[]"`
	Game        GameDto                `json:"game" ts:"GameDto"`
}

// LogUpdatePayload contains game log entries sent via WebSocket
type LogUpdatePayload struct {
	Logs []StateDiffDto `json:"logs" ts:"StateDiffDto[]"`
}

// ConfirmStartingCardSelectionMessage represents confirm starting card selection message
type ConfirmStartingCardSelectionMessage struct {
	GameID   string `json:"gameId" ts:"string"`
	PlayerID string `json:"playerId" ts:"string"`
}

// PlayerTakeoverPayload contains data for player takeover requests
type PlayerTakeoverPayload struct {
	GameID         string `json:"gameId" ts:"string"`
	TargetPlayerID string `json:"targetPlayerId" ts:"string"`
}

// SpectatorConnectPayload contains spectator connection data.
type SpectatorConnectPayload struct {
	SpectatorName string `json:"spectatorName" ts:"string"`
	GameID        string `json:"gameId" ts:"string"`
}

// SpectatorConnectedPayload contains data about a newly connected spectator.
type SpectatorConnectedPayload struct {
	SpectatorID string  `json:"spectatorId" ts:"string"`
	Game        GameDto `json:"game" ts:"GameDto"`
}

// SpectatorDisconnectedPayload contains data about a disconnected spectator.
type SpectatorDisconnectedPayload struct {
	SpectatorID string `json:"spectatorId" ts:"string"`
	GameID      string `json:"gameId" ts:"string"`
}

// ChatMessagePayload contains a chat message from a client.
type ChatMessagePayload struct {
	Message string `json:"message" ts:"string"`
}

// ChatUpdatePayload contains a new chat message broadcast to all clients.
type ChatUpdatePayload struct {
	ChatMessage ChatMessageDto `json:"chatMessage" ts:"ChatMessageDto"`
}
