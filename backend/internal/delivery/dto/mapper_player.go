package dto

import (
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/milestones"
	"terraforming-mars-backend/internal/standardprojects"
)

// toResourcesDto converts shared.Resources to ResourcesDto.
func toResourcesDto(res shared.Resources) ResourcesDto {
	return ResourcesDto{
		Credits:  res.Credits,
		Steel:    res.Steel,
		Titanium: res.Titanium,
		Plants:   res.Plants,
		Energy:   res.Energy,
		Heat:     res.Heat,
	}
}

// toProductionDto converts shared.Production to ProductionDto.
func toProductionDto(prod shared.Production) ProductionDto {
	return ProductionDto{
		Credits:  prod.Credits,
		Steel:    prod.Steel,
		Titanium: prod.Titanium,
		Plants:   prod.Plants,
		Energy:   prod.Energy,
		Heat:     prod.Heat,
	}
}

// calculateResourceDelta calculates the difference between two Resources and returns a ResourcesDto.
func calculateResourceDelta(before, after shared.Resources) ResourcesDto {
	return ResourcesDto{
		Credits:  after.Credits - before.Credits,
		Steel:    after.Steel - before.Steel,
		Titanium: after.Titanium - before.Titanium,
		Plants:   after.Plants - before.Plants,
		Energy:   after.Energy - before.Energy,
		Heat:     after.Heat - before.Heat,
	}
}

// ToPlayerDto converts Player to PlayerDto
func ToPlayerDto(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry, stdProjRegistry standardprojects.StandardProjectRegistry, awardRegistry awards.AwardRegistry, milestoneRegistry milestones.MilestoneRegistry) PlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	corporation := getCorporationCard(p, cardRegistry)
	playedCardIDs := p.PlayedCards().Cards()
	playedCards := getPlayedCards(playedCardIDs, cardRegistry)
	handCards := mapPlayerCards(p, g, cardRegistry)
	standardProjects := mapPlayerStandardProjects(p, g, cardRegistry, stdProjRegistry)
	milestones := mapPlayerMilestones(p, g, cardRegistry, milestoneRegistry)
	awards := mapPlayerAwards(p, g, awardRegistry)
	var pendingTileSelection *PendingTileSelectionDto
	var forcedFirstAction *ForcedFirstActionDto
	currentTurn := g.CurrentTurn()
	if currentTurn != nil && currentTurn.PlayerID() == p.ID() {
		pendingTileSelection = convertPendingTileSelection(g.GetPendingTileSelection(p.ID()))
		forcedFirstAction = convertForcedFirstAction(g.GetForcedFirstAction(p.ID()))
	}
	if g.CurrentPhase() == shared.GamePhaseStartingSelection ||
		g.CurrentPhase() == shared.GamePhaseInitApplyCorp ||
		g.CurrentPhase() == shared.GamePhaseInitApplyPrelude {
		pendingTileSelection = convertPendingTileSelection(g.GetPendingTileSelection(p.ID()))
		forcedFirstAction = convertForcedFirstAction(g.GetForcedFirstAction(p.ID()))
	}

	return PlayerDto{
		ID:               p.ID(),
		Name:             p.Name(),
		PlayerType:       string(p.PlayerType()),
		BotStatus:        string(p.BotStatus()),
		BotDifficulty:    string(p.BotDifficulty()),
		BotSpeed:         string(p.BotSpeed()),
		Color:            p.Color(),
		Resources:        toResourcesDto(resources),
		Production:       toProductionDto(production),
		TerraformRating:  resourcesComponent.TerraformRating(),
		Status:           playerStatus(p, g),
		Corporation:      corporation,
		Cards:            handCards, // PlayerCardDto[] with state
		PlayedCards:      playedCards,
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		TotalActions:     getTotalActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		IsExited:         p.HasExited(),
		Effects:          convertPlayerEffects(p.Effects().List(), p, g, cardRegistry),
		Actions:          convertPlayerActions(p.Actions().List(), p, g, cardRegistry),
		StandardProjects: standardProjects, // PlayerStandardProjectDto[] with state
		Milestones:       milestones,       // PlayerMilestoneDto[] with eligibility
		Awards:           awards,           // PlayerAwardDto[] with eligibility

		SelectCorporationPhase:         convertSelectCorporationPhase(g.GetSelectCorporationPhase(p.ID()), cardRegistry),
		SelectStartingCardsPhase:       convertSelectStartingCardsPhase(g.GetSelectStartingCardsPhase(p.ID()), cardRegistry),
		SelectPreludeCardsPhase:        convertSelectPreludeCardsPhase(g.GetSelectPreludeCardsPhase(p.ID()), cardRegistry),
		ProductionPhase:                convertProductionPhase(g.GetProductionPhase(p.ID()), cardRegistry),
		StartingCards:                  []CardDto{},
		PendingTileSelection:           pendingTileSelection,
		PendingCardSelection:           convertPendingCardSelection(p.Selection().GetPendingCardSelection(), p, g, cardRegistry),
		PendingCardDrawSelection:       convertPendingCardDrawSelection(p.Selection().GetPendingCardDrawSelection(), p, g, cardRegistry),
		PendingCardDiscardSelection:    convertPendingCardDiscardSelection(p.Selection().GetPendingCardDiscardSelection()),
		PendingBehaviorChoiceSelection: convertPendingBehaviorChoiceSelection(p.Selection().GetPendingBehaviorChoiceSelection(), p, g, cardRegistry),
		PendingStealTargetSelection:    convertPendingStealTargetSelection(p.Selection().GetPendingStealTargetSelection()),
		PendingColonyResourceSelection: convertPendingColonyResourceFromQueue(p.Selection().GetPendingColonyResourceQueue()),
		PendingAwardFundSelection:      convertPendingAwardFundSelection(p.Selection().GetPendingAwardFundSelection()),
		PendingColonySelection:         convertPendingColonySelection(p.Selection().GetPendingColonySelection()),
		PendingFreeTradeSelection:      convertPendingFreeTradeSelection(p.Selection().GetPendingFreeTradeSelection()),
		ForcedFirstAction:              forcedFirstAction,
		ResourceStorage:                p.Resources().Storage(),
		PaymentSubstitutes:             convertPaymentSubstitutes(p.Resources().PaymentSubstitutes()),
		StoragePaymentSubstitutes:      convertStoragePaymentSubstitutes(p.Resources().StoragePaymentSubstitutes()),
		GenerationalEvents:             convertGenerationalEvents(p.GenerationalEvents().GetAll()),
		VPGranters:                     toVPGranterDtos(p.VPGranters().GetAll()),
		BonusTags:                      convertBonusTags(p.BonusTags()),
		ActionCosts:                    mapPlayerActionCosts(p, g, cardRegistry),
	}
}

