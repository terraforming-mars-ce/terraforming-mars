package dto

// DiffValueStringDto represents old/new values for string fields
type DiffValueStringDto struct {
	Old string `json:"old" ts:"string"`
	New string `json:"new" ts:"string"`
}

// DiffValueIntDto represents old/new values for integer fields
type DiffValueIntDto struct {
	Old int `json:"old" ts:"number"`
	New int `json:"new" ts:"number"`
}

// DiffValueBoolDto represents old/new values for boolean fields
type DiffValueBoolDto struct {
	Old bool `json:"old" ts:"boolean"`
	New bool `json:"new" ts:"boolean"`
}

// TilePlacementDto records a single tile placement on the board
type TilePlacementDto struct {
	HexID    string `json:"hexId" ts:"string"`
	TileType string `json:"tileType" ts:"string"`
	OwnerID  string `json:"ownerId,omitempty" ts:"string | undefined"`
}

// BoardChangesDto contains all changes to the game board
type BoardChangesDto struct {
	TilesPlaced []TilePlacementDto `json:"tilesPlaced,omitempty" ts:"TilePlacementDto[] | undefined"`
}

// PlayerChangesDto contains all changes to a single player's state
type PlayerChangesDto struct {
	Credits            *DiffValueIntDto    `json:"credits,omitempty" ts:"DiffValueIntDto | undefined"`
	Steel              *DiffValueIntDto    `json:"steel,omitempty" ts:"DiffValueIntDto | undefined"`
	Titanium           *DiffValueIntDto    `json:"titanium,omitempty" ts:"DiffValueIntDto | undefined"`
	Plants             *DiffValueIntDto    `json:"plants,omitempty" ts:"DiffValueIntDto | undefined"`
	Energy             *DiffValueIntDto    `json:"energy,omitempty" ts:"DiffValueIntDto | undefined"`
	Heat               *DiffValueIntDto    `json:"heat,omitempty" ts:"DiffValueIntDto | undefined"`
	TerraformRating    *DiffValueIntDto    `json:"terraformRating,omitempty" ts:"DiffValueIntDto | undefined"`
	CreditsProduction  *DiffValueIntDto    `json:"creditsProduction,omitempty" ts:"DiffValueIntDto | undefined"`
	SteelProduction    *DiffValueIntDto    `json:"steelProduction,omitempty" ts:"DiffValueIntDto | undefined"`
	TitaniumProduction *DiffValueIntDto    `json:"titaniumProduction,omitempty" ts:"DiffValueIntDto | undefined"`
	PlantsProduction   *DiffValueIntDto    `json:"plantsProduction,omitempty" ts:"DiffValueIntDto | undefined"`
	EnergyProduction   *DiffValueIntDto    `json:"energyProduction,omitempty" ts:"DiffValueIntDto | undefined"`
	HeatProduction     *DiffValueIntDto    `json:"heatProduction,omitempty" ts:"DiffValueIntDto | undefined"`
	CardsAdded         []string            `json:"cardsAdded,omitempty" ts:"string[] | undefined"`
	CardsRemoved       []string            `json:"cardsRemoved,omitempty" ts:"string[] | undefined"`
	CardsPlayed        []string            `json:"cardsPlayed,omitempty" ts:"string[] | undefined"`
	Corporation        *DiffValueStringDto `json:"corporation,omitempty" ts:"DiffValueStringDto | undefined"`
	Passed             *DiffValueBoolDto   `json:"passed,omitempty" ts:"DiffValueBoolDto | undefined"`
}

// GameChangesDto contains all changes in a single state transition
type GameChangesDto struct {
	Status              *DiffValueStringDto          `json:"status,omitempty" ts:"DiffValueStringDto | undefined"`
	Phase               *DiffValueStringDto          `json:"phase,omitempty" ts:"DiffValueStringDto | undefined"`
	Generation          *DiffValueIntDto             `json:"generation,omitempty" ts:"DiffValueIntDto | undefined"`
	CurrentTurnPlayerID *DiffValueStringDto          `json:"currentTurnPlayerId,omitempty" ts:"DiffValueStringDto | undefined"`
	Temperature         *DiffValueIntDto             `json:"temperature,omitempty" ts:"DiffValueIntDto | undefined"`
	Oxygen              *DiffValueIntDto             `json:"oxygen,omitempty" ts:"DiffValueIntDto | undefined"`
	Oceans              *DiffValueIntDto             `json:"oceans,omitempty" ts:"DiffValueIntDto | undefined"`
	PlayerChanges       map[string]*PlayerChangesDto `json:"playerChanges,omitempty" ts:"Record<string, PlayerChangesDto> | undefined"`
	BoardChanges        *BoardChangesDto             `json:"boardChanges,omitempty" ts:"BoardChangesDto | undefined"`
}

// CalculatedOutputDto represents an actual output value that was applied
type CalculatedOutputDto struct {
	ResourceType string `json:"resourceType" ts:"string"`
	Amount       int    `json:"amount" ts:"number"`
	IsScaled     bool   `json:"isScaled" ts:"boolean"`
}

// ComputedBehaviorValueDto holds pre-computed per-condition output values for a behavior.
// Target uses the format "behaviors::N" where N is the behavior index.
type ComputedBehaviorValueDto struct {
	Target  string                `json:"target" ts:"string"`
	Outputs []CalculatedOutputDto `json:"outputs" ts:"CalculatedOutputDto[]"`
}

// LogDisplayDataDto contains pre-computed display information for log entries
type LogDisplayDataDto struct {
	Behaviors    []CardBehaviorDto `json:"behaviors,omitempty" ts:"CardBehaviorDto[] | undefined"`
	Tags         []CardTag         `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	VPConditions []VPConditionDto  `json:"vpConditions,omitempty" ts:"VPConditionDto[] | undefined"`
}

// StateDiffDto represents the difference between two consecutive game states
type StateDiffDto struct {
	SequenceNumber    int64                 `json:"sequenceNumber" ts:"number"`
	Timestamp         string                `json:"timestamp" ts:"string"`
	GameID            string                `json:"gameId" ts:"string"`
	Changes           *GameChangesDto       `json:"changes" ts:"GameChangesDto"`
	Source            string                `json:"source" ts:"string"`
	SourceType        string                `json:"sourceType" ts:"string"`
	PlayerID          string                `json:"playerId" ts:"string"`
	Description       string                `json:"description" ts:"string"`
	ChoiceIndex       *int                  `json:"choiceIndex,omitempty" ts:"number | undefined"`
	CalculatedOutputs []CalculatedOutputDto `json:"calculatedOutputs,omitempty" ts:"CalculatedOutputDto[] | undefined"`
	DisplayData       *LogDisplayDataDto    `json:"displayData,omitempty" ts:"LogDisplayDataDto | undefined"`
}

// DiffLogDto contains the complete history of state changes for a game
type DiffLogDto struct {
	GameID          string         `json:"gameId" ts:"string"`
	Diffs           []StateDiffDto `json:"diffs" ts:"StateDiffDto[]"`
	CurrentSequence int64          `json:"currentSequence" ts:"number"`
}
