package action_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

func setupTestEnvironment(t *testing.T) (*game.Game, *player.Player, cards.CardRegistry) {
	// Initialize logger
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test game with default settings
	settings := game.GameSettings{
		MaxPlayers:      5,
		DevelopmentMode: true,
	}
	g := game.NewGame("test-game", "player1", settings)

	// Create and add test player
	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Create minimal card registry for testing
	testCards := []gamecards.Card{
		{
			ID:   "card1",
			Name: "Test Card",
			Type: gamecards.CardTypeAutomated,
			Cost: 10,
			Tags: []shared.CardTag{shared.TagBuilding},
			Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
				{
					Type: gamecards.RequirementTemperature,
					Min:  intPtr(-10),
				},
			}},
		},
		{
			ID:   "card2",
			Name: "Expensive Card",
			Type: gamecards.CardTypeAutomated,
			Cost: 50,
			Tags: []shared.CardTag{shared.TagSpace},
		},
		{
			ID:   "card3",
			Name: "Card With Multiple Requirements",
			Type: gamecards.CardTypeAutomated,
			Cost: 15,
			Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
				{
					Type: gamecards.RequirementOxygen,
					Min:  intPtr(5),
				},
				{
					Type: gamecards.RequirementOceans,
					Min:  intPtr(3),
				},
			}},
		},
	}

	cardRegistry := cards.NewInMemoryCardRegistry(testCards)

	// Set game to action phase
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	return g, p, cardRegistry
}

func intPtr(val int) *int {
	return &val
}

// TestCalculatePlayerCardState_Available verifies that a playable card has no errors
func TestCalculatePlayerCardState_Available(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player enough credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Set temperature to meet requirement
	ctx := context.Background()
	g.GlobalParameters().SetTemperature(ctx, -10)

	// Get test card
	card, err := cardRegistry.GetByID("card1")
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	// Calculate state
	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	// Verify card is available
	if !state.Available() {
		t.Errorf("Expected card to be available, got errors: %+v", state.Errors)
	}

	if len(state.Errors) != 0 {
		t.Errorf("Expected no errors, got %d errors", len(state.Errors))
	}

	if len(state.Cost) == 0 {
		t.Fatal("Expected cost to be set")
	}

	if state.Cost["credit"] != 10 {
		t.Errorf("Expected cost 10, got %d", state.Cost["credit"])
	}
}

// TestCalculatePlayerCardState_InsufficientCredits verifies affordability check
func TestCalculatePlayerCardState_InsufficientCredits(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player insufficient credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 5,
	})

	// Get test card (cost 10)
	card, err := cardRegistry.GetByID("card1")
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	// Calculate state
	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	// Verify card is NOT available
	if state.Available() {
		t.Error("Expected card to be unavailable due to insufficient credits")
	}

	// Verify insufficient-credits error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientCredits && err.Category == player.ErrorCategoryCost {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected insufficient-credits error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardState_TemperatureRequirement verifies requirement validation
