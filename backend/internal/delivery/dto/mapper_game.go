package dto

import (
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
)

// ToGameDto converts Game to GameDto with personalized view
// The playerID parameter determines which player is "currentPlayer" vs "otherPlayers"
func ToGameDto(g *game.Game, cardRegistry cards.CardRegistry, playerID string) GameDto {
	players := g.GetAllPlayers()

	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	var viewingPlayer *player.Player
	for _, p := range players {
		if p.ID() == playerID {
			viewingPlayer = p
			currentPlayer = ToPlayerDto(p, g, cardRegistry)
		} else {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(p, g, cardRegistry))
		}
	}

	if viewingPlayer == nil && len(players) > 0 {
		otherPlayers = make([]OtherPlayerDto, 0)
		currentPlayer = ToPlayerDto(players[0], g, cardRegistry)
		for i := 1; i < len(players); i++ {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(players[i], g, cardRegistry))
		}
		playerID = players[0].ID()
	}

	settings := g.Settings()
	settingsDto := GameSettingsDto{
		MaxPlayers:       settings.MaxPlayers,
		VenusNextEnabled: settings.VenusNextEnabled,
		DevelopmentMode:  settings.DevelopmentMode,
		DemoGame:         settings.DemoGame,
		CardPacks:        settings.CardPacks,
	}

	globalParams := g.GlobalParameters()
	globalParamsDto := GlobalParametersDto{
		Temperature: globalParams.Temperature(),
		Oxygen:      globalParams.Oxygen(),
		Oceans:      globalParams.Oceans(),
		Venus:       globalParams.Venus(),
	}

	board := g.Board()
	tiles := board.Tiles()
	tileDtos := make([]TileDto, len(tiles))
	for i, tile := range tiles {
		tileDtos[i] = TileDto{
			Coordinates: HexPositionDto{
				Q: tile.Coordinates.Q,
				R: tile.Coordinates.R,
				S: tile.Coordinates.S,
			},
			Type:        string(tile.Type),
			OwnerID:     tile.OwnerID,
			Tags:        tile.Tags,
			Bonuses:     convertTileBonuses(tile.Bonuses),
			Location:    string(tile.Location),
			DisplayName: tile.DisplayName,
			ReservedBy:  tile.ReservedBy,
		}
		if tile.OccupiedBy != nil {
			occupant := &TileOccupantDto{
				Type: string(tile.OccupiedBy.Type),
				Tags: tile.OccupiedBy.Tags,
			}
			tileDtos[i].OccupiedBy = occupant
		}
	}

	paymentConstants := PaymentConstantsDto{
		SteelValue:    2, // Default steel value
		TitaniumValue: 3, // Default titanium value
	}

	var finalScoreDtos []FinalScoreDto
	if g.Status() == game.GameStatusCompleted {
		finalScores := g.GetFinalScores()
		if finalScores != nil {
			finalScoreDtos = make([]FinalScoreDto, len(finalScores))
			for i, fs := range finalScores {
				finalScoreDtos[i] = FinalScoreDto{
					PlayerID:    fs.PlayerID,
					PlayerName:  fs.PlayerName,
					VPBreakdown: ToVPBreakdownDto(fs.Breakdown),
					IsWinner:    fs.IsWinner,
					Placement:   fs.Placement,
				}
			}
		}
	}

	triggeredEffects := g.GetTriggeredEffects()
	var triggeredEffectDtos []TriggeredEffectDto
	if len(triggeredEffects) > 0 {
		triggeredEffectDtos = make([]TriggeredEffectDto, len(triggeredEffects))
		for i, effect := range triggeredEffects {
			triggeredEffectDtos[i] = ToTriggeredEffectDto(effect)
		}
	}

	return GameDto{
		ID:               g.ID(),
		Status:           GameStatus(g.Status()),
		Settings:         settingsDto,
		HostPlayerID:     g.HostPlayerID(),
		CurrentPhase:     GamePhase(g.CurrentPhase()),
		GlobalParameters: globalParamsDto,
		CurrentPlayer:    currentPlayer,
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  playerID,
		CurrentTurn:      getCurrentTurnPlayerID(g),
		Generation:       g.Generation(),
		TurnOrder:        g.TurnOrder(),
		Board: BoardDto{
			Tiles: tileDtos,
		},
		PaymentConstants:   paymentConstants,
		Milestones:         ToMilestonesDto(g, cardRegistry),
		Awards:             ToAwardsDto(g, cardRegistry),
		AwardResults:       ToAwardResultsDto(g, cardRegistry),
		FinalScores:        finalScoreDtos,
		TriggeredEffects:   triggeredEffectDtos,
		PlaceableTileTypes: ToPlaceableTileTypeDtos(),
	}
}

