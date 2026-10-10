package bot_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"openmars/internal/service/bot"
	"openmars/test/testutil"
)

func TestPersonaCatalog_ShippedCatalogLoads(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	catalog, err := bot.LoadPersonaCatalog(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "assets", "bot", "personas.json"))
	testutil.AssertNoError(t, err, "shipped persona catalog should load")
	name, persona := catalog.AssignIdentity(42, nil)
	testutil.AssertTrue(t, name != "" && persona != "", "an identity should be assigned")
}

func TestPersonaCatalog_RejectsDuplicateNames(t *testing.T) {
	_, err := bot.NewPersonaCatalog([]bot.Persona{
		{ID: "a", Voice: "x", Names: []string{"HAL"}},
		{ID: "b", Voice: "y", Names: []string{"HAL"}},
	})
	testutil.AssertError(t, err, "a name may belong to one persona only")
}

func TestPersonaCatalog_AssignIdentityIsDeterministicAndAvoidsTakenNames(t *testing.T) {
	catalog, err := bot.NewPersonaCatalog([]bot.Persona{
		{ID: "rival", Voice: "x", Names: []string{"SHODAN", "Skynet"}},
		{ID: "optimist", Voice: "y", Names: []string{"Wall-E"}},
	})
	testutil.AssertNoError(t, err, "catalog should be valid")

	name1, persona1 := catalog.AssignIdentity(7, []string{"Alice"})
	name2, persona2 := catalog.AssignIdentity(7, []string{"Alice"})
	testutil.AssertEqual(t, name1, name2, "same seed gives the same name")
	testutil.AssertEqual(t, persona1, persona2, "same seed gives the same persona")

	name, persona := catalog.AssignIdentity(7, []string{"SHODAN", "Skynet"})
	testutil.AssertEqual(t, "Wall-E", name, "only the free name is left")
	testutil.AssertEqual(t, "optimist", persona, "the persona follows the name")

	name, _ = catalog.AssignIdentity(7, []string{"SHODAN", "Skynet", "Wall-E"})
	testutil.AssertTrue(t, name != "SHODAN" && name != "Skynet" && name != "Wall-E", "a numbered name is made when all are taken")
}
