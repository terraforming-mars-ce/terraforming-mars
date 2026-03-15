package projectfunding

import (
	"fmt"

	"terraforming-mars-backend/internal/game/projectfunding"
)

// ProjectFundingRegistry provides lookup functionality for project funding definitions
type ProjectFundingRegistry interface {
	GetByID(projectID string) (*projectfunding.ProjectDefinition, error)
	GetAll() []projectfunding.ProjectDefinition
}

// InMemoryProjectFundingRegistry implements ProjectFundingRegistry with an in-memory map
type InMemoryProjectFundingRegistry struct {
	projects map[string]projectfunding.ProjectDefinition
}

// NewInMemoryProjectFundingRegistry creates a new project funding registry from a slice of definitions
func NewInMemoryProjectFundingRegistry(projectList []projectfunding.ProjectDefinition) *InMemoryProjectFundingRegistry {
	projectMap := make(map[string]projectfunding.ProjectDefinition, len(projectList))
	for _, p := range projectList {
		projectMap[p.ID] = p
	}
	return &InMemoryProjectFundingRegistry{projects: projectMap}
}

// GetByID retrieves a project funding definition by ID
func (r *InMemoryProjectFundingRegistry) GetByID(projectID string) (*projectfunding.ProjectDefinition, error) {
	p, exists := r.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}
	return &p, nil
}

// GetAll returns all project funding definitions
func (r *InMemoryProjectFundingRegistry) GetAll() []projectfunding.ProjectDefinition {
	result := make([]projectfunding.ProjectDefinition, 0, len(r.projects))
	for _, p := range r.projects {
		result = append(result, p)
	}
	return result
}
