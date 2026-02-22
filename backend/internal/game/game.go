package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

type VPCardInfo struct {
	CardID       string
	CardName     string
	Description  string
	VPConditions []player.VPCondition
	Tags         []shared.CardTag
}

type VPCardLookup interface {
	LookupVPCard(cardID string) (*VPCardInfo, error)
}

type gameVPRecalculationContext struct {
	game *Game
}

func (ctx *gameVPRecalculationContext) GetCardStorage(playerID string, cardID string) int {
	p, err := ctx.game.GetPlayer(playerID)
	if err != nil {
		return 0
	}
	return p.Resources().GetCardStorage(cardID)
}

func (ctx *gameVPRecalculationContext) CountPlayerTagsByType(playerID string, tagType shared.CardTag) int {
	p, err := ctx.game.GetPlayer(playerID)
	if err != nil {
		return 0
	}
	count := 0
	if ctx.game.vpCardLookup == nil {
		return 0
	}
	for _, cardID := range p.PlayedCards().Cards() {
		cardInfo, err := ctx.game.vpCardLookup.LookupVPCard(cardID)
		if err != nil {
			continue
		}
		for _, tag := range cardInfo.Tags {
			if tag == tagType {
				count++
			}
		}
	}
	return count
}

func (ctx *gameVPRecalculationContext) CountAllTilesOfType(tileType shared.ResourceType) int {
	tiles := ctx.game.board.Tiles()
	count := 0
	for _, tile := range tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tileType {
			count++
		}
	}
	return count
}

// Game represents a unified game entity containing all game state
// All fields are private with public methods for access and mutation
type Game struct {
	mu               sync.RWMutex
	id               string
	createdAt        time.Time
	updatedAt        time.Time
	status           GameStatus
	settings         GameSettings
	hostPlayerID     string
	currentPhase     GamePhase
	globalParameters *global_parameters.GlobalParameters
	currentTurn      *Turn // Tracks active player and available actions (nullable)
	generation       int
	board            *board.Board
	deck             *deck.Deck
	players          map[string]*player.Player
	turnOrder        []string // Ordered list of player IDs for turn sequence
	eventBus         *events.EventBusImpl

	milestones *Milestones
	awards     *Awards

	finalScores []FinalScore
	winnerID    string
	isTie       bool

	vpCardLookup VPCardLookup

	triggeredEffects []TriggeredEffect

	pendingTileSelections      map[string]*player.PendingTileSelection
	pendingTileSelectionQueues map[string]*player.PendingTileSelectionQueue
	forcedFirstActions         map[string]*player.ForcedFirstAction
	productionPhases           map[string]*player.ProductionPhase
	selectStartingCardsPhases  map[string]*player.SelectStartingCardsPhase
}

// NewGame creates a new game with the given settings
// Creates its own EventBus for synchronous event handling
func NewGame(
	id string,
	hostPlayerID string,
	settings GameSettings,
) *Game {
	now := time.Now()

	eventBus := events.NewEventBus()

	initTemp := DefaultTemperature
	initOxy := DefaultOxygen
	initOcean := DefaultOceans
	if settings.Temperature != nil {
		initTemp = *settings.Temperature
	}
	if settings.Oxygen != nil {
		initOxy = *settings.Oxygen
	}
	if settings.Oceans != nil {
		initOcean = *settings.Oceans
	}

	g := &Game{
		id:                         id,
		createdAt:                  now,
		updatedAt:                  now,
		status:                     GameStatusLobby,
		settings:                   settings,
		hostPlayerID:               hostPlayerID,
		currentPhase:               GamePhaseWaitingForGameStart,
		globalParameters:           global_parameters.NewGlobalParametersWithValues(id, initTemp, initOxy, initOcean, eventBus),
		generation:                 1,
		board:                      board.NewBoardWithTiles(id, board.GenerateMarsBoard(), eventBus),
		deck:                       nil,
		players:                    make(map[string]*player.Player),
		turnOrder:                  []string{},
		eventBus:                   eventBus,
		milestones:                 NewMilestones(id, eventBus),
		awards:                     NewAwards(id, eventBus),
		pendingTileSelections:      make(map[string]*player.PendingTileSelection),
		pendingTileSelectionQueues: make(map[string]*player.PendingTileSelectionQueue),
		forcedFirstActions:         make(map[string]*player.ForcedFirstAction),
		productionPhases:           make(map[string]*player.ProductionPhase),
		selectStartingCardsPhases:  make(map[string]*player.SelectStartingCardsPhase),
	}

	g.subscribeToGenerationalEvents()

	return g
}

