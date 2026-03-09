package dto

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingSelection     GamePhase = "starting_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseDemoSetup             GamePhase = "demo_setup"
	GamePhaseInitApplyCorp         GamePhase = "init_apply_corp"
	GamePhaseInitApplyPrelude      GamePhase = "init_apply_prelude"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseComplete              GamePhase = "complete"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// CardType represents different types of cards
type CardType string

const (
	CardTypeAutomated   CardType = "automated"
	CardTypeActive      CardType = "active"
	CardTypeEvent       CardType = "event"
	CardTypeCorporation CardType = "corporation"
	CardTypePrelude     CardType = "prelude"
)

// StandardProject represents the different types of standard projects
type StandardProject string

const (
	StandardProjectSellPatents StandardProject = "sell-patents"
	StandardProjectPowerPlant  StandardProject = "power-plant"
	StandardProjectAsteroid    StandardProject = "asteroid"
	StandardProjectAquifer     StandardProject = "aquifer"
	StandardProjectGreenery    StandardProject = "greenery"
	StandardProjectCity        StandardProject = "city"

	StandardProjectConvertPlantsToGreenery  StandardProject = "convert-plants-to-greenery"
	StandardProjectConvertHeatToTemperature StandardProject = "convert-heat-to-temperature"
)

// CardTag represents different card categories and attributes
type CardTag string

const (
	TagSpace    CardTag = "space"
	TagEarth    CardTag = "earth"
	TagScience  CardTag = "science"
	TagPower    CardTag = "power"
	TagBuilding CardTag = "building"
	TagMicrobe  CardTag = "microbe"
	TagAnimal   CardTag = "animal"
	TagPlant    CardTag = "plant"
	TagEvent    CardTag = "event"
	TagCity     CardTag = "city"
	TagVenus    CardTag = "venus"
	TagJovian   CardTag = "jovian"
	TagWildlife CardTag = "wildlife"
	TagWild     CardTag = "wild"
)

// ResourceType represents different types of resources for client consumption
// This is a 1:1 mapping from types.ResourceType
type ResourceType string

const (
	ResourceTypeCredit   ResourceType = "credit"
	ResourceTypeSteel    ResourceType = "steel"
	ResourceTypeTitanium ResourceType = "titanium"
	ResourceTypePlant    ResourceType = "plant"
	ResourceTypeEnergy   ResourceType = "energy"
	ResourceTypeHeat     ResourceType = "heat"
	ResourceTypeMicrobe  ResourceType = "microbe"
	ResourceTypeAnimal   ResourceType = "animal"
	ResourceTypeFloater  ResourceType = "floater"
	ResourceTypeScience  ResourceType = "science"
	ResourceTypeAsteroid ResourceType = "asteroid"
	ResourceTypeFighter  ResourceType = "fighter"
	ResourceTypeDisease  ResourceType = "disease"

	ResourceTypeCardDraw ResourceType = "card-draw"
	ResourceTypeCardTake ResourceType = "card-take"
	ResourceTypeCardPeek ResourceType = "card-peek"

	ResourceTypeCityPlacement     ResourceType = "city-placement"
	ResourceTypeOceanPlacement    ResourceType = "ocean-placement"
	ResourceTypeGreeneryPlacement ResourceType = "greenery-placement"

	ResourceTypeCityTile     ResourceType = "city-tile"
	ResourceTypeOceanTile    ResourceType = "ocean-tile"
	ResourceTypeGreeneryTile ResourceType = "greenery-tile"
	ResourceTypeColonyTile   ResourceType = "colony-tile"

	ResourceTypeTemperature ResourceType = "temperature"
	ResourceTypeOxygen      ResourceType = "oxygen"
	ResourceTypeVenus       ResourceType = "venus"
	ResourceTypeTR          ResourceType = "tr"

	ResourceTypeCreditProduction   ResourceType = "credit-production"
	ResourceTypeSteelProduction    ResourceType = "steel-production"
	ResourceTypeTitaniumProduction ResourceType = "titanium-production"
	ResourceTypePlantProduction    ResourceType = "plant-production"
	ResourceTypeEnergyProduction   ResourceType = "energy-production"
	ResourceTypeHeatProduction     ResourceType = "heat-production"

	ResourceTypeEffect ResourceType = "effect"
	ResourceTypeTag    ResourceType = "tag"

	ResourceTypeGlobalParameterLenience ResourceType = "global-parameter-lenience"
	ResourceTypeDefense                 ResourceType = "defense"
	ResourceTypeDiscount                ResourceType = "discount"
	ResourceTypeValueModifier           ResourceType = "value-modifier"
)

// TargetType represents different targeting scopes for resource conditions for client consumption
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player"
	TargetSelfCard   TargetType = "self-card"
	TargetAnyCard    TargetType = "any-card"
	TargetAnyPlayer  TargetType = "any-player"
	TargetOpponent   TargetType = "opponent"
	TargetNone       TargetType = "none"
)

// CardApplyLocation represents different locations where card conditions can be evaluated for client consumption
type CardApplyLocation string

const (
	CardApplyLocationAnywhere CardApplyLocation = "anywhere"
	CardApplyLocationMars     CardApplyLocation = "mars"
)

// RequirementType represents different card requirement types for client consumption
type RequirementType string

const (
	RequirementTemperature RequirementType = "temperature"
	RequirementOxygen      RequirementType = "oxygen"
	RequirementOceans      RequirementType = "oceans"
	RequirementVenus       RequirementType = "venus"
	RequirementCities      RequirementType = "cities"
	RequirementGreeneries  RequirementType = "greeneries"
	RequirementTags        RequirementType = "tags"
	RequirementProduction  RequirementType = "production"
	RequirementTR          RequirementType = "tr"
	RequirementResource    RequirementType = "resource"
)

// VPConditionType represents different types of VP conditions for client consumption
type VPConditionType string

const (
	VPConditionFixed       VPConditionType = "fixed"
	VPConditionPer         VPConditionType = "per"
	VPConditionResourcesOn VPConditionType = "resources-on"
)

// TriggerType represents different trigger conditions for client consumption
type TriggerType string

const (
	TriggerOceanPlaced           TriggerType = "ocean-placed"
	TriggerGlobalParameterRaised TriggerType = "global-parameter-raised"
	TriggerCityPlaced            TriggerType = "city-placed"
	TriggerCardPlayed            TriggerType = "card-played"
	TriggerTagPlayed             TriggerType = "tag-played"
	TriggerTilePlaced            TriggerType = "tile-placed"
)

// ResourceTriggerType represents different trigger types for resource exchanges for client consumption
type ResourceTriggerType string

