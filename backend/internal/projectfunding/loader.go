package projectfunding

import (
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/game/projectfunding"
)

// LoadProjectsFromJSON loads project funding definitions from a JSON file
func LoadProjectsFromJSON(filepath string) ([]projectfunding.ProjectDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project funding file: %w", err)
	}

	var projects []projectfunding.ProjectDefinition
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse project funding JSON: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found in file: %s", filepath)
	}

	return projects, nil
}
