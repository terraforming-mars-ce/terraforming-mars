package shared

import (
	"slices"
	"time"
)

// GameStatus represents the current status of a game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingSelection     GamePhase = "starting_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseInitApplyCorp         GamePhase = "init_apply_corp"
	GamePhaseInitApplyPrelude      GamePhase = "init_apply_prelude"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseFinalPhase            GamePhase = "final_phase"
	GamePhaseComplete              GamePhase = "complete"
)

// GameSettings contains configurable game parameters
type GameSettings struct {
	MaxPlayers         int
	Temperature        *int
	Oxygen             *int
	Oceans             *int
	Venus              *int
	VenusNextEnabled   bool
	DevelopmentMode    bool
	DemoGame           bool
	CardPacks          []string
	Generation         *int
	ClaudeAPIKey       string
	ClaudeModel        string
	SelectedMilestones []string
	SelectedAwards     []string
}

// Card pack constants
const (
	PackBaseGame       = "base-game"
	PackFuture         = "future"
	PackPrelude        = "prelude"
	PackVenus          = "venus-next"
	PackExperimental   = "experimental"
	PackColonies       = "colonies"
	PackProjectFunding = "project-funding"
)

// EnabledPacks returns a set of all enabled pack names for filtering.
func (s GameSettings) EnabledPacks() map[string]bool {
	packs := make(map[string]bool, len(s.CardPacks)+1)
	for _, pack := range s.CardPacks {
		packs[pack] = true
	}
	if s.VenusNextEnabled {
		packs[PackVenus] = true
	}
	return packs
}

// HasPrelude returns true if the prelude card pack is enabled
func (s GameSettings) HasPrelude() bool {
	return slices.Contains(s.CardPacks, PackPrelude)
}

// HasColonies returns true if the Colonies expansion is enabled
func (s GameSettings) HasColonies() bool {
	return slices.Contains(s.CardPacks, PackColonies)
}

// HasProjectFunding returns true if the Project Funding expansion is enabled
func (s GameSettings) HasProjectFunding() bool {
	return slices.Contains(s.CardPacks, PackProjectFunding)
}

// HasVenus returns true if the Venus expansion is enabled
func (s GameSettings) HasVenus() bool {
	return s.VenusNextEnabled
}

// DefaultCardPacks returns the default card packs
func DefaultCardPacks() []string {
	return []string{PackBaseGame, PackPrelude}
}

// SourceType identifies the type of action that caused a state change
type SourceType string

const (
	SourceTypeCardPlay                 SourceType = "card_play"
	SourceTypeCardAction               SourceType = "card_action"
	SourceTypeStandardProject          SourceType = "standard_project"
	SourceTypePassiveEffect            SourceType = "passive_effect"
	SourceTypeResourceConvert          SourceType = "resource_convert"
	SourceTypeGameEvent                SourceType = "game_event"
	SourceTypeInitial                  SourceType = "initial"
	SourceTypeAward                    SourceType = "award"
	SourceTypeMilestone                SourceType = "milestone"
	SourceTypeActionAdded              SourceType = "action_added"
	SourceTypeEffectAdded              SourceType = "effect_added"
	SourceTypeGlobalBonus              SourceType = "global_bonus"
	SourceTypeColonyTrade              SourceType = "colony_trade"
	SourceTypeColonyBuild              SourceType = "colony_build"
	SourceTypeProjectFundingSeat       SourceType = "project_funding_seat"
	SourceTypeProjectFundingCompletion SourceType = "project_funding_completion"
)

// Chat constants
const (
	MaxChatMessages      = 200
	MaxChatMessageLength = 500
)

// Spectator constants and colors
const MaxSpectators = 4

// PlayerColors is the palette of 10 visually distinct colors available to players.
var PlayerColors = []string{
	"#e53935", "#1e88e5", "#43a047", "#ffb300", "#8e24aa",
	"#00acc1", "#f4511e", "#3949ab", "#c0ca33", "#d81b60",
}

// SpectatorColors is the palette of colors assigned to spectators.
var SpectatorColors = []string{"#9b9b9b", "#7eb8da", "#c4a6e8", "#e8c4a6"}

// SpectatorState represents a spectator's data
type SpectatorState struct {
	ID    string
	Name  string
	Color string
}

// ChatMessage represents a single chat message
type ChatMessage struct {
	SenderID    string
	SenderName  string
	SenderColor string
	Message     string
	Timestamp   time.Time
	IsSpectator bool
}

// ClaimedMilestone represents a milestone that has been claimed
type ClaimedMilestone struct {
	Type       MilestoneType
	PlayerID   string
	Generation int
	ClaimedAt  time.Time
}

// FundedAward represents an award that has been funded
type FundedAward struct {
	Type           AwardType
	FundedByPlayer string
	FundingOrder   int
	FundingCost    int
	FundedAt       time.Time
}

// FinalScore holds the final scoring breakdown for a player
type FinalScore struct {
	PlayerID   string
	PlayerName string
	Breakdown  VPBreakdown
	Credits    int
	Placement  int
	IsWinner   bool
}

// VPBreakdown contains the victory point breakdown
type VPBreakdown struct {
	TerraformRating   int
	CardVP            int
	CardVPDetails     []CardVPDetail
	MilestoneVP       int
	AwardVP           int
	GreeneryVP        int
	GreeneryVPDetails []GreeneryVPDetail
	CityVP            int
	CityVPDetails     []CityVPDetail
	TotalVP           int
}

// CardVPDetail contains VP details for a single card
type CardVPDetail struct {
	CardID     string
	CardName   string
	Conditions []CardVPConditionDetail
	TotalVP    int
}

// CardVPConditionDetail contains details of a single VP condition evaluation
type CardVPConditionDetail struct {
	ConditionType  string
	Amount         int
	Count          int
	MaxTrigger     *int
	ActualTriggers int
	TotalVP        int
	Explanation    string
}

// GreeneryVPDetail contains VP details for a greenery tile
type GreeneryVPDetail struct {
	Coordinate string
	VP         int
}

// CityVPDetail contains VP details for a city tile
type CityVPDetail struct {
	CityCoordinate     string
	AdjacentGreeneries []string
	VP                 int
}

// TriggeredEffect records a triggered passive effect for notification
type TriggeredEffect struct {
	CardName          string
	PlayerID          string
	SourceType        SourceType
	Outputs           []BehaviorCondition
	CalculatedOutputs []CalculatedOutput
	Behaviors         []CardBehavior
	VPConditions      []VPConditionForLog
}

// CalculatedOutput represents an actual output value that was applied
type CalculatedOutput struct {
	ResourceType string
	Amount       int
	IsScaled     bool
}

// VPConditionForLog is a simplified VP condition for log display
type VPConditionForLog struct {
	Amount     int
	Condition  string
	MaxTrigger *int
	Per        *PerCondition
}

// PendingDemoChoices holds a player's card selections made during the demo lobby phase
type PendingDemoChoices struct {
	CorporationID   string
	PreludeIDs      []string
	CardIDs         []string
	Resources       Resources
	Production      Production
	TerraformRating int
}

// DeferredStartingChoices holds choices that are applied after init
type DeferredStartingChoices struct {
	CorporationID   string
	PreludeIDs      []string
	CardIDs         []string
	CorpApplied     bool
	PreludesApplied bool
}
