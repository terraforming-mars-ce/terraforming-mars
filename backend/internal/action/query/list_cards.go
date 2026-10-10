package query

import (
	"context"
	"log/slog"
	"sort"

	gamecards "openmars/internal/game/cards"
)

// ListCardsResult represents the result of listing cards
type ListCardsResult struct {
	Cards      []gamecards.Card
	TotalCount int
	Offset     int
	Limit      int
}

// ListCardsAction handles querying all cards with pagination
type ListCardsAction struct {
	cardRegistry gamecards.CardRegistry
	logger       *slog.Logger
}

// NewListCardsAction creates a new list cards query action
func NewListCardsAction(
	cardRegistry gamecards.CardRegistry,
	logger *slog.Logger,
) *ListCardsAction {
	return &ListCardsAction{
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute retrieves cards with pagination
func (a *ListCardsAction) Execute(ctx context.Context, offset, limit int) (*ListCardsResult, error) {
	log := a.logger.With(
		slog.Int("offset", offset),
		slog.Int("limit", limit),
	)
	log.Debug("Querying cards")

	allCards := a.cardRegistry.GetAll()

	sort.Slice(allCards, func(i, j int) bool {
		return allCards[i].ID < allCards[j].ID
	})

	totalCount := len(allCards)

	start := offset
	if start > totalCount {
		start = totalCount
	}

	end := start + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedCards := allCards[start:end]

	log.Debug("Cards query completed",
		slog.Int("total_count", totalCount),
		slog.Int("returned_count", len(paginatedCards)),
	)

	return &ListCardsResult{
		Cards:      paginatedCards,
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}, nil
}
