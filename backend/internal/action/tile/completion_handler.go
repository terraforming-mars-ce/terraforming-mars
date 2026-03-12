package tile

import (
	"context"
	"fmt"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// Callback types for tile completion
const (
	CallbackConvertPlantsToGreenery = "convert-plants-to-greenery"
	CallbackStandardProjectGreenery = "standard-project-greenery"
	CallbackStandardProjectAquifer  = "standard-project-aquifer"
	CallbackAdjacentSteal           = "adjacent-steal"
)

// TileCompletionHandlerFunc is the signature for tile completion callbacks
type TileCompletionHandlerFunc func(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *shared.TileCompletionCallback) error

// TileCompletionRegistry holds registered completion handlers
type TileCompletionRegistry struct {
	handlers  map[string]TileCompletionHandlerFunc
	stateRepo game.GameStateRepository
}

// NewTileCompletionRegistry creates a new registry with default handlers
func NewTileCompletionRegistry(stateRepo game.GameStateRepository) *TileCompletionRegistry {
	r := &TileCompletionRegistry{
		handlers:  make(map[string]TileCompletionHandlerFunc),
		stateRepo: stateRepo,
	}
	r.registerDefaultHandlers()
	return r
}

func (r *TileCompletionRegistry) registerDefaultHandlers() {
	r.handlers[CallbackConvertPlantsToGreenery] = r.handleConvertPlantsToGreenery
	r.handlers[CallbackStandardProjectGreenery] = r.handleStandardProjectGreenery
	r.handlers[CallbackStandardProjectAquifer] = r.handleStandardProjectAquifer
	r.handlers[CallbackAdjacentSteal] = r.handleAdjacentSteal
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

func (r *TileCompletionRegistry) handleAdjacentSteal(_ context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *shared.TileCompletionCallback) error {
	amount, _ := callback.Data["amount"].(int)
	source, _ := callback.Data["source"].(string)
	sourceCardID, _ := callback.Data["sourceCardID"].(string)

	coords, err := parseHexPosition(result.Hex)
	if err != nil {
		return fmt.Errorf("failed to parse hex for adjacent steal: %w", err)
	}

	neighbors := coords.GetNeighbors()
	eligiblePlayerIDs := make(map[string]bool)

	for _, neighbor := range neighbors {
		neighborTile, tileErr := g.Board().GetTile(neighbor)
		if tileErr != nil {
			continue
		}
		if neighborTile.OwnerID != nil && *neighborTile.OwnerID != playerID {
			eligiblePlayerIDs[*neighborTile.OwnerID] = true
		}
	}

	if len(eligiblePlayerIDs) == 0 {
		return nil
	}

	ids := make([]string, 0, len(eligiblePlayerIDs))
	for id := range eligiblePlayerIDs {
		ids = append(ids, id)
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		return fmt.Errorf("player not found: %w", err)
	}

	p.Selection().SetPendingStealTargetSelection(&shared.PendingStealTargetSelection{
		EligiblePlayerIDs: ids,
		ResourceType:      shared.ResourceCredit,
		Amount:            amount,
		Source:            source,
		SourceCardID:      sourceCardID,
	})

	return nil
}
