package tile

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	baseaction "terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/action/turn_management"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// TilePlacementResult contains information about a completed tile placement
type TilePlacementResult struct {
	TileType    string
	Source      string
	Hex         string
	OxygenSteps int
	TRGained    int
	OceanPlaced bool
	OnComplete  *player.TileCompletionCallback
}

// SelectTileAction handles the business logic for selecting a tile position
type SelectTileAction struct {
	baseaction.BaseAction
	completionRegistry *TileCompletionRegistry
}

// NewSelectTileAction creates a new select tile action
func NewSelectTileAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *zap.Logger,
) *SelectTileAction {
	return &SelectTileAction{
		BaseAction:         baseaction.NewBaseActionWithStateRepo(gameRepo, cardRegistry, stateRepo),
		completionRegistry: NewTileCompletionRegistry(stateRepo),
	}
}

// Execute performs the select tile action and returns placement result
func (a *SelectTileAction) Execute(ctx context.Context, gameID string, playerID string, selectedHex string) (*TilePlacementResult, error) {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "select_tile"))
	log.Debug("Selecting tile", zap.String("hex", selectedHex))

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return nil, err
	}

	phase := g.CurrentPhase()
	if phase != game.GamePhaseStartingSelection &&
		phase != game.GamePhaseInitApplyCorp &&
		phase != game.GamePhaseInitApplyPrelude {
		if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
			return nil, err
		}
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return nil, err
	}

	pendingTileSelection := g.GetPendingTileSelection(playerID)
	if pendingTileSelection == nil {
		log.Warn("No pending tile selection found")
		return nil, fmt.Errorf("no pending tile selection found for player %s", playerID)
	}

	hexIsValid := false
	for _, availableHex := range pendingTileSelection.AvailableHexes {
		if availableHex == selectedHex {
			hexIsValid = true
			break
		}
	}
	if !hexIsValid {
		log.Warn("Invalid hex selection",
			zap.String("selected_hex", selectedHex),
			zap.Strings("available_hexes", pendingTileSelection.AvailableHexes))
		return nil, fmt.Errorf("selected hex %s is not valid for placement", selectedHex)
	}

	coords, err := parseHexPosition(selectedHex)
	if err != nil {
		log.Warn("Failed to parse hex coordinates", zap.String("hex", selectedHex), zap.Error(err))
		return nil, fmt.Errorf("invalid hex format: %w", err)
	}

	tileType := pendingTileSelection.TileType

	// Handle clear differently - removes occupant/reservation from a tile (admin debug tool)
	if tileType == "clear" {
		// Check if tile is an ocean before clearing, so we can decrement the ocean count
		clearedTile, tileErr := g.Board().GetTile(*coords)
		if tileErr != nil {
			log.Warn("Failed to get tile for clear", zap.Error(tileErr))
			return nil, fmt.Errorf("failed to get tile: %w", tileErr)
		}
		wasOcean := clearedTile.OccupiedBy != nil && clearedTile.OccupiedBy.Type == shared.ResourceOceanTile

		if err := g.Board().ClearTileOccupant(ctx, *coords); err != nil {
			log.Warn("Failed to clear tile occupant", zap.Error(err))
			return nil, fmt.Errorf("failed to clear tile: %w", err)
		}

		if wasOcean {
			currentOceans := g.GlobalParameters().Oceans()
			if currentOceans > 0 {
				if err := g.GlobalParameters().SetOceans(ctx, currentOceans-1); err != nil {
					log.Warn("Failed to decrement ocean count", zap.Error(err))
				}
			}
		}

		log.Debug("Tile cleared",
			zap.String("position", selectedHex))

		result := &TilePlacementResult{
			TileType:   tileType,
			Source:     pendingTileSelection.Source,
			Hex:        selectedHex,
			OnComplete: pendingTileSelection.OnComplete,
		}

		if err := g.SetPendingTileSelection(ctx, playerID, nil); err != nil {
			return nil, fmt.Errorf("failed to clear pending tile selection: %w", err)
		}

		if err := g.ProcessNextTile(ctx, playerID); err != nil {
			return nil, fmt.Errorf("failed to process next tile: %w", err)
		}

		baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)

		log.Debug("Tile cleared",
			zap.String("position", selectedHex))
		return result, nil
	}

	// Handle tile destruction - removes any tile (including oceans)
	if tileType == "tile-destruction" {
		destroyedTile, tileErr := g.Board().GetTile(*coords)
		if tileErr != nil {
			log.Warn("Failed to get tile for destruction", zap.Error(tileErr))
			return nil, fmt.Errorf("failed to get tile: %w", tileErr)
		}
		wasOcean := destroyedTile.OccupiedBy != nil && destroyedTile.OccupiedBy.Type == shared.ResourceOceanTile

		if err := g.Board().ClearTileOccupant(ctx, *coords); err != nil {
			log.Warn("Failed to destroy tile", zap.Error(err))
			return nil, fmt.Errorf("failed to destroy tile: %w", err)
		}

		if wasOcean {
			currentOceans := g.GlobalParameters().Oceans()
			if currentOceans > 0 {
				if err := g.GlobalParameters().SetOceans(ctx, currentOceans-1); err != nil {
					log.Warn("Failed to decrement ocean count", zap.Error(err))
				}
			}
		}

		result := &TilePlacementResult{
			TileType:   tileType,
			Source:     pendingTileSelection.Source,
			Hex:        selectedHex,
			OnComplete: pendingTileSelection.OnComplete,
		}

		if err := g.SetPendingTileSelection(ctx, playerID, nil); err != nil {
			return nil, fmt.Errorf("failed to clear pending tile selection: %w", err)
		}

		if err := g.ProcessNextTile(ctx, playerID); err != nil {
			return nil, fmt.Errorf("failed to process next tile: %w", err)
		}

		if err := a.completionRegistry.Handle(ctx, g, playerID, result, result.OnComplete); err != nil {
			log.Warn("Failed to handle completion callback", zap.Error(err))
		}

		baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)

		log.Info("Tile destroyed",
			zap.String("position", selectedHex),
			zap.Bool("was_ocean", wasOcean))
		return result, nil
	}

	// Handle land claims differently - they reserve a tile instead of placing an occupant
	if tileType == "land-claim" {
		if err := g.Board().ReserveTile(ctx, *coords, playerID); err != nil {
			log.Warn("Failed to reserve tile", zap.Error(err))
			return nil, fmt.Errorf("failed to reserve tile: %w", err)
		}

		log.Debug("Tile reserved for land claim",
			zap.String("position", selectedHex))

		result := &TilePlacementResult{
			TileType:   tileType,
			Source:     pendingTileSelection.Source,
			Hex:        selectedHex,
			OnComplete: pendingTileSelection.OnComplete,
		}

		if err := g.SetPendingTileSelection(ctx, playerID, nil); err != nil {
			return nil, fmt.Errorf("failed to clear pending tile selection: %w", err)
		}

		if err := g.ProcessNextTile(ctx, playerID); err != nil {
			return nil, fmt.Errorf("failed to process next tile: %w", err)
		}

		if err := a.completionRegistry.Handle(ctx, g, playerID, result, result.OnComplete); err != nil {
			log.Warn("Failed to handle completion callback", zap.Error(err))
		}

		baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)

		log.Debug("Land claim reserved",
			zap.String("position", selectedHex))
		return result, nil
	}

	occupant := board.TileOccupant{
		Type: mapTileTypeToResourceType(tileType),
		Tags: []string{},
	}

	if err := g.Board().UpdateTileOccupancy(ctx, *coords, occupant, playerID); err != nil {
		log.Warn("Failed to place tile", zap.Error(err))
		return nil, fmt.Errorf("failed to place tile: %w", err)
	}

	log.Debug("Tile placed on board",
		zap.String("tile_type", tileType),
		zap.String("position", selectedHex))

	placedTile, err := g.Board().GetTile(*coords)
	if err != nil {
		log.Warn("Failed to get placed tile for bonus check", zap.Error(err))
	} else if len(placedTile.Bonuses) > 0 {
		log.Debug("Tile has bonuses", zap.Int("bonus_count", len(placedTile.Bonuses)))

		resourceBonuses := make(map[string]int)

		for _, bonus := range placedTile.Bonuses {
			switch bonus.Type {
			case shared.ResourceSteel, shared.ResourceTitanium, shared.ResourcePlant, shared.ResourceCredit:
				p.Resources().Add(map[shared.ResourceType]int{
					bonus.Type: bonus.Amount,
				})
				log.Debug("Awarded resource bonus",
					zap.String("resource", string(bonus.Type)),
					zap.Int("amount", bonus.Amount))

				resourceBonuses[string(bonus.Type)] = bonus.Amount

			case shared.ResourceCardDraw:
				deck := g.Deck()
				cardIDs, err := deck.DrawProjectCards(ctx, bonus.Amount)
				if err != nil {
					log.Warn("Failed to draw cards for bonus", zap.Error(err))
					continue
				}

				baseaction.AddCardsToPlayerHand(cardIDs, p, g, a.CardRegistry(), log)

				g.AddTriggeredEffect(game.TriggeredEffect{
					CardName:   "Tile Bonus",
					PlayerID:   playerID,
					SourceType: game.SourceTypeCardPlay,
					CalculatedOutputs: []game.CalculatedOutput{
						{ResourceType: string(shared.ResourceCardDraw), Amount: len(cardIDs)},
					},
				})

				log.Debug("Awarded card draw bonus",
					zap.Int("cards_drawn", len(cardIDs)),
					zap.Strings("card_ids", cardIDs))

			default:
				log.Warn(" Unhandled tile bonus type",
					zap.String("type", string(bonus.Type)),
					zap.Int("amount", bonus.Amount))
			}
		}

		if len(resourceBonuses) > 0 {
			events.Publish(g.EventBus(), events.PlacementBonusGainedEvent{
				GameID:       gameID,
				PlayerID:     playerID,
				Resources:    resourceBonuses,
				SourceCardID: pendingTileSelection.SourceCardID,
				Q:            coords.Q,
				R:            coords.R,
				S:            coords.S,
				Timestamp:    time.Now(),
			})
			log.Debug("Published PlacementBonusGainedEvent",
				zap.Any("resources", resourceBonuses))
		}

		// Clear bonuses from tile after claiming
		if err := g.Board().ClearTileBonuses(ctx, *coords); err != nil {
			log.Warn("Failed to clear tile bonuses", zap.Error(err))
		}
	}

	// Award 2 M€ per adjacent ocean tile (applies to all tile types, including oceans)
	neighbors := coords.GetNeighbors()
	adjacentOceanCount := 0
	for _, neighbor := range neighbors {
		neighborTile, err := g.Board().GetTile(neighbor)
		if err != nil {
			continue
		}
		if neighborTile.OccupiedBy != nil && neighborTile.OccupiedBy.Type == shared.ResourceOceanTile {
			adjacentOceanCount++
		}
	}
	if adjacentOceanCount > 0 {
		oceanBonus := adjacentOceanCount * 2
		p.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredit: oceanBonus,
		})

		g.AddTriggeredEffect(game.TriggeredEffect{
			CardName:   "Ocean Adjacency",
			PlayerID:   playerID,
			SourceType: game.SourceTypeGameEvent,
			CalculatedOutputs: []game.CalculatedOutput{
				{ResourceType: string(shared.ResourceCredit), Amount: oceanBonus},
			},
		})

		log.Debug("Awarded ocean adjacency bonus",
			zap.Int("adjacent_oceans", adjacentOceanCount),
			zap.Int("credits_awarded", oceanBonus))
	}

	result := &TilePlacementResult{
		TileType:   tileType,
		Source:     pendingTileSelection.Source,
		Hex:        selectedHex,
		OnComplete: pendingTileSelection.OnComplete,
	}

	switch tileType {
	case "city":
		log.Debug("City placed (no TR bonus)")

	case "greenery", "world-tree":
		actualSteps, err := g.GlobalParameters().IncreaseOxygen(ctx, 1, playerID)
		if err != nil {
			return nil, fmt.Errorf("failed to increase oxygen: %w", err)
		}
		result.OxygenSteps = actualSteps

		if actualSteps > 0 {
			p.Resources().UpdateTerraformRating(1)
			result.TRGained = 1
			log.Debug("Increased oxygen and TR for forest placement",
				zap.Int("oxygen_steps", actualSteps),
				zap.Int("tr_gained", 1))
		} else {
			log.Debug("Forest placed but oxygen already maxed")
		}

	case "ocean":
		success, err := g.GlobalParameters().PlaceOcean(ctx, playerID)
		if err != nil {
			return nil, fmt.Errorf("failed to place ocean: %w", err)
		}
		result.OceanPlaced = success

		if success {
			p.Resources().UpdateTerraformRating(1)
			result.TRGained = 1
			log.Debug("Placed ocean and increased TR",
				zap.Int("tr_gained", 1))
		} else {
			log.Debug("Ocean placed but ocean count already maxed")
		}

	case "volcano":
		log.Debug("Volcano placed (no TR bonus)")
	}

	if err := g.SetPendingTileSelection(ctx, playerID, nil); err != nil {
		return nil, fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	if err := g.ProcessNextTile(ctx, playerID); err != nil {
		return nil, fmt.Errorf("failed to process next tile: %w", err)
	}

	// Invoke completion callback to log the action
	if err := a.completionRegistry.Handle(ctx, g, playerID, result, result.OnComplete); err != nil {
		log.Warn("Failed to handle completion callback", zap.Error(err))
	}

	switch g.CurrentPhase() {
	case game.GamePhaseStartingSelection:
		a.checkStartingSelectionCompletion(ctx, g, log)
	case game.GamePhaseInitApplyCorp, game.GamePhaseInitApplyPrelude:
		// During init phases, tile placement completes and the frontend sends the next confirm
	default:
		baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)
	}

	log.Info("Tile placed",
		zap.String("tile_type", tileType),
		zap.String("position", selectedHex))
	return result, nil
}

