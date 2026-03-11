package confirmation

import (
	"context"
	"fmt"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// ConfirmColonyResourceAction handles confirming a card storage target for colony resource placement
type ConfirmColonyResourceAction struct {
	baseaction.BaseAction
}

// NewConfirmColonyResourceAction creates a new confirm colony resource action
func NewConfirmColonyResourceAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *ConfirmColonyResourceAction {
	return &ConfirmColonyResourceAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
	}
}

// Execute performs the confirm colony resource action
func (a *ConfirmColonyResourceAction) Execute(ctx context.Context, gameID string, playerID string, targetCardID string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "confirm_colony_resource"),
		zap.String("target_card_id", targetCardID),
	)
	log.Debug("Confirming colony resource placement")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	selection := p.Selection().GetPendingColonyResourceSelection()
	if selection == nil {
		log.Warn("No pending colony resource selection found")
		return fmt.Errorf("no pending colony resource selection found")
	}

	// Clear the selection first
	p.Selection().SetPendingColonyResourceSelection(nil)

	// If empty target, player skipped (no eligible card or chose to skip)
	if targetCardID == "" {
		log.Debug("Player skipped colony resource placement",
			zap.String("resource_type", selection.ResourceType),
			zap.Int("amount", selection.Amount))
		return nil
	}

	// Validate the target card can store this resource type
	if a.CardRegistry() != nil {
		card, cardErr := a.CardRegistry().GetByID(targetCardID)
		if cardErr != nil {
			return fmt.Errorf("target card not found: %s", targetCardID)
		}
		if card.ResourceStorage == nil || card.ResourceStorage.Type != shared.ResourceType(selection.ResourceType) {
			return fmt.Errorf("card %s cannot store %s", targetCardID, selection.ResourceType)
		}
	}

	// Validate the card belongs to the player (played cards or corporation)
	cardBelongsToPlayer := false
	for _, cardID := range p.PlayedCards().Cards() {
		if cardID == targetCardID {
			cardBelongsToPlayer = true
			break
		}
	}
	if p.CorporationID() == targetCardID {
		cardBelongsToPlayer = true
	}
	if !cardBelongsToPlayer {
		return fmt.Errorf("card %s does not belong to player", targetCardID)
	}

	// Apply the resources to the card's storage
	p.Resources().AddToStorage(targetCardID, selection.Amount)

	log.Info("Colony resource placed",
		zap.String("resource_type", selection.ResourceType),
		zap.Int("amount", selection.Amount),
		zap.String("target_card", targetCardID),
		zap.String("source", selection.Source))

	return nil
}
