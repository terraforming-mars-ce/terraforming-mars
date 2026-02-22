package board

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

// Tile type string constants for placement operations
const (
	TileTypeCity     = "city"
	TileTypeGreenery = "greenery"
	TileTypeOcean    = "ocean"
)

// BoardTag represents a tag on a board tile for reserved areas
type BoardTag = string

const (
	BoardTagNoctisCity     BoardTag = "noctis-city"
	BoardTagGanymedeColony BoardTag = "ganymede-colony"
	BoardTagVolcanic       BoardTag = "volcanic"
)

// TileLocation represents the celestial body where tiles are located
type TileLocation string

const (
	// TileLocationMars represents tiles on the Mars surface
	TileLocationMars TileLocation = "mars"
)

// TileBonus represents a resource bonus provided by a tile when occupied
type TileBonus struct {
	Type   shared.ResourceType `json:"type"`
	Amount int                 `json:"amount"`
}

// TileOccupant represents what currently occupies a tile
type TileOccupant struct {
	Type shared.ResourceType `json:"type"`
	Tags []string            `json:"tags"`
}

// Tile represents a single hexagonal tile on the game board
type Tile struct {
	Coordinates shared.HexPosition  `json:"coordinates"`
	Tags        []string            `json:"tags"`
	Type        shared.ResourceType `json:"type"`
	Location    TileLocation        `json:"location"`
	DisplayName *string             `json:"displayName,omitempty"`
	Bonuses     []TileBonus         `json:"bonuses"`
	OccupiedBy  *TileOccupant       `json:"occupiedBy,omitempty"`
	OwnerID     *string             `json:"ownerId,omitempty"`
	ReservedBy  *string             `json:"reservedBy,omitempty" ts:"reservedBy"`
}

// Board represents the complete game board state with encapsulated tiles
type Board struct {
	mu       sync.RWMutex
	gameID   string
	tiles    []Tile
	eventBus *events.EventBusImpl
}

// NewBoard creates a new empty board
func NewBoard(gameID string, eventBus *events.EventBusImpl) *Board {
	return &Board{
		gameID:   gameID,
		tiles:    []Tile{},
		eventBus: eventBus,
	}
}

// NewBoardWithTiles creates a new board with the provided tiles
func NewBoardWithTiles(gameID string, tiles []Tile, eventBus *events.EventBusImpl) *Board {
	tilesCopy := make([]Tile, len(tiles))
	copy(tilesCopy, tiles)
	return &Board{
		gameID:   gameID,
		tiles:    tilesCopy,
		eventBus: eventBus,
	}
}

