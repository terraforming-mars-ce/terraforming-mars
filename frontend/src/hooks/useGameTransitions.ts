import { useEffect, useRef } from "react";
import { useGameStore } from "@/stores/gameStore.ts";
import { useUIOverlayStore } from "@/stores/uiOverlayStore.ts";
import { useAppPhaseStore, gameIdOf } from "@/stores/appPhaseStore.ts";
import { useCardPlayFlowStore } from "@/stores/cardPlayFlowStore.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import {
  GamePhaseAction,
  GamePhaseComplete,
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  GamePhaseProductionAndCardDraw,
  GamePhaseStartingSelection,
  GameStatusActive,
  GameStatusCompleted,
  GameStatusLobby,
} from "@/types/generated/api-types.ts";

export function useGameTransitions(
  playProductionSound: () => Promise<void>,
  playGameStartSound: () => Promise<void>,
  notificationQueueDoneAt: React.RefObject<number>,
): void {
  const gamePhase = useGameStore((s) => s.game?.currentPhase);
  const gameStatus = useGameStore((s) => s.game?.status);
  const gameId = useGameStore((s) => s.game?.id);
  const isConnected = useGameStore((s) => s.isConnected);
  const corpData = useGameStore((s) => s.currentPlayer?.corporation);
  const productionPhase = useGameStore((s) => s.currentPlayer?.productionPhase);
  const selectCorpPhase = useGameStore((s) => s.game?.currentPlayer?.selectCorporationPhase);
  const pendingCardSelection = useGameStore((s) => s.game?.currentPlayer?.pendingCardSelection);
  const pendingCardDrawSelection = useGameStore(
    (s) => s.game?.currentPlayer?.pendingCardDrawSelection,
  );
  const pendingCardDiscardSelection = useGameStore(
    (s) => s.game?.currentPlayer?.pendingCardDiscardSelection,
  );
  const pendingBehaviorChoice = useGameStore(
    (s) => s.game?.currentPlayer?.pendingBehaviorChoiceSelection,
  );
  const pendingStealTarget = useGameStore(
    (s) => s.game?.currentPlayer?.pendingStealTargetSelection,
  );
  const pendingColonyResource = useGameStore(
    (s) => s.game?.currentPlayer?.pendingColonyResourceSelection,
  );
  const pendingColonyPlacement = useGameStore((s) => s.game?.currentPlayer?.pendingColonySelection);
  const pendingFreeTrade = useGameStore((s) => s.game?.currentPlayer?.pendingFreeTradeSelection);
  const initPhase = useGameStore((s) => s.game?.initPhase);
  const hostPlayerId = useGameStore((s) => s.game?.hostPlayerId);
  const currentPlayerId = useGameStore((s) => s.currentPlayer?.id);

  const phase = useAppPhaseStore((s) => s.phase);
  const isSkyboxReady = useAppPhaseStore((s) => s.isSkyboxReady);
  const isGpuReady = useAppPhaseStore((s) => s.isGpuReady);

  const isLobbyPhase = gameStatus === GameStatusLobby;

  const isStartingSelectionPhase =
    gameStatus === GameStatusActive && gamePhase === GamePhaseStartingSelection;

  const isInitApplyPhase =
    gameStatus === GameStatusActive &&
    (gamePhase === GamePhaseInitApplyCorp || gamePhase === GamePhaseInitApplyPrelude);

  const wasInLobby = useRef(false);
  const isInitialMount = useRef(true);
  const loadingEnteredAt = useRef<number | null>(null);

  useEffect(() => {
    window.history.pushState(null, "", window.location.href);
    const handlePopState = () => {
      window.history.pushState(null, "", window.location.href);
      const { isConnected, game } = useGameStore.getState();
      if (isConnected && game) {
        useUIOverlayStore.getState().setShowCloseGameConfirm(true);
      }
    };
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  useEffect(() => {
    const { isConnected, game } = useGameStore.getState();
    if (!isConnected || !game) {
      return;
    }
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [isConnected, gamePhase]);

  useEffect(() => {
    const { setCorporationData } = useGameStore.getState();
    if (corpData) {
      setCorporationData(corpData);
    } else {
      setCorporationData(null);
    }
  }, [corpData]);

  useEffect(() => {
    if (!gamePhase) {
      return;
    }
    const { setShowCorp } = useGameStore.getState();
    if (
      gamePhase === GamePhaseInitApplyPrelude ||
      gamePhase === GamePhaseAction ||
      gamePhase === GamePhaseProductionAndCardDraw ||
      gamePhase === GamePhaseComplete
    ) {
      setShowCorp(true);
    }
  }, [gamePhase]);

  useEffect(() => {
    const { setDisplayedInitPlayerId } = useGameStore.getState();
    const initPlayerId = initPhase?.currentPlayerId ?? null;
    if (gamePhase !== GamePhaseInitApplyCorp && gamePhase !== GamePhaseInitApplyPrelude) {
      setDisplayedInitPlayerId(null);
      return;
    }
    const now = Date.now();
    const queueRemaining = Math.max(0, notificationQueueDoneAt.current - now);
    if (queueRemaining === 0) {
      setDisplayedInitPlayerId(initPlayerId);
      return;
    }
    const timer = setTimeout(() => {
      setDisplayedInitPlayerId(initPlayerId);
    }, queueRemaining);
    return () => clearTimeout(timer);
  }, [gamePhase, initPhase?.currentPlayerId, notificationQueueDoneAt]);

  useEffect(() => {
    const {
      showProductionPhaseModal,
      setShowProductionPhaseModal,
      setOpenProductionToCardSelection,
    } = useUIOverlayStore.getState();
    const currentPlayer = useGameStore.getState().currentPlayer;

    const hasProductionData =
      currentPlayer?.productionPhase &&
      !currentPlayer.productionPhase.selectionComplete &&
      currentPlayer.productionPhase.availableCards;

    if (hasProductionData && !showProductionPhaseModal) {
      if (!isInitialMount.current) {
        void playProductionSound();
      }
      setShowProductionPhaseModal(true);
      setOpenProductionToCardSelection(false);
    } else if (!hasProductionData && showProductionPhaseModal) {
      setShowProductionPhaseModal(false);
    }

    if (isInitialMount.current) {
      isInitialMount.current = false;
    }
  }, [productionPhase, playProductionSound]);

  useEffect(() => {
    const { showStartingSelection, setShowStartingSelection, setIsStartingSelectionHidden } =
      useUIOverlayStore.getState();

    const hasStartingData =
      gamePhase === GamePhaseStartingSelection &&
      gameStatus === GameStatusActive &&
      selectCorpPhase &&
      selectCorpPhase.availableCorporations.length > 0;

    if (hasStartingData && !showStartingSelection) {
      setShowStartingSelection(true);
    } else if (showStartingSelection && !selectCorpPhase) {
      setShowStartingSelection(false);
      setIsStartingSelectionHidden(false);
    }
  }, [gamePhase, gameStatus, selectCorpPhase]);

  useEffect(() => {
    const { showPendingCardSelection, setShowPendingCardSelection } = useUIOverlayStore.getState();

    if (pendingCardSelection && !showPendingCardSelection) {
      setShowPendingCardSelection(true);
    } else if (!pendingCardSelection && showPendingCardSelection) {
      setShowPendingCardSelection(false);
    }
  }, [pendingCardSelection]);

  useEffect(() => {
    const { showCardDrawSelection, setShowCardDrawSelection } = useUIOverlayStore.getState();

    if (pendingCardDrawSelection && !showCardDrawSelection) {
      setShowCardDrawSelection(true);
    } else if (!pendingCardDrawSelection && showCardDrawSelection) {
      setShowCardDrawSelection(false);
    }
  }, [pendingCardDrawSelection]);

  useEffect(() => {
    const { showCardDiscardSelection, setShowCardDiscardSelection } = useUIOverlayStore.getState();

    if (pendingCardDiscardSelection && !showCardDiscardSelection) {
      setShowCardDiscardSelection(true);
    } else if (!pendingCardDiscardSelection && showCardDiscardSelection) {
      setShowCardDiscardSelection(false);
    }
  }, [pendingCardDiscardSelection]);

  useEffect(() => {
    const { showBehaviorChoiceSelection, setShowBehaviorChoiceSelection } =
      useCardPlayFlowStore.getState();

    if (pendingBehaviorChoice && !showBehaviorChoiceSelection) {
      setShowBehaviorChoiceSelection(true);
    } else if (!pendingBehaviorChoice && showBehaviorChoiceSelection) {
      setShowBehaviorChoiceSelection(false);
    }
  }, [pendingBehaviorChoice]);

  useEffect(() => {
    const { showStealTargetSelection, setShowStealTargetSelection } = useUIOverlayStore.getState();

    if (pendingStealTarget && !showStealTargetSelection) {
      setShowStealTargetSelection(true);
    } else if (!pendingStealTarget && showStealTargetSelection) {
      setShowStealTargetSelection(false);
    }
  }, [pendingStealTarget]);

  useEffect(() => {
    const { showColonyResourceSelection, setShowColonyResourceSelection } =
      useUIOverlayStore.getState();

    if (pendingColonyResource && !showColonyResourceSelection) {
      setShowColonyResourceSelection(true);
    } else if (!pendingColonyResource && showColonyResourceSelection) {
      setShowColonyResourceSelection(false);
    }
  }, [pendingColonyResource]);

  useEffect(() => {
    const { showColonyPlacementSelection, setShowColonyPlacementSelection } =
      useUIOverlayStore.getState();

    if (pendingColonyPlacement && !showColonyPlacementSelection) {
      setShowColonyPlacementSelection(true);
    } else if (!pendingColonyPlacement && showColonyPlacementSelection) {
      setShowColonyPlacementSelection(false);
    }
  }, [pendingColonyPlacement]);

  useEffect(() => {
    const { showFreeTradeSelection, setShowFreeTradeSelection } = useUIOverlayStore.getState();

    if (pendingFreeTrade && !showFreeTradeSelection) {
      setShowFreeTradeSelection(true);
    } else if (!pendingFreeTrade && showFreeTradeSelection) {
      setShowFreeTradeSelection(false);
    }
  }, [pendingFreeTrade]);

  useEffect(() => {
    if (isLobbyPhase) {
      useAppPhaseStore.getState().setLobbyMounted(true);
    }
  }, [isLobbyPhase]);

  useEffect(() => {
    const { setPhase } = useAppPhaseStore.getState();
    const activeGameId = gameId ?? gameIdOf(phase);
    if (!activeGameId) {
      return;
    }

    if (isLobbyPhase) {
      if (phase.kind !== "lobby") {
        setPhase({ kind: "lobby", gameId: activeGameId });
      }
      wasInLobby.current = true;
      return;
    }

    if (!wasInLobby.current || phase.kind !== "lobby") {
      return;
    }

    setPhase({ kind: "loading", gameId: activeGameId });
  }, [isLobbyPhase, gameId, phase]);

  useEffect(() => {
    if (phase.kind !== "loading") {
      loadingEnteredAt.current = null;
      return;
    }
    if (loadingEnteredAt.current === null) {
      loadingEnteredAt.current = Date.now();
    }
    if (!isSkyboxReady || !isGpuReady) {
      return;
    }
    const activeGameId = phase.gameId;
    const elapsed = Date.now() - loadingEnteredAt.current;
    const remaining = Math.max(0, 1000 - elapsed);
    const advance = () => {
      useAppPhaseStore.getState().setPhase({ kind: "fadeOutLobby", gameId: activeGameId });
    };
    if (remaining === 0) {
      advance();
      return;
    }
    const timer = window.setTimeout(advance, remaining);
    return () => window.clearTimeout(timer);
  }, [phase, isSkyboxReady, isGpuReady]);

  useEffect(() => {
    const { setPhase } = useAppPhaseStore.getState();
    if (phase.kind === "fadeOutLobby") {
      const timer = setTimeout(
        () => setPhase({ kind: "marsRevealed", gameId: phase.gameId }),
        1500,
      );
      return () => clearTimeout(timer);
    }
    if (phase.kind === "animateUI") {
      const timer = setTimeout(() => setPhase({ kind: "playing", gameId: phase.gameId }), 2500);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [phase]);

  useEffect(() => {
    if (phase.kind === "marsRevealed" && !isStartingSelectionPhase) {
      useAppPhaseStore.getState().setPhase({ kind: "animateUI", gameId: phase.gameId });
    }
  }, [phase, isStartingSelectionPhase]);

  useEffect(() => {
    const { setMarsRevealedReady } = useAppPhaseStore.getState();

    if (phase.kind === "marsRevealed") {
      const timer = setTimeout(() => setMarsRevealedReady(true), 2000);
      return () => clearTimeout(timer);
    }
    setMarsRevealedReady(false);
    return undefined;
  }, [phase]);

  useEffect(() => {
    if (phase.kind === "fadeOutLobby") {
      void playGameStartSound();
    }
  }, [phase, playGameStartSound]);

  useEffect(() => {
    const activeGameId = gameId ?? gameIdOf(phase);
    if (!activeGameId) {
      return;
    }
    if (
      isStartingSelectionPhase &&
      (phase.kind === "checking" ||
        phase.kind === "connecting" ||
        phase.kind === "selecting" ||
        phase.kind === "joining" ||
        phase.kind === "spectating") &&
      !wasInLobby.current &&
      isSkyboxReady &&
      isGpuReady
    ) {
      useAppPhaseStore.getState().setPhase({ kind: "marsRevealed", gameId: activeGameId });
    }
  }, [isStartingSelectionPhase, phase, isSkyboxReady, isGpuReady, gameId]);

  useEffect(() => {
    if (gameStatus !== GameStatusActive) {
      return;
    }
    if (isLobbyPhase || isStartingSelectionPhase) {
      return;
    }
    if (!isSkyboxReady || !isGpuReady) {
      return;
    }
    if (wasInLobby.current) {
      return;
    }
    const activeGameId = gameId ?? gameIdOf(phase);
    if (!activeGameId) {
      return;
    }
    if (
      phase.kind === "checking" ||
      phase.kind === "connecting" ||
      phase.kind === "selecting" ||
      phase.kind === "joining" ||
      phase.kind === "spectating"
    ) {
      useAppPhaseStore.getState().setPhase({ kind: "playing", gameId: activeGameId });
    }
  }, [
    gameStatus,
    isLobbyPhase,
    isStartingSelectionPhase,
    isSkyboxReady,
    isGpuReady,
    gameId,
    phase,
  ]);

  useEffect(() => {
    if (gameStatus !== GameStatusCompleted) {
      return;
    }
    const activeGameId = gameId ?? gameIdOf(phase);
    if (!activeGameId) {
      return;
    }
    if (phase.kind !== "completed") {
      useAppPhaseStore.getState().setPhase({ kind: "completed", gameId: activeGameId });
    }
  }, [gameStatus, gameId, phase]);

  useEffect(() => {
    if (!isInitApplyPhase || !initPhase?.waitingForConfirm) {
      return;
    }
    if (hostPlayerId !== currentPlayerId) {
      return;
    }

    const animationDone = phase.kind === "playing" || phase.kind === "completed";
    if (!animationDone) {
      return;
    }

    if (initPhase.hasPendingTiles) {
      return;
    }

    const now = Date.now();
    const queueRemaining = Math.max(0, notificationQueueDoneAt.current - now);
    const delay = queueRemaining + 750;

    const timer = setTimeout(() => {
      void globalWebSocketManager.confirmInitAdvance();
    }, delay);

    return () => clearTimeout(timer);
  }, [
    isInitApplyPhase,
    initPhase?.waitingForConfirm,
    initPhase?.confirmVersion,
    initPhase?.currentPlayerIndex,
    initPhase?.hasPendingTiles,
    hostPlayerId,
    currentPlayerId,
    phase,
    notificationQueueDoneAt,
  ]);
}
