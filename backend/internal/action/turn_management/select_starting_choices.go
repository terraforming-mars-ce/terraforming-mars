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
	"terraforming-mars-backend/internal/game/shared"
)

// SelectStartingChoicesAction handles the combined selection of corporation, preludes, and starting cards
type SelectStartingChoicesAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	corpProc     *gamecards.CorporationProcessor
	logger       *zap.Logger
}

// NewSelectStartingChoicesAction creates a new select starting choices action
func NewSelectStartingChoicesAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *SelectStartingChoicesAction {
	return &SelectStartingChoicesAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		corpProc:     gamecards.NewCorporationProcessor(cardRegistry, logger),
		logger:       logger,
	}
}

// Execute performs the combined starting selection action
func (a *SelectStartingChoicesAction) Execute(ctx context.Context, gameID string, playerID string, corporationID string, preludeIDs []string, cardIDs []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_starting_choices"),
		zap.String("corporation_id", corporationID),
		zap.Strings("prelude_ids", preludeIDs),
		zap.Strings("card_ids", cardIDs),
	)
	log.Info("🎮 Player selecting starting choices")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	if g.CurrentPhase() != game.GamePhaseStartingSelection {
		log.Error("Game not in starting selection phase", zap.String("phase", string(g.CurrentPhase())))
		return fmt.Errorf("game not in starting selection phase")
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	if err := a.selectCorporation(ctx, g, p, corporationID, log); err != nil {
		return err
	}

	if err := a.selectPreludes(ctx, g, p, preludeIDs, log); err != nil {
		return err
	}

	if err := a.selectStartingCards(ctx, g, p, cardIDs, log); err != nil {
		return err
	}

	a.checkAndAdvancePhase(ctx, g, log)

	log.Info("🎉 Starting choices completed successfully")
	return nil
}

func (a *SelectStartingChoicesAction) selectCorporation(ctx context.Context, g *game.Game, p *player.Player, corporationID string, log *zap.Logger) error {
	corpPhase := g.GetSelectCorporationPhase(p.ID())
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
			GameID:    g.ID(),
			PlayerID:  p.ID(),
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

	if err := a.corpProc.SetupForcedFirstAction(ctx, corpCard, g, p.ID()); err != nil {
		log.Error("Failed to setup forced first action", zap.Error(err))
		return fmt.Errorf("failed to setup forced first action: %w", err)
	}

	if err := g.SetSelectCorporationPhase(ctx, p.ID(), nil); err != nil {
		log.Error("Failed to clear corporation phase", zap.Error(err))
		return fmt.Errorf("failed to clear corporation phase: %w", err)
	}
	log.Info("✅ Corporation selection complete")

	return nil
}

func (a *SelectStartingChoicesAction) selectPreludes(ctx context.Context, g *game.Game, p *player.Player, preludeIDs []string, log *zap.Logger) error {
	preludePhase := g.GetSelectPreludeCardsPhase(p.ID())
	if preludePhase == nil {
		if len(preludeIDs) > 0 {
			return fmt.Errorf("prelude cards submitted but player has no prelude phase")
		}
		return nil
	}

	if len(preludeIDs) != preludePhase.MaxSelectable {
		log.Error("Wrong number of preludes selected",
			zap.Int("selected", len(preludeIDs)),
			zap.Int("required", preludePhase.MaxSelectable))
		return fmt.Errorf("must select exactly %d preludes, got %d", preludePhase.MaxSelectable, len(preludeIDs))
	}

	availableSet := make(map[string]bool, len(preludePhase.AvailablePreludes))
	for _, id := range preludePhase.AvailablePreludes {
		availableSet[id] = true
	}
	for _, id := range preludeIDs {
		if !availableSet[id] {
			log.Error("Selected prelude not available", zap.String("prelude_id", id))
			return fmt.Errorf("prelude %s not available for selection", id)
		}
	}

	for _, preludeID := range preludeIDs {
		if err := a.applyPrelude(ctx, g, p, preludeID, log); err != nil {
			return fmt.Errorf("failed to apply prelude %s: %w", preludeID, err)
		}
	}

	selectedSet := make(map[string]bool, len(preludeIDs))
	for _, id := range preludeIDs {
		selectedSet[id] = true
	}
	var unselected []string
	for _, id := range preludePhase.AvailablePreludes {
		if !selectedSet[id] {
			unselected = append(unselected, id)
		}
	}
	if len(unselected) > 0 {
		if err := g.Deck().Discard(ctx, unselected); err != nil {
			log.Error("Failed to discard unselected preludes", zap.Error(err))
			return fmt.Errorf("failed to discard unselected preludes: %w", err)
		}
	}

	if err := g.SetSelectPreludeCardsPhase(ctx, p.ID(), nil); err != nil {
		log.Error("Failed to clear prelude phase", zap.Error(err))
		return fmt.Errorf("failed to clear prelude phase: %w", err)
	}
	log.Info("✅ Prelude selection complete")

	return nil
}

