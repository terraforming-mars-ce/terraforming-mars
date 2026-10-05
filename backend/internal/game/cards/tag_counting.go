package cards

import (
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// TagCountContext permits wild tags only for the player performing an action.
// A zero context counts actual tags, as required for triggers and scoring.
type TagCountContext struct {
	ActorID           string
	ExcludeWildCardID string
}

// TagCounts keeps actual tags and wild tags separate until a query is resolved.
type TagCounts map[shared.CardTag]int

func tagsFromCards(cards []*Card) TagCounts {
	counts := TagCounts{}
	for _, card := range cards {
		for _, tag := range card.Tags {
			if card.Type != CardTypeEvent || tag == shared.TagEvent {
				counts[tag]++
			}
		}
	}
	return counts
}

// PlayerTagCounts collects tags on face-up cards, the corporation, and bonus tags.
func PlayerTagCounts(p *player.Player, registry CardRegistryInterface) TagCounts {
	counts := TagCounts{}
	if p == nil {
		return counts
	}
	if registry != nil {
		ids := append([]string(nil), p.PlayedCards().Cards()...)
		if id := p.CorporationID(); id != "" && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
		for _, id := range ids {
			if card, err := registry.GetByID(id); err == nil {
				for tag, n := range tagsFromCards([]*Card{card}) {
					counts[tag] += n
				}
			}
		}
	}
	for tag, n := range p.BonusTags() {
		counts[tag] += n
	}
	return counts
}

func substitutableTag(tag shared.CardTag) bool {
	switch tag {
	case shared.TagSpace, shared.TagEarth, shared.TagScience, shared.TagPower, shared.TagBuilding,
		shared.TagMicrobe, shared.TagAnimal, shared.TagPlant, shared.TagCity, shared.TagVenus, shared.TagJovian, shared.TagWildlife:
		return true
	}
	return false
}

func eligibleWildTags(p *player.Player, registry CardRegistryInterface, context TagCountContext, counts TagCounts) int {
	if p == nil || context.ActorID == "" || context.ActorID != p.ID() {
		return 0
	}
	wild := counts[shared.TagWild]
	if registry != nil && context.ExcludeWildCardID != "" && (slices.Contains(p.PlayedCards().Cards(), context.ExcludeWildCardID) || p.CorporationID() == context.ExcludeWildCardID) {
		if card, err := registry.GetByID(context.ExcludeWildCardID); err == nil && card.Type != CardTypeEvent {
			for _, tag := range card.Tags {
				if tag == shared.TagWild {
					wild--
				}
			}
		}
	}
	return max(0, wild)
}

// CountPlayerTags counts a union of tag types, using each eligible wild at most once.
func CountPlayerTags(p *player.Player, registry CardRegistryInterface, tags []shared.CardTag, context TagCountContext) int {
	counts := PlayerTagCounts(p, registry)
	total := 0
	seen := map[shared.CardTag]bool{}
	useWild := false
	for _, tag := range tags {
		if seen[tag] {
			continue
		}
		seen[tag] = true
		total += counts[tag]
		useWild = useWild || substitutableTag(tag)
	}
	if useWild && !seen[shared.TagWild] {
		total += eligibleWildTags(p, registry, context, counts)
	}
	return total
}

// ValidateTagRequirements allocates one shared wild pool across all tag requirements.
func ValidateTagRequirements(requirements []Requirement, counts TagCounts) error {
	type bounds struct {
		min int
		max *int
	}
	byTag := map[shared.CardTag]bounds{}
	for _, req := range requirements {
		if req.Type != RequirementTags {
			continue
		}
		if req.Tag == nil {
			return fmt.Errorf("tag requirement missing tag specification")
		}
		b := byTag[*req.Tag]
		if req.Min != nil {
			b.min = max(b.min, *req.Min)
		}
		if req.Max != nil && (b.max == nil || *req.Max < *b.max) {
			value := *req.Max
			b.max = &value
		}
		byTag[*req.Tag] = b
	}
	missing := 0
	for tag, b := range byTag {
		actual := counts[tag]
		if b.max != nil && (actual > *b.max || b.min > *b.max) {
			return fmt.Errorf("tag requirement not met: max %d %s tags", *b.max, tag)
		}
		deficit := max(0, b.min-actual)
		if deficit > 0 && !substitutableTag(tag) {
			return fmt.Errorf("tag requirement not met: need %d %s tags, have %d", b.min, tag, actual)
		}
		missing += deficit
	}
	if missing > counts[shared.TagWild] {
		return fmt.Errorf("tag requirements need %d additional tags, but only %d wild tags are available", missing, counts[shared.TagWild])
	}
	return nil
}

// ValidateChoiceTagRequirements uses the same allocation rules for choice requirements.
func ValidateChoiceTagRequirements(requirements []shared.ChoiceRequirement, counts TagCounts) error {
	converted := make([]Requirement, 0, len(requirements))
	for _, req := range requirements {
		if req.Type == "tags" {
			converted = append(converted, Requirement{Type: RequirementTags, Tag: req.Tag, Min: req.Min, Max: req.Max})
		}
	}
	return ValidateTagRequirements(converted, counts)
}
