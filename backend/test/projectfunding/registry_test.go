package projectfunding_test

import (
	"testing"

	pfLoader "terraforming-mars-backend/internal/projectfunding"
	"terraforming-mars-backend/test/testutil"
)

func TestRegistry_GetByID_Found(t *testing.T) {
	defs, _ := pfLoader.LoadProjectsFromJSON(jsonPath)
	registry := pfLoader.NewInMemoryProjectFundingRegistry(defs)

	def, err := registry.GetByID(defs[0].ID)
	testutil.AssertNoError(t, err, "Should find project by ID")
	testutil.AssertEqual(t, defs[0].Name, def.Name, "Name should match")
}

func TestRegistry_GetByID_NotFound(t *testing.T) {
	defs, _ := pfLoader.LoadProjectsFromJSON(jsonPath)
	registry := pfLoader.NewInMemoryProjectFundingRegistry(defs)

	_, err := registry.GetByID("nonexistent_id")
	testutil.AssertError(t, err, "Should fail for unknown ID")
}

func TestRegistry_GetAll(t *testing.T) {
	defs, _ := pfLoader.LoadProjectsFromJSON(jsonPath)
	registry := pfLoader.NewInMemoryProjectFundingRegistry(defs)

	all := registry.GetAll()
	testutil.AssertEqual(t, len(defs), len(all), "GetAll should return all projects")
}

func TestRegistry_GetAll_ReturnsDefensiveCopy(t *testing.T) {
	defs, _ := pfLoader.LoadProjectsFromJSON(jsonPath)
	registry := pfLoader.NewInMemoryProjectFundingRegistry(defs)

	all1 := registry.GetAll()
	originalLen := len(all1)

	all1 = all1[:0]

	all2 := registry.GetAll()
	testutil.AssertEqual(t, originalLen, len(all2), "Modifying returned slice should not affect registry")
}
