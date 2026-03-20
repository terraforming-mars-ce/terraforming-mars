import { useCallback, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import GameLayout from "./GameLayout.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import ProductionPhaseModal from "../../ui/modals/ProductionPhaseModal.tsx";
import PaymentSelectionPopover from "../../ui/popover/PaymentSelectionPopover.tsx";
import DebugDropdown from "../../ui/debug/DebugDropdown.tsx";
import DevModeChip from "../../ui/debug/DevModeChip.tsx";
import PerformanceWindow from "../../ui/debug/PerformanceWindow.tsx";
import FeedbackWindow from "../../ui/debug/FeedbackWindow.tsx";

import { WindowManagerProvider } from "../../ui/debug/WindowManager.tsx";
import WaitingRoomOverlay from "../../ui/overlay/WaitingRoomOverlay.tsx";
import PlayerSelectionOverlay from "../../ui/overlay/PlayerSelectionOverlay.tsx";
import JoinGameOverlay from "../../ui/overlay/JoinGameOverlay.tsx";
import SpectateGameOverlay from "../../ui/overlay/SpectateGameOverlay.tsx";
import DemoSetupOverlay from "../../ui/overlay/DemoSetupOverlay.tsx";
import TabConflictOverlay from "../../ui/overlay/TabConflictOverlay.tsx";
import StartingCardSelectionOverlay from "../../ui/overlay/StartingCardSelectionOverlay.tsx";
import PendingCardSelectionOverlay from "../../ui/overlay/PendingCardSelectionOverlay.tsx";
import CardDrawSelectionOverlay from "../../ui/overlay/CardDrawSelectionOverlay.tsx";
import CardDiscardSelectionOverlay from "../../ui/overlay/CardDiscardSelectionOverlay.tsx";
import AwardFundSelectionPopover from "../../ui/popover/AwardFundSelectionPopover.tsx";
import CardFanOverlay, { CardFanOverlayHandle } from "../../ui/overlay/CardFanOverlay.tsx";
import CorporationOverlay from "../../ui/overlay/CorporationOverlay.tsx";
import LoadingOverlay from "../../game/view/LoadingOverlay.tsx";
import GameEventBanner from "../../ui/overlay/GameEventBanner.tsx";
import { useGameEvent } from "@/hooks/useGameEvent.ts";
import { usePlayedCardNotification } from "@/hooks/usePlayedCardNotification.ts";
import ChatOverlay from "../../ui/overlay/ChatOverlay.tsx";
import GameButton from "../../ui/buttons/GameButton.tsx";
import { BotDifficultyChip, BotSpeedChip } from "../../ui/display/BotChips.tsx";
import GameMenuModal from "../../ui/overlay/GameMenuModal.tsx";
import MainMenuHamburger from "../../ui/buttons/MainMenuHamburger.tsx";
import SpaceBackground from "../../3d/SpaceBackground.tsx";
import EndGameBottomBar from "../../ui/endgame/EndGameBottomBar.tsx";
import { VPCountingProvider } from "../../../contexts/VPCountingContext.tsx";
import { useVPCountingAnimation } from "@/hooks/useVPCountingAnimation.ts";

import { useGameHistory } from "@/hooks/useGameHistory.ts";
import { useGameReplay } from "@/hooks/useGameReplay.ts";
import {
  historyEntryToGameDto,
  buildCardLookup,
  historyPlayerToPlayerDto,
} from "@/utils/historyToGameDto.ts";
import ChoiceSelectionPopover from "../../ui/popover/ChoiceSelectionPopover.tsx";
import CardStorageSelectionPopover from "../../ui/popover/CardStorageSelectionPopover.tsx";
import TargetPlayerSelectionPopover from "../../ui/popover/TargetPlayerSelectionPopover.tsx";
import CardResourceSelectionPopover from "../../ui/popover/CardResourceSelectionPopover.tsx";
import AmountSelectionPopover from "../../ui/popover/AmountSelectionPopover.tsx";
import ActionReusePopover from "../../ui/popover/ActionReusePopover.tsx";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { clearGameSession } from "@/utils/sessionStorage.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";
import { useSpaceBackground } from "@/contexts/SpaceBackgroundContext.tsx";
import { useNotifications } from "@/contexts/NotificationContext.tsx";
import {
  GamePhaseComplete,
  GamePhaseDemoSetup,
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  GamePhaseStartingSelection,
  GameStatusActive,
  GameStatusCompleted,
  GameStatusLobby,
  ResourceType,
} from "@/types/generated/api-types.ts";
import type { PlayerDto, OtherPlayerDto } from "@/types/generated/api-types.ts";
import { PlayerListHandle } from "../../ui/list/PlayerList.tsx";

import { useGameStore } from "@/stores/gameStore.ts";
import { useUIOverlayStore } from "@/stores/uiOverlayStore.ts";
import { useCardPlayFlowStore } from "@/stores/cardPlayFlowStore.ts";
import { useTransitionStore } from "@/stores/transitionStore.ts";
import { useSpectateStore } from "@/stores/spectateStore.ts";
import { useCardPlayFlow } from "@/hooks/useCardPlayFlow.ts";
import { useWebSocketConnection } from "@/hooks/useWebSocketConnection.ts";
import { useGameInitialization } from "@/hooks/useGameInitialization.ts";
import { useGameTransitions } from "@/hooks/useGameTransitions.ts";
import { useGameHotkeys } from "@/hooks/useGameHotkeys.ts";

export default function GameInterface() {
  const location = useLocation();
  const navigate = useNavigate();
  const { gameId: urlGameId } = useParams<{ gameId?: string }>();
  const { playProductionSound, playGameStartSound } = useSoundEffects();
  const { showNotification } = useNotifications();
  const { isLoaded: isSpaceBgLoaded } = useSpaceBackground();

  // Refs that stay in the component
  const cardFanRef = useRef<CardFanOverlayHandle>(null);
  const playerListRef = useRef<PlayerListHandle>(null);
  const notificationQueueDoneAt = useRef<number>(0);
  const vpCountingStartedRef = useRef(false);
  const winnerBannerShownRef = useRef(false);

  // Game event and played card notification hooks
  const {
    currentEvent,
    enqueue: enqueueGameEvent,
    dismissCurrent: dismissGameEvent,
  } = useGameEvent();
  const playedCardNotification = usePlayedCardNotification();
  const { enqueue: enqueuePlayedCard } = playedCardNotification;

  // --- Stores ---
  const game = useGameStore((s) => s.game);
  const currentPlayer = useGameStore((s) => s.currentPlayer);
  const playerId = useGameStore((s) => s.playerId);
  const isConnected = useGameStore((s) => s.isConnected);
  const isSpectator = useGameStore((s) => s.isSpectator);
  const isReconnecting = useGameStore((s) => s.isReconnecting);
  const reconnectionStep = useGameStore((s) => s.reconnectionStep);
  const loadingPhase = useGameStore((s) => s.loadingPhase);
  const gameForSelection = useGameStore((s) => s.gameForSelection);
  const changedPaths = useGameStore((s) => s.changedPaths);
  const triggeredEffects = useGameStore((s) => s.triggeredEffects);
  const chatMessages = useGameStore((s) => s.chatMessages);
  const corporationData = useGameStore((s) => s.corporationData);
  const showCorp = useGameStore((s) => s.showCorp);
  const displayedInitPlayerId = useGameStore((s) => s.displayedInitPlayerId);

  const showCardsPlayedModal = useUIOverlayStore((s) => s.showCardsPlayedModal);
  const showProductionPhaseModal = useUIOverlayStore((s) => s.showProductionPhaseModal);
  const isProductionModalHidden = useUIOverlayStore((s) => s.isProductionModalHidden);
  const openProductionToCardSelection = useUIOverlayStore((s) => s.openProductionToCardSelection);
  const showDebugDropdown = useUIOverlayStore((s) => s.showDebugDropdown);
  const showPerformanceWindow = useUIOverlayStore((s) => s.showPerformanceWindow);
  const showFeedbackWindow = useUIOverlayStore((s) => s.showFeedbackWindow);
  const showTabConflict = useUIOverlayStore((s) => s.showTabConflict);
  const conflictingTabInfo = useUIOverlayStore((s) => s.conflictingTabInfo);
  const showLeaveGameConfirm = useUIOverlayStore((s) => s.showLeaveGameConfirm);
  const showCloseGameConfirm = useUIOverlayStore((s) => s.showCloseGameConfirm);
  const showEndGameConfirm = useUIOverlayStore((s) => s.showEndGameConfirm);
  const showStartingSelection = useUIOverlayStore((s) => s.showStartingSelection);
  const isStartingSelectionHidden = useUIOverlayStore((s) => s.isStartingSelectionHidden);
  const showPendingCardSelection = useUIOverlayStore((s) => s.showPendingCardSelection);
  const showCardDrawSelection = useUIOverlayStore((s) => s.showCardDrawSelection);
  const showCardDiscardSelection = useUIOverlayStore((s) => s.showCardDiscardSelection);
  const showCorporationOverlay = useUIOverlayStore((s) => s.showCorporationOverlay);
  const showStealTargetSelection = useUIOverlayStore((s) => s.showStealTargetSelection);
  const showColonyResourceSelection = useUIOverlayStore((s) => s.showColonyResourceSelection);

  const showBehaviorChoiceSelection = useCardPlayFlowStore((s) => s.showBehaviorChoiceSelection);
  const cardPendingChoice = useCardPlayFlowStore((s) => s.cardPendingChoice);
  const pendingCardBehaviorIndex = useCardPlayFlowStore((s) => s.pendingCardBehaviorIndex);
  const showChoiceSelection = useCardPlayFlowStore((s) => s.showChoiceSelection);
  const actionPendingChoice = useCardPlayFlowStore((s) => s.actionPendingChoice);
  const showActionChoiceSelection = useCardPlayFlowStore((s) => s.showActionChoiceSelection);
  const pendingActionReuse = useCardPlayFlowStore((s) => s.pendingActionReuse);
  const showActionReuseSelection = useCardPlayFlowStore((s) => s.showActionReuseSelection);
  const pendingBehaviorChoiceStorage = useCardPlayFlowStore((s) => s.pendingBehaviorChoiceStorage);
  const showBehaviorChoiceStorage = useCardPlayFlowStore((s) => s.showBehaviorChoiceStorage);
  const pendingCardStorage = useCardPlayFlowStore((s) => s.pendingCardStorage);
  const showCardStorageSelection = useCardPlayFlowStore((s) => s.showCardStorageSelection);
  const pendingCardPayment = useCardPlayFlowStore((s) => s.pendingCardPayment);
  const showPaymentSelection = useCardPlayFlowStore((s) => s.showPaymentSelection);
  const pendingActionStorage = useCardPlayFlowStore((s) => s.pendingActionStorage);
  const showActionStorageSelection = useCardPlayFlowStore((s) => s.showActionStorageSelection);
  const pendingTargetPlayer = useCardPlayFlowStore((s) => s.pendingTargetPlayer);
  const showTargetPlayerSelection = useCardPlayFlowStore((s) => s.showTargetPlayerSelection);
  const pendingActionTargetPlayer = useCardPlayFlowStore((s) => s.pendingActionTargetPlayer);
  const showActionTargetPlayerSelection = useCardPlayFlowStore(
    (s) => s.showActionTargetPlayerSelection,
  );
  const pendingCardResourceInput = useCardPlayFlowStore((s) => s.pendingCardResourceInput);
  const showCardResourceSelection = useCardPlayFlowStore((s) => s.showCardResourceSelection);
  const pendingVariableAmount = useCardPlayFlowStore((s) => s.pendingVariableAmount);
  const showAmountSelection = useCardPlayFlowStore((s) => s.showAmountSelection);

  const transitionPhase = useTransitionStore((s) => s.transitionPhase);
  const isSkyboxReady = useTransitionStore((s) => s.isSkyboxReady);
  const isGpuReady = useTransitionStore((s) => s.isGpuReady);
  const marsRevealedReady = useTransitionStore((s) => s.marsRevealedReady);
  const overlayVisible = useTransitionStore((s) => s.overlayVisible);
  const lobbyMounted = useTransitionStore((s) => s.lobbyMounted);

  const spectatePlayerId = useSpectateStore((s) => s.spectatePlayerId);
  const replaySpectatePlayerId = useSpectateStore((s) => s.replaySpectatePlayerId);

  // --- Hooks ---
  const flow = useCardPlayFlow();
  const { attemptReconnection: _attemptReconnection } = useWebSocketConnection(
    navigate,
    enqueueGameEvent,
    enqueuePlayedCard,
    notificationQueueDoneAt,
    showNotification,
  );
  const init = useGameInitialization({ navigate, location, urlGameId });
  useGameTransitions(playProductionSound, playGameStartSound, notificationQueueDoneAt);
  useGameHotkeys(cardFanRef, playerListRef);

  // --- VP counting animation ---
  const isGameComplete =
    game?.currentPhase === GamePhaseComplete && game?.status === GameStatusCompleted;
  const vpCounting = useVPCountingAnimation(isGameComplete ? (game?.finalScores ?? []) : []);

  if (isGameComplete && !vpCountingStartedRef.current && (game?.finalScores?.length ?? 0) > 0) {
    vpCountingStartedRef.current = true;
    vpCounting.start();
  }

  if (vpCounting.state.isComplete && !winnerBannerShownRef.current && game?.finalScores?.length) {
    winnerBannerShownRef.current = true;
    const sorted = [...game.finalScores].sort(
      (a, b) => b.vpBreakdown.totalVP - a.vpBreakdown.totalVP,
    );
    const w = sorted[0];
    const allPlayers = [game.currentPlayer, ...(game.otherPlayers ?? [])];
    const wp = allPlayers.find((p) => p?.id === w.playerId);
    enqueueGameEvent({
      title: `${w.playerName} WINS!`,
      duration: 12000,
      playerName: w.playerName,
      playerColor: wp?.color ?? "#f59e0b",
      size: "large",
    });
  }

  // --- Game history and replay ---
  const { entries: historyEntries } = useGameHistory(isGameComplete ? (game?.id ?? null) : null);
  const replay = useGameReplay(historyEntries);
  const hasHistory = (historyEntries?.length ?? 0) > 0;

  // Endgame panel tracking
  const [endgamePanel, setEndgamePanel] = useState<"score" | "graphs" | "replay">("score");
  const endgameFadeUI = isGameComplete && endgamePanel !== "replay";

  const handleEndgamePanelChange = useCallback(
    (panel: "score" | "graphs" | "replay") => {
      setEndgamePanel(panel);
      const isCounting = vpCounting.state.isActive && !vpCounting.state.isComplete;
      if (panel !== "score" && isCounting) {
        vpCounting.skip();
      }
      if (panel !== "score") {
        dismissGameEvent();
      }
      if (panel === "replay" && !replay.isActive) {
        replay.start();
      }
      if (panel !== "replay" && replay.isActive) {
        replay.exit();
        useSpectateStore.getState().setReplaySpectatePlayerId(null);
      }
    },
    [vpCounting, replay, dismissGameEvent],
  );

  const cardLookup = useMemo(() => {
    if (!game) {
      return new Map();
    }
    return buildCardLookup(game);
  }, [game]);

  const replayGameState = useMemo(() => {
    if (!replay.isActive || !replay.currentEntry || !game) {
      return null;
    }
    return historyEntryToGameDto(replay.currentEntry, game, cardLookup);
  }, [replay.isActive, replay.currentEntry, game, cardLookup]);

  const replayViewAsPlayer = useMemo(() => {
    if (!replay.isActive || !replay.currentEntry || !replaySpectatePlayerId) {
      return null;
    }
    const historyPlayer = replay.currentEntry.players[replaySpectatePlayerId];
    if (!historyPlayer) {
      return null;
    }
    return historyPlayerToPlayerDto(historyPlayer, cardLookup);
  }, [replay.isActive, replay.currentEntry, replaySpectatePlayerId, cardLookup]);

  // --- Spectate player derived values ---
  const spectatePlayer = useMemo(() => {
    if (!spectatePlayerId || !game) {
      return null;
    }
    if (game.currentPlayer?.id === spectatePlayerId) {
      return game.currentPlayer;
    }
    return game.otherPlayers?.find((p) => p.id === spectatePlayerId) ?? null;
  }, [spectatePlayerId, game]);

  const spectatePlayerColor = useMemo(() => {
    if (!spectatePlayerId || !game) {
      return "#6496ff";
    }
    if (game.currentPlayer?.id === spectatePlayerId) {
      return game.currentPlayer.color || "#6496ff";
    }
    const other = game.otherPlayers?.find((p) => p.id === spectatePlayerId);
    return other?.color || "#6496ff";
  }, [spectatePlayerId, game]);

  const playerColorMap = useMemo(() => {
    if (!game) {
      return new Map<string, string>();
    }
    const map = new Map<string, string>();
    if (game.currentPlayer?.id && game.currentPlayer.color) {
      map.set(game.currentPlayer.id, game.currentPlayer.color);
    }
    game.otherPlayers?.forEach((p) => {
      if (p.id && p.color) {
        map.set(p.id, p.color);
      }
    });
    game.spectators?.forEach((s) => {
      if (s.id && s.color) {
        map.set(s.id, s.color);
      }
    });
    return map;
  }, [game]);

  const handlePlayerClick = useCallback(
    (player: PlayerDto | OtherPlayerDto) => {
      const phase = game?.currentPhase;
      if (phase === GamePhaseInitApplyCorp || phase === GamePhaseInitApplyPrelude) {
        return;
      }
      useSpectateStore.getState().toggleSpectatePlayer(player.id, game?.currentPlayer?.id);
    },
    [game?.currentPlayer?.id, game?.currentPhase],
  );

  const handleStopSpectating = useCallback(() => {
    useSpectateStore.getState().clearSpectating();
  }, []);

  // --- Computed phase booleans ---
  const isLobbyPhase = game?.status === GameStatusLobby;
  const isPreGamePhase = isLobbyPhase;

  const showWaitingForPlayers =
    game?.status === GameStatusActive &&
    game?.currentPhase === GamePhaseStartingSelection &&
    !game?.currentPlayer?.selectCorporationPhase &&
    !game?.currentPlayer?.selectPreludeCardsPhase &&
    !game?.currentPlayer?.selectStartingCardsPhase &&
    !game?.currentPlayer?.pendingTileSelection;

  // --- Simple WebSocket handlers that stay here ---
  const handleStartingChoicesConfirm = useCallback(
    async (corporationId: string, preludeIds: string[], cardIds: string[]) => {
      try {
        await globalWebSocketManager.selectStartingChoices(corporationId, preludeIds, cardIds);
      } catch (error) {
        console.error("Failed to select starting choices:", error);
      }
    },
    [],
  );

  const handlePendingCardSelection = useCallback(async (selectedCardIds: string[]) => {
    try {
      await globalWebSocketManager.selectCards(selectedCardIds);
    } catch (error) {
      console.error("Failed to select cards:", error);
    }
  }, []);

  const handleCardDrawConfirm = useCallback(async (cardsToTake: string[], cardsToBuy: string[]) => {
    try {
      await globalWebSocketManager.confirmCardDraw(cardsToTake, cardsToBuy);
    } catch (error) {
      console.error("Failed to confirm card draw:", error);
    }
  }, []);

  const handleCardDiscardConfirm = useCallback(async (cardsToDiscard: string[]) => {
    try {
      await globalWebSocketManager.confirmCardDiscard(cardsToDiscard);
    } catch (error) {
      console.error("Failed to confirm card discard:", error);
    }
  }, []);

  // --- Leave/end game handlers ---
  const handleLeaveGame = useCallback(() => {
    useUIOverlayStore.getState().setShowLeaveGameConfirm(true);
  }, []);

  const handleConfirmLeaveGame = useCallback(() => {
    useUIOverlayStore.getState().setShowLeaveGameConfirm(false);
    useUIOverlayStore.getState().setShowCloseGameConfirm(false);
    clearGameSession();
    globalWebSocketManager.disconnect();
    setTimeout(() => {
      navigate("/", { replace: true });
    }, 100);
  }, [navigate]);

  const handleEndGame = useCallback(() => {
    useUIOverlayStore.getState().setShowEndGameConfirm(true);
  }, []);

  const handleConfirmEndGame = useCallback(() => {
    useUIOverlayStore.getState().setShowEndGameConfirm(false);
    void globalWebSocketManager.endGame();
  }, []);

  // --- Bottom bar callbacks ---
  const bottomBarCallbacks = useMemo(
    () => ({
      onOpenCardsPlayedModal: () => useUIOverlayStore.getState().setShowCardsPlayedModal(true),
      onActionSelect: flow.handleActionSelect,
      onConvertPlantsToGreenery: flow.handleConvertPlantsToGreenery,
      onConvertHeatToTemperature: flow.handleConvertHeatToTemperature,
    }),
    [
      flow.handleActionSelect,
      flow.handleConvertPlantsToGreenery,
      flow.handleConvertHeatToTemperature,
    ],
  );

  // --- Loading message ---
  const loadingMessage = (() => {
    if (
      loadingPhase === "selecting" ||
      loadingPhase === "joining" ||
      loadingPhase === "spectating"
    ) {
      if (!isSpaceBgLoaded) {
        return "Loading 3D environment...";
      }
      return "Loading game...";
    }
    if (loadingPhase === "checking") {
      return "Loading game...";
    }
    if (loadingPhase === "connecting") {
      return "Connecting...";
    }
    if (isReconnecting && reconnectionStep) {
      if (reconnectionStep === "game") {
        return "Reconnecting to game...";
      }
      if (reconnectionStep === "environment") {
        return "Loading 3D environment...";
      }
    }
    if (!isSkyboxReady) {
      return "Loading 3D environment...";
    }
    return "Connecting to game...";
  })();

  // --- Loading state ---
  const isFullyLoaded =
    (loadingPhase === "selecting" && isSpaceBgLoaded) ||
    (loadingPhase === "joining" && isSpaceBgLoaded) ||
    (loadingPhase === "spectating" && isSpaceBgLoaded) ||
    (isConnected &&
      !!game &&
      !isReconnecting &&
      (isSkyboxReady || transitionPhase === "lobby") &&
      (isGpuReady || transitionPhase === "lobby"));

  const handleSkyboxReady = useCallback(() => {
    useTransitionStore.getState().setSkyboxReady(true);
  }, []);

  const handleGpuReady = useCallback(() => {
    useTransitionStore.getState().setGpuReady(true);
  }, []);

  const handleLoadingTransitionEnd = useCallback(() => {
    const ts = useTransitionStore.getState();
    if (ts.transitionPhase === "loading") {
      ts.setTransitionPhase("fadeOutLobby");
    } else {
      ts.setOverlayVisible(false);
    }
  }, []);

  // --- Backdrop ---
  const shouldShowBackdrop =
    (showStartingSelection &&
      !isStartingSelectionHidden &&
      (marsRevealedReady || transitionPhase === "idle")) ||
    showWaitingForPlayers;

  // --- Card fan visibility ---
  const hideCardFanForModals =
    showStartingSelection ||
    showPendingCardSelection ||
    showCardDrawSelection ||
    showCardDiscardSelection ||
    showBehaviorChoiceSelection ||
    isPreGamePhase;

  const cardFanTransitionClass = (() => {
    if (spectatePlayerId) {
      return "opacity-0 pointer-events-none";
    }
    if (transitionPhase === "animateUI") {
      return "animate-[uiFadeIn_1200ms_ease-out_both]";
    }
    if (
      transitionPhase === "loading" ||
      transitionPhase === "fadeOutLobby" ||
      transitionPhase === "marsRevealed"
    ) {
      return "opacity-0";
    }
    return "";
  })();

  return (
    <VPCountingProvider
      state={vpCounting.state}
      controls={{
        start: vpCounting.start,
        skip: vpCounting.skip,
        goToPhase: vpCounting.goToPhase,
        phases: vpCounting.phases,
      }}
    >
      {game?.settings?.developmentMode && <DevModeChip />}

      {shouldShowBackdrop && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-[999] animate-[backdropFadeIn_0.3s_ease-out]" />
      )}

      <style>{`
        @keyframes backdropFadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes fadeOut {
          from { opacity: 1; }
          to { opacity: 0; }
        }
        @keyframes modalFadeIn {
          from { opacity: 0; transform: scale(0.95); }
          to { opacity: 1; transform: scale(1); }
        }
        @keyframes uiFadeIn {
          from { opacity: 0; transform: translateY(20px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>

      {game &&
        loadingPhase !== "selecting" &&
        loadingPhase !== "joining" &&
        loadingPhase !== "spectating" && (
          <GameLayout
            ref={playerListRef}
            gameState={replayGameState ?? game}
            currentPlayer={replayViewAsPlayer ?? (replay.isActive ? null : currentPlayer)}
            playedCards={replayViewAsPlayer?.playedCards ?? currentPlayer?.playedCards ?? []}
            corporationCard={replayViewAsPlayer?.corporation ?? corporationData}
            showCorporation={!!replayViewAsPlayer || showCorp}
            initTurnPlayerId={displayedInitPlayerId}
            showStartingSelection={showStartingSelection}
            transitionPhase={transitionPhase}
            animateHexEntrance={
              transitionPhase === "marsRevealed" ||
              transitionPhase === "animateUI" ||
              transitionPhase === "complete"
            }
            startDark={transitionPhase === "loading" || transitionPhase === "lobby"}
            tilesHidden={
              transitionPhase === "loading" ||
              transitionPhase === "lobby" ||
              transitionPhase === "fadeOutLobby"
            }
            changedPaths={changedPaths}
            triggeredEffects={triggeredEffects}
            bottomBarCallbacks={bottomBarCallbacks}
            onStandardProjectSelect={flow.handleStandardProjectSelect}
            onLeaveGame={handleLeaveGame}
            onEndGame={handleEndGame}
            onSkyboxReady={handleSkyboxReady}
            onGpuReady={handleGpuReady}
            onPlayerClick={handlePlayerClick}
            spectatingPlayer={spectatePlayer}
            spectatingCorporation={spectatePlayer?.corporation ?? null}
            spectatePlayerColor={spectatePlayerColor}
            onStopSpectating={handleStopSpectating}
            isGameSpectator={isSpectator}
            chatMessages={chatMessages}
            onSendChatMessage={(msg) => void globalWebSocketManager.sendChatMessage(msg)}
            isLobbyPhase={isLobbyPhase}
            playerColorMap={playerColorMap}
            endgameFadeUI={endgameFadeUI}
            isEndgame={isGameComplete}
            activeEndgamePanel={endgamePanel}
            onEndgamePanelChange={handleEndgamePanelChange}
            hasHistory={hasHistory}
            playedCardNotification={playedCardNotification.currentNotification}
            isPlayedCardPinned={playedCardNotification.isPinned}
            onPlayedCardTogglePin={playedCardNotification.togglePin}
            onPlayedCardAdvance={playedCardNotification.advance}
          />
        )}

      <CardsPlayedModal
        isVisible={showCardsPlayedModal}
        onClose={() => useUIOverlayStore.getState().setShowCardsPlayedModal(false)}
        cards={(spectatePlayer?.playedCards ?? currentPlayer?.playedCards) || []}
      />

      <ProductionPhaseModal
        isOpen={showProductionPhaseModal && !isProductionModalHidden}
        gameState={game}
        onClose={() => {
          useUIOverlayStore.getState().setShowProductionPhaseModal(false);
          useUIOverlayStore.getState().setIsProductionModalHidden(false);
          useUIOverlayStore.getState().setOpenProductionToCardSelection(false);
        }}
        onHide={() => {
          useUIOverlayStore.getState().setIsProductionModalHidden(true);
          useUIOverlayStore.getState().setOpenProductionToCardSelection(false);
        }}
        openDirectlyToCardSelection={openProductionToCardSelection}
      />

      <WindowManagerProvider>
        <DebugDropdown
          isVisible={showDebugDropdown}
          onClose={() => useUIOverlayStore.getState().setShowDebugDropdown(false)}
          gameState={game}
          changedPaths={changedPaths}
        />

        <PerformanceWindow
          isVisible={showPerformanceWindow}
          onClose={() => useUIOverlayStore.getState().setShowPerformanceWindow(false)}
        />

        <FeedbackWindow
          isVisible={showFeedbackWindow}
          onClose={() => useUIOverlayStore.getState().setShowFeedbackWindow(false)}
          gameState={game}
        />
      </WindowManagerProvider>

      {(transitionPhase === "lobby" ||
        transitionPhase === "loading" ||
        transitionPhase === "fadeOutLobby" ||
        loadingPhase === "selecting" ||
        loadingPhase === "joining" ||
        loadingPhase === "spectating") && (
        <div
          className={
            transitionPhase === "fadeOutLobby" ? "animate-[fadeOut_1500ms_ease-out_forwards]" : ""
          }
        >
          <SpaceBackground animationSpeed={0.5} overlayOpacity={0.3} />
        </div>
      )}

      {lobbyMounted && game && (playerId || isSpectator) && (
        <>
          <WaitingRoomOverlay
            game={game}
            playerId={playerId ?? "spectator"}
            visible={isLobbyPhase}
            onExited={() => useTransitionStore.getState().setLobbyMounted(false)}
          />
          {isLobbyPhase && (
            <ChatOverlay
              messages={chatMessages}
              onSendMessage={(msg) => void globalWebSocketManager.sendChatMessage(msg)}
              isLobby={true}
              playerColorMap={playerColorMap}
            />
          )}
        </>
      )}

      {game?.currentPhase === GamePhaseDemoSetup && game && playerId && (
        <DemoSetupOverlay game={game} playerId={playerId} />
      )}

      {showTabConflict && conflictingTabInfo && (
        <TabConflictOverlay
          activeGameInfo={conflictingTabInfo}
          onTakeOver={init.handleTabTakeOver}
          onCancel={init.handleTabCancel}
        />
      )}

      {loadingPhase === "selecting" && gameForSelection && (
        <PlayerSelectionOverlay
          game={gameForSelection}
          onSelectPlayer={(pid, playerName) => void init.handlePlayerSelected(pid, playerName)}
          onSpectate={init.handleSpectatorConnected}
          onCancel={init.handlePlayerSelectionCancel}
        />
      )}

      {loadingPhase === "joining" && gameForSelection && (
        <JoinGameOverlay game={gameForSelection} onCancel={init.handlePlayerSelectionCancel} />
      )}

      {loadingPhase === "spectating" && gameForSelection && (
        <SpectateGameOverlay
          game={gameForSelection}
          onCancel={init.handlePlayerSelectionCancel}
          onConnected={init.handleSpectatorConnected}
        />
      )}

      {showLeaveGameConfirm && (
        <GameMenuModal
          title="Leave game?"
          showBackdrop={true}
          onClose={() => useUIOverlayStore.getState().setShowLeaveGameConfirm(false)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            You can reconnect to the game again without losing any progress.
          </p>
          <div className="flex gap-4 justify-center">
            <GameButton
              buttonType="secondary"
              onClick={() => useUIOverlayStore.getState().setShowLeaveGameConfirm(false)}
            >
              Cancel
            </GameButton>
            <GameButton variant="error" onClick={handleConfirmLeaveGame}>
              Leave
            </GameButton>
          </div>
        </GameMenuModal>
      )}

      {showCloseGameConfirm && (
        <GameMenuModal
          title="Close game?"
          showBackdrop={true}
          onClose={() => useUIOverlayStore.getState().setShowCloseGameConfirm(false)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            You can reconnect to the game again without losing any progress.
          </p>
          <div className="flex gap-4 justify-center">
            <GameButton
              buttonType="secondary"
              onClick={() => useUIOverlayStore.getState().setShowCloseGameConfirm(false)}
            >
              Cancel
            </GameButton>
            <GameButton variant="error" onClick={handleConfirmLeaveGame}>
              Close
            </GameButton>
          </div>
        </GameMenuModal>
      )}

      {showEndGameConfirm && (
        <GameMenuModal
          title="End game?"
          showBackdrop={true}
          onClose={() => useUIOverlayStore.getState().setShowEndGameConfirm(false)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            This will end the game for all players. This action cannot be undone.
          </p>
          <div className="flex gap-4 justify-center">
            <GameButton
              buttonType="secondary"
              onClick={() => useUIOverlayStore.getState().setShowEndGameConfirm(false)}
            >
              Cancel
            </GameButton>
            <GameButton variant="error" onClick={handleConfirmEndGame}>
              End game
            </GameButton>
          </div>
        </GameMenuModal>
      )}

      {showStartingSelection && !isStartingSelectionHidden && game && (
        <MainMenuHamburger
          gameId={game.id}
          onLeaveGame={handleLeaveGame}
          onEndGame={playerId === game.hostPlayerId ? handleEndGame : undefined}
        />
      )}
      <StartingCardSelectionOverlay
        isOpen={
          showStartingSelection &&
          !isStartingSelectionHidden &&
          (marsRevealedReady || transitionPhase === "idle")
        }
        availableCorporations={
          game?.currentPlayer?.selectCorporationPhase?.availableCorporations || []
        }
        availablePreludes={game?.currentPlayer?.selectPreludeCardsPhase?.availablePreludes || []}
        maxSelectablePreludes={game?.currentPlayer?.selectPreludeCardsPhase?.maxSelectable || 2}
        cards={game?.currentPlayer?.selectStartingCardsPhase?.availableCards || []}
        playerCredits={currentPlayer?.resources?.credits || 40}
        onConfirm={handleStartingChoicesConfirm}
        onHide={() => useUIOverlayStore.getState().setIsStartingSelectionHidden(true)}
      />

      {showStartingSelection && isStartingSelectionHidden && marsRevealedReady && (
        <GameButton
          className="fixed top-[80px] left-[70%] !py-3.5 !px-7 !text-base !border-space-blue-400 text-shadow-glow shadow-[0_4px_15px_rgba(0,0,0,0.5),0_0_20px_rgba(30,60,150,0.4)] z-[1000] whitespace-nowrap hover:!border-space-blue-500 hover:shadow-[0_6px_20px_rgba(0,0,0,0.6),0_0_35px_rgba(30,60,150,0.6)] active:shadow-[0_2px_10px_rgba(0,0,0,0.4),0_0_20px_rgba(30,60,150,0.4)]"
          onClick={() => useUIOverlayStore.getState().setIsStartingSelectionHidden(false)}
        >
          Return to Selection
        </GameButton>
      )}

      {showWaitingForPlayers && game && (
        <>
          <MainMenuHamburger
            gameId={game.id}
            onLeaveGame={handleLeaveGame}
            onEndGame={playerId === game.hostPlayerId ? handleEndGame : undefined}
          />
          <div className="fixed inset-0 z-[1000] flex items-center justify-center">
            <div className="w-[450px] max-w-[90vw] bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)] animate-[modalFadeIn_0.3s_ease-out]">
              <div className="text-center mb-6">
                <h2 className="font-orbitron text-white text-[24px] m-0 mb-2 text-shadow-glow font-bold tracking-wider">
                  Waiting for players...
                </h2>
              </div>

              <div className="mb-6">
                <h3 className="text-white text-sm font-semibold mb-2 uppercase tracking-wide">
                  Players
                </h3>
                <div className="flex flex-col gap-2">
                  {(() => {
                    const allPlayers: {
                      id: string;
                      name: string;
                      isReady: boolean;
                      isSelf: boolean;
                      playerType: string;
                      botDifficulty?: string;
                      botSpeed?: string;
                    }[] = [];

                    if (game.currentPlayer) {
                      allPlayers.push({
                        id: game.currentPlayer.id,
                        name: game.currentPlayer.name,
                        isReady:
                          !game.currentPlayer.selectCorporationPhase &&
                          !game.currentPlayer.selectPreludeCardsPhase &&
                          !game.currentPlayer.selectStartingCardsPhase &&
                          !!game.currentPlayer.corporation,
                        isSelf: true,
                        playerType: game.currentPlayer.playerType,
                        botDifficulty: game.currentPlayer.botDifficulty || undefined,
                        botSpeed: game.currentPlayer.botSpeed || undefined,
                      });
                    }

                    game.otherPlayers?.forEach((other) => {
                      allPlayers.push({
                        id: other.id,
                        name: other.name,
                        isReady:
                          !other.selectStartingCardsPhase &&
                          !other.selectCorporationPhase &&
                          !other.selectPreludeCardsPhase &&
                          !!other.corporation,
                        isSelf: false,
                        playerType: other.playerType,
                        botDifficulty: other.botDifficulty || undefined,
                        botSpeed: other.botSpeed || undefined,
                      });
                    });

                    const ordered = game.turnOrder?.length
                      ? game.turnOrder
                          .map((pid) => allPlayers.find((p) => p.id === pid))
                          .filter((p) => p !== undefined)
                      : allPlayers;

                    return ordered.map((player) => (
                      <div
                        key={player.id}
                        className="flex justify-between items-center py-2 px-3 bg-black/40 rounded-lg border border-space-blue-600/50"
                      >
                        <span className="text-white text-sm font-medium">{player.name}</span>
                        <div className="flex gap-1.5 items-center">
                          {player.isSelf && (
                            <span className="bg-space-blue-800 text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                              You
                            </span>
                          )}
                          {player.playerType === "bot" && (
                            <>
                              <BotDifficultyChip difficulty={player.botDifficulty} />
                              <BotSpeedChip speed={player.botSpeed} />
                            </>
                          )}
                          {player.isReady ? (
                            <span className="flex items-center gap-1 bg-emerald-700/80 text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                              <svg
                                width="10"
                                height="10"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="3"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                              >
                                <polyline points="20 6 9 17 4 12" />
                              </svg>
                              Ready
                            </span>
                          ) : (
                            <span className="flex items-center gap-1 bg-white/10 text-white/70 py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                              <div className="w-2.5 h-2.5 border border-white/50 border-t-transparent rounded-full animate-spin" />
                              Selecting...
                            </span>
                          )}
                        </div>
                      </div>
                    ));
                  })()}
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {game?.currentPlayer?.pendingCardSelection && (
        <PendingCardSelectionOverlay
          isOpen={showPendingCardSelection}
          selection={game.currentPlayer.pendingCardSelection}
          playerCredits={currentPlayer?.resources?.credits || 0}
          onSelectCards={handlePendingCardSelection}
        />
      )}

      {game?.currentPlayer?.pendingCardDrawSelection && (
        <CardDrawSelectionOverlay
          isOpen={showCardDrawSelection}
          selection={game.currentPlayer.pendingCardDrawSelection}
          playerCredits={currentPlayer?.resources?.credits || 0}
          onConfirm={handleCardDrawConfirm}
        />
      )}

      {game?.currentPlayer?.pendingCardDiscardSelection && currentPlayer && (
        <CardDiscardSelectionOverlay
          isOpen={showCardDiscardSelection}
          selection={game.currentPlayer.pendingCardDiscardSelection}
          handCards={currentPlayer.cards || []}
          onConfirm={handleCardDiscardConfirm}
        />
      )}

      {game &&
        (currentPlayer || replayViewAsPlayer) &&
        (game.currentPhase !== GamePhaseComplete || !!replayViewAsPlayer) && (
          <div className={`transition-all duration-300 ease-in-out ${cardFanTransitionClass}`}>
            <CardFanOverlay
              ref={cardFanRef}
              cards={replayViewAsPlayer?.cards ?? currentPlayer?.cards ?? []}
              hideWhenModalOpen={hideCardFanForModals}
              onCardSelect={(_cardId) => {}}
              onPlayCard={flow.handlePlayCard}
            />
          </div>
        )}

      <CorporationOverlay
        visible={showCorporationOverlay}
        onClose={() => useUIOverlayStore.getState().setShowCorporationOverlay(false)}
        currentPlayer={replayViewAsPlayer ?? currentPlayer}
        otherPlayers={game?.otherPlayers ?? []}
      />

      {game &&
        playerId &&
        game.currentPhase === GamePhaseComplete &&
        game.status === GameStatusCompleted &&
        game.finalScores &&
        game.finalScores.length > 0 && (
          <>
            <EndGameBottomBar
              game={game}
              playerId={playerId}
              historyEntries={historyEntries}
              activePanel={endgamePanel}
              isReplayActive={replay.isActive}
              replayIndex={replay.currentIndex}
              replayTotal={replay.totalStates}
              replayPlaying={replay.isPlaying}
              replaySpeed={replay.playbackSpeed}
              onReplayPlay={replay.play}
              onReplayPause={replay.pause}
              onReplaySeek={replay.seekTo}
              onReplayStepForward={replay.stepForward}
              onReplayStepBackward={replay.stepBackward}
              onReplaySpeedChange={replay.setPlaybackSpeed}
              replaySpectatePlayerId={replaySpectatePlayerId}
              onReplaySpectatePlayerChange={(id) =>
                useSpectateStore.getState().setReplaySpectatePlayerId(id)
              }
            />
          </>
        )}

      {cardPendingChoice && (
        <ChoiceSelectionPopover
          cardId={cardPendingChoice.id}
          cardName={cardPendingChoice.name}
          behaviors={cardPendingChoice.behaviors || []}
          behaviorIndex={pendingCardBehaviorIndex}
          onChoiceSelect={flow.handleChoiceSelect}
          onCancel={flow.handleChoiceCancel}
          isVisible={showChoiceSelection}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {actionPendingChoice && (
        <ChoiceSelectionPopover
          cardId={actionPendingChoice.cardId}
          cardName={actionPendingChoice.cardName}
          behaviors={[actionPendingChoice.behavior]}
          behaviorIndex={0}
          onChoiceSelect={flow.handleActionChoiceSelect}
          onCancel={flow.handleActionChoiceCancel}
          isVisible={showActionChoiceSelection}
          isAction={true}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {pendingActionReuse && currentPlayer?.actions && (
        <ActionReusePopover
          isVisible={showActionReuseSelection}
          onClose={flow.handleActionReuseCancel}
          actions={currentPlayer.actions}
          reuseSourceCardId={pendingActionReuse.cardId}
          onActionSelect={flow.handleActionReuseSelect}
          gameState={game ?? undefined}
        />
      )}

      {game?.currentPlayer?.pendingBehaviorChoiceSelection && (
        <ChoiceSelectionPopover
          cardId={game.currentPlayer.pendingBehaviorChoiceSelection.sourceCardId}
          cardName={`Triggered: ${game.currentPlayer.pendingBehaviorChoiceSelection.source}`}
          behaviors={[
            {
              choices: game.currentPlayer.pendingBehaviorChoiceSelection.choices,
            },
          ]}
          behaviorIndex={0}
          onChoiceSelect={flow.handleBehaviorChoiceSelect}
          onCancel={() => {}}
          isVisible={showBehaviorChoiceSelection}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {pendingBehaviorChoiceStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingBehaviorChoiceStorage.resourceType}
          amount={pendingBehaviorChoiceStorage.amount}
          selectorTags={pendingBehaviorChoiceStorage.selectorTags}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={currentPlayer?.corporation}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={flow.handleBehaviorChoiceStorageSelect}
          onCancel={flow.handleBehaviorChoiceStorageCancel}
          isVisible={showBehaviorChoiceStorage}
        />
      )}

      {pendingCardStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingCardStorage.resourceType}
          amount={pendingCardStorage.amount}
          selectorTags={pendingCardStorage.selectorTags}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={currentPlayer?.corporation}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={flow.handleCardStorageSelect}
          onCancel={flow.handleCardStorageCancel}
          isVisible={showCardStorageSelection}
        />
      )}

      {pendingCardPayment && game && currentPlayer && (
        <PaymentSelectionPopover
          cardId={pendingCardPayment.card.id}
          card={pendingCardPayment.card}
          playerResources={currentPlayer.resources}
          paymentConstants={game.paymentConstants}
          playerPaymentSubstitutes={currentPlayer.paymentSubstitutes}
          storagePaymentSubstitutes={currentPlayer.storagePaymentSubstitutes}
          resourceStorage={currentPlayer.resourceStorage}
          onConfirm={flow.handlePaymentConfirm}
          onCancel={flow.handlePaymentCancel}
          isVisible={showPaymentSelection}
        />
      )}

      {pendingActionStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingActionStorage.resourceType}
          amount={pendingActionStorage.amount}
          selectorTags={pendingActionStorage.selectorTags}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={currentPlayer?.corporation}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={flow.handleActionStorageSelect}
          onCancel={flow.handleActionStorageCancel}
          isVisible={showActionStorageSelection}
        />
      )}

      {pendingTargetPlayer && game && (
        <TargetPlayerSelectionPopover
          resourceType={pendingTargetPlayer.resourceType}
          amount={pendingTargetPlayer.amount}
          isSteal={pendingTargetPlayer.isSteal}
          players={[
            ...(currentPlayer
              ? [
                  {
                    id: currentPlayer.id,
                    name: currentPlayer.name,
                    resources: currentPlayer.resources,
                    production: currentPlayer.production,
                  },
                ]
              : []),
            ...(game.otherPlayers || []).map((p) => ({
              id: p.id,
              name: p.name,
              resources: p.resources,
              production: p.production,
            })),
          ]}
          onPlayerSelect={flow.handleTargetPlayerSelect}
          onCancel={flow.handleTargetPlayerCancel}
          isVisible={showTargetPlayerSelection}
        />
      )}

      {pendingActionTargetPlayer && game && (
        <TargetPlayerSelectionPopover
          resourceType={pendingActionTargetPlayer.resourceType}
          amount={pendingActionTargetPlayer.amount}
          isSteal={pendingActionTargetPlayer.isSteal}
          players={[
            ...(currentPlayer
              ? [
                  {
                    id: currentPlayer.id,
                    name: currentPlayer.name,
                    resources: currentPlayer.resources,
                    production: currentPlayer.production,
                  },
                ]
              : []),
            ...(game.otherPlayers || []).map((p) => ({
              id: p.id,
              name: p.name,
              resources: p.resources,
              production: p.production,
            })),
          ]}
          onPlayerSelect={flow.handleActionTargetPlayerSelect}
          onCancel={flow.handleActionTargetPlayerCancel}
          isVisible={showActionTargetPlayerSelection}
        />
      )}

      {game?.currentPlayer?.pendingStealTargetSelection && game && (
        <TargetPlayerSelectionPopover
          resourceType={game.currentPlayer.pendingStealTargetSelection.resourceType as ResourceType}
          amount={game.currentPlayer.pendingStealTargetSelection.amount}
          isSteal={true}
          players={(game.otherPlayers || [])
            .filter((p) =>
              game.currentPlayer!.pendingStealTargetSelection!.eligiblePlayerIds.includes(p.id),
            )
            .map((p) => ({
              id: p.id,
              name: p.name,
              resources: p.resources,
              production: p.production,
            }))}
          onPlayerSelect={flow.handleStealTargetSelect}
          onCancel={flow.handleStealTargetSkip}
          isVisible={showStealTargetSelection}
          mandatory
        />
      )}

      {game?.currentPlayer?.pendingColonyResourceSelection && currentPlayer && (
        <CardStorageSelectionPopover
          resourceType={
            game.currentPlayer.pendingColonyResourceSelection.resourceType as ResourceType
          }
          amount={game.currentPlayer.pendingColonyResourceSelection.amount}
          reason={game.currentPlayer.pendingColonyResourceSelection.reason}
          mandatory
          playedCards={currentPlayer.playedCards || []}
          corporationCard={currentPlayer.corporation}
          resourceStorage={currentPlayer.resourceStorage}
          onCardSelect={flow.handleColonyResourceSelect}
          onCancel={flow.handleColonyResourceSkip}
          isVisible={showColonyResourceSelection}
        />
      )}

      {game?.currentPlayer?.pendingAwardFundSelection && (
        <AwardFundSelectionPopover
          isOpen={true}
          selection={game.currentPlayer.pendingAwardFundSelection}
          gameState={game}
        />
      )}

      {pendingCardResourceInput && (
        <CardResourceSelectionPopover
          resourceType={pendingCardResourceInput.resourceType}
          amount={pendingCardResourceInput.amount}
          excludeCardId={pendingCardResourceInput.cardId}
          players={[
            ...(currentPlayer
              ? [
                  {
                    id: currentPlayer.id,
                    name: currentPlayer.name,
                    playedCards: currentPlayer.playedCards,
                    resourceStorage: currentPlayer.resourceStorage,
                  },
                ]
              : []),
            ...(game?.otherPlayers || []).map((p) => ({
              id: p.id,
              name: p.name,
              playedCards: p.playedCards,
              resourceStorage: p.resourceStorage,
            })),
          ]}
          onCardSelect={flow.handleCardResourceSelect}
          onCancel={flow.handleCardResourceCancel}
          isVisible={showCardResourceSelection}
        />
      )}

      {pendingVariableAmount && (
        <AmountSelectionPopover
          cardName={pendingVariableAmount.cardName}
          resourceLabel={pendingVariableAmount.resourceLabel}
          maxAmount={pendingVariableAmount.maxAmount}
          onAmountSelect={flow.handleAmountSelect}
          onCancel={flow.handleAmountCancel}
          isVisible={showAmountSelection}
        />
      )}

      {showProductionPhaseModal && isProductionModalHidden && (
        <GameButton
          className="fixed top-[80px] left-[70%] !py-3.5 !px-7 !text-base !border-space-blue-400 text-shadow-glow shadow-[0_4px_15px_rgba(0,0,0,0.5),0_0_20px_rgba(30,60,150,0.4)] z-[1000] whitespace-nowrap hover:!border-space-blue-500 hover:shadow-[0_6px_20px_rgba(0,0,0,0.6),0_0_35px_rgba(30,60,150,0.6)] active:shadow-[0_2px_10px_rgba(0,0,0,0.4),0_0_20px_rgba(30,60,150,0.4)]"
          onClick={() => {
            useUIOverlayStore.getState().setIsProductionModalHidden(false);
            useUIOverlayStore.getState().setOpenProductionToCardSelection(true);
          }}
        >
          Return to Production
        </GameButton>
      )}

      <GameEventBanner
        event={endgamePanel === "score" ? currentEvent : null}
        onDismiss={dismissGameEvent}
      />

      {overlayVisible && (
        <LoadingOverlay
          isLoaded={isFullyLoaded}
          message={loadingMessage}
          onTransitionEnd={handleLoadingTransitionEnd}
        />
      )}
    </VPCountingProvider>
  );
}
