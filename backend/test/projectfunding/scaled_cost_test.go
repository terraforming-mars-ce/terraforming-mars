package projectfunding_test

import (
	"fmt"
	"testing"

	pf "terraforming-mars-backend/internal/game/projectfunding"
	"terraforming-mars-backend/test/testutil"
)

func TestScaledSeatCost_Baseline(t *testing.T) {
	for _, base := range []int{5, 6, 8, 10, 12, 15, 18, 21} {
		got := pf.ScaledSeatCost(base, pf.SeatCostBaselinePlayers)
		testutil.AssertEqual(t, base, got, fmt.Sprintf("baseline should preserve base cost %d", base))
	}
}

func TestScaledSeatCost_Linear(t *testing.T) {
	tests := []struct {
		base, players, want int
	}{
		{18, 2, 9},
		{18, 3, 13},
		{18, 4, 18},
		{18, 5, 22},
		{10, 2, 5},
		{10, 5, 12},
		{5, 2, 2},
		{5, 5, 6},
	}
	for _, tc := range tests {
		got := pf.ScaledSeatCost(tc.base, tc.players)
		testutil.AssertEqual(t, tc.want, got,
			fmt.Sprintf("ScaledSeatCost(%d, %d)", tc.base, tc.players))
	}
}

func TestScaledSeatCost_Floor(t *testing.T) {
	testutil.AssertEqual(t, 1, pf.ScaledSeatCost(1, 1), "should floor at 1")
	testutil.AssertEqual(t, 1, pf.ScaledSeatCost(3, 1), "3*1/4 would be 0; floored to 1")
}

func TestScaledSeatCost_NonPositive(t *testing.T) {
	testutil.AssertEqual(t, 0, pf.ScaledSeatCost(0, 4), "zero base returns zero")
	testutil.AssertEqual(t, -2, pf.ScaledSeatCost(-2, 4), "negative base returned unchanged")
}

func TestScaledSeatCount_Baseline(t *testing.T) {
	for _, base := range []int{1, 5, 6, 7} {
		got := pf.ScaledSeatCount(base, pf.SeatCostBaselinePlayers)
		testutil.AssertEqual(t, base, got, fmt.Sprintf("baseline should preserve seat count %d", base))
	}
}

func TestScaledSeatCount_RoundsToNearest(t *testing.T) {
	tests := []struct {
		base, players, want int
	}{
		// base=6 (Orbital Station)
		{6, 1, 2}, // 6/4 = 1.5 -> 2
		{6, 2, 3},
		{6, 3, 5}, // 18/4 = 4.5 -> 5
		{6, 5, 8}, // 30/4 = 7.5 -> 8
		// base=5 (Solar Forge)
		{5, 1, 1}, // 5/4 = 1.25 -> 1
		{5, 2, 3}, // 10/4 = 2.5 -> 3
		{5, 3, 4}, // 15/4 = 3.75 -> 4
		{5, 5, 6},
		// base=7 (Terraforming Hub)
		{7, 2, 4}, // 14/4 = 3.5 -> 4
		{7, 3, 5}, // 21/4 = 5.25 -> 5
		{7, 5, 9}, // 35/4 = 8.75 -> 9
	}
	for _, tc := range tests {
		got := pf.ScaledSeatCount(tc.base, tc.players)
		testutil.AssertEqual(t, tc.want, got,
			fmt.Sprintf("ScaledSeatCount(%d, %d)", tc.base, tc.players))
	}
}

func TestScaledSeatCount_Floor(t *testing.T) {
	testutil.AssertEqual(t, 1, pf.ScaledSeatCount(1, 1), "1*1/4 rounded would be 0; floored to 1")
}

func TestScaledSeats_Truncates(t *testing.T) {
	seats := []pf.SeatDefinition{
		{Cost: 6}, {Cost: 8}, {Cost: 10}, {Cost: 12}, {Cost: 15}, {Cost: 18},
	}
	got := pf.ScaledSeats(seats, 2) // 2 players -> 3 seats
	testutil.AssertEqual(t, 3, len(got), "2-player game truncates to 3 seats")
	testutil.AssertEqual(t, 3, got[0].Cost, "scaled cost of first seat")
	testutil.AssertEqual(t, 4, got[1].Cost, "scaled cost of second seat")
	testutil.AssertEqual(t, 5, got[2].Cost, "scaled cost of third seat")
}

