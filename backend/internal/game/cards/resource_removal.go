package cards

import (
	"fmt"
	"github.com/google/uuid"
	"slices"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ResourceRemovalTargets returns the current removable amount for each eligible player.
func ResourceRemovalTargets(g *game.Game, actor *player.Player, output *shared.BasicResourceCondition, placement *shared.HexPosition, registry CardRegistryInterface) (map[string]int, error) {
	if g == nil || actor == nil || output == nil || output.Target != "any-player" || output.Amount >= 0 || output.TargetRestriction == nil {
		return nil, fmt.Errorf("invalid restricted resource removal")
	}
	restriction := output.TargetRestriction
	if restriction.Adjacent != "" && restriction.Adjacent != "self-card" {
		return nil, fmt.Errorf("unsupported adjacency restriction")
	}
	if len(restriction.Selectors) > 0 && registry == nil {
		return nil, fmt.Errorf("selector restrictions require a card registry")
	}
	targets := make(map[string]int)
	for _, target := range g.GetAllPlayers() {
		if target.HasExited() || IsResourceProtected(actor, target, output.ResourceType, "") {
			continue
		}
		if restriction.Adjacent != "" {
			if placement == nil {
				continue
			}
			adjacent := false
			for _, position := range placement.GetNeighbors() {
				tile, err := g.Board().GetTile(position)
				if err == nil && tile.OccupiedBy != nil && tile.OwnerID != nil && *tile.OwnerID == target.ID() {
					adjacent = true
					break
				}
			}
			if !adjacent {
				continue
			}
		}
		if len(restriction.Selectors) > 0 {
			ids := append([]string(nil), target.PlayedCards().Cards()...)
			if corp := target.CorporationID(); corp != "" && !slices.Contains(ids, corp) {
				ids = append(ids, corp)
			}
			matched := false
			for _, id := range ids {
				card, err := registry.GetByID(id)
				if err != nil {
					return nil, err
				}
				if card.Type != CardTypeEvent && MatchesAnySelector(card, restriction.Selectors) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		amount := min(-output.Amount, target.Resources().Get().GetAmount(output.ResourceType))
		if amount > 0 {
			targets[target.ID()] = amount
		}
	}
	return targets, nil
}

// QueueResourceRemoval opens an optional selection only when resources can be removed.
func QueueResourceRemoval(g *game.Game, p *player.Player, output *shared.BasicResourceCondition, placement *shared.HexPosition, sourceCardID, source string, registry CardRegistryInterface) error {
	targets, err := ResourceRemovalTargets(g, p, output, placement, registry)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	if p.Selection().GetPendingResourceRemovalSelection() != nil {
		return fmt.Errorf("resource removal already pending")
	}
	ids := make([]string, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	p.Selection().SetPendingResourceRemovalSelection(&shared.PendingResourceRemovalSelection{
		ID: uuid.NewString(), Output: shared.CloneCondition(output).(*shared.BasicResourceCondition), Placement: placement,
		EligiblePlayerIDs: ids, MaxAmounts: targets, ResourceType: output.ResourceType, Amount: -output.Amount, Source: source, SourceCardID: sourceCardID,
	})
	return nil
}
