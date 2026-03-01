package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	admin "terraforming-mars-backend/internal/action/admin"
	awardAction "terraforming-mars-backend/internal/action/award"
	cardAction "terraforming-mars-backend/internal/action/card"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	connAction "terraforming-mars-backend/internal/action/connection"
	gameAction "terraforming-mars-backend/internal/action/game"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	query "terraforming-mars-backend/internal/action/query"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/cards"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"
	"terraforming-mars-backend/internal/service/bugreport"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var Version = "localbuild"

func main() {
	// Load env from dev.env if present (does not override existing env vars)
	_ = godotenv.Load("dev.env")

	logLevel := os.Getenv("TM_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	// Initialize logger
	if err := logger.Init(&logLevel); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Shutdown()

	log := logger.Get()
	log.Info("🚀 Starting Terraforming Mars backend server")
	log.Info("Version: " + Version)
	log.Info("Log level set to " + logLevel)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// ========== Initialize Card Registry ==========
	// Get working directory to build absolute path
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory", zap.Error(err))
	}

	cardPath := filepath.Join(wd, "assets", "terraforming_mars_cards.json")
	log.Info("📂 Loading cards from", zap.String("path", cardPath))

	cardData, err := cards.LoadCardsFromJSON(cardPath)
	if err != nil {
		log.Fatal("Failed to load cards", zap.Error(err))
	}
	cardRegistry := cards.NewInMemoryCardRegistry(cardData)
	log.Info("🃏 Card registry initialized", zap.Int("card_count", len(cardData)))

	// ========== Initialize Game Repository (Single Source of Truth) ==========
	gameRepo := game.NewInMemoryGameRepository()
	log.Info("🎮 Game repository initialized")

	// ========== Initialize Game State Repository (Diff Logging) ==========
	stateRepo := game.NewInMemoryGameStateRepository()
	log.Info("📊 Game state repository initialized")

	// ========== Initialize Bug Report Service ==========
	bugReportService := bugreport.NewService(log)
	caps := bugReportService.Capabilities()
	log.Info("🐛 Bug report service",
		zap.Bool("github_app", caps.GitHubApp),
		zap.Bool("claude", caps.Claude))

	// ========== Initialize WebSocket Hub ==========
	hub := core.NewHub()
	log.Info("🔌 WebSocket hub initialized")

	// ========== Initialize Game State Broadcaster (Automatic Broadcasting) ==========
	broadcaster := wsHandler.NewBroadcaster(gameRepo, stateRepo, hub, cardRegistry)
	log.Info("📡 Game state broadcaster initialized (provides automatic broadcasting for all games)")

	// ========== Initialize Game Actions ==========

	// Game lifecycle (5)
	createGameAction := gameAction.NewCreateGameAction(gameRepo, cardRegistry, log)
	createDemoLobbyAction := gameAction.NewCreateDemoLobbyAction(gameRepo, cardRegistry, log)
	joinGameAction := gameAction.NewJoinGameAction(gameRepo, cardRegistry, log)
	confirmDemoSetupAction := gameAction.NewConfirmDemoSetupAction(gameRepo, cardRegistry, log)
	finalScoringAction := gameAction.NewFinalScoringAction(gameRepo, cardRegistry, log)

	// Milestones & Awards (2)
	claimMilestoneAction := milestoneAction.NewClaimMilestoneAction(gameRepo, cardRegistry, stateRepo, log)
	fundAwardAction := awardAction.NewFundAwardAction(gameRepo, cardRegistry, stateRepo, log)

	// Card actions (2)
	playCardAction := cardAction.NewPlayCardAction(gameRepo, cardRegistry, stateRepo, log)
	useCardActionAction := cardAction.NewUseCardActionAction(gameRepo, cardRegistry, stateRepo, log)

	// Standard projects (6)
	launchAsteroidAction := stdprojAction.NewLaunchAsteroidAction(gameRepo, stateRepo, log)
	buildPowerPlantAction := stdprojAction.NewBuildPowerPlantAction(gameRepo, cardRegistry, stateRepo, log)
	buildAquiferAction := stdprojAction.NewBuildAquiferAction(gameRepo, stateRepo, log)
	buildCityAction := stdprojAction.NewBuildCityAction(gameRepo, stateRepo, log)
	plantGreeneryAction := stdprojAction.NewPlantGreeneryAction(gameRepo, stateRepo, log)
	sellPatentsAction := stdprojAction.NewSellPatentsAction(gameRepo, stateRepo, log)

	// Resource conversions (2)
	convertHeatAction := resconvAction.NewConvertHeatToTemperatureAction(gameRepo, cardRegistry, stateRepo, log)
	convertPlantsAction := resconvAction.NewConvertPlantsToGreeneryAction(gameRepo, cardRegistry, stateRepo, log)

	// Tile selection (1)
	selectTileAction := tileAction.NewSelectTileAction(gameRepo, cardRegistry, stateRepo, log)

	// Turn management (5)
	startGameAction := turnAction.NewStartGameAction(gameRepo, log)
	skipActionAction := turnAction.NewSkipActionAction(gameRepo, finalScoringAction, log)
	selectStartingChoicesAction := turnAction.NewSelectStartingChoicesAction(gameRepo, cardRegistry, log)

	// Confirmations (4)
	confirmSellPatentsAction := confirmAction.NewConfirmSellPatentsAction(gameRepo, log)
	confirmProductionCardsAction := confirmAction.NewConfirmProductionCardsAction(gameRepo, cardRegistry, log)
	confirmCardDrawAction := confirmAction.NewConfirmCardDrawAction(gameRepo, cardRegistry, log)
	confirmCardDiscardAction := confirmAction.NewConfirmCardDiscardAction(gameRepo, cardRegistry, log)
	confirmBehaviorChoiceAction := confirmAction.NewConfirmBehaviorChoiceAction(gameRepo, cardRegistry, log)

	// Connection management (3)
	playerDisconnectedAction := connAction.NewPlayerDisconnectedAction(gameRepo, log)
	playerTakeoverAction := connAction.NewPlayerTakeoverAction(gameRepo, cardRegistry, log)
	kickPlayerAction := connAction.NewKickPlayerAction(gameRepo, log)

	// Admin actions (9)
	adminSetPhaseAction := admin.NewSetPhaseAction(gameRepo, log)
	adminSetCurrentTurnAction := admin.NewSetCurrentTurnAction(gameRepo, log)
	adminSetResourcesAction := admin.NewSetResourcesAction(gameRepo, log)
	adminSetProductionAction := admin.NewSetProductionAction(gameRepo, log)
	adminSetGlobalParametersAction := admin.NewSetGlobalParametersAction(gameRepo, log)
	adminGiveCardAction := admin.NewGiveCardAction(gameRepo, cardRegistry, log)
	adminSetCorporationAction := admin.NewSetCorporationAction(gameRepo, cardRegistry, log)
	adminStartTileSelectionAction := admin.NewStartTileSelectionAction(gameRepo, log)
	adminSetTRAction := admin.NewSetTRAction(gameRepo, log)

	// Query actions for HTTP (5)
	getGameAction := query.NewGetGameAction(gameRepo, log)
	getGameLogsAction := query.NewGetGameLogsAction(stateRepo, log)
	listGamesAction := query.NewListGamesAction(gameRepo, log)
	listCardsAction := query.NewListCardsAction(cardRegistry, log)
	getPlayerAction := query.NewGetPlayerAction(gameRepo, log)

	log.Info("✅ All actions initialized")

	// ========== Register WebSocket Handlers ==========
	wsHandler.RegisterHandlers(
		hub,
		broadcaster,
		// Game lifecycle
		createGameAction,
		joinGameAction,
		confirmDemoSetupAction,
		// Card actions
		playCardAction,
		useCardActionAction,
		// Standard projects
		launchAsteroidAction,
		buildPowerPlantAction,
		buildAquiferAction,
		buildCityAction,
		plantGreeneryAction,
		sellPatentsAction,
		// Resource conversions
		convertHeatAction,
		convertPlantsAction,
		// Tile selection
		selectTileAction,
		// Turn management
		startGameAction,
		skipActionAction,
		selectStartingChoicesAction,
		// Confirmations
		confirmSellPatentsAction,
		confirmProductionCardsAction,
		confirmCardDrawAction,
		confirmCardDiscardAction,
		confirmBehaviorChoiceAction,
		// Connection
		playerDisconnectedAction,
		playerTakeoverAction,
		kickPlayerAction,
		// Milestones & Awards
		claimMilestoneAction,
		fundAwardAction,
		// Admin actions
		adminSetPhaseAction,
		adminSetCurrentTurnAction,
		adminSetResourcesAction,
		adminSetProductionAction,
		adminSetGlobalParametersAction,
		adminGiveCardAction,
		adminSetCorporationAction,
		adminStartTileSelectionAction,
		adminSetTRAction,
	)

	log.Info("🎯 WebSocket handlers registered")

	// ========== Start WebSocket Hub ==========
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)
	log.Info("🔌 WebSocket hub running")

	// ========== Setup HTTP Router ==========
	mainRouter := mux.NewRouter()
	mainRouter.Use(httpmiddleware.CORS) // Apply CORS to all routes

	apiRouter := httpHandler.SetupRouter(
		createGameAction,
		createDemoLobbyAction,
		getGameAction,
		getGameLogsAction,
		listGamesAction,
		listCardsAction,
		getPlayerAction,
		cardRegistry,
		bugReportService,
	)

	// Mount API router
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)

	// Create WebSocket handler
	wsHttpHandler := core.NewHandler(hub)

	// Add WebSocket endpoint
	mainRouter.HandleFunc("/ws", wsHttpHandler.ServeWS)

	log.Info("🌐 HTTP routes configured")
	log.Info("   📌 POST /api/v1/games - Create game")
	log.Info("   📌 POST /api/v1/games/demo/lobby - Create demo lobby")
	log.Info("   📌 GET  /api/v1/games - List games")
	log.Info("   📌 GET  /api/v1/games/{gameId} - Get game")
	log.Info("   📌 GET  /api/v1/games/{gameId}/logs - Get game logs")
	log.Info("   📌 GET  /api/v1/cards - List cards")
	log.Info("   📌 GET  /api/v1/games/{gameId}/players/{playerId} - Get player")
	log.Info("   📌 POST /api/v1/bugs - Submit bug report")
	log.Info("   📌 WS   /ws - WebSocket endpoint")
	log.Info("   ℹ️  Game creation available via both HTTP POST and WebSocket 'create-game'")

	// ========== Setup HTTP Server ==========
	server := &http.Server{
		Addr:         ":3001",
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Info("🌍 HTTP server listening on :3001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	log.Info("✅ Server started successfully")

	// Wait for shutdown signal
	<-quit

	log.Info("🛑 Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
	} else {
		log.Info("✅ HTTP server stopped")
	}

	// Cancel WebSocket hub context
	cancel()
	log.Info("✅ WebSocket hub stopped")

	log.Info("✅ Server shutdown complete")
}
