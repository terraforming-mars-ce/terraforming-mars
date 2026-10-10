package core

import (
	"fmt"
	"sync"

	"openmars/internal/delivery/dto"
)

// ActionRegistry manages the registration and lookup of action handlers
type ActionRegistry struct {
	mu       sync.RWMutex
	handlers map[dto.ActionType]ActionHandler
}

// NewActionRegistry creates a new action registry
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		handlers: make(map[dto.ActionType]ActionHandler),
	}
}

// Register registers an action handler for a specific action type
func (ar *ActionRegistry) Register(actionType dto.ActionType, handler ActionHandler) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.handlers[actionType] = handler
}

// GetHandler retrieves a handler for the specified action type
func (ar *ActionRegistry) GetHandler(actionType dto.ActionType) (ActionHandler, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	handler, exists := ar.handlers[actionType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for action type: %s", actionType)
	}

	return handler, nil
}
