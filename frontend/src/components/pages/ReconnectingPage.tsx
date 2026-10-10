import GameButton from "@/components/ui/buttons/GameButton.tsx";
import React, { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { apiService } from "../../services/apiService.ts";
import { clearGameSession, getGameSession, saveGameSession } from "../../utils/sessionStorage.ts";

const ReconnectingPage: React.FC = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isReconnecting, setIsReconnecting] = useState(true);
  const reconnectionAttempted = useRef(false);

  useEffect(() => {
    const attemptReconnection = async () => {
      // Prevent multiple reconnection attempts (React Strict Mode protection)
      if (reconnectionAttempted.current) {
        return;
      }
      reconnectionAttempted.current = true;

      try {
        setError(null);
        setIsReconnecting(true);

        // Check localStorage for saved game data
        const savedGameData = getGameSession();
        if (!savedGameData) {
          // No saved data, clear storage and go to landing page
          clearGameSession();
          navigate("/", { replace: true });
          return;
        }

        const { gameId, playerId, playerName } = savedGameData;
        if (!gameId || !playerName || !playerId) {
          // Invalid data, clear storage and go to landing page
          clearGameSession();
          navigate("/", { replace: true });
          return;
        }

        // Verify game still exists
        const game = await apiService.getGame(gameId);
        if (!game) {
          // Game no longer exists, automatically clear storage and redirect
          clearGameSession();
          navigate("/", { replace: true });
          return;
        }

        // Ensure WebSocket is ready and attempt reconnection
        // playerConnect sends the reconnect message, game state updates come via WebSocket
        await globalWebSocketManager.playerConnect(playerName, gameId, playerId);

        // CRITICAL FIX: Set the current player ID in globalWebSocketManager
        // This ensures the GameInterface component knows which player this client represents
        globalWebSocketManager.setCurrentPlayerId(playerId);

        // Update localStorage with fresh data
        saveGameSession({
          gameId: game.id,
          playerId: playerId,
          playerName: playerName,
        });

        // Navigate to game with reconnected state
        // Game state will be updated reactively via WebSocket game-updated events
        navigate("/game", {
          state: {
            game: game,
            playerId: playerId,
            playerName: playerName,
            isReconnection: true,
          },
          replace: true,
        });
      } catch (error: any) {
        console.error("Reconnection failed:", error);
        setError(error.message || "Reconnection failed");
      } finally {
        setIsReconnecting(false);
      }
    };

    attemptReconnection();
  }, [navigate]);

  const handleReturnToMenu = () => {
    clearGameSession();
    navigate("/", { replace: true });
  };

  return (
    <div className="bg-space-black text-white min-h-dvh flex items-center justify-center font-sans">
      <div className="max-w-[600px] w-full py-10 px-5 max-[768px]:py-5 max-[768px]:px-[15px]">
        <div className="text-center">
          <h1 className="text-5xl text-white mb-[60px] text-shadow-glow-strong font-bold font-orbitron max-[768px]:text-4xl max-[768px]:mb-10 max-[480px]:text-[28px] max-[480px]:mb-[30px]">
            Open Mars
          </h1>

          {isReconnecting ? (
            <div className="flex flex-col items-center gap-6">
              <div className="w-[60px] h-[60px] border-4 border-space-blue-500 border-t-space-blue-900 rounded-full animate-spin shadow-glow max-[768px]:w-[50px] max-[768px]:h-[50px]"></div>
              <h2 className="text-[32px] text-space-blue-900 m-0 text-shadow-glow-strong font-bold font-orbitron max-[768px]:text-2xl max-[480px]:text-xl">
                Reconnecting to game...
              </h2>
              <p className="text-lg text-white/80 m-0 max-w-[400px] max-[768px]:text-base">
                Please wait while we restore your connection
              </p>
            </div>
          ) : error ? (
            <div className="flex flex-col items-center gap-5">
              <div className="text-[64px] mb-2">!</div>
              <h2 className="text-[28px] text-error-red m-0 text-shadow-glow-strong font-bold font-orbitron max-[768px]:text-2xl max-[480px]:text-xl">
                Reconnection Failed
              </h2>
              <p className="text-base text-white/90 m-0 max-w-[400px] leading-[1.5] max-[768px]:text-sm">
                {error}
              </p>
              <GameButton
                emphasis="quiet"
                className="bg-space-blue-600 border-2 border-space-blue-900 rounded-none py-4 px-8 text-lg font-bold text-white cursor-pointer transition-all duration-300 shadow-glow mt-4 hover:bg-space-blue-900 hover:-translate-y-0.5 hover:shadow-glow-lg max-[768px]:py-3.5 max-[768px]:px-7 max-[768px]:text-base max-[480px]:py-3 max-[480px]:px-6 max-[480px]:text-sm font-orbitron"
                onClick={handleReturnToMenu}
              >
                Return to Main Menu
              </GameButton>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
};

export default ReconnectingPage;
