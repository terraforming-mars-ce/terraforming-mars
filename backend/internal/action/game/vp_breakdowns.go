package game

import (
	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/milestones"
)

// ComputePlayerVPBreakdowns is the single source of truth for per-player VP
// breakdowns. Both FinalScoringAction (game end) and the history snapshot enricher
// (per-snapshot projected score) call this so any future VP rule change applies
// to both code paths automatically.
func ComputePlayerVPBreakdowns(
	g *game.Game,
	cardRegistry cards.CardRegistry,
	awardRegistry awards.AwardRegistry,
	milestoneRegistry milestones.MilestoneRegistry,
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