const (
	ResourceTriggerManual                     ResourceTriggerType = "manual"
	ResourceTriggerAuto                       ResourceTriggerType = "auto"
	ResourceTriggerAutoCorporationFirstAction ResourceTriggerType = "auto-corporation-first-action"
	ResourceTriggerAutoCorporationStart       ResourceTriggerType = "auto-corporation-start"
)

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// TileRestrictionsDto represents tile placement restrictions for client consumption
type TileRestrictionsDto struct {
	BoardTags         []string `json:"boardTags,omitempty" ts:"string[] | undefined"`
	Adjacency         string   `json:"adjacency,omitempty" ts:"string | undefined"`         // "none" = no adjacent occupied tiles
	OnTileType        string   `json:"onTileType,omitempty" ts:"string | undefined"`        // "ocean" = only on ocean spaces
	AdjacentToType    string   `json:"adjacentToType,omitempty" ts:"string | undefined"`    // "city", "greenery" = must be adjacent to this tile type
	MinAdjacentOfType *int     `json:"minAdjacentOfType,omitempty" ts:"number | undefined"` // min count of adjacent tiles of AdjacentToType
	AdjacentToOwned   *bool    `json:"adjacentToOwned,omitempty" ts:"boolean | undefined"`  // must be adjacent to a tile owned by the placing player
	OnBonusType       []string `json:"onBonusType,omitempty" ts:"string[] | undefined"`     // tile must have one of these bonus types
}

// SelectorDto represents matching criteria for cards, resources, or projects.
// Multiple fields within a Selector use AND logic (all must match).
// Multiple Selectors in a slice use OR logic (any match is sufficient).
type SelectorDto struct {
	Tags                 []CardTag         `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	CardTypes            []CardType        `json:"cardTypes,omitempty" ts:"CardType[] | undefined"`
	Resources            []string          `json:"resources,omitempty" ts:"string[] | undefined"`
	StandardProjects     []StandardProject `json:"standardProjects,omitempty" ts:"StandardProject[] | undefined"`
	RequiredOriginalCost *MinMaxValueDto   `json:"requiredOriginalCost,omitempty" ts:"MinMaxValueDto | undefined"`
	VP                   *MinMaxValueDto   `json:"vp,omitempty" ts:"MinMaxValueDto | undefined"`
	GlobalParameters     []string          `json:"globalParameters,omitempty" ts:"string[] | undefined"`
}

// ResourceConditionDto represents a resource condition for client consumption
type ResourceConditionDto struct {
	Type             ResourceType         `json:"type" ts:"ResourceType"`
	Amount           int                  `json:"amount" ts:"number"`
	Target           TargetType           `json:"target" ts:"TargetType"`
	Selectors        []SelectorDto        `json:"selectors,omitempty" ts:"SelectorDto[] | undefined"`
	MaxTrigger       *int                 `json:"maxTrigger,omitempty" ts:"number | undefined"`
	Per              *PerConditionDto     `json:"per,omitempty" ts:"PerConditionDto | undefined"`
	TileRestrictions *TileRestrictionsDto `json:"tileRestrictions,omitempty" ts:"TileRestrictionsDto | undefined"`
	TileType         string               `json:"tileType,omitempty" ts:"string | undefined"`
	VariableAmount   *bool                `json:"variableAmount,omitempty" ts:"boolean | undefined"`
	Optional         *bool                `json:"optional,omitempty" ts:"boolean | undefined"`
	PaymentAllowed   []ResourceType       `json:"paymentAllowed,omitempty" ts:"ResourceType[] | undefined"`
}

// PerConditionDto represents a per condition for client consumption
type PerConditionDto struct {
	Type     ResourceType       `json:"type" ts:"ResourceType"`
	Amount   int                `json:"amount" ts:"number"`
	Location *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	Target   *TargetType        `json:"target,omitempty" ts:"TargetType | undefined"`
	Tag      *CardTag           `json:"tag,omitempty" ts:"CardTag | undefined"`
}

// ChoiceDto represents a choice for client consumption
type ChoiceDto struct {
	OriginalIndex int                    `json:"originalIndex" ts:"number"`
	Inputs        []ResourceConditionDto `json:"inputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Outputs       []ResourceConditionDto `json:"outputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Requirements  *CardRequirementsDto   `json:"requirements,omitempty" ts:"CardRequirementsDto | undefined"`
	Available     bool                   `json:"available" ts:"boolean"`
	Errors        []StateErrorDto        `json:"errors" ts:"StateErrorDto[]"`
}

// TriggerDto represents a trigger for client consumption
type TriggerDto struct {
	Type      ResourceTriggerType          `json:"type" ts:"ResourceTriggerType"`
	Condition *ResourceTriggerConditionDto `json:"condition,omitempty" ts:"ResourceTriggerConditionDto | undefined"`
}

// MinMaxValueDto represents a minimum and/or maximum value constraint for client consumption
type MinMaxValueDto struct {
	Min *int `json:"min,omitempty" ts:"number | undefined"`
	Max *int `json:"max,omitempty" ts:"number | undefined"`
}

// ResourceTriggerConditionDto represents a resource trigger condition for client consumption
type ResourceTriggerConditionDto struct {
	Type                   TriggerType                     `json:"type" ts:"TriggerType"`
	ResourceTypes          []ResourceType                  `json:"resourceTypes,omitempty" ts:"ResourceType[] | undefined"`
	Location               *CardApplyLocation              `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	Selectors              []SelectorDto                   `json:"selectors,omitempty" ts:"SelectorDto[] | undefined"`
	Target                 *TargetType                     `json:"target,omitempty" ts:"TargetType | undefined"`
	RequiredOriginalCost   *MinMaxValueDto                 `json:"requiredOriginalCost,omitempty" ts:"MinMaxValueDto | undefined"`
	RequiredResourceChange map[ResourceType]MinMaxValueDto `json:"requiredResourceChange,omitempty" ts:"Record<ResourceType, MinMaxValueDto> | undefined"`
	OnBonusType            []string                        `json:"onBonusType,omitempty" ts:"string[] | undefined"`
}

// CardBehaviorDto represents a card behavior for client consumption
type CardBehaviorDto struct {
	Description                   string                            `json:"description,omitempty" ts:"string | undefined"`
	Triggers                      []TriggerDto                      `json:"triggers,omitempty" ts:"TriggerDto[] | undefined"`
	Inputs                        []ResourceConditionDto            `json:"inputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Outputs                       []ResourceConditionDto            `json:"outputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Choices                       []ChoiceDto                       `json:"choices,omitempty" ts:"ChoiceDto[] | undefined"`
	ChoicePolicy                  string                            `json:"choicePolicy,omitempty" ts:"string | undefined"`
	GenerationalEventRequirements []GenerationalEventRequirementDto `json:"generationalEventRequirements,omitempty" ts:"GenerationalEventRequirementDto[] | undefined"`
	Group                         string                            `json:"group,omitempty" ts:"string | undefined"`
}

// PaymentConstantsDto represents payment conversion rates
type PaymentConstantsDto struct {
	SteelValue    int `json:"steelValue" ts:"number"`
	TitaniumValue int `json:"titaniumValue" ts:"number"`
}

// CardRequirementsDto wraps requirement items with a description for client consumption
type CardRequirementsDto struct {
	Description string           `json:"description,omitempty" ts:"string | undefined"`
	Items       []RequirementDto `json:"items" ts:"RequirementDto[]"`
}

