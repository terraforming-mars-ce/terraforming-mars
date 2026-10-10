package payment_test

import (
	"context"
	"testing"

	cardAction "openmars/internal/action/card"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

// TestValueModifier_StoredOnPlayerResources tests that value modifiers are stored correctly
func TestValueModifier_StoredOnPlayerResources(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	// Get player
	players := testGame.GetAllPlayers()
	player := players[0]

	// Initial state: no value modifiers
	titaniumModifier := player.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 0, titaniumModifier, "Initial titanium value modifier should be 0")

	// Add value modifier (simulating playing Phobolog)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Verify modifier was stored
	titaniumModifier = player.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 1, titaniumModifier, "Titanium value modifier should be 1 after adding")

	// Add another modifier (simulating playing Advanced Alloys)
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1)

	// Verify modifiers stack
	titaniumModifier = player.Resources().GetValueModifier(shared.ResourceTitanium)
	testutil.AssertEqual(t, 2, titaniumModifier, "Titanium value modifier should stack to 2")

	_ = ctx // context available for future use
}

// TestValueModifier_PlayCardWithModifiedTitanium tests full play card flow with value modifier
func TestValueModifier_PlayCardWithModifiedTitanium(t *testing.T) {
	// Setup: Create game with player who has value modifier
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID(testutil.CardID("Tharsis Republic"))

	// Start the game
	testutil.StartTestGame(t, testGame)

	asteroidID := testutil.CardID("Asteroid")

	// Give player titanium and add value modifier
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 5,
		shared.ResourceCredit:   50, // Extra credits for buying other cards if needed
	})
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // +1 titanium value

	// Add a space-tagged card to hand that costs 14 (asteroid)
	player.Hand().AddCard(asteroidID)

	// Play card with 4 titanium (4 * 4 MC = 16 MC covers 14 cost)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.
		PaymentSource{Target: "self-player", Resource: "titanium"}, TargetResource: shared.ResourceCredit,
		Amount: 4}},

	// 4 titanium at 4 MC each = 16 MC
	}

	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), asteroidID, payment, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should be able to play 14-cost card with 4 titanium at value 4")

	// Verify titanium was deducted
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Titanium, "Should have 3 titanium remaining (5 - 4 + 2 from Asteroid behavior)")

	// Verify card is no longer in hand
	testutil.AssertFalse(t, player.Hand().HasCard(asteroidID), "Card should be removed from hand")
}

// TestValueModifier_MixedPaymentWithModifier tests mixed payment (credits + modified titanium)
func TestValueModifier_MixedPaymentWithModifier(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID(testutil.CardID("Tharsis Republic"))

	// Start the game
	testutil.StartTestGame(t, testGame)

	asteroidID := testutil.CardID("Asteroid")

	// Give player resources and value modifier
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 2,
		shared.ResourceCredit:   50,
	})
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // +1 titanium value

	// Add a space-tagged card to hand that costs 14
	player.Hand().AddCard(asteroidID)

	// Play card with 2 titanium (8 MC) + 6 credits = 14 MC exactly
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.
		PaymentSource{Target: "self-player", Resource: "credit"}, TargetResource: shared.ResourceCredit,
		Amount: 6}, {Source: shared.PaymentSource{Target: "self-player",
		Resource:               "titanium"}, TargetResource: shared.
		ResourceCredit, Amount: 2}},

	// 2 titanium at 4 MC each = 8 MC
	}

	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), asteroidID, payment, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Should be able to play 14-cost card with 2 titanium (8) + 6 credits")

	// Verify resources were deducted
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Titanium, "Should have 2 titanium remaining (0 + 2 from Asteroid behavior)")
	testutil.AssertEqual(t, 44, resources.Credits, "Should have 44 credits remaining (50 - 6)")
}

