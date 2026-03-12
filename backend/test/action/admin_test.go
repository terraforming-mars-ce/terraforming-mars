package action_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action/admin"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- GiveCard ---

func TestGiveCard_Success(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	cardID := testutil.CardID("Asteroid Mining")
	action := admin.NewGiveCardAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, cardID)
	testutil.AssertNoError(t, err, "GiveCard should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertTrue(t, p.Hand().HasCard(cardID), "Player should have the card in hand")
}

func TestGiveCard_MultipleCards(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewGiveCardAction(repo, cardRegistry, logger)

	card1 := testutil.CardID("Asteroid Mining")
	card2 := testutil.CardID("Power Plant")

	err := action.Execute(ctx, testGame.ID(), playerID, card1)
	testutil.AssertNoError(t, err, "GiveCard should succeed for first card")

	err = action.Execute(ctx, testGame.ID(), playerID, card2)
	testutil.AssertNoError(t, err, "GiveCard should succeed for second card")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertTrue(t, p.Hand().HasCard(card1), "Player should have the first card")
	testutil.AssertTrue(t, p.Hand().HasCard(card2), "Player should have the second card")
}

func TestGiveCard_InvalidGame(t *testing.T) {
	_, repo, cardRegistry, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewGiveCardAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player", testutil.CardID("Asteroid Mining"))
	testutil.AssertError(t, err, "GiveCard should fail for invalid game")
}

func TestGiveCard_InvalidPlayer(t *testing.T) {
	testGame, repo, cardRegistry, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewGiveCardAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", testutil.CardID("Asteroid Mining"))
	testutil.AssertError(t, err, "GiveCard should fail for invalid player")
}

// --- SetResources ---

func TestSetResources_Success(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetResourcesAction(repo, logger)
	resources := shared.Resources{
		Credits:  100,
		Steel:    10,
		Titanium: 5,
		Plants:   8,
		Energy:   3,
		Heat:     7,
	}
	err := action.Execute(ctx, testGame.ID(), playerID, resources)
	testutil.AssertNoError(t, err, "SetResources should succeed")

	p, _ := testGame.GetPlayer(playerID)
	got := p.Resources().Get()
	testutil.AssertEqual(t, 100, got.Credits, "Credits should be 100")
	testutil.AssertEqual(t, 10, got.Steel, "Steel should be 10")
	testutil.AssertEqual(t, 5, got.Titanium, "Titanium should be 5")
	testutil.AssertEqual(t, 8, got.Plants, "Plants should be 8")
	testutil.AssertEqual(t, 3, got.Energy, "Energy should be 3")
	testutil.AssertEqual(t, 7, got.Heat, "Heat should be 7")
}

func TestSetResources_ZeroValues(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// First set some resources
	action := admin.NewSetResourcesAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, shared.Resources{Credits: 50, Steel: 10})
	testutil.AssertNoError(t, err, "SetResources should succeed")

	// Then set to zero
	err = action.Execute(ctx, testGame.ID(), playerID, shared.Resources{})
	testutil.AssertNoError(t, err, "SetResources to zero should succeed")

	p, _ := testGame.GetPlayer(playerID)
	got := p.Resources().Get()
	testutil.AssertEqual(t, 0, got.Credits, "Credits should be 0")
	testutil.AssertEqual(t, 0, got.Steel, "Steel should be 0")
}

func TestSetResources_OnlyAffectsTargetPlayer(t *testing.T) {
	testGame, repo, _, player1, player2 := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p2, _ := testGame.GetPlayer(player2)
	originalP2Resources := p2.Resources().Get()

	action := admin.NewSetResourcesAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), player1, shared.Resources{Credits: 999})
	testutil.AssertNoError(t, err, "SetResources should succeed")

	p2After, _ := testGame.GetPlayer(player2)
	testutil.AssertEqual(t, originalP2Resources.Credits, p2After.Resources().Get().Credits, "Player 2 credits should be unchanged")
}

