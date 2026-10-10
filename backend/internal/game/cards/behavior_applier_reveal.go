package cards

import (
	"context"
	"fmt"
	"openmars/internal/game"
	"openmars/internal/game/shared"
	"strings"
)

// ValidateRevealOutputs performs all reveal preconditions before inputs are charged.
func ValidateRevealOutputs(outputs []shared.BehaviorCondition, g *game.Game, registry CardRegistryInterface) error {
	count := 0
	for _, output := range outputs {
		reveal, ok := output.(*shared.CardRevealCondition)
		if !ok {
			continue
		}
		count++
		if count > 1 {
			return fmt.Errorf("use a single card-reveal output with the required amount")
		}
		if errors := shared.ValidateResourceCondition(reveal, false); len(errors) > 0 {
			return fmt.Errorf("invalid reveal: %s", strings.Join(errors, "; "))
		}
		if g == nil || registry == nil {
			return fmt.Errorf("reveal requires a game and card registry")
		}
		available := append(g.Deck().ProjectCards(), g.Deck().DiscardPile()...)
		if len(available) < reveal.Amount {
			return fmt.Errorf("not enough cards available to reveal")
		}
		for _, id := range available {
			if _, err := registry.GetByID(id); err != nil {
				return fmt.Errorf("unknown reveal card %s: %w", id, err)
			}
		}
	}
	if count > 0 {
		for _, output := range outputs {
			switch output.GetResourceType() {
			case shared.ResourceCardPeek, shared.ResourceCardTake, shared.ResourceCardBuy, shared.ResourceCopy, shared.ResourceColonyTrackStep:
				return fmt.Errorf("reveal and selection outputs must use separate behaviors")
			}
		}
	}
	return nil
}

func (a *BehaviorApplier) applyCardReveal(ctx context.Context, reveal *shared.CardRevealCondition) ([]shared.CalculatedOutput, error) {
	if a.player == nil || a.sourceCardID == "" {
		return nil, fmt.Errorf("reveal requires a player and source card")
	}
	if a.player.Selection().GetPendingCardReveal() != nil {
		return nil, fmt.Errorf("a reveal is already awaiting acknowledgement")
	}
	ids, err := a.game.Deck().DrawProjectCards(ctx, reveal.Amount)
	if err != nil {
		return nil, err
	}
	pending := &shared.PendingCardReveal{Source: a.source, SourceCardID: a.sourceCardID}
	for _, id := range ids {
		card, err := a.cardRegistry.GetByID(id)
		if err != nil {
			return nil, err
		}
		matched := reveal.OnMatch != nil && MatchesAnySelector(card, reveal.OnMatch.Selectors)
		pending.Cards = append(pending.Cards, shared.RevealedCard{CardID: id, Name: card.Name, Matched: matched})
		if matched {
			rewards, err := a.ApplyOutputsAndGetCalculated(ctx, reveal.OnMatch.Outputs)
			if err != nil {
				return nil, err
			}
			pending.Rewards = append(pending.Rewards, rewards...)
		}
	}
	if err := a.game.Deck().Discard(ctx, ids); err != nil {
		return nil, err
	}
	a.player.Selection().SetPendingCardReveal(pending)
	return pending.Rewards, nil
}