// RequirementDto represents a card requirement for client consumption
type RequirementDto struct {
	Type     RequirementType    `json:"type" ts:"RequirementType"`
	Min      *int               `json:"min,omitempty" ts:"number | undefined"`
	Max      *int               `json:"max,omitempty" ts:"number | undefined"`
	Location *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	Tag      *CardTag           `json:"tag,omitempty" ts:"CardTag | undefined"`
	Resource *ResourceType      `json:"resource,omitempty" ts:"ResourceType | undefined"`
}

// ResourceStorageDto represents a card's resource storage for client consumption
type ResourceStorageDto struct {
	Type        ResourceType `json:"type" ts:"ResourceType"`
	Capacity    *int         `json:"capacity,omitempty" ts:"number | undefined"`
	Starting    int          `json:"starting" ts:"number"`
	Description string       `json:"description,omitempty" ts:"string | undefined"`
}

// VPConditionDto represents a victory point condition for client consumption
type VPConditionDto struct {
	Amount      int              `json:"amount" ts:"number"`
	Condition   VPConditionType  `json:"condition" ts:"VPConditionType"`
	MaxTrigger  *int             `json:"maxTrigger,omitempty" ts:"number | undefined"`
	Per         *PerConditionDto `json:"per,omitempty" ts:"PerConditionDto | undefined"`
	Description string           `json:"description,omitempty" ts:"string | undefined"`
}

// CardDto represents a card for client consumption
type CardDto struct {
	ID              string               `json:"id" ts:"string"`
	Name            string               `json:"name" ts:"string"`
	Type            CardType             `json:"type" ts:"CardType"`
	Cost            int                  `json:"cost" ts:"number"`
	Description     string               `json:"description" ts:"string"`
	Pack            string               `json:"pack" ts:"string"`
	Tags            []CardTag            `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	Requirements    *CardRequirementsDto `json:"requirements,omitempty" ts:"CardRequirementsDto | undefined"`
	Behaviors       []CardBehaviorDto    `json:"behaviors,omitempty" ts:"CardBehaviorDto[] | undefined"`
	ResourceStorage *ResourceStorageDto  `json:"resourceStorage,omitempty" ts:"ResourceStorageDto | undefined"`
	VPConditions    []VPConditionDto     `json:"vpConditions,omitempty" ts:"VPConditionDto[] | undefined"`

	StartingResources  *ResourceSet `json:"startingResources,omitempty" ts:"ResourceSet | undefined"`
	StartingProduction *ResourceSet `json:"startingProduction,omitempty" ts:"ResourceSet | undefined"`
}

// SelectCorporationPhaseDto represents corporation selection state for the current player
type SelectCorporationPhaseDto struct {
	AvailableCorporations []CardDto `json:"availableCorporations" ts:"CardDto[]"`
}

// SelectCorporationOtherPlayerDto represents corporation selection state for other players
type SelectCorporationOtherPlayerDto struct{}

type SelectStartingCardsPhaseDto struct {
	AvailableCards []CardDto `json:"availableCards" ts:"CardDto[]"`
}

type SelectStartingCardsOtherPlayerDto struct{}

// SelectPreludeCardsPhaseDto represents prelude card selection state for the current player
type SelectPreludeCardsPhaseDto struct {
	AvailablePreludes []CardDto `json:"availablePreludes" ts:"CardDto[]"`
	MaxSelectable     int       `json:"maxSelectable" ts:"number"`
}

// SelectPreludeCardsOtherPlayerDto represents prelude card selection state for other players
type SelectPreludeCardsOtherPlayerDto struct {
	// Empty - other players don't see selection details
}

// ProductionPhaseDto represents card selection and production phase state for a player
type ProductionPhaseDto struct {
	AvailableCards    []CardDto    `json:"availableCards" ts:"CardDto[]"`  // Cards available for selection
	SelectionComplete bool         `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
	BeforeResources   ResourcesDto `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources    ResourcesDto `json:"afterResources" ts:"ResourcesDto"`
	ResourceDelta     ResourcesDto `json:"resourceDelta" ts:"ResourceDelta"`
	EnergyConverted   int          `json:"energyConverted" ts:"number"`
	CreditsIncome     int          `json:"creditsIncome" ts:"number"`
}

type ProductionPhaseOtherPlayerDto struct {
	SelectionComplete bool         `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
	BeforeResources   ResourcesDto `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources    ResourcesDto `json:"afterResources" ts:"ResourcesDto"`
	ResourceDelta     ResourcesDto `json:"resourceDelta" ts:"ResourceDelta"`
	EnergyConverted   int          `json:"energyConverted" ts:"number"`
	CreditsIncome     int          `json:"creditsIncome" ts:"number"`
}

// GameSettingsDto contains configurable game parameters
type GameSettingsDto struct {
	MaxPlayers            int      `json:"maxPlayers" ts:"number"`
	VenusNextEnabled      bool     `json:"venusNextEnabled" ts:"boolean"`
	DevelopmentMode       bool     `json:"developmentMode" ts:"boolean"`
	DemoGame              bool     `json:"demoGame" ts:"boolean"`
	CardPacks             []string `json:"cardPacks,omitempty" ts:"string[] | undefined"`
	HasClaudeAPIKey       bool     `json:"hasClaudeApiKey" ts:"boolean"`
	ClaudeModel           string   `json:"claudeModel,omitempty" ts:"string | undefined"`
	AvailablePlayerColors []string `json:"availablePlayerColors" ts:"string[]"`
}

// GlobalParametersDto represents the terraforming progress
type GlobalParametersDto struct {
	Temperature int `json:"temperature" ts:"number"` // Range: -30 to +8°C
	Oxygen      int `json:"oxygen" ts:"number"`      // Range: 0-14%
	Oceans      int `json:"oceans" ts:"number"`      // Range: 0-9
	MaxOceans   int `json:"maxOceans" ts:"number"`   // Dynamic max, starts at 9
	Venus       int `json:"venus" ts:"number"`       // Range: 0-30%
}

// ResourcesDto represents a player's resources
type ResourcesDto struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// ProductionDto represents a player's production values
type ProductionDto struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// PaymentSubstituteDto represents an alternative resource that can be used as payment for credits
type PaymentSubstituteDto struct {
	ResourceType   ResourceType `json:"resourceType" ts:"ResourceType"`
	ConversionRate int          `json:"conversionRate" ts:"number"`
}

// StoragePaymentSubstituteDto represents card storage resources that can be used as M€ payment
type StoragePaymentSubstituteDto struct {
	CardID         string       `json:"cardId" ts:"string"`
	ResourceType   ResourceType `json:"resourceType" ts:"ResourceType"`
	ConversionRate int          `json:"conversionRate" ts:"number"`
}

// StateErrorCode represents error codes for entity state validation.
// All codes use kebab-case for consistency with JSON serialization.
type StateErrorCode string

const (
	ErrorCodeNotYourTurn StateErrorCode = "not-your-turn"
	ErrorCodeWrongPhase  StateErrorCode = "wrong-phase"

	ErrorCodeInsufficientCredits StateErrorCode = "insufficient-credits"

	ErrorCodeInsufficientResources StateErrorCode = "insufficient-resources"
	ErrorCodeTooManyResources      StateErrorCode = "too-many-resources"

	ErrorCodeTemperatureTooLow  StateErrorCode = "temperature-too-low"
	ErrorCodeTemperatureTooHigh StateErrorCode = "temperature-too-high"
	ErrorCodeOxygenTooLow       StateErrorCode = "oxygen-too-low"
	ErrorCodeOxygenTooHigh      StateErrorCode = "oxygen-too-high"
	ErrorCodeOceansTooLow       StateErrorCode = "oceans-too-low"
	ErrorCodeOceansTooHigh      StateErrorCode = "oceans-too-high"
	ErrorCodeTRTooLow           StateErrorCode = "tr-too-low"
	ErrorCodeTRTooHigh          StateErrorCode = "tr-too-high"

	ErrorCodeInsufficientTags       StateErrorCode = "insufficient-tags"
	ErrorCodeTooManyTags            StateErrorCode = "too-many-tags"
	ErrorCodeInsufficientProduction StateErrorCode = "insufficient-production"

	ErrorCodeNoOceanTiles         StateErrorCode = "no-ocean-tiles"
	ErrorCodeNoCityPlacements     StateErrorCode = "no-city-placements"
	ErrorCodeNoGreeneryPlacements StateErrorCode = "no-greenery-placements"
	ErrorCodeNoCardsInHand        StateErrorCode = "no-cards-in-hand"
	ErrorCodeInvalidProjectType   StateErrorCode = "invalid-project-type"
	ErrorCodeInvalidRequirement   StateErrorCode = "invalid-requirement"

	ErrorCodeInvalidCardType StateErrorCode = "invalid-card-type"
)

// StateErrorCategory represents categories for error grouping.
// Categories enable UI filtering and display organization.
type StateErrorCategory string

const (
	ErrorCategoryTurn          StateErrorCategory = "turn"
	ErrorCategoryPhase         StateErrorCategory = "phase"
	ErrorCategoryCost          StateErrorCategory = "cost"
	ErrorCategoryInput         StateErrorCategory = "input"
	ErrorCategoryRequirement   StateErrorCategory = "requirement"
	ErrorCategoryAvailability  StateErrorCategory = "availability"
	ErrorCategoryConfiguration StateErrorCategory = "configuration"
	ErrorCategoryInternal      StateErrorCategory = "internal"
)

// StateErrorDto represents a specific reason why an entity (card, action, project) is unavailable
// Part of the Player-Scoped Card Architecture for rich error information
type StateErrorDto struct {
	Code     StateErrorCode     `json:"code" ts:"StateErrorCode"`         // Error code (e.g., ErrorCodeInsufficientCredits)
	Category StateErrorCategory `json:"category" ts:"StateErrorCategory"` // Error category (e.g., ErrorCategoryCost)
	Message  string             `json:"message" ts:"string"`              // Human-readable error message
}

// StateWarningCode represents warning codes for entity state validation.
// All codes use kebab-case for consistency with JSON serialization.
type StateWarningCode string

// StateWarningDto represents a non-blocking warning about an action
// Warnings inform the player of potential issues without preventing the action
type StateWarningDto struct {
	Code    StateWarningCode `json:"code" ts:"StateWarningCode"`
	Message string           `json:"message" ts:"string"`
}

// PlayerCardDto represents a card in a player's hand with calculated playability state
// Part of the Player-Scoped Card Architecture
type PlayerCardDto struct {
	ID              string               `json:"id" ts:"string"`
	Name            string               `json:"name" ts:"string"`
	Type            CardType             `json:"type" ts:"CardType"`
	Cost            int                  `json:"cost" ts:"number"` // Original card cost (same as CardDto.Cost)
	Description     string               `json:"description" ts:"string"`
	Pack            string               `json:"pack" ts:"string"`
	Tags            []CardTag            `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	Requirements    *CardRequirementsDto `json:"requirements,omitempty" ts:"CardRequirementsDto | undefined"`
	Behaviors       []CardBehaviorDto    `json:"behaviors,omitempty" ts:"CardBehaviorDto[] | undefined"`
	ResourceStorage *ResourceStorageDto  `json:"resourceStorage,omitempty" ts:"ResourceStorageDto | undefined"`
	VPConditions    []VPConditionDto     `json:"vpConditions,omitempty" ts:"VPConditionDto[] | undefined"`

	Available     bool              `json:"available" ts:"boolean"`                                      // Computed: len(Errors) == 0
	Errors        []StateErrorDto   `json:"errors" ts:"StateErrorDto[]"`                                 // Single source of truth for availability
	Warnings      []StateWarningDto `json:"warnings,omitempty" ts:"StateWarningDto[] | undefined"`       // Non-blocking warnings
	EffectiveCost int               `json:"effectiveCost" ts:"number"`                                   // Effective cost after discounts (credits)
	Discounts     map[string]int    `json:"discounts,omitempty" ts:"Record<string, number> | undefined"` // Discount amounts per resource type (if any)
}

