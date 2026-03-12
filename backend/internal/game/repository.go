package game

import (
	"context"

	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
)

// GameRepository manages the collection of active games
type GameRepository interface {
	Get(ctx context.Context, gameID string) (*Game, error)
	Create(ctx context.Context, game *Game) error
	Delete(ctx context.Context, gameID string) error
	List(ctx context.Context, status *shared.GameStatus) ([]*Game, error)
	Exists(ctx context.Context, gameID string) bool
	DataStore() *datastore.DataStore
}