// TestValueModifier_InsufficientPaymentRejected tests that insufficient payment with modifier is rejected
func TestValueModifier_InsufficientPaymentRejected(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	// Get player and set corporation
	players := testGame.GetAllPlayers()
	player := players[0]
	player.SetCorporationID(testutil.CardID("Tharsis Republic"))

	// Start the game
	testutil.StartTestGame(t, testGame)

	asteroidID := testutil.CardID("Asteroid")

	// Give player resources and value modifier
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceTitanium: 3,
		shared.ResourceCredit:   0,
	})
	player.Resources().AddValueModifier(shared.ResourceTitanium, 1) // +1 titanium value

	// Add a space-tagged card to hand that costs 14
	player.Hand().AddCard(asteroidID)

	// Try to play card with only 3 titanium (12 MC) - not enough for 14 cost
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := shared.Payment{Allocations: []shared.PaymentAllocation{{Source: shared.
		PaymentSource{Target: "self-player", Resource: "titanium"}, TargetResource: shared.ResourceCredit,
		Amount: 3}},

	// 3 titanium at 4 MC each = 12 MC (not enough for 14)
	}

	err := playCardAction.Execute(ctx, testGame.ID(), player.ID(), asteroidID, payment, nil, nil, nil, nil, nil)
	testutil.AssertError(t, err, "Should reject payment of 12 MC for 14 cost card")

	// Verify titanium was NOT deducted (action failed)
	resources := player.Resources().Get()
	testutil.AssertEqual(t, 3, resources.Titanium, "Titanium should not be deducted on failed payment")
}

func TestPaymentQuote_MetalRatesAndRestrictions(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		tags                    []shared.CardTag
		steel, titanium         int
		allowed                 []shared.ResourceType
		wantSteel, wantTitanium int
	}{
		{name: "no tags"},
		{name: "base metals", tags: []shared.CardTag{shared.TagBuilding, shared.TagSpace}, wantSteel: 2, wantTitanium: 3},
		{name: "stacked modifiers", tags: []shared.CardTag{shared.TagBuilding, shared.TagSpace}, steel: 2, titanium: 2, wantSteel: 4, wantTitanium: 5},
		{name: "explicit action currency", allowed: []shared.ResourceType{shared.ResourceTitanium}, titanium: 1, wantTitanium: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _, registry, id, _ := testutil.SetupTwoPlayerGame(t)
			p, _ := g.GetPlayer(id)
			p.Resources().AddValueModifier(shared.ResourceSteel, tc.steel)
			p.Resources().AddValueModifier(shared.ResourceTitanium, tc.titanium)
			p.Resources().Set(shared.Resources{Steel: 4, Titanium: 4})
			action := shared.ActionCardPlaying
			if tc.allowed != nil {
				action = "card-action"
			}
			quote, err := gamecards.QuotePayment(p, g, registry, gamecards.PaymentContext{Costs: map[shared.ResourceType]int{shared.ResourceCredit: 12}, Card: &gamecards.Card{Tags: tc.tags}, Action: action, PaymentAllowed: tc.allowed})
			testutil.AssertNoError(t, err, "quote")
			rates := map[shared.ResourceType]int{}
			for _, o := range quote.Options {
				rates[o.Source.Resource] = o.ConversionRate
			}
			testutil.AssertEqual(t, tc.wantSteel, rates[shared.ResourceSteel], "steel rate")
			testutil.AssertEqual(t, tc.wantTitanium, rates[shared.ResourceTitanium], "titanium rate")
			for _, o := range quote.Options {
				if o.Source.Resource == shared.ResourceCredit {
					continue
				}
				units := (12 + o.ConversionRate - 1) / o.ConversionRate
				plan, err := gamecards.ValidatePayment(quote, shared.Payment{Allocations: []shared.PaymentAllocation{{Source: o.Source, TargetResource: shared.ResourceCredit, Amount: units}}})
				if units > o.Available {
					testutil.AssertError(t, err, "cannot overdraw")
				} else {
					testutil.AssertNoError(t, err, "covers with indivisible resources")
					testutil.AssertEqual(t, units, plan.Resources[o.Source.Resource], "spend exact units")
				}
			}
		})
	}
}

