package game

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingSelection     GamePhase = "starting_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseDemoSetup             GamePhase = "demo_setup"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseComplete              GamePhase = "complete"
)
