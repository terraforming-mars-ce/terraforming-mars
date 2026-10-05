package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	awardAction "terraforming-mars-backend/internal/action/award"
	cardAction "terraforming-mars-backend/internal/action/card"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game/shared"
)

// CommandDispatcher maps bot JSONL commands to direct action calls.
type CommandDispatcher struct {
	confirmColonyPlacement *confirmAction.ConfirmColonyPlacementAction
	confirmColonyResource  *confirmAction.ConfirmColonyResourceAction
	confirmAwardFund       *confirmAction.ConfirmAwardFundAction
	confirmResourceRemoval *confirmAction.ConfirmResourceRemovalAction
	confirmCardReveal      *confirmAction.ConfirmCardRevealAction
	confirmEffectSelection *confirmAction.ConfirmEffectSelectionAction
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
	executeStandardProject *stdprojAction.ExecuteStandardProjectAction
	convertHeat            *resconvAction.ConvertHeatToTemperatureAction
	convertPlants          *resconvAction.ConvertPlantsToGreeneryAction
	claimMilestone         *milestoneAction.ClaimMilestoneAction
	fundAward              *awardAction.FundAwardAction
	confirmInitAdvance     *turnAction.ConfirmInitAdvanceAction
	logger                 *slog.Logger
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
	confirmEffectSelection *confirmAction.ConfirmEffectSelectionAction,
	confirmCardReveal *confirmAction.ConfirmCardRevealAction,
	confirmSellPatents *confirmAction.ConfirmSellPatentsAction,
	executeStandardProject *stdprojAction.ExecuteStandardProjectAction,
	convertHeat *resconvAction.ConvertHeatToTemperatureAction,
	convertPlants *resconvAction.ConvertPlantsToGreeneryAction,
	claimMilestone *milestoneAction.ClaimMilestoneAction,
	fundAward *awardAction.FundAwardAction,
	confirmInitAdvance *turnAction.ConfirmInitAdvanceAction,
	confirmResourceRemoval *confirmAction.ConfirmResourceRemovalAction,
	confirmColonyPlacement *confirmAction.ConfirmColonyPlacementAction,
	confirmColonyResource *confirmAction.ConfirmColonyResourceAction,
	confirmAwardFund *confirmAction.ConfirmAwardFundAction,
	logger *slog.Logger,
) *CommandDispatcher {
	return &CommandDispatcher{
		confirmColonyPlacement: confirmColonyPlacement, confirmColonyResource: confirmColonyResource, confirmAwardFund: confirmAwardFund,
		confirmResourceRemoval: confirmResourceRemoval,
		confirmEffectSelection: confirmEffectSelection,
		confirmCardReveal:      confirmCardReveal,
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
		executeStandardProject: executeStandardProject,
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

	d.logger.Debug("Dispatching bot command",
		slog.String("game_id", gameID),
		slog.String("player_id", playerID),
		slog.String("type", envelope.Type))

	var paymentEnvelope struct {
		Payment shared.Payment `json:"payment"`
	}
	if len(envelope.Payload) > 0 {
		if err := json.Unmarshal(envelope.Payload, &paymentEnvelope); err != nil {
			return err
		}
	}
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
	case "action.acknowledge-card-receipt":
		var p struct {
			ReceiptID string `json:"receiptId"`
		}
		if err := json.Unmarshal(envelope.Payload, &p); err != nil {
			return err
		}
		return d.confirmCardDraw.AcknowledgeReceipt(ctx, gameID, playerID, p.ReceiptID)
	case "action.confirm-colony-placement":
		var p struct {
			ColonyID string `json:"colonyId"`
		}
		if err := json.Unmarshal(envelope.Payload, &p); err != nil {
			return err
		}
		return d.confirmColonyPlacement.Execute(ctx, gameID, playerID, p.ColonyID)
	case "action.confirm-colony-resource":
		var p struct {
			CardID string `json:"cardId"`
		}
		if err := json.Unmarshal(envelope.Payload, &p); err != nil {
			return err
		}
		return d.confirmColonyResource.Execute(ctx, gameID, playerID, p.CardID)
	case "action.confirm-award-fund":
		var p struct {
			AwardType string `json:"awardType"`
		}
		if err := json.Unmarshal(envelope.Payload, &p); err != nil {
			return err
		}
		return d.confirmAwardFund.Execute(ctx, gameID, playerID, p.AwardType)
	case "action.card.card-draw-confirmed":
		return d.dispatchConfirmCardDraw(ctx, gameID, playerID, envelope.Payload)
	case "action.card.card-discard-confirmed":
		return d.dispatchConfirmCardDiscard(ctx, gameID, playerID, envelope.Payload)
	case "action.confirm-card-reveal":
		return d.confirmCardReveal.Execute(ctx, gameID, playerID)
	case "action.card.confirm-resource-removal":
		var p struct {
			SelectionID    string `json:"selectionId"`
			TargetPlayerID string `json:"targetPlayerId"`
			Amount         *int   `json:"amount"`
		}
		if err := json.Unmarshal(envelope.Payload, &p); err != nil {
			return err
		}
		if p.Amount == nil {
			return fmt.Errorf("missing amount")
		}
		return d.confirmResourceRemoval.Execute(ctx, gameID, playerID, p.SelectionID, p.TargetPlayerID, *p.Amount)
	case "action.confirm-effect-selection":
		var p struct {
			OptionIndex *int `json:"optionIndex"`
		}
		if err := json.Unmarshal(envelope.Payload, &p); err != nil {
			return err
		}
		if p.OptionIndex == nil {
			return fmt.Errorf("missing optionIndex")
		}
		return d.confirmEffectSelection.Execute(ctx, gameID, playerID, *p.OptionIndex)
	case "action.card.behavior-choice-confirmed":
		return d.dispatchConfirmBehaviorChoice(ctx, gameID, playerID, envelope.Payload)
	case "action.card.select-cards":
		return d.dispatchConfirmSellPatents(ctx, gameID, playerID, envelope.Payload)
	case "action.standard-project":
		return d.dispatchStandardProject(ctx, gameID, playerID, envelope.Payload)
	case "action.standard-project.sell-patents":
		return d.executeStandardProject.Execute(ctx, gameID, playerID, "sell-patents", paymentEnvelope.Payment)
	case "action.standard-project.confirm-sell-patents":
		return d.dispatchConfirmSellPatents(ctx, gameID, playerID, envelope.Payload)
	case "action.standard-project.launch-asteroid":
		return d.executeStandardProject.Execute(ctx, gameID, playerID, "asteroid", paymentEnvelope.Payment)
	case "action.standard-project.build-power-plant":
		return d.executeStandardProject.Execute(ctx, gameID, playerID, "power-plant", paymentEnvelope.Payment)
	case "action.standard-project.build-aquifer":
		return d.executeStandardProject.Execute(ctx, gameID, playerID, "aquifer", paymentEnvelope.Payment)
	case "action.standard-project.plant-greenery":
		return d.executeStandardProject.Execute(ctx, gameID, playerID, "greenery", paymentEnvelope.Payment)
	case "action.standard-project.build-city":
		return d.executeStandardProject.Execute(ctx, gameID, playerID, "city", paymentEnvelope.Payment)
	case "action.resource-conversion.convert-heat-to-temperature":
		return d.convertHeat.Execute(ctx, gameID, playerID, paymentEnvelope.Payment)
	case "action.resource-conversion.convert-plants-to-greenery":
		return d.convertPlants.Execute(ctx, gameID, playerID, paymentEnvelope.Payment)
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
		CardID             string         `json:"cardId"`
		Payment            shared.Payment `json:"payment"`
		ChoiceIndex        *int           `json:"choiceIndex,omitempty"`
		CardStorageSources []string       `json:"cardStorageSources,omitempty"`
		CardStorageTargets []string       `json:"cardStorageTargets,omitempty"`
		TargetPlayerID     *string        `json:"targetPlayerId,omitempty"`
		SelectedAmount     *int           `json:"selectedAmount,omitempty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse play-card payload: %w", err)
	}

	return d.playCard.Execute(ctx, gameID, playerID, p.CardID, p.Payment, p.ChoiceIndex, p.CardStorageTargets, p.TargetPlayerID, p.SelectedAmount, p.CardStorageSources)
}

func (d *CommandDispatcher) dispatchUseCardAction(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		ReuseSourceCardID  *string         `json:"reuseSourceCardId,omitempty"`
		CardID             string          `json:"cardId"`
		BehaviorIndex      int             `json:"behaviorIndex"`
		ChoiceIndex        *int            `json:"choiceIndex,omitempty"`
		CardStorageSources []string        `json:"cardStorageSources,omitempty"`
		CardStorageTargets []string        `json:"cardStorageTargets,omitempty"`
		TargetPlayerID     *string         `json:"targetPlayerId,omitempty"`
		SourceCardForInput *string         `json:"sourceCardForInput,omitempty"`
		SelectedAmount     *int            `json:"selectedAmount,omitempty"`
		Payment            *shared.Payment `json:"payment,omitempty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse card-action payload: %w", err)
	}

	return d.useCardAction.Execute(ctx, gameID, playerID, p.CardID, p.BehaviorIndex, p.ChoiceIndex, p.CardStorageTargets, p.TargetPlayerID, p.SourceCardForInput, p.SelectedAmount, p.Payment, p.ReuseSourceCardID, p.CardStorageSources)
}

