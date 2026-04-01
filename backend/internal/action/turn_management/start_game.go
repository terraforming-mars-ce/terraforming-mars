package turn_management

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/colony"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/projectfunding"
	"terraforming-mars-backend/internal/game/shared"
	pfRegistry "terraforming-mars-backend/internal/projectfunding"
)

// BotStarter starts bot sessions when a game begins.
type BotStarter interface {
	StartBot(gameID, playerID, botName, difficulty, speed string, settings shared.GameSettings) error
}

// StartGameAction handles the business logic for starting games
// NOTE: Deck initialization is handled separately before calling this action
type StartGameAction struct {
	gameRepo               game.GameRepository
	colonyRegistry         colonies.ColonyRegistry
	projectFundingRegistry pfRegistry.ProjectFundingRegistry
	botStarter             BotStarter
	logger                 *zap.Logger
}

// NewStartGameAction creates a new start game action
func NewStartGameAction(
	gameRepo game.GameRepository,
	colonyRegistry colonies.ColonyRegistry,
	projectFundingRegistry pfRegistry.ProjectFundingRegistry,
	botStarter BotStarter,
	logger *zap.Logger,
) *StartGameAction {
	return &StartGameAction{
		gameRepo:               gameRepo,
		colonyRegistry:         colonyRegistry,
		projectFundingRegistry: projectFundingRegistry,
		botStarter:             botStarter,
		logger:                 logger,
	}
}

// Execute performs the start game action
func (a *StartGameAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "start_game"),
	)
	log.Debug("Starting game")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game is in lobby status
	if g.Status() != shared.GameStatusLobby {
		log.Warn("Game is not in lobby", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	// 3. BUSINESS LOGIC: Validate player is the host
	if g.HostPlayerID() != playerID {
		log.Warn("Only host can start the game",
			zap.String("host_id", g.HostPlayerID()),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("only host can start the game")
	}

	// 4. Get all players
	players := g.GetAllPlayers()
	log.Debug("Starting game with players", zap.Int("player_count", len(players)))

	// 5. BUSINESS LOGIC: Randomize and set turn order
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(playerIDs), func(i, j int) {
		playerIDs[i], playerIDs[j] = playerIDs[j], playerIDs[i]
	})
	if err := g.SetTurnOrder(ctx, playerIDs); err != nil {
		log.Error("Failed to set turn order", zap.Error(err))
		return fmt.Errorf("failed to set turn order: %w", err)
	}
	log.Debug("Randomized turn order", zap.Strings("turn_order", playerIDs))

	// 5b. BUSINESS LOGIC: Initialize colony tiles if colonies pack enabled
	if g.Settings().HasColonies() {
		a.initializeColonies(g, playerIDs, rng, log)
	}

	// 5c. BUSINESS LOGIC: Initialize project funding if enabled
	if g.Settings().HasProjectFunding() {
		a.initializeProjectFunding(g, log)
	}

	// 6. BUSINESS LOGIC: Ensure deck is initialized
	deck := g.Deck()
	if deck == nil {
		log.Error("Game deck not initialized")
		return fmt.Errorf("game deck not initialized - must initialize deck before starting game")
	}

	// 7. BUSINESS LOGIC: Update game status to Active
	if err := g.UpdateStatus(ctx, shared.GameStatusActive); err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	// 8. BUSINESS LOGIC: Set first player's turn (use randomized turn order)
	if len(playerIDs) > 0 {
		firstPlayerID := playerIDs[0]
		if err := g.SetCurrentTurn(ctx, firstPlayerID, 0); err != nil {
			log.Error("Failed to set current turn", zap.Error(err))
			return fmt.Errorf("failed to set current turn: %w", err)
		}
		log.Debug("Set initial turn", zap.String("first_player_id", firstPlayerID))
	}

	// 9. BUSINESS LOGIC: Demo games use pre-selected choices, normal games distribute cards
	if g.Settings().DemoGame {
		if err := a.startDemoGame(ctx, g, players, log); err != nil {
			return err
		}
	} else {
		if err := g.UpdatePhase(ctx, shared.GamePhaseStartingSelection); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}

		if err := a.distributeCorporations(ctx, g, players); err != nil {
			log.Error("Failed to distribute corporations", zap.Error(err))
			return fmt.Errorf("failed to distribute corporations: %w", err)
		}
		log.Debug("Corporations distributed to all players")

		if g.Settings().HasPrelude() {
			if err := a.distributePreludeCards(ctx, g, players); err != nil {
				log.Error("Failed to distribute prelude cards", zap.Error(err))
				return fmt.Errorf("failed to distribute prelude cards: %w", err)
			}
			log.Debug("Prelude cards distributed to all players")
		}

		if err := a.distributeProjectCards(ctx, g, players); err != nil {
			log.Error("Failed to distribute project cards", zap.Error(err))
			return fmt.Errorf("failed to distribute project cards: %w", err)
		}
		log.Debug("Project cards distributed to all players")
	}

	// Start bot sessions for any bot players
	if a.botStarter != nil {
		settings := g.Settings()
		for _, p := range players {
			if p.IsBot() {
				if err := a.botStarter.StartBot(gameID, p.ID(), p.Name(), string(p.BotDifficulty()), string(p.BotSpeed()), settings); err != nil {
					log.Error("Failed to start bot",
						zap.String("bot_player_id", p.ID()),
						zap.Error(err))
				}
			}
		}
	}

	log.Info("Game started")
	return nil
}

