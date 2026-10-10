package bot

import (
	"fmt"
	"sort"

	"openmars/internal/action"
	"openmars/internal/game"
	"openmars/internal/game/cards"
	"openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// quotePayment builds a legal payment for an intent. Without preferences the server's
// default allocation is used; preferences spend the named substitutes first.
func (ts *ToolServer) quotePayment(g *game.Game, p *player.Player, intent action.PaymentIntent, prefer map[string]int) (shared.Payment, error) {
	quote, err := action.QuoteActionPayment(g, p, ts.registries.Cards, ts.registries.StandardProjects, ts.registries.Milestones, ts.registries.Awards, intent)
	if err != nil {
		return shared.Payment{}, fmt.Errorf("quote payment: %w", err)
	}
	if len(prefer) == 0 {
		return cards.DefaultPayment(quote)
	}
	return PreferredPayment(quote, prefer)
}

// PreferredPayment spends the preferred substitute sources first (never overpaying),
// then covers the rest of each cost from the player's own pool of that resource.
func PreferredPayment(quote shared.PaymentQuote, prefer map[string]int) (shared.Payment, error) {
	costTypes := make([]shared.ResourceType, 0, len(quote.Costs))
	for rt, cost := range quote.Costs {
		if cost > 0 {
			costTypes = append(costTypes, rt)
		}
	}
	sort.Slice(costTypes, func(i, j int) bool { return costTypes[i] < costTypes[j] })

	used := map[shared.PaymentSource]int{}
	allocations := []shared.PaymentAllocation{}
	for _, rt := range costTypes {
		remaining := quote.Costs[rt]
		for _, o := range quote.Options {
			if o.TargetResource != rt || o.Source.Resource == rt || o.ConversionRate <= 0 {
				continue
			}
			wanted := prefer[string(o.Source.Resource)]
			units := min(wanted, o.Available-used[o.Source], remaining/o.ConversionRate)
			if units <= 0 {
				continue
			}
			used[o.Source] += units
			remaining -= units * o.ConversionRate
			allocations = append(allocations, shared.PaymentAllocation{Source: o.Source, TargetResource: rt, Amount: units})
		}
		if remaining <= 0 {
			continue
		}
		native := shared.PaymentSource{Target: "self-player", Resource: rt}
		for _, o := range quote.Options {
			if o.Source != native || o.TargetResource != rt {
				continue
			}
			units := min(remaining, o.Available-used[native])
			if units > 0 {
				used[native] += units
				remaining -= units
				allocations = append(allocations, shared.PaymentAllocation{Source: native, TargetResource: rt, Amount: units})
			}
		}
		if remaining > 0 {
			return shared.Payment{}, fmt.Errorf("cannot cover %d %s with the preferred sources and your own %s", remaining, rt, rt)
		}
	}
	return shared.Payment{Allocations: allocations}, nil
}