func (d *CommandDispatcher) dispatchSelectStartingChoices(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		Payment       shared.Payment `json:"payment"`
		CorporationID string         `json:"corporationId"`
		PreludeIDs    []string       `json:"preludeIds"`
		CardIDs       []string       `json:"cardIds"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse select-starting-choices payload: %w", err)
	}
	return d.selectStartingChoices.Execute(ctx, gameID, playerID, p.CorporationID, p.PreludeIDs, p.CardIDs, p.Payment)
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
		Payment shared.Payment `json:"payment"`
		CardIDs []string       `json:"cardIds"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse confirm-production-cards payload: %w", err)
	}
	return d.confirmProductionCards.Execute(ctx, gameID, playerID, p.CardIDs, false, p.Payment)
}

func (d *CommandDispatcher) dispatchConfirmCardDraw(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		Payment     shared.Payment `json:"payment"`
		CardsToTake []string       `json:"cardsToTake"`
		CardsToBuy  []string       `json:"cardsToBuy"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse card-draw-confirmed payload: %w", err)
	}
	return d.confirmCardDraw.Execute(ctx, gameID, playerID, p.CardsToTake, p.CardsToBuy, p.Payment)
}

func (d *CommandDispatcher) dispatchConfirmCardDiscard(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		ResolutionID   string   `json:"resolutionId"`
		CardsToDiscard []string `json:"cardsToDiscard"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse card-discard-confirmed payload: %w", err)
	}
	return d.confirmCardDiscard.Execute(ctx, gameID, playerID, p.ResolutionID, p.CardsToDiscard)
}

