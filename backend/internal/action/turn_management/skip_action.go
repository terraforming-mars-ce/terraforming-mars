package turn_management

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"
	gameaction "terraforming-mars-backend/internal/action/game"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
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
	logger *zap.Logger,
) *SkipActionAction {
	return &SkipActionAction{
		BaseAction:         baseaction.NewBaseAction(gameRepo, nil),
		finalScoringAction: finalScoringAction,
	}
}

// Execute performs the skip action
func (a *SkipActionAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "skip_action"))
	log.Info("⏭️ Skipping player turn")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	turnOrder := g.TurnOrder()

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
		if p != nil && !p.HasPassed() {
			activePlayerCount++
		}
	}

	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		log.Error("No current turn set")
		return fmt.Errorf("no current turn set")
	}
	availableActions := currentTurn.ActionsRemaining()
	isPassing := availableActions == 2 || availableActions == -1 || len(turnOrder) == 1
	if isPassing {
		currentPlayer.SetPassed(true)

		log.Debug("Player PASSED (marked as passed for generation)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", availableActions))

		if activePlayerCount == 2 {
			for _, id := range turnOrder {
				p, _ := g.GetPlayer(id)
				if p != nil && !p.HasPassed() && p.ID() != playerID {
					if err := g.SetCurrentTurn(ctx, p.ID(), -1); err != nil {
						log.Error("Failed to grant unlimited actions to last player", zap.Error(err))
						return fmt.Errorf("failed to grant unlimited actions: %w", err)
					}
					log.Info("🏃 Last active player granted unlimited actions due to others passing",
						zap.String("player_id", p.ID()))
				}
			}
		}
	} else {
		// SKIP: Player is done with their turn but not passing for the generation
		// Don't consume action - just advance to next player
		log.Debug("Player SKIPPED (turn advanced, not passed)",
			zap.String("player_id", playerID),
			zap.Int("available_actions", availableActions))
	}

	passedCount := 0
	for _, id := range turnOrder {
		p, _ := g.GetPlayer(id)
		if p != nil && p.HasPassed() {
			passedCount++
		}
	}

	allPlayersFinished := passedCount == len(turnOrder)

	log.Debug("Checking generation end condition",
		zap.Int("passed_count", passedCount),
		zap.Int("total_players", len(turnOrder)),
		zap.Bool("all_players_finished", allPlayersFinished))

	if allPlayersFinished {
		if g.GlobalParameters().IsMaxed() {
			log.Info("🏆 All global parameters maxed - triggering final scoring",
				zap.String("game_id", gameID),
				zap.Int("generation", g.Generation()))

			err = a.finalScoringAction.Execute(ctx, gameID)
			if err != nil {
				log.Error("Failed to execute final scoring", zap.Error(err))
				return fmt.Errorf("failed to execute final scoring: %w", err)
			}

			log.Info("✅ Game ended, final scores calculated")
			return nil
		}

		log.Info("🏭 All players finished their turns - executing production phase",
			zap.String("game_id", gameID),
			zap.Int("generation", g.Generation()),
			zap.Int("passed_players", passedCount))

		err = a.executeProductionPhase(ctx, g, g.GetAllPlayers())
		if err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}

		log.Info("✅ Production phase completed, new generation started")
		return nil
	}

	nextPlayerIndex := (currentPlayerIndex + 1) % len(turnOrder)
	for i := 0; i < len(turnOrder); i++ {
		nextPlayer, _ := g.GetPlayer(turnOrder[nextPlayerIndex])
		if nextPlayer != nil && !nextPlayer.HasPassed() {
			break
		}
		nextPlayerIndex = (nextPlayerIndex + 1) % len(turnOrder)
	}

	nextPlayerID := turnOrder[nextPlayerIndex]
	nextActions := 2

	nonPassedCount := 0
	for _, id := range turnOrder {
		p, _ := g.GetPlayer(id)
		if p != nil && !p.HasPassed() {
			nonPassedCount++
		}
	}
	if nonPassedCount == 1 {
		nextActions = -1
		log.Info("🏃 Next player is the last non-passed player, granting unlimited actions",
			zap.String("player_id", nextPlayerID))
	}

	err = g.SetCurrentTurn(ctx, nextPlayerID, nextActions)
	if err != nil {
		log.Error("Failed to update current turn", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("✅ Player turn skipped, advanced to next player",
		zap.String("previous_player", playerID),
		zap.String("current_player", nextPlayerID))

	return nil
}

// executeProductionPhase handles the production phase when all players have passed
func (a *SkipActionAction) executeProductionPhase(ctx context.Context, gameInstance *game.Game, players []*playerPkg.Player) error {
	log := a.GetLogger().With(zap.String("game_id", gameInstance.ID()))
	log.Info("🏭 Starting production phase",
		zap.Int("player_count", len(players)),
		zap.Int("generation", gameInstance.Generation()))

	deck := gameInstance.Deck()
	if deck == nil {
		return fmt.Errorf("game deck is nil")
	}

	for _, p := range players {
		currentResources := p.Resources().Get()
		energyConverted := currentResources.Energy

		production := p.Resources().Production()
		tr := p.Resources().TerraformRating()
		newResources := shared.Resources{
			Credits:  currentResources.Credits + production.Credits + tr,
			Steel:    currentResources.Steel + production.Steel,
			Titanium: currentResources.Titanium + production.Titanium,
			Plants:   currentResources.Plants + production.Plants,
			Energy:   production.Energy,
			Heat:     currentResources.Heat + energyConverted + production.Heat,
		}

		p.Resources().Set(newResources)
		p.SetPassed(false)

		drawnCards := []string{}
		for i := range 4 {
			cardIDs, err := deck.DrawProjectCards(ctx, 1)
			if err != nil || len(cardIDs) == 0 {
				log.Debug("⚠️ Deck empty or error drawing card, stopping at card draw",
					zap.Int("cards_drawn", len(drawnCards)),
					zap.Int("attempt", i),
					zap.Error(err))
				break
			}
			drawnCards = append(drawnCards, cardIDs[0])
		}

		productionPhaseData := &playerPkg.ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   currentResources,
			AfterResources:    newResources,
			EnergyConverted:   energyConverted,
			CreditsIncome:     production.Credits + tr,
		}

		log.Info("📋 Setting production phase data for player",
			zap.String("player_id", p.ID()),
			zap.Int("available_cards", len(drawnCards)))

		err := gameInstance.SetProductionPhase(ctx, p.ID(), productionPhaseData)
		if err != nil {
			log.Error("❌ Failed to set production phase", zap.Error(err))
			return fmt.Errorf("failed to set production phase: %w", err)
		}

		log.Info("✅ Production phase data set successfully",
			zap.String("player_id", p.ID()),
			zap.Int("cards_drawn", len(drawnCards)),
			zap.Int("credits_income", productionPhaseData.CreditsIncome),
			zap.Int("energy_converted", energyConverted))
	}

	oldGeneration := gameInstance.Generation()
	if err := gameInstance.AdvanceGeneration(ctx); err != nil {
		return fmt.Errorf("failed to increment generation: %w", err)
	}
	newGeneration := gameInstance.Generation()

	// Rotate turn order for new generation (starting player rotates each generation)
	turnOrder := gameInstance.TurnOrder()
	if len(turnOrder) > 1 {
		rotatedOrder := make([]string, len(turnOrder))
		copy(rotatedOrder, turnOrder[1:])
		rotatedOrder[len(turnOrder)-1] = turnOrder[0]
		if err := gameInstance.SetTurnOrder(ctx, rotatedOrder); err != nil {
			return fmt.Errorf("failed to rotate turn order: %w", err)
		}
		turnOrder = rotatedOrder
		log.Info("🔄 Turn order rotated for new generation",
			zap.Strings("new_turn_order", turnOrder))
	}

	if len(turnOrder) > 0 {
		firstPlayerID := turnOrder[0]
		actionsForNewGeneration := 2
		if len(turnOrder) == 1 {
			actionsForNewGeneration = -1
		}
		if err := gameInstance.SetCurrentTurn(ctx, firstPlayerID, actionsForNewGeneration); err != nil {
			return fmt.Errorf("failed to set current turn: %w", err)
		}
	}

	log.Info("🔄 Updating game phase to production_and_card_draw",
		zap.String("current_phase", string(gameInstance.CurrentPhase())),
		zap.String("new_phase", string(game.GamePhaseProductionAndCardDraw)))

	err := gameInstance.UpdatePhase(ctx, game.GamePhaseProductionAndCardDraw)
	if err != nil {
		log.Error("❌ Failed to update phase", zap.Error(err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("✅ Game phase updated successfully",
		zap.String("phase", string(gameInstance.CurrentPhase())))

	log.Info("🎉 Production phase complete, generation advanced",
		zap.Int("old_generation", oldGeneration),
		zap.Int("new_generation", newGeneration))

	return nil
}
