package testutil

import (
	colonyAction "openmars/internal/action/colony"
	"openmars/internal/game/shared"
)

// TradePayment pays the base cost of a colony trade with the given resource: 9 credits,
// 3 energy or 3 titanium.
func TradePayment(paymentType colonyAction.TradePaymentType) shared.Payment {
	switch paymentType {
	case colonyAction.TradePaymentCredits:
		return shared.NativePayment(shared.ResourceCredit, 9)
	case colonyAction.TradePaymentTitanium:
		return shared.NativePayment(shared.ResourceTitanium, 3)
	default:
		return shared.NativePayment(shared.ResourceEnergy, 3)
	}
}
