package dto

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
)

// ToSpectatorGameDto creates a GameDto for spectators where all players are shown
// as OtherPlayerDto (no hidden information like hand cards or pending selections).
func ToSpectatorGameDto(g *game.Game, cardRegistry cards.CardRegistry) GameDto {
	players := g.GetAllPlayers()

	otherPlayers := make([]OtherPlayerDto, 0, len(players))
	for _, p := range players {
		otherPlayers = append(otherPlayers, ToOtherPlayerDto(p, g, cardRegistry))
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

	tiles := g.Board().Tiles()
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
			tileDtos[i].OccupiedBy = &TileOccupantDto{
				Type: string(tile.OccupiedBy.Type),
				Tags: tile.OccupiedBy.Tags,
			}
		}
	}

	paymentConstants := PaymentConstantsDto{
		SteelValue:    2,
		TitaniumValue: 3,
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
		CurrentPlayer:    PlayerDto{},
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  "",
		IsSpectator:      true,
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
		InitPhase:          initPhaseDto,
		Spectators:         toSpectatorDtos(g),
		ChatMessages:       toChatMessageDtos(g),
	}
}

func convertTileBonusesForSpectator(bonuses []board.TileBonus) []TileBonusDto {
	return convertTileBonuses(bonuses)
}
