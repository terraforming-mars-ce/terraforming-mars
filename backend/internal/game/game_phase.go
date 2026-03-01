package game

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseCorporationSelection  GamePhase = "corporation_selection"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhasePreludeSelection      GamePhase = "prelude_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseDemoSetup             GamePhase = "demo_setup" // Demo games: players set corp, cards, resources
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseComplete              GamePhase = "complete"
)
