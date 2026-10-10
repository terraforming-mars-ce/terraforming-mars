package turn_management

import (
	"context"
	"fmt"
	"log/slog"
	baseaction "openmars/internal/action"
	gameaction "openmars/internal/action/game"

	"openmars/internal/game"
	playerPkg "openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// SkipActionAction handles the business logic for skipping/passing player turns
type SkipActionAction struct {
	baseaction.BaseAction
	finalScoringAction *gameaction.FinalScoringAction
}

// NewSkipActionAction creates a new skip action action
func NewSkipActionAction(
	gameRepo game.GameRepository,
	finalScoringAction *gameaction.FinalScoringAction,
	logger *slog.Logger,
) *SkipActionAction {
	return &SkipActionAction{
		BaseAction:         baseaction.NewBaseAction(gameRepo, nil),
		finalScoringAction: finalScoringAction,
	}
}

// Execute performs the skip action
func (a *SkipActionAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(slog.String("action", "skip_action"))
	log.Debug("Skipping player turn")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateNoPendingSelections(g, playerID, log); err != nil {
		return err
	}

	turnOrder := g.TurnOrder()

	currentTurn := g.CurrentTurn()
	if currentTurn != nil {
		currentTurn.IncrementGlobalActionCounter()
	}

	currentPlayer, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Current player not found in game")
		return fmt.Errorf("player not found in game")
	}

	currentPlayerIndex := -1
	for i, id := range turnOrder {
		if id == playerID {
			currentPlayerIndex = i
			break
		}
	}

	if currentPlayerIndex == -1 {
		log.Error("Current player not found in turn order")
		return fmt.Errorf("player not found in turn order")
	}

	activePlayerCount := 0
	for _, id := range turnOrder {
		p, _ := g.GetPlayer(id)
		if p != nil && !p.HasPassed() && !p.HasExited() {
			activePlayerCount++
		}
	}

	if currentTurn == nil {
		log.Error("No current turn set")
		return fmt.Errorf("no current turn set")
	}
	availableActions := currentTurn.ActionsRemaining()
	isPassing := availableActions == 2 || availableActions == -1 || len(turnOrder) == 1
	if isPassing {
		currentPlayer.SetPassed(true)

		log.Debug("Player PASSED (marked as passed for generation)",
			slog.String("player_id", playerID),
			slog.Int("available_actions", availableActions))

		if activePlayerCount == 2 {
			for _, id := range turnOrder {
				p, _ := g.GetPlayer(id)
				if p != nil && !p.HasPassed() && !p.HasExited() && p.ID() != playerID {
					if err := g.SetCurrentTurn(ctx, p.ID(), -1); err != nil {
						log.Error("Failed to grant unlimited actions to last player", slog.Any("error", err))
						return fmt.Errorf("failed to grant unlimited actions: %w", err)
					}
					log.Debug("Last active player granted unlimited actions due to others passing",
						slog.String("player_id", p.ID()))
				}
			}
		}
	} else {
		// SKIP: Player is done with their turn but not passing for the generation
		// Don't consume action - just advance to next player
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			slog.String("player_id", playerID),
			slog.Int("available_actions", availableActions))
	}

	passedOrExitedCount := 0
	for _, id := range turnOrder {
		p, _ := g.GetPlayer(id)
		if p != nil && (p.HasPassed() || p.HasExited()) {
			passedOrExitedCount++
		}
	}

	allPlayersFinished := passedOrExitedCount == len(turnOrder)

	log.Debug("Checking generation end condition",
		slog.Int("passed_or_exited_count", passedOrExitedCount),
		slog.Int("total_players", len(turnOrder)),
		slog.Bool("all_players_finished", allPlayersFinished))

	if allPlayersFinished {
		var activePlayers []*playerPkg.Player
		for _, p := range g.GetAllPlayers() {
			if !p.HasExited() {
				activePlayers = append(activePlayers, p)
			}
		}

		if g.CurrentPhase() == shared.GamePhaseFinalPhase {
			log.Debug("All players finished final phase - triggering final scoring",
				slog.String("game_id", gameID))
			if err := a.finalScoringAction.Execute(ctx, gameID); err != nil {
				log.Error("Failed to execute final scoring", slog.Any("error", err))
				return fmt.Errorf("failed to execute final scoring: %w", err)
			}
			log.Info("Game ended after final phase")
			return nil
		}

		if g.GlobalParameters().IsMaxed() {
			log.Debug("All global parameters maxed - running final production phase",
				slog.String("game_id", gameID),
				slog.Int("generation", g.Generation()))

			err = ExecuteFinalProductionPhase(ctx, g, activePlayers, log)
			if err != nil {
				log.Error("Failed to execute final production phase", slog.Any("error", err))
				return fmt.Errorf("failed to execute final production phase: %w", err)
			}

			log.Info("Final production phase started, awaiting player confirmation")
			return nil
		}

		log.Debug("All players finished their turns - executing production phase",
			slog.String("game_id", gameID),
			slog.Int("generation", g.Generation()),
			slog.Int("passed_or_exited_players", passedOrExitedCount))

		err = ExecuteProductionPhase(ctx, g, activePlayers, log)
		if err != nil {
			log.Error("Failed to execute production phase", slog.Any("error", err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		log.Info("New generation started")
		return nil
	}

	nextPlayerIndex := (currentPlayerIndex + 1) % len(turnOrder)
	for i := 0; i < len(turnOrder); i++ {
		nextPlayer, _ := g.GetPlayer(turnOrder[nextPlayerIndex])
		if nextPlayer != nil && !nextPlayer.HasPassed() && !nextPlayer.HasExited() {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(turnOrder)
	}

	nextPlayerID := turnOrder[nextPlayerIndex]
	nextActions := 2

	nonPassedCount := 0
	for _, id := range turnOrder {
		p, _ := g.GetPlayer(id)
		if p != nil && !p.HasPassed() && !p.HasExited() {
			nonPassedCount++
		}
	}
	if nonPassedCount == 1 {
		nextActions = -1
		log.Debug("Next player is the last non-passed player, granting unlimited actions",
			slog.String("player_id", nextPlayerID))
	}

	err = g.SetCurrentTurn(ctx, nextPlayerID, nextActions)
	if err != nil {
		log.Error("Failed to update current turn", slog.Any("error", err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Player turn skipped, advanced to next player",
		slog.String("previous_player", playerID),
		slog.String("current_player", nextPlayerID))

	return nil
}
