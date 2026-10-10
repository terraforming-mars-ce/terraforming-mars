package game

import (
	"context"
	"fmt"
	"log/slog"
	"unicode/utf8"

	"openmars/internal/action"
	"openmars/internal/game"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
)

// JoinGameAction handles players joining games
// New architecture: Uses only GameRepository + logger, events handle broadcasting
type JoinGameAction struct {
	gameRepo          game.GameRepository
	cardRegistry      gamecards.CardRegistry
	colonyBonusLookup gamecards.ColonyBonusLookup
	logger            *slog.Logger
}

// JoinGameResult contains the result of joining a game
type JoinGameResult struct {
	PlayerID string
}

// NewJoinGameAction creates a new join game action
func NewJoinGameAction(
	gameRepo game.GameRepository,
	cardRegistry gamecards.CardRegistry,
	logger *slog.Logger,
	colonyBonusLookup ...gamecards.ColonyBonusLookup,
) *JoinGameAction {
	var lookup gamecards.ColonyBonusLookup
	if len(colonyBonusLookup) > 0 {
		lookup = colonyBonusLookup[0]
	}
	return &JoinGameAction{
		gameRepo:          gameRepo,
		cardRegistry:      cardRegistry,
		colonyBonusLookup: lookup,
		logger:            logger,
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
		slog.String("game_id", gameID),
		slog.String("player_name", playerName),
	)
	log.Debug("Player joining game")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Game not found", slog.Any("error", err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	if lobby := g.ResumeLobby(); lobby != nil && !lobby.Claimed[playerID] {
		return nil, fmt.Errorf("choose an available seat in the resume lobby")
	}
	// 2. Check for reconnection (playerID provided and player exists in game)
	existingPlayer, err := g.GetPlayer(playerID)
	if err == nil && existingPlayer != nil {
		// Reconnection case - skip lobby check, just update connection status
		log.Debug("Player reconnecting", slog.String("player_id", playerID))
		existingPlayer.SetConnected(true)

		return &JoinGameResult{PlayerID: playerID}, nil
	}

	// 3. Validate player name length (only for new joins)
	if utf8.RuneCountInString(playerName) > game.MaxPlayerNameLength {
		log.Warn("Player name exceeds maximum length",
			slog.Int("max_length", game.MaxPlayerNameLength))
		return nil, fmt.Errorf("player name exceeds maximum length of %d characters", game.MaxPlayerNameLength)
	}

	// 4. Validate game is in lobby status (only for new joins)
	if g.Status() != shared.GameStatusLobby {
		log.Warn("Game is not in lobby", slog.String("status", string(g.Status())))
		return nil, fmt.Errorf("game is not in lobby: %s", g.Status())
	}

	// 4. Check if player with same name already exists (idempotent join)
	existingPlayers := g.GetAllPlayers()
	for _, p := range existingPlayers {
		if p.Name() == playerName {
			log.Debug("Player already exists, returning existing ID",
				slog.String("player_id", p.ID()))

			return &JoinGameResult{PlayerID: p.ID()}, nil
		}
	}

	// 5. Check max players only for new players
	maxPlayers := g.Settings().MaxPlayers
	if maxPlayers == 0 {
		maxPlayers = game.DefaultMaxPlayers
	}
	if len(existingPlayers) >= maxPlayers {
		log.Error("Game is full", slog.Int("max_players", maxPlayers))
		return nil, fmt.Errorf("game is full")
	}

	// 6. Check if this will be the first player (before adding)
	isFirstPlayer := len(existingPlayers) == 0

	// 7. If first player, set as host BEFORE adding (so auto-broadcast includes hostPlayerID)
	if isFirstPlayer {
		err = g.SetHostPlayerID(ctx, playerID)
		if err != nil {
			log.Error("Failed to set host player", slog.Any("error", err))
			return nil, fmt.Errorf("failed to set host player: %w", err)
		}
		log.Debug("Player set as host")
	}

	// 8. Create and add player to game (publishes PlayerJoinedEvent which auto-broadcasts)
	newPlayer, err := g.AddNewPlayer(ctx, playerID, playerName)
	if err != nil {
		log.Error("Failed to add player to game", slog.Any("error", err))
		return nil, fmt.Errorf("failed to add player to game: %w", err)
	}
	action.SetupPlayerCardStore(newPlayer, g, a.cardRegistry, a.colonyBonusLookup)
	log.Debug("Player added to game")

	// Broadcasting handled automatically via PlayerJoinedEvent:
	// g.AddNewPlayer() publishes the event → broadcaster sends personalized state.
	log.Info("Player joined game")
	return &JoinGameResult{PlayerID: newPlayer.ID()}, nil
}