// GenerateMarsBoard creates the standard Terraforming Mars board layout
// Returns a hexagonal grid with ocean spaces, bonus tiles, and land tiles
func GenerateMarsBoard() []Tile {
	tiles := []Tile{}

	oceanSpaces := map[shared.HexPosition]bool{
		{Q: -4, R: 0, S: 4}:  true,
		{Q: -3, R: -1, S: 4}: true,
		{Q: -1, R: -2, S: 3}: true,
		{Q: 1, R: 1, S: -2}:  true,
		{Q: 2, R: -1, S: -1}: true,
		{Q: 3, R: -2, S: -1}: true,
		{Q: 0, R: 3, S: -3}:  true,
		{Q: -2, R: 4, S: -2}: true,
		{Q: 1, R: 3, S: -4}:  true,
	}

	bonusTiles := map[shared.HexPosition]TileBonus{
		{Q: -3, R: 1, S: 2}:  {Type: shared.ResourceSteel, Amount: 2},
		{Q: -2, R: 0, S: 2}:  {Type: shared.ResourceSteel, Amount: 2},
		{Q: 2, R: 1, S: -3}:  {Type: shared.ResourceTitanium, Amount: 3},
		{Q: 3, R: 0, S: -3}:  {Type: shared.ResourceTitanium, Amount: 3},
		{Q: -1, R: 2, S: -1}: {Type: shared.ResourcePlant, Amount: 2},
		{Q: 0, R: 2, S: -2}:  {Type: shared.ResourcePlant, Amount: 2},
		{Q: 1, R: -3, S: 2}:  {Type: shared.ResourceCardDraw, Amount: 2},
		{Q: 2, R: -3, S: 1}:  {Type: shared.ResourceCardDraw, Amount: 2},
		{Q: -1, R: -1, S: 2}: {Type: shared.ResourceCredit, Amount: 3},
	}

	type taggedTileInfo struct {
		Tags        []string
		DisplayName string
	}
	taggedTiles := map[shared.HexPosition]taggedTileInfo{
		{Q: -4, R: 2, S: 2}:  {Tags: []string{BoardTagNoctisCity}, DisplayName: "Noctis City"},
		{Q: 4, R: -2, S: -2}: {Tags: []string{BoardTagGanymedeColony}, DisplayName: "Ganymede Colony"},
		{Q: 0, R: -2, S: 2}:  {Tags: []string{BoardTagVolcanic}, DisplayName: "Tharsis Tholus"},
		{Q: -1, R: 0, S: 1}:  {Tags: []string{BoardTagVolcanic}, DisplayName: "Ascraeus Mons"},
		{Q: -2, R: 1, S: 1}:  {Tags: []string{BoardTagVolcanic}, DisplayName: "Pavonis Mons"},
		{Q: -3, R: 2, S: 1}:  {Tags: []string{BoardTagVolcanic}, DisplayName: "Arsia Mons"},
	}

	radius := 4
	for q := -radius; q <= radius; q++ {
		r1 := max(-radius, -q-radius)
		r2 := min(radius, -q+radius)

		for r := r1; r <= r2; r++ {
			s := -q - r
			pos := shared.HexPosition{Q: q, R: r, S: s}

			// Determine tile type and bonuses
			var tileType shared.ResourceType
			var bonuses []TileBonus

			if oceanSpaces[pos] {
				tileType = shared.ResourceOceanSpace
			} else {
				tileType = shared.ResourceLandTile
			}

			// Add bonus if this position has one
			if bonus, hasBonus := bonusTiles[pos]; hasBonus {
				bonuses = append(bonuses, bonus)
			}

			// Build tags and display name for tagged tiles
			var tags []string
			var displayName *string
			if tagInfo, hasTag := taggedTiles[pos]; hasTag {
				tags = tagInfo.Tags
				displayName = &tagInfo.DisplayName
			} else {
				tags = []string{}
			}

			tile := Tile{
				Coordinates: pos,
				Type:        tileType,
				Location:    TileLocationMars,
				Tags:        tags,
				DisplayName: displayName,
				Bonuses:     bonuses,
				OccupiedBy:  nil,
				OwnerID:     nil,
			}

			tiles = append(tiles, tile)
		}
	}

	return tiles
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Tiles returns a deep copy of all tiles to prevent external mutation
func (b *Board) Tiles() []Tile {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.deepCopyTiles()
}

// GetTile returns a copy of a specific tile by coordinates
func (b *Board) GetTile(coords shared.HexPosition) (*Tile, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			tileCopy := b.deepCopyTile(&b.tiles[i])
			return tileCopy, nil
		}
	}

	return nil, fmt.Errorf("tile not found at coordinates %v", coords)
}

// SetTiles replaces all tiles (used for board generation)
func (b *Board) SetTiles(ctx context.Context, tiles []Tile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	b.tiles = make([]Tile, len(tiles))
	copy(b.tiles, tiles)
	b.mu.Unlock()

	return nil
}

// UpdateTileOccupancy updates a tile's occupancy state and publishes TilePlacedEvent
func (b *Board) UpdateTileOccupancy(ctx context.Context, coords shared.HexPosition, occupant TileOccupant, ownerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var found bool

	b.mu.Lock()
	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			b.tiles[i].OccupiedBy = &occupant
			b.tiles[i].OwnerID = &ownerID
			b.tiles[i].ReservedBy = nil // Clear reservation when tile is occupied
			found = true
			break
		}
	}
	b.mu.Unlock()

	if !found {
		return fmt.Errorf("tile not found at coordinates %v", coords)
	}

	if b.eventBus != nil {
		events.Publish(b.eventBus, events.TilePlacedEvent{
			GameID:   b.gameID,
			PlayerID: ownerID,
			TileType: string(occupant.Type),
			Q:        coords.Q,
			R:        coords.R,
			S:        coords.S,
		})
	}

	return nil
}

