package dto

import "time"

// GameHistoryEntryDto represents a single historical game state snapshot.
type GameHistoryEntryDto struct {
	Sequence     int64                           `json:"sequence" ts:"number"`
	Timestamp    time.Time                       `json:"timestamp" ts:"string"`
	Generation   int                             `json:"generation" ts:"number"`
	Phase        GamePhase                       `json:"phase" ts:"GamePhase"`
	Temperature  int                             `json:"temperature" ts:"number"`
	Oxygen       int                             `json:"oxygen" ts:"number"`
	Oceans       int                             `json:"oceans" ts:"number"`
	Venus        int                             `json:"venus" ts:"number"`
	ActionNumber int                             `json:"actionNumber" ts:"number"`
	Board        BoardDto                        `json:"board" ts:"BoardDto"`
	Players      map[string]GameHistoryPlayerDto `json:"players" ts:"Record<string, GameHistoryPlayerDto>"`
	Milestones   []ClaimedMilestoneDto           `json:"milestones" ts:"ClaimedMilestoneDto[]"`
	Awards       []FundedAwardDto                `json:"awards" ts:"FundedAwardDto[]"`
	Settings     GameSettingsDto                 `json:"settings" ts:"GameSettingsDto"`
}

// GameHistoryPlayerDto contains the player data needed for graphs and replay.
type GameHistoryPlayerDto struct {
	ID              string         `json:"id" ts:"string"`
	Name            string         `json:"name" ts:"string"`
	Color           string         `json:"color" ts:"string"`
	TerraformRating int            `json:"terraformRating" ts:"number"`
	Credits         int            `json:"credits" ts:"number"`
	Steel           int            `json:"steel" ts:"number"`
	Titanium        int            `json:"titanium" ts:"number"`
	Plants          int            `json:"plants" ts:"number"`
	Energy          int            `json:"energy" ts:"number"`
	Heat            int            `json:"heat" ts:"number"`
	PlayedCardCount int            `json:"playedCardCount" ts:"number"`
	Production      ProductionDto  `json:"production" ts:"ProductionDto"`
	PlayedCardIDs   []string       `json:"playedCardIds" ts:"string[]"`
	HandCardIDs     []string       `json:"handCardIds" ts:"string[]"`
	CorporationID   string         `json:"corporationId" ts:"string"`
	ResourceStorage map[string]int `json:"resourceStorage" ts:"Record<string, number>"`
}

// ClaimedMilestoneDto represents a claimed milestone in history.
type ClaimedMilestoneDto struct {
	Type     string `json:"type" ts:"string"`
	PlayerID string `json:"playerId" ts:"string"`
}

// FundedAwardDto represents a funded award in history.
type FundedAwardDto struct {
	Type     string `json:"type" ts:"string"`
	PlayerID string `json:"playerId" ts:"string"`
}

// GetGameHistoryResponse is the HTTP response for the game history endpoint.
type GetGameHistoryResponse struct {
	Entries []GameHistoryEntryDto `json:"entries" ts:"GameHistoryEntryDto[]"`
}
