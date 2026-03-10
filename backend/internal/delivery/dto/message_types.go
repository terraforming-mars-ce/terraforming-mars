package dto

// MessageType represents different types of WebSocket messages
type MessageType string

const (
	MessageTypePlayerConnect MessageType = "player-connect"
	MessageTypeJoinGame      MessageType = "join-game"

	MessageTypeGameUpdated            MessageType = "game-updated"
	MessageTypePlayerConnected        MessageType = "player-connected"
	MessageTypePlayerReconnected      MessageType = "player-reconnected"
	MessageTypePlayerDisconnected     MessageType = "player-disconnected"
	MessageTypeError                  MessageType = "error"
	MessageTypeFullState              MessageType = "full-state"
	MessageTypeProductionPhaseStarted MessageType = "production-phase-started"
	MessageTypeLogUpdate              MessageType = "log-update"

	MessageTypeActionSellPatents        MessageType = "action.standard-project.sell-patents"
	MessageTypeActionConfirmSellPatents MessageType = "action.standard-project.confirm-sell-patents"
	MessageTypeActionLaunchAsteroid     MessageType = "action.standard-project.launch-asteroid"
	MessageTypeActionBuildPowerPlant    MessageType = "action.standard-project.build-power-plant"
	MessageTypeActionBuildAquifer       MessageType = "action.standard-project.build-aquifer"
	MessageTypeActionPlantGreenery      MessageType = "action.standard-project.plant-greenery"
	MessageTypeActionBuildCity          MessageType = "action.standard-project.build-city"

	MessageTypeActionConvertPlantsToGreenery  MessageType = "action.resource-conversion.convert-plants-to-greenery"
	MessageTypeActionConvertHeatToTemperature MessageType = "action.resource-conversion.convert-heat-to-temperature"

	MessageTypeCreateGame               MessageType = "create-game"
	MessageTypeAddBot                   MessageType = "add-bot"
	MessageTypeActionStartGame          MessageType = "action.game-management.start-game"
	MessageTypeActionSkipAction         MessageType = "action.game-management.skip-action"
	MessageTypeActionConfirmDemoSetup   MessageType = "action.game-management.confirm-demo-setup"
	MessageTypeActionConfirmInitAdvance MessageType = "action.game-management.confirm-init-advance"

	MessageTypeActionClaimMilestone MessageType = "action.milestone.claim-milestone"
	MessageTypeActionFundAward      MessageType = "action.award.fund-award"

	MessageTypeActionTileSelected MessageType = "action.tile-selection.tile-selected"

	MessageTypeActionPlayCard                MessageType = "action.card.play-card"
	MessageTypeActionCardAction              MessageType = "action.card.card-action"
	MessageTypeActionSelectStartingChoices   MessageType = "action.card.select-starting-choices"
	MessageTypeActionSelectCards             MessageType = "action.card.select-cards"
	MessageTypeActionConfirmProductionCards  MessageType = "action.card.confirm-production-cards"
	MessageTypeActionCardDrawConfirmed       MessageType = "action.card.card-draw-confirmed"
	MessageTypeActionCardDiscardConfirmed    MessageType = "action.card.card-discard-confirmed"
	MessageTypeActionBehaviorChoiceConfirmed MessageType = "action.card.behavior-choice-confirmed"
	MessageTypeActionConfirmStealTarget      MessageType = "action.card.confirm-steal-target"

	MessageTypeAdminCommand MessageType = "admin-command"

	MessageTypeRequestLogs MessageType = "request-logs"

	MessageTypePlayerTakeover MessageType = "player-takeover"
	MessageTypeKickPlayer     MessageType = "kick-player"
	MessageTypePlayerKicked   MessageType = "player-kicked"
	MessageTypeConvertToBot   MessageType = "convert-to-bot"
	MessageTypeEndGame        MessageType = "end-game"
	MessageTypeGameEnded      MessageType = "game-ended"

	MessageTypeSetPlayerColor        MessageType = "set-player-color"
	MessageTypeSpectatorConnect      MessageType = "spectator-connect"
	MessageTypeSpectatorConnected    MessageType = "spectator-connected"
	MessageTypeSpectatorDisconnected MessageType = "spectator-disconnected"
	MessageTypeChatMessage           MessageType = "chat-message"
	MessageTypeChatUpdate            MessageType = "chat-update"
	MessageTypeKickSpectator         MessageType = "kick-spectator"
	MessageTypeSpectatorKicked       MessageType = "spectator-kicked"
)
