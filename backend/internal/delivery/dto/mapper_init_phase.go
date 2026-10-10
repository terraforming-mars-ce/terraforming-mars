package dto

import (
	"openmars/internal/game"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/player"
	"openmars/internal/game/shared"
)

// buildInitPhaseDto describes the init_apply_corp / init_apply_prelude showcase state,
// or returns nil outside those phases.
func buildInitPhaseDto(g *game.Game, players []*player.Player, cardRegistry gamecards.CardRegistry) *InitPhaseDto {
	phase := g.CurrentPhase()
	if phase != shared.GamePhaseInitApplyCorp && phase != shared.GamePhaseInitApplyPrelude {
		return nil
	}

	turnOrder := g.TurnOrder()
	idx := g.InitPhasePlayerIndex()
	currentInitPlayerID := ""
	if idx < len(turnOrder) {
		currentInitPlayerID = turnOrder[idx]
	}

	activePlayers := 0
	for _, p := range players {
		if !p.HasExited() {
			activePlayers++
		}
	}

	dto := &InitPhaseDto{
		CurrentPlayerID:    currentInitPlayerID,
		CurrentPlayerIndex: idx,
		TotalPlayers:       activePlayers,
		WaitingForConfirm:  g.InitPhaseWaitingForConfirm(),
		ConfirmVersion:     g.InitPhaseConfirmVersion(),
		HasPreludePhase:    g.Settings().HasPrelude(),
		Stage:              initPhaseStage(g, phase, currentInitPlayerID),
		Preludes:           []CardDto{},
	}
	if currentInitPlayerID == "" {
		return dto
	}

	if phase == shared.GamePhaseInitApplyPrelude {
		if choices := g.GetDeferredStartingChoices(currentInitPlayerID); choices != nil {
			dto.Preludes = getPlayedCards(choices.PreludeIDs, cardRegistry)
			dto.PreludesPlayed = choices.PreludesAppliedCount
		}
	}

	dto.HasPendingSelection = g.HasAnyPendingSelection(currentInitPlayerID) ||
		g.GetPendingTileSelectionQueue(currentInitPlayerID) != nil
	dto.PendingSourceCardID = pendingSourceCardID(g, currentInitPlayerID)

	return dto
}

func initPhaseStage(g *game.Game, phase shared.GamePhase, playerID string) InitPhaseStage {
	if g.InitPhaseRoster() {
		return InitPhaseStageRoster
	}
	choices := g.GetDeferredStartingChoices(playerID)
	if choices == nil {
		return InitPhaseStageReveal
	}
	applied := choices.CorpApplied
	if phase == shared.GamePhaseInitApplyPrelude {
		applied = choices.PreludesAppliedCount > 0
	}
	if applied {
		return InitPhaseStageApplied
	}
	return InitPhaseStageReveal
}

func pendingSourceCardID(g *game.Game, playerID string) string {
	if selection := g.GetPendingTileSelection(playerID); selection != nil && selection.SourceCardID != "" {
		return selection.SourceCardID
	}
	if queue := g.GetPendingTileSelectionQueue(playerID); queue != nil && queue.SourceCardID != "" {
		return queue.SourceCardID
	}
	p, err := g.GetPlayer(playerID)
	if err != nil {
		return ""
	}
	if colony := p.Selection().GetPendingColonySelection(); colony != nil {
		return colony.SourceCardID
	}
	return ""
}
