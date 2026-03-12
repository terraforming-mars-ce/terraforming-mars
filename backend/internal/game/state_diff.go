package game

import (
	"time"

	"terraforming-mars-backend/internal/game/shared"
)

// DiffValueString represents old/new values for string fields
type DiffValueString struct {
	Old string
	New string
}

// DiffValueInt represents old/new values for integer fields
type DiffValueInt struct {
	Old int
	New int
}

// DiffValueBool represents old/new values for boolean fields
type DiffValueBool struct {
	Old bool
	New bool
}

// TilePlacement records a single tile placement on the board
type TilePlacement struct {
	HexID    string
	TileType string
	OwnerID  string
}

// BoardChanges contains all changes to the game board
type BoardChanges struct {
	TilesPlaced []TilePlacement
}

// PlayerChanges contains all changes to a single player's state
type PlayerChanges struct {
	Credits            *DiffValueInt
	Steel              *DiffValueInt
	Titanium           *DiffValueInt
	Plants             *DiffValueInt
	Energy             *DiffValueInt
	Heat               *DiffValueInt
	TerraformRating    *DiffValueInt
	CreditsProduction  *DiffValueInt
	SteelProduction    *DiffValueInt
	TitaniumProduction *DiffValueInt
	PlantsProduction   *DiffValueInt
	EnergyProduction   *DiffValueInt
	HeatProduction     *DiffValueInt
	CardsAdded         []string
	CardsRemoved       []string
	CardsPlayed        []string
	Corporation        *DiffValueString
	Passed             *DiffValueBool
}

// GameChanges contains all changes in a single state transition
type GameChanges struct {
	Status              *DiffValueString
	Phase               *DiffValueString
	Generation          *DiffValueInt
	CurrentTurnPlayerID *DiffValueString
	Temperature         *DiffValueInt
	Oxygen              *DiffValueInt
	Oceans              *DiffValueInt
	PlayerChanges       map[string]*PlayerChanges
	BoardChanges        *BoardChanges
}

// LogDisplayData contains pre-computed display information for log entries
type LogDisplayData struct {
	Behaviors    []shared.CardBehavior
	Tags         []shared.CardTag
	VPConditions []shared.VPConditionForLog
}

// StateDiff represents the difference between two consecutive game states
type StateDiff struct {
	SequenceNumber    int64
	Timestamp         time.Time
	GameID            string
	Changes           *GameChanges
	Source            string
	SourceType        shared.SourceType
	PlayerID          string
	Description       string
	ChoiceIndex       *int                      // For cards with choices, which choice was selected (0-indexed)
	CalculatedOutputs []shared.CalculatedOutput // Actual values applied (for scaled outputs like "per X tags")
	DisplayData       *LogDisplayData           // Pre-computed display information for log entries
}

// DiffLog contains the complete history of state changes for a game
type DiffLog struct {
	GameID          string
	Diffs           []StateDiff
	CurrentSequence int64
}

// NewDiffLog creates a new empty diff log for a game
func NewDiffLog(gameID string) *DiffLog {
	return &DiffLog{
		GameID:          gameID,
		Diffs:           []StateDiff{},
		CurrentSequence: 0,
	}
}

// Append adds a new diff to the log and returns the sequence number
func (dl *DiffLog) Append(changes *GameChanges, source string, sourceType shared.SourceType, playerID, description string) int64 {
	return dl.AppendWithChoice(changes, source, sourceType, playerID, description, nil)
}

// AppendWithChoice adds a new diff with an optional choice index and returns the sequence number
func (dl *DiffLog) AppendWithChoice(changes *GameChanges, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int) int64 {
	return dl.AppendWithChoiceAndOutputs(changes, source, sourceType, playerID, description, choiceIndex, nil)
}

// AppendWithChoiceAndOutputs adds a new diff with optional choice index and calculated outputs
func (dl *DiffLog) AppendWithChoiceAndOutputs(changes *GameChanges, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []shared.CalculatedOutput) int64 {
	return dl.AppendFull(changes, source, sourceType, playerID, description, choiceIndex, calculatedOutputs, nil)
}

// AppendFull adds a new diff with all optional fields including display data
func (dl *DiffLog) AppendFull(changes *GameChanges, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []shared.CalculatedOutput, displayData *LogDisplayData) int64 {
	dl.CurrentSequence++
	diff := StateDiff{
		SequenceNumber:    dl.CurrentSequence,
		Timestamp:         time.Now(),
		GameID:            dl.GameID,
		Changes:           changes,
		Source:            source,
		SourceType:        sourceType,
		PlayerID:          playerID,
		Description:       description,
		ChoiceIndex:       choiceIndex,
		CalculatedOutputs: calculatedOutputs,
		DisplayData:       displayData,
	}
	dl.Diffs = append(dl.Diffs, diff)
	return dl.CurrentSequence
}

// GetAll returns all diffs in chronological order
func (dl *DiffLog) GetAll() []StateDiff {
	result := make([]StateDiff, len(dl.Diffs))
	copy(result, dl.Diffs)
	return result
}

// diffInt compares two integers and returns a DiffValueInt if different
func diffInt(old, new int) *DiffValueInt {
	if old == new {
		return nil
	}
	return &DiffValueInt{Old: old, New: new}
}

// diffString compares two strings and returns a DiffValueString if different
func diffString(old, new string) *DiffValueString {
	if old == new {
		return nil
	}
	return &DiffValueString{Old: old, New: new}
}

// diffBool compares two booleans and returns a DiffValueBool if different
func diffBool(old, new bool) *DiffValueBool {
	if old == new {
		return nil
	}
	return &DiffValueBool{Old: old, New: new}
}

// diffStringSlice computes added and removed strings between two slices
func diffStringSlice(old, new []string) (added, removed []string) {
	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)

	for _, s := range old {
		oldSet[s] = true
	}
	for _, s := range new {
		newSet[s] = true
	}

	for s := range newSet {
		if !oldSet[s] {
			added = append(added, s)
		}
	}
	for s := range oldSet {
		if !newSet[s] {
			removed = append(removed, s)
		}
	}

	return added, removed
}