func TestSetResources_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetResourcesAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player", shared.Resources{Credits: 10})
	testutil.AssertError(t, err, "SetResources should fail for invalid game")
}

func TestSetResources_InvalidPlayer(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetResourcesAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", shared.Resources{Credits: 10})
	testutil.AssertError(t, err, "SetResources should fail for invalid player")
}

// --- SetProduction ---

func TestSetProduction_Success(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetProductionAction(repo, logger)
	production := shared.Production{
		Credits:  5,
		Steel:    3,
		Titanium: 2,
		Plants:   1,
		Energy:   4,
		Heat:     6,
	}
	err := action.Execute(ctx, testGame.ID(), playerID, production)
	testutil.AssertNoError(t, err, "SetProduction should succeed")

	p, _ := testGame.GetPlayer(playerID)
	got := p.Resources().Production()
	testutil.AssertEqual(t, 5, got.Credits, "Credits production should be 5")
	testutil.AssertEqual(t, 3, got.Steel, "Steel production should be 3")
	testutil.AssertEqual(t, 2, got.Titanium, "Titanium production should be 2")
	testutil.AssertEqual(t, 1, got.Plants, "Plants production should be 1")
	testutil.AssertEqual(t, 4, got.Energy, "Energy production should be 4")
	testutil.AssertEqual(t, 6, got.Heat, "Heat production should be 6")
}

func TestSetProduction_NegativeCredits(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetProductionAction(repo, logger)
	production := shared.Production{Credits: -5}
	err := action.Execute(ctx, testGame.ID(), playerID, production)
	testutil.AssertNoError(t, err, "SetProduction with negative credits should succeed")

	p, _ := testGame.GetPlayer(playerID)
	got := p.Resources().Production()
	testutil.AssertEqual(t, -5, got.Credits, "Credits production should be -5")
}

func TestSetProduction_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetProductionAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player", shared.Production{})
	testutil.AssertError(t, err, "SetProduction should fail for invalid game")
}

func TestSetProduction_InvalidPlayer(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetProductionAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", shared.Production{})
	testutil.AssertError(t, err, "SetProduction should fail for invalid player")
}

// --- SetTR ---

func TestSetTR_Success(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetTRAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, 30)
	testutil.AssertNoError(t, err, "SetTR should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertEqual(t, 30, p.Resources().TerraformRating(), "TR should be 30")
}

func TestSetTR_HighValue(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetTRAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, 100)
	testutil.AssertNoError(t, err, "SetTR with high value should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertEqual(t, 100, p.Resources().TerraformRating(), "TR should be 100")
}

func TestSetTR_ZeroValue(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetTRAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, 0)
	testutil.AssertNoError(t, err, "SetTR to 0 should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertEqual(t, 0, p.Resources().TerraformRating(), "TR should be 0")
}

func TestSetTR_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetTRAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player", 30)
	testutil.AssertError(t, err, "SetTR should fail for invalid game")
}

func TestSetTR_InvalidPlayer(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetTRAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", 30)
	testutil.AssertError(t, err, "SetTR should fail for invalid player")
}

// --- SetGlobalParameters ---

func TestSetGlobalParameters_Success(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetGlobalParametersAction(repo, logger)
	params := admin.SetGlobalParametersRequest{
		Temperature: 4,
		Oxygen:      7,
		Oceans:      5,
		Venus:       10,
	}
	err := action.Execute(ctx, testGame.ID(), params)
	testutil.AssertNoError(t, err, "SetGlobalParameters should succeed")

	gp := testGame.GlobalParameters()
	testutil.AssertEqual(t, 4, gp.Temperature(), "Temperature should be 4")
	testutil.AssertEqual(t, 7, gp.Oxygen(), "Oxygen should be 7")
	testutil.AssertEqual(t, 5, gp.Oceans(), "Oceans should be 5")
	testutil.AssertEqual(t, 10, gp.Venus(), "Venus should be 10")
}

