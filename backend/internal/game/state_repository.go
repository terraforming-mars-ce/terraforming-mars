package game

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/game/shared"
)

// GameStateRepository manages game state with diff tracking
type GameStateRepository interface {
	Write(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string) (*StateDiff, error)
	WriteWithChoice(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int) (*StateDiff, error)
	WriteWithChoiceAndOutputs(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []shared.CalculatedOutput) (*StateDiff, error)
	WriteFull(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []shared.CalculatedOutput, displayData *LogDisplayData) (*StateDiff, error)
	Get(ctx context.Context, gameID string) (*Game, error)
	GetDiff(ctx context.Context, gameID string) ([]StateDiff, error)
}

// GameSnapshot represents a serializable snapshot of game state for diffing
type GameSnapshot struct {
	Status      string
	Phase       string
	Generation  int
	CurrentTurn string
	Temperature int
	Oxygen      int
	Oceans      int
	Players     map[string]*PlayerSnapshot
	Tiles       map[string]*TileSnapshot
}

// PlayerSnapshot represents a serializable snapshot of player state
type PlayerSnapshot struct {
	Credits            int
	Steel              int
	Titanium           int
	Plants             int
	Energy             int
	Heat               int
	TerraformRating    int
	CreditsProduction  int
	SteelProduction    int
	TitaniumProduction int
	PlantsProduction   int
	EnergyProduction   int
	HeatProduction     int
	Corporation        string
	Passed             bool
	HandCardIDs        []string
	PlayedCardIDs      []string
}

// TileSnapshot represents a serializable snapshot of a tile
type TileSnapshot struct {
	HexID    string
	TileType string
	OwnerID  string
}

// InMemoryGameStateRepository implements GameStateRepository using in-memory storage
type InMemoryGameStateRepository struct {
	mu        sync.RWMutex
	snapshots map[string]*GameSnapshot
	diffLogs  map[string]*DiffLog
}

// NewInMemoryGameStateRepository creates a new in-memory game state repository
func NewInMemoryGameStateRepository() *InMemoryGameStateRepository {
	return &InMemoryGameStateRepository{
		snapshots: make(map[string]*GameSnapshot),
		diffLogs:  make(map[string]*DiffLog),
	}
}

// Write stores the current game state and computes a diff from the previous state
func (r *InMemoryGameStateRepository) Write(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string) (*StateDiff, error) {
	return r.WriteWithChoice(ctx, gameID, game, source, sourceType, playerID, description, nil)
}

// WriteWithChoice stores the current game state with an optional choice index
func (r *InMemoryGameStateRepository) WriteWithChoice(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int) (*StateDiff, error) {
	return r.WriteWithChoiceAndOutputs(ctx, gameID, game, source, sourceType, playerID, description, choiceIndex, nil)
}

// WriteWithChoiceAndOutputs stores the current game state with optional choice index and calculated outputs
func (r *InMemoryGameStateRepository) WriteWithChoiceAndOutputs(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []shared.CalculatedOutput) (*StateDiff, error) {
	return r.WriteFull(ctx, gameID, game, source, sourceType, playerID, description, choiceIndex, calculatedOutputs, nil)
}

