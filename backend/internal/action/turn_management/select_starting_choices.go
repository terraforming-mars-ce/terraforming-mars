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

// SelectStartingChoicesAction handles the combined selection of corporation, preludes, and starting cards.
// Selections are validated and stored but effects are NOT applied until the init_apply phases.
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

// Execute validates and stores starting selections without applying effects.
// Effects are deferred to init_apply_corp and init_apply_prelude phases.
func (a *SelectStartingChoicesAction) Execute(ctx context.Context, gameID string, playerID string, corporationID string, preludeIDs []string, cardIDs []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_starting_choices"),
		zap.String("corporation_id", corporationID),
		zap.Strings("prelude_ids", preludeIDs),
		zap.Strings("card_ids", cardIDs),
	)
	log.Debug("Player selecting starting choices")

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

	if err := a.validateCorporation(g, p, corporationID, log); err != nil {
		return err
	}

	if err := a.validatePreludes(g, p, preludeIDs, log); err != nil {
		return err
	}

	if err := a.validateStartingCards(g, p, corporationID, cardIDs, log); err != nil {
		return err
	}

	p.SetCorporationID(corporationID)

	if err := g.SetDeferredStartingChoices(ctx, playerID, &game.DeferredStartingChoices{
		CorporationID: corporationID,
		PreludeIDs:    preludeIDs,
		CardIDs:       cardIDs,
	}); err != nil {
		return fmt.Errorf("failed to store deferred starting choices: %w", err)
	}

	// Remove unchosen preludes permanently (preludes must never enter the discard/draw cycle)
	preludePhase := g.GetSelectPreludeCardsPhase(p.ID())
	if preludePhase != nil {
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
			if err := g.Deck().Remove(ctx, unselected); err != nil {
				log.Error("Failed to remove unselected preludes", zap.Error(err))
				return fmt.Errorf("failed to remove unselected preludes: %w", err)
			}
		}
	}

	// Discard unchosen project cards
	selectionPhase := g.GetSelectStartingCardsPhase(p.ID())
	if selectionPhase != nil {
		selectedSet := make(map[string]bool, len(cardIDs))
		for _, id := range cardIDs {
			selectedSet[id] = true
		}
		var unselected []string
		for _, cardID := range selectionPhase.AvailableCards {
			if !selectedSet[cardID] {
				unselected = append(unselected, cardID)
			}
		}
		if len(unselected) > 0 {
			if err := g.Deck().Discard(ctx, unselected); err != nil {
				log.Error("Failed to discard unselected project cards", zap.Error(err))
				return fmt.Errorf("failed to discard unselected project cards: %w", err)
			}
		}
	}

	// Clear all selection phases to signal completion
	if err := g.SetSelectCorporationPhase(ctx, p.ID(), nil); err != nil {
		return fmt.Errorf("failed to clear corporation phase: %w", err)
	}
	if err := g.SetSelectPreludeCardsPhase(ctx, p.ID(), nil); err != nil {
		return fmt.Errorf("failed to clear prelude phase: %w", err)
	}
	if err := g.SetSelectStartingCardsPhase(ctx, p.ID(), nil); err != nil {
		return fmt.Errorf("failed to clear starting cards phase: %w", err)
	}

	a.checkAndAdvanceToInitApplyCorp(ctx, g, log)

	log.Info("Starting choices stored")
	return nil
}

func (a *SelectStartingChoicesAction) validateCorporation(g *game.Game, p *player.Player, corporationID string, log *zap.Logger) error {
	corpPhase := g.GetSelectCorporationPhase(p.ID())
	if corpPhase == nil {
		return fmt.Errorf("not in corporation selection phase")
	}

	if p.HasCorporation() {
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
		return fmt.Errorf("corporation %s not available", corporationID)
	}

	corpCard, err := a.cardRegistry.GetByID(corporationID)
	if err != nil {
		return fmt.Errorf("corporation card not found: %s", corporationID)
	}

	if corpCard.Type != gamecards.CardTypeCorporation {
		return fmt.Errorf("card %s is not a corporation card", corporationID)
	}

	return nil
}

func (a *SelectStartingChoicesAction) validatePreludes(g *game.Game, p *player.Player, preludeIDs []string, log *zap.Logger) error {
	preludePhase := g.GetSelectPreludeCardsPhase(p.ID())
	if preludePhase == nil {
		if len(preludeIDs) > 0 {
			return fmt.Errorf("prelude cards submitted but player has no prelude phase")
		}
		return nil
	}

	if len(preludeIDs) != preludePhase.MaxSelectable {
		return fmt.Errorf("must select exactly %d preludes, got %d", preludePhase.MaxSelectable, len(preludeIDs))
	}

	availableSet := make(map[string]bool, len(preludePhase.AvailablePreludes))
	for _, id := range preludePhase.AvailablePreludes {
		availableSet[id] = true
	}
	for _, id := range preludeIDs {
		if !availableSet[id] {
			return fmt.Errorf("prelude %s not available for selection", id)
		}
	}

	return nil
}

