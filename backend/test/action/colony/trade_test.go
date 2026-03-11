package colony_test

import (
	"context"
	"testing"

	colonyAction "terraforming-mars-backend/internal/action/colony"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

func setupColonyGame(t *testing.T) (*game.Game, game.GameRepository, colonies.ColonyRegistry, string, string) {
	t.Helper()
	testGame, repo, cardRegistry, player1, player2 := testutil.SetupTwoPlayerGame(t)

	colonyDefs, err := colonies.LoadColoniesFromJSON("../../../assets/terraforming_mars_colonies.json")
	if err != nil {
		t.Fatalf("Failed to load colonies: %v", err)
	}
	colonyRegistry := colonies.NewInMemoryColonyRegistry(colonyDefs)

	settings := testGame.Settings()
	settings.CardPacks = append(settings.CardPacks, game.PackColonies)
	testGame.UpdateSettings(context.Background(), settings)

	// Give both players energy for trading
	p1, _ := testGame.GetPlayer(player1)
	p2, _ := testGame.GetPlayer(player2)
	p1.Resources().Add(map[shared.ResourceType]int{shared.ResourceEnergy: 10})
	p2.Resources().Add(map[shared.ResourceType]int{shared.ResourceEnergy: 10})

	// Enable trade fleets
	testGame.SetTradeFleetAvailable(player1, true)
	testGame.SetTradeFleetAvailable(player2, true)

	_ = cardRegistry
	return testGame, repo, colonyRegistry, player1, player2
}

func setupColonyTile(g *game.Game, colonyID string, markerPosition int, playerColonies []string) {
	states := g.ColonyTileStates()
	states = append(states, &colony.TileState{
		DefinitionID:   colonyID,
		MarkerPosition: markerPosition,
		PlayerColonies: playerColonies,
	})
	g.SetColonyTileStates(states)
}

func TestTrade_ImmediateResources_CreditsAdded(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	p, _ := testGame.GetPlayer(playerID)
	creditsBefore := p.Resources().Get().Credits

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Luna should succeed")

	creditsAfter := p.Resources().Get().Credits
	// Luna step 3 gives 7 credits
	testutil.AssertEqual(t, creditsBefore+7, creditsAfter, "Should gain 7 credits from Luna trade at position 3")
}

func TestTrade_ColonyBonusGivenToOwners(t *testing.T) {
	testGame, repo, colonyRegistry, player1, player2 := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Player 2 has a colony on Luna
	setupColonyTile(testGame, "luna", 3, []string{player2})

	p2, _ := testGame.GetPlayer(player2)
	creditsBefore := p2.Resources().Get().Credits

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), player1, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Luna should succeed")

	creditsAfter := p2.Resources().Get().Credits
	// Luna colony bonus is 2 credits
	testutil.AssertEqual(t, creditsBefore+2, creditsAfter, "Colony owner should gain 2 credits bonus")
}

func TestTrade_TraderWithColony_GetsBothIncomeAndBonus(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Trader has a colony on Luna, marker at position 3 (7 credits)
	setupColonyTile(testGame, "luna", 3, []string{playerID})

	p, _ := testGame.GetPlayer(playerID)
	creditsBefore := p.Resources().Get().Credits

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Luna should succeed")

	creditsAfter := p.Resources().Get().Credits
	// Luna step 3 = 7 credits trade income + 2 credits colony bonus = 9 total
	testutil.AssertEqual(t, creditsBefore+9, creditsAfter, "Trader with colony should gain trade income + colony bonus")
}

