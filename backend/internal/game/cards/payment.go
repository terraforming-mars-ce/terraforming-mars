package cards

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// PaymentContext describes the cost being paid, not the card granting an exchange.
type PaymentContext struct {
	Costs             map[shared.ResourceType]int
	Action            string
	Card              *Card
	StandardProject   shared.StandardProject
	PaymentAllowed    []shared.ResourceType
	ReservedResources map[shared.ResourceType]int
	ReservedStorage   map[string]int
}

// PaymentPlan contains fully validated deductions. Validation never mutates state.
type PaymentPlan struct {
	Resources map[shared.ResourceType]int
	Storage   map[string]int
	Payment   shared.Payment
}

func paymentSelectorMatches(context PaymentContext, selector shared.Selector) bool {
	if len(selector.Actions) > 0 && !slices.Contains(selector.Actions, context.Action) {
		return false
	}
	if len(selector.StandardProjects) > 0 && !slices.Contains(selector.StandardProjects, context.StandardProject) {
		return false
	}
	hasCard := HasCardSelectors([]shared.Selector{selector})
	if hasCard && (context.Card == nil || context.Action != shared.ActionCardPlaying || !MatchesSelector(context.Card, selector)) {
		return false
	}
	return hasCard || len(selector.Actions) > 0 || len(selector.StandardProjects) > 0
}

func paymentRuleMatches(context PaymentContext, rule shared.PaymentSubstitute) bool {
	if len(rule.Selectors) == 0 {
		return true
	}
	for _, selector := range rule.Selectors {
		if paymentSelectorMatches(context, selector) {
			return true
		}
	}
	return false
}

// QuotePayment resolves all legal direct exchanges for the given costs.
func QuotePayment(p *player.Player, g *game.Game, registry CardRegistryInterface, context PaymentContext) (shared.PaymentQuote, error) {
	quote := shared.PaymentQuote{Costs: map[shared.ResourceType]int{}, Options: []shared.PaymentOption{}}
	for rt, cost := range context.Costs {
		if !shared.IsBasicPaymentResource(rt) || cost < 0 {
			return quote, fmt.Errorf("invalid payment cost %s", rt)
		}
		quote.Costs[rt] = cost
	}
	rules := p.Resources().PaymentSubstitutes()
	for rt := range quote.Costs {
		rules = append(rules, shared.PaymentSubstitute{Source: shared.PaymentSource{Target: "self-player", Resource: rt}, TargetResource: rt, ConversionRate: 1})
	}
	for _, rt := range []shared.ResourceType{shared.ResourceSteel, shared.ResourceTitanium} {
		allowed := slices.Contains(context.PaymentAllowed, rt)
		if context.Action == shared.ActionCardPlaying && context.Card != nil {
			allowed = allowed || (rt == shared.ResourceSteel && HasTag(context.Card, shared.TagBuilding)) || (rt == shared.ResourceTitanium && HasTag(context.Card, shared.TagSpace))
		}
		if allowed {
			rules = append(rules, shared.PaymentSubstitute{Source: shared.PaymentSource{Target: "self-player", Resource: rt}, TargetResource: shared.ResourceCredit, ConversionRate: p.Resources().ResourceValue(rt)})
		}
	}
	options := map[shared.PaymentSource]map[shared.ResourceType]shared.PaymentOption{}
	resources := p.Resources().Get()
	for _, rule := range rules {
		if quote.Costs[rule.TargetResource] <= 0 || rule.ConversionRate <= 0 || !paymentRuleMatches(context, rule) {
			continue
		}
		var available int
		switch rule.Source.Target {
		case "self-player":
			if rule.Source.CardID != "" || !shared.IsBasicPaymentResource(rule.Source.Resource) {
				return quote, fmt.Errorf("invalid player payment source")
			}
			available = resources.GetAmount(rule.Source.Resource) - context.ReservedResources[rule.Source.Resource]
		case "self-card":
			if registry == nil || rule.Source.CardID == "" {
				return quote, fmt.Errorf("payment source requires card registry")
			}
			card, err := registry.GetByID(rule.Source.CardID)
			if err != nil || card.ResourceStorage == nil || card.ResourceStorage.Type != rule.Source.Resource || !slices.Contains(ownedStorageCards(p), rule.Source.CardID) {
				continue
			}
			available = p.Resources().GetCardStorage(rule.Source.CardID) - context.ReservedStorage[rule.Source.CardID]
		default:
			return quote, fmt.Errorf("invalid payment source target")
		}
		if IsResourceProtected(p, p, rule.Source.Resource, rule.Source.CardID) {
			continue
		}
		available = max(0, available)
		if options[rule.Source] == nil {
			options[rule.Source] = map[shared.ResourceType]shared.PaymentOption{}
		}
		if options[rule.Source][rule.TargetResource].ConversionRate < rule.ConversionRate {
			options[rule.Source][rule.TargetResource] = shared.PaymentOption{Source: rule.Source, TargetResource: rule.TargetResource, ConversionRate: rule.ConversionRate, Available: available}
		}
	}
	for _, destinations := range options {
		for _, option := range destinations {
			quote.Options = append(quote.Options, option)
		}
	}
	sort.Slice(quote.Options, func(i, j int) bool {
		a, b := quote.Options[i], quote.Options[j]
		return fmt.Sprint(a.TargetResource, a.Source.Target, a.Source.Resource, a.Source.CardID) < fmt.Sprint(b.TargetResource, b.Source.Target, b.Source.Resource, b.Source.CardID)
	})
	return quote, nil
}

