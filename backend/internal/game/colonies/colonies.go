package colonies

import (
	"slices"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/logger"
)

const maxColoniesPerTile = 3

type Colonies struct {
	ds       *datastore.DataStore
	gameID   string
	eventBus *events.EventBusImpl
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
		logger.Get().Warn("Failed to update game state", zap.String("game_id", c.gameID), zap.Error(err))
	}
}

func (c *Colonies) read(fn func(s *datastore.GameState)) {
	if err := c.ds.ReadGame(c.gameID, fn); err != nil {
		logger.Get().Warn("Failed to read game state", zap.String("game_id", c.gameID), zap.Error(err))
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
		if len(cs.PlayerColonies) >= maxColoniesPerTile {
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
		if !cs.TradedThisGen {
			ids = append(ids, cs.DefinitionID)
		}
	}
	return ids
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

func (c *Colonies) GetTradeFleetAvailable(playerID string) bool {
	var v bool
	c.read(func(s *datastore.GameState) {
		if s.TradeFleets == nil {
			return
		}
		v = s.TradeFleets[playerID]
	})
	return v
}

func (c *Colonies) SetTradeFleetAvailable(playerID string, available bool) {
	c.update(func(s *datastore.GameState) {
		if s.TradeFleets == nil {
			s.TradeFleets = make(map[string]bool)
		}
		s.TradeFleets[playerID] = available
		s.UpdatedAt = time.Now()
	})
}
