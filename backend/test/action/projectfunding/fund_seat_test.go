package projectfunding_test

import (
	"context"
	"testing"

	pfAction "terraforming-mars-backend/internal/action/projectfunding"
	"terraforming-mars-backend/internal/game"
	pf "terraforming-mars-backend/internal/game/projectfunding"
	"terraforming-mars-backend/internal/game/shared"
	pfLoader "terraforming-mars-backend/internal/projectfunding"
	"terraforming-mars-backend/test/testutil"
)

func setupProjectFundingGame(t *testing.T) (*game.Game, game.GameRepository, pfLoader.ProjectFundingRegistry, string, string) {
	t.Helper()
	testGame, repo, _, player1, player2 := testutil.SetupTwoPlayerGame(t)

	pfDefs, err := pfLoader.LoadProjectsFromJSON("../../../assets/terraforming_mars_project_funding.json")
	if err != nil {
		t.Fatalf("Failed to load project funding: %v", err)
	}
	pfRegistry := pfLoader.NewInMemoryProjectFundingRegistry(pfDefs)

	settings := testGame.Settings()
	settings.CardPacks = append(settings.CardPacks, shared.PackProjectFunding)
	testGame.UpdateSettings(context.Background(), settings)

	return testGame, repo, pfRegistry, player1, player2
}

func setupProjectState(g *game.Game, projectID string, seatOwners []string) {
	states := g.ProjectFundingStates()
	states = append(states, &pf.ProjectState{
		DefinitionID: projectID,
		SeatOwners:   seatOwners,
		IsCompleted:  false,
	})
	g.SetProjectFundingStates(states)
}

func newAction(repo game.GameRepository, pfReg pfLoader.ProjectFundingRegistry) *pfAction.FundSeatAction {
	stateRepo := game.NewInMemoryGameStateRepository()
	return pfAction.NewFundSeatAction(repo, pfReg, stateRepo)
}

// --- Validation Tests ---

func TestFundSeat_ExpansionNotEnabled_Fails(t *testing.T) {
	testGame, repo, _, player1, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	pfDefs, _ := pfLoader.LoadProjectsFromJSON("../../../assets/terraforming_mars_project_funding.json")
	pfRegistry := pfLoader.NewInMemoryProjectFundingRegistry(pfDefs)

	setupProjectState(testGame, "pf_orbital_station", nil)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertError(t, err, "Should fail when project funding expansion is not enabled")
}

func TestFundSeat_WrongPhase_Fails(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", nil)
	if err := testGame.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw); err != nil {
		t.Fatalf("Failed to update phase: %v", err)
	}

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertError(t, err, "Should fail when not in action phase")
}

func TestFundSeat_NotCurrentTurn_Fails(t *testing.T) {
	testGame, repo, pfRegistry, _, player2 := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", nil)

	p2, _ := testGame.GetPlayer(player2)
	testutil.SetPlayerCredits(ctx, p2, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player2, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertError(t, err, "Should fail when not player's turn")
}

func TestFundSeat_InvalidProjectID_Fails(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "nonexistent_project", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertError(t, err, "Should fail with invalid project ID")
}

func TestFundSeat_ProjectAlreadyCompleted_Fails(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	states := testGame.ProjectFundingStates()
	states = append(states, &pf.ProjectState{
		DefinitionID: "pf_orbital_station",
		SeatOwners:   []string{"other1", "other2"},
		IsCompleted:  true,
	})
	testGame.SetProjectFundingStates(states)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertError(t, err, "Should fail when project is completed")
}

// --- Basic Purchase Tests ---

func TestFundSeat_BasicPurchase_Success(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", nil)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertNoError(t, err, "Buy first seat should succeed")

	state := testGame.GetProjectFundingState("pf_orbital_station")
	testutil.AssertEqual(t, 1, len(state.SeatOwners), "Should have 1 seat owner")
	testutil.AssertEqual(t, player1, state.SeatOwners[0], "Owner should be player1")
	testutil.AssertEqual(t, 94, testutil.GetPlayerCredits(p1), "Should deduct 6 credits (100 - 6 = 94)")
}

func TestFundSeat_InsufficientCredits_Fails(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", nil)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 3)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertError(t, err, "Should fail with insufficient credits")
}

func TestFundSeat_SecondSeat_HigherCost(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", []string{"other"})

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 8})
	testutil.AssertNoError(t, err, "Buy second seat should succeed")

	state := testGame.GetProjectFundingState("pf_orbital_station")
	testutil.AssertEqual(t, 2, len(state.SeatOwners), "Should have 2 seat owners")
	testutil.AssertEqual(t, 92, testutil.GetPlayerCredits(p1), "Second seat costs 8 (100 - 8 = 92)")
}

