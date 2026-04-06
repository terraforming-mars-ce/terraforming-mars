package colony

import (
	"context"
	"fmt"
	"slices"
	"time"

	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecolony "terraforming-mars-backend/internal/game/colony"
	gameplayer "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

const (
	BuildColonyCost = 17
)

// BuildColonyAction handles the business logic for building a colony on a colony tile
type BuildColonyAction struct {
	baseaction.BaseAction
	colonyRegistry colonies.ColonyRegistry
	cardRegistry   cards.CardRegistry
}

// NewBuildColonyAction creates a new build colony action
func NewBuildColonyAction(
	gameRepo game.GameRepository,
	colonyRegistry colonies.ColonyRegistry,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *BuildColonyAction {
	return &BuildColonyAction{
		BaseAction:     baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
		colonyRegistry: colonyRegistry,
		cardRegistry:   cardRegistry,
	}
}

// Execute performs the build colony action
func (a *BuildColonyAction) Execute(ctx context.Context, gameID string, playerID string, colonyID string) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("action", "build_colony"),
		zap.String("colony_id", colonyID),
	)
	log.Debug("Building colony")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, shared.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateActionsRemaining(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateNoPendingSelections(g, playerID, log); err != nil {
		return err
	}

	if !g.HasColonies() {
		return fmt.Errorf("colonies expansion is not enabled")
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	tileState := g.Colonies().GetState(colonyID)
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

	if slices.Contains(tileState.PlayerColonies, playerID) {
		return fmt.Errorf("player already has a colony on this tile")
	}

	resources := player.Resources().Get()
	if resources.Credits < BuildColonyCost {
		log.Warn("Insufficient credits for colony",
			zap.Int("cost", BuildColonyCost),
			zap.Int("player_credits", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", BuildColonyCost, resources.Credits)
	}

	// Deduct cost
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -BuildColonyCost,
	})

	// Place colony
	slotIndex := len(tileState.PlayerColonies)
	tileState.PlayerColonies = append(tileState.PlayerColonies, playerID)

	if tileState.MarkerPosition < len(tileState.PlayerColonies) {
		tileState.MarkerPosition = len(tileState.PlayerColonies)
	}

	// Apply placement reward
	calculatedOutputs := []shared.CalculatedOutput{
		{ResourceType: "colony", Amount: 1},
	}
	if slotIndex < len(definition.Colonies) {
		slot := definition.Colonies[slotIndex]
		for _, reward := range slot.Reward {
			if reward.Amount > 0 {
				pending := applyOutput(ctx, g, player, reward.Type, reward.Amount, a.cardRegistry, log)
				if pending != nil {
					setPendingColonyResource(player, pending, definition.Name, colonyID, "build", a.cardRegistry, log)
				}
				calculatedOutputs = append(calculatedOutputs, shared.CalculatedOutput{
					ResourceType: reward.Type,
					Amount:       reward.Amount,
				})
			}
		}
	}

	g.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:          "Build Colony: " + definition.Name,
		PlayerID:          playerID,
		SourceType:        shared.SourceTypeColonyBuild,
		CalculatedOutputs: calculatedOutputs,
	})

	a.WriteStateLogFull(ctx, g, "Build Colony: "+definition.Name, shared.SourceTypeColonyBuild,
		playerID, fmt.Sprintf("Built colony on %s", definition.Name), nil, calculatedOutputs, nil)

	events.Publish(g.EventBus(), events.ColonyBuiltEvent{
		GameID:    g.ID(),
		PlayerID:  playerID,
		ColonyID:  colonyID,
		Timestamp: time.Now(),
	})

	a.ConsumePlayerAction(g, log)

	log.Info("Colony built",
		zap.String("colony_id", colonyID),
		zap.Int("slot_index", slotIndex))

	return nil
}

// PlaceColonyOnTile places a colony on the given tile for free (no cost deduction).
// It applies the placement reward, records the triggered effect, and publishes ColonyBuiltEvent.
// Used by card-triggered colony placements (e.g., Mining Colony, Poseidon first action).
func PlaceColonyOnTile(
	ctx context.Context,
	g *game.Game,
	player *gameplayer.Player,
	definition *gamecolony.ColonyDefinition,
	tileState *gamecolony.ColonyState,
	cardRegistry cards.CardRegistry,
	source string,
	log *zap.Logger,
) error {
	slotIndex := len(tileState.PlayerColonies)
	tileState.PlayerColonies = append(tileState.PlayerColonies, player.ID())

	if tileState.MarkerPosition < len(tileState.PlayerColonies) {
		tileState.MarkerPosition = len(tileState.PlayerColonies)
	}

	calculatedOutputs := []shared.CalculatedOutput{
		{ResourceType: "colony", Amount: 1},
	}
	if slotIndex < len(definition.Colonies) {
		slot := definition.Colonies[slotIndex]
		for _, reward := range slot.Reward {
			if reward.Amount > 0 {
				pending := applyOutput(ctx, g, player, reward.Type, reward.Amount, cardRegistry, log)
				if pending != nil {
					setPendingColonyResource(player, pending, definition.Name, tileState.DefinitionID, "build", cardRegistry, log)
				}
				calculatedOutputs = append(calculatedOutputs, shared.CalculatedOutput{
					ResourceType: reward.Type,
					Amount:       reward.Amount,
				})
			}
		}
	}

	g.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:          source + ": " + definition.Name,
		PlayerID:          player.ID(),
		SourceType:        shared.SourceTypeColonyBuild,
		CalculatedOutputs: calculatedOutputs,
	})

	events.Publish(g.EventBus(), events.ColonyBuiltEvent{
		GameID:    g.ID(),
		PlayerID:  player.ID(),
		ColonyID:  tileState.DefinitionID,
		Timestamp: time.Now(),
	})

	log.Debug("Colony placed (free)",
		zap.String("colony_id", tileState.DefinitionID),
		zap.String("player_id", player.ID()),
		zap.Int("slot_index", slotIndex))

	return nil
}
