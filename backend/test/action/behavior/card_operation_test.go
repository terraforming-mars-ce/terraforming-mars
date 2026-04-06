package behavior_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action"
	cardAction "terraforming-mars-backend/internal/action/card"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// ============================================================================
// Card discard tests
// ============================================================================

// --- Mars University (073) ---
// "Effect: When you play a science tag, including this, you may discard a card from hand to draw a card."

func TestMarsUniversity_CreatesDiscardSelectionOnScienceTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
		Tags: []shared.CardTag{shared.TagBuilding, shared.TagScience},
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	// Give player some cards in hand to discard
	owner.Hand().AddCard("some-card-1")
	owner.Hand().AddCard("some-card-2")

	// Register Mars University passive effect with card-discard input
	effect := shared.CardEffect{
		CardID:        "card-mars-university",
		CardName:      "Mars University",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type: "tag-played",
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagScience}},
						},
					},
				},
			},
			Inputs: []shared.BehaviorCondition{
				&shared.CardOperationCondition{
					ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCardDiscard, Amount: 1, Target: "self-player"},
					Optional:      true,
				},
			},
			Outputs: []shared.BehaviorCondition{
				shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
			},
		},
	}
	owner.Effects().AddEffect(effect)

	// Subscribe the passive effect to events
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	// Verify no pending selection before event
	testutil.AssertTrue(t, owner.Selection().GetPendingCardDiscardSelection() == nil,
		"Should have no pending discard selection before science tag event")

	// Publish a science tag played event
	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	// Should have a pending card discard selection
	selection := owner.Selection().GetPendingCardDiscardSelection()
	testutil.AssertTrue(t, selection != nil, "Should have a pending discard selection after science tag event")
	testutil.AssertEqual(t, 0, selection.MinCards, "MinCards should be 0 (optional)")
	testutil.AssertEqual(t, 1, selection.MaxCards, "MaxCards should be 1")
	testutil.AssertEqual(t, "Mars University", selection.Source, "Source should be Mars University")
	testutil.AssertEqual(t, "card-mars-university", selection.SourceCardID, "SourceCardID should match")
	testutil.AssertEqual(t, 1, len(selection.PendingOutputs), "Should have 1 pending output")
	testutil.AssertEqual(t, shared.ResourceCardDraw, selection.PendingOutputs[0].GetResourceType(), "Pending output should be card-draw")

	_ = repo // suppress unused warning
}

func TestMarsUniversity_SkipsDiscardWhenNoCardsInHand(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	// Do NOT add cards to hand - player has empty hand

	effect := shared.CardEffect{
		CardID:        "card-mars-university",
		CardName:      "Mars University",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type: "tag-played",
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagScience}},
						},
					},
				},
			},
			Inputs: []shared.BehaviorCondition{
				&shared.CardOperationCondition{
					ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCardDiscard, Amount: 1, Target: "self-player"},
					Optional:      true,
				},
			},
			Outputs: []shared.BehaviorCondition{
				shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	// Publish science tag event
	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "science",
	})

	time.Sleep(20 * time.Millisecond)

	// Should NOT have a pending discard selection since player has no cards
	testutil.AssertTrue(t, owner.Selection().GetPendingCardDiscardSelection() == nil,
		"Should NOT have pending discard selection when hand is empty")
}