// WriteFull stores the current game state with all optional fields including display data
func (r *InMemoryGameStateRepository) WriteFull(ctx context.Context, gameID string, game *Game, source string, sourceType shared.SourceType, playerID, description string, choiceIndex *int, calculatedOutputs []shared.CalculatedOutput, displayData *LogDisplayData) (*StateDiff, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if game == nil {
		return nil, fmt.Errorf("game cannot be nil")
	}

	newSnapshot := captureGameSnapshot(game)

	r.mu.Lock()
	defer r.mu.Unlock()

	oldSnapshot := r.snapshots[gameID]
	changes := computeSnapshotChanges(oldSnapshot, newSnapshot)

	if r.diffLogs[gameID] == nil {
		r.diffLogs[gameID] = NewDiffLog(gameID)
	}

	seqNum := r.diffLogs[gameID].AppendFull(changes, source, sourceType, playerID, description, choiceIndex, calculatedOutputs, displayData)
	r.snapshots[gameID] = newSnapshot

	return &StateDiff{
		SequenceNumber:    seqNum,
		Timestamp:         r.diffLogs[gameID].Diffs[len(r.diffLogs[gameID].Diffs)-1].Timestamp,
		GameID:            gameID,
		Changes:           changes,
		Source:            source,
		SourceType:        sourceType,
		PlayerID:          playerID,
		Description:       description,
		ChoiceIndex:       choiceIndex,
		CalculatedOutputs: calculatedOutputs,
		DisplayData:       displayData,
	}, nil
}

// Get retrieves the current game from the main GameRepository
// Note: This repository only tracks state history; use GameRepository for current game access
func (r *InMemoryGameStateRepository) Get(ctx context.Context, gameID string) (*Game, error) {
	return nil, fmt.Errorf("GameStateRepository.Get not implemented - use GameRepository.Get instead")
}

// GetDiff retrieves all diffs for the specified game in chronological order
func (r *InMemoryGameStateRepository) GetDiff(ctx context.Context, gameID string) ([]StateDiff, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	diffLog, exists := r.diffLogs[gameID]
	if !exists {
		return nil, fmt.Errorf("game %s not found", gameID)
	}

	return diffLog.GetAll(), nil
}

// captureGameSnapshot creates a snapshot of the current game state
func captureGameSnapshot(game *Game) *GameSnapshot {
	snapshot := &GameSnapshot{
		Status:      string(game.Status()),
		Phase:       string(game.CurrentPhase()),
		Generation:  game.Generation(),
		Temperature: game.GlobalParameters().Temperature(),
		Oxygen:      game.GlobalParameters().Oxygen(),
		Oceans:      game.GlobalParameters().Oceans(),
		Players:     make(map[string]*PlayerSnapshot),
		Tiles:       make(map[string]*TileSnapshot),
	}

	if turn := game.CurrentTurn(); turn != nil {
		snapshot.CurrentTurn = turn.PlayerID()
	}

	for _, p := range game.GetAllPlayers() {
		resources := p.Resources().Get()
		production := p.Resources().Production()

		playerSnapshot := &PlayerSnapshot{
			Credits:            resources.Credits,
			Steel:              resources.Steel,
			Titanium:           resources.Titanium,
			Plants:             resources.Plants,
			Energy:             resources.Energy,
			Heat:               resources.Heat,
			TerraformRating:    p.Resources().TerraformRating(),
			CreditsProduction:  production.Credits,
			SteelProduction:    production.Steel,
			TitaniumProduction: production.Titanium,
			PlantsProduction:   production.Plants,
			EnergyProduction:   production.Energy,
			HeatProduction:     production.Heat,
			Passed:             p.HasPassed(),
			HandCardIDs:        make([]string, 0),
			PlayedCardIDs:      make([]string, 0),
			Corporation:        p.CorporationID(),
		}

		for _, cardID := range p.Hand().Cards() {
			playerSnapshot.HandCardIDs = append(playerSnapshot.HandCardIDs, cardID)
		}

		for _, cardID := range p.PlayedCards().Cards() {
			playerSnapshot.PlayedCardIDs = append(playerSnapshot.PlayedCardIDs, cardID)
		}

		snapshot.Players[p.ID()] = playerSnapshot
	}

	for _, tile := range game.Board().Tiles() {
		if tile.OccupiedBy != nil {
			hexID := tile.Coordinates.String()
			tileSnapshot := &TileSnapshot{
				HexID:    hexID,
				TileType: string(tile.OccupiedBy.Type),
			}
			if tile.OwnerID != nil {
				tileSnapshot.OwnerID = *tile.OwnerID
			}
			snapshot.Tiles[hexID] = tileSnapshot
		}
	}

	return snapshot
}

