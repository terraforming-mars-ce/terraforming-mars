package dto

import "encoding/json"

// CreateGameRequest represents the request body for creating a game
type CreateGameRequest struct {
	MaxPlayers         int      `json:"maxPlayers" binding:"required,min=1,max=10"`
	VenusNextEnabled   bool     `json:"venusNextEnabled"`
	DevelopmentMode    bool     `json:"developmentMode"`
	DemoGame           bool     `json:"demoGame"`
	CardPacks          []string `json:"cardPacks,omitempty"`
	ClaudeAPIKey       string   `json:"claudeApiKey,omitempty"`
	SelectedMilestones []string `json:"selectedMilestones,omitempty"`
	SelectedAwards     []string `json:"selectedAwards,omitempty"`
}

// CreateGameResponse represents the response for creating a game
type CreateGameResponse struct {
	Game GameDto `json:"game"`
}

// JoinGameRequest represents the request body for joining a game
type JoinGameRequest struct {
	PlayerName string `json:"playerName" binding:"required,min=1,max=50"`
}

// JoinGameResponse represents the response for joining a game
type JoinGameResponse struct {
	Game     GameDto `json:"game"`
	PlayerID string  `json:"playerId"`
}

// GetGameResponse represents the response for getting a game
type GetGameResponse struct {
	Game GameDto `json:"game"`
}

// ListGamesResponse represents the response for listing games
type ListGamesResponse struct {
	Games []GameDto `json:"games"`
}

// GetPlayerResponse represents the response for getting a player
type GetPlayerResponse struct {
	Player PlayerDto `json:"player"`
}

// ListCardsResponse represents the response for listing cards with pagination
type ListCardsResponse struct {
	Cards      []CardDto `json:"cards"`
	TotalCount int       `json:"totalCount"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
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
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags"`
	Author      string          `json:"author,omitempty"`
	GameState   json.RawMessage `json:"gameState,omitempty"`
}

// FeedbackDto represents a feedback submission's current state
type FeedbackDto struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	StatusMessage string `json:"statusMessage"`
	IssueURL      string `json:"issueUrl,omitempty"`
}

// FeedbackResponse represents the response for feedback operations
type FeedbackResponse struct {
	Report FeedbackDto `json:"report"`
}

// FeedbackStatusResponse represents the response for feedback service availability
type FeedbackStatusResponse struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}
