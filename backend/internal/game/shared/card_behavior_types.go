package shared

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      string                    `json:"type"`
	Condition *ResourceTriggerCondition `json:"condition,omitempty"`
}

// TriggerType constants for card behavior triggers
const (
	TriggerTypeAuto   = "auto"   // Automatic trigger (immediate effect when card is played)
	TriggerTypeManual = "manual" // Manual trigger (blue card action, activated by player)
)

// MinMaxValue represents a min/max value constraint
type MinMaxValue struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// Selector represents matching criteria for cards, resources, or projects.
// Multiple fields within a Selector use AND logic (all must match).
// Multiple Selectors in a slice use OR logic (any match is sufficient).
type Selector struct {
	Tags                 []CardTag         `json:"tags,omitempty"`
	CardTypes            []string          `json:"cardTypes,omitempty"`
	Resources            []string          `json:"resources,omitempty"`
	StandardProjects     []StandardProject `json:"standardProjects,omitempty"`
	RequiredOriginalCost *MinMaxValue      `json:"requiredOriginalCost,omitempty"`
	VP                   *MinMaxValue      `json:"vp,omitempty"`
	GlobalParameters     []string          `json:"globalParameters,omitempty"`
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type                 string         `json:"type"`
	ResourceTypes        []ResourceType `json:"resourceTypes,omitempty"`
	Location             *string        `json:"location,omitempty"`
	Selectors            []Selector     `json:"selectors,omitempty"`
	Target               *string        `json:"target,omitempty"`
	RequiredOriginalCost *MinMaxValue   `json:"requiredOriginalCost,omitempty"`
	OnBonusType          []string       `json:"onBonusType,omitempty"`
}

// TileRestrictions represents restrictions for tile placement
type TileRestrictions struct {
	BoardTags         []string `json:"boardTags,omitempty" ts:"string[]"`
	Adjacency         string   `json:"adjacency,omitempty" ts:"string"`                     // "none" = no adjacent occupied tiles
	OnTileType        string   `json:"onTileType,omitempty" ts:"string"`                    // "ocean" = only on ocean spaces
	AdjacentToType    string   `json:"adjacentToType,omitempty" ts:"string"`                // "city", "greenery" = must be adjacent to this tile type
	MinAdjacentOfType *int     `json:"minAdjacentOfType,omitempty" ts:"number | undefined"` // min count of adjacent tiles of AdjacentToType
	AdjacentToOwned   bool     `json:"adjacentToOwned,omitempty" ts:"boolean | undefined"`  // must be adjacent to a tile owned by the placing player
	OnBonusType       []string `json:"onBonusType,omitempty" ts:"string[]"`                 // tile must have one of these bonus types (e.g., "steel", "titanium")
}

// Temporary effect expiry constants
const (
	TemporaryNextCard      = "next-card"      // Effect expires after the next card is played
	TemporaryGenerationEnd = "generation-end" // Effect expires at end of generation
)

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	ResourceType               ResourceType       `json:"type"`
	Amount                     int                `json:"amount"`
	Target                     string             `json:"target"`
	Selectors                  []Selector         `json:"selectors,omitempty"`
	MaxTrigger                 *int               `json:"maxTrigger,omitempty"`
	Per                        *PerCondition      `json:"per,omitempty"`
	TileRestrictions           *TileRestrictions  `json:"tileRestrictions,omitempty" ts:"TileRestrictions | undefined"`
	TileType                   string             `json:"tileType,omitempty" ts:"string | undefined"` // For tile-placement: specifies the tile type to place
	VariableAmount             bool               `json:"variableAmount,omitempty" ts:"boolean | undefined"`
	Temporary                  string             `json:"temporary,omitempty" ts:"string | undefined"`              // "next-card" or "generation-end"
	Optional                   bool               `json:"optional,omitempty" ts:"boolean | undefined"`              // Player can skip this input
	PaymentAllowed             []ResourceType     `json:"paymentAllowed,omitempty" ts:"ResourceType[] | undefined"` // Alternative resources accepted as payment (e.g., ["titanium"] for "titanium may be used")
	TargetRestriction          *TargetRestriction `json:"targetRestriction,omitempty" ts:"TargetRestriction | undefined"`
	AllowDuplicatePlayerColony bool               `json:"allowDuplicatePlayerColony,omitempty" ts:"boolean | undefined"`
}

// TargetRestriction restricts which players can be targeted by an output.
type TargetRestriction struct {
	Adjacent string `json:"adjacent,omitempty" ts:"string"` // "self-card" = only players with tiles adjacent to this card's tile placement
}

// PerCondition represents what to count for conditional resource gains.
// Used by card behaviors, VP conditions, and award quantifiers.
type PerCondition struct {
	ResourceType       ResourceType  `json:"type"`
	Amount             int           `json:"amount"`
	Location           *string       `json:"location,omitempty"`
	Target             *string       `json:"target,omitempty"`
	Tag                *CardTag      `json:"tag,omitempty"`
	AdjacentToTileType *ResourceType `json:"adjacentToTileType,omitempty"`
	AdjacentToSelfTile bool          `json:"adjacentToSelfTile,omitempty"`
}

// ChoiceRequirement represents a requirement that gates whether a choice is available to a player.
// Uses raw string types to avoid circular dependency with cards/ package.
type ChoiceRequirement struct {
	Type     string        `json:"type"`                                             // Same values as cards.RequirementType (e.g. "tags", "temperature")
	Min      *int          `json:"min,omitempty"`                                    // Minimum value
	Max      *int          `json:"max,omitempty"`                                    // Maximum value
	Location *string       `json:"location,omitempty"`                               // Location constraint (e.g. "mars", "anywhere")
	Tag      *CardTag      `json:"tag,omitempty"`                                    // Tag to count (for type "tags")
	Resource *ResourceType `json:"resource,omitempty" ts:"ResourceType | undefined"` // Resource type (for type "production" or "resource")
}

// ChoiceRequirements wraps a list of requirements for a choice option
type ChoiceRequirements struct {
	Items []ChoiceRequirement `json:"items"`
}

// Choice represents a player choice option
type Choice struct {
	Inputs       []ResourceCondition `json:"inputs,omitempty"`
	Outputs      []ResourceCondition `json:"outputs,omitempty"`
	Requirements *ChoiceRequirements `json:"requirements,omitempty"` // If set, choice is only available when requirements are met
}
