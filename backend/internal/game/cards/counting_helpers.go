package cards

import (
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CountPlayerTiles counts tiles owned by a player on the board.
// If tileType is nil, counts all tiles owned by the player.
// If tileType is specified, counts only tiles of that type.
func CountPlayerTiles(playerID string, b *board.Board, tileType *shared.ResourceType) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OwnerID == nil || *tile.OwnerID != playerID {
			continue
		}
		if tile.OccupiedBy == nil {
			continue
		}
		if tileType != nil && tile.OccupiedBy.Type != *tileType {
			continue
		}
		count++
	}
	return count
}

// CountPlayerTagsByType counts tags of a specific type across all played cards and corporation for a player.
// Wild tags count toward any tag type. Event cards are excluded unless counting TagEvent.
// Optional extraTags allows counting tags from a card not yet in played cards (e.g., the card being played).
func CountPlayerTagsByType(p *player.Player, cardRegistry CardRegistryInterface, tagType shared.CardTag, extraTags ...[]shared.CardTag) int {
	count := 0

	for _, cardID := range p.PlayedCards().Cards() {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}
		if card.Type == CardTypeEvent && tagType != shared.TagEvent {
			continue
		}
		count += countTagsInList(card.Tags, tagType)
	}

	if corpID := p.CorporationID(); corpID != "" {
		if corp, err := cardRegistry.GetByID(corpID); err == nil {
			count += countTagsInList(corp.Tags, tagType)
		}
	}

	for _, tags := range extraTags {
		count += countTagsInList(tags, tagType)
	}

	// Include bonus tags from effects like Home Schooled
	count += p.BonusTagCount(tagType)

	return count
}

// countTagsInList counts occurrences of a tag in a slice, including wild tags.
func countTagsInList(tags []shared.CardTag, target shared.CardTag) int {
	count := 0
	for _, tag := range tags {
		if tag == target || tag == shared.TagWild {
			count++
		}
	}
	return count
}

// HasTag checks if a card has a specific tag.
func HasTag(card *Card, tag shared.CardTag) bool {
	for _, cardTag := range card.Tags {
		if cardTag == tag {
			return true
		}
	}
	return false
}

// CountAllTilesOfType counts all tiles of a specific type on the board, regardless of owner.
func CountAllTilesOfType(b *board.Board, tileType shared.ResourceType) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OccupiedBy != nil && tile.OccupiedBy.Type == tileType {
			count++
		}
	}
	return count
}

// CountTilesOfTypeByLocation counts tiles of a specific type, optionally filtered by location.
// If location is "mars", only counts tiles with TileLocationMars.
// If location is nil or "anywhere", counts all tiles of that type.
func CountTilesOfTypeByLocation(b *board.Board, tileType shared.ResourceType, location *string) int {
	count := 0
	tiles := b.Tiles()
	for _, tile := range tiles {
		if tile.OccupiedBy == nil || tile.OccupiedBy.Type != tileType {
			continue
		}
		if location != nil && *location == "mars" && tile.Location != board.TileLocationMars {
			continue
		}
		count++
	}
	return count
}

// CountAllPlayersTagsByType sums tag counts of a specific type across all players.
func CountAllPlayersTagsByType(players []*player.Player, cardRegistry CardRegistryInterface, tagType shared.CardTag) int {
	count := 0
	for _, p := range players {
		count += CountPlayerTagsByType(p, cardRegistry, tagType)
	}
	return count
}

// CountOtherPlayersTagsByType sums tag counts of a specific type across all players except the given one.
func CountOtherPlayersTagsByType(players []*player.Player, excludePlayerID string, cardRegistry CardRegistryInterface, tagType shared.CardTag) int {
	count := 0
	for _, p := range players {
		if p.ID() == excludePlayerID {
			continue
		}
		count += CountPlayerTagsByType(p, cardRegistry, tagType)
	}
	return count
}
