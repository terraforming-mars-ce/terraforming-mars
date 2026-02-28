package action_test

import (
	"context"
	"slices"
	"testing"

	"terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/test/testutil"
)

func TestDeckRecycling_UnselectedCardsFromCardDraw(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	log := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	// Draw 4 cards to set up a pending card draw selection
	drawnCards, err := testGame.Deck().DrawProjectCards(ctx, 4)
	testutil.AssertNoError(t, err, "draw project cards")

	// Set up pending card draw selection (buy up to 4 cards at 3 MC each)
	testutil.SetPlayerCredits(ctx, p, 100)
	p.Selection().SetPendingCardDrawSelection(&player.PendingCardDrawSelection{
		AvailableCards: drawnCards,
		FreeTakeCount:  0,
		MaxBuyCount:    4,
		CardBuyCost:    3,
		Source:         "research-phase",
	})

	// Player buys only 1 card, the other 3 should be discarded
	action := confirmation.NewConfirmCardDrawAction(repo, cardRegistry, log)
	err = action.Execute(ctx, testGame.ID(), playerID, []string{}, []string{drawnCards[0]})
	testutil.AssertNoError(t, err, "confirm card draw")

	discardPile := testGame.Deck().DiscardPile()
	testutil.AssertEqual(t, 3, len(discardPile), "discard pile should have 3 unselected cards")

	for _, cardID := range drawnCards[1:] {
		testutil.AssertTrue(t, slices.Contains(discardPile, cardID),
			"discard pile should contain unselected card "+cardID)
	}
}

func TestDeckRecycling_DiscardedCardsFromHand(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	log := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	// Add cards to player hand
	cardsInHand := []string{"card-power-plant", "card-asteroid", "card-water-import"}
	for _, cardID := range cardsInHand {
		p.Hand().AddCard(cardID)
	}

	// Set up pending card discard selection
	p.Selection().SetPendingCardDiscardSelection(&player.PendingCardDiscardSelection{
		MinCards: 1,
		MaxCards: 2,
		Source:   "test-discard",
	})

	// Player discards 2 cards
	cardsToDiscard := []string{"card-power-plant", "card-asteroid"}
	action := confirmation.NewConfirmCardDiscardAction(repo, cardRegistry, log)
	err := action.Execute(ctx, testGame.ID(), playerID, cardsToDiscard)
	testutil.AssertNoError(t, err, "confirm card discard")

	discardPile := testGame.Deck().DiscardPile()
	testutil.AssertEqual(t, 2, len(discardPile), "discard pile should have 2 discarded cards")

	for _, cardID := range cardsToDiscard {
		testutil.AssertTrue(t, slices.Contains(discardPile, cardID),
			"discard pile should contain discarded card "+cardID)
	}
}

func TestDeckRecycling_SoldPatents(t *testing.T) {
	testGame, repo, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	log := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	// Add cards to player hand
	cardsInHand := []string{"card-power-plant", "card-asteroid"}
	for _, cardID := range cardsInHand {
		p.Hand().AddCard(cardID)
	}

	// Set up pending card selection for sell patents
	p.Selection().SetPendingCardSelection(&player.PendingCardSelection{
		AvailableCards: cardsInHand,
		MinCards:       0,
		MaxCards:       len(cardsInHand),
		CardRewards:    map[string]int{"card-power-plant": 1, "card-asteroid": 1},
		Source:         "sell-patents",
	})

	action := confirmation.NewConfirmSellPatentsAction(repo, log)
	err := action.Execute(ctx, testGame.ID(), playerID, cardsInHand)
	testutil.AssertNoError(t, err, "confirm sell patents")

	discardPile := testGame.Deck().DiscardPile()
	testutil.AssertEqual(t, 2, len(discardPile), "discard pile should have 2 sold cards")

	for _, cardID := range cardsInHand {
		testutil.AssertTrue(t, slices.Contains(discardPile, cardID),
			"discard pile should contain sold card "+cardID)
	}
}

