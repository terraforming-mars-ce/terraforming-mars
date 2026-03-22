package game

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/projectfunding"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

type VPCardInfo struct {
	CardID       string
	CardName     string
	CardType     string
	Description  string
	VPConditions []shared.VPCondition
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
		if cardInfo.CardType == "event" && tagType != shared.TagEvent {
			continue
		}
		for _, tag := range cardInfo.Tags {
			if tag == tagType || tag == shared.TagWild {
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

func (ctx *gameVPRecalculationContext) CountPlayerTilesOfType(playerID string, tileType shared.ResourceType) int {
	tiles := ctx.game.board.Tiles()
	count := 0
	for _, tile := range tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tileType &&
			tile.OwnerID != nil && *tile.OwnerID == playerID {
			count++
		}
	}
	return count
}

func (ctx *gameVPRecalculationContext) CountAdjacentTilesForCard(cardID string, tileType shared.ResourceType) int {
	tiles := ctx.game.board.Tiles()
	sourceTag := "source:" + cardID

	var sourceTile *board.Tile
	for i := range tiles {
		if tiles[i].OccupiedBy == nil {
			continue
		}
		for _, tag := range tiles[i].OccupiedBy.Tags {
			if tag == sourceTag {
				sourceTile = &tiles[i]
				break
			}
		}
		if sourceTile != nil {
			break
		}
	}

	if sourceTile == nil {
		return 0
	}

	neighbors := sourceTile.Coordinates.GetNeighbors()
	count := 0
	for _, tile := range tiles {
		if tile.OccupiedBy == nil || tile.OccupiedBy.Type != tileType {
			continue
		}
		for _, neighbor := range neighbors {
			if tile.Coordinates == neighbor {
				count++
				break
			}
		}
	}
	return count
}

type Game struct {
	mu               sync.RWMutex
	ds               *datastore.DataStore
	id               string
	globalParameters *global_parameters.GlobalParameters
	currentTurn      *Turn
	board            *board.Board
	deck             *deck.Deck
	players          map[string]*player.Player
	eventBus         *events.EventBusImpl
	milestones       *Milestones
	awards           *Awards
	vpCardLookup     VPCardLookup
}

func (g *Game) update(fn func(s *datastore.GameState)) {
	if err := g.ds.UpdateGame(g.id, fn); err != nil {
		logger.Get().Warn("Failed to update game state", zap.String("game_id", g.id), zap.Error(err))
	}
}

func (g *Game) read(fn func(s *datastore.GameState)) {
	if err := g.ds.ReadGame(g.id, fn); err != nil {
		logger.Get().Warn("Failed to read game state", zap.String("game_id", g.id), zap.Error(err))
	}
}

// NewGame creates a new game with the given settings.
// The DataStore is used as the single point of entry for all state reads/writes.
func NewGame(
	ds *datastore.DataStore,
	id string,
	hostPlayerID string,
	settings shared.GameSettings,
) *Game {
	now := time.Now()

	eventBus := events.NewEventBus()

	initTemp := DefaultTemperature
	initOxy := DefaultOxygen
	initOcean := DefaultOceans
	initVenus := DefaultVenus
	if settings.Temperature != nil {
		initTemp = *settings.Temperature
	}
	if settings.Oxygen != nil {
		initOxy = *settings.Oxygen
	}
	if settings.Oceans != nil {
		initOcean = *settings.Oceans
	}
	if settings.Venus != nil {
		initVenus = *settings.Venus
	}

	state := &datastore.GameState{
		ID:                         id,
		CreatedAt:                  now,
		UpdatedAt:                  now,
		Status:                     shared.GameStatusLobby,
		Settings:                   settings,
		HostPlayerID:               hostPlayerID,
		CurrentPhase:               shared.GamePhaseWaitingForGameStart,
		Generation:                 1,
		Temperature:                initTemp,
		Oxygen:                     initOxy,
		Oceans:                     initOcean,
		MaxOceans:                  global_parameters.MaxOceans,
		Venus:                      initVenus,
		PlayerOrder:                []string{},
		TurnOrder:                  []string{},
		Players:                    make(map[string]*datastore.PlayerState),
		ClaimedMilestones:          []shared.ClaimedMilestone{},
		FundedAwards:               []shared.FundedAward{},
		Spectators:                 make(map[string]*shared.SpectatorState),
		ChatMessages:               []shared.ChatMessage{},
		PendingTileSelections:      make(map[string]*shared.PendingTileSelection),
		PendingTileSelectionQueues: make(map[string]*shared.PendingTileSelectionQueue),
		ForcedFirstActions:         make(map[string]*shared.ForcedFirstAction),
		ProductionPhases:           make(map[string]*shared.ProductionPhase),
		SelectCorporationPhases:    make(map[string]*shared.SelectCorporationPhase),
		SelectStartingCardsPhases:  make(map[string]*shared.SelectStartingCardsPhase),
		SelectPreludeCardsPhases:   make(map[string]*shared.SelectPreludeCardsPhase),
		DeferredStartingChoices:    make(map[string]*shared.DeferredStartingChoices),
		TradeFleets:                make(map[string]bool),
	}

	// Insert state into DataStore so components can read/write through it
	txn := ds.BeginTxn()
	if err := txn.InsertGame(state); err != nil {
		logger.Get().Error("Failed to insert game state", zap.String("game_id", id), zap.Error(err))
	}
	txn.Commit()

	ds.RecordInitialHistory(state)

	g := &Game{
		ds:               ds,
		id:               id,
		globalParameters: global_parameters.NewGlobalParameters(ds, id, eventBus),
		board:            board.NewBoardWithTiles(&state.Tiles, id, board.GenerateMarsBoard(settings.VenusNextEnabled), eventBus),
		players:          make(map[string]*player.Player),
		eventBus:         eventBus,
		milestones:       NewMilestones(ds, id, eventBus),
		awards:           NewAwards(ds, id, eventBus),
	}

	g.subscribeToGenerationalEvents()
	g.subscribeToOceanSpaceEvents()
	g.subscribeToGlobalParameterBonuses()

	return g
}

// ID returns the game ID
func (g *Game) ID() string {
	return g.id
}

func (g *Game) CreatedAt() time.Time {
	var v time.Time
	g.read(func(s *datastore.GameState) { v = s.CreatedAt })
	return v
}

func (g *Game) UpdatedAt() time.Time {
	var v time.Time
	g.read(func(s *datastore.GameState) { v = s.UpdatedAt })
	return v
}

func (g *Game) Status() shared.GameStatus {
	var v shared.GameStatus
	g.read(func(s *datastore.GameState) { v = s.Status })
	return v
}

func (g *Game) Settings() shared.GameSettings {
	var v shared.GameSettings
	g.read(func(s *datastore.GameState) { v = s.Settings })
	return v
}

func (g *Game) UpdateSettings(ctx context.Context, settings shared.GameSettings) {
	g.update(func(s *datastore.GameState) {
		s.Settings = settings
		s.UpdatedAt = time.Now()
	})
}

func (g *Game) HostPlayerID() string {
	var v string
	g.read(func(s *datastore.GameState) { v = s.HostPlayerID })
	return v
}

func (g *Game) State() *datastore.GameState {
	state, _ := g.ds.GetGame(g.id)
	return state
}

// EventBus returns the event bus for publishing domain events
func (g *Game) EventBus() *events.EventBusImpl {
	return g.eventBus
}

func (g *Game) CurrentPhase() shared.GamePhase {
	var v shared.GamePhase
	g.read(func(s *datastore.GameState) { v = s.CurrentPhase })
	return v
}

func (g *Game) Generation() int {
	var v int
	g.read(func(s *datastore.GameState) { v = s.Generation })
	return v
}

func (g *Game) PlayerOrder() []string {
	var order []string
	g.read(func(s *datastore.GameState) {
		order = make([]string, len(s.PlayerOrder))
		copy(order, s.PlayerOrder)
	})
	return order
}

func (g *Game) TurnOrder() []string {
	var order []string
	g.read(func(s *datastore.GameState) {
		order = make([]string, len(s.TurnOrder))
		copy(order, s.TurnOrder)
	})
	return order
}

func (g *Game) CurrentTurn() *Turn {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentTurn
}

func (g *Game) GlobalParameters() *global_parameters.GlobalParameters {
	return g.globalParameters
}

func (g *Game) Board() *board.Board {
	return g.board
}

func (g *Game) Deck() *deck.Deck {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.deck
}

func (g *Game) SetDeck(d *deck.Deck) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.deck = d
	g.update(func(s *datastore.GameState) { s.UpdatedAt = time.Now() })
}

func (g *Game) InitDeck(projectCardIDs, corpIDs, preludeIDs []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.deck = deck.NewDeck(g.ds, g.id, projectCardIDs, corpIDs, preludeIDs)
	g.update(func(s *datastore.GameState) { s.UpdatedAt = time.Now() })
}

func (g *Game) Milestones() *Milestones {
	return g.milestones
}

func (g *Game) Awards() *Awards {
	return g.awards
}

func (g *Game) GetFinalScores() []shared.FinalScore {
	var result []shared.FinalScore
	g.read(func(s *datastore.GameState) {
		if s.FinalScores == nil {
			return
		}
		result = make([]shared.FinalScore, len(s.FinalScores))
		copy(result, s.FinalScores)
	})
	return result
}

func (g *Game) GetWinnerID() string {
	var v string
	g.read(func(s *datastore.GameState) { v = s.WinnerID })
	return v
}

func (g *Game) IsTie() bool {
	var v bool
	g.read(func(s *datastore.GameState) { v = s.IsTie })
	return v
}

func (g *Game) GetPlayer(playerID string) (*player.Player, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	p, exists := g.players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found in game %s", playerID, g.id)
	}
	return p, nil
}

