package game

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// JoinGameAction handles players joining games
// New architecture: Uses only GameRepository + logger, events handle broadcasting
type JoinGameAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// JoinGameResult contains the result of joining a game
type JoinGameResult struct {
	PlayerID string
	GameDto  dto.GameDto
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *JoinGameAction {
	return &JoinGameAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute performs the join game action
// playerID is required and must be generated at handler level for proper connection registration
func (a *JoinGameAction) Execute(
	ctx context.Context,
	gameID string,
	playerName string,
	playerID string,
) (*JoinGameResult, error) {

	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_name", playerName),
	)
	log.Debug("Player joining game")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// 2. Check for reconnection (playerID provided and player exists in game)
	existingPlayer, err := g.GetPlayer(playerID)
	if err == nil && existingPlayer != nil {
		// Reconnection case - skip lobby check, just update connection status
		log.Debug("Player reconnecting", zap.String("player_id", playerID))
		existingPlayer.SetConnected(true)

		gameDto := dto.ToGameDto(g, a.cardRegistry, playerID)
		return &JoinGameResult{
			PlayerID: playerID,
			GameDto:  gameDto,
		}, nil
	}

	// 3. Validate game is in lobby status (only for new joins)
	if g.Status() != shared.GameStatusLobby {
		log.Warn("Game is not in lobby", zap.String("status", string(g.Status())))
		return nil, fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	// 4. Check if player with same name already exists (idempotent join)
	existingPlayers := g.GetAllPlayers()
	for _, p := range existingPlayers {
		if p.Name() == playerName {
			log.Debug("Player already exists, returning existing ID",
				zap.String("player_id", p.ID()))

			// Return the existing game state with personalized view
			gameDto := dto.ToGameDto(g, a.cardRegistry, p.ID())
			return &JoinGameResult{
				PlayerID: p.ID(),
				GameDto:  gameDto,
			}, nil
		}
	}

	// 5. Check max players only for new players
	maxPlayers := g.Settings().MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = game.DefaultMaxPlayers
	}
	if len(existingPlayers) >= maxPlayers {
		log.Error("Game is full", zap.Int("max_players", maxPlayers))
		return nil, fmt.Errorf("game is full")
	}

	// 6. Check if this will be the first player (before adding)
	isFirstPlayer := len(existingPlayers) == 0

	// 7. If first player, set as host BEFORE adding (so auto-broadcast includes hostPlayerID)
	if isFirstPlayer {
		err = g.SetHostPlayerID(ctx, playerID)
		if err != nil {
			log.Error("Failed to set host player", zap.Error(err))
			return nil, fmt.Errorf("failed to set host player: %w", err)
		}
		log.Debug("Player set as host")
	}

	// 8. Create and add player to game (publishes PlayerJoinedEvent which auto-broadcasts)
	newPlayer, err := g.AddNewPlayer(ctx, playerID, playerName)
	if err != nil {
		log.Error("Failed to add player to game", zap.Error(err))
		return nil, fmt.Errorf("failed to add player to game: %w", err)
	}
	action.SetupPlayerCardStore(newPlayer, g, a.cardRegistry)
	log.Debug("Player added to game")

	// 10. Convert to DTO with personalized view for the joining player
	gameDto := dto.ToGameDto(g, a.cardRegistry, newPlayer.ID())

	// Note: Broadcasting handled automatically via PlayerJoinedEvent
	// g.AddPlayer() publishes event → SessionManager subscribes → broadcasts

	log.Info("Player joined game")
	return &JoinGameResult{
		PlayerID: newPlayer.ID(),
		GameDto:  gameDto,
	}, nil
}