func (a *StartGameAction) initializeColonies(g *game.Game, playerIDs []string, rng *rand.Rand, log *zap.Logger) {
	allColonies := a.colonyRegistry.GetAll()
	if len(allColonies) == 0 {
		log.Warn("No colony definitions available")
		return
	}

	// Select N+2 colonies (min 5)
	numToSelect := len(playerIDs) + 2
	if numToSelect < 5 {
		numToSelect = 5
	}
	if numToSelect > len(allColonies) {
		numToSelect = len(allColonies)
	}

	// Shuffle and pick
	rng.Shuffle(len(allColonies), func(i, j int) {
		allColonies[i], allColonies[j] = allColonies[j], allColonies[i]
	})
	selected := allColonies[:numToSelect]

	// Initialize tile states
	states := make([]*colony.ColonyState, len(selected))
	for i, def := range selected {
		states[i] = &colony.ColonyState{
			DefinitionID:   def.ID,
			MarkerPosition: 1,
			PlayerColonies: []string{},
			TradedThisGen:  false,
		}
	}
	g.Colonies().SetStates(states)
	g.InitializeTradeFleets(playerIDs)

	log.Debug("Colony tiles initialized",
		zap.Int("colony_count", len(states)),
		zap.Int("player_count", len(playerIDs)))
}

func (a *StartGameAction) initializeProjectFunding(g *game.Game, log *zap.Logger) {
	if a.projectFundingRegistry == nil {
		log.Warn("No project funding registry available")
		return
	}

	allProjects := a.projectFundingRegistry.GetAll()
	if len(allProjects) == 0 {
		log.Warn("No project funding definitions available")
		return
	}

	states := make([]*projectfunding.ProjectState, len(allProjects))
	for i, def := range allProjects {
		states[i] = &projectfunding.ProjectState{
			DefinitionID: def.ID,
			SeatOwners:   []string{},
			IsCompleted:  false,
		}
	}
	g.SetProjectFundingStates(states)

	log.Debug("Project funding initialized", zap.Int("project_count", len(states)))
}

