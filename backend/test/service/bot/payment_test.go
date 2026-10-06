package bot_test

import (
	"testing"

	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"
)

func TestPreferredPayment_SpendsSubstitutesFirstWithoutOverpaying(t *testing.T) {
	steel := shared.PaymentSource{Target: "self-player", Resource: shared.ResourceSteel}
	credit := shared.PaymentSource{Target: "self-player", Resource: shared.ResourceCredit}
	quote := shared.PaymentQuote{
		Costs: map[shared.ResourceType]int{shared.ResourceCredit: 13},
		Options: []shared.PaymentOption{
			{Source: credit, TargetResource: shared.ResourceCredit, ConversionRate: 1, Available: 20},
			{Source: steel, TargetResource: shared.ResourceCredit, ConversionRate: 2, Available: 10},
		},
	}

	payment, err := bot.PreferredPayment(quote, map[string]int{"steel": 10})
	testutil.AssertNoError(t, err, "payment should be built")
	testutil.AssertEqual(t, 2, len(payment.Allocations), "steel then credits")
	testutil.AssertEqual(t, steel, payment.Allocations[0].Source, "steel first")
	testutil.AssertEqual(t, 6, payment.Allocations[0].Amount, "6 steel covers 12 without overpaying")
	testutil.AssertEqual(t, 1, payment.Allocations[1].Amount, "1 credit covers the rest")
}

func TestPreferredPayment_FailsWhenUnaffordable(t *testing.T) {
	credit := shared.PaymentSource{Target: "self-player", Resource: shared.ResourceCredit}
	quote := shared.PaymentQuote{
		Costs:   map[shared.ResourceType]int{shared.ResourceCredit: 13},
		Options: []shared.PaymentOption{{Source: credit, TargetResource: shared.ResourceCredit, ConversionRate: 1, Available: 5}},
	}
	_, err := bot.PreferredPayment(quote, map[string]int{"steel": 3})
	testutil.AssertError(t, err, "not enough to pay")
}
