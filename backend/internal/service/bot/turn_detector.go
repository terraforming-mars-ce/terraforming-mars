package bot

import (
	"terraforming-mars-backend/internal/delivery/dto"
)

// IsMyTurn checks if it's currently the given player's turn to act.
func IsMyTurn(game *dto.GameDto, myPlayerID string) bool {
	if game == nil || myPlayerID == "" {
		return false
	}

	p := &game.CurrentPlayer

	// Action phase and it's our turn
	if game.CurrentPhase == dto.GamePhaseAction && game.CurrentTurn != nil && *game.CurrentTurn == myPlayerID {
		return true
	}

	// Player status checks
	if p.Status == dto.PlayerStatusActive {
		return true
	}
	if p.Status == dto.PlayerStatusSelectingStartingCards {
		return true
	}
	if p.Status == dto.PlayerStatusSelectingProductionCards {
		return true
	}

	// Pending selections
	if p.PendingTileSelection != nil {
		return true
	}
	if p.PendingCardSelection != nil {
		return true
	}
	if p.PendingCardDrawSelection != nil {
		return true
	}
	if p.PendingCardDiscardSelection != nil {
		return true
	}
	if p.PendingBehaviorChoiceSelection != nil {
		return true
	}

	// Forced first action
	if p.ForcedFirstAction != nil && !p.ForcedFirstAction.Completed {
		return true
	}

	// Starting selection phase — only if player still has choices to make
	if game.CurrentPhase == dto.GamePhaseStartingSelection {
		if p.SelectCorporationPhase != nil || p.SelectStartingCardsPhase != nil || p.SelectPreludeCardsPhase != nil {
			return true
		}
	}

	// Production phase with incomplete selection
	if game.CurrentPhase == dto.GamePhaseProductionAndCardDraw &&
		p.ProductionPhase != nil &&
		!p.ProductionPhase.SelectionComplete {
		return true
	}

	return false
}

// GetPendingActionType returns the type of pending action requiring resolution.
func GetPendingActionType(game *dto.GameDto) string {
	if game == nil {
		return ""
	}
	p := &game.CurrentPlayer

	if p.PendingTileSelection != nil {
		return "tile-selection"
	}
	if p.PendingCardSelection != nil {
		return "card-selection"
	}
	if p.PendingCardDrawSelection != nil {
		return "card-draw-selection"
	}
	if p.PendingCardDiscardSelection != nil {
		return "card-discard-selection"
	}
	if p.PendingBehaviorChoiceSelection != nil {
		return "behavior-choice-selection"
	}
	if p.ForcedFirstAction != nil && !p.ForcedFirstAction.Completed {
		return "forced-first-action"
	}
	return ""
}