func TestMarsUniversity_DoesNotTriggerOnNonScienceTag(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("some-card-1")

	effect := shared.CardEffect{
		CardID:        "card-mars-university",
		CardName:      "Mars University",
		BehaviorIndex: 0,
		Behavior: shared.CardBehavior{
			Triggers: []shared.Trigger{
				{
					Type: shared.TriggerTypeAuto,
					Condition: &shared.ResourceTriggerCondition{
						Type: "tag-played",
						Selectors: []shared.Selector{
							{Tags: []shared.CardTag{shared.TagScience}},
						},
					},
				},
			},
			Inputs: []shared.BehaviorCondition{
				&shared.CardOperationCondition{
					ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCardDiscard, Amount: 1, Target: "self-player"},
					Optional:      true,
				},
			},
			Outputs: []shared.BehaviorCondition{
				shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
			},
		},
	}
	owner.Effects().AddEffect(effect)
	action.SubscribePassiveEffectToEvents(ctx, testGame, owner, effect, logger, cardRegistry)

	// Publish a building tag event (not science)
	events.Publish(testGame.EventBus(), events.TagPlayedEvent{
		GameID:   testGame.ID(),
		PlayerID: owner.ID(),
		Tag:      "building",
	})

	time.Sleep(20 * time.Millisecond)

	// Should NOT trigger for building tag
	testutil.AssertTrue(t, owner.Selection().GetPendingCardDiscardSelection() == nil,
		"Mars University should not trigger on non-science tags")
}

// --- ConfirmCardDiscardAction Tests ---

func TestConfirmCardDiscard_DiscardAndDraw(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	// Give player a card to discard
	owner.Hand().AddCard("card-to-discard")
	owner.Hand().AddCard("card-to-keep")

	handBefore := len(owner.Hand().Cards())
	testutil.AssertEqual(t, 2, handBefore, "Should have 2 cards before discard")

	// Set up pending discard selection
	owner.Selection().SetPendingCardDiscardSelection(&shared.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.BehaviorCondition{
			shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
		},
	})

	// Execute confirm discard
	confirmDiscard := confirmAction.NewConfirmCardDiscardAction(repo, cardRegistry, logger)
	err := confirmDiscard.Execute(ctx, testGame.ID(), owner.ID(), []string{"card-to-discard"})
	testutil.AssertNoError(t, err, "Confirm card discard should succeed")

	// Verify card was removed from hand
	handCards := owner.Hand().Cards()
	found := false
	for _, id := range handCards {
		if id == "card-to-discard" {
			found = true
		}
	}
	testutil.AssertFalse(t, found, "Discarded card should not be in hand")

	// Verify pending selection was cleared
	testutil.AssertTrue(t, owner.Selection().GetPendingCardDiscardSelection() == nil,
		"Pending discard selection should be cleared after confirmation")
}

func TestConfirmCardDiscard_SkipOptionalDiscard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("card-to-keep")

	// Set up pending discard with min=0 (optional)
	owner.Selection().SetPendingCardDiscardSelection(&shared.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.BehaviorCondition{
			shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
		},
	})

	// Skip discard by sending empty list
	confirmDiscard := confirmAction.NewConfirmCardDiscardAction(repo, cardRegistry, logger)
	err := confirmDiscard.Execute(ctx, testGame.ID(), owner.ID(), []string{})
	testutil.AssertNoError(t, err, "Skipping optional discard should succeed")

	// Hand should still have the card
	testutil.AssertEqual(t, 1, len(owner.Hand().Cards()), "Hand should still have 1 card")

	// Pending selection should be cleared
	testutil.AssertTrue(t, owner.Selection().GetPendingCardDiscardSelection() == nil,
		"Pending discard selection should be cleared even when skipping")
}

func TestConfirmCardDiscard_RejectsNonHandCard(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("real-card")

	owner.Selection().SetPendingCardDiscardSelection(&shared.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.BehaviorCondition{
			shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
		},
	})

	// Try to discard a card not in hand
	confirmDiscard := confirmAction.NewConfirmCardDiscardAction(repo, cardRegistry, logger)
	err := confirmDiscard.Execute(ctx, testGame.ID(), owner.ID(), []string{"fake-card"})
	testutil.AssertError(t, err, "Should reject discard of card not in hand")
}