func TestCalculatePlayerCardState_TemperatureRequirement(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player enough credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Set temperature BELOW requirement
	ctx := context.Background()
	g.GlobalParameters().SetTemperature(ctx, -20)

	// Get test card (requires temp >= -10)
	card, err := cardRegistry.GetByID("card1")
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	// Calculate state
	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	// Verify card is NOT available
	if state.Available() {
		t.Error("Expected card to be unavailable due to temperature requirement")
	}

	// Verify temperature-too-low error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeTemperatureTooLow && err.Category == player.ErrorCategoryRequirement {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected temperature-too-low error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardState_MultipleRequirements verifies multiple requirement checks
func TestCalculatePlayerCardState_MultipleRequirements(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player enough credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Set global parameters to NOT meet requirements
	ctx := context.Background()
	g.GlobalParameters().SetOxygen(ctx, 2) // Need 5
	g.GlobalParameters().PlaceOcean(ctx)   // 1 ocean, need 3
	g.GlobalParameters().PlaceOcean(ctx)   // 2 oceans, need 3

	// Get test card (requires oxygen >= 5 AND oceans >= 3)
	card, err := cardRegistry.GetByID("card3")
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	// Calculate state
	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	// Verify card is NOT available
	if state.Available() {
		t.Error("Expected card to be unavailable due to multiple unmet requirements")
	}

	// Verify we have multiple errors (oxygen AND oceans)
	if len(state.Errors) < 2 {
		t.Errorf("Expected at least 2 errors, got %d: %+v", len(state.Errors), state.Errors)
	}

	// Check for both error types
	hasOxygenError := false
	hasOceanError := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeOxygenTooLow {
			hasOxygenError = true
		}
		if err.Code == player.ErrorCodeOceansTooLow {
			hasOceanError = true
		}
	}

	if !hasOxygenError {
		t.Error("Expected oxygen-too-low error")
	}
	if !hasOceanError {
		t.Error("Expected oceans-too-low error")
	}
}

// TestCalculatePlayerCardState_WrongPhase verifies phase validation
func TestCalculatePlayerCardState_WrongPhase(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Set game to WRONG phase (waiting for game start instead of action)
	ctx := context.Background()
	if err := g.UpdatePhase(ctx, game.GamePhaseWaitingForGameStart); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	// Get test card
	card, err := cardRegistry.GetByID("card1")
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	// Calculate state
	state := action.CalculatePlayerCardState(card, p, g, cardRegistry)

	// Verify card is NOT available
	if state.Available() {
		t.Error("Expected card to be unavailable due to wrong phase")
	}

	// Verify wrong-phase error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeWrongPhase && err.Category == player.ErrorCategoryPhase {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected wrong-phase error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardActionState_Available verifies usable action
func TestCalculatePlayerCardActionState_Available(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	// Set current turn to this player
	ctx := context.Background()
	if err := g.SetCurrentTurn(ctx, "player1", 2); err != nil {
		t.Fatalf("Failed to set current turn: %v", err)
	}

	// Give player resources for action inputs
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 10,
	})

	// Create test action with input requirements
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceSteel, Amount: 2},
		},
	}

	pca := player.NewPlayerCardAction("card1", 0, behavior)

	// Calculate state
	state := action.CalculatePlayerCardActionState("card1", behavior, pca.TimesUsedThisGeneration(), p, g)

	// Verify action is available
	if !state.Available() {
		t.Errorf("Expected action to be available, got errors: %+v", state.Errors)
	}

	if len(state.Errors) != 0 {
		t.Errorf("Expected no errors, got %d errors", len(state.Errors))
	}
}

// TestCalculatePlayerCardActionState_InsufficientResources verifies input check
func TestCalculatePlayerCardActionState_InsufficientResources(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	// Set current turn to this player
	ctx := context.Background()
	if err := g.SetCurrentTurn(ctx, "player1", 2); err != nil {
		t.Fatalf("Failed to set current turn: %v", err)
	}

	// Give player INSUFFICIENT resources
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 2, // Need 4
	})

	// Create test action requiring 4 energy
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 4},
		},
	}

	pca := player.NewPlayerCardAction("card1", 0, behavior)

	// Calculate state
	state := action.CalculatePlayerCardActionState("card1", behavior, pca.TimesUsedThisGeneration(), p, g)

	// Verify action is NOT available
	if state.Available() {
		t.Error("Expected action to be unavailable due to insufficient resources")
	}

	// Verify insufficient-resources error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientResources && err.Category == player.ErrorCategoryInput {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected insufficient-resources error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardActionState_NotPlayerTurn verifies turn check