// PlayerEffectDto represents ongoing effects that a player has active for client consumption
// Aligned with PlayerActionDto structure for consistent behavior handling
type PlayerEffectDto struct {
	CardID        string          `json:"cardId" ts:"string"`            // ID of the card that provides this effect
	CardName      string          `json:"cardName" ts:"string"`          // Name of the card for display purposes
	BehaviorIndex int             `json:"behaviorIndex" ts:"number"`     // Which behavior on the card this effect represents
	Behavior      CardBehaviorDto `json:"behavior" ts:"CardBehaviorDto"` // The actual behavior definition with inputs/outputs
	// Note: No PlayCount since effects are ongoing, not per-generation like actions
}

// PlayerActionDto represents an action that a player can take for client consumption
// Enhanced with calculated usability state from Player-Scoped Card Architecture
type PlayerActionDto struct {
	CardID        string          `json:"cardId" ts:"string"`            // ID of the card that provides this action
	CardName      string          `json:"cardName" ts:"string"`          // Name of the card for display purposes
	BehaviorIndex int             `json:"behaviorIndex" ts:"number"`     // Which behavior on the card this action represents
	Behavior      CardBehaviorDto `json:"behavior" ts:"CardBehaviorDto"` // The actual behavior definition with inputs/outputs

	TimesUsedThisTurn       int `json:"timesUsedThisTurn" ts:"number"`       // Times used this turn
	TimesUsedThisGeneration int `json:"timesUsedThisGeneration" ts:"number"` // Times used this generation

	Available bool            `json:"available" ts:"boolean"`      // Computed: action is usable
	Errors    []StateErrorDto `json:"errors" ts:"StateErrorDto[]"` // Reasons why action is not usable
}

