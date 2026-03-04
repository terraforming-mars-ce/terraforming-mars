package bot_test

import (
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/test/testutil"
)

func TestIsMyTurn_ActionPhase_MyTurn(t *testing.T) {
	myID := "player-1"
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseAction,
		CurrentTurn:  &myID,
		CurrentPlayer: dto.PlayerDto{
			Status: dto.PlayerStatusActive,
		},
	}

	testutil.AssertTrue(t, bot.IsMyTurn(game, myID), "Should be my turn in action phase")
}

func TestIsMyTurn_ActionPhase_NotMyTurn(t *testing.T) {
	otherID := "player-2"
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseAction,
		CurrentTurn:  &otherID,
		CurrentPlayer: dto.PlayerDto{
			Status: dto.PlayerStatusWaiting,
		},
	}

	testutil.AssertFalse(t, bot.IsMyTurn(game, "player-1"), "Should not be my turn when other player is active")
}

func TestIsMyTurn_StartingSelection(t *testing.T) {
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseStartingSelection,
		CurrentPlayer: dto.PlayerDto{
			Status: dto.PlayerStatusSelectingStartingCards,
		},
	}

	testutil.AssertTrue(t, bot.IsMyTurn(game, "player-1"), "Should be my turn during starting selection")
}

func TestIsMyTurn_ProductionPhase(t *testing.T) {
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseProductionAndCardDraw,
		CurrentPlayer: dto.PlayerDto{
			Status:          dto.PlayerStatusSelectingProductionCards,
			ProductionPhase: &dto.ProductionPhaseDto{},
		},
	}

	testutil.AssertTrue(t, bot.IsMyTurn(game, "player-1"), "Should be my turn during production phase")
}

func TestIsMyTurn_PendingTileSelection(t *testing.T) {
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseAction,
		CurrentPlayer: dto.PlayerDto{
			PendingTileSelection: &dto.PendingTileSelectionDto{},
		},
	}

	testutil.AssertTrue(t, bot.IsMyTurn(game, "player-1"), "Should be my turn with pending tile selection")
}

func TestIsMyTurn_NilGame(t *testing.T) {
	testutil.AssertFalse(t, bot.IsMyTurn(nil, "player-1"), "Should not be my turn with nil game")
}

func TestIsMyTurn_EmptyPlayerID(t *testing.T) {
	game := &dto.GameDto{
		CurrentPhase: dto.GamePhaseAction,
	}
	testutil.AssertFalse(t, bot.IsMyTurn(game, ""), "Should not be my turn with empty player ID")
}

func TestGetPendingActionType_TileSelection(t *testing.T) {
	game := &dto.GameDto{
		CurrentPlayer: dto.PlayerDto{
			PendingTileSelection: &dto.PendingTileSelectionDto{},
		},
	}
	testutil.AssertEqual(t, "tile-selection", bot.GetPendingActionType(game), "Should detect tile-selection")
}

func TestGetPendingActionType_CardSelection(t *testing.T) {
	game := &dto.GameDto{
		CurrentPlayer: dto.PlayerDto{
			PendingCardSelection: &dto.PendingCardSelectionDto{},
		},
	}
	testutil.AssertEqual(t, "card-selection", bot.GetPendingActionType(game), "Should detect card-selection")
}

func TestGetPendingActionType_None(t *testing.T) {
	game := &dto.GameDto{
		CurrentPlayer: dto.PlayerDto{},
	}
	testutil.AssertEqual(t, "", bot.GetPendingActionType(game), "Should return empty when no pending action")
}

func TestGetPendingActionType_NilGame(t *testing.T) {
	testutil.AssertEqual(t, "", bot.GetPendingActionType(nil), "Should return empty for nil game")
}
