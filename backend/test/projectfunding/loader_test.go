package projectfunding_test

import (
	"testing"

	pfLoader "terraforming-mars-backend/internal/projectfunding"
	"terraforming-mars-backend/test/testutil"
)

const jsonPath = "../../assets/terraforming_mars_project_funding.json"

func TestLoadProjectsFromJSON_Success(t *testing.T) {
	defs, err := pfLoader.LoadProjectsFromJSON(jsonPath)
	testutil.AssertNoError(t, err, "Should load projects from JSON")
	testutil.AssertTrue(t, len(defs) >= 1, "Should have at least 1 project")
}

func TestLoadProjectsFromJSON_InvalidPath(t *testing.T) {
	_, err := pfLoader.LoadProjectsFromJSON("/nonexistent/path.json")
	testutil.AssertError(t, err, "Should fail with invalid path")
}

func TestLoadProjectsFromJSON_ValidatesStructure(t *testing.T) {
	defs, err := pfLoader.LoadProjectsFromJSON(jsonPath)
	testutil.AssertNoError(t, err, "Should load projects")

	for _, def := range defs {
		testutil.AssertTrue(t, def.ID != "", "Project ID should not be empty")
		testutil.AssertTrue(t, def.Name != "", "Project name should not be empty")
		testutil.AssertTrue(t, len(def.Seats) > 0, "Project should have at least 1 seat")
		testutil.AssertTrue(t, len(def.RewardTiers) > 0, "Project should have at least 1 reward tier")
		hasCompletionEffect := len(def.CompletionEffect.Rewards) > 0 || len(def.CompletionEffect.GlobalEffects) > 0
		testutil.AssertTrue(t, hasCompletionEffect, "Project should have completion rewards or global effects")
	}
}

func TestLoadProjectsFromJSON_SeatCostsEscalate(t *testing.T) {
	defs, err := pfLoader.LoadProjectsFromJSON(jsonPath)
	testutil.AssertNoError(t, err, "Should load projects")

	for _, def := range defs {
		for i := 1; i < len(def.Seats); i++ {
			testutil.AssertTrue(t, def.Seats[i].Cost >= def.Seats[i-1].Cost,
				"Seat costs should be non-decreasing for project "+def.ID)
		}
	}
}

func TestLoadProjectsFromJSON_UniqueIDs(t *testing.T) {
	defs, err := pfLoader.LoadProjectsFromJSON(jsonPath)
	testutil.AssertNoError(t, err, "Should load projects")

	seen := map[string]bool{}
	for _, def := range defs {
		testutil.AssertFalse(t, seen[def.ID], "Duplicate project ID: "+def.ID)
		seen[def.ID] = true
	}
}

func TestLoadProjectsFromJSON_RewardTiersAscending(t *testing.T) {
	defs, err := pfLoader.LoadProjectsFromJSON(jsonPath)
	testutil.AssertNoError(t, err, "Should load projects")

	for _, def := range defs {
		for i := 1; i < len(def.RewardTiers); i++ {
			testutil.AssertTrue(t, def.RewardTiers[i].SeatsOwned > def.RewardTiers[i-1].SeatsOwned,
				"Reward tiers should have ascending seatsOwned for project "+def.ID)
		}
	}
}

func TestLoadProjectsFromJSON_StyleFieldsPresent(t *testing.T) {
	defs, err := pfLoader.LoadProjectsFromJSON(jsonPath)
	testutil.AssertNoError(t, err, "Should load projects")

	for _, def := range defs {
		testutil.AssertTrue(t, def.Style.Color != "", "Style color should not be empty for "+def.ID)
		testutil.AssertTrue(t, def.Style.Icon != "", "Style icon should not be empty for "+def.ID)
	}
}