func TestSetGlobalParameters_ZeroValues(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetGlobalParametersAction(repo, logger)
	params := admin.SetGlobalParametersRequest{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
		Venus:       0,
	}
	err := action.Execute(ctx, testGame.ID(), params)
	testutil.AssertNoError(t, err, "SetGlobalParameters to zero should succeed")

	gp := testGame.GlobalParameters()
	testutil.AssertEqual(t, 0, gp.Temperature(), "Temperature should be 0")
	testutil.AssertEqual(t, 0, gp.Oxygen(), "Oxygen should be 0")
	testutil.AssertEqual(t, 0, gp.Oceans(), "Oceans should be 0")
	testutil.AssertEqual(t, 0, gp.Venus(), "Venus should be 0")
}

func TestSetGlobalParameters_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetGlobalParametersAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", admin.SetGlobalParametersRequest{})
	testutil.AssertError(t, err, "SetGlobalParameters should fail for invalid game")
}

// --- SetPhase ---

func TestSetPhase_Success(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetPhaseAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), shared.GamePhaseProductionAndCardDraw)
	testutil.AssertNoError(t, err, "SetPhase should succeed")

	testutil.AssertEqual(t, shared.GamePhaseProductionAndCardDraw, testGame.CurrentPhase(), "Phase should be production_and_card_draw")
}

func TestSetPhase_BackToAction(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetPhaseAction(repo, logger)

	err := action.Execute(ctx, testGame.ID(), shared.GamePhaseProductionAndCardDraw)
	testutil.AssertNoError(t, err, "SetPhase to production should succeed")

	err = action.Execute(ctx, testGame.ID(), shared.GamePhaseAction)
	testutil.AssertNoError(t, err, "SetPhase back to action should succeed")

	testutil.AssertEqual(t, shared.GamePhaseAction, testGame.CurrentPhase(), "Phase should be action")
}

func TestSetPhase_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetPhaseAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", shared.GamePhaseAction)
	testutil.AssertError(t, err, "SetPhase should fail for invalid game")
}

// --- SetCurrentTurn ---

func TestSetCurrentTurn_Success(t *testing.T) {
	testGame, repo, _, player1, player2 := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Initially player1 has the turn
	testutil.AssertEqual(t, player1, testGame.CurrentTurn().PlayerID(), "Player 1 should have the initial turn")

	action := admin.NewSetCurrentTurnAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), player2)
	testutil.AssertNoError(t, err, "SetCurrentTurn should succeed")

	testutil.AssertEqual(t, player2, testGame.CurrentTurn().PlayerID(), "Player 2 should now have the turn")
}

func TestSetCurrentTurn_SwitchBackAndForth(t *testing.T) {
	testGame, repo, _, player1, player2 := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCurrentTurnAction(repo, logger)

	err := action.Execute(ctx, testGame.ID(), player2)
	testutil.AssertNoError(t, err, "SetCurrentTurn to player2 should succeed")
	testutil.AssertEqual(t, player2, testGame.CurrentTurn().PlayerID(), "Player 2 should have the turn")

	err = action.Execute(ctx, testGame.ID(), player1)
	testutil.AssertNoError(t, err, "SetCurrentTurn back to player1 should succeed")
	testutil.AssertEqual(t, player1, testGame.CurrentTurn().PlayerID(), "Player 1 should have the turn again")
}

func TestSetCurrentTurn_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCurrentTurnAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player")
	testutil.AssertError(t, err, "SetCurrentTurn should fail for invalid game")
}

func TestSetCurrentTurn_InvalidPlayer(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCurrentTurnAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player")
	testutil.AssertError(t, err, "SetCurrentTurn should fail for invalid player")
}

// --- SetCorporation ---

func TestSetCorporation_Success(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	corpID := testutil.CardID("CrediCor")
	err := action.Execute(ctx, testGame.ID(), playerID, corpID)
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertEqual(t, corpID, p.CorporationID(), "Corporation should be CrediCor")
}