func TestScaledSeats_RepeatsLastWhenGrowing(t *testing.T) {
	seats := []pf.SeatDefinition{
		{Cost: 6}, {Cost: 8}, {Cost: 10}, {Cost: 12}, {Cost: 15}, {Cost: 18},
	}
	got := pf.ScaledSeats(seats, 5) // 5 players -> 8 seats
	testutil.AssertEqual(t, 8, len(got), "5-player game grows to 8 seats")
	// First 6 seats reflect JSON costs scaled by 5/4.
	testutil.AssertEqual(t, 7, got[0].Cost, "6*5/4")
	testutil.AssertEqual(t, 22, got[5].Cost, "18*5/4")
	// Extra seats repeat the final entry's scaled cost.
	testutil.AssertEqual(t, 22, got[6].Cost, "extra seat reuses last scaled cost")
	testutil.AssertEqual(t, 22, got[7].Cost, "extra seat reuses last scaled cost")
}

func TestScaledSeats_PreservesPaymentSubstitutes(t *testing.T) {
	seats := []pf.SeatDefinition{
		{Cost: 6, PaymentSubstitutes: []pf.PaymentSubstitute{{ResourceType: "titanium", ConversionRate: 3}}},
		{Cost: 8, PaymentSubstitutes: []pf.PaymentSubstitute{{ResourceType: "titanium", ConversionRate: 3}}},
	}
	got := pf.ScaledSeats(seats, 5)
	for i, s := range got {
		testutil.AssertEqual(t, 1, len(s.PaymentSubstitutes),
			fmt.Sprintf("seat %d preserves payment substitutes", i))
		testutil.AssertEqual(t, 3, s.PaymentSubstitutes[0].ConversionRate,
			fmt.Sprintf("seat %d conversion rate is unchanged", i))
	}
}

func TestScaledSeats_EmptyInput(t *testing.T) {
	got := pf.ScaledSeats(nil, 4)
	testutil.AssertTrue(t, got == nil, "nil input returns nil")
}

func TestScaledRewardTiers_Baseline(t *testing.T) {
	tiers := []pf.RewardTier{
		{SeatsOwned: 2, Rewards: []pf.Output{{Type: "credit", Amount: 5}}},
		{SeatsOwned: 4, Rewards: []pf.Output{{Type: "tr", Amount: 1}}},
	}
	got := pf.ScaledRewardTiers(tiers, pf.SeatCostBaselinePlayers)
	testutil.AssertEqual(t, 2, got[0].SeatsOwned, "baseline preserves first threshold")
	testutil.AssertEqual(t, 4, got[1].SeatsOwned, "baseline preserves second threshold")
}

func TestScaledRewardTiers_Scales(t *testing.T) {
	tiers := []pf.RewardTier{
		{SeatsOwned: 2}, {SeatsOwned: 4}, {SeatsOwned: 6},
	}
	tests := []struct {
		players        int
		wantThresholds []int
	}{
		{2, []int{1, 2, 3}},
		{3, []int{2, 3, 5}}, // 6/4=1.5->2, 12/4=3, 18/4=4.5->5
		{5, []int{3, 5, 8}}, // 10/4=2.5->3, 20/4=5, 30/4=7.5->8
	}
	for _, tc := range tests {
		got := pf.ScaledRewardTiers(tiers, tc.players)
		for i, want := range tc.wantThresholds {
			testutil.AssertEqual(t, want, got[i].SeatsOwned,
				fmt.Sprintf("players=%d tier %d threshold", tc.players, i))
		}
	}
}

func TestScaledRewardTiers_FloorsAtOne(t *testing.T) {
	tiers := []pf.RewardTier{{SeatsOwned: 1}}
	got := pf.ScaledRewardTiers(tiers, 1) // 1*1/4 rounded = 0 -> floor to 1
	testutil.AssertEqual(t, 1, got[0].SeatsOwned, "threshold floored at 1")
}

func TestScaledRewardTiers_Empty(t *testing.T) {
	got := pf.ScaledRewardTiers(nil, 4)
	testutil.AssertTrue(t, got == nil, "nil input returns nil")
}
