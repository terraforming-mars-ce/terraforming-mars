package confirmation

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmStealTargetAction handles confirming a steal target player selection after tile placement
type ConfirmStealTargetAction struct {
	baseaction.BaseAction
}

// NewConfirmStealTargetAction creates a new confirm steal target action
func NewConfirmStealTargetAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ConfirmStealTargetAction {
	return &ConfirmStealTargetAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// Execute performs the confirm steal target action
func (a *ConfirmStealTargetAction) Execute(ctx context.Context, gameID string, playerID string, targetPlayerID string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_steal_target"),
		zap.String("target_player_id", targetPlayerID),
	)
	log.Debug("Confirming steal target selection")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	selection := p.Selection().GetPendingStealTargetSelection()
	if selection == nil {
		log.Warn("No pending steal target selection found")
		return fmt.Errorf("no pending steal target selection found")
	}

	p.Selection().SetPendingStealTargetSelection(nil)

	if targetPlayerID == "" {
		log.Debug("Player skipped steal")
		baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)

		description := fmt.Sprintf("Skipped steal from %s", selection.Source)
		a.WriteStateLog(ctx, g, selection.Source, game.SourceTypeCardPlay, playerID, description)

		log.Info("Steal target skipped",
			zap.String("source", selection.Source))
		return nil
	}

	eligible := false
	for _, id := range selection.EligiblePlayerIDs {
		if id == targetPlayerID {
			eligible = true
			break
		}
	}
	if !eligible {
		return fmt.Errorf("player %s is not an eligible steal target", targetPlayerID)
	}

	targetPlayer, err := g.GetPlayer(targetPlayerID)
	if err != nil {
		return fmt.Errorf("target player not found: %w", err)
	}

	resourceType := selection.ResourceType
	targetResources := targetPlayer.Resources().Get()
	available := getResourceByType(targetResources, resourceType)
	stolenAmount := min(selection.Amount, available)

	if stolenAmount > 0 {
		targetPlayer.Resources().Add(map[shared.ResourceType]int{
			resourceType: -stolenAmount,
		})
		p.Resources().Add(map[shared.ResourceType]int{
			resourceType: stolenAmount,
		})
	}

	baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)

	calculatedOutputs := []game.CalculatedOutput{
		{ResourceType: string(resourceType), Amount: stolenAmount},
	}
	description := fmt.Sprintf("Stole %d %s from %s", stolenAmount, resourceType, targetPlayer.Name())
	a.WriteStateLogFull(ctx, g, selection.Source, game.SourceTypeCardPlay, playerID, description, nil, calculatedOutputs, nil)

	log.Info("Steal target confirmed",
		zap.String("source", selection.Source),
		zap.Int("stolen_amount", stolenAmount))
	return nil
}

func getResourceByType(r shared.Resources, rt shared.ResourceType) int {
	switch rt {
	case shared.ResourceCredit:
		return r.Credits
	case shared.ResourceSteel:
		return r.Steel
	case shared.ResourceTitanium:
		return r.Titanium
	case shared.ResourcePlant:
		return r.Plants
	case shared.ResourceEnergy:
		return r.Energy
	case shared.ResourceHeat:
		return r.Heat
	default:
		return 0
	}
}