func TestSetCorporation_ChangeCorporation(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)

	corp1 := testutil.CardID("CrediCor")
	err := action.Execute(ctx, testGame.ID(), playerID, corp1)
	testutil.AssertNoError(t, err, "SetCorporation to CrediCor should succeed")

	corp2 := testutil.CardID("Ecoline")
	err = action.Execute(ctx, testGame.ID(), playerID, corp2)
	testutil.AssertNoError(t, err, "SetCorporation to EcoLine should succeed")

	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertEqual(t, corp2, p.CorporationID(), "Corporation should now be EcoLine")
}

func TestSetCorporation_InvalidGame(t *testing.T) {
	_, repo, cardRegistry, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player", testutil.CardID("CrediCor"))
	testutil.AssertError(t, err, "SetCorporation should fail for invalid game")
}

func TestSetCorporation_InvalidPlayer(t *testing.T) {
	testGame, repo, cardRegistry, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", testutil.CardID("CrediCor"))
	testutil.AssertError(t, err, "SetCorporation should fail for invalid player")
}

func TestSetCorporation_InvalidCardID(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "nonexistent-card-id")
	testutil.AssertError(t, err, "SetCorporation should fail for invalid card ID")
}

func TestSetCorporation_RegistersVPGranter(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	corpID := testutil.CardID("Arklight")
	err := action.Execute(ctx, testGame.ID(), playerID, corpID)
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	granters := p.VPGranters().GetAll()
	testutil.AssertEqual(t, 1, len(granters), "Arklight should register 1 VP granter")
	testutil.AssertEqual(t, "Arklight", granters[0].CardName, "VP granter should be for Arklight")
}

func TestSetCorporation_ChangeCorporation_ClearsOldVPGranter(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)

	// Set Arklight (has VP granter: 1 VP per 2 animals)
	arklightID := testutil.CardID("Arklight")
	err := action.Execute(ctx, testGame.ID(), playerID, arklightID)
	testutil.AssertNoError(t, err, "SetCorporation to Arklight should succeed")

	p, _ := testGame.GetPlayer(playerID)
	granters := p.VPGranters().GetAll()
	testutil.AssertEqual(t, 1, len(granters), "Should have 1 VP granter after setting Arklight")
	testutil.AssertEqual(t, "Arklight", granters[0].CardName, "VP granter should be for Arklight")

	// Switch to Celestic (has VP granter: 1 VP per 3 floaters)
	celesticID := testutil.CardID("Celestic")
	err = action.Execute(ctx, testGame.ID(), playerID, celesticID)
	testutil.AssertNoError(t, err, "SetCorporation to Celestic should succeed")

	granters = p.VPGranters().GetAll()
	testutil.AssertEqual(t, 1, len(granters), "Should still have 1 VP granter after switching to Celestic")
	testutil.AssertEqual(t, "Celestic", granters[0].CardName, "VP granter should now be for Celestic")
}

func TestSetCorporation_NoVPGranterForCorpWithoutVP(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	corpID := testutil.CardID("CrediCor")
	err := action.Execute(ctx, testGame.ID(), playerID, corpID)
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	p, _ := testGame.GetPlayer(playerID)
	granters := p.VPGranters().GetAll()
	testutil.AssertEqual(t, 0, len(granters), "CrediCor should not register any VP granters")
}

func TestSetCorporation_NonCorporationCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Asteroid Mining"))
	testutil.AssertError(t, err, "SetCorporation should fail for non-corporation card")
}

// --- StartTileSelection ---

func TestStartTileSelection_Success(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewStartTileSelectionAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, board.TileTypeCity)
	testutil.AssertNoError(t, err, "StartTileSelection should succeed")

	// Queue is auto-processed into a PendingTileSelection
	selection := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, selection != nil, "Pending tile selection should be set")
	testutil.AssertEqual(t, board.TileTypeCity, selection.TileType, "Tile type should be city")
}

