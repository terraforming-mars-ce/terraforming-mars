package dto

import (
	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
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
func ToPlayerDto(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) PlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	corporation := getCorporationCard(p, cardRegistry)
	playedCardIDs := p.PlayedCards().Cards()
	playedCards := getPlayedCards(playedCardIDs, cardRegistry)
	handCards := mapPlayerCards(p, g, cardRegistry)
	standardProjects := mapPlayerStandardProjects(p, g, cardRegistry)
	milestones := mapPlayerMilestones(p, g, cardRegistry)
	awards := mapPlayerAwards(p, g)
	var pendingTileSelection *PendingTileSelectionDto
	var forcedFirstAction *ForcedFirstActionDto
	currentTurn := g.CurrentTurn()
	if currentTurn != nil && currentTurn.PlayerID() == p.ID() {
		pendingTileSelection = convertPendingTileSelection(g.GetPendingTileSelection(p.ID()))
		forcedFirstAction = convertForcedFirstAction(g.GetForcedFirstAction(p.ID()))
	}
	if g.CurrentPhase() == game.GamePhaseStartingSelection ||
		g.CurrentPhase() == game.GamePhaseInitApplyCorp ||
		g.CurrentPhase() == game.GamePhaseInitApplyPrelude {
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
		Status:           playerStatus(p),
		Corporation:      corporation,
		Cards:            handCards, // PlayerCardDto[] with state
		PlayedCards:      playedCards,
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		IsExited:         p.HasExited(),
		Effects:          convertPlayerEffects(p.Effects().List()),
		Actions:          convertPlayerActions(p.Actions().List(), p, g),
		StandardProjects: standardProjects, // PlayerStandardProjectDto[] with state
		Milestones:       milestones,       // PlayerMilestoneDto[] with eligibility
		Awards:           awards,           // PlayerAwardDto[] with eligibility

		SelectCorporationPhase:         convertSelectCorporationPhase(g.GetSelectCorporationPhase(p.ID()), cardRegistry),
		SelectStartingCardsPhase:       convertSelectStartingCardsPhase(g.GetSelectStartingCardsPhase(p.ID()), cardRegistry),
		SelectPreludeCardsPhase:        convertSelectPreludeCardsPhase(g.GetSelectPreludeCardsPhase(p.ID()), cardRegistry),
		ProductionPhase:                convertProductionPhase(g.GetProductionPhase(p.ID()), cardRegistry),
		StartingCards:                  []CardDto{},
		PendingTileSelection:           pendingTileSelection,
		PendingCardSelection:           convertPendingCardSelection(p.Selection().GetPendingCardSelection(), cardRegistry),
		PendingCardDrawSelection:       convertPendingCardDrawSelection(p.Selection().GetPendingCardDrawSelection(), cardRegistry),
		PendingCardDiscardSelection:    convertPendingCardDiscardSelection(p.Selection().GetPendingCardDiscardSelection()),
		PendingBehaviorChoiceSelection: convertPendingBehaviorChoiceSelection(p.Selection().GetPendingBehaviorChoiceSelection(), p, g, cardRegistry),
		ForcedFirstAction:              forcedFirstAction,
		ResourceStorage:                p.Resources().Storage(),
		PaymentSubstitutes:             convertPaymentSubstitutes(p.Resources().PaymentSubstitutes()),
		StoragePaymentSubstitutes:      convertStoragePaymentSubstitutes(p.Resources().StoragePaymentSubstitutes()),
		GenerationalEvents:             convertGenerationalEvents(p.GenerationalEvents().GetAll()),
		VPGranters:                     toVPGranterDtos(p.VPGranters().GetAll()),
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
		Status:           playerStatus(p),
		Corporation:      corporation,
		HandCardCount:    handCardCount,
		PlayedCards:      playedCards,
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		IsExited:         p.HasExited(),
		Effects:          convertPlayerEffects(p.Effects().List()),
		Actions:          convertPlayerActions(p.Actions().List(), p, g),

		SelectCorporationPhase:    convertSelectCorporationPhaseForOtherPlayer(g.GetSelectCorporationPhase(p.ID())),
		SelectStartingCardsPhase:  convertSelectStartingCardsPhaseForOtherPlayer(g.GetSelectStartingCardsPhase(p.ID())),
		SelectPreludeCardsPhase:   convertSelectPreludeCardsPhaseForOtherPlayer(g.GetSelectPreludeCardsPhase(p.ID())),
		ProductionPhase:           convertProductionPhaseForOtherPlayer(g.GetProductionPhase(p.ID())),
		ResourceStorage:           p.Resources().Storage(),
		PaymentSubstitutes:        convertPaymentSubstitutes(p.Resources().PaymentSubstitutes()),
		StoragePaymentSubstitutes: convertStoragePaymentSubstitutes(p.Resources().StoragePaymentSubstitutes()),
	}
}

