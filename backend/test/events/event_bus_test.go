package events_test

import (
	"sync"
	"testing"
	"time"

	"openmars/internal/events"
)

// TestEventBusSubscribeAndPublish tests basic subscribe and publish functionality
func TestEventBusSubscribeAndPublish(t *testing.T) {
	bus := events.NewEventBus()

	// Track if handler was called
	var handlerCalled bool
	var receivedEvent events.TemperatureChangedEvent

	// Subscribe to TemperatureChangedEvent
	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		handlerCalled = true
		receivedEvent = event
	})

	// Publish event
	testEvent := events.TemperatureChangedEvent{
		GameID:    "game-123",
		OldValue:  -30,
		NewValue:  -28,
		ChangedBy: "player-456",
		Timestamp: time.Now(),
	}
	events.Publish(bus, testEvent)

	// Verify handler was called
	if !handlerCalled {
		t.Error("Expected handler to be called, but it wasn't")
	}

	// Verify event data
	if receivedEvent.GameID != testEvent.GameID {
		t.Errorf("Expected GameID %s, got %s", testEvent.GameID, receivedEvent.GameID)
	}
	if receivedEvent.OldValue != testEvent.OldValue {
		t.Errorf("Expected OldValue %d, got %d", testEvent.OldValue, receivedEvent.OldValue)
	}
	if receivedEvent.NewValue != testEvent.NewValue {
		t.Errorf("Expected NewValue %d, got %d", testEvent.NewValue, receivedEvent.NewValue)
	}
}

// TestEventBusMultipleSubscribers tests that multiple subscribers receive the same event
func TestEventBusMultipleSubscribers(t *testing.T) {
	bus := events.NewEventBus()

	// Track calls for each handler
	var handler1Called, handler2Called, handler3Called bool

	// Subscribe multiple handlers
	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		handler1Called = true
	})

	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		handler2Called = true
	})

	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		handler3Called = true
	})

	// Publish event
	testEvent := events.TemperatureChangedEvent{
		GameID:   "game-123",
		OldValue: -30,
		NewValue: -28,
	}
	events.Publish(bus, testEvent)

	// Verify all handlers were called
	if !handler1Called {
		t.Error("Handler 1 was not called")
	}
	if !handler2Called {
		t.Error("Handler 2 was not called")
	}
	if !handler3Called {
		t.Error("Handler 3 was not called")
	}
}

// TestEventBusTypeSafety tests that only matching event types trigger handlers
func TestEventBusTypeSafety(t *testing.T) {
	bus := events.NewEventBus()

	var tempHandlerCalled bool
	var oxygenHandlerCalled bool

	// Subscribe to different event types
	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		tempHandlerCalled = true
	})

	_ = events.Subscribe(bus, func(event events.OxygenChangedEvent) {
		oxygenHandlerCalled = true
	})

	// Publish TemperatureChangedEvent
	events.Publish(bus, events.TemperatureChangedEvent{
		GameID:   "game-123",
		OldValue: -30,
		NewValue: -28,
	})

	// Only temperature handler should be called
	if !tempHandlerCalled {
		t.Error("Temperature handler was not called")
	}
	if oxygenHandlerCalled {
		t.Error("Oxygen handler should not have been called")
	}

	// Reset flags
	tempHandlerCalled = false
	oxygenHandlerCalled = false

	// Publish OxygenChangedEvent
	events.Publish(bus, events.OxygenChangedEvent{
		GameID:   "game-123",
		OldValue: 0,
		NewValue: 2,
	})

	// Only oxygen handler should be called
	if tempHandlerCalled {
		t.Error("Temperature handler should not have been called")
	}
	if !oxygenHandlerCalled {
		t.Error("Oxygen handler was not called")
	}
}