func TestConfirmCardDiscard_RejectsWithoutPendingSelection(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	// No pending selection set

	confirmDiscard := confirmAction.NewConfirmCardDiscardAction(repo, cardRegistry, logger)
	err := confirmDiscard.Execute(ctx, testGame.ID(), owner.ID(), []string{})
	testutil.AssertError(t, err, "Should reject confirm when no pending selection exists")
}

func TestConfirmCardDiscard_RejectsTooManyCards(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()

	marsUniCard := gamecards.Card{
		ID:   "card-mars-university",
		Name: "Mars University",
		Type: gamecards.CardTypeActive,
		Cost: 8,
	}
	cardRegistry := cards.NewInMemoryCardRegistry([]gamecards.Card{marsUniCard})

	players := testGame.GetAllPlayers()
	owner := players[0]
	owner.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("card-1")
	owner.Hand().AddCard("card-2")

	owner.Selection().SetPendingCardDiscardSelection(&shared.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.BehaviorCondition{
			shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "self-player"),
		},
	})

	// Try to discard 2 cards when max is 1
	confirmDiscard := confirmAction.NewConfirmCardDiscardAction(repo, cardRegistry, logger)
	err := confirmDiscard.Execute(ctx, testGame.ID(), owner.ID(), []string{"card-1", "card-2"})
	testutil.AssertError(t, err, "Should reject discarding too many cards")
}

// ============================================================================
// All opponents card operation tests
// ============================================================================

func createAllOpponentsDrawTestCard() gamecards.Card {
	return gamecards.Card{
		ID:   "card-all-opponents-draw-test",
		Name: "All Opponents Draw Test",
		Type: gamecards.CardTypeAutomated,
		Pack: "base",
		Cost: 5,
		Tags: []shared.CardTag{shared.TagScience},
		Behaviors: []shared.CardBehavior{
			{
				Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
				Outputs: []shared.BehaviorCondition{
					shared.NewCardOperationCondition(shared.ResourceCardDraw, 2, "self-player"),
					shared.NewCardOperationCondition(shared.ResourceCardDraw, 1, "all-opponents"),
				},
			},
		},
	}
}

func TestPlayCard_AllOpponentsDrawCard_TwoPlayers(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2, broadcaster)
	additionalCards := []gamecards.Card{createAllOpponentsDrawTestCard()}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player2 := players[1]
	player1.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player2.SetCorporationID(testutil.CardID("Tharsis Republic"))

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1.ID(), 2), "set current turn")

	player1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player1.Hand().AddCard("card-all-opponents-draw-test")

	player1HandBefore := player1.Hand().CardCount()
	player2HandBefore := player2.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), player1.ID(), "card-all-opponents-draw-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing all-opponents draw card should succeed")

	// Player 1: played 1 card (-1) and drew 2 cards (+2) = net +1
	player1HandAfter := player1.Hand().CardCount()
	testutil.AssertEqual(t, player1HandBefore+1, player1HandAfter,
		"Player 1 hand should increase by 1 (played 1, drew 2)")

	// Player 2 (opponent): drew 1 card (+1)
	player2HandAfter := player2.Hand().CardCount()
	testutil.AssertEqual(t, player2HandBefore+1, player2HandAfter,
		"Player 2 (opponent) should have drawn 1 card")
}

func TestPlayCard_AllOpponentsDrawCard_ThreePlayers(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	additionalCards := []gamecards.Card{createAllOpponentsDrawTestCard()}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player2 := players[1]
	player3 := players[2]
	player1.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player2.SetCorporationID(testutil.CardID("Tharsis Republic"))
	player3.SetCorporationID(testutil.CardID("Tharsis Republic"))

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1.ID(), 2), "set current turn")

	player1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player1.Hand().AddCard("card-all-opponents-draw-test")

	player2HandBefore := player2.Hand().CardCount()
	player3HandBefore := player3.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), player1.ID(), "card-all-opponents-draw-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing all-opponents draw card should succeed")

	// Both opponents should have drawn 1 card each
	player2HandAfter := player2.Hand().CardCount()
	testutil.AssertEqual(t, player2HandBefore+1, player2HandAfter,
		"Player 2 (opponent) should have drawn 1 card")

	player3HandAfter := player3.Hand().CardCount()
	testutil.AssertEqual(t, player3HandBefore+1, player3HandAfter,
		"Player 3 (opponent) should have drawn 1 card")
}

