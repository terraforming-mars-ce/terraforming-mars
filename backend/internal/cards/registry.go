package cards

import (
	"fmt"

	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
)

// CardRegistry provides lookup functionality for card data
type CardRegistry interface {
	// GetByID retrieves a card by its ID
	GetByID(cardID string) (*gamecards.Card, error)

	// GetAll returns all cards in the registry
	GetAll() []gamecards.Card
}

// InMemoryCardRegistry implements CardRegistry with an in-memory map
type InMemoryCardRegistry struct {
	cards map[string]gamecards.Card
}

// NewInMemoryCardRegistry creates a new card registry from a slice of cards
func NewInMemoryCardRegistry(cardList []gamecards.Card) *InMemoryCardRegistry {
	cardMap := make(map[string]gamecards.Card, len(cardList))
	for _, card := range cardList {
		cardMap[card.ID] = card
	}

	return &InMemoryCardRegistry{
		cards: cardMap,
	}
}

// GetByID retrieves a card by its ID, returning a copy to prevent mutation
func (r *InMemoryCardRegistry) GetByID(cardID string) (*gamecards.Card, error) {
	card, exists := r.cards[cardID]
	if !exists {
		return nil, fmt.Errorf("card not found: %s", cardID)
	}

	// Return a deep copy to prevent external mutation
	cardCopy := card.DeepCopy()
	return &cardCopy, nil
}

// GetAll returns all cards in the registry
func (r *InMemoryCardRegistry) GetAll() []gamecards.Card {
	cardList := make([]gamecards.Card, 0, len(r.cards))
	for _, card := range r.cards {
		cardList = append(cardList, card.DeepCopy())
	}
	return cardList
}

// GetCardIDsByPacks filters cards by pack and separates them by type.
// Returns project card IDs, corporation IDs, and prelude IDs.
func GetCardIDsByPacks(registry CardRegistry, packs []string) (projectCards, corps, preludes []string) {
	allCards := registry.GetAll()

	packMap := make(map[string]bool, len(packs))
	for _, pack := range packs {
		packMap[pack] = true
	}

	for _, card := range allCards {
		if !packMap[card.Pack] {
			continue
		}

		switch card.Type {
		case gamecards.CardTypeCorporation:
			corps = append(corps, card.ID)
		case gamecards.CardTypePrelude:
			preludes = append(preludes, card.ID)
		default:
			projectCards = append(projectCards, card.ID)
		}
	}

	return projectCards, corps, preludes
}

type VPCardLookupAdapter struct {
	registry CardRegistry
}

func NewVPCardLookupAdapter(registry CardRegistry) *VPCardLookupAdapter {
	return &VPCardLookupAdapter{registry: registry}
}

func (a *VPCardLookupAdapter) LookupVPCard(cardID string) (*game.VPCardInfo, error) {
	card, err := a.registry.GetByID(cardID)
	if err != nil {
		return nil, err
	}

	vpConditions := make([]shared.VPCondition, len(card.VPConditions))
	for i, vc := range card.VPConditions {
		vpConditions[i] = convertVPCondition(vc)
	}

	tags := make([]shared.CardTag, len(card.Tags))
	copy(tags, card.Tags)

	return &game.VPCardInfo{
		CardID:       card.ID,
		CardName:     card.Name,
		CardType:     string(card.Type),
		Description:  card.Description,
		VPConditions: vpConditions,
		Tags:         tags,
	}, nil
}

func convertVPCondition(vc gamecards.VictoryPointCondition) shared.VPCondition {
	cond := shared.VPCondition{
		Amount:     vc.Amount,
		Condition:  string(vc.Condition),
		MaxTrigger: vc.MaxTrigger,
	}

	if vc.Per != nil {
		perCond := &shared.VPPerCondition{
			ResourceType: shared.ResourceType(vc.Per.Type),
			Amount:       vc.Per.Amount,
			Tag:          vc.Per.Tag,
		}
		if vc.Per.Target != nil {
			target := string(*vc.Per.Target)
			perCond.Target = &target
		}
		cond.Per = perCond
	}

	return cond
}
