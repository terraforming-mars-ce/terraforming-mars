package admin

import (
	"context"
	"encoding/json"

	"terraforming-mars-backend/internal/action/admin"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster interface for broadcasting game state
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// AdminCommandHandler handles admin commands via WebSocket (development mode only)
type AdminCommandHandler struct {
	setPhaseAction            *admin.SetPhaseAction
	setCurrentTurnAction      *admin.SetCurrentTurnAction
	setResourcesAction        *admin.SetResourcesAction
	setProductionAction       *admin.SetProductionAction
	setGlobalParametersAction *admin.SetGlobalParametersAction
	giveCardAction            *admin.GiveCardAction
	setCorporationAction      *admin.SetCorporationAction
	startTileSelectionAction  *admin.StartTileSelectionAction
	setTRAction               *admin.SetTRAction
	broadcaster               Broadcaster
	logger                    *zap.Logger
}

// NewAdminCommandHandler creates a new admin command handler
func NewAdminCommandHandler(
	setPhaseAction *admin.SetPhaseAction,
	setCurrentTurnAction *admin.SetCurrentTurnAction,
	setResourcesAction *admin.SetResourcesAction,
	setProductionAction *admin.SetProductionAction,
	setGlobalParametersAction *admin.SetGlobalParametersAction,
	giveCardAction *admin.GiveCardAction,
	setCorporationAction *admin.SetCorporationAction,
	startTileSelectionAction *admin.StartTileSelectionAction,
	setTRAction *admin.SetTRAction,
	broadcaster Broadcaster,
) *AdminCommandHandler {
	return &AdminCommandHandler{
		setPhaseAction:            setPhaseAction,
		setCurrentTurnAction:      setCurrentTurnAction,
		setResourcesAction:        setResourcesAction,
		setProductionAction:       setProductionAction,
		setGlobalParametersAction: setGlobalParametersAction,
		giveCardAction:            giveCardAction,
		setCorporationAction:      setCorporationAction,
		startTileSelectionAction:  startTileSelectionAction,
		setTRAction:               setTRAction,
		broadcaster:               broadcaster,
		logger:                    logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *AdminCommandHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Debug("Processing admin command")

	_, gameID := connection.GetPlayer()
	if gameID == "" {
		log.Error("No game ID found for connection")
		h.sendError(connection, "Not connected to a game")
		return
	}

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		h.sendError(connection, "Invalid payload format")
		return
	}

	commandType, ok := payloadMap["commandType"].(string)
	if !ok {
		log.Error("Missing or invalid commandType")
		h.sendError(connection, "Missing or invalid commandType")
		return
	}

	commandPayload, ok := payloadMap["payload"]
	if !ok {
		log.Error("Missing command payload")
		h.sendError(connection, "Missing command payload")
		return
	}

	log.Debug("Admin command received",
		zap.String("command_type", commandType),
		zap.String("game_id", gameID))

	var err error
	switch dto.AdminCommandType(commandType) {
	case dto.AdminCommandTypeGiveCard:
		err = h.handleGiveCard(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetPhase:
		err = h.handleSetPhase(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetResources:
		err = h.handleSetResources(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetProduction:
		err = h.handleSetProduction(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetGlobalParams:
		err = h.handleSetGlobalParams(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetCurrentTurn:
		err = h.handleSetCurrentTurn(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetCorporation:
		err = h.handleSetCorporation(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeStartTileSelection:
		err = h.handleStartTileSelection(ctx, gameID, commandPayload)
	case dto.AdminCommandTypeSetTR:
		err = h.handleSetTR(ctx, gameID, commandPayload)
	default:
		log.Error("Unknown admin command type", zap.String("command_type", commandType))
		h.sendError(connection, "Unknown admin command type: "+commandType)
		return
	}

	if err != nil {
		log.Error("Admin command failed", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Debug("Admin command executed")

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("Broadcasted game state after admin command")
}

func (h *AdminCommandHandler) handleGiveCard(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid give-card payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	cardID, _ := payloadMap["cardId"].(string)

	if playerID == "" || cardID == "" {
		return &adminError{message: "Missing playerId or cardId"}
	}

	return h.giveCardAction.Execute(ctx, gameID, playerID, cardID)
}

func (h *AdminCommandHandler) handleSetPhase(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-phase payload"}
	}

	phase, _ := payloadMap["phase"].(string)
	if phase == "" {
		return &adminError{message: "Missing phase"}
	}

	return h.setPhaseAction.Execute(ctx, gameID, game.GamePhase(phase))
}

func (h *AdminCommandHandler) handleSetResources(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-resources payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	if playerID == "" {
		return &adminError{message: "Missing playerId"}
	}

	resourcesData, ok := payloadMap["resources"].(map[string]interface{})
	if !ok {
		return &adminError{message: "Missing or invalid resources"}
	}

	resources := shared.Resources{
		Credits:  getIntFromMap(resourcesData, "credits"),
		Steel:    getIntFromMap(resourcesData, "steel"),
		Titanium: getIntFromMap(resourcesData, "titanium"),
		Plants:   getIntFromMap(resourcesData, "plants"),
		Energy:   getIntFromMap(resourcesData, "energy"),
		Heat:     getIntFromMap(resourcesData, "heat"),
	}

	return h.setResourcesAction.Execute(ctx, gameID, playerID, resources)
}

func (h *AdminCommandHandler) handleSetProduction(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-production payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	if playerID == "" {
		return &adminError{message: "Missing playerId"}
	}

	productionData, ok := payloadMap["production"].(map[string]interface{})
	if !ok {
		return &adminError{message: "Missing or invalid production"}
	}

	production := shared.Production{
		Credits:  getIntFromMap(productionData, "credits"),
		Steel:    getIntFromMap(productionData, "steel"),
		Titanium: getIntFromMap(productionData, "titanium"),
		Plants:   getIntFromMap(productionData, "plants"),
		Energy:   getIntFromMap(productionData, "energy"),
		Heat:     getIntFromMap(productionData, "heat"),
	}

	return h.setProductionAction.Execute(ctx, gameID, playerID, production)
}

func (h *AdminCommandHandler) handleSetGlobalParams(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-global-params payload"}
	}

	globalParamsData, ok := payloadMap["globalParameters"].(map[string]interface{})
	if !ok {
		return &adminError{message: "Missing or invalid globalParameters"}
	}

	params := admin.SetGlobalParametersRequest{
		Temperature: getIntFromMap(globalParamsData, "temperature"),
		Oxygen:      getIntFromMap(globalParamsData, "oxygen"),
		Oceans:      getIntFromMap(globalParamsData, "oceans"),
		Venus:       getIntFromMap(globalParamsData, "venus"),
	}

	return h.setGlobalParametersAction.Execute(ctx, gameID, params)
}

func (h *AdminCommandHandler) handleSetCurrentTurn(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-current-turn payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	if playerID == "" {
		return &adminError{message: "Missing playerId"}
	}

	return h.setCurrentTurnAction.Execute(ctx, gameID, playerID)
}

func (h *AdminCommandHandler) handleSetCorporation(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-corporation payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	corporationID, _ := payloadMap["corporationId"].(string)

	if playerID == "" || corporationID == "" {
		return &adminError{message: "Missing playerId or corporationId"}
	}

	return h.setCorporationAction.Execute(ctx, gameID, playerID, corporationID)
}

func (h *AdminCommandHandler) handleStartTileSelection(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid start-tile-selection payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	tileType, _ := payloadMap["tileType"].(string)

	if playerID == "" || tileType == "" {
		return &adminError{message: "Missing playerId or tileType"}
	}

	return h.startTileSelectionAction.Execute(ctx, gameID, playerID, tileType)
}

func (h *AdminCommandHandler) handleSetTR(ctx context.Context, gameID string, payload interface{}) error {
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		return &adminError{message: "Invalid set-tr payload"}
	}

	playerID, _ := payloadMap["playerId"].(string)
	if playerID == "" {
		return &adminError{message: "Missing playerId"}
	}

	terraformRating := getIntFromMap(payloadMap, "terraformRating")

	return h.setTRAction.Execute(ctx, gameID, playerID, terraformRating)
}

// sendError sends an error message to the client
func (h *AdminCommandHandler) sendError(connection *core.Connection, errorMessage string) {
	_, gameID := connection.GetPlayer()
	connection.Send <- dto.WebSocketMessage{
		Type:   dto.MessageTypeError,
		GameID: gameID,
		Payload: dto.ErrorPayload{
			Message: errorMessage,
		},
	}
}

// adminError is a simple error type for admin command errors
type adminError struct {
	message string
}

func (e *adminError) Error() string {
	return e.message
}

// getIntFromMap safely extracts an int from a map[string]interface{}
func getIntFromMap(m map[string]interface{}, key string) int {
	val, ok := m[key]
	if !ok {
		return 0
	}

	// Handle both float64 (JSON default) and int
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return 0
	}
}