// ToPlaceableTileTypeDtos converts the board PlaceableTileTypes registry to DTOs
func ToPlaceableTileTypeDtos() []PlaceableTileTypeDto {
	dtos := make([]PlaceableTileTypeDto, len(board.PlaceableTileTypes))
	for i, pt := range board.PlaceableTileTypes {
		dtos[i] = PlaceableTileTypeDto{
			Type:  pt.Type,
			Label: pt.Label,
			Group: pt.Group,
		}
	}
	return dtos
}

// getCurrentTurnPlayerID extracts the player ID from the current turn
func getCurrentTurnPlayerID(g *game.Game) *string {
	turn := g.CurrentTurn()
	if turn == nil {
		return nil
	}
	playerID := turn.PlayerID()
	return &playerID
}

// convertTileBonuses converts TileBonus to DTO
func convertTileBonuses(bonuses []board.TileBonus) []TileBonusDto {
	dtos := make([]TileBonusDto, len(bonuses))
	for i, bonus := range bonuses {
		dtos[i] = TileBonusDto{
			Type:   string(bonus.Type),
			Amount: bonus.Amount,
		}
	}
	return dtos
}

// ToMilestonesDto converts all milestones to DTOs including claim status and per-player progress
func ToMilestonesDto(g *game.Game, cardRegistry cards.CardRegistry) []MilestoneDto {
	milestones := g.Milestones()
	players := g.GetAllPlayers()
	b := g.Board()

	dtos := make([]MilestoneDto, len(game.AllMilestones))
	for i, info := range game.AllMilestones {
		var claimedBy *string
		isClaimed := milestones.IsClaimed(info.Type)
		if isClaimed {
			for _, claimed := range milestones.ClaimedMilestones() {
				if claimed.Type == info.Type {
					claimedBy = &claimed.PlayerID
					break
				}
			}
		}

		playerProgress := make(map[string]int, len(players))
		for _, p := range players {
			playerProgress[p.ID()] = gamecards.GetPlayerMilestoneProgress(info.Type, p, b, cardRegistry)
		}

		dtos[i] = MilestoneDto{
			Type:           string(info.Type),
			Name:           info.Name,
			Description:    info.Description,
			IsClaimed:      isClaimed,
			ClaimedBy:      claimedBy,
			ClaimCost:      game.MilestoneClaimCost,
			Required:       info.Requirement,
			PlayerProgress: playerProgress,
		}
	}
	return dtos
}

// ToAwardsDto converts all awards to DTOs including funding status and per-player scores
func ToAwardsDto(g *game.Game, cardRegistry cards.CardRegistry) []AwardDto {
	awards := g.Awards()
	players := g.GetAllPlayers()
	b := g.Board()

	dtos := make([]AwardDto, len(game.AllAwards))
	fundedCount := awards.FundedCount()

	for i, info := range game.AllAwards {
		var fundedBy *string
		isFunded := awards.IsFunded(info.Type)
		fundingCost := game.AwardFundingCosts[0]

		if isFunded {
			for _, funded := range awards.FundedAwards() {
				if funded.Type == info.Type {
					fundedBy = &funded.FundedByPlayer
					fundingCost = funded.FundingCost
					break
				}
			}
		} else {
			if fundedCount < game.MaxFundedAwards {
				fundingCost = game.AwardFundingCosts[fundedCount]
			}
		}

		playerProgress := make(map[string]int, len(players))
		for _, p := range players {
			playerProgress[p.ID()] = gamecards.CalculateAwardScore(info.Type, p, b, cardRegistry)
		}

		dtos[i] = AwardDto{
			Type:           string(info.Type),
			Name:           info.Name,
			Description:    info.Description,
			IsFunded:       isFunded,
			FundedBy:       fundedBy,
			FundingCost:    fundingCost,
			PlayerProgress: playerProgress,
		}
	}
	return dtos
}

