package game

import (
	"context"
	"sort"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
)

// FinalScoringAction handles the business logic for calculating final scores and ending the game
type FinalScoringAction struct {
	gameRepo     game.GameRepository
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewFinalScoringAction creates a new final scoring action
func NewFinalScoringAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *FinalScoringAction {
	return &FinalScoringAction{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// PlayerScore holds a player's score with breakdown for sorting
type PlayerScore struct {
	PlayerID   string
	PlayerName string
	Breakdown  gamecards.VPBreakdown
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
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active, skipping final scoring", zap.String("status", string(g.Status())))
		return nil
	}

	// 3. Get all players
	allPlayers := g.GetAllPlayers()
	if len(allPlayers) == 0 {
		log.Warn("No players in game")
		return nil
	}

	// 4. Prepare milestone and award data for VP calculation
	claimedMilestones := convertToClaimedMilestoneInfo(g.Milestones().ClaimedMilestones())
	fundedAwards := convertToFundedAwardInfo(g.Awards().FundedAwards())

	// 5. Calculate VP for each player
	scores := make([]PlayerScore, len(allPlayers))
	for i, p := range allPlayers {
		breakdown := gamecards.CalculatePlayerVP(
			p,
			g.Board(),
			claimedMilestones,
			fundedAwards,
			allPlayers,
			a.cardRegistry,
		)
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

	// 8. Convert to game.FinalScore and store in game
	finalScores := make([]game.FinalScore, len(scores))
	for i, s := range scores {
		finalScores[i] = game.FinalScore{
			PlayerID:   s.PlayerID,
			PlayerName: s.PlayerName,
			Breakdown: game.VPBreakdown{
				TerraformRating:   s.Breakdown.TerraformRating,
				CardVP:            s.Breakdown.CardVP,
				CardVPDetails:     convertCardVPDetails(s.Breakdown.CardVPDetails),
				MilestoneVP:       s.Breakdown.MilestoneVP,
				AwardVP:           s.Breakdown.AwardVP,
				GreeneryVP:        s.Breakdown.GreeneryVP,
				GreeneryVPDetails: convertGreeneryVPDetails(s.Breakdown.GreeneryVPDetails),
				CityVP:            s.Breakdown.CityVP,
				CityVPDetails:     convertCityVPDetails(s.Breakdown.CityVPDetails),
				TotalVP:           s.Breakdown.TotalVP,
			},
			Credits:   s.Credits,
			Placement: i + 1, // 1-indexed placement
			IsWinner:  s.PlayerID == winnerID,
		}
	}
	err = g.SetFinalScores(ctx, finalScores, winnerID, isTie)
	if err != nil {
		log.Error("Failed to set final scores", zap.Error(err))
		return err
	}

	// 9. Update game status to completed
	err = g.UpdateStatus(ctx, game.GameStatusCompleted)
	if err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return err
	}

	// 10. Update game phase to complete
	err = g.UpdatePhase(ctx, game.GamePhaseComplete)
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
func convertToClaimedMilestoneInfo(claimed []game.ClaimedMilestone) []gamecards.ClaimedMilestoneInfo {
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
func convertToFundedAwardInfo(funded []game.FundedAward) []gamecards.FundedAwardInfo {
	result := make([]gamecards.FundedAwardInfo, len(funded))
	for i, a := range funded {
		result[i] = gamecards.FundedAwardInfo{
			Type: string(a.Type),
		}
	}
	return result
}

// convertCardVPDetails converts gamecards.CardVPDetail to game.CardVPDetail
func convertCardVPDetails(details []gamecards.CardVPDetail) []game.CardVPDetail {
	result := make([]game.CardVPDetail, len(details))
	for i, d := range details {
		conditions := make([]game.CardVPConditionDetail, len(d.Conditions))
		for j, c := range d.Conditions {
			conditions[j] = game.CardVPConditionDetail{
				ConditionType:  c.ConditionType,
				Amount:         c.Amount,
				Count:          c.Count,
				MaxTrigger:     c.MaxTrigger,
				ActualTriggers: c.ActualTriggers,
				TotalVP:        c.TotalVP,
				Explanation:    c.Explanation,
			}
		}
		result[i] = game.CardVPDetail{
			CardID:     d.CardID,
			CardName:   d.CardName,
			Conditions: conditions,
			TotalVP:    d.TotalVP,
		}
	}
	return result
}

// convertGreeneryVPDetails converts gamecards.GreeneryVPDetail to game.GreeneryVPDetail
func convertGreeneryVPDetails(details []gamecards.GreeneryVPDetail) []game.GreeneryVPDetail {
	result := make([]game.GreeneryVPDetail, len(details))
	for i, d := range details {
		result[i] = game.GreeneryVPDetail{
			Coordinate: d.Coordinate,
			VP:         d.VP,
		}
	}
	return result
}

// convertCityVPDetails converts gamecards.CityVPDetail to game.CityVPDetail
func convertCityVPDetails(details []gamecards.CityVPDetail) []game.CityVPDetail {
	result := make([]game.CityVPDetail, len(details))
	for i, d := range details {
		result[i] = game.CityVPDetail{
			CityCoordinate:     d.CityCoordinate,
			AdjacentGreeneries: d.AdjacentGreeneries,
			VP:                 d.VP,
		}
	}
	return result
}
