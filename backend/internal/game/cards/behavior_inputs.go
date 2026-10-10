package cards

import (
	"fmt"
	"slices"

	"openmars/internal/game/shared"
)

// WithInputCardIDs sets the sources of any-card inputs in input order.
func (a *BehaviorApplier) WithInputCardIDs(ids []string) *BehaviorApplier {
	a.inputCardIDs = ids
	return a
}

// WithReservedInputs includes card payment in input affordability checks.
func (a *BehaviorApplier) WithReservedInputs(resources map[shared.ResourceType]int, storage map[string]int) *BehaviorApplier {
	a.reservedResources, a.reservedStorage = resources, storage
	return a
}

func (a *BehaviorApplier) inputStorageMatches(id string, input shared.BehaviorCondition) bool {
	if a.player == nil || a.cardRegistry == nil || !slices.Contains(ownedStorageCards(a.player), id) {
		return false
	}
	card, err := a.cardRegistry.GetByID(id)
	if err != nil || card.ResourceStorage == nil {
		return false
	}
	if input.GetResourceType() != shared.ResourceCardResource && card.ResourceStorage.Type != input.GetResourceType() {
		return false
	}
	selectors := shared.GetSelectors(input)
	if len(selectors) > 0 && !slices.ContainsFunc(selectors, func(s shared.Selector) bool { return MatchesSelector(card, s) }) {
		return false
	}
	return !IsResourceProtected(a.player, a.player, card.ResourceStorage.Type, id)
}

func (a *BehaviorApplier) resolveInputSources(inputs []shared.BehaviorCondition) (map[int]string, error) {
	sources := make(map[int]string)
	index := 0
	for i, input := range inputs {
		if !IsStorageResourceType(input.GetResourceType()) {
			continue
		}
		switch input.GetTarget() {
		case "self-card":
			if a.sourceCardID == "" {
				return nil, fmt.Errorf("storage input requires a source card")
			}
			sources[i] = a.sourceCardID
		case "any-card":
			if index >= len(a.inputCardIDs) || !a.inputStorageMatches(a.inputCardIDs[index], input) {
				return nil, fmt.Errorf("select an eligible owned card for %s input", input.GetResourceType())
			}
			sources[i] = a.inputCardIDs[index]
			index++
		default:
			return nil, fmt.Errorf("invalid storage input target %s", input.GetTarget())
		}
	}
	if index != len(a.inputCardIDs) {
		return nil, fmt.Errorf("unexpected storage input sources")
	}
	return sources, nil
}

// QuoteInputs validates fixed inputs and reserves their resources before quoting basic costs.
func (a *BehaviorApplier) QuoteInputs(inputs []shared.BehaviorCondition) (shared.PaymentQuote, error) {
	if a.player == nil {
		return shared.PaymentQuote{}, fmt.Errorf("cannot validate inputs without player")
	}
	if a.selectedAmount < 0 {
		return shared.PaymentQuote{}, fmt.Errorf("selected amount must be nonnegative")
	}
	sources, err := a.resolveInputSources(inputs)
	if err != nil {
		return shared.PaymentQuote{}, err
	}
	storageCosts := make(map[string]int)
	resourceCosts := make(map[shared.ResourceType]int)
	basicCosts := map[shared.ResourceType]int{}
	var allowed []shared.ResourceType
	for id, amount := range a.reservedStorage {
		storageCosts[id] = amount
	}
	for rt, amount := range a.reservedResources {
		resourceCosts[rt] = amount
	}
	for i, input := range inputs {
		rt, amount := input.GetResourceType(), input.GetAmount()
		if amount < 0 {
			return shared.PaymentQuote{}, fmt.Errorf("input amount must be nonnegative")
		}
		if shared.IsVariableAmount(input) {
			if amount > 0 && a.selectedAmount > int(^uint(0)>>1)/amount {
				return shared.PaymentQuote{}, fmt.Errorf("selected amount is too large")
			}
			amount *= a.selectedAmount
		}
		if amount == 0 {
			continue
		}
		if IsResourceProtected(a.player, a.player, rt, sources[i]) {
			return shared.PaymentQuote{}, fmt.Errorf("%s are protected", rt)
		}
		if IsStorageResourceType(rt) {
			storageCosts[sources[i]] += amount
			continue
		}
		if shared.IsBasicPaymentResource(rt) {
			basicCosts[rt] += amount
			allowed = append(allowed, shared.GetPaymentAllowed(input)...)
			continue
		}

		resourceCosts[rt] += amount
	}
	for id, amount := range storageCosts {
		if a.player.Resources().GetCardStorage(id) < amount {
			return shared.PaymentQuote{}, fmt.Errorf("insufficient resources on card %s", id)
		}
	}
	for rt, amount := range resourceCosts {
		if err := a.validateInputAmount(rt, amount, a.player.Resources().Get()); err != nil {
			return shared.PaymentQuote{}, err
		}
	}
	quote, err := QuotePayment(a.player, a.game, a.cardRegistry, PaymentContext{Costs: basicCosts, Action: "card-action", PaymentAllowed: allowed, ReservedResources: resourceCosts, ReservedStorage: storageCosts})
	if err != nil {
		return shared.PaymentQuote{}, err
	}
	return quote, nil
}