// ToAwardResultsDto converts funded awards to placement results
func ToAwardResultsDto(g *game.Game, cardRegistry cards.CardRegistry) []AwardResultDto {
	fundedAwards := g.Awards().FundedAwards()
	results := make([]AwardResultDto, 0, len(fundedAwards))

	for _, funded := range fundedAwards {
		placements := gamecards.ScoreAward(funded.Type, g.GetAllPlayers(), g.Board(), cardRegistry)

		firstPlace := make([]string, 0)
		secondPlace := make([]string, 0)
		for _, p := range placements {
			if p.Placement == 1 {
				firstPlace = append(firstPlace, p.PlayerID)
			} else if p.Placement == 2 {
				secondPlace = append(secondPlace, p.PlayerID)
			}
		}

		results = append(results, AwardResultDto{
			AwardType:      string(funded.Type),
			FirstPlaceIds:  firstPlace,
			SecondPlaceIds: secondPlace,
		})
	}
	return results
}

// ToCardVPConditionDetailDto converts a card VP condition detail to DTO
func ToCardVPConditionDetailDto(detail game.CardVPConditionDetail) CardVPConditionDetailDto {
	return CardVPConditionDetailDto{
		ConditionType:  detail.ConditionType,
		Amount:         detail.Amount,
		Count:          detail.Count,
		MaxTrigger:     detail.MaxTrigger,
		ActualTriggers: detail.ActualTriggers,
		TotalVP:        detail.TotalVP,
		Explanation:    detail.Explanation,
	}
}

// ToCardVPDetailDto converts a card VP detail to DTO
func ToCardVPDetailDto(detail game.CardVPDetail) CardVPDetailDto {
	return CardVPDetailDto{
		CardID:     detail.CardID,
		CardName:   detail.CardName,
		Conditions: mapSlice(detail.Conditions, ToCardVPConditionDetailDto),
		TotalVP:    detail.TotalVP,
	}
}

// ToGreeneryVPDetailDto converts a greenery VP detail to DTO
func ToGreeneryVPDetailDto(detail game.GreeneryVPDetail) GreeneryVPDetailDto {
	return GreeneryVPDetailDto{
		Coordinate: detail.Coordinate,
		VP:         detail.VP,
	}
}

// ToCityVPDetailDto converts a city VP detail to DTO
func ToCityVPDetailDto(detail game.CityVPDetail) CityVPDetailDto {
	return CityVPDetailDto{
		CityCoordinate:     detail.CityCoordinate,
		AdjacentGreeneries: detail.AdjacentGreeneries,
		VP:                 detail.VP,
	}
}

// ToVPBreakdownDto converts a VP breakdown to DTO
func ToVPBreakdownDto(breakdown game.VPBreakdown) VPBreakdownDto {
	return VPBreakdownDto{
		TerraformRating:   breakdown.TerraformRating,
		CardVP:            breakdown.CardVP,
		CardVPDetails:     mapSlice(breakdown.CardVPDetails, ToCardVPDetailDto),
		MilestoneVP:       breakdown.MilestoneVP,
		AwardVP:           breakdown.AwardVP,
		GreeneryVP:        breakdown.GreeneryVP,
		GreeneryVPDetails: mapSlice(breakdown.GreeneryVPDetails, ToGreeneryVPDetailDto),
		CityVP:            breakdown.CityVP,
		CityVPDetails:     mapSlice(breakdown.CityVPDetails, ToCityVPDetailDto),
		TotalVP:           breakdown.TotalVP,
	}
}