func TestTrade_CardTargetedResources_CombinedWhenTraderHasColony(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// Trader has a colony on Titan, marker at position 6 (4 floaters)
	setupColonyTile(testGame, "titan", 6, []string{playerID})

	// Give trader a card with floater storage
	p, _ := testGame.GetPlayer(playerID)
	aerialMappersID := testutil.CardID("Aerial Mappers")
	p.PlayedCards().AddCard(aerialMappersID, "Aerial Mappers", "active", []string{"venus"})

	action := colonyAction.NewTradeAction(repo, colonyRegistry, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "titan", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Titan should succeed")

	// Should have a pending selection with combined amount: 4 (trade) + 1 (bonus) = 5
	selection := p.Selection().GetPendingColonyResourceSelection()
	testutil.AssertTrue(t, selection != nil, "Should have pending colony resource selection")
	testutil.AssertEqual(t, 5, selection.Amount, "Pending floaters should be 4 (trade) + 1 (bonus) = 5")
	testutil.AssertEqual(t, "floater", selection.ResourceType, "Resource type should be floater")
	testutil.AssertEqual(t, "trade", selection.Reason, "Reason should be trade for the trader")
}

func TestTrade_CardTargetedResources_TradeOnlyWithoutColony(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// No colonies on Titan, marker at position 6 (4 floaters)
	setupColonyTile(testGame, "titan", 6, nil)

	p, _ := testGame.GetPlayer(playerID)
	aerialMappersID := testutil.CardID("Aerial Mappers")
	p.PlayedCards().AddCard(aerialMappersID, "Aerial Mappers", "active", []string{"venus"})

	action := colonyAction.NewTradeAction(repo, colonyRegistry, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "titan", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Titan should succeed")

	selection := p.Selection().GetPendingColonyResourceSelection()
	testutil.AssertTrue(t, selection != nil, "Should have pending colony resource selection")
	testutil.AssertEqual(t, 4, selection.Amount, "Pending floaters should be 4 (trade only)")
	testutil.AssertEqual(t, "trade", selection.Reason, "Reason should be trade for the trader")
}

func TestTrade_ColonyBonusReason_SetToColonyTax(t *testing.T) {
	testGame, repo, colonyRegistry, player1, player2 := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()
	cardRegistry := testutil.CreateTestCardRegistry()

	// Player 2 has a colony on Titan, player 1 trades
	setupColonyTile(testGame, "titan", 6, []string{player2})

	// Give player 2 a card with floater storage to receive the bonus
	p2, _ := testGame.GetPlayer(player2)
	aerialMappersID := testutil.CardID("Aerial Mappers")
	p2.PlayedCards().AddCard(aerialMappersID, "Aerial Mappers", "active", []string{"venus"})

	action := colonyAction.NewTradeAction(repo, colonyRegistry, cardRegistry, stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), player1, "titan", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Titan should succeed")

	selection := p2.Selection().GetPendingColonyResourceSelection()
	testutil.AssertTrue(t, selection != nil, "Colony owner should have pending colony resource selection")
	testutil.AssertEqual(t, "colony-tax", selection.Reason, "Reason should be colony-tax for non-trader colony owner")
}

func TestTrade_MarkerAtZero_NoTradeIncome(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Ganymede at position 0 gives 0 plants
	setupColonyTile(testGame, "ganymede", 0, nil)

	p, _ := testGame.GetPlayer(playerID)
	plantsBefore := p.Resources().Get().Plants

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "ganymede", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade with Ganymede at position 0 should succeed")

	plantsAfter := p.Resources().Get().Plants
	testutil.AssertEqual(t, plantsBefore, plantsAfter, "Should gain 0 plants at position 0")
}

func TestTrade_ResetsMarkerPosition(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// 2 colonies on Luna, marker at position 5
	setupColonyTile(testGame, "luna", 5, []string{"other-1", "other-2"})

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade should succeed")

	tileState := testGame.GetColonyTileState("luna")
	// Marker resets to number of colonies
	testutil.AssertEqual(t, 2, tileState.MarkerPosition, "Marker should reset to number of colonies (2)")
	testutil.AssertTrue(t, tileState.TradedThisGen, "Colony should be marked as traded")
	testutil.AssertEqual(t, playerID, tileState.TraderID, "Trader ID should be set")
}

func TestTrade_AlreadyTraded_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)
	tileState := testGame.GetColonyTileState("luna")
	tileState.TradedThisGen = true

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertError(t, err, "Should fail when colony already traded")
}

func TestTrade_InsufficientEnergy_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	// Remove energy
	p, _ := testGame.GetPlayer(playerID)
	r := p.Resources().Get()
	r.Energy = 0
	p.Resources().Set(r)

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertError(t, err, "Should fail with insufficient energy")
}