// PlayerStandardProjectDto represents a standard project with availability state
// Part of the Player-Scoped Card Architecture
type PlayerStandardProjectDto struct {
	ProjectType string         `json:"projectType" ts:"string"`              // Standard project type (e.g., "sell_patents", "aquifer")
	BaseCost    map[string]int `json:"baseCost" ts:"Record<string, number>"` // Base cost per resource type (e.g., {"credits": 23} or {"plants": 8})

	Available     bool                   `json:"available" ts:"boolean"`                                      // Computed: project is available
	Errors        []StateErrorDto        `json:"errors" ts:"StateErrorDto[]"`                                 // Reasons why project is not available
	EffectiveCost map[string]int         `json:"effectiveCost" ts:"Record<string, number>"`                   // Cost per resource type after discounts
	Discounts     map[string]int         `json:"discounts,omitempty" ts:"Record<string, number> | undefined"` // Discount amounts per resource type (if any)
	Metadata      map[string]interface{} `json:"metadata,omitempty" ts:"Record<string, any> | undefined"`     // Project-specific context (e.g., oceansRemaining)
}

// ForcedFirstActionDto represents an action that must be completed as the player's first turn action
type ForcedFirstActionDto struct {
	ActionType    string `json:"actionType" ts:"string"`    // Type of action: "city_placement", "card_draw", etc.
	CorporationID string `json:"corporationId" ts:"string"` // Corporation that requires this action
	Completed     bool   `json:"completed" ts:"boolean"`    // Whether the forced action has been completed
	Description   string `json:"description" ts:"string"`   // Human-readable description for UI
}

// PendingTileSelectionDto represents a pending tile placement action for client consumption
type PendingTileSelectionDto struct {
	TileType       string   `json:"tileType" ts:"string"`         // "city", "greenery", "ocean"
	AvailableHexes []string `json:"availableHexes" ts:"string[]"` // Backend-calculated valid hex coordinates
	Source         string   `json:"source" ts:"string"`           // What triggered this selection (card ID, standard project, etc.)
}

// PendingCardSelectionDto represents a pending card selection action (e.g., sell patents, card effects)
type PendingCardSelectionDto struct {
	AvailableCards []CardDto      `json:"availableCards" ts:"CardDto[]"`           // Card IDs player can select from
	CardCosts      map[string]int `json:"cardCosts" ts:"Record<string, number>"`   // Card ID -> cost to select (0 for sell patents, 3 for buying cards)
	CardRewards    map[string]int `json:"cardRewards" ts:"Record<string, number>"` // Card ID -> reward for selecting (1 MC for sell patents)
	Source         string         `json:"source" ts:"string"`                      // What triggered this selection ("sell-patents", card ID, etc.)
	MinCards       int            `json:"minCards" ts:"number"`                    // Minimum cards to select (0 for sell patents)
	MaxCards       int            `json:"maxCards" ts:"number"`                    // Maximum cards to select (hand size for sell patents)
}

// PendingCardDrawSelectionDto represents a pending card draw/peek/take/buy action from card effects
type PendingCardDrawSelectionDto struct {
	AvailableCards []CardDto `json:"availableCards" ts:"CardDto[]"` // Cards shown to player (drawn or peeked)
	FreeTakeCount  int       `json:"freeTakeCount" ts:"number"`     // Number of cards to take for free (mandatory for card-draw, 0 = optional)
	MaxBuyCount    int       `json:"maxBuyCount" ts:"number"`       // Maximum cards to buy (optional, 0 = no buying allowed)
	CardBuyCost    int       `json:"cardBuyCost" ts:"number"`       // Cost per card when buying (typically 3 MC, 0 if no buying)
	Source         string    `json:"source" ts:"string"`            // Card ID or action that triggered this
	PlayAsPrelude  bool      `json:"playAsPrelude" ts:"boolean"`    // When true, selected card is played as prelude
}

// PendingCardDiscardSelectionDto represents a pending card discard action from card effects
type PendingCardDiscardSelectionDto struct {
	MinCards     int    `json:"minCards" ts:"number"`     // 0 if optional (player can skip)
	MaxCards     int    `json:"maxCards" ts:"number"`     // Maximum cards to discard
	Source       string `json:"source" ts:"string"`       // Card name that triggered this
	SourceCardID string `json:"sourceCardId" ts:"string"` // Card ID that triggered this
}

// PendingBehaviorChoiceSelectionDto represents a pending behavior choice from a passive triggered effect
type PendingBehaviorChoiceSelectionDto struct {
	Choices      []ChoiceDto `json:"choices" ts:"ChoiceDto[]"`
	Source       string      `json:"source" ts:"string"`
	SourceCardID string      `json:"sourceCardId" ts:"string"`
}

// PlayerStatus represents the current status of a player in the game
type PlayerStatus string

const (
	PlayerStatusSelectingStartingCards   PlayerStatus = "selecting-starting-cards"
	PlayerStatusSelectingProductionCards PlayerStatus = "selecting-production-cards"
	PlayerStatusWaiting                  PlayerStatus = "waiting"
	PlayerStatusActive                   PlayerStatus = "active"
	PlayerStatusExited                   PlayerStatus = "exited"
)

