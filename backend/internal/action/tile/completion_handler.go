package tile

import (
	"context"
	"fmt"

	baseaction "openmars/internal/action"
	"openmars/internal/game"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
)

// Callback types for tile completion
const (
	CallbackConvertPlantsToGreenery = "convert-plants-to-greenery"
	CallbackStandardProjectGreenery = "standard-project-greenery"
	CallbackStandardProjectAquifer  = "standard-project-aquifer"
	CallbackAdjacentRemoval         = "adjacent-removal"
)

// TileCompletionHandlerFunc is the signature for tile completion callbacks
type TileCompletionHandlerFunc func(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *shared.TileCompletionCallback) error

// TileCompletionRegistry holds registered completion handlers
type TileCompletionRegistry struct {
	handlers  map[string]TileCompletionHandlerFunc
	stateRepo game.GameStateRepository
	registry  gamecards.CardRegistry
}

// NewTileCompletionRegistry creates a new registry with default handlers
func NewTileCompletionRegistry(stateRepo game.GameStateRepository, registry gamecards.CardRegistry) *TileCompletionRegistry {
	r := &TileCompletionRegistry{
		handlers:  make(map[string]TileCompletionHandlerFunc),
		stateRepo: stateRepo,
		registry:  registry,
	}
	r.registerDefaultHandlers()
	return r
}

func (r *TileCompletionRegistry) registerDefaultHandlers() {
	r.handlers[CallbackConvertPlantsToGreenery] = r.handleConvertPlantsToGreenery
	r.handlers[CallbackStandardProjectGreenery] = r.handleStandardProjectGreenery
	r.handlers[CallbackStandardProjectAquifer] = r.handleStandardProjectAquifer
	r.handlers[CallbackAdjacentRemoval] = r.handleAdjacentRemoval
}

// Handle invokes the appropriate handler for the callback type
// If no callback is registered, no log is created - only use cases that explicitly register callbacks get logs
func (r *TileCompletionRegistry) Handle(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *shared.TileCompletionCallback) error {
	if callback == nil {
		return nil
	}

	handler, exists := r.handlers[callback.Type]
	if !exists {
		return nil
	}

	return handler(ctx, g, playerID, result, callback)
}

func (r *TileCompletionRegistry) handleConvertPlantsToGreenery(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, _ *shared.TileCompletionCallback) error {
	outputs := []shared.CalculatedOutput{
		{ResourceType: string(shared.ResourceGreeneryPlacement), Amount: 1, IsScaled: false},
	}
	if result.OxygenSteps > 0 {
		outputs = append(outputs, shared.CalculatedOutput{ResourceType: string(shared.ResourceOxygen), Amount: result.OxygenSteps, IsScaled: false})
	}
	if result.TRGained > 0 {
		outputs = append(outputs, shared.CalculatedOutput{ResourceType: string(shared.ResourceTR), Amount: result.TRGained, IsScaled: false})
	}

	displayData := baseaction.GetStandardProjectDisplayData("Convert Plants")
	_, err := r.stateRepo.WriteFull(ctx, g.ID(), g, "Convert Plants", shared.SourceTypeResourceConvert, playerID, "Converted plants to greenery", nil, outputs, displayData)
	return err
}

func (r *TileCompletionRegistry) handleStandardProjectGreenery(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, _ *shared.TileCompletionCallback) error {
	outputs := []shared.CalculatedOutput{
		{ResourceType: string(shared.ResourceGreeneryPlacement), Amount: 1, IsScaled: false},
	}
	if result.OxygenSteps > 0 {
		outputs = append(outputs, shared.CalculatedOutput{ResourceType: string(shared.ResourceOxygen), Amount: result.OxygenSteps, IsScaled: false})
	}
	if result.TRGained > 0 {
		outputs = append(outputs, shared.CalculatedOutput{ResourceType: string(shared.ResourceTR), Amount: result.TRGained, IsScaled: false})
	}

	displayData := baseaction.GetStandardProjectDisplayData("Standard Project: Greenery")
	_, err := r.stateRepo.WriteFull(ctx, g.ID(), g, "Standard Project: Greenery", shared.SourceTypeStandardProject, playerID, "Planted greenery", nil, outputs, displayData)
	return err
}

func (r *TileCompletionRegistry) handleStandardProjectAquifer(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, _ *shared.TileCompletionCallback) error {
	outputs := []shared.CalculatedOutput{
		{ResourceType: string(shared.ResourceOceanPlacement), Amount: 1, IsScaled: false},
	}
	if result.TRGained > 0 {
		outputs = append(outputs, shared.CalculatedOutput{ResourceType: string(shared.ResourceTR), Amount: result.TRGained, IsScaled: false})
	}

	displayData := baseaction.GetStandardProjectDisplayData("Standard Project: Aquifer")
	_, err := r.stateRepo.WriteFull(ctx, g.ID(), g, "Standard Project: Aquifer", shared.SourceTypeStandardProject, playerID, "Built aquifer", nil, outputs, displayData)
	return err
}

func (r *TileCompletionRegistry) handleAdjacentRemoval(_ context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *shared.TileCompletionCallback) error {
	output, ok := callback.Data["output"].(*shared.BasicResourceCondition)
	if !ok {
		return fmt.Errorf("missing resource removal output")
	}
	coords, err := parseHexPosition(result.Hex)
	if err != nil {
		return err
	}
	p, err := g.GetPlayer(playerID)
	if err != nil {
		return err
	}
	source, _ := callback.Data["source"].(string)
	sourceCardID, _ := callback.Data["sourceCardID"].(string)
	return gamecards.QueueResourceRemoval(g, p, output, coords, sourceCardID, source, r.registry)
}
