package core_test

import (
	"context"
	"testing"

	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	gameaction "terraforming-mars-backend/internal/action/game"
	turnmgmt "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestLastRound_FinalProductionRunsBeforeScoring(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	testutil.SetPlayerCredits(ctx, p1, 10)
	testutil.SetPlayerCredits(ctx, p2, 10)

	p1.Resources().SetProduction(shared.Production{Credits: 5})
	p2.Resources().SetProduction(shared.Production{Credits: 3})

	p1TR := p1.Resources().TerraformRating()
	p2TR := p2.Resources().TerraformRating()

	gp := g.GlobalParameters()
	for !gp.IsMaxed() {
		if gp.Temperature() < 8 {
			if _, err := gp.IncreaseTemperature(ctx, 1, ""); err != nil {
				t.Fatalf("increase temp: %v", err)
			}
		}
		if gp.Oxygen() < 14 {
			if _, err := gp.IncreaseOxygen(ctx, 1, ""); err != nil {
				t.Fatalf("increase oxygen: %v", err)
			}
		}
		if gp.Oceans() < 9 {
			if _, err := gp.PlaceOcean(ctx, ""); err != nil {
				t.Fatalf("place ocean: %v", err)
			}
		}
	}

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)
	confirmProdAction := confirmAction.NewConfirmProductionCardsAction(repo, cardRegistry, finalScoringAction, logger)

	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn p1")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")

	// After all pass, game should be in production phase (not completed yet)
	testutil.AssertEqual(t, shared.GamePhaseProductionAndCardDraw, g.CurrentPhase(), "Game should be in production phase")

	// Resources should already reflect production
	p1After, _ := g.GetPlayer(p1ID)
	p2After, _ := g.GetPlayer(p2ID)
	expectedP1Credits := 10 + 5 + p1TR
	expectedP2Credits := 10 + 3 + p2TR
	testutil.AssertEqual(t, expectedP1Credits, p1After.Resources().Get().Credits, "P1 credits should include final production")
	testutil.AssertEqual(t, expectedP2Credits, p2After.Resources().Get().Credits, "P2 credits should include final production")

	// Both players confirm production (no cards to select)
	err = confirmProdAction.Execute(ctx, g.ID(), p1ID, []string{})
	testutil.AssertNoError(t, err, "P1 confirm production")
	err = confirmProdAction.Execute(ctx, g.ID(), p2ID, []string{})
	testutil.AssertNoError(t, err, "P2 confirm production")

	// Now game should be completed
	testutil.AssertEqual(t, shared.GameStatusCompleted, g.Status(), "Game should be completed after confirming final production")
}

func TestLastRound_NoCardDrawInFinalProduction(t *testing.T) {
	g, repo, cardRegistry, p1ID, p2ID := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()

	p1, _ := g.GetPlayer(p1ID)
	p2, _ := g.GetPlayer(p2ID)

	initialP1HandSize := len(p1.Hand().Cards())
	initialP2HandSize := len(p2.Hand().Cards())

	gp := g.GlobalParameters()
	for !gp.IsMaxed() {
		if gp.Temperature() < 8 {
			if _, err := gp.IncreaseTemperature(ctx, 1, ""); err != nil {
				t.Fatalf("increase temp: %v", err)
			}
		}
		if gp.Oxygen() < 14 {
			if _, err := gp.IncreaseOxygen(ctx, 1, ""); err != nil {
				t.Fatalf("increase oxygen: %v", err)
			}
		}
		if gp.Oceans() < 9 {
			if _, err := gp.PlaceOcean(ctx, ""); err != nil {
				t.Fatalf("place ocean: %v", err)
			}
		}
	}

	finalScoringAction := gameaction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnmgmt.NewSkipActionAction(repo, finalScoringAction, logger)

	err := g.SetCurrentTurn(ctx, p1ID, 2)
	testutil.AssertNoError(t, err, "Set current turn p1")
	err = skipAction.Execute(ctx, g.ID(), p1ID)
	testutil.AssertNoError(t, err, "P1 pass")
	err = skipAction.Execute(ctx, g.ID(), p2ID)
	testutil.AssertNoError(t, err, "P2 pass")

	// Production phase data should have empty available cards
	p1Phase := g.GetProductionPhase(p1ID)
	p2Phase := g.GetProductionPhase(p2ID)
	testutil.AssertEqual(t, 0, len(p1Phase.AvailableCards), "P1 should have no available cards in final production")
	testutil.AssertEqual(t, 0, len(p2Phase.AvailableCards), "P2 should have no available cards in final production")

	// Hand sizes should not have changed
	p1After, _ := g.GetPlayer(p1ID)
	p2After, _ := g.GetPlayer(p2ID)
	testutil.AssertEqual(t, initialP1HandSize, len(p1After.Hand().Cards()), "P1 hand size should not change in final production")
	testutil.AssertEqual(t, initialP2HandSize, len(p2After.Hand().Cards()), "P2 hand size should not change in final production")
}

func TestLastRound_IsLastRoundDtoField(t *testing.T) {
	g, _, cardRegistry, _, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	gameDto := dto.ToGameDtoFull(g, cardRegistry, g.TurnOrder()[0], dto.Registries{})
	testutil.AssertFalse(t, gameDto.IsLastRound, "IsLastRound should be false when parameters not maxed")

	gp := g.GlobalParameters()
	for !gp.IsMaxed() {
		if gp.Temperature() < 8 {
			if _, err := gp.IncreaseTemperature(ctx, 1, ""); err != nil {
				t.Fatalf("increase temp: %v", err)
			}
		}
		if gp.Oxygen() < 14 {
			if _, err := gp.IncreaseOxygen(ctx, 1, ""); err != nil {
				t.Fatalf("increase oxygen: %v", err)
			}
		}
		if gp.Oceans() < 9 {
			if _, err := gp.PlaceOcean(ctx, ""); err != nil {
				t.Fatalf("place ocean: %v", err)
			}
		}
	}

	gameDto = dto.ToGameDtoFull(g, cardRegistry, g.TurnOrder()[0], dto.Registries{})
	testutil.AssertTrue(t, gameDto.IsLastRound, "IsLastRound should be true when all parameters maxed")
}
