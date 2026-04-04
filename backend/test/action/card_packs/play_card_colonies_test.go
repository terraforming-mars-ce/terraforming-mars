package card_packs_test

import (
	"context"
	"testing"

	cardAction "terraforming-mars-backend/internal/action/card"
	"terraforming-mars-backend/internal/colonies"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/colony"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// =============================================================================
// Helper: set up a game with colonies enabled
// =============================================================================

func setupColoniesGame(t *testing.T) (*game.Game, game.GameRepository, colonies.ColonyRegistry, string, string) {
	t.Helper()
	testGame, repo, _, player1, player2 := testutil.SetupTwoPlayerGame(t)

	colonyDefs, err := colonies.LoadColoniesFromJSON("../../../assets/terraforming_mars_colonies.json")
	if err != nil {
		t.Fatalf("Failed to load colonies: %v", err)
	}
	colonyRegistry := colonies.NewInMemoryColonyRegistry(colonyDefs)

	settings := testGame.Settings()
	settings.CardPacks = append(settings.CardPacks, shared.PackColonies)
	testGame.UpdateSettings(context.Background(), settings)

	return testGame, repo, colonyRegistry, player1, player2
}

func addColony(g *game.Game, colonyID string, markerPosition int, playerColonies []string) {
	states := g.Colonies().States()
	states = append(states, &colony.ColonyState{
		DefinitionID:   colonyID,
		MarkerPosition: markerPosition,
		PlayerColonies: playerColonies,
	})
	g.Colonies().SetStates(states)
}

// =============================================================================
// Titan Floating Launch-Pad (C44, active, colonies)
// Auto: Add 2 floaters to any jovian card.
// Manual choice A: Add 1 floater to any jovian card
// Manual choice B: Spend 1 floater here to trade for free
// =============================================================================

func TestTitanFloatingLaunchPad_AutoBehavior_PlacesFloatersOnSelf(t *testing.T) {
	testGame, repo, _, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Titan Floating Launch-Pad")

	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	// cardStorageTargets: target self (the card being played) for the 2 floaters
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, []string{card.ID}, nil, nil)
	testutil.AssertNoError(t, err, "Titan Floating Launch-Pad should play successfully")
	testutil.AssertTrue(t, p.PlayedCards().Contains(card.ID), "Card should be in played cards")

	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 2, storage, "Card should have 2 floaters from auto behavior")
}

func TestTitanFloatingLaunchPad_ManualAction_AddFloater(t *testing.T) {
	testGame, repo, _, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Titan Floating Launch-Pad")

	// Set up card as already played with 1 floater
	p.PlayedCards().AddCard(card.ID, card.Name, string(card.Type), []string{"jovian"})
	p.Resources().AddToStorage(card.ID, 1)

	// Register manual action (behavior index 1 is the manual action)
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        card.ID,
			CardName:      card.Name,
			BehaviorIndex: 1,
			Behavior:      card.Behaviors[1],
		},
	})

	// Use choice 0: add 1 floater to a jovian card (target self)
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, card.ID, 1, &choiceIndex, []string{card.ID}, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Manual action choice A should succeed")

	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 2, storage, "Card should have 2 floaters (1 original + 1 from action)")
}

func TestTitanFloatingLaunchPad_ManualAction_SpendFloaterForFreeTrade(t *testing.T) {
	testGame, repo, _, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Titan Floating Launch-Pad")

	// Set up card as already played with 2 floaters
	p.PlayedCards().AddCard(card.ID, card.Name, string(card.Type), []string{"jovian"})
	p.Resources().AddToStorage(card.ID, 2)

	// Set up colony tiles and trade fleet for free trade
	addColony(testGame, "luna", 3, nil)
	testGame.Colonies().SetTradeFleetAvailable(playerID, true)

	// Register manual action
	p.Actions().SetActions([]shared.CardAction{
		{
			CardID:        card.ID,
			CardName:      card.Name,
			BehaviorIndex: 1,
			Behavior:      card.Behaviors[1],
		},
	})

	// Use choice 1: spend 1 floater to trade for free
	choiceIndex := 1
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err := useAction.Execute(ctx, testGame.ID(), playerID, card.ID, 1, &choiceIndex, nil, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Manual action choice B (free trade) should succeed")

	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 1, storage, "Card should have 1 floater (2 - 1 spent)")

	// Should have a pending free trade selection
	pendingTrade := p.Selection().GetPendingFreeTradeSelection()
	testutil.AssertTrue(t, pendingTrade != nil, "Should have pending free trade selection")
}

