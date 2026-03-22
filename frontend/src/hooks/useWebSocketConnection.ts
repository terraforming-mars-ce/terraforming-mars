import { useEffect, useRef } from "react";
import type { NavigateFunction } from "react-router-dom";
import type { MutableRefObject } from "react";
import { config } from "@/config.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import { skyboxCache } from "@/services/SkyboxCache.ts";
import { useGameStore } from "@/stores/gameStore.ts";
import { useUIOverlayStore } from "@/stores/uiOverlayStore.ts";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";
import { deepClone, findChangedPaths } from "@/utils/deepCompare.ts";
import { clearGameSession, getGameSession, saveGameSession } from "@/utils/sessionStorage.ts";
import type {
  ChatMessageDto,
  GameDto,
  FullStatePayload,
  PlayerDisconnectedPayload,
  StateDiffDto,
} from "@/types/generated/api-types.ts";
import { GamePhaseAction, GamePhaseInitApplyCorp } from "@/types/generated/api-types.ts";
import type { GameEvent } from "@/hooks/useGameEvent.ts";
import type { PlayedCardNotification } from "@/hooks/usePlayedCardNotification.ts";

type GameEventInput = Omit<GameEvent, "id">;
type PlayedCardNotificationInput = Omit<PlayedCardNotification, "id">;

interface ShowNotificationOptions {
  message: string;
  type: "error" | "warning" | "info";
  duration?: number;
}

