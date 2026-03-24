import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { GameDto } from "@/types/generated/api-types";
import { globalWebSocketManager } from "@/services/globalWebSocketManager";
import { skyboxCache } from "@/services/SkyboxCache";
import { saveGameSession } from "@/utils/sessionStorage";
import { useNotifications } from "@/contexts/NotificationContext";

interface UseJoinGameOptions {
  game: GameDto | null;
}

interface UseJoinGameReturn {
  playerName: string;
  setPlayerName: (name: string) => void;
  isLoading: boolean;
  loadingStep: "game" | "environment" | null;
  handleJoin: () => Promise<void>;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  loadingMessage: string;
}

export function useJoinGame({ game }: UseJoinGameOptions): UseJoinGameReturn {
  const navigate = useNavigate();
  const { showNotification } = useNotifications();
  const [playerName, setPlayerName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [loadingStep, setLoadingStep] = useState<"game" | "environment" | null>(null);

  const handleJoin = async () => {
    if (!game) {
      showNotification({ message: "No game selected", type: "error" });
      return;
    }

    const trimmed = playerName.trim();

    if (!trimmed) {
      showNotification({ message: "Please enter your name", type: "error" });
      return;
    }

    if (trimmed.length < 2) {
      showNotification({ message: "Name must be at least 2 characters long", type: "error" });
      return;
    }

    setIsLoading(true);
    setLoadingStep("game");

    try {
      if (!skyboxCache.isReady()) {
        setLoadingStep("environment");
        await skyboxCache.preload();
      }

      setLoadingStep("game");
      await globalWebSocketManager.initialize();

      const handleGameUpdated = (gameData: any) => {
        const allPlayers = [gameData.currentPlayer, ...(gameData.otherPlayers || [])].filter(
          Boolean,
        );
        const connectedPlayer = allPlayers.find((p: any) => p.name === trimmed);

        if (connectedPlayer) {
          saveGameSession({
            gameId: game.id,
            playerId: connectedPlayer.id,
            playerName: trimmed,
            joinedAt: new Date().toISOString(),
          });

          navigate("/game", {
            state: {
              game: gameData,
              playerId: connectedPlayer.id,
              playerName: trimmed,
            },
          });

          globalWebSocketManager.off("game-updated", handleGameUpdated);
        }
      };

      globalWebSocketManager.on("game-updated", handleGameUpdated);
      globalWebSocketManager.playerConnect(trimmed, game.id);
    } catch (err) {
      if (err instanceof Error) {
        if (err.message.includes("WebSocket")) {
          showNotification({ message: "Connection failed. Please try again.", type: "error" });
        } else {
          showNotification({ message: err.message || "Failed to join game", type: "error" });
        }
      } else {
        showNotification({ message: "Failed to join game", type: "error" });
      }
    } finally {
      setIsLoading(false);
      setLoadingStep(null);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleJoin();
    }
  };

  const loadingMessage = (() => {
    if (loadingStep === "environment") return "Loading...";
    return "Joining game...";
  })();

  return {
    playerName,
    setPlayerName,
    isLoading,
    loadingStep,
    handleJoin,
    handleKeyDown,
    loadingMessage,
  };
}