// ValidatePayment validates and normalizes a selection against a current server quote.
func ValidatePayment(quote shared.PaymentQuote, payment shared.Payment) (PaymentPlan, error) {
	plan := PaymentPlan{Resources: map[shared.ResourceType]int{}, Storage: map[string]int{}, Payment: shared.Payment{Allocations: []shared.PaymentAllocation{}}}
	type key struct {
		source      shared.PaymentSource
		destination shared.ResourceType
	}
	options := map[key]shared.PaymentOption{}
	for _, o := range quote.Options {
		options[key{o.Source, o.TargetResource}] = o
	}
	quantities := map[key]int{}
	for _, a := range payment.Allocations {
		if a.Amount < 0 {
			return plan, fmt.Errorf("payment amounts must be nonnegative")
		}
		k := key{a.Source, a.TargetResource}
		o, ok := options[k]
		if !ok {
			return plan, fmt.Errorf("ineligible payment source for %s", a.TargetResource)
		}
		if a.Amount == 0 {
			continue
		}
		if quantities[k] > math.MaxInt-a.Amount {
			return plan, fmt.Errorf("payment amount is too large")
		}
		if !(o.Source.Target == "self-player" && o.Source.Resource == o.TargetResource) && (a.Amount > o.Available || quantities[k] > o.Available-a.Amount) {
			return plan, fmt.Errorf("insufficient payment resources")
		}
		if quantities[k]+a.Amount > math.MaxInt/o.ConversionRate {
			return plan, fmt.Errorf("payment amount is too large")
		}
		quantities[k] += a.Amount
	}
	totals := map[shared.ResourceType]int{}
	// Non-native selections are exact; native currency fills only the remaining cost.
	for pass := 0; pass < 2; pass++ {
		for _, o := range quote.Options {
			native := o.Source.Target == "self-player" && o.Source.Resource == o.TargetResource
			if native != (pass == 1) {
				continue
			}
			amount := quantities[key{o.Source, o.TargetResource}]
			if native {
				amount = min(amount, max(0, quote.Costs[o.TargetResource]-totals[o.TargetResource]))
			}
			if amount == 0 {
				continue
			}
			if totals[o.TargetResource] > math.MaxInt-amount*o.ConversionRate {
				return plan, fmt.Errorf("payment value is too large")
			}
			used := plan.Resources[o.Source.Resource]
			if o.Source.Target == "self-card" {
				used = plan.Storage[o.Source.CardID]
			}
			if amount > o.Available || used > o.Available-amount {
				return plan, fmt.Errorf("payment exceeds available source pool")
			}
			totals[o.TargetResource] += amount * o.ConversionRate
			if o.Source.Target == "self-card" {
				plan.Storage[o.Source.CardID] += amount
			} else {
				plan.Resources[o.Source.Resource] += amount
			}
			plan.Payment.Allocations = append(plan.Payment.Allocations, shared.PaymentAllocation{Source: o.Source, TargetResource: o.TargetResource, Amount: amount})
		}
	}
	for _, o := range quote.Options {
		used := plan.Resources[o.Source.Resource]
		if o.Source.Target == "self-card" {
			used = plan.Storage[o.Source.CardID]
		}
		if used > o.Available {
			return plan, fmt.Errorf("payment uses the same resource pool more than once")
		}
	}
	for rt, cost := range quote.Costs {
		if totals[rt] < cost {
			return plan, fmt.Errorf("insufficient %s payment: need %d, selected %d", rt, cost, totals[rt])
		}
	}
	return plan, nil
}

// ApplyPayment spends a validated payment plan exactly once.
func ApplyPayment(p *player.Player, plan PaymentPlan) {
	deltas := map[shared.ResourceType]int{}
	for rt, amount := range plan.Resources {
		deltas[rt] = -amount
	}
	p.Resources().Add(deltas)
	for id, amount := range plan.Storage {
		p.Resources().AddToStorage(id, -amount)
	}
}

