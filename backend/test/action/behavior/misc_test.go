package behavior_test

import (
	"testing"

	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestMisc_ExtraActionsGranted(t *testing.T) {
	testGame, _, cardRegistry, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	actionsBefore := testGame.CurrentTurn().ActionsRemaining()
	output := shared.NewMiscCondition(shared.ResourceExtraActions, 2, "none")
	applyOutputs(t, p, testGame, cardRegistry, output)

	actionsAfter := testGame.CurrentTurn().ActionsRemaining()
	testutil.AssertEqual(t, actionsBefore+2, actionsAfter, "remaining actions should increase by 2")
}

func TestMisc_BonusTagsWithPer(t *testing.T) {
	testGame, _, _, playerID, _ := testutil.SetupTwoPlayerGame(t)
	p, _ := testGame.GetPlayer(playerID)

	jovianTag := shared.TagJovian

	// Add 3 played cards with jovian tag using synthetic cards
	syntheticCards := []gamecards.Card{
		{ID: "jov-a", Name: "JovianA", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}},
		{ID: "jov-b", Name: "JovianB", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}},
		{ID: "jov-c", Name: "JovianC", Type: gamecards.CardTypeAutomated, Tags: []shared.CardTag{shared.TagJovian}},
	}
	cardRegistry := testutil.CreateTestCardRegistryWithAdditionalCards(syntheticCards)

	p.PlayedCards().AddCard("jov-a", "JovianA", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("jov-b", "JovianB", "automated", []string{"jovian"})
	p.PlayedCards().AddCard("jov-c", "JovianC", "automated", []string{"jovian"})

	// The MiscCondition Per field tells the handler which tag type to count and grant.
	// The behavior applier's general Per scaling also applies, so the final bonus count is
	// (tag_count * amount) computed by the general scaler, then multiplied again by tag_count
	// inside applyMiscOutput. With 3 jovian tags and amount=1: scaled_amount=3, bonusCount=3*3=9.
	output := &shared.MiscCondition{
		ConditionBase: shared.ConditionBase{ResourceType: shared.ResourceBonusTags, Amount: 1, Target: "none"},
		Per: &shared.PerCondition{
			ResourceType: shared.ResourceType("tag"),
			Amount:       1,
			Tag:          &jovianTag,
		},
		Selectors: []shared.Selector{
			{Tags: []shared.CardTag{shared.TagJovian}},
		},
	}
	applyOutputs(t, p, testGame, cardRegistry, output)

	bonusTags := p.BonusTags()
	// General Per scaling (3) * misc handler's internal counting (3) = 9
	testutil.AssertEqual(t, 9, bonusTags[shared.TagJovian], "player should have bonus jovian tags from Per scaling")
}

func TestMisc_FreeTrade(t *testing.T) {
	// TODO: Free trade requires colonies expansion (CreateTestGameWithColonies helper does not exist).
	// Setting up colonies manually is complex (colony tile states, trade fleets, etc.).
	// This test should be implemented when a colony test helper becomes available.
	t.Skip("requires colonies expansion setup")
}
