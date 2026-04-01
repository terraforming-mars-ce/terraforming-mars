package colonies

import (
	"fmt"

	"terraforming-mars-backend/internal/game/colony"
)

// ColonyRegistry provides lookup functionality for colony tile definitions
type ColonyRegistry interface {
	GetByID(colonyID string) (*colony.ColonyDefinition, error)
	GetAll() []colony.ColonyDefinition
}

// InMemoryColonyRegistry implements ColonyRegistry with an in-memory map
type InMemoryColonyRegistry struct {
	colonies map[string]colony.ColonyDefinition
}

// NewInMemoryColonyRegistry creates a new colony registry from a slice of definitions
func NewInMemoryColonyRegistry(colonyList []colony.ColonyDefinition) *InMemoryColonyRegistry {
	colonyMap := make(map[string]colony.ColonyDefinition, len(colonyList))
	for _, c := range colonyList {
		colonyMap[c.ID] = c
	}
	return &InMemoryColonyRegistry{colonies: colonyMap}
}

// GetByID retrieves a colony tile definition by ID
func (r *InMemoryColonyRegistry) GetByID(colonyID string) (*colony.ColonyDefinition, error) {
	c, exists := r.colonies[colonyID]
	if !exists {
		return nil, fmt.Errorf("colony not found: %s", colonyID)
	}
	return &c, nil
}

// GetAll returns all colony tile definitions
func (r *InMemoryColonyRegistry) GetAll() []colony.ColonyDefinition {
	result := make([]colony.ColonyDefinition, 0, len(r.colonies))
	for _, c := range r.colonies {
		result = append(result, c)
	}
	return result
}
