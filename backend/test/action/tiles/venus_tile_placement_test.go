package tiles_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func TestMaxwellBase_PlacesOnVenusTile_WhenVenusEnabled(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithVenus(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Maxwell Base")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Resources().AddProduction(map[shared.ResourceType]int{
		shared.ResourceEnergyProduction: 2,
	})
	p.Hand().AddCard(card.ID)
	testGame.GlobalParameters().SetVenus(ctx, 12)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Maxwell Base should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")

	// Venus tile Maxwell Base is at coordinates 100,0,-100
	testutil.AssertTrue(t, len(selection.AvailableHexes) == 1,
		"Should have exactly 1 available hex (the Maxwell Base reserved area on Venus)")
	testutil.AssertEqual(t, "100,0,-100", selection.AvailableHexes[0],
		"Available hex should be the Maxwell Base Venus tile")
}

func TestStratopolis_PlacesOnVenusTile_WhenVenusEnabled(t *testing.T) {
	broadcaster := testutil.NewMockBroadcaster()
	testGame, repo := testutil.CreateTestGameWithVenus(t, 1, broadcaster)
	logger := testutil.TestLogger()
	ctx := context.Background()
	card := testutil.GetCardByName("Stratopolis")
	cardRegistry := testutil.CreateTestCardRegistry()
	players := testGame.GetAllPlayers()
	p := players[0]
	p.SetCorporationID(testutil.CardID("Tharsis Republic"))
	testGame.UpdateStatus(ctx, game.GameStatusActive)
	testGame.UpdatePhase(ctx, game.GamePhaseAction)
	testGame.SetCurrentTurn(ctx, p.ID(), 2)

	// Add science tags via played cards (Stratopolis requires 2 science tags)
	sciCard1 := testutil.GetCardByName("Search For Life")
	sciCard2 := testutil.GetCardByName("Physics Complex")
	p.PlayedCards().AddCard(sciCard1.ID, sciCard1.Name, string(sciCard1.Type), []string{"science"})
	p.PlayedCards().AddCard(sciCard2.ID, sciCard2.Name, string(sciCard2.Type), []string{"science"})
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 22}
	err := playCardAction.Execute(ctx, testGame.ID(), p.ID(), card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Stratopolis should play successfully")

	selection := testGame.GetPendingTileSelection(p.ID())
	testutil.AssertTrue(t, selection != nil, "Should have pending city tile selection")

	// Venus tile Stratopolis is at coordinates 102,0,-102
	testutil.AssertTrue(t, len(selection.AvailableHexes) == 1,
		"Should have exactly 1 available hex (the Stratopolis reserved area on Venus)")
	testutil.AssertEqual(t, "102,0,-102", selection.AvailableHexes[0],
		"Available hex should be the Stratopolis Venus tile")
}
