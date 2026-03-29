package card_effects_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestPredatorsStealAnimalFromOtherPlayer(t *testing.T) {
	testGame, repo, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	targetCardID := "test-birds"
	other.PlayedCards().AddCard(targetCardID, "Birds", "active", []string{"animal"})
	other.Resources().AddToStorage(targetCardID, 3)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "steal-from-any-card"}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Predators steal action should succeed")

	testutil.AssertEqual(t, 2, other.Resources().GetCardStorage(targetCardID), "Target card should have 2 animals after steal")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(predatorsID), "Predators card should have 1 animal after steal")
}

func TestPredatorsStealFromOwnCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	targetCardID := "test-fish"
	p.PlayedCards().AddCard(targetCardID, "Fish", "active", []string{"animal"})
	p.Resources().AddToStorage(targetCardID, 2)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "steal-from-any-card"}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Predators steal from own card should succeed")

	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(targetCardID), "Target card should have 1 animal after steal")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(predatorsID), "Predators card should have 1 animal after steal")
}

func TestPredatorsRejectsWithNoSourceCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "steal-from-any-card"}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Predators without source card should be rejected")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(predatorsID), "Predators card should have 0 animals when no source card specified")
}

func TestPredatorsStealFromCardWithZeroAnimals(t *testing.T) {
	testGame, repo, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	predatorsID := "test-predators"
	p.PlayedCards().AddCard(predatorsID, "Predators", "active", []string{"animal"})
	p.Resources().AddToStorage(predatorsID, 0)

	targetCardID := "test-birds"
	other.PlayedCards().AddCard(targetCardID, "Birds", "active", []string{"animal"})
	other.Resources().AddToStorage(targetCardID, 0)

	predatorsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceAnimal, Amount: 1, Target: "steal-from-any-card"}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        predatorsID,
			CardName:      "Predators",
			BehaviorIndex: 0,
			Behavior:      predatorsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, predatorsID, 0, nil, []string{predatorsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Predators steal from empty card should succeed")

	testutil.AssertEqual(t, 0, other.Resources().GetCardStorage(targetCardID), "Target card should still have 0 animals")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(predatorsID), "Predators card should have 0 animals when source had none")
}

func TestAntsStealMicrobeFromOtherPlayer(t *testing.T) {
	testGame, repo, cardRegistry, playerID, otherPlayerID := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)
	other, _ := testGame.GetPlayer(otherPlayerID)

	antsID := "test-ants"
	p.PlayedCards().AddCard(antsID, "Ants", "active", []string{"microbe"})
	p.Resources().AddToStorage(antsID, 0)

	targetCardID := "test-decomposers"
	other.PlayedCards().AddCard(targetCardID, "Decomposers", "active", []string{"microbe"})
	other.Resources().AddToStorage(targetCardID, 5)

	antsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "steal-from-any-card"}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        antsID,
			CardName:      "Ants",
			BehaviorIndex: 0,
			Behavior:      antsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, antsID, 0, nil, []string{antsID}, nil, &targetCardID, nil, nil, nil)
	testutil.AssertNoError(t, err, "Ants steal action should succeed")

	testutil.AssertEqual(t, 4, other.Resources().GetCardStorage(targetCardID), "Target card should have 4 microbes after steal")
	testutil.AssertEqual(t, 1, p.Resources().GetCardStorage(antsID), "Ants card should have 1 microbe after steal")
}

func TestAntsRejectsWithNoSourceCard(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	p, _ := testGame.GetPlayer(playerID)

	antsID := "test-ants"
	p.PlayedCards().AddCard(antsID, "Ants", "active", []string{"microbe"})
	p.Resources().AddToStorage(antsID, 0)

	antsBehavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Outputs: []shared.BehaviorCondition{
			&shared.CardStorageCondition{ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceMicrobe, Amount: 1, Target: "steal-from-any-card"}},
		},
	}
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        antsID,
			CardName:      "Ants",
			BehaviorIndex: 0,
			Behavior:      antsBehavior,
		},
	})

	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	ctx := context.Background()

	err := useAction.Execute(ctx, testGame.ID(), playerID, antsID, 0, nil, []string{antsID}, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Ants without source card should be rejected")
	testutil.AssertEqual(t, 0, p.Resources().GetCardStorage(antsID), "Ants card should have 0 microbes when no source card specified")
}
