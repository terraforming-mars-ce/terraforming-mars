package player

import (
	"log/slog"
	"openmars/internal/events"
	"openmars/internal/game/datastore"
	"openmars/internal/game/shared"
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

// Player represents a player in the game.
type Player struct {
	ds       *datastore.DataStore
	gameID   string
	playerID string
	eventBus *events.EventBusImpl

	hand               *Hand
	playedCards        *PlayedCards
	resources          *PlayerResources
	selection          *Selection
	actions            *Actions
	effects            *Effects
	generationalEvents *GenerationalEvents
	vpGranters         *VPGranters
	cardStateStore     *CardStateStore
}

// NewPlayer creates a new human player view backed by the DataStore.
func NewPlayer(ds *datastore.DataStore, gameID string, playerID string, eventBus *events.EventBusImpl) *Player {
	return &Player{
		ds:                 ds,
		gameID:             gameID,
		playerID:           playerID,
		eventBus:           eventBus,
		hand:               newHand(ds, eventBus, gameID, playerID),
		playedCards:        newPlayedCards(ds, eventBus, gameID, playerID),
		resources:          newResources(ds, eventBus, gameID, playerID),
		selection:          newSelection(ds, eventBus, gameID, playerID),
		actions:            NewActions(ds, gameID, playerID),
		effects:            NewEffects(ds, eventBus, gameID, playerID),
		generationalEvents: newGenerationalEvents(ds, gameID, playerID),
		vpGranters:         NewVPGranters(ds, eventBus, gameID, playerID),
		cardStateStore:     NewCardStateStore(),
	}
}

func (p *Player) update(fn func(s *datastore.PlayerState)) {
	if err := p.ds.UpdatePlayer(p.gameID, p.playerID, fn); err != nil {
		slog.Default().Warn("Failed to update player state", slog.String("game_id", p.gameID), slog.String("player_id", p.playerID), slog.Any("error", err))
	}
}

func (p *Player) read(fn func(s *datastore.PlayerState)) {
	if err := p.ds.ReadPlayer(p.gameID, p.playerID, fn); err != nil {
		slog.Default().Warn("Failed to read player state", slog.String("game_id", p.gameID), slog.String("player_id", p.playerID), slog.Any("error", err))
	}
}

func (p *Player) ID() string {
	return p.playerID
}

func (p *Player) Name() string {
	var name string
	p.read(func(s *datastore.PlayerState) {
		name = s.Name
	})
	return name
}

func (p *Player) GameID() string {
	return p.gameID
}

func (p *Player) PlayerType() PlayerType {
	var pt PlayerType
	p.read(func(s *datastore.PlayerState) {
		pt = PlayerType(s.PlayerType)
	})
	return pt
}

func (p *Player) IsBot() bool {
	var isBot bool
	p.read(func(s *datastore.PlayerState) {
		isBot = s.PlayerType == string(PlayerTypeBot)
	})
	return isBot
}

func (p *Player) BotStatus() BotStatus {
	var status BotStatus
	p.read(func(s *datastore.PlayerState) {
		status = BotStatus(s.BotStatus)
	})
	return status
}

func (p *Player) SetBotStatus(status BotStatus) {
	p.update(func(s *datastore.PlayerState) {
		s.BotStatus = string(status)
	})
}

func (p *Player) BotPersona() string {
	var persona string
	p.read(func(s *datastore.PlayerState) {
		persona = s.BotPersona
	})
	return persona
}

func (p *Player) BotError() string {
	var botErr string
	p.read(func(s *datastore.PlayerState) {
		botErr = s.BotError
	})
	return botErr
}

func (p *Player) SetPlayerType(playerType PlayerType) {
	p.update(func(s *datastore.PlayerState) {
		s.PlayerType = string(playerType)
	})
}

func (p *Player) SetBotPersona(persona string) {
	p.update(func(s *datastore.PlayerState) {
		s.BotPersona = persona
	})
}

// SetBotError records why the bot is failing; an empty reason clears it.
func (p *Player) SetBotError(reason string) {
	p.update(func(s *datastore.PlayerState) {
		s.BotError = reason
	})
}

func (p *Player) IsConnected() bool {
	var connected bool
	p.read(func(s *datastore.PlayerState) {
		connected = s.Connected
	})
	return connected
}

func (p *Player) SetConnected(connected bool) {
	p.update(func(s *datastore.PlayerState) {
		s.Connected = connected
	})
}

func (p *Player) CorporationID() string {
	var corpID string
	p.read(func(s *datastore.PlayerState) {
		corpID = s.CorporationID
	})
	return corpID
}

func (p *Player) SetCorporationID(corporationID string) {
	p.update(func(s *datastore.PlayerState) {
		s.CorporationID = corporationID
	})
}

func (p *Player) HasCorporation() bool {
	var has bool
	p.read(func(s *datastore.PlayerState) {
		has = s.CorporationID != ""
	})
	return has
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

func (p *Player) CardStateStore() *CardStateStore {
	return p.cardStateStore
}

func (p *Player) Color() string {
	var color string
	p.read(func(s *datastore.PlayerState) {
		color = s.Color
	})
	return color
}

func (p *Player) SetColor(color string) {
	p.update(func(s *datastore.PlayerState) {
		s.Color = color
	})
}

func (p *Player) HasPassed() bool {
	var passed bool
	p.read(func(s *datastore.PlayerState) {
		passed = s.HasPassed
	})
	return passed
}

func (p *Player) SetPassed(passed bool) {
	p.update(func(s *datastore.PlayerState) {
		s.HasPassed = passed
	})
}

func (p *Player) HasExited() bool {
	var exited bool
	p.read(func(s *datastore.PlayerState) {
		exited = s.HasExited
	})
	return exited
}

func (p *Player) SetExited(exited bool) {
	p.update(func(s *datastore.PlayerState) {
		s.HasExited = exited
	})
}

// PendingDemoChoices returns the player's pending demo lobby selections
func (p *Player) PendingDemoChoices() *shared.PendingDemoChoices {
	var choices *shared.PendingDemoChoices
	p.read(func(s *datastore.PlayerState) {
		choices = s.PendingDemoChoices
	})
	return choices
}

// SetPendingDemoChoices stores the player's demo lobby card selections
func (p *Player) SetPendingDemoChoices(choices *shared.PendingDemoChoices) {
	p.update(func(s *datastore.PlayerState) {
		s.PendingDemoChoices = choices
	})
}

// HasPendingDemoChoices returns true if the player has made demo lobby selections
func (p *Player) HasPendingDemoChoices() bool {
	var has bool
	p.read(func(s *datastore.PlayerState) {
		has = s.PendingDemoChoices != nil
	})
	return has
}

// BonusTags returns the player's bonus tags map (tag type -> count)
func (p *Player) BonusTags() map[shared.CardTag]int {
	var result map[shared.CardTag]int
	p.read(func(s *datastore.PlayerState) {
		result = make(map[shared.CardTag]int, len(s.BonusTags))
		for k, v := range s.BonusTags {
			result[k] = v
		}
	})
	return result
}

// AddBonusTags adds bonus tags of the specified type
func (p *Player) AddBonusTags(tag shared.CardTag, count int) {
	p.update(func(s *datastore.PlayerState) {
		if s.BonusTags == nil {
			s.BonusTags = make(map[shared.CardTag]int)
		}
		s.BonusTags[tag] = s.BonusTags[tag] + count
	})
}

// BonusTagCount returns the number of bonus tags of the specified type
func (p *Player) BonusTagCount(tag shared.CardTag) int {
	var count int
	p.read(func(s *datastore.PlayerState) {
		if s.BonusTags == nil {
			return
		}
		count = s.BonusTags[tag]
	})
	return count
}

// EventBus returns the event bus (for use by actions that need to subscribe)
func (p *Player) EventBus() *events.EventBusImpl {
	return p.eventBus
}