// ToOtherPlayerDto converts Player to OtherPlayerDto
func ToOtherPlayerDto(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) OtherPlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	corporation := getCorporationCard(p, cardRegistry)
	playedCardIDs := p.PlayedCards().Cards()
	playedCards := getPlayedCards(playedCardIDs, cardRegistry)
	handCardCount := len(p.Hand().Cards())

	return OtherPlayerDto{
		ID:               p.ID(),
		Name:             p.Name(),
		PlayerType:       string(p.PlayerType()),
		BotStatus:        string(p.BotStatus()),
		BotDifficulty:    string(p.BotDifficulty()),
		BotSpeed:         string(p.BotSpeed()),
		Color:            p.Color(),
		Resources:        toResourcesDto(resources),
		Production:       toProductionDto(production),
		TerraformRating:  resourcesComponent.TerraformRating(),
		Status:           playerStatus(p, g),
		Corporation:      corporation,
		HandCardCount:    handCardCount,
		PlayedCards:      playedCards,
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		TotalActions:     getTotalActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		IsExited:         p.HasExited(),
		Effects:          convertPlayerEffects(p.Effects().List(), p, g, cardRegistry),
		Actions:          convertPlayerActions(p.Actions().List(), p, g, cardRegistry),

		SelectCorporationPhase:    convertSelectCorporationPhaseForOtherPlayer(g.GetSelectCorporationPhase(p.ID())),
		SelectStartingCardsPhase:  convertSelectStartingCardsPhaseForOtherPlayer(g.GetSelectStartingCardsPhase(p.ID())),
		SelectPreludeCardsPhase:   convertSelectPreludeCardsPhaseForOtherPlayer(g.GetSelectPreludeCardsPhase(p.ID())),
		ProductionPhase:           convertProductionPhaseForOtherPlayer(g.GetProductionPhase(p.ID())),
		ResourceStorage:           p.Resources().Storage(),
		PaymentSubstitutes:        convertPaymentSubstitutes(p.Resources().PaymentSubstitutes()),
		StoragePaymentSubstitutes: convertStoragePaymentSubstitutes(p.Resources().StoragePaymentSubstitutes()),
		VPGranters:                toVPGranterDtos(p.VPGranters().GetAll()),
		BonusTags:                 convertBonusTags(p.BonusTags()),
	}
}

func convertBonusTags(tags map[shared.CardTag]int) map[string]int {
	if len(tags) == 0 {
		return map[string]int{}
	}
	result := make(map[string]int, len(tags))
	for k, v := range tags {
		result[string(k)] = v
	}
	return result
}

func playerStatus(p *player.Player, g *game.Game) PlayerStatus {
	if p.HasExited() {
		return PlayerStatusExited
	}
	if g.HasAnyPendingSelection(p.ID()) {
		return PlayerStatusSelection
	}
	return PlayerStatusWaiting
}

