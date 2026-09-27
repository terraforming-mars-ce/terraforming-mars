import { CARD_PACKS, VENUS_PACK } from "@/constants/cardPacks.ts";
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

const displayName = (id: string) =>
  id.replace(/[-_]+/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
const packLabels = new Map([...CARD_PACKS, VENUS_PACK].map((pack) => [pack.id, pack.label]));

const UUID_V4_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

const JoinGamePage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { showNotification } = useNotifications();
  const [availableGames, setAvailableGames] = useState<GameDto[]>([]);
  const [mapNames, setMapNames] = useState<Record<string, string>>({});
  const [isLoadingGames, setIsLoadingGames] = useState(false);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [initialCode, setInitialCode] = useState<string | undefined>(undefined);
  const [isFadedIn, setIsFadedIn] = useState(false);

  const [showEnterCodePopover, setShowEnterCodePopover] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [loadError, setLoadError] = useState("");
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
    setLoadError("");
    try {
      const games = await apiService.listGames();
      const lobbyGames = games.filter((g) => g.status === "lobby");
      const activeGames = games.filter((g) => g.status === "active");
      setAvailableGames([...lobbyGames, ...activeGames]);

      setIsInitialLoad(false);
    } catch {
      setLoadError("Could not load games. Try refreshing.");
    } finally {
      setIsInitialLoad(false);
      setIsLoadingGames(false);
    }
  };

  useEffect(() => {
    void fetchGames();
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void apiService
      .getGameOptions(controller.signal)
      .then((options) => {
        if (!controller.signal.aborted) {
          setMapNames(Object.fromEntries(options.availableMaps.map((map) => [map.id, map.name])));
        }
      })
      .catch(() => {});
    return () => controller.abort();
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

  const filteredGames = availableGames.filter((game) => {
    const names = [
      game.currentPlayer?.name,
      ...(game.otherPlayers ?? []).map((player) => player.name),
    ];
    return names.some((name) => name?.toLowerCase().includes(searchQuery.trim().toLowerCase()));
  });

  return (
    <div
      className={`text-white transition-opacity duration-200 ${isFadedIn ? "opacity-100" : "opacity-0"}`}
    >
      <div className="menu-shell relative" style={{ zIndex: Z_INDEX.UI_BASE }}>
        <div className="fixed top-[30px] left-[30px]" style={{ zIndex: Z_INDEX.TOP_MENU_BAR }}>
          <BackButton onClick={handleBackToHome} />
        </div>
        <div className="w-full">
          <div className="text-left">
            <h1 className="menu-title">Browse games</h1>

            <div className="game-panel p-5 sm:p-8">
              <div className="flex flex-wrap items-center gap-3 mb-6">
                {(availableGames.length > 0 || searchQuery) && (
                  <div className="relative flex-1 min-w-[180px]">
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
                      className="w-full bg-space-black-darker/80 border border-white/20 rounded-none py-2 pl-10 pr-3 text-white text-sm outline-none placeholder:text-white/40 focus:border-white/40 transition-colors backdrop-blur-space"
                    />
                  </div>
                )}
                <GameButton as="link" to="/create" className="mr-auto">
                  Create game
                </GameButton>
                <GameButton
                  ref={enterCodeButtonRef}
                  emphasis="secondary"
                  size="sm"
                  onClick={() => setShowEnterCodePopover(true)}
                  className="shrink-0"
                >
                  Join with code
                </GameButton>
                <GameButton
                  emphasis="secondary"
                  size="sm"
                  aria-label="Refresh games"
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

              <div className="min-h-[240px] max-h-[55vh] overflow-y-auto">
                {loadError ? (
                  <p role="alert" className="text-red-300 py-8">
                    {loadError}
                  </p>
                ) : isInitialLoad && isLoadingGames ? (
                  <div className="text-white/50 text-sm text-center py-8">Loading games...</div>
                ) : availableGames.length === 0 && !isLoadingGames ? (
                  <div className="text-white/50 text-sm text-center py-8">No public games yet.</div>
                ) : (
                  <div className="flex flex-col gap-3">
                    {filteredGames.length === 0 && (
                      <p className="text-white/60 py-8">No matching games.</p>
                    )}
                    {filteredGames.map((game) => {
                      const playerCount =
                        (game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0);
                      const maxPlayers = game.settings?.maxPlayers || 10;
                      const hostName =
                        game.currentPlayer?.name || game.otherPlayers?.[0]?.name || "Unknown";
                      const isActive = game.status === "active";
                      return (
                        <div
                          key={game.id}
                          className="flex flex-wrap items-center justify-between gap-4 border-b border-white/10 py-5"
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
                            <span className="text-white/50 text-xs">
                              {mapNames[game.settings.mapId] ?? displayName(game.settings.mapId)} ·{" "}
                              {(game.settings.cardPacks ?? [])
                                .map((id) => packLabels.get(id) ?? displayName(id))
                                .join(", ")}
                            </span>
                          </div>
                          <div className="flex gap-2 shrink-0">
                            <GameButton
                              ref={(el) => {
                                spectateButtonRefs.current.set(game.id, el);
                              }}
                              emphasis="secondary"
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
                                disabled={playerCount >= maxPlayers}
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
