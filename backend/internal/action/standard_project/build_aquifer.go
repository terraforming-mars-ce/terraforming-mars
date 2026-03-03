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

const (
	// BuildAquiferCost is the megacredit cost to build an aquifer via standard project
	BuildAquiferCost = 18
)

// BuildAquiferAction handles the business logic for the build aquifer standard project
type BuildAquiferAction struct {
	baseaction.BaseAction
}

// NewBuildAquiferAction creates a new build aquifer action
func NewBuildAquiferAction(
	gameRepo game.GameRepository,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *BuildAquiferAction {
	return &BuildAquiferAction{
		BaseAction: baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
	}
}

// Execute performs the build aquifer action
func (a *BuildAquiferAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "build_aquifer"))
	log.Info("💧 Building aquifer (ocean tile)")

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

	if err := baseaction.ValidateActionsRemaining(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	resources := player.Resources().Get()
	if resources.Credits < BuildAquiferCost {
		log.Warn("Insufficient credits for aquifer",
			zap.Int("cost", BuildAquiferCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildAquiferCost, resources.Credits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -BuildAquiferCost,
	})

	events.Publish(g.EventBus(), events.StandardProjectPlayedEvent{
		GameID:      g.ID(),
		PlayerID:    playerID,
		ProjectType: string(shared.StandardProjectAquifer),
		ProjectCost: BuildAquiferCost,
		Timestamp:   time.Now(),
	})

	resources = player.Resources().Get()
	log.Info("💰 Deducted aquifer cost",
		zap.Int("cost", BuildAquiferCost),
		zap.Int("remaining_credits", resources.Credits))

	queue := &playerPkg.PendingTileSelectionQueue{
		Items:  []string{"ocean"},
		Source: "standard-project-aquifer",
		OnComplete: &playerPkg.TileCompletionCallback{
			Type: "standard-project-aquifer",
		},
	}
	if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
		return fmt.Errorf("failed to queue tile placement: %w", err)
	}

	log.Info("📋 Created tile queue for ocean placement (auto-processed by SetPendingTileSelectionQueue)")

	a.ConsumePlayerAction(g, log)

	log.Info("✅ Aquifer ocean tile queued for placement",
		zap.Int("remaining_credits", resources.Credits))
	return nil
}