func convertSelectCorporationPhase(phase *shared.SelectCorporationPhase, cardRegistry cards.CardRegistry) *SelectCorporationPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectCorporationPhaseDto{
		AvailableCorporations: getPlayedCards(phase.AvailableCorporations, cardRegistry),
	}
}

func convertSelectCorporationPhaseForOtherPlayer(phase *shared.SelectCorporationPhase) *SelectCorporationOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectCorporationOtherPlayerDto{}
}

func convertSelectStartingCardsPhase(phase *shared.SelectStartingCardsPhase, cardRegistry cards.CardRegistry) *SelectStartingCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsPhaseDto{
		AvailableCards: getPlayedCards(phase.AvailableCards, cardRegistry),
	}
}

func convertSelectStartingCardsPhaseForOtherPlayer(phase *shared.SelectStartingCardsPhase) *SelectStartingCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsOtherPlayerDto{}
}

// convertSelectPreludeCardsPhase converts SelectPreludeCardsPhase to DTO
func convertSelectPreludeCardsPhase(phase *shared.SelectPreludeCardsPhase, cardRegistry cards.CardRegistry) *SelectPreludeCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectPreludeCardsPhaseDto{
		AvailablePreludes: getPlayedCards(phase.AvailablePreludes, cardRegistry),
		MaxSelectable:     phase.MaxSelectable,
	}
}

// convertSelectPreludeCardsPhaseForOtherPlayer converts SelectPreludeCardsPhase to DTO for other players
func convertSelectPreludeCardsPhaseForOtherPlayer(phase *shared.SelectPreludeCardsPhase) *SelectPreludeCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectPreludeCardsOtherPlayerDto{}
}

// convertProductionPhase converts production phase data to DTO for current player
func convertProductionPhase(phase *shared.ProductionPhase, cardRegistry cards.CardRegistry) *ProductionPhaseDto {
	if phase == nil {
		return nil
	}

	return &ProductionPhaseDto{
		AvailableCards:    getPlayedCards(phase.AvailableCards, cardRegistry),
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   toResourcesDto(phase.BeforeResources),
		AfterResources:    toResourcesDto(phase.AfterResources),
		ResourceDelta:     calculateResourceDelta(phase.BeforeResources, phase.AfterResources),
		EnergyConverted:   phase.EnergyConverted,
		CreditsIncome:     phase.CreditsIncome,
	}
}

// convertProductionPhaseForOtherPlayer converts production phase data to DTO for other players
func convertProductionPhaseForOtherPlayer(phase *shared.ProductionPhase) *ProductionPhaseOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &ProductionPhaseOtherPlayerDto{
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   toResourcesDto(phase.BeforeResources),
		AfterResources:    toResourcesDto(phase.AfterResources),
		ResourceDelta:     calculateResourceDelta(phase.BeforeResources, phase.AfterResources),
		EnergyConverted:   phase.EnergyConverted,
		CreditsIncome:     phase.CreditsIncome,
	}
}