func TestDeckRecycling_UnselectedStartingCards(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	ctx := context.Background()
	log := testutil.TestLogger()

	err := testGame.UpdateStatus(ctx, game.GameStatusActive)
	testutil.AssertNoError(t, err, "update game status")
	err = testGame.UpdatePhase(ctx, game.GamePhaseStartingCardSelection)
	testutil.AssertNoError(t, err, "update game phase")

	playerID := "player-1"

	// Set up the starting cards phase with 10 available project cards and 2 corporations
	availableCards := []string{
		"card-power-plant", "card-asteroid", "card-water-import", "card-ai-central",
		"card-aquifer-pumping", "card-asteroid-mining", "card-biomass-combustors",
		"card-building-industries", "card-capital", "card-carbonate-processing",
	}
	availableCorps := []string{"corp-tharsis-republic", "corp-mining-guild"}

	err = testGame.SetSelectStartingCardsPhase(ctx, playerID, &player.SelectStartingCardsPhase{
		AvailableCards:        availableCards,
		AvailableCorporations: availableCorps,
	})
	testutil.AssertNoError(t, err, "set starting cards phase")

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 100)

	// Player selects 3 out of 10 project cards
	selectedCards := []string{"card-power-plant", "card-asteroid", "card-water-import"}
	action := turn_management.NewSelectStartingCardsAction(repo, cardRegistry, log)
	err = action.Execute(ctx, testGame.ID(), playerID, selectedCards, "corp-tharsis-republic")
	testutil.AssertNoError(t, err, "select starting cards")

	discardPile := testGame.Deck().DiscardPile()
	testutil.AssertEqual(t, 7, len(discardPile), "discard pile should have 7 unselected project cards")

	// Verify the selected cards are NOT in the discard pile
	for _, cardID := range selectedCards {
		testutil.AssertFalse(t, slices.Contains(discardPile, cardID),
			"discard pile should not contain selected card "+cardID)
	}

	// Verify unselected cards ARE in the discard pile
	for _, cardID := range availableCards {
		if !slices.Contains(selectedCards, cardID) {
			testutil.AssertTrue(t, slices.Contains(discardPile, cardID),
				"discard pile should contain unselected card "+cardID)
		}
	}

	// Verify corporations are NOT in the discard pile
	testutil.AssertFalse(t, slices.Contains(discardPile, "corp-tharsis-republic"),
		"discard pile should not contain selected corporation")
	testutil.AssertFalse(t, slices.Contains(discardPile, "corp-mining-guild"),
		"discard pile should not contain unselected corporation")
}

func TestDeckRecycling_AutoShuffleOnEmptyDrawPile(t *testing.T) {
	ctx := context.Background()

	// Create a deck with only 2 project cards in draw pile
	gameDeck := deck.NewDeck("test-game", []string{"card-a", "card-b"}, nil, nil)

	// Draw all cards to empty the draw pile
	drawn, err := gameDeck.DrawProjectCards(ctx, 2)
	testutil.AssertNoError(t, err, "initial draw")
	testutil.AssertEqual(t, 2, len(drawn), "should draw 2 cards")
	testutil.AssertEqual(t, 0, gameDeck.GetAvailableCardCount(), "draw pile should be empty")

	// Add cards to discard pile
	err = gameDeck.Discard(ctx, []string{"card-c", "card-d", "card-e"})
	testutil.AssertNoError(t, err, "discard cards")
	testutil.AssertEqual(t, 3, len(gameDeck.DiscardPile()), "discard pile should have 3 cards")

	// Drawing should auto-shuffle the discard pile back in
	drawn, err = gameDeck.DrawProjectCards(ctx, 2)
	testutil.AssertNoError(t, err, "draw after auto-shuffle")
	testutil.AssertEqual(t, 2, len(drawn), "should draw 2 cards after auto-shuffle")
	testutil.AssertEqual(t, 1, gameDeck.GetAvailableCardCount(), "draw pile should have 1 remaining card")
	testutil.AssertEqual(t, 0, len(gameDeck.DiscardPile()), "discard pile should be empty after shuffle")
	testutil.AssertEqual(t, 1, gameDeck.ShuffleCount(), "shuffle count should be 1")
}

func TestDeckRecycling_AutoShuffleFailsWhenBothPilesEmpty(t *testing.T) {
	ctx := context.Background()

	// Create a deck with only 1 project card
	gameDeck := deck.NewDeck("test-game", []string{"card-a"}, nil, nil)

	// Draw the only card
	_, err := gameDeck.DrawProjectCards(ctx, 1)
	testutil.AssertNoError(t, err, "initial draw")

	// Both draw pile and discard pile are empty — draw should fail
	_, err = gameDeck.DrawProjectCards(ctx, 1)
	testutil.AssertError(t, err, "draw from empty deck should fail")
}
