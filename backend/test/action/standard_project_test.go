package action_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	spAction "terraforming-mars-backend/internal/action/standard_project"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/standardprojects"
	"terraforming-mars-backend/test/testutil"
)

func loadStandardProjectRegistry(t *testing.T) standardprojects.StandardProjectRegistry {
	t.Helper()
	_, currentFile, _, _ := runtime.Caller(0)
	stdProjPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "assets", "terraforming_mars_standard_projects.json")
	stdProjData, err := standardprojects.LoadStandardProjectsFromJSON(stdProjPath)
	if err != nil {
		t.Fatalf("Failed to load standard projects JSON: %v", err)
	}
	return standardprojects.NewInMemoryStandardProjectRegistry(stdProjData)
}

func setupStandardProjectTest(t *testing.T) (*game.Game, game.GameRepository, cards.CardRegistry, standardprojects.StandardProjectRegistry, string) {
	t.Helper()
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	stdProjRegistry := loadStandardProjectRegistry(t)
	return testGame, repo, cardRegistry, stdProjRegistry, playerID
}

// --- Registry Tests ---

func TestStandardProjectRegistry_LoadFromJSON(t *testing.T) {
	registry := loadStandardProjectRegistry(t)

	allProjects := registry.GetAll()
	testutil.AssertEqual(t, 7, len(allProjects), "Should load all standard projects from JSON")
}

func TestStandardProjectRegistry_LookupByID(t *testing.T) {
	registry := loadStandardProjectRegistry(t)

	expectedIDs := []string{"sell-patents", "power-plant", "asteroid", "aquifer", "greenery", "city"}

	for _, id := range expectedIDs {
		t.Run(id, func(t *testing.T) {
			def, err := registry.GetByID(id)
			testutil.AssertNoError(t, err, "Should find standard project: "+id)
			testutil.AssertEqual(t, id, def.ID, "Project ID should match")
		})
	}
}

func TestStandardProjectRegistry_UnknownIDReturnsError(t *testing.T) {
	registry := loadStandardProjectRegistry(t)

	_, err := registry.GetByID("nonexistent-project")
	testutil.AssertError(t, err, "Should return error for unknown project ID")
}

func TestStandardProjectRegistry_ProjectCosts(t *testing.T) {
	registry := loadStandardProjectRegistry(t)

	expectedCosts := map[string]int{
		"sell-patents": 0,
		"power-plant":  11,
		"asteroid":     14,
		"aquifer":      18,
		"greenery":     23,
		"city":         25,
	}

	for id, expectedCost := range expectedCosts {
		t.Run(id, func(t *testing.T) {
			def, err := registry.GetByID(id)
			testutil.AssertNoError(t, err, "Should find project")
			testutil.AssertEqual(t, expectedCost, def.CreditCost(), "Cost mismatch for "+id)
		})
	}
}

// --- Execute Action Tests ---

func TestStandardProject_PowerPlant(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	initialCredits := testutil.GetPlayerCredits(p)
	initialEnergyProd := p.Resources().Production().Energy

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "power-plant")
	testutil.AssertNoError(t, err, "Power Plant should succeed")

	afterCredits := testutil.GetPlayerCredits(p)
	afterEnergyProd := p.Resources().Production().Energy

	testutil.AssertEqual(t, initialCredits-11, afterCredits, "Should deduct 11 credits")
	testutil.AssertEqual(t, initialEnergyProd+1, afterEnergyProd, "Energy production should increase by 1")
}

func TestStandardProject_Asteroid(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	initialCredits := testutil.GetPlayerCredits(p)
	initialTemp := testGame.GlobalParameters().Temperature()
	initialTR := p.Resources().TerraformRating()

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "asteroid")
	testutil.AssertNoError(t, err, "Asteroid should succeed")

	afterCredits := testutil.GetPlayerCredits(p)
	afterTemp := testGame.GlobalParameters().Temperature()
	afterTR := p.Resources().TerraformRating()

	testutil.AssertEqual(t, initialCredits-14, afterCredits, "Should deduct 14 credits")
	testutil.AssertEqual(t, initialTemp+2, afterTemp, "Temperature should increase by 2 (1 step = 2 degrees)")
	testutil.AssertEqual(t, initialTR+1, afterTR, "TR should increase by 1")
}

func TestStandardProject_Aquifer(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	initialCredits := testutil.GetPlayerCredits(p)

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "aquifer")
	testutil.AssertNoError(t, err, "Aquifer should succeed")

	afterCredits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, initialCredits-18, afterCredits, "Should deduct 18 credits")

	tileSel := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, tileSel != nil, "Should have a pending tile selection")
	testutil.AssertEqual(t, "ocean", tileSel.TileType, "Pending tile selection should be for ocean")
}

func TestStandardProject_Greenery(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	initialCredits := testutil.GetPlayerCredits(p)

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "greenery")
	testutil.AssertNoError(t, err, "Greenery should succeed")

	afterCredits := testutil.GetPlayerCredits(p)
	testutil.AssertEqual(t, initialCredits-23, afterCredits, "Should deduct 23 credits")

	tileSel := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, tileSel != nil, "Should have a pending tile selection")
	testutil.AssertEqual(t, "greenery", tileSel.TileType, "Pending tile selection should be for greenery")
}

