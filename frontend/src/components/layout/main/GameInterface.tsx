import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import GameLayout from "./GameLayout.tsx";
import CardsPlayedModal from "../../ui/modals/CardsPlayedModal.tsx";
import EffectsModal from "../../ui/modals/EffectsModal.tsx";
import ActionsModal from "../../ui/modals/ActionsModal.tsx";
import ProductionPhaseModal from "../../ui/modals/ProductionPhaseModal.tsx";
import PaymentSelectionPopover from "../../ui/popover/PaymentSelectionPopover.tsx";
import DebugDropdown from "../../ui/debug/DebugDropdown.tsx";
import DevModeChip from "../../ui/debug/DevModeChip.tsx";
import PerformanceWindow from "../../ui/debug/PerformanceWindow.tsx";
import TilePlacerWindow from "../../ui/debug/TilePlacerWindow.tsx";
import { WindowManagerProvider } from "../../ui/debug/WindowManager.tsx";
import WaitingRoomOverlay from "../../ui/overlay/WaitingRoomOverlay.tsx";
import PlayerSelectionOverlay from "../../ui/overlay/PlayerSelectionOverlay.tsx";
import JoinGameOverlay from "../../ui/overlay/JoinGameOverlay.tsx";
import DemoSetupOverlay from "../../ui/overlay/DemoSetupOverlay.tsx";
import TabConflictOverlay from "../../ui/overlay/TabConflictOverlay.tsx";
import StartingCardSelectionOverlay from "../../ui/overlay/StartingCardSelectionOverlay.tsx";
import PendingCardSelectionOverlay from "../../ui/overlay/PendingCardSelectionOverlay.tsx";
import CardDrawSelectionOverlay from "../../ui/overlay/CardDrawSelectionOverlay.tsx";
import CardFanOverlay from "../../ui/overlay/CardFanOverlay.tsx";
import LoadingOverlay from "../../game/view/LoadingOverlay.tsx";
import MainMenuSettingsButton from "../../ui/buttons/MainMenuSettingsButton.tsx";
import GameMenuButton from "../../ui/buttons/GameMenuButton.tsx";
import GameMenuModal from "../../ui/overlay/GameMenuModal.tsx";
import SpaceBackground from "../../3d/SpaceBackground.tsx";
import EndGameOverlay, { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import { TileHighlightMode } from "../../game/board/Tile.tsx";
import ChoiceSelectionPopover from "../../ui/popover/ChoiceSelectionPopover.tsx";
import CardStorageSelectionPopover from "../../ui/popover/CardStorageSelectionPopover.tsx";
import TargetPlayerSelectionPopover from "../../ui/popover/TargetPlayerSelectionPopover.tsx";
import CardResourceSelectionPopover from "../../ui/popover/CardResourceSelectionPopover.tsx";
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
  FullStatePayload,
  GameDto,
  GamePhaseComplete,
  GamePhaseDemoSetup,
  GamePhaseStartingCardSelection,
  GameStatusActive,
  GameStatusCompleted,
  GameStatusLobby,
  PlayerDisconnectedPayload,
  PlayerDto,
  PlayerActionDto,
  ResourceType,
  TriggeredEffectDto,
} from "@/types/generated/api-types.ts";
import { shouldShowPaymentModal, createDefaultPayment } from "@/utils/paymentUtils.ts";
import { deepClone, findChangedPaths } from "@/utils/deepCompare.ts";
import { StandardProject } from "@/types/cards.tsx";

type TransitionPhase = "idle" | "lobby" | "loading" | "fadeOutLobby" | "animateUI" | "complete";

type LoadingPhase = "checking" | "selecting" | "joining" | "connecting" | "ready";

