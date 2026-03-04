package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	awardAction "terraforming-mars-backend/internal/action/award"
	cardAction "terraforming-mars-backend/internal/action/card"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// CommandDispatcher maps bot JSONL commands to direct action calls.
type CommandDispatcher struct {
	playCard               *cardAction.PlayCardAction
	useCardAction          *cardAction.UseCardActionAction
	skipAction             *turnAction.SkipActionAction
	selectStartingChoices  *turnAction.SelectStartingChoicesAction
	selectTile             *tileAction.SelectTileAction
	confirmProductionCards *confirmAction.ConfirmProductionCardsAction
	confirmCardDraw        *confirmAction.ConfirmCardDrawAction
	confirmCardDiscard     *confirmAction.ConfirmCardDiscardAction
	confirmBehaviorChoice  *confirmAction.ConfirmBehaviorChoiceAction
	confirmSellPatents     *confirmAction.ConfirmSellPatentsAction
	launchAsteroid         *stdprojAction.LaunchAsteroidAction
	buildPowerPlant        *stdprojAction.BuildPowerPlantAction
	buildAquifer           *stdprojAction.BuildAquiferAction
	buildCity              *stdprojAction.BuildCityAction
	plantGreenery          *stdprojAction.PlantGreeneryAction
	sellPatents            *stdprojAction.SellPatentsAction
	convertHeat            *resconvAction.ConvertHeatToTemperatureAction
	convertPlants          *resconvAction.ConvertPlantsToGreeneryAction
	claimMilestone         *milestoneAction.ClaimMilestoneAction
	fundAward              *awardAction.FundAwardAction
	confirmInitAdvance     *turnAction.ConfirmInitAdvanceAction
	logger                 *zap.Logger
}

// NewCommandDispatcher creates a new dispatcher with all action references.
func NewCommandDispatcher(
	playCard *cardAction.PlayCardAction,
	useCardAction *cardAction.UseCardActionAction,
	skipAction *turnAction.SkipActionAction,
	selectStartingChoices *turnAction.SelectStartingChoicesAction,
	selectTile *tileAction.SelectTileAction,
	confirmProductionCards *confirmAction.ConfirmProductionCardsAction,
	confirmCardDraw *confirmAction.ConfirmCardDrawAction,
	confirmCardDiscard *confirmAction.ConfirmCardDiscardAction,
	confirmBehaviorChoice *confirmAction.ConfirmBehaviorChoiceAction,
	confirmSellPatents *confirmAction.ConfirmSellPatentsAction,
	launchAsteroid *stdprojAction.LaunchAsteroidAction,
	buildPowerPlant *stdprojAction.BuildPowerPlantAction,
	buildAquifer *stdprojAction.BuildAquiferAction,
	buildCity *stdprojAction.BuildCityAction,
	plantGreenery *stdprojAction.PlantGreeneryAction,
	sellPatents *stdprojAction.SellPatentsAction,
	convertHeat *resconvAction.ConvertHeatToTemperatureAction,
	convertPlants *resconvAction.ConvertPlantsToGreeneryAction,
	claimMilestone *milestoneAction.ClaimMilestoneAction,
	fundAward *awardAction.FundAwardAction,
	confirmInitAdvance *turnAction.ConfirmInitAdvanceAction,
	logger *zap.Logger,
) *CommandDispatcher {
	return &CommandDispatcher{
		playCard:               playCard,
		useCardAction:          useCardAction,
		skipAction:             skipAction,
		selectStartingChoices:  selectStartingChoices,
		selectTile:             selectTile,
		confirmProductionCards: confirmProductionCards,
		confirmCardDraw:        confirmCardDraw,
		confirmCardDiscard:     confirmCardDiscard,
		confirmBehaviorChoice:  confirmBehaviorChoice,
		confirmSellPatents:     confirmSellPatents,
		launchAsteroid:         launchAsteroid,
		buildPowerPlant:        buildPowerPlant,
		buildAquifer:           buildAquifer,
		buildCity:              buildCity,
		plantGreenery:          plantGreenery,
		sellPatents:            sellPatents,
		convertHeat:            convertHeat,
		convertPlants:          convertPlants,
		claimMilestone:         claimMilestone,
		fundAward:              fundAward,
		confirmInitAdvance:     confirmInitAdvance,
		logger:                 logger,
	}
}

