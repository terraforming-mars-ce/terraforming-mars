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
	MaxClaimedMilestones = 3
)

type Milestones struct {
	ds       *datastore.DataStore
	gameID   string
	eventBus *events.EventBusImpl
}

func NewMilestones(ds *datastore.DataStore, gameID string, eventBus *events.EventBusImpl) *Milestones {
	return &Milestones{
		ds:       ds,
		gameID:   gameID,
		eventBus: eventBus,
	}
}

func (m *Milestones) update(fn func(s *datastore.GameState)) {
	if err := m.ds.UpdateGame(m.gameID, fn); err != nil {
		logger.Get().Warn("Failed to update game state", zap.String("game_id", m.gameID), zap.Error(err))
	}
}

func (m *Milestones) read(fn func(s *datastore.GameState)) {
	if err := m.ds.ReadGame(m.gameID, fn); err != nil {
		logger.Get().Warn("Failed to read game state", zap.String("game_id", m.gameID), zap.Error(err))
	}
}

func (m *Milestones) ClaimedMilestones() []shared.ClaimedMilestone {
	var result []shared.ClaimedMilestone
	m.read(func(s *datastore.GameState) {
		result = make([]shared.ClaimedMilestone, len(s.ClaimedMilestones))
		copy(result, s.ClaimedMilestones)
	})
	return result
}

func (m *Milestones) IsClaimed(milestoneType shared.MilestoneType) bool {
	var claimed bool
	m.read(func(s *datastore.GameState) {
		for _, c := range s.ClaimedMilestones {
			if c.Type == milestoneType {
				claimed = true
				return
			}
		}
	})
	return claimed
}

func (m *Milestones) IsClaimedBy(milestoneType shared.MilestoneType, playerID string) bool {
	var claimed bool
	m.read(func(s *datastore.GameState) {
		for _, c := range s.ClaimedMilestones {
			if c.Type == milestoneType && c.PlayerID == playerID {
				claimed = true
				return
			}
		}
	})
	return claimed
}

func (m *Milestones) CanClaimMore() bool {
	var can bool
	m.read(func(s *datastore.GameState) {
		can = len(s.ClaimedMilestones) < MaxClaimedMilestones
	})
	return can
}

func (m *Milestones) ClaimedCount() int {
	var count int
	m.read(func(s *datastore.GameState) { count = len(s.ClaimedMilestones) })
	return count
}

func (m *Milestones) GetClaimedByPlayer(playerID string) []shared.ClaimedMilestone {
	var result []shared.ClaimedMilestone
	m.read(func(s *datastore.GameState) {
		for _, c := range s.ClaimedMilestones {
			if c.PlayerID == playerID {
				result = append(result, c)
			}
		}
	})
	return result
}

func (m *Milestones) ClaimMilestone(ctx context.Context, milestoneType shared.MilestoneType, playerID string, generation int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var claimErr error
	m.update(func(s *datastore.GameState) {
		if len(s.ClaimedMilestones) >= MaxClaimedMilestones {
			claimErr = fmt.Errorf("maximum milestones (%d) already claimed", MaxClaimedMilestones)
			return
		}

		for _, c := range s.ClaimedMilestones {
			if c.Type == milestoneType {
				claimErr = fmt.Errorf("milestone %s is already claimed", milestoneType)
				return
			}
		}

		s.ClaimedMilestones = append(s.ClaimedMilestones, shared.ClaimedMilestone{
			Type:       milestoneType,
			PlayerID:   playerID,
			Generation: generation,
			ClaimedAt:  time.Now(),
		})
	})
	if claimErr != nil {
		return claimErr
	}

	if m.eventBus != nil {
		events.Publish(m.eventBus, events.MilestoneClaimedEvent{
			GameID:        m.gameID,
			PlayerID:      playerID,
			MilestoneType: string(milestoneType),
			Timestamp:     time.Now(),
		})
	}

	return nil
}
