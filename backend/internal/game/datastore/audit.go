package datastore

import (
	"time"

	"terraforming-mars-backend/internal/game/shared"
)

// GameSnapshot is a point-in-time capture of game state.
type GameSnapshot struct {
	GameID      string
	Status      string
	Phase       string
	Generation  int
	CurrentTurn string
	Temperature int
	Oxygen      int
	Oceans      int
	Players     map[string]*PlayerSnapshot
	Tiles       map[string]*TileSnapshot
}

type PlayerSnapshot struct {
	Credits            int
	Steel              int
	Titanium           int
	Plants             int
	Energy             int
	Heat               int
	TerraformRating    int
	CreditsProduction  int
	SteelProduction    int
	TitaniumProduction int
	PlantsProduction   int
	EnergyProduction   int
	HeatProduction     int
	Corporation        string
	Passed             bool
	HandCardIDs        []string
	PlayedCardIDs      []string
}

type TileSnapshot struct {
	HexID    string
	TileType string
	OwnerID  string
}

type DiffLog struct {
	GameID          string
	Diffs           []StateDiff
	CurrentSequence int64
}

type StateDiff struct {
	SequenceNumber    int64
	Timestamp         time.Time
	GameID            string
	Changes           *GameChanges
	Source            string
	SourceType        string
	PlayerID          string
	Description       string
	ChoiceIndex       *int
	CalculatedOutputs []shared.CalculatedOutput
	DisplayData       *LogDisplayData
}

type LogDisplayData struct {
	Behaviors    []shared.CardBehavior
	Tags         []shared.CardTag
	VPConditions []shared.VPConditionForLog
}

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

type BoardChanges struct {
	TilesPlaced []TilePlacement
}

type TilePlacement struct {
	HexID    string
	TileType string
	OwnerID  string
}

type DiffValueString struct {
	Old string
	New string
}

type DiffValueInt struct {
	Old int
	New int
}

type DiffValueBool struct {
	Old bool
	New bool
}