// Dispatch parses a raw JSON command and calls the appropriate action.
func (d *CommandDispatcher) Dispatch(ctx context.Context, gameID, playerID string, rawJSON json.RawMessage) error {
	var envelope struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(rawJSON, &envelope); err != nil {
		return fmt.Errorf("parse command envelope: %w", err)
	}

	d.logger.Info("📤 Dispatching bot command",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("type", envelope.Type))

	switch envelope.Type {
	case "action.card.play-card":
		return d.dispatchPlayCard(ctx, gameID, playerID, envelope.Payload)
	case "action.card.card-action":
		return d.dispatchUseCardAction(ctx, gameID, playerID, envelope.Payload)
	case "action.game-management.skip-action":
		return d.skipAction.Execute(ctx, gameID, playerID)
	case "action.card.select-starting-choices":
		return d.dispatchSelectStartingChoices(ctx, gameID, playerID, envelope.Payload)
	case "action.tile-selection.tile-selected":
		return d.dispatchSelectTile(ctx, gameID, playerID, envelope.Payload)
	case "action.card.confirm-production-cards":
		return d.dispatchConfirmProductionCards(ctx, gameID, playerID, envelope.Payload)
	case "action.card.card-draw-confirmed":
		return d.dispatchConfirmCardDraw(ctx, gameID, playerID, envelope.Payload)
	case "action.card.card-discard-confirmed":
		return d.dispatchConfirmCardDiscard(ctx, gameID, playerID, envelope.Payload)
	case "action.card.behavior-choice-confirmed":
		return d.dispatchConfirmBehaviorChoice(ctx, gameID, playerID, envelope.Payload)
	case "action.card.select-cards":
		return d.dispatchConfirmSellPatents(ctx, gameID, playerID, envelope.Payload)
	case "action.standard-project.sell-patents":
		return d.sellPatents.Execute(ctx, gameID, playerID)
	case "action.standard-project.confirm-sell-patents":
		return d.dispatchConfirmSellPatents(ctx, gameID, playerID, envelope.Payload)
	case "action.standard-project.launch-asteroid":
		return d.launchAsteroid.Execute(ctx, gameID, playerID)
	case "action.standard-project.build-power-plant":
		return d.buildPowerPlant.Execute(ctx, gameID, playerID)
	case "action.standard-project.build-aquifer":
		return d.buildAquifer.Execute(ctx, gameID, playerID)
	case "action.standard-project.plant-greenery":
		return d.plantGreenery.Execute(ctx, gameID, playerID)
	case "action.standard-project.build-city":
		return d.buildCity.Execute(ctx, gameID, playerID)
	case "action.resource-conversion.convert-heat-to-temperature":
		return d.convertHeat.Execute(ctx, gameID, playerID)
	case "action.resource-conversion.convert-plants-to-greenery":
		return d.convertPlants.Execute(ctx, gameID, playerID)
	case "action.game-management.confirm-init-advance":
		return d.confirmInitAdvance.Execute(ctx, gameID, playerID)
	case "action.milestone.claim-milestone":
		return d.dispatchClaimMilestone(ctx, gameID, playerID, envelope.Payload)
	case "action.award.fund-award":
		return d.dispatchFundAward(ctx, gameID, playerID, envelope.Payload)
	default:
		return fmt.Errorf("unknown command type: %s", envelope.Type)
	}
}

