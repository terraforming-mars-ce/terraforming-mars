package standardprojects

import (
	"fmt"

	"terraforming-mars-backend/internal/game/standardproject"
)

// StandardProjectRegistry provides lookup functionality for standard project definitions
type StandardProjectRegistry interface {
	GetByID(projectID string) (*standardproject.StandardProjectDefinition, error)
	GetAll() []standardproject.StandardProjectDefinition
}

// InMemoryStandardProjectRegistry implements StandardProjectRegistry with an in-memory map
type InMemoryStandardProjectRegistry struct {
	projects map[string]standardproject.StandardProjectDefinition
	order    []string
}

// NewInMemoryStandardProjectRegistry creates a new registry from a slice of definitions
func NewInMemoryStandardProjectRegistry(projectList []standardproject.StandardProjectDefinition) *InMemoryStandardProjectRegistry {
	projectMap := make(map[string]standardproject.StandardProjectDefinition, len(projectList))
	order := make([]string, 0, len(projectList))
	for _, p := range projectList {
		projectMap[p.ID] = p
		order = append(order, p.ID)
	}
	return &InMemoryStandardProjectRegistry{projects: projectMap, order: order}
}

// GetByID retrieves a standard project definition by ID
func (r *InMemoryStandardProjectRegistry) GetByID(projectID string) (*standardproject.StandardProjectDefinition, error) {
	p, exists := r.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("standard project not found: %s", projectID)
	}
	return &p, nil
}

// GetAll returns all standard project definitions in their original JSON order
func (r *InMemoryStandardProjectRegistry) GetAll() []standardproject.StandardProjectDefinition {
	result := make([]standardproject.StandardProjectDefinition, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.projects[id])
	}
	return result
}