func TestCalculatePlayerCardActionState_NotPlayerTurn(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	// Set current turn to DIFFERENT player
	ctx := context.Background()
	if err := g.SetCurrentTurn(ctx, "player2", 2); err != nil {
		t.Fatalf("Failed to set current turn: %v", err)
	}

	// Create test action
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
	}

	pca := player.NewPlayerCardAction("card1", 0, behavior)

	// Calculate state
	state := action.CalculatePlayerCardActionState("card1", behavior, pca.TimesUsedThisGeneration(), p, g)

	// Verify action is NOT available
	if state.Available() {
		t.Error("Expected action to be unavailable because it's not player's turn")
	}

	// Verify not-your-turn error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeNotYourTurn && err.Category == player.ErrorCategoryTurn {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected not-your-turn error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerStandardProjectState_Available verifies affordable project
func TestCalculatePlayerStandardProjectState_Available(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player enough credits for asteroid (cost 14)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Calculate state for asteroid project
	state := action.CalculatePlayerStandardProjectState(
		shared.StandardProjectAsteroid,
		p,
		g,
		cardRegistry,
	)

	// Verify project is available
	if !state.Available() {
		t.Errorf("Expected project to be available, got errors: %+v", state.Errors)
	}

	if len(state.Errors) != 0 {
		t.Errorf("Expected no errors, got %d errors", len(state.Errors))
	}

	expectedCost := shared.StandardProjectCost[shared.StandardProjectAsteroid]
	if len(state.Cost) == 0 {
		t.Fatal("Expected cost to be set")
	}
	if state.Cost["credit"] != expectedCost {
		t.Errorf("Expected cost %d, got %d", expectedCost, state.Cost["credit"])
	}
}

// TestCalculatePlayerStandardProjectState_InsufficientCredits verifies affordability
func TestCalculatePlayerStandardProjectState_InsufficientCredits(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player insufficient credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 5,
	})

	// Calculate state for asteroid project (cost 14)
	state := action.CalculatePlayerStandardProjectState(
		shared.StandardProjectAsteroid,
		p,
		g,
		cardRegistry,
	)

	// Verify project is NOT available
	if state.Available() {
		t.Error("Expected project to be unavailable due to insufficient credits")
	}

	// Verify insufficient-credits error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientCredits && err.Category == player.ErrorCategoryCost {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected insufficient-credits error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerStandardProjectState_NoOceansRemaining verifies ocean availability
func TestCalculatePlayerStandardProjectState_NoOceansRemaining(t *testing.T) {
	g, p, cardRegistry := setupTestEnvironment(t)

	// Give player enough credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Place all 9 oceans
	ctx := context.Background()
	for i := 0; i < 9; i++ {
		if _, err := g.GlobalParameters().PlaceOcean(ctx); err != nil {
			t.Fatalf("Failed to place ocean: %v", err)
		}
	}

	// Calculate state for aquifer project
	state := action.CalculatePlayerStandardProjectState(
		shared.StandardProjectAquifer,
		p,
		g,
		cardRegistry,
	)

	// Verify project is NOT available
	if state.Available() {
		t.Error("Expected aquifer project to be unavailable when no oceans remaining")
	}

	// Verify no-ocean-tiles error exists
	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeNoOceanTiles && err.Category == player.ErrorCategoryAvailability {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected no-ocean-tiles error, got errors: %+v", state.Errors)
	}

	// Verify metadata shows oceansRemaining = 0
	if oceansRemaining, ok := state.Metadata["oceansRemaining"].(int); !ok || oceansRemaining != 0 {
		t.Errorf("Expected oceansRemaining=0 in metadata, got %v", state.Metadata["oceansRemaining"])
	}
}

