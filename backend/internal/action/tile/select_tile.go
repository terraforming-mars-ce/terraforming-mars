package tile

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	baseaction "terraforming-mars-backend/internal/action"
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
	log.Info("🎯 Selecting tile", zap.String("hex", selectedHex))

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return nil, err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return nil, err
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

	// Handle clear differently - removes occupant from a tile (admin debug tool)
	if tileType == "clear" {
		if err := g.Board().ClearTileOccupant(ctx, *coords); err != nil {
			log.Warn("Failed to clear tile occupant", zap.Error(err))
			return nil, fmt.Errorf("failed to clear tile: %w", err)
		}

		log.Info("🧹 Tile cleared",
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

		log.Info("✅ Tile cleared successfully",
			zap.String("position", selectedHex))
		return result, nil
	}

	// Handle land claims differently - they reserve a tile instead of placing an occupant
	if tileType == "land-claim" {
		if err := g.Board().ReserveTile(ctx, *coords, playerID); err != nil {
			log.Warn("Failed to reserve tile", zap.Error(err))
			return nil, fmt.Errorf("failed to reserve tile: %w", err)
		}

		log.Info("🏴 Tile reserved for land claim",
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

		log.Info("✅ Land claim reserved successfully",
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

	log.Info("🏗️ Tile placed on board",
		zap.String("tile_type", tileType),
		zap.String("position", selectedHex))

	placedTile, err := g.Board().GetTile(*coords)
	if err != nil {
		log.Warn("Failed to get placed tile for bonus check", zap.Error(err))
	} else if len(placedTile.Bonuses) > 0 {
		log.Info("🎁 Tile has bonuses", zap.Int("bonus_count", len(placedTile.Bonuses)))

		resourceBonuses := make(map[string]int)

		for _, bonus := range placedTile.Bonuses {
			switch bonus.Type {
			case shared.ResourceSteel, shared.ResourceTitanium, shared.ResourcePlant, shared.ResourceCredit:
				p.Resources().Add(map[shared.ResourceType]int{
					bonus.Type: bonus.Amount,
				})
				log.Info("💰 Awarded resource bonus",
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

				for _, cardID := range cardIDs {
					p.Hand().AddCard(cardID)
				}

				log.Info("🃏 Awarded card draw bonus",
					zap.Int("cards_drawn", len(cardIDs)),
					zap.Strings("card_ids", cardIDs))

			default:
				log.Warn("⚠️  Unhandled tile bonus type",
					zap.String("type", string(bonus.Type)),
					zap.Int("amount", bonus.Amount))
			}
		}

		if len(resourceBonuses) > 0 {
			events.Publish(g.EventBus(), events.PlacementBonusGainedEvent{
				GameID:    gameID,
				PlayerID:  playerID,
				Resources: resourceBonuses,
				Q:         coords.Q,
				R:         coords.R,
				S:         coords.S,
				Timestamp: time.Now(),
			})
			log.Info("📢 Published PlacementBonusGainedEvent",
				zap.Any("resources", resourceBonuses))
		}

		// Clear bonuses from tile after claiming
		if err := g.Board().ClearTileBonuses(ctx, *coords); err != nil {
			log.Warn("Failed to clear tile bonuses", zap.Error(err))
		}
	}

	result := &TilePlacementResult{
		TileType:   tileType,
		Source:     pendingTileSelection.Source,
		Hex:        selectedHex,
		OnComplete: pendingTileSelection.OnComplete,
	}

	switch tileType {
	case "city":
		log.Info("🏙️ City placed (no TR bonus)")

	case "greenery":
		actualSteps, err := g.GlobalParameters().IncreaseOxygen(ctx, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to increase oxygen: %w", err)
		}
		result.OxygenSteps = actualSteps

		if actualSteps > 0 {
			p.Resources().UpdateTerraformRating(1)
			result.TRGained = 1
			log.Info("🌿 Increased oxygen and TR for greenery placement",
				zap.Int("oxygen_steps", actualSteps),
				zap.Int("tr_gained", 1))
		} else {
			log.Info("🌿 Greenery placed but oxygen already maxed")
		}

	case "ocean":
		success, err := g.GlobalParameters().PlaceOcean(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to place ocean: %w", err)
		}
		result.OceanPlaced = success

		if success {
			p.Resources().UpdateTerraformRating(1)
			result.TRGained = 1
			log.Info("🌊 Placed ocean and increased TR",
				zap.Int("tr_gained", 1))
		} else {
			log.Info("🌊 Ocean placed but ocean count already maxed")
		}

	case "volcano":
		log.Info("🌋 Volcano placed (no TR bonus)")
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

	baseaction.AutoAdvanceTurnIfNeeded(g, playerID, log)

	log.Info("✅ Tile selected and placed successfully",
		zap.String("tile_type", tileType),
		zap.String("position", selectedHex))
	return result, nil
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
	default:
		return shared.ResourceType(tileType)
	}
}
