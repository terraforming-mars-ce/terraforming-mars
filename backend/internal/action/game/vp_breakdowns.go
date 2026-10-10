package game

import (
	"openmars/internal/game"
	"openmars/internal/game/award"
	gamecards "openmars/internal/game/cards"
	"openmars/internal/game/milestone"
	"openmars/internal/game/shared"
)

// ComputePlayerVPBreakdowns is the single source of truth for per-player VP
// breakdowns. Both FinalScoringAction (game end) and the history snapshot enricher
// (per-snapshot projected score) call this so any future VP rule change applies
// to both code paths automatically.
func ComputePlayerVPBreakdowns(
	g *game.Game,
	cardRegistry gamecards.CardRegistry,
	awardRegistry award.AwardRegistry,
	milestoneRegistry milestone.MilestoneRegistry,
) map[string]shared.VPBreakdown {
	allPlayers := g.GetAllPlayers()
	if len(allPlayers) == 0 {
		return nil
	}

	claimedMilestones := convertToClaimedMilestoneInfo(g.Milestones().ClaimedMilestones())
	fundedAwards := convertToFundedAwardInfo(g.Awards().FundedAwards())

	out := make(map[string]shared.VPBreakdown, len(allPlayers))
	for _, p := range allPlayers {
		breakdown := gamecards.CalculatePlayerVP(
			p, g, claimedMilestones, fundedAwards, allPlayers,
			cardRegistry, awardRegistry, milestoneRegistry,
		)
		out[p.ID()] = toSharedVPBreakdown(breakdown)
	}
	return out
}

func toSharedVPBreakdown(b gamecards.VPBreakdown) shared.VPBreakdown {
	return shared.VPBreakdown{
		TerraformRating:   b.TerraformRating,
		CardVP:            b.CardVP,
		CardVPDetails:     convertCardVPDetails(b.CardVPDetails),
		MilestoneVP:       b.MilestoneVP,
		AwardVP:           b.AwardVP,
		GreeneryVP:        b.GreeneryVP,
		GreeneryVPDetails: convertGreeneryVPDetails(b.GreeneryVPDetails),
		CityVP:            b.CityVP,
		CityVPDetails:     convertCityVPDetails(b.CityVPDetails),
		TotalVP:           b.TotalVP,
	}
}