func TestPaymentValidation_SharedPoolsAndAtomicity(t *testing.T) {
	heat := shared.PaymentSource{Target: "self-player", Resource: shared.ResourceHeat}
	floater := shared.PaymentSource{Target: "self-card", Resource: shared.ResourceFloater, CardID: "host"}
	quote := shared.PaymentQuote{Costs: map[shared.ResourceType]int{shared.ResourceCredit: 4, shared.ResourceHeat: 4}, Options: []shared.PaymentOption{
		{Source: heat, TargetResource: shared.ResourceCredit, ConversionRate: 1, Available: 6},
		{Source: heat, TargetResource: shared.ResourceHeat, ConversionRate: 1, Available: 6},
		{Source: floater, TargetResource: shared.ResourceHeat, ConversionRate: 2, Available: 2},
	}}
	for _, tc := range []struct {
		name        string
		allocations []shared.PaymentAllocation
		ok          bool
	}{
		{"shared pool overdraw", []shared.PaymentAllocation{{Source: heat, TargetResource: shared.ResourceCredit, Amount: 4}, {Source: heat, TargetResource: shared.ResourceHeat, Amount: 4}}, false},
		{"direct sources", []shared.PaymentAllocation{{Source: heat, TargetResource: shared.ResourceCredit, Amount: 4}, {Source: floater, TargetResource: shared.ResourceHeat, Amount: 2}}, true},
		{"no chained conversion", []shared.PaymentAllocation{{Source: floater, TargetResource: shared.ResourceCredit, Amount: 2}, {Source: heat, TargetResource: shared.ResourceHeat, Amount: 4}}, false},
		{"negative count", []shared.PaymentAllocation{{Source: heat, TargetResource: shared.ResourceCredit, Amount: -1}}, false},
		{"duplicate overdraw", []shared.PaymentAllocation{{Source: floater, TargetResource: shared.ResourceHeat, Amount: 2}, {Source: floater, TargetResource: shared.ResourceHeat, Amount: 2}}, false},
		{"forged card", []shared.PaymentAllocation{{Source: shared.PaymentSource{Target: "self-card", Resource: shared.ResourceFloater, CardID: "other"}, TargetResource: shared.ResourceHeat, Amount: 2}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gamecards.ValidatePayment(quote, shared.Payment{Allocations: tc.allocations})
			if tc.ok {
				testutil.AssertNoError(t, err, "valid plan")
			} else {
				testutil.AssertError(t, err, "invalid plan")
			}
		})
	}
	payment, err := gamecards.DefaultPayment(quote)
	testutil.AssertNoError(t, err, "finds feasible shared pool allocation")
	_, err = gamecards.ValidatePayment(quote, payment)
	testutil.AssertNoError(t, err, "default allocation valid")
}

func TestPaymentQuote_SelectorsReservationsAndGrantOwnership(t *testing.T) {
	g, _, registry, id, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := g.GetPlayer(id)
	source := shared.PaymentSource{Target: "self-player", Resource: shared.ResourceHeat}
	p.Resources().Set(shared.Resources{Heat: 6})
	rule := shared.PaymentSubstitute{Source: source, TargetResource: shared.ResourceCredit, ConversionRate: 1, GrantedByCardID: "corp", Selectors: []shared.Selector{{Actions: []string{shared.ActionCardPlaying}, Tags: []shared.CardTag{shared.TagPlant, shared.TagScience}}, {Actions: []string{shared.ActionCardBuying}}}}
	p.Resources().AddPaymentSubstitute(rule)
	rule.ConversionRate = 2
	rule.GrantedByCardID = "project"
	p.Resources().AddPaymentSubstitute(rule)
	for _, tc := range []struct {
		name, action string
		tags         []shared.CardTag
		eligible     bool
	}{
		{"AND tags", shared.ActionCardPlaying, []shared.CardTag{shared.TagPlant}, false},
		{"complete match", shared.ActionCardPlaying, []shared.CardTag{shared.TagPlant, shared.TagScience}, true},
		{"OR selectors", shared.ActionCardBuying, nil, true},
		{"wrong action", "card-action", []shared.CardTag{shared.TagPlant, shared.TagScience}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q, err := gamecards.QuotePayment(p, g, registry, gamecards.PaymentContext{Costs: map[shared.ResourceType]int{shared.ResourceCredit: 8}, Action: tc.action, Card: &gamecards.Card{Tags: tc.tags}, ReservedResources: map[shared.ResourceType]int{shared.ResourceHeat: 2}})
			testutil.AssertNoError(t, err, "quote")
			found := false
			for _, o := range q.Options {
				if o.Source == source {
					found = true
					testutil.AssertEqual(t, 2, o.ConversionRate, "overlapping grants use highest rate")
					testutil.AssertEqual(t, 4, o.Available, "reserve existing input costs")
				}
			}
			testutil.AssertEqual(t, tc.eligible, found, "eligibility")
		})
	}
	rules := p.Resources().PaymentSubstitutes()
	rules[0].Selectors[0].Tags[0] = shared.TagVenus
	testutil.AssertEqual(t, shared.TagPlant, p.Resources().PaymentSubstitutes()[0].Selectors[0].Tags[0], "rules returned independently")
	p.Resources().RemovePaymentSubstitutes("corp")
	testutil.AssertEqual(t, 1, len(p.Resources().PaymentSubstitutes()), "only corporation rules removed")
	testutil.AssertEqual(t, "project", p.Resources().PaymentSubstitutes()[0].GrantedByCardID, "project grant preserved")
}