func playerStatus(p *player.Player) PlayerStatus {
	if p.HasExited() {
		return PlayerStatusExited
	}
	return PlayerStatusWaiting
}

func convertSelectCorporationPhase(phase *player.SelectCorporationPhase, cardRegistry cards.CardRegistry) *SelectCorporationPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectCorporationPhaseDto{
		AvailableCorporations: getPlayedCards(phase.AvailableCorporations, cardRegistry),
	}
}

func convertSelectCorporationPhaseForOtherPlayer(phase *player.SelectCorporationPhase) *SelectCorporationOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectCorporationOtherPlayerDto{}
}

func convertSelectStartingCardsPhase(phase *player.SelectStartingCardsPhase, cardRegistry cards.CardRegistry) *SelectStartingCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsPhaseDto{
		AvailableCards: getPlayedCards(phase.AvailableCards, cardRegistry),
	}
}

func convertSelectStartingCardsPhaseForOtherPlayer(phase *player.SelectStartingCardsPhase) *SelectStartingCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsOtherPlayerDto{}
}

// convertSelectPreludeCardsPhase converts SelectPreludeCardsPhase to DTO
func convertSelectPreludeCardsPhase(phase *player.SelectPreludeCardsPhase, cardRegistry cards.CardRegistry) *SelectPreludeCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectPreludeCardsPhaseDto{
		AvailablePreludes: getPlayedCards(phase.AvailablePreludes, cardRegistry),
		MaxSelectable:     phase.MaxSelectable,
	}
}

// convertSelectPreludeCardsPhaseForOtherPlayer converts SelectPreludeCardsPhase to DTO for other players
func convertSelectPreludeCardsPhaseForOtherPlayer(phase *player.SelectPreludeCardsPhase) *SelectPreludeCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectPreludeCardsOtherPlayerDto{}
}

// convertProductionPhase converts production phase data to DTO for current player
func convertProductionPhase(phase *player.ProductionPhase, cardRegistry cards.CardRegistry) *ProductionPhaseDto {
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
func convertProductionPhaseForOtherPlayer(phase *player.ProductionPhase) *ProductionPhaseOtherPlayerDto {
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
func convertPlayerEffects(effects []player.CardEffect) []PlayerEffectDto {
	if len(effects) == 0 {
		return []PlayerEffectDto{}
	}

	dtos := make([]PlayerEffectDto, len(effects))
	for i, effect := range effects {
		dtos[i] = PlayerEffectDto{
			CardID:        effect.CardID,
			CardName:      effect.CardName,
			BehaviorIndex: effect.BehaviorIndex,
			Behavior:      toCardBehaviorDto(effect.Behavior),
		}
	}
	return dtos
}

// convertPlayerActions converts CardAction slice to PlayerActionDto slice
// Calculates state for each action to determine availability and errors.
func convertPlayerActions(actions []player.CardAction, p *player.Player, g *game.Game) []PlayerActionDto {
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
		)

		dtos[i] = PlayerActionDto{
			CardID:                  act.CardID,
			CardName:                act.CardName,
			BehaviorIndex:           act.BehaviorIndex,
			Behavior:                toCardBehaviorDto(act.Behavior),
			TimesUsedThisTurn:       act.TimesUsedThisTurn,
			TimesUsedThisGeneration: act.TimesUsedThisGeneration,
			Available:               state.Available(),
			Errors:                  convertStateErrors(state.Errors),
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
		}
	}
	return dtos
}

