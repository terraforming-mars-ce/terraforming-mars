package player

import (
	"terraforming-mars-backend/internal/events"
)

// PlayerType represents the type of player (human or bot)
type PlayerType string

const (
	PlayerTypeHuman PlayerType = "human"
	PlayerTypeBot   PlayerType = "bot"
)

// BotStatus represents the readiness state of a bot player
type BotStatus string

const (
	BotStatusNone     BotStatus = ""
	BotStatusLoading  BotStatus = "loading"
	BotStatusReady    BotStatus = "ready"
	BotStatusFailed   BotStatus = "failed"
	BotStatusThinking BotStatus = "thinking"
)

// BotDifficulty represents the difficulty level of a bot player
type BotDifficulty string

const (
	BotDifficultyNormal  BotDifficulty = "normal"
	BotDifficultyHard    BotDifficulty = "hard"
	BotDifficultyExtreme BotDifficulty = "extreme"
)

// BotSpeed represents the speed/model tier of a bot player
type BotSpeed string

const (
	BotSpeedFast    BotSpeed = "fast"
	BotSpeedNormal  BotSpeed = "normal"
	BotSpeedThinker BotSpeed = "thinker"
)

// Player represents a player in the game with delegated component management
type Player struct {
	id                 string
	name               string
	gameID             string
	connected          bool
	playerType         PlayerType
	botStatus          BotStatus
	botDifficulty      BotDifficulty
	botSpeed           BotSpeed
	eventBus           *events.EventBusImpl
	corporationID      string
	color              string
	hasPassed          bool
	hasExited          bool
	demoSetupConfirmed bool

	hand               *Hand
	playedCards        *PlayedCards
	resources          *PlayerResources
	selection          *Selection
	actions            *Actions
	effects            *Effects
	generationalEvents *GenerationalEvents
	vpGranters         *VPGranters
}

// NewPlayer creates a new human player with initialized components
// playerID must be provided (generated at handler level for session persistence)
func NewPlayer(eventBus *events.EventBusImpl, gameID, playerID, name string) *Player {
	return &Player{
		id:                 playerID,
		name:               name,
		gameID:             gameID,
		connected:          true,
		playerType:         PlayerTypeHuman,
		eventBus:           eventBus,
		corporationID:      "",
		hasPassed:          false,
		hand:               newHand(eventBus, gameID, playerID),
		playedCards:        newPlayedCards(eventBus, gameID, playerID),
		resources:          newResources(eventBus, gameID, playerID),
		selection:          newSelection(eventBus, gameID, playerID),
		actions:            NewActions(),
		effects:            NewEffects(eventBus),
		generationalEvents: newGenerationalEvents(),
		vpGranters:         NewVPGranters(eventBus, gameID, playerID),
	}
}

// NewBotPlayer creates a new bot player (always connected, type bot)
func NewBotPlayer(eventBus *events.EventBusImpl, gameID, playerID, name string, difficulty BotDifficulty, speed BotSpeed) *Player {
	return &Player{
		id:                 playerID,
		name:               name,
		gameID:             gameID,
		connected:          true,
		playerType:         PlayerTypeBot,
		botStatus:          BotStatusLoading,
		botDifficulty:      difficulty,
		botSpeed:           speed,
		eventBus:           eventBus,
		corporationID:      "",
		hasPassed:          false,
		hand:               newHand(eventBus, gameID, playerID),
		playedCards:        newPlayedCards(eventBus, gameID, playerID),
		resources:          newResources(eventBus, gameID, playerID),
		selection:          newSelection(eventBus, gameID, playerID),
		actions:            NewActions(),
		effects:            NewEffects(eventBus),
		generationalEvents: newGenerationalEvents(),
		vpGranters:         NewVPGranters(eventBus, gameID, playerID),
	}
}

func (p *Player) ID() string {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) GameID() string {
	return p.gameID
}

func (p *Player) PlayerType() PlayerType {
	return p.playerType
}

func (p *Player) IsBot() bool {
	return p.playerType == PlayerTypeBot
}

func (p *Player) BotStatus() BotStatus {
	return p.botStatus
}

func (p *Player) SetBotStatus(status BotStatus) {
	p.botStatus = status
}

func (p *Player) BotDifficulty() BotDifficulty {
	return p.botDifficulty
}

func (p *Player) BotSpeed() BotSpeed {
	return p.botSpeed
}

func (p *Player) SetPlayerType(playerType PlayerType) {
	p.playerType = playerType
}

func (p *Player) SetBotDifficulty(difficulty BotDifficulty) {
	p.botDifficulty = difficulty
}

func (p *Player) SetBotSpeed(speed BotSpeed) {
	p.botSpeed = speed
}

func (p *Player) IsConnected() bool {
	return p.connected
}

func (p *Player) SetConnected(connected bool) {
	p.connected = connected

	if p.eventBus != nil {
	}
}

func (p *Player) CorporationID() string {
	return p.corporationID
}

func (p *Player) SetCorporationID(corporationID string) {
	p.corporationID = corporationID
}

func (p *Player) HasCorporation() bool {
	return p.corporationID != ""
}

func (p *Player) Hand() *Hand {
	return p.hand
}

func (p *Player) PlayedCards() *PlayedCards {
	return p.playedCards
}

func (p *Player) Resources() *PlayerResources {
	return p.resources
}

func (p *Player) Selection() *Selection {
	return p.selection
}

func (p *Player) Actions() *Actions {
	return p.actions
}

func (p *Player) Effects() *Effects {
	return p.effects
}

func (p *Player) GenerationalEvents() *GenerationalEvents {
	return p.generationalEvents
}

func (p *Player) VPGranters() *VPGranters {
	return p.vpGranters
}

func (p *Player) Color() string {
	return p.color
}

func (p *Player) SetColor(color string) {
	p.color = color
}

func (p *Player) HasPassed() bool {
	return p.hasPassed
}

func (p *Player) SetPassed(passed bool) {
	p.hasPassed = passed

	if p.eventBus != nil {
	}
}

func (p *Player) HasExited() bool {
	return p.hasExited
}

func (p *Player) SetExited(exited bool) {
	p.hasExited = exited
}

func (p *Player) DemoSetupConfirmed() bool {
	return p.demoSetupConfirmed
}

func (p *Player) SetDemoSetupConfirmed(confirmed bool) {
	p.demoSetupConfirmed = confirmed
}
