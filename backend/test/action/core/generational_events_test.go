package core_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

func setupGenerationalEventsTestEnvironment(t *testing.T) (*game.Game, *player.Player) {
	logLevel := "error"
	if err := logger.Init(&logLevel); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	settings := game.GameSettings{
		MaxPlayers:      5,
		DevelopmentMode: true,
	}
	g := game.NewGame("test-game", "player1", settings)

	ctx := context.Background()
	p := player.NewPlayer(g.EventBus(), "test-game", "player1", "Test Player")
	if err := g.AddPlayer(ctx, p); err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		t.Fatalf("Failed to set game phase: %v", err)
	}

	return g, p
}

func TestGenerationalEventTracking_TRRaise(t *testing.T) {
	tests := []struct {
		name          string
		trRaises      int
		expectedCount int
	}{
		{
			name:          "single TR raise increments count to 1",
			trRaises:      1,
			expectedCount: 1,
		},
		{
			name:          "multiple TR raises accumulate",
			trRaises:      3,
			expectedCount: 3,
		},
		{
			name:          "no TR raises results in zero count",
			trRaises:      0,
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, p := setupGenerationalEventsTestEnvironment(t)

			for i := 0; i < tc.trRaises; i++ {
				events.Publish(g.EventBus(), events.TerraformRatingChangedEvent{
					GameID:    g.ID(),
					PlayerID:  p.ID(),
					OldRating: 20 + i,
					NewRating: 21 + i,
					Timestamp: time.Now(),
				})
			}

			count := p.GenerationalEvents().GetCount(shared.GenerationalEventTRRaise)
			if count != tc.expectedCount {
				t.Errorf("Expected TR raise count %d, got %d", tc.expectedCount, count)
			}
		})
	}
}

func TestGenerationalEventTracking_OceanPlacement(t *testing.T) {
	tests := []struct {
		name            string
		oceanPlacements int
		expectedCount   int
	}{
		{
			name:            "single ocean placement increments count to 1",
			oceanPlacements: 1,
			expectedCount:   1,
		},
		{
			name:            "multiple ocean placements accumulate",
			oceanPlacements: 2,
			expectedCount:   2,
		},
		{
			name:            "no ocean placements results in zero count",
			oceanPlacements: 0,
			expectedCount:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, p := setupGenerationalEventsTestEnvironment(t)

			for i := 0; i < tc.oceanPlacements; i++ {
				events.Publish(g.EventBus(), events.TilePlacedEvent{
					GameID:    g.ID(),
					PlayerID:  p.ID(),
					TileType:  "ocean",
					Q:         i,
					R:         0,
					S:         -i,
					Timestamp: time.Now(),
				})
			}

			count := p.GenerationalEvents().GetCount(shared.GenerationalEventOceanPlacement)
			if count != tc.expectedCount {
				t.Errorf("Expected ocean placement count %d, got %d", tc.expectedCount, count)
			}
		})
	}
}

func TestGenerationalEventTracking_GenerationAdvanceClearsEvents(t *testing.T) {
	tests := []struct {
		name               string
		initialTRRaises    int
		initialOceans      int
		advanceGeneration  bool
		expectedTRCount    int
		expectedOceanCount int
	}{
		{
			name:               "generation advance clears all events",
			initialTRRaises:    2,
			initialOceans:      1,
			advanceGeneration:  true,
			expectedTRCount:    0,
			expectedOceanCount: 0,
		},
		{
			name:               "events persist without generation advance",
			initialTRRaises:    2,
			initialOceans:      1,
			advanceGeneration:  false,
			expectedTRCount:    2,
			expectedOceanCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, p := setupGenerationalEventsTestEnvironment(t)

			for i := 0; i < tc.initialTRRaises; i++ {
				events.Publish(g.EventBus(), events.TerraformRatingChangedEvent{
					GameID:    g.ID(),
					PlayerID:  p.ID(),
					OldRating: 20 + i,
					NewRating: 21 + i,
					Timestamp: time.Now(),
				})
			}

			for i := 0; i < tc.initialOceans; i++ {
				events.Publish(g.EventBus(), events.TilePlacedEvent{
					GameID:    g.ID(),
					PlayerID:  p.ID(),
					TileType:  "ocean",
					Q:         i,
					R:         0,
					S:         -i,
					Timestamp: time.Now(),
				})
			}

			if tc.advanceGeneration {
				ctx := context.Background()
				if err := g.AdvanceGeneration(ctx); err != nil {
					t.Fatalf("Failed to advance generation: %v", err)
				}
			}

			trCount := p.GenerationalEvents().GetCount(shared.GenerationalEventTRRaise)
			if trCount != tc.expectedTRCount {
				t.Errorf("Expected TR raise count %d after generation advance, got %d", tc.expectedTRCount, trCount)
			}

			oceanCount := p.GenerationalEvents().GetCount(shared.GenerationalEventOceanPlacement)
			if oceanCount != tc.expectedOceanCount {
				t.Errorf("Expected ocean placement count %d after generation advance, got %d", tc.expectedOceanCount, oceanCount)
			}
		})
	}
}