export function useWebSocketConnection(
  navigate: NavigateFunction,
  enqueueGameEvent: (input: GameEventInput) => void,
  enqueuePlayedCard: (input: PlayedCardNotificationInput) => void,
  notificationQueueDoneAt: MutableRefObject<number>,
  showNotification: (options: ShowNotificationOptions) => string,
): { attemptReconnection: () => Promise<void> } {
  const isWebSocketInitialized = useRef(false);
  const previousGameRef = useRef<GameDto | null>(null);
  const hasReceivedInitialLogs = useRef(false);

  const {
    playTemperatureSound,
    playOxygenSound,
    playVenusSound,
    playYourTurnSound,
    playAwardFundedSound,
  } = useSoundEffects();

  const attemptReconnection = async () => {
    try {
      const savedGameData = getGameSession();
      if (!savedGameData) {
        clearGameSession();
        navigate("/", { replace: true });
        return;
      }

      const { gameId, playerId, playerName } = savedGameData;

      useGameStore.getState().setReconnectionStep("game");

      const response = await fetch(`${config.apiUrl}/games/${gameId}?playerId=${playerId}`);
      if (!response.ok) {
        clearGameSession();
        navigate("/", { replace: true });
        return;
      }

      const gameData = await response.json();

      const store = useGameStore.getState();
      store.setGame(gameData.game);
      store.setPlayerId(playerId);

      const player = gameData.game.currentPlayer;
      store.setCurrentPlayer(player || null);

      if (!skyboxCache.isReady()) {
        useGameStore.getState().setReconnectionStep("environment");
        await skyboxCache.preload();
      }

      globalWebSocketManager.playerConnect(playerName, gameId, playerId);
    } catch (error) {
      console.error("Reconnection failed:", error);
      const store = useGameStore.getState();
      store.setIsReconnecting(false);
      store.setReconnectionStep(null);
    }
  };

  useEffect(() => {
    if (isWebSocketInitialized.current) {
      return;
    }

    const initialPlayerId = useGameStore.getState().playerId;
    if (initialPlayerId) {
      globalWebSocketManager.setCurrentPlayerId(initialPlayerId);
    }

    const handleGameUpdated = (updatedGame: GameDto) => {
      const myPlayerId = useGameStore.getState().playerId;
      if (!myPlayerId && !updatedGame.isSpectator) {
        return;
      }

      if (previousGameRef.current) {
        const changes = findChangedPaths(previousGameRef.current, updatedGame);
        useGameStore.getState().setChangedPaths(changes);

        const prevTemp = previousGameRef.current.globalParameters?.temperature;
        const newTemp = updatedGame.globalParameters?.temperature;
        if (prevTemp !== undefined && newTemp !== undefined && newTemp > prevTemp) {
          void playTemperatureSound();
        }

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

        const prevVenus = previousGameRef.current.globalParameters?.venus;
        const newVenus = updatedGame.globalParameters?.venus;
        if (prevVenus !== undefined && newVenus !== undefined && newVenus > prevVenus) {
          void playVenusSound();
        }

        const prevTurn = previousGameRef.current.currentTurn;
        const prevPhase = previousGameRef.current.currentPhase;
        const newTurn = updatedGame.currentTurn;
        const newPhase = updatedGame.currentPhase;
        const isNowMyActionTurn = newTurn === myPlayerId && newPhase === GamePhaseAction;
        const wasMyActionTurn = prevTurn === myPlayerId && prevPhase === GamePhaseAction;
        if (isNowMyActionTurn && !wasMyActionTurn) {
          void playYourTurnSound();
          enqueueGameEvent({ title: "YOUR TURN", duration: 2500 });
        }

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

        const prevAwards = previousGameRef.current.awards;
        const newAwards = updatedGame.awards;
        if (prevAwards && newAwards) {
          for (const aw of newAwards) {
            if (aw.isFunded && aw.fundedBy) {
              const prev = prevAwards.find((p) => p.type === aw.type);
              if (prev && !prev.isFunded) {
                void playAwardFundedSound();
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

        if (!previousGameRef.current.isLastRound && updatedGame.isLastRound) {
          enqueueGameEvent({ title: "LAST ROUND", duration: 4000 });
        }

        const prevProjects = previousGameRef.current.projectFunding;
        const newProjects = updatedGame.projectFunding;
        if (prevProjects && newProjects) {
          for (const proj of newProjects) {
            const prev = prevProjects.find((p) => p.id === proj.id);
            if (prev && proj.seatOwners.length > prev.seatOwners.length) {
              const newOwner = proj.seatOwners[proj.seatOwners.length - 1];
              const allPlayers = [updatedGame.currentPlayer, ...(updatedGame.otherPlayers ?? [])];
              const fundPlayer = allPlayers.find((p) => p.id === newOwner.playerId);
              enqueueGameEvent({
                title: "PROJECT FUNDED",
                achievementName: proj.name,
                playerName: fundPlayer?.name ?? "Unknown",
                playerColor: fundPlayer?.color ?? "#64c8ff",
                duration: 4000,
              });
            }
          }
        }

        if (changes.size > 0) {
          setTimeout(() => {
            useGameStore.getState().setChangedPaths(new Set());
          }, 1500);
        }
      }

      if (!previousGameRef.current && updatedGame.chatMessages?.length) {
        useGameStore.getState().setChatMessages(updatedGame.chatMessages);
      }

      previousGameRef.current = deepClone(updatedGame);

      if (updatedGame.triggeredEffects && updatedGame.triggeredEffects.length > 0) {
        const store = useGameStore.getState();
        store.setTriggeredEffects(updatedGame.triggeredEffects);
        const notificationCount = updatedGame.triggeredEffects.length;
        notificationQueueDoneAt.current = Date.now() + notificationCount * 2500;
        setTimeout(() => useGameStore.getState().setTriggeredEffects([]), 100);

        if (updatedGame.currentPhase === GamePhaseInitApplyCorp) {
          const hasMyEffect = updatedGame.triggeredEffects.some(
            (e) => e.playerId === updatedGame.currentPlayer?.id,
          );
          if (hasMyEffect) {
            store.setShowCorp(true);
          }
        }
      }

      const store = useGameStore.getState();
      store.setGameRef(updatedGame);
      store.setGame(updatedGame);
      store.setIsConnected(true);

      if (store.isReconnecting) {
        store.setIsReconnecting(false);
        store.setReconnectionStep(null);
      }

      const updatedPlayer = updatedGame.currentPlayer;
      store.setCurrentPlayer(updatedPlayer || null);

      if (updatedPlayer && !updatedPlayer.corporation) {
        useUIOverlayStore.getState().setShowCorporationModal(true);
      } else {
        useUIOverlayStore.getState().setShowCorporationModal(false);
      }
    };

    const handleFullState = (statePayload: FullStatePayload) => {
      if (statePayload.game) {
        handleGameUpdated(statePayload.game);
      }
    };

    const handleLogUpdate = (logs: StateDiffDto[]) => {
      if (!hasReceivedInitialLogs.current) {
        hasReceivedInitialLogs.current = true;
        return;
      }
      const store = useGameStore.getState();
      const latestGame = store.gameRef;
      const myPlayerId = store.playerId;
      if (!latestGame || !myPlayerId || store.isReconnecting) {
        return;
      }
      for (const log of logs) {
        if (log.sourceType !== "card_play") {
          continue;
        }
        if (log.playerId === myPlayerId) {
          continue;
        }
        const playerChanges = log.changes?.playerChanges?.[log.playerId];
        const cardIds = playerChanges?.cardsPlayed;
        if (!cardIds || cardIds.length === 0) {
          continue;
        }
        const otherPlayer = latestGame.otherPlayers?.find((p) => p.id === log.playerId);
        if (!otherPlayer) {
          continue;
        }
        for (const cardId of cardIds) {
          const card = otherPlayer.playedCards.find((c) => c.id === cardId);
          if (!card) {
            continue;
          }
          enqueuePlayedCard({
            card,
            playerName: otherPlayer.name,
            playerColor: otherPlayer.color,
          });
        }
      }
    };

    const handleError = () => {};

    const handlePlayerKicked = () => {
      clearGameSession();
      globalWebSocketManager.disconnect();
      navigate("/", { replace: true });
      showNotification({
        message: "You were kicked from the game",
        type: "info",
      });
    };

    const handleGameEnded = () => {
      clearGameSession();
      globalWebSocketManager.disconnect();
      navigate("/", { replace: true });
      showNotification({
        message: "The host ended the game",
        type: "info",
      });
    };

    const handleDisconnect = () => {
      if (globalWebSocketManager.isGracefulDisconnect()) {
        return;
      }

      useGameStore.getState().setIsConnected(false);

      if (useGameStore.getState().playerId) {
        useGameStore.getState().setIsReconnecting(true);

        const savedGameData = getGameSession();
        if (savedGameData) {
          void attemptReconnection();
        } else {
          clearGameSession();
          navigate("/", { replace: true });
        }
      }
    };

    const handlePlayerDisconnected = (_payload: PlayerDisconnectedPayload) => {};

    const handleMaxReconnectsReached = () => {
      clearGameSession();
      navigate("/", { state: { error: "Server is down", persistent: true } });
    };

    const handleChatUpdate = (chatMessage: ChatMessageDto) => {
      useGameStore.getState().addChatMessage(chatMessage);
    };

    const handleSpectatorKicked = () => {
      globalWebSocketManager.disconnect();
      navigate("/", { replace: true });
    };

    const handleSpectatorIdReceived = (payload: { spectatorId: string }) => {
      const gameId = globalWebSocketManager.gameId;
      if (!gameId || !payload.spectatorId) {
        return;
      }
      useGameStore.getState().setPlayerId(payload.spectatorId);
      const savedSession = getGameSession();
      if (savedSession && savedSession.gameId === gameId) {
        saveGameSession({
          ...savedSession,
          playerId: payload.spectatorId,
          isSpectator: true,
        });
      }
    };

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
  }, []);

  return { attemptReconnection };
}
