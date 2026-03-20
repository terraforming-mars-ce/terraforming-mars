package game

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// MaxFundedAwards is the maximum number of awards that can be funded in a game
const MaxFundedAwards = 3

// Awards manages award funding state for a game
type Awards struct {
	ds       *datastore.DataStore
	gameID   string
	eventBus *events.EventBusImpl
}

// NewAwards creates a new Awards state tracker
func NewAwards(ds *datastore.DataStore, gameID string, eventBus *events.EventBusImpl) *Awards {
	return &Awards{
		ds:       ds,
		gameID:   gameID,
		eventBus: eventBus,
	}
}

func (a *Awards) update(fn func(s *datastore.GameState)) {
	if err := a.ds.UpdateGame(a.gameID, fn); err != nil {
		logger.Get().Warn("Failed to update game state", zap.String("game_id", a.gameID), zap.Error(err))
	}
}

func (a *Awards) read(fn func(s *datastore.GameState)) {
	if err := a.ds.ReadGame(a.gameID, fn); err != nil {
		logger.Get().Warn("Failed to read game state", zap.String("game_id", a.gameID), zap.Error(err))
	}
}

// FundedAwards returns a copy of all funded awards
func (a *Awards) FundedAwards() []shared.FundedAward {
	var result []shared.FundedAward
	a.read(func(s *datastore.GameState) {
		result = make([]shared.FundedAward, len(s.FundedAwards))
		copy(result, s.FundedAwards)
	})
	return result
}

// IsFunded returns true if the given award has been funded
func (a *Awards) IsFunded(awardType shared.AwardType) bool {
	var funded bool
	a.read(func(s *datastore.GameState) {
		for _, f := range s.FundedAwards {
			if f.Type == awardType {
				funded = true
				return
			}
		}
	})
	return funded
}

// IsFundedBy returns true if the given award was funded by the specified player
func (a *Awards) IsFundedBy(awardType shared.AwardType, playerID string) bool {
	var funded bool
	a.read(func(s *datastore.GameState) {
		for _, f := range s.FundedAwards {
			if f.Type == awardType && f.FundedByPlayer == playerID {
				funded = true
				return
			}
		}
	})
	return funded
}

// CanFundMore returns true if fewer than MaxFundedAwards have been funded
func (a *Awards) CanFundMore() bool {
	var can bool
	a.read(func(s *datastore.GameState) {
		can = len(s.FundedAwards) < MaxFundedAwards
	})
	return can
}

// FundedCount returns the number of currently funded awards
func (a *Awards) FundedCount() int {
	var count int
	a.read(func(s *datastore.GameState) { count = len(s.FundedAwards) })
	return count
}

// FundAward funds an award for a player at the given cost
func (a *Awards) FundAward(ctx context.Context, awardType shared.AwardType, playerID string, fundingCost int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var fundErr error
	a.update(func(s *datastore.GameState) {
		if len(s.FundedAwards) >= MaxFundedAwards {
			fundErr = fmt.Errorf("maximum awards (%d) already funded", MaxFundedAwards)
			return
		}

		for _, f := range s.FundedAwards {
			if f.Type == awardType {
				fundErr = fmt.Errorf("award %s is already funded", awardType)
				return
			}
		}

		s.FundedAwards = append(s.FundedAwards, shared.FundedAward{
			Type:           awardType,
			FundedByPlayer: playerID,
			FundingOrder:   len(s.FundedAwards),
			FundingCost:    fundingCost,
			FundedAt:       time.Now(),
		})
	})
	if fundErr != nil {
		return fundErr
	}

	if a.eventBus != nil {
		events.Publish(a.eventBus, events.AwardFundedEvent{
			GameID:      a.gameID,
			PlayerID:    playerID,
			AwardType:   string(awardType),
			FundingCost: fundingCost,
			Timestamp:   time.Now(),
		})
	}

	return nil
}