// ToFinalScoreDto creates a final score DTO for a player
func ToFinalScoreDto(playerID, playerName string, breakdown game.VPBreakdown, isWinner bool, placement int) FinalScoreDto {
	return FinalScoreDto{
		PlayerID:    playerID,
		PlayerName:  playerName,
		VPBreakdown: ToVPBreakdownDto(breakdown),
		IsWinner:    isWinner,
		Placement:   placement,
	}
}

// toVPGranterDtos converts a slice of VPGranter to VPGranterDto slice with per-condition breakdown
func toVPGranterDtos(granters []player.VPGranter) []VPGranterDto {
	if len(granters) == 0 {
		return []VPGranterDto{}
	}

	dtos := make([]VPGranterDto, len(granters))
	for i, g := range granters {
		conditions := make([]VPGranterConditionDto, len(g.VPConditions))
		for j, cond := range g.VPConditions {
			conditions[j] = toVPGranterConditionDto(cond)
		}
		dtos[i] = VPGranterDto{
			CardID:        g.CardID,
			CardName:      g.CardName,
			Description:   g.Description,
			ComputedValue: g.ComputedValue,
			Conditions:    conditions,
		}
	}
	return dtos
}

func toVPGranterConditionDto(cond player.VPCondition) VPGranterConditionDto {
	dto := VPGranterConditionDto{
		Amount:        cond.Amount,
		ConditionType: string(cond.Condition),
	}

	switch cond.Condition {
	case player.VPConditionFixed, player.VPConditionOnce:
		dto.ComputedVP = cond.Amount
		dto.Explanation = fmt.Sprintf("%d VP", cond.Amount)

	case player.VPConditionPer:
		if cond.Per != nil {
			perType := string(cond.Per.ResourceType)
			if cond.Per.Tag != nil {
				perType = string(*cond.Per.Tag)
			}
			if cond.Per.Target != nil && *cond.Per.Target == "self-card" {
				perType = string(cond.Per.ResourceType)
			}
			dto.PerType = &perType
			dto.PerAmount = &cond.Per.Amount
			dto.Explanation = fmt.Sprintf("%d VP per %d %s", cond.Amount, cond.Per.Amount, perType)
		}
	}

	return dto
}

// ToTriggeredEffectDto converts a triggered effect to DTO
func ToTriggeredEffectDto(effect game.TriggeredEffect) TriggeredEffectDto {
	outputDtos := make([]ResourceConditionDto, len(effect.Outputs))
	for i, output := range effect.Outputs {
		outputDtos[i] = toResourceConditionDto(output)
	}

	var calculatedOutputDtos []CalculatedOutputDto
	if len(effect.CalculatedOutputs) > 0 {
		calculatedOutputDtos = make([]CalculatedOutputDto, len(effect.CalculatedOutputs))
		for i, co := range effect.CalculatedOutputs {
			calculatedOutputDtos[i] = CalculatedOutputDto{
				ResourceType: co.ResourceType,
				Amount:       co.Amount,
				IsScaled:     co.IsScaled,
			}
		}
	}

	var behaviorDtos []CardBehaviorDto
	if len(effect.Behaviors) > 0 {
		behaviorDtos = make([]CardBehaviorDto, len(effect.Behaviors))
		for i, b := range effect.Behaviors {
			behaviorDtos[i] = toCardBehaviorDto(b)
		}
	}

	var vpConditionDtos []VPConditionDto
	if len(effect.VPConditions) > 0 {
		vpConditionDtos = make([]VPConditionDto, len(effect.VPConditions))
		for i, vp := range effect.VPConditions {
			vpConditionDtos[i] = toVPConditionForLogDto(vp)
		}
	}

	return TriggeredEffectDto{
		CardName:          effect.CardName,
		PlayerID:          effect.PlayerID,
		SourceType:        string(effect.SourceType),
		Outputs:           outputDtos,
		CalculatedOutputs: calculatedOutputDtos,
		Behaviors:         behaviorDtos,
		VPConditions:      vpConditionDtos,
	}
}
