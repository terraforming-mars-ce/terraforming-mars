package player

import (
	"sync"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

// Selection manages player-specific card selection state
type Selection struct {
	mu                             sync.RWMutex
	selectCorporationPhase         *SelectCorporationPhase
	selectStartingCardsPhase       *SelectStartingCardsPhase
	selectPreludeCardsPhase        *SelectPreludeCardsPhase
	pendingCardSelection           *PendingCardSelection
	pendingCardDrawSelection       *PendingCardDrawSelection
	pendingCardDiscardSelection    *PendingCardDiscardSelection
	pendingBehaviorChoiceSelection *PendingBehaviorChoiceSelection
	pendingStealTargetSelection    *PendingStealTargetSelection
	pendingColonyResourceSelection *PendingColonyResourceSelection
	eventBus                       *events.EventBusImpl
	gameID                         string
	playerID                       string
}

func newSelection(eventBus *events.EventBusImpl, gameID, playerID string) *Selection {
	return &Selection{
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (s *Selection) GetSelectCorporationPhase() *SelectCorporationPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectCorporationPhase
}

func (s *Selection) SetSelectCorporationPhase(phase *SelectCorporationPhase) {
	s.mu.Lock()
	s.selectCorporationPhase = phase
	s.mu.Unlock()
}

func (s *Selection) GetSelectStartingCardsPhase() *SelectStartingCardsPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectStartingCardsPhase
}

func (s *Selection) SetSelectStartingCardsPhase(phase *SelectStartingCardsPhase) {
	s.mu.Lock()
	s.selectStartingCardsPhase = phase
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetSelectPreludeCardsPhase() *SelectPreludeCardsPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectPreludeCardsPhase
}

func (s *Selection) SetSelectPreludeCardsPhase(phase *SelectPreludeCardsPhase) {
	s.mu.Lock()
	s.selectPreludeCardsPhase = phase
	s.mu.Unlock()
}

func (s *Selection) GetPendingCardSelection() *PendingCardSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardSelection
}

func (s *Selection) SetPendingCardSelection(selection *PendingCardSelection) {
	s.mu.Lock()
	s.pendingCardSelection = selection
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetPendingCardDrawSelection() *PendingCardDrawSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardDrawSelection
}

func (s *Selection) SetPendingCardDrawSelection(selection *PendingCardDrawSelection) {
	s.mu.Lock()
	s.pendingCardDrawSelection = selection
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetPendingCardDiscardSelection() *PendingCardDiscardSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingCardDiscardSelection
}

func (s *Selection) SetPendingCardDiscardSelection(selection *PendingCardDiscardSelection) {
	s.mu.Lock()
	s.pendingCardDiscardSelection = selection
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetPendingBehaviorChoiceSelection() *PendingBehaviorChoiceSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingBehaviorChoiceSelection
}

func (s *Selection) SetPendingBehaviorChoiceSelection(selection *PendingBehaviorChoiceSelection) {
	s.mu.Lock()
	s.pendingBehaviorChoiceSelection = selection
	s.mu.Unlock()

	if s.eventBus != nil {
	}
}

func (s *Selection) GetPendingStealTargetSelection() *PendingStealTargetSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingStealTargetSelection
}

func (s *Selection) SetPendingStealTargetSelection(selection *PendingStealTargetSelection) {
	s.mu.Lock()
	s.pendingStealTargetSelection = selection
	s.mu.Unlock()
}

func (s *Selection) GetPendingColonyResourceSelection() *PendingColonyResourceSelection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingColonyResourceSelection
}

func (s *Selection) SetPendingColonyResourceSelection(selection *PendingColonyResourceSelection) {
	s.mu.Lock()
	s.pendingColonyResourceSelection = selection
	s.mu.Unlock()
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
	SourceCardID        string // Card that triggered this selection (for card actions)
	SourceBehaviorIndex int    // Behavior index of the card action
	PlayAsPrelude       bool   // When true, selected card is played as prelude instead of added to hand
}

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
	BeforeResources   shared.Resources
	AfterResources    shared.Resources
	EnergyConverted   int
	CreditsIncome     int
}

// TileCompletionCallback stores info about what to call when tile placement completes
type TileCompletionCallback struct {
	Type string
	Data map[string]interface{}
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string
	AvailableHexes []string
	Source         string
	SourceCardID   string
	OnComplete     *TileCompletionCallback
}

// PendingTileSelectionQueue represents a queue of tile placements
type PendingTileSelectionQueue struct {
	Items            []string
	Source           string
	SourceCardID     string
	OnComplete       *TileCompletionCallback
	TileRestrictions *shared.TileRestrictions
}

// PendingCardDiscardSelection represents a pending card discard action
// The player must select card(s) to discard from hand, after which pending outputs are applied.
type PendingCardDiscardSelection struct {
	MinCards       int                        // 0 if optional (player can skip), 1+ if mandatory
	MaxCards       int                        // Maximum cards to discard
	Source         string                     // Card name that triggered this selection
	SourceCardID   string                     // Card ID that triggered this selection
	PendingOutputs []shared.ResourceCondition // Outputs to apply after discard completes
}

// PendingBehaviorChoiceSelection represents a pending behavior choice from a passive triggered effect.
// The player must select which choice to apply.
type PendingBehaviorChoiceSelection struct {
	Choices      []shared.Choice
	Source       string // Card name that owns the effect
	SourceCardID string // Card ID that owns the effect
}

// PendingStealTargetSelection represents a pending target player selection for
// post-tile-placement resource steal (e.g., Flooding card).
type PendingStealTargetSelection struct {
	EligiblePlayerIDs []string
	ResourceType      shared.ResourceType
	Amount            int
	Source            string
	SourceCardID      string
}

// PendingColonyResourceSelection represents a pending card storage selection
// from a colony trade or build that produced card-targeted resources (microbes, animals, floaters).
type PendingColonyResourceSelection struct {
	ResourceType string // "microbe", "animal", "floater"
	Amount       int
	Source       string // Colony name
	ColonyID     string
	Reason       string // "trade", "colony-bonus", "build"
}

// ForcedFirstAction represents an action that must be completed as first action
type ForcedFirstAction struct {
	ActionType    string
	CorporationID string
	Source        string
	Completed     bool
	Description   string
}