// convertPlayerEffects converts CardEffect slice to PlayerEffectDto slice
func convertPlayerEffects(effects []shared.CardEffect, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []PlayerEffectDto {
	if len(effects) == 0 {
		return []PlayerEffectDto{}
	}

	board := g.Board()
	allPlayers := g.GetAllPlayers()

	dtos := make([]PlayerEffectDto, len(effects))
	for i, effect := range effects {
		var computedValues []ComputedBehaviorValueDto
		var outputs []CalculatedOutputDto
		for _, outputBC := range effect.Behavior.Outputs {
			per := shared.GetPerCondition(outputBC)
			if per == nil {
				continue
			}
			count := gamecards.CountPerCondition(per, effect.CardID, p, board, cardRegistry, allPlayers)
			if per.Amount > 0 {
				multiplier := count / per.Amount
				actualAmount := outputBC.GetAmount() * multiplier
				outputs = append(outputs, CalculatedOutputDto{
					ResourceType: string(outputBC.GetResourceType()),
					Amount:       actualAmount,
					IsScaled:     true,
				})
			}
		}
		if len(outputs) > 0 {
			computedValues = []ComputedBehaviorValueDto{{
				Target:  "behaviors::0",
				Outputs: outputs,
			}}
		}
		dtos[i] = PlayerEffectDto{
			CardID:         effect.CardID,
			CardName:       effect.CardName,
			BehaviorIndex:  effect.BehaviorIndex,
			Behavior:       toCardBehaviorDto(effect.Behavior),
			ComputedValues: computedValues,
		}
	}
	return dtos
}

// convertPlayerActions converts CardAction slice to PlayerActionDto slice
// Calculates state for each action to determine availability and errors.
func convertPlayerActions(actions []shared.CardAction, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []PlayerActionDto {
	if len(actions) == 0 {
		return []PlayerActionDto{}
	}

	dtos := make([]PlayerActionDto, len(actions))
	for i, act := range actions {
		// Calculate state for this action
		state := action.CalculatePlayerCardActionState(
			act.CardID,
			act.Behavior,
			act.TimesUsedThisGeneration,
			p,
			g,
			cardRegistry,
		)

		behaviorDto := toCardBehaviorDto(act.Behavior)

		if act.Behavior.ChoicePolicy != nil && len(behaviorDto.Choices) > 0 {
			production := p.Resources().Production()
			validIndices := shared.FilterChoiceIndicesByPolicy(act.Behavior.Choices, act.Behavior.ChoicePolicy, production)
			filtered := make([]ChoiceDto, 0, len(validIndices))
			for _, idx := range validIndices {
				choice := behaviorDto.Choices[idx]
				choice.OriginalIndex = idx
				filtered = append(filtered, choice)
			}
			behaviorDto.Choices = filtered
		}

		dtos[i] = PlayerActionDto{
			CardID:                  act.CardID,
			CardName:                act.CardName,
			BehaviorIndex:           act.BehaviorIndex,
			Behavior:                behaviorDto,
			TimesUsedThisTurn:       act.TimesUsedThisTurn,
			TimesUsedThisGeneration: act.TimesUsedThisGeneration,
			Available:               state.Available(),
			Errors:                  convertStateErrors(state.Errors),
			Warnings:                convertStateWarnings(state.Warnings),
			ComputedValues:          convertComputedValues(state.ComputedValues),
		}
	}
	return dtos
}

// convertComputedValues converts ComputedBehaviorValue slice to DTO slice
func convertComputedValues(values []player.ComputedBehaviorValue) []ComputedBehaviorValueDto {
	if len(values) == 0 {
		return nil
	}

	dtos := make([]ComputedBehaviorValueDto, len(values))
	for i, v := range values {
		outputs := make([]CalculatedOutputDto, len(v.Outputs))
		for j, o := range v.Outputs {
			outputs[j] = CalculatedOutputDto{
				ResourceType: o.ResourceType,
				Amount:       o.Amount,
				IsScaled:     o.IsScaled,
			}
		}
		dtos[i] = ComputedBehaviorValueDto{
			Target:  v.Target,
			Outputs: outputs,
		}
	}
	return dtos
}

// convertPaymentSubstitutes converts PaymentSubstitute slice to PaymentSubstituteDto slice
func convertPaymentSubstitutes(substitutes []shared.PaymentSubstitute) []PaymentSubstituteDto {
	if len(substitutes) == 0 {
		return []PaymentSubstituteDto{}
	}

	dtos := make([]PaymentSubstituteDto, len(substitutes))
	for i, sub := range substitutes {
		dtos[i] = PaymentSubstituteDto{
			ResourceType:   ResourceType(sub.ResourceType),
			ConversionRate: sub.ConversionRate,
		}
	}
	return dtos
}

// convertStoragePaymentSubstitutes converts StoragePaymentSubstitute slice to DTO slice
func convertStoragePaymentSubstitutes(substitutes []shared.StoragePaymentSubstitute) []StoragePaymentSubstituteDto {
	if len(substitutes) == 0 {
		return []StoragePaymentSubstituteDto{}
	}

	dtos := make([]StoragePaymentSubstituteDto, len(substitutes))
	for i, sub := range substitutes {
		dtos[i] = StoragePaymentSubstituteDto{
			CardID:         sub.CardID,
			ResourceType:   ResourceType(sub.ResourceType),
			ConversionRate: sub.ConversionRate,
			Selectors:      mapSlice(sub.Selectors, toSelectorDto),
		}
	}
	return dtos
}

// convertPendingCardSelection converts PendingCardSelection to DTO with playability state
func convertPendingCardSelection(selection *shared.PendingCardSelection, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) *PendingCardSelectionDto {
	if selection == nil {
		return nil
	}

	availableCards := make([]PlayerCardDto, 0, len(selection.AvailableCards))
	for _, cardID := range selection.AvailableCards {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}
		state := action.CalculatePendingCardPlayability(card, p, g, cardRegistry)
		availableCards = append(availableCards, ToPlayerCardDto(card, state))
	}

	return &PendingCardSelectionDto{
		AvailableCards: availableCards,
		CardCosts:      selection.CardCosts,
		CardRewards:    selection.CardRewards,
		Source:         selection.Source,
		MinCards:       selection.MinCards,
		MaxCards:       selection.MaxCards,
	}
}

// convertPendingCardDrawSelection converts PendingCardDrawSelection to DTO with playability state
func convertPendingCardDrawSelection(selection *shared.PendingCardDrawSelection, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) *PendingCardDrawSelectionDto {
	if selection == nil {
		return nil
	}

	availableCards := make([]PlayerCardDto, 0, len(selection.AvailableCards))
	for _, cardID := range selection.AvailableCards {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}
		state := action.CalculatePendingCardPlayability(card, p, g, cardRegistry)
		availableCards = append(availableCards, ToPlayerCardDto(card, state))
	}

	return &PendingCardDrawSelectionDto{
		AvailableCards: availableCards,
		FreeTakeCount:  selection.FreeTakeCount,
		MaxBuyCount:    selection.MaxBuyCount,
		CardBuyCost:    selection.CardBuyCost,
		Source:         selection.Source,
		PlayAsPrelude:  selection.PlayAsPrelude,
	}
}