func TestTrade_NoTradeFleet_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)
	testGame.SetTradeFleetAvailable(playerID, false)

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertError(t, err, "Should fail without trade fleet")
}

func TestTrade_DeductsEnergyCost(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	p, _ := testGame.GetPlayer(playerID)
	energyBefore := p.Resources().Get().Energy

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade should succeed")

	energyAfter := p.Resources().Get().Energy
	testutil.AssertEqual(t, energyBefore-3, energyAfter, "Should deduct 3 energy")
}

func TestTrade_ConsumesTradeFleet(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade should succeed")

	testutil.AssertFalse(t, testGame.GetTradeFleetAvailable(playerID), "Trade fleet should be consumed")
}

func TestTrade_MultipleColonyOwners_AllGetBonus(t *testing.T) {
	testGame, repo, colonyRegistry, player1, player2 := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	// Both players have colonies on Io (bonus: 2 heat each)
	setupColonyTile(testGame, "io", 4, []string{player1, player2})

	p1, _ := testGame.GetPlayer(player1)
	p2, _ := testGame.GetPlayer(player2)
	p1Heat := p1.Resources().Get().Heat
	p2Heat := p2.Resources().Get().Heat

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), player1, "io", colonyAction.TradePaymentEnergy)
	testutil.AssertNoError(t, err, "Trade should succeed")

	// Io step 4 = 8 heat trade income, bonus = 2 heat per colony
	p1HeatAfter := p1.Resources().Get().Heat
	p2HeatAfter := p2.Resources().Get().Heat

	// Player 1 (trader + colony owner): 8 heat trade + 2 heat bonus = 10
	testutil.AssertEqual(t, p1Heat+10, p1HeatAfter, "Trader with colony should gain trade income + bonus")
	// Player 2 (colony owner only): 2 heat bonus
	testutil.AssertEqual(t, p2Heat+2, p2HeatAfter, "Non-trader colony owner should gain bonus only")
}

func TestTrade_PayWithCredits(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 20)
	creditsBefore := p.Resources().Get().Credits
	energyBefore := p.Resources().Get().Energy

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentCredits)
	testutil.AssertNoError(t, err, "Trade with credits should succeed")

	testutil.AssertEqual(t, creditsBefore-9+7, p.Resources().Get().Credits, "Should deduct 9 credits cost and gain 7 from Luna")
	testutil.AssertEqual(t, energyBefore, p.Resources().Get().Energy, "Energy should be unchanged")
}

func TestTrade_PayWithTitanium(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	p, _ := testGame.GetPlayer(playerID)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceTitanium: 5})
	titaniumBefore := p.Resources().Get().Titanium
	energyBefore := p.Resources().Get().Energy

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentTitanium)
	testutil.AssertNoError(t, err, "Trade with titanium should succeed")

	testutil.AssertEqual(t, titaniumBefore-3, p.Resources().Get().Titanium, "Should deduct 3 titanium")
	testutil.AssertEqual(t, energyBefore, p.Resources().Get().Energy, "Energy should be unchanged")
}

func TestTrade_InsufficientCredits_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	p, _ := testGame.GetPlayer(playerID)
	testutil.SetPlayerCredits(ctx, p, 5)

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentCredits)
	testutil.AssertError(t, err, "Should fail with insufficient credits")
}

func TestTrade_InsufficientTitanium_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentTitanium)
	testutil.AssertError(t, err, "Should fail with insufficient titanium")
}

func TestTrade_InvalidPaymentType_Fails(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColonyGame(t)
	ctx := context.Background()
	stateRepo := game.NewInMemoryGameStateRepository()
	logger := testutil.TestLogger()

	setupColonyTile(testGame, "luna", 3, nil)

	action := colonyAction.NewTradeAction(repo, colonyRegistry, testutil.CreateTestCardRegistry(), stateRepo, logger)
	err := action.Execute(ctx, testGame.ID(), playerID, "luna", colonyAction.TradePaymentType("invalid"))
	testutil.AssertError(t, err, "Should fail with invalid payment type")
}