func (d *CommandDispatcher) dispatchConfirmBehaviorChoice(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		ResolutionID       string   `json:"resolutionId"`
		ChoiceIndex        int      `json:"choiceIndex"`
		CardStorageSources []string `json:"cardStorageSources,omitempty"`
		CardStorageTargets []string `json:"cardStorageTargets,omitempty"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse behavior-choice-confirmed payload: %w", err)
	}
	return d.confirmBehaviorChoice.Execute(ctx, gameID, playerID, p.ResolutionID, p.ChoiceIndex, p.CardStorageTargets)
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
		Payment       shared.Payment `json:"payment"`
		MilestoneType string         `json:"milestoneType"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse claim-milestone payload: %w", err)
	}
	return d.claimMilestone.Execute(ctx, gameID, playerID, p.MilestoneType, p.Payment)
}

func (d *CommandDispatcher) dispatchFundAward(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		Payment   shared.Payment `json:"payment"`
		AwardType string         `json:"awardType"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse fund-award payload: %w", err)
	}
	return d.fundAward.Execute(ctx, gameID, playerID, p.AwardType, p.Payment)
}

func (d *CommandDispatcher) dispatchStandardProject(ctx context.Context, gameID, playerID string, payload json.RawMessage) error {
	var p struct {
		Payment   shared.Payment `json:"payment"`
		ProjectID string         `json:"projectId"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse standard-project payload: %w", err)
	}
	return d.executeStandardProject.Execute(ctx, gameID, playerID, p.ProjectID, p.Payment)
}
