package bot

import (
	"fmt"
	"strings"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SummarizeGameState produces a human-readable text summary of the game state.
func SummarizeGameState(game *dto.GameDto, myPlayerID string) string {
	if game == nil {
		return "No game state available."
	}

	var lines []string

	lines = append(lines, formatGameInfo(game, myPlayerID))
	lines = append(lines, formatGlobalParams(&game.GlobalParameters))

	p := &game.CurrentPlayer
	lines = append(lines, formatPendingActions(p))
	lines = append(lines, formatPlayerStatus(p))
	lines = append(lines, formatHand(p.Cards))
	lines = append(lines, formatCardActions(p.Actions))
	lines = append(lines, formatStandardProjects(p.StandardProjects))
	lines = append(lines, formatMilestones(p.Milestones))
	lines = append(lines, formatAwards(p.Awards))
	lines = append(lines, formatOpponents(game.OtherPlayers))
	lines = append(lines, formatBoard(game.Board.Tiles))

	if len(game.FinalScores) > 0 {
		lines = append(lines, formatFinalScores(game))
	}

	var filtered []string
	for _, l := range lines {
		if l != "" {
			filtered = append(filtered, l)
		}
	}
	return strings.Join(filtered, "\n\n")
}

func formatGameInfo(game *dto.GameDto, myPlayerID string) string {
	turnInfo := "N/A"
	if game.CurrentTurn != nil {
		if *game.CurrentTurn == myPlayerID {
			turnInfo = "YOUR TURN"
		} else {
			turnInfo = fmt.Sprintf("Waiting for %s", findPlayerName(game, *game.CurrentTurn))
		}
	}

	return strings.Join([]string{
		"=== GAME INFO ===",
		fmt.Sprintf("Game ID: %s", game.ID),
		fmt.Sprintf("Phase: %s", string(game.CurrentPhase)),
		fmt.Sprintf("Status: %s", string(game.Status)),
		fmt.Sprintf("Generation: %d", game.Generation),
		fmt.Sprintf("Turn: %s", turnInfo),
		fmt.Sprintf("Players: %d", len(game.TurnOrder)),
		fmt.Sprintf("Turn order: %s", formatTurnOrder(game)),
	}, "\n")
}

func formatTurnOrder(game *dto.GameDto) string {
	var names []string
	for _, id := range game.TurnOrder {
		names = append(names, findPlayerName(game, id))
	}
	return strings.Join(names, " -> ")
}

func formatGlobalParams(gp *dto.GlobalParametersDto) string {
	return strings.Join([]string{
		"=== GLOBAL PARAMETERS ===",
		fmt.Sprintf("Temperature: %d°C (target: 8°C)", gp.Temperature),
		fmt.Sprintf("Oxygen: %d%% (target: 14%%)", gp.Oxygen),
		fmt.Sprintf("Oceans: %d/%d", gp.Oceans, gp.MaxOceans),
		fmt.Sprintf("Venus: %d%% (target: 30%%)", gp.Venus),
	}, "\n")
}

func formatPendingActions(player *dto.PlayerDto) string {
	var parts []string

	if player.PendingTileSelection != nil {
		parts = append(parts, formatPendingTile(player.PendingTileSelection))
	}
	if player.PendingCardSelection != nil {
		parts = append(parts, formatPendingCardSelection(player.PendingCardSelection))
	}
	if player.PendingCardDrawSelection != nil {
		parts = append(parts, formatPendingCardDraw(player.PendingCardDrawSelection))
	}
	if player.PendingCardDiscardSelection != nil {
		parts = append(parts, formatPendingCardDiscard(player.PendingCardDiscardSelection))
	}
	if player.PendingBehaviorChoiceSelection != nil {
		parts = append(parts, formatPendingBehaviorChoice(player.PendingBehaviorChoiceSelection))
	}
	if player.ForcedFirstAction != nil && !player.ForcedFirstAction.Completed {
		parts = append(parts, formatForcedAction(player.ForcedFirstAction))
	}
	if player.SelectCorporationPhase != nil || player.SelectStartingCardsPhase != nil || player.SelectPreludeCardsPhase != nil {
		parts = append(parts, formatStartingSelection(player))
	}
	if player.ProductionPhase != nil && !player.ProductionPhase.SelectionComplete {
		parts = append(parts, formatProductionPhase(player))
	}

	if len(parts) == 0 {
		return ""
	}

	return ">>> PENDING ACTIONS (must resolve) <<<\n" + strings.Join(parts, "\n\n")
}

func formatPendingTile(sel *dto.PendingTileSelectionDto) string {
	return strings.Join([]string{
		fmt.Sprintf("TILE PLACEMENT REQUIRED: Place a %s tile", sel.TileType),
		fmt.Sprintf("Source: %s", sel.Source),
		fmt.Sprintf("Available hexes: %s", strings.Join(sel.AvailableHexes, ", ")),
		"Send a tile-selected command with q, r, s coordinates.",
	}, "\n")
}

func formatPendingCardSelection(sel *dto.PendingCardSelectionDto) string {
	var cardLines []string
	for _, c := range sel.AvailableCards {
		cost := sel.CardCosts[c.ID]
		reward := sel.CardRewards[c.ID]
		info := ""
		if cost > 0 {
			info = fmt.Sprintf(" (cost: %dM€)", cost)
		} else if reward > 0 {
			info = fmt.Sprintf(" (reward: %dM€)", reward)
		}
		cardLines = append(cardLines, fmt.Sprintf("  - %s [%s]%s", c.Name, c.ID, info))
	}

	return strings.Join([]string{
		fmt.Sprintf("CARD SELECTION REQUIRED (source: %s)", sel.Source),
		fmt.Sprintf("Select %d-%d cards:", sel.MinCards, sel.MaxCards),
		strings.Join(cardLines, "\n"),
		`Send a select-cards command with cardIds.`,
	}, "\n")
}

func formatPendingCardDraw(sel *dto.PendingCardDrawSelectionDto) string {
	var cardLines []string
	for _, c := range sel.AvailableCards {
		cardLines = append(cardLines, fmt.Sprintf("  - %s [%s]: %s", c.Name, c.ID, c.Description))
	}

	buyCostStr := ""
	if sel.CardBuyCost > 0 {
		buyCostStr = fmt.Sprintf(" (%dM€ each)", sel.CardBuyCost)
	}

	return strings.Join([]string{
		fmt.Sprintf("CARD DRAW SELECTION (source: %s)", sel.Source),
		fmt.Sprintf("Free takes: %d, Max buy: %d%s", sel.FreeTakeCount, sel.MaxBuyCount, buyCostStr),
		strings.Join(cardLines, "\n"),
		`Send a card-draw-confirmed command with cardsToTake and cardsToBuy.`,
	}, "\n")
}

func formatPendingCardDiscard(sel *dto.PendingCardDiscardSelectionDto) string {
	return strings.Join([]string{
		fmt.Sprintf("CARD DISCARD REQUIRED (source: %s)", sel.Source),
		fmt.Sprintf("Discard %d-%d cards from hand.", sel.MinCards, sel.MaxCards),
		`Send a card-discard-confirmed command with cardsToDiscard.`,
	}, "\n")
}

func formatPendingBehaviorChoice(sel *dto.PendingBehaviorChoiceSelectionDto) string {
	var choiceLines []string
	for i, c := range sel.Choices {
		desc := formatBehaviorBrief(c.Inputs, c.Outputs)
		avail := ""
		if !c.Available {
			avail = " [UNAVAILABLE]"
		}
		choiceLines = append(choiceLines, fmt.Sprintf("  %d: %s%s", i, desc, avail))
	}

	return strings.Join([]string{
		fmt.Sprintf("BEHAVIOR CHOICE REQUIRED (source: %s)", sel.Source),
		strings.Join(choiceLines, "\n"),
		`Send a behavior-choice-confirmed command with choiceIndex.`,
	}, "\n")
}

func formatForcedAction(fa *dto.ForcedFirstActionDto) string {
	return strings.Join([]string{
		fmt.Sprintf("FORCED FIRST ACTION: %s", fa.Description),
		fmt.Sprintf("Action type: %s", fa.ActionType),
		fmt.Sprintf("Corporation: %s", fa.CorporationID),
	}, "\n")
}

func formatStartingSelection(player *dto.PlayerDto) string {
	parts := []string{"STARTING SELECTION REQUIRED"}

	if player.SelectCorporationPhase != nil {
		var corps []string
		for _, c := range player.SelectCorporationPhase.AvailableCorporations {
			corps = append(corps, fmt.Sprintf("  - %s [%s]: %s", c.Name, c.ID, c.Description))
		}
		parts = append(parts, "Corporations:\n"+strings.Join(corps, "\n"))
	}

	if player.SelectPreludeCardsPhase != nil {
		var preludes []string
		for _, c := range player.SelectPreludeCardsPhase.AvailablePreludes {
			preludes = append(preludes, fmt.Sprintf("  - %s [%s]: %s", c.Name, c.ID, c.Description))
		}
		parts = append(parts, fmt.Sprintf("Preludes (pick %d):\n%s",
			player.SelectPreludeCardsPhase.MaxSelectable,
			strings.Join(preludes, "\n")))
	}

	if player.SelectStartingCardsPhase != nil {
		var cards []string
		for _, c := range player.SelectStartingCardsPhase.AvailableCards {
			tags := ""
			if len(c.Tags) > 0 {
				tagStrs := make([]string, len(c.Tags))
				for i, t := range c.Tags {
					tagStrs[i] = string(t)
				}
				tags = " " + strings.Join(tagStrs, ", ")
			}
			cards = append(cards, fmt.Sprintf("  - %s [%s] (%dM€) [%s]%s: %s",
				c.Name, c.ID, c.Cost, string(c.Type), tags, c.Description))
		}
		parts = append(parts, "Starting cards (pick any to buy at 3M€ each):\n"+strings.Join(cards, "\n"))
	}

	parts = append(parts, "Send a select-starting-choices command with corporationId, preludeIds, and cardIds.")

	return strings.Join(parts, "\n")
}

func formatProductionPhase(player *dto.PlayerDto) string {
	pp := player.ProductionPhase
	var cards []string
	for _, c := range pp.AvailableCards {
		cards = append(cards, fmt.Sprintf("  - %s [%s]", c.Name, c.ID))
	}

	cardList := "  (no cards available)"
	if len(cards) > 0 {
		cardList = strings.Join(cards, "\n")
	}

	return strings.Join([]string{
		"PRODUCTION PHASE - Select cards to buy:",
		cardList,
		`Send a confirm-production-cards command with cardIds.`,
	}, "\n")
}

func formatPlayerStatus(player *dto.PlayerDto) string {
	r := &player.Resources
	p := &player.Production

	corpName := "None"
	if player.Corporation != nil {
		corpName = player.Corporation.Name
	}

	lines := []string{
		"=== YOUR STATUS ===",
		fmt.Sprintf("Name: %s | Corporation: %s | TR: %d", player.Name, corpName, player.TerraformRating),
		fmt.Sprintf("Status: %s | Actions remaining: %d | Passed: %v", string(player.Status), player.AvailableActions, player.Passed),
		"",
		"Resources (amount / production):",
		fmt.Sprintf("  Credits:  %d / %s", r.Credits, formatProd(p.Credits)),
		fmt.Sprintf("  Steel:    %d / %s", r.Steel, formatProd(p.Steel)),
		fmt.Sprintf("  Titanium: %d / %s", r.Titanium, formatProd(p.Titanium)),
		fmt.Sprintf("  Plants:   %d / %s", r.Plants, formatProd(p.Plants)),
		fmt.Sprintf("  Energy:   %d / %s", r.Energy, formatProd(p.Energy)),
		fmt.Sprintf("  Heat:     %d / %s", r.Heat, formatProd(p.Heat)),
	}

	if len(player.PlayedCards) > 0 {
		var names []string
		for _, c := range player.PlayedCards {
			names = append(names, c.Name)
		}
		lines = append(lines, "", fmt.Sprintf("Played cards (%d): %s", len(player.PlayedCards), strings.Join(names, ", ")))
	}

	if len(player.ResourceStorage) > 0 {
		var storage []string
		for k, v := range player.ResourceStorage {
			if v > 0 {
				storage = append(storage, fmt.Sprintf("%s: %d", k, v))
			}
		}
		if len(storage) > 0 {
			lines = append(lines, fmt.Sprintf("Resource storage: %s", strings.Join(storage, ", ")))
		}
	}

	if len(player.PaymentSubstitutes) > 0 {
		var subs []string
		for _, s := range player.PaymentSubstitutes {
			subs = append(subs, fmt.Sprintf("%s (%d:1)", string(s.ResourceType), s.ConversionRate))
		}
		lines = append(lines, fmt.Sprintf("Payment substitutes: %s", strings.Join(subs, ", ")))
	}

	if len(player.Effects) > 0 {
		var effects []string
		for _, e := range player.Effects {
			effects = append(effects, e.CardName)
		}
		lines = append(lines, fmt.Sprintf("Active effects: %s", strings.Join(effects, ", ")))
	}

	return strings.Join(lines, "\n")
}

func formatHand(cards []dto.PlayerCardDto) string {
	if len(cards) == 0 {
		return "=== HAND (0 cards) ===\n(empty)"
	}

	header := fmt.Sprintf("=== HAND (%d cards) ===", len(cards))
	var cardLines []string
	for _, c := range cards {
		avail := "PLAYABLE"
		if !c.Available {
			avail = "BLOCKED"
		}

		errInfo := ""
		if !c.Available && len(c.Errors) > 0 {
			var msgs []string
			for _, e := range c.Errors {
				msgs = append(msgs, e.Message)
			}
			errInfo = fmt.Sprintf(" (%s)", strings.Join(msgs, "; "))
		}

		tags := ""
		if len(c.Tags) > 0 {
			tagStrs := make([]string, len(c.Tags))
			for i, t := range c.Tags {
				tagStrs[i] = string(t)
			}
			tags = fmt.Sprintf(" [%s]", strings.Join(tagStrs, ", "))
		}

		discount := ""
		if c.EffectiveCost < c.Cost {
			discount = fmt.Sprintf(" (discounted from %d)", c.Cost)
		}

		line := fmt.Sprintf("  - %s [%s] | %dM€%s | %s%s | %s%s",
			c.Name, c.ID, c.EffectiveCost, discount, string(c.Type), tags, avail, errInfo)

		cardLines = append(cardLines, line)
	}

	return header + "\n" + strings.Join(cardLines, "\n")
}

func formatCardActions(actions []dto.PlayerActionDto) string {
	if len(actions) == 0 {
		return ""
	}

	header := "=== CARD ACTIONS ==="
	var actionLines []string
	for _, a := range actions {
		avail := "AVAILABLE"
		if !a.Available {
			avail = "BLOCKED"
		}

		errInfo := ""
		if !a.Available && len(a.Errors) > 0 {
			var msgs []string
			for _, e := range a.Errors {
				msgs = append(msgs, e.Message)
			}
			errInfo = fmt.Sprintf(" (%s)", strings.Join(msgs, "; "))
		}

		usedInfo := ""
		if a.TimesUsedThisTurn > 0 {
			usedInfo = fmt.Sprintf(" [used %dx this turn]", a.TimesUsedThisTurn)
		}

		desc := formatBehaviorBrief(a.Behavior.Inputs, a.Behavior.Outputs)

		line := fmt.Sprintf("  - %s [%s] behavior#%d | %s%s%s",
			a.CardName, a.CardID, a.BehaviorIndex, avail, usedInfo, errInfo)
		if desc != "" {
			line += fmt.Sprintf("\n    %s", desc)
		}

		actionLines = append(actionLines, line)
	}

	return header + "\n" + strings.Join(actionLines, "\n")
}

func formatStandardProjects(projects []dto.PlayerStandardProjectDto) string {
	if len(projects) == 0 {
		return ""
	}

	header := "=== STANDARD PROJECTS ==="
	var lines []string
	for _, p := range projects {
		avail := "AVAILABLE"
		if !p.Available {
			avail = "BLOCKED"
		}

		errInfo := ""
		if !p.Available && len(p.Errors) > 0 {
			errInfo = fmt.Sprintf(" (%s)", p.Errors[0].Message)
		}

		var costParts []string
		for k, v := range p.EffectiveCost {
			costParts = append(costParts, fmt.Sprintf("%d %s", v, k))
		}
		costStr := strings.Join(costParts, ", ")

		lines = append(lines, fmt.Sprintf("  - %s | %s | %s%s", p.ProjectType, costStr, avail, errInfo))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func formatMilestones(milestones []dto.PlayerMilestoneDto) string {
	if len(milestones) == 0 {
		return ""
	}

	header := "=== MILESTONES ==="
	var lines []string
	for _, m := range milestones {
		claimed := "NOT YET"
		if m.IsClaimed {
			by := ""
			if m.ClaimedBy != nil {
				by = *m.ClaimedBy
			}
			claimed = fmt.Sprintf("CLAIMED by %s", by)
		} else if m.Available {
			claimed = fmt.Sprintf("CLAIMABLE (%dM€)", m.ClaimCost)
		}
		lines = append(lines, fmt.Sprintf("  - %s: %s | Progress: %d/%d | %s",
			m.Name, m.Description, m.Progress, m.Required, claimed))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func formatAwards(awards []dto.PlayerAwardDto) string {
	if len(awards) == 0 {
		return ""
	}

	header := "=== AWARDS ==="
	var lines []string
	for _, a := range awards {
		funded := "NOT AVAILABLE"
		if a.IsFunded {
			by := ""
			if a.FundedBy != nil {
				by = *a.FundedBy
			}
			funded = fmt.Sprintf("FUNDED by %s", by)
		} else if a.Available {
			funded = fmt.Sprintf("FUNDABLE (%dM€)", a.FundingCost)
		}
		lines = append(lines, fmt.Sprintf("  - %s: %s | %s", a.Name, a.Description, funded))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func formatOpponents(others []dto.OtherPlayerDto) string {
	if len(others) == 0 {
		return ""
	}

	header := "=== OPPONENTS ==="
	var lines []string
	for _, o := range others {
		r := &o.Resources
		corpName := "None"
		if o.Corporation != nil {
			corpName = o.Corporation.Name
		}
		lines = append(lines, strings.Join([]string{
			fmt.Sprintf("  %s (%s) | TR: %d | Status: %s | Passed: %v",
				o.Name, corpName, o.TerraformRating, string(o.Status), o.Passed),
			fmt.Sprintf("    Resources: %dM€, %d steel, %d ti, %d plants, %d energy, %d heat",
				r.Credits, r.Steel, r.Titanium, r.Plants, r.Energy, r.Heat),
			fmt.Sprintf("    Cards in hand: %d | Played: %d cards",
				o.HandCardCount, len(o.PlayedCards)),
		}, "\n"))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func formatBoard(tiles []dto.TileDto) string {
	var occupied []dto.TileDto
	for _, t := range tiles {
		if t.OccupiedBy != nil {
			occupied = append(occupied, t)
		}
	}

	if len(occupied) == 0 {
		return ""
	}

	header := "=== BOARD ==="

	var lines []string
	for _, t := range occupied {
		coord := fmt.Sprintf("(%d,%d,%d)", t.Coordinates.Q, t.Coordinates.R, t.Coordinates.S)
		occ := ""
		if t.OccupiedBy != nil {
			ownerStr := ""
			if t.OwnerID != nil {
				ownerStr = fmt.Sprintf(" (owner: %s)", *t.OwnerID)
			}
			occ = fmt.Sprintf(" | %s%s", t.OccupiedBy.Type, ownerStr)
		}
		name := ""
		if t.DisplayName != nil {
			name = " " + *t.DisplayName
		}
		bonuses := ""
		if len(t.Bonuses) > 0 {
			var bonusParts []string
			for _, b := range t.Bonuses {
				bonusParts = append(bonusParts, fmt.Sprintf("%dx %s", b.Amount, b.Type))
			}
			bonuses = " | bonuses: " + strings.Join(bonusParts, ", ")
		}
		lines = append(lines, fmt.Sprintf("  %s%s%s%s", coord, name, occ, bonuses))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func formatFinalScores(game *dto.GameDto) string {
	if len(game.FinalScores) == 0 {
		return ""
	}

	header := "=== FINAL SCORES ==="
	var lines []string
	for _, s := range game.FinalScores {
		vp := s.VPBreakdown
		winner := ""
		if s.IsWinner {
			winner = " (WINNER)"
		}
		lines = append(lines, strings.Join([]string{
			fmt.Sprintf("  #%d %s%s: %d VP", s.Placement, s.PlayerName, winner, vp.TotalVP),
			fmt.Sprintf("    TR: %d, Cards: %d, Greenery: %d, City: %d, Milestones: %d, Awards: %d",
				vp.TerraformRating, vp.CardVP, vp.GreeneryVP, vp.CityVP, vp.MilestoneVP, vp.AwardVP),
		}, "\n"))
	}

	return header + "\n" + strings.Join(lines, "\n")
}

func formatProd(val int) string {
	if val >= 0 {
		return fmt.Sprintf("+%d", val)
	}
	return fmt.Sprintf("%d", val)
}

func formatBehaviorBrief(inputs, outputs []dto.ResourceConditionDto) string {
	var parts []string

	if len(inputs) > 0 {
		var items []string
		for _, i := range inputs {
			items = append(items, fmt.Sprintf("%d %s", i.Amount, string(i.Type)))
		}
		parts = append(parts, "Costs: "+strings.Join(items, ", "))
	}
	if len(outputs) > 0 {
		var items []string
		for _, o := range outputs {
			items = append(items, fmt.Sprintf("%d %s", o.Amount, string(o.Type)))
		}
		parts = append(parts, "Gives: "+strings.Join(items, ", "))
	}

	return strings.Join(parts, " -> ")
}

func findPlayerName(game *dto.GameDto, playerID string) string {
	if game.CurrentPlayer.ID == playerID {
		return game.CurrentPlayer.Name
	}
	for _, o := range game.OtherPlayers {
		if o.ID == playerID {
			return o.Name
		}
	}
	return playerID
}

func formatRecentLog(diffs []game.StateDiff, maxEntries int) string {
	if len(diffs) == 0 {
		return ""
	}

	start := 0
	if len(diffs) > maxEntries {
		start = len(diffs) - maxEntries
	}
	recent := diffs[start:]

	var lines []string
	for _, d := range recent {
		if d.Description == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("  - %s", d.Description))
	}

	if len(lines) == 0 {
		return ""
	}

	return "=== RECENT GAME LOG ===\n" + strings.Join(lines, "\n")
}

func formatRecentChat(messages []shared.ChatMessage, maxEntries int) string {
	if len(messages) == 0 {
		return ""
	}

	start := 0
	if len(messages) > maxEntries {
		start = len(messages) - maxEntries
	}
	recent := messages[start:]

	var lines []string
	for _, m := range recent {
		lines = append(lines, fmt.Sprintf("  %s: %s", m.SenderName, m.Message))
	}

	return "=== RECENT CHAT ===\n" + strings.Join(lines, "\n")
}
