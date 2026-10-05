package colonies

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/datastore"
)

const maxColoniesPerTile = 3

type Colonies struct {
	definitions []colony.ColonyDefinition
	ds          *datastore.DataStore
	gameID      string
	eventBus    *events.EventBusImpl
}

func NewColonies(ds *datastore.DataStore, gameID string, eventBus *events.EventBusImpl) *Colonies {
	return &Colonies{
		ds:       ds,
		gameID:   gameID,
		eventBus: eventBus,
	}
}

func (c *Colonies) update(fn func(s *datastore.GameState)) {
	if err := c.ds.UpdateGame(c.gameID, fn); err != nil {
		slog.Default().Warn("Failed to update game state", slog.String("game_id", c.gameID), slog.Any("error", err))
	}
}

func (c *Colonies) read(fn func(s *datastore.GameState)) {
	if err := c.ds.ReadGame(c.gameID, fn); err != nil {
		slog.Default().Warn("Failed to read game state", slog.String("game_id", c.gameID), slog.Any("error", err))
	}
}

func (c *Colonies) States() []*colony.ColonyState {
	var result []*colony.ColonyState
	c.read(func(s *datastore.GameState) { result = s.ColonyStates })
	return result
}

func (c *Colonies) SetStates(states []*colony.ColonyState) {
	c.update(func(s *datastore.GameState) {
		s.ColonyStates = states
		s.UpdatedAt = time.Now()
	})
}

func (c *Colonies) GetState(colonyID string) *colony.ColonyState {
	var result *colony.ColonyState
	c.read(func(s *datastore.GameState) {
		for _, state := range s.ColonyStates {
			if state.DefinitionID == colonyID {
				result = state
				return
			}
		}
	})
	return result
}

func (c *Colonies) GetAvailableIDs() []string {
	states := c.States()
	ids := make([]string, 0, len(states))
	for _, cs := range states {
		ids = append(ids, cs.DefinitionID)
	}
	return ids
}

// GetPlaceableIDs returns colony IDs where the player can place a colony
// (not full and player doesn't already have a colony there, unless allowDuplicate is true).
func (c *Colonies) GetPlaceableIDs(playerID string, allowDuplicate bool) []string {
	states := c.States()
	ids := make([]string, 0, len(states))
	for _, cs := range states {
		if cs.AwaitingResource != "" || len(cs.PlayerColonies) >= maxColoniesPerTile {
			continue
		}
		if !allowDuplicate && slices.Contains(cs.PlayerColonies, playerID) {
			continue
		}
		ids = append(ids, cs.DefinitionID)
	}
	return ids
}

// GetTradeableIDs returns colony IDs that haven't been traded this generation.
func (c *Colonies) GetTradeableIDs() []string {
	states := c.States()
	ids := make([]string, 0, len(states))
	for _, cs := range states {
		if cs.AwaitingResource == "" && !cs.TradedThisGen {
			ids = append(ids, cs.DefinitionID)
		}
	}
	return ids
}

// CountPlayerColonies counts how many colonies a specific player has across all colony tiles.
func (c *Colonies) CountPlayerColonies(playerID string) int {
	var total int
	c.read(func(s *datastore.GameState) {
		for _, state := range s.ColonyStates {
			for _, id := range state.PlayerColonies {
				if id == playerID {
					total++
				}
			}
		}
	})
	return total
}

func (c *Colonies) CountAllColonies() int {
	var total int
	c.read(func(s *datastore.GameState) {
		for _, state := range s.ColonyStates {
			total += len(state.PlayerColonies)
		}
	})
	return total
}

// TradeFleet returns a copy of the player's fleet state.
func (c *Colonies) TradeFleet(playerID string) colony.TradeFleet {
	var fleet colony.TradeFleet
	c.read(func(s *datastore.GameState) { fleet = s.TradeFleets[playerID] })
	return fleet
}

// AddTradeFleets increases permanent capacity without resetting used fleets.
func (c *Colonies) AddTradeFleets(playerID string, amount int) {
	if amount <= 0 {
		return
	}
	c.update(func(s *datastore.GameState) {
		if s.TradeFleets == nil {
			s.TradeFleets = make(map[string]colony.TradeFleet)
		}
		fleet := s.TradeFleets[playerID]
		fleet.Capacity += amount
		s.TradeFleets[playerID] = fleet
		s.UpdatedAt = time.Now()
	})
}

// UseTradeFleet consumes one available fleet, rejecting exhaustion.
func (c *Colonies) UseTradeFleet(playerID string) error {
	var result error
	c.update(func(s *datastore.GameState) {
		fleet := s.TradeFleets[playerID]
		if fleet.Available() == 0 {
			result = fmt.Errorf("trade fleet is not available")
			return
		}
		fleet.Used++
		s.TradeFleets[playerID] = fleet
		s.UpdatedAt = time.Now()
	})
	return result
}

// ResetTradeFleets returns all fleets at generation rollover, retaining capacity.
func (c *Colonies) ResetTradeFleets() {
	c.update(func(s *datastore.GameState) {
		for id, fleet := range s.TradeFleets {
			fleet.Used = 0
			s.TradeFleets[id] = fleet
		}
		s.UpdatedAt = time.Now()
	})
}

// MoveTradeMarkers applies a validated group of track movements in one state update.
func (c *Colonies) MoveTradeMarkers(moves map[string]int) {
	c.update(func(s *datastore.GameState) {
		for _, state := range s.ColonyStates {
			state.MarkerPosition += moves[state.DefinitionID]
		}
		s.UpdatedAt = time.Now()
	})
}

// SetDefinitions installs the colony catalog for unused-tile effects.
func (c *Colonies) SetDefinitions(definitions []colony.ColonyDefinition) {
	c.definitions = append([]colony.ColonyDefinition{}, definitions...)
}

// UnusedDefinitions returns colony tiles that have not entered this game.
func (c *Colonies) UnusedDefinitions() []colony.ColonyDefinition {
	result := []colony.ColonyDefinition{}
	for _, def := range c.definitions {
		if c.GetState(def.ID) == nil {
			result = append(result, def)
		}
	}
	return result
}

// AddTile introduces an unowned colony tile, without colony-building rewards.
func (c *Colonies) AddTile(state *colony.ColonyState) error {
	var result error
	c.update(func(s *datastore.GameState) {
		for _, existing := range s.ColonyStates {
			if existing.DefinitionID == state.DefinitionID {
				result = fmt.Errorf("colony tile is already in play")
				return
			}
		}
		s.ColonyStates = append(s.ColonyStates, state)
	})
	if result == nil && c.eventBus != nil {
		events.Publish(c.eventBus, events.GameStateChangedEvent{GameID: c.gameID, Timestamp: time.Now()})
	}
	return result
}

// ActivateResource opens resource-dependent colonies when a matching resource holder enters play.
func (c *Colonies) ActivateResource(resource string) {
	if resource == "" {
		return
	}
	changed := false
	c.update(func(s *datastore.GameState) {
		for _, state := range s.ColonyStates {
			if state.AwaitingResource == resource {
				state.AwaitingResource = ""
				state.MarkerPosition = 1
				changed = true
			}
		}
	})
	if changed && c.eventBus != nil {
		events.Publish(c.eventBus, events.GameStateChangedEvent{GameID: c.gameID, Timestamp: time.Now()})
	}
}
