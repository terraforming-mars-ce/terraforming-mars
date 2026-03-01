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

// StartGameAction handles the business logic for starting games
// NOTE: Deck initialization is handled separately before calling this action
type StartGameAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewStartGameAction creates a new start game action
func NewStartGameAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *StartGameAction {
	return &StartGameAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the start game action
func (a *StartGameAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "start_game"),
	)
	log.Info("🎮 Starting game")

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
	log.Info("🎮 Starting game with players", zap.Int("player_count", len(players)))

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
	log.Info("🎲 Randomized turn order", zap.Strings("turn_order", playerIDs))

	// 5b. BUSINESS LOGIC: Assign stable colors based on turn order
	playerColors := []string{"#b91c2b", "#232dc7", "#3abe3a", "#ffa502", "#a55eea", "#26d0ce"}
	for i, pid := range playerIDs {
		p, err := g.GetPlayer(pid)
		if err != nil {
			log.Error("Failed to get player for color assignment", zap.Error(err))
			return fmt.Errorf("failed to get player for color assignment: %w", err)
		}
		p.SetColor(playerColors[i%len(playerColors)])
	}

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
		log.Info("✅ Set initial turn", zap.String("first_player_id", firstPlayerID))
	}

	// 9. BUSINESS LOGIC: Demo games go to DemoSetup phase, normal games to StartingCardSelection
	if g.Settings().DemoGame {
		// Demo game: go to demo setup phase where players configure their setup
		if err := g.UpdatePhase(ctx, game.GamePhaseDemoSetup); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}
		log.Info("🎮 Demo game entering setup phase")
	} else {
		// Normal game: go to starting card selection phase
		if err := g.UpdatePhase(ctx, game.GamePhaseStartingCardSelection); err != nil {
			log.Error("Failed to update game phase", zap.Error(err))
			return fmt.Errorf("failed to update game phase: %w", err)
		}

		// Distribute starting cards to all players
		if err := a.distributeStartingCards(ctx, g, players); err != nil {
			log.Error("Failed to distribute starting cards", zap.Error(err))
			return fmt.Errorf("failed to distribute starting cards: %w", err)
		}
		log.Info("✅ Starting cards distributed to all players")
	}

	log.Info("🎉 Game started successfully")
	return nil
}

// distributeStartingCards gives each player 10 project cards and 2 corporations
func (a *StartGameAction) distributeStartingCards(ctx context.Context, gameInstance *game.Game, players []*playerPkg.Player) error {
	log := a.logger.With(zap.String("game_id", gameInstance.ID()))
	log.Debug("Distributing starting cards to players", zap.Int("player_count", len(players)))

	deck := gameInstance.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		// Draw 10 project cards from game deck
		projectCardIDs, err := deck.DrawProjectCards(ctx, 10)
		if err != nil {
			return fmt.Errorf("failed to draw project cards for player %s: %w", p.ID(), err)
		}

		// Draw 2 corporation cards from game deck
		corporationIDs, err := deck.DrawCorporations(ctx, 2)
		if err != nil {
			return fmt.Errorf("failed to draw corporations for player %s: %w", p.ID(), err)
		}

		// Set starting cards selection phase for player (phase state managed by Game)
		selectionPhase := &playerPkg.SelectStartingCardsPhase{
			AvailableCards:        projectCardIDs,
			AvailableCorporations: corporationIDs,
		}
		if err := gameInstance.SetSelectStartingCardsPhase(ctx, p.ID(), selectionPhase); err != nil {
			return fmt.Errorf("failed to set selection phase for player %s: %w", p.ID(), err)
		}

		log.Info("✅ Distributed cards to player",
			zap.String("player_id", p.ID()),
			zap.Int("project_cards", len(projectCardIDs)),
			zap.Int("corporations", len(corporationIDs)),
			zap.Strings("corporation_ids", corporationIDs))
	}

	return nil
}
