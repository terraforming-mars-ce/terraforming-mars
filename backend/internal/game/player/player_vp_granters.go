package player

import (
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/shared"
)

type VPConditionType string

const (
	VPConditionFixed VPConditionType = "fixed"
	VPConditionPer   VPConditionType = "per"
	VPConditionOnce  VPConditionType = "once"
)

type VPPerCondition struct {
	ResourceType shared.ResourceType
	Amount       int
	Target       *string
	Tag          *shared.CardTag
}

type VPCondition struct {
	Amount     int
	Condition  VPConditionType
	MaxTrigger *int
	Per        *VPPerCondition
}

type VPRecalculationContext interface {
	GetCardStorage(playerID string, cardID string) int
	CountPlayerTagsByType(playerID string, tagType shared.CardTag) int
	CountAllTilesOfType(tileType shared.ResourceType) int
}

type VPGranter struct {
	CardID        string
	CardName      string
	Description   string
	VPConditions  []VPCondition
	ComputedValue int
}

type VPGranters struct {
	mu       sync.RWMutex
	granters []VPGranter
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

func NewVPGranters(eventBus *events.EventBusImpl, gameID, playerID string) *VPGranters {
	return &VPGranters{
		granters: make([]VPGranter, 0),
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (vg *VPGranters) Add(granter VPGranter) {
	vg.mu.Lock()
	defer vg.mu.Unlock()
	vg.granters = append(vg.granters, granter)
}

func (vg *VPGranters) Prepend(granter VPGranter) {
	vg.mu.Lock()
	defer vg.mu.Unlock()
	vg.granters = append([]VPGranter{granter}, vg.granters...)
}

// RemoveByCardID removes all VP granters for the given card ID.
func (vg *VPGranters) RemoveByCardID(cardID string) {
	vg.mu.Lock()
	defer vg.mu.Unlock()
	filtered := vg.granters[:0]
	for _, g := range vg.granters {
		if g.CardID != cardID {
			filtered = append(filtered, g)
		}
	}
	vg.granters = filtered
}

func (vg *VPGranters) GetAll() []VPGranter {
	vg.mu.RLock()
	defer vg.mu.RUnlock()
	result := make([]VPGranter, len(vg.granters))
	copy(result, vg.granters)
	return result
}

func (vg *VPGranters) TotalComputedVP() int {
	vg.mu.RLock()
	defer vg.mu.RUnlock()
	total := 0
	for _, g := range vg.granters {
		total += g.ComputedValue
	}
	return total
}

const vpTargetSelfCard = "self-card"

func (vg *VPGranters) RecalculateAll(ctx VPRecalculationContext) {
	vg.mu.Lock()
	defer vg.mu.Unlock()

	oldTotal := 0
	for _, g := range vg.granters {
		oldTotal += g.ComputedValue
	}

	for i := range vg.granters {
		vg.granters[i].ComputedValue = evaluateGranterVP(&vg.granters[i], vg.playerID, ctx)
	}

	newTotal := 0
	for _, g := range vg.granters {
		newTotal += g.ComputedValue
	}

	if oldTotal != newTotal {
		events.Publish(vg.eventBus, events.VictoryPointsChangedEvent{
			GameID:    vg.gameID,
			PlayerID:  vg.playerID,
			OldPoints: oldTotal,
			NewPoints: newTotal,
			Source:    "vp-granters",
			Timestamp: time.Now(),
		})
	}
}

func evaluateGranterVP(granter *VPGranter, playerID string, ctx VPRecalculationContext) int {
	total := 0
	for _, cond := range granter.VPConditions {
		total += evaluateVPCondition(cond, granter.CardID, playerID, ctx)
	}
	return total
}

func evaluateVPCondition(cond VPCondition, cardID string, playerID string, ctx VPRecalculationContext) int {
	switch cond.Condition {
	case VPConditionFixed, VPConditionOnce:
		return cond.Amount

	case VPConditionPer:
		if cond.Per == nil {
			return 0
		}
		count := countPerCondition(cond.Per, cardID, playerID, ctx)
		if cond.Per.Amount <= 0 {
			return 0
		}
		triggers := count / cond.Per.Amount

		if cond.MaxTrigger != nil && *cond.MaxTrigger >= 0 && triggers > *cond.MaxTrigger {
			triggers = *cond.MaxTrigger
		}

		return cond.Amount * triggers

	default:
		return 0
	}
}

func countPerCondition(per *VPPerCondition, cardID string, playerID string, ctx VPRecalculationContext) int {
	if per.Target != nil && *per.Target == vpTargetSelfCard {
		return ctx.GetCardStorage(playerID, cardID)
	}

	if per.Tag != nil {
		return ctx.CountPlayerTagsByType(playerID, *per.Tag)
	}

	switch per.ResourceType {
	case shared.ResourceOceanTile, shared.ResourceCityTile, shared.ResourceGreeneryTile, shared.ResourceColonyTile:
		return ctx.CountAllTilesOfType(per.ResourceType)
	default:
		return ctx.CountPlayerTagsByType(playerID, shared.CardTag(per.ResourceType))
	}
}
