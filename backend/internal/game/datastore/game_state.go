package datastore

import (
	"time"

	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/projectfunding"
	"terraforming-mars-backend/internal/game/shared"
)

// GameState holds all game data.
type GameState struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Status       shared.GameStatus
	Settings     shared.GameSettings
	HostPlayerID string

	CurrentPhase shared.GamePhase
	Generation   int

	Temperature int
	Oxygen      int
	Oceans      int
	MaxOceans   int
	Venus       int

	CurrentTurnPlayerID     string // empty = no current turn
	CurrentTurnActions      int    // -1 = unlimited, 0 = none, >0 = specific count
	CurrentTurnTotalActions int
	GlobalActionCounter     int

	Tiles []board.Tile

	ProjectCards   []string
	Corporations   []string
	DiscardPile    []string
	RemovedCards   []string
	PreludeCards   []string
	DrawnCardCount int
	ShuffleCount   int

	PlayerOrder []string
	TurnOrder   []string
	Players     map[string]*PlayerState

	ClaimedMilestones []shared.ClaimedMilestone
	FundedAwards      []shared.FundedAward

	FinalScores []shared.FinalScore
	WinnerID    string
	IsTie       bool

	Spectators   map[string]*shared.SpectatorState
	ChatMessages []shared.ChatMessage

	PendingTileSelections      map[string]*shared.PendingTileSelection
	PendingTileSelectionQueues map[string]*shared.PendingTileSelectionQueue
	ForcedFirstActions         map[string]*shared.ForcedFirstAction
	ProductionPhases           map[string]*shared.ProductionPhase
	SelectCorporationPhases    map[string]*shared.SelectCorporationPhase
	SelectStartingCardsPhases  map[string]*shared.SelectStartingCardsPhase
	SelectPreludeCardsPhases   map[string]*shared.SelectPreludeCardsPhase
	DeferredStartingChoices    map[string]*shared.DeferredStartingChoices

	InitPhasePlayerIndex       int
	InitPhaseWaitingForConfirm bool
	InitPhaseConfirmVersion    int

	NextGenTurnOrderFrozen bool

	ColonyStates         []*colony.ColonyState
	TradeFleets          map[string]bool
	ProjectFundingStates []*projectfunding.ProjectState

	TriggeredEffects []shared.TriggeredEffect
}

// GameStateHistoryEntry stores a full GameState snapshot at a point in time.
type GameStateHistoryEntry struct {
	GameID    string
	Sequence  int64
	Timestamp time.Time
	State     *GameState
}

// PlayerState holds a single player's data.
type PlayerState struct {
	ID            string
	Name          string
	Connected     bool
	PlayerType    string
	BotStatus     string
	BotDifficulty string
	BotSpeed      string

	CorporationID      string
	Color              string
	HasPassed          bool
	HasExited          bool
	PendingDemoChoices *shared.PendingDemoChoices

	HandCardIDs   []string
	PlayedCardIDs []string

	Resources       shared.Resources
	Production      shared.Production
	TerraformRating int

	ResourceStorage map[string]int

	PaymentSubstitutes        []shared.PaymentSubstitute
	StoragePaymentSubstitutes []shared.StoragePaymentSubstitute
	ValueModifiers            map[shared.ResourceType]int

	SelectCorporationPhase         *shared.SelectCorporationPhase
	SelectStartingCardsPhase       *shared.SelectStartingCardsPhase
	SelectPreludeCardsPhase        *shared.SelectPreludeCardsPhase
	PendingCardSelection           *shared.PendingCardSelection
	PendingCardDrawSelection       *shared.PendingCardDrawSelection
	PendingCardDiscardSelection    *shared.PendingCardDiscardSelection
	PendingBehaviorChoiceSelection *shared.PendingBehaviorChoiceSelection
	PendingStealTargetSelection    *shared.PendingStealTargetSelection
	PendingColonyResourceSelection *shared.PendingColonyResourceSelection
	PendingColonyResourceQueue     []shared.PendingColonyResourceSelection
	PendingAwardFundSelection      *shared.PendingAwardFundSelection
	PendingColonySelection         *shared.PendingColonySelection
	PendingFreeTradeSelection      *shared.PendingFreeTradeSelection

	Actions []shared.CardAction

	Effects []shared.CardEffect

	BonusTags map[shared.CardTag]int

	GenerationalEvents map[shared.GenerationalEvent]int

	VPGranters []shared.VPGranter
}
