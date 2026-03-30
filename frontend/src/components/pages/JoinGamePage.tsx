import React, { useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { GameDto } from "../../types/generated/api-types.ts";
import GameButton from "../ui/buttons/GameButton.tsx";
import BackButton from "../ui/buttons/BackButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";
import EnterCodePopover from "../ui/popover/EnterCodePopover.tsx";
import SpectatePopover from "../ui/popover/SpectatePopover.tsx";
import JoinGamePopover from "../ui/popover/JoinGamePopover.tsx";
import { useNotifications } from "../../contexts/NotificationContext.tsx";

const UUID_V4_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const JoinGamePage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { showNotification } = useNotifications();
  const [availableGames, setAvailableGames] = useState<GameDto[]>([]);
  const [isLoadingGames, setIsLoadingGames] = useState(false);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [initialCode, setInitialCode] = useState<string | undefined>(undefined);
  const [isFadedIn, setIsFadedIn] = useState(false);

  const [showEnterCodePopover, setShowEnterCodePopover] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [animationKey, setAnimationKey] = useState(0);
  const [spectateGameId, setSpectateGameId] = useState<string | null>(null);
  const [joinGame, setJoinGame] = useState<GameDto | null>(null);
  const enterCodeButtonRef = useRef<HTMLButtonElement>(null);
  const spectateButtonRefs = useRef<Map<string, HTMLButtonElement | null>>(new Map());
  const joinButtonRefs = useRef<Map<string, HTMLButtonElement | null>>(new Map());

  useEffect(() => {
    setTimeout(() => {
      setIsFadedIn(true);
    }, 10);
  }, []);

  const fetchGames = async () => {
    setIsLoadingGames(true);
    try {
      const games = await apiService.listGames();
      const lobbyGames = games.filter((g) => g.status === "lobby");
      const activeGames = games.filter((g) => g.status === "active");
      setAvailableGames([...lobbyGames, ...activeGames]);
      setAnimationKey((prev) => prev + 1);
      setIsInitialLoad(false);
    } catch {
      setAvailableGames([]);
    } finally {
      setIsLoadingGames(false);
    }
  };

  useEffect(() => {
    void fetchGames();
  }, []);

  useEffect(() => {
    const urlParams = new URLSearchParams(location.search);
    const codeParam = urlParams.get("code");

    if (codeParam && UUID_V4_REGEX.test(codeParam)) {
      setInitialCode(codeParam);
      setShowEnterCodePopover(true);
    }
  }, [location.search]);

  const handleBackToHome = () => {
    navigate("/");
  };

  const handleGameValidated = (game: GameDto) => {
    setShowEnterCodePopover(false);
    setJoinGame(game);
  };

  const handleJoinGame = async (game: GameDto) => {
    const existingGame = await apiService.getGame(game.id);
    if (!existingGame) {
      showNotification({ message: "Game no longer exists", type: "info" });
      void fetchGames();
      return;
    }
    setJoinGame(game);
  };

  return (
    <div
      className={`transition-opacity duration-300 ease-in ${isFadedIn ? "opacity-100" : "opacity-0"}`}
    >
      <div className="relative z-[1] flex items-start justify-center w-full min-h-screen pt-[15vh]">
        <div className="fixed top-[30px] left-[30px]" style={{ zIndex: Z_INDEX.TOP_MENU_BAR }}>
          <BackButton onClick={handleBackToHome} />
        </div>
        <div className="max-w-[600px] w-full px-5 py-10">
          <div className="text-center">
            <h1 className="font-orbitron text-[42px] text-white mb-8 text-shadow-glow font-bold tracking-wider">
              Browse games
            </h1>

            <div className="max-w-[500px] mx-auto">
              <div className="flex items-center gap-3 mb-6">
                <div className="relative flex-1">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth={2}
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none z-10"
                  >
                    <circle cx="11" cy="11" r="8" />
                    <path d="M21 21l-4.35-4.35" />
                  </svg>
                  <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search games..."
                    spellCheck={false}
                    autoComplete="off"
                    className="w-full bg-space-black-darker/80 border border-white/20 rounded-lg py-2 pl-10 pr-3 text-white text-sm outline-none placeholder:text-white/40 focus:border-white/40 transition-colors backdrop-blur-space"
                  />
                </div>
                <GameButton
                  ref={enterCodeButtonRef}
                  buttonType="secondary"
                  size="sm"
                  onClick={() => setShowEnterCodePopover(true)}
                  className="shrink-0"
                >
                  Enter code
                </GameButton>
                <GameButton
                  buttonType="secondary"
                  size="sm"
                  onClick={() => void fetchGames()}
                  disabled={isLoadingGames}
                  className={`shrink-0 p-2 transition-opacity duration-300${isLoadingGames ? " opacity-40" : ""}`}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth={2}
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className="w-5 h-5"
                  >
                    <path d="M21 12a9 9 0 1 1-6.22-8.56" />
                    <polyline points="21 3 21 9 15 9" />
                  </svg>
                </GameButton>
              </div>

              <div className="h-[400px] overflow-y-auto">
                {isInitialLoad && isLoadingGames ? (
                  <div className="text-white/50 text-sm text-center py-8">Loading games...</div>
                ) : availableGames.length === 0 && !isLoadingGames ? (
                  <div className="text-white/50 text-sm text-center py-8">
                    No games available. Create a new game or enter a game code.
                  </div>
                ) : (
                  <div key={animationKey} className="flex flex-col gap-3">
                    {availableGames
                      .filter((game) => {
                        if (!searchQuery.trim()) return true;
                        const query = searchQuery.toLowerCase();
                        const hostName = (
                          game.currentPlayer?.name ||
                          game.otherPlayers?.[0]?.name ||
                          ""
                        ).toLowerCase();
                        const playerNames = [
                          ...(game.currentPlayer ? [game.currentPlayer.name] : []),
                          ...(game.otherPlayers?.map((p) => p.name) || []),
                        ];
                        return (
                          hostName.includes(query) ||
                          playerNames.some((n) => n.toLowerCase().includes(query))
                        );
                      })
                      .map((game, index) => {
                        const playerCount =
                          (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0);
                        const maxPlayers = game.settings?.maxPlayers || 10;
                        const hostName =
                          game.currentPlayer?.name || game.otherPlayers?.[0]?.name || "Unknown";
                        const isActive = game.status === "active";
                        return (
                          <div
                            key={game.id}
                            className="flex items-center justify-between bg-space-black-darker/95 border border-white/20 rounded-xl p-4 backdrop-blur-space animate-[slide-down_0.3s_ease-out_both]"
                            style={{ animationDelay: `${index * 60}ms` }}
                          >
                            <div className="flex flex-col gap-1 min-w-0 text-left">
                              <span className="text-white text-sm font-medium truncate">
                                {hostName}
                              </span>
                              <span className="text-white/50 text-xs">
                                {playerCount}/{maxPlayers} Players
                                {isActive && game.generation != null && (
                                  <span className="ml-2 text-white/35">Gen {game.generation}</span>
                                )}
                              </span>
                            </div>
                            <div className="flex gap-2 shrink-0 ml-4">
                              <GameButton
                                ref={(el) => {
                                  spectateButtonRefs.current.set(game.id, el);
                                }}
                                buttonType="secondary"
                                size="sm"
                                onClick={() => setSpectateGameId(game.id)}
                              >
                                Spectate
                              </GameButton>
                              {!isActive && (
                                <GameButton
                                  ref={(el) => {
                                    joinButtonRefs.current.set(game.id, el);
                                  }}
                                  size="sm"
                                  onClick={() => void handleJoinGame(game)}
                                >
                                  Join
                                </GameButton>
                              )}
                            </div>
                          </div>
                        );
                      })}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      {joinGame && (
        <JoinGamePopover
          isVisible={!!joinGame}
          onClose={() => setJoinGame(null)}
          game={joinGame}
          anchorRef={{
            current: joinButtonRefs.current.get(joinGame.id) ?? enterCodeButtonRef.current,
          }}
        />
      )}

      <EnterCodePopover
        isVisible={showEnterCodePopover}
        onClose={() => setShowEnterCodePopover(false)}
        onGameValidated={handleGameValidated}
        initialCode={initialCode}
        anchorRef={enterCodeButtonRef}
      />

      {spectateGameId && (
        <SpectatePopover
          isVisible={!!spectateGameId}
          onClose={() => setSpectateGameId(null)}
          gameId={spectateGameId}
          anchorRef={{ current: spectateButtonRefs.current.get(spectateGameId) ?? null }}
        />
      )}
    </div>
  );
};

export default JoinGamePage;