func (a *SelectStartingChoicesAction) applyPrelude(ctx context.Context, g *game.Game, p *player.Player, preludeID string, log *zap.Logger) error {
	card, err := a.cardRegistry.GetByID(preludeID)
	if err != nil {
		log.Error("Failed to fetch prelude card", zap.String("prelude_id", preludeID), zap.Error(err))
		return fmt.Errorf("prelude card not found: %s", preludeID)
	}

	if card.Type != gamecards.CardTypePrelude {
		log.Error("Card is not a prelude", zap.String("card_type", string(card.Type)))
		return fmt.Errorf("card %s is not a prelude card", preludeID)
	}

	tags := make([]string, len(card.Tags))
	for i, tag := range card.Tags {
		tags[i] = string(tag)
	}
	p.PlayedCards().AddCard(card.ID, card.Name, string(card.Type), tags)

	for behaviorIndex, behavior := range card.Behaviors {
		if !gamecards.HasAutoTrigger(behavior) {
			continue
		}

		_, outputs := behavior.ExtractInputsOutputs(nil)

		applier := gamecards.NewBehaviorApplier(p, g, card.Name, log).
			WithSourceCardID(card.ID).
			WithCardRegistry(a.cardRegistry).
			WithSourceType(game.SourceTypeCardPlay).
			WithOnCardsAddedToHand(baseaction.MakeCardDrawCallback(p, g, a.cardRegistry))

		calculatedOutputs, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
		if err != nil {
			return fmt.Errorf("failed to apply prelude behavior %d: %w", behaviorIndex, err)
		}

		if len(calculatedOutputs) > 0 {
			g.AddTriggeredEffect(game.TriggeredEffect{
				CardName:          card.Name,
				PlayerID:          p.ID(),
				SourceType:        game.SourceTypeCardPlay,
				Behaviors:         []shared.CardBehavior{behavior},
				CalculatedOutputs: calculatedOutputs,
			})
		}

		if gamecards.HasPersistentEffects(behavior) {
			effect := player.CardEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			p.Effects().AddEffect(effect)

			events.Publish(g.EventBus(), events.PlayerEffectsChangedEvent{
				GameID:    g.ID(),
				PlayerID:  p.ID(),
				Timestamp: time.Now(),
			})
		}
	}

	for behaviorIndex, behavior := range card.Behaviors {
		if !gamecards.HasConditionalTrigger(behavior) {
			continue
		}

		effect := player.CardEffect{
			CardID:        card.ID,
			CardName:      card.Name,
			BehaviorIndex: behaviorIndex,
			Behavior:      behavior,
		}
		p.Effects().AddEffect(effect)
		baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, a.cardRegistry)
	}

	for behaviorIndex, behavior := range card.Behaviors {
		if !gamecards.HasManualTrigger(behavior) {
			continue
		}

		p.Actions().AddAction(player.CardAction{
			CardID:        card.ID,
			CardName:      card.Name,
			BehaviorIndex: behaviorIndex,
			Behavior:      behavior,
		})
	}

	return nil
}

func (a *SelectStartingChoicesAction) selectStartingCards(ctx context.Context, g *game.Game, p *player.Player, cardIDs []string, log *zap.Logger) error {
	selectionPhase := g.GetSelectStartingCardsPhase(p.ID())
	if selectionPhase == nil {
		log.Error("Player not in starting card selection phase")
		return fmt.Errorf("not in starting card selection phase")
	}

	availableSet := make(map[string]bool)
	for _, id := range selectionPhase.AvailableCards {
		availableSet[id] = true
	}
	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	cost := len(cardIDs) * 3

	resources := p.Resources().Get()
	if resources.Credits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, resources.Credits)
	}

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: -cost,
	})

	baseaction.AddCardsToPlayerHand(cardIDs, p, g, a.cardRegistry, log)

	selectedSet := make(map[string]bool, len(cardIDs))
	for _, id := range cardIDs {
		selectedSet[id] = true
	}
	var unselectedProjectCards []string
	for _, cardID := range selectionPhase.AvailableCards {
		if !selectedSet[cardID] {
			unselectedProjectCards = append(unselectedProjectCards, cardID)
		}
	}
	if len(unselectedProjectCards) > 0 {
		if err := g.Deck().Discard(ctx, unselectedProjectCards); err != nil {
			log.Error("Failed to discard unselected project cards", zap.Error(err))
			return fmt.Errorf("failed to discard unselected project cards: %w", err)
		}
	}

	if err := g.SetSelectStartingCardsPhase(ctx, p.ID(), nil); err != nil {
		log.Error("Failed to clear starting cards phase", zap.Error(err))
		return fmt.Errorf("failed to clear starting cards phase: %w", err)
	}
	log.Info("✅ Starting card selection complete")

	return nil
}

