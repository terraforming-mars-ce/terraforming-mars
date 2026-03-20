package awards

import (
	"fmt"

	"terraforming-mars-backend/internal/game/award"
)

// AwardRegistry provides lookup functionality for award definitions
type AwardRegistry interface {
	GetByID(awardID string) (*award.AwardDefinition, error)
	GetAll() []award.AwardDefinition
}

// InMemoryAwardRegistry implements AwardRegistry with an in-memory map
type InMemoryAwardRegistry struct {
	awards map[string]award.AwardDefinition
	order  []string
}

// NewInMemoryAwardRegistry creates a new registry from a slice of definitions
func NewInMemoryAwardRegistry(awardList []award.AwardDefinition) *InMemoryAwardRegistry {
	awardMap := make(map[string]award.AwardDefinition, len(awardList))
	order := make([]string, 0, len(awardList))
	for _, a := range awardList {
		awardMap[a.ID] = a
		order = append(order, a.ID)
	}
	return &InMemoryAwardRegistry{awards: awardMap, order: order}
}

// GetByID retrieves an award definition by ID
func (r *InMemoryAwardRegistry) GetByID(awardID string) (*award.AwardDefinition, error) {
	a, exists := r.awards[awardID]
	if !exists {
		return nil, fmt.Errorf("award not found: %s", awardID)
	}
	return &a, nil
}

// GetAll returns all award definitions in their original JSON order
func (r *InMemoryAwardRegistry) GetAll() []award.AwardDefinition {
	result := make([]award.AwardDefinition, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.awards[id])
	}
	return result
}
