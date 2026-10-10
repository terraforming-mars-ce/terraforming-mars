package turn_management

import (
	"context"
	"fmt"
	"log/slog"

	"openmars/internal/game"
	playerPkg "openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// ExecuteProductionPhase handles the production phase when all players have passed.
// It calculates production, draws cards, advances the generation, rotates turn order,
// and transitions the game to the production_and_card_draw phase.
func ExecuteProductionPhase(ctx context.Context, g *game.Game, players []*playerPkg.Player, log *slog.Logger) error {
	log = log.With(slog.String("game_id", g.ID()))
	log.Debug("Starting production phase",
		slog.Int("player_count", len(players)),
		slog.Int("generation", g.Generation()))

	// Solar Phase: advance colony markers and reset trade fleets
	if g.HasColonies() {
		for _, state := range g.Colonies().States() {
			if state.AwaitingResource != "" {
				continue
			}
			state.TradedThisGen = false
			state.TraderID = ""
			if state.MarkerPosition < 6 {
				state.MarkerPosition++
			}
		}
		g.Colonies().ResetTradeFleets()
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
					slog.Int("cards_drawn", len(drawnCards)),
					slog.Int("attempt", i),
					slog.Any("error", err))
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
			slog.String("player_id", p.ID()),
			slog.Int("available_cards", len(drawnCards)))

		err := g.SetProductionPhase(ctx, p.ID(), productionPhaseData)
		if err != nil {
			log.Error("Failed to set production phase", slog.Any("error", err))
			return fmt.Errorf("failed to set production phase: %w", err)
		}

		log.Debug("Production phase data set",
			slog.String("player_id", p.ID()),
			slog.Int("cards_drawn", len(drawnCards)),
			slog.Int("credits_income", productionPhaseData.CreditsIncome),
			slog.Int("energy_converted", energyConverted))
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
			slog.Any("new_turn_order", turnOrder))
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
		slog.String("current_phase", string(g.CurrentPhase())),
		slog.String("new_phase", string(shared.GamePhaseProductionAndCardDraw)))

	err := g.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw)
	if err != nil {
		log.Error("Failed to update phase", slog.Any("error", err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("Production phase complete, generation advanced",
		slog.Int("old_generation", oldGeneration),
		slog.Int("new_generation", newGeneration))

	return nil
}

// ExecuteFinalProductionPhase runs the production phase for the final generation.
// Per TM rules, there is no research phase (no card drawing) after the final production.
// Sets up ProductionPhase data for the modal and transitions to production_and_card_draw.
func ExecuteFinalProductionPhase(ctx context.Context, g *game.Game, players []*playerPkg.Player, log *slog.Logger) error {
	log = log.With(slog.String("game_id", g.ID()))
	log.Debug("Starting final production phase",
		slog.Int("player_count", len(players)),
		slog.Int("generation", g.Generation()))

	if g.HasColonies() {
		for _, state := range g.Colonies().States() {
			if state.AwaitingResource != "" {
				continue
			}
			state.TradedThisGen = false
			state.TraderID = ""
			if state.MarkerPosition < 6 {
				state.MarkerPosition++
			}
		}
		g.Colonies().ResetTradeFleets()
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
			log.Error("Failed to set final production phase", slog.Any("error", err))
			return fmt.Errorf("failed to set final production phase: %w", err)
		}
		log.Debug("Final production applied",
			slog.String("player_id", p.ID()),
			slog.Int("credits_income", production.Credits+tr),
			slog.Int("energy_converted", energyConverted))
	}

	if err := g.UpdatePhase(ctx, shared.GamePhaseProductionAndCardDraw); err != nil {
		log.Error("Failed to update phase", slog.Any("error", err))
		return fmt.Errorf("failed to update phase: %w", err)
	}

	log.Info("Final production phase set up, awaiting player confirmation")
	return nil
}