func (g *Game) GetAllPlayers() []*player.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	playerOrder := g.PlayerOrder()
	players := make([]*player.Player, 0, len(playerOrder))
	for _, id := range playerOrder {
		if p, exists := g.players[id]; exists {
			players = append(players, p)
		}
	}
	return players
}

// AddNewPlayer creates a new human player backed by this game's state and adds them to the game
func (g *Game) AddNewPlayer(ctx context.Context, playerID, playerName string) (*player.Player, error) {
	if err := g.ds.UpdateGame(g.id, func(s *datastore.GameState) {
		s.Players[playerID] = &datastore.PlayerState{
			ID:                 playerID,
			Name:               playerName,
			Connected:          true,
			PlayerType:         "human",
			TerraformRating:    20,
			HandCardIDs:        []string{},
			PlayedCardIDs:      []string{},
			ResourceStorage:    make(map[string]int),
			BonusTags:          make(map[shared.CardTag]int),
			GenerationalEvents: make(map[shared.GenerationalEvent]int),
		}
	}); err != nil {
		logger.Get().Error("Failed to add player to game state", zap.String("game_id", g.id), zap.String("player_id", playerID), zap.Error(err))
	}
	p := player.NewPlayer(g.ds, g.id, playerID, g.eventBus)
	if err := g.AddPlayer(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// AddNewBotPlayer creates a new bot player backed by this game's state and adds them to the game
func (g *Game) AddNewBotPlayer(ctx context.Context, botID, botName string, difficulty player.BotDifficulty, speed player.BotSpeed) (*player.Player, error) {
	if err := g.ds.UpdateGame(g.id, func(s *datastore.GameState) {
		s.Players[botID] = &datastore.PlayerState{
			ID:                 botID,
			Name:               botName,
			Connected:          false,
			PlayerType:         "bot",
			BotStatus:          string(player.BotStatusLoading),
			BotDifficulty:      string(difficulty),
			BotSpeed:           string(speed),
			TerraformRating:    20,
			HandCardIDs:        []string{},
			PlayedCardIDs:      []string{},
			ResourceStorage:    make(map[string]int),
			BonusTags:          make(map[shared.CardTag]int),
			GenerationalEvents: make(map[shared.GenerationalEvent]int),
		}
	}); err != nil {
		logger.Get().Error("Failed to add bot player to game state", zap.String("game_id", g.id), zap.String("bot_id", botID), zap.Error(err))
	}
	p := player.NewPlayer(g.ds, g.id, botID, g.eventBus)
	if err := g.AddPlayer(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (g *Game) AddPlayer(ctx context.Context, p *player.Player) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	if _, exists := g.players[p.ID()]; exists {
		g.mu.Unlock()
		return fmt.Errorf("player %s already exists in game %s", p.ID(), g.id)
	}

	if p.Color() == "" {
		taken := make(map[string]bool, len(g.players))
		for _, existing := range g.players {
			if existing.Color() != "" {
				taken[existing.Color()] = true
			}
		}
		for _, c := range shared.PlayerColors {
			if !taken[c] {
				p.SetColor(c)
				break
			}
		}
	}

	g.players[p.ID()] = p
	g.update(func(s *datastore.GameState) {
		s.PlayerOrder = append(s.PlayerOrder, p.ID())
		s.UpdatedAt = time.Now()
	})
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.PlayerJoinedEvent{
			GameID:   g.id,
			PlayerID: p.ID(),
		})
	}

	return nil
}

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
	g.update(func(s *datastore.GameState) {
		for i, id := range s.PlayerOrder {
			if id == playerID {
				s.PlayerOrder = append(s.PlayerOrder[:i], s.PlayerOrder[i+1:]...)
				break
			}
		}
		s.UpdatedAt = time.Now()
	})
	g.mu.Unlock()

	return nil
}

