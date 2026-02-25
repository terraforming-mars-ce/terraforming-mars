package card_effects_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/action"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Give player some cards in hand to discard
	owner.Hand().AddCard("some-card-1")
	owner.Hand().AddCard("some-card-2")

	// Register Mars University passive effect with card-discard input
	effect := player.CardEffect{
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
			Inputs: []shared.ResourceCondition{
				{
					ResourceType: shared.ResourceCardDiscard,
					Amount:       1,
					Target:       "self-player",
					Optional:     true,
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
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
	testutil.AssertEqual(t, shared.ResourceCardDraw, selection.PendingOutputs[0].ResourceType, "Pending output should be card-draw")

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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Do NOT add cards to hand - player has empty hand

	effect := player.CardEffect{
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
			Inputs: []shared.ResourceCondition{
				{
					ResourceType: shared.ResourceCardDiscard,
					Amount:       1,
					Target:       "self-player",
					Optional:     true,
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("some-card-1")

	effect := player.CardEffect{
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
			Inputs: []shared.ResourceCondition{
				{
					ResourceType: shared.ResourceCardDiscard,
					Amount:       1,
					Target:       "self-player",
					Optional:     true,
				},
			},
			Outputs: []shared.ResourceCondition{
				{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	// Give player a card to discard
	owner.Hand().AddCard("card-to-discard")
	owner.Hand().AddCard("card-to-keep")

	handBefore := len(owner.Hand().Cards())
	testutil.AssertEqual(t, 2, handBefore, "Should have 2 cards before discard")

	// Set up pending discard selection
	owner.Selection().SetPendingCardDiscardSelection(&player.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("card-to-keep")

	// Set up pending discard with min=0 (optional)
	owner.Selection().SetPendingCardDiscardSelection(&player.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("real-card")

	owner.Selection().SetPendingCardDiscardSelection(&player.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
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
	owner.SetCorporationID("corp-tharsis-republic")
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
	owner.SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	owner.Hand().AddCard("card-1")
	owner.Hand().AddCard("card-2")

	owner.Selection().SetPendingCardDiscardSelection(&player.PendingCardDiscardSelection{
		MinCards:     0,
		MaxCards:     1,
		Source:       "Mars University",
		SourceCardID: "card-mars-university",
		PendingOutputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCardDraw, Amount: 1, Target: "self-player"},
		},
	})

	// Try to discard 2 cards when max is 1
	confirmDiscard := confirmAction.NewConfirmCardDiscardAction(repo, cardRegistry, logger)
	err := confirmDiscard.Execute(ctx, testGame.ID(), owner.ID(), []string{"card-1", "card-2"})
	testutil.AssertError(t, err, "Should reject discarding too many cards")
}
