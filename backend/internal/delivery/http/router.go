package http

import (
	"net/http"

	gameaction "terraforming-mars-backend/internal/action/game"
	"terraforming-mars-backend/internal/action/query"
	"terraforming-mars-backend/internal/cards"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"
	"terraforming-mars-backend/internal/service/bugreport"

	"github.com/gorilla/mux"
)

// SetupRouter creates HTTP router
// Includes both query (GET) and mutation (POST) endpoints
func SetupRouter(
	createGameAction *gameaction.CreateGameAction,
	createDemoLobbyAction *gameaction.CreateDemoLobbyAction,
	getGameAction *query.GetGameAction,
	getGameLogsAction *query.GetGameLogsAction,
	listGamesAction *query.ListGamesAction,
	listCardsAction *query.ListCardsAction,
	getPlayerAction *query.GetPlayerAction,
	cardRegistry cards.CardRegistry,
	bugReportService *bugreport.Service,
) *mux.Router {
	gameHandler := NewGameHandler(createGameAction, createDemoLobbyAction, getGameAction, getGameLogsAction, listGamesAction, listCardsAction, cardRegistry)
	playerHandler := NewPlayerHandler(getPlayerAction, getGameAction, cardRegistry)
	healthHandler := NewHealthHandler()
	bugReportHandler := NewBugReportHandler(bugReportService)

	router := mux.NewRouter()
	router.Use(httpmiddleware.Recovery)
	router.Use(httpmiddleware.CORS)
	router.Use(httpmiddleware.LoggingMiddleware)
	router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/health", healthHandler.HealthCheck).Methods(http.MethodGet)

	gameRoutes := api.PathPrefix("/games").Subrouter()
	gameRoutes.HandleFunc("", gameHandler.CreateGame).Methods(http.MethodPost)
	gameRoutes.HandleFunc("", gameHandler.ListGames).Methods(http.MethodGet)
	gameRoutes.HandleFunc("/demo/lobby", gameHandler.CreateDemoLobby).Methods(http.MethodPost)
	gameRoutes.HandleFunc("/{gameId}", gameHandler.GetGame).Methods(http.MethodGet)
	gameRoutes.HandleFunc("/{gameId}/logs", gameHandler.GetGameLogs).Methods(http.MethodGet)

	playerRoutes := api.PathPrefix("/games/{gameId}/players").Subrouter()
	playerRoutes.HandleFunc("/{playerId}", playerHandler.GetPlayer).Methods(http.MethodGet)

	api.HandleFunc("/cards", gameHandler.ListCards).Methods(http.MethodGet)

	api.HandleFunc("/bugs", bugReportHandler.SubmitBugReport).Methods(http.MethodPost)
	api.HandleFunc("/bugs/status", bugReportHandler.GetStatus).Methods(http.MethodGet)
	api.HandleFunc("/bugs/{id}", bugReportHandler.GetReport).Methods(http.MethodGet)

	return router
}
