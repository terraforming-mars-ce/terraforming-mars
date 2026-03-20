package standardprojects

import (
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/game/standardproject"
)

// LoadStandardProjectsFromJSON loads standard project definitions from a JSON file
func LoadStandardProjectsFromJSON(filepath string) ([]standardproject.StandardProjectDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read standard projects file: %w", err)
	}

	var projects []standardproject.StandardProjectDefinition
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse standard projects JSON: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no standard projects found in file: %s", filepath)
	}

	return projects, nil
}
