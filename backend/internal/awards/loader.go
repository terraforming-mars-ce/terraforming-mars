package awards

import (
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/game/award"
)

// LoadAwardsFromJSON loads award definitions from a JSON file
func LoadAwardsFromJSON(filepath string) ([]award.AwardDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read awards file: %w", err)
	}

	var awards []award.AwardDefinition
	if err := json.Unmarshal(data, &awards); err != nil {
		return nil, fmt.Errorf("failed to parse awards JSON: %w", err)
	}

	if len(awards) == 0 {
		return nil, fmt.Errorf("no awards found in file: %s", filepath)
	}

	return awards, nil
}