// convertPendingCardDiscardSelection converts PendingCardDiscardSelection to DTO
func convertPendingCardDiscardSelection(selection *shared.PendingCardDiscardSelection) *PendingCardDiscardSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingCardDiscardSelectionDto{
		MinCards:     selection.MinCards,
		MaxCards:     selection.MaxCards,
		Source:       selection.Source,
		SourceCardID: selection.SourceCardID,
	}
}

func convertPendingBehaviorChoiceSelection(selection *shared.PendingBehaviorChoiceSelection, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) *PendingBehaviorChoiceSelectionDto {
	if selection == nil {
		return nil
	}

	choices := make([]ChoiceDto, len(selection.Choices))
	for i, choice := range selection.Choices {
		choices[i] = toChoiceDtoWithState(i, choice, p, g, cardRegistry)
	}

	return &PendingBehaviorChoiceSelectionDto{
		Choices:      choices,
		Source:       selection.Source,
		SourceCardID: selection.SourceCardID,
	}
}

func convertPendingStealTargetSelection(selection *shared.PendingStealTargetSelection) *PendingStealTargetSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingStealTargetSelectionDto{
		EligiblePlayerIDs: selection.EligiblePlayerIDs,
		ResourceType:      string(selection.ResourceType),
		Amount:            selection.Amount,
		Source:            selection.Source,
		SourceCardID:      selection.SourceCardID,
	}
}

func convertPendingColonyResourceFromQueue(queue []shared.PendingColonyResourceSelection) *PendingColonyResourceSelectionDto {
	if len(queue) == 0 {
		return nil
	}
	selection := queue[0]
	return &PendingColonyResourceSelectionDto{
		ResourceType: selection.ResourceType,
		Amount:       selection.Amount,
		Source:       selection.Source,
		ColonyID:     selection.ColonyID,
		Reason:       ColonyResourceReason(selection.Reason),
	}
}

// toChoiceDtoWithState maps a choice to DTO with computed errors from the state calculator.
func toChoiceDtoWithState(index int, choice shared.Choice, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) ChoiceDto {
	errors := action.CalculateChoiceErrors(choice, p, g, cardRegistry)
	return ChoiceDto{
		OriginalIndex: index,
		Inputs:        mapSlice(choice.Inputs, toResourceConditionDto),
		Outputs:       mapSlice(choice.Outputs, toResourceConditionDto),
		Requirements:  toChoiceRequirementsDto(choice.Requirements),
		Available:     len(errors) == 0,
		Errors:        convertStateErrors(errors),
	}
}

func convertPendingAwardFundSelection(selection *shared.PendingAwardFundSelection) *PendingAwardFundSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingAwardFundSelectionDto{
		AvailableAwards: selection.AvailableAwards,
		Source:          selection.Source,
	}
}

func convertPendingColonySelection(selection *shared.PendingColonySelection) *PendingColonySelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingColonySelectionDto{
		AvailableColonyIDs:         selection.AvailableColonyIDs,
		AllowDuplicatePlayerColony: selection.AllowDuplicatePlayerColony,
		Source:                     selection.Source,
		SourceCardID:               selection.SourceCardID,
	}
}

func convertPendingFreeTradeSelection(selection *shared.PendingFreeTradeSelection) *PendingFreeTradeSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingFreeTradeSelectionDto{
		AvailableColonyIDs: selection.AvailableColonyIDs,
		Source:             selection.Source,
		SourceCardID:       selection.SourceCardID,
	}
}

// convertForcedFirstAction converts ForcedFirstAction to DTO
func convertForcedFirstAction(action *shared.ForcedFirstAction) *ForcedFirstActionDto {
	if action == nil {
		return nil
	}

	return &ForcedFirstActionDto{
		ActionType:    action.ActionType,
		CorporationID: action.CorporationID,
		Completed:     action.Completed,
		Description:   action.Description,
	}
}

// convertPendingTileSelection converts PendingTileSelection to DTO
func convertPendingTileSelection(selection *shared.PendingTileSelection) *PendingTileSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingTileSelectionDto{
		TileType:       selection.TileType,
		AvailableHexes: selection.AvailableHexes,
		Source:         selection.Source,
	}
}

// getAvailableActionsForPlayer returns the available actions for a player
// Actions are now at game level, so only the current player has actions
func getAvailableActionsForPlayer(g *game.Game, playerID string) int {
	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		return 0
	}

	if currentTurn.PlayerID() == playerID {
		return currentTurn.ActionsRemaining()
	}

	return 0
}