func (g *Game) AddSpectator(ctx context.Context, s *Spectator) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var addErr error
	g.update(func(state *datastore.GameState) {
		if len(state.Spectators) >= shared.MaxSpectators {
			addErr = fmt.Errorf("game %s already has the maximum number of spectators (%d)", g.id, shared.MaxSpectators)
			return
		}
		if _, exists := state.Spectators[s.ID()]; exists {
			addErr = fmt.Errorf("spectator %s already exists in game %s", s.ID(), g.id)
			return
		}
		state.Spectators[s.ID()] = &shared.SpectatorState{
			ID:    s.ID(),
			Name:  s.Name(),
			Color: s.Color(),
		}
		state.UpdatedAt = time.Now()
	})
	return addErr
}

func (g *Game) RemoveSpectator(ctx context.Context, spectatorID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var removeErr error
	g.update(func(state *datastore.GameState) {
		if _, exists := state.Spectators[spectatorID]; !exists {
			removeErr = fmt.Errorf("spectator %s not found in game %s", spectatorID, g.id)
			return
		}
		delete(state.Spectators, spectatorID)
		state.UpdatedAt = time.Now()
	})
	return removeErr
}

func (g *Game) GetSpectator(spectatorID string) (*Spectator, error) {
	var result *Spectator
	var getErr error
	g.read(func(state *datastore.GameState) {
		ss, exists := state.Spectators[spectatorID]
		if !exists {
			getErr = fmt.Errorf("spectator %s not found in game %s", spectatorID, g.id)
			return
		}
		result = NewSpectator(ss.ID, ss.Name, ss.Color)
	})
	return result, getErr
}

func (g *Game) GetAllSpectators() []*Spectator {
	var spectators []*Spectator
	g.read(func(state *datastore.GameState) {
		spectators = make([]*Spectator, 0, len(state.Spectators))
		for _, ss := range state.Spectators {
			spectators = append(spectators, NewSpectator(ss.ID, ss.Name, ss.Color))
		}
	})
	return spectators
}

func (g *Game) SpectatorCount() int {
	var count int
	g.read(func(state *datastore.GameState) { count = len(state.Spectators) })
	return count
}

func (g *Game) NextSpectatorColor() string {
	var color string
	g.read(func(state *datastore.GameState) {
		idx := len(state.Spectators) % len(shared.SpectatorColors)
		color = shared.SpectatorColors[idx]
	})
	return color
}