// PlayerDto represents a player in the game for client consumption
type PlayerDto struct {
	ID               string                     `json:"id" ts:"string"`
	Name             string                     `json:"name" ts:"string"`
	PlayerType       string                     `json:"playerType" ts:"string"`
	BotStatus        string                     `json:"botStatus,omitempty" ts:"string | undefined"`
	BotDifficulty    string                     `json:"botDifficulty,omitempty" ts:"string | undefined"`
	BotSpeed         string                     `json:"botSpeed,omitempty" ts:"string | undefined"`
	Color            string                     `json:"color" ts:"string"`
	Status           PlayerStatus               `json:"status" ts:"PlayerStatus"`
	Corporation      *CardDto                   `json:"corporation" ts:"CardDto | null"`
	Cards            []PlayerCardDto            `json:"cards" ts:"PlayerCardDto[]"` // Hand cards with playability state (Player-Scoped Architecture)
	Resources        ResourcesDto               `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto              `json:"production" ts:"ProductionDto"`
	TerraformRating  int                        `json:"terraformRating" ts:"number"`
	PlayedCards      []CardDto                  `json:"playedCards" ts:"CardDto[]"` // Full card details for all played cards
	Passed           bool                       `json:"passed" ts:"boolean"`
	AvailableActions int                        `json:"availableActions" ts:"number"`
	TotalActions     int                        `json:"totalActions" ts:"number"`
	IsConnected      bool                       `json:"isConnected" ts:"boolean"`
	IsExited         bool                       `json:"isExited" ts:"boolean"`
	Effects          []PlayerEffectDto          `json:"effects" ts:"PlayerEffectDto[]"`                   // Active ongoing effects (discounts, special abilities, etc.)
	Actions          []PlayerActionDto          `json:"actions" ts:"PlayerActionDto[]"`                   // Available actions from played cards with manual triggers
	StandardProjects []PlayerStandardProjectDto `json:"standardProjects" ts:"PlayerStandardProjectDto[]"` // Standard projects with availability state (Player-Scoped Architecture)
	Milestones       []PlayerMilestoneDto       `json:"milestones" ts:"PlayerMilestoneDto[]"`             // Milestones with player eligibility state
	Awards           []PlayerAwardDto           `json:"awards" ts:"PlayerAwardDto[]"`                     // Awards with player eligibility state

	SelectCorporationPhase         *SelectCorporationPhaseDto         `json:"selectCorporationPhase" ts:"SelectCorporationPhaseDto | null"`
	SelectStartingCardsPhase       *SelectStartingCardsPhaseDto       `json:"selectStartingCardsPhase" ts:"SelectStartingCardsPhaseDto | null"`
	SelectPreludeCardsPhase        *SelectPreludeCardsPhaseDto        `json:"selectPreludeCardsPhase" ts:"SelectPreludeCardsPhaseDto | null"`
	ProductionPhase                *ProductionPhaseDto                `json:"productionPhase" ts:"ProductionPhaseDto | null"`
	StartingCards                  []CardDto                          `json:"startingCards" ts:"CardDto[]"`
	PendingTileSelection           *PendingTileSelectionDto           `json:"pendingTileSelection" ts:"PendingTileSelectionDto | null"`
	PendingCardSelection           *PendingCardSelectionDto           `json:"pendingCardSelection" ts:"PendingCardSelectionDto | null"`
	PendingCardDrawSelection       *PendingCardDrawSelectionDto       `json:"pendingCardDrawSelection" ts:"PendingCardDrawSelectionDto | null"`
	PendingCardDiscardSelection    *PendingCardDiscardSelectionDto    `json:"pendingCardDiscardSelection" ts:"PendingCardDiscardSelectionDto | null"`
	PendingBehaviorChoiceSelection *PendingBehaviorChoiceSelectionDto `json:"pendingBehaviorChoiceSelection" ts:"PendingBehaviorChoiceSelectionDto | null"`
	ForcedFirstAction              *ForcedFirstActionDto              `json:"forcedFirstAction" ts:"ForcedFirstActionDto | null"`
	ResourceStorage                map[string]int                     `json:"resourceStorage" ts:"Record<string, number>"`
	PaymentSubstitutes             []PaymentSubstituteDto             `json:"paymentSubstitutes" ts:"PaymentSubstituteDto[]"`
	StoragePaymentSubstitutes      []StoragePaymentSubstituteDto      `json:"storagePaymentSubstitutes" ts:"StoragePaymentSubstituteDto[]"`
	GenerationalEvents             []PlayerGenerationalEventEntryDto  `json:"generationalEvents" ts:"PlayerGenerationalEventEntryDto[]"`
	VPGranters                     []VPGranterDto                     `json:"vpGranters" ts:"VPGranterDto[]"`
	BonusTags                      map[string]int                     `json:"bonusTags" ts:"Record<string, number>"`
}

// OtherPlayerDto represents another player from the viewing player's perspective (limited data)
type OtherPlayerDto struct {
	ID               string            `json:"id" ts:"string"`
	Name             string            `json:"name" ts:"string"`
	PlayerType       string            `json:"playerType" ts:"string"`
	BotStatus        string            `json:"botStatus,omitempty" ts:"string | undefined"`
	BotDifficulty    string            `json:"botDifficulty,omitempty" ts:"string | undefined"`
	BotSpeed         string            `json:"botSpeed,omitempty" ts:"string | undefined"`
	Color            string            `json:"color" ts:"string"`
	Status           PlayerStatus      `json:"status" ts:"PlayerStatus"`
	Corporation      *CardDto          `json:"corporation" ts:"CardDto | null"`
	HandCardCount    int               `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        ResourcesDto      `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto     `json:"production" ts:"ProductionDto"`
	TerraformRating  int               `json:"terraformRating" ts:"number"`
	PlayedCards      []CardDto         `json:"playedCards" ts:"CardDto[]"` // Played cards are public - full card details
	Passed           bool              `json:"passed" ts:"boolean"`
	AvailableActions int               `json:"availableActions" ts:"number"`
	TotalActions     int               `json:"totalActions" ts:"number"`
	IsConnected      bool              `json:"isConnected" ts:"boolean"`
	IsExited         bool              `json:"isExited" ts:"boolean"`
	Effects          []PlayerEffectDto `json:"effects" ts:"PlayerEffectDto[]"`
	Actions          []PlayerActionDto `json:"actions" ts:"PlayerActionDto[]"`

	SelectCorporationPhase    *SelectCorporationOtherPlayerDto   `json:"selectCorporationPhase" ts:"SelectCorporationOtherPlayerDto | null"`
	SelectStartingCardsPhase  *SelectStartingCardsOtherPlayerDto `json:"selectStartingCardsPhase" ts:"SelectStartingCardsOtherPlayerDto | null"`
	SelectPreludeCardsPhase   *SelectPreludeCardsOtherPlayerDto  `json:"selectPreludeCardsPhase" ts:"SelectPreludeCardsOtherPlayerDto | null"`
	ProductionPhase           *ProductionPhaseOtherPlayerDto     `json:"productionPhase" ts:"ProductionPhaseOtherPlayerDto | null"`
	ResourceStorage           map[string]int                     `json:"resourceStorage" ts:"Record<string, number>"`
	PaymentSubstitutes        []PaymentSubstituteDto             `json:"paymentSubstitutes" ts:"PaymentSubstituteDto[]"`
	StoragePaymentSubstitutes []StoragePaymentSubstituteDto      `json:"storagePaymentSubstitutes" ts:"StoragePaymentSubstituteDto[]"`
	VPGranters                []VPGranterDto                     `json:"vpGranters" ts:"VPGranterDto[]"`
	BonusTags                 map[string]int                     `json:"bonusTags" ts:"Record<string, number>"`
}

// GameDto represents a game for client consumption (clean architecture)
type GameDto struct {
	ID                 string                 `json:"id" ts:"string"`
	Status             GameStatus             `json:"status" ts:"GameStatus"`
	Settings           GameSettingsDto        `json:"settings" ts:"GameSettingsDto"`
	HostPlayerID       string                 `json:"hostPlayerId" ts:"string"`
	CurrentPhase       GamePhase              `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters   GlobalParametersDto    `json:"globalParameters" ts:"GlobalParametersDto"`
	CurrentPlayer      PlayerDto              `json:"currentPlayer" ts:"PlayerDto"`       // Viewing player's full data
	OtherPlayers       []OtherPlayerDto       `json:"otherPlayers" ts:"OtherPlayerDto[]"` // Other players' limited data
	ViewingPlayerID    string                 `json:"viewingPlayerId" ts:"string"`        // The player viewing this game state
	CurrentTurn        *string                `json:"currentTurn" ts:"string|null"`       // Whose turn it is (nullable)
	Generation         int                    `json:"generation" ts:"number"`
	PlayerOrder        []string               `json:"playerOrder" ts:"string[]"`                                        // Player IDs in join order
	TurnOrder          []string               `json:"turnOrder" ts:"string[]"`                                          // Turn order of all players in game
	Board              BoardDto               `json:"board" ts:"BoardDto"`                                              // Game board with tiles and occupancy state
	PaymentConstants   PaymentConstantsDto    `json:"paymentConstants" ts:"PaymentConstantsDto"`                        // Conversion rates for alternative payments
	Milestones         []MilestoneDto         `json:"milestones" ts:"MilestoneDto[]"`                                   // All milestones with claim status
	Awards             []AwardDto             `json:"awards" ts:"AwardDto[]"`                                           // All awards with funding status
	AwardResults       []AwardResultDto       `json:"awardResults" ts:"AwardResultDto[]"`                               // Current award placements (1st/2nd place per award)
	FinalScores        []FinalScoreDto        `json:"finalScores,omitempty" ts:"FinalScoreDto[] | undefined"`           // Final scores (only when game completed)
	TriggeredEffects   []TriggeredEffectDto   `json:"triggeredEffects,omitempty" ts:"TriggeredEffectDto[] | undefined"` // Recently triggered passive effects
	PlaceableTileTypes []PlaceableTileTypeDto `json:"placeableTileTypes" ts:"PlaceableTileTypeDto[]"`                   // Available tile types for the demo tile picker
	InitPhase          *InitPhaseDto          `json:"initPhase,omitempty" ts:"InitPhaseDto | undefined"`
	Spectators         []SpectatorDto         `json:"spectators" ts:"SpectatorDto[]"`
	ChatMessages       []ChatMessageDto       `json:"chatMessages" ts:"ChatMessageDto[]"`
	IsSpectator        bool                   `json:"isSpectator" ts:"boolean"`
}