// convertPendingCardSelection converts PendingCardSelection to DTO
func convertPendingCardSelection(selection *player.PendingCardSelection, cardRegistry cards.CardRegistry) *PendingCardSelectionDto {
	if selection == nil {
		return nil
	}

	availableCards := getPlayedCards(selection.AvailableCards, cardRegistry)

	return &PendingCardSelectionDto{
		AvailableCards: availableCards,
		CardCosts:      selection.CardCosts,
		CardRewards:    selection.CardRewards,
		Source:         selection.Source,
		MinCards:       selection.MinCards,
		MaxCards:       selection.MaxCards,
	}
}

// convertPendingCardDrawSelection converts PendingCardDrawSelection to DTO
func convertPendingCardDrawSelection(selection *player.PendingCardDrawSelection, cardRegistry cards.CardRegistry) *PendingCardDrawSelectionDto {
	if selection == nil {
		return nil
	}

	availableCards := getPlayedCards(selection.AvailableCards, cardRegistry)

	return &PendingCardDrawSelectionDto{
		AvailableCards: availableCards,
		FreeTakeCount:  selection.FreeTakeCount,
		MaxBuyCount:    selection.MaxBuyCount,
		CardBuyCost:    selection.CardBuyCost,
		Source:         selection.Source,
	}
}

