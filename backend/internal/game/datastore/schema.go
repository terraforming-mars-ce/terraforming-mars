package datastore

import (
	"fmt"

	"github.com/hashicorp/go-memdb"

	"terraforming-mars-backend/internal/game/shared"
)

func createSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"games": {
				Name: "games",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"status": {
						Name:         "status",
						Unique:       false,
						AllowMissing: true,
						Indexer:      &gameStatusIndexer{},
					},
				},
			},
			"snapshots": {
				Name: "snapshots",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "GameID"},
					},
				},
			},
			"difflogs": {
				Name: "difflogs",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "GameID"},
					},
				},
			},
		},
	}
}

type gameStatusIndexer struct{}

func (g *gameStatusIndexer) FromObject(raw any) (bool, []byte, error) {
	gs, ok := raw.(*GameState)
	if !ok {
		return false, nil, nil
	}
	val := string(gs.Status) + "\x00"
	return true, []byte(val), nil
}

func (g *gameStatusIndexer) FromArgs(args ...any) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide exactly one argument")
	}
	switch v := args[0].(type) {
	case string:
		return []byte(v + "\x00"), nil
	case shared.GameStatus:
		return []byte(string(v) + "\x00"), nil
	default:
		return nil, fmt.Errorf("argument must be a string or GameStatus: %#v", args[0])
	}
}
