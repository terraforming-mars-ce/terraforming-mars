package admin

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
)

// SetCorporationAction handles the admin action to set a player's corporation
type SetCorporationAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	corpProc     *gamecards.CorporationProcessor
	logger       *zap.Logger
}

// NewSetCorporationAction creates a new set corporation admin action
func NewSetCorporationAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SetCorporationAction {
	return &SetCorporationAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		corpProc:     gamecards.NewCorporationProcessor(cardRegistry, logger),
		logger:       logger,
	}
}

// Execute performs the set corporation admin action
func (a *SetCorporationAction) Execute(ctx context.Context, gameID string, playerID string, corporationID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "admin_set_corporation"),
		zap.String("corporation_id", corporationID),
	)
	log.Debug("Admin: Setting player corporation")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	oldCorpID := player.CorporationID()
	if oldCorpID != "" {
		log.Debug("Clearing old corporation effects", zap.String("old_corporation_id", oldCorpID))

		player.Effects().RemoveEffectsByCardID(oldCorpID)
		player.Actions().RemoveActionsByCardID(oldCorpID)
		player.PlayedCards().RemoveCard(oldCorpID)
		player.Resources().RemoveCardStorage(oldCorpID)
		player.Resources().ClearPaymentSubstitutes()
		player.Resources().ClearValueModifiers()

		log.Debug("Old corporation effects cleared")
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

	player.SetCorporationID(corporationID)

	corpTags := make([]string, len(corpCard.Tags))
	for i, tag := range corpCard.Tags {
		corpTags[i] = string(tag)
	}
	player.PlayedCards().AddCard(corporationID, corpCard.Name, string(corpCard.Type), corpTags)

	if corpCard.ResourceStorage != nil {
		player.Resources().AddToStorage(corporationID, corpCard.ResourceStorage.Starting)
	}

	log.Debug("Corporation ID set", zap.String("corporation_name", corpCard.Name))

	// Register trigger effects BEFORE applying starting effects so that
	// production-increased triggers (e.g. Manutech) fire on starting production
	triggerEffects := a.corpProc.GetTriggerEffects(corpCard)
	for _, effect := range triggerEffects {
		player.Effects().AddEffect(effect)
		log.Debug("Registered trigger effect",
			zap.String("card_name", effect.CardName),
			zap.Int("behavior_index", effect.BehaviorIndex))

		baseaction.SubscribePassiveEffectToEvents(ctx, g, player, effect, log, a.cardRegistry)
	}

	if err := a.corpProc.ApplyStartingEffects(ctx, corpCard, player, g); err != nil {
		log.Error("Failed to apply corporation starting effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation starting effects: %w", err)
	}

	if err := a.corpProc.ApplyAutoEffects(ctx, corpCard, player, g); err != nil {
		log.Error("Failed to apply corporation auto effects", zap.Error(err))
		return fmt.Errorf("failed to apply corporation auto effects: %w", err)
	}

	autoEffects := a.corpProc.GetAutoEffects(corpCard)
	for _, effect := range autoEffects {
		player.Effects().AddEffect(effect)
		log.Debug("Registered auto effect",
			zap.String("card_name", effect.CardName),
			zap.Int("behavior_index", effect.BehaviorIndex))
	}

	// Publish TagPlayedEvent for each corporation tag (triggers Saturn Systems, etc.)
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
		player.Actions().AddAction(action)
		log.Debug("Registered manual action",
			zap.String("card_name", action.CardName),
			zap.Int("behavior_index", action.BehaviorIndex))
	}

	if err := a.corpProc.SetupForcedFirstAction(ctx, corpCard, g, playerID); err != nil {
		log.Error("Failed to setup forced first action", zap.Error(err))
		return fmt.Errorf("failed to setup forced first action: %w", err)
	}

	log.Info("Admin set corporation completed with all effects applied",
		zap.String("corporation_name", corpCard.Name))
	return nil
}
