package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	admin "openmars/internal/action/admin"
	awardAction "openmars/internal/action/award"
	cardAction "openmars/internal/action/card"
	colonyAction "openmars/internal/action/colony"
	confirmAction "openmars/internal/action/confirmation"
	connAction "openmars/internal/action/connection"
	gameAction "openmars/internal/action/game"
	milestoneAction "openmars/internal/action/milestone"
	pfAction "openmars/internal/action/projectfunding"
	query "openmars/internal/action/query"
	resconvAction "openmars/internal/action/resource_conversion"
	stdprojAction "openmars/internal/action/standard_project"
	tileAction "openmars/internal/action/tile"
	turnAction "openmars/internal/action/turn_management"
	"openmars/internal/changelog"
	"openmars/internal/delivery/dto"
	httpHandler "openmars/internal/delivery/http"
	"openmars/internal/delivery/web"
	wsHandler "openmars/internal/delivery/websocket"
	"openmars/internal/delivery/websocket/core"
	"openmars/internal/game"
	"openmars/internal/game/award"
	"openmars/internal/game/board"
	"openmars/internal/game/cards"
	"openmars/internal/game/colony"
	"openmars/internal/game/datastore"
	msLoader "openmars/internal/game/milestone"
	pfLoader "openmars/internal/game/projectfunding"
	"openmars/internal/game/shared"
	stdprojLoader "openmars/internal/game/standardproject"
	"openmars/internal/logger"
	httpmiddleware "openmars/internal/middleware/http"
	"openmars/internal/service/bot"
	"openmars/internal/service/bugreport"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var Version = "localbuild"

var serverAliasPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// loadServerMeta reads the identity a gateway uses to list and verify this server.
func loadServerMeta() (dto.MetaResponse, error) {
	alias := os.Getenv("OPENMARS_SERVER_ALIAS")
	if alias == "" {
		alias = "local"
	}
	if !serverAliasPattern.MatchString(alias) {
		return dto.MetaResponse{}, fmt.Errorf("OPENMARS_SERVER_ALIAS %q must match %s", alias, serverAliasPattern)
	}
	name := os.Getenv("OPENMARS_SERVER_NAME")
	if name == "" {
		name = alias
	}
	return dto.MetaResponse{Alias: alias, Name: name, Version: Version}, nil
}