// SpectatorDto represents a spectator visible to all clients.
type SpectatorDto struct {
	ID    string `json:"id" ts:"string"`
	Name  string `json:"name" ts:"string"`
	Color string `json:"color" ts:"string"`
}

// ChatMessageDto represents a chat message sent by a player or spectator.
type ChatMessageDto struct {
	SenderID    string `json:"senderId" ts:"string"`
	SenderName  string `json:"senderName" ts:"string"`
	SenderColor string `json:"senderColor" ts:"string"`
	Message     string `json:"message" ts:"string"`
	Timestamp   string `json:"timestamp" ts:"string"`
	IsSpectator bool   `json:"isSpectator" ts:"boolean"`
}

// PlaceableTileTypeDto represents a tile type available for placement in the demo tile picker
type PlaceableTileTypeDto struct {
	Type  string `json:"type" ts:"string"`
	Label string `json:"label" ts:"string"`
	Group string `json:"group" ts:"string"`
}

// InitPhaseDto represents the state of the init_apply_corp or init_apply_prelude phase
type InitPhaseDto struct {
	CurrentPlayerID    string `json:"currentPlayerId" ts:"string"`
	CurrentPlayerIndex int    `json:"currentPlayerIndex" ts:"number"`
	TotalPlayers       int    `json:"totalPlayers" ts:"number"`
	WaitingForConfirm  bool   `json:"waitingForConfirm" ts:"boolean"`
	ConfirmVersion     int    `json:"confirmVersion" ts:"number"`
	HasPreludePhase    bool   `json:"hasPreludePhase" ts:"boolean"`
	HasPendingTiles    bool   `json:"hasPendingTiles" ts:"boolean"`
}

// Board-related DTOs for tygo generation

// TileBonusDto represents a resource bonus provided by a tile when occupied
type TileBonusDto struct {
	Type   string `json:"type" ts:"string"`
	Amount int    `json:"amount" ts:"number"`
}

// TileOccupantDto represents what currently occupies a tile
type TileOccupantDto struct {
	Type string   `json:"type" ts:"string"`
	Tags []string `json:"tags" ts:"string[]"`
}

// TileDto represents a single hexagonal tile on the game board
type TileDto struct {
	Coordinates HexPositionDto   `json:"coordinates" ts:"HexPositionDto"`
	Tags        []string         `json:"tags" ts:"string[]"`
	Type        string           `json:"type" ts:"string"`
	Location    string           `json:"location" ts:"string"`
	DisplayName *string          `json:"displayName,omitempty" ts:"string|null"`
	Bonuses     []TileBonusDto   `json:"bonuses" ts:"TileBonusDto[]"`
	OccupiedBy  *TileOccupantDto `json:"occupiedBy,omitempty" ts:"TileOccupantDto|null"`
	OwnerID     *string          `json:"ownerId,omitempty" ts:"string|null"`
	ReservedBy  *string          `json:"reservedBy,omitempty" ts:"string|null"`
}

// BoardDto represents the game board containing all tiles
type BoardDto struct {
	Tiles []TileDto `json:"tiles" ts:"TileDto[]"`
}

// MilestoneDto represents a milestone for client consumption
type MilestoneDto struct {
	Type           string         `json:"type" ts:"string"`
	Name           string         `json:"name" ts:"string"`
	Description    string         `json:"description" ts:"string"`
	IsClaimed      bool           `json:"isClaimed" ts:"boolean"`
	ClaimedBy      *string        `json:"claimedBy" ts:"string | null"`
	ClaimCost      int            `json:"claimCost" ts:"number"`
	Required       int            `json:"required" ts:"number"`
	PlayerProgress map[string]int `json:"playerProgress" ts:"Record<string, number>"`
}

// AwardDto represents an award for client consumption
type AwardDto struct {
	Type           string         `json:"type" ts:"string"`
	Name           string         `json:"name" ts:"string"`
	Description    string         `json:"description" ts:"string"`
	IsFunded       bool           `json:"isFunded" ts:"boolean"`
	FundedBy       *string        `json:"fundedBy" ts:"string | null"`
	FundingCost    int            `json:"fundingCost" ts:"number"`
	PlayerProgress map[string]int `json:"playerProgress" ts:"Record<string, number>"`
}

// AwardResultDto represents the placement results for a single funded award
type AwardResultDto struct {
	AwardType      string   `json:"awardType" ts:"string"`
	FirstPlaceIds  []string `json:"firstPlaceIds" ts:"string[]"`
	SecondPlaceIds []string `json:"secondPlaceIds" ts:"string[]"`
}

// PlayerMilestoneDto represents a milestone with player-specific eligibility state
type PlayerMilestoneDto struct {
	Type        string          `json:"type" ts:"string"`
	Name        string          `json:"name" ts:"string"`
	Description string          `json:"description" ts:"string"`
	ClaimCost   int             `json:"claimCost" ts:"number"`
	IsClaimed   bool            `json:"isClaimed" ts:"boolean"`
	ClaimedBy   *string         `json:"claimedBy" ts:"string | null"`
	Available   bool            `json:"available" ts:"boolean"`      // Can this player claim this milestone?
	Progress    int             `json:"progress" ts:"number"`        // Current progress towards requirement
	Required    int             `json:"required" ts:"number"`        // Requirement threshold
	Errors      []StateErrorDto `json:"errors" ts:"StateErrorDto[]"` // Reasons why not available
}