// DefaultPayment finds a legal allocation without double-counting shared pools.
// Native sources are preferred; callers can offer the quote when a choice is available.
func DefaultPayment(quote shared.PaymentQuote) (shared.Payment, error) {
	costs := make([]shared.ResourceType, 0, len(quote.Costs))
	for rt, cost := range quote.Costs {
		if cost > 0 {
			costs = append(costs, rt)
		}
	}
	sort.Slice(costs, func(i, j int) bool { return costs[i] < costs[j] })
	available := map[shared.PaymentSource]int{}
	for _, o := range quote.Options {
		available[o.Source] = o.Available
	}
	selected := []shared.PaymentAllocation{}
	var payCost func(int) bool
	payCost = func(index int) bool {
		if index == len(costs) {
			return true
		}
		rt := costs[index]
		options := []shared.PaymentOption{}
		for _, o := range quote.Options {
			if o.TargetResource == rt {
				options = append(options, o)
			}
		}
		sort.SliceStable(options, func(i, j int) bool {
			return options[i].Source.Target == "self-player" && options[i].Source.Resource == rt && !(options[j].Source.Target == "self-player" && options[j].Source.Resource == rt)
		})
		var allocate func(int, int) bool
		allocate = func(i, remaining int) bool {
			if remaining <= 0 {
				return payCost(index + 1)
			}
			if i == len(options) {
				return false
			}
			capacity := 0
			for _, o := range options[i:] {
				units := min(available[o.Source], remaining/o.ConversionRate+1)
				capacity += units * o.ConversionRate
				if capacity >= remaining {
					break
				}
			}
			if capacity < remaining {
				return false
			}
			o := options[i]
			limit := min(available[o.Source], remaining/o.ConversionRate)
			if remaining%o.ConversionRate != 0 {
				limit = min(available[o.Source], limit+1)
			}
			for n := limit; n >= 0; n-- {
				length := len(selected)
				available[o.Source] -= n
				if n > 0 {
					selected = append(selected, shared.PaymentAllocation{Source: o.Source, TargetResource: rt, Amount: n})
				}
				if allocate(i+1, remaining-n*o.ConversionRate) {
					return true
				}
				selected = selected[:length]
				available[o.Source] += n
			}
			return false
		}
		return allocate(0, quote.Costs[rt])
	}
	if !payCost(0) {
		return shared.Payment{}, fmt.Errorf("cannot afford payment")
	}
	return shared.Payment{Allocations: selected}, nil
}

// QuoteStartingPayment projects only corporation starting resources, before preludes or actions.
func QuoteStartingPayment(registry CardRegistry, corporationID string, count int, free bool) (shared.PaymentQuote, error) {
	corp, err := registry.GetByID(corporationID)
	if err != nil {
		return shared.PaymentQuote{}, err
	}
	if count < 0 {
		return shared.PaymentQuote{}, fmt.Errorf("invalid card count")
	}
	cost := max(0, 3-CalculateActionDiscountsFromCard(corp, shared.ActionCardBuying)) * count
	if free {
		cost = 0
	}
	quote := shared.PaymentQuote{Costs: map[shared.ResourceType]int{shared.ResourceCredit: cost}, Options: []shared.PaymentOption{}}
	resources := map[shared.ResourceType]int{}
	rules := []shared.PaymentSubstitute{{Source: shared.PaymentSource{Target: "self-player", Resource: shared.ResourceCredit}, TargetResource: shared.ResourceCredit, ConversionRate: 1}}
	for _, b := range corp.Behaviors {
		starts := false
		for _, t := range b.Triggers {
			if t.Type == string(ResourceTriggerAutoCorporationStart) {
				starts = true
			}
		}
		for _, o := range b.Outputs {
			if starts && o.GetTarget() == "self-player" && shared.IsBasicPaymentResource(o.GetResourceType()) {
				resources[o.GetResourceType()] += o.GetAmount()
			}
			if c, ok := o.(*shared.PaymentSubstituteCondition); ok && (starts || HasAutoTrigger(b)) {
				rules = append(rules, shared.PaymentSubstitute{Source: c.Source, TargetResource: c.TargetResource, ConversionRate: c.Amount, Selectors: c.Selectors})
			}
		}
	}
	context := PaymentContext{Action: shared.ActionCardBuying}
	for _, r := range rules {
		if cost > 0 && r.Source.Target == "self-player" && r.TargetResource == shared.ResourceCredit && paymentRuleMatches(context, r) {
			quote.Options = append(quote.Options, shared.PaymentOption{Source: r.Source, TargetResource: r.TargetResource, ConversionRate: r.ConversionRate, Available: max(0, resources[r.Source.Resource])})
		}
	}
	return quote, nil
}

// PaymentCapacity returns the value of distinct source pools for a single destination.
func PaymentCapacity(p *player.Player, g *game.Game, registry CardRegistryInterface, action string, destination shared.ResourceType) int {
	quote, err := QuotePayment(p, g, registry, PaymentContext{Costs: map[shared.ResourceType]int{destination: 1}, Action: action})
	if err != nil {
		return 0
	}
	total := 0
	for _, o := range quote.Options {
		if o.Available > (math.MaxInt-total)/o.ConversionRate {
			return math.MaxInt
		}
		total += o.Available * o.ConversionRate
	}
	return total
}
