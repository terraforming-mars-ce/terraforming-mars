package turn_management

import (
	"context"
	"fmt"
	"time"

	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
)

// SelectCorporationAction handles the business logic for selecting a corporation
type SelectCorporationAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	corpProc     *gamecards.CorporationProcessor
	logger       *zap.Logger
}

// NewSelectCorporationAction creates a new select corporation action
func NewSelectCorporationAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectCorporationAction {
	return &SelectCorporationAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		corpProc:     gamecards.NewCorporationProcessor(cardRegistry, logger),
		logger:       logger,
	}
}

// Execute performs the select corporation action
func (a *SelectCorporationAction) Execute(ctx context.Context, gameID string, playerID string, corporationID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_corporation"),
		zap.String("corporation_id", corporationID),
	)
	log.Info("🏢 Player selecting corporation")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	corpPhase := g.GetSelectCorporationPhase(playerID)
	if corpPhase == nil {
		log.Error("Player not in corporation selection phase")
		return fmt.Errorf("not in corporation selection phase")
	}

	if p.HasCorporation() {
		log.Error("Corporation already selected")
		return fmt.Errorf("corporation already selected")
	}

	corpAvailable := false
	for _, corpID := range corpPhase.AvailableCorporations {
		if corpID == corporationID {
			corpAvailable = true
			break
		}
	}
	if !corpAvailable {
		log.Error("Selected corporation not available")
		return fmt.Errorf("corporation %s not available", corporationID)
	}

	corpCard, err := a.cardRegistry.GetByID(corporationID)
	if err != nil {
		log.Error("Failed to fetch corporation card", zap.Error(err))
		return fmt.Errorf("corporation card not found: %s", corporationID)
	}

	if corpCard.Type != gamecards.CardTypeCorporation {
		log.Error("Card is not a corporation", zap.String("card_type", string(corpCard.Type)))
		return fmt.Errorf("card %s is not a corporation card", corporationID)
	}

	p.SetCorporationID(corporationID)
	log.Info("✅ Corporation selected", zap.String("corporation_id", corporationID))

	if err := a.corpProc.ApplyStartingEffects(ctx, corpCard, p, g); err != nil {
		log.Error("Failed to apply corporation starting effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation starting effects: %w", err)
	}

	if err := a.corpProc.ApplyAutoEffects(ctx, corpCard, p, g); err != nil {
		log.Error("Failed to apply corporation auto effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation auto effects: %w", err)
	}

	autoEffects := a.corpProc.GetAutoEffects(corpCard)
	for _, effect := range autoEffects {
		p.Effects().AddEffect(effect)
	}

	triggerEffects := a.corpProc.GetTriggerEffects(corpCard)
	for _, effect := range triggerEffects {
		p.Effects().AddEffect(effect)
		baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, a.cardRegistry)
	}

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

	manualActions := a.corpProc.GetManualActions(corpCard)
	for _, action := range manualActions {
		p.Actions().AddAction(action)
	}

	if err := a.corpProc.SetupForcedFirstAction(ctx, corpCard, g, playerID); err != nil {
		log.Error("Failed to setup forced first action", zap.Error(err))
		return fmt.Errorf("failed to setup forced first action: %w", err)
	}

	if err := g.SetSelectCorporationPhase(ctx, playerID, nil); err != nil {
		log.Error("Failed to clear corporation phase", zap.Error(err))
		return fmt.Errorf("failed to clear corporation phase: %w", err)
	}
	log.Info("✅ Corporation selection marked complete")

	allPlayers := g.GetAllPlayers()
	allComplete := true
	for _, pl := range allPlayers {
		if !pl.HasCorporation() {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed corporation selection")

		if g.Settings().HasPrelude() {
			if err := g.UpdatePhase(ctx, game.GamePhasePreludeSelection); err != nil {
				log.Error("Failed to transition to prelude selection phase", zap.Error(err))
				return fmt.Errorf("failed to transition to prelude selection phase: %w", err)
			}

			if err := a.distributePreludeCards(ctx, g, allPlayers); err != nil {
				log.Error("Failed to distribute prelude cards", zap.Error(err))
				return fmt.Errorf("failed to distribute prelude cards: %w", err)
			}
			log.Info("✅ Prelude cards distributed to all players")
		} else {
			if err := g.UpdatePhase(ctx, game.GamePhaseStartingCardSelection); err != nil {
				log.Error("Failed to transition to starting card selection phase", zap.Error(err))
				return fmt.Errorf("failed to transition to starting card selection phase: %w", err)
			}

			if err := a.distributeProjectCards(ctx, g, allPlayers); err != nil {
				log.Error("Failed to distribute project cards", zap.Error(err))
				return fmt.Errorf("failed to distribute project cards: %w", err)
			}
			log.Info("✅ Project cards distributed to all players")
		}
	}

	log.Info("🎉 Corporation selection completed successfully")
	return nil
}

func (a *SelectCorporationAction) distributePreludeCards(ctx context.Context, g *game.Game, players []*player.Player) error {
	deck := g.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		preludeIDs, err := deck.DrawPreludeCards(ctx, 4)
		if err != nil {
			return fmt.Errorf("failed to draw prelude cards for player %s: %w", p.ID(), err)
		}

		phase := &player.SelectPreludeCardsPhase{
			AvailablePreludes: preludeIDs,
			MaxSelectable:     2,
		}
		if err := g.SetSelectPreludeCardsPhase(ctx, p.ID(), phase); err != nil {
			return fmt.Errorf("failed to set prelude phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}

func (a *SelectCorporationAction) distributeProjectCards(ctx context.Context, g *game.Game, players []*player.Player) error {
	deck := g.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		projectCardIDs, err := deck.DrawProjectCards(ctx, 10)
		if err != nil {
			return fmt.Errorf("failed to draw project cards for player %s: %w", p.ID(), err)
		}

		selectionPhase := &player.SelectStartingCardsPhase{
			AvailableCards: projectCardIDs,
		}
		if err := g.SetSelectStartingCardsPhase(ctx, p.ID(), selectionPhase); err != nil {
			return fmt.Errorf("failed to set selection phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}
