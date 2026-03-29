package dto

// ActionType represents different types of game actions
type ActionType string

const (
	ActionTypeSelectStartingCard ActionType = "select-starting-card"
	ActionTypeSelectCards        ActionType = "select-cards"

	ActionTypeStartGame  ActionType = "start-game"
	ActionTypeSkipAction ActionType = "skip-action"
	ActionTypePlayCard   ActionType = "play-card"
	ActionTypeCardAction ActionType = "card-action"

	ActionTypeSellPatents     ActionType = "sell-patents"
	ActionTypeBuildPowerPlant ActionType = "build-power-plant"
	ActionTypeLaunchAsteroid  ActionType = "launch-asteroid"
	ActionTypeBuildAquifer    ActionType = "build-aquifer"
	ActionTypePlantGreenery   ActionType = "plant-greenery"
	ActionTypeBuildCity       ActionType = "build-city"

	ActionTypeConvertPlantsToGreenery  ActionType = "convert-plants-to-greenery"
	ActionTypeConvertHeatToTemperature ActionType = "convert-heat-to-temperature"
)

// SelectStartingCardAction represents selecting starting cards and corporation
type SelectStartingCardAction struct {
	Type          ActionType `json:"type"`
	CardIDs       []string   `json:"cardIds"`
	CorporationID string     `json:"corporationId"`
}

// StartGameAction represents starting the game (host only)
type StartGameAction struct {
	Type ActionType `json:"type"`
}

// SkipAction represents skipping a player's turn
type SkipAction struct {
	Type ActionType `json:"type"`
}

// PlayCardAction represents playing a card from hand
type PlayCardAction struct {
	CardID             string         `json:"cardId"`
	Payment            CardPaymentDto `json:"payment"`                      // Required: payment breakdown (credits, steel, titanium)
	ChoiceIndex        *int           `json:"choiceIndex,omitempty"`        // Optional: index of choice to play (for cards with choices)
	CardStorageTargets []string       `json:"cardStorageTargets,omitempty"` // Optional: target card IDs for resource storage (positional, one per any-card output)
}

// PlayCardActionAction represents playing a card action from player's action list
type PlayCardActionAction struct {
	CardID             string   `json:"cardId"`
	BehaviorIndex      int      `json:"behaviorIndex"`
	ChoiceIndex        *int     `json:"choiceIndex,omitempty"`        // Optional: index of choice to play (for actions with choices)
	CardStorageTargets []string `json:"cardStorageTargets,omitempty"` // Optional: target card IDs for resource storage (positional, one per any-card output)
}

// HexPositionDto represents a position on the Mars board
type HexPositionDto struct {
	Q int `json:"q"`
	R int `json:"r"`
	S int `json:"s"`
}

// Standard Project Actions

// SellPatentsAction represents selling patent cards for megacredits (initiates card selection)
type SellPatentsAction struct {
	Type ActionType `json:"type"`
}

// BuildPowerPlantAction represents building a power plant
type BuildPowerPlantAction struct {
	Type ActionType `json:"type"`
}

// LaunchAsteroidAction represents launching an asteroid
type LaunchAsteroidAction struct {
	Type ActionType `json:"type"`
}

// BuildAquiferAction represents building an aquifer
type BuildAquiferAction struct {
	Type        ActionType     `json:"type"`
	HexPosition HexPositionDto `json:"hexPosition"`
}

// PlantGreeneryAction represents planting greenery
type PlantGreeneryAction struct {
	Type        ActionType     `json:"type"`
	HexPosition HexPositionDto `json:"hexPosition"`
}

// BuildCityAction represents building a city
type BuildCityAction struct {
	Type        ActionType     `json:"type"`
	HexPosition HexPositionDto `json:"hexPosition"`
}

// ActionSelectStartingCardRequest contains the action data for select starting card actions
type ActionSelectStartingCardRequest struct {
	Type          ActionType `json:"type"`
	CardIDs       []string   `json:"cardIds"`
	CorporationID string     `json:"corporationId"` // Corporation selected alongside starting cards
}

// ActionSelectProductionCardsRequest contains the action data for select production card actions
type ActionSelectProductionCardsRequest struct {
	Type    ActionType `json:"type"`
	CardIDs []string   `json:"cardIds"`
}

// GetAction returns the select starting card action
func (ap *ActionSelectStartingCardRequest) GetAction() *SelectStartingCardAction {
	return &SelectStartingCardAction{Type: ap.Type, CardIDs: ap.CardIDs, CorporationID: ap.CorporationID}
}

