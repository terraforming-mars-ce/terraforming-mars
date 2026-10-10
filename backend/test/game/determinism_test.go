package game_test

import (
	"slices"
	"testing"

	"openmars/internal/game"
	"openmars/internal/game/board"
	"openmars/internal/game/colony"
	"openmars/internal/game/shared"
	"openmars/test/testutil"
)

// buildSeededGame reproduces a game from a seed via the replay primitive (which uses
// the production deck-construction path) and returns the shuffled project-card order.
func buildSeededGame(t *testing.T, seed uint64) []string {
	t.Helper()
	g, _, _ := testutil.NewSeededGame(t, seed, shared.GameSettings{})
	return g.Deck().ProjectCards()
}

func TestDeckShuffle_SameSeedIsReproducible(t *testing.T) {
	const seed uint64 = 0xC0FFEE

	first := buildSeededGame(t, seed)
	second := buildSeededGame(t, seed)

	if len(first) == 0 {
		t.Fatal("expected a non-empty project deck")
	}
	if !slices.Equal(first, second) {
		t.Fatalf("same seed produced different deck orders:\n first[:5]=%v\nsecond[:5]=%v",
			first[:min(5, len(first))], second[:min(5, len(second))])
	}
}

func TestDeckShuffle_DifferentSeedsDiffer(t *testing.T) {
	a := buildSeededGame(t, 1)
	b := buildSeededGame(t, 2)

	if slices.Equal(a, b) {
		t.Fatal("different seeds produced identical deck orders (RNG not seed-dependent)")
	}
}

// TestSetupRand_SameSeedIsReproducible guards the setup RNG stream (turn order,
// milestone/award/colony selection) used by StartGameAction.
func TestSetupRand_SameSeedIsReproducible(t *testing.T) {
	repo := testutil.NewTestGameRepository(t)
	settings := shared.GameSettings{MaxPlayers: 4}
	mk := func() []int {
		g := game.NewGame(repo.DataStore(), "setup-test", "", settings, board.GenerateMarsBoard(false))
		g.SetSeed(42)
		return g.SetupRand().Perm(10)
	}
	if !slices.Equal(mk(), mk()) {
		t.Fatal("SetupRand with the same seed produced different sequences")
	}
}

// colonyIDs builds a registry from a fixed input order and returns GetAll's IDs.
func colonyIDs(reg colony.ColonyRegistry) []string {
	all := reg.GetAll()
	ids := make([]string, len(all))
	for i, c := range all {
		ids[i] = c.ID
	}
	return ids
}

// TestColonyRegistry_GetAllPreservesOrder guards the root cause of the colony
// non-determinism: GetAll must return definitions in a stable load order, not
// in Go's randomized map-iteration order. Without a stable order, the seeded
// shuffle in StartGameAction.initializeColonies picks different colonies each
// run even for a fixed seed.
func TestColonyRegistry_GetAllPreservesOrder(t *testing.T) {
	input := []colony.ColonyDefinition{
		{ID: "luna"}, {ID: "triton"}, {ID: "ceres"}, {ID: "enceladus"},
		{ID: "io"}, {ID: "titan"}, {ID: "callisto"}, {ID: "ganymede"},
	}
	want := []string{"luna", "triton", "ceres", "enceladus", "io", "titan", "callisto", "ganymede"}

	// Every construction must yield the exact input order (map iteration would
	// vary run-to-run, so repeating catches the regression probabilistically).
	for i := 0; i < 20; i++ {
		reg := colony.NewInMemoryColonyRegistry(input)
		if got := colonyIDs(reg); !slices.Equal(got, want) {
			t.Fatalf("GetAll did not preserve load order (iteration %d): got %v, want %v", i, got, want)
		}
	}
}

// TestColonySelection_SameSeedIsReproducible exercises the real selection path
// (registry GetAll shuffled by the seed-derived SetupRand) and asserts a fixed
// seed reproduces the same picks while different seeds diverge.
func TestColonySelection_SameSeedIsReproducible(t *testing.T) {
	input := []colony.ColonyDefinition{
		{ID: "luna"}, {ID: "triton"}, {ID: "ceres"}, {ID: "enceladus"},
		{ID: "io"}, {ID: "titan"}, {ID: "callisto"}, {ID: "ganymede"},
	}
	repo := testutil.NewTestGameRepository(t)
	settings := shared.GameSettings{MaxPlayers: 4}

	// selectWithSeed mirrors StartGameAction.initializeColonies: shuffle GetAll
	// with the game's SetupRand, then take the first 5.
	selectWithSeed := func(seed uint64) []string {
		g := game.NewGame(repo.DataStore(), "colony-test", "", settings, board.GenerateMarsBoard(false))
		g.SetSeed(seed)
		rng := g.SetupRand()
		ids := colonyIDs(colony.NewInMemoryColonyRegistry(input))
		rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
		return ids[:5]
	}

	if !slices.Equal(selectWithSeed(0xC0FFEE), selectWithSeed(0xC0FFEE)) {
		t.Fatal("same seed produced different colony selections")
	}
	if slices.Equal(selectWithSeed(1), selectWithSeed(2)) {
		t.Fatal("different seeds produced identical colony selections (selection not seed-dependent)")
	}
}
