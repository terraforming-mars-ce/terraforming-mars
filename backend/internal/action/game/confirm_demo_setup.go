package game

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	internalgame "terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmDemoSetupAction handles a player confirming their demo setup configuration
type ConfirmDemoSetupAction struct {
	gameRepo     internalgame.GameRepository
	cardRegistry cards.CardRegistry
	corpProc     *gamecards.CorporationProcessor
	logger       *zap.Logger
}

// NewConfirmDemoSetupAction creates a new confirm demo setup action
func NewConfirmDemoSetupAction(
	gameRepo internalgame.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConfirmDemoSetupAction {
	return &ConfirmDemoSetupAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		corpProc:     gamecards.NewCorporationProcessor(cardRegistry, logger),
		logger:       logger,
	}
}

// Execute performs the confirm demo setup action
func (a *ConfirmDemoSetupAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	request *dto.ConfirmDemoSetupRequest,
) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "confirm_demo_setup"),
	)
	log.Debug("Player confirming demo setup")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Validate game is in DemoSetup phase
	if g.CurrentPhase() != shared.GamePhaseDemoSetup {
		log.Warn("Game is not in demo setup phase", zap.String("phase", string(g.CurrentPhase())))
		return fmt.Errorf("game is not in demo setup phase: %s", g.CurrentPhase())
	}

	// 3. Get the player
	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Set corporation - either specified or random
	var corporationID string
	if request.CorporationID != nil && *request.CorporationID != "" {
		corporationID = *request.CorporationID
	} else {
		// Select random corporation by filtering all cards for corporation type
		allCards := a.cardRegistry.GetAll()
		var corporations []gamecards.Card
		for _, card := range allCards {
			if card.Type == gamecards.CardTypeCorporation {
				corporations = append(corporations, card)
			}
		}
		if len(corporations) > 0 {
			randomIndex := rand.Intn(len(corporations))
			corporationID = corporations[randomIndex].ID
		}
	}

	// 4a. Apply corporation with all effects
	if corporationID != "" {
		p.SetCorporationID(corporationID)
		log.Debug("Set corporation ID", zap.String("corporation_id", corporationID))

		// Fetch corporation card and apply effects
		corpCard, err := a.cardRegistry.GetByID(corporationID)
		if err != nil {
			log.Error("Failed to fetch corporation card", zap.Error(err))
			return fmt.Errorf("corporation card not found: %s", corporationID)
		}

		// Apply corporation auto effects (payment substitutes, value modifiers, etc.)
		if err := a.corpProc.ApplyAutoEffects(ctx, corpCard, p, g); err != nil {
			log.Error("Failed to apply corporation auto effects", zap.Error(err))
			return fmt.Errorf("failed to apply corporation auto effects: %w", err)
		}

		// Register corporation auto effects for display
		autoEffects := a.corpProc.GetAutoEffects(corpCard)
		for _, effect := range autoEffects {
			p.Effects().AddEffect(effect)
			log.Debug("Registered auto effect",
				zap.String("card_name", effect.CardName),
				zap.Int("behavior_index", effect.BehaviorIndex))
		}

		// Register corporation trigger effects and subscribe to events
		triggerEffects := a.corpProc.GetTriggerEffects(corpCard)
		for _, effect := range triggerEffects {
			p.Effects().AddEffect(effect)
			log.Debug("Registered trigger effect",
				zap.String("card_name", effect.CardName),
				zap.Int("behavior_index", effect.BehaviorIndex))

			// Subscribe trigger effects to relevant events
			action.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, a.cardRegistry)
		}

		// Register corporation manual actions
		manualActions := a.corpProc.GetManualActions(corpCard)
		for _, act := range manualActions {
			p.Actions().AddAction(act)
			log.Debug("Registered manual action",
				zap.String("card_name", act.CardName),
				zap.Int("behavior_index", act.BehaviorIndex))
		}

		// Setup forced first action if corporation requires it
		if err := a.corpProc.SetupForcedFirstAction(ctx, corpCard, g, playerID); err != nil {
			log.Error("Failed to setup forced first action", zap.Error(err))
			return fmt.Errorf("failed to setup forced first action: %w", err)
		}

		g.RegisterCorporationVPGranter(playerID, corporationID)

		log.Debug("Applied corporation effects", zap.String("corporation_name", corpCard.Name))
	}

	// 5. Add cards to hand
	if len(request.CardIDs) > 0 {
		for _, cardID := range request.CardIDs {
			p.Hand().AddCard(cardID)
		}
		log.Debug("Added cards to hand", zap.Int("card_count", len(request.CardIDs)))
	}

	// 6. Set resources
	resources := shared.Resources{
		Credits:  request.Resources.Credits,
		Steel:    request.Resources.Steel,
		Titanium: request.Resources.Titanium,
		Plants:   request.Resources.Plants,
		Energy:   request.Resources.Energy,
		Heat:     request.Resources.Heat,
	}
	p.Resources().Set(resources)
	log.Debug("Set resources", zap.Any("resources", resources))

	// 7. Set production
	production := shared.Production{
		Credits:  request.Production.Credits,
		Steel:    request.Production.Steel,
		Titanium: request.Production.Titanium,
		Plants:   request.Production.Plants,
		Energy:   request.Production.Energy,
		Heat:     request.Production.Heat,
	}
	p.Resources().SetProduction(production)
	log.Debug("Set production", zap.Any("production", production))

	// 7a. Publish TagPlayedEvent for corporation tags AFTER production is set
	// This ensures trigger effects (like Saturn Systems) modify production correctly
	if corporationID != "" {
		corpCard, err := a.cardRegistry.GetByID(corporationID)
		if err == nil {
			for _, tag := range corpCard.Tags {
				events.Publish(g.EventBus(), events.TagPlayedEvent{
					GameID:    gameID,
					PlayerID:  playerID,
					CardID:    corporationID,
					CardName:  corpCard.Name,
					Tag:       string(tag),
					Timestamp: time.Now(),
				})
			}
		}
	}

	// 8. Set terraform rating
	p.Resources().SetTerraformRating(request.TerraformRating)
	log.Debug("Set terraform rating", zap.Int("rating", request.TerraformRating))

	// 9. If host, set global parameters and generation
	isHost := g.HostPlayerID() == playerID
	if isHost && request.GlobalParameters != nil {
		gp := g.GlobalParameters()
		if err := gp.SetTemperature(ctx, request.GlobalParameters.Temperature); err != nil {
			log.Error("Failed to set temperature", zap.Error(err))
			return fmt.Errorf("failed to set temperature: %w", err)
		}
		if err := gp.SetOxygen(ctx, request.GlobalParameters.Oxygen); err != nil {
			log.Error("Failed to set oxygen", zap.Error(err))
			return fmt.Errorf("failed to set oxygen: %w", err)
		}
		if err := gp.SetOceans(ctx, request.GlobalParameters.Oceans); err != nil {
			log.Error("Failed to set oceans", zap.Error(err))
			return fmt.Errorf("failed to set oceans: %w", err)
		}
		log.Debug("Set global parameters",
			zap.Int("temperature", request.GlobalParameters.Temperature),
			zap.Int("oxygen", request.GlobalParameters.Oxygen),
			zap.Int("oceans", request.GlobalParameters.Oceans))
	}

	if isHost && request.Generation != nil {
		if err := g.SetGeneration(ctx, *request.Generation); err != nil {
			log.Error("Failed to set generation", zap.Error(err))
			return fmt.Errorf("failed to set generation: %w", err)
		}
		log.Debug("Set generation", zap.Int("generation", *request.Generation))
	}

	// 10. Mark player as having confirmed demo setup
	p.SetDemoSetupConfirmed(true)
	log.Info("Player confirmed demo setup")

	// 11. Check if all players have confirmed
	allConfirmed := true
	for _, pl := range g.GetAllPlayers() {
		if !pl.DemoSetupConfirmed() {
			allConfirmed = false
			break
		}
	}

	// 12. If all players confirmed, transition to Action phase
	if allConfirmed {
		if err := g.UpdatePhase(ctx, shared.GamePhaseAction); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}
		log.Debug("All players confirmed, transitioning to Action phase")

		// Set first player turn with appropriate action count
		turnOrder := g.TurnOrder()
		if len(turnOrder) > 0 {
			firstPlayerID := turnOrder[0]
			availableActions := 2
			if len(turnOrder) == 1 {
				availableActions = -1 // Unlimited for solo mode
			}
			if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}
			log.Debug("Set first player turn with actions",
				zap.String("player_id", firstPlayerID),
				zap.Int("available_actions", availableActions))
		}
	}

	return nil
}
