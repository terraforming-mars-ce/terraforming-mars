package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"openmars/internal/action/admin"
	"openmars/internal/delivery/dto"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game"
	"openmars/internal/game/shared"
	"openmars/internal/logger"
)

// Broadcaster interface for broadcasting game state
type Broadcaster interface {
	BroadcastGameState(gameID string, playerIDs []string)
}

// AdminCommandHandler handles admin commands. Only players of a game in development mode
// may send them.
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
	restartGameAction         *admin.RestartGameAction
	setActionsRemainingAction *admin.SetActionsRemainingAction
	gameRepo                  game.GameRepository
	broadcaster               Broadcaster
	logger                    *slog.Logger
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
	restartGameAction *admin.RestartGameAction,
	setActionsRemainingAction *admin.SetActionsRemainingAction,
	gameRepo game.GameRepository,
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
		restartGameAction:         restartGameAction,
		setActionsRemainingAction: setActionsRemainingAction,
		gameRepo:                  gameRepo,
		broadcaster:               broadcaster,
		logger:                    logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *AdminCommandHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		slog.String("connection_id", connection.ID),
		slog.String("message_type", string(message.Type)),
	)

	log.Debug("Processing admin command")

	id := connection.Identity()
	if !id.IsPlayer() {
		connection.SendError(message.Type, "Admin commands are only available to players in a game")
		return
	}
	g, err := h.gameRepo.Get(ctx, id.GameID)
	if err != nil {
		connection.SendError(message.Type, err.Error())
		return
	}
	if !g.Settings().DevelopmentMode {
		connection.SendError(message.Type, "Admin commands are only available in development mode")
		return
	}
	gameID := id.GameID

	payloadMap, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.SendError(message.Type, "Invalid payload format")
		return
	}

	commandType, ok := payloadMap["commandType"].(string)
	if !ok {
		log.Error("Missing or invalid commandType")
		connection.SendError(message.Type, "Missing or invalid commandType")
		return
	}

	commandPayload, ok := payloadMap["payload"]
	if !ok {
		log.Error("Missing command payload")
		connection.SendError(message.Type, "Missing command payload")
		return
	}

	log.Debug("Admin command received",
		slog.String("command_type", commandType),
		slog.String("game_id", gameID))

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
	case dto.AdminCommandTypeRestartGame:
		err = h.restartGameAction.Execute(ctx, gameID)
	case dto.AdminCommandTypeSetActionsRemaining:
		err = h.handleSetActionsRemaining(ctx, gameID, commandPayload)
	default:
		log.Error("Unknown admin command type", slog.String("command_type", commandType))
		connection.SendError(message.Type, "Unknown admin command type: "+commandType)
		return
	}

	if err != nil {
		log.Error("Admin command failed", slog.Any("error", err))
		connection.SendError(message.Type, err.Error())
		return
	}

	log.Debug("Admin command executed")

	h.broadcaster.BroadcastGameState(gameID, nil)
	log.Debug("Broadcasted game state after admin command")
}

func (h *AdminCommandHandler) handleGiveCard(ctx context.Context, gameID string, payload any) error {
	var cmd dto.GiveCardAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" || cmd.CardID == "" {
		return errors.New("missing playerId or cardId")
	}
	return h.giveCardAction.Execute(ctx, gameID, cmd.PlayerID, cmd.CardID)
}

func (h *AdminCommandHandler) handleSetPhase(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetPhaseAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	phase := shared.GamePhase(cmd.Phase)
	if !shared.IsGamePhase(phase) {
		return fmt.Errorf("unknown phase %q", cmd.Phase)
	}
	return h.setPhaseAction.Execute(ctx, gameID, phase)
}

func (h *AdminCommandHandler) handleSetResources(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetResourcesAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" {
		return errors.New("missing playerId")
	}
	r := cmd.Resources
	return h.setResourcesAction.Execute(ctx, gameID, cmd.PlayerID, shared.Resources{
		Credits: r.Credits, Steel: r.Steel, Titanium: r.Titanium, Plants: r.Plants, Energy: r.Energy, Heat: r.Heat,
	})
}

func (h *AdminCommandHandler) handleSetProduction(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetProductionAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" {
		return errors.New("missing playerId")
	}
	p := cmd.Production
	return h.setProductionAction.Execute(ctx, gameID, cmd.PlayerID, shared.Production{
		Credits: p.Credits, Steel: p.Steel, Titanium: p.Titanium, Plants: p.Plants, Energy: p.Energy, Heat: p.Heat,
	})
}

func (h *AdminCommandHandler) handleSetGlobalParams(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetGlobalParamsAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	g := cmd.GlobalParameters
	return h.setGlobalParametersAction.Execute(ctx, gameID, admin.SetGlobalParametersRequest{
		Temperature: g.Temperature, Oxygen: g.Oxygen, Oceans: g.Oceans, Venus: g.Venus,
	})
}

func (h *AdminCommandHandler) handleSetCurrentTurn(ctx context.Context, gameID string, payload any) error {
	var cmd struct {
		PlayerID string `json:"playerId"`
	}
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" {
		return errors.New("missing playerId")
	}
	return h.setCurrentTurnAction.Execute(ctx, gameID, cmd.PlayerID)
}

func (h *AdminCommandHandler) handleSetCorporation(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetCorporationAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" || cmd.CorporationID == "" {
		return errors.New("missing playerId or corporationId")
	}
	return h.setCorporationAction.Execute(ctx, gameID, cmd.PlayerID, cmd.CorporationID)
}

func (h *AdminCommandHandler) handleStartTileSelection(ctx context.Context, gameID string, payload any) error {
	var cmd dto.StartTileSelectionAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" || cmd.TileType == "" {
		return errors.New("missing playerId or tileType")
	}
	return h.startTileSelectionAction.Execute(ctx, gameID, cmd.PlayerID, cmd.TileType)
}

func (h *AdminCommandHandler) handleSetTR(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetTRAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	if cmd.PlayerID == "" {
		return errors.New("missing playerId")
	}
	return h.setTRAction.Execute(ctx, gameID, cmd.PlayerID, cmd.TerraformRating)
}

func (h *AdminCommandHandler) handleSetActionsRemaining(ctx context.Context, gameID string, payload any) error {
	var cmd dto.SetActionsRemainingAdminCommand
	if err := decode(payload, &cmd); err != nil {
		return err
	}
	return h.setActionsRemainingAction.Execute(ctx, gameID, cmd.Actions)
}

// decode reads a command payload into its typed form, rejecting values of the wrong
// type instead of reading them as zero.
func decode(payload any, out any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("invalid command payload: %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("invalid command payload: %w", err)
	}
	return nil
}
