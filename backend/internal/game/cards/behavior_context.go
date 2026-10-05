package cards

import (
	"fmt"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// WithTriggeringCard retains the event card independently of the effect source.
func (a *BehaviorApplier) WithTriggeringCard(cardID, playerID string) *BehaviorApplier {
	a.triggeringCardID = cardID
	a.triggeringPlayerID = playerID
	return a
}

// ResolveTriggeringCard validates the fixed destination of a triggered storage output.
func ResolveTriggeringCard(g *game.Game, registry CardRegistryInterface, cardID, playerID string, output shared.BehaviorCondition) (*player.Player, *Card, error) {
	if g == nil || registry == nil || cardID == "" || playerID == "" {
		return nil, nil, fmt.Errorf("missing triggering card context")
	}
	owner, err := g.GetPlayer(playerID)
	if err != nil {
		return nil, nil, err
	}
	if !owner.PlayedCards().Contains(cardID) && owner.CorporationID() != cardID {
		return nil, nil, fmt.Errorf("triggering card is not in play")
	}
	card, err := registry.GetByID(cardID)
	if err != nil {
		return nil, nil, err
	}
	if card.ResourceStorage == nil {
		return nil, nil, fmt.Errorf("%s cannot store resources", card.Name)
	}
	resource := output.GetResourceType()
	if resource != shared.ResourceCardResource && card.ResourceStorage.Type != resource {
		return nil, nil, fmt.Errorf("%s cannot store %s", card.Name, resource)
	}
	if selectors := shared.GetSelectors(output); len(selectors) > 0 && !MatchesAnySelector(card, selectors) {
		return nil, nil, fmt.Errorf("triggering card does not match selectors")
	}
	return owner, card, nil
}

// ValidateStorageDestination checks ownership, storage type and selectors for a resource gain.
// enteringCardID allows the card currently being played to receive its own resources.
func ValidateStorageDestination(p *player.Player, registry CardRegistryInterface, output shared.BehaviorCondition, id, enteringCardID string) error {
	if p == nil || registry == nil {
		return fmt.Errorf("resource storage requires a player and card registry")
	}
	if id == "" || (!p.PlayedCards().Contains(id) && p.CorporationID() != id && id != enteringCardID) {
		return fmt.Errorf("target card is not yours or is not in play")
	}
	card, err := registry.GetByID(id)
	if err != nil {
		return err
	}
	if card.ResourceStorage == nil || (output.GetResourceType() != shared.ResourceCardResource && card.ResourceStorage.Type != output.GetResourceType()) {
		return fmt.Errorf("card cannot store that resource")
	}
	if selectors := shared.GetSelectors(output); len(selectors) > 0 && !MatchesAnySelector(card, selectors) {
		return fmt.Errorf("target card does not match selectors")
	}
	return nil
}

// StorageDestinations lists the player's eligible cards and corporation for a resource gain.
func StorageDestinations(p *player.Player, registry CardRegistryInterface, output shared.BehaviorCondition, enteringCardID string) []string {
	if p == nil || registry == nil {
		return nil
	}
	ids := append([]string(nil), p.PlayedCards().Cards()...)
	if p.CorporationID() != "" {
		ids = append(ids, p.CorporationID())
	}
	if enteringCardID != "" && !p.PlayedCards().Contains(enteringCardID) && p.CorporationID() != enteringCardID {
		ids = append(ids, enteringCardID)
	}
	var eligible []string
	for _, id := range ids {
		if ValidateStorageDestination(p, registry, output, id, enteringCardID) == nil {
			eligible = append(eligible, id)
		}
	}
	return eligible
}