func TestStandardProject_City(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	initialCredits := testutil.GetPlayerCredits(p)
	initialCreditProd := p.Resources().Production().Credits

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "city")
	testutil.AssertNoError(t, err, "City should succeed")

	afterCredits := testutil.GetPlayerCredits(p)
	afterCreditProd := p.Resources().Production().Credits

	testutil.AssertEqual(t, initialCredits-25, afterCredits, "Should deduct 25 credits")
	testutil.AssertEqual(t, initialCreditProd+1, afterCreditProd, "Credit production should increase by 1")

	tileSel := testGame.GetPendingTileSelection(playerID)
	testutil.AssertTrue(t, tileSel != nil, "Should have a pending tile selection")
	testutil.AssertEqual(t, "city", tileSel.TileType, "Pending tile selection should be for city")
}

func TestStandardProject_SellPatents(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	p.Hand().AddCard("test-card-1")
	p.Hand().AddCard("test-card-2")

	initialActions := testGame.CurrentTurn().ActionsRemaining()

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "sell-patents")
	testutil.AssertNoError(t, err, "Sell Patents should succeed")

	pendingSel := p.Selection().GetPendingCardSelection()
	testutil.AssertTrue(t, pendingSel != nil, "Should have a pending card selection")
	testutil.AssertEqual(t, "sell-patents", pendingSel.Source, "Pending selection source should be sell-patents")
	testutil.AssertEqual(t, 2, len(pendingSel.AvailableCards), "Should have 2 available cards")

	afterActions := testGame.CurrentTurn().ActionsRemaining()
	testutil.AssertEqual(t, initialActions, afterActions, "Sell Patents should NOT consume an action")
}

func TestStandardProject_SellPatentsNoCards(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "sell-patents")
	testutil.AssertError(t, err, "Sell Patents with no cards should fail")
}

// --- Affordability Tests ---

func TestStandardProject_InsufficientCredits(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		credits   int
	}{
		{"Power Plant with 10 credits", "power-plant", 10},
		{"Asteroid with 13 credits", "asteroid", 13},
		{"Aquifer with 17 credits", "aquifer", 17},
		{"Greenery with 22 credits", "greenery", 22},
		{"City with 24 credits", "city", 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
			logger := testutil.TestLogger()
			ctx := context.Background()

			p, _ := testGame.GetPlayer(playerID)
			testutil.SetPlayerCredits(ctx, p, tt.credits)

			action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
			err := action.Execute(ctx, testGame.ID(), playerID, tt.projectID)
			testutil.AssertError(t, err, "Should reject with insufficient credits")
		})
	}
}

func TestStandardProject_ExactCreditsSucceeds(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		credits   int
	}{
		{"Power Plant with exactly 11 credits", "power-plant", 11},
		{"Asteroid with exactly 14 credits", "asteroid", 14},
		{"Aquifer with exactly 18 credits", "aquifer", 18},
		{"Greenery with exactly 23 credits", "greenery", 23},
		{"City with exactly 25 credits", "city", 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
			logger := testutil.TestLogger()
			ctx := context.Background()

			p, _ := testGame.GetPlayer(playerID)
			testutil.SetPlayerCredits(ctx, p, tt.credits)

			action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
			err := action.Execute(ctx, testGame.ID(), playerID, tt.projectID)
			testutil.AssertNoError(t, err, "Should succeed with exact credits")

			afterCredits := testutil.GetPlayerCredits(p)
			testutil.AssertEqual(t, 0, afterCredits, "Should have 0 credits remaining")
		})
	}
}

// --- Unknown Project Test ---

func TestStandardProject_UnknownProjectID(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "nonexistent-project")
	testutil.AssertError(t, err, "Should reject unknown project ID")
}

// --- Action Consumption Tests ---

func TestStandardProject_ConsumesAction(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	initialActions := testGame.CurrentTurn().ActionsRemaining()
	testutil.AssertEqual(t, 2, initialActions, "Should start with 2 actions")

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "power-plant")
	testutil.AssertNoError(t, err, "Power Plant should succeed")

	afterActions := testGame.CurrentTurn().ActionsRemaining()
	testutil.AssertEqual(t, initialActions-1, afterActions, "Should consume 1 action")
}

func TestStandardProject_SellPatentsDoesNotConsumeAction(t *testing.T) {
	testGame, repo, cardRegistry, stdProjRegistry, playerID := setupStandardProjectTest(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	p.Hand().AddCard("test-card-1")

	initialActions := testGame.CurrentTurn().ActionsRemaining()

	action := spAction.NewExecuteStandardProjectAction(repo, cardRegistry, stdProjRegistry, nil, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "sell-patents")
	testutil.AssertNoError(t, err, "Sell Patents should succeed")

	afterActions := testGame.CurrentTurn().ActionsRemaining()
	testutil.AssertEqual(t, initialActions, afterActions, "Sell Patents should not consume an action")
}