// ValidateInputs checks all costs together without changing game state.
func (a *BehaviorApplier) ValidateInputs(inputs []shared.BehaviorCondition) error {
	quote, err := a.QuoteInputs(inputs)
	if err != nil {
		return err
	}
	payment := shared.Payment{}
	if a.actionPayment != nil {
		payment = *a.actionPayment
	} else {
		for rt, amount := range quote.Costs {
			payment.Allocations = append(payment.Allocations, shared.NativePayment(rt, amount).Allocations...)
		}
	}
	plan, err := ValidatePayment(quote, payment)
	if err != nil {
		return err
	}
	a.inputPaymentPlan = plan

	return nil
}

// InputOptions describes server-calculated choices for paying a behavior's inputs.
type InputOptions struct {
	StorageSources [][]string
	VariableAmount *VariableInputAmount
}

// VariableInputAmount gives the inclusive range of affordable repetitions.
type VariableInputAmount struct {
	ResourceType shared.ResourceType
	Min          int
	Max          int
}

// InputOptions calculates eligible owned sources and variable amount bounds.
func (a *BehaviorApplier) InputOptions(inputs []shared.BehaviorCondition) InputOptions {
	options := InputOptions{}
	for _, input := range inputs {
		rt, amount := input.GetResourceType(), input.GetAmount()
		if IsStorageResourceType(rt) && input.GetTarget() == "any-card" {
			ids := []string{}
			for _, id := range ownedStorageCards(a.player) {
				if a.inputStorageMatches(id, input) && a.player.Resources().GetCardStorage(id) >= amount {
					ids = append(ids, id)
				}
			}
			options.StorageSources = append(options.StorageSources, ids)
		}
		if !shared.IsVariableAmount(input) || amount <= 0 {
			continue
		}
		available := a.player.Resources().Get().GetAmount(rt)
		if shared.IsProductionResourceType(rt) {
			available = a.player.Resources().Production().GetAmount(rt) - shared.ProductionMinimum(rt)
		}
		if IsStorageResourceType(rt) {
			available = a.player.Resources().GetCardStorage(a.sourceCardID)
		}
		maximum := max(0, available/amount)
		if options.VariableAmount == nil || maximum < options.VariableAmount.Max {
			options.VariableAmount = &VariableInputAmount{ResourceType: rt, Max: maximum}
		}
	}
	return options
}

// InputReservations identifies fixed pools which cannot also pay a card's price.
func (a *BehaviorApplier) InputReservations(inputs []shared.BehaviorCondition) (map[shared.ResourceType]int, map[string]int, error) {
	if err := a.ValidateInputs(inputs); err != nil {
		return nil, nil, err
	}
	resources := a.inputPaymentPlan.Resources
	storage := a.inputPaymentPlan.Storage
	sources, err := a.resolveInputSources(inputs)
	if err != nil {
		return nil, nil, err
	}
	for i, input := range inputs {
		if !IsStorageResourceType(input.GetResourceType()) {
			continue
		}
		amount := input.GetAmount()
		if shared.IsVariableAmount(input) {
			amount *= a.selectedAmount
		}
		storage[sources[i]] += amount
	}
	return resources, storage, nil
}