// checkAndAdvancePhase checks if all players have completed all starting selections and advances to action phase
func (a *SelectStartingChoicesAction) checkAndAdvancePhase(ctx context.Context, g *game.Game, log *zap.Logger) {
	allPlayers := g.GetAllPlayers()

	for _, p := range allPlayers {
		if p.HasExited() {
			continue
		}
		if g.GetSelectCorporationPhase(p.ID()) != nil {
			log.Info("⏳ Waiting for other players to complete starting selection")
			return
		}
		if g.GetSelectPreludeCardsPhase(p.ID()) != nil {
			log.Info("⏳ Waiting for other players to complete starting selection")
			return
		}
		if g.GetSelectStartingCardsPhase(p.ID()) != nil {
			log.Info("⏳ Waiting for other players to complete starting selection")
			return
		}
		if g.GetPendingTileSelection(p.ID()) != nil {
			log.Info("⏳ Waiting for player to complete pending tile selection", zap.String("player_id", p.ID()))
			return
		}
		if g.GetPendingTileSelectionQueue(p.ID()) != nil {
			log.Info("⏳ Waiting for player to complete pending tile queue", zap.String("player_id", p.ID()))
			return
		}
	}

	log.Info("🎉 All players completed starting selection, advancing to action phase")

	AdvanceToActionPhase(ctx, g, allPlayers, log)
}

// AdvanceToActionPhase transitions the game from starting_selection to action phase.
// It creates deferred forced first action tile queues and sets the first player's turn.
func AdvanceToActionPhase(ctx context.Context, g *game.Game, allPlayers []*player.Player, log *zap.Logger) {
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		log.Error("Failed to transition game phase", zap.Error(err))
		return
	}

	for _, p := range allPlayers {
		if p.HasExited() {
			continue
		}
		forcedAction := g.GetForcedFirstAction(p.ID())
		if forcedAction == nil || forcedAction.Completed {
			continue
		}

		tileType := forcedActionToTileType(forcedAction.ActionType)
		if tileType == "" {
			continue
		}

		queue := &player.PendingTileSelectionQueue{
			Items:  []string{tileType},
			Source: "corporation-starting-action",
		}
		if err := g.SetPendingTileSelectionQueue(ctx, p.ID(), queue); err != nil {
			log.Error("Failed to create deferred forced action tile queue",
				zap.String("player_id", p.ID()),
				zap.Error(err))
			continue
		}
		log.Info("🎯 Created deferred forced action tile queue",
			zap.String("player_id", p.ID()),
			zap.String("tile_type", tileType))
	}

	activePlayerCount := 0
	for _, p := range allPlayers {
		if !p.HasExited() {
			activePlayerCount++
		}
	}

	turnOrder := g.TurnOrder()
	if len(turnOrder) > 0 {
		// Find first non-exited player in turn order
		firstPlayerID := ""
		for _, id := range turnOrder {
			p, err := g.GetPlayer(id)
			if err == nil && !p.HasExited() {
				firstPlayerID = p.ID()
				break
			}
		}
		if firstPlayerID == "" {
			return
		}

		availableActions := 2
		if activePlayerCount == 1 {
			availableActions = -1
			log.Info("🎮 Solo mode detected - setting unlimited actions")
		}

		if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
			log.Error("Failed to set current turn", zap.Error(err))
			return
		}

		log.Info("✅ Set first player turn with actions",
			zap.String("first_player_id", firstPlayerID),
			zap.Int("available_actions", availableActions))
	}
}

func forcedActionToTileType(actionType string) string {
	switch actionType {
	case "city-placement":
		return "city"
	case "greenery-placement":
		return "greenery"
	case "ocean-placement":
		return "ocean"
	default:
		return ""
	}
}