// getTotalActionsForPlayer returns the total actions granted this turn
func getTotalActionsForPlayer(g *game.Game, playerID string) int {
	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		return 0
	}

	if currentTurn.PlayerID() == playerID {
		return currentTurn.TotalActions()
	}

	return 0
}

// convertStateErrors converts EntityState errors to DTOs.
// Since domain and DTO enums have identical string values, we cast between them.
func convertStateErrors(errors []player.StateError) []StateErrorDto {
	result := make([]StateErrorDto, len(errors))
	for i, err := range errors {
		result[i] = StateErrorDto{
			Code:     StateErrorCode(err.Code),
			Category: StateErrorCategory(err.Category),
			Message:  err.Message,
		}
	}
	return result
}

// convertStateWarnings converts EntityState warnings to DTOs.
// Since domain and DTO enums have identical string values, we cast between them.
func convertStateWarnings(warnings []player.StateWarning) []StateWarningDto {
	if len(warnings) == 0 {
		return nil
	}
	result := make([]StateWarningDto, len(warnings))
	for i, warn := range warnings {
		result[i] = StateWarningDto{
			Code:    StateWarningCode(warn.Code),
			Message: warn.Message,
		}
	}
	return result
}

// ToPlayerCardDto converts a card and its computed state to a PlayerCardDto.
func ToPlayerCardDto(card *gamecards.Card, state player.EntityState) PlayerCardDto {
	discounts := make(map[string]int)
	if discountData, ok := state.Metadata["discounts"].(map[string]int); ok {
		discounts = discountData
	}

	tags := make([]CardTag, len(card.Tags))
	for i, tag := range card.Tags {
		tags[i] = CardTag(tag)
	}

	requirements := toCardRequirementsDto(card.Requirements)

	var behaviors []CardBehaviorDto
	if len(card.Behaviors) > 0 {
		behaviors = make([]CardBehaviorDto, len(card.Behaviors))
		for i, behavior := range card.Behaviors {
			behaviors[i] = toCardBehaviorDto(behavior)
		}
	}

	var resourceStorage *ResourceStorageDto
	if card.ResourceStorage != nil {
		storage := toResourceStorageDto(*card.ResourceStorage)
		resourceStorage = &storage
	}

	var vpConditions []VPConditionDto
	if len(card.VPConditions) > 0 {
		vpConditions = make([]VPConditionDto, len(card.VPConditions))
		for i, vp := range card.VPConditions {
			vpConditions[i] = toVPConditionDto(vp)
		}
	}

	effectiveCost := 0
	if state.Cost != nil {
		if credits, ok := state.Cost[string(shared.ResourceCredit)]; ok {
			effectiveCost = credits
		}
	}

	return PlayerCardDto{
		ID:              card.ID,
		Name:            card.Name,
		Type:            CardType(card.Type),
		Cost:            card.Cost,
		Description:     card.Description,
		Pack:            card.Pack,
		Tags:            tags,
		Requirements:    requirements,
		Behaviors:       behaviors,
		ResourceStorage: resourceStorage,
		VPConditions:    vpConditions,
		Available:       state.Available(),
		Errors:          convertStateErrors(state.Errors),
		Warnings:        convertStateWarnings(state.Warnings),
		EffectiveCost:   effectiveCost,
		Discounts:       discounts,
		ComputedValues:  convertComputedValues(state.ComputedValues),
	}
}

// mapPlayerCards converts hand cards to DTOs using CardStateStore for cached state.
// Enriches behavior choices with computed errors from the state calculator.
func mapPlayerCards(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []PlayerCardDto {
	handCardIDs := p.Hand().Cards()
	result := make([]PlayerCardDto, 0, len(handCardIDs))

	for _, cardID := range handCardIDs {
		state, exists := p.CardStateStore().GetState(cardID)
		if !exists {
			continue
		}
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}

		dto := ToPlayerCardDto(card, state)

		// Enrich choices with computed errors and apply choice policy filtering
		for bi, behavior := range card.Behaviors {
			if bi < len(dto.Behaviors) {
				if behavior.ChoicePolicy != nil && len(dto.Behaviors[bi].Choices) > 0 {
					production := p.Resources().Production()
					validIndices := shared.FilterChoiceIndicesByPolicy(behavior.Choices, behavior.ChoicePolicy, production)
					filtered := make([]ChoiceDto, 0, len(validIndices))
					for _, idx := range validIndices {
						choiceDto := dto.Behaviors[bi].Choices[idx]
						choiceDto.OriginalIndex = idx
						choiceErrors := action.CalculateChoiceErrors(behavior.Choices[idx], p, g, cardRegistry)
						choiceDto.Available = len(choiceErrors) == 0
						choiceDto.Errors = convertStateErrors(choiceErrors)
						filtered = append(filtered, choiceDto)
					}
					dto.Behaviors[bi].Choices = filtered
				} else {
					for ci, choice := range behavior.Choices {
						if ci < len(dto.Behaviors[bi].Choices) {
							choiceErrors := action.CalculateChoiceErrors(choice, p, g, cardRegistry)
							dto.Behaviors[bi].Choices[ci].Available = len(choiceErrors) == 0
							dto.Behaviors[bi].Choices[ci].Errors = convertStateErrors(choiceErrors)
						}
					}
				}
			}
		}

		result = append(result, dto)
	}

	return result
}

