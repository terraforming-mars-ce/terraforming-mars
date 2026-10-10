package player_test

import (
	"testing"

	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

func TestAddProduction_CreditProductionAllowsNegative(t *testing.T) {
	// Setup: Create game with a player
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1)
	players := testGame.GetAllPlayers()
	player := players[0]

	// Give player 2 MC production
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 2,
	})
	testutil.AssertEqual(t, 2, player.Resources().Production().Credits, "Should have 2 MC production")

	// Reduce MC production by 5 (should result in -3)
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: -5,
	})
	testutil.AssertEqual(t, -3, player.Resources().Production().Credits, "MC production should be -3 (allowed negative)")
}

func TestAddProduction_CreditProductionClampsToMinusFive(t *testing.T) {
	// Setup: Create game with a player
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1)
	players := testGame.GetAllPlayers()
	player := players[0]

	// Give player 0 MC production (starting state)
	production := player.Resources().Production()
	testutil.AssertEqual(t, 0, production.Credits, "Should start with 0 MC production")

	// Reduce MC production by 10 (should clamp to -5)
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: -10,
	})
	testutil.AssertEqual(t, shared.MinCreditProduction, player.Resources().Production().Credits, "MC production should be clamped to -5")
}

func TestAddProduction_OtherProductionClampsToZero(t *testing.T) {
	// Setup: Create game with a player
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1)
	players := testGame.GetAllPlayers()
	player := players[0]

	// Give player 2 steel production
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceSteelProduction: 2,
	})
	testutil.AssertEqual(t, 2, player.Resources().Production().Steel, "Should have 2 steel production")

	// Reduce steel production by 5 (should clamp to 0)
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceSteelProduction: -5,
	})
	testutil.AssertEqual(t, shared.MinOtherProduction, player.Resources().Production().Steel, "Steel production should be clamped to 0")
}

func TestAddProduction_AllNonCreditProductionClampsToZero(t *testing.T) {
	// Setup: Create game with a player
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1)
	players := testGame.GetAllPlayers()
	player := players[0]

	// Give player 1 of each non-MC production
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceSteelProduction:    1,
		shared.ResourceTitaniumProduction: 1,
		shared.ResourcePlantProduction:    1,
		shared.ResourceEnergyProduction:   1,
		shared.ResourceHeatProduction:     1,
	})

	// Reduce all by 5 (should all clamp to 0)
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceSteelProduction:    -5,
		shared.ResourceTitaniumProduction: -5,
		shared.ResourcePlantProduction:    -5,
		shared.ResourceEnergyProduction:   -5,
		shared.ResourceHeatProduction:     -5,
	})

	production := player.Resources().Production()
	testutil.AssertEqual(t, 0, production.Steel, "Steel production should be clamped to 0")
	testutil.AssertEqual(t, 0, production.Titanium, "Titanium production should be clamped to 0")
	testutil.AssertEqual(t, 0, production.Plants, "Plants production should be clamped to 0")
	testutil.AssertEqual(t, 0, production.Energy, "Energy production should be clamped to 0")
	testutil.AssertEqual(t, 0, production.Heat, "Heat production should be clamped to 0")
}

func TestAddProduction_ExactlyMinusFive(t *testing.T) {
	// Setup: Create game with a player
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1)
	players := testGame.GetAllPlayers()
	player := players[0]

	// Give player 3 MC production
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: 3,
	})

	// Reduce MC production by 8 (should result in exactly -5)
	player.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceCreditProduction: -8,
	})
	testutil.AssertEqual(t, -5, player.Resources().Production().Credits, "MC production should be exactly -5")
}
