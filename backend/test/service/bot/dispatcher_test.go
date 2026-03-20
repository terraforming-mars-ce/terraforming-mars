package bot_test

import (
	"context"
	"encoding/json"
	"testing"

	turnAction "terraforming-mars-backend/internal/action/turn_management"
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
		nil, logger,
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
		nil, logger,
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
		nil, logger,
	)

	rawCmd := json.RawMessage(`{"type": "action.game-management.skip-action", "payload": {}}`)
	err := dispatcher.Dispatch(context.Background(), g.ID(), p1, rawCmd)
	testutil.AssertNoError(t, err, "Skip action should succeed")

	player, _ := g.GetPlayer(p1)
	testutil.AssertTrue(t, player.HasPassed(), "Player should have passed after skip")
}
