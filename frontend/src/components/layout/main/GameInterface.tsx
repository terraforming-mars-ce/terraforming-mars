import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { config } from "../../../config.ts";
import GameLayout from "./GameLayout.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import ProductionPhaseModal from "../../ui/modals/ProductionPhaseModal.tsx";
import PaymentSelectionPopover from "../../ui/popover/PaymentSelectionPopover.tsx";
import DebugDropdown from "../../ui/debug/DebugDropdown.tsx";
import DevModeChip from "../../ui/debug/DevModeChip.tsx";
import PerformanceWindow from "../../ui/debug/PerformanceWindow.tsx";
import BugReportWindow from "../../ui/debug/BugReportWindow.tsx";

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
import CardFanOverlay, { CardFanOverlayHandle } from "../../ui/overlay/CardFanOverlay.tsx";
import LoadingOverlay from "../../game/view/LoadingOverlay.tsx";
import GameEventBanner from "../../ui/overlay/GameEventBanner.tsx";
import { useGameEvent } from "@/hooks/useGameEvent.ts";
import ChatOverlay from "../../ui/overlay/ChatOverlay.tsx";
import MainMenuSettingsButton from "../../ui/buttons/MainMenuSettingsButton.tsx";
import GameMenuButton from "../../ui/buttons/GameMenuButton.tsx";
import { BotDifficultyChip, BotSpeedChip } from "../../ui/display/BotChips.tsx";
import GameMenuModal from "../../ui/overlay/GameMenuModal.tsx";
import SpaceBackground from "../../3d/SpaceBackground.tsx";
import EndGameOverlay, { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import { TileHighlightMode } from "../../game/board/Tile.tsx";
import ChoiceSelectionPopover from "../../ui/popover/ChoiceSelectionPopover.tsx";
import CardStorageSelectionPopover from "../../ui/popover/CardStorageSelectionPopover.tsx";
import TargetPlayerSelectionPopover from "../../ui/popover/TargetPlayerSelectionPopover.tsx";
import CardResourceSelectionPopover from "../../ui/popover/CardResourceSelectionPopover.tsx";
import AmountSelectionPopover from "../../ui/popover/AmountSelectionPopover.tsx";
import ActionReusePopover from "../../ui/popover/ActionReusePopover.tsx";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { apiService } from "@/services/apiService.ts";
import { getTabManager } from "@/utils/tabManager.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";
import { useSpaceBackground } from "@/contexts/SpaceBackgroundContext.tsx";
import { useNotifications } from "@/contexts/NotificationContext.tsx";
import { skyboxCache } from "@/services/SkyboxCache.ts";
import { audioService } from "@/services/audioService.ts";
import { clearGameSession, getGameSession, saveGameSession } from "@/utils/sessionStorage.ts";
import {
  CardDto,
  CardPaymentDto,
  ChatMessageDto,
  FullStatePayload,
  GameDto,
  GamePhaseAction,
  GamePhaseComplete,
  GamePhaseDemoSetup,
  GamePhaseProductionAndCardDraw,
  GamePhaseInitApplyCorp,
  GamePhaseInitApplyPrelude,
  GamePhaseStartingSelection,
  GameStatusActive,
  GameStatusCompleted,
  GameStatusLobby,
  PlayerDisconnectedPayload,
  PlayerDto,
  OtherPlayerDto,
  PlayerActionDto,
  PlayerCardDto,
  ResourceType,
  StateDiffDto,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import { shouldShowPaymentModal, createDefaultPayment } from "@/utils/paymentUtils.ts";
import { deepClone, findChangedPaths } from "@/utils/deepCompare.ts";
import { StandardProject } from "@/types/cards.tsx";

type TransitionPhase = "idle" | "lobby" | "loading" | "fadeOutLobby" | "animateUI" | "complete";

type LoadingPhase = "checking" | "selecting" | "joining" | "spectating" | "connecting" | "ready";

export default function GameInterface() {
  const location = useLocation();
  const navigate = useNavigate();
  const { gameId: urlGameId } = useParams<{ gameId?: string }>();
  const { playProductionSound, playTemperatureSound, playOxygenSound, playYourTurnSound } =
    useSoundEffects();
  const { showNotification } = useNotifications();
  const [game, setGame] = useState<GameDto | null>(null);
  const {
    currentEvent,
    enqueue: enqueueGameEvent,
    dismissCurrent: dismissGameEvent,
  } = useGameEvent();
  const [isConnected, setIsConnected] = useState(false);
  const [transitionPhase, setTransitionPhase] = useState<TransitionPhase>("idle");
  const wasInLobby = useRef(false);
  const cardFanRef = useRef<CardFanOverlayHandle>(null);
  const [isReconnecting, setIsReconnecting] = useState(false);
  const [reconnectionStep, setReconnectionStep] = useState<"game" | "environment" | null>(null);
  const [currentPlayer, setCurrentPlayer] = useState<PlayerDto | null>(null);
  const [playerId, setPlayerId] = useState<string | null>(null); // Track player ID separately
  const [loadingPhase, setLoadingPhase] = useState<LoadingPhase>("checking");
  const [gameForSelection, setGameForSelection] = useState<GameDto | null>(null);
  const [_showCorporationModal, setShowCorporationModal] = useState(false);
  const [corporationData, setCorporationData] = useState<CardDto | null>(null);

  // New modal states
  const [showCardsPlayedModal, setShowCardsPlayedModal] = useState(false);
  const [showDebugDropdown, setShowDebugDropdown] = useState(false);
  const [showPerformanceWindow, setShowPerformanceWindow] = useState(false);
  const [showBugReportWindow, setShowBugReportWindow] = useState(false);
  const [spectatePlayerId, setSpectatePlayerId] = useState<string | null>(null);
  const [isSpectator, setIsSpectator] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessageDto[]>([]);
  const [showCorp, setShowCorp] = useState(false);
  const [displayedInitPlayerId, setDisplayedInitPlayerId] = useState<string | null>(null);

  // Set corporation data directly from player (backend now sends full CardDto)
  useEffect(() => {
    if (currentPlayer?.corporation) {
      setCorporationData(currentPlayer.corporation);
    } else {
      setCorporationData(null);
    }
  }, [currentPlayer?.corporation]);

  // Corp reveal: set to true when entering a phase where the corp has been applied (handles reconnection)
  useEffect(() => {
    if (!game) return;
    const phase = game.currentPhase;
    if (
      phase === GamePhaseInitApplyPrelude ||
      phase === GamePhaseAction ||
      phase === GamePhaseProductionAndCardDraw ||
      phase === GamePhaseComplete
    ) {
      setShowCorp(true);
    }
  }, [game?.currentPhase]);

  // Delay init phase turn highlight until notification queue drains
  useEffect(() => {
    const initPlayerId = game?.initPhase?.currentPlayerId ?? null;
    const phase = game?.currentPhase;
    if (phase !== GamePhaseInitApplyCorp && phase !== GamePhaseInitApplyPrelude) {
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
  }, [game?.currentPhase, game?.initPhase?.currentPlayerId]);

  // Production phase modal state
  const [showProductionPhaseModal, setShowProductionPhaseModal] = useState(false);
  const [isProductionModalHidden, setIsProductionModalHidden] = useState(false);
  const [openProductionToCardSelection, setOpenProductionToCardSelection] = useState(false);
  const isInitialMount = useRef(true);

  // Starting selection state (corporation + preludes + project cards)
  const [showStartingSelection, setShowStartingSelection] = useState(false);

  // Pending card selection state (for sell patents, etc.)
  const [showPendingCardSelection, setShowPendingCardSelection] = useState(false);

  // Card draw selection state (for card-draw/peek/take/buy effects)
  const [showCardDrawSelection, setShowCardDrawSelection] = useState(false);

  // Card discard selection state (for passive effects like Mars University)
  const [showCardDiscardSelection, setShowCardDiscardSelection] = useState(false);

  // End game tile highlighting state
  const [tileHighlightMode, setTileHighlightMode] = useState<TileHighlightMode>(null);

  // End game VP indicators state
  const [vpIndicators, setVPIndicators] = useState<TileVPIndicator[]>([]);

  // Choice selection state (for card play)
  const [showChoiceSelection, setShowChoiceSelection] = useState(false);
  const [cardPendingChoice, setCardPendingChoice] = useState<PlayerCardDto | null>(null);
  const [pendingCardBehaviorIndex, setPendingCardBehaviorIndex] = useState(0);

  // Action choice selection state (for playing actions with choices)
  const [showActionChoiceSelection, setShowActionChoiceSelection] = useState(false);
  const [actionPendingChoice, setActionPendingChoice] = useState<PlayerActionDto | null>(null);

  // Action reuse state (for Viron-style abilities)
  const [showActionReuseSelection, setShowActionReuseSelection] = useState(false);
  const [pendingActionReuse, setPendingActionReuse] = useState<{
    cardId: string;
    behaviorIndex: number;
  } | null>(null);
  const activeReuseSourceCardId = useRef<string | undefined>(undefined);

  // Steal target selection state (from server pending selection after tile placement)
  const [showStealTargetSelection, setShowStealTargetSelection] = useState(false);

  // Colony resource selection state (from server pending selection after colony trade/build)
  const [showColonyResourceSelection, setShowColonyResourceSelection] = useState(false);

  // Passive triggered behavior choice state (from server pending selection)
  const [showBehaviorChoiceSelection, setShowBehaviorChoiceSelection] = useState(false);
  const [pendingBehaviorChoiceStorage, setPendingBehaviorChoiceStorage] = useState<{
    choiceIndex: number;
    allStorageNeeds: Array<{ resourceType: ResourceType; amount: number; selectorTags?: string[] }>;
    collectedTargets: string[];
    currentIndex: number;
    resourceType: ResourceType;
    amount: number;
    selectorTags?: string[];
  } | null>(null);
  const [showBehaviorChoiceStorage, setShowBehaviorChoiceStorage] = useState(false);

  // Payment selection state
  const [showPaymentSelection, setShowPaymentSelection] = useState(false);
  const [pendingCardPayment, setPendingCardPayment] = useState<{
    card: PlayerCardDto;
    choiceIndex?: number;
    cardStorageTargets?: string[];
  } | null>(null);

  // Card storage selection state
  const [showCardStorageSelection, setShowCardStorageSelection] = useState(false);
  const [pendingCardStorage, setPendingCardStorage] = useState<{
    cardId: string;
    choiceIndex?: number;
    allStorageNeeds: Array<{ resourceType: ResourceType; amount: number; selectorTags?: string[] }>;
    collectedTargets: string[];
    currentIndex: number;
    resourceType: ResourceType;
    amount: number;
    selectorTags?: string[];
  } | null>(null);

  // Action storage selection state
  const [showActionStorageSelection, setShowActionStorageSelection] = useState(false);
  const [pendingActionStorage, setPendingActionStorage] = useState<{
    cardId: string;
    behaviorIndex: number;
    choiceIndex?: number;
    allStorageNeeds: Array<{ resourceType: ResourceType; amount: number; selectorTags?: string[] }>;
    collectedTargets: string[];
    currentIndex: number;
    resourceType: ResourceType;
    amount: number;
    selectorTags?: string[];
  } | null>(null);

  // Target player selection state (for any-player resource/production removal)
  const [showTargetPlayerSelection, setShowTargetPlayerSelection] = useState(false);
  const [pendingTargetPlayer, setPendingTargetPlayer] = useState<{
    cardId: string;
    payment: CardPaymentDto;
    choiceIndex?: number;
    cardStorageTargets?: string[];
    selectedAmount?: number;
    resourceType: ResourceType;
    amount: number;
    isSteal: boolean;
  } | null>(null);

  // Action target player selection state
  const [showActionTargetPlayerSelection, setShowActionTargetPlayerSelection] = useState(false);
  const [pendingActionTargetPlayer, setPendingActionTargetPlayer] = useState<{
    cardId: string;
    behaviorIndex: number;
    choiceIndex?: number;
    cardStorageTargets?: string[];
    resourceType: ResourceType;
    amount: number;
    isSteal: boolean;
  } | null>(null);

  // Card resource selection state (for steal-from-any-card inputs like Predators/Ants)
  const [showCardResourceSelection, setShowCardResourceSelection] = useState(false);
  const [pendingCardResourceInput, setPendingCardResourceInput] = useState<{
    cardId: string;
    behaviorIndex: number;
    choiceIndex?: number;
    cardStorageTargets?: string[];
    resourceType: ResourceType;
    amount: number;
  } | null>(null);

  // Variable amount selection state (for variableAmount cards like Insulation, Power Infrastructure)
  const [showAmountSelection, setShowAmountSelection] = useState(false);
  const [pendingVariableAmount, setPendingVariableAmount] = useState<{
    type: "play-card" | "card-action";
    cardId: string;
    cardName: string;
    payment?: CardPaymentDto;
    choiceIndex?: number;
    cardStorageTargets?: string[];
    behaviorIndex?: number;
    resourceLabel: string;
    maxAmount: number;
  } | null>(null);

  // Tab management
  const [showTabConflict, setShowTabConflict] = useState(false);
  const [conflictingTabInfo, setConflictingTabInfo] = useState<{
    gameId: string;
    playerName: string;
  } | null>(null);

  // Leave game confirmation
  const [showLeaveGameConfirm, setShowLeaveGameConfirm] = useState(false);
  // End game confirmation
  const [showEndGameConfirm, setShowEndGameConfirm] = useState(false);

  // Change detection
  const previousGameRef = useRef<GameDto | null>(null);
  const [changedPaths, setChangedPaths] = useState<Set<string>>(new Set());

  // Triggered effects notifications
  const [triggeredEffects, setTriggeredEffects] = useState<TriggeredEffectDto[]>([]);
  const notificationQueueDoneAt = useRef<number>(0);

  const [isSkyboxReady, setIsSkyboxReady] = useState(() => skyboxCache.isReady());
  const [isGpuReady, setIsGpuReady] = useState(false);
  const [overlayVisible, setOverlayVisible] = useState(true);
  const { isLoaded: isSpaceBgLoaded } = useSpaceBackground();
  const handleSkyboxReady = useCallback(() => setIsSkyboxReady(true), []);
  const handleGpuReady = useCallback(() => setIsGpuReady(true), []);
  const handleLoadingTransitionEnd = useCallback(() => {
    if (transitionPhase === "loading") {
      setTransitionPhase("fadeOutLobby");
    } else {
      setOverlayVisible(false);
    }
  }, [transitionPhase]);

  const isFullyLoaded =
    (loadingPhase === "selecting" && isSpaceBgLoaded) ||
    (loadingPhase === "joining" && isSpaceBgLoaded) ||
    (loadingPhase === "spectating" && isSpaceBgLoaded) ||
    (isConnected &&
      !!game &&
      !isReconnecting &&
      (isSkyboxReady || transitionPhase === "lobby") &&
      (isGpuReady || transitionPhase === "lobby"));

  // WebSocket stability
  const isWebSocketInitialized = useRef(false);
  const currentPlayerIdRef = useRef<string | null>(null);

  // Stable WebSocket event handlers using useCallback
  const handleGameUpdated = useCallback(
    (updatedGame: GameDto) => {
      const playerId = currentPlayerIdRef.current;
      if (!playerId && !updatedGame.isSpectator) return;

      // Detect changes before updating
      if (previousGameRef.current) {
        const changes = findChangedPaths(previousGameRef.current, updatedGame);
        setChangedPaths(changes);

        // Play temperature increase sound when temperature goes up
        const prevTemp = previousGameRef.current.globalParameters?.temperature;
        const newTemp = updatedGame.globalParameters?.temperature;
        if (prevTemp !== undefined && newTemp !== undefined && newTemp > prevTemp) {
          void playTemperatureSound();
        }

        // Play oxygen increase sound when oxygen goes up, or when a greenery is placed at max oxygen
        const prevOxygen = previousGameRef.current.globalParameters?.oxygen;
        const newOxygen = updatedGame.globalParameters?.oxygen;
        const greeneryTypes = new Set([
          "greenery-tile",
          "ecological-zone-tile",
          "natural-preserve-tile",
        ]);
        const prevGreeneryCount =
          previousGameRef.current.board?.tiles?.filter(
            (t) => t.occupiedBy && greeneryTypes.has(t.occupiedBy.type),
          ).length ?? 0;
        const newGreeneryCount =
          updatedGame.board?.tiles?.filter(
            (t) => t.occupiedBy && greeneryTypes.has(t.occupiedBy.type),
          ).length ?? 0;
        const greeneryPlaced = newGreeneryCount > prevGreeneryCount;
        if (prevOxygen !== undefined && newOxygen !== undefined && newOxygen > prevOxygen) {
          void playOxygenSound();
        } else if (greeneryPlaced) {
          void playOxygenSound();
        }

        // Detect when it becomes this player's action turn
        const prevTurn = previousGameRef.current.currentTurn;
        const prevPhase = previousGameRef.current.currentPhase;
        const newTurn = updatedGame.currentTurn;
        const newPhase = updatedGame.currentPhase;
        const isNowMyActionTurn = newTurn === playerId && newPhase === GamePhaseAction;
        const wasMyActionTurn = prevTurn === playerId && prevPhase === GamePhaseAction;
        if (isNowMyActionTurn && !wasMyActionTurn) {
          void playYourTurnSound();
          enqueueGameEvent({ title: "YOUR TURN", duration: 2500 });
        }

        // Detect milestone claims
        const prevMilestones = previousGameRef.current.milestones;
        const newMilestones = updatedGame.milestones;
        if (prevMilestones && newMilestones) {
          for (const ms of newMilestones) {
            if (ms.isClaimed && ms.claimedBy) {
              const prev = prevMilestones.find((p) => p.type === ms.type);
              if (prev && !prev.isClaimed) {
                const allPlayers = [updatedGame.currentPlayer, ...(updatedGame.otherPlayers ?? [])];
                const claimPlayer = allPlayers.find((p) => p.id === ms.claimedBy);
                enqueueGameEvent({
                  title: "MILESTONE CLAIMED",
                  achievementName: ms.name,
                  playerName: claimPlayer?.name ?? "Unknown",
                  playerColor: claimPlayer?.color ?? "#64c8ff",
                  duration: 4000,
                });
              }
            }
          }
        }

        // Detect award funding
        const prevAwards = previousGameRef.current.awards;
        const newAwards = updatedGame.awards;
        if (prevAwards && newAwards) {
          for (const aw of newAwards) {
            if (aw.isFunded && aw.fundedBy) {
              const prev = prevAwards.find((p) => p.type === aw.type);
              if (prev && !prev.isFunded) {
                const allPlayers = [updatedGame.currentPlayer, ...(updatedGame.otherPlayers ?? [])];
                const fundPlayer = allPlayers.find((p) => p.id === aw.fundedBy);
                enqueueGameEvent({
                  title: "AWARD FUNDED",
                  achievementName: aw.name,
                  playerName: fundPlayer?.name ?? "Unknown",
                  playerColor: fundPlayer?.color ?? "#64c8ff",
                  duration: 4000,
                });
              }
            }
          }
        }

        // Clear changed paths after animation completes
        if (changes.size > 0) {
          setTimeout(() => {
            setChangedPaths(new Set());
          }, 1500);
        }
      }

      if (!previousGameRef.current && updatedGame.chatMessages?.length) {
        setChatMessages(updatedGame.chatMessages);
      }

      // Store the previous state for next comparison
      previousGameRef.current = deepClone(updatedGame);

      // Extract triggered effects for notifications (clear after short delay to allow component to process)
      if (updatedGame.triggeredEffects && updatedGame.triggeredEffects.length > 0) {
        setTriggeredEffects(updatedGame.triggeredEffects);
        const notificationCount = updatedGame.triggeredEffects.length;
        notificationQueueDoneAt.current = Date.now() + notificationCount * 2500;
        setTimeout(() => setTriggeredEffects([]), 100);

        // Reveal corp logo when we see triggered effects for our player during init_apply_corp
        if (updatedGame.currentPhase === GamePhaseInitApplyCorp) {
          const hasMyEffect = updatedGame.triggeredEffects.some(
            (e) => e.playerId === updatedGame.currentPlayer?.id,
          );
          if (hasMyEffect) setShowCorp(true);
        }
      }

      setGame(updatedGame);
      setIsConnected(true);

      // If we were reconnecting, mark reconnection as successful
      if (isReconnecting) {
        setIsReconnecting(false);
        setReconnectionStep(null);
      }

      // Set current player from updated game data
      const updatedPlayer = updatedGame.currentPlayer;
      setCurrentPlayer(updatedPlayer || null);

      // Show corporation modal if player hasn't selected a corporation yet
      if (updatedPlayer && !updatedPlayer.corporation) {
        setShowCorporationModal(true);
      } else {
        setShowCorporationModal(false);
      }
    },
    [isReconnecting, playTemperatureSound, playOxygenSound, playYourTurnSound],
  );

  const handleLogUpdate = useCallback((_logs: StateDiffDto[]) => {}, []);

  const handleFullState = useCallback(
    (statePayload: FullStatePayload) => {
      // Handle full-state message (e.g., on reconnection)
      if (statePayload.game) {
        handleGameUpdated(statePayload.game);
      }
    },
    [handleGameUpdated],
  );

  const handleError = useCallback(() => {
    // Could show error modal
  }, []);

  const handlePlayerKicked = useCallback(() => {
    clearGameSession();
    globalWebSocketManager.disconnect();
    navigate("/", { replace: true });
    showNotification({
      message: "You were kicked from the game",
      type: "info",
    });
  }, [navigate, showNotification]);

  const handleGameEnded = useCallback(() => {
    clearGameSession();
    globalWebSocketManager.disconnect();
    navigate("/", { replace: true });
    showNotification({
      message: "The host ended the game",
      type: "info",
    });
  }, [navigate, showNotification]);

  const handleDisconnect = useCallback(() => {
    // Skip disconnect handling if this was an intentional leave
    if (globalWebSocketManager.isGracefulDisconnect()) {
      return;
    }

    // WebSocket connection closed - this client lost connection
    setIsConnected(false);

    // Only start reconnection if we were actually connected to a game
    if (currentPlayerIdRef.current) {
      // Start in-place reconnection instead of redirecting
      setIsReconnecting(true);

      const savedGameData = getGameSession();
      if (savedGameData) {
        // Attempt to reconnect in place
        void attemptReconnection();
      } else {
        // No saved game data, go to main menu
        clearGameSession();
        navigate("/", { replace: true });
      }
    }
  }, [navigate]);

  const handlePlayerDisconnected = useCallback(
    (_payload: PlayerDisconnectedPayload) => {
      // Handle when any player disconnects (NOT this client)
      // Player disconnected from the game
      // Note: PlayerDisconnectedPayload no longer contains game data
      // Game state updates will come through separate game-updated events
    },
    [handleGameUpdated],
  );

  const handleMaxReconnectsReached = useCallback(() => {
    clearGameSession();
    navigate("/", { state: { error: "Server is down", persistent: true } });
  }, [navigate]);

  const handleChatUpdate = useCallback((chatMessage: ChatMessageDto) => {
    setChatMessages((prev) => [...prev, chatMessage]);
  }, []);

  const handleSpectatorKicked = useCallback(() => {
    globalWebSocketManager.disconnect();
    navigate("/", { replace: true });
  }, [navigate]);

  const handleSpectatorIdReceived = useCallback((payload: { spectatorId: string }) => {
    const gameId = globalWebSocketManager.gameId;
    if (!gameId || !payload.spectatorId) return;
    currentPlayerIdRef.current = payload.spectatorId;
    const savedSession = getGameSession();
    if (savedSession && savedSession.gameId === gameId) {
      saveGameSession({
        ...savedSession,
        playerId: payload.spectatorId,
        isSpectator: true,
      });
    }
  }, []);

  // Check if we should show production phase modal based on game state
  useEffect(() => {
    if (!game || !currentPlayer) return;

    // Check if current player has production phase data
    const hasProductionData =
      currentPlayer.productionPhase &&
      !currentPlayer.productionPhase.selectionComplete &&
      currentPlayer.productionPhase.availableCards;

    if (hasProductionData && !showProductionPhaseModal) {
      // Only play sound if this is not the initial mount (skip on page reload)
      if (!isInitialMount.current) {
        void playProductionSound();
      }
      setShowProductionPhaseModal(true);
      // Reset the flag for opening directly to card selection on new production phase
      setOpenProductionToCardSelection(false);
    } else if (!hasProductionData && showProductionPhaseModal) {
      // Production phase is over, hide the modal
      setShowProductionPhaseModal(false);
    }

    // Mark that initial mount is complete
    if (isInitialMount.current) {
      isInitialMount.current = false;
    }
  }, [currentPlayer?.productionPhase, game, showProductionPhaseModal, playProductionSound]);

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

  // Handle pending card selection (sell patents, etc.)
  const handlePendingCardSelection = useCallback(async (selectedCardIds: string[]) => {
    try {
      await globalWebSocketManager.selectCards(selectedCardIds);
      // Overlay closes automatically when backend clears pendingCardSelection
    } catch (error) {
      console.error("Failed to select cards:", error);
    }
  }, []);

  // Handle card draw selection confirmation
  const handleCardDrawConfirm = useCallback(async (cardsToTake: string[], cardsToBuy: string[]) => {
    try {
      await globalWebSocketManager.confirmCardDraw(cardsToTake, cardsToBuy);
    } catch (error) {
      console.error("Failed to confirm card draw:", error);
    }
  }, []);

  // Handle card discard selection confirmation (passive effects like Mars University)
  const handleCardDiscardConfirm = useCallback(async (cardsToDiscard: string[]) => {
    try {
      await globalWebSocketManager.confirmCardDiscard(cardsToDiscard);
    } catch (error) {
      console.error("Failed to confirm card discard:", error);
    }
  }, []);

  // Helper function to get ALL any-card storage selections from outputs
  const getAllAnyCardStorageSelections = useCallback(
    (
      outputs: any[] | undefined,
    ): Array<{
      resourceType: ResourceType;
      amount: number;
      target: string;
      selectorTags?: string[];
    }> => {
      if (!outputs) return [];

      const storageResources = [
        "animal",
        "microbe",
        "floater",
        "science",
        "asteroid",
        "card-resource",
      ] as ResourceType[];

      const results: Array<{
        resourceType: ResourceType;
        amount: number;
        target: string;
        selectorTags?: string[];
      }> = [];

      for (const output of outputs) {
        if (
          (output.target === "any-card" || output.target === "self-card") &&
          storageResources.includes(output.type as ResourceType)
        ) {
          let selectorTags: string[] | undefined;
          if (output.selectors) {
            selectorTags = output.selectors.flatMap((s: any) => s.tags || []);
          }
          // Only include any-card outputs (self-card auto-resolves)
          if (output.target === "any-card") {
            results.push({
              resourceType: output.type as ResourceType,
              amount: output.amount || 1,
              target: output.target as string,
              selectorTags,
            });
          }
        }
      }

      return results;
    },
    [],
  );

  // Helper function to check if outputs need card storage selection (returns first match)
  const needsCardStorageSelection = useCallback(
    (
      outputs: any[] | undefined,
    ): {
      resourceType: ResourceType;
      amount: number;
      target: string;
      selectorTags?: string[];
    } | null => {
      if (!outputs) return null;

      const storageResources = [
        "animal",
        "microbe",
        "floater",
        "science",
        "asteroid",
        "card-resource",
      ] as ResourceType[];

      for (const output of outputs) {
        if (
          (output.target === "any-card" || output.target === "self-card") &&
          storageResources.includes(output.type as ResourceType)
        ) {
          let selectorTags: string[] | undefined;
          if (output.selectors) {
            selectorTags = output.selectors.flatMap((s: any) => s.tags || []);
          }
          return {
            resourceType: output.type as ResourceType,
            amount: output.amount || 1,
            target: output.target as string,
            selectorTags,
          };
        }
      }

      return null;
    },
    [],
  );

  const needsTargetPlayerSelection = useCallback(
    (
      outputs: any[] | undefined,
    ): {
      resourceType: ResourceType;
      amount: number;
      isSteal: boolean;
    } | null => {
      if (!outputs) return null;

      // Solo mode: no other players, skip target selection
      if (!game?.otherPlayers || game.otherPlayers.length === 0) return null;

      const targetableResources = [
        "credit",
        "steel",
        "titanium",
        "plant",
        "energy",
        "heat",
        "credit-production",
        "steel-production",
        "titanium-production",
        "plant-production",
        "energy-production",
        "heat-production",
      ] as ResourceType[];

      for (const output of outputs) {
        if (output.targetRestriction) continue; // Deferred to post-tile-placement
        if (
          (output.target === "any-player" || output.target === "steal-any-player") &&
          targetableResources.includes(output.type as ResourceType)
        ) {
          return {
            resourceType: output.type as ResourceType,
            amount: output.amount || 1,
            isSteal: output.target === "steal-any-player",
          };
        }
      }

      return null;
    },
    [game?.otherPlayers],
  );

  const needsCardResourceInput = useCallback(
    (inputs: any[] | undefined): { resourceType: ResourceType; amount: number } | null => {
      if (!inputs) return null;

      for (const input of inputs) {
        if (input.target === "steal-from-any-card") {
          return {
            resourceType: input.type as ResourceType,
            amount: input.amount || 1,
          };
        }
      }

      return null;
    },
    [],
  );

  // Helper to detect variableAmount in a behavior and compute the max selectable amount
  const getVariableAmountInfo = useCallback(
    (
      inputs: any[] | undefined,
      outputs: any[] | undefined,
    ): { resourceLabel: string; maxAmount: number } | null => {
      if (!currentPlayer) return null;

      // Check inputs for variableAmount (e.g., Power Infrastructure: spend energy)
      if (inputs) {
        for (const input of inputs) {
          if (!input.variableAmount) continue;
          const resType = input.type as string;
          let max = 0;
          const resources = currentPlayer.resources;
          if (resType === "energy") max = resources.energy;
          else if (resType === "heat") max = resources.heat;
          else if (resType === "credit") max = resources.credits;
          else if (resType === "steel") max = resources.steel;
          else if (resType === "titanium") max = resources.titanium;
          else if (resType === "plant") max = resources.plants;
          if (max > 0) return { resourceLabel: resType, maxAmount: max };
        }
      }

      // Check outputs for variableAmount with negative amounts (e.g., Insulation: decrease heat production)
      if (outputs) {
        for (const output of outputs) {
          if (!output.variableAmount || output.amount >= 0) continue;
          const resType = output.type as string;
          let max = 0;
          const production = currentPlayer.production;
          if (resType === "heat-production") max = production.heat;
          else if (resType === "energy-production") max = production.energy;
          else if (resType === "credit-production") max = production.credits;
          else if (resType === "steel-production") max = production.steel;
          else if (resType === "titanium-production") max = production.titanium;
          else if (resType === "plant-production") max = production.plants;
          const label = resType.replace("-production", " production");
          if (max > 0) return { resourceLabel: label, maxAmount: max };
        }
      }

      return null;
    },
    [currentPlayer],
  );

  const finalizePlayCard = useCallback(
    async (
      cardId: string,
      payment: CardPaymentDto,
      choiceIndex?: number,
      cardStorageTargets?: string[],
      cardForBehaviors?: CardDto,
      selectedAmount?: number,
    ) => {
      // Check if any auto-trigger behavior has variableAmount and we haven't collected it yet
      const card = cardForBehaviors || currentPlayer?.cards.find((c) => c.id === cardId);
      if (card && selectedAmount === undefined) {
        const autoTriggerBehaviors = card.behaviors?.filter((b) =>
          b.triggers?.some((t) => t.type === "auto"),
        );
        for (const behavior of autoTriggerBehaviors || []) {
          const matchedChoice =
            choiceIndex !== undefined
              ? behavior.choices?.find((c) => c.originalIndex === choiceIndex)
              : undefined;
          const outputs = matchedChoice ? matchedChoice.outputs : behavior.outputs;
          const inputs = matchedChoice ? matchedChoice.inputs : behavior.inputs;
          const variableInfo = getVariableAmountInfo(inputs, outputs);
          if (variableInfo) {
            setPendingVariableAmount({
              type: "play-card",
              cardId,
              cardName: card.name,
              payment,
              choiceIndex,
              cardStorageTargets,
              resourceLabel: variableInfo.resourceLabel,
              maxAmount: variableInfo.maxAmount,
            });
            setShowAmountSelection(true);
            return;
          }
        }
      }

      // Check if any auto-trigger behavior needs target player selection
      if (card) {
        const autoTriggerBehaviors = card.behaviors?.filter((b) =>
          b.triggers?.some((t) => t.type === "auto"),
        );
        for (const behavior of autoTriggerBehaviors || []) {
          const matchedChoice2 =
            choiceIndex !== undefined
              ? behavior.choices?.find((c) => c.originalIndex === choiceIndex)
              : undefined;
          const outputs = matchedChoice2 ? matchedChoice2.outputs : behavior.outputs;
          const targetInfo = needsTargetPlayerSelection(outputs);
          if (targetInfo) {
            setPendingTargetPlayer({
              cardId,
              payment,
              choiceIndex,
              cardStorageTargets,
              selectedAmount,
              resourceType: targetInfo.resourceType,
              amount: targetInfo.amount,
              isSteal: targetInfo.isSteal,
            });
            setShowTargetPlayerSelection(true);
            return;
          }
        }
      }

      await globalWebSocketManager.playCard(
        cardId,
        payment,
        choiceIndex,
        cardStorageTargets,
        undefined,
        selectedAmount,
      );
    },
    [currentPlayer?.cards, needsTargetPlayerSelection, getVariableAmountInfo],
  );

  const handlePlayCard = useCallback(
    async (cardId: string) => {
      try {
        // Block card plays when it's not your turn
        if (game?.currentTurn !== game?.viewingPlayerId) {
          throw new Error("Not your turn");
        }

        // Block card plays when tile selection is pending
        if (currentPlayer?.pendingTileSelection) {
          return;
        }

        // Find the card to check if it has choices
        const card = currentPlayer?.cards.find((c) => c.id === cardId);
        if (!card) {
          console.error(`Card ${cardId} not found in player's hand`);
          return;
        }

        // Check if any AUTO-triggered behavior has choices
        // Manual-triggered behaviors (actions) will show choices when the action is played
        const behaviorWithChoices = card.behaviors?.findIndex(
          (b) => b.choices && b.choices.length > 0 && b.triggers?.some((t) => t.type === "auto"),
        );

        if (
          behaviorWithChoices !== undefined &&
          behaviorWithChoices >= 0 &&
          card.behaviors?.[behaviorWithChoices]?.choices
        ) {
          // Card has auto-triggered choices, show the choice selection popover
          setCardPendingChoice(card);
          setPendingCardBehaviorIndex(behaviorWithChoices);
          setShowChoiceSelection(true);
        } else {
          // No auto-triggered choices — check storage first, then payment

          // Collect ALL any-card storage needs from auto-trigger behaviors
          const autoTriggerBehaviors = card.behaviors?.filter((b) =>
            b.triggers?.some((t) => t.type === "auto"),
          );

          const allStorageNeeds: Array<{
            resourceType: ResourceType;
            amount: number;
            selectorTags?: string[];
          }> = [];
          for (const behavior of autoTriggerBehaviors || []) {
            const selections = getAllAnyCardStorageSelections(behavior.outputs);
            for (const sel of selections) {
              allStorageNeeds.push({
                resourceType: sel.resourceType,
                amount: sel.amount,
                selectorTags: sel.selectorTags,
              });
            }
          }

          if (allStorageNeeds.length > 0) {
            // Start sequential storage prompting
            const first = allStorageNeeds[0];
            setPendingCardStorage({
              cardId: card.id,
              choiceIndex: undefined,
              allStorageNeeds,
              collectedTargets: [],
              currentIndex: 0,
              resourceType: first.resourceType,
              amount: first.amount,
              selectorTags: first.selectorTags,
            });
            setShowCardStorageSelection(true);
          } else if (
            currentPlayer &&
            shouldShowPaymentModal(
              card,
              currentPlayer.resources,
              currentPlayer.paymentSubstitutes,
              currentPlayer.storagePaymentSubstitutes,
              currentPlayer.resourceStorage,
            )
          ) {
            // No any-card storage needed, show payment selection modal
            setPendingCardPayment({
              card: card,
              choiceIndex: undefined,
            });
            setShowPaymentSelection(true);
          } else {
            // No storage, no payment modal — play directly
            const payment = createDefaultPayment(card.effectiveCost);
            await finalizePlayCard(cardId, payment, undefined, undefined, card);
          }
        }
      } catch (error) {
        console.error(`❌ Failed to play card ${cardId}:`, error);
        throw error; // Re-throw to allow CardFanOverlay to handle the error
      }
    },
    [
      currentPlayer?.cards,
      getAllAnyCardStorageSelections,
      finalizePlayCard,
      game?.currentTurn,
      game?.viewingPlayerId,
    ],
  );

  const handleChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      if (!cardPendingChoice || !currentPlayer) return;

      try {
        setShowChoiceSelection(false);

        // Check storage first, then payment
        const behavior = cardPendingChoice.behaviors?.[pendingCardBehaviorIndex];
        const selectedChoice = behavior?.choices?.find((c) => c.originalIndex === choiceIndex);

        // Collect ALL any-card storage needs from choice outputs and behavior outputs
        const allStorageNeeds: Array<{
          resourceType: ResourceType;
          amount: number;
          selectorTags?: string[];
        }> = [];
        const choiceSelections = getAllAnyCardStorageSelections(selectedChoice?.outputs);
        for (const sel of choiceSelections) {
          allStorageNeeds.push({
            resourceType: sel.resourceType,
            amount: sel.amount,
            selectorTags: sel.selectorTags,
          });
        }
        if (allStorageNeeds.length === 0) {
          const behaviorSelections = getAllAnyCardStorageSelections(behavior?.outputs);
          for (const sel of behaviorSelections) {
            allStorageNeeds.push({
              resourceType: sel.resourceType,
              amount: sel.amount,
              selectorTags: sel.selectorTags,
            });
          }
        }

        if (allStorageNeeds.length > 0) {
          // Start sequential storage prompting
          const first = allStorageNeeds[0];
          setPendingCardStorage({
            cardId: cardPendingChoice.id,
            choiceIndex: choiceIndex,
            allStorageNeeds,
            collectedTargets: [],
            currentIndex: 0,
            resourceType: first.resourceType,
            amount: first.amount,
            selectorTags: first.selectorTags,
          });
          setShowCardStorageSelection(true);
          setCardPendingChoice(null);
          setPendingCardBehaviorIndex(0);
        } else if (
          shouldShowPaymentModal(
            cardPendingChoice,
            currentPlayer.resources,
            currentPlayer.paymentSubstitutes,
            currentPlayer.storagePaymentSubstitutes,
            currentPlayer.resourceStorage,
          )
        ) {
          // No any-card storage needed, show payment selection modal
          setPendingCardPayment({
            card: cardPendingChoice,
            choiceIndex: choiceIndex,
          });
          setShowPaymentSelection(true);
          setCardPendingChoice(null);
          setPendingCardBehaviorIndex(0);
        } else {
          // No storage, no payment modal — play directly
          const payment = createDefaultPayment(cardPendingChoice.effectiveCost);
          await finalizePlayCard(
            cardPendingChoice.id,
            payment,
            choiceIndex,
            undefined,
            cardPendingChoice,
          );
          setCardPendingChoice(null);
          setPendingCardBehaviorIndex(0);
        }
      } catch (error) {
        console.error(
          `❌ Failed to play card ${cardPendingChoice.id} with choice ${choiceIndex}:`,
          error,
        );
        setCardPendingChoice(null);
        setPendingCardBehaviorIndex(0);
      }
    },
    [
      cardPendingChoice,
      currentPlayer,
      pendingCardBehaviorIndex,
      getAllAnyCardStorageSelections,
      finalizePlayCard,
    ],
  );

  const handleChoiceCancel = useCallback(() => {
    setShowChoiceSelection(false);
    setCardPendingChoice(null);
    setPendingCardBehaviorIndex(0);
  }, []);

  const handleActionChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      if (!actionPendingChoice) return;

      try {
        setShowActionChoiceSelection(false);

        // Get the selected choice by originalIndex (choices may have been filtered by policy)
        const selectedChoice = actionPendingChoice.behavior.choices?.find(
          (c) => c.originalIndex === choiceIndex,
        );

        // Check for self-card storage first
        const storageInfo = needsCardStorageSelection(selectedChoice?.outputs);

        if (storageInfo && storageInfo.target === "self-card") {
          // Self-card target: check target player needs
          const targetInfo = needsTargetPlayerSelection(selectedChoice?.outputs);
          if (targetInfo) {
            setPendingActionTargetPlayer({
              cardId: actionPendingChoice.cardId,
              behaviorIndex: actionPendingChoice.behaviorIndex,
              choiceIndex,
              cardStorageTargets: [actionPendingChoice.cardId],
              resourceType: targetInfo.resourceType,
              amount: targetInfo.amount,
              isSteal: targetInfo.isSteal,
            });
            setShowActionTargetPlayerSelection(true);
            setActionPendingChoice(null);
          } else {
            const reuseId = activeReuseSourceCardId.current;
            activeReuseSourceCardId.current = undefined;
            await globalWebSocketManager.playCardAction(
              actionPendingChoice.cardId,
              actionPendingChoice.behaviorIndex,
              choiceIndex,
              [actionPendingChoice.cardId],
              undefined,
              undefined,
              undefined,
              undefined,
              reuseId,
            );
            setActionPendingChoice(null);
          }
        } else {
          // Collect ALL any-card storage needs
          const allStorageNeeds: Array<{
            resourceType: ResourceType;
            amount: number;
            selectorTags?: string[];
          }> = [];
          const selections = getAllAnyCardStorageSelections(selectedChoice?.outputs);
          for (const sel of selections) {
            allStorageNeeds.push({
              resourceType: sel.resourceType,
              amount: sel.amount,
              selectorTags: sel.selectorTags,
            });
          }

          if (allStorageNeeds.length > 0) {
            const first = allStorageNeeds[0];
            setPendingActionStorage({
              cardId: actionPendingChoice.cardId,
              behaviorIndex: actionPendingChoice.behaviorIndex,
              choiceIndex: choiceIndex,
              allStorageNeeds,
              collectedTargets: [],
              currentIndex: 0,
              resourceType: first.resourceType,
              amount: first.amount,
              selectorTags: first.selectorTags,
            });
            setShowActionStorageSelection(true);
            setActionPendingChoice(null);
          } else {
            // No card storage needed, check target player needs
            const targetInfo = needsTargetPlayerSelection(selectedChoice?.outputs);
            if (targetInfo) {
              setPendingActionTargetPlayer({
                cardId: actionPendingChoice.cardId,
                behaviorIndex: actionPendingChoice.behaviorIndex,
                choiceIndex,
                resourceType: targetInfo.resourceType,
                amount: targetInfo.amount,
                isSteal: targetInfo.isSteal,
              });
              setShowActionTargetPlayerSelection(true);
              setActionPendingChoice(null);
            } else {
              const reuseId = activeReuseSourceCardId.current;
              activeReuseSourceCardId.current = undefined;
              await globalWebSocketManager.playCardAction(
                actionPendingChoice.cardId,
                actionPendingChoice.behaviorIndex,
                choiceIndex,
                undefined,
                undefined,
                undefined,
                undefined,
                undefined,
                reuseId,
              );
              setActionPendingChoice(null);
            }
          }
        }
      } catch (error) {
        console.error(
          `Failed to play action ${actionPendingChoice.cardId} with choice ${choiceIndex}:`,
          error,
        );
        setActionPendingChoice(null);
        activeReuseSourceCardId.current = undefined;
      }
    },
    [
      actionPendingChoice,
      needsCardStorageSelection,
      getAllAnyCardStorageSelections,
      needsTargetPlayerSelection,
    ],
  );

  const handleActionChoiceCancel = useCallback(() => {
    setShowActionChoiceSelection(false);
    setActionPendingChoice(null);
  }, []);

  const handleActionReuseSelect = useCallback(
    (targetAction: PlayerActionDto) => {
      if (!pendingActionReuse) return;
      setShowActionReuseSelection(false);

      // For the reused action, we need to handle its sub-flows (choices, storage, etc.)
      // Set the reuse source so downstream playCardAction calls include it
      activeReuseSourceCardId.current = pendingActionReuse.cardId;

      // Re-enter the normal action handling flow for the target action
      // Check if this action has choices
      if (targetAction.behavior.choices && targetAction.behavior.choices.length > 0) {
        setActionPendingChoice(targetAction);
        setShowActionChoiceSelection(true);
      } else {
        // Check for self-card storage
        const storageInfo = needsCardStorageSelection(targetAction.behavior.outputs);
        if (storageInfo && storageInfo.target === "self-card") {
          void globalWebSocketManager.playCardAction(
            targetAction.cardId,
            targetAction.behaviorIndex,
            undefined,
            [targetAction.cardId],
            undefined,
            undefined,
            undefined,
            undefined,
            pendingActionReuse.cardId,
          );
        } else {
          void globalWebSocketManager.playCardAction(
            targetAction.cardId,
            targetAction.behaviorIndex,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            pendingActionReuse.cardId,
          );
        }
        activeReuseSourceCardId.current = undefined;
      }
      setPendingActionReuse(null);
    },
    [pendingActionReuse, needsCardStorageSelection],
  );

  const handleActionReuseCancel = useCallback(() => {
    setShowActionReuseSelection(false);
    setPendingActionReuse(null);
    activeReuseSourceCardId.current = undefined;
  }, []);

  // Passive triggered behavior choice handlers
  const handleBehaviorChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      const pendingBehaviorChoice = game?.currentPlayer?.pendingBehaviorChoiceSelection;
      if (!pendingBehaviorChoice) return;

      try {
        const selectedChoice = pendingBehaviorChoice.choices.find(
          (c) => c.originalIndex === choiceIndex,
        );

        // Collect ALL any-card storage needs
        const allStorageNeeds: Array<{
          resourceType: ResourceType;
          amount: number;
          selectorTags?: string[];
        }> = [];
        const selections = getAllAnyCardStorageSelections(selectedChoice?.outputs);
        for (const sel of selections) {
          allStorageNeeds.push({
            resourceType: sel.resourceType,
            amount: sel.amount,
            selectorTags: sel.selectorTags,
          });
        }

        if (allStorageNeeds.length > 0) {
          const first = allStorageNeeds[0];
          setPendingBehaviorChoiceStorage({
            choiceIndex,
            allStorageNeeds,
            collectedTargets: [],
            currentIndex: 0,
            resourceType: first.resourceType,
            amount: first.amount,
            selectorTags: first.selectorTags,
          });
          setShowBehaviorChoiceStorage(true);
          setShowBehaviorChoiceSelection(false);
        } else {
          await globalWebSocketManager.confirmBehaviorChoice(choiceIndex);
        }
      } catch (error) {
        console.error("Failed to confirm behavior choice:", error);
      }
    },
    [game?.currentPlayer?.pendingBehaviorChoiceSelection, getAllAnyCardStorageSelections],
  );

  const handleBehaviorChoiceStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingBehaviorChoiceStorage) return;

      try {
        const newCollected = [...pendingBehaviorChoiceStorage.collectedTargets, targetCardId];
        const nextIndex = pendingBehaviorChoiceStorage.currentIndex + 1;

        if (nextIndex < pendingBehaviorChoiceStorage.allStorageNeeds.length) {
          // More storage needs remain - advance to next prompt
          const next = pendingBehaviorChoiceStorage.allStorageNeeds[nextIndex];
          setPendingBehaviorChoiceStorage({
            ...pendingBehaviorChoiceStorage,
            collectedTargets: newCollected,
            currentIndex: nextIndex,
            resourceType: next.resourceType,
            amount: next.amount,
            selectorTags: next.selectorTags,
          });
        } else {
          // All collected - send to backend
          setShowBehaviorChoiceStorage(false);
          await globalWebSocketManager.confirmBehaviorChoice(
            pendingBehaviorChoiceStorage.choiceIndex,
            newCollected,
          );
          setPendingBehaviorChoiceStorage(null);
        }
      } catch (error) {
        console.error("Failed to confirm behavior choice with storage target:", error);
        setPendingBehaviorChoiceStorage(null);
      }
    },
    [pendingBehaviorChoiceStorage],
  );

  const handleBehaviorChoiceStorageCancel = useCallback(() => {
    setShowBehaviorChoiceStorage(false);
    setPendingBehaviorChoiceStorage(null);
    setShowBehaviorChoiceSelection(true);
  }, []);

  // Steal target selection callbacks (deferred adjacent steal)
  const handleStealTargetSelect = useCallback(async (targetPlayerId: string) => {
    void globalWebSocketManager.confirmStealTarget(targetPlayerId);
  }, []);

  const handleStealTargetSkip = useCallback(async () => {
    void globalWebSocketManager.confirmStealTarget("");
  }, []);

  // Colony resource selection callbacks (card storage for colony trade/build rewards)
  const handleColonyResourceSelect = useCallback(async (cardId: string) => {
    void globalWebSocketManager.confirmColonyResource(cardId);
  }, []);

  const handleColonyResourceSkip = useCallback(async () => {
    void globalWebSocketManager.confirmColonyResource("");
  }, []);

  // Payment selection callbacks
  const handlePaymentConfirm = useCallback(
    async (payment: CardPaymentDto) => {
      if (!pendingCardPayment || !currentPlayer) return;

      try {
        setShowPaymentSelection(false);

        // Payment is the last step — storage (if needed) was already selected
        await finalizePlayCard(
          pendingCardPayment.card.id,
          payment,
          pendingCardPayment.choiceIndex,
          pendingCardPayment.cardStorageTargets,
          pendingCardPayment.card,
        );

        setPendingCardPayment(null);
      } catch (error) {
        console.error(`❌ Failed to play card with payment:`, error);
        setPendingCardPayment(null);
      }
    },
    [pendingCardPayment, currentPlayer, finalizePlayCard],
  );

  const handlePaymentCancel = useCallback(() => {
    setShowPaymentSelection(false);
    setPendingCardPayment(null);
  }, []);

  const handleCardStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingCardStorage || !currentPlayer) return;

      try {
        const newCollected = [...pendingCardStorage.collectedTargets, targetCardId];
        const nextIndex = pendingCardStorage.currentIndex + 1;

        if (nextIndex < pendingCardStorage.allStorageNeeds.length) {
          // More storage needs remain - advance to next prompt
          const next = pendingCardStorage.allStorageNeeds[nextIndex];
          setPendingCardStorage({
            ...pendingCardStorage,
            collectedTargets: newCollected,
            currentIndex: nextIndex,
            resourceType: next.resourceType,
            amount: next.amount,
            selectorTags: next.selectorTags,
          });
          return;
        }

        // All collected - proceed with payment or finalize
        setShowCardStorageSelection(false);
        const card = currentPlayer.cards.find((c) => c.id === pendingCardStorage.cardId);

        if (
          card &&
          shouldShowPaymentModal(
            card,
            currentPlayer.resources,
            currentPlayer.paymentSubstitutes,
            currentPlayer.storagePaymentSubstitutes,
            currentPlayer.resourceStorage,
          )
        ) {
          setPendingCardPayment({
            card: card,
            choiceIndex: pendingCardStorage.choiceIndex,
            cardStorageTargets: newCollected,
          });
          setShowPaymentSelection(true);
          setPendingCardStorage(null);
          return;
        }

        // No payment modal needed — finalize with default payment
        const payment = createDefaultPayment(card?.effectiveCost ?? 0);
        await finalizePlayCard(
          pendingCardStorage.cardId,
          payment,
          pendingCardStorage.choiceIndex,
          newCollected,
          card,
        );
        setPendingCardStorage(null);
      } catch (error) {
        console.error(
          `❌ Failed to play card ${pendingCardStorage.cardId} with card storage target ${targetCardId}:`,
          error,
        );
        setPendingCardStorage(null);
      }
    },
    [pendingCardStorage, currentPlayer, finalizePlayCard],
  );

  const handleCardStorageCancel = useCallback(() => {
    setShowCardStorageSelection(false);
    setPendingCardStorage(null);
  }, []);

  const handleActionStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingActionStorage) return;

      try {
        const newCollected = [...pendingActionStorage.collectedTargets, targetCardId];
        const nextIndex = pendingActionStorage.currentIndex + 1;

        if (nextIndex < pendingActionStorage.allStorageNeeds.length) {
          // More storage needs remain - advance to next prompt
          const next = pendingActionStorage.allStorageNeeds[nextIndex];
          setPendingActionStorage({
            ...pendingActionStorage,
            collectedTargets: newCollected,
            currentIndex: nextIndex,
            resourceType: next.resourceType,
            amount: next.amount,
            selectorTags: next.selectorTags,
          });
          return;
        }

        // All collected - proceed
        setShowActionStorageSelection(false);

        // Check if the action also needs target player selection
        const actionPlayer = currentPlayer;
        const action = actionPlayer?.actions?.find(
          (a) =>
            a.cardId === pendingActionStorage.cardId &&
            a.behaviorIndex === pendingActionStorage.behaviorIndex,
        );
        const outputs =
          pendingActionStorage.choiceIndex !== undefined
            ? action?.behavior.choices?.find(
                (c) => c.originalIndex === pendingActionStorage.choiceIndex,
              )?.outputs
            : action?.behavior.outputs;
        const targetInfo = needsTargetPlayerSelection(outputs);

        if (targetInfo) {
          setPendingActionTargetPlayer({
            cardId: pendingActionStorage.cardId,
            behaviorIndex: pendingActionStorage.behaviorIndex,
            choiceIndex: pendingActionStorage.choiceIndex,
            cardStorageTargets: newCollected,
            resourceType: targetInfo.resourceType,
            amount: targetInfo.amount,
            isSteal: targetInfo.isSteal,
          });
          setShowActionTargetPlayerSelection(true);
          setPendingActionStorage(null);
          return;
        }

        await globalWebSocketManager.playCardAction(
          pendingActionStorage.cardId,
          pendingActionStorage.behaviorIndex,
          pendingActionStorage.choiceIndex,
          newCollected,
        );
        setPendingActionStorage(null);
      } catch (error) {
        console.error(
          `❌ Failed to play action ${pendingActionStorage.cardId} with card storage target ${targetCardId}:`,
          error,
        );
        setPendingActionStorage(null);
      }
    },
    [pendingActionStorage, currentPlayer, needsTargetPlayerSelection],
  );

  const handleActionStorageCancel = useCallback(() => {
    setShowActionStorageSelection(false);
    setPendingActionStorage(null);
  }, []);

  const handleTargetPlayerSelect = useCallback(
    async (targetPlayerId: string) => {
      if (!pendingTargetPlayer) return;

      try {
        setShowTargetPlayerSelection(false);
        await globalWebSocketManager.playCard(
          pendingTargetPlayer.cardId,
          pendingTargetPlayer.payment,
          pendingTargetPlayer.choiceIndex,
          pendingTargetPlayer.cardStorageTargets,
          targetPlayerId,
          pendingTargetPlayer.selectedAmount,
        );
        setPendingTargetPlayer(null);
      } catch (error) {
        console.error(
          `❌ Failed to play card ${pendingTargetPlayer.cardId} with target player ${targetPlayerId}:`,
          error,
        );
        setPendingTargetPlayer(null);
      }
    },
    [pendingTargetPlayer],
  );

  const handleTargetPlayerCancel = useCallback(() => {
    setShowTargetPlayerSelection(false);
    setPendingTargetPlayer(null);
  }, []);

  // Amount selection handlers for variableAmount cards (Insulation, Power Infrastructure)
  const handleAmountSelect = useCallback(
    async (amount: number) => {
      if (!pendingVariableAmount) return;

      try {
        setShowAmountSelection(false);
        if (pendingVariableAmount.type === "play-card") {
          await finalizePlayCard(
            pendingVariableAmount.cardId,
            pendingVariableAmount.payment!,
            pendingVariableAmount.choiceIndex,
            pendingVariableAmount.cardStorageTargets,
            undefined,
            amount,
          );
        } else if (pendingVariableAmount.type === "card-action") {
          await globalWebSocketManager.playCardAction(
            pendingVariableAmount.cardId,
            pendingVariableAmount.behaviorIndex!,
            pendingVariableAmount.choiceIndex,
            pendingVariableAmount.cardStorageTargets,
            undefined,
            undefined,
            amount,
          );
        }
        setPendingVariableAmount(null);
      } catch (error) {
        console.error(`❌ Failed to execute with amount ${amount}:`, error);
        setPendingVariableAmount(null);
      }
    },
    [pendingVariableAmount, finalizePlayCard],
  );

  const handleAmountCancel = useCallback(() => {
    setShowAmountSelection(false);
    setPendingVariableAmount(null);
  }, []);

  const handleActionTargetPlayerSelect = useCallback(
    async (targetPlayerId: string) => {
      if (!pendingActionTargetPlayer) return;

      try {
        setShowActionTargetPlayerSelection(false);
        await globalWebSocketManager.playCardAction(
          pendingActionTargetPlayer.cardId,
          pendingActionTargetPlayer.behaviorIndex,
          pendingActionTargetPlayer.choiceIndex,
          pendingActionTargetPlayer.cardStorageTargets,
          targetPlayerId,
        );
        setPendingActionTargetPlayer(null);
      } catch (error) {
        console.error(
          `❌ Failed to play action ${pendingActionTargetPlayer.cardId} with target player ${targetPlayerId}:`,
          error,
        );
        setPendingActionTargetPlayer(null);
      }
    },
    [pendingActionTargetPlayer],
  );

  const handleActionTargetPlayerCancel = useCallback(() => {
    setShowActionTargetPlayerSelection(false);
    setPendingActionTargetPlayer(null);
  }, []);

  const handleCardResourceSelect = useCallback(
    async (sourceCardId: string) => {
      if (!pendingCardResourceInput) return;

      try {
        setShowCardResourceSelection(false);
        await globalWebSocketManager.playCardAction(
          pendingCardResourceInput.cardId,
          pendingCardResourceInput.behaviorIndex,
          pendingCardResourceInput.choiceIndex,
          pendingCardResourceInput.cardStorageTargets,
          undefined,
          sourceCardId,
        );
        setPendingCardResourceInput(null);
      } catch (error) {
        console.error(
          `Failed to play action ${pendingCardResourceInput.cardId} with source card ${sourceCardId}:`,
          error,
        );
        setPendingCardResourceInput(null);
      }
    },
    [pendingCardResourceInput],
  );

  const handleCardResourceCancel = useCallback(() => {
    setShowCardResourceSelection(false);
    setPendingCardResourceInput(null);
  }, []);

  // Attempt reconnection to the game
  const attemptReconnection = useCallback(async () => {
    try {
      const savedGameData = getGameSession();
      if (!savedGameData) {
        console.log("No saved game data for reconnection, returning to landing page");
        clearGameSession();
        navigate("/", { replace: true });
        return;
      }

      const { gameId, playerId, playerName } = savedGameData;

      // Step 1: Reconnect to game
      setReconnectionStep("game");

      // Fetch current game state from server first (with playerId for personalized view)
      const response = await fetch(`${config.apiUrl}/games/${gameId}?playerId=${playerId}`);
      if (!response.ok) {
        // Game doesn't exist, automatically clear storage and redirect
        console.log(
          `Game not found (status: ${response.status}), clearing session and returning to landing page`,
        );
        clearGameSession();
        navigate("/", { replace: true });
        return;
      }

      const gameData = await response.json();

      // Update local state with fetched game data
      setGame(gameData.game);
      setPlayerId(playerId);

      // Set current player from fetched game data
      const player = gameData.game.currentPlayer;
      setCurrentPlayer(player || null);

      // Store player ID for WebSocket handlers
      currentPlayerIdRef.current = playerId;

      // Step 2: Ensure 3D environment is loaded
      if (!skyboxCache.isReady()) {
        setReconnectionStep("environment");
        await skyboxCache.preload();
      }

      // Now establish WebSocket connection (non-blocking)
      globalWebSocketManager.playerConnect(playerName, gameId, playerId);
    } catch (error) {
      console.error("❌ Reconnection failed:", error);
      setIsReconnecting(false);
      setReconnectionStep(null);
      // Don't navigate away - let user try manual reconnection
      // or they can manually navigate to home if needed
      console.error("Failed to reconnect to game. Please check your connection and try again.");
    }
  }, [navigate]);

  // Setup WebSocket listeners using global manager - only initialize once
  const setupWebSocketListeners = useCallback(() => {
    if (isWebSocketInitialized.current) {
      return () => {}; // Already initialized, return empty cleanup
    }

    globalWebSocketManager.on("game-updated", handleGameUpdated);
    globalWebSocketManager.on("full-state", handleFullState);
    globalWebSocketManager.on("log-update", handleLogUpdate);
    globalWebSocketManager.on("player-disconnected", handlePlayerDisconnected);
    globalWebSocketManager.on("player-kicked", handlePlayerKicked);
    globalWebSocketManager.on("game-ended", handleGameEnded);
    globalWebSocketManager.on("error", handleError);
    globalWebSocketManager.on("disconnect", handleDisconnect);
    globalWebSocketManager.on("max-reconnects-reached", handleMaxReconnectsReached);
    globalWebSocketManager.on("chat-update", handleChatUpdate);
    globalWebSocketManager.on("spectator-kicked", handleSpectatorKicked);
    globalWebSocketManager.on("spectator-connected", handleSpectatorIdReceived);

    isWebSocketInitialized.current = true;

    return () => {
      globalWebSocketManager.off("game-updated", handleGameUpdated);
      globalWebSocketManager.off("full-state", handleFullState);
      globalWebSocketManager.off("log-update", handleLogUpdate);
      globalWebSocketManager.off("player-disconnected", handlePlayerDisconnected);
      globalWebSocketManager.off("player-kicked", handlePlayerKicked);
      globalWebSocketManager.off("game-ended", handleGameEnded);
      globalWebSocketManager.off("error", handleError);
      globalWebSocketManager.off("disconnect", handleDisconnect);
      globalWebSocketManager.off("max-reconnects-reached", handleMaxReconnectsReached);
      globalWebSocketManager.off("chat-update", handleChatUpdate);
      globalWebSocketManager.off("spectator-kicked", handleSpectatorKicked);
      globalWebSocketManager.off("spectator-connected", handleSpectatorIdReceived);
      isWebSocketInitialized.current = false;
    };
  }, [
    handleGameUpdated,
    handleFullState,
    handleLogUpdate,
    handlePlayerDisconnected,
    handlePlayerKicked,
    handleGameEnded,
    handleError,
    handleDisconnect,
    handleMaxReconnectsReached,
    handleChatUpdate,
    handleSpectatorKicked,
    handleSpectatorIdReceived,
  ]);

  // Handle action selection from card actions
  const handleActionSelect = useCallback(
    (action: PlayerActionDto) => {
      // Block actions when tile selection is pending
      if (currentPlayer?.pendingTileSelection) {
        return;
      }

      // Check if this is an action-reuse action (e.g., Viron)
      const isActionReuse = action.behavior.outputs?.some(
        (o: { type: string }) => o.type === "action-reuse",
      );
      if (isActionReuse) {
        setPendingActionReuse({
          cardId: action.cardId,
          behaviorIndex: action.behaviorIndex,
        });
        setShowActionReuseSelection(true);
        return;
      }

      // Check if this action has choices
      if (action.behavior.choices && action.behavior.choices.length > 0) {
        // Action has choices, show the choice selection popover
        setActionPendingChoice(action);
        setShowActionChoiceSelection(true);
      } else {
        // No choices, check if behavior has variableAmount inputs/outputs
        const variableInfo = getVariableAmountInfo(action.behavior.inputs, action.behavior.outputs);
        if (variableInfo) {
          setPendingVariableAmount({
            type: "card-action",
            cardId: action.cardId,
            cardName: action.cardName,
            behaviorIndex: action.behaviorIndex,
            resourceLabel: variableInfo.resourceLabel,
            maxAmount: variableInfo.maxAmount,
          });
          setShowAmountSelection(true);
          return;
        }

        // Check if inputs need card resource selection (steal-from-any-card)
        const cardResourceInfo = needsCardResourceInput(action.behavior.inputs);

        if (cardResourceInfo) {
          // Determine cardStorageTargets for the output side
          const storageInfo = needsCardStorageSelection(action.behavior.outputs);
          const cardStorageTargets =
            storageInfo?.target === "self-card" ? [action.cardId] : undefined;

          setPendingCardResourceInput({
            cardId: action.cardId,
            behaviorIndex: action.behaviorIndex,
            cardStorageTargets,
            resourceType: cardResourceInfo.resourceType,
            amount: cardResourceInfo.amount,
          });
          setShowCardResourceSelection(true);
        } else {
          // Check for self-card storage first
          const storageInfo = needsCardStorageSelection(action.behavior.outputs);

          if (storageInfo && storageInfo.target === "self-card") {
            // Self-card target: skip popover and send directly with the source card as target
            const targetInfo = needsTargetPlayerSelection(action.behavior.outputs);
            if (targetInfo) {
              setPendingActionTargetPlayer({
                cardId: action.cardId,
                behaviorIndex: action.behaviorIndex,
                cardStorageTargets: [action.cardId],
                resourceType: targetInfo.resourceType,
                amount: targetInfo.amount,
                isSteal: targetInfo.isSteal,
              });
              setShowActionTargetPlayerSelection(true);
            } else {
              void globalWebSocketManager.playCardAction(
                action.cardId,
                action.behaviorIndex,
                undefined,
                [action.cardId],
              );
            }
          } else {
            // Collect ALL any-card storage needs
            const allStorageNeeds: Array<{
              resourceType: ResourceType;
              amount: number;
              selectorTags?: string[];
            }> = [];
            const selections = getAllAnyCardStorageSelections(action.behavior.outputs);
            for (const sel of selections) {
              allStorageNeeds.push({
                resourceType: sel.resourceType,
                amount: sel.amount,
                selectorTags: sel.selectorTags,
              });
            }

            if (allStorageNeeds.length > 0) {
              const first = allStorageNeeds[0];
              setPendingActionStorage({
                cardId: action.cardId,
                behaviorIndex: action.behaviorIndex,
                allStorageNeeds,
                collectedTargets: [],
                currentIndex: 0,
                resourceType: first.resourceType,
                amount: first.amount,
                selectorTags: first.selectorTags,
              });
              setShowActionStorageSelection(true);
            } else {
              // No card storage needed, check target player needs
              const targetInfo = needsTargetPlayerSelection(action.behavior.outputs);
              if (targetInfo) {
                setPendingActionTargetPlayer({
                  cardId: action.cardId,
                  behaviorIndex: action.behaviorIndex,
                  resourceType: targetInfo.resourceType,
                  amount: targetInfo.amount,
                  isSteal: targetInfo.isSteal,
                });
                setShowActionTargetPlayerSelection(true);
              } else {
                void globalWebSocketManager.playCardAction(action.cardId, action.behaviorIndex);
              }
            }
          }
        }
      }
    },
    [
      currentPlayer?.pendingTileSelection,
      needsCardStorageSelection,
      getAllAnyCardStorageSelections,
      needsTargetPlayerSelection,
      needsCardResourceInput,
      getVariableAmountInfo,
    ],
  );

  // Standard project selection handler
  const handleStandardProjectSelect = useCallback(
    (project: StandardProject) => {
      // Block standard projects when tile selection is pending
      if (currentPlayer?.pendingTileSelection) {
        return;
      }

      // All standard projects execute immediately
      // Backend will create tile queue for projects requiring placement
      switch (project) {
        case StandardProject.SELL_PATENTS:
          // Initiate sell patents - backend will create pendingCardSelection
          void globalWebSocketManager.sellPatents();
          break;
        case StandardProject.POWER_PLANT:
          void globalWebSocketManager.buildPowerPlant();
          break;
        case StandardProject.ASTEROID:
          void globalWebSocketManager.launchAsteroid();
          break;
        case StandardProject.AQUIFER:
          void globalWebSocketManager.buildAquifer();
          break;
        case StandardProject.GREENERY:
          void globalWebSocketManager.plantGreenery();
          break;
        case StandardProject.CITY:
          void globalWebSocketManager.buildCity();
          break;
      }
    },
    [currentPlayer?.pendingTileSelection],
  );

  // Resource conversion handlers
  const handleConvertPlantsToGreenery = useCallback(() => {
    // Block if tile selection is already pending
    if (currentPlayer?.pendingTileSelection) {
      return;
    }

    // Initiate plant conversion (backend creates pending tile selection)
    void globalWebSocketManager.convertPlantsToGreenery();
  }, [currentPlayer?.pendingTileSelection]);

  const handleConvertHeatToTemperature = useCallback(() => {
    // Convert heat to temperature directly (no tile selection needed)
    void globalWebSocketManager.convertHeatToTemperature();
  }, []);

  // Leave game handler - shows confirmation dialog
  const handleLeaveGame = useCallback(() => {
    setShowLeaveGameConfirm(true);
  }, []);

  // Confirm leave game - clears session and disconnects
  const handleConfirmLeaveGame = useCallback(() => {
    clearGameSession();
    globalWebSocketManager.disconnect();
    // Small delay ensures WebSocket close is processed before navigation
    setTimeout(() => {
      navigate("/", { replace: true });
    }, 100);
  }, [navigate]);

  // End game handler - shows confirmation dialog (host only)
  const handleEndGame = useCallback(() => {
    setShowEndGameConfirm(true);
  }, []);

  // Confirm end game - sends end-game message to server
  const handleConfirmEndGame = useCallback(() => {
    setShowEndGameConfirm(false);
    void globalWebSocketManager.endGame();
  }, []);

  // Tab conflict handlers
  const handleTabTakeOver = () => {
    if (conflictingTabInfo) {
      const tabManager = getTabManager();
      tabManager.forceTakeOver(conflictingTabInfo.gameId, conflictingTabInfo.playerName);
      setShowTabConflict(false);
      setConflictingTabInfo(null);

      // Now initialize the game with the route state
      const routeState = location.state as {
        game?: GameDto;
        playerId?: string;
        playerName?: string;
      } | null;

      if (routeState?.game && routeState?.playerId && routeState?.playerName) {
        setGame(routeState.game);
        setIsConnected(true);

        // Store game data for reconnection
        saveGameSession({
          gameId: routeState.game.id,
          playerId: routeState.playerId,
          playerName: routeState.playerName,
          timestamp: Date.now(),
        });

        // Set current player from game data
        const player = routeState.game.currentPlayer;
        setCurrentPlayer(player || null);

        // Store player ID for WebSocket handlers and component state
        currentPlayerIdRef.current = routeState.playerId;
        setPlayerId(routeState.playerId);
      }
    }
  };

  const handleTabCancel = () => {
    setShowTabConflict(false);
    setConflictingTabInfo(null);
    // Return to main menu
    navigate("/", { replace: true });
  };

  const handlePlayerSelected = useCallback(
    async (selectedPlayerId: string, playerName: string) => {
      if (!gameForSelection) return;

      setLoadingPhase("connecting");

      const tabManager = getTabManager();
      const canClaim = await tabManager.claimTab(gameForSelection.id, playerName);

      if (!canClaim) {
        const activeTabInfo = tabManager.getActiveTabInfo();
        if (activeTabInfo) {
          setConflictingTabInfo(activeTabInfo);
          setShowTabConflict(true);
          return;
        }
      }

      saveGameSession({
        gameId: gameForSelection.id,
        playerId: selectedPlayerId,
        playerName: playerName,
        timestamp: Date.now(),
      });

      currentPlayerIdRef.current = selectedPlayerId;
      setPlayerId(selectedPlayerId);
      globalWebSocketManager.setCurrentPlayerId(selectedPlayerId);

      await globalWebSocketManager.playerTakeover(selectedPlayerId, gameForSelection.id);

      setLoadingPhase("ready");
    },
    [gameForSelection],
  );

  const handlePlayerSelectionCancel = useCallback(() => {
    navigate("/", { replace: true });
  }, [navigate]);

  const handleSpectatorConnected = useCallback(() => {
    setIsSpectator(true);
    setLoadingPhase("ready");
    setIsConnected(true);
  }, []);

  useEffect(() => {
    let aborted = false;

    const initializeGame = async () => {
      setLoadingPhase("checking");

      const routeState = location.state as {
        game?: GameDto;
        playerId?: string;
        playerName?: string;
        isReconnection?: boolean;
        spectatorName?: string;
      } | null;

      // 1. Determine game ID and player ID sources (priority order)
      const savedSession = getGameSession();
      const gameId = urlGameId || routeState?.game?.id || savedSession?.gameId;

      if (!gameId) {
        navigate("/", { replace: true });
        return;
      }

      // 2. Validate game exists (and player is still in game if we have a saved session)
      let fetchedGame: GameDto | null = null;
      try {
        const playerIdForApi = savedSession?.isSpectator ? undefined : savedSession?.playerId;
        fetchedGame = await apiService.getGame(gameId, playerIdForApi);
      } catch {
        navigate("/", {
          replace: true,
          state: { error: "Could not find game" },
        });
        return;
      }

      if (aborted) return;

      if (!fetchedGame) {
        clearGameSession();
        navigate("/", {
          replace: true,
          state: { error: "Could not find game" },
        });
        return;
      }

      // 3. Check cached session
      const cachedForThisGame = savedSession && savedSession.gameId === gameId;

      if (cachedForThisGame && savedSession.isSpectator) {
        setIsSpectator(true);
        setIsConnected(true);
        setLoadingPhase("ready");
        await globalWebSocketManager.spectatorConnect(savedSession.playerName, gameId);
        return;
      }

      if (cachedForThisGame && savedSession.playerId) {
        // Check if cached player exists in game
        const allPlayers = [fetchedGame.currentPlayer, ...(fetchedGame.otherPlayers || [])].filter(
          Boolean,
        );

        const cachedPlayer = allPlayers.find((p) => p?.id === savedSession.playerId);

        if (cachedPlayer) {
          // Player exists in game
          if (!cachedPlayer.isConnected) {
            // Player is disconnected, auto-reconnect
            setLoadingPhase("connecting");

            const tabManager = getTabManager();
            const canClaim = await tabManager.claimTab(fetchedGame.id, savedSession.playerName);
            if (aborted) return;

            if (!canClaim) {
              const activeTabInfo = tabManager.getActiveTabInfo();
              if (activeTabInfo) {
                setConflictingTabInfo(activeTabInfo);
                setShowTabConflict(true);
                return;
              }
            }

            currentPlayerIdRef.current = savedSession.playerId;
            setPlayerId(savedSession.playerId);
            globalWebSocketManager.setCurrentPlayerId(savedSession.playerId);

            await globalWebSocketManager.playerTakeover(savedSession.playerId, fetchedGame.id);
            if (aborted) return;

            setLoadingPhase("ready");
            return;
          }
          // Player is connected from somewhere else - show selection
        }
      }

      // 4. If we have full route state with player info, use that directly
      if (routeState?.game && routeState?.playerId && routeState?.playerName) {
        const tabManager = getTabManager();
        const canClaim = await tabManager.claimTab(routeState.game.id, routeState.playerName);
        if (aborted) return;

        if (!canClaim) {
          const activeTabInfo = tabManager.getActiveTabInfo();
          if (activeTabInfo) {
            setConflictingTabInfo(activeTabInfo);
            setShowTabConflict(true);
            return;
          }
        }

        setGame(routeState.game);
        setIsConnected(true);

        saveGameSession({
          gameId: routeState.game.id,
          playerId: routeState.playerId,
          playerName: routeState.playerName,
          timestamp: Date.now(),
        });

        const player = routeState.game.currentPlayer;
        setCurrentPlayer(player || null);

        currentPlayerIdRef.current = routeState.playerId;
        setPlayerId(routeState.playerId);

        globalWebSocketManager.setCurrentPlayerId(routeState.playerId);

        void globalWebSocketManager.playerConnect(
          routeState.playerName,
          routeState.game.id,
          routeState.playerId,
        );

        setLoadingPhase("ready");
        return;
      }

      if (aborted) return;

      // 5. Check if spectator name was passed from browse page popover
      if (routeState?.spectatorName) {
        setIsSpectator(true);
        setIsConnected(true);
        setLoadingPhase("ready");
        saveGameSession({
          gameId,
          playerId: "",
          playerName: routeState.spectatorName,
          isSpectator: true,
        });
        await globalWebSocketManager.spectatorConnect(routeState.spectatorName, gameId);
        return;
      }

      // 6. Check if this is a join or spectate link
      const urlParams = new URLSearchParams(window.location.search);
      const linkType = urlParams.get("type");

      if (linkType === "spectate") {
        setGameForSelection(fetchedGame);
        setLoadingPhase("spectating");
        return;
      }

      if (linkType === "join") {
        setGameForSelection(fetchedGame);
        setLoadingPhase("joining");
        return;
      }

      // 6. Show player selection
      setGameForSelection(fetchedGame);
      setLoadingPhase("selecting");
    };

    void initializeGame();
    return () => {
      aborted = true;
    };
  }, [location.state, navigate, urlGameId]);

  // Register event listeners when component mounts, unregister on unmount
  useEffect(() => {
    // Store player ID in global manager for event handling
    if (currentPlayerIdRef.current) {
      globalWebSocketManager.setCurrentPlayerId(currentPlayerIdRef.current);
    }

    return setupWebSocketListeners();
  }, [setupWebSocketListeners]);

  // Listen for debug dropdown toggle from TopMenuBar
  useEffect(() => {
    const handleToggleDebug = () => {
      setShowDebugDropdown((prev) => !prev);
    };

    window.addEventListener("toggle-debug-dropdown", handleToggleDebug);
    return () => {
      window.removeEventListener("toggle-debug-dropdown", handleToggleDebug);
    };
  }, []);

  useEffect(() => {
    const handleTogglePerf = () => {
      setShowPerformanceWindow((prev) => !prev);
    };
    window.addEventListener("toggle-performance-window", handleTogglePerf);
    return () => {
      window.removeEventListener("toggle-performance-window", handleTogglePerf);
    };
  }, []);

  useEffect(() => {
    const handleToggleBugReport = () => {
      setShowBugReportWindow((prev) => !prev);
    };
    window.addEventListener("toggle-bug-report-window", handleToggleBugReport);
    return () => {
      window.removeEventListener("toggle-bug-report-window", handleToggleBugReport);
    };
  }, []);

  // Show/hide starting selection overlay based on backend state
  useEffect(() => {
    const corpPhase = game?.currentPlayer?.selectCorporationPhase;
    const hasStartingData =
      game?.currentPhase === GamePhaseStartingSelection &&
      game?.status === GameStatusActive &&
      corpPhase &&
      corpPhase.availableCorporations.length > 0;

    if (hasStartingData && !showStartingSelection) {
      setShowStartingSelection(true);
    } else if (showStartingSelection && !corpPhase) {
      setShowStartingSelection(false);
    }
  }, [
    game?.currentPhase,
    game?.status,
    game?.currentPlayer?.selectCorporationPhase,
    showStartingSelection,
  ]);

  // Show/hide pending card selection overlay (sell patents, etc.)
  useEffect(() => {
    const pendingSelection = game?.currentPlayer?.pendingCardSelection;

    if (pendingSelection && !showPendingCardSelection) {
      setShowPendingCardSelection(true);
    } else if (!pendingSelection && showPendingCardSelection) {
      setShowPendingCardSelection(false);
    }
  }, [game?.currentPlayer?.pendingCardSelection, showPendingCardSelection]);

  // Show/hide card draw selection overlay
  useEffect(() => {
    const pendingDrawSelection = game?.currentPlayer?.pendingCardDrawSelection;

    if (pendingDrawSelection && !showCardDrawSelection) {
      setShowCardDrawSelection(true);
    } else if (!pendingDrawSelection && showCardDrawSelection) {
      setShowCardDrawSelection(false);
    }
  }, [game?.currentPlayer?.pendingCardDrawSelection, showCardDrawSelection]);

  // Show/hide card discard selection overlay (passive effects like Mars University)
  useEffect(() => {
    const pendingDiscardSelection = game?.currentPlayer?.pendingCardDiscardSelection;

    if (pendingDiscardSelection && !showCardDiscardSelection) {
      setShowCardDiscardSelection(true);
    } else if (!pendingDiscardSelection && showCardDiscardSelection) {
      setShowCardDiscardSelection(false);
    }
  }, [game?.currentPlayer?.pendingCardDiscardSelection, showCardDiscardSelection]);

  // Show/hide behavior choice selection popover (passive triggered effects)
  useEffect(() => {
    const pendingBehaviorChoice = game?.currentPlayer?.pendingBehaviorChoiceSelection;

    if (pendingBehaviorChoice && !showBehaviorChoiceSelection) {
      setShowBehaviorChoiceSelection(true);
    } else if (!pendingBehaviorChoice && showBehaviorChoiceSelection) {
      setShowBehaviorChoiceSelection(false);
    }
  }, [game?.currentPlayer?.pendingBehaviorChoiceSelection, showBehaviorChoiceSelection]);

  // Show/hide steal target selection popover (deferred adjacent steal from tile placement)
  useEffect(() => {
    const pending = game?.currentPlayer?.pendingStealTargetSelection;

    if (pending && !showStealTargetSelection) {
      setShowStealTargetSelection(true);
    } else if (!pending && showStealTargetSelection) {
      setShowStealTargetSelection(false);
    }
  }, [game?.currentPlayer?.pendingStealTargetSelection, showStealTargetSelection]);

  // Show/hide colony resource selection popover (card storage for colony trade/build rewards)
  useEffect(() => {
    const pending = game?.currentPlayer?.pendingColonyResourceSelection;

    if (pending && !showColonyResourceSelection) {
      setShowColonyResourceSelection(true);
    } else if (!pending && showColonyResourceSelection) {
      setShowColonyResourceSelection(false);
    }
  }, [game?.currentPlayer?.pendingColonyResourceSelection, showColonyResourceSelection]);

  // Demo keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.ctrlKey || event.metaKey) {
        switch (event.key) {
          case "1":
            event.preventDefault();
            setShowCardsPlayedModal(true);
            break;
          case "4":
            event.preventDefault();
            // Actions are now handled via popover in BottomResourceBar
            break;
          case "5":
            event.preventDefault();
            break;
        }
      }
    };

    window.addEventListener("keydown", handleKeyPress);
    return () => window.removeEventListener("keydown", handleKeyPress);
  }, []);

  const loadingMessage = (() => {
    if (
      loadingPhase === "selecting" ||
      loadingPhase === "joining" ||
      loadingPhase === "spectating"
    ) {
      if (!isSpaceBgLoaded) return "Loading 3D environment...";
      return "Loading game...";
    }
    if (loadingPhase === "checking") return "Loading game...";
    if (loadingPhase === "connecting") return "Connecting...";
    if (isReconnecting && reconnectionStep) {
      if (reconnectionStep === "game") return "Reconnecting to game...";
      if (reconnectionStep === "environment") return "Loading 3D environment...";
    }
    if (!isSkyboxReady) return "Loading 3D environment...";
    return "Connecting to game...";
  })();

  // Check if game is in lobby phase
  // Spectate: derive player from live game state so data stays fresh via WebSocket
  const spectatePlayer = useMemo(() => {
    if (!spectatePlayerId || !game) return null;
    if (game.currentPlayer?.id === spectatePlayerId) return game.currentPlayer;
    return game.otherPlayers?.find((p) => p.id === spectatePlayerId) ?? null;
  }, [spectatePlayerId, game]);

  const spectatePlayerColor = useMemo(() => {
    if (!spectatePlayerId || !game) return "#6496ff";
    if (game.currentPlayer?.id === spectatePlayerId) return game.currentPlayer.color || "#6496ff";
    const other = game.otherPlayers?.find((p) => p.id === spectatePlayerId);
    return other?.color || "#6496ff";
  }, [spectatePlayerId, game]);

  const playerColorMap = useMemo(() => {
    if (!game) return new Map<string, string>();
    const map = new Map<string, string>();
    if (game.currentPlayer?.id && game.currentPlayer.color) {
      map.set(game.currentPlayer.id, game.currentPlayer.color);
    }
    game.otherPlayers?.forEach((p) => {
      if (p.id && p.color) map.set(p.id, p.color);
    });
    game.spectators?.forEach((s) => {
      if (s.id && s.color) map.set(s.id, s.color);
    });
    return map;
  }, [game]);

  const handlePlayerClick = useCallback(
    (player: PlayerDto | OtherPlayerDto) => {
      const phase = game?.currentPhase;
      if (phase === GamePhaseInitApplyCorp || phase === GamePhaseInitApplyPrelude) return;
      if (player.id === game?.currentPlayer?.id) {
        setSpectatePlayerId(null);
        return;
      }
      setSpectatePlayerId((prev) => (prev === player.id ? null : player.id));
    },
    [game?.currentPlayer?.id, game?.currentPhase],
  );

  const handleStopSpectating = useCallback(() => {
    setSpectatePlayerId(null);
  }, []);

  // ESC key to stop spectating
  useEffect(() => {
    if (!spectatePlayerId) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setSpectatePlayerId(null);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [spectatePlayerId]);

  const isLobbyPhase = game?.status === GameStatusLobby;

  // Pre-game phase covers lobby and starting selection
  const isPreGamePhase =
    isLobbyPhase ||
    (game?.status === GameStatusActive && game?.currentPhase === GamePhaseStartingSelection);

  const isInitApplyPhase =
    game?.status === GameStatusActive &&
    (game?.currentPhase === GamePhaseInitApplyCorp ||
      game?.currentPhase === GamePhaseInitApplyPrelude);

  // Show waiting modal when player has finished selection but others haven't
  const showWaitingForPlayers =
    game?.status === GameStatusActive &&
    game?.currentPhase === GamePhaseStartingSelection &&
    !game?.currentPlayer?.selectCorporationPhase &&
    !game?.currentPlayer?.selectPreludeCardsPhase &&
    !game?.currentPlayer?.selectStartingCardsPhase &&
    !game?.currentPlayer?.pendingTileSelection;

  const [lobbyMounted, setLobbyMounted] = useState(false);

  useEffect(() => {
    if (isLobbyPhase) setLobbyMounted(true);
  }, [isLobbyPhase]);

  // Transition state machine: lobby → loading → fadeOutLobby → animateUI → complete
  useEffect(() => {
    if (isPreGamePhase) {
      setTransitionPhase("lobby");
      wasInLobby.current = true;
      return;
    }

    if (!wasInLobby.current || transitionPhase !== "lobby") return;

    setTransitionPhase("loading");
    setOverlayVisible(true);
    audioService.stopAmbientWithDuration(2000);
  }, [isPreGamePhase]);

  useEffect(() => {
    if (transitionPhase === "fadeOutLobby") {
      const timer = setTimeout(() => setTransitionPhase("animateUI"), 1500);
      return () => clearTimeout(timer);
    }
    if (transitionPhase === "animateUI") {
      const timer = setTimeout(() => setTransitionPhase("complete"), 2500);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [transitionPhase]);

  // On mount: if game is already active (reload/reconnect), stop ambient immediately
  useEffect(() => {
    if (game && !isPreGamePhase && !wasInLobby.current) {
      audioService.stopAmbient();
    }
  }, [game, isPreGamePhase]);

  // Auto-advance init phase after transition animation and notification queue complete
  // Host always sends confirms for all players (bots handle tile placements only)
  useEffect(() => {
    if (!isInitApplyPhase || !game?.initPhase?.waitingForConfirm) return;
    if (game.hostPlayerId !== currentPlayer?.id) return;

    const animationDone = transitionPhase === "complete" || transitionPhase === "idle";
    if (!animationDone) return;

    if (game.initPhase.hasPendingTiles) return;

    // Wait for any notification queue to finish before confirming
    const now = Date.now();
    const queueRemaining = Math.max(0, notificationQueueDoneAt.current - now);
    const delay = queueRemaining + 750;

    const timer = setTimeout(() => {
      void globalWebSocketManager.confirmInitAdvance();
    }, delay);

    return () => clearTimeout(timer);
  }, [
    isInitApplyPhase,
    game?.initPhase?.waitingForConfirm,
    game?.initPhase?.confirmVersion,
    game?.initPhase?.currentPlayerIndex,
    game?.initPhase?.hasPendingTiles,
    game?.hostPlayerId,
    currentPlayer?.id,
    transitionPhase,
  ]);

  const bottomBarCallbacks = useMemo(
    () => ({
      onOpenCardsPlayedModal: () => setShowCardsPlayedModal(true),
      onActionSelect: handleActionSelect,
      onConvertPlantsToGreenery: handleConvertPlantsToGreenery,
      onConvertHeatToTemperature: handleConvertHeatToTemperature,
    }),
    [handleActionSelect, handleConvertPlantsToGreenery, handleConvertHeatToTemperature],
  );

  // Hotkeys: Space (toggle hand fan), Enter (pass/skip turn)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable) {
        return;
      }

      const anyModalOpen =
        showStartingSelection ||
        showPendingCardSelection ||
        showCardDrawSelection ||
        showCardDiscardSelection ||
        showBehaviorChoiceSelection ||
        showStealTargetSelection ||
        showColonyResourceSelection ||
        showProductionPhaseModal ||
        showPaymentSelection ||
        isPreGamePhase ||
        isInitApplyPhase;

      if (e.key === " ") {
        e.preventDefault();
        if (!anyModalOpen && !spectatePlayerId && !isSpectator) {
          cardFanRef.current?.toggleExpand();
        }
      }

      if (e.key === "Enter") {
        e.preventDefault();
        if (
          !anyModalOpen &&
          game?.currentPhase === GamePhaseAction &&
          currentPlayer?.id === game?.currentTurn &&
          !currentPlayer?.pendingTileSelection
        ) {
          void globalWebSocketManager.skipAction();
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [
    spectatePlayerId,
    isSpectator,
    showStartingSelection,
    showPendingCardSelection,
    showCardDrawSelection,
    showCardDiscardSelection,
    showBehaviorChoiceSelection,
    showProductionPhaseModal,
    showPaymentSelection,
    isPreGamePhase,
    isInitApplyPhase,
    game?.currentPhase,
    game?.currentTurn,
    currentPlayer?.id,
    currentPlayer?.pendingTileSelection,
  ]);

  // Check if we need the persistent backdrop (during overlay transitions)
  const shouldShowBackdrop = showStartingSelection;

  return (
    <>
      {/* Dev Mode Chip - Always visible in dev mode */}
      {game?.settings?.developmentMode && <DevModeChip />}

      {/* Persistent backdrop for card selection / waiting overlay */}
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
            gameState={game}
            currentPlayer={currentPlayer}
            playedCards={currentPlayer?.playedCards || []}
            corporationCard={corporationData}
            showCorporation={showCorp}
            initTurnPlayerId={displayedInitPlayerId}
            showStartingSelection={showStartingSelection}
            transitionPhase={transitionPhase}
            animateHexEntrance={
              transitionPhase === "fadeOutLobby" ||
              transitionPhase === "animateUI" ||
              transitionPhase === "complete"
            }
            changedPaths={changedPaths}
            tileHighlightMode={tileHighlightMode}
            vpIndicators={vpIndicators}
            triggeredEffects={triggeredEffects}
            bottomBarCallbacks={bottomBarCallbacks}
            onStandardProjectSelect={handleStandardProjectSelect}
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
          />
        )}

      <CardsPlayedModal
        isVisible={showCardsPlayedModal}
        onClose={() => setShowCardsPlayedModal(false)}
        cards={(spectatePlayer?.playedCards ?? currentPlayer?.playedCards) || []}
      />

      <ProductionPhaseModal
        isOpen={showProductionPhaseModal && !isProductionModalHidden}
        gameState={game}
        onClose={() => {
          setShowProductionPhaseModal(false);
          setIsProductionModalHidden(false);
          setOpenProductionToCardSelection(false);
        }}
        onHide={() => {
          setIsProductionModalHidden(true);
          setOpenProductionToCardSelection(false);
        }}
        openDirectlyToCardSelection={openProductionToCardSelection}
      />

      <WindowManagerProvider>
        <DebugDropdown
          isVisible={showDebugDropdown}
          onClose={() => setShowDebugDropdown(false)}
          gameState={game}
          changedPaths={changedPaths}
        />

        <PerformanceWindow
          isVisible={showPerformanceWindow}
          onClose={() => setShowPerformanceWindow(false)}
        />

        <BugReportWindow
          isVisible={showBugReportWindow}
          onClose={() => setShowBugReportWindow(false)}
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
            onExited={() => setLobbyMounted(false)}
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

      {/* Demo setup overlay - shown after start game in demo mode */}
      {game?.currentPhase === GamePhaseDemoSetup && game && playerId && (
        <DemoSetupOverlay game={game} playerId={playerId} />
      )}

      {showTabConflict && conflictingTabInfo && (
        <TabConflictOverlay
          activeGameInfo={conflictingTabInfo}
          onTakeOver={handleTabTakeOver}
          onCancel={handleTabCancel}
        />
      )}

      {loadingPhase === "selecting" && gameForSelection && (
        <PlayerSelectionOverlay
          game={gameForSelection}
          onSelectPlayer={(playerId, playerName) => void handlePlayerSelected(playerId, playerName)}
          onSpectate={handleSpectatorConnected}
          onCancel={handlePlayerSelectionCancel}
        />
      )}

      {loadingPhase === "joining" && gameForSelection && (
        <JoinGameOverlay game={gameForSelection} onCancel={handlePlayerSelectionCancel} />
      )}

      {loadingPhase === "spectating" && gameForSelection && (
        <SpectateGameOverlay
          game={gameForSelection}
          onCancel={handlePlayerSelectionCancel}
          onConnected={handleSpectatorConnected}
        />
      )}

      {/* Leave game confirmation dialog */}
      {showLeaveGameConfirm && (
        <GameMenuModal
          title="Leave game?"
          showBackdrop={true}
          onClose={() => setShowLeaveGameConfirm(false)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            You can reconnect to the game again without losing any progress.
          </p>
          <div className="flex gap-4 justify-center">
            <GameMenuButton variant="secondary" onClick={() => setShowLeaveGameConfirm(false)}>
              Cancel
            </GameMenuButton>
            <GameMenuButton variant="error" onClick={handleConfirmLeaveGame}>
              Leave
            </GameMenuButton>
          </div>
        </GameMenuModal>
      )}

      {/* End game confirmation dialog (host only) */}
      {showEndGameConfirm && (
        <GameMenuModal
          title="End game?"
          showBackdrop={true}
          onClose={() => setShowEndGameConfirm(false)}
          zIndex={10000}
        >
          <p className="text-white/80 text-center mb-6">
            This will end the game for all players. This action cannot be undone.
          </p>
          <div className="flex gap-4 justify-center">
            <GameMenuButton variant="secondary" onClick={() => setShowEndGameConfirm(false)}>
              Cancel
            </GameMenuButton>
            <GameMenuButton variant="error" onClick={handleConfirmEndGame}>
              End game
            </GameMenuButton>
          </div>
        </GameMenuModal>
      )}

      {/* Starting selection overlay (corporation + preludes + project cards) */}
      <StartingCardSelectionOverlay
        isOpen={showStartingSelection}
        availableCorporations={
          game?.currentPlayer?.selectCorporationPhase?.availableCorporations || []
        }
        availablePreludes={game?.currentPlayer?.selectPreludeCardsPhase?.availablePreludes || []}
        maxSelectablePreludes={game?.currentPlayer?.selectPreludeCardsPhase?.maxSelectable || 2}
        cards={game?.currentPlayer?.selectStartingCardsPhase?.availableCards || []}
        playerCredits={currentPlayer?.resources?.credits || 40}
        onConfirm={handleStartingChoicesConfirm}
      />

      {/* Waiting for other players to finish card selection */}
      {showWaitingForPlayers && game && (
        <>
          <MainMenuSettingsButton />
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

      {/* Pending card selection overlay (sell patents, etc.) */}
      {game?.currentPlayer?.pendingCardSelection && (
        <PendingCardSelectionOverlay
          isOpen={showPendingCardSelection}
          selection={game.currentPlayer.pendingCardSelection}
          playerCredits={currentPlayer?.resources?.credits || 0}
          onSelectCards={handlePendingCardSelection}
        />
      )}

      {/* Card draw selection overlay (card-draw/peek/take/buy effects) */}
      {game?.currentPlayer?.pendingCardDrawSelection && (
        <CardDrawSelectionOverlay
          isOpen={showCardDrawSelection}
          selection={game.currentPlayer.pendingCardDrawSelection}
          playerCredits={currentPlayer?.resources?.credits || 0}
          onConfirm={handleCardDrawConfirm}
        />
      )}

      {/* Card discard selection overlay (passive effects like Mars University) */}
      {game?.currentPlayer?.pendingCardDiscardSelection && currentPlayer && (
        <CardDiscardSelectionOverlay
          isOpen={showCardDiscardSelection}
          selection={game.currentPlayer.pendingCardDiscardSelection}
          handCards={currentPlayer.cards || []}
          onConfirm={handleCardDiscardConfirm}
        />
      )}

      {/* Card fan overlay for hand cards (animated out when spectating another player) */}
      {game && currentPlayer && (
        <div
          className={`transition-all duration-300 ease-in-out ${
            spectatePlayerId
              ? "opacity-0 pointer-events-none"
              : transitionPhase === "animateUI"
                ? "animate-[uiFadeIn_1200ms_ease-out_both]"
                : transitionPhase === "loading" || transitionPhase === "fadeOutLobby"
                  ? "opacity-0"
                  : ""
          }`}
        >
          <CardFanOverlay
            ref={cardFanRef}
            cards={currentPlayer.cards || []}
            hideWhenModalOpen={
              showStartingSelection ||
              showPendingCardSelection ||
              showCardDrawSelection ||
              showCardDiscardSelection ||
              showBehaviorChoiceSelection ||
              isPreGamePhase
            }
            onCardSelect={(_cardId) => {
              // TODO: Implement card selection logic (view details, etc.)
            }}
            onPlayCard={handlePlayCard}
          />
        </div>
      )}

      {/* End game overlay - shown when game is completed */}
      {game &&
        playerId &&
        game.currentPhase === GamePhaseComplete &&
        game.status === GameStatusCompleted &&
        game.finalScores &&
        game.finalScores.length > 0 && (
          <EndGameOverlay
            game={game}
            playerId={playerId}
            onTileHighlight={setTileHighlightMode}
            onVPIndicators={setVPIndicators}
            onReturnToMenu={() => {
              clearGameSession();
              navigate("/");
            }}
          />
        )}

      {/* Choice selection popover for card play */}
      {cardPendingChoice && (
        <ChoiceSelectionPopover
          cardId={cardPendingChoice.id}
          cardName={cardPendingChoice.name}
          behaviors={cardPendingChoice.behaviors || []}
          behaviorIndex={pendingCardBehaviorIndex}
          onChoiceSelect={handleChoiceSelect}
          onCancel={handleChoiceCancel}
          isVisible={showChoiceSelection}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {/* Choice selection popover for actions */}
      {actionPendingChoice && (
        <ChoiceSelectionPopover
          cardId={actionPendingChoice.cardId}
          cardName={actionPendingChoice.cardName}
          behaviors={[actionPendingChoice.behavior]}
          behaviorIndex={0}
          onChoiceSelect={handleActionChoiceSelect}
          onCancel={handleActionChoiceCancel}
          isVisible={showActionChoiceSelection}
          isAction={true}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {/* Action reuse popover (Viron) */}
      {pendingActionReuse && currentPlayer?.actions && (
        <ActionReusePopover
          isVisible={showActionReuseSelection}
          onClose={handleActionReuseCancel}
          actions={currentPlayer.actions}
          reuseSourceCardId={pendingActionReuse.cardId}
          onActionSelect={handleActionReuseSelect}
          gameState={game ?? undefined}
        />
      )}

      {/* Passive triggered behavior choice popover */}
      {game?.currentPlayer?.pendingBehaviorChoiceSelection && (
        <ChoiceSelectionPopover
          cardId={game.currentPlayer.pendingBehaviorChoiceSelection.sourceCardId}
          cardName={`Triggered: ${game.currentPlayer.pendingBehaviorChoiceSelection.source}`}
          behaviors={[{ choices: game.currentPlayer.pendingBehaviorChoiceSelection.choices }]}
          behaviorIndex={0}
          onChoiceSelect={handleBehaviorChoiceSelect}
          onCancel={() => {}}
          isVisible={showBehaviorChoiceSelection}
          playerResources={currentPlayer?.resources}
          resourceStorage={currentPlayer?.resourceStorage}
        />
      )}

      {/* Behavior choice storage selection popover */}
      {pendingBehaviorChoiceStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingBehaviorChoiceStorage.resourceType}
          amount={pendingBehaviorChoiceStorage.amount}
          selectorTags={pendingBehaviorChoiceStorage.selectorTags}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={currentPlayer?.corporation}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={handleBehaviorChoiceStorageSelect}
          onCancel={handleBehaviorChoiceStorageCancel}
          isVisible={showBehaviorChoiceStorage}
        />
      )}

      {/* Card storage selection popover */}
      {pendingCardStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingCardStorage.resourceType}
          amount={pendingCardStorage.amount}
          selectorTags={pendingCardStorage.selectorTags}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={currentPlayer?.corporation}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={handleCardStorageSelect}
          onCancel={handleCardStorageCancel}
          isVisible={showCardStorageSelection}
        />
      )}

      {/* Payment selection popover */}
      {pendingCardPayment && game && currentPlayer && (
        <PaymentSelectionPopover
          cardId={pendingCardPayment.card.id}
          card={pendingCardPayment.card}
          playerResources={currentPlayer.resources}
          paymentConstants={game.paymentConstants}
          playerPaymentSubstitutes={currentPlayer.paymentSubstitutes}
          storagePaymentSubstitutes={currentPlayer.storagePaymentSubstitutes}
          resourceStorage={currentPlayer.resourceStorage}
          onConfirm={handlePaymentConfirm}
          onCancel={handlePaymentCancel}
          isVisible={showPaymentSelection}
        />
      )}

      {/* Action storage selection popover */}
      {pendingActionStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingActionStorage.resourceType}
          amount={pendingActionStorage.amount}
          selectorTags={pendingActionStorage.selectorTags}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={currentPlayer?.corporation}
          resourceStorage={currentPlayer?.resourceStorage}
          onCardSelect={handleActionStorageSelect}
          onCancel={handleActionStorageCancel}
          isVisible={showActionStorageSelection}
        />
      )}

      {/* Target player selection popover (card play flow) */}
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
          onPlayerSelect={handleTargetPlayerSelect}
          onCancel={handleTargetPlayerCancel}
          isVisible={showTargetPlayerSelection}
        />
      )}

      {/* Target player selection popover (action flow) */}
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
          onPlayerSelect={handleActionTargetPlayerSelect}
          onCancel={handleActionTargetPlayerCancel}
          isVisible={showActionTargetPlayerSelection}
        />
      )}

      {/* Steal target selection popover (deferred adjacent steal from tile placement) */}
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
          onPlayerSelect={handleStealTargetSelect}
          onCancel={handleStealTargetSkip}
          isVisible={showStealTargetSelection}
          mandatory
        />
      )}

      {/* Colony resource selection popover (card storage for colony trade/build rewards) */}
      {game?.currentPlayer?.pendingColonyResourceSelection && currentPlayer && (
        <CardStorageSelectionPopover
          resourceType={
            game.currentPlayer.pendingColonyResourceSelection.resourceType as ResourceType
          }
          amount={game.currentPlayer.pendingColonyResourceSelection.amount}
          playedCards={currentPlayer.playedCards || []}
          corporationCard={currentPlayer.corporation}
          resourceStorage={currentPlayer.resourceStorage}
          onCardSelect={handleColonyResourceSelect}
          onCancel={handleColonyResourceSkip}
          isVisible={showColonyResourceSelection}
        />
      )}

      {/* Card resource selection popover (steal-from-any-card like Predators/Ants) */}
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
          onCardSelect={handleCardResourceSelect}
          onCancel={handleCardResourceCancel}
          isVisible={showCardResourceSelection}
        />
      )}

      {/* Amount selection popover (for variableAmount cards like Insulation, Power Infrastructure) */}
      {pendingVariableAmount && (
        <AmountSelectionPopover
          cardName={pendingVariableAmount.cardName}
          resourceLabel={pendingVariableAmount.resourceLabel}
          maxAmount={pendingVariableAmount.maxAmount}
          onAmountSelect={handleAmountSelect}
          onCancel={handleAmountCancel}
          isVisible={showAmountSelection}
        />
      )}

      {showProductionPhaseModal && isProductionModalHidden && (
        <GameMenuButton
          variant="primary"
          className="fixed top-[80px] left-[70%] !py-3.5 !px-7 !text-base !border-space-blue-400 text-shadow-glow shadow-[0_4px_15px_rgba(0,0,0,0.5),0_0_20px_rgba(30,60,150,0.4)] z-[1000] whitespace-nowrap hover:!border-space-blue-500 hover:shadow-[0_6px_20px_rgba(0,0,0,0.6),0_0_35px_rgba(30,60,150,0.6)] active:shadow-[0_2px_10px_rgba(0,0,0,0.4),0_0_20px_rgba(30,60,150,0.4)]"
          onClick={() => {
            setIsProductionModalHidden(false);
            setOpenProductionToCardSelection(true);
          }}
        >
          Return to Production
        </GameMenuButton>
      )}

      <GameEventBanner event={currentEvent} onDismiss={dismissGameEvent} />

      {/* Loading overlay rendered LAST to ensure it covers all UI elements */}
      {overlayVisible && (
        <LoadingOverlay
          isLoaded={isFullyLoaded}
          message={loadingMessage}
          onTransitionEnd={handleLoadingTransitionEnd}
        />
      )}
    </>
  );
}