// ClearTileOccupant removes the occupant and owner from a tile (admin debug tool)
func (b *Board) ClearTileOccupant(ctx context.Context, coords shared.HexPosition) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var found bool

	b.mu.Lock()
	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			b.tiles[i].OccupiedBy = nil
			b.tiles[i].OwnerID = nil
			found = true
			break
		}
	}
	b.mu.Unlock()

	if !found {
		return fmt.Errorf("tile not found at coordinates %v", coords)
	}

	if b.eventBus != nil {
		events.Publish(b.eventBus, events.GameStateChangedEvent{
			GameID:    b.gameID,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// ClearTileBonuses removes all bonuses from a tile after they have been claimed
func (b *Board) ClearTileBonuses(ctx context.Context, coords shared.HexPosition) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var found bool

	b.mu.Lock()
	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			b.tiles[i].Bonuses = nil
			found = true
			break
		}
	}
	b.mu.Unlock()

	if !found {
		return fmt.Errorf("tile not found at coordinates %v", coords)
	}

	return nil
}

// ReserveTile reserves a tile for exclusive future placement by a player
func (b *Board) ReserveTile(ctx context.Context, coords shared.HexPosition, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var found bool

	b.mu.Lock()
	for i := range b.tiles {
		if b.tiles[i].Coordinates == coords {
			if b.tiles[i].OccupiedBy != nil {
				b.mu.Unlock()
				return fmt.Errorf("cannot reserve tile at %v: already occupied", coords)
			}
			if b.tiles[i].ReservedBy != nil {
				b.mu.Unlock()
				return fmt.Errorf("cannot reserve tile at %v: already reserved by another player", coords)
			}
			b.tiles[i].ReservedBy = &playerID
			found = true
			break
		}
	}
	b.mu.Unlock()

	if !found {
		return fmt.Errorf("tile not found at coordinates %v", coords)
	}

	if b.eventBus != nil {
		events.Publish(b.eventBus, events.GameStateChangedEvent{
			GameID:    b.gameID,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// deepCopyTiles creates a deep copy of all tiles
func (b *Board) deepCopyTiles() []Tile {
	tiles := make([]Tile, len(b.tiles))
	for i := range b.tiles {
		tiles[i] = *b.deepCopyTile(&b.tiles[i])
	}
	return tiles
}

// deepCopyTile creates a deep copy of a single tile
func (b *Board) deepCopyTile(tile *Tile) *Tile {
	tileCopy := *tile

	tileCopy.Tags = make([]string, len(tile.Tags))
	copy(tileCopy.Tags, tile.Tags)

	tileCopy.Bonuses = make([]TileBonus, len(tile.Bonuses))
	copy(tileCopy.Bonuses, tile.Bonuses)

	if tile.DisplayName != nil {
		displayNameCopy := *tile.DisplayName
		tileCopy.DisplayName = &displayNameCopy
	}

	if tile.OccupiedBy != nil {
		occupantCopy := *tile.OccupiedBy
		occupantCopy.Tags = make([]string, len(tile.OccupiedBy.Tags))
		copy(occupantCopy.Tags, tile.OccupiedBy.Tags)
		tileCopy.OccupiedBy = &occupantCopy
	}

	if tile.OwnerID != nil {
		ownerIDCopy := *tile.OwnerID
		tileCopy.OwnerID = &ownerIDCopy
	}

	if tile.ReservedBy != nil {
		reservedByCopy := *tile.ReservedBy
		tileCopy.ReservedBy = &reservedByCopy
	}

	return &tileCopy
}
