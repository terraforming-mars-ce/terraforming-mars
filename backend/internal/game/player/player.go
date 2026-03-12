package player

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
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
		logger.Get().Warn("Failed to update player state", zap.String("game_id", p.gameID), zap.String("player_id", p.playerID), zap.Error(err))
	}
}

func (p *Player) read(fn func(s *datastore.PlayerState)) {
	if err := p.ds.ReadPlayer(p.gameID, p.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", p.gameID), zap.String("player_id", p.playerID), zap.Error(err))
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

func (p *Player) BotDifficulty() BotDifficulty {
	var diff BotDifficulty
	p.read(func(s *datastore.PlayerState) {
		diff = BotDifficulty(s.BotDifficulty)
	})
	return diff
}

func (p *Player) BotSpeed() BotSpeed {
	var speed BotSpeed
	p.read(func(s *datastore.PlayerState) {
		speed = BotSpeed(s.BotSpeed)
	})
	return speed
}

func (p *Player) SetPlayerType(playerType PlayerType) {
	p.update(func(s *datastore.PlayerState) {
		s.PlayerType = string(playerType)
	})
}

func (p *Player) SetBotDifficulty(difficulty BotDifficulty) {
	p.update(func(s *datastore.PlayerState) {
		s.BotDifficulty = string(difficulty)
	})
}

func (p *Player) SetBotSpeed(speed BotSpeed) {
	p.update(func(s *datastore.PlayerState) {
		s.BotSpeed = string(speed)
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

func (p *Player) DemoSetupConfirmed() bool {
	var confirmed bool
	p.read(func(s *datastore.PlayerState) {
		confirmed = s.DemoSetupConfirmed
	})
	return confirmed
}

func (p *Player) SetDemoSetupConfirmed(confirmed bool) {
	p.update(func(s *datastore.PlayerState) {
		s.DemoSetupConfirmed = confirmed
	})
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