func (a *SelectStartingChoicesAction) validateStartingCards(g *game.Game, p *player.Player, corporationID string, cardIDs []string, log *zap.Logger) error {
	selectionPhase := g.GetSelectStartingCardsPhase(p.ID())
	if selectionPhase == nil {
		return fmt.Errorf("not in starting card selection phase")
	}

	availableSet := make(map[string]bool)
	for _, id := range selectionPhase.AvailableCards {
		availableSet[id] = true
	}
	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	cost := len(cardIDs) * 3
	startingCredits := getCorpStartingCredits(a.cardRegistry, corporationID)
	if startingCredits < cost {
		return fmt.Errorf("insufficient credits: need %d, corp provides %d", cost, startingCredits)
	}

	return nil
}

// getCorpStartingCredits calculates the starting credits a corporation provides
// by examining its auto-corporation-start behaviors
func getCorpStartingCredits(cardRegistry cards.CardRegistry, corporationID string) int {
	corpCard, err := cardRegistry.GetByID(corporationID)
	if err != nil {
		return 0
	}

	credits := 0
	for _, behavior := range corpCard.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == string(gamecards.ResourceTriggerAutoCorporationStart) {
				for _, output := range behavior.Outputs {
					if output.ResourceType == shared.ResourceCredit {
						credits += output.Amount
					}
				}
			}
		}
	}
	return credits
}

// checkAndAdvanceToInitApplyCorp checks if all players have stored their choices
// and transitions to the init_apply_corp phase
func (a *SelectStartingChoicesAction) checkAndAdvanceToInitApplyCorp(ctx context.Context, g *game.Game, log *zap.Logger) {
	allPlayers := g.GetAllPlayers()
	turnOrder := g.TurnOrder()

	for _, p := range allPlayers {
		if p.HasExited() {
			continue
		}
		if g.GetDeferredStartingChoices(p.ID()) == nil {
			log.Debug("Waiting for other players to complete starting selection")
			return
		}
	}

	log.Debug("All players stored starting choices, advancing to init_apply_corp phase")

	if err := g.UpdatePhase(ctx, game.GamePhaseInitApplyCorp); err != nil {
		log.Error("Failed to transition to init_apply_corp phase", zap.Error(err))
		return
	}

	firstPlayerID := findFirstActivePlayer(g, turnOrder)
	if firstPlayerID == "" {
		return
	}

	firstIndex := findPlayerIndex(turnOrder, firstPlayerID)
	if err := g.SetInitPhasePlayerIndex(ctx, firstIndex); err != nil {
		log.Error("Failed to set init phase player index", zap.Error(err))
		return
	}

	if err := g.SetInitPhaseWaitingForConfirm(ctx, true); err != nil {
		log.Error("Failed to set waiting for confirm", zap.Error(err))
		return
	}
}

// ApplyCorpForPlayer applies corporation effects for a single player during init_apply_corp phase.
// This includes starting effects, auto effects, registering triggers/actions,
// and purchasing project cards.
func ApplyCorpForPlayer(ctx context.Context, g *game.Game, playerID string, cardRegistry cards.CardRegistry, corpProc *gamecards.CorporationProcessor, log *zap.Logger) error {
	choices := g.GetDeferredStartingChoices(playerID)
	if choices == nil {
		return fmt.Errorf("no deferred starting choices for player %s", playerID)
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		return fmt.Errorf("player not found: %s", playerID)
	}

	corpCard, err := cardRegistry.GetByID(choices.CorporationID)
	if err != nil {
		return fmt.Errorf("corporation card not found: %s", choices.CorporationID)
	}

	if corpCard.ResourceStorage != nil {
		p.Resources().AddToStorage(choices.CorporationID, corpCard.ResourceStorage.Starting)
	}

	log.Debug("Applying corporation effects",
		zap.String("player_id", playerID),
		zap.String("corporation", corpCard.Name))

	// Register trigger effects BEFORE applying starting effects so that
	// production-increased triggers (e.g. Manutech) fire on starting production
	triggerEffects := corpProc.GetTriggerEffects(corpCard)
	for _, effect := range triggerEffects {
		p.Effects().AddEffect(effect)
		baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, cardRegistry)
		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:   corpCard.Name,
			PlayerID:   p.ID(),
			SourceType: game.SourceTypeEffectAdded,
			Behaviors:  []shared.CardBehavior{effect.Behavior},
		})
	}

	if err := corpProc.ApplyStartingEffects(ctx, corpCard, p, g); err != nil {
		return fmt.Errorf("failed to apply corporation starting effects: %w", err)
	}

	if err := corpProc.ApplyAutoEffects(ctx, corpCard, p, g); err != nil {
		return fmt.Errorf("failed to apply corporation auto effects: %w", err)
	}

	autoEffects := corpProc.GetAutoEffects(corpCard)
	for _, effect := range autoEffects {
		p.Effects().AddEffect(effect)
		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:   corpCard.Name,
			PlayerID:   p.ID(),
			SourceType: game.SourceTypeEffectAdded,
			Behaviors:  []shared.CardBehavior{effect.Behavior},
		})
	}

	for _, tag := range corpCard.Tags {
		events.Publish(g.EventBus(), events.TagPlayedEvent{
			GameID:    g.ID(),
			PlayerID:  p.ID(),
			CardID:    choices.CorporationID,
			CardName:  corpCard.Name,
			Tag:       string(tag),
			Timestamp: time.Now(),
		})
	}

	g.RegisterCorporationVPGranter(p.ID(), choices.CorporationID)

	manualActions := corpProc.GetManualActions(corpCard)
	for _, action := range manualActions {
		p.Actions().AddAction(action)
		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:   corpCard.Name,
			PlayerID:   p.ID(),
			SourceType: game.SourceTypeActionAdded,
			Behaviors:  []shared.CardBehavior{action.Behavior},
		})
	}

	corpProc.WithOnCardsAddedToHand(baseaction.MakeCardDrawCallback(p, g, cardRegistry))
	if err := corpProc.SetupForcedFirstAction(ctx, corpCard, g, p.ID()); err != nil {
		return fmt.Errorf("failed to setup forced first action: %w", err)
	}

	// Purchase project cards now that starting credits are applied
	cost := len(choices.CardIDs) * 3
	if cost > 0 {
		p.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: -cost,
		})
	}
	if len(choices.CardIDs) > 0 {
		baseaction.AddCardsToPlayerHand(choices.CardIDs, p, g, cardRegistry, log)
	}

	g.MarkCorpApplied(playerID)

	log.Debug("Corporation effects and card purchase complete",
		zap.String("player_id", playerID),
		zap.String("corporation", corpCard.Name))

	return nil
}

