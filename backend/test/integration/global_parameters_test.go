package integration_test

import (
	"context"
	"fmt"
	"testing"

	resconvAction "openmars/internal/action/resource_conversion"
	"openmars/internal/events"
	"openmars/internal/game"
	"openmars/internal/game/cards"
	"openmars/internal/game/global_parameters"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

func setupActiveGameForGlobalParams(t *testing.T) (*game.Game, game.GameRepository, cards.CardRegistry, string) {
	t.Helper()

	testGame, repo := testutil.CreateTestGameWithPlayers(t, 2)
	cardRegistry := testutil.CreateTestCardRegistry()
	testutil.StartTestGame(t, testGame)

	playerID := testGame.CurrentTurn().PlayerID()
	return testGame, repo, cardRegistry, playerID
}

func convertHeat(t *testing.T, repo game.GameRepository, cardRegistry cards.CardRegistry, g *game.Game, playerID string) error {
	t.Helper()
	action := resconvAction.NewConvertHeatToTemperatureAction(repo, cardRegistry, nil, testutil.TestLogger())
	return action.Execute(context.Background(), g.ID(), playerID, shared.NativePayment(shared.ResourceHeat, 8))
}

func TestGlobalParameters_TemperatureProgression(t *testing.T) {
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 24)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 2), "set current turn")
	initialTemp := testGame.GlobalParameters().Temperature()
	initialTR := player.Resources().TerraformRating()

	for i := range 2 {
		testutil.AssertNoError(t, convertHeat(t, repo, cardRegistry, testGame, playerID), fmt.Sprintf("conversion %d", i+1))
	}

	testutil.AssertEqual(t, initialTemp+4, testGame.GlobalParameters().Temperature(), "two conversions raise the temperature two steps of 2°C")
	testutil.AssertEqual(t, initialTR+2, player.Resources().TerraformRating(), "each step raises TR by 1")
	testutil.AssertEqual(t, 8, player.Resources().Get().Heat, "each conversion costs 8 heat")
	testutil.AssertErrorContains(t, convertHeat(t, repo, cardRegistry, testGame, playerID), "not your turn", "a third conversion in a two-action turn")
	testutil.AssertEqual(t, 8, player.Resources().Get().Heat, "a rejected conversion costs nothing")
	testutil.AssertEqual(t, initialTemp+4, testGame.GlobalParameters().Temperature(), "a rejected conversion leaves the temperature")
}

func TestGlobalParameters_TemperatureStopsAtMaximum(t *testing.T) {
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()
	testutil.AssertNoError(t, testGame.GlobalParameters().SetTemperature(ctx, global_parameters.MaxTemperature-2), "set temperature")
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 16)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 2), "set current turn")
	initialTR := player.Resources().TerraformRating()

	testutil.AssertNoError(t, convertHeat(t, repo, cardRegistry, testGame, playerID), "the last step")
	testutil.AssertEqual(t, global_parameters.MaxTemperature, testGame.GlobalParameters().Temperature(), "the last step reaches the maximum")
	testutil.AssertEqual(t, initialTR+1, player.Resources().TerraformRating(), "the last step raises TR")

	testutil.AssertNoError(t, convertHeat(t, repo, cardRegistry, testGame, playerID), "converting at the maximum")
	testutil.AssertEqual(t, global_parameters.MaxTemperature, testGame.GlobalParameters().Temperature(), "the temperature never exceeds the maximum")
	testutil.AssertEqual(t, initialTR+1, player.Resources().TerraformRating(), "converting at the maximum raises no TR")
}

func TestGlobalParameters_StartAtTheirMinimums(t *testing.T) {
	testGame, _, _, _ := setupActiveGameForGlobalParams(t)
	params := testGame.GlobalParameters()
	testutil.AssertEqual(t, global_parameters.MinTemperature, params.Temperature(), "temperature starts at its minimum")
	testutil.AssertEqual(t, 0, params.Oxygen(), "oxygen starts at 0")
	testutil.AssertEqual(t, 0, params.Oceans(), "no oceans at the start")
}

func TestGlobalParameters_RaisingTemperaturePublishesAnEvent(t *testing.T) {
	testGame, repo, cardRegistry, playerID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()
	player, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerHeat(ctx, player, 8)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 2), "set current turn")
	initialTemp := testGame.GlobalParameters().Temperature()

	var published []events.TemperatureChangedEvent
	events.Subscribe(testGame.EventBus(), func(e events.TemperatureChangedEvent) {
		published = append(published, e)
	})
	testutil.AssertNoError(t, convertHeat(t, repo, cardRegistry, testGame, playerID), "convert heat")

	testutil.AssertEqual(t, 1, len(published), "one temperature event")
	testutil.AssertEqual(t, initialTemp, published[0].OldValue, "the event carries the old temperature")
	testutil.AssertEqual(t, initialTemp+2, published[0].NewValue, "the event carries the new temperature")
	testutil.AssertEqual(t, testGame.ID(), published[0].GameID, "the event names the game")
}

func TestGlobalParameters_EachPlayerGetsTRForTheirOwnSteps(t *testing.T) {
	testGame, repo, cardRegistry, player1ID := setupActiveGameForGlobalParams(t)
	ctx := context.Background()
	var player2ID string
	for _, p := range testGame.GetAllPlayers() {
		if p.ID() != player1ID {
			player2ID = p.ID()
		}
	}
	player1, _ := testGame.GetPlayer(player1ID)
	player2, _ := testGame.GetPlayer(player2ID)
	tr1, tr2 := player1.Resources().TerraformRating(), player2.Resources().TerraformRating()
	initialTemp := testGame.GlobalParameters().Temperature()

	for _, id := range []string{player1ID, player2ID} {
		p, _ := testGame.GetPlayer(id)
		testutil.SetPlayerHeat(ctx, p, 8)
		testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, id, 2), "set current turn")
		testutil.AssertNoError(t, convertHeat(t, repo, cardRegistry, testGame, id), "convert heat")
	}

	testutil.AssertEqual(t, initialTemp+4, testGame.GlobalParameters().Temperature(), "both players raised the temperature")
	testutil.AssertEqual(t, tr1+1, player1.Resources().TerraformRating(), "player 1 gains TR for their step")
	testutil.AssertEqual(t, tr2+1, player2.Resources().TerraformRating(), "player 2 gains TR for their step")
}
