import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { GameDto } from "@/types/generated/api-types";
import { globalWebSocketManager } from "@/services/globalWebSocketManager";
import { saveGameSession, getRememberedPlayerName } from "@/utils/sessionStorage";
import { useNotifications } from "@/contexts/NotificationContext";

export function useJoinGame({ game }: { game: GameDto | null }) {
  const navigate = useNavigate();
  const { showNotification } = useNotifications();
  const [playerName, setPlayerName] = useState(getRememberedPlayerName);
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const busy = useRef(false);
  const cancel = useRef<(() => void) | null>(null);
  const mounted = useRef(true);
  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      cancel.current?.();
    };
  }, []);

  const handleJoin = async (target = game) => {
    if (busy.current || !target) {
      return;
    }
    setErrorMessage(null);
    const name = playerName.trim();
    if (name.length < 2) {
      setErrorMessage("Name must be at least 2 characters long");
      showNotification({ message: "Name must be at least 2 characters long", type: "error" });
      return;
    }
    busy.current = true;
    setIsLoading(true);
    try {
      await globalWebSocketManager.initialize();
      if (!mounted.current) {
        return;
      }
      const connected = await new Promise<GameDto>((resolve, reject) => {
        const cleanup = () => {
          globalWebSocketManager.off("game-updated", updated);
          globalWebSocketManager.off("error", failed);
          globalWebSocketManager.off("disconnect", disconnected);
          cancel.current = null;
        };
        const updated = (state: GameDto) => {
          if (state.id === target.id && state.currentPlayer?.name === name) {
            cleanup();
            resolve(state);
          }
        };
        const failed = (error: { message?: string; error?: string }) => {
          cleanup();
          reject(new Error(error.message ?? error.error ?? "Could not join game"));
        };
        const disconnected = () => failed({ error: "Connection lost. Try again." });
        cancel.current = () => {
          cleanup();
          reject(new Error("Join cancelled"));
        };
        globalWebSocketManager.on("game-updated", updated);
        globalWebSocketManager.on("error", failed);
        void globalWebSocketManager
          .playerConnect(name, target.id)
          .then(() => {
            if (cancel.current) {
              globalWebSocketManager.on("disconnect", disconnected);
            }
          })
          .catch(failed);
      });
      if (!mounted.current || !connected.currentPlayer) {
        return;
      }
      const playerId = connected.currentPlayer.id;
      saveGameSession({
        gameId: connected.id,
        playerId,
        playerName: name,
        joinedAt: new Date().toISOString(),
      });
      navigate("/game", { state: { game: connected, playerId, playerName: name } });
    } catch (error) {
      if (mounted.current) {
        setErrorMessage(error instanceof Error ? error.message : "Could not join game");
        showNotification({
          message: error instanceof Error ? error.message : "Could not join game",
          type: "error",
        });
      }
    } finally {
      busy.current = false;
      if (mounted.current) {
        setIsLoading(false);
      }
    }
  };
  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      void handleJoin();
    }
  };
  return {
    playerName,
    errorMessage,
    setPlayerName,
    isLoading,
    handleJoin,
    handleKeyDown,
    loadingMessage: "Joining game...",
  };
}