// convertPendingCardDiscardSelection converts PendingCardDiscardSelection to DTO
func convertPendingCardDiscardSelection(selection *player.PendingCardDiscardSelection) *PendingCardDiscardSelectionDto {
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

func convertPendingBehaviorChoiceSelection(selection *player.PendingBehaviorChoiceSelection, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) *PendingBehaviorChoiceSelectionDto {
	if selection == nil {
		return nil
	}

	choices := make([]ChoiceDto, len(selection.Choices))
	for i, choice := range selection.Choices {
		choices[i] = toChoiceDtoWithState(choice, p, g, cardRegistry)
	}

	return &PendingBehaviorChoiceSelectionDto{
		Choices:      choices,
		Source:       selection.Source,
		SourceCardID: selection.SourceCardID,
	}
}

// toChoiceDtoWithState maps a choice to DTO with computed errors from the state calculator.
func toChoiceDtoWithState(choice shared.Choice, p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) ChoiceDto {
	errors := action.CalculateChoiceErrors(choice, p, g, cardRegistry)
	return ChoiceDto{
		Inputs:       mapSlice(choice.Inputs, toResourceConditionDto),
		Outputs:      mapSlice(choice.Outputs, toResourceConditionDto),
		Requirements: toChoiceRequirementsDto(choice.Requirements),
		Available:    len(errors) == 0,
		Errors:       convertStateErrors(errors),
	}
}

// convertForcedFirstAction converts ForcedFirstAction to DTO
func convertForcedFirstAction(action *player.ForcedFirstAction) *ForcedFirstActionDto {
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
func convertPendingTileSelection(selection *player.PendingTileSelection) *PendingTileSelectionDto {
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

// ToPlayerCardDto converts a PlayerCard to PlayerCardDto with state information
func ToPlayerCardDto(pc *player.PlayerCard) PlayerCardDto {
	state := pc.State()

	cardAny := pc.Card()
	card, ok := cardAny.(*gamecards.Card)
	if !ok {
		// Defensive: return empty DTO if card type is wrong (should not happen)
		return PlayerCardDto{
			Available:     false,
			Errors:        []StateErrorDto{{Code: ErrorCodeInvalidCardType, Category: ErrorCategoryInternal, Message: "Invalid card type"}},
			EffectiveCost: 0,
		}
	}

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
		// Card data (same as CardDto)
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
	}
}

// mapPlayerCards converts cached PlayerCard instances from hand to DTOs.
// Enriches behavior choices with computed errors from the state calculator.
func mapPlayerCards(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []PlayerCardDto {
	handCardIDs := p.Hand().Cards()
	result := make([]PlayerCardDto, 0, len(handCardIDs))

	for _, cardID := range handCardIDs {
		// Get cached PlayerCard from hand
		pc, exists := p.Hand().GetPlayerCard(cardID)
		if !exists {
			// PlayerCard not cached - skip (should not happen if architecture is working correctly)
			continue
		}

		dto := ToPlayerCardDto(pc)

		// Enrich choices with computed errors
		if card, ok := pc.Card().(*gamecards.Card); ok {
			for bi, behavior := range card.Behaviors {
				if bi < len(dto.Behaviors) {
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
// Will be rewritten when standard projects are part of the cards JSON DB as metadata cards (instead of hard coded).
// NOTE: Conversion projects (plants→greenery, heat→temperature) are NOT included here - they
// are handled separately via resource buttons in the bottom bar.
func mapPlayerStandardProjects(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []PlayerStandardProjectDto {
	projectTypes := []shared.StandardProject{
		shared.StandardProjectSellPatents,
		shared.StandardProjectPowerPlant,
		shared.StandardProjectAsteroid,
		shared.StandardProjectAquifer,
		shared.StandardProjectGreenery,
		shared.StandardProjectCity,
	}

	result := make([]PlayerStandardProjectDto, 0, len(projectTypes))
	for _, projectType := range projectTypes {
		// Calculate state using the state calculator
		state := action.CalculatePlayerStandardProjectState(projectType, p, g, cardRegistry)

		baseCost := action.GetStandardProjectBaseCosts(projectType)

		discounts := make(map[string]int)
		if discountData, ok := state.Metadata["discounts"].(map[string]int); ok {
			discounts = discountData
		}

		dto := PlayerStandardProjectDto{
			ProjectType:   string(projectType),
			BaseCost:      baseCost,
			Available:     state.Available(),
			Errors:        convertStateErrors(state.Errors),
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
func mapPlayerMilestones(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) []PlayerMilestoneDto {
	result := make([]PlayerMilestoneDto, 0, len(game.AllMilestones))
	gameMilestones := g.Milestones()

	for _, info := range game.AllMilestones {
		// Calculate state on-the-fly using the state calculator
		state := action.CalculateMilestoneState(info.Type, p, g, cardRegistry)

		// Get claim status from game-level milestones
		isClaimed := gameMilestones.IsClaimed(info.Type)
		var claimedBy *string
		for _, claimed := range gameMilestones.ClaimedMilestones() {
			if claimed.Type == info.Type {
				claimedBy = &claimed.PlayerID
				break
			}
		}

		// Extract progress from metadata
		progress := 0
		if prog, ok := state.Metadata["progress"].(int); ok {
			progress = prog
		}

		dto := PlayerMilestoneDto{
			Type:        string(info.Type),
			Name:        info.Name,
			Description: info.Description,
			ClaimCost:   game.MilestoneClaimCost,
			IsClaimed:   isClaimed,
			ClaimedBy:   claimedBy,
			Available:   state.Available(),
			Progress:    progress,
			Required:    info.Requirement,
			Errors:      convertStateErrors(state.Errors),
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
func mapPlayerAwards(p *player.Player, g *game.Game) []PlayerAwardDto {
	result := make([]PlayerAwardDto, 0, len(game.AllAwards))
	gameAwards := g.Awards()
	currentCost := gameAwards.GetCurrentFundingCost()

	for _, info := range game.AllAwards {
		// Calculate state on-the-fly using the state calculator
		state := action.CalculateAwardState(info.Type, p, g)

		// Get funding status from game-level awards
		isFunded := gameAwards.IsFunded(info.Type)
		var fundedBy *string
		for _, funded := range gameAwards.FundedAwards() {
			if funded.Type == info.Type {
				fundedBy = &funded.FundedByPlayer
				break
			}
		}

		dto := PlayerAwardDto{
			Type:        string(info.Type),
			Name:        info.Name,
			Description: info.Description,
			FundingCost: currentCost,
			IsFunded:    isFunded,
			FundedBy:    fundedBy,
			Available:   state.Available(),
			Errors:      convertStateErrors(state.Errors),
		}
		result = append(result, dto)
	}

	return result
}
