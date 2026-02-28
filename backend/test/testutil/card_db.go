package testutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"
)

var (
	realCardDB cards.CardRegistry
	cardByName map[string]gamecards.Card
	cardByID   map[string]gamecards.Card
	loadOnce   sync.Once
)

func loadCards() {
	_, currentFile, _, _ := runtime.Caller(0)
	jsonPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "assets", "terraforming_mars_cards.json")

	cardList, err := cards.LoadCardsFromJSON(jsonPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load card DB for tests: %v", err))
	}

	cardByName = make(map[string]gamecards.Card, len(cardList))
	cardByID = make(map[string]gamecards.Card, len(cardList))
	for _, c := range cardList {
		cardByName[c.Name] = c
		cardByID[c.ID] = c
	}

	realCardDB = cards.NewInMemoryCardRegistry(cardList)
}

// GetCardDB returns the real card registry loaded from the JSON database.
// The JSON is loaded once and shared across all tests.
func GetCardDB() cards.CardRegistry {
	loadOnce.Do(loadCards)
	return realCardDB
}

// GetCardByName returns a card from the real DB by its name.
// Panics if the card is not found.
func GetCardByName(name string) gamecards.Card {
	loadOnce.Do(loadCards)
	card, ok := cardByName[name]
	if !ok {
		panic(fmt.Sprintf("card not found by name: %s", name))
	}
	return card
}

// GetCardByID returns a card from the real DB by its ID.
// Panics if the card is not found.
func GetCardByID(id string) gamecards.Card {
	loadOnce.Do(loadCards)
	card, ok := cardByID[id]
	if !ok {
		panic(fmt.Sprintf("card not found by ID: %s", id))
	}
	return card
}

// CardID returns the real card ID for a given card name.
// Useful for tests that need to reference cards by name but pass IDs to actions.
func CardID(name string) string {
	return GetCardByName(name).ID
}
