// Package app is the composition root: it builds every repository, registry, action and
// transport of the server and returns one HTTP handler. The server binary and the
// functional tests both use it, so tests run exactly the production wiring.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

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
	httpmiddleware "openmars/internal/middleware/http"
	"openmars/internal/service/bot"
	"openmars/internal/service/bugreport"

	"github.com/gorilla/mux"
)

// Config holds everything the server needs from its environment.
type Config struct {
	// AssetsDir holds cards.json, maps.json and the other game data files.
	AssetsDir string
	// WebDir holds the built frontend. It is not served when the directory does not exist.
	WebDir string
	// ChangelogDir holds the release notes. An empty changelog is served when it does not exist.
	ChangelogDir string
	Meta         dto.MetaResponse
	// BotRunner runs bot model calls. Nil uses the claude CLI.
	BotRunner bot.Runner
	Logger    *slog.Logger
}

// App is a fully wired server.
type App struct {
	hub     *core.Hub
	ds      *datastore.DataStore
	handler http.Handler
	cancel  context.CancelFunc
	hubDone chan struct{}
}

// New builds the server and starts its background work: the WebSocket hub and the bot
// tool server. Close stops both.
func New(cfg Config) (*App, error) {
	log := cfg.Logger
	asset := func(parts ...string) string {
		return filepath.Join(append([]string{cfg.AssetsDir}, parts...)...)
	}

	changelogResponse := dto.ToChangelogResponse(nil)
	if _, err := os.Stat(cfg.ChangelogDir); err != nil {
		log.Warn("Changelog directory not found, serving an empty changelog", slog.String("path", cfg.ChangelogDir))
	} else {
		entries, err := changelog.LoadAll(changelog.Player, cfg.ChangelogDir)
		if err != nil {
			return nil, fmt.Errorf("load changelog: %w", err)
		}
		changelogResponse = dto.ToChangelogResponse(entries)
		log.Debug("Changelog loaded", slog.Int("entry_count", len(entries)))
	}

	cardData, err := cards.LoadCardsFromJSON(asset("cards.json"))
	if err != nil {
		return nil, fmt.Errorf("load cards: %w", err)
	}
	cardRegistry := cards.NewInMemoryCardRegistry(cardData)

	colonyData, err := colony.LoadColoniesFromJSON(asset("colonies.json"))
	if err != nil {
		return nil, fmt.Errorf("load colonies: %w", err)
	}
	colonyRegistry := colony.NewInMemoryColonyRegistry(colonyData)

	pfData, err := pfLoader.LoadProjectsFromJSON(asset("project_funding.json"))
	if err != nil {
		return nil, fmt.Errorf("load project funding: %w", err)
	}
	pfRegistry := pfLoader.NewInMemoryProjectFundingRegistry(pfData)

	stdProjData, err := stdprojLoader.LoadStandardProjectsFromJSON(asset("standard_projects.json"))
	if err != nil {
		return nil, fmt.Errorf("load standard projects: %w", err)
	}
	stdProjRegistry := stdprojLoader.NewInMemoryStandardProjectRegistry(stdProjData)

	awardData, err := award.LoadAwardsFromJSON(asset("awards.json"))
	if err != nil {
		return nil, fmt.Errorf("load awards: %w", err)
	}
	awardRegistry := award.NewInMemoryAwardRegistry(awardData)

	milestoneData, err := msLoader.LoadMilestonesFromJSON(asset("milestones.json"))
	if err != nil {
		return nil, fmt.Errorf("load milestones: %w", err)
	}
	milestoneRegistry := msLoader.NewInMemoryMilestoneRegistry(milestoneData)

	mapRegistry, err := board.LoadMapsFromJSON(asset("maps.json"))
	if err != nil {
		return nil, fmt.Errorf("load maps: %w", err)
	}

	personas, err := bot.LoadPersonaCatalog(asset("bot", "personas.json"))
	if err != nil {
		return nil, fmt.Errorf("load bot personas: %w", err)
	}
	strategy, err := bot.LoadStrategyGuide(asset("bot", "strategy.md"))
	if err != nil {
		return nil, fmt.Errorf("load bot strategy guide: %w", err)
	}
	log.Debug("Game data loaded",
		slog.Int("card_count", len(cardData)),
		slog.Int("map_count", len(mapRegistry.ListMaps())))

	ds, err := datastore.NewDataStore()
	if err != nil {
		return nil, fmt.Errorf("create datastore: %w", err)
	}
	gameRepo := game.NewMemDBGameRepository(ds, datastore.NewRuntimeManager())
	hub := core.NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	// Per-snapshot VP enrichment uses the same helper as final scoring, so the projected
	// score on each history entry matches the scoreboard exactly. It runs on its own
	// goroutine, so it reads the game on the hub goroutine instead of racing the action
	// that is still changing it.
	ds.SetSnapshotEnricher(func(state *datastore.GameState) map[string]shared.VPBreakdown {
		var breakdowns map[string]shared.VPBreakdown
		_ = hub.Do(ctx, func() {
			g, err := gameRepo.Get(ctx, state.ID)
			if err != nil || g == nil {
				return
			}
			breakdowns = gameAction.ComputePlayerVPBreakdowns(g, cardRegistry, awardRegistry, milestoneRegistry)
		})
		return breakdowns
	})

	stateRepo := game.NewInMemoryGameStateRepository()
	bugReportService := bugreport.NewService(log)

	broadcaster := wsHandler.NewBroadcaster(gameRepo, stateRepo, hub, cardRegistry, colonyRegistry, pfRegistry, stdProjRegistry, awardRegistry, milestoneRegistry, dto.MapPreviews(mapRegistry))

	createGameAction := gameAction.NewCreateGameAction(gameRepo, cardRegistry, mapRegistry, log)
	updateGameSettingsAction := gameAction.NewUpdateGameSettingsAction(gameRepo, cardRegistry, mapRegistry, log)
	joinGameAction := gameAction.NewJoinGameAction(gameRepo, cardRegistry, log, colonyRegistry)
	selectDemoChoicesAction := gameAction.NewSelectDemoChoicesAction(gameRepo, cardRegistry, log)
	finalScoringAction := gameAction.NewFinalScoringAction(gameRepo, cardRegistry, awardRegistry, milestoneRegistry, log)

	claimMilestoneAction := milestoneAction.NewClaimMilestoneAction(gameRepo, cardRegistry, stateRepo, milestoneRegistry, log)
	fundAwardAction := awardAction.NewFundAwardAction(gameRepo, cardRegistry, stateRepo, awardRegistry, log)

	colonyTradeAction := colonyAction.NewTradeAction(gameRepo, colonyRegistry, cardRegistry, stateRepo, log)
	colonyBuildAction := colonyAction.NewBuildColonyAction(gameRepo, colonyRegistry, cardRegistry, stateRepo, log)

	fundSeatAction := pfAction.NewFundSeatAction(gameRepo, pfRegistry, stateRepo)

	playCardAction := cardAction.NewPlayCardAction(gameRepo, cardRegistry, stateRepo, log, colonyRegistry)
	useCardActionAction := cardAction.NewUseCardActionAction(gameRepo, cardRegistry, stateRepo, log)

	executeStandardProjectAction := stdprojAction.NewExecuteStandardProjectAction(gameRepo, cardRegistry, stdProjRegistry, stateRepo, log)

	convertHeatAction := resconvAction.NewConvertHeatToTemperatureAction(gameRepo, cardRegistry, stateRepo, log)
	convertPlantsAction := resconvAction.NewConvertPlantsToGreeneryAction(gameRepo, cardRegistry, stateRepo, log)

	selectTileAction := tileAction.NewSelectTileAction(gameRepo, cardRegistry, stateRepo, log)

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

	skipActionAction := turnAction.NewSkipActionAction(gameRepo, finalScoringAction, log)
	selectStartingChoicesAction := turnAction.NewSelectStartingChoicesAction(gameRepo, cardRegistry, awardRegistry, log)
	confirmInitAdvanceAction := turnAction.NewConfirmInitAdvanceAction(gameRepo, cardRegistry, awardRegistry, stateRepo, log)

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

	if err := botTools.Start(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("start bot tool server: %w", err)
	}

	botRunner := cfg.BotRunner
	if botRunner == nil {
		botRunner = bot.NewCLIRunner(log)
	}
	botController := bot.NewBotController(gameRepo, stateRepo, cardRegistry, broadcaster, botRunner, botTools, personas, bot.DefaultConfig(strategy), log)
	broadcaster.SetBotNotifier(botController)
	addBotAction := gameAction.NewAddBotAction(gameRepo, cardRegistry, botController, log, colonyRegistry)
	startGameAction := turnAction.NewStartGameAction(gameRepo, colonyRegistry, pfRegistry, milestoneRegistry, awardRegistry, botController, log)
	convertToBotAction := gameAction.NewConvertToBotAction(gameRepo, botController, log)

	playerDisconnectedAction := connAction.NewPlayerDisconnectedAction(gameRepo, log)
	playerTakeoverAction := connAction.NewPlayerTakeoverAction(gameRepo, cardRegistry, log)
	kickPlayerAction := connAction.NewKickPlayerAction(gameRepo, botController, finalScoringAction, log)
	endGameAction := connAction.NewEndGameAction(gameRepo, botController, log)

	setPlayerColorAction := connAction.NewSetPlayerColorAction(gameRepo, log)
	spectateGameAction := connAction.NewSpectateGameAction(gameRepo, log)
	spectatorDisconnectedAction := connAction.NewSpectatorDisconnectedAction(gameRepo, log)
	kickSpectatorAction := connAction.NewKickSpectatorAction(gameRepo, log)
	sendChatMessageAction := connAction.NewSendChatMessageAction(gameRepo, log)

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

	wsHandler.RegisterHandlers(
		hub,
		broadcaster,
		gameRepo,
		joinGameAction,
		addBotAction,
		selectDemoChoicesAction,
		updateGameSettingsAction,
		playCardAction,
		useCardActionAction,
		executeStandardProjectAction,
		convertHeatAction,
		convertPlantsAction,
		selectTileAction,
		startGameAction,
		skipActionAction,
		selectStartingChoicesAction,
		confirmInitAdvanceAction,
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
		playerDisconnectedAction,
		playerTakeoverAction,
		kickPlayerAction,
		endGameAction,
		setPlayerColorAction,
		spectateGameAction,
		spectatorDisconnectedAction,
		kickSpectatorAction,
		sendChatMessageAction,
		convertToBotAction,
		botController,
		claimMilestoneAction,
		fundAwardAction,
		colonyTradeAction,
		colonyBuildAction,
		fundSeatAction,
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

	router := mux.NewRouter()
	router.PathPrefix("/api/v1").Handler(httpHandler.SetupRouter(
		createGameAction,
		query.NewGetGameAction(gameRepo, log),
		query.NewGetGameLogsAction(stateRepo, log),
		query.NewGetGameHistoryAction(ds, log),
		query.NewListGamesAction(gameRepo, log),
		query.NewListCardsAction(cardRegistry, log),
		query.NewGetPlayerAction(gameRepo, log),
		cardRegistry,
		milestoneRegistry,
		awardRegistry,
		bugReportService,
		cfg.Meta,
		cfg.ChangelogDir,
		changelogResponse,
	))
	router.HandleFunc("/ws", core.NewHandler(hub).ServeWS)

	// Registered last: mux matches in order, and the frontend owns every other path.
	if info, err := os.Stat(cfg.WebDir); err == nil && info.IsDir() {
		staticHandler, err := web.NewStaticHandler(cfg.WebDir)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("index frontend: %w", err)
		}
		router.PathPrefix("/").Handler(httpmiddleware.Recovery(staticHandler))
	} else {
		log.Info("Frontend not served: no web directory", slog.String("dir", cfg.WebDir))
	}

	hubDone := make(chan struct{})
	go func() {
		defer close(hubDone)
		hub.Run(ctx)
	}()

	return &App{hub: hub, ds: ds, handler: router, cancel: cancel, hubDone: hubDone}, nil
}

// Handler serves the API, the WebSocket endpoint and the frontend.
func (a *App) Handler() http.Handler {
	return a.handler
}

// MessageTypes lists every message type clients can send over the WebSocket.
func (a *App) MessageTypes() []dto.MessageType {
	return a.hub.MessageTypes()
}

// ConnectionCount returns the number of open WebSocket connections.
func (a *App) ConnectionCount() int {
	return a.hub.Manager().ConnectionCount()
}

// Close stops the hub, closing every WebSocket connection, the bot tool server and any
// history enrichment still running.
func (a *App) Close() {
	a.cancel()
	<-a.hubDone
	a.ds.WaitForPendingEnrichments()
}
