package player

import (
	"terraforming-mars-backend/internal/events"
)

// Player represents a player in the game with delegated component management
type Player struct {
	id                 string
	name               string
	gameID             string
	connected          bool
	eventBus           *events.EventBusImpl
	corporationID      string
	color              string
	hasPassed          bool
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

// NewPlayer creates a new player with initialized components
// playerID must be provided (generated at handler level for session persistence)
func NewPlayer(eventBus *events.EventBusImpl, gameID, playerID, name string) *Player {
	return &Player{
		id:                 playerID,
		name:               name,
		gameID:             gameID,
		connected:          true,
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

func (p *Player) DemoSetupConfirmed() bool {
	return p.demoSetupConfirmed
}

func (p *Player) SetDemoSetupConfirmed(confirmed bool) {
	p.demoSetupConfirmed = confirmed
}