// ID returns the game ID
func (g *Game) ID() string {
	return g.id
}

// CreatedAt returns when the game was created
func (g *Game) CreatedAt() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.createdAt
}

// UpdatedAt returns when the game was last updated
func (g *Game) UpdatedAt() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.updatedAt
}

// Status returns the current game status
func (g *Game) Status() GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status
}

// Settings returns a copy of the game settings
func (g *Game) Settings() GameSettings {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.settings
}

// HostPlayerID returns the host player ID
func (g *Game) HostPlayerID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hostPlayerID
}

// EventBus returns the event bus for publishing domain events
func (g *Game) EventBus() *events.EventBusImpl {
	return g.eventBus
}

// CurrentPhase returns the current game phase
func (g *Game) CurrentPhase() GamePhase {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentPhase
}

// Generation returns the current generation number
func (g *Game) Generation() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.generation
}

// TurnOrder returns a copy of the turn order
func (g *Game) TurnOrder() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	order := make([]string, len(g.turnOrder))
	copy(order, g.turnOrder)
	return order
}

// CurrentTurn returns the current turn information (may be nil)
func (g *Game) CurrentTurn() *Turn {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentTurn
}

// GlobalParameters returns the global parameters component
func (g *Game) GlobalParameters() *global_parameters.GlobalParameters {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.globalParameters
}

// Board returns the board component
func (g *Game) Board() *board.Board {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.board
}

// Deck returns the deck component
func (g *Game) Deck() *deck.Deck {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.deck
}

// SetDeck sets the deck for this game (called during initialization)
func (g *Game) SetDeck(d *deck.Deck) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.deck = d
	g.updatedAt = time.Now()
}

// Milestones returns the milestones component
func (g *Game) Milestones() *Milestones {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.milestones
}

// Awards returns the awards component
func (g *Game) Awards() *Awards {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.awards
}

// GetFinalScores returns a copy of the final scores (empty if game hasn't ended)
func (g *Game) GetFinalScores() []FinalScore {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.finalScores == nil {
		return nil
	}
	result := make([]FinalScore, len(g.finalScores))
	copy(result, g.finalScores)
	return result
}

// GetWinnerID returns the winner ID (empty if game hasn't ended)
func (g *Game) GetWinnerID() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.winnerID
}

// IsTie returns true if the game ended in a tie
func (g *Game) IsTie() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isTie
}

// GetPlayer returns a player by ID
func (g *Game) GetPlayer(playerID string) (*player.Player, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	p, exists := g.players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, g.id)
	}
	return p, nil
}

// GetAllPlayers returns all players in the game
func (g *Game) GetAllPlayers() []*player.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	players := make([]*player.Player, 0, len(g.players))
	for _, p := range g.players {
		players = append(players, p)
	}
	return players
}

// AddPlayer adds a new player to the game and publishes PlayerJoinedEvent
func (g *Game) AddPlayer(ctx context.Context, p *player.Player) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if _, exists := g.players[p.ID()]; exists {
		g.mu.Unlock()
		return fmt.Errorf("player %s already exists in game %s", p.ID(), g.id)
	}

	g.players[p.ID()] = p
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.PlayerJoinedEvent{
			GameID:   g.id,
			PlayerID: p.ID(),
		})
	}

	return nil
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if _, exists := g.players[playerID]; !exists {
		g.mu.Unlock()
		return fmt.Errorf("player %s not found in game %s", playerID, g.id)
	}

	delete(g.players, playerID)
	g.updatedAt = time.Now()
	g.mu.Unlock()

	return nil
}

