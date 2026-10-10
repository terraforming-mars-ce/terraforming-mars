package game

import (
	"fmt"
	"maps"
	"reflect"

	"openmars/internal/events"
	"openmars/internal/game/board"
	"openmars/internal/game/colonies"
	"openmars/internal/game/datastore"
	"openmars/internal/game/deck"
	"openmars/internal/game/global_parameters"
	"openmars/internal/game/player"
)

// RestoreGame binds runtime views to an existing state without performing setup.
// The caller owns the state and registers the completed game only after hydration.
func RestoreGame(ds *datastore.DataStore, state *datastore.GameState) (*Game, error) {
	if state == nil || state.ID == "" {
		return nil, fmt.Errorf("missing game state")
	}
	// Runtime-only initialization must not mutate immutable historical snapshots.
	restored := *state
	restored.Players = maps.Clone(state.Players)
	for id, p := range restored.Players {
		if p == nil {
			return nil, fmt.Errorf("player %s is null", id)
		}
		copy := *p
		initializeMaps(reflect.ValueOf(&copy))
		restored.Players[id] = &copy
	}
	state = &restored
	initializeMaps(reflect.ValueOf(state))
	txn := ds.BeginTxn()
	defer txn.Abort()
	if err := txn.InsertGame(state); err != nil {
		return nil, err
	}
	txn.Commit()
	bus := events.NewEventBus()
	g := &Game{
		ds: ds, id: state.ID, eventBus: bus,
		board:            board.NewBoard(&state.Tiles, state.ID, bus),
		globalParameters: global_parameters.NewGlobalParameters(ds, state.ID, bus),
		colonies:         colonies.NewColonies(ds, state.ID, bus),
		deck:             deck.NewDeckView(ds, state.ID),
		players:          make(map[string]*player.Player),
		milestones:       NewMilestones(ds, state.ID, bus),
		awards:           NewAwards(ds, state.ID, bus),
	}
	if state.CurrentTurnPlayerID != "" {
		g.currentTurn = NewTurn(ds, state.ID)
	}
	for id := range state.Players {
		g.players[id] = player.NewPlayer(ds, state.ID, id, bus)
	}
	g.subscribeToGenerationalEvents()
	g.subscribeToOceanSpaceEvents()
	g.subscribeToGlobalParameterBonuses()
	return g, nil
}

// ResumeLobby contains only the temporary seat assignments of an imported game.
type ResumeLobby struct {
	SavedNames map[string]string
	Claimed    map[string]bool
}

// ResumeLobby returns the session state while players gather, or nil during play.
func (g *Game) ResumeLobby() *ResumeLobby {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return copyResumeLobby(g.resumeLobby)
}

// SetResumeLobby changes the session gate without changing the saved gameplay phase.
func (g *Game) SetResumeLobby(lobby *ResumeLobby) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.resumeLobby = copyResumeLobby(lobby)
}

func copyResumeLobby(lobby *ResumeLobby) *ResumeLobby {
	if lobby == nil {
		return nil
	}
	return &ResumeLobby{SavedNames: maps.Clone(lobby.SavedNames), Claimed: maps.Clone(lobby.Claimed)}
}

func initializeMaps(v reflect.Value) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		if f.Kind() == reflect.Map {
			if f.IsNil() {
				f.Set(reflect.MakeMap(f.Type()))
			}
		}
	}
}
