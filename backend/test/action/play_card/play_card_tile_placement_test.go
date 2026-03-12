package play_card_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Mangrove (059) ---
// "Place a Greenery tile on an area reserved for ocean and raise oxygen 1 step."

func TestMangrove_GreeneryOnOceanTileRestriction(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	mangrove := gamecards.Card{
		ID:   "card-mangrove",
		Name: "Mangrove",
		Type: gamecards.CardTypeAutomated,
		Cost: 12,
		Tags: []shared.CardTag{shared.TagPlant},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{
						ResourceType: shared.ResourceGreeneryPlacement,
						Amount:       1,
						Target:       "none",
						TileRestrictions: &shared.TileRestrictions{
							OnTileType: "ocean",
						},
					},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{mangrove})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-mangrove")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 12}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-mangrove", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Mangrove should play successfully")

	// After playing, ProcessNextTile moves the tile from the queue to PendingTileSelection
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection")
	testutil.AssertEqual(t, "greenery", selection.TileType, "Pending tile type should be greenery")
}

// --- Land Claim (066) ---
// "Place your marker on a non-reserved area. Only you may place a tile here."

func TestLandClaim_CreatesLandClaimTileSelection(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	landClaim := gamecards.Card{
		ID:   "card-land-claim",
		Name: "Land Claim",
		Type: gamecards.CardTypeEvent,
		Cost: 1,
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceLandClaim, Amount: 1, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{landClaim})

	players := testGame.GetAllPlayers()
	p := players[0]
	other := players[1]
	p.SetCorporationID("corp-tharsis-republic")
	other.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-land-claim")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 1}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-land-claim", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Land Claim should play successfully")

	// After playing, ProcessNextTile moves the tile from the queue to PendingTileSelection
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection for land claim")
	testutil.AssertEqual(t, "land-claim", selection.TileType, "Pending tile type should be land-claim")
}

// --- Artificial Lake (116) ---
// "Place 1 ocean tile on an area not reserved for ocean."

func TestArtificialLake_OceanPlacement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	artificialLake := gamecards.Card{
		ID:   "card-artificial-lake",
		Name: "Artificial Lake",
		Type: gamecards.CardTypeAutomated,
		Cost: 15,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-6)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{artificialLake})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Meet temperature requirement
	testutil.AssertNoError(t, testGame.GlobalParameters().SetTemperature(ctx, 0), "set temperature")

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-artificial-lake")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-artificial-lake", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Artificial Lake should play successfully")

	// After playing, ProcessNextTile moves the tile from the queue to PendingTileSelection
	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending tile selection for ocean placement")
	testutil.AssertEqual(t, "ocean", selection.TileType, "Pending tile type should be ocean")
}

func TestArtificialLake_FailsWithoutTemperatureRequirement(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	artificialLake := gamecards.Card{
		ID:   "card-artificial-lake",
		Name: "Artificial Lake",
		Type: gamecards.CardTypeAutomated,
		Cost: 15,
		Tags: []shared.CardTag{shared.TagBuilding},
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-6)},
		}},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
				},
			},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{artificialLake})

	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID("corp-tharsis-republic")

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, p.ID(), 2), "set current turn")

	// Temperature below requirement (-30 default, need -6)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Hand().AddCard("card-artificial-lake")

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 15}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), "card-artificial-lake", payment, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Artificial Lake should fail without meeting temperature requirement")
	testutil.AssertTrue(t, p.Hand().HasCard("card-artificial-lake"), "Card should still be in hand")
}