// IsPlayerColorAvailable returns true if the given color is in the palette and
// not taken by another player (excluding the specified player).
func (g *Game) IsPlayerColorAvailable(color string, excludePlayerID string) bool {
	validColor := false
	for _, c := range shared.PlayerColors {
		if c == color {
			validColor = true
			break
		}
	}
	if !validColor {
		return false
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, p := range g.players {
		if p.ID() != excludePlayerID && p.Color() == color {
			return false
		}
	}
	return true
}

func (g *Game) AddChatMessage(ctx context.Context, msg shared.ChatMessage) {
	if ctx.Err() != nil {
		return
	}

	g.update(func(s *datastore.GameState) {
		s.ChatMessages = append(s.ChatMessages, msg)
		if len(s.ChatMessages) > shared.MaxChatMessages {
			s.ChatMessages = s.ChatMessages[len(s.ChatMessages)-shared.MaxChatMessages:]
		}
		s.UpdatedAt = time.Now()
	})
}

func (g *Game) GetChatMessages() []shared.ChatMessage {
	var msgs []shared.ChatMessage
	g.read(func(s *datastore.GameState) {
		msgs = make([]shared.ChatMessage, len(s.ChatMessages))
		copy(msgs, s.ChatMessages)
	})
	return msgs
}

func (g *Game) UpdateStatus(ctx context.Context, newStatus shared.GameStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldStatus shared.GameStatus
	g.update(func(s *datastore.GameState) {
		oldStatus = s.Status
		s.Status = newStatus
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil && oldStatus != newStatus {
		events.Publish(g.eventBus, events.GameStatusChangedEvent{
			GameID:    g.id,
			OldStatus: string(oldStatus),
			NewStatus: string(newStatus),
		})
	}

	return nil
}

func (g *Game) SetFinalScores(ctx context.Context, scores []shared.FinalScore, winnerID string, isTie bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		s.FinalScores = make([]shared.FinalScore, len(scores))
		copy(s.FinalScores, scores)
		s.WinnerID = winnerID
		s.IsTie = isTie
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) UpdatePhase(ctx context.Context, newPhase shared.GamePhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldPhase shared.GamePhase
	g.update(func(s *datastore.GameState) {
		oldPhase = s.CurrentPhase
		s.CurrentPhase = newPhase
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil && oldPhase != newPhase {
		events.Publish(g.eventBus, events.GamePhaseChangedEvent{
			GameID:   g.id,
			OldPhase: string(oldPhase),
			NewPhase: string(newPhase),
		})
	}

	return nil
}

func (g *Game) AdvanceGeneration(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldGeneration, newGeneration int
	g.update(func(s *datastore.GameState) {
		oldGeneration = s.Generation
		s.Generation++
		newGeneration = s.Generation
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GenerationAdvancedEvent{
			GameID:        g.id,
			OldGeneration: oldGeneration,
			NewGeneration: newGeneration,
		})
	}

	return nil
}

func (g *Game) SetGeneration(ctx context.Context, generation int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var oldGeneration, newGeneration int
	g.update(func(s *datastore.GameState) {
		oldGeneration = s.Generation
		s.Generation = generation
		newGeneration = s.Generation
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil && oldGeneration != newGeneration {
		events.Publish(g.eventBus, events.GenerationAdvancedEvent{
			GameID:        g.id,
			OldGeneration: oldGeneration,
			NewGeneration: newGeneration,
		})
	}

	return nil
}

func (g *Game) SetCurrentTurn(ctx context.Context, playerID string, actionsRemaining int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.mu.Lock()
	g.update(func(s *datastore.GameState) {
		s.CurrentTurnPlayerID = playerID
		s.CurrentTurnActions = actionsRemaining
		s.CurrentTurnTotalActions = actionsRemaining
		s.UpdatedAt = time.Now()
	})
	g.currentTurn = NewTurn(g.ds, g.id)
	g.mu.Unlock()

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) SetTurnOrder(ctx context.Context, turnOrder []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		s.TurnOrder = make([]string, len(turnOrder))
		copy(s.TurnOrder, turnOrder)
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) SetHostPlayerID(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		s.HostPlayerID = playerID
		s.UpdatedAt = time.Now()
	})

	return nil
}

func (g *Game) NextPlayer() *string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	turnOrder := g.TurnOrder()
	if g.currentTurn == nil || len(turnOrder) == 0 {
		return nil
	}

	currentPlayerID := g.currentTurn.PlayerID()

	currentIndex := -1
	for i, playerID := range turnOrder {
		if playerID == currentPlayerID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return &turnOrder[0]
	}

	nextIndex := (currentIndex + 1) % len(turnOrder)
	return &turnOrder[nextIndex]
}

func (g *Game) GetPendingTileSelection(playerID string) *shared.PendingTileSelection {
	var result *shared.PendingTileSelection
	g.read(func(s *datastore.GameState) {
		selection, exists := s.PendingTileSelections[playerID]
		if !exists || selection == nil {
			return
		}
		selectionCopy := *selection
		result = &selectionCopy
	})
	return result
}

func (g *Game) SetPendingTileSelection(ctx context.Context, playerID string, selection *shared.PendingTileSelection) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if selection == nil {
			delete(s.PendingTileSelections, playerID)
		} else {
			selectionCopy := *selection
			s.PendingTileSelections[playerID] = &selectionCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) GetPendingTileSelectionQueue(playerID string) *shared.PendingTileSelectionQueue {
	var result *shared.PendingTileSelectionQueue
	g.read(func(s *datastore.GameState) {
		queue, exists := s.PendingTileSelectionQueues[playerID]
		if !exists || queue == nil {
			return
		}
		queueCopy := *queue
		result = &queueCopy
	})
	return result
}

func (g *Game) SetPendingTileSelectionQueue(ctx context.Context, playerID string, queue *shared.PendingTileSelectionQueue) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if queue == nil {
			delete(s.PendingTileSelectionQueues, playerID)
		} else {
			queueCopy := *queue
			s.PendingTileSelectionQueues[playerID] = &queueCopy
		}
		s.UpdatedAt = time.Now()
	})

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

func (g *Game) SetTileQueueOnComplete(_ context.Context, playerID string, callback *shared.TileCompletionCallback) {
	g.update(func(s *datastore.GameState) {
		if queue, exists := s.PendingTileSelectionQueues[playerID]; exists && queue != nil {
			queue.OnComplete = callback
		}
		if sel, exists := s.PendingTileSelections[playerID]; exists && sel != nil {
			sel.OnComplete = callback
		}
	})
}

func (g *Game) AppendToPendingTileSelectionQueue(ctx context.Context, playerID string, tileTypes []string, source string, sourceCardID string, tileRestrictions *shared.TileRestrictions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(tileTypes) == 0 {
		return nil
	}

	wasEmpty := false
	g.update(func(s *datastore.GameState) {
		existingQueue, exists := s.PendingTileSelectionQueues[playerID]
		var items []string
		var queueSource string
		var queueSourceCardID string
		var queueTileRestrictions *shared.TileRestrictions

		if exists && existingQueue != nil {
			items = existingQueue.Items
			queueSource = existingQueue.Source
			queueSourceCardID = existingQueue.SourceCardID
			queueTileRestrictions = existingQueue.TileRestrictions
		} else {
			items = []string{}
			queueSource = source
			queueSourceCardID = sourceCardID
			queueTileRestrictions = tileRestrictions
			wasEmpty = true
		}

		if exists && existingQueue != nil && len(existingQueue.Items) == 0 {
			wasEmpty = true
		}

		items = append(items, tileTypes...)

		s.PendingTileSelectionQueues[playerID] = &shared.PendingTileSelectionQueue{
			Items:            items,
			TileRestrictions: queueTileRestrictions,
			Source:           queueSource,
			SourceCardID:     queueSourceCardID,
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	if wasEmpty {
		if err := g.ProcessNextTile(ctx, playerID); err != nil {
			return fmt.Errorf("failed to auto-process first queued tile: %w", err)
		}
	}

	return nil
}

func (g *Game) GetForcedFirstAction(playerID string) *shared.ForcedFirstAction {
	var result *shared.ForcedFirstAction
	g.read(func(s *datastore.GameState) {
		action, exists := s.ForcedFirstActions[playerID]
		if !exists || action == nil {
			return
		}
		actionCopy := *action
		result = &actionCopy
	})
	return result
}

func (g *Game) SetForcedFirstAction(ctx context.Context, playerID string, action *shared.ForcedFirstAction) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if action == nil {
			delete(s.ForcedFirstActions, playerID)
		} else {
			actionCopy := *action
			s.ForcedFirstActions[playerID] = &actionCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) GetProductionPhase(playerID string) *shared.ProductionPhase {
	var result *shared.ProductionPhase
	g.read(func(s *datastore.GameState) {
		phase, exists := s.ProductionPhases[playerID]
		if !exists || phase == nil {
			return
		}
		phaseCopy := *phase
		result = &phaseCopy
	})
	return result
}

func (g *Game) SetProductionPhase(ctx context.Context, playerID string, phase *shared.ProductionPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if phase == nil {
			delete(s.ProductionPhases, playerID)
		} else {
			phaseCopy := *phase
			s.ProductionPhases[playerID] = &phaseCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) GetSelectCorporationPhase(playerID string) *shared.SelectCorporationPhase {
	var result *shared.SelectCorporationPhase
	g.read(func(s *datastore.GameState) {
		phase, exists := s.SelectCorporationPhases[playerID]
		if !exists || phase == nil {
			return
		}
		phaseCopy := *phase
		result = &phaseCopy
	})
	return result
}

func (g *Game) SetSelectCorporationPhase(ctx context.Context, playerID string, phase *shared.SelectCorporationPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if phase == nil {
			delete(s.SelectCorporationPhases, playerID)
		} else {
			phaseCopy := *phase
			s.SelectCorporationPhases[playerID] = &phaseCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) GetSelectStartingCardsPhase(playerID string) *shared.SelectStartingCardsPhase {
	var result *shared.SelectStartingCardsPhase
	g.read(func(s *datastore.GameState) {
		phase, exists := s.SelectStartingCardsPhases[playerID]
		if !exists || phase == nil {
			return
		}
		phaseCopy := *phase
		result = &phaseCopy
	})
	return result
}

func (g *Game) SetSelectStartingCardsPhase(ctx context.Context, playerID string, phase *shared.SelectStartingCardsPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if phase == nil {
			delete(s.SelectStartingCardsPhases, playerID)
		} else {
			phaseCopy := *phase
			s.SelectStartingCardsPhases[playerID] = &phaseCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) GetSelectPreludeCardsPhase(playerID string) *shared.SelectPreludeCardsPhase {
	var result *shared.SelectPreludeCardsPhase
	g.read(func(s *datastore.GameState) {
		phase, exists := s.SelectPreludeCardsPhases[playerID]
		if !exists || phase == nil {
			return
		}
		phaseCopy := *phase
		result = &phaseCopy
	})
	return result
}

func (g *Game) SetSelectPreludeCardsPhase(ctx context.Context, playerID string, phase *shared.SelectPreludeCardsPhase) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if phase == nil {
			delete(s.SelectPreludeCardsPhases, playerID)
		} else {
			phaseCopy := *phase
			s.SelectPreludeCardsPhases[playerID] = &phaseCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) ProcessNextTile(ctx context.Context, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var nextTileType string
	var source string
	var sourceCardID string
	var onComplete *shared.TileCompletionCallback
	var tileRestrictions *shared.TileRestrictions
	var found bool

	g.update(func(s *datastore.GameState) {
		queue, exists := s.PendingTileSelectionQueues[playerID]
		if !exists || queue == nil || len(queue.Items) == 0 {
			return
		}
		found = true

		nextTileType = queue.Items[0]
		remainingItems := queue.Items[1:]
		source = queue.Source
		sourceCardID = queue.SourceCardID
		onComplete = queue.OnComplete
		tileRestrictions = queue.TileRestrictions

		if len(remainingItems) > 0 {
			s.PendingTileSelectionQueues[playerID] = &shared.PendingTileSelectionQueue{
				Items:            remainingItems,
				TileRestrictions: tileRestrictions,
				Source:           source,
				SourceCardID:     sourceCardID,
				OnComplete:       onComplete,
			}
		} else {
			delete(s.PendingTileSelectionQueues, playerID)
		}
	})

	if !found {
		return nil
	}

	availableHexes := g.calculateAvailableHexesForTile(nextTileType, playerID, tileRestrictions)

	if len(availableHexes) == 0 {
		return g.ProcessNextTile(ctx, playerID)
	}

	err := g.SetPendingTileSelection(ctx, playerID, &shared.PendingTileSelection{
		TileType:       nextTileType,
		AvailableHexes: availableHexes,
		Source:         source,
		SourceCardID:   sourceCardID,
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

	// Helper to count adjacent occupied tiles of a specific type
	countAdjacentOfType := func(tile board.Tile, tileOccupantType string) int {
		count := 0
		var targetType shared.ResourceType
		switch tileOccupantType {
		case "city":
			targetType = shared.ResourceCityTile
		case "greenery":
			targetType = shared.ResourceGreeneryTile
		case "ocean":
			targetType = shared.ResourceOceanTile
		default:
			targetType = shared.ResourceType(tileOccupantType + "-tile")
		}
		for _, neighborPos := range tile.Coordinates.GetNeighbors() {
			for _, neighborTile := range tiles {
				if neighborTile.Coordinates.Equals(neighborPos) && neighborTile.OccupiedBy != nil && neighborTile.OccupiedBy.Type == targetType {
					count++
					break
				}
			}
		}
		return count
	}

	// Helper to check if tile has an adjacent tile of a specific type owned by the player
	hasAdjacentOwnedOfType := func(tile board.Tile, tileOccupantType string) bool {
		var targetType shared.ResourceType
		switch tileOccupantType {
		case "city":
			targetType = shared.ResourceCityTile
		case "greenery":
			targetType = shared.ResourceGreeneryTile
		case "ocean":
			targetType = shared.ResourceOceanTile
		default:
			targetType = shared.ResourceType(tileOccupantType + "-tile")
		}
		for _, neighborPos := range tile.Coordinates.GetNeighbors() {
			for _, neighborTile := range tiles {
				if neighborTile.Coordinates.Equals(neighborPos) &&
					neighborTile.OccupiedBy != nil &&
					neighborTile.OccupiedBy.Type == targetType &&
					neighborTile.OwnerID != nil &&
					*neighborTile.OwnerID == playerID {
					return true
				}
			}
		}
		return false
	}

	// Helper to check if tile has any adjacent tile owned by the player (regardless of type)
	hasAnyAdjacentOwned := func(tile board.Tile) bool {
		for _, neighborPos := range tile.Coordinates.GetNeighbors() {
			for _, neighborTile := range tiles {
				if neighborTile.Coordinates.Equals(neighborPos) &&
					neighborTile.OccupiedBy != nil &&
					neighborTile.OwnerID != nil &&
					*neighborTile.OwnerID == playerID {
					return true
				}
			}
		}
		return false
	}

	// Helper to check if a tile has one of the specified bonus types
	hasBonusOfType := func(tile board.Tile, bonusTypes []string) bool {
		for _, bonus := range tile.Bonuses {
			for _, bonusType := range bonusTypes {
				if string(bonus.Type) == bonusType {
					return true
				}
			}
		}
		return false
	}

	// Helper to check if a tile passes adjacentToType/minAdjacentOfType/adjacentToOwned restrictions
	passesAdjacentRestrictions := func(tile board.Tile) bool {
		if tileRestrictions == nil {
			return true
		}
		// Check adjacentToType + minAdjacentOfType
		if tileRestrictions.AdjacentToType != "" {
			minRequired := 1
			if tileRestrictions.MinAdjacentOfType != nil {
				minRequired = *tileRestrictions.MinAdjacentOfType
			}
			if countAdjacentOfType(tile, tileRestrictions.AdjacentToType) < minRequired {
				return false
			}
			// If both adjacentToType and adjacentToOwned are set, check owned of that type
			if tileRestrictions.AdjacentToOwned && !hasAdjacentOwnedOfType(tile, tileRestrictions.AdjacentToType) {
				return false
			}
		} else if tileRestrictions.AdjacentToOwned {
			// AdjacentToOwned without AdjacentToType: adjacent to any owned tile
			if !hasAnyAdjacentOwned(tile) {
				return false
			}
		}
		return true
	}

	for _, tile := range tiles {
		// Clear targets occupied or reserved tiles (inverse of normal placement)
		if tileType == "clear" {
			if tile.OccupiedBy != nil || tile.ReservedBy != nil {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
			continue
		}

		// Tile replacement targets any occupied non-ocean tile
		if strings.HasPrefix(tileType, "tile-replacement:") {
			if tile.OccupiedBy != nil && tile.OccupiedBy.Type != shared.ResourceOceanTile {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}
			continue
		}

		// Tile destruction targets any occupied tile on the board
		if tileType == "tile-destruction" {
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
					logger.Get().Debug("Tile available for city (board tag match)",
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
				logger.Get().Debug("Skipping reserved tile for normal city placement",
					zap.String("tile", tile.Coordinates.String()),
					zap.Strings("tile_tags", tile.Tags))
				continue
			}

			// Handle "no adjacent tiles" restriction (Research Outpost)
			if adjacency == "none" {
				if !hasAnyAdjacentOccupied(tile) {
					availableHexes = append(availableHexes, tile.Coordinates.String())
					logger.Get().Debug("Tile available for city (no adjacent tiles)",
						zap.String("tile", tile.Coordinates.String()))
				}
				continue // Skip normal city adjacency rules
			}

			// Handle adjacentToType restriction (e.g., Urbanized Area: adjacent to 2+ cities)
			// This overrides the normal "no adjacent cities" rule
			if tileRestrictions != nil && tileRestrictions.AdjacentToType != "" {
				if passesAdjacentRestrictions(tile) {
					availableHexes = append(availableHexes, tile.Coordinates.String())
				}
				continue
			}

			// Check city adjacency rule (no adjacent cities)
			hasAdjacentCity := false
			neighbors := tile.Coordinates.GetNeighbors()

			logger.Get().Debug("Checking city placement",
				zap.String("tile", tile.Coordinates.String()),
				zap.Int("neighbor_count", len(neighbors)))

			for _, neighborPos := range neighbors {
				for _, neighborTile := range tiles {
					if neighborTile.Coordinates.Equals(neighborPos) {
						occupantType := ""
						if neighborTile.OccupiedBy != nil {
							occupantType = string(neighborTile.OccupiedBy.Type)
						}

						logger.Get().Debug("Checking neighbor",
							zap.String("neighbor_pos", neighborPos.String()),
							zap.String("neighbor_tile", neighborTile.Coordinates.String()),
							zap.Bool("occupied", neighborTile.OccupiedBy != nil),
							zap.String("occupant_type", occupantType))

						if neighborTile.OccupiedBy != nil && neighborTile.OccupiedBy.Type == shared.ResourceCityTile {
							hasAdjacentCity = true
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
				logger.Get().Debug("Tile available for city",
					zap.String("tile", tile.Coordinates.String()))
			} else {
				logger.Get().Debug("Tile unavailable for city (adjacent city)",
					zap.String("tile", tile.Coordinates.String()))
			}

		case "greenery", "world-tree":
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
				// Apply adjacency restrictions if set (e.g., Ecological Zone: adjacent to greenery)
				if !passesAdjacentRestrictions(tile) {
					continue
				}
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

		case "mohole":
			if tile.Type == shared.ResourceOceanSpace {
				availableHexes = append(availableHexes, tile.Coordinates.String())
			}

		default:
			// Handle ocean-space placement (e.g., Mohole Area)
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

			// If boardTags specified, only allow tiles with matching tags
			if len(boardTags) > 0 {
				if tileHasRequiredTag(tile, boardTags) {
					availableHexes = append(availableHexes, tile.Coordinates.String())
				}
				continue
			}

			// Exclude reserved areas from normal placement
			if tileHasAnyTag(tile) {
				continue
			}

			// Must be on land
			if tile.Type != shared.ResourceLandTile {
				continue
			}

			// Check bonus type restriction (e.g., Mining Area/Mining Rights)
			if tileRestrictions != nil && len(tileRestrictions.OnBonusType) > 0 {
				if !hasBonusOfType(tile, tileRestrictions.OnBonusType) {
					continue
				}
			}

			// Handle "no adjacent tiles" restriction (e.g., Natural Preserve)
			if adjacency == "none" {
				if !hasAnyAdjacentOccupied(tile) {
					availableHexes = append(availableHexes, tile.Coordinates.String())
				}
				continue
			}

			// Apply adjacency restrictions if set
			if !passesAdjacentRestrictions(tile) {
				continue
			}

			availableHexes = append(availableHexes, tile.Coordinates.String())
		}
	}

	if len(availableHexes) == 0 && tileRestrictions != nil {
		// Board tags fallback (e.g., Noctis City already occupied)
		canFallback := len(boardTags) > 0
		// AdjacentToOwned-only fallback: greenery must be placed adjacent to own tiles if possible,
		// but if no owned tiles exist, placement is allowed anywhere (TM rules)
		if tileRestrictions.AdjacentToOwned && tileRestrictions.AdjacentToType == "" && len(tileRestrictions.OnBonusType) == 0 {
			canFallback = true
		}
		if canFallback {
			logger.Get().Debug("No tiles match restrictions, falling back to normal placement",
				zap.String("tile_type", tileType))
			return g.calculateAvailableHexesForTile(tileType, playerID, nil)
		}
	}

	return availableHexes
}

// CountAvailableHexesForTile returns the number of valid hex positions for placing a tile
// This is used by state calculators to determine if tile-placing actions are available
func (g *Game) CountAvailableHexesForTile(tileType string, playerID string, tileRestrictions *shared.TileRestrictions) int {
	return len(g.calculateAvailableHexesForTile(tileType, playerID, tileRestrictions))
}

// CalculateAvailableHexesForTile returns the list of valid hex coordinate strings for placing a tile
func (g *Game) CalculateAvailableHexesForTile(tileType string, playerID string, tileRestrictions *shared.TileRestrictions) []string {
	return g.calculateAvailableHexesForTile(tileType, playerID, tileRestrictions)
}

func (g *Game) AddTriggeredEffect(effect shared.TriggeredEffect) {
	g.update(func(s *datastore.GameState) {
		s.TriggeredEffects = append(s.TriggeredEffects, effect)
	})
}

func (g *Game) AppendToLastTriggeredEffect(playerID string, outputs []shared.CalculatedOutput) {
	g.update(func(s *datastore.GameState) {
		for i := len(s.TriggeredEffects) - 1; i >= 0; i-- {
			if s.TriggeredEffects[i].PlayerID == playerID {
				s.TriggeredEffects[i].CalculatedOutputs = append(s.TriggeredEffects[i].CalculatedOutputs, outputs...)
				return
			}
		}
	})
}

func (g *Game) GetTriggeredEffects() []shared.TriggeredEffect {
	var result []shared.TriggeredEffect
	g.read(func(s *datastore.GameState) {
		result = make([]shared.TriggeredEffect, len(s.TriggeredEffects))
		copy(result, s.TriggeredEffects)
	})
	return result
}

func (g *Game) ClearTriggeredEffects() {
	g.update(func(s *datastore.GameState) {
		s.TriggeredEffects = nil
	})
}

func (g *Game) SetVPCardLookup(lookup VPCardLookup) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.vpCardLookup = lookup
	g.subscribeToVPEvents()
}

// RegisterCorporationVPGranter registers VP conditions from a corporation card.
func (g *Game) RegisterCorporationVPGranter(playerID string, corporationID string) {
	if g.vpCardLookup == nil {
		return
	}
	cardInfo, err := g.vpCardLookup.LookupVPCard(corporationID)
	if err != nil || len(cardInfo.VPConditions) == 0 {
		return
	}
	p, err := g.GetPlayer(playerID)
	if err != nil {
		return
	}
	granter := shared.VPGranter{
		CardID:       cardInfo.CardID,
		CardName:     cardInfo.CardName,
		Description:  cardInfo.Description,
		VPConditions: cardInfo.VPConditions,
	}
	p.VPGranters().Prepend(granter)
	g.recalculatePlayerVP(p)
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

		granter := shared.VPGranter{
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
			// Clear temporary "generation-end" effects
			p.Effects().RemoveTemporaryEffects(shared.TemporaryGenerationEnd)
			// Also clear any "next-card" effects that weren't consumed
			p.Effects().RemoveTemporaryEffects(shared.TemporaryNextCard)
		}
	})
}

func (g *Game) subscribeToOceanSpaceEvents() {
	events.Subscribe(g.eventBus, func(e events.TilePlacedEvent) {
		if e.TileType == string(shared.ResourceOceanTile) || e.TileType == "ocean" {
			return
		}

		coords := shared.HexPosition{Q: e.Q, R: e.R, S: e.S}
		tile, err := g.board.GetTile(coords)
		if err != nil {
			return
		}
		if tile.Type != shared.ResourceOceanSpace {
			return
		}

		freeOceanSpaces := g.board.FreeOceanSpaces()
		gp := g.globalParameters
		oceansRemaining := gp.GetMaxOceans() - gp.Oceans()
		if freeOceanSpaces < oceansRemaining {
			gp.ReduceMaxOceans(gp.Oceans() + freeOceanSpaces)
		}
	})
}

func (g *Game) subscribeToGlobalParameterBonuses() {
	log := logger.Get()

	events.Subscribe(g.eventBus, func(e events.TemperatureChangedEvent) {
		if e.ChangedBy == "" {
			return
		}
		p, err := g.GetPlayer(e.ChangedBy)
		if err != nil {
			return
		}

		if e.OldValue < -24 && e.NewValue >= -24 {
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceHeatProduction: 1,
			})
			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   "Temperature Bonus",
				PlayerID:   e.ChangedBy,
				SourceType: shared.SourceTypeGlobalBonus,
				CalculatedOutputs: []shared.CalculatedOutput{
					{ResourceType: string(shared.ResourceHeatProduction), Amount: 1},
				},
			})
			log.Debug("Temperature bonus: +1 heat production at -24C", zap.String("player_id", e.ChangedBy))
		}

		if e.OldValue < -20 && e.NewValue >= -20 {
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceHeatProduction: 1,
			})
			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   "Temperature Bonus",
				PlayerID:   e.ChangedBy,
				SourceType: shared.SourceTypeGlobalBonus,
				CalculatedOutputs: []shared.CalculatedOutput{
					{ResourceType: string(shared.ResourceHeatProduction), Amount: 1},
				},
			})
			log.Debug("Temperature bonus: +1 heat production at -20C", zap.String("player_id", e.ChangedBy))
		}

		if e.OldValue < 0 && e.NewValue >= 0 {
			ctx := context.Background()
			if err := g.AppendToPendingTileSelectionQueue(ctx, e.ChangedBy, []string{"ocean"}, "Temperature Bonus", "", nil); err != nil {
				log.Warn("Failed to queue ocean tile for temperature bonus", zap.Error(err))
			}
			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   "Temperature Bonus",
				PlayerID:   e.ChangedBy,
				SourceType: shared.SourceTypeGlobalBonus,
				CalculatedOutputs: []shared.CalculatedOutput{
					{ResourceType: string(shared.ResourceOceanTile), Amount: 1},
				},
			})
			log.Debug("Temperature bonus: place ocean at 0C", zap.String("player_id", e.ChangedBy))
		}
	})

	events.Subscribe(g.eventBus, func(e events.OxygenChangedEvent) {
		if e.ChangedBy == "" {
			return
		}

		if e.OldValue < 8 && e.NewValue >= 8 {
			ctx := context.Background()
			actualSteps, err := g.globalParameters.IncreaseTemperature(ctx, 1, e.ChangedBy)
			if err != nil {
				logger.Get().Warn("Failed to increase temperature from oxygen bonus", zap.Error(err))
			}
			if actualSteps > 0 {
				p, pErr := g.GetPlayer(e.ChangedBy)
				if pErr == nil {
					p.Resources().UpdateTerraformRating(actualSteps)
				}
			}
			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   "Oxygen Bonus",
				PlayerID:   e.ChangedBy,
				SourceType: shared.SourceTypeGlobalBonus,
				CalculatedOutputs: []shared.CalculatedOutput{
					{ResourceType: string(shared.ResourceTemperature), Amount: 1},
				},
			})
			log.Debug("Oxygen bonus: +1 temperature step at 8%", zap.String("player_id", e.ChangedBy))
		}
	})

	events.Subscribe(g.eventBus, func(e events.VenusChangedEvent) {
		if e.ChangedBy == "" {
			return
		}
		p, err := g.GetPlayer(e.ChangedBy)
		if err != nil {
			return
		}

		if e.OldValue < 8 && e.NewValue >= 8 {
			ctx := context.Background()
			cardIDs, drawErr := g.deck.DrawProjectCards(ctx, 1)
			if drawErr == nil && len(cardIDs) > 0 {
				p.Hand().AddCard(cardIDs[0])
			}
			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   "Venus Bonus",
				PlayerID:   e.ChangedBy,
				SourceType: shared.SourceTypeGlobalBonus,
				CalculatedOutputs: []shared.CalculatedOutput{
					{ResourceType: "card-draw", Amount: 1},
				},
			})
			log.Debug("Venus bonus: draw 1 card at 8%", zap.String("player_id", e.ChangedBy))
		}

		if e.OldValue < 16 && e.NewValue >= 16 {
			p.Resources().UpdateTerraformRating(1)
			g.AddTriggeredEffect(shared.TriggeredEffect{
				CardName:   "Venus Bonus",
				PlayerID:   e.ChangedBy,
				SourceType: shared.SourceTypeGlobalBonus,
				CalculatedOutputs: []shared.CalculatedOutput{
					{ResourceType: string(shared.ResourceTR), Amount: 1},
				},
			})
			log.Debug("Venus bonus: +1 TR at 16%", zap.String("player_id", e.ChangedBy))
		}
	})
}