func TestStartTileSelection_DifferentTileTypes(t *testing.T) {
	tileTypes := []string{
		board.TileTypeCity,
		board.TileTypeGreenery,
		board.TileTypeOcean,
	}

	for _, tileType := range tileTypes {
		t.Run(tileType, func(t *testing.T) {
			testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
			logger := testutil.TestLogger()
			ctx := context.Background()

			action := admin.NewStartTileSelectionAction(repo, logger)
			err := action.Execute(ctx, testGame.ID(), playerID, tileType)
			testutil.AssertNoError(t, err, "StartTileSelection should succeed for "+tileType)

			selection := testGame.GetPendingTileSelection(playerID)
			testutil.AssertTrue(t, selection != nil, "Pending tile selection should be set")
			testutil.AssertEqual(t, tileType, selection.TileType, "Tile type should match")
		})
	}
}

func TestStartTileSelection_InvalidTileType(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewStartTileSelectionAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "invalid-tile-type")
	testutil.AssertError(t, err, "StartTileSelection should fail for invalid tile type")
}

func TestStartTileSelection_InvalidGame(t *testing.T) {
	_, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewStartTileSelectionAction(repo, logger)
	err := action.Execute(ctx, "nonexistent-game", "some-player", board.TileTypeCity)
	testutil.AssertError(t, err, "StartTileSelection should fail for invalid game")
}

func TestStartTileSelection_InvalidPlayer(t *testing.T) {
	testGame, repo, _, _, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := admin.NewStartTileSelectionAction(repo, logger)
	err := action.Execute(ctx, testGame.ID(), "nonexistent-player", board.TileTypeCity)
	testutil.AssertError(t, err, "StartTileSelection should fail for invalid player")
}

// --- Combined admin operations ---

func TestAdminActions_CombinedSetup(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Set corporation
	setCorp := admin.NewSetCorporationAction(repo, cardRegistry, logger)
	err := setCorp.Execute(ctx, testGame.ID(), playerID, testutil.CardID("CrediCor"))
	testutil.AssertNoError(t, err, "SetCorporation should succeed")

	// Set resources
	setRes := admin.NewSetResourcesAction(repo, logger)
	err = setRes.Execute(ctx, testGame.ID(), playerID, shared.Resources{Credits: 50, Steel: 5, Titanium: 3})
	testutil.AssertNoError(t, err, "SetResources should succeed")

	// Set production
	setProd := admin.NewSetProductionAction(repo, logger)
	err = setProd.Execute(ctx, testGame.ID(), playerID, shared.Production{Credits: 3, Steel: 2})
	testutil.AssertNoError(t, err, "SetProduction should succeed")

	// Set TR
	setTR := admin.NewSetTRAction(repo, logger)
	err = setTR.Execute(ctx, testGame.ID(), playerID, 25)
	testutil.AssertNoError(t, err, "SetTR should succeed")

	// Give card
	giveCard := admin.NewGiveCardAction(repo, cardRegistry, logger)
	cardID := testutil.CardID("Asteroid Mining")
	err = giveCard.Execute(ctx, testGame.ID(), playerID, cardID)
	testutil.AssertNoError(t, err, "GiveCard should succeed")

	// Verify all state
	p, _ := testGame.GetPlayer(playerID)
	testutil.AssertEqual(t, testutil.CardID("CrediCor"), p.CorporationID(), "Corporation should be CrediCor")
	testutil.AssertEqual(t, 50, p.Resources().Get().Credits, "Credits should be 50")
	testutil.AssertEqual(t, 5, p.Resources().Get().Steel, "Steel should be 5")
	testutil.AssertEqual(t, 3, p.Resources().Get().Titanium, "Titanium should be 3")
	testutil.AssertEqual(t, 3, p.Resources().Production().Credits, "Credits production should be 3")
	testutil.AssertEqual(t, 2, p.Resources().Production().Steel, "Steel production should be 2")
	testutil.AssertEqual(t, 25, p.Resources().TerraformRating(), "TR should be 25")
	testutil.AssertTrue(t, p.Hand().HasCard(cardID), "Player should have the card")
}
