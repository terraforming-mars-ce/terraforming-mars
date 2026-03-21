package websocket

import (
	adminAction "terraforming-mars-backend/internal/action/admin"
	awardAction "terraforming-mars-backend/internal/action/award"
	cardAction "terraforming-mars-backend/internal/action/card"
	colonyAction "terraforming-mars-backend/internal/action/colony"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	connAction "terraforming-mars-backend/internal/action/connection"
	gameAction "terraforming-mars-backend/internal/action/game"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	pfAction "terraforming-mars-backend/internal/action/projectfunding"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/admin"
	"terraforming-mars-backend/internal/delivery/websocket/handler/award"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card"
	colonyHandler "terraforming-mars-backend/internal/delivery/websocket/handler/colony"
	"terraforming-mars-backend/internal/delivery/websocket/handler/confirmation"
	"terraforming-mars-backend/internal/delivery/websocket/handler/connection"
	"terraforming-mars-backend/internal/delivery/websocket/handler/game"
	"terraforming-mars-backend/internal/delivery/websocket/handler/milestone"
	pfHandler "terraforming-mars-backend/internal/delivery/websocket/handler/projectfunding"
	"terraforming-mars-backend/internal/delivery/websocket/handler/resource_conversion"
	"terraforming-mars-backend/internal/delivery/websocket/handler/standard_project"
	"terraforming-mars-backend/internal/delivery/websocket/handler/tile"
	"terraforming-mars-backend/internal/delivery/websocket/handler/turn_management"
	gameModel "terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
)

