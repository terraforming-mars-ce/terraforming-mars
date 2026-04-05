package dto

import "encoding/json"

// CreateGameRequest represents the request body for creating a game
type CreateGameRequest struct {
	MaxPlayers         int      `json:"maxPlayers" binding:"required,min=1,max=10" ts:"number"`
	VenusNextEnabled   bool     `json:"venusNextEnabled" ts:"boolean"`
	DevelopmentMode    bool     `json:"developmentMode" ts:"boolean"`
	CardPacks          []string `json:"cardPacks,omitempty" ts:"string[] | undefined"`
	ClaudeAPIKey       string   `json:"claudeApiKey,omitempty" ts:"string | undefined"`
	SelectedMilestones []string `json:"selectedMilestones,omitempty" ts:"string[] | undefined"`
	SelectedAwards     []string `json:"selectedAwards,omitempty" ts:"string[] | undefined"`
}

// CreateGameResponse represents the response for creating a game
type CreateGameResponse struct {
	Game GameDto `json:"game" ts:"GameDto"`
}

// JoinGameRequest represents the request body for joining a game
type JoinGameRequest struct {
	PlayerName string `json:"playerName" binding:"required,min=1,max=50"`
}

// JoinGameResponse represents the response for joining a game
type JoinGameResponse struct {
	Game     GameDto `json:"game" ts:"GameDto"`
	PlayerID string  `json:"playerId" ts:"string"`
}

// GetGameResponse represents the response for getting a game
type GetGameResponse struct {
	Game GameDto `json:"game" ts:"GameDto"`
}

// ListGamesResponse represents the response for listing games
type ListGamesResponse struct {
	Games []GameDto `json:"games" ts:"GameDto[]"`
}

// GetPlayerResponse represents the response for getting a player
type GetPlayerResponse struct {
	Player PlayerDto `json:"player" ts:"PlayerDto"`
}

// ListCardsResponse represents the response for listing cards with pagination
type ListCardsResponse struct {
	Cards      []CardDto `json:"cards" ts:"CardDto[]"`
	TotalCount int       `json:"totalCount" ts:"number"`
	Offset     int       `json:"offset" ts:"number"`
	Limit      int       `json:"limit" ts:"number"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" ts:"string"`
	Code    string `json:"code,omitempty" ts:"string"`
	Details string `json:"details,omitempty" ts:"string"`
}

// CreateDemoLobbyRequest represents the request body for creating a demo lobby
type CreateDemoLobbyRequest struct {
	PlayerCount int      `json:"playerCount" ts:"number"`
	CardPacks   []string `json:"cardPacks,omitempty" ts:"string[] | undefined"`
	PlayerName  string   `json:"playerName,omitempty" ts:"string | undefined"`
}

// CreateDemoLobbyResponse represents the response for creating a demo lobby
type CreateDemoLobbyResponse struct {
	Game     GameDto `json:"game" ts:"GameDto"`
	PlayerID string  `json:"playerId" ts:"string"`
}

// MilestoneAwardItemDto represents a single milestone or award for selection
type MilestoneAwardItemDto struct {
	ID          string `json:"id" ts:"string"`
	Name        string `json:"name" ts:"string"`
	Description string `json:"description" ts:"string"`
}

// ListMilestonesAwardsResponse represents the response for listing milestones and awards
type ListMilestonesAwardsResponse struct {
	Milestones []MilestoneAwardItemDto `json:"milestones" ts:"MilestoneAwardItemDto[]"`
	Awards     []MilestoneAwardItemDto `json:"awards" ts:"MilestoneAwardItemDto[]"`
}

// FeedbackRequest represents the request body for submitting feedback
type FeedbackRequest struct {
	Title       string          `json:"title" ts:"string"`
	Description string          `json:"description" ts:"string"`
	Tags        []string        `json:"tags" ts:"string[]"`
	Author      string          `json:"author,omitempty" ts:"string | undefined"`
	GameState   json.RawMessage `json:"gameState,omitempty" ts:"GameDto | undefined"`
}

// FeedbackDto represents a feedback submission's current state
type FeedbackDto struct {
	ID            string `json:"id" ts:"string"`
	Status        string `json:"status" ts:"string"`
	StatusMessage string `json:"statusMessage" ts:"string"`
	IssueURL      string `json:"issueUrl,omitempty" ts:"string | undefined"`
}

// FeedbackResponse represents the response for feedback operations
type FeedbackResponse struct {
	Report FeedbackDto `json:"report" ts:"FeedbackDto"`
}

// FeedbackStatusResponse represents the response for feedback service availability
type FeedbackStatusResponse struct {
	Available bool   `json:"available" ts:"boolean"`
	Reason    string `json:"reason,omitempty" ts:"string | undefined"`
}
