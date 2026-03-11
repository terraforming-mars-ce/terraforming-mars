package colonies

import (
	"encoding/json"
	"fmt"
	"os"

	"terraforming-mars-backend/internal/game/colony"
)

// LoadColoniesFromJSON loads colony tile definitions from a JSON file
func LoadColoniesFromJSON(filepath string) ([]colony.ColonyTileDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read colony file: %w", err)
	}

	var colonies []colony.ColonyTileDefinition
	if err := json.Unmarshal(data, &colonies); err != nil {
		return nil, fmt.Errorf("failed to parse colony JSON: %w", err)
	}

	if len(colonies) == 0 {
		return nil, fmt.Errorf("no colonies found in file: %s", filepath)
	}

	return colonies, nil
}