func (a *StartGameAction) startDemoGame(ctx context.Context, g *game.Game, players []*playerPkg.Player, log *zap.Logger) error {
	settings := g.Settings()
	deck := g.Deck()
	turnOrder := g.TurnOrder()

	// Validate all human players have made selections; auto-assign bots
	for _, p := range players {
		if p.IsBot() {
			if !p.HasPendingDemoChoices() {
				// Auto-assign random corporation for bots
				corpIDs, err := deck.DrawCorporations(ctx, 1)
				if err != nil {
					return fmt.Errorf("failed to draw corporation for bot %s: %w", p.ID(), err)
				}
				p.SetPendingDemoChoices(&shared.PendingDemoChoices{
					CorporationID: corpIDs[0],
				})
			}
		} else if !p.HasPendingDemoChoices() {
			return fmt.Errorf("player %s has not selected cards", p.Name())
		}
	}

	// Convert PendingDemoChoices to DeferredStartingChoices for each player
	for _, p := range players {
		choices := p.PendingDemoChoices()
		if err := g.SetDeferredStartingChoices(ctx, p.ID(), &shared.DeferredStartingChoices{
			CorporationID: choices.CorporationID,
			PreludeIDs:    choices.PreludeIDs,
			CardIDs:       choices.CardIDs,
		}); err != nil {
			return fmt.Errorf("failed to set deferred choices for player %s: %w", p.ID(), err)
		}
		p.SetCorporationID(choices.CorporationID)
	}

	// Apply global parameter overrides from settings
	gp := g.GlobalParameters()
	if settings.Temperature != nil {
		if err := gp.SetTemperature(ctx, *settings.Temperature); err != nil {
			return fmt.Errorf("failed to set temperature: %w", err)
		}
	}
	if settings.Oxygen != nil {
		if err := gp.SetOxygen(ctx, *settings.Oxygen); err != nil {
			return fmt.Errorf("failed to set oxygen: %w", err)
		}
	}
	if settings.Oceans != nil {
		if err := gp.SetOceans(ctx, *settings.Oceans); err != nil {
			return fmt.Errorf("failed to set oceans: %w", err)
		}
	}
	if settings.Generation != nil {
		if err := g.SetGeneration(ctx, *settings.Generation); err != nil {
			return fmt.Errorf("failed to set generation: %w", err)
		}
	}

	// Transition to InitApplyCorp phase (same as normal flow)
	if err := g.UpdatePhase(ctx, shared.GamePhaseInitApplyCorp); err != nil {
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	firstPlayerID := findFirstActivePlayer(g, turnOrder)
	if firstPlayerID != "" {
		firstIndex := findPlayerIndex(turnOrder, firstPlayerID)
		if err := g.SetInitPhasePlayerIndex(ctx, firstIndex); err != nil {
			return fmt.Errorf("failed to set init phase player index: %w", err)
		}
		if err := g.SetInitPhaseWaitingForConfirm(ctx, true); err != nil {
			return fmt.Errorf("failed to set waiting for confirm: %w", err)
		}
	}

	log.Debug("Demo game entering init_apply_corp phase")
	return nil
}

func (a *StartGameAction) distributeCorporations(ctx context.Context, g *game.Game, players []*playerPkg.Player) error {
	deck := g.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		corporationIDs, err := deck.DrawCorporations(ctx, 2)
		if err != nil {
			return fmt.Errorf("failed to draw corporations for player %s: %w", p.ID(), err)
		}

		phase := &shared.SelectCorporationPhase{
			AvailableCorporations: corporationIDs,
		}
		if err := g.SetSelectCorporationPhase(ctx, p.ID(), phase); err != nil {
			return fmt.Errorf("failed to set corporation phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}

func (a *StartGameAction) distributePreludeCards(ctx context.Context, g *game.Game, players []*playerPkg.Player) error {
	deck := g.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		preludeIDs, err := deck.DrawPreludeCards(ctx, 4)
		if err != nil {
			return fmt.Errorf("failed to draw prelude cards for player %s: %w", p.ID(), err)
		}

		phase := &shared.SelectPreludeCardsPhase{
			AvailablePreludes: preludeIDs,
			MaxSelectable:     2,
		}
		if err := g.SetSelectPreludeCardsPhase(ctx, p.ID(), phase); err != nil {
			return fmt.Errorf("failed to set prelude phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}

func (a *StartGameAction) distributeProjectCards(ctx context.Context, g *game.Game, players []*playerPkg.Player) error {
	deck := g.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		projectCardIDs, err := deck.DrawProjectCards(ctx, 10)
		if err != nil {
			return fmt.Errorf("failed to draw project cards for player %s: %w", p.ID(), err)
		}

		phase := &shared.SelectStartingCardsPhase{
			AvailableCards: projectCardIDs,
		}
		if err := g.SetSelectStartingCardsPhase(ctx, p.ID(), phase); err != nil {
			return fmt.Errorf("failed to set selection phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}
