package behavior_test

import (
	"context"
	"testing"

	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestGlobalParameter_TemperatureIncrease(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	trBefore := p.Resources().TerraformRating()
	output := shared.NewGlobalParameterCondition(shared.ResourceTemperature, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1 from temperature raise")
}

func TestGlobalParameter_OxygenIncrease(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	trBefore := p.Resources().TerraformRating()
	output := shared.NewGlobalParameterCondition(shared.ResourceOxygen, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1 from oxygen raise")
}

func TestGlobalParameter_VenusIncrease(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithVenus(t, 2, broadcaster)
	testutil.StartTestGame(t, testGame)
	cardRegistry := testutil.CreateTestCardRegistry()

	p := testGame.GetAllPlayers()[0]

	trBefore := p.Resources().TerraformRating()
	output := shared.NewGlobalParameterCondition(shared.ResourceVenus, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+1, trAfter, "TR should increase by 1 from venus raise")
}

func TestGlobalParameter_TRDirect(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	trBefore := p.Resources().TerraformRating()
	output := shared.NewGlobalParameterCondition(shared.ResourceTR, 2, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+2, trAfter, "TR should increase by exactly 2")
}

func TestGlobalParameter_TRWithPerCondition(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	jovianTag := shared.TagJovian
	selfTarget := "self-player"

	// Add 3 played cards with jovian tag using synthetic cards
	syntheticCards := []gamecards.Card{
		{ID: "jov-1", Name: "Jovian1", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}},
		{ID: "jov-2", Name: "Jovian2", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}},
		{ID: "jov-3", Name: "Jovian3", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(syntheticCards)

	p.PlayedCards().AddCard("jov-1", "Jovian1", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("jov-2", "Jovian2", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("jov-3", "Jovian3", "automated", []string{"jovian"})

	trBefore := p.Resources().TerraformRating()

	output := &shared.GlobalParameterCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceTR, Amount: 1, Target: "none"},
		Per: &shared.PerCondition{
			ResourceType: shared.ResourceType("tag"),
			Amount:       1,
			Target:       &selfTarget,
			Tag:          &jovianTag,
		},
	}
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore+3, trAfter, "TR should increase by 3 (1 per jovian tag)")
}

func TestGlobalParameter_TemperatureAtMax(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)
	ctx := context.Background()

	for testGame.GlobalParameters().Temperature() < global_parameters.MaxTemperature {
		_, err := testGame.GlobalParameters().IncreaseTemperature(ctx, 1, playerID)
		testutil.AssertNoError(t, err, "raising temperature")
	}

	trBefore := p.Resources().TerraformRating()
	output := shared.NewGlobalParameterCondition(shared.ResourceTemperature, 1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore, trAfter, "TR should not change when temperature is already at max")
}

func TestGlobalParameter_NegativeTR(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	trBefore := p.Resources().TerraformRating()
	output := shared.NewGlobalParameterCondition(shared.ResourceTR, -1, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	trAfter := p.Resources().TerraformRating()
	testutil.AssertEqual(t, trBefore-1, trAfter, "TR should decrease by 1")
}
