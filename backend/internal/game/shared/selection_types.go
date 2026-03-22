package shared

// SelectCorporationPhase represents the corporation selection phase state
type SelectCorporationPhase struct {
	AvailableCorporations []string
}

// SelectStartingCardsPhase represents the starting cards selection phase state
type SelectStartingCardsPhase struct {
	AvailableCards    []string
	SelectionComplete bool
}

// SelectPreludeCardsPhase represents the prelude card selection phase state
type SelectPreludeCardsPhase struct {
	AvailablePreludes []string
	MaxSelectable     int
}

// ProductionPhase represents the production phase state for a player
type ProductionPhase struct {
	AvailableCards    []string
	SelectionComplete bool
	BeforeResources   Resources
	AfterResources    Resources
	EnergyConverted   int
	CreditsIncome     int
}

// PendingTileSelection represents a pending tile placement
type PendingTileSelection struct {
	TileType       string
	AvailableHexes []string
	Source         string
	SourceCardID   string
	OnComplete     *TileCompletionCallback
}

// TileCompletionCallback stores info about what to call when tile placement completes
type TileCompletionCallback struct {
	Type string
	Data map[string]any
}

// PendingTileSelectionQueue represents a queue of tile placements
type PendingTileSelectionQueue struct {
	Items            []string
	Source           string
	SourceCardID     string
	OnComplete       *TileCompletionCallback
	TileRestrictions *TileRestrictions
}

// PendingCardSelection represents a pending card selection
type PendingCardSelection struct {
	AvailableCards []string
	MinCards       int
	MaxCards       int
	CardCosts      map[string]int
	CardRewards    map[string]int
	Source         string
}

// PendingCardDrawSelection represents a pending card draw/peek/take/buy action
type PendingCardDrawSelection struct {
	AvailableCards      []string
	FreeTakeCount       int
	MaxBuyCount         int
	CardBuyCost         int
	Source              string
	SourceCardID        string
	SourceBehaviorIndex int
	PlayAsPrelude       bool
}

// PendingCardDiscardSelection represents a pending card discard action
type PendingCardDiscardSelection struct {
	MinCards       int
	MaxCards       int
	Source         string
	SourceCardID   string
	PendingOutputs []ResourceCondition
}

// PendingBehaviorChoiceSelection represents a pending behavior choice
type PendingBehaviorChoiceSelection struct {
	Choices      []Choice
	Source       string
	SourceCardID string
}

// PendingStealTargetSelection represents a pending steal target selection
type PendingStealTargetSelection struct {
	EligiblePlayerIDs []string
	ResourceType      ResourceType
	Amount            int
	Source            string
	SourceCardID      string
}

// PendingColonyResourceSelection represents a pending colony resource selection
type PendingColonyResourceSelection struct {
	ResourceType string
	Amount       int
	Source       string
	ColonyID     string
	Reason       string
}

// PendingAwardFundSelection represents a pending award fund selection (e.g., Vitor forced first action)
type PendingAwardFundSelection struct {
	AvailableAwards []string
	Source          string
}

// PendingColonySelection represents a pending colony tile selection from a card effect
type PendingColonySelection struct {
	AvailableColonyIDs         []string
	AllowDuplicatePlayerColony bool
	Source                     string
	SourceCardID               string
}

// PendingFreeTradeSelection represents a pending free trade colony selection from a card effect
type PendingFreeTradeSelection struct {
	AvailableColonyIDs []string
	Source             string
	SourceCardID       string
}

// ForcedFirstAction represents an action that must be completed first
type ForcedFirstAction struct {
	ActionType    string
	CorporationID string
	Source        string
	Completed     bool
	Description   string
}