// --- Payment Substitute Tests ---

func TestFundSeat_PayWithSteel_InvalidSeat_Fails(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	// Orbital station first seat has no steel substitution
	setupProjectState(testGame, "pf_orbital_station", nil)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)
	p1.Resources().Add(map[shared.ResourceType]int{shared.ResourceSteel: 10})

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 0, Steel: 3})
	testutil.AssertError(t, err, "Should fail when steel not allowed for this seat")
}

// --- Multi-player Seat Ownership Tests ---

func TestFundSeat_SamePlayerMultipleSeats(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", nil)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 500)

	action := newAction(repo, pfRegistry)

	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertNoError(t, err, "Buy first seat should succeed")
	if err := testGame.SetCurrentTurn(ctx, player1, 2); err != nil {
		t.Fatalf("Failed to set current turn: %v", err)
	}

	err = action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 8})
	testutil.AssertNoError(t, err, "Buy second seat should succeed")
	if err := testGame.SetCurrentTurn(ctx, player1, 2); err != nil {
		t.Fatalf("Failed to set current turn: %v", err)
	}

	err = action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 10})
	testutil.AssertNoError(t, err, "Buy third seat should succeed")

	state := testGame.GetProjectFundingState("pf_orbital_station")
	testutil.AssertEqual(t, 3, len(state.SeatOwners), "Should have 3 seat owners")
	for i := 0; i < 3; i++ {
		testutil.AssertEqual(t, player1, state.SeatOwners[i], "All seats should belong to player1")
	}
}

// --- Completion & Reward Tests ---

func TestFundSeat_Completion_SetsIsCompleted(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	// Fill all seats except the last one
	def, _ := pfRegistry.GetByID("pf_orbital_station")
	owners := make([]string, len(def.Seats)-1)
	for i := range owners {
		owners[i] = "other_player"
	}
	setupProjectState(testGame, "pf_orbital_station", owners)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 500)

	lastSeatCost := def.Seats[len(def.Seats)-1].Cost

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: lastSeatCost})
	testutil.AssertNoError(t, err, "Buy last seat should succeed")

	state := testGame.GetProjectFundingState("pf_orbital_station")
	testutil.AssertTrue(t, state.IsCompleted, "Project should be completed")
}

func TestFundSeat_Completion_AllPlayersGetCompletionEffect(t *testing.T) {
	testGame, repo, pfRegistry, player1, player2 := setupProjectFundingGame(t)
	ctx := context.Background()

	// pf_orbital_station completion effect: card-draw (5 cards for each player)
	def, _ := pfRegistry.GetByID("pf_orbital_station")
	owners := make([]string, len(def.Seats)-1)
	for i := range owners {
		owners[i] = player1
	}
	setupProjectState(testGame, "pf_orbital_station", owners)

	p1, _ := testGame.GetPlayer(player1)
	p2, _ := testGame.GetPlayer(player2)
	testutil.SetPlayerCredits(ctx, p1, 500)

	p2HandBefore := len(p2.Hand().Cards())
	lastSeatCost := def.Seats[len(def.Seats)-1].Cost

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: lastSeatCost})
	testutil.AssertNoError(t, err, "Buy last seat should succeed")

	// Player2 should receive the card-draw completion effect even though they have no seats
	p2HandAfter := len(p2.Hand().Cards())
	testutil.AssertTrue(t, p2HandAfter > p2HandBefore, "Player2 should receive cards from completion effect even without seats")
}

func TestFundSeat_NotCompleted_WhenSeatsRemain(t *testing.T) {
	testGame, repo, pfRegistry, player1, _ := setupProjectFundingGame(t)
	ctx := context.Background()

	setupProjectState(testGame, "pf_orbital_station", nil)

	p1, _ := testGame.GetPlayer(player1)
	testutil.SetPlayerCredits(ctx, p1, 100)

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: 6})
	testutil.AssertNoError(t, err, "Buy first seat should succeed")

	state := testGame.GetProjectFundingState("pf_orbital_station")
	testutil.AssertFalse(t, state.IsCompleted, "Project should not be completed with seats remaining")
}

// --- Global Effect Tests ---

