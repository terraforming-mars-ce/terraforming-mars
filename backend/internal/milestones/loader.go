package milestones

import (
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/game/milestone"
)

// LoadMilestonesFromJSON loads milestone definitions from a JSON file
func LoadMilestonesFromJSON(filepath string) ([]milestone.MilestoneDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read milestones file: %w", err)
	}

	var defs []milestone.MilestoneDefinition
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("failed to parse milestones JSON: %w", err)
	}

	if len(defs) == 0 {
		return nil, fmt.Errorf("no milestones found in file: %s", filepath)
	}

	return defs, nil
}
