import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { apiService } from "../../services/apiService";
import { globalWebSocketManager } from "../../services/globalWebSocketManager";
import { GameSettingsDto } from "../../types/generated/api-types.ts";
import { skyboxCache } from "../../services/SkyboxCache.ts";
import { saveGameSession } from "../../utils/sessionStorage.ts";
import LoadingOverlay from "../game/view/LoadingOverlay.tsx";
import GameIcon from "../ui/display/GameIcon.tsx";
import InfoTooltip from "../ui/display/InfoTooltip.tsx";
import { useNotifications } from "../../contexts/NotificationContext.tsx";

const CreateGamePage: React.FC = () => {
  const navigate = useNavigate();
  const { showNotification } = useNotifications();
  const [playerName, setPlayerName] = useState("");
  const [developmentMode, setDevelopmentMode] = useState(true);
  const [maxPlayers, setMaxPlayers] = useState(4);
  const [venusNextEnabled, setVenusEnabled] = useState(false);
  const [selectedPacks, setSelectedPacks] = useState<string[]>(["base-game"]);
  const [isLoading, setIsLoading] = useState(false);
  const [loadingStep, setLoadingStep] = useState<"game" | "environment" | null>(null);
  const [skyboxReady, setSkyboxReady] = useState(false);
  const [isFadedIn, setIsFadedIn] = useState(false);
  const [showMore, setShowMore] = useState(false);
  const [isCreatingDemo, setIsCreatingDemo] = useState(false);
  const [claudeApiKey, setClaudeApiKey] = useState("");

  // Check if skybox is already loaded on component mount
  useEffect(() => {
    if (skyboxCache.isReady()) {
      setSkyboxReady(true);
    }
    // Trigger fade in animation
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
      // Step 1: Create game
      const gameSettings: GameSettingsDto = {
        maxPlayers: maxPlayers,
        venusNextEnabled: venusNextEnabled,
        developmentMode: developmentMode,
        cardPacks: selectedPacks,
        demoGame: false,
        hasClaudeApiKey: !!claudeApiKey.trim(),
        availablePlayerColors: [],
      };

      const game = await apiService.createGame(gameSettings, claudeApiKey.trim() || undefined);

      // Step 2: Load 3D environment if not already loaded
      if (!skyboxReady) {
        setLoadingStep("environment");
        await skyboxCache.preload();
      }

      // Step 3: Ensure WebSocket is connected BEFORE setting up listener
      setLoadingStep("game");
      await globalWebSocketManager.initialize();

      // Step 4: Set up one-time listener for game-updated event
      const handleGameUpdated = (gameData: any) => {
        // Extract player info from game data
        const allPlayers = [gameData.currentPlayer, ...(gameData.otherPlayers || [])].filter(
          Boolean,
        );

        const connectedPlayer = allPlayers.find((p: any) => p.name === playerName.trim());

        if (connectedPlayer) {
          // Store game data
          const storedData = {
            gameId: gameData.id,
            playerId: connectedPlayer.id,
            playerName: playerName.trim(),
            createdAt: new Date().toISOString(),
          };
          localStorage.setItem("terraforming-mars-game", JSON.stringify(storedData));

          // Navigate to the main game interface with the complete game state
          navigate("/game", {
            state: {
              game: gameData,
              playerId: connectedPlayer.id,
              playerName: playerName.trim(),
            },
          });

          // Clean up listener
          globalWebSocketManager.off("game-updated", handleGameUpdated);
        }
      };

      // Register listener BEFORE sending connect message
      globalWebSocketManager.on("game-updated", handleGameUpdated);

      // Step 5: Connect player to the game via WebSocket (non-blocking)
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

  const handlePackToggle = (pack: string) => {
    setSelectedPacks((prev) => {
      if (prev.includes(pack)) {
        if (prev.length === 1) {
          return prev;
        }
        return prev.filter((p) => p !== pack);
      } else {
        return [...prev, pack];
      }
    });
  };

  const handleDemoGame = async () => {
    if (isCreatingDemo) return;

    setIsCreatingDemo(true);

    try {
      const result = await apiService.createDemoLobby({
        playerCount: 5,
        playerName: "You",
      });

      saveGameSession({
        gameId: result.game.id,
        playerId: result.playerId,
        playerName: "You",
      });

      await globalWebSocketManager.initialize();
      await globalWebSocketManager.playerConnect("You", result.game.id, result.playerId);

      navigate("/game", {
        state: {
          game: result.game,
          playerId: result.playerId,
          playerName: "You",
        },
      });
    } catch (err) {
      showNotification({
        message: err instanceof Error ? err.message : "Failed to create demo game",
        type: "error",
      });
      setIsCreatingDemo(false);
    }
  };

  const getLoadingMessage = () => {
    if (loadingStep === "game") return "Creating game...";
    if (loadingStep === "environment") return "Loading 3D environment...";
    return "Loading...";
  };

  return (
    <div
      className={`bg-transparent text-white min-h-screen flex items-center justify-center font-sans relative z-10 transition-opacity duration-300 ease-in ${isFadedIn ? "opacity-100" : "opacity-0"}`}
    >
      <div className="relative z-[1] flex items-center justify-center w-full min-h-screen">
        <button
          onClick={handleBackToHome}
          className="fixed top-[30px] left-[30px] bg-space-black-darker/80 border border-white/20 rounded-lg py-2.5 px-4 text-white text-sm cursor-pointer hover:bg-white/20 transition-colors backdrop-blur-space z-[100]"
        >
          ← Back
        </button>
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
                  <div className="w-4 h-6 brightness-0 invert transition-all duration-200 group-hover:drop-shadow-[0_0_8px_rgba(255,255,255,0.8)] group-hover:scale-110">
                    <GameIcon iconType="arrow" size="small" />
                  </div>
                </button>
              </div>

              <div className="flex items-center justify-center gap-6 mt-3">
                <button
                  type="button"
                  onClick={() => setShowMore(!showMore)}
                  className="text-white/50 text-sm py-1 px-3 cursor-pointer hover:text-white/70 transition-colors bg-transparent border-none"
                >
                  Settings
                </button>
                <button
                  type="button"
                  onClick={() => void handleDemoGame()}
                  disabled={isCreatingDemo}
                  className="text-white/50 text-sm py-1 px-3 cursor-pointer hover:text-white/70 transition-colors bg-transparent border-none disabled:opacity-50 disabled:cursor-default"
                >
                  {isCreatingDemo ? "Creating..." : "Demo game"}
                </button>
              </div>

              <div
                className={`absolute left-0 right-0 top-full mt-1 z-10 bg-space-black-darker/95 border border-white/20 rounded-xl p-4 transition-all duration-300 ${showMore ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-2 pointer-events-none"}`}
              >
                <div className="mb-4">
                  <h3 className="text-white text-sm font-semibold mb-3 text-center">Settings</h3>
                  <label className="flex items-center gap-3 py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                    <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2">
                      Max Players
                    </span>
                    <input
                      type="number"
                      min={1}
                      max={10}
                      value={maxPlayers}
                      onChange={(e) => {
                        const val = parseInt(e.target.value, 10);
                        if (val >= 1 && val <= 10) setMaxPlayers(val);
                      }}
                      disabled={isLoading}
                      className="w-16 bg-black/50 border border-white/20 rounded-lg py-1 px-2 text-white text-sm text-center outline-none focus:border-white/60 transition-colors disabled:opacity-60 disabled:cursor-default"
                    />
                  </label>
                  <label className="flex items-center gap-3 cursor-pointer py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                    <input
                      type="checkbox"
                      checked={developmentMode}
                      onChange={(e) => setDevelopmentMode(e.target.checked)}
                      disabled={isLoading}
                      className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-default"
                    />
                    <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2">
                      Development Mode
                      <InfoTooltip size="medium">
                        Enable admin commands for debugging and testing. Allows you to give cards to
                        players, modify resources/production, change game phases, and adjust global
                        parameters through the debug panel.
                      </InfoTooltip>
                    </span>
                  </label>
                </div>

                <div>
                  <h3 className="text-white text-sm font-semibold mb-3 text-center">Card Packs</h3>
                  <div className="flex flex-col gap-2">
                    <label className="flex items-center gap-3 cursor-pointer py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                      <input
                        type="checkbox"
                        checked={selectedPacks.includes("base-game")}
                        onChange={() => handlePackToggle("base-game")}
                        disabled={
                          isLoading ||
                          (selectedPacks.includes("base-game") && selectedPacks.length === 1)
                        }
                        className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-default"
                      />
                      <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2 flex-1">
                        Base Game
                        <span className="text-white/50 text-xs">(22 cards)</span>
                        <InfoTooltip size="small">
                          Includes tested cards with comprehensive test coverage. All cards have
                          verified implementations.
                        </InfoTooltip>
                      </span>
                    </label>

                    <label className="flex items-center gap-3 cursor-pointer py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                      <input
                        type="checkbox"
                        checked={selectedPacks.includes("prelude")}
                        onChange={() => handlePackToggle("prelude")}
                        disabled={
                          isLoading ||
                          (selectedPacks.includes("prelude") && selectedPacks.length === 1)
                        }
                        className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-default"
                      />
                      <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2 flex-1">
                        Prelude
                        <span className="text-white/50 text-xs">(35 cards)</span>
                        <InfoTooltip size="small">
                          Each player receives 4 prelude cards and keeps 2 for free. These give an
                          early boost with resources, production, or other effects.
                        </InfoTooltip>
                      </span>
                    </label>

                    <label className="flex items-center gap-3 cursor-pointer py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                      <input
                        type="checkbox"
                        checked={venusNextEnabled}
                        onChange={(e) => setVenusEnabled(e.target.checked)}
                        disabled={isLoading}
                        className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-default"
                      />
                      <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2 flex-1">
                        Venus Next
                        <InfoTooltip size="small">
                          Adds the Venus globe with tile placements and the Venus global parameter
                          track. Venus Next expansion cards are automatically included.
                        </InfoTooltip>
                      </span>
                    </label>

                    <label className="flex items-center gap-3 cursor-pointer py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                      <input
                        type="checkbox"
                        checked={selectedPacks.includes("experimental")}
                        onChange={() => handlePackToggle("experimental")}
                        disabled={
                          isLoading ||
                          (selectedPacks.includes("experimental") && selectedPacks.length === 1)
                        }
                        className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-default"
                      />
                      <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2 flex-1">
                        Experimental
                        <span className="text-white/50 text-xs">(4 cards)</span>
                        <InfoTooltip size="small">
                          Experimental cards with new mechanics: extra actions, bonus tags, special
                          tiles, and tile destruction.
                        </InfoTooltip>
                      </span>
                    </label>

                    <label className="flex items-center gap-3 cursor-pointer py-2 px-2 rounded hover:bg-white/5 transition-all duration-200">
                      <input
                        type="checkbox"
                        checked={selectedPacks.includes("future")}
                        onChange={() => handlePackToggle("future")}
                        disabled={
                          isLoading ||
                          (selectedPacks.includes("future") && selectedPacks.length === 1)
                        }
                        className="w-[18px] h-[18px] accent-space-blue-solid cursor-pointer m-0 disabled:opacity-60 disabled:cursor-default"
                      />
                      <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2 flex-1">
                        Future Content
                        <span className="text-white/50 text-xs">(431 cards)</span>
                        <InfoTooltip size="small">
                          Includes complex and untested cards for future implementation. May have
                          incomplete effects or bugs.
                        </InfoTooltip>
                      </span>
                    </label>
                  </div>
                </div>

                <div className="mt-4 pt-4 border-t border-white/10">
                  <h3 className="text-white text-sm font-semibold mb-3 text-center">AI Bot</h3>
                  <label className="flex flex-col gap-2 py-2 px-2 rounded">
                    <span className="text-white text-sm font-medium leading-none m-0 flex items-center gap-2">
                      Claude token
                      <InfoTooltip size="small">
                        Provide your own Claude token to enable AI bot players. The token is sent
                        with the game settings and used server-side for Claude invocations.
                      </InfoTooltip>
                    </span>
                    <input
                      type="password"
                      value={claudeApiKey}
                      onChange={(e) => setClaudeApiKey(e.target.value)}
                      placeholder="sk-ant-..."
                      disabled={isLoading}
                      spellCheck={false}
                      autoComplete="off"
                      className="w-full bg-black/50 border border-white/20 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-white/60 transition-colors disabled:opacity-60 disabled:cursor-default placeholder:text-white/30"
                    />
                  </label>
                </div>
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
