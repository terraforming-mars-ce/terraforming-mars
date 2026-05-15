import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager";
import { skyboxCache } from "../../services/SkyboxCache.ts";
import LoadingOverlay from "../game/view/LoadingOverlay.tsx";
import BackButton from "../ui/buttons/BackButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useNotifications } from "../../contexts/NotificationContext.tsx";

const CreateGamePage: React.FC = () => {
  const navigate = useNavigate();
  const { showNotification } = useNotifications();
  const [playerName, setPlayerName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [loadingStep, setLoadingStep] = useState<"game" | "environment" | null>(null);
  const [skyboxReady, setSkyboxReady] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);

  useEffect(() => {
    if (skyboxCache.isReady()) {
      setSkyboxReady(true);
    }
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!playerName.trim()) {
      showNotification({ message: "Please enter your name", type: "error" });
      return;
    }

    if (playerName.trim().length < 2) {
      showNotification({ message: "Name must be at least 2 characters long", type: "error" });
      return;
    }

    setIsLoading(true);
    setLoadingStep("game");

    try {
      const game = await apiService.createGame();

      if (!skyboxReady) {
        setLoadingStep("environment");
        await skyboxCache.preload();
      }

      setLoadingStep("game");
      await globalWebSocketManager.initialize();

      const handleGameUpdated = (gameData: any) => {
        const allPlayers = [gameData.currentPlayer, ...(gameData.otherPlayers || [])].filter(
          Boolean,
        );

        const connectedPlayer = allPlayers.find((p: any) => p.name === playerName.trim());

        if (connectedPlayer) {
          const storedData = {
            gameId: gameData.id,
            playerId: connectedPlayer.id,
            playerName: playerName.trim(),
            createdAt: new Date().toISOString(),
          };
          localStorage.setItem("terraforming-mars-game", JSON.stringify(storedData));

          navigate("/game", {
            state: {
              game: gameData,
              playerId: connectedPlayer.id,
              playerName: playerName.trim(),
            },
          });

          globalWebSocketManager.off("game-updated", handleGameUpdated);
        }
      };

      globalWebSocketManager.on("game-updated", handleGameUpdated);

      globalWebSocketManager.playerConnect(playerName.trim(), game.id);
    } catch (err) {
      showNotification({
        message: err instanceof Error ? err.message : "Failed to create game",
        type: "error",
      });
    } finally {
      setIsLoading(false);
      setLoadingStep(null);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPlayerName(e.target.value);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      void handleSubmit(e as React.FormEvent);
    }
  };

  const handleBackToHome = () => {
    navigate("/");
  };

  const getLoadingMessage = () => {
    if (loadingStep === "game") return "Creating game...";
    if (loadingStep === "environment") return "Loading...";
    return "Loading...";
  };

  return (
    <div
      className={`bg-transparent text-white min-h-screen flex items-center justify-center font-sans relative z-10 transition-opacity duration-300 ease-in ${isFadedIn ? "opacity-100" : "opacity-0"}`}
    >
      <div className="relative z-[1] flex items-center justify-center w-full min-h-screen">
        <div className="fixed top-[30px] left-[30px]" style={{ zIndex: Z_INDEX.TOP_MENU_BAR }}>
          <BackButton onClick={handleBackToHome} />
        </div>
        <div className="max-w-[600px] w-full px-5 py-10">
          <div className="text-center">
            <h1 className="font-orbitron text-[42px] text-white mb-[60px] text-shadow-glow font-bold tracking-wider">
              Create a new game
            </h1>

            <form onSubmit={handleSubmit} className="relative max-w-[400px] mx-auto">
              <div className="relative flex items-center bg-space-black-darker/95 border border-white/20 rounded-xl p-0 transition-all duration-200 backdrop-blur-space focus-within:border-white/60 focus-within:shadow-[0_0_20px_rgba(255,255,255,0.1)] overflow-hidden">
                <input
                  type="text"
                  value={playerName}
                  onChange={handleInputChange}
                  onKeyDown={handleKeyDown}
                  placeholder="Enter your name"
                  disabled={isLoading}
                  spellCheck={false}
                  autoComplete="off"
                  autoCorrect="off"
                  className="flex-1 bg-transparent border-none py-5 px-6 text-white text-lg outline-none placeholder:text-white/50 disabled:opacity-60"
                  autoFocus
                  maxLength={50}
                />

                <button
                  type="submit"
                  disabled={isLoading || !playerName.trim()}
                  className="bg-transparent border-none py-4 px-5 cursor-pointer flex items-center justify-center transition-all duration-200 disabled:cursor-default disabled:opacity-60 group"
                >
                  <svg
                    width="22"
                    height="22"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="text-white/70 transition-all duration-200 group-hover:text-white group-hover:drop-shadow-[0_0_8px_rgba(255,255,255,0.8)]"
                  >
                    <polyline points="9 6 15 12 9 18" />
                  </svg>
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      {isLoading && <LoadingOverlay isLoaded={false} message={getLoadingMessage()} />}
    </div>
  );
};

export default CreateGamePage;