// ApplyPreludesForPlayer applies all prelude card effects for a single player
// during the init_apply_prelude phase.
func ApplyPreludesForPlayer(ctx context.Context, g *game.Game, playerID string, cardRegistry cards.CardRegistry, log *zap.Logger) error {
	choices := g.GetDeferredStartingChoices(playerID)
	if choices == nil {
		return fmt.Errorf("no deferred starting choices for player %s", playerID)
	}

	if len(choices.PreludeIDs) == 0 {
		return nil
	}

	p, err := g.GetPlayer(playerID)
	if err != nil {
		return fmt.Errorf("player not found: %s", playerID)
	}

	log.Debug("Applying prelude effects",
		zap.String("player_id", playerID),
		zap.Strings("preludes", choices.PreludeIDs))

	for _, preludeID := range choices.PreludeIDs {
		if err := ApplyPreludeCard(ctx, g, p, preludeID, cardRegistry, log); err != nil {
			return fmt.Errorf("failed to apply prelude %s: %w", preludeID, err)
		}
	}

	g.MarkPreludesApplied(playerID)

	log.Debug("Prelude effects complete", zap.String("player_id", playerID))
	return nil
}

// ApplyPreludeCard applies a single prelude card's effects: adds to played cards,
// processes auto behaviors, registers trigger effects and manual actions.
func ApplyPreludeCard(ctx context.Context, g *game.Game, p *player.Player, preludeID string, cardRegistry cards.CardRegistry, log *zap.Logger) error {
	card, err := cardRegistry.GetByID(preludeID)
	if err != nil {
		return fmt.Errorf("prelude card not found: %s", preludeID)
	}

	if card.Type != gamecards.CardTypePrelude {
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
			WithCardRegistry(cardRegistry).
			WithSourceType(game.SourceTypeCardPlay).
			WithOnCardsAddedToHand(baseaction.MakeCardDrawCallback(p, g, cardRegistry))

		_, err := applier.ApplyOutputsAndGetCalculated(ctx, outputs)
		if err != nil {
			return fmt.Errorf("failed to apply prelude behavior %d: %w", behaviorIndex, err)
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
		baseaction.SubscribePassiveEffectToEvents(ctx, g, p, effect, log, cardRegistry)
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

// AdvanceToActionPhase transitions the game to the action phase and sets the first player's turn.
func AdvanceToActionPhase(ctx context.Context, g *game.Game, allPlayers []*player.Player, log *zap.Logger) {
	if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
		log.Error("Failed to transition game phase", zap.Error(err))
		return
	}

	activePlayerCount := 0
	for _, p := range allPlayers {
		if !p.HasExited() {
			activePlayerCount++
		}
	}

	turnOrder := g.TurnOrder()
	if len(turnOrder) > 0 {
		firstPlayerID := findFirstActivePlayer(g, turnOrder)
		if firstPlayerID == "" {
			return
		}

		availableActions := 2
		if activePlayerCount == 1 {
			availableActions = -1
			log.Debug("Solo mode detected - setting unlimited actions")
		}

		if err := g.SetCurrentTurn(ctx, firstPlayerID, availableActions); err != nil {
			log.Error("Failed to set current turn", zap.Error(err))
			return
		}

		log.Debug("Set first player turn with actions",
			zap.String("first_player_id", firstPlayerID),
			zap.Int("available_actions", availableActions))
	}
}

func findFirstActivePlayer(g *game.Game, turnOrder []string) string {
	for _, id := range turnOrder {
		p, err := g.GetPlayer(id)
		if err == nil && !p.HasExited() {
			return p.ID()
		}
	}
	return ""
}
