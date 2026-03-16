import React, { useEffect, useState, useCallback, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import GameButton from "../ui/buttons/GameButton.tsx";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { useSpaceBackground } from "../../contexts/SpaceBackgroundContext.tsx";
import { useNotifications } from "../../contexts/NotificationContext.tsx";
import { GameDto } from "../../types/generated/api-types.ts";
import { getCorporationLogo } from "../../utils/corporationLogos.tsx";
import { clearGameSession } from "../../utils/sessionStorage.ts";

const FADE_DURATION_MS = 300;

const GameLandingPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { showNotification } = useNotifications();
  const [isFadingOut, setIsFadingOut] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);
  const { preloadSkybox } = useSpaceBackground();
  const [savedGameData, setSavedGameData] = useState<{
    game: GameDto;
    playerId: string;
    playerName: string;
    isSpectator?: boolean;
  } | null>(null);
  const [isDismissing, setIsDismissing] = useState(false);
  const reconnectCardRef = useRef<HTMLDivElement>(null);
  const processedErrorRef = useRef<string | null>(null);

  useEffect(() => {
    const checkExistingGame = async () => {
      try {
        // Preload skybox in parallel with game check
        void preloadSkybox();

        // Check localStorage for existing game
        const savedGameDataString = localStorage.getItem("terraforming-mars-game");
        if (savedGameDataString) {
          const { gameId, playerId, playerName, isSpectator } = JSON.parse(savedGameDataString);

          if (gameId && (playerId || isSpectator)) {
            const game = await apiService.getGame(gameId, isSpectator ? undefined : playerId);
            if (!game) {
              throw new Error("Saved game not found on server");
            }

            setSavedGameData({
              game: game,
              playerId: playerId,
              playerName: playerName,
              isSpectator: isSpectator,
            });
          }
        }
      } catch {
        // Clear invalid saved game data
        clearGameSession();
        setSavedGameData(null);
      }
    };

    void checkExistingGame();
  }, [preloadSkybox]);

  useEffect(() => {
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  useEffect(() => {
    const state = location.state as { error?: string; persistent?: boolean } | null;
    if (state?.error && processedErrorRef.current !== state.error) {
      processedErrorRef.current = state.error;
      showNotification({
        message: state.error,
        type: "error",
        duration: state.persistent ? 0 : undefined,
      });
      window.history.replaceState({}, document.title);
    }
  }, [location.state, showNotification]);

  // Factory for creating fade-out navigation handlers
  const createFadeNavigate = useCallback(
    (path: string) => (e: React.MouseEvent<HTMLAnchorElement>) => {
      // Allow CTRL+Click, CMD+Click, and middle mouse button to open in new tab
      if (e.ctrlKey || e.metaKey || e.button === 1) {
        return;
      }
      // For normal clicks, prevent default and use fade-out animation
      e.preventDefault();
      setIsFadingOut(true);
      setTimeout(() => {
        navigate(path);
      }, FADE_DURATION_MS);
    },
    [navigate],
  );

  const handleCreateGame = createFadeNavigate("/create");
  const handleJoinGame = createFadeNavigate("/join");

  const handleReconnect = async () => {
    if (!savedGameData) return;

    const isSpectator = !!savedGameData.isSpectator;

    setIsFadingOut(true);
    setTimeout(async () => {
      try {
        // Verify game still exists before attempting reconnection
        const game = await apiService.getGame(savedGameData.game.id);
        if (!game) {
          clearGameSession();
          showNotification({ message: "Game no longer exists", type: "error" });
          setIsFadingOut(false);
          setSavedGameData(null);
          return;
        }

        if (isSpectator) {
          navigate(`/game/${savedGameData.game.id}`, {
            state: { spectatorName: savedGameData.playerName },
          });
        } else {
          await globalWebSocketManager.playerConnect(
            savedGameData.playerName,
            savedGameData.game.id,
            savedGameData.playerId,
          );

          navigate("/game", {
            state: {
              game: savedGameData.game,
              playerId: savedGameData.playerId,
              playerName: savedGameData.playerName,
            },
          });
        }
      } catch {
        showNotification({ message: "Failed to reconnect to game", type: "error" });
        setIsFadingOut(false);
      }
    }, FADE_DURATION_MS);
  };

  const handleDismiss = () => {
    setIsDismissing(true);
  };

  const handleDismissTransitionEnd = () => {
    if (isDismissing) {
      clearGameSession();
      globalWebSocketManager.disconnect();
      setSavedGameData(null);
      setIsDismissing(false);
    }
  };

  return (
    <div
      className={`min-h-screen text-white font-sans transition-opacity duration-300 ease-out relative z-10 ${isFadingOut || !isFadedIn ? "opacity-0" : "opacity-100"}`}
    >
      <div className="relative z-[1] w-full min-h-screen flex flex-col items-center justify-center">
        <div className="text-center px-5 py-5">
          <h1 className="font-orbitron text-[56px] text-white mb-[60px] text-shadow-glow-strong font-bold tracking-wider-2xl text-center mx-auto leading-tight">
            TERRAFORMING
            <br />
            MARS
          </h1>

          <div className="flex gap-5 justify-center">
            <GameButton
              as="link"
              to="/create"
              size="lg"
              linkOnClick={handleCreateGame}
              className="px-10 py-5"
            >
              New Game
            </GameButton>

            <GameButton
              as="link"
              to="/join"
              size="lg"
              linkOnClick={handleJoinGame}
              className="px-10 py-5"
            >
              Browse
            </GameButton>
          </div>
        </div>

        <div className="mt-12 flex flex-col items-center gap-5">
          {savedGameData && (
            <div
              ref={reconnectCardRef}
              onTransitionEnd={handleDismissTransitionEnd}
              className={`transition-all duration-300 ${isDismissing ? "opacity-0 translate-y-4" : "opacity-100 translate-y-0"}`}
            >
              <div className="relative w-[500px] bg-space-black-darker/90 border-2 border-space-blue-500 rounded-xl p-8 backdrop-blur-space">
                <button
                  onClick={handleDismiss}
                  className="absolute top-3 right-3 w-8 h-8 flex items-center justify-center text-white/40 hover:text-white/80 transition-colors rounded-full hover:bg-white/10"
                >
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                  >
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                </button>

                {(() => {
                  const isLobby = savedGameData.game.currentPhase === "waiting_for_game_start";
                  const isSpectator = !!savedGameData.isSpectator;
                  const playerCount =
                    (savedGameData.game.currentPlayer ? 1 : 0) +
                    (savedGameData.game.otherPlayers?.length || 0);

                  const buttonLabel = isSpectator
                    ? "RETURN AS SPECTATOR"
                    : isLobby
                      ? "RETURN TO LOBBY"
                      : "RECONNECT";

                  return (
                    <>
                      {!isLobby && !isSpectator && (
                        <div className="mb-6 flex justify-center">
                          {savedGameData.game.currentPlayer?.corporation ? (
                            getCorporationLogo(
                              savedGameData.game.currentPlayer.corporation.name.toLowerCase(),
                            )
                          ) : (
                            <div className="text-white/60 text-sm italic">No Corporation</div>
                          )}
                        </div>
                      )}

                      <div className="flex justify-center gap-6 mb-4 text-white/90 text-base">
                        <div className="flex items-center gap-2">
                          {isLobby ? (
                            <>
                              <span className="font-semibold">
                                {playerCount}/{savedGameData.game.settings.maxPlayers}
                              </span>
                              <span className="text-white/70">Players</span>
                            </>
                          ) : (
                            <>
                              <span className="font-semibold">{playerCount}</span>
                              <span className="text-white/70">
                                {playerCount === 1 ? "Player" : "Players"}
                              </span>
                            </>
                          )}
                        </div>
                        {!isLobby && (
                          <>
                            <div className="text-white/40">•</div>
                            <div className="flex items-center gap-2">
                              <span className="text-white/70">Generation</span>
                              <span className="font-semibold">{savedGameData.game.generation}</span>
                            </div>
                          </>
                        )}
                      </div>

                      <GameButton
                        size="lg"
                        onClick={() => void handleReconnect()}
                        className="w-full"
                      >
                        {buttonLabel}
                      </GameButton>
                    </>
                  );
                })()}
              </div>
            </div>
          )}
        </div>

        <GameButton
          as="link"
          to="/cards"
          buttonType="secondary"
          size="sm"
          className="fixed bottom-[30px] right-[30px]"
        >
          View Cards
        </GameButton>
      </div>
    </div>
  );
};

export default GameLandingPage;
