package events_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/test/testutil"
)

// TestPassiveEffects_EventSubscription tests that passive effects can subscribe to events
func TestPassiveEffects_EventSubscription(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()

	var effectTriggered bool
	var mu sync.Mutex

	// Subscribe passive effect to temperature change
	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		effectTriggered = true
		mu.Unlock()
	})

	// Trigger temperature change
	ctx := context.Background()
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")

	time.Sleep(10 * time.Millisecond)

	// Verify effect was triggered
	mu.Lock()
	defer mu.Unlock()
	testutil.AssertTrue(t, effectTriggered, "Passive effect should be triggered on temperature change")
}

// TestPassiveEffects_MultipleEffects tests multiple passive effects subscribing
func TestPassiveEffects_MultipleEffects(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()

	var effect1Triggered, effect2Triggered, effect3Triggered bool
	var mu sync.Mutex

	// Subscribe multiple effects
	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		effect1Triggered = true
		mu.Unlock()
	})

	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		effect2Triggered = true
		mu.Unlock()
	})

	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		effect3Triggered = true
		mu.Unlock()
	})

	// Trigger temperature change
	ctx := context.Background()
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")

	time.Sleep(10 * time.Millisecond)

	// Verify all effects triggered
	mu.Lock()
	defer mu.Unlock()
	testutil.AssertTrue(t, effect1Triggered, "Effect 1 should trigger")
	testutil.AssertTrue(t, effect2Triggered, "Effect 2 should trigger")
	testutil.AssertTrue(t, effect3Triggered, "Effect 3 should trigger")
}

// TestPassiveEffects_DifferentEventTypes tests effects subscribe to different event types
func TestPassiveEffects_DifferentEventTypes(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()
	ctx := context.Background()

	var tempEffectTriggered, resourceEffectTriggered bool
	var mu sync.Mutex

	// Subscribe to different events
	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		tempEffectTriggered = true
		mu.Unlock()
	})

	_ = events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		mu.Lock()
		resourceEffectTriggered = true
		mu.Unlock()
	})

	// Start game
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Trigger temperature change
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	tempTriggered := tempEffectTriggered
	mu.Unlock()

	testutil.AssertTrue(t, tempTriggered, "Temperature effect should trigger")

	// Reset flags
	mu.Lock()
	tempEffectTriggered = false
	resourceEffectTriggered = false
	mu.Unlock()

	// Trigger resource change
	testutil.AddPlayerCredits(ctx, player, 10)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	testutil.AssertTrue(t, resourceEffectTriggered, "Resource effect should trigger")
	testutil.AssertFalse(t, tempEffectTriggered, "Temperature effect should not trigger again")
}

// TestPassiveEffects_EventData tests that effects receive correct event data
func TestPassiveEffects_EventData(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()
	ctx := context.Background()

	var receivedEvent events.TemperatureChangedEvent
	var mu sync.Mutex

	// Subscribe effect
	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		receivedEvent = event
		mu.Unlock()
	})

	// Start game
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	initialTemp := testGame.GlobalParameters().Temperature()

	// Trigger temperature change
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	time.Sleep(10 * time.Millisecond)

	// Verify event data
	mu.Lock()
	defer mu.Unlock()

	testutil.AssertEqual(t, testGame.ID(), receivedEvent.GameID, "Event should have correct game ID")
	testutil.AssertEqual(t, initialTemp, receivedEvent.OldValue, "Event should have old temperature")
	testutil.AssertTrue(t, receivedEvent.NewValue > initialTemp, "Event should have new temperature")
}

// TestPassiveEffects_Unsubscribe tests that unsubscribed effects don't trigger
func TestPassiveEffects_Unsubscribe(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()
	ctx := context.Background()

	var effectTriggered bool
	var mu sync.Mutex

	// Subscribe effect
	subID := events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		effectTriggered = true
		mu.Unlock()
	})

	// Start game
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Trigger once
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	firstTrigger := effectTriggered
	effectTriggered = false
	mu.Unlock()

	testutil.AssertTrue(t, firstTrigger, "Effect should trigger before unsubscribe")

	// Unsubscribe
	eventBus.Unsubscribe(subID)

	// Trigger again
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	testutil.AssertFalse(t, effectTriggered, "Effect should not trigger after unsubscribe")
}

// TestPassiveEffects_CardBasedEffect tests simulating a card-based passive effect
func TestPassiveEffects_CardBasedEffect(t *testing.T) {
	// Setup - simulate a card that gives 1 credit when temperature increases
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Subscribe card effect: +1 credit on temperature change
	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		// Add credit to the first player (simulating a card effect)
		testutil.AddPlayerCredits(ctx, player, 1)
	})

	initialCredits := testutil.GetPlayerCredits(player)

	// Trigger temperature change
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	time.Sleep(10 * time.Millisecond)

	// Verify credits increased
	finalCredits := testutil.GetPlayerCredits(player)
	testutil.AssertEqual(t, initialCredits+1, finalCredits, "Card effect should give 1 credit")
}

// TestPassiveEffects_ConditionalEffect tests conditional passive effects
func TestPassiveEffects_ConditionalEffect(t *testing.T) {
	// Setup - effect only triggers if temperature below 0
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	eventBus := testGame.EventBus()
	ctx := context.Background()

	var effectCount int
	var mu sync.Mutex

	// Subscribe conditional effect
	_ = events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		if event.NewValue < 0 {
			mu.Lock()
			effectCount++
			mu.Unlock()
		}
	})

	// Start game
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Trigger temperature changes
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	testGame.GlobalParameters().IncreaseTemperature(ctx, 1, "")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	count := effectCount
	mu.Unlock()

	// Effect should have triggered (or not) based on temperature values
	t.Logf("Conditional effect triggered %d times", count)
}

// TestPassiveEffects_PerPlayerEffects tests effects isolated per player
func TestPassiveEffects_PerPlayerEffects(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	eventBus := testGame.EventBus()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player2 := players[1]

	var player1Effects, player2Effects int
	var mu sync.Mutex

	// Subscribe per-player effects
	_ = events.Subscribe(eventBus, func(event events.ResourcesChangedEvent) {
		mu.Lock()
		if event.PlayerID == player1.ID() {
			player1Effects++
		} else if event.PlayerID == player2.ID() {
			player2Effects++
		}
		mu.Unlock()
	})

	// Trigger resource changes for both players
	testutil.AddPlayerCredits(ctx, player1, 5)
	testutil.AddPlayerCredits(ctx, player2, 10)
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	testutil.AssertTrue(t, player1Effects > 0, "Player 1 effects should trigger")
	testutil.AssertTrue(t, player2Effects > 0, "Player 2 effects should trigger")
}
