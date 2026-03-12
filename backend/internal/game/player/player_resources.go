package player

import (
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// PlayerResources manages player resources, production, and scoring.
type PlayerResources struct {
	ds       *datastore.DataStore
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

func newResources(ds *datastore.DataStore, eventBus *events.EventBusImpl, gameID, playerID string) *PlayerResources {
	return &PlayerResources{
		ds:       ds,
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (r *PlayerResources) update(fn func(s *datastore.PlayerState)) {
	if err := r.ds.UpdatePlayer(r.gameID, r.playerID, fn); err != nil {
		logger.Get().Warn("Failed to update player state", zap.String("game_id", r.gameID), zap.String("player_id", r.playerID), zap.Error(err))
	}
}

func (r *PlayerResources) read(fn func(s *datastore.PlayerState)) {
	if err := r.ds.ReadPlayer(r.gameID, r.playerID, fn); err != nil {
		logger.Get().Warn("Failed to read player state", zap.String("game_id", r.gameID), zap.String("player_id", r.playerID), zap.Error(err))
	}
}

func (r *PlayerResources) Get() shared.Resources {
	var res shared.Resources
	r.read(func(s *datastore.PlayerState) {
		res = s.Resources
	})
	return res
}

func (r *PlayerResources) Production() shared.Production {
	var prod shared.Production
	r.read(func(s *datastore.PlayerState) {
		prod = s.Production
	})
	return prod
}

func (r *PlayerResources) TerraformRating() int {
	var tr int
	r.read(func(s *datastore.PlayerState) {
		tr = s.TerraformRating
	})
	return tr
}

func (r *PlayerResources) Storage() map[string]int {
	var storageCopy map[string]int
	r.read(func(s *datastore.PlayerState) {
		storageCopy = make(map[string]int, len(s.ResourceStorage))
		for k, v := range s.ResourceStorage {
			storageCopy[k] = v
		}
	})
	return storageCopy
}

const (
	baseSteelValue    = 2
	baseTitaniumValue = 3
)

// PaymentSubstitutes returns all payment substitutes including steel/titanium with dynamic values.
func (r *PlayerResources) PaymentSubstitutes() []shared.PaymentSubstitute {
	var substitutes []shared.PaymentSubstitute
	r.read(func(s *datastore.PlayerState) {
		substitutes = []shared.PaymentSubstitute{
			{ResourceType: shared.ResourceSteel, ConversionRate: baseSteelValue + s.ValueModifiers[shared.ResourceSteel]},
			{ResourceType: shared.ResourceTitanium, ConversionRate: baseTitaniumValue + s.ValueModifiers[shared.ResourceTitanium]},
		}
		substitutes = append(substitutes, s.PaymentSubstitutes...)
	})
	return substitutes
}

func (r *PlayerResources) AddPaymentSubstitute(resourceType shared.ResourceType, conversionRate int) {
	r.update(func(s *datastore.PlayerState) {
		s.PaymentSubstitutes = append(s.PaymentSubstitutes, shared.PaymentSubstitute{
			ResourceType:   resourceType,
			ConversionRate: conversionRate,
		})
	})
}

// ValueModifiers returns a copy of the value modifiers map
func (r *PlayerResources) ValueModifiers() map[shared.ResourceType]int {
	var modifiersCopy map[shared.ResourceType]int
	r.read(func(s *datastore.PlayerState) {
		modifiersCopy = make(map[shared.ResourceType]int, len(s.ValueModifiers))
		for k, v := range s.ValueModifiers {
			modifiersCopy[k] = v
		}
	})
	return modifiersCopy
}

// AddValueModifier adds a value modifier for a resource type
func (r *PlayerResources) AddValueModifier(resourceType shared.ResourceType, amount int) {
	r.update(func(s *datastore.PlayerState) {
		if s.ValueModifiers == nil {
			s.ValueModifiers = make(map[shared.ResourceType]int)
		}
		s.ValueModifiers[resourceType] += amount
	})
}

// GetValueModifier returns the total value modifier for a resource type
func (r *PlayerResources) GetValueModifier(resourceType shared.ResourceType) int {
	var val int
	r.read(func(s *datastore.PlayerState) {
		val = s.ValueModifiers[resourceType]
	})
	return val
}

func (r *PlayerResources) Set(resources shared.Resources) {
	r.update(func(s *datastore.PlayerState) {
		s.Resources = resources
	})

	if r.eventBus != nil {
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			Changes:   make(map[string]int),
			Timestamp: time.Now(),
		})
	}
}

func (r *PlayerResources) SetProduction(production shared.Production) {
	var oldProduction, newProduction shared.Production
	r.update(func(s *datastore.PlayerState) {
		oldProduction = s.Production
		s.Production = production
		newProduction = s.Production
	})

	if r.eventBus != nil {
		resourceTypes := []struct {
			name     string
			oldValue int
			newValue int
		}{
			{"credits", oldProduction.Credits, newProduction.Credits},
			{"steel", oldProduction.Steel, newProduction.Steel},
			{"titanium", oldProduction.Titanium, newProduction.Titanium},
			{"plants", oldProduction.Plants, newProduction.Plants},
			{"energy", oldProduction.Energy, newProduction.Energy},
			{"heat", oldProduction.Heat, newProduction.Heat},
		}

		for _, rt := range resourceTypes {
			events.Publish(r.eventBus, events.ProductionChangedEvent{
				GameID:        r.gameID,
				PlayerID:      r.playerID,
				ResourceType:  rt.name,
				OldProduction: rt.oldValue,
				NewProduction: rt.newValue,
				Timestamp:     time.Now(),
			})
		}
	}
}

func (r *PlayerResources) SetTerraformRating(tr int) {
	var oldRating int
	r.update(func(s *datastore.PlayerState) {
		oldRating = s.TerraformRating
		s.TerraformRating = tr
	})

	if r.eventBus != nil {
		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			OldRating: oldRating,
			NewRating: tr,
			Timestamp: time.Now(),
		})
	}
}