// PlayerAwardDto represents an award with player-specific eligibility state
type PlayerAwardDto struct {
	Type        string          `json:"type" ts:"string"`
	Name        string          `json:"name" ts:"string"`
	Description string          `json:"description" ts:"string"`
	FundingCost int             `json:"fundingCost" ts:"number"` // Current cost to fund (increases as more are funded)
	IsFunded    bool            `json:"isFunded" ts:"boolean"`
	FundedBy    *string         `json:"fundedBy" ts:"string | null"`
	Available   bool            `json:"available" ts:"boolean"`      // Can this player fund this award?
	Errors      []StateErrorDto `json:"errors" ts:"StateErrorDto[]"` // Reasons why not available
}

// VPGranterConditionDto represents a single VP condition's computed breakdown for client consumption
type VPGranterConditionDto struct {
	Amount        int     `json:"amount" ts:"number"`
	ConditionType string  `json:"conditionType" ts:"string"`
	PerType       *string `json:"perType,omitempty" ts:"string | undefined"`
	PerAmount     *int    `json:"perAmount,omitempty" ts:"number | undefined"`
	Count         int     `json:"count" ts:"number"`
	ComputedVP    int     `json:"computedVP" ts:"number"`
	Explanation   string  `json:"explanation" ts:"string"`
}

// VPGranterDto represents a VP source from a played card or corporation for client consumption
type VPGranterDto struct {
	CardID        string                  `json:"cardId" ts:"string"`
	CardName      string                  `json:"cardName" ts:"string"`
	Description   string                  `json:"description" ts:"string"`
	ComputedValue int                     `json:"computedValue" ts:"number"`
	Conditions    []VPGranterConditionDto `json:"conditions" ts:"VPGranterConditionDto[]"`
}

// CardVPConditionDetailDto represents the detailed calculation of a single VP condition
type CardVPConditionDetailDto struct {
	ConditionType  string `json:"conditionType" ts:"string"` // "fixed", "per", "once"
	Amount         int    `json:"amount" ts:"number"`        // VP amount per trigger or fixed amount
	Count          int    `json:"count" ts:"number"`         // Items counted (for "per" conditions)
	MaxTrigger     *int   `json:"maxTrigger" ts:"number | undefined"`
	ActualTriggers int    `json:"actualTriggers" ts:"number"` // Actual triggers after applying max
	TotalVP        int    `json:"totalVP" ts:"number"`        // Final VP from this condition
	Explanation    string `json:"explanation" ts:"string"`    // Human-readable breakdown
}

// CardVPDetailDto represents VP calculation for a single card
type CardVPDetailDto struct {
	CardID     string                     `json:"cardId" ts:"string"`
	CardName   string                     `json:"cardName" ts:"string"`
	Conditions []CardVPConditionDetailDto `json:"conditions" ts:"CardVPConditionDetailDto[]"`
	TotalVP    int                        `json:"totalVP" ts:"number"`
}

// GreeneryVPDetailDto represents VP from a single greenery tile
type GreeneryVPDetailDto struct {
	Coordinate string `json:"coordinate" ts:"string"` // Format: "q,r,s"
	VP         int    `json:"vp" ts:"number"`         // Always 1 per greenery
}

// CityVPDetailDto represents VP from a single city tile and its adjacent greeneries
type CityVPDetailDto struct {
	CityCoordinate     string   `json:"cityCoordinate" ts:"string"`       // Format: "q,r,s"
	AdjacentGreeneries []string `json:"adjacentGreeneries" ts:"string[]"` // Coordinates of adjacent greenery tiles
	VP                 int      `json:"vp" ts:"number"`                   // Number of adjacent greeneries
}

// VPBreakdownDto represents a breakdown of victory points for client consumption
type VPBreakdownDto struct {
	TerraformRating   int                   `json:"terraformRating" ts:"number"`
	CardVP            int                   `json:"cardVP" ts:"number"`
	CardVPDetails     []CardVPDetailDto     `json:"cardVPDetails" ts:"CardVPDetailDto[]"` // Per-card VP breakdown
	MilestoneVP       int                   `json:"milestoneVP" ts:"number"`
	AwardVP           int                   `json:"awardVP" ts:"number"`
	GreeneryVP        int                   `json:"greeneryVP" ts:"number"`
	GreeneryVPDetails []GreeneryVPDetailDto `json:"greeneryVPDetails" ts:"GreeneryVPDetailDto[]"` // Per-greenery VP breakdown
	CityVP            int                   `json:"cityVP" ts:"number"`
	CityVPDetails     []CityVPDetailDto     `json:"cityVPDetails" ts:"CityVPDetailDto[]"` // Per-city VP breakdown with adjacencies
	TotalVP           int                   `json:"totalVP" ts:"number"`
}

// FinalScoreDto represents a player's final score for client consumption
type FinalScoreDto struct {
	PlayerID    string         `json:"playerId" ts:"string"`
	PlayerName  string         `json:"playerName" ts:"string"`
	VPBreakdown VPBreakdownDto `json:"vpBreakdown" ts:"VPBreakdownDto"`
	IsWinner    bool           `json:"isWinner" ts:"boolean"`
	Placement   int            `json:"placement" ts:"number"`
}

// TriggeredEffectDto represents a card effect that was triggered for client notification
type TriggeredEffectDto struct {
	CardName          string                 `json:"cardName" ts:"string"`
	PlayerID          string                 `json:"playerId" ts:"string"`
	SourceType        string                 `json:"sourceType" ts:"string"`
	Outputs           []ResourceConditionDto `json:"outputs" ts:"ResourceConditionDto[]"`
	CalculatedOutputs []CalculatedOutputDto  `json:"calculatedOutputs,omitempty" ts:"CalculatedOutputDto[] | undefined"`
	Behaviors         []CardBehaviorDto      `json:"behaviors,omitempty" ts:"CardBehaviorDto[] | undefined"`
	VPConditions      []VPConditionDto       `json:"vpConditions,omitempty" ts:"VPConditionDto[] | undefined"`
}

// GenerationalEvent represents events tracked within a generation for conditional card behaviors
type GenerationalEvent string

const (
	GenerationalEventTRRaise           GenerationalEvent = "tr-raise"
	GenerationalEventOceanPlacement    GenerationalEvent = "ocean-placement"
	GenerationalEventCityPlacement     GenerationalEvent = "city-placement"
	GenerationalEventGreeneryPlacement GenerationalEvent = "greenery-placement"
)

// PlayerGenerationalEventEntryDto represents a player's tracked generational event for client consumption
type PlayerGenerationalEventEntryDto struct {
	Event GenerationalEvent `json:"event" ts:"GenerationalEvent"`
	Count int               `json:"count" ts:"number"`
}

// GenerationalEventRequirementDto represents a requirement based on generational events for card behaviors
type GenerationalEventRequirementDto struct {
	Event  GenerationalEvent `json:"event" ts:"GenerationalEvent"`
	Count  *MinMaxValueDto   `json:"count,omitempty" ts:"MinMaxValueDto | undefined"`
	Target *TargetType       `json:"target,omitempty" ts:"TargetType | undefined"`
}