// ActionStartGameRequest contains the action data for start game actions
type ActionStartGameRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the start game action
func (ap *ActionStartGameRequest) GetAction() *StartGameAction {
	return &StartGameAction{Type: ap.Type}
}

// ActionSkipActionRequest contains the action data for skip action actions
type ActionSkipActionRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the skip action action
func (ap *ActionSkipActionRequest) GetAction() *SkipAction {
	return &SkipAction{Type: ap.Type}
}

// ConfirmDemoSetupRequest contains the player's demo setup configuration
type ConfirmDemoSetupRequest struct {
	CorporationID    *string              `json:"corporationId,omitempty"`
	CardIDs          []string             `json:"cardIds"`
	Resources        ResourcesDto         `json:"resources"`
	Production       ProductionDto        `json:"production"`
	TerraformRating  int                  `json:"terraformRating"`
	GlobalParameters *GlobalParametersDto `json:"globalParameters,omitempty"` // Host only
	Generation       *int                 `json:"generation,omitempty"`       // Host only
}

// ActionPlayCardRequest contains the action data for play card actions
type ActionPlayCardRequest struct {
	Type               ActionType     `json:"type"`
	CardID             string         `json:"cardId"`
	Payment            CardPaymentDto `json:"payment"`                      // Required: payment breakdown (credits, steel, titanium)
	ChoiceIndex        *int           `json:"choiceIndex,omitempty"`        // Optional: index of choice to play (for cards with choices)
	CardStorageTargets []string       `json:"cardStorageTargets,omitempty"` // Optional: target card IDs for resource storage (positional, one per any-card output)
}

// GetAction returns the play card action
func (ap *ActionPlayCardRequest) GetAction() *PlayCardAction {
	return &PlayCardAction{CardID: ap.CardID, Payment: ap.Payment, ChoiceIndex: ap.ChoiceIndex, CardStorageTargets: ap.CardStorageTargets}
}

// ActionPlayCardActionRequest contains the action data for play card action actions
type ActionPlayCardActionRequest struct {
	Type               ActionType `json:"type"`
	CardID             string     `json:"cardId"`
	BehaviorIndex      int        `json:"behaviorIndex"`
	ChoiceIndex        *int       `json:"choiceIndex,omitempty"`        // Optional: index of choice to play (for actions with choices)
	CardStorageTargets []string   `json:"cardStorageTargets,omitempty"` // Optional: target card IDs for resource storage (positional, one per any-card output)
}

// GetAction returns the play card action action
func (ap *ActionPlayCardActionRequest) GetAction() *PlayCardActionAction {
	return &PlayCardActionAction{CardID: ap.CardID, BehaviorIndex: ap.BehaviorIndex, ChoiceIndex: ap.ChoiceIndex, CardStorageTargets: ap.CardStorageTargets}
}

// Standard Project Action Requests

// ActionSellPatentsRequest contains the action data for sell patents actions (initiates card selection)
type ActionSellPatentsRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the sell patents action
func (ap *ActionSellPatentsRequest) GetAction() *SellPatentsAction {
	return &SellPatentsAction{Type: ap.Type}
}

// ActionBuildPowerPlantRequest contains the action data for build power plant actions
type ActionBuildPowerPlantRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the build power plant action
func (ap *ActionBuildPowerPlantRequest) GetAction() *BuildPowerPlantAction {
	return &BuildPowerPlantAction{Type: ap.Type}
}

// ActionLaunchAsteroidRequest contains the action data for launch asteroid actions
type ActionLaunchAsteroidRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the launch asteroid action
func (ap *ActionLaunchAsteroidRequest) GetAction() *LaunchAsteroidAction {
	return &LaunchAsteroidAction{Type: ap.Type}
}

// ActionBuildAquiferRequest contains the action data for build aquifer actions
type ActionBuildAquiferRequest struct {
	Type        ActionType     `json:"type"`
	HexPosition HexPositionDto `json:"hexPosition"`
}

// GetAction returns the build aquifer action
func (ap *ActionBuildAquiferRequest) GetAction() *BuildAquiferAction {
	return &BuildAquiferAction{Type: ap.Type, HexPosition: ap.HexPosition}
}

// ActionPlantGreeneryRequest contains the action data for plant greenery actions
type ActionPlantGreeneryRequest struct {
	Type        ActionType     `json:"type"`
	HexPosition HexPositionDto `json:"hexPosition"`
}

// GetAction returns the plant greenery action
func (ap *ActionPlantGreeneryRequest) GetAction() *PlantGreeneryAction {
	return &PlantGreeneryAction{Type: ap.Type, HexPosition: ap.HexPosition}
}