func (d *CommandDispatcher) dispatchPlayCard(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		CardID             string          `json:"cardId"`
		Payment            playCardPayment `json:"payment"`
		ChoiceIndex        *int            `json:"choiceIndex,omitempty"`
		CardStorageTargets []string        `json:"cardStorageTargets,omitempty"`
		TargetPlayerID     *string         `json:"targetPlayerId,omitempty"`
		SelectedAmount     *int            `json:"selectedAmount,omitempty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse play-card payload: %w", err)
	}

	payment := cardAction.PaymentRequest{
		Credits:            p.Payment.Credits,
		Steel:              p.Payment.Steel,
		Titanium:           p.Payment.Titanium,
		Substitutes:        make(map[shared.ResourceType]int),
		StorageSubstitutes: make(map[string]int),
	}
	for k, v := range p.Payment.Substitutes {
		payment.Substitutes[shared.ResourceType(k)] = v
	}
	maps.Copy(payment.StorageSubstitutes, p.Payment.StorageSubstitutes)

	return d.playCard.Execute(ctx, gameID, playerID, p.CardID, payment, p.ChoiceIndex, p.CardStorageTargets, p.TargetPlayerID, p.SelectedAmount)
}

type playCardPayment struct {
	Credits            int            `json:"credits"`
	Steel              int            `json:"steel"`
	Titanium           int            `json:"titanium"`
	Substitutes        map[string]int `json:"substitutes,omitempty"`
	StorageSubstitutes map[string]int `json:"storageSubstitutes,omitempty"`
}

func (d *CommandDispatcher) dispatchUseCardAction(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		CardID             string   `json:"cardId"`
		BehaviorIndex      int      `json:"behaviorIndex"`
		ChoiceIndex        *int     `json:"choiceIndex,omitempty"`
		CardStorageTargets []string `json:"cardStorageTargets,omitempty"`
		TargetPlayerID     *string  `json:"targetPlayerId,omitempty"`
		SourceCardForInput *string  `json:"sourceCardForInput,omitempty"`
		SelectedAmount     *int     `json:"selectedAmount,omitempty"`
		Payment            *struct {
			Credits  int `json:"credits"`
			Steel    int `json:"steel"`
			Titanium int `json:"titanium"`
		} `json:"payment,omitempty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse card-action payload: %w", err)
	}

	var actionPayment *gamecards.CardPayment
	if p.Payment != nil {
		actionPayment = &gamecards.CardPayment{
			Credits:  p.Payment.Credits,
			Steel:    p.Payment.Steel,
			Titanium: p.Payment.Titanium,
		}
	}

	return d.useCardAction.Execute(ctx, gameID, playerID, p.CardID, p.BehaviorIndex, p.ChoiceIndex, p.CardStorageTargets, p.TargetPlayerID, p.SourceCardForInput, p.SelectedAmount, actionPayment)
}

func (d *CommandDispatcher) dispatchSelectStartingChoices(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		CorporationID string   `json:"corporationId"`
		PreludeIDs    []string `json:"preludeIds"`
		CardIDs       []string `json:"cardIds"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse select-starting-choices payload: %w", err)
	}
	return d.selectStartingChoices.Execute(ctx, gameID, playerID, p.CorporationID, p.PreludeIDs, p.CardIDs)
}

func (d *CommandDispatcher) dispatchSelectTile(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		Hex string `json:"hex"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse tile-selected payload: %w", err)
	}
	_, err := d.selectTile.Execute(ctx, gameID, playerID, p.Hex)
	return err
}

func (d *CommandDispatcher) dispatchConfirmProductionCards(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		CardIDs []string `json:"cardIds"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse confirm-production-cards payload: %w", err)
	}
	return d.confirmProductionCards.Execute(ctx, gameID, playerID, p.CardIDs)
}

func (d *CommandDispatcher) dispatchConfirmCardDraw(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		CardsToTake []string `json:"cardsToTake"`
		CardsToBuy  []string `json:"cardsToBuy"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse card-draw-confirmed payload: %w", err)
	}
	return d.confirmCardDraw.Execute(ctx, gameID, playerID, p.CardsToTake, p.CardsToBuy)
}

func (d *CommandDispatcher) dispatchConfirmCardDiscard(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		CardsToDiscard []string `json:"cardsToDiscard"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse card-discard-confirmed payload: %w", err)
	}
	return d.confirmCardDiscard.Execute(ctx, gameID, playerID, p.CardsToDiscard)
}

func (d *CommandDispatcher) dispatchConfirmBehaviorChoice(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		ChoiceIndex        int      `json:"choiceIndex"`
		CardStorageTargets []string `json:"cardStorageTargets,omitempty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse behavior-choice-confirmed payload: %w", err)
	}
	return d.confirmBehaviorChoice.Execute(ctx, gameID, playerID, p.ChoiceIndex, p.CardStorageTargets)
}

func (d *CommandDispatcher) dispatchConfirmSellPatents(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		SelectedCardIDs []string `json:"selectedCardIds"`
		CardIDs         []string `json:"cardIds"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse confirm-sell-patents payload: %w", err)
	}
	ids := p.SelectedCardIDs
	if len(ids) == 0 {
		ids = p.CardIDs
	}
	return d.confirmSellPatents.Execute(ctx, gameID, playerID, ids)
}

func (d *CommandDispatcher) dispatchClaimMilestone(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		MilestoneType string `json:"milestoneType"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse claim-milestone payload: %w", err)
	}
	return d.claimMilestone.Execute(ctx, gameID, playerID, p.MilestoneType)
}

func (d *CommandDispatcher) dispatchFundAward(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		AwardType string `json:"awardType"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse fund-award payload: %w", err)
	}
	return d.fundAward.Execute(ctx, gameID, playerID, p.AwardType)
}
