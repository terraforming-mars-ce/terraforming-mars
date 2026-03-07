package turn_management

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
)

// BotStarter starts bot sessions when a game begins.
type BotStarter interface {
	StartBot(gameID, playerID, botName, difficulty, speed string, settings game.GameSettings) error
}

// StartGameAction handles the business logic for starting games
// NOTE: Deck initialization is handled separately before calling this action
type StartGameAction struct {
	gameRepo   game.GameRepository
	botStarter BotStarter
	logger     *zap.Logger
}

// NewStartGameAction creates a new start game action
func NewStartGameAction(
	gameRepo game.GameRepository,
	botStarter BotStarter,
	logger *zap.Logger,
) *StartGameAction {
	return &StartGameAction{
		gameRepo:   gameRepo,
		botStarter: botStarter,
		logger:     logger,
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
	if g.Status() != game.GameStatusLobby {
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

	// 6. BUSINESS LOGIC: Ensure deck is initialized
	deck := g.Deck()
	if deck == nil {
		log.Error("Game deck not initialized")
		return fmt.Errorf("game deck not initialized - must initialize deck before starting game")
	}

	// 7. BUSINESS LOGIC: Update game status to Active
	if err := g.UpdateStatus(ctx, game.GameStatusActive); err != nil {
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

	// 9. BUSINESS LOGIC: Demo games go to DemoSetup phase, normal games to CorporationSelection
	if g.Settings().DemoGame {
		if err := g.UpdatePhase(ctx, game.GamePhaseDemoSetup); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}
		log.Debug("Demo game entering setup phase")
	} else {
		if err := g.UpdatePhase(ctx, game.GamePhaseStartingSelection); err != nil {
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

		phase := &playerPkg.SelectCorporationPhase{
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

		phase := &playerPkg.SelectPreludeCardsPhase{
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

		phase := &playerPkg.SelectStartingCardsPhase{
			AvailableCards: projectCardIDs,
		}
		if err := g.SetSelectStartingCardsPhase(ctx, p.ID(), phase); err != nil {
			return fmt.Errorf("failed to set selection phase for player %s: %w", p.ID(), err)
		}
	}

	return nil
}