// mapPlayerStandardProjects calculates state for all standard projects and converts to DTOs.
// Uses the state calculator to compute availability and effective costs for each project.
// NOTE: Conversion projects (plants→greenery, heat→temperature) are NOT included here - they
// are handled separately via resource buttons in the bottom bar.
func mapPlayerStandardProjects(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry, stdProjRegistry standardprojects.StandardProjectRegistry) []PlayerStandardProjectDto {
	if stdProjRegistry == nil {
		return nil
	}
	allDefinitions := stdProjRegistry.GetAll()
	settings := g.Settings()
	enabledPacks := make(map[string]bool, len(settings.CardPacks))
	for _, pack := range settings.CardPacks {
		enabledPacks[pack] = true
	}
	if settings.VenusNextEnabled {
		enabledPacks[shared.PackVenus] = true
	}

	result := make([]PlayerStandardProjectDto, 0, len(allDefinitions))

	for _, def := range allDefinitions {
		if def.Pack != "" && !enabledPacks[def.Pack] {
			continue
		}
		projectType := shared.StandardProject(def.ID)
		state := action.CalculatePlayerStandardProjectState(projectType, p, g, cardRegistry)

		baseCost := map[string]int{}
		if creditCost := def.CreditCost(); creditCost > 0 {
			baseCost[string(shared.ResourceCredit)] = creditCost
		}

		discounts := make(map[string]int)
		if discountData, ok := state.Metadata["discounts"].(map[string]int); ok {
			discounts = discountData
		}

		behaviorDtos := make([]CardBehaviorDto, 0, len(def.Behaviors))
		for _, b := range def.Behaviors {
			behaviorDtos = append(behaviorDtos, toCardBehaviorDto(b))
		}

		dto := PlayerStandardProjectDto{
			ProjectType:   def.ID,
			Name:          def.Name,
			Description:   def.Description,
			Behaviors:     behaviorDtos,
			Style:         &StyleDto{Color: def.Style.Color, Icon: def.Style.Icon},
			BaseCost:      baseCost,
			Available:     state.Available(),
			Errors:        convertStateErrors(state.Errors),
			Warnings:      convertStateWarnings(state.Warnings),
			EffectiveCost: state.Cost,
			Discounts:     discounts,
			Metadata:      state.Metadata,
		}

		result = append(result, dto)
	}

	return result
}

// mapPlayerMilestones calculates state for all milestones and converts to DTOs.
// Uses the state calculator to compute availability on-the-fly (same pattern as standard projects).
func mapPlayerMilestones(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry, milestoneRegistry milestones.MilestoneRegistry) []PlayerMilestoneDto {
	if milestoneRegistry == nil {
		return nil
	}
	allDefs := milestoneRegistry.GetAll()
	settings := g.Settings()
	enabledPacks := make(map[string]bool, len(settings.CardPacks))
	for _, pack := range settings.CardPacks {
		enabledPacks[pack] = true
	}
	if settings.VenusNextEnabled {
		enabledPacks[shared.PackVenus] = true
	}

	result := make([]PlayerMilestoneDto, 0, len(allDefs))
	gameMilestones := g.Milestones()

	for _, def := range allDefs {
		if def.Pack != "" && !enabledPacks[def.Pack] {
			continue
		}
		milestoneType := shared.MilestoneType(def.ID)
		state := action.CalculateMilestoneState(milestoneType, p, g, cardRegistry, milestoneRegistry)

		isClaimed := gameMilestones.IsClaimed(milestoneType)
		var claimedBy *string
		for _, claimed := range gameMilestones.ClaimedMilestones() {
			if claimed.Type == milestoneType {
				claimedBy = &claimed.PlayerID
				break
			}
		}

		progress := 0
		if prog, ok := state.Metadata["progress"].(int); ok {
			progress = prog
		}

		rewardDtos := buildMilestoneRewardDtos(def.Reward)

		var styleDtoPtr *StyleDto
		if def.Style.Color != "" || def.Style.Icon != "" {
			styleDtoPtr = &StyleDto{Color: def.Style.Color, Icon: def.Style.Icon}
		}

		dto := PlayerMilestoneDto{
			Type:        def.ID,
			Name:        def.Name,
			Description: def.Description,
			ClaimCost:   def.ClaimCost,
			IsClaimed:   isClaimed,
			ClaimedBy:   claimedBy,
			Available:   state.Available(),
			Progress:    progress,
			Required:    def.GetRequired(),
			Errors:      convertStateErrors(state.Errors),
			Reward:      rewardDtos,
			Style:       styleDtoPtr,
		}
		result = append(result, dto)
	}

	return result
}

