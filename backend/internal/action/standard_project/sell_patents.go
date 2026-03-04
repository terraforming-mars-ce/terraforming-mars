package standard_project

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// SellPatentsAction handles the business logic for initiating sell patents standard project
// This is Phase 1: creates pending card selection for player to choose which cards to sell
type SellPatentsAction struct {
	baseaction.BaseAction
}

// NewSellPatentsAction creates a new sell patents action
func NewSellPatentsAction(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *SellPatentsAction {
	return &SellPatentsAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
	}
}

// Execute performs the sell patents action (Phase 1: initiate card selection)
func (a *SellPatentsAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "sell_patents"))
	log.Info("🏛️ Initiating sell patents")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	playerCards := player.Hand().Cards()
	if len(playerCards) == 0 {
		log.Warn("Player has no cards to sell")
		return fmt.Errorf("no cards available to sell")
	}

	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range playerCards {
		cardCosts[cardID] = 0
		cardRewards[cardID] = 1
	}

	pendingSelection := &playerPkg.PendingCardSelection{
		Source:         "sell-patents",
		AvailableCards: playerCards,
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		MinCards:       0,
		MaxCards:       len(playerCards),
	}

	player.Selection().SetPendingCardSelection(pendingSelection)

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: string(shared.StandardProjectSellPatents),
		ProjectCost: 0,
		Timestamp:   time.Now(),
	})

	log.Info("📋 Created pending card selection for sell patents",
		zap.Int("available_cards", len(playerCards)))

	log.Info("✅ Sell patents initiated successfully, awaiting card selection")
	return nil
}
