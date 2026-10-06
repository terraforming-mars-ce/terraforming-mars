package bot_test

import (
	"strings"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"
)

func TestSummarizeGameState_ShowsWhatTheBotNeeds(t *testing.T) {
	myID := "bot"
	owner := "human"
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseAction,
		CurrentTurn:  &myID,
		CurrentPlayer: dto.PlayerDto{
			ID:   myID,
			Name: "HAL",
			Cards: []dto.PlayerCardDto{{
				ID: "c1", Name: "Mining Rights", Cost: 9, EffectiveCost: 9, Available: true,
				Description: []dto.CardDescriptionSectionDto{{Text: "Gain 1 steel production"}},
			}},
			PendingFreeTradeSelection: &dto.PendingFreeTradeSelectionDto{AvailableColonyIDs: []string{"luna"}, Source: "Titan Floating Launch-pad"},
		},
		OtherPlayers: []dto.OtherPlayerDto{{ID: owner, Name: "Alice"}},
		Board: dto.BoardDto{Tiles: []dto.TileDto{
			{Coordinates: dto.HexPositionDto{Q: 0, R: 0, S: 0}, Type: "land-tile", Location: "mars"},
			{Coordinates: dto.HexPositionDto{Q: 1, R: -1, S: 0}, Type: "ocean-tile", Location: "mars"},
			{Coordinates: dto.HexPositionDto{Q: 2, R: -2, S: 0}, Type: "land-tile", Location: "mars", OccupiedBy: &dto.TileOccupantDto{Type: "city"}, OwnerID: &owner},
		}},
		Colonies: []dto.ColonyDto{{
			Active: true, ID: "luna", Name: "Luna", MarkerPosition: 3, TradeAvailable: true,
			TradeOptions: []dto.ColonyTradeOptionDto{{TrackSteps: 0, MarkerPosition: 3, Outputs: []dto.ColonyOutputDto{{Type: "credit", Amount: 7}}}},
		}},
		TradeFleets: map[string]dto.TradeFleetDto{myID: {Total: 1, Available: 1}},
		ProjectFunding: []dto.ProjectFundingDto{{
			ID: "pf1", Name: "Space Elevator Fund", Description: "Shared project", NextSeatCost: 10, CanBuySeat: true,
		}},
	}

	summary := bot.SummarizeGameState(game, myID)
	for _, want := range []string{
		"Gain 1 steel production",
		"Free land-tile spaces: 0,0,0",
		"Free ocean-tile spaces: 1,-1,0",
		"2,-2,0 | city (owner: Alice)",
		"=== COLONIES ===",
		"Your trade fleets: 1 available of 1",
		"trackSteps=0 -> marker 3, you gain: 7 credit",
		"=== PROJECT FUNDING ===",
		"project_fund_seat",
		"FREE TRADE from Titan Floating Launch-pad",
	} {
		testutil.AssertTrue(t, strings.Contains(summary, want), "summary should contain: "+want)
	}
	testutil.AssertFalse(t, strings.Contains(summary, "action.confirm"), "no old command names remain")
}

func TestSnapshotLog_NamesTheActorAndWhoWasHurt(t *testing.T) {
	myID, humanID := "bot", "human"
	snap := &bot.Snapshot{
		PlayerID: myID,
		View: dto.GameDto{
			CurrentPlayer: dto.PlayerDto{ID: myID, Name: "Wall-E"},
			OtherPlayers:  []dto.OtherPlayerDto{{ID: humanID, Name: "Saffron"}},
		},
		Log: []game.StateDiff{{
			PlayerID:    humanID,
			Description: "Played Giant Ice Asteroid for 36 credits",
			Changes: &game.GameChanges{
				Temperature: &game.DiffValueInt{Old: -30, New: -26},
				PlayerChanges: map[string]*game.PlayerChanges{
					myID: {Plants: &game.DiffValueInt{Old: 6, New: 0}},
				},
			},
		}},
	}

	text := snap.Describe()
	testutil.AssertTrue(t, strings.Contains(text, "Saffron: Played Giant Ice Asteroid"), "the actor is named")
	testutil.AssertTrue(t, strings.Contains(text, "you (Wall-E) plants -6"), "the bot sees its own loss")
	testutil.AssertTrue(t, strings.Contains(text, "temperature +4"), "global changes are listed")
}

func TestSummarizeGameState_ExplainsBlockedMilestones(t *testing.T) {
	myID, otherID := "bot", "human"
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseAction,
		CurrentPlayer: dto.PlayerDto{ID: myID, Name: "CASE", Milestones: []dto.PlayerMilestoneDto{
			{Type: "tactician", Name: "Tactician", Progress: 6, Required: 5, ClaimCost: 8, Errors: []dto.StateErrorDto{{Message: "Insufficient credits"}}},
			{Type: "mayor", Name: "Mayor", IsClaimed: true, ClaimedBy: &otherID},
		}},
		OtherPlayers: []dto.OtherPlayerDto{{ID: otherID, Name: "Saffron"}},
	}
	summary := bot.SummarizeGameState(game, myID)
	testutil.AssertTrue(t, strings.Contains(summary, "QUALIFIED but blocked (Insufficient credits); costs 8M€"), "blocked reason is shown")
	testutil.AssertTrue(t, strings.Contains(summary, "CLAIMED by Saffron"), "claimer is named")
}