// TestCalculatePlayerCardState_SteelNotCountedForNonBuildingCard verifies that steel
// is NOT counted as a payment substitute for cards without the Building tag.
func TestCalculatePlayerCardState_SteelNotCountedForNonBuildingCard(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	// Card without Building tag, costs 20
	nonBuildingCard := &gamecards.Card{
		ID:   "non-building",
		Name: "Non Building Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 20,
		Tags: []shared.CardTag{shared.TagScience},
	}

	// Player has 5 credits + 10 steel (worth 20 MC for Building cards)
	// Without steel substitute, player can only afford 5 MC
	p.Resources().Set(shared.Resources{Credits: 5, Steel: 10})

	state := action.CalculatePlayerCardState(nonBuildingCard, p, g, cards.NewInMemoryCardRegistry(nil))

	if state.Available() {
		t.Error("Expected card to be unavailable: steel should NOT count for non-Building card")
	}

	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientCredits {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected insufficient-credits error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardState_SteelCountedForBuildingCard verifies that steel
// IS counted as a payment substitute for cards with the Building tag.
func TestCalculatePlayerCardState_SteelCountedForBuildingCard(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	buildingCard := &gamecards.Card{
		ID:   "building-card",
		Name: "Building Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 20,
		Tags: []shared.CardTag{shared.TagBuilding},
	}

	// Player has 5 credits + 10 steel (worth 20 MC) = 25 effective MC
	p.Resources().Set(shared.Resources{Credits: 5, Steel: 10})

	state := action.CalculatePlayerCardState(buildingCard, p, g, cards.NewInMemoryCardRegistry(nil))

	if !state.Available() {
		t.Errorf("Expected Building card to be affordable with steel, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardState_TitaniumNotCountedForNonSpaceCard verifies that titanium
// is NOT counted as a payment substitute for cards without the Space tag.
func TestCalculatePlayerCardState_TitaniumNotCountedForNonSpaceCard(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	nonSpaceCard := &gamecards.Card{
		ID:   "non-space",
		Name: "Non Space Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 20,
		Tags: []shared.CardTag{shared.TagPlant},
	}

	// Player has 5 credits + 5 titanium (worth 15 MC for Space cards)
	// Without titanium substitute, player can only afford 5 MC
	p.Resources().Set(shared.Resources{Credits: 5, Titanium: 5})

	state := action.CalculatePlayerCardState(nonSpaceCard, p, g, cards.NewInMemoryCardRegistry(nil))

	if state.Available() {
		t.Error("Expected card to be unavailable: titanium should NOT count for non-Space card")
	}

	found := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientCredits {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected insufficient-credits error, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardState_TitaniumCountedForSpaceCard verifies that titanium
// IS counted as a payment substitute for cards with the Space tag.
func TestCalculatePlayerCardState_TitaniumCountedForSpaceCard(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	spaceCard := &gamecards.Card{
		ID:   "space-card",
		Name: "Space Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 20,
		Tags: []shared.CardTag{shared.TagSpace},
	}

	// Player has 5 credits + 5 titanium (worth 15 MC) = 20 effective MC
	p.Resources().Set(shared.Resources{Credits: 5, Titanium: 5})

	state := action.CalculatePlayerCardState(spaceCard, p, g, cards.NewInMemoryCardRegistry(nil))

	if !state.Available() {
		t.Errorf("Expected Space card to be affordable with titanium, got errors: %+v", state.Errors)
	}
}

// TestCalculatePlayerCardState_HeatSubstituteWorksForAllCards verifies that Helion-style
// heat-to-credit substitutes work regardless of card tags.
func TestCalculatePlayerCardState_HeatSubstituteWorksForAllCards(t *testing.T) {
	g, p, _ := setupTestEnvironment(t)

	// Card with no Building/Space tags
	genericCard := &gamecards.Card{
		ID:   "generic-card",
		Name: "Generic Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 20,
		Tags: []shared.CardTag{shared.TagScience},
	}

	// Player has 5 credits + 20 heat, with heat-to-credit substitute (like Helion)
	p.Resources().Set(shared.Resources{Credits: 5, Heat: 20})
	p.Resources().AddPaymentSubstitute(shared.ResourceHeat, 1) // 1 heat = 1 MC

	state := action.CalculatePlayerCardState(genericCard, p, g, cards.NewInMemoryCardRegistry(nil))

	if !state.Available() {
		t.Errorf("Expected card to be affordable with heat substitute, got errors: %+v", state.Errors)
	}
}