// TestEventBusUnsubscribe tests that unsubscribed handlers don't receive events
func TestEventBusUnsubscribe(t *testing.T) {
	bus := events.NewEventBus()

	var handlerCalled bool

	// Subscribe and get subscription ID
	subID := events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		handlerCalled = true
	})

	// Publish event - handler should be called
	events.Publish(bus, events.TemperatureChangedEvent{
		GameID:   "game-123",
		OldValue: -30,
		NewValue: -28,
	})

	if !handlerCalled {
		t.Error("Handler was not called before unsubscribe")
	}

	// Unsubscribe
	bus.Unsubscribe(subID)

	// Reset flag
	handlerCalled = false

	// Publish event again - handler should NOT be called
	events.Publish(bus, events.TemperatureChangedEvent{
		GameID:   "game-123",
		OldValue: -28,
		NewValue: -26,
	})

	if handlerCalled {
		t.Error("Handler was called after unsubscribe")
	}
}

// TestEventBusConcurrentPublish tests thread-safety with concurrent publishers
func TestEventBusConcurrentPublish(t *testing.T) {
	bus := events.NewEventBus()

	// Track received events
	var mu sync.Mutex
	receivedEvents := make([]events.TemperatureChangedEvent, 0)

	// Subscribe handler
	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	})

	// Publish events concurrently
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			events.Publish(bus, events.TemperatureChangedEvent{
				GameID:   "game-123",
				OldValue: index,
				NewValue: index + 2,
			})
		}(i)
	}

	wg.Wait()

	// Verify all events were received
	mu.Lock()
	defer mu.Unlock()

	if len(receivedEvents) != numGoroutines {
		t.Errorf("Expected %d events, got %d", numGoroutines, len(receivedEvents))
	}
}

// TestEventBusConcurrentSubscribe tests thread-safety with concurrent subscriptions
func TestEventBusConcurrentSubscribe(t *testing.T) {
	bus := events.NewEventBus()

	const numSubscribers = 10
	var wg sync.WaitGroup
	wg.Add(numSubscribers)

	// Track calls
	var mu sync.Mutex
	callCounts := make(map[int]int)

	// Subscribe concurrently
	for i := 0; i < numSubscribers; i++ {
		go func(index int) {
			defer wg.Done()
			_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
				mu.Lock()
				callCounts[index]++
				mu.Unlock()
			})
		}(i)
	}

	wg.Wait()

	// Publish event
	events.Publish(bus, events.TemperatureChangedEvent{
		GameID:   "game-123",
		OldValue: -30,
		NewValue: -28,
	})

	// Verify all handlers were called
	mu.Lock()
	defer mu.Unlock()

	if len(callCounts) != numSubscribers {
		t.Errorf("Expected %d handlers to be called, got %d", numSubscribers, len(callCounts))
	}

	for i := 0; i < numSubscribers; i++ {
		if callCounts[i] != 1 {
			t.Errorf("Handler %d was called %d times, expected 1", i, callCounts[i])
		}
	}
}

// TestEventBusDifferentEventTypes tests multiple different event types
func TestEventBusDifferentEventTypes(t *testing.T) {
	bus := events.NewEventBus()

	var tempEventReceived events.TemperatureChangedEvent
	var resourceEventReceived events.ResourcesChangedEvent
	var trEventReceived events.TerraformRatingChangedEvent

	_ = events.Subscribe(bus, func(event events.TemperatureChangedEvent) {
		tempEventReceived = event
	})

	_ = events.Subscribe(bus, func(event events.ResourcesChangedEvent) {
		resourceEventReceived = event
	})

	_ = events.Subscribe(bus, func(event events.TerraformRatingChangedEvent) {
		trEventReceived = event
	})

	// Publish different event types
	events.Publish(bus, events.TemperatureChangedEvent{
		GameID:   "game-123",
		NewValue: -26,
	})

	events.Publish(bus, events.ResourcesChangedEvent{
		GameID:       "game-123",
		PlayerID:     "player-456",
		ResourceType: "plants",
		NewAmount:    5,
	})

	events.Publish(bus, events.TerraformRatingChangedEvent{
		GameID:    "game-123",
		PlayerID:  "player-456",
		NewRating: 25,
	})

	// Verify all events were received correctly
	if tempEventReceived.NewValue != -26 {
		t.Errorf("Temperature event not received correctly")
	}
	if resourceEventReceived.ResourceType != "plants" || resourceEventReceived.NewAmount != 5 {
		t.Errorf("Resource event not received correctly")
	}
	if trEventReceived.NewRating != 25 {
		t.Errorf("TR event not received correctly")
	}
}