export default function GameInterface() {
  const location = useLocation();
  const navigate = useNavigate();
  const { gameId: urlGameId } = useParams<{ gameId?: string }>();
  const { playProductionSound, playTemperatureSound, playWaterPlacementSound, playOxygenSound } =
    useSoundEffects();
  const { showNotification } = useNotifications();
  const [game, setGame] = useState<GameDto | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [transitionPhase, setTransitionPhase] = useState<TransitionPhase>("idle");
  const wasInLobby = useRef(false);
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
  const [showCardEffectsModal, setShowCardEffectsModal] = useState(false);
  const [showActionsModal, setShowActionsModal] = useState(false);
  const [showDebugDropdown, setShowDebugDropdown] = useState(false);
  const [showPerformanceWindow, setShowPerformanceWindow] = useState(false);
  const [tilePlacerPlayerId, setTilePlacerPlayerId] = useState<string | null>(null);

  // Set corporation data directly from player (backend now sends full CardDto)
  useEffect(() => {
    if (currentPlayer?.corporation) {
      setCorporationData(currentPlayer.corporation);
    } else {
      setCorporationData(null);
    }
  }, [currentPlayer?.corporation]);

  // Production phase modal state
  const [showProductionPhaseModal, setShowProductionPhaseModal] = useState(false);
  const [isProductionModalHidden, setIsProductionModalHidden] = useState(false);
  const [openProductionToCardSelection, setOpenProductionToCardSelection] = useState(false);
  const isInitialMount = useRef(true);

  // Card selection state
  const [showCardSelection, setShowCardSelection] = useState(false);
  const [cardDetails, setCardDetails] = useState<CardDto[]>([]);

  // Pending card selection state (for sell patents, etc.)
  const [showPendingCardSelection, setShowPendingCardSelection] = useState(false);

  // Card draw selection state (for card-draw/peek/take/buy effects)
  const [showCardDrawSelection, setShowCardDrawSelection] = useState(false);

  // End game tile highlighting state
  const [tileHighlightMode, setTileHighlightMode] = useState<TileHighlightMode>(null);

  // End game VP indicators state
  const [vpIndicators, setVPIndicators] = useState<TileVPIndicator[]>([]);

  // Choice selection state (for card play)
  const [showChoiceSelection, setShowChoiceSelection] = useState(false);
  const [cardPendingChoice, setCardPendingChoice] = useState<CardDto | null>(null);
  const [pendingCardBehaviorIndex, setPendingCardBehaviorIndex] = useState(0);

  // Action choice selection state (for playing actions with choices)
  const [showActionChoiceSelection, setShowActionChoiceSelection] = useState(false);
  const [actionPendingChoice, setActionPendingChoice] = useState<PlayerActionDto | null>(null);

  // Payment selection state
  const [showPaymentSelection, setShowPaymentSelection] = useState(false);
  const [pendingCardPayment, setPendingCardPayment] = useState<{
    card: CardDto;
    choiceIndex?: number;
  } | null>(null);

  // Card storage selection state
  const [showCardStorageSelection, setShowCardStorageSelection] = useState(false);
  const [pendingCardStorage, setPendingCardStorage] = useState<{
    cardId: string;
    payment: CardPaymentDto;
    choiceIndex?: number;
    resourceType: ResourceType;
    amount: number;
  } | null>(null);

  // Action storage selection state
  const [showActionStorageSelection, setShowActionStorageSelection] = useState(false);
  const [pendingActionStorage, setPendingActionStorage] = useState<{
    cardId: string;
    behaviorIndex: number;
    choiceIndex?: number;
    resourceType: ResourceType;
    amount: number;
  } | null>(null);

  // Target player selection state (for any-player resource/production removal)
  const [showTargetPlayerSelection, setShowTargetPlayerSelection] = useState(false);
  const [pendingTargetPlayer, setPendingTargetPlayer] = useState<{
    cardId: string;
    payment: CardPaymentDto;
    choiceIndex?: number;
    cardStorageTarget?: string;
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
    cardStorageTarget?: string;
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
    cardStorageTarget?: string;
    resourceType: ResourceType;
    amount: number;
  } | null>(null);

  // Tab management
  const [showTabConflict, setShowTabConflict] = useState(false);
  const [conflictingTabInfo, setConflictingTabInfo] = useState<{
    gameId: string;
    playerName: string;
  } | null>(null);

  // Leave game confirmation
  const [showLeaveGameConfirm, setShowLeaveGameConfirm] = useState(false);

  // Change detection
  const previousGameRef = useRef<GameDto | null>(null);
  const [changedPaths, setChangedPaths] = useState<Set<string>>(new Set());

  // Triggered effects notifications
  const [triggeredEffects, setTriggeredEffects] = useState<TriggeredEffectDto[]>([]);

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
      if (!playerId) return;

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

        // Play water placement sound when ocean count increases
        const prevOceans = previousGameRef.current.globalParameters?.oceans;
        const newOceans = updatedGame.globalParameters?.oceans;
        if (prevOceans !== undefined && newOceans !== undefined && newOceans > prevOceans) {
          void playWaterPlacementSound();
        }

        // Play oxygen increase sound when oxygen goes up
        const prevOxygen = previousGameRef.current.globalParameters?.oxygen;
        const newOxygen = updatedGame.globalParameters?.oxygen;
        if (prevOxygen !== undefined && newOxygen !== undefined && newOxygen > prevOxygen) {
          void playOxygenSound();
        }

        // Clear changed paths after animation completes
        if (changes.size > 0) {
          setTimeout(() => {
            setChangedPaths(new Set());
          }, 1500);
        }
      }

      // Store the previous state for next comparison
      previousGameRef.current = deepClone(updatedGame);

      // Extract triggered effects for notifications (clear after short delay to allow component to process)
      if (updatedGame.triggeredEffects && updatedGame.triggeredEffects.length > 0) {
        setTriggeredEffects(updatedGame.triggeredEffects);
        // Clear after component has processed them
        setTimeout(() => setTriggeredEffects([]), 100);
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
    [isReconnecting, playTemperatureSound, playWaterPlacementSound, playOxygenSound],
  );

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
    showNotification({ message: "You were kicked from the game", type: "info" });
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

  // Check if we should show production phase modal based on game state
  useEffect(() => {
    if (!game || !currentPlayer) return;

    // Check if current player has production phase data
    const hasProductionData =
      currentPlayer.productionPhase &&
      currentPlayer.productionPhase.availableCards &&
      currentPlayer.productionPhase.availableCards.length >= 0;

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

  const handleCardSelection = useCallback(
    async (selectedCardIds: string[], corporationId: string) => {
      try {
        // Send card and corporation selection to server - commits immediately
        await globalWebSocketManager.selectStartingCard(selectedCardIds, corporationId);
        // Modal will close automatically when backend clears startingSelection
      } catch (error) {
        console.error("Failed to select cards:", error);
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
      // Overlay closes automatically when backend clears pendingCardDrawSelection
    } catch (error) {
      console.error("Failed to confirm card draw:", error);
    }
  }, []);

  // Helper function to check if outputs need card storage selection
  const needsCardStorageSelection = useCallback(
    (
      outputs: any[] | undefined,
    ): { resourceType: ResourceType; amount: number; target: string } | null => {
      if (!outputs) return null;

      const storageResources = [
        "animal",
        "microbe",
        "floater",
        "science",
        "asteroid",
      ] as ResourceType[];

      for (const output of outputs) {
        if (
          (output.target === "any-card" || output.target === "self-card") &&
          storageResources.includes(output.type as ResourceType)
        ) {
          return {
            resourceType: output.type as ResourceType,
            amount: output.amount || 1,
            target: output.target as string,
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
    ): { resourceType: ResourceType; amount: number; isSteal: boolean } | null => {
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

  const finalizePlayCard = useCallback(
    async (
      cardId: string,
      payment: CardPaymentDto,
      choiceIndex?: number,
      cardStorageTarget?: string,
      cardForBehaviors?: CardDto,
    ) => {
      // Check if any auto-trigger behavior needs target player selection
      const card = cardForBehaviors || currentPlayer?.cards.find((c) => c.id === cardId);
      if (card) {
        const autoTriggerBehaviors = card.behaviors?.filter((b) =>
          b.triggers?.some((t) => t.type === "auto"),
        );
        for (const behavior of autoTriggerBehaviors || []) {
          const outputs =
            choiceIndex !== undefined ? behavior.choices?.[choiceIndex]?.outputs : behavior.outputs;
          const targetInfo = needsTargetPlayerSelection(outputs);
          if (targetInfo) {
            setPendingTargetPlayer({
              cardId,
              payment,
              choiceIndex,
              cardStorageTarget,
              resourceType: targetInfo.resourceType,
              amount: targetInfo.amount,
              isSteal: targetInfo.isSteal,
            });
            setShowTargetPlayerSelection(true);
            return;
          }
        }
      }

      await globalWebSocketManager.playCard(cardId, payment, choiceIndex, cardStorageTarget);
    },
    [currentPlayer?.cards, needsTargetPlayerSelection],
  );

  const handlePlayCard = useCallback(
    async (cardId: string) => {
      try {
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
          // No auto-triggered choices, check if we need payment modal
          if (
            currentPlayer &&
            shouldShowPaymentModal(card, currentPlayer.resources, currentPlayer.paymentSubstitutes)
          ) {
            // Show payment selection modal
            setPendingCardPayment({
              card: card,
              choiceIndex: undefined,
            });
            setShowPaymentSelection(true);
          } else {
            // No payment modal needed, use default all-credits payment
            const payment = createDefaultPayment(card.cost);

            // Check if card needs storage selection
            const autoTriggerBehaviors = card.behaviors?.filter((b) =>
              b.triggers?.some((t) => t.type === "auto"),
            );

            let storageNeeded: {
              resourceType: ResourceType;
              amount: number;
              target: string;
            } | null = null;
            for (const behavior of autoTriggerBehaviors || []) {
              storageNeeded = needsCardStorageSelection(behavior.outputs);
              if (storageNeeded) break;
            }

            if (storageNeeded) {
              if (storageNeeded.target === "self-card") {
                // Self-card target: backend uses sourceCardID, no popover needed
                await finalizePlayCard(cardId, payment, undefined, undefined, card);
              } else {
                // Show storage selection popover for any-card targets
                setPendingCardStorage({
                  cardId: card.id,
                  payment: payment,
                  resourceType: storageNeeded.resourceType,
                  amount: storageNeeded.amount,
                });
                setShowCardStorageSelection(true);
              }
            } else {
              // No storage needed, play the card directly
              await finalizePlayCard(cardId, payment, undefined, undefined, card);
            }
          }
        }
      } catch (error) {
        console.error(`❌ Failed to play card ${cardId}:`, error);
        throw error; // Re-throw to allow CardFanOverlay to handle the error
      }
    },
    [currentPlayer?.cards, needsCardStorageSelection, finalizePlayCard],
  );

  const handleChoiceSelect = useCallback(
    async (choiceIndex: number) => {
      if (!cardPendingChoice || !currentPlayer) return;

      try {
        setShowChoiceSelection(false);

        // Check if we need payment modal
        if (
          shouldShowPaymentModal(
            cardPendingChoice,
            currentPlayer.resources,
            currentPlayer.paymentSubstitutes,
          )
        ) {
          // Show payment selection modal
          setPendingCardPayment({
            card: cardPendingChoice,
            choiceIndex: choiceIndex,
          });
          setShowPaymentSelection(true);
          setCardPendingChoice(null);
          setPendingCardBehaviorIndex(0);
        } else {
          // No payment modal needed, use default all-credits payment
          const payment = createDefaultPayment(cardPendingChoice.cost);

          // Get the selected choice
          const behavior = cardPendingChoice.behaviors?.[pendingCardBehaviorIndex];
          const selectedChoice = behavior?.choices?.[choiceIndex];

          // Check if the selected choice outputs need card storage selection
          const storageInfo = needsCardStorageSelection(selectedChoice?.outputs);

          if (storageInfo) {
            if (storageInfo.target === "self-card") {
              // Self-card target: backend uses sourceCardID, no popover needed
              await finalizePlayCard(
                cardPendingChoice.id,
                payment,
                choiceIndex,
                undefined,
                cardPendingChoice,
              );
              setCardPendingChoice(null);
              setPendingCardBehaviorIndex(0);
            } else {
              // Show card storage selection popover for any-card targets
              setPendingCardStorage({
                cardId: cardPendingChoice.id,
                payment: payment,
                choiceIndex: choiceIndex,
                resourceType: storageInfo.resourceType,
                amount: storageInfo.amount,
              });
              setShowCardStorageSelection(true);
              setCardPendingChoice(null);
              setPendingCardBehaviorIndex(0);
            }
          } else {
            // No card storage needed, play the card directly
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
      needsCardStorageSelection,
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

        // Get the selected choice
        const selectedChoice = actionPendingChoice.behavior.choices?.[choiceIndex];

        // Check if the selected choice outputs need card storage selection
        const storageInfo = needsCardStorageSelection(selectedChoice?.outputs);

        if (storageInfo) {
          if (storageInfo.target === "self-card") {
            // Self-card target: check target player needs
            const targetInfo = needsTargetPlayerSelection(selectedChoice?.outputs);
            if (targetInfo) {
              setPendingActionTargetPlayer({
                cardId: actionPendingChoice.cardId,
                behaviorIndex: actionPendingChoice.behaviorIndex,
                choiceIndex,
                cardStorageTarget: actionPendingChoice.cardId,
                resourceType: targetInfo.resourceType,
                amount: targetInfo.amount,
                isSteal: targetInfo.isSteal,
              });
              setShowActionTargetPlayerSelection(true);
              setActionPendingChoice(null);
            } else {
              await globalWebSocketManager.playCardAction(
                actionPendingChoice.cardId,
                actionPendingChoice.behaviorIndex,
                choiceIndex,
                actionPendingChoice.cardId,
              );
              setActionPendingChoice(null);
            }
          } else {
            // Show action storage selection popover for any-card targets
            setPendingActionStorage({
              cardId: actionPendingChoice.cardId,
              behaviorIndex: actionPendingChoice.behaviorIndex,
              choiceIndex: choiceIndex,
              resourceType: storageInfo.resourceType,
              amount: storageInfo.amount,
            });
            setShowActionStorageSelection(true);
            setActionPendingChoice(null);
          }
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
            await globalWebSocketManager.playCardAction(
              actionPendingChoice.cardId,
              actionPendingChoice.behaviorIndex,
              choiceIndex,
            );
            setActionPendingChoice(null);
          }
        }
      } catch (error) {
        console.error(
          `❌ Failed to play action ${actionPendingChoice.cardId} with choice ${choiceIndex}:`,
          error,
        );
        setActionPendingChoice(null);
      }
    },
    [actionPendingChoice, needsCardStorageSelection, needsTargetPlayerSelection],
  );

  const handleActionChoiceCancel = useCallback(() => {
    setShowActionChoiceSelection(false);
    setActionPendingChoice(null);
  }, []);

  // Payment selection callbacks
  const handlePaymentConfirm = useCallback(
    async (payment: CardPaymentDto) => {
      if (!pendingCardPayment || !currentPlayer) return;

      try {
        setShowPaymentSelection(false);

        // Check if card storage selection is still needed
        const autoTriggerBehaviors = pendingCardPayment.card.behaviors?.filter(
          (b) => b.triggers && b.triggers.length > 0 && b.triggers.some((t) => t.type === "auto"),
        );

        let storageNeeded: {
          resourceType: ResourceType;
          amount: number;
          target: string;
        } | null = null;
        for (const behavior of autoTriggerBehaviors || []) {
          storageNeeded = needsCardStorageSelection(behavior.outputs);
          if (storageNeeded) break;
        }

        if (storageNeeded) {
          if (storageNeeded.target === "self-card") {
            // Self-card target: backend uses sourceCardID, no popover needed
            await finalizePlayCard(
              pendingCardPayment.card.id,
              payment,
              pendingCardPayment.choiceIndex,
              undefined,
              pendingCardPayment.card,
            );
          } else {
            // Show storage selection popover for any-card targets
            setPendingCardStorage({
              cardId: pendingCardPayment.card.id,
              payment: payment,
              choiceIndex: pendingCardPayment.choiceIndex,
              resourceType: storageNeeded.resourceType,
              amount: storageNeeded.amount,
            });
            setShowCardStorageSelection(true);
          }
        } else {
          // No storage needed, play the card directly
          await finalizePlayCard(
            pendingCardPayment.card.id,
            payment,
            pendingCardPayment.choiceIndex,
            undefined,
            pendingCardPayment.card,
          );
        }

        setPendingCardPayment(null);
      } catch (error) {
        console.error(`❌ Failed to play card with payment:`, error);
        setPendingCardPayment(null);
      }
    },
    [pendingCardPayment, currentPlayer, needsCardStorageSelection, finalizePlayCard],
  );

  const handlePaymentCancel = useCallback(() => {
    setShowPaymentSelection(false);
    setPendingCardPayment(null);
  }, []);

  const handleCardStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingCardStorage) return;

      try {
        setShowCardStorageSelection(false);

        // Find the card to check for target player needs
        const card = currentPlayer?.cards.find((c) => c.id === pendingCardStorage.cardId);
        const autoTriggerBehaviors = card?.behaviors?.filter((b) =>
          b.triggers?.some((t) => t.type === "auto"),
        );
        let targetInfo: { resourceType: ResourceType; amount: number; isSteal: boolean } | null =
          null;
        for (const behavior of autoTriggerBehaviors || []) {
          const outputs =
            pendingCardStorage.choiceIndex !== undefined
              ? behavior.choices?.[pendingCardStorage.choiceIndex]?.outputs
              : behavior.outputs;
          targetInfo = needsTargetPlayerSelection(outputs);
          if (targetInfo) break;
        }

        if (targetInfo) {
          setPendingTargetPlayer({
            cardId: pendingCardStorage.cardId,
            payment: pendingCardStorage.payment,
            choiceIndex: pendingCardStorage.choiceIndex,
            cardStorageTarget: targetCardId,
            resourceType: targetInfo.resourceType,
            amount: targetInfo.amount,
            isSteal: targetInfo.isSteal,
          });
          setShowTargetPlayerSelection(true);
          setPendingCardStorage(null);
          return;
        }

        await globalWebSocketManager.playCard(
          pendingCardStorage.cardId,
          pendingCardStorage.payment,
          pendingCardStorage.choiceIndex,
          targetCardId,
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
    [pendingCardStorage, currentPlayer?.cards, needsTargetPlayerSelection],
  );

  const handleCardStorageCancel = useCallback(() => {
    setShowCardStorageSelection(false);
    setPendingCardStorage(null);
  }, []);

  const handleActionStorageSelect = useCallback(
    async (targetCardId: string) => {
      if (!pendingActionStorage) return;

      try {
        setShowActionStorageSelection(false);

        // Check if the action also needs target player selection
        // We look at the action's behavior outputs
        const actionPlayer = currentPlayer;
        const action = actionPlayer?.actions?.find(
          (a) =>
            a.cardId === pendingActionStorage.cardId &&
            a.behaviorIndex === pendingActionStorage.behaviorIndex,
        );
        const outputs =
          pendingActionStorage.choiceIndex !== undefined
            ? action?.behavior.choices?.[pendingActionStorage.choiceIndex]?.outputs
            : action?.behavior.outputs;
        const targetInfo = needsTargetPlayerSelection(outputs);

        if (targetInfo) {
          setPendingActionTargetPlayer({
            cardId: pendingActionStorage.cardId,
            behaviorIndex: pendingActionStorage.behaviorIndex,
            choiceIndex: pendingActionStorage.choiceIndex,
            cardStorageTarget: targetCardId,
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
          targetCardId,
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
          pendingTargetPlayer.cardStorageTarget,
          targetPlayerId,
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

  const handleActionTargetPlayerSelect = useCallback(
    async (targetPlayerId: string) => {
      if (!pendingActionTargetPlayer) return;

      try {
        setShowActionTargetPlayerSelection(false);
        await globalWebSocketManager.playCardAction(
          pendingActionTargetPlayer.cardId,
          pendingActionTargetPlayer.behaviorIndex,
          pendingActionTargetPlayer.choiceIndex,
          pendingActionTargetPlayer.cardStorageTarget,
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
          pendingCardResourceInput.cardStorageTarget,
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
      const response = await fetch(
        `http://localhost:3001/api/v1/games/${gameId}?playerId=${playerId}`,
      );
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
    globalWebSocketManager.on("player-disconnected", handlePlayerDisconnected);
    globalWebSocketManager.on("player-kicked", handlePlayerKicked);
    globalWebSocketManager.on("error", handleError);
    globalWebSocketManager.on("disconnect", handleDisconnect);
    globalWebSocketManager.on("max-reconnects-reached", handleMaxReconnectsReached);

    isWebSocketInitialized.current = true;

    return () => {
      globalWebSocketManager.off("game-updated", handleGameUpdated);
      globalWebSocketManager.off("full-state", handleFullState);
      globalWebSocketManager.off("player-disconnected", handlePlayerDisconnected);
      globalWebSocketManager.off("player-kicked", handlePlayerKicked);
      globalWebSocketManager.off("error", handleError);
      globalWebSocketManager.off("disconnect", handleDisconnect);
      globalWebSocketManager.off("max-reconnects-reached", handleMaxReconnectsReached);
      isWebSocketInitialized.current = false;
    };
  }, [
    handleGameUpdated,
    handleFullState,
    handlePlayerDisconnected,
    handlePlayerKicked,
    handleError,
    handleDisconnect,
    handleMaxReconnectsReached,
  ]);

  // Handle action selection from card actions
  const handleActionSelect = useCallback(
    (action: PlayerActionDto) => {
      // Block actions when tile selection is pending
      if (currentPlayer?.pendingTileSelection) {
        return;
      }

      // Check if this action has choices
      if (action.behavior.choices && action.behavior.choices.length > 0) {
        // Action has choices, show the choice selection popover
        setActionPendingChoice(action);
        setShowActionChoiceSelection(true);
      } else {
        // No choices, check if inputs need card resource selection (steal-from-any-card)
        const cardResourceInfo = needsCardResourceInput(action.behavior.inputs);

        if (cardResourceInfo) {
          // Determine cardStorageTarget for the output side
          const storageInfo = needsCardStorageSelection(action.behavior.outputs);
          const cardStorageTarget = storageInfo?.target === "self-card" ? action.cardId : undefined;

          setPendingCardResourceInput({
            cardId: action.cardId,
            behaviorIndex: action.behaviorIndex,
            cardStorageTarget,
            resourceType: cardResourceInfo.resourceType,
            amount: cardResourceInfo.amount,
          });
          setShowCardResourceSelection(true);
        } else {
          // No card resource input, check if action outputs need card storage selection
          const storageInfo = needsCardStorageSelection(action.behavior.outputs);

          if (storageInfo) {
            if (storageInfo.target === "self-card") {
              // Self-card target: skip popover and send directly with the source card as target
              // Check target player needs
              const targetInfo = needsTargetPlayerSelection(action.behavior.outputs);
              if (targetInfo) {
                setPendingActionTargetPlayer({
                  cardId: action.cardId,
                  behaviorIndex: action.behaviorIndex,
                  cardStorageTarget: action.cardId,
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
                  action.cardId,
                );
              }
            } else {
              // Show action storage selection popover for any-card targets
              setPendingActionStorage({
                cardId: action.cardId,
                behaviorIndex: action.behaviorIndex,
                resourceType: storageInfo.resourceType,
                amount: storageInfo.amount,
              });
              setShowActionStorageSelection(true);
            }
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
    },
    [
      currentPlayer?.pendingTileSelection,
      needsCardStorageSelection,
      needsTargetPlayerSelection,
      needsCardResourceInput,
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

  useEffect(() => {
    let aborted = false;

    const initializeGame = async () => {
      setLoadingPhase("checking");

      const routeState = location.state as {
        game?: GameDto;
        playerId?: string;
        playerName?: string;
        isReconnection?: boolean;
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
        fetchedGame = await apiService.getGame(gameId, savedSession?.playerId);
      } catch {
        navigate("/", { replace: true, state: { error: "Could not find game" } });
        return;
      }

      if (aborted) return;

      if (!fetchedGame) {
        clearGameSession();
        navigate("/", { replace: true, state: { error: "Could not find game" } });
        return;
      }

      // 3. Check cached session
      const cachedForThisGame = savedSession && savedSession.gameId === gameId;

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

      // 5. Check if this is a join link
      const urlParams = new URLSearchParams(window.location.search);
      const isJoinLink = urlParams.get("type") === "join";

      if (isJoinLink) {
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

  // Extract card details directly from game data (backend now sends full card objects)
  const extractCardDetails = useCallback((cards: CardDto[]) => {
    setCardDetails(cards);
  }, []);

  // Show/hide starting card selection overlay based on backend state
  useEffect(() => {
    const cards = game?.currentPlayer?.selectStartingCardsPhase?.availableCards;
    const hasCardSelection = cards && cards.length > 0;

    if (
      game?.currentPhase === GamePhaseStartingCardSelection &&
      game?.status === GameStatusActive &&
      hasCardSelection &&
      !showCardSelection
    ) {
      extractCardDetails(cards);
      setShowCardSelection(true);
    } else if (showCardSelection && !hasCardSelection) {
      setShowCardSelection(false);
    }
  }, [
    game?.currentPhase,
    game?.status,
    game?.currentPlayer?.selectStartingCardsPhase?.availableCards,
    showCardSelection,
    extractCardDetails,
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
            setShowCardEffectsModal(true);
            break;
          case "d":
          case "D":
            event.preventDefault();
            setShowDebugDropdown(!showDebugDropdown);
            break;
        }
      }
    };

    window.addEventListener("keydown", handleKeyPress);
    return () => window.removeEventListener("keydown", handleKeyPress);
  }, [showDebugDropdown]);

  const loadingMessage = (() => {
    if (loadingPhase === "selecting" || loadingPhase === "joining") {
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
  const isLobbyPhase = game?.status === GameStatusLobby;

  // Pre-game phase covers both lobby AND starting card selection
  const isPreGamePhase =
    isLobbyPhase ||
    (game?.status === GameStatusActive && game?.currentPhase === GamePhaseStartingCardSelection);

  // Show waiting modal when player has finished card selection but others haven't
  const showWaitingForPlayers =
    game?.status === GameStatusActive &&
    game?.currentPhase === GamePhaseStartingCardSelection &&
    !game?.currentPlayer?.selectStartingCardsPhase &&
    !!currentPlayer?.corporation;

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
    if (transitionPhase !== "fadeOutLobby") return;

    const animateTimer = setTimeout(() => {
      setTransitionPhase("animateUI");
    }, 1500);

    const completeTimer = setTimeout(() => {
      setTransitionPhase("complete");
    }, 1500 + 2500);

    return () => {
      clearTimeout(animateTimer);
      clearTimeout(completeTimer);
    };
  }, [transitionPhase]);

  // On mount: if game is already active (reload/reconnect), stop ambient immediately
  useEffect(() => {
    if (game && !isPreGamePhase && !wasInLobby.current) {
      audioService.stopAmbient();
    }
  }, [game, isPreGamePhase]);

  const bottomBarCallbacks = useMemo(
    () => ({
      onOpenCardEffectsModal: () => setShowCardEffectsModal(true),
      onOpenCardsPlayedModal: () => setShowCardsPlayedModal(true),
      onOpenActionsModal: () => setShowActionsModal(true),
      onActionSelect: handleActionSelect,
      onConvertPlantsToGreenery: handleConvertPlantsToGreenery,
      onConvertHeatToTemperature: handleConvertHeatToTemperature,
    }),
    [handleActionSelect, handleConvertPlantsToGreenery, handleConvertHeatToTemperature],
  );

  // Check if we need the persistent backdrop (during overlay transitions)
  const shouldShowBackdrop = showCardSelection;

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

      {game && loadingPhase !== "selecting" && loadingPhase !== "joining" && (
        <GameLayout
          gameState={game}
          currentPlayer={currentPlayer}
          playedCards={currentPlayer?.playedCards || []}
          corporationCard={corporationData}
          showCardSelection={showCardSelection}
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
          onSkyboxReady={handleSkyboxReady}
          onGpuReady={handleGpuReady}
        />
      )}

      <CardsPlayedModal
        isVisible={showCardsPlayedModal}
        onClose={() => setShowCardsPlayedModal(false)}
        cards={currentPlayer?.playedCards || []}
      />

      <EffectsModal
        isVisible={showCardEffectsModal}
        onClose={() => setShowCardEffectsModal(false)}
        effects={currentPlayer?.effects || []}
      />

      <ActionsModal
        isVisible={showActionsModal}
        onClose={() => setShowActionsModal(false)}
        actions={currentPlayer?.actions || []}
        onActionSelect={handleActionSelect}
        gameState={game ?? undefined}
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
          onOpenTilePlacer={(playerId) => setTilePlacerPlayerId(playerId)}
        />

        <PerformanceWindow
          isVisible={showPerformanceWindow}
          onClose={() => setShowPerformanceWindow(false)}
        />

        {tilePlacerPlayerId && game && (
          <TilePlacerWindow
            playerId={tilePlacerPlayerId}
            playerName={
              [game.currentPlayer, ...game.otherPlayers].find((p) => p.id === tilePlacerPlayerId)
                ?.name || tilePlacerPlayerId
            }
            onClose={() => setTilePlacerPlayerId(null)}
          />
        )}
      </WindowManagerProvider>

      {(transitionPhase === "lobby" ||
        transitionPhase === "loading" ||
        transitionPhase === "fadeOutLobby" ||
        loadingPhase === "selecting" ||
        loadingPhase === "joining") && (
        <div
          className={
            transitionPhase === "fadeOutLobby" ? "animate-[fadeOut_1500ms_ease-out_forwards]" : ""
          }
        >
          <SpaceBackground animationSpeed={0.5} overlayOpacity={0.3} />
        </div>
      )}

      {lobbyMounted && game && playerId && (
        <WaitingRoomOverlay
          game={game}
          playerId={playerId}
          visible={isLobbyPhase}
          onExited={() => setLobbyMounted(false)}
        />
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
          onCancel={handlePlayerSelectionCancel}
        />
      )}

      {loadingPhase === "joining" && gameForSelection && (
        <JoinGameOverlay game={gameForSelection} onCancel={handlePlayerSelectionCancel} />
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

      {/* Starting card selection overlay */}
      <StartingCardSelectionOverlay
        isOpen={showCardSelection}
        cards={cardDetails}
        availableCorporations={
          game?.currentPlayer?.selectStartingCardsPhase?.availableCorporations || []
        }
        playerCredits={currentPlayer?.resources?.credits || 40}
        onSelectCards={handleCardSelection}
      />

      {/* Waiting for other players to finish card selection */}
      {showWaitingForPlayers && game && (
        <>
          <div className="z-[10000]">
            <MainMenuSettingsButton />
          </div>
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
                    }[] = [];

                    if (game.currentPlayer) {
                      allPlayers.push({
                        id: game.currentPlayer.id,
                        name: game.currentPlayer.name,
                        isReady:
                          !game.currentPlayer.selectStartingCardsPhase &&
                          !!game.currentPlayer.corporation,
                        isSelf: true,
                      });
                    }

                    game.otherPlayers?.forEach((other) => {
                      allPlayers.push({
                        id: other.id,
                        name: other.name,
                        isReady: !other.selectStartingCardsPhase && !!other.corporation,
                        isSelf: false,
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

      {/* Card fan overlay for hand cards */}
      {game && currentPlayer && (
        <div
          className={
            transitionPhase === "animateUI"
              ? "animate-[uiFadeIn_1200ms_ease-out_both]"
              : transitionPhase === "loading" || transitionPhase === "fadeOutLobby"
                ? "opacity-0"
                : ""
          }
        >
          <CardFanOverlay
            cards={currentPlayer.cards || []}
            hideWhenModalOpen={
              showCardSelection ||
              showPendingCardSelection ||
              showCardDrawSelection ||
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

      {/* Card storage selection popover */}
      {pendingCardStorage && (
        <CardStorageSelectionPopover
          resourceType={pendingCardStorage.resourceType}
          amount={pendingCardStorage.amount}
          playedCards={currentPlayer?.playedCards || []}
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
          playedCards={currentPlayer?.playedCards || []}
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
