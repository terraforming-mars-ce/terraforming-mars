package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	admin "terraforming-mars-backend/internal/action/admin"
	awardAction "terraforming-mars-backend/internal/action/award"
	cardAction "terraforming-mars-backend/internal/action/card"
	colonyAction "terraforming-mars-backend/internal/action/colony"
	confirmAction "terraforming-mars-backend/internal/action/confirmation"
	connAction "terraforming-mars-backend/internal/action/connection"
	gameAction "terraforming-mars-backend/internal/action/game"
	milestoneAction "terraforming-mars-backend/internal/action/milestone"
	pfAction "terraforming-mars-backend/internal/action/projectfunding"
	query "terraforming-mars-backend/internal/action/query"
	resconvAction "terraforming-mars-backend/internal/action/resource_conversion"
	stdprojAction "terraforming-mars-backend/internal/action/standard_project"
	tileAction "terraforming-mars-backend/internal/action/tile"
	turnAction "terraforming-mars-backend/internal/action/turn_management"
	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/colonies"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/datastore"
	"terraforming-mars-backend/internal/logger"
	httpmiddleware "terraforming-mars-backend/internal/middleware/http"
	msLoader "terraforming-mars-backend/internal/milestones"
	pfLoader "terraforming-mars-backend/internal/projectfunding"
	"terraforming-mars-backend/internal/service/bot"
	"terraforming-mars-backend/internal/service/bugreport"
	stdprojLoader "terraforming-mars-backend/internal/standardprojects"

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
	defer func() {
		if err := logger.Shutdown(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown logger: %v\n", err)
		}
	}()

	log := logger.Get()
	log.Info("Starting Terraforming Mars backend server")
	log.Info("Version: " + Version)
	log.Debug("Log level set to " + logLevel)

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
	log.Debug("Loading cards from", zap.String("path", cardPath))

	cardData, err := cards.LoadCardsFromJSON(cardPath)
	if err != nil {
		log.Fatal("Failed to load cards", zap.Error(err))
	}
	cardRegistry := cards.NewInMemoryCardRegistry(cardData)
	log.Debug("Card registry initialized", zap.Int("card_count", len(cardData)))

	// ========== Initialize Colony Registry ==========
	colonyPath := filepath.Join(wd, "assets", "terraforming_mars_colonies.json")
	log.Debug("Loading colonies from", zap.String("path", colonyPath))

	colonyData, err := colonies.LoadColoniesFromJSON(colonyPath)
	if err != nil {
		log.Fatal("Failed to load colonies", zap.Error(err))
	}
	colonyRegistry := colonies.NewInMemoryColonyRegistry(colonyData)
	log.Debug("Colony registry initialized", zap.Int("colony_count", len(colonyData)))

	// ========== Initialize Project Funding Registry ==========
	pfPath := filepath.Join(wd, "assets", "terraforming_mars_project_funding.json")
	log.Debug("Loading project funding from", zap.String("path", pfPath))

	pfData, err := pfLoader.LoadProjectsFromJSON(pfPath)
	if err != nil {
		log.Fatal("Failed to load project funding", zap.Error(err))
	}
	pfRegistry := pfLoader.NewInMemoryProjectFundingRegistry(pfData)
	log.Debug("Project funding registry initialized", zap.Int("project_count", len(pfData)))

	// ========== Initialize Standard Project Registry ==========
	stdProjPath := filepath.Join(wd, "assets", "terraforming_mars_standard_projects.json")
	log.Debug("Loading standard projects from", zap.String("path", stdProjPath))

	stdProjData, err := stdprojLoader.LoadStandardProjectsFromJSON(stdProjPath)
	if err != nil {
		log.Fatal("Failed to load standard projects", zap.Error(err))
	}
	stdProjRegistry := stdprojLoader.NewInMemoryStandardProjectRegistry(stdProjData)
	log.Debug("Standard project registry initialized", zap.Int("project_count", len(stdProjData)))

	// ========== Initialize Award Registry ==========
	awardPath := filepath.Join(wd, "assets", "terraforming_mars_awards.json")
	log.Debug("Loading awards from", zap.String("path", awardPath))

	awardData, err := awards.LoadAwardsFromJSON(awardPath)
	if err != nil {
		log.Fatal("Failed to load awards", zap.Error(err))
	}
	awardRegistry := awards.NewInMemoryAwardRegistry(awardData)
	log.Debug("Award registry initialized", zap.Int("award_count", len(awardData)))

	// ========== Initialize Milestone Registry ==========
	milestonePath := filepath.Join(wd, "assets", "terraforming_mars_milestones.json")
	log.Debug("Loading milestones from", zap.String("path", milestonePath))

	milestoneData, err := msLoader.LoadMilestonesFromJSON(milestonePath)
	if err != nil {
		log.Fatal("Failed to load milestones", zap.Error(err))
	}
	milestoneRegistry := msLoader.NewInMemoryMilestoneRegistry(milestoneData)
	log.Debug("Milestone registry initialized", zap.Int("milestone_count", len(milestoneData)))

	// ========== Initialize Game Repository (Single Source of Truth) ==========
	ds, err := datastore.NewDataStore()
	if err != nil {
		log.Fatal("Failed to create datastore", zap.Error(err))
	}
	rm := datastore.NewRuntimeManager()
	gameRepo := game.NewMemDBGameRepository(ds, rm)
	log.Debug("Game repository initialized")

	// ========== Initialize Game State Repository (Diff Logging) ==========
	stateRepo := game.NewInMemoryGameStateRepository()
	log.Debug("Game state repository initialized")

	// ========== Initialize Bug Report Service ==========
	bugReportService := bugreport.NewService(log)
	caps := bugReportService.Capabilities()
	log.Debug("Feedback service",
		zap.Bool("github_app", caps.GitHubApp))

	// ========== Initialize WebSocket Hub ==========
	hub := core.NewHub()
	log.Debug("WebSocket hub initialized")

	// ========== Initialize Game State Broadcaster (Automatic Broadcasting) ==========
	broadcaster := wsHandler.NewBroadcaster(gameRepo, stateRepo, hub, cardRegistry, colonyRegistry, pfRegistry, stdProjRegistry, awardRegistry, milestoneRegistry)
	log.Debug("Game state broadcaster initialized (provides automatic broadcasting for all games)")

	// ========== Initialize Game Actions ==========

	// Game lifecycle (6)
	createGameAction := gameAction.NewCreateGameAction(gameRepo, cardRegistry, log)
	createDemoLobbyAction := gameAction.NewCreateDemoLobbyAction(gameRepo, cardRegistry, log)
	joinGameAction := gameAction.NewJoinGameAction(gameRepo, cardRegistry, log)
	healthChecker := bot.NewHealthChecker(log)
	addBotAction := gameAction.NewAddBotAction(gameRepo, cardRegistry, healthChecker, broadcaster, log)
	confirmDemoSetupAction := gameAction.NewConfirmDemoSetupAction(gameRepo, cardRegistry, awardRegistry, log)
	finalScoringAction := gameAction.NewFinalScoringAction(gameRepo, cardRegistry, awardRegistry, milestoneRegistry, log)

	// Milestones & Awards (2)
	claimMilestoneAction := milestoneAction.NewClaimMilestoneAction(gameRepo, cardRegistry, stateRepo, milestoneRegistry, log)
	fundAwardAction := awardAction.NewFundAwardAction(gameRepo, cardRegistry, stateRepo, awardRegistry, log)

	// Colony actions (2)
	colonyTradeAction := colonyAction.NewTradeAction(gameRepo, colonyRegistry, cardRegistry, stateRepo, log)
	colonyBuildAction := colonyAction.NewBuildColonyAction(gameRepo, colonyRegistry, cardRegistry, stateRepo, log)

	// Project funding actions (1)
	fundSeatAction := pfAction.NewFundSeatAction(gameRepo, pfRegistry, stateRepo)

	// Card actions (2)
	playCardAction := cardAction.NewPlayCardAction(gameRepo, cardRegistry, stateRepo, log)
	useCardActionAction := cardAction.NewUseCardActionAction(gameRepo, cardRegistry, stateRepo, log)

	// Standard projects (1 unified action)
	executeStandardProjectAction := stdprojAction.NewExecuteStandardProjectAction(gameRepo, cardRegistry, stdProjRegistry, stateRepo, log)

	// Resource conversions (2)
	convertHeatAction := resconvAction.NewConvertHeatToTemperatureAction(gameRepo, cardRegistry, stateRepo, log)
	convertPlantsAction := resconvAction.NewConvertPlantsToGreeneryAction(gameRepo, cardRegistry, stateRepo, log)

	// Tile selection (1)
	selectTileAction := tileAction.NewSelectTileAction(gameRepo, cardRegistry, stateRepo, log)

	// Confirmations (6)
	confirmSellPatentsAction := confirmAction.NewConfirmSellPatentsAction(gameRepo, stateRepo, log)
	confirmProductionCardsAction := confirmAction.NewConfirmProductionCardsAction(gameRepo, cardRegistry, finalScoringAction, log)
	confirmCardDrawAction := confirmAction.NewConfirmCardDrawAction(gameRepo, cardRegistry, log)
	confirmCardDiscardAction := confirmAction.NewConfirmCardDiscardAction(gameRepo, cardRegistry, log)
	confirmBehaviorChoiceAction := confirmAction.NewConfirmBehaviorChoiceAction(gameRepo, cardRegistry, log)
	confirmStealTargetAction := confirmAction.NewConfirmStealTargetAction(gameRepo, cardRegistry, stateRepo, log)
	confirmColonyResourceAction := confirmAction.NewConfirmColonyResourceAction(gameRepo, cardRegistry, stateRepo, log)
	confirmAwardFundAction := confirmAction.NewConfirmAwardFundAction(gameRepo, cardRegistry, awardRegistry, log)
	confirmColonyPlacementAction := confirmAction.NewConfirmColonyPlacementAction(gameRepo, cardRegistry, colonyRegistry, log)
	confirmFreeTradeAction := confirmAction.NewConfirmFreeTradeAction(gameRepo, cardRegistry, colonyRegistry, stateRepo)

	// Turn management (4)
	skipActionAction := turnAction.NewSkipActionAction(gameRepo, finalScoringAction, log)
	selectStartingChoicesAction := turnAction.NewSelectStartingChoicesAction(gameRepo, cardRegistry, awardRegistry, log)
	confirmInitAdvanceAction := turnAction.NewConfirmInitAdvanceAction(gameRepo, cardRegistry, awardRegistry, stateRepo, log)

	// Bot service
	commandDispatcher := bot.NewCommandDispatcher(
		playCardAction, useCardActionAction,
		skipActionAction, selectStartingChoicesAction,
		selectTileAction,
		confirmProductionCardsAction, confirmCardDrawAction,
		confirmCardDiscardAction, confirmBehaviorChoiceAction,
		confirmSellPatentsAction,
		executeStandardProjectAction,
		convertHeatAction, convertPlantsAction,
		claimMilestoneAction, fundAwardAction,
		confirmInitAdvanceAction,
		log,
	)
	botController := bot.NewBotController(gameRepo, stateRepo, cardRegistry, commandDispatcher, broadcaster, log)
	broadcaster.SetBotNotifier(botController)

	startGameAction := turnAction.NewStartGameAction(gameRepo, colonyRegistry, pfRegistry, botController, log)

	// Game management (convert to bot)
	convertToBotAction := gameAction.NewConvertToBotAction(gameRepo, botController, log)

	// Connection management (4)
	playerDisconnectedAction := connAction.NewPlayerDisconnectedAction(gameRepo, log)
	playerTakeoverAction := connAction.NewPlayerTakeoverAction(gameRepo, cardRegistry, log)
	kickPlayerAction := connAction.NewKickPlayerAction(gameRepo, botController, finalScoringAction, log)
	endGameAction := connAction.NewEndGameAction(gameRepo, botController, log)

	// Spectator & chat actions (5)
	setPlayerColorAction := connAction.NewSetPlayerColorAction(gameRepo, log)
	spectateGameAction := connAction.NewSpectateGameAction(gameRepo, log)
	spectatorDisconnectedAction := connAction.NewSpectatorDisconnectedAction(gameRepo, log)
	kickSpectatorAction := connAction.NewKickSpectatorAction(gameRepo, log)
	sendChatMessageAction := connAction.NewSendChatMessageAction(gameRepo, log)

	// Admin actions (9)
	adminSetPhaseAction := admin.NewSetPhaseAction(gameRepo, log)
	adminSetCurrentTurnAction := admin.NewSetCurrentTurnAction(gameRepo, log)
	adminSetResourcesAction := admin.NewSetResourcesAction(gameRepo, log)
	adminSetProductionAction := admin.NewSetProductionAction(gameRepo, log)
	adminSetGlobalParametersAction := admin.NewSetGlobalParametersAction(gameRepo, log)
	adminGiveCardAction := admin.NewGiveCardAction(gameRepo, cardRegistry, log)
	adminSetCorporationAction := admin.NewSetCorporationAction(gameRepo, cardRegistry, awardRegistry, log)
	adminStartTileSelectionAction := admin.NewStartTileSelectionAction(gameRepo, log)
	adminSetTRAction := admin.NewSetTRAction(gameRepo, log)

	// Query actions for HTTP (6)
	getGameAction := query.NewGetGameAction(gameRepo, log)
	getGameLogsAction := query.NewGetGameLogsAction(stateRepo, log)
	getGameHistoryAction := query.NewGetGameHistoryAction(ds, log)
	listGamesAction := query.NewListGamesAction(gameRepo, log)
	listCardsAction := query.NewListCardsAction(cardRegistry, log)
	getPlayerAction := query.NewGetPlayerAction(gameRepo, log)

	log.Debug("All actions initialized")

	// ========== Register WebSocket Handlers ==========
	wsHandler.RegisterHandlers(
		hub,
		broadcaster,
		gameRepo,
		// Game lifecycle
		createGameAction,
		joinGameAction,
		addBotAction,
		confirmDemoSetupAction,
		// Card actions
		playCardAction,
		useCardActionAction,
		// Standard projects
		executeStandardProjectAction,
		// Resource conversions
		convertHeatAction,
		convertPlantsAction,
		// Tile selection
		selectTileAction,
		// Turn management
		startGameAction,
		skipActionAction,
		selectStartingChoicesAction,
		confirmInitAdvanceAction,
		// Confirmations
		confirmSellPatentsAction,
		confirmProductionCardsAction,
		confirmCardDrawAction,
		confirmCardDiscardAction,
		confirmBehaviorChoiceAction,
		confirmStealTargetAction,
		confirmColonyResourceAction,
		confirmAwardFundAction,
		confirmColonyPlacementAction,
		confirmFreeTradeAction,
		// Connection
		playerDisconnectedAction,
		playerTakeoverAction,
		kickPlayerAction,
		endGameAction,
		// Spectator & chat
		setPlayerColorAction,
		spectateGameAction,
		spectatorDisconnectedAction,
		kickSpectatorAction,
		sendChatMessageAction,
		// Convert to bot
		convertToBotAction,
		// Milestones & Awards
		claimMilestoneAction,
		fundAwardAction,
		// Colony actions
		colonyTradeAction,
		colonyBuildAction,
		// Project funding actions
		fundSeatAction,
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

	log.Debug("WebSocket handlers registered")

	// ========== Start WebSocket Hub ==========
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)
	log.Debug("WebSocket hub running")

	// ========== Setup HTTP Router ==========
	mainRouter := mux.NewRouter()
	mainRouter.Use(httpmiddleware.CORS) // Apply CORS to all routes

	apiRouter := httpHandler.SetupRouter(
		createGameAction,
		createDemoLobbyAction,
		getGameAction,
		getGameLogsAction,
		getGameHistoryAction,
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

	log.Debug("HTTP routes configured")

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
		log.Info("HTTP server listening on :3001")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	log.Info("Server started")

	// Wait for shutdown signal
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
	} else {
		log.Debug("HTTP server stopped")
	}

	// Cancel WebSocket hub context
	cancel()
	log.Debug("WebSocket hub stopped")

	log.Info("Server shutdown complete")
}