func (r *PlayerResources) Add(changes map[shared.ResourceType]int) {
	r.update(func(s *datastore.PlayerState) {
		for resourceType, amount := range changes {
			switch resourceType {
			case shared.ResourceCredit:
				s.Resources.Credits += amount
			case shared.ResourceSteel:
				s.Resources.Steel += amount
			case shared.ResourceTitanium:
				s.Resources.Titanium += amount
			case shared.ResourcePlant:
				s.Resources.Plants += amount
			case shared.ResourceEnergy:
				s.Resources.Energy += amount
			case shared.ResourceHeat:
				s.Resources.Heat += amount
			}
		}
	})

	if r.eventBus != nil {
		changesMap := make(map[string]int, len(changes))
		for resourceType, amount := range changes {
			changesMap[string(resourceType)] = amount
		}

		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			Changes:   changesMap,
			Timestamp: time.Now(),
		})
	}
}

func (r *PlayerResources) AddProduction(changes map[shared.ResourceType]int) {
	var oldProduction, newProduction shared.Production
	r.update(func(s *datastore.PlayerState) {
		oldProduction = s.Production
		for resourceType, amount := range changes {
			switch resourceType {
			case shared.ResourceCreditProduction:
				s.Production.Credits += amount
				if s.Production.Credits < shared.MinCreditProduction {
					s.Production.Credits = shared.MinCreditProduction
				}
			case shared.ResourceSteelProduction:
				s.Production.Steel += amount
				if s.Production.Steel < shared.MinOtherProduction {
					s.Production.Steel = shared.MinOtherProduction
				}
			case shared.ResourceTitaniumProduction:
				s.Production.Titanium += amount
				if s.Production.Titanium < shared.MinOtherProduction {
					s.Production.Titanium = shared.MinOtherProduction
				}
			case shared.ResourcePlantProduction:
				s.Production.Plants += amount
				if s.Production.Plants < shared.MinOtherProduction {
					s.Production.Plants = shared.MinOtherProduction
				}
			case shared.ResourceEnergyProduction:
				s.Production.Energy += amount
				if s.Production.Energy < shared.MinOtherProduction {
					s.Production.Energy = shared.MinOtherProduction
				}
			case shared.ResourceHeatProduction:
				s.Production.Heat += amount
				if s.Production.Heat < shared.MinOtherProduction {
					s.Production.Heat = shared.MinOtherProduction
				}
			}
		}
		newProduction = s.Production
	})

	if r.eventBus != nil {
		for resourceType := range changes {
			var oldValue, newValue int
			resourceName := string(resourceType)

			switch resourceType {
			case shared.ResourceCreditProduction:
				oldValue = oldProduction.Credits
				newValue = newProduction.Credits
				resourceName = "credits"
			case shared.ResourceSteelProduction:
				oldValue = oldProduction.Steel
				newValue = newProduction.Steel
				resourceName = "steel"
			case shared.ResourceTitaniumProduction:
				oldValue = oldProduction.Titanium
				newValue = newProduction.Titanium
				resourceName = "titanium"
			case shared.ResourcePlantProduction:
				oldValue = oldProduction.Plants
				newValue = newProduction.Plants
				resourceName = "plants"
			case shared.ResourceEnergyProduction:
				oldValue = oldProduction.Energy
				newValue = newProduction.Energy
				resourceName = "energy"
			case shared.ResourceHeatProduction:
				oldValue = oldProduction.Heat
				newValue = newProduction.Heat
				resourceName = "heat"
			}

			events.Publish(r.eventBus, events.ProductionChangedEvent{
				GameID:        r.gameID,
				PlayerID:      r.playerID,
				ResourceType:  resourceName,
				OldProduction: oldValue,
				NewProduction: newValue,
				Timestamp:     time.Now(),
			})
		}
	}
}