func (g *Game) GetDeferredStartingChoices(playerID string) *shared.DeferredStartingChoices {
	var result *shared.DeferredStartingChoices
	g.read(func(s *datastore.GameState) {
		choices, exists := s.DeferredStartingChoices[playerID]
		if !exists || choices == nil {
			return
		}
		choicesCopy := *choices
		result = &choicesCopy
	})
	return result
}

func (g *Game) SetDeferredStartingChoices(ctx context.Context, playerID string, choices *shared.DeferredStartingChoices) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		if choices == nil {
			delete(s.DeferredStartingChoices, playerID)
		} else {
			choicesCopy := *choices
			s.DeferredStartingChoices[playerID] = &choicesCopy
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) MarkCorpApplied(playerID string) {
	g.update(func(s *datastore.GameState) {
		if choices, ok := s.DeferredStartingChoices[playerID]; ok && choices != nil {
			choices.CorpApplied = true
		}
	})
}

func (g *Game) MarkPreludesApplied(playerID string) {
	g.update(func(s *datastore.GameState) {
		if choices, ok := s.DeferredStartingChoices[playerID]; ok && choices != nil {
			choices.PreludesApplied = true
		}
	})
}

func (g *Game) InitPhasePlayerIndex() int {
	var v int
	g.read(func(s *datastore.GameState) { v = s.InitPhasePlayerIndex })
	return v
}

