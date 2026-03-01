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

	MessageTypeCreateGame             MessageType = "create-game"
	MessageTypeActionStartGame        MessageType = "action.game-management.start-game"
	MessageTypeActionSkipAction       MessageType = "action.game-management.skip-action"
	MessageTypeActionConfirmDemoSetup MessageType = "action.game-management.confirm-demo-setup"

	MessageTypeActionClaimMilestone MessageType = "action.milestone.claim-milestone"
	MessageTypeActionFundAward      MessageType = "action.award.fund-award"

	MessageTypeActionTileSelected MessageType = "action.tile-selection.tile-selected"

	MessageTypeActionPlayCard                MessageType = "action.card.play-card"
	MessageTypeActionCardAction              MessageType = "action.card.card-action"
	MessageTypeActionSelectCorporation       MessageType = "action.card.select-corporation"
	MessageTypeActionSelectStartingCard      MessageType = "action.card.select-starting-card"
	MessageTypeActionSelectPreludeCards      MessageType = "action.card.select-prelude-cards"
	MessageTypeActionSelectCards             MessageType = "action.card.select-cards"
	MessageTypeActionConfirmProductionCards  MessageType = "action.card.confirm-production-cards"
	MessageTypeActionCardDrawConfirmed       MessageType = "action.card.card-draw-confirmed"
	MessageTypeActionCardDiscardConfirmed    MessageType = "action.card.card-discard-confirmed"
	MessageTypeActionBehaviorChoiceConfirmed MessageType = "action.card.behavior-choice-confirmed"

	MessageTypeAdminCommand MessageType = "admin-command"

	MessageTypeRequestLogs MessageType = "request-logs"

	MessageTypePlayerTakeover MessageType = "player-takeover"
	MessageTypeKickPlayer     MessageType = "kick-player"
	MessageTypePlayerKicked   MessageType = "player-kicked"
)