// RegisterHandlers registers all action handlers with the hub.
func RegisterHandlers(
	hub *core.Hub,
	broadcaster *Broadcaster,
	gameRepo gameModel.GameRepository,
	createGameAction *gameAction.CreateGameAction,
	joinGameAction *gameAction.JoinGameAction,
	addBotAction *gameAction.AddBotAction,
	confirmDemoSetupAction *gameAction.ConfirmDemoSetupAction,
	playCardAction *cardAction.PlayCardAction,
	useCardActionAction *cardAction.UseCardActionAction,
	executeStandardProjectAction *stdprojAction.ExecuteStandardProjectAction,
	convertHeatAction *resconvAction.ConvertHeatToTemperatureAction,
	convertPlantsAction *resconvAction.ConvertPlantsToGreeneryAction,
	selectTileAction *tileAction.SelectTileAction,
	startGameAction *turnAction.StartGameAction,
	skipActionAction *turnAction.SkipActionAction,
	selectStartingChoicesAction *turnAction.SelectStartingChoicesAction,
	confirmInitAdvanceAction *turnAction.ConfirmInitAdvanceAction,
	confirmSellPatentsAction *confirmAction.ConfirmSellPatentsAction,
	confirmProductionCardsAction *confirmAction.ConfirmProductionCardsAction,
	confirmCardDrawAction *confirmAction.ConfirmCardDrawAction,
	confirmCardDiscardAction *confirmAction.ConfirmCardDiscardAction,
	confirmBehaviorChoiceAction *confirmAction.ConfirmBehaviorChoiceAction,
	confirmStealTargetAction *confirmAction.ConfirmStealTargetAction,
	confirmColonyResourceAction *confirmAction.ConfirmColonyResourceAction,
	confirmAwardFundAction *confirmAction.ConfirmAwardFundAction,
	confirmColonyPlacementAction *confirmAction.ConfirmColonyPlacementAction,
	playerDisconnectedAction *connAction.PlayerDisconnectedAction,
	playerTakeoverAction *connAction.PlayerTakeoverAction,
	kickPlayerAction *connAction.KickPlayerAction,
	endGameAction *connAction.EndGameAction,
	setPlayerColorAction *connAction.SetPlayerColorAction,
	spectateGameAction *connAction.SpectateGameAction,
	spectatorDisconnectedAction *connAction.SpectatorDisconnectedAction,
	kickSpectatorAction *connAction.KickSpectatorAction,
	sendChatMessageAction *connAction.SendChatMessageAction,
	convertToBotAction *gameAction.ConvertToBotAction,
	claimMilestoneAction *milestoneAction.ClaimMilestoneAction,
	fundAwardAction *awardAction.FundAwardAction,
	colonyTradeAction *colonyAction.TradeAction,
	colonyBuildAction *colonyAction.BuildColonyAction,
	fundSeatAction *pfAction.FundSeatAction,
	adminSetPhaseAction *adminAction.SetPhaseAction,
	adminSetCurrentTurnAction *adminAction.SetCurrentTurnAction,
	adminSetResourcesAction *adminAction.SetResourcesAction,
	adminSetProductionAction *adminAction.SetProductionAction,
	adminSetGlobalParametersAction *adminAction.SetGlobalParametersAction,
	adminGiveCardAction *adminAction.GiveCardAction,
	adminSetCorporationAction *adminAction.SetCorporationAction,
	adminStartTileSelectionAction *adminAction.StartTileSelectionAction,
	adminSetTRAction *adminAction.SetTRAction,
) {
	log := logger.Get()
	log.Debug("Registering WebSocket handlers")

	createGameHandler := game.NewCreateGameHandler(createGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeCreateGame, createGameHandler)

	joinGameHandler := game.NewJoinGameHandler(joinGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypePlayerConnect, joinGameHandler)
	hub.RegisterHandler(dto.MessageTypeJoinGame, joinGameHandler)

	addBotHandler := game.NewAddBotHandler(addBotAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeAddBot, addBotHandler)

	confirmDemoSetupHandler := game.NewConfirmDemoSetupHandler(confirmDemoSetupAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmDemoSetup, confirmDemoSetupHandler)

	playCardHandler := card.NewPlayCardHandler(playCardAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionPlayCard, playCardHandler)

	useCardActionHandler := card.NewUseCardActionHandler(useCardActionAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionCardAction, useCardActionHandler)

	stdProjHandler := standard_project.NewExecuteHandler(executeStandardProjectAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionStandardProject, stdProjHandler)
	// Legacy message types for backwards compatibility
	hub.RegisterHandler(dto.MessageTypeActionSellPatents, stdProjHandler)
	hub.RegisterHandler(dto.MessageTypeActionLaunchAsteroid, stdProjHandler)
	hub.RegisterHandler(dto.MessageTypeActionBuildPowerPlant, stdProjHandler)
	hub.RegisterHandler(dto.MessageTypeActionBuildAquifer, stdProjHandler)
	hub.RegisterHandler(dto.MessageTypeActionPlantGreenery, stdProjHandler)
	hub.RegisterHandler(dto.MessageTypeActionBuildCity, stdProjHandler)

	convertHeatHandler := resource_conversion.NewConvertHeatHandler(convertHeatAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConvertHeatToTemperature, convertHeatHandler)

	convertPlantsHandler := resource_conversion.NewConvertPlantsHandler(convertPlantsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConvertPlantsToGreenery, convertPlantsHandler)

	selectTileHandler := tile.NewSelectTileHandler(selectTileAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionTileSelected, selectTileHandler)

	startGameHandler := turn_management.NewStartGameHandler(startGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionStartGame, startGameHandler)

	skipActionHandler := turn_management.NewSkipActionHandler(skipActionAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionSkipAction, skipActionHandler)

	selectStartingChoicesHandler := turn_management.NewSelectStartingChoicesHandler(selectStartingChoicesAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionSelectStartingChoices, selectStartingChoicesHandler)

	confirmInitAdvanceHandler := turn_management.NewConfirmInitAdvanceHandler(confirmInitAdvanceAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmInitAdvance, confirmInitAdvanceHandler)

	confirmSellPatentsHandler := confirmation.NewConfirmSellPatentsHandler(confirmSellPatentsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmSellPatents, confirmSellPatentsHandler)

	confirmProductionCardsHandler := confirmation.NewConfirmProductionCardsHandler(confirmProductionCardsAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmProductionCards, confirmProductionCardsHandler)

	confirmCardDrawHandler := confirmation.NewConfirmCardDrawHandler(confirmCardDrawAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionCardDrawConfirmed, confirmCardDrawHandler)

	confirmCardDiscardHandler := confirmation.NewConfirmCardDiscardHandler(confirmCardDiscardAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionCardDiscardConfirmed, confirmCardDiscardHandler)

	confirmBehaviorChoiceHandler := confirmation.NewConfirmBehaviorChoiceHandler(confirmBehaviorChoiceAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionBehaviorChoiceConfirmed, confirmBehaviorChoiceHandler)

	confirmStealTargetHandler := confirmation.NewConfirmStealTargetHandler(confirmStealTargetAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmStealTarget, confirmStealTargetHandler)

	confirmColonyResourceHandler := confirmation.NewConfirmColonyResourceHandler(confirmColonyResourceAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmColonyResource, confirmColonyResourceHandler)

	confirmAwardFundHandler := confirmation.NewConfirmAwardFundHandler(confirmAwardFundAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmAwardFund, confirmAwardFundHandler)

	confirmColonyPlacementHandler := confirmation.NewConfirmColonyPlacementHandler(confirmColonyPlacementAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionConfirmColonyPlacement, confirmColonyPlacementHandler)

	requestLogsHandler := connection.NewRequestLogsHandler(broadcaster)
	hub.RegisterHandler(dto.MessageTypeRequestLogs, requestLogsHandler)

	playerDisconnectedHandler := connection.NewPlayerDisconnectedHandler(playerDisconnectedAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypePlayerDisconnected, playerDisconnectedHandler)

	playerTakeoverHandler := connection.NewPlayerTakeoverHandler(playerTakeoverAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypePlayerTakeover, playerTakeoverHandler)

	kickPlayerHandler := connection.NewKickPlayerHandler(kickPlayerAction, broadcaster, hub)
	hub.RegisterHandler(dto.MessageTypeKickPlayer, kickPlayerHandler)

	endGameHandler := connection.NewEndGameHandler(endGameAction, hub)
	hub.RegisterHandler(dto.MessageTypeEndGame, endGameHandler)

	setPlayerColorHandler := connection.NewSetPlayerColorHandler(setPlayerColorAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeSetPlayerColor, setPlayerColorHandler)

	spectateGameHandler := connection.NewSpectateGameHandler(spectateGameAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeSpectatorConnect, spectateGameHandler)

	spectatorDisconnectedHandler := connection.NewSpectatorDisconnectedHandler(spectatorDisconnectedAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeSpectatorDisconnected, spectatorDisconnectedHandler)

	kickSpectatorHandler := connection.NewKickSpectatorHandler(kickSpectatorAction, broadcaster, hub)
	hub.RegisterHandler(dto.MessageTypeKickSpectator, kickSpectatorHandler)

	chatMessageHandler := connection.NewChatMessageHandler(sendChatMessageAction, broadcaster, gameRepo)
	hub.RegisterHandler(dto.MessageTypeChatMessage, chatMessageHandler)

	convertToBotHandler := game.NewConvertToBotHandler(convertToBotAction, broadcaster, hub)
	hub.RegisterHandler(dto.MessageTypeConvertToBot, convertToBotHandler)

	claimMilestoneHandler := milestone.NewClaimMilestoneHandler(claimMilestoneAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionClaimMilestone, claimMilestoneHandler)

	fundAwardHandler := award.NewFundAwardHandler(fundAwardAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionFundAward, fundAwardHandler)

	colonyTradeHandler := colonyHandler.NewTradeHandler(colonyTradeAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionColonyTrade, colonyTradeHandler)

	colonyBuildHandler := colonyHandler.NewBuildColonyHandler(colonyBuildAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionColonyBuild, colonyBuildHandler)

	fundSeatHandler := pfHandler.NewFundSeatHandler(fundSeatAction, broadcaster)
	hub.RegisterHandler(dto.MessageTypeActionProjectFundingSeat, fundSeatHandler)

	adminCommandHandler := admin.NewAdminCommandHandler(
		adminSetPhaseAction,
		adminSetCurrentTurnAction,
		adminSetResourcesAction,
		adminSetProductionAction,
		adminSetGlobalParametersAction,
		adminGiveCardAction,
		adminSetCorporationAction,
		adminStartTileSelectionAction,
		adminSetTRAction,
		broadcaster,
	)
	hub.RegisterHandler(dto.MessageTypeAdminCommand, adminCommandHandler)

	log.Debug("WebSocket handlers registered")
}