// UpdateStatus updates the game status and publishes GameStatusChangedEvent
func (g *Game) UpdateStatus(ctx context.Context, newStatus GameStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldStatus GameStatus

	g.mu.Lock()
	oldStatus = g.status
	g.status = newStatus
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil && oldStatus != newStatus {
		events.Publish(g.eventBus, events.GameStatusChangedEvent{
			GameID:    g.id,
			OldStatus: string(oldStatus),
			NewStatus: string(newStatus),
		})
	}

	return nil
}

// SetFinalScores sets the final scores for the game
func (g *Game) SetFinalScores(ctx context.Context, scores []FinalScore, winnerID string, isTie bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.finalScores = make([]FinalScore, len(scores))
	copy(g.finalScores, scores)
	g.winnerID = winnerID
	g.isTie = isTie
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdatePhase updates the game phase and publishes GamePhaseChangedEvent
func (g *Game) UpdatePhase(ctx context.Context, newPhase GamePhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldPhase GamePhase

	g.mu.Lock()
	oldPhase = g.currentPhase
	g.currentPhase = newPhase
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil && oldPhase != newPhase {
		events.Publish(g.eventBus, events.GamePhaseChangedEvent{
			GameID:   g.id,
			OldPhase: string(oldPhase),
			NewPhase: string(newPhase),
		})
	}

	return nil
}

// AdvanceGeneration advances the game to the next generation and publishes GenerationAdvancedEvent
func (g *Game) AdvanceGeneration(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldGeneration, newGeneration int

	g.mu.Lock()
	oldGeneration = g.generation
	g.generation++
	newGeneration = g.generation
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GenerationAdvancedEvent{
			GameID:        g.id,
			OldGeneration: oldGeneration,
			NewGeneration: newGeneration,
		})
	}

	return nil
}

// SetGeneration sets the generation number directly and publishes GenerationAdvancedEvent
// This is used for demo/admin purposes to set an arbitrary generation
func (g *Game) SetGeneration(ctx context.Context, generation int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldGeneration, newGeneration int

	g.mu.Lock()
	oldGeneration = g.generation
	g.generation = generation
	newGeneration = g.generation
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil && oldGeneration != newGeneration {
		events.Publish(g.eventBus, events.GenerationAdvancedEvent{
			GameID:        g.id,
			OldGeneration: oldGeneration,
			NewGeneration: newGeneration,
		})
	}

	return nil
}

