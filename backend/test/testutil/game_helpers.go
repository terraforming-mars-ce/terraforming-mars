package testutil

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// StartTestGame transitions a game from lobby to active status
func StartTestGame(t *testing.T, g *game.Game) {
	t.Helper()
	ctx := context.Background()

	// Set corporations for all players (required before starting)
	players := g.GetAllPlayers()
	for _, p := range players {
		if !p.HasCorporation() {
			p.SetCorporationID(CardID("Tharsis Republic"))
		}
	}

	// Set turn order
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}
	err := g.SetTurnOrder(ctx, playerIDs)
	if err != nil {
		t.Fatalf("Failed to set turn order: %v", err)
	}

	// Update status to active
	err = g.UpdateStatus(ctx, shared.GameStatusActive)
	if err != nil {
		t.Fatalf("Failed to update game status: %v", err)
	}

	// Update phase
	err = g.UpdatePhase(ctx, shared.GamePhaseAction)
	if err != nil {
		t.Fatalf("Failed to update game phase: %v", err)
	}

	// Set current turn
	if len(playerIDs) > 0 {
		err = g.SetCurrentTurn(ctx, playerIDs[0], 2)
		if err != nil {
			t.Fatalf("Failed to set current turn: %v", err)
		}
	}
}

// SetupTwoPlayerGame creates a started 2-player game with card registry.
func SetupTwoPlayerGame(t *testing.T) (*game.Game, game.GameRepository, cards.CardRegistry, string, string) {
	t.Helper()

	broadcaster := NewMockBroadcaster()
	testGame, repo := CreateTestGameWithPlayers(t, 2, broadcaster)
	cardRegistry := CreateTestCardRegistry()
	StartTestGame(t, testGame)

	turnOrder := testGame.TurnOrder()

	return testGame, repo, cardRegistry, turnOrder[0], turnOrder[1]
}

// SetupMultiPlayerGame creates a started N-player game with card registry.
// Returns the game, repo, card registry, and player IDs in turn order.
func SetupMultiPlayerGame(t *testing.T, numPlayers int) (*game.Game, game.GameRepository, cards.CardRegistry, []string) {
	t.Helper()

	broadcaster := NewMockBroadcaster()
	testGame, repo := CreateTestGameWithPlayers(t, numPlayers, broadcaster)
	cardRegistry := CreateTestCardRegistry()
	StartTestGame(t, testGame)

	turnOrder := make([]string, len(testGame.TurnOrder()))
	copy(turnOrder, testGame.TurnOrder())

	return testGame, repo, cardRegistry, turnOrder
}

// SetupSoloGame creates a started 1-player game with unlimited actions.
func SetupSoloGame(t *testing.T) (*game.Game, game.GameRepository, cards.CardRegistry, string) {
	t.Helper()

	broadcaster := NewMockBroadcaster()
	testGame, repo := CreateTestGameWithPlayers(t, 1, broadcaster)
	cardRegistry := CreateTestCardRegistry()
	StartTestGame(t, testGame)

	turnOrder := testGame.TurnOrder()
	playerID := turnOrder[0]
	err := testGame.SetCurrentTurn(context.Background(), playerID, -1)
	if err != nil {
		t.Fatalf("Failed to set solo unlimited actions: %v", err)
	}

	return testGame, repo, cardRegistry, playerID
}

// AddBotToGame adds a bot player to a game in lobby and returns it.
func AddBotToGame(t *testing.T, g *game.Game, repo game.GameRepository, cardRegistry cards.CardRegistry, name, difficulty, speed string) *player.Player {
	t.Helper()
	ctx := context.Background()
	botID := "bot-" + name
	bot, err := g.AddNewBotPlayer(ctx, botID, name, player.BotDifficulty(difficulty), player.BotSpeed(speed))
	if err != nil {
		t.Fatalf("Failed to add bot: %v", err)
	}
	action.SetupPlayerCardStore(bot, g, cardRegistry)
	return bot
}

// SetPlayerHeat sets a player's heat resource
func SetPlayerHeat(ctx context.Context, p *player.Player, amount int) {
	resources := p.Resources().Get()
	resources.Heat = amount
	p.Resources().Set(resources)
}

// GetPlayerHeat gets a player's heat resource
func GetPlayerHeat(p *player.Player) int {
	return p.Resources().Get().Heat
}

// SetPlayerCredits sets a player's credits
func SetPlayerCredits(ctx context.Context, p *player.Player, amount int) {
	resources := p.Resources().Get()
	resources.Credits = amount
	p.Resources().Set(resources)
}

// GetPlayerCredits gets a player's credits
func GetPlayerCredits(p *player.Player) int {
	return p.Resources().Get().Credits
}

// AddPlayerCredits adds credits to a player
func AddPlayerCredits(ctx context.Context, p *player.Player, amount int) {
	changes := map[shared.ResourceType]int{
		shared.ResourceCredit: amount,
	}
	p.Resources().Add(changes)
}

// AddPlayerHeat adds heat to a player
func AddPlayerHeat(ctx context.Context, p *player.Player, amount int) {
	changes := map[shared.ResourceType]int{
		shared.ResourceHeat: amount,
	}
	p.Resources().Add(changes)
}
