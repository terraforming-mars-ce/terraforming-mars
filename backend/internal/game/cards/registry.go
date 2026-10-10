package cards

import (
	"fmt"

	"openmars/internal/game"
	"openmars/internal/game/shared"
)

// CardRegistry provides lookup functionality for card data
type CardRegistry interface {
	// GetByID retrieves a card by its ID
	GetByID(cardID string) (*Card, error)

	// GetAll returns all cards in the registry
	GetAll() []Card
}

// InMemoryCardRegistry implements CardRegistry with an in-memory map
type InMemoryCardRegistry struct {
	cards map[string]Card
	order []string // preserves load (JSON) order so GetAll is deterministic
}

// NewInMemoryCardRegistry creates a new card registry from a slice of cards
func NewInMemoryCardRegistry(cardList []Card) *InMemoryCardRegistry {
	cardMap := make(map[string]Card, len(cardList))
	order := make([]string, 0, len(cardList))
	for _, card := range cardList {
		if _, exists := cardMap[card.ID]; !exists {
			order = append(order, card.ID)
		}
		cardMap[card.ID] = card
	}

	return &InMemoryCardRegistry{
		cards: cardMap,
		order: order,
	}
}

// GetByID retrieves a card by its ID, returning a copy to prevent mutation
func (r *InMemoryCardRegistry) GetByID(cardID string) (*Card, error) {
	card, exists := r.cards[cardID]
	if !exists {
		return nil, fmt.Errorf("card not found: %s", cardID)
	}

	// Return a deep copy to prevent external mutation
	cardCopy := card.DeepCopy()
	return &cardCopy, nil
}

// GetAll returns all cards in the registry, in their original load (JSON) order
// so that deck construction is deterministic for a given seed.
func (r *InMemoryCardRegistry) GetAll() []Card {
	cardList := make([]Card, 0, len(r.order))
	for _, id := range r.order {
		cardList = append(cardList, r.cards[id].DeepCopy())
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
		case CardTypeCorporation:
			corps = append(corps, card.ID)
		case CardTypePrelude:
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
		Description:  card.Description.PlainText(),
		VPConditions: vpConditions,
		Tags:         tags,
	}, nil
}

func convertVPCondition(vc VictoryPointCondition) shared.VPCondition {
	cond := shared.VPCondition{
		Amount:     vc.Amount,
		Condition:  string(vc.Condition),
		MaxTrigger: vc.MaxTrigger,
		Per:        vc.Per,
	}
	return cond
}
