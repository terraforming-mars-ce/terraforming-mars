import React, { useEffect, useState, useCallback, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import GameButton from "../ui/buttons/GameButton.tsx";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager.ts";
import { useNotifications } from "../../contexts/NotificationContext.tsx";
import { GameDto } from "../../types/generated/api-types.ts";
import { getCorporationLogo } from "../../utils/corporationLogos.tsx";
import { clearGameSession, getGameSession } from "../../utils/sessionStorage.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import OpenMarsLogo from "@/components/ui/OpenMarsLogo.tsx";

const FADE_DURATION_MS = 300;

const GameLandingPage: React.FC = () => {
  const { isCompact } = useLayoutMode();
  const navigate = useNavigate();
  const location = useLocation();
  const { showNotification } = useNotifications();
  const [isFadingOut, setIsFadingOut] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);
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
        // Check localStorage for existing game
        const savedGame = getGameSession();
        if (savedGame) {
          const { gameId, playerId, playerName, isSpectator } = savedGame;

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
  }, []);

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
      className={`min-h-dvh text-white font-sans transition-opacity duration-300 ease-out relative ${isFadingOut || !isFadedIn ? "opacity-0" : "opacity-100"}`}
      style={{ zIndex: Z_INDEX.UI_BASE }}
    >
      <div className="menu-shell relative items-start [@media(max-height:500px)]:flex-row [@media(max-height:500px)]:items-center [@media(max-height:500px)]:justify-between [@media(max-height:500px)]:gap-8">
        <div className="text-left py-5 w-[520px] max-w-full compact:w-[340px] [@media(max-height:500px)]:py-0 [@media(max-height:500px)]:shrink-0">
          <h1 className="mb-10 leading-none compact:mb-5 [@media(max-height:500px)]:mb-5">
            <OpenMarsLogo surface="menu" />
          </h1>

          <div className="flex flex-wrap justify-center gap-4">
            <GameButton
              as="link"
              to="/create"
              size="lg"
              onClick={handleCreateGame}
              className="!px-8 compact:!px-5"
            >
              New game
            </GameButton>

            <GameButton
              as="link"
              to="/join"
              emphasis="secondary"
              size="lg"
              onClick={handleJoinGame}
              className="!px-8 compact:!px-5"
            >
              Browse games
            </GameButton>
          </div>
        </div>

        <div className="mt-10 flex flex-col items-start gap-5 max-w-full min-w-0 compact:mt-6 compact:w-full compact:landscape:w-auto [@media(max-height:500px)]:mt-0">
          {savedGameData && (
            <div
              ref={reconnectCardRef}
              onTransitionEnd={handleDismissTransitionEnd}
              className={`transition-all duration-300 ${isDismissing ? "opacity-0 translate-y-4" : "opacity-100 translate-y-0"}`}
            >
              <div className="relative w-[500px] max-w-full game-panel p-6 [@media(max-height:500px)]:w-[360px] [@media(max-height:500px)]:p-4 compact:!w-[300px] compact:!p-3">
                <button
                  onClick={handleDismiss}
                  aria-label="Dismiss"
                  className="absolute top-3 right-3 compact:top-1 compact:right-1 w-8 h-8 flex items-center justify-center text-white/40 hover:text-white/80 transition-colors rounded-full hover:bg-white/10 before:absolute before:-inset-1.5 before:content-['']"
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
                        <div className="mb-6 flex justify-center [@media(max-height:500px)]:mb-3 compact:!mb-2">
                          {savedGameData.game.currentPlayer?.corporation ? (
                            getCorporationLogo(
                              savedGameData.game.currentPlayer.corporation.name.toLowerCase(),
                              "w-[220px] h-[110px] [@media(max-height:500px)]:w-[128px] [@media(max-height:500px)]:h-[64px] compact:!w-[112px] compact:!h-[48px]",
                            )
                          ) : (
                            <div className="text-white/60 text-sm italic">No Corporation</div>
                          )}
                        </div>
                      )}

                      <div className="flex flex-wrap justify-center gap-x-6 gap-y-1 mb-4 text-white/90 text-base [@media(max-height:500px)]:mb-3 compact:!mb-2 compact:text-sm compact:gap-x-4">
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
                        size={isCompact ? "md" : "lg"}
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
      </div>
    </div>
  );
};

export default GameLandingPage;
