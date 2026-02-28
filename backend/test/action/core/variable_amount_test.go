package core_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Insulation (152) ---
// "Decrease your heat production any number of steps and increase your M€ production the same number of steps."

func TestInsulation_DecreaseHeatProductionIncreaseCreditProduction(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	// Give player resources and production
	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceHeatProduction: 5,
	})

	// Add Insulation to hand
	p.Hand().AddCard(testutil.CardID("Insulation"))

	// Play Insulation with selectedAmount=3 (decrease 3 heat production, increase 3 credit production)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 2}
	selectedAmount := 3
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Insulation"), payment, nil, nil, nil, &selectedAmount)
	testutil.AssertNoError(t, err, "Insulation should play successfully")

	// Verify production changes
	production := p.Resources().Production()
	testutil.AssertEqual(t, 2, production.Heat, "Heat production should decrease from 5 to 2")
	testutil.AssertEqual(t, 3, production.Credits, "Credit production should increase by 3")
}

func TestInsulation_SelectAmountZero(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceHeatProduction: 5,
	})

	p.Hand().AddCard(testutil.CardID("Insulation"))

	// Play Insulation with selectedAmount=0 (no change)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 2}
	selectedAmount := 0
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Insulation"), payment, nil, nil, nil, &selectedAmount)
	testutil.AssertNoError(t, err, "Insulation with 0 amount should play successfully")

	// Verify no production changes
	production := p.Resources().Production()
	testutil.AssertEqual(t, 5, production.Heat, "Heat production should remain 5")
	testutil.AssertEqual(t, 0, production.Credits, "Credit production should remain 0")
}

func TestInsulation_SelectMaxAmount(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 100,
	})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceHeatProduction: 3,
	})

	p.Hand().AddCard(testutil.CardID("Insulation"))

	// Play Insulation with selectedAmount=3 (all heat production)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 2}
	selectedAmount := 3
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, testutil.CardID("Insulation"), payment, nil, nil, nil, &selectedAmount)
	testutil.AssertNoError(t, err, "Insulation with max amount should play successfully")

	production := p.Resources().Production()
	testutil.AssertEqual(t, 0, production.Heat, "Heat production should be 0")
	testutil.AssertEqual(t, 3, production.Credits, "Credit production should increase by 3")
}

// --- Power Infrastructure (194) ---
// "Action: Spend any number of energy to gain that amount of M€."

func TestPowerInfrastructure_SpendEnergyGainCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-power-infrastructure"
	p.PlayedCards().AddCard(cardID, "Power Infrastructure", "active", []string{"building", "power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 5,
		shared.ResourceCredit: 10,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player", VariableAmount: true},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player", VariableAmount: true},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Power Infrastructure",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Use action with selectedAmount=3 (spend 3 energy, gain 3 credits)
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 3
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure action should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Energy, "Should have 2 energy after spending 3")
	testutil.AssertEqual(t, 13, resources.Credits, "Should have 13 credits after gaining 3")
}

func TestPowerInfrastructure_SpendZeroEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-power-infrastructure"
	p.PlayedCards().AddCard(cardID, "Power Infrastructure", "active", []string{"building", "power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 5,
		shared.ResourceCredit: 10,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player", VariableAmount: true},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player", VariableAmount: true},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Power Infrastructure",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Use action with selectedAmount=0 (spend 0, gain 0)
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 0
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure with 0 amount should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 5, resources.Energy, "Energy should remain 5")
	testutil.AssertEqual(t, 10, resources.Credits, "Credits should remain 10")
}

func TestPowerInfrastructure_SpendAllEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-power-infrastructure"
	p.PlayedCards().AddCard(cardID, "Power Infrastructure", "active", []string{"building", "power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 5,
		shared.ResourceCredit: 0,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player", VariableAmount: true},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player", VariableAmount: true},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Power Infrastructure",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Use action with selectedAmount=5 (spend all energy)
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 5
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, &selectedAmount, nil)
	testutil.AssertNoError(t, err, "Power Infrastructure with max amount should succeed")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Energy, "Should have 0 energy after spending all")
	testutil.AssertEqual(t, 5, resources.Credits, "Should have 5 credits after gaining 5")
}

func TestPowerInfrastructure_FailsWhenInsufficientEnergy(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "test-power-infrastructure"
	p.PlayedCards().AddCard(cardID, "Power Infrastructure", "active", []string{"building", "power"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceEnergy: 2,
		shared.ResourceCredit: 10,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceEnergy, Amount: 1, Target: "self-player", VariableAmount: true},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceCredit, Amount: 1, Target: "self-player", VariableAmount: true},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Power Infrastructure",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Try to spend more energy than available
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	selectedAmount := 5
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, &selectedAmount, nil)
	testutil.AssertError(t, err, "Should fail when trying to spend more energy than available")

	// Verify resources unchanged
	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Energy, "Energy should remain 2")
	testutil.AssertEqual(t, 10, resources.Credits, "Credits should remain 10")
}