func main() {
	// Load env from dev.env if present (does not override existing env vars)
	_ = godotenv.Load("dev.env")

	logLevel := os.Getenv("OPENMARS_LOG_LEVEL")
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
	log.Info("Starting Open Mars backend server")
	log.Info("Version: " + Version)
	log.Debug("Log level set to " + logLevel)

	addr := os.Getenv("OPENMARS_ADDR")
	if addr == "" {
		addr = ":3001"
	}
	webDir := os.Getenv("OPENMARS_WEB_DIR")
	if webDir == "" {
		webDir = "web"
	}

	serverMeta, err := loadServerMeta()
	if err != nil {
		log.Error("Invalid server identity", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Server alias: " + serverMeta.Alias)

	changelogDir := os.Getenv("OPENMARS_CHANGELOG_DIR")
	if changelogDir == "" {
		changelogDir = "changelog"
	}
	changelogResponse := dto.ToChangelogResponse(nil)
	if _, err := os.Stat(changelogDir); err != nil {
		log.Warn("Changelog directory not found, serving an empty changelog", slog.String("path", changelogDir))
	} else {
		entries, err := changelog.LoadAll(changelogDir)
		if err != nil {
			log.Error("Failed to load changelog", slog.Any("error", err))
			os.Exit(1)
		}
		changelogResponse = dto.ToChangelogResponse(entries)
		log.Debug("Changelog loaded", slog.Int("entry_count", len(entries)))
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// ========== Initialize Card Registry ==========
	// Get working directory to build absolute path
	wd, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get working directory", slog.Any("error", err))
		os.Exit(1)
	}

	cardPath := filepath.Join(wd, "assets", "cards.json")
	log.Debug("Loading cards from", slog.String("path", cardPath))

	cardData, err := cards.LoadCardsFromJSON(cardPath)
	if err != nil {
		log.Error("Failed to load cards", slog.Any("error", err))
		os.Exit(1)
	}
	cardRegistry := cards.NewInMemoryCardRegistry(cardData)
	log.Debug("Card registry initialized", slog.Int("card_count", len(cardData)))

	// ========== Initialize Colony Registry ==========
	colonyPath := filepath.Join(wd, "assets", "colonies.json")
	log.Debug("Loading colonies from", slog.String("path", colonyPath))

	colonyData, err := colony.LoadColoniesFromJSON(colonyPath)
	if err != nil {
		log.Error("Failed to load colonies", slog.Any("error", err))
		os.Exit(1)
	}
	colonyRegistry := colony.NewInMemoryColonyRegistry(colonyData)
	log.Debug("Colony registry initialized", slog.Int("colony_count", len(colonyData)))

	// ========== Initialize Project Funding Registry ==========
	pfPath := filepath.Join(wd, "assets", "project_funding.json")
	log.Debug("Loading project funding from", slog.String("path", pfPath))

	pfData, err := pfLoader.LoadProjectsFromJSON(pfPath)
	if err != nil {
		log.Error("Failed to load project funding", slog.Any("error", err))
		os.Exit(1)
	}
	pfRegistry := pfLoader.NewInMemoryProjectFundingRegistry(pfData)
	log.Debug("Project funding registry initialized", slog.Int("project_count", len(pfData)))

	// ========== Initialize Standard Project Registry ==========
	stdProjPath := filepath.Join(wd, "assets", "standard_projects.json")
	log.Debug("Loading standard projects from", slog.String("path", stdProjPath))

	stdProjData, err := stdprojLoader.LoadStandardProjectsFromJSON(stdProjPath)
	if err != nil {
		log.Error("Failed to load standard projects", slog.Any("error", err))
		os.Exit(1)
	}
	stdProjRegistry := stdprojLoader.NewInMemoryStandardProjectRegistry(stdProjData)
	log.Debug("Standard project registry initialized", slog.Int("project_count", len(stdProjData)))

	// ========== Initialize Award Registry ==========
	awardPath := filepath.Join(wd, "assets", "awards.json")
	log.Debug("Loading awards from", slog.String("path", awardPath))

	awardData, err := award.LoadAwardsFromJSON(awardPath)
	if err != nil {
		log.Error("Failed to load awards", slog.Any("error", err))
		os.Exit(1)
	}
	awardRegistry := award.NewInMemoryAwardRegistry(awardData)
	log.Debug("Award registry initialized", slog.Int("award_count", len(awardData)))

	// ========== Initialize Milestone Registry ==========
	milestonePath := filepath.Join(wd, "assets", "milestones.json")
	log.Debug("Loading milestones from", slog.String("path", milestonePath))

	milestoneData, err := msLoader.LoadMilestonesFromJSON(milestonePath)
	if err != nil {
		log.Error("Failed to load milestones", slog.Any("error", err))
		os.Exit(1)
	}
	milestoneRegistry := msLoader.NewInMemoryMilestoneRegistry(milestoneData)
	log.Debug("Milestone registry initialized", slog.Int("milestone_count", len(milestoneData)))

	// ========== Initialize Map Registry ==========
	mapPath := filepath.Join(wd, "assets", "maps.json")
	log.Debug("Loading maps from", slog.String("path", mapPath))

	mapRegistry, err := board.LoadMapsFromJSON(mapPath)
	if err != nil {
		log.Error("Failed to load maps", slog.Any("error", err))
		os.Exit(1)
	}
	log.Debug("Map registry initialized", slog.Int("map_count", len(mapRegistry.ListMaps())))

	// ========== Initialize Game Repository (Single Source of Truth) ==========
	ds, err := datastore.NewDataStore()
	if err != nil {
		log.Error("Failed to create datastore", slog.Any("error", err))
		os.Exit(1)
	}
	rm := datastore.NewRuntimeManager()
	gameRepo := game.NewMemDBGameRepository(ds, rm)
	log.Debug("Game repository initialized")

	// Per-snapshot VP enrichment. Uses the same ComputePlayerVPBreakdowns helper as
	// FinalScoringAction so the projected score on each history entry matches the
	// scoreboard formula exactly.
	ds.SetSnapshotEnricher(func(state *datastore.GameState) map[string]shared.VPBreakdown {
		g, err := gameRepo.Get(context.Background(), state.ID)
		if err != nil || g == nil {
			return nil
		}
		return gameAction.ComputePlayerVPBreakdowns(g, cardRegistry, awardRegistry, milestoneRegistry)
	})

	// ========== Initialize Game State Repository (Diff Logging) ==========
	stateRepo := game.NewInMemoryGameStateRepository()
	log.Debug("Game state repository initialized")

	// ========== Initialize Bug Report Service ==========
	bugReportService := bugreport.NewService(log)
	caps := bugReportService.Capabilities()
	log.Debug("Feedback service",
		slog.Bool("github_app", caps.GitHubApp))

	// ========== Initialize WebSocket Hub ==========
	hub := core.NewHub()
	log.Debug("WebSocket hub initialized")

	// ========== Initialize Game State Broadcaster (Automatic Broadcasting) ==========
	availableMaps := dto.MapPreviews(mapRegistry)
	broadcaster := wsHandler.NewBroadcaster(gameRepo, stateRepo, hub, cardRegistry, colonyRegistry, pfRegistry, stdProjRegistry, awardRegistry, milestoneRegistry, availableMaps)
	log.Debug("Game state broadcaster initialized (provides automatic broadcasting for all games)")

	// ========== Initialize Game Actions ==========

	// Game lifecycle (6)
	createGameAction := gameAction.NewCreateGameAction(gameRepo, cardRegistry, mapRegistry, log)
	updateGameSettingsAction := gameAction.NewUpdateGameSettingsAction(gameRepo, cardRegistry, mapRegistry, log)
	joinGameAction := gameAction.NewJoinGameAction(gameRepo, cardRegistry, log, colonyRegistry)
	selectDemoChoicesAction := gameAction.NewSelectDemoChoicesAction(gameRepo, cardRegistry, log)
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
	playCardAction := cardAction.NewPlayCardAction(gameRepo, cardRegistry, stateRepo, log, colonyRegistry)
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
	confirmCardDiscardAction := confirmAction.NewConfirmCardDiscardAction(gameRepo, cardRegistry, stateRepo, log)
	confirmBehaviorChoiceAction := confirmAction.NewConfirmBehaviorChoiceAction(gameRepo, cardRegistry, stateRepo, log)
	confirmResourceRemovalAction := confirmAction.NewConfirmResourceRemovalAction(gameRepo, cardRegistry, stateRepo, log)
	confirmColonyResourceAction := confirmAction.NewConfirmColonyResourceAction(gameRepo, cardRegistry, stateRepo, log)
	confirmAwardFundAction := confirmAction.NewConfirmAwardFundAction(gameRepo, cardRegistry, awardRegistry, log)
	confirmColonyPlacementAction := confirmAction.NewConfirmColonyPlacementAction(gameRepo, cardRegistry, colonyRegistry, log)
	confirmCardRevealAction := confirmAction.NewConfirmCardRevealAction(gameRepo)
	confirmEffectSelectionAction := confirmAction.NewConfirmEffectSelectionAction(gameRepo, cardRegistry, colonyRegistry, stateRepo)
	confirmFreeTradeAction := confirmAction.NewConfirmFreeTradeAction(gameRepo, cardRegistry, colonyRegistry, stateRepo)

	// Turn management (4)
	skipActionAction := turnAction.NewSkipActionAction(gameRepo, finalScoringAction, log)
	selectStartingChoicesAction := turnAction.NewSelectStartingChoicesAction(gameRepo, cardRegistry, awardRegistry, log)
	confirmInitAdvanceAction := turnAction.NewConfirmInitAdvanceAction(gameRepo, cardRegistry, awardRegistry, stateRepo, log)

	// Bot service
	personas, err := bot.LoadPersonaCatalog(filepath.Join(wd, "assets", "bot", "personas.json"))
	if err != nil {
		log.Error("Failed to load bot personas", slog.Any("error", err))
		os.Exit(1)
	}
	strategy, err := bot.LoadStrategyGuide(filepath.Join(wd, "assets", "bot", "strategy.md"))
	if err != nil {
		log.Error("Failed to load bot strategy guide", slog.Any("error", err))
		os.Exit(1)
	}
	botTools := bot.NewToolServer(bot.Actions{
		PlayCard:               playCardAction,
		UseCardAction:          useCardActionAction,
		SkipAction:             skipActionAction,
		SelectStartingChoices:  selectStartingChoicesAction,
		ConfirmInitAdvance:     confirmInitAdvanceAction,
		SelectTile:             selectTileAction,
		ConfirmProductionCards: confirmProductionCardsAction,
		ConfirmCardDraw:        confirmCardDrawAction,
		ConfirmCardDiscard:     confirmCardDiscardAction,
		ConfirmBehaviorChoice:  confirmBehaviorChoiceAction,
		ConfirmEffectSelection: confirmEffectSelectionAction,
		ConfirmCardReveal:      confirmCardRevealAction,
		ConfirmSellPatents:     confirmSellPatentsAction,
		ConfirmResourceRemoval: confirmResourceRemovalAction,
		ConfirmColonyPlacement: confirmColonyPlacementAction,
		ConfirmColonyResource:  confirmColonyResourceAction,
		ConfirmAwardFund:       confirmAwardFundAction,
		ConfirmFreeTrade:       confirmFreeTradeAction,
		ExecuteStandardProject: executeStandardProjectAction,
		ConvertHeat:            convertHeatAction,
		ConvertPlants:          convertPlantsAction,
		ClaimMilestone:         claimMilestoneAction,
		FundAward:              fundAwardAction,
		ColonyTrade:            colonyTradeAction,
		ColonyBuild:            colonyBuildAction,
		FundSeat:               fundSeatAction,
	}, bot.Registries{
		Cards:            cardRegistry,
		StandardProjects: stdProjRegistry,
		Milestones:       milestoneRegistry,
		Awards:           awardRegistry,
	}, gameRepo, hub, log)
	botToolsCtx, stopBotTools := context.WithCancel(context.Background())
	defer stopBotTools()
	if err := botTools.Start(botToolsCtx); err != nil {
		log.Error("Failed to start bot tool server", slog.Any("error", err))
		os.Exit(1)
	}
	botController := bot.NewBotController(gameRepo, stateRepo, cardRegistry, broadcaster, bot.NewCLIRunner(log), botTools, personas, bot.DefaultConfig(strategy), log)
	broadcaster.SetBotNotifier(botController)
	addBotAction := gameAction.NewAddBotAction(gameRepo, cardRegistry, botController, log, colonyRegistry)

	startGameAction := turnAction.NewStartGameAction(gameRepo, colonyRegistry, pfRegistry, milestoneRegistry, awardRegistry, botController, log)

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

	// Admin actions (11)
	adminSetPhaseAction := admin.NewSetPhaseAction(gameRepo, log)
	adminSetCurrentTurnAction := admin.NewSetCurrentTurnAction(gameRepo, log)
	adminSetResourcesAction := admin.NewSetResourcesAction(gameRepo, log)
	adminSetProductionAction := admin.NewSetProductionAction(gameRepo, log)
	adminSetGlobalParametersAction := admin.NewSetGlobalParametersAction(gameRepo, log)
	adminGiveCardAction := admin.NewGiveCardAction(gameRepo, cardRegistry, log)
	adminSetCorporationAction := admin.NewSetCorporationAction(gameRepo, cardRegistry, awardRegistry, log)
	adminStartTileSelectionAction := admin.NewStartTileSelectionAction(gameRepo, log)
	adminSetTRAction := admin.NewSetTRAction(gameRepo, log)
	adminSetActionsRemainingAction := admin.NewSetActionsRemainingAction(gameRepo, log)
	adminRestartGameAction := admin.NewRestartGameAction(gameRepo, createGameAction, startGameAction, cardRegistry, colonyRegistry, botController, log)

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
		selectDemoChoicesAction,
		updateGameSettingsAction,
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
		confirmResourceRemovalAction,
		confirmColonyResourceAction,
		confirmAwardFundAction,
		confirmColonyPlacementAction,
		confirmFreeTradeAction,
		confirmEffectSelectionAction,
		confirmCardRevealAction,
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
		botController,
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
		adminRestartGameAction,
		adminSetActionsRemainingAction,
	)

	log.Debug("WebSocket handlers registered")

	// ========== Start WebSocket Hub ==========
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)
	log.Debug("WebSocket hub running")

	// ========== Setup HTTP Router ==========
	mainRouter := mux.NewRouter()

	apiRouter := httpHandler.SetupRouter(
		createGameAction,
		getGameAction,
		getGameLogsAction,
		getGameHistoryAction,
		listGamesAction,
		listCardsAction,
		getPlayerAction,
		cardRegistry,
		milestoneRegistry,
		awardRegistry,
		bugReportService,
		serverMeta,
		changelogResponse,
	)

	// Mount API router
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)

	// Create WebSocket handler
	wsHttpHandler := core.NewHandler(hub)

	// Add WebSocket endpoint
	mainRouter.HandleFunc("/ws", wsHttpHandler.ServeWS)

	// Registered last: mux matches in order, and the frontend owns every other path
	if info, err := os.Stat(webDir); err == nil && info.IsDir() {
		staticHandler, err := web.NewStaticHandler(webDir)
		if err != nil {
			log.Error("Failed to index frontend", slog.Any("error", err))
			os.Exit(1)
		}
		mainRouter.PathPrefix("/").Handler(httpmiddleware.Recovery(staticHandler))
	} else {
		log.Info("Frontend not served: no web directory", slog.String("dir", webDir))
	}

	log.Debug("HTTP routes configured")

	// ========== Setup HTTP Server ==========
	server := &http.Server{
		Addr:         addr,
		Handler:      mainRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Info("HTTP server listening", slog.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start HTTP server", slog.Any("error", err))
			os.Exit(1)
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
		log.Error("Failed to gracefully shutdown HTTP server", slog.Any("error", err))
	} else {
		log.Debug("HTTP server stopped")
	}

	// Cancel WebSocket hub context
	cancel()
	log.Debug("WebSocket hub stopped")

	log.Info("Server shutdown complete")
}