// computeSnapshotChanges computes the diff between two game snapshots
func computeSnapshotChanges(old, new *GameSnapshot) *GameChanges {
	changes := &GameChanges{}

	if old == nil {
		changes.Status = &DiffValueString{Old: "", New: new.Status}
		changes.Phase = &DiffValueString{Old: "", New: new.Phase}
		changes.Generation = &DiffValueInt{Old: 0, New: new.Generation}
		changes.Temperature = &DiffValueInt{Old: 0, New: new.Temperature}
		changes.Oxygen = &DiffValueInt{Old: 0, New: new.Oxygen}
		changes.Oceans = &DiffValueInt{Old: 0, New: new.Oceans}

		if new.CurrentTurn != "" {
			changes.CurrentTurnPlayerID = &DiffValueString{Old: "", New: new.CurrentTurn}
		}

		changes.PlayerChanges = computeInitialSnapshotPlayerChanges(new)
		changes.BoardChanges = computeInitialSnapshotBoardChanges(new)

		return changes
	}

	changes.Status = diffString(old.Status, new.Status)
	changes.Phase = diffString(old.Phase, new.Phase)
	changes.Generation = diffInt(old.Generation, new.Generation)
	changes.Temperature = diffInt(old.Temperature, new.Temperature)
	changes.Oxygen = diffInt(old.Oxygen, new.Oxygen)
	changes.Oceans = diffInt(old.Oceans, new.Oceans)
	changes.CurrentTurnPlayerID = diffString(old.CurrentTurn, new.CurrentTurn)

	changes.PlayerChanges = computeSnapshotPlayerChanges(old, new)
	changes.BoardChanges = computeSnapshotBoardChanges(old, new)

	return changes
}

// computeInitialSnapshotPlayerChanges computes player changes for initial state
func computeInitialSnapshotPlayerChanges(snapshot *GameSnapshot) map[string]*PlayerChanges {
	playerChanges := make(map[string]*PlayerChanges)

	for playerID, player := range snapshot.Players {
		pc := computePlayerSnapshotChange(nil, player)
		if pc != nil {
			playerChanges[playerID] = pc
		}
	}

	if len(playerChanges) == 0 {
		return nil
	}
	return playerChanges
}

// computeSnapshotPlayerChanges computes changes to all players
func computeSnapshotPlayerChanges(old, new *GameSnapshot) map[string]*PlayerChanges {
	playerChanges := make(map[string]*PlayerChanges)

	for playerID, newPlayer := range new.Players {
		oldPlayer := old.Players[playerID]
		pc := computePlayerSnapshotChange(oldPlayer, newPlayer)
		if pc != nil {
			playerChanges[playerID] = pc
		}
	}

	if len(playerChanges) == 0 {
		return nil
	}
	return playerChanges
}

