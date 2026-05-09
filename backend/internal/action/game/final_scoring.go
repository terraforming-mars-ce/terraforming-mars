package game

import (
	"context"
	"sort"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/milestones"
)

// FinalScoringAction handles the business logic for calculating final scores and ending the game
type FinalScoringAction struct {
	gameRepo          game.GameRepository
	cardRegistry      cards.CardRegistry
	awardRegistry     awards.AwardRegistry
	milestoneRegistry milestones.MilestoneRegistry
	logger            *zap.Logger
}

// NewFinalScoringAction creates a new final scoring action
func NewFinalScoringAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	awardRegistry awards.AwardRegistry,
	milestoneRegistry milestones.MilestoneRegistry,
	logger *zap.Logger,
) *FinalScoringAction {
	return &FinalScoringAction{
		gameRepo:          gameRepo,
		cardRegistry:      cardRegistry,
		awardRegistry:     awardRegistry,
		milestoneRegistry: milestoneRegistry,
		logger:            logger,
	}
}

// PlayerScore holds a player's score with breakdown for sorting
type PlayerScore struct {
	PlayerID   string
	PlayerName string
	Breakdown  shared.VPBreakdown
	Credits    int // For tiebreaker
}

// Execute performs the final scoring action
func (a *FinalScoringAction) Execute(ctx context.Context, gameID string) error {
	log := a.logger.With(zap.String("game_id", gameID))
	log.Debug("Starting final scoring")

	// 1. Fetch game
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return err
	}

	// 2. Validate game is active
	if g.Status() != shared.GameStatusActive {
		log.Warn("Game is not active, skipping final scoring", zap.String("status", string(g.Status())))
		return nil
	}

	// 3. Get all players
	allPlayers := g.GetAllPlayers()
	if len(allPlayers) == 0 {
		log.Warn("No players in game")
		return nil
	}

	// 4. Compute VP breakdowns via the shared single-source-of-truth helper.
	breakdowns := ComputePlayerVPBreakdowns(g, a.cardRegistry, a.awardRegistry, a.milestoneRegistry)

	// 5. Build PlayerScore entries for sorting.
	scores := make([]PlayerScore, len(allPlayers))
	for i, p := range allPlayers {
		breakdown := breakdowns[p.ID()]
		scores[i] = PlayerScore{
			PlayerID:   p.ID(),
			PlayerName: p.Name(),
			Breakdown:  breakdown,
			Credits:    p.Resources().Get().Credits,
		}
		log.Debug("Player VP calculated",
			zap.String("player_id", p.ID()),
			zap.String("player_name", p.Name()),
			zap.Int("total_vp", breakdown.TotalVP),
			zap.Int("tr", breakdown.TerraformRating),
			zap.Int("card_vp", breakdown.CardVP),
			zap.Int("milestone_vp", breakdown.MilestoneVP),
			zap.Int("award_vp", breakdown.AwardVP),
			zap.Int("greenery_vp", breakdown.GreeneryVP),
			zap.Int("city_vp", breakdown.CityVP),
		)
	}

	// 6. Sort by total VP (descending), then credits (descending) for tiebreaker
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Breakdown.TotalVP != scores[j].Breakdown.TotalVP {
			return scores[i].Breakdown.TotalVP > scores[j].Breakdown.TotalVP
		}
		return scores[i].Credits > scores[j].Credits
	})

	// 7. Determine winner and check for ties
	winnerID := scores[0].PlayerID
	isTie := false
	if len(scores) > 1 {
		// Check if top players have same VP and credits (true tie)
		if scores[0].Breakdown.TotalVP == scores[1].Breakdown.TotalVP &&
			scores[0].Credits == scores[1].Credits {
			isTie = true
		}
	}

	log.Debug("Winner determined",
		zap.String("winner_id", winnerID),
		zap.String("winner_name", scores[0].PlayerName),
		zap.Int("winning_vp", scores[0].Breakdown.TotalVP),
		zap.Bool("is_tie", isTie),
	)

	// 8. Build shared.FinalScore entries and store in game.
	finalScores := make([]shared.FinalScore, len(scores))
	for i, s := range scores {
		finalScores[i] = shared.FinalScore{
			PlayerID:   s.PlayerID,
			PlayerName: s.PlayerName,
			Breakdown:  s.Breakdown,
			Credits:    s.Credits,
			Placement:  i + 1, // 1-indexed placement
			IsWinner:   s.PlayerID == winnerID,
		}
	}
	err = g.SetFinalScores(ctx, finalScores, winnerID, isTie)
	if err != nil {
		log.Error("Failed to set final scores", zap.Error(err))
		return err
	}

	// 9. Update game status to completed
	err = g.UpdateStatus(ctx, shared.GameStatusCompleted)
	if err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return err
	}

	// 10. Update game phase to complete
	err = g.UpdatePhase(ctx, shared.GamePhaseComplete)
	if err != nil {
		log.Error("Failed to update game phase", zap.Error(err))
		return err
	}

	// 11. Publish GameEndedEvent
	events.Publish(g.EventBus(), events.GameEndedEvent{
		GameID:    gameID,
		WinnerID:  winnerID,
		IsTie:     isTie,
		Timestamp: time.Now(),
	})

	log.Info("Final scoring complete, game ended")
	return nil
}

