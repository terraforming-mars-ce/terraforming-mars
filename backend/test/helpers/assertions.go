package helpers

import (
	"testing"

	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"

	"github.com/stretchr/testify/assert"
)

// AssertResourcesEqual checks if actual resources match expected resources
func AssertResourcesEqual(t *testing.T, expected, actual shared.Resources, msgAndArgs ...interface{}) {
	assert.Equal(t, expected.Credits, actual.Credits, "credits mismatch")
	assert.Equal(t, expected.Steel, actual.Steel, "steel mismatch")
	assert.Equal(t, expected.Titanium, actual.Titanium, "titanium mismatch")
	assert.Equal(t, expected.Plants, actual.Plants, "plants mismatch")
	assert.Equal(t, expected.Energy, actual.Energy, "energy mismatch")
	assert.Equal(t, expected.Heat, actual.Heat, "heat mismatch")
}

// AssertProductionEqual checks if actual production matches expected production
func AssertProductionEqual(t *testing.T, expected, actual shared.Production, msgAndArgs ...interface{}) {
	assert.Equal(t, expected.Credits, actual.Credits, "credits production mismatch")
	assert.Equal(t, expected.Steel, actual.Steel, "steel production mismatch")
	assert.Equal(t, expected.Titanium, actual.Titanium, "titanium production mismatch")
	assert.Equal(t, expected.Plants, actual.Plants, "plants production mismatch")
	assert.Equal(t, expected.Energy, actual.Energy, "energy production mismatch")
	assert.Equal(t, expected.Heat, actual.Heat, "heat production mismatch")
}

// AssertGlobalParametersEqual checks if actual global parameters match expected
func AssertGlobalParametersEqual(t *testing.T, expected, actual *global_parameters.GlobalParameters) {
	assert.Equal(t, expected.Temperature(), actual.Temperature(), "temperature mismatch")
	assert.Equal(t, expected.Oxygen(), actual.Oxygen(), "oxygen mismatch")
	assert.Equal(t, expected.Oceans(), actual.Oceans(), "oceans mismatch")
}

// AssertGamePhase checks if game is in expected phase
func AssertGamePhase(t *testing.T, expected shared.GamePhase, actual shared.GamePhase) {
	assert.Equal(t, expected, actual, "game phase mismatch")
}

// AssertGameStatus checks if game is in expected status
func AssertGameStatus(t *testing.T, expected shared.GameStatus, actual shared.GameStatus) {
	assert.Equal(t, expected, actual, "game status mismatch")
}

// AssertPlayerHasCard checks if player has a specific card in their played cards
func AssertPlayerHasCard(t *testing.T, p *player.Player, cardID string) {
	playedCards := p.PlayedCards().Cards()
	for _, playedCardID := range playedCards {
		if playedCardID == cardID {
			return
		}
	}
	t.Errorf("Player %s does not have card %s in played cards", p.ID(), cardID)
}

// AssertPlayerDoesNotHaveCard checks if player does NOT have a specific card
func AssertPlayerDoesNotHaveCard(t *testing.T, p *player.Player, cardID string) {
	playedCards := p.PlayedCards().Cards()
	for _, playedCardID := range playedCards {
		if playedCardID == cardID {
			t.Errorf("Player %s should not have card %s in played cards", p.ID(), cardID)
			return
		}
	}
}

// AssertPlayerTags checks if player has expected tag counts
func AssertPlayerTags(t *testing.T, expected map[shared.CardTag]int, actual map[shared.CardTag]int) {
	for tag, expectedCount := range expected {
		actualCount, ok := actual[tag]
		if !ok {
			t.Errorf("Player missing tag %v", tag)
			continue
		}
		assert.Equal(t, expectedCount, actualCount, "tag %v count mismatch", tag)
	}
}

// AssertResourceChange checks if resource changed by expected amount
func AssertResourceChange(t *testing.T, resourceName string, expectedChange int, before, after int) {
	actualChange := after - before
	assert.Equal(t, expectedChange, actualChange, "%s should have changed by %d, but changed by %d", resourceName, expectedChange, actualChange)
}

// AssertCreditsChange checks if credits changed by expected amount
func AssertCreditsChange(t *testing.T, expectedChange int, beforeResources, afterResources shared.Resources) {
	AssertResourceChange(t, "credits", expectedChange, beforeResources.Credits, afterResources.Credits)
}

// AssertTerraformRatingChange checks if TR changed by expected amount
func AssertTerraformRatingChange(t *testing.T, expectedChange int, before, after int) {
	actualChange := after - before
	assert.Equal(t, expectedChange, actualChange, "terraform rating should have changed by %d, but changed by %d", expectedChange, actualChange)
}

// AssertGameStateQuality performs comprehensive validation of game state structure
func AssertGameStateQuality(t *testing.T, gameState map[string]interface{}, expectedPlayerCount int) {
	if gameState == nil {
		t.Fatalf("Game state should not be nil")
		return
	}

	// Check game status
	status, ok := gameState["status"].(string)
	if !ok || status == "" {
		t.Fatalf("Game should have non-empty status field")
	}

	// Check game phase
	phase, ok := gameState["phase"].(string)
	if !ok || phase == "" {
		t.Fatalf("Game should have non-empty phase field")
	}

	// Check player count
	actualPlayerCount := CountPlayersInGameState(t, gameState)
	assert.Equal(t, expectedPlayerCount, actualPlayerCount, "player count mismatch")

	// Check global parameters
	globalParams := ExtractGlobalParameters(gameState)
	if globalParams == nil {
		t.Fatalf("Game should have globalParameters")
	}

	// Verify global parameters have required fields
	if _, ok := globalParams["temperature"]; !ok {
		t.Errorf("Global parameters missing temperature")
	}
	if _, ok := globalParams["oxygen"]; !ok {
		t.Errorf("Global parameters missing oxygen")
	}
	if _, ok := globalParams["oceans"]; !ok {
		t.Errorf("Global parameters missing oceans")
	}
}
