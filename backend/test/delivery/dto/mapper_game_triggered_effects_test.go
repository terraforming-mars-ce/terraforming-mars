package dto_test

import (
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// TestToGameDto_TriggeredEffectsNotCleared verifies that calling ToGameDto
// multiple times for different players returns the same triggered effects.
// This is critical for multiplayer: the mapper must not mutate game state.
func TestToGameDto_TriggeredEffectsNotCleared(t *testing.T) {
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 3, mockBroadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()

	players := testGame.GetAllPlayers()

	testGame.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:   "Test Card",
		PlayerID:   players[0].ID(),
		SourceType: shared.SourceTypePassiveEffect,
	})

	// Build DTOs for each player sequentially (same order as broadcaster loop)
	for _, p := range players {
		gameDto := dto.ToGameDto(testGame, cardRegistry, p.ID())

		if len(gameDto.TriggeredEffects) != 1 {
			t.Fatalf("player %s: expected 1 triggered effect, got %d", p.ID(), len(gameDto.TriggeredEffects))
		}
		if gameDto.TriggeredEffects[0].CardName != "Test Card" {
			t.Fatalf("player %s: expected card name 'Test Card', got '%s'", p.ID(), gameDto.TriggeredEffects[0].CardName)
		}
	}

	// Effects should still be present on the game until explicitly cleared
	effects := testGame.GetTriggeredEffects()
	if len(effects) != 1 {
		t.Fatalf("expected triggered effects to still be present on game, got %d", len(effects))
	}
}
