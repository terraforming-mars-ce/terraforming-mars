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

const (
	MaxFundedAwards = 3
	AwardFirstVP    = 5
	AwardSecondVP   = 2
)

var AwardFundingCosts = []int{8, 14, 20}

type AwardInfo struct {
	Type        shared.AwardType
	Name        string
	Description string
}

var AllAwards = []AwardInfo{
	{Type: shared.AwardLandlord, Name: "Landlord", Description: "Most tiles on Mars"},
	{Type: shared.AwardBanker, Name: "Banker", Description: "Highest MC production"},
	{Type: shared.AwardScientist, Name: "Scientist", Description: "Most science tags in play"},
	{Type: shared.AwardThermalist, Name: "Thermalist", Description: "Most heat resources"},
	{Type: shared.AwardMiner, Name: "Miner", Description: "Most steel and titanium resources"},
}

type Awards struct {
	ds       *datastore.DataStore
	gameID   string
	eventBus *events.EventBusImpl
}

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

func (a *Awards) FundedAwards() []shared.FundedAward {
	var result []shared.FundedAward
	a.read(func(s *datastore.GameState) {
		result = make([]shared.FundedAward, len(s.FundedAwards))
		copy(result, s.FundedAwards)
	})
	return result
}

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

func (a *Awards) CanFundMore() bool {
	var can bool
	a.read(func(s *datastore.GameState) {
		can = len(s.FundedAwards) < MaxFundedAwards
	})
	return can
}

func (a *Awards) FundedCount() int {
	var count int
	a.read(func(s *datastore.GameState) { count = len(s.FundedAwards) })
	return count
}

func (a *Awards) GetCurrentFundingCost() int {
	var cost int
	a.read(func(s *datastore.GameState) {
		count := len(s.FundedAwards)
		if count >= MaxFundedAwards {
			return
		}
		cost = AwardFundingCosts[count]
	})
	return cost
}

func (a *Awards) FundAward(ctx context.Context, awardType shared.AwardType, playerID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var fundErr error
	var fundingCost int
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

		fundingCost = AwardFundingCosts[len(s.FundedAwards)]
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

func GetAwardInfo(awardType shared.AwardType) (AwardInfo, bool) {
	for _, info := range AllAwards {
		if info.Type == awardType {
			return info, true
		}
	}
	return AwardInfo{}, false
}
