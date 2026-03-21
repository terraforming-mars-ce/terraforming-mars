package confirmation

import (
	"context"
	"fmt"
	"slices"

	"go.uber.org/zap"

	baseaction "terraforming-mars-backend/internal/action"
	colonyaction "terraforming-mars-backend/internal/action/colony"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/game"
)

// ConfirmColonyPlacementAction handles confirming a colony placement from a card effect
type ConfirmColonyPlacementAction struct {
	baseaction.BaseAction
	colonyRegistry colonies.ColonyRegistry
}

// NewConfirmColonyPlacementAction creates a new confirm colony placement action
func NewConfirmColonyPlacementAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	colonyRegistry colonies.ColonyRegistry,
	logger *zap.Logger,
) *ConfirmColonyPlacementAction {
	return &ConfirmColonyPlacementAction{
		BaseAction:     baseaction.NewBaseAction(gameRepo, cardRegistry),
		colonyRegistry: colonyRegistry,
	}
}

// Execute places the selected colony for free and clears the pending selection
func (a *ConfirmColonyPlacementAction) Execute(ctx context.Context, gameID string, playerID string, colonyID string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_colony_placement"),
		zap.String("colony_id", colonyID),
	)
	log.Debug("Confirming colony placement")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	pending := p.Selection().GetPendingColonySelection()
	if pending == nil {
		return fmt.Errorf("no pending colony selection")
	}

	if !slices.Contains(pending.AvailableColonyIDs, colonyID) {
		return fmt.Errorf("colony %s is not available for selection", colonyID)
	}

	tileState := g.GetColonyTileState(colonyID)
	if tileState == nil {
		return fmt.Errorf("colony tile not found: %s", colonyID)
	}

	definition, err := a.colonyRegistry.GetByID(colonyID)
	if err != nil {
		return fmt.Errorf("colony definition not found: %w", err)
	}

	maxColonies := len(definition.Colonies)
	if len(tileState.PlayerColonies) >= maxColonies {
		return fmt.Errorf("colony tile is full: %d/%d colonies", len(tileState.PlayerColonies), maxColonies)
	}

	if !pending.AllowDuplicatePlayerColony {
		for _, pid := range tileState.PlayerColonies {
			if pid == playerID {
				return fmt.Errorf("player already has a colony on this tile")
			}
		}
	}

	source := pending.Source
	if source == "" {
		source = "Card Effect"
	}

	if err := colonyaction.PlaceColonyOnTile(ctx, g, p, definition, tileState, a.CardRegistry(), source, log); err != nil {
		return fmt.Errorf("failed to place colony: %w", err)
	}

	p.Selection().SetPendingColonySelection(nil)

	forcedAction := g.GetForcedFirstAction(playerID)
	if forcedAction != nil && forcedAction.ActionType == "colony-placement" {
		if err := g.SetForcedFirstAction(ctx, playerID, nil); err != nil {
			log.Error("Failed to clear forced colony placement action", zap.Error(err))
		}
	}

	log.Info("Colony placed from card effect",
		zap.String("colony_id", colonyID))

	return nil
}