// ActionBuildCityRequest contains the action data for build city actions
type ActionBuildCityRequest struct {
	Type        ActionType     `json:"type"`
	HexPosition HexPositionDto `json:"hexPosition"`
}

// GetAction returns the build city action
func (ap *ActionBuildCityRequest) GetAction() *BuildCityAction {
	return &BuildCityAction{Type: ap.Type, HexPosition: ap.HexPosition}
}

// ActionConvertPlantsToGreeneryRequest contains the action data for initiating plant conversion
type ActionConvertPlantsToGreeneryRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the convert plants to greenery action
func (ap *ActionConvertPlantsToGreeneryRequest) GetAction() *ConvertPlantsToGreeneryAction {
	return &ConvertPlantsToGreeneryAction{Type: ap.Type}
}

// ActionConvertHeatToTemperatureRequest contains the action data for converting heat to temperature
type ActionConvertHeatToTemperatureRequest struct {
	Type ActionType `json:"type"`
}

// GetAction returns the convert heat to temperature action
func (ap *ActionConvertHeatToTemperatureRequest) GetAction() *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{Type: ap.Type}
}

// ConvertPlantsToGreeneryAction represents converting 8 plants to a greenery tile
type ConvertPlantsToGreeneryAction struct {
	Type ActionType `json:"type"`
}

// ConvertHeatToTemperatureAction represents converting 8 heat to raise temperature
type ConvertHeatToTemperatureAction struct {
	Type ActionType `json:"type"`
}

// Admin Command Types (Development Mode Only)

// AdminCommandType represents different types of admin commands
type AdminCommandType string

const (
	AdminCommandTypeGiveCard           AdminCommandType = "give-card"
	AdminCommandTypeSetPhase           AdminCommandType = "set-phase"
	AdminCommandTypeSetResources       AdminCommandType = "set-resources"
	AdminCommandTypeSetProduction      AdminCommandType = "set-production"
	AdminCommandTypeSetGlobalParams    AdminCommandType = "set-global-params"
	AdminCommandTypeStartTileSelection AdminCommandType = "start-tile-selection"
	AdminCommandTypeSetCurrentTurn     AdminCommandType = "set-current-turn"
	AdminCommandTypeSetCorporation     AdminCommandType = "set-corporation"
	AdminCommandTypeSetTR              AdminCommandType = "set-tr"
)

// AdminCommandRequest contains the admin command data
type AdminCommandRequest struct {
	CommandType AdminCommandType `json:"commandType"`
	Payload     interface{}      `json:"payload"`
}

// GiveCardAdminCommand represents giving a card to a player
type GiveCardAdminCommand struct {
	PlayerID string `json:"playerId"`
	CardID   string `json:"cardId"`
}

// SetPhaseAdminCommand represents setting the game phase
type SetPhaseAdminCommand struct {
	Phase string `json:"phase"`
}

// SetResourcesAdminCommand represents setting a player's resources
type SetResourcesAdminCommand struct {
	PlayerID  string       `json:"playerId"`
	Resources ResourcesDto `json:"resources"`
}

// SetProductionAdminCommand represents setting a player's production
type SetProductionAdminCommand struct {
	PlayerID   string        `json:"playerId"`
	Production ProductionDto `json:"production"`
}

// SetGlobalParamsAdminCommand represents setting global parameters
type SetGlobalParamsAdminCommand struct {
	GlobalParameters GlobalParametersDto `json:"globalParameters"`
}

// StartTileSelectionAdminCommand represents starting tile selection for testing
type StartTileSelectionAdminCommand struct {
	PlayerID string `json:"playerId"`
	TileType string `json:"tileType"`
}

// SetCorporationAdminCommand represents setting a player's corporation
type SetCorporationAdminCommand struct {
	PlayerID      string `json:"playerId"`
	CorporationID string `json:"corporationId"`
}

// SetTRAdminCommand represents setting a player's terraform rating
type SetTRAdminCommand struct {
	PlayerID        string `json:"playerId"`
	TerraformRating int    `json:"terraformRating"`
}

// CardPaymentDto represents how a player is paying for a card
type CardPaymentDto struct {
	Credits            int            `json:"credits"`                      // MC spent
	Steel              int            `json:"steel"`                        // Steel resources used (2 MC value each)
	Titanium           int            `json:"titanium"`                     // Titanium resources used (3 MC value each)
	Substitutes        map[string]int `json:"substitutes,omitempty"`        // Payment substitutes (e.g., heat for Helion)
	StorageSubstitutes map[string]int `json:"storageSubstitutes,omitempty"` // Storage payment substitutes (e.g., floaters from Dirigibles)
}
