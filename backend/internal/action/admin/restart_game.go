package admin

import (
	"context"
	"fmt"
	"log/slog"

	baseaction "terraforming-mars-backend/internal/action"
	gameAction "terraforming-mars-backend/internal/action/game"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// BotGameStopper stops all bot sessions of a game.
type BotGameStopper interface {
	StopAllBotsForGame(gameID string)
}

// RestartGameAction rebuilds a running game under the same ID with the same players
// and settings, then deals a fresh start (development mode only).
type RestartGameAction struct {
	gameRepo          game.GameRepository
	createGame        *gameAction.CreateGameAction
	startGame         *turnAction.StartGameAction
	cardRegistry      gamecards.CardRegistry
	colonyBonusLookup gamecards.ColonyBonusLookup
	bots              BotGameStopper
	logger            *slog.Logger
}

// NewRestartGameAction creates a new restart game admin action
func NewRestartGameAction(
	gameRepo game.GameRepository,
	createGame *gameAction.CreateGameAction,
	startGame *turnAction.StartGameAction,
	cardRegistry gamecards.CardRegistry,
	colonyBonusLookup gamecards.ColonyBonusLookup,
	bots BotGameStopper,
	logger *slog.Logger,
) *RestartGameAction {
	return &RestartGameAction{
		gameRepo:          gameRepo,
		createGame:        createGame,
		startGame:         startGame,
		cardRegistry:      cardRegistry,
		colonyBonusLookup: colonyBonusLookup,
		bots:              bots,
		logger:            logger,
	}
}

// Execute restarts the game with a fresh deal
func (a *RestartGameAction) Execute(ctx context.Context, gameID string) error {
	log := a.logger.With(
		slog.String("game_id", gameID),
		slog.String("action", "admin_restart_game"),
	)
	log.Debug("Admin: Restarting game")

	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %s", gameID)
	}
	if !g.Settings().DevelopmentMode {
		return fmt.Errorf("restart is only available in development mode")
	}
	if g.Status() == shared.GameStatusLobby {
		return fmt.Errorf("game has not started")
	}

	state := g.State()
	if state == nil {
		return fmt.Errorf("game state not found: %s", gameID)
	}
	hostID := state.HostPlayerID
	identities := make([]game.PlayerIdentity, 0, len(state.PlayerOrder))
	hostPresent := false
	for _, id := range state.PlayerOrder {
		ps, ok := state.Players[id]
		if !ok || ps.HasExited {
			continue
		}
		identities = append(identities, game.IdentityOf(ps))
		if id == hostID {
			hostPresent = true
		}
	}
	if !hostPresent {
		return fmt.Errorf("host is no longer in the game")
	}

	if a.bots != nil {
		a.bots.StopAllBotsForGame(gameID)
	}

	if err := a.gameRepo.DataStore().DeleteGameHistory(gameID); err != nil {
		return err
	}

	newGame, err := a.createGame.Build(gameID, state.Settings)
	if err != nil {
		return fmt.Errorf("failed to rebuild game: %w", err)
	}
	if err := newGame.SetHostPlayerID(ctx, hostID); err != nil {
		return fmt.Errorf("failed to set host: %w", err)
	}
	for _, identity := range identities {
		p, err := newGame.AddRestoredPlayer(ctx, identity)
		if err != nil {
			return err
		}
		baseaction.SetupPlayerCardStore(p, newGame, a.cardRegistry, a.colonyBonusLookup)
	}
	newGame.RestoreSocial(state.Spectators, state.ChatMessages)

	if err := a.gameRepo.Create(ctx, newGame); err != nil {
		return fmt.Errorf("failed to register rebuilt game: %w", err)
	}

	if err := a.startGame.Execute(ctx, gameID, hostID); err != nil {
		return fmt.Errorf("failed to start rebuilt game: %w", err)
	}

	log.Info("Admin restarted game", slog.Int("player_count", len(identities)))
	return nil
}
