package payment_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// --- Water Import From Europa (012) ---
// "Action: Pay 12 M€ to place an ocean tile. Titanium may be used as if playing a space card."

func TestWaterImportFromEuropa_PayWithCreditsOnly(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 20,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{
				ResourceType:   shared.ResourceCredit,
				Amount:         12,
				Target:         "self-player",
				PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
			},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	payment := &gamecards.CardPayment{Credits: 12}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment)
	testutil.AssertNoError(t, err, "Water Import action should succeed with credits only")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 8, resources.Credits, "Should have 8 credits after paying 12")
}

func TestWaterImportFromEuropa_PayWithTitaniumAndCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   6,
		shared.ResourceTitanium: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{
				ResourceType:   shared.ResourceCredit,
				Amount:         12,
				Target:         "self-player",
				PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
			},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay 6 credits + 2 titanium (value 3 each = 6) = 12 total
	payment := &gamecards.CardPayment{Credits: 6, Titanium: 2}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment)
	testutil.AssertNoError(t, err, "Water Import action should succeed with titanium + credits")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Credits, "Should have 0 credits after paying 6")
	testutil.AssertEqual(t, 1, resources.Titanium, "Should have 1 titanium after spending 2")
}

func TestWaterImportFromEuropa_FailInsufficientPayment(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   3,
		shared.ResourceTitanium: 1,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{
				ResourceType:   shared.ResourceCredit,
				Amount:         12,
				Target:         "self-player",
				PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
			},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay 3 credits + 1 titanium (value 3) = 6, need 12
	payment := &gamecards.CardPayment{Credits: 3, Titanium: 1}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment)
	testutil.AssertError(t, err, "Should fail with insufficient payment")
}

func TestWaterImportFromEuropa_FailSteelNotAllowed(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 6,
		shared.ResourceSteel:  5,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{
				ResourceType:   shared.ResourceCredit,
				Amount:         12,
				Target:         "self-player",
				PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
			},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Try to pay with steel (not allowed)
	payment := &gamecards.CardPayment{Credits: 6, Steel: 3}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment)
	testutil.AssertError(t, err, "Should fail when using steel (not allowed)")
}

func TestWaterImportFromEuropa_NoPaymentFallsBackToCredits(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit: 15,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{
				ResourceType:   shared.ResourceCredit,
				Amount:         12,
				Target:         "self-player",
				PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
			},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// No payment provided — should fall back to credits-only
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should succeed with no payment (falls back to credits)")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Credits, "Should have 3 credits after paying 12")
}

// --- Rotator Impacts (243) ---
// Choice 1: "Spend 6 M€ to add an asteroid resource (titanium may be used)"

func TestRotatorImpacts_Choice1_PayWithTitanium(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "243"
	p.PlayedCards().AddCard(cardID, "Rotator Impacts", "active", []string{"space"})
	p.Resources().AddToStorage(cardID, 0)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   0,
		shared.ResourceTitanium: 2,
	})

	choiceIndex := 0
	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Choices: []shared.Choice{
			{
				Inputs: []shared.ResourceCondition{
					{
						ResourceType:   shared.ResourceCredit,
						Amount:         6,
						Target:         "self-player",
						PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
					},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAsteroid, Amount: 1, Target: "self-card"},
				},
			},
			{
				Inputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceAsteroid, Amount: 1, Target: "self-card"},
				},
				Outputs: []shared.ResourceCondition{
					{ResourceType: shared.ResourceVenus, Amount: 1, Target: "none"},
				},
			},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Rotator Impacts",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay with 2 titanium (3 MC each = 6 MC total)
	payment := &gamecards.CardPayment{Credits: 0, Titanium: 2}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, &choiceIndex, nil, nil, nil, nil, payment)
	testutil.AssertNoError(t, err, "Rotator Impacts choice 1 should succeed with titanium")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Titanium, "Should have 0 titanium after spending 2")
}

func TestWaterImportFromEuropa_TitaniumWithValueModifier(t *testing.T) {
	testGame, repo, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)

	cardID := "012"
	p.PlayedCards().AddCard(cardID, "Water Import From Europa", "active", []string{"jovian", "space"})

	// Add titanium value modifier (+1, so titanium = 4 MC each)
	p.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	p.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredit:   0,
		shared.ResourceTitanium: 3,
	})

	behavior := shared.CardBehavior{
		Triggers: []shared.Trigger{{Type: shared.TriggerTypeManual}},
		Inputs: []shared.ResourceCondition{
			{
				ResourceType:   shared.ResourceCredit,
				Amount:         12,
				Target:         "self-player",
				PaymentAllowed: []shared.ResourceType{shared.ResourceTitanium},
			},
		},
		Outputs: []shared.ResourceCondition{
			{ResourceType: shared.ResourceOceanPlacement, Amount: 1, Target: "none"},
		},
	}
	p.Actions().SetActions([]player.CardAction{
		{
			CardID:        cardID,
			CardName:      "Water Import From Europa",
			BehaviorIndex: 0,
			Behavior:      behavior,
		},
	})

	// Pay with 3 titanium (4 MC each with modifier = 12 MC total)
	payment := &gamecards.CardPayment{Credits: 0, Titanium: 3}
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, cardID, 0, nil, nil, nil, nil, nil, payment)
	testutil.AssertNoError(t, err, "Should succeed with titanium value modifier")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 0, resources.Titanium, "Should have 0 titanium after spending 3")
}