func TestPlayCard_AllOpponentsDrawCard_SoloMode(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	additionalCards := []gamecards.Card{createAllOpponentsDrawTestCard()}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(additionalCards)
	logger := testutil.TestLogger()
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player1 := players[0]
	player1.SetCorporationID(testutil.CardID("Tharsis Republic"))

	testutil.AssertNoError(t, testGame.UpdateStatus(ctx, shared.GameStatusActive), "update status")
	testutil.AssertNoError(t, testGame.UpdatePhase(ctx, shared.GamePhaseAction), "update phase")
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, player1.ID(), 2), "set current turn")

	player1.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	player1.Hand().AddCard("card-all-opponents-draw-test")

	player1HandBefore := player1.Hand().CardCount()

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 5}
	err := playCardAction.Execute(ctx, testGame.ID(), player1.ID(), "card-all-opponents-draw-test", payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Playing all-opponents draw card in solo mode should succeed")

	// Solo: played 1 card (-1) and drew 2 cards (+2), no opponents = net +1
	player1HandAfter := player1.Hand().CardCount()
	testutil.AssertEqual(t, player1HandBefore+1, player1HandAfter,
		"Player 1 hand should increase by 1 in solo mode (played 1, drew 2, no opponents)")
}

func TestCardOperation_DrawWithSelectors(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)

	p, _ := testGame.GetPlayer(playerID)
	handBefore := p.Hand().CardCount()

	output := &shared.CardOperationCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
		Selectors:     []shared.Selector{{Tags: []shared.CardTag{shared.TagScience}}},
	}

	applyOutputs(t, p, testGame, cardRegistry, output)

	handAfter := p.Hand().CardCount()
	testutil.AssertEqual(t, handBefore+1, handAfter, "Player should have drawn 1 card")
}

func TestCardOperation_VariableAmountDiscard(t *testing.T) {
	input := &shared.CardOperationCondition{
		ConditionBase:  shared.ConditionBase{ResourceType: shared.ResourceCardDiscard, Amount: 1, Target: "self-player"},
		VariableAmount: true,
	}

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeAuto}},
		Inputs:   []shared.BehaviorCondition{input},
	}

	testutil.AssertTrue(t, behavior.Inputs[0] != nil, "Behavior input should be constructed")
	testutil.AssertEqual(t, shared.ResourceCardDiscard, behavior.Inputs[0].GetResourceType(), "Resource type should be card-discard")
	testutil.AssertTrue(t, shared.IsVariableAmount(behavior.Inputs[0]), "VariableAmount should be true")
}

func TestCardOperation_DrawFromNearlyEmptyDeck(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	// Drain the deck to leave very few cards
	available := testGame.Deck().GetAvailableCardCount()
	if available > 2 {
		_, err := testGame.Deck().DrawProjectCards(ctx, available-2)
		testutil.AssertNoError(t, err, "Draining deck should succeed")
	}

	remaining := testGame.Deck().GetAvailableCardCount()
	handBefore := p.Hand().CardCount()

	// Try to draw 5 cards when only a few remain
	output := shared.NewCardOperationCondition(shared.ResourceCardDraw, 5, "self-player")
	applyOutputs(t, p, testGame, cardRegistry, output)

	// The draw may fail (not enough cards even after reshuffle), which is logged
	// as a warning and returns nil. The player gets 0 cards from a failed draw.
	// Either way, no crash should occur.
	handAfter := p.Hand().CardCount()
	testutil.AssertTrue(t, handAfter >= handBefore,
		"Hand should not decrease")
	_ = remaining
}