// checkStartingSelectionCompletion checks if all players finished starting selection and tile placements,
// then advances to action phase
func (a *SelectTileAction) checkStartingSelectionCompletion(ctx context.Context, g *game.Game, log *zap.Logger) {
	allPlayers := g.GetAllPlayers()
	for _, p := range allPlayers {
		if g.GetSelectCorporationPhase(p.ID()) != nil {
			return
		}
		if g.GetSelectPreludeCardsPhase(p.ID()) != nil {
			return
		}
		if g.GetSelectStartingCardsPhase(p.ID()) != nil {
			return
		}
		if g.GetPendingTileSelection(p.ID()) != nil {
			return
		}
		if g.GetPendingTileSelectionQueue(p.ID()) != nil {
			return
		}
	}

	log.Debug("All starting selections and tiles resolved, advancing to action phase")

	turn_management.AdvanceToActionPhase(ctx, g, allPlayers, log)
}

func parseHexPosition(hexStr string) (*shared.HexPosition, error) {
	parts := strings.Split(hexStr, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected 3 coordinates, got %d", len(parts))
	}

	q, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid q coordinate: %w", err)
	}

	r, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid r coordinate: %w", err)
	}

	s, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid s coordinate: %w", err)
	}

	if q+r+s != 0 {
		return nil, fmt.Errorf("invalid cube coordinates: q+r+s must equal 0")
	}

	return &shared.HexPosition{Q: q, R: r, S: s}, nil
}

func mapTileTypeToResourceType(tileType string) shared.ResourceType {
	switch tileType {
	case "city":
		return shared.ResourceCityTile
	case "greenery":
		return shared.ResourceGreeneryTile
	case "ocean":
		return shared.ResourceOceanTile
	case "volcano":
		return shared.ResourceVolcanoTile
	case "world-tree":
		return shared.ResourceWorldTreeTile
	default:
		return shared.ResourceType(tileType + "-tile")
	}
}
