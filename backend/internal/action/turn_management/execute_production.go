package turn_management

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ExecuteProductionPhase handles the production phase when all players have passed.
// It calculates production, draws cards, advances the generation, rotates turn order,
// and transitions the game to the production_and_card_draw phase.
func ExecuteProductionPhase(ctx context.Context, g *game.Game, players []*playerPkg.Player, log *zap.Logger) error {
	log = log.With(zap.String("game_id", g.ID()))
	log.Debug("Starting production phase",
		zap.Int("player_count", len(players)),
		zap.Int("generation", g.Generation()))

	// Solar Phase: advance colony markers and reset trade fleets
	if g.HasColonies() {
		for _, state := range g.Colonies().States() {
			state.TradedThisGen = false
			state.TraderID = ""
			if state.MarkerPosition < 6 {
				state.MarkerPosition++
			}
		}
		for _, p := range players {
			g.Colonies().SetTradeFleetAvailable(p.ID(), true)
		}
		log.Debug("Solar phase complete: colony markers advanced, trade fleets reset")
	}

	deck := g.Deck()
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
				log.Debug("Deck empty or error drawing card, stopping at card draw",
					zap.Int("cards_drawn", len(drawnCards)),
					zap.Int("attempt", i),
					zap.Error(err))
				break
			}
			drawnCards = append(drawnCards, cardIDs[0])
		}

		productionPhaseData := &shared.ProductionPhase{
			AvailableCards:    drawnCards,
			SelectionComplete: false,
			BeforeResources:   currentResources,
			AfterResources:    newResources,
			EnergyConverted:   energyConverted,
			CreditsIncome:     production.Credits + tr,
		}

		log.Debug("Setting production phase data for player",
			zap.String("player_id", p.ID()),
			zap.Int("available_cards", len(drawnCards)))

		err := g.SetProductionPhase(ctx, p.ID(), productionPhaseData)
		if err != nil {
			log.Error("Failed to set production phase", zap.Error(err))
			return fmt.Errorf("failed to set production phase: %w", err)
		}

		log.Debug("Production phase data set",
			zap.String("player_id", p.ID()),
			zap.Int("cards_drawn", len(drawnCards)),
			zap.Int("credits_income", productionPhaseData.CreditsIncome),
			zap.Int("energy_converted", energyConverted))
	}

	oldGeneration := g.Generation()
	if err := g.AdvanceGeneration(ctx); err != nil {
		return fmt.Errorf("failed to increment generation: %w", err)
	}
	newGeneration := g.Generation()

	turnOrder := g.TurnOrder()
	if g.IsNextGenTurnOrderFrozen() {
		g.SetNextGenTurnOrderFrozen(false)
		log.Debug("Turn order frozen, skipping rotation for this generation")
	} else if len(turnOrder) > 1 {
		var activePart []string
		var exitedPart []string
		for _, id := range turnOrder {
			p, _ := g.GetPlayer(id)
			if p != nil && p.HasExited() {
				exitedPart = append(exitedPart, id)
			} else {
				activePart = append(activePart, id)
			}
		}
		if len(activePart) > 1 {
			rotated := make([]string, 0, len(activePart))
			rotated = append(rotated, activePart[1:]...)
			rotated = append(rotated, activePart[0])
			activePart = rotated
		}
		rotatedOrder := append(activePart, exitedPart...)
		if err := g.SetTurnOrder(ctx, rotatedOrder); err != nil {
			return fmt.Errorf("failed to rotate turn order: %w", err)
		}
		turnOrder = rotatedOrder
		log.Debug("Turn order rotated for new generation",
			zap.Strings("new_turn_order", turnOrder))
	}

	if len(turnOrder) > 0 {
		// Find first non-exited player in rotated turn order
		firstPlayerID := ""
		activeCount := 0
		for _, id := range turnOrder {
			p, _ := g.GetPlayer(id)
			if p != nil && !p.HasExited() {
				activeCount++
				if firstPlayerID == "" {
					firstPlayerID = id
				}
			}
		}
		if firstPlayerID != "" {
			actionsForNewGeneration := 2
			if activeCount == 1 {
				actionsForNewGeneration = -1
			}
			if err := g.SetCurrentTurn(ctx, firstPlayerID, actionsForNewGeneration); err != nil {
				return fmt.Errorf("failed to set current turn: %w", err)
			}
		}
	}

	log.Debug("Updating game phase to production_and_card_draw",
		zap.String("current_phase", string(g.CurrentPhase())),
		zap.String("new_phase", string(shared.GamePhaseProductionAndCardDraw)))

	err := g.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw)
	if err != nil {
		log.Error("Failed to update phase", zap.Error(err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("Production phase complete, generation advanced",
		zap.Int("old_generation", oldGeneration),
		zap.Int("new_generation", newGeneration))

	return nil
}

// ExecuteFinalProductionPhase runs the production phase for the final generation.
// Per TM rules, there is no research phase (no card drawing) after the final production.
// Sets up ProductionPhase data for the modal and transitions to production_and_card_draw.
func ExecuteFinalProductionPhase(ctx context.Context, g *game.Game, players []*playerPkg.Player, log *zap.Logger) error {
	log = log.With(zap.String("game_id", g.ID()))
	log.Debug("Starting final production phase",
		zap.Int("player_count", len(players)),
		zap.Int("generation", g.Generation()))

	if g.HasColonies() {
		for _, state := range g.Colonies().States() {
			state.TradedThisGen = false
			state.TraderID = ""
			if state.MarkerPosition < 6 {
				state.MarkerPosition++
			}
		}
		for _, p := range players {
			g.Colonies().SetTradeFleetAvailable(p.ID(), true)
		}
		log.Debug("Solar phase complete: colony markers advanced, trade fleets reset")
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

		productionPhaseData := &shared.ProductionPhase{
			AvailableCards:    []string{},
			SelectionComplete: false,
			BeforeResources:   currentResources,
			AfterResources:    newResources,
			EnergyConverted:   energyConverted,
			CreditsIncome:     production.Credits + tr,
		}
		if err := g.SetProductionPhase(ctx, p.ID(), productionPhaseData); err != nil {
			log.Error("Failed to set final production phase", zap.Error(err))
			return fmt.Errorf("failed to set final production phase: %w", err)
		}
		log.Debug("Final production applied",
			zap.String("player_id", p.ID()),
			zap.Int("credits_income", production.Credits+tr),
			zap.Int("energy_converted", energyConverted))
	}

	if err := g.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw); err != nil {
		log.Error("Failed to update phase", zap.Error(err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("Final production phase set up, awaiting player confirmation")
	return nil
}
