package player

import (
	"log/slog"
	"time"

	"openmars/internal/events"
	"openmars/internal/game/datastore"
)

// PlayedCards manages all cards a player has played.
type PlayedCards struct {
	ds       *datastore.DataStore
	eventBus *events.EventBusImpl
	gameID   string
	playerID string
}

func newPlayedCards(ds *datastore.DataStore, eventBus *events.EventBusImpl, gameID, playerID string) *PlayedCards {
	return &PlayedCards{
		ds:       ds,
		eventBus: eventBus,
		gameID:   gameID,
		playerID: playerID,
	}
}

func (pc *PlayedCards) update(fn func(s *datastore.PlayerState)) {
	if err := pc.ds.UpdatePlayer(pc.gameID, pc.playerID, fn); err != nil {
		slog.Default().Warn("Failed to update player state", slog.String("game_id", pc.gameID), slog.String("player_id", pc.playerID), slog.Any("error", err))
	}
}

func (pc *PlayedCards) read(fn func(s *datastore.PlayerState)) {
	if err := pc.ds.ReadPlayer(pc.gameID, pc.playerID, fn); err != nil {
		slog.Default().Warn("Failed to read player state", slog.String("game_id", pc.gameID), slog.String("player_id", pc.playerID), slog.Any("error", err))
	}
}

// Cards returns a copy of all played cards
func (pc *PlayedCards) Cards() []string {
	var cardsCopy []string
	pc.read(func(s *datastore.PlayerState) {
		cardsCopy = make([]string, len(s.PlayedCardIDs))
		copy(cardsCopy, s.PlayedCardIDs)
	})
	return cardsCopy
}

// Contains checks if a specific card has been played
func (pc *PlayedCards) Contains(cardID string) bool {
	var found bool
	pc.read(func(s *datastore.PlayerState) {
		for _, id := range s.PlayedCardIDs {
			if id == cardID {
				found = true
				return
			}
		}
	})
	return found
}

// AddCard adds a card to played cards
func (pc *PlayedCards) AddCard(cardID, cardName, cardType string, tags []string) {
	pc.update(func(s *datastore.PlayerState) {
		s.PlayedCardIDs = append(s.PlayedCardIDs, cardID)
	})

	if pc.eventBus != nil {
		events.Publish(pc.eventBus, events.CardPlayedEvent{
			GameID:    pc.gameID,
			PlayerID:  pc.playerID,
			CardID:    cardID,
			CardName:  cardName,
			CardType:  cardType,
			Timestamp: time.Now(),
		})

		for _, tag := range tags {
			events.Publish(pc.eventBus, events.TagPlayedEvent{
				GameID:    pc.gameID,
				PlayerID:  pc.playerID,
				CardID:    cardID,
				CardName:  cardName,
				Tag:       tag,
				Timestamp: time.Now(),
			})
		}
	}
}

// RemoveCard removes a card from played cards
func (pc *PlayedCards) RemoveCard(cardID string) bool {
	var removed bool
	pc.update(func(s *datastore.PlayerState) {
		for i, id := range s.PlayedCardIDs {
			if id == cardID {
				s.PlayedCardIDs = append(s.PlayedCardIDs[:i], s.PlayedCardIDs[i+1:]...)
				removed = true
				return
			}
		}
	})
	return removed
}

// SetCards replaces all played cards
func (pc *PlayedCards) SetCards(cards []string) {
	pc.update(func(s *datastore.PlayerState) {
		if cards == nil {
			s.PlayedCardIDs = []string{}
		} else {
			s.PlayedCardIDs = make([]string, len(cards))
			copy(s.PlayedCardIDs, cards)
		}
	})
}

// Count returns the number of played cards
func (pc *PlayedCards) Count() int {
	var count int
	pc.read(func(s *datastore.PlayerState) {
		count = len(s.PlayedCardIDs)
	})
	return count
}
