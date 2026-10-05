package bot_test

import (
	"context"
	"encoding/json"
	"testing"

	"terraforming-mars-backend/internal/action/admin"
	cardAction "terraforming-mars-backend/internal/action/card"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"

	gameAction "terraforming-mars-backend/internal/action/game"
)

func TestDispatcher_UnknownType(t *testing.T) {
	logger := testutil.TestLogger()

	dispatcher := bot.NewCommandDispatcher(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, logger,
	)

	rawCmd := json.RawMessage(`{"type": "unknown.command", "payload": {}}`)
	err := dispatcher.Dispatch(context.Background(), "game-1", "player-1", rawCmd)
	testutil.AssertError(t, err, "Should error on unknown command type")
}

func TestDispatcher_InvalidJSON(t *testing.T) {
	logger := testutil.TestLogger()

	dispatcher := bot.NewCommandDispatcher(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, logger,
	)

	rawCmd := json.RawMessage(`not json`)
	err := dispatcher.Dispatch(context.Background(), "game-1", "player-1", rawCmd)
	testutil.AssertError(t, err, "Should error on invalid JSON")
}

func TestDispatcher_SkipAction(t *testing.T) {
	g, repo, cardRegistry, p1, _ := testutil.SetupTwoPlayerGame(t)
	logger := testutil.TestLogger()

	finalScoringAction := gameAction.NewFinalScoringAction(repo, cardRegistry, nil, nil, logger)
	skipAction := turnAction.NewSkipActionAction(repo, finalScoringAction, logger)

	dispatcher := bot.NewCommandDispatcher(
		nil, nil, skipAction, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, logger,
	)

	rawCmd := json.RawMessage(`{"type": "action.game-management.skip-action", "payload": {}}`)
	err := dispatcher.Dispatch(context.Background(), g.ID(), p1, rawCmd)
	testutil.AssertNoError(t, err, "Skip action should succeed")

	player, _ := g.GetPlayer(p1)
	testutil.AssertTrue(t, player.HasPassed(), "Player should have passed after skip")
}

func TestDispatcher_ForwardsActionReuse(t *testing.T) {
	g, repo, registry, id, _ := testutil.SetupTwoPlayerGame(t)
	ctx := context.Background()
	logger := testutil.TestLogger()
	testutil.AssertNoError(t, admin.NewSetCorporationAction(repo, registry, nil, logger).Execute(ctx, g.ID(), id, "V05"), "Viron")
	p, _ := g.GetPlayer(id)
	target := testutil.GetCardByName("Power Infrastructure")
	p.PlayedCards().AddCard(target.ID, target.Name, "active", nil)
	p.Actions().SetActions(append(p.Actions().List(), shared.CardAction{CardID: target.ID, CardName: target.Name, Behavior: target.Behaviors[0], TimesUsedThisGeneration: 1}))
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceEnergy: 3})
	before := p.Resources().Get().Credits
	dispatcher := bot.NewCommandDispatcher(nil, cardAction.NewUseCardActionAction(repo, registry, nil, logger), nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, logger)
	err := dispatcher.Dispatch(ctx, g.ID(), id, json.RawMessage(`{"type":"action.card.card-action","payload":{"cardId":"194","behaviorIndex":0,"selectedAmount":3,"reuseSourceCardId":"V05"}}`))
	testutil.AssertNoError(t, err, "bot reuses target")
	testutil.AssertEqual(t, before+3, p.Resources().Get().Credits, "output applied")
	testutil.AssertEqual(t, 1, p.Actions().List()[0].TimesUsedThisGeneration, "Viron used")
}
