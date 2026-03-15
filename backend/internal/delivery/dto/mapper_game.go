package dto

import (
	"fmt"
	"time"

	"slices"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// ToGameDto converts Game to GameDto with personalized view
// The playerID parameter determines which player is "currentPlayer" vs "otherPlayers"
func ToGameDto(g *game.Game, cardRegistry cards.CardRegistry, playerID string, colonyRegistry ...colonies.ColonyRegistry) GameDto {
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
		MaxPlayers:            settings.MaxPlayers,
		VenusNextEnabled:      settings.VenusNextEnabled,
		DevelopmentMode:       settings.DevelopmentMode,
		DemoGame:              settings.DemoGame,
		CardPacks:             settings.CardPacks,
		HasClaudeAPIKey:       settings.ClaudeAPIKey != "",
		ClaudeModel:           settings.ClaudeModel,
		AvailablePlayerColors: shared.PlayerColors,
	}

	globalParams := g.GlobalParameters()
	globalParamsDto := GlobalParametersDto{
		Temperature: globalParams.Temperature(),
		Oxygen:      globalParams.Oxygen(),
		Oceans:      globalParams.Oceans(),
		MaxOceans:   globalParams.GetMaxOceans(),
		Venus:       globalParams.Venus(),
		Bonuses:     buildGlobalParameterBonuses(settings.VenusNextEnabled),
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
	if g.Status() == shared.GameStatusCompleted {
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

	var initPhaseDto *InitPhaseDto
	phase := g.CurrentPhase()
	if phase == shared.GamePhaseInitApplyCorp || phase == shared.GamePhaseInitApplyPrelude {
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

		hasPendingTiles := false
		if currentInitPlayerID != "" {
			hasPendingTiles = g.GetPendingTileSelection(currentInitPlayerID) != nil ||
				g.GetPendingTileSelectionQueue(currentInitPlayerID) != nil
		}

		initPhaseDto = &InitPhaseDto{
			CurrentPlayerID:    currentInitPlayerID,
			CurrentPlayerIndex: idx,
			TotalPlayers:       activePlayers,
			WaitingForConfirm:  g.InitPhaseWaitingForConfirm(),
			ConfirmVersion:     g.InitPhaseConfirmVersion(),
			HasPreludePhase:    g.Settings().HasPrelude(),
			HasPendingTiles:    hasPendingTiles,
		}
	}

	result := GameDto{
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
		PlayerOrder:      g.PlayerOrder(),
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
		InitPhase:          initPhaseDto,
		Spectators:         toSpectatorDtos(g),
		ChatMessages:       toChatMessageDtos(g),
	}

	if g.HasColonies() && len(colonyRegistry) > 0 && colonyRegistry[0] != nil {
		result.ColonyTiles = toColonyTileDtos(g, colonyRegistry[0], playerID)
		result.TradeFleetAvailable = g.GetTradeFleetAvailable(playerID)
	}

	return result
}

func toSpectatorDtos(g *game.Game) []SpectatorDto {
	spectators := g.GetAllSpectators()
	dtos := make([]SpectatorDto, len(spectators))
	for i, s := range spectators {
		dtos[i] = SpectatorDto{
			ID:    s.ID(),
			Name:  s.Name(),
			Color: s.Color(),
		}
	}
	return dtos
}

func toChatMessageDtos(g *game.Game) []ChatMessageDto {
	messages := g.GetChatMessages()
	dtos := make([]ChatMessageDto, len(messages))
	for i, msg := range messages {
		dtos[i] = ChatMessageDto{
			SenderID:    msg.SenderID,
			SenderName:  msg.SenderName,
			SenderColor: msg.SenderColor,
			Message:     msg.Message,
			Timestamp:   msg.Timestamp.Format(time.RFC3339),
			IsSpectator: msg.IsSpectator,
		}
	}
	return dtos
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
func ToCardVPConditionDetailDto(detail shared.CardVPConditionDetail) CardVPConditionDetailDto {
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
func ToCardVPDetailDto(detail shared.CardVPDetail) CardVPDetailDto {
	return CardVPDetailDto{
		CardID:     detail.CardID,
		CardName:   detail.CardName,
		Conditions: mapSlice(detail.Conditions, ToCardVPConditionDetailDto),
		TotalVP:    detail.TotalVP,
	}
}

// ToGreeneryVPDetailDto converts a greenery VP detail to DTO
func ToGreeneryVPDetailDto(detail shared.GreeneryVPDetail) GreeneryVPDetailDto {
	return GreeneryVPDetailDto{
		Coordinate: detail.Coordinate,
		VP:         detail.VP,
	}
}

// ToCityVPDetailDto converts a city VP detail to DTO
func ToCityVPDetailDto(detail shared.CityVPDetail) CityVPDetailDto {
	return CityVPDetailDto{
		CityCoordinate:     detail.CityCoordinate,
		AdjacentGreeneries: detail.AdjacentGreeneries,
		VP:                 detail.VP,
	}
}

// ToVPBreakdownDto converts a VP breakdown to DTO
func ToVPBreakdownDto(breakdown shared.VPBreakdown) VPBreakdownDto {
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
func ToFinalScoreDto(playerID, playerName string, breakdown shared.VPBreakdown, isWinner bool, placement int) FinalScoreDto {
	return FinalScoreDto{
		PlayerID:    playerID,
		PlayerName:  playerName,
		VPBreakdown: ToVPBreakdownDto(breakdown),
		IsWinner:    isWinner,
		Placement:   placement,
	}
}

// toVPGranterDtos converts a slice of VPGranter to VPGranterDto slice with per-condition breakdown
func toVPGranterDtos(granters []shared.VPGranter) []VPGranterDto {
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

func toVPGranterConditionDto(cond shared.VPCondition) VPGranterConditionDto {
	dto := VPGranterConditionDto{
		Amount:        cond.Amount,
		ConditionType: string(cond.Condition),
	}

	switch cond.Condition {
	case shared.VPConditionFixed, shared.VPConditionOnce:
		dto.ComputedVP = cond.Amount
		dto.Explanation = fmt.Sprintf("%d VP", cond.Amount)

	case shared.VPConditionPer:
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
			dto.AdjacentToSelfTile = cond.Per.AdjacentToSelfTile
			dto.Explanation = fmt.Sprintf("%d VP per %d %s", cond.Amount, cond.Per.Amount, perType)
		}
	}

	return dto
}

// ToTriggeredEffectDto converts a triggered effect to DTO
func ToTriggeredEffectDto(effect shared.TriggeredEffect) TriggeredEffectDto {
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
			vpConditionDtos[i] = VPConditionDto{
				Amount:     vp.Amount,
				Condition:  VPConditionType(vp.Condition),
				MaxTrigger: vp.MaxTrigger,
				Per:        ptrCast(vp.Per, toPerConditionDto),
			}
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

func buildGlobalParameterBonuses(venusEnabled bool) []GlobalParameterBonusDto {
	bonuses := []GlobalParameterBonusDto{
		{Parameter: "temperature", Threshold: -24, RewardType: "heat-production", RewardAmount: 1},
		{Parameter: "temperature", Threshold: -20, RewardType: "heat-production", RewardAmount: 1},
		{Parameter: "temperature", Threshold: 0, RewardType: "ocean-placement", RewardAmount: 1},
		{Parameter: "oxygen", Threshold: 8, RewardType: "temperature", RewardAmount: 1},
	}
	if venusEnabled {
		bonuses = append(bonuses,
			GlobalParameterBonusDto{Parameter: "venus", Threshold: 8, RewardType: "card-draw", RewardAmount: 1},
			GlobalParameterBonusDto{Parameter: "venus", Threshold: 16, RewardType: "tr", RewardAmount: 1},
		)
	}
	return bonuses
}

func toColonyTileDtos(g *game.Game, colonyRegistry colonies.ColonyRegistry, playerID string) []ColonyTileDto {
	tileStates := g.ColonyTileStates()
	if len(tileStates) == 0 {
		return nil
	}

	dtos := make([]ColonyTileDto, 0, len(tileStates))
	for _, state := range tileStates {
		def, err := colonyRegistry.GetByID(state.DefinitionID)
		if err != nil {
			continue
		}

		steps := make([]ColonyStepDto, len(def.Steps))
		for i, s := range def.Steps {
			outputs := make([]ColonyOutputDto, len(s.Outputs))
			for j, o := range s.Outputs {
				outputs[j] = ColonyOutputDto{Type: o.Type, Amount: o.Amount}
			}
			steps[i] = ColonyStepDto{Outputs: outputs}
		}

		colonyBonus := make([]ColonyOutputDto, len(def.ColonyBonus))
		for i, b := range def.ColonyBonus {
			colonyBonus[i] = ColonyOutputDto{Type: b.Type, Amount: b.Amount}
		}

		colonySlots := make([]ColonySlotDto, len(def.Colonies))
		for i, c := range def.Colonies {
			reward := make([]ColonyOutputDto, len(c.Reward))
			for j, r := range c.Reward {
				reward[j] = ColonyOutputDto{Type: r.Type, Amount: r.Amount}
			}
			colonySlots[i] = ColonySlotDto{Reward: reward}
		}

		playerColonies := state.PlayerColonies
		if playerColonies == nil {
			playerColonies = []string{}
		}

		// Calculate trade availability
		tradeAvailable := true
		var tradeErrors []StateErrorDto
		if state.TradedThisGen {
			tradeAvailable = false
			tradeErrors = append(tradeErrors, StateErrorDto{
				Code:    StateErrorCode("colony-already-traded"),
				Message: "This colony has already been traded this generation",
			})
		}
		if !g.GetTradeFleetAvailable(playerID) {
			tradeAvailable = false
			tradeErrors = append(tradeErrors, StateErrorDto{
				Code:    StateErrorCode("fleet-unavailable"),
				Message: "Your trade fleet is not available",
			})
		}
		playerObj, _ := g.GetPlayer(playerID)
		if playerObj != nil {
			resources := playerObj.Resources().Get()
			canAffordAny := resources.Credits >= 9 || resources.Energy >= 3 || resources.Titanium >= 3
			if !canAffordAny {
				tradeAvailable = false
				tradeErrors = append(tradeErrors, StateErrorDto{
					Code:    StateErrorCode("insufficient-resources"),
					Message: "Cannot afford trade: need 9 MC, 3 energy, or 3 titanium",
				})
			}
		}

		// Calculate build availability
		buildAvailable := true
		var buildErrors []StateErrorDto
		maxColonies := len(def.Colonies)
		if len(state.PlayerColonies) >= maxColonies {
			buildAvailable = false
			buildErrors = append(buildErrors, StateErrorDto{
				Code:    StateErrorCode("colony-full"),
				Message: "This colony tile is full",
			})
		}
		if slices.Contains(state.PlayerColonies, playerID) {
			buildAvailable = false
			buildErrors = append(buildErrors, StateErrorDto{
				Code:    StateErrorCode("already-has-colony"),
				Message: "You already have a colony here",
			})
		}
		if playerObj != nil {
			resources := playerObj.Resources().Get()
			if resources.Credits < 17 {
				buildAvailable = false
				buildErrors = append(buildErrors, StateErrorDto{
					Code:    StateErrorCode("insufficient-credits"),
					Message: fmt.Sprintf("Insufficient credits: need 17, have %d", resources.Credits),
				})
			}
		}

		dtos = append(dtos, ColonyTileDto{
			ID:             def.ID,
			Name:           def.Name,
			Steps:          steps,
			ColonyBonus:    colonyBonus,
			Colonies:       colonySlots,
			MarkerPosition: state.MarkerPosition,
			PlayerColonies: playerColonies,
			TradedThisGen:  state.TradedThisGen,
			TraderID:       state.TraderID,
			Style: ColonyStyleDto{
				Color: def.Style.Color,
				Icon:  def.Style.Icon,
			},
			TradeAvailable: tradeAvailable,
			BuildAvailable: buildAvailable,
			TradeErrors:    tradeErrors,
			BuildErrors:    buildErrors,
		})
	}

	return dtos
}