func (r *PlayerResources) UpdateTerraformRating(delta int) {
	var oldRating, newRating int
	r.update(func(s *datastore.PlayerState) {
		oldRating = s.TerraformRating
		s.TerraformRating += delta
		newRating = s.TerraformRating
	})

	if r.eventBus != nil {
		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			OldRating: oldRating,
			NewRating: newRating,
			Timestamp: time.Now(),
		})
	}
}

// AddToStorage adds resources to a specific card's storage
func (r *PlayerResources) AddToStorage(cardID string, amount int) {
	var oldAmount, newAmount int
	r.update(func(s *datastore.PlayerState) {
		if s.ResourceStorage == nil {
			s.ResourceStorage = make(map[string]int)
		}
		oldAmount = s.ResourceStorage[cardID]
		s.ResourceStorage[cardID] += amount
		newAmount = s.ResourceStorage[cardID]
	})

	if r.eventBus != nil {
		events.Publish(r.eventBus, events.ResourceStorageChangedEvent{
			GameID:    r.gameID,
			PlayerID:  r.playerID,
			CardID:    cardID,
			OldAmount: oldAmount,
			NewAmount: newAmount,
			Timestamp: time.Now(),
		})
	}
}

// GetCardStorage returns the amount of resources stored on a specific card
func (r *PlayerResources) GetCardStorage(cardID string) int {
	var val int
	r.read(func(s *datastore.PlayerState) {
		val = s.ResourceStorage[cardID]
	})
	return val
}

// RemoveCardStorage removes the storage entry for a specific card
func (r *PlayerResources) RemoveCardStorage(cardID string) {
	r.update(func(s *datastore.PlayerState) {
		delete(s.ResourceStorage, cardID)
	})
}

// ClearPaymentSubstitutes removes all non-standard payment substitutes
func (r *PlayerResources) ClearPaymentSubstitutes() {
	r.update(func(s *datastore.PlayerState) {
		s.PaymentSubstitutes = []shared.PaymentSubstitute{}
	})
}

// ClearValueModifiers resets all value modifiers to zero
func (r *PlayerResources) ClearValueModifiers() {
	r.update(func(s *datastore.PlayerState) {
		s.ValueModifiers = make(map[shared.ResourceType]int)
	})
}

// AddStoragePaymentSubstitute registers a card's storage resources as usable for payment
func (r *PlayerResources) AddStoragePaymentSubstitute(sub shared.StoragePaymentSubstitute) {
	r.update(func(s *datastore.PlayerState) {
		s.StoragePaymentSubstitutes = append(s.StoragePaymentSubstitutes, sub)
	})
}

// StoragePaymentSubstitutes returns all storage payment substitutes
func (r *PlayerResources) StoragePaymentSubstitutes() []shared.StoragePaymentSubstitute {
	var result []shared.StoragePaymentSubstitute
	r.read(func(s *datastore.PlayerState) {
		result = make([]shared.StoragePaymentSubstitute, len(s.StoragePaymentSubstitutes))
		copy(result, s.StoragePaymentSubstitutes)
	})
	return result
}

// GetStoragePaymentSubstitute returns the storage payment substitute for a specific card, or nil
func (r *PlayerResources) GetStoragePaymentSubstitute(cardID string) *shared.StoragePaymentSubstitute {
	var result *shared.StoragePaymentSubstitute
	r.read(func(s *datastore.PlayerState) {
		for _, sub := range s.StoragePaymentSubstitutes {
			if sub.CardID == cardID {
				subCopy := sub
				result = &subCopy
				return
			}
		}
	})
	return result
}