// computePlayerSnapshotChange computes changes for a single player
func computePlayerSnapshotChange(old, new *PlayerSnapshot) *PlayerChanges {
	pc := &PlayerChanges{}
	hasChanges := false

	var oldCredits, oldSteel, oldTitanium, oldPlants, oldEnergy, oldHeat int
	var oldTR int
	var oldCredProd, oldSteelProd, oldTitaniumProd, oldPlantsProd, oldEnergyProd, oldHeatProd int
	var oldPassed bool
	var oldCorp string
	var oldHand, oldPlayed []string

	if old != nil {
		oldCredits = old.Credits
		oldSteel = old.Steel
		oldTitanium = old.Titanium
		oldPlants = old.Plants
		oldEnergy = old.Energy
		oldHeat = old.Heat
		oldTR = old.TerraformRating
		oldCredProd = old.CreditsProduction
		oldSteelProd = old.SteelProduction
		oldTitaniumProd = old.TitaniumProduction
		oldPlantsProd = old.PlantsProduction
		oldEnergyProd = old.EnergyProduction
		oldHeatProd = old.HeatProduction
		oldPassed = old.Passed
		oldCorp = old.Corporation
		oldHand = old.HandCardIDs
		oldPlayed = old.PlayedCardIDs
	}

	if d := diffInt(oldCredits, new.Credits); d != nil {
		pc.Credits = d
		hasChanges = true
	}
	if d := diffInt(oldSteel, new.Steel); d != nil {
		pc.Steel = d
		hasChanges = true
	}
	if d := diffInt(oldTitanium, new.Titanium); d != nil {
		pc.Titanium = d
		hasChanges = true
	}
	if d := diffInt(oldPlants, new.Plants); d != nil {
		pc.Plants = d
		hasChanges = true
	}
	if d := diffInt(oldEnergy, new.Energy); d != nil {
		pc.Energy = d
		hasChanges = true
	}
	if d := diffInt(oldHeat, new.Heat); d != nil {
		pc.Heat = d
		hasChanges = true
	}

	if d := diffInt(oldCredProd, new.CreditsProduction); d != nil {
		pc.CreditsProduction = d
		hasChanges = true
	}
	if d := diffInt(oldSteelProd, new.SteelProduction); d != nil {
		pc.SteelProduction = d
		hasChanges = true
	}
	if d := diffInt(oldTitaniumProd, new.TitaniumProduction); d != nil {
		pc.TitaniumProduction = d
		hasChanges = true
	}
	if d := diffInt(oldPlantsProd, new.PlantsProduction); d != nil {
		pc.PlantsProduction = d
		hasChanges = true
	}
	if d := diffInt(oldEnergyProd, new.EnergyProduction); d != nil {
		pc.EnergyProduction = d
		hasChanges = true
	}
	if d := diffInt(oldHeatProd, new.HeatProduction); d != nil {
		pc.HeatProduction = d
		hasChanges = true
	}

	if d := diffInt(oldTR, new.TerraformRating); d != nil {
		pc.TerraformRating = d
		hasChanges = true
	}

	if d := diffBool(oldPassed, new.Passed); d != nil {
		pc.Passed = d
		hasChanges = true
	}

	if d := diffString(oldCorp, new.Corporation); d != nil {
		pc.Corporation = d
		hasChanges = true
	}

	added, removed := diffStringSlice(oldHand, new.HandCardIDs)
	if len(added) > 0 {
		pc.CardsAdded = added
		hasChanges = true
	}
	if len(removed) > 0 {
		pc.CardsRemoved = removed
		hasChanges = true
	}

	playedAdded, _ := diffStringSlice(oldPlayed, new.PlayedCardIDs)
	if len(playedAdded) > 0 {
		pc.CardsPlayed = playedAdded
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}
	return pc
}

// computeInitialSnapshotBoardChanges computes board changes for initial state
func computeInitialSnapshotBoardChanges(snapshot *GameSnapshot) *BoardChanges {
	if len(snapshot.Tiles) == 0 {
		return nil
	}

	placements := make([]TilePlacement, 0, len(snapshot.Tiles))
	for _, tile := range snapshot.Tiles {
		placements = append(placements, TilePlacement{
			HexID:    tile.HexID,
			TileType: tile.TileType,
			OwnerID:  tile.OwnerID,
		})
	}

	return &BoardChanges{TilesPlaced: placements}
}

// computeSnapshotBoardChanges computes changes to the board
func computeSnapshotBoardChanges(old, new *GameSnapshot) *BoardChanges {
	var placements []TilePlacement

	for hexID, newTile := range new.Tiles {
		if _, exists := old.Tiles[hexID]; !exists {
			placements = append(placements, TilePlacement{
				HexID:    newTile.HexID,
				TileType: newTile.TileType,
				OwnerID:  newTile.OwnerID,
			})
		}
	}

	if len(placements) == 0 {
		return nil
	}
	return &BoardChanges{TilesPlaced: placements}
}
