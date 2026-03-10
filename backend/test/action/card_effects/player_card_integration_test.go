package card_effects_test

import (
	"context"
	"terraforming-mars-backend/test/testutil"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// TestPlayerCard_EventDrivenStateUpdate verifies state updates on domain events
func TestPlayerCard_EventDrivenStateUpdate(t *testing.T) {
	// Initialize logger
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test game
	settings := game.GameSettings{MaxPlayers: 5, DevelopmentMode: true}
	g := game.NewGame("test-game", "player1", settings)

	// Add player
	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Set game to action phase
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	// Create card with temperature requirement
	card := &gamecards.Card{
		ID:   "test-card",
		Name: "Test Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-10)},
		}},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{*card})

	// Give player enough credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Set temperature BELOW requirement
	g.GlobalParameters().SetTemperature(ctx, -20)

	// Create PlayerCard with event listeners
	pc := action.CreateAndCachePlayerCard(card, p, g, cardRegistry)

	// Initial state should be unavailable (temperature too low)
	state := pc.State()
	if state.Available() {
		t.Error("Expected card to be unavailable initially")
	}

	// Verify temperature-too-low error
	foundTempError := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeTemperatureTooLow {
			foundTempError = true
			break
		}
	}
	if !foundTempError {
		t.Errorf("Expected %s error initially, got errors: %+v", player.ErrorCodeTemperatureTooLow, state.Errors)
	}

	// TRIGGER EVENT: Increase temperature to meet requirement
	_, err := g.GlobalParameters().IncreaseTemperature(ctx, 10, "")
	if err != nil {
		t.Fatalf("Failed to increase temperature: %v", err)
	}

	// Give event handlers time to execute (synchronous but need to yield)
	time.Sleep(10 * time.Millisecond)

	// State should now be updated automatically (temperature requirement met)
	updatedState := pc.State()
	if !updatedState.Available() {
		t.Errorf("Expected card to be available after temperature increase, errors: %+v", updatedState.Errors)
	}

	// Verify no temperature error
	for _, err := range updatedState.Errors {
		if err.Code == player.ErrorCodeTemperatureTooLow {
			t.Errorf("Expected no %s error after temperature increase", player.ErrorCodeTemperatureTooLow)
		}
	}
}

// TestPlayerCard_ResourceChangeEventUpdate verifies state updates on resource changes
func TestPlayerCard_ResourceChangeEventUpdate(t *testing.T) {
	// Initialize logger
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test game
	settings := game.GameSettings{MaxPlayers: 5, DevelopmentMode: true}
	g := game.NewGame("test-game", "player1", settings)

	// Add player
	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Set game to action phase
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	// Create expensive card
	card := &gamecards.Card{
		ID:   "expensive-card",
		Name: "Expensive Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 50,
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{*card})

	// Give player insufficient credits initially
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 30,
	})

	// Create PlayerCard with event listeners
	pc := action.CreateAndCachePlayerCard(card, p, g, cardRegistry)

	// Initial state should be unavailable (insufficient credits)
	state := pc.State()
	if state.Available() {
		t.Error("Expected card to be unavailable with insufficient credits")
	}

	// Verify insufficient-credits error
	foundCreditError := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeInsufficientCredits {
			foundCreditError = true
			break
		}
	}
	if !foundCreditError {
		t.Errorf("Expected %s error initially, got errors: %+v", player.ErrorCodeInsufficientCredits, state.Errors)
	}

	// TRIGGER EVENT: Add credits to player
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 25,
	})

	// Give event handlers time to execute
	time.Sleep(10 * time.Millisecond)

	// State should now be updated automatically (affordable)
	updatedState := pc.State()
	if !updatedState.Available() {
		t.Errorf("Expected card to be available after gaining credits, errors: %+v", updatedState.Errors)
	}

	// Verify no credit error
	for _, err := range updatedState.Errors {
		if err.Code == player.ErrorCodeInsufficientCredits {
			t.Errorf("Expected no %s error after gaining credits", player.ErrorCodeInsufficientCredits)
		}
	}
}

// TestPlayerCard_PhaseChangeEventUpdate verifies state updates on phase changes
func TestPlayerCard_PhaseChangeEventUpdate(t *testing.T) {
	// Initialize logger
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test game
	settings := game.GameSettings{MaxPlayers: 5, DevelopmentMode: true}
	g := game.NewGame("test-game", "player1", settings)

	// Add player
	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Start in WRONG phase
	if err := g.UpdatePhase(ctx, game.GamePhaseWaitingForGameStart); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	// Create simple card
	card := &gamecards.Card{
		ID:   "test-card",
		Name: "Test Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{*card})

	// Give player credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Create PlayerCard with event listeners
	pc := action.CreateAndCachePlayerCard(card, p, g, cardRegistry)

	// Initial state should be unavailable (wrong phase)
	state := pc.State()
	if state.Available() {
		t.Error("Expected card to be unavailable in wrong phase")
	}

	// Verify wrong-phase error
	foundPhaseError := false
	for _, err := range state.Errors {
		if err.Code == player.ErrorCodeWrongPhase {
			foundPhaseError = true
			break
		}
	}
	if !foundPhaseError {
		t.Errorf("Expected %s error initially, got errors: %+v", player.ErrorCodeWrongPhase, state.Errors)
	}

	// TRIGGER EVENT: Change phase to action
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to update phase: %v", err)
	}

	// Give event handlers time to execute
	time.Sleep(10 * time.Millisecond)

	// State should now be updated automatically (correct phase)
	updatedState := pc.State()
	if !updatedState.Available() {
		t.Errorf("Expected card to be available in action phase, errors: %+v", updatedState.Errors)
	}

	// Verify no phase error
	for _, err := range updatedState.Errors {
		if err.Code == player.ErrorCodeWrongPhase {
			t.Errorf("Expected no %s error in action phase", player.ErrorCodeWrongPhase)
		}
	}
}