func TestFundSeat_Completion_TerraformingHub_TRAndOxygen(t *testing.T) {
	testGame, repo, pfRegistry, player1, player2 := setupProjectFundingGame(t)
	ctx := context.Background()

	// pf_terraforming_hub: completion gives +3 TR to all players + raises oxygen 2 steps
	def, _ := pfRegistry.GetByID("pf_terraforming_hub")
	// Use "other_player" for all preceding seats so P1 only owns the final seat (no tier reward)
	owners := make([]string, len(def.Seats)-1)
	for i := range owners {
		owners[i] = "other_player"
	}
	setupProjectState(testGame, "pf_terraforming_hub", owners)

	p1, _ := testGame.GetPlayer(player1)
	p2, _ := testGame.GetPlayer(player2)
	testutil.SetPlayerCredits(ctx, p1, 500)

	p1TRBefore := p1.Resources().TerraformRating()
	p2TRBefore := p2.Resources().TerraformRating()
	oxygenBefore := testGame.GlobalParameters().Oxygen()
	lastSeatCost := def.Seats[len(def.Seats)-1].Cost

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_terraforming_hub", pfAction.FundSeatPayment{Credits: lastSeatCost})
	testutil.AssertNoError(t, err, "Buy last seat should succeed")

	testutil.AssertEqual(t, p1TRBefore+5, p1.Resources().TerraformRating(), "P1 should gain +5 TR from completion")
	testutil.AssertEqual(t, p2TRBefore+5, p2.Resources().TerraformRating(), "P2 should gain +5 TR from completion")
	testutil.AssertEqual(t, oxygenBefore+2, testGame.GlobalParameters().Oxygen(), "Oxygen should increase by 2")
}

func TestFundSeat_Completion_ProductionChoice_SetForAllPlayers(t *testing.T) {
	testGame, repo, pfRegistry, player1, player2 := setupProjectFundingGame(t)
	ctx := context.Background()

	// pf_solar_forge: completion gives each player production choice
	def, _ := pfRegistry.GetByID("pf_solar_forge")
	owners := make([]string, len(def.Seats)-1)
	for i := range owners {
		owners[i] = player1
	}
	setupProjectState(testGame, "pf_solar_forge", owners)

	p1, _ := testGame.GetPlayer(player1)
	p2, _ := testGame.GetPlayer(player2)
	testutil.SetPlayerCredits(ctx, p1, 500)

	lastSeatCost := def.Seats[len(def.Seats)-1].Cost

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_solar_forge", pfAction.FundSeatPayment{Credits: lastSeatCost})
	testutil.AssertNoError(t, err, "Buy last seat should succeed")

	p1Choice := p1.Selection().GetPendingBehaviorChoiceSelection()
	p2Choice := p2.Selection().GetPendingBehaviorChoiceSelection()

	testutil.AssertTrue(t, p1Choice != nil, "P1 should have pending production choice")
	testutil.AssertTrue(t, p2Choice != nil, "P2 should have pending production choice")
	testutil.AssertEqual(t, 6, len(p1Choice.Choices), "Should have 6 production type choices")
	testutil.AssertEqual(t, "project-funding-completion", p1Choice.Source, "Source should be project-funding-completion")
}

func TestFundSeat_Completion_MassCardDraw(t *testing.T) {
	testGame, repo, pfRegistry, player1, player2 := setupProjectFundingGame(t)
	ctx := context.Background()

	// pf_orbital_station: completion deals 20 cards to each player
	def, _ := pfRegistry.GetByID("pf_orbital_station")
	owners := make([]string, len(def.Seats)-1)
	for i := range owners {
		owners[i] = player1
	}
	setupProjectState(testGame, "pf_orbital_station", owners)

	p1, _ := testGame.GetPlayer(player1)
	p2, _ := testGame.GetPlayer(player2)
	testutil.SetPlayerCredits(ctx, p1, 500)

	p1HandBefore := len(p1.Hand().Cards())
	p2HandBefore := len(p2.Hand().Cards())
	lastSeatCost := def.Seats[len(def.Seats)-1].Cost

	action := newAction(repo, pfRegistry)
	err := action.Execute(ctx, testGame.ID(), player1, "pf_orbital_station", pfAction.FundSeatPayment{Credits: lastSeatCost})
	testutil.AssertNoError(t, err, "Buy last seat should succeed")

	p1HandAfter := len(p1.Hand().Cards())
	p2HandAfter := len(p2.Hand().Cards())

	testutil.AssertEqual(t, p1HandBefore+20, p1HandAfter, "P1 should receive 20 cards")
	testutil.AssertEqual(t, p2HandBefore+20, p2HandAfter, "P2 should receive 20 cards")
}