func (g *Game) SetInitPhasePlayerIndex(ctx context.Context, index int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		s.InitPhasePlayerIndex = index
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) InitPhaseWaitingForConfirm() bool {
	var v bool
	g.read(func(s *datastore.GameState) { v = s.InitPhaseWaitingForConfirm })
	return v
}

func (g *Game) InitPhaseConfirmVersion() int {
	var v int
	g.read(func(s *datastore.GameState) { v = s.InitPhaseConfirmVersion })
	return v
}

func (g *Game) SetInitPhaseWaitingForConfirm(ctx context.Context, waiting bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	g.update(func(s *datastore.GameState) {
		s.InitPhaseWaitingForConfirm = waiting
		if waiting {
			s.InitPhaseConfirmVersion++
		}
		s.UpdatedAt = time.Now()
	})

	if g.eventBus != nil {
		events.Publish(g.eventBus, events.GameStateChangedEvent{
			GameID:    g.id,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (g *Game) HasColonies() bool {
	return g.Settings().HasColonies()
}

func (g *Game) ColonyTileStates() []*colony.TileState {
	var result []*colony.TileState
	g.read(func(s *datastore.GameState) { result = s.ColonyTileStates })
	return result
}

// GetAvailableColonyIDs returns the definition IDs of all colony tiles in the game.
func (g *Game) GetAvailableColonyIDs() []string {
	tiles := g.ColonyTileStates()
	ids := make([]string, 0, len(tiles))
	for _, ts := range tiles {
		ids = append(ids, ts.DefinitionID)
	}
	return ids
}

// GetPlaceableColonyIDs returns colony IDs where the player can actually place a colony
// (not full and player doesn't already have a colony there, unless allowDuplicate is true).
func (g *Game) GetPlaceableColonyIDs(playerID string, allowDuplicate bool) []string {
	const maxColoniesPerTile = 3
	tiles := g.ColonyTileStates()
	ids := make([]string, 0, len(tiles))
	for _, ts := range tiles {
		if len(ts.PlayerColonies) >= maxColoniesPerTile {
			continue
		}
		if !allowDuplicate && slices.Contains(ts.PlayerColonies, playerID) {
			continue
		}
		ids = append(ids, ts.DefinitionID)
	}
	return ids
}

// GetTradeableColonyIDs returns colony IDs that haven't been traded this generation.
func (g *Game) GetTradeableColonyIDs() []string {
	tiles := g.ColonyTileStates()
	ids := make([]string, 0, len(tiles))
	for _, ts := range tiles {
		if !ts.TradedThisGen {
			ids = append(ids, ts.DefinitionID)
		}
	}
	return ids
}

func (g *Game) SetColonyTileStates(states []*colony.TileState) {
	g.update(func(s *datastore.GameState) {
		s.ColonyTileStates = states
		s.UpdatedAt = time.Now()
	})
}

func (g *Game) GetColonyTileState(colonyID string) *colony.TileState {
	var result *colony.TileState
	g.read(func(s *datastore.GameState) {
		for _, state := range s.ColonyTileStates {
			if state.DefinitionID == colonyID {
				result = state
				return
			}
		}
	})
	return result
}

// CountAllColonies returns the total number of colonies placed across all colony tiles.
func (g *Game) CountAllColonies() int {
	var total int
	g.read(func(s *datastore.GameState) {
		for _, state := range s.ColonyTileStates {
			total += len(state.PlayerColonies)
		}
	})
	return total
}

func (g *Game) GetTradeFleetAvailable(playerID string) bool {
	var v bool
	g.read(func(s *datastore.GameState) {
		if s.TradeFleets == nil {
			return
		}
		v = s.TradeFleets[playerID]
	})
	return v
}

func (g *Game) SetTradeFleetAvailable(playerID string, available bool) {
	g.update(func(s *datastore.GameState) {
		if s.TradeFleets == nil {
			s.TradeFleets = make(map[string]bool)
		}
		s.TradeFleets[playerID] = available
		s.UpdatedAt = time.Now()
	})
}

func (g *Game) InitializeTradeFleets(playerIDs []string) {
	g.update(func(s *datastore.GameState) {
		s.TradeFleets = make(map[string]bool, len(playerIDs))
		for _, id := range playerIDs {
			s.TradeFleets[id] = true
		}
		s.UpdatedAt = time.Now()
	})
}

// HasProjectFunding returns true if the project funding expansion is enabled
func (g *Game) HasProjectFunding() bool {
	return g.Settings().HasProjectFunding()
}

// ProjectFundingStates returns the project funding states
func (g *Game) ProjectFundingStates() []*projectfunding.ProjectState {
	var result []*projectfunding.ProjectState
	g.read(func(s *datastore.GameState) { result = s.ProjectFundingStates })
	return result
}

// SetProjectFundingStates sets the project funding states
func (g *Game) SetProjectFundingStates(states []*projectfunding.ProjectState) {
	g.update(func(s *datastore.GameState) {
		s.ProjectFundingStates = states
		s.UpdatedAt = time.Now()
	})
}

// IsNextGenTurnOrderFrozen returns true if turn order rotation is skipped next generation.
func (g *Game) IsNextGenTurnOrderFrozen() bool {
	var v bool
	g.read(func(s *datastore.GameState) { v = s.NextGenTurnOrderFrozen })
	return v
}

// SetNextGenTurnOrderFrozen sets whether turn order rotation is skipped next generation.
func (g *Game) SetNextGenTurnOrderFrozen(frozen bool) {
	g.update(func(s *datastore.GameState) {
		s.NextGenTurnOrderFrozen = frozen
		s.UpdatedAt = time.Now()
	})
}

// GetProjectFundingState returns the state for a specific project
func (g *Game) GetProjectFundingState(projectID string) *projectfunding.ProjectState {
	var result *projectfunding.ProjectState
	g.read(func(s *datastore.GameState) {
		for _, state := range s.ProjectFundingStates {
			if state.DefinitionID == projectID {
				result = state
				return
			}
		}
	})
	return result
}