// convertGenerationalEvents converts PlayerGenerationalEventEntry slice to DTO slice
func convertGenerationalEvents(entries []shared.PlayerGenerationalEventEntry) []PlayerGenerationalEventEntryDto {
	if len(entries) == 0 {
		return []PlayerGenerationalEventEntryDto{}
	}

	dtos := make([]PlayerGenerationalEventEntryDto, len(entries))
	for i, entry := range entries {
		dtos[i] = PlayerGenerationalEventEntryDto{
			Event: GenerationalEvent(entry.Event),
			Count: entry.Count,
		}
	}
	return dtos
}

// mapPlayerAwards calculates state for all awards and converts to DTOs.
// Uses the state calculator to compute availability on-the-fly (same pattern as standard projects).
func mapPlayerAwards(p *player.Player, g *game.Game, awardRegistry awards.AwardRegistry) []PlayerAwardDto {
	if awardRegistry == nil {
		return nil
	}
	allDefs := awardRegistry.GetAll()
	settings := g.Settings()
	enabledPacks := make(map[string]bool, len(settings.CardPacks))
	for _, pack := range settings.CardPacks {
		enabledPacks[pack] = true
	}
	if settings.VenusNextEnabled {
		enabledPacks[shared.PackVenus] = true
	}

	result := make([]PlayerAwardDto, 0, len(allDefs))
	gameAwards := g.Awards()
	fundedCount := gameAwards.FundedCount()

	for _, def := range allDefs {
		if def.Pack != "" && !enabledPacks[def.Pack] {
			continue
		}
		awardType := shared.AwardType(def.ID)
		state := action.CalculateAwardState(awardType, p, g, awardRegistry)

		isFunded := gameAwards.IsFunded(awardType)
		var fundedBy *string
		for _, funded := range gameAwards.FundedAwards() {
			if funded.Type == awardType {
				fundedBy = &funded.FundedByPlayer
				break
			}
		}

		var styleDtoPtr *StyleDto
		if def.Style.Color != "" || def.Style.Icon != "" {
			styleDtoPtr = &StyleDto{Color: def.Style.Color, Icon: def.Style.Icon}
		}

		dto := PlayerAwardDto{
			Type:        def.ID,
			Name:        def.Name,
			Description: def.Description,
			FundingCost: def.GetCostForFundedCount(fundedCount),
			IsFunded:    isFunded,
			FundedBy:    fundedBy,
			Available:   state.Available(),
			Errors:      convertStateErrors(state.Errors),
			Style:       styleDtoPtr,
		}
		result = append(result, dto)
	}

	return result
}

func mapPlayerActionCosts(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []ActionCostDto {
	calc := gamecards.NewRequirementModifierCalculator(cardRegistry)

	cardBuyDiscounts := calc.CalculateActionDiscounts(p, shared.ActionCardBuying)
	creditDiscount := cardBuyDiscounts[shared.ResourceCredit]
	cardBuyBase := 3
	cardBuyEffective := cardBuyBase - creditDiscount
	if cardBuyEffective < 0 {
		cardBuyEffective = 0
	}

	result := []ActionCostDto{
		{
			ActionType: shared.ActionCardBuying,
			Costs: []ActionCostEntryDto{
				{
					Resource:      string(shared.ResourceCredit),
					BaseCost:      cardBuyBase,
					EffectiveCost: cardBuyEffective,
					Discount:      creditDiscount,
				},
			},
		},
	}

	if g.HasColonies() {
		tradeDiscounts := calc.CalculateActionDiscounts(p, shared.ActionColonyTrade)
		tradeCosts := []ActionCostEntryDto{
			{
				Resource:      string(shared.ResourceCredit),
				BaseCost:      9,
				EffectiveCost: max(9-tradeDiscounts[shared.ResourceCredit], 0),
				Discount:      tradeDiscounts[shared.ResourceCredit],
			},
			{
				Resource:      string(shared.ResourceEnergy),
				BaseCost:      3,
				EffectiveCost: max(3-tradeDiscounts[shared.ResourceEnergy], 0),
				Discount:      tradeDiscounts[shared.ResourceEnergy],
			},
			{
				Resource:      string(shared.ResourceTitanium),
				BaseCost:      3,
				EffectiveCost: max(3-tradeDiscounts[shared.ResourceTitanium], 0),
				Discount:      tradeDiscounts[shared.ResourceTitanium],
			},
		}
		result = append(result, ActionCostDto{
			ActionType: shared.ActionColonyTrade,
			Costs:      tradeCosts,
		})
	}

	return result
}