// SetCurrentTurn sets the current turn to a specific player with a specific action count
func (g *Game) SetCurrentTurn(ctx context.Context, playerID string, actionsRemaining int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.currentTurn = NewTurn(playerID, actionsRemaining)
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// SetTurnOrder sets the turn order for the game
func (g *Game) SetTurnOrder(ctx context.Context, turnOrder []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.turnOrder = make([]string, len(turnOrder))
	copy(g.turnOrder, turnOrder)
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// SetHostPlayerID sets the host player ID
func (g *Game) SetHostPlayerID(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.hostPlayerID = playerID
	g.updatedAt = time.Now()
	g.mu.Unlock()

	return nil
}

// NextPlayer returns the next player ID in turn order based on current turn
// Returns nil if CurrentTurn is nil, turnOrder is empty, or no players exist
func (g *Game) NextPlayer() *string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.currentTurn == nil || len(g.turnOrder) == 0 {
		return nil
	}

	currentPlayerID := g.currentTurn.PlayerID()

	currentIndex := -1
	for i, playerID := range g.turnOrder {
		if playerID == currentPlayerID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return &g.turnOrder[0]
	}

	nextIndex := (currentIndex + 1) % len(g.turnOrder)
	return &g.turnOrder[nextIndex]
}

// GetPendingTileSelection returns the pending tile selection for a player
func (g *Game) GetPendingTileSelection(playerID string) *player.PendingTileSelection {
	g.mu.RLock()
	defer g.mu.RUnlock()

	selection, exists := g.pendingTileSelections[playerID]
	if !exists || selection == nil {
		return nil
	}
	selectionCopy := *selection
	return &selectionCopy
}

// SetPendingTileSelection sets the pending tile selection for a player
func (g *Game) SetPendingTileSelection(ctx context.Context, playerID string, selection *player.PendingTileSelection) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if selection == nil {
		delete(g.pendingTileSelections, playerID)
	} else {
		selectionCopy := *selection
		g.pendingTileSelections[playerID] = &selectionCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetPendingTileSelectionQueue returns the tile selection queue for a player
func (g *Game) GetPendingTileSelectionQueue(playerID string) *player.PendingTileSelectionQueue {
	g.mu.RLock()
	defer g.mu.RUnlock()

	queue, exists := g.pendingTileSelectionQueues[playerID]
	if !exists || queue == nil {
		return nil
	}
	queueCopy := *queue
	return &queueCopy
}

// SetPendingTileSelectionQueue sets the tile selection queue for a player
func (g *Game) SetPendingTileSelectionQueue(ctx context.Context, playerID string, queue *player.PendingTileSelectionQueue) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if queue == nil {
		delete(g.pendingTileSelectionQueues, playerID)
	} else {
		queueCopy := *queue
		g.pendingTileSelectionQueues[playerID] = &queueCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	if queue != nil && len(queue.Items) > 0 {
		if err := g.ProcessNextTile(ctx, playerID); err != nil {
			return fmt.Errorf("failed to auto-process first queued tile: %w", err)
		}
	}

	return nil
}

// AppendToPendingTileSelectionQueue atomically appends tile types to a player's tile selection queue
// This is thread-safe and prevents race conditions when multiple tiles need to be queued
// tileRestrictions restricts placement based on board tags or adjacency rules (nil for normal placement)
func (g *Game) AppendToPendingTileSelectionQueue(ctx context.Context, playerID string, tileTypes []string, source string, tileRestrictions *shared.TileRestrictions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(tileTypes) == 0 {
		return nil // Nothing to append
	}

	g.mu.Lock()
	existingQueue, exists := g.pendingTileSelectionQueues[playerID]
	var items []string
	var queueSource string
	var queueTileRestrictions *shared.TileRestrictions

	if exists && existingQueue != nil {
		items = existingQueue.Items
		queueSource = existingQueue.Source
		queueTileRestrictions = existingQueue.TileRestrictions
	} else {
		items = []string{}
		queueSource = source
		queueTileRestrictions = tileRestrictions
	}

	items = append(items, tileTypes...)

	g.pendingTileSelectionQueues[playerID] = &player.PendingTileSelectionQueue{
		Items:            items,
		TileRestrictions: queueTileRestrictions,
		Source:           queueSource,
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	// Automatically process first tile if this was the first tile added to an empty queue
	if !exists || existingQueue == nil || len(existingQueue.Items) == 0 {
		if err := g.ProcessNextTile(ctx, playerID); err != nil {
			return fmt.Errorf("failed to auto-process first queued tile: %w", err)
		}
	}

	return nil
}

// GetForcedFirstAction returns the forced first action for a player
func (g *Game) GetForcedFirstAction(playerID string) *player.ForcedFirstAction {
	g.mu.RLock()
	defer g.mu.RUnlock()

	action, exists := g.forcedFirstActions[playerID]
	if !exists || action == nil {
		return nil
	}
	actionCopy := *action
	return &actionCopy
}

// SetForcedFirstAction sets the forced first action for a player
func (g *Game) SetForcedFirstAction(ctx context.Context, playerID string, action *player.ForcedFirstAction) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if action == nil {
		delete(g.forcedFirstActions, playerID)
	} else {
		actionCopy := *action
		g.forcedFirstActions[playerID] = &actionCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetProductionPhase returns the production phase state for a player
func (g *Game) GetProductionPhase(playerID string) *player.ProductionPhase {
	g.mu.RLock()
	defer g.mu.RUnlock()

	phase, exists := g.productionPhases[playerID]
	if !exists || phase == nil {
		return nil
	}
	phaseCopy := *phase
	return &phaseCopy
}

// SetProductionPhase sets the production phase state for a player
func (g *Game) SetProductionPhase(ctx context.Context, playerID string, phase *player.ProductionPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if phase == nil {
		delete(g.productionPhases, playerID)
	} else {
		phaseCopy := *phase
		g.productionPhases[playerID] = &phaseCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetSelectStartingCardsPhase returns the select starting cards phase state for a player
func (g *Game) GetSelectStartingCardsPhase(playerID string) *player.SelectStartingCardsPhase {
	g.mu.RLock()
	defer g.mu.RUnlock()

	phase, exists := g.selectStartingCardsPhases[playerID]
	if !exists || phase == nil {
		return nil
	}
	phaseCopy := *phase
	return &phaseCopy
}

// SetSelectStartingCardsPhase sets the select starting cards phase state for a player
func (g *Game) SetSelectStartingCardsPhase(ctx context.Context, playerID string, phase *player.SelectStartingCardsPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if phase == nil {
		delete(g.selectStartingCardsPhases, playerID)
	} else {
		phaseCopy := *phase
		g.selectStartingCardsPhases[playerID] = &phaseCopy
	}
	g.updatedAt = time.Now()
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// ProcessNextTile pops the next tile from a player's tile queue and creates a PendingTileSelection
// This method converts the queue into an actionable tile selection for the player
func (g *Game) ProcessNextTile(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	queue, exists := g.pendingTileSelectionQueues[playerID]
	if !exists || queue == nil || len(queue.Items) == 0 {
		g.mu.Unlock()
		return nil
	}

	nextTileType := queue.Items[0]
	remainingItems := queue.Items[1:]
	source := queue.Source
	onComplete := queue.OnComplete
	tileRestrictions := queue.TileRestrictions

	if len(remainingItems) > 0 {
		g.pendingTileSelectionQueues[playerID] = &player.PendingTileSelectionQueue{
			Items:            remainingItems,
			TileRestrictions: tileRestrictions,
			Source:           source,
			OnComplete:       onComplete,
		}
	} else {
		delete(g.pendingTileSelectionQueues, playerID)
	}
	g.mu.Unlock()

	availableHexes := g.calculateAvailableHexesForTile(nextTileType, playerID, tileRestrictions)

	if len(availableHexes) == 0 {
		return g.ProcessNextTile(ctx, playerID)
	}

	err := g.SetPendingTileSelection(ctx, playerID, &player.PendingTileSelection{
		TileType:       nextTileType,
		AvailableHexes: availableHexes,
		Source:         source,
		OnComplete:     onComplete,
	})

	return err
}

// calculateAvailableHexesForTile returns a list of valid hex positions for placing a tile
// tileRestrictions controls placement rules:
//   - BoardTags: restricts to tiles with matching tags (e.g., Noctis City)
//   - Adjacency: "none" means no adjacent occupied tiles allowed (Research Outpost)
//
// For cities: if BoardTags is set, only matching tiles are valid (ignoring adjacency);
// if Adjacency is "none", tiles must have no adjacent occupied tiles;
// otherwise, tagged tiles (reserved areas) are excluded and normal adjacency rules apply
func (g *Game) calculateAvailableHexesForTile(tileType string, playerID string, tileRestrictions *shared.TileRestrictions) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.board == nil {
		return []string{}
	}

	tiles := g.board.Tiles()
	availableHexes := []string{}

	// Extract restrictions
	var boardTags []string
	var adjacency string
	if tileRestrictions != nil {
		boardTags = tileRestrictions.BoardTags
		adjacency = tileRestrictions.Adjacency
	}

	// Helper to check if tile has any of the required board tags
	tileHasRequiredTag := func(tile board.Tile, requiredTags []string) bool {
		for _, reqTag := range requiredTags {
			for _, tileTag := range tile.Tags {
				if tileTag == reqTag {
					return true
				}
			}
		}
		return false
	}

	// Helper to check if tile has any tags (is a reserved area)
	tileHasAnyTag := func(tile board.Tile) bool {
		return len(tile.Tags) > 0
	}

	// Helper to check if a tile has any adjacent occupied tiles
	hasAnyAdjacentOccupied := func(tile board.Tile) bool {
		for _, neighborPos := range tile.Coordinates.GetNeighbors() {
			for _, neighborTile := range tiles {
				if neighborTile.Coordinates.Equals(neighborPos) && neighborTile.OccupiedBy != nil {
					return true
				}
			}
		}
		return false
	}

	// Helper to check if tile is reserved by another player
	isReservedByOther := func(tile board.Tile) bool {
		return tile.ReservedBy != nil && *tile.ReservedBy != playerID
	}

	for _, tile := range tiles {
		// Clear targets occupied tiles (inverse of normal placement)
		if tileType == "clear" {
			if tile.OccupiedBy != nil {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
			continue
		}

		// Skip tiles that are already occupied
		if tile.OccupiedBy != nil {
			continue
		}

		switch tileType {
		case "land-claim":
			// Land claim can only be placed on unoccupied, unreserved land tiles
			if tile.Type != shared.ResourceLandTile {
				continue
			}
			// Exclude reserved areas (tagged tiles like Noctis City)
			if tileHasAnyTag(tile) {
				continue
			}
			// Exclude tiles already reserved by anyone
			if tile.ReservedBy != nil {
				continue
			}
			availableHexes = append(availableHexes, tile.Coordinates.String())

		case "city":
			if tile.Type != shared.ResourceLandTile {
				continue
			}

			// If boardTags specified, only allow tiles with matching tags (Noctis City case)
			if len(boardTags) > 0 {
				if tileHasRequiredTag(tile, boardTags) {
					availableHexes = append(availableHexes, tile.Coordinates.String())
					logger.Get().Debug("✅ Tile available for city (board tag match)",
						zap.String("tile", tile.Coordinates.String()),
						zap.Strings("board_tags", boardTags))
				}
				continue
			}

			// Skip tiles reserved by other players (current player can use their own reserved tiles)
			if isReservedByOther(tile) {
				continue
			}

			// Normal city placement: exclude reserved areas (tagged tiles)
			if tileHasAnyTag(tile) {
				logger.Get().Debug("⏭️ Skipping reserved tile for normal city placement",
					zap.String("tile", tile.Coordinates.String()),
					zap.Strings("tile_tags", tile.Tags))
				continue
			}

			// Handle "no adjacent tiles" restriction (Research Outpost)
			if adjacency == "none" {
				if !hasAnyAdjacentOccupied(tile) {
					availableHexes = append(availableHexes, tile.Coordinates.String())
					logger.Get().Debug("✅ Tile available for city (no adjacent tiles)",
						zap.String("tile", tile.Coordinates.String()))
				}
				continue // Skip normal city adjacency rules
			}

			// Check city adjacency rule (no adjacent cities)
			hasAdjacentCity := false
			neighbors := tile.Coordinates.GetNeighbors()

			logger.Get().Debug("🔍 Checking city placement",
				zap.String("tile", tile.Coordinates.String()),
				zap.Int("neighbor_count", len(neighbors)))

			for _, neighborPos := range neighbors {
				for _, neighborTile := range tiles {
					if neighborTile.Coordinates.Equals(neighborPos) {
						occupantType := ""
						if neighborTile.OccupiedBy != nil {
							occupantType = string(neighborTile.OccupiedBy.Type)
						}

						logger.Get().Debug("🔎 Checking neighbor",
							zap.String("neighbor_pos", neighborPos.String()),
							zap.String("neighbor_tile", neighborTile.Coordinates.String()),
							zap.Bool("occupied", neighborTile.OccupiedBy != nil),
							zap.String("occupant_type", occupantType))

						if neighborTile.OccupiedBy != nil && neighborTile.OccupiedBy.Type == shared.ResourceCityTile {
							hasAdjacentCity = true
							logger.Get().Info("🚫 City adjacency violation detected",
								zap.String("tile", tile.Coordinates.String()),
								zap.String("adjacent_city", neighborTile.Coordinates.String()))
							break
						}
					}
				}
				if hasAdjacentCity {
					break
				}
			}

			if !hasAdjacentCity {
				availableHexes = append(availableHexes, tile.Coordinates.String())
				logger.Get().Debug("✅ Tile available for city",
					zap.String("tile", tile.Coordinates.String()))
			} else {
				logger.Get().Debug("❌ Tile unavailable for city (adjacent city)",
					zap.String("tile", tile.Coordinates.String()))
			}

		case "greenery":
			// Check if restricted to ocean tiles (Mangrove card)
			if tileRestrictions != nil && tileRestrictions.OnTileType == "ocean" {
				if tile.Type == shared.ResourceOceanSpace {
					availableHexes = append(availableHexes, tile.Coordinates.String())
				}
				continue
			}

			// Skip tiles reserved by other players (current player can use their own reserved tiles)
			if isReservedByOther(tile) {
				continue
			}

			// Exclude reserved areas from normal greenery placement
			if len(boardTags) == 0 && tileHasAnyTag(tile) {
				continue
			}
			if tile.Type == shared.ResourceLandTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		case "ocean":
			if tile.Type == shared.ResourceOceanSpace {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		case "volcano":
			if !tileHasRequiredTag(tile, []string{board.BoardTagVolcanic}) {
				continue
			}
			if tile.Type == shared.ResourceLandTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		default:
			// Skip tiles reserved by other players (current player can use their own reserved tiles)
			if isReservedByOther(tile) {
				continue
			}
			// Exclude reserved areas from normal placement
			if len(boardTags) == 0 && tileHasAnyTag(tile) {
				continue
			}
			if tile.Type == shared.ResourceLandTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
		}
	}

	// If boardTags specified but no matching tiles found (e.g., Noctis City already occupied),
	// fall back to normal placement rules
	if len(boardTags) > 0 && len(availableHexes) == 0 {
		logger.Get().Info("🔄 No tiles match board tags, falling back to normal placement",
			zap.Strings("board_tags", boardTags),
			zap.String("tile_type", tileType))
		return g.calculateAvailableHexesForTile(tileType, playerID, nil)
	}

	return availableHexes
}

// CountAvailableHexesForTile returns the number of valid hex positions for placing a tile
// This is used by state calculators to determine if tile-placing actions are available
func (g *Game) CountAvailableHexesForTile(tileType string, playerID string, tileRestrictions *shared.TileRestrictions) int {
	return len(g.calculateAvailableHexesForTile(tileType, playerID, tileRestrictions))
}

// TriggeredEffect represents a card effect that was triggered (for frontend notifications)
type TriggeredEffect struct {
	CardName string
	PlayerID string
	Outputs  []shared.ResourceCondition
}

// AddTriggeredEffect records a triggered effect for notification
func (g *Game) AddTriggeredEffect(effect TriggeredEffect) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.triggeredEffects = append(g.triggeredEffects, effect)
}

// GetTriggeredEffects returns all triggered effects since last clear
func (g *Game) GetTriggeredEffects() []TriggeredEffect {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]TriggeredEffect, len(g.triggeredEffects))
	copy(result, g.triggeredEffects)
	return result
}

// ClearTriggeredEffects clears the triggered effects list
func (g *Game) ClearTriggeredEffects() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.triggeredEffects = nil
}

func (g *Game) SetVPCardLookup(lookup VPCardLookup) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.vpCardLookup = lookup
	g.subscribeToVPEvents()
}

func (g *Game) recalculatePlayerVP(p *player.Player) {
	if g.vpCardLookup == nil {
		return
	}
	ctx := &gameVPRecalculationContext{game: g}
	p.VPGranters().RecalculateAll(ctx)
}

func (g *Game) recalculateAllPlayersVP() {
	for _, p := range g.GetAllPlayers() {
		g.recalculatePlayerVP(p)
	}
}

func (g *Game) subscribeToVPEvents() {
	events.Subscribe(g.eventBus, func(e events.CardPlayedEvent) {
		if g.vpCardLookup == nil {
			return
		}
		cardInfo, err := g.vpCardLookup.LookupVPCard(e.CardID)
		if err != nil {
			return
		}
		if len(cardInfo.VPConditions) == 0 {
			return
		}

		p, err := g.GetPlayer(e.PlayerID)
		if err != nil {
			return
		}

		granter := player.VPGranter{
			CardID:       cardInfo.CardID,
			CardName:     cardInfo.CardName,
			Description:  cardInfo.Description,
			VPConditions: cardInfo.VPConditions,
		}
		p.VPGranters().Add(granter)
		g.recalculatePlayerVP(p)
	})

	events.Subscribe(g.eventBus, func(e events.ResourceStorageChangedEvent) {
		p, err := g.GetPlayer(e.PlayerID)
		if err != nil {
			return
		}
		g.recalculatePlayerVP(p)
	})

	events.Subscribe(g.eventBus, func(e events.TagPlayedEvent) {
		p, err := g.GetPlayer(e.PlayerID)
		if err != nil {
			return
		}
		g.recalculatePlayerVP(p)
	})

	events.Subscribe(g.eventBus, func(e events.TilePlacedEvent) {
		g.recalculateAllPlayersVP()
	})

	events.Subscribe(g.eventBus, func(e events.CorporationSelectedEvent) {
		if g.vpCardLookup == nil {
			return
		}
		cardInfo, err := g.vpCardLookup.LookupVPCard(e.CorporationID)
		if err != nil {
			return
		}
		if len(cardInfo.VPConditions) == 0 {
			return
		}

		p, err := g.GetPlayer(e.PlayerID)
		if err != nil {
			return
		}

		granter := player.VPGranter{
			CardID:       cardInfo.CardID,
			CardName:     cardInfo.CardName,
			Description:  cardInfo.Description,
			VPConditions: cardInfo.VPConditions,
		}
		p.VPGranters().Prepend(granter)
		g.recalculatePlayerVP(p)
	})
}

func (g *Game) subscribeToGenerationalEvents() {
	events.Subscribe(g.eventBus, func(e events.TerraformRatingChangedEvent) {
		if e.NewRating > e.OldRating {
			p, err := g.GetPlayer(e.PlayerID)
			if err != nil {
				return
			}
			p.GenerationalEvents().Increment(shared.GenerationalEventTRRaise)
		}
	})

	events.Subscribe(g.eventBus, func(e events.TilePlacedEvent) {
		p, err := g.GetPlayer(e.PlayerID)
		if err != nil {
			return
		}

		switch e.TileType {
		case "ocean":
			p.GenerationalEvents().Increment(shared.GenerationalEventOceanPlacement)
		case "city":
			p.GenerationalEvents().Increment(shared.GenerationalEventCityPlacement)
		case "greenery":
			p.GenerationalEvents().Increment(shared.GenerationalEventGreeneryPlacement)
		}
	})

	events.Subscribe(g.eventBus, func(e events.GenerationAdvancedEvent) {
		for _, p := range g.GetAllPlayers() {
			p.GenerationalEvents().Clear()
		}
	})
}