func TestGenerationalEventRequirement_TRRaiseUnavailableWhenNotRaised(t *testing.T) {
	g, p := setupGenerationalEventsTestEnvironment(t)

	minOne := 1
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs:   []shared.ResourceCondition{{ResourceType: shared.ResourceCredit, Amount: 3}},
		Outputs:  []shared.ResourceCondition{{ResourceType: shared.ResourceTR, Amount: 1}},
		GenerationalEventRequirements: []shared.GenerationalEventRequirement{
			{
				Event: shared.GenerationalEventTRRaise,
				Count: &shared.MinMax{Min: &minOne},
			},
		},
	}

	state := action.CalculatePlayerCardActionState("test-card", behavior, 0, p, g)

	hasGenerationalEventError := false
	for _, stateErr := range state.Errors {
		if stateErr.Code == player.ErrorCodeGenerationalEventNotMet {
			hasGenerationalEventError = true
			break
		}
	}

	if !hasGenerationalEventError {
		t.Errorf("Expected generational event requirement error when TR has not been raised, got errors: %v", state.Errors)
	}
}

func TestGenerationalEventRequirement_TRRaiseAvailableWhenRaised(t *testing.T) {
	g, p := setupGenerationalEventsTestEnvironment(t)

	events.Publish(g.EventBus(), events.TerraformRatingChangedEvent{
		GameID:    g.ID(),
		PlayerID:  p.ID(),
		OldRating: 20,
		NewRating: 21,
		Timestamp: time.Now(),
	})

	minOne := 1
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs:   []shared.ResourceCondition{{ResourceType: shared.ResourceCredit, Amount: 3}},
		Outputs:  []shared.ResourceCondition{{ResourceType: shared.ResourceTR, Amount: 1}},
		GenerationalEventRequirements: []shared.GenerationalEventRequirement{
			{
				Event: shared.GenerationalEventTRRaise,
				Count: &shared.MinMax{Min: &minOne},
			},
		},
	}

	state := action.CalculatePlayerCardActionState("test-card", behavior, 0, p, g)

	for _, stateErr := range state.Errors {
		if stateErr.Code == player.ErrorCodeGenerationalEventNotMet {
			t.Errorf("Expected no generational event requirement error when TR has been raised, but got: %v", stateErr)
		}
	}
}

func TestGenerationalEventRequirement_OceanPlacementMinTwoUnavailableWithOne(t *testing.T) {
	g, p := setupGenerationalEventsTestEnvironment(t)

	events.Publish(g.EventBus(), events.TilePlacedEvent{
		GameID:    g.ID(),
		PlayerID:  p.ID(),
		TileType:  "ocean",
		Q:         0,
		R:         0,
		S:         0,
		Timestamp: time.Now(),
	})

	minTwo := 2
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs:   []shared.ResourceCondition{{ResourceType: shared.ResourceCredit, Amount: 5}},
		Outputs:  []shared.ResourceCondition{{ResourceType: shared.ResourceTR, Amount: 1}},
		GenerationalEventRequirements: []shared.GenerationalEventRequirement{
			{
				Event: shared.GenerationalEventOceanPlacement,
				Count: &shared.MinMax{Min: &minTwo},
			},
		},
	}

	state := action.CalculatePlayerCardActionState("test-card", behavior, 0, p, g)

	hasGenerationalEventError := false
	for _, stateErr := range state.Errors {
		if stateErr.Code == player.ErrorCodeGenerationalEventNotMet {
			hasGenerationalEventError = true
			break
		}
	}

	if !hasGenerationalEventError {
		t.Errorf("Expected generational event requirement error when only 1 ocean placed but min 2 required, got errors: %v", state.Errors)
	}
}