// convertToClaimedMilestoneInfo converts game milestones to the format expected by VP calculator
func convertToClaimedMilestoneInfo(claimed []shared.ClaimedMilestone) []gamecards.ClaimedMilestoneInfo {
	result := make([]gamecards.ClaimedMilestoneInfo, len(claimed))
	for i, m := range claimed {
		result[i] = gamecards.ClaimedMilestoneInfo{
			Type:     string(m.Type),
			PlayerID: m.PlayerID,
		}
	}
	return result
}

// convertToFundedAwardInfo converts game awards to the format expected by VP calculator
func convertToFundedAwardInfo(funded []shared.FundedAward) []gamecards.FundedAwardInfo {
	result := make([]gamecards.FundedAwardInfo, len(funded))
	for i, a := range funded {
		result[i] = gamecards.FundedAwardInfo{
			Type: string(a.Type),
		}
	}
	return result
}

// convertCardVPDetails converts gamecards.CardVPDetail to shared.CardVPDetail
func convertCardVPDetails(details []gamecards.CardVPDetail) []shared.CardVPDetail {
	result := make([]shared.CardVPDetail, len(details))
	for i, d := range details {
		conditions := make([]shared.CardVPConditionDetail, len(d.Conditions))
		for j, c := range d.Conditions {
			conditions[j] = shared.CardVPConditionDetail{
				ConditionType:  c.ConditionType,
				Amount:         c.Amount,
				Count:          c.Count,
				MaxTrigger:     c.MaxTrigger,
				ActualTriggers: c.ActualTriggers,
				TotalVP:        c.TotalVP,
				Explanation:    c.Explanation,
			}
		}
		result[i] = shared.CardVPDetail{
			CardID:     d.CardID,
			CardName:   d.CardName,
			Conditions: conditions,
			TotalVP:    d.TotalVP,
		}
	}
	return result
}

// convertGreeneryVPDetails converts gamecards.GreeneryVPDetail to shared.GreeneryVPDetail
func convertGreeneryVPDetails(details []gamecards.GreeneryVPDetail) []shared.GreeneryVPDetail {
	result := make([]shared.GreeneryVPDetail, len(details))
	for i, d := range details {
		result[i] = shared.GreeneryVPDetail{
			Coordinate: d.Coordinate,
			VP:         d.VP,
		}
	}
	return result
}

// convertCityVPDetails converts gamecards.CityVPDetail to shared.CityVPDetail
func convertCityVPDetails(details []gamecards.CityVPDetail) []shared.CityVPDetail {
	result := make([]shared.CityVPDetail, len(details))
	for i, d := range details {
		result[i] = shared.CityVPDetail{
			CityCoordinate:     d.CityCoordinate,
			AdjacentGreeneries: d.AdjacentGreeneries,
			VP:                 d.VP,
		}
	}
	return result
}
