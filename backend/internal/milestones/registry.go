package milestones

import (
	"fmt"

	"terraforming-mars-backend/internal/game/milestone"
)

// MilestoneRegistry provides lookup functionality for milestone definitions
type MilestoneRegistry interface {
	GetByID(milestoneID string) (*milestone.MilestoneDefinition, error)
	GetAll() []milestone.MilestoneDefinition
}

// InMemoryMilestoneRegistry implements MilestoneRegistry with an in-memory map
type InMemoryMilestoneRegistry struct {
	milestones map[string]milestone.MilestoneDefinition
	order      []string
}

// NewInMemoryMilestoneRegistry creates a new registry from a slice of definitions
func NewInMemoryMilestoneRegistry(defs []milestone.MilestoneDefinition) *InMemoryMilestoneRegistry {
	m := make(map[string]milestone.MilestoneDefinition, len(defs))
	order := make([]string, 0, len(defs))
	for _, d := range defs {
		m[d.ID] = d
		order = append(order, d.ID)
	}
	return &InMemoryMilestoneRegistry{milestones: m, order: order}
}

// GetByID retrieves a milestone definition by ID
func (r *InMemoryMilestoneRegistry) GetByID(milestoneID string) (*milestone.MilestoneDefinition, error) {
	d, exists := r.milestones[milestoneID]
	if !exists {
		return nil, fmt.Errorf("milestone not found: %s", milestoneID)
	}
	return &d, nil
}

// GetAll returns all milestone definitions in their original JSON order
func (r *InMemoryMilestoneRegistry) GetAll() []milestone.MilestoneDefinition {
	result := make([]milestone.MilestoneDefinition, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.milestones[id])
	}
	return result
}