func TestTitanFloatingLaunchPad_UsableOnSameGeneration(t *testing.T) {
	testGame, repo, _, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Titan Floating Launch-Pad")

	// Give player enough credits and extra actions
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceCredit: 100})
	p.Hand().AddCard(card.ID)
	testutil.AssertNoError(t, testGame.SetCurrentTurn(ctx, playerID, 3), "set current turn")

	// Play the card (auto behavior places 2 floaters on self)
	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger)
	payment := cardAction.PaymentRequest{Credits: 18}
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, []string{card.ID}, nil, nil)
	testutil.AssertNoError(t, err, "Card should play successfully")

	storage := p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 2, storage, "Card should have 2 floaters from auto behavior")

	// Now use manual action choice 0: add 1 floater
	choiceIndex := 0
	useAction := cardAction.NewUseCardActionAction(repo, cardRegistry, nil, logger)
	err = useAction.Execute(ctx, testGame.ID(), playerID, card.ID, 1, &choiceIndex, []string{card.ID}, nil, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Manual action should work on same generation as play")

	storage = p.Resources().GetCardStorage(card.ID)
	testutil.AssertEqual(t, 3, storage, "Card should have 3 floaters (2 + 1 from manual action)")
}

// =============================================================================
// Productive Outpost (C30, automated, colonies)
// "Gain all your colony bonuses"
// =============================================================================

func TestProductiveOutpost_NoColonies(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Productive Outpost")

	// Set up colony tiles but player has no colonies
	addColony(testGame, "luna", 3, nil)
	addColony(testGame, "io", 2, nil)

	creditsBefore := p.Resources().Get().Credits
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger, colonyRegistry)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Productive Outpost should play successfully")

	creditsAfter := p.Resources().Get().Credits
	testutil.AssertEqual(t, creditsBefore, creditsAfter, "Credits should not change when player has no colonies")
}

func TestProductiveOutpost_SingleColony_Luna(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Productive Outpost")

	// Player has a colony on Luna (bonus: 2 credits)
	addColony(testGame, "luna", 3, []string{playerID})

	testutil.SetPlayerCredits(ctx, p, 0)
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger, colonyRegistry)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Productive Outpost should play successfully")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Credits, "Should gain 2 credits from Luna colony bonus")
}

func TestProductiveOutpost_MultipleColonies(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Productive Outpost")

	// Player has colonies on Luna (bonus: 2 credits) and Io (bonus: 2 heat)
	addColony(testGame, "luna", 3, []string{playerID})
	addColony(testGame, "io", 2, []string{playerID})

	testutil.SetPlayerCredits(ctx, p, 0)
	p.Resources().Add(map[shared.ResourceType]int{shared.ResourceHeat: 0})
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger, colonyRegistry)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Productive Outpost should play successfully")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Credits, "Should gain 2 credits from Luna colony bonus")
	testutil.AssertEqual(t, 2, resources.Heat, "Should gain 2 heat from Io colony bonus")
}

func TestProductiveOutpost_OnlyOwnColonies(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, otherPlayerID := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Productive Outpost")

	// Player has colony on Luna, other player has colony on Io
	addColony(testGame, "luna", 3, []string{playerID})
	addColony(testGame, "io", 2, []string{otherPlayerID})

	testutil.SetPlayerCredits(ctx, p, 0)
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger, colonyRegistry)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Productive Outpost should play successfully")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Credits, "Should gain 2 credits from own Luna colony only")
	testutil.AssertEqual(t, 0, resources.Heat, "Should not gain heat from other player's Io colony")
}

func TestProductiveOutpost_MultipleColoniesOnDifferentTiles(t *testing.T) {
	testGame, repo, colonyRegistry, playerID, _ := setupColoniesGame(t)
	cardRegistry := testutil.CreateTestCardRegistry()
	logger := testutil.TestLogger()
	ctx := context.Background()

	p, _ := testGame.GetPlayer(playerID)
	card := testutil.GetCardByName("Productive Outpost")

	// Player has colonies on Luna (2 credits), Ceres (2 steel), and Ganymede (1 plant)
	addColony(testGame, "luna", 3, []string{playerID})
	addColony(testGame, "ceres", 2, []string{playerID})
	addColony(testGame, "ganymede", 1, []string{playerID})

	testutil.SetPlayerCredits(ctx, p, 0)
	p.Hand().AddCard(card.ID)

	playCardAction := cardAction.NewPlayCardAction(repo, cardRegistry, nil, logger, colonyRegistry)
	payment := cardAction.PaymentRequest{Credits: 0}
	err := playCardAction.Execute(ctx, testGame.ID(), playerID, card.ID, payment, nil, nil, nil, nil)
	testutil.AssertNoError(t, err, "Productive Outpost should play successfully")

	resources := p.Resources().Get()
	testutil.AssertEqual(t, 2, resources.Credits, "Should gain 2 credits from Luna")
	testutil.AssertEqual(t, 2, resources.Steel, "Should gain 2 steel from Ceres")
	testutil.AssertEqual(t, 1, resources.Plants, "Should gain 1 plant from Ganymede")
}