// TestPlayerCard_CleanupPreventsMemoryLeak verifies event listener cleanup
func TestPlayerCard_CleanupPreventsMemoryLeak(t *testing.T) {
	// Initialize logger
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test game
	settings := game.GameSettings{MaxPlayers: 5, DevelopmentMode: true}
	g := game.NewGame("test-game", "player1", settings)

	// Add player
	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Set game to action phase
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	// Create card
	card := &gamecards.Card{
		ID:   "test-card",
		Name: "Test Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{*card})

	// Give player credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	// Create PlayerCard and add to hand
	pc := action.CreateAndCachePlayerCard(card, p, g, cardRegistry)
	p.Hand().AddCard(card.ID)
	p.Hand().AddPlayerCard(card.ID, pc)

	// Get initial state timestamp
	initialState := pc.State()
	initialTimestamp := initialState.LastCalculated

	// Remove card from hand (should trigger cleanup)
	removed := p.Hand().RemoveCard(card.ID)
	if !removed {
		t.Fatal("Failed to remove card from hand")
	}

	// Give time for cleanup to complete
	time.Sleep(5 * time.Millisecond)

	// Verify state is no longer updated after cleanup
	// Change temperature (would normally trigger state recalculation)
	g.GlobalParameters().SetTemperature(ctx, 10)

	// Give time for any handlers to execute
	time.Sleep(10 * time.Millisecond)

	// State should NOT have changed (listeners were cleaned up)
	// LastCalculated timestamp should still be the initial one
	stateAfterCleanup := pc.State()
	if stateAfterCleanup.LastCalculated.After(initialTimestamp) {
		t.Errorf("State was recalculated after cleanup (should not happen). Initial: %v, After: %v",
			initialTimestamp, stateAfterCleanup.LastCalculated)
	}

	// Also try triggering resource change (another event type)
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	time.Sleep(10 * time.Millisecond)

	// Still should not have recalculated
	finalState := pc.State()
	if finalState.LastCalculated.After(initialTimestamp) {
		t.Error("State was recalculated after cleanup on resource change (should not happen)")
	}
}

// TestPlayerCard_MultipleCardsIndependentState verifies each card has independent state
func TestPlayerCard_MultipleCardsIndependentState(t *testing.T) {
	// Initialize logger
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create test game
	settings := game.GameSettings{MaxPlayers: 5, DevelopmentMode: true}
	g := game.NewGame("test-game", "player1", settings)

	// Add player
	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Set game to action phase
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	// Create two cards with different requirements
	card1 := &gamecards.Card{
		ID:   "card1",
		Name: "Low Temp Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(-10)},
		}},
	}

	card2 := &gamecards.Card{
		ID:   "card2",
		Name: "High Temp Card",
		Type: gamecards.CardTypeAutomated,
		Cost: 10,
		Requirements: &gamecards.CardRequirements{Items: []gamecards.Requirement{
			{Type: gamecards.RequirementTemperature, Min: testutil.IntPtr(10)},
		}},
	}

	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{*card1, *card2})

	// Give player credits
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 30,
	})

	// Set temperature to 0 (meets card1 requirement, not card2)
	g.GlobalParameters().SetTemperature(ctx, 0)

	// Create both PlayerCards
	pc1 := action.CreateAndCachePlayerCard(card1, p, g, cardRegistry)
	pc2 := action.CreateAndCachePlayerCard(card2, p, g, cardRegistry)

	// Verify card1 is available (temp >= -10)
	state1 := pc1.State()
	if !state1.Available() {
		t.Errorf("Expected card1 to be available, errors: %+v", state1.Errors)
	}

	// Verify card2 is NOT available (temp < 10)
	state2 := pc2.State()
	if state2.Available() {
		t.Error("Expected card2 to be unavailable")
	}

	// Verify card2 has temperature error
	foundTempError := false
	for _, err := range state2.Errors {
		if err.Code == player.ErrorCodeTemperatureTooLow {
			foundTempError = true
			break
		}
	}
	if !foundTempError {
		t.Errorf("Expected card2 to have %s error, got errors: %+v", player.ErrorCodeTemperatureTooLow, state2.Errors)
	}

	// Increase temperature to 10 (now both cards should be available)
	g.GlobalParameters().SetTemperature(ctx, 10)
	time.Sleep(10 * time.Millisecond)

	// Both cards should now be available
	updatedState1 := pc1.State()
	updatedState2 := pc2.State()

	if !updatedState1.Available() {
		t.Errorf("Expected card1 to still be available, errors: %+v", updatedState1.Errors)
	}

	if !updatedState2.Available() {
		t.Errorf("Expected card2 to now be available, errors: %+v", updatedState2.Errors)
	}
}
