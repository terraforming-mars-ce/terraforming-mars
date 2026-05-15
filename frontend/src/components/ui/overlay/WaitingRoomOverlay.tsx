import React, { useRef, useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { GameDto, OtherPlayerDto, PlayerDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";
import GameButton from "../buttons/GameButton.tsx";
import GameMenuModal from "./GameMenuModal.tsx";
import DemoSetupOverlay from "./DemoSetupOverlay.tsx";
import LobbySettingsOverlay from "./LobbySettingsOverlay.tsx";
import LobbyMapInfoPanel from "../lobby/LobbyMapInfoPanel.tsx";
import { BotDifficultyChip, BotSpeedChip } from "../display/BotChips.tsx";
import MainMenuHamburger from "../buttons/MainMenuHamburger.tsx";

interface WaitingRoomOverlayProps {
  game: GameDto;
  playerId: string;
  visible?: boolean;
  onExited?: () => void;
}

interface LeavingPlayer {
  id: string;
  name: string;
}

function getOrderedPlayers<T>(playerMap: Map<string, T>, game: GameDto): T[] {
  const order = [game.turnOrder, game.playerOrder].find((o) => o && o.length > 0);
  if (!order) {
    return Array.from(playerMap.values());
  }
  return order.map((pid) => playerMap.get(pid)).filter((p) => p !== undefined);
}

const ColorPicker: React.FC<{
  colors: string[];
  takenColors: Set<string>;
  currentColor: string;
  onSelect: (color: string) => void;
}> = ({ colors, takenColors, currentColor, onSelect }) => (
  <div
    className="flex flex-wrap gap-1.5"
    style={{ width: colors.length > 5 ? `${5 * 28 + 4 * 6}px` : undefined }}
  >
    {colors.map((color) => {
      const isTaken = takenColors.has(color);
      const isSelected = color === currentColor;
      return (
        <button
          key={color}
          onClick={() => !isTaken && onSelect(color)}
          className={`w-7 h-7 rounded-full border-2 transition-all flex-shrink-0 ${
            isSelected
              ? "border-white scale-110"
              : isTaken
                ? "border-white/10 opacity-25 cursor-default"
                : "border-transparent cursor-pointer hover:scale-110"
          }`}
          style={{ backgroundColor: color }}
          disabled={isTaken}
        />
      );
    })}
  </div>
);

const WaitingRoomOverlay: React.FC<WaitingRoomOverlayProps> = ({
  game,
  playerId,
  visible,
  onExited,
}) => {
  const navigate = useNavigate();
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/game/${game.id}?type=join`;
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false);
  const [leaveConfirmVisible, setLeaveConfirmVisible] = useState(true);
  const [pendingLeave, setPendingLeave] = useState(false);
  const [showBotDropdown, setShowBotDropdown] = useState(false);
  const [colorPickerForPlayer, setColorPickerForPlayer] = useState<string | null>(null);
  const botDropdownRef = useRef<HTMLDivElement>(null);
  const colorPickerRef = useRef<HTMLDivElement>(null);

  const isDemoGame = game.settings.demoGame;
  const [showDemoSetup, setShowDemoSetup] = useState(false);
  const [showLobbySettings, setShowLobbySettings] = useState(false);

  const handleStartGame = () => {
    if (!isHost) return;
    void globalWebSocketManager.startGame();
  };

  const hasCurrentPlayer = !!game.currentPlayer?.id;
  const playerCount = (hasCurrentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0);

  const allBotsReady = React.useMemo(() => {
    const bots: { botStatus?: string }[] = [];
    if (hasCurrentPlayer && game.currentPlayer?.playerType === "bot") bots.push(game.currentPlayer);
    game.otherPlayers?.forEach((p) => {
      if (p.playerType === "bot") bots.push(p);
    });
    return bots.every((b) => b.botStatus === "ready");
  }, [hasCurrentPlayer, game.currentPlayer, game.otherPlayers]);

  const allDemoPlayersReady = React.useMemo(() => {
    if (!isDemoGame) {
      return true;
    }
    const humans: { demoReady?: boolean }[] = [];
    if (hasCurrentPlayer && game.currentPlayer?.playerType !== "bot") {
      humans.push(game.currentPlayer!);
    }
    game.otherPlayers?.forEach((p) => {
      if (p.playerType !== "bot") {
        humans.push(p);
      }
    });
    return humans.every((h) => h.demoReady);
  }, [isDemoGame, hasCurrentPlayer, game.currentPlayer, game.otherPlayers]);

  const allPlayers = React.useMemo(() => {
    const players: { id: string; name: string; playerType: string }[] = [];
    if (hasCurrentPlayer && game.currentPlayer)
      players.push({
        id: game.currentPlayer.id,
        name: game.currentPlayer.name,
        playerType: game.currentPlayer.playerType,
      });
    game.otherPlayers?.forEach((p) =>
      players.push({ id: p.id, name: p.name, playerType: p.playerType }),
    );
    return players;
  }, [hasCurrentPlayer, game.currentPlayer, game.otherPlayers]);

  const takenColorsFor = React.useCallback(
    (targetId: string) => {
      const colors = new Set<string>();
      const addPlayerColor = (p: PlayerDto | OtherPlayerDto) => {
        if (p.color && p.id !== targetId) colors.add(p.color);
      };
      if (hasCurrentPlayer && game.currentPlayer) addPlayerColor(game.currentPlayer);
      game.otherPlayers?.forEach(addPlayerColor);
      return colors;
    },
    [hasCurrentPlayer, game.currentPlayer, game.otherPlayers],
  );

  const getPlayerColor = React.useCallback(
    (pid: string) => {
      if (hasCurrentPlayer && game.currentPlayer?.id === pid)
        return game.currentPlayer?.color || "";
      return game.otherPlayers?.find((p) => p.id === pid)?.color || "";
    },
    [hasCurrentPlayer, game.currentPlayer, game.otherPlayers],
  );

  const handleColorSelect = (color: string) => {
    if (colorPickerForPlayer) {
      const targetId = colorPickerForPlayer === playerId ? undefined : colorPickerForPlayer;
      void globalWebSocketManager.setPlayerColor(color, targetId);
    }
    setColorPickerForPlayer(null);
  };

  const currentPlayerIds = React.useMemo(() => new Set(allPlayers.map((p) => p.id)), [allPlayers]);

  const prevPlayerIdsRef = useRef<Set<string>>(new Set());
  const [newPlayerIds, setNewPlayerIds] = useState<Set<string>>(new Set());
  const [leavingPlayers, setLeavingPlayers] = useState<LeavingPlayer[]>([]);
  const prevPlayersRef = useRef<Map<string, string>>(new Map());

  const handleAnimationEnd = useCallback((id: string) => {
    setNewPlayerIds((prev) => {
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  }, []);

  useEffect(() => {
    const prevIds = prevPlayerIdsRef.current;

    const joined = new Set<string>();
    for (const id of currentPlayerIds) {
      if (!prevIds.has(id)) joined.add(id);
    }

    const left: LeavingPlayer[] = [];
    for (const id of prevIds) {
      if (!currentPlayerIds.has(id)) {
        left.push({ id, name: prevPlayersRef.current.get(id) ?? "Player" });
      }
    }

    if (joined.size > 0) setNewPlayerIds(joined);

    if (left.length > 0) {
      setLeavingPlayers((prev) => [...prev, ...left]);
      setTimeout(() => {
        setLeavingPlayers((prev) => prev.filter((p) => !left.some((l) => l.id === p.id)));
      }, 300);
    }

    prevPlayerIdsRef.current = new Set(currentPlayerIds);
    const nameMap = new Map<string, string>();
    allPlayers.forEach((p) => nameMap.set(p.id, p.name));
    prevPlayersRef.current = nameMap;
  }, [currentPlayerIds, allPlayers]);

  const handleCancelLeave = () => {
    setLeaveConfirmVisible(false);
    setPendingLeave(false);
  };

  const handleConfirmLeave = () => {
    setLeaveConfirmVisible(false);
    setPendingLeave(true);
  };

  const handleLeaveConfirmExited = () => {
    setShowLeaveConfirm(false);
    setLeaveConfirmVisible(true);
    if (pendingLeave) {
      setPendingLeave(false);
      globalWebSocketManager.disconnect();
      void navigate("/");
    }
  };

  const openLeaveConfirm = () => {
    setShowLeaveConfirm(true);
    setLeaveConfirmVisible(true);
  };

  useEffect(() => {
    if (!showBotDropdown && !colorPickerForPlayer) return;
    const handleClick = (e: MouseEvent) => {
      if (
        showBotDropdown &&
        botDropdownRef.current &&
        !botDropdownRef.current.contains(e.target as Node)
      ) {
        setShowBotDropdown(false);
      }
      if (
        colorPickerForPlayer &&
        colorPickerRef.current &&
        !colorPickerRef.current.contains(e.target as Node)
      ) {
        setColorPickerForPlayer(null);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [showBotDropdown, colorPickerForPlayer]);

  const handleAddBot = (difficulty: string, speed: string) => {
    setShowBotDropdown(false);
    void globalWebSocketManager.addBot(undefined, difficulty, speed);
  };

  return (
    <>
      <MainMenuHamburger gameId={game.id} onLeaveGame={openLeaveConfirm} />
      <LobbyMapInfoPanel game={game} playerId={playerId} />

      {/* Leave Confirmation Modal */}
      {showLeaveConfirm && (
        <GameMenuModal
          title="Leave game?"
          subtitle={isHost && playerCount > 1 ? "Another player will become the host" : undefined}
          visible={leaveConfirmVisible}
          onExited={handleLeaveConfirmExited}
          showBackdrop={true}
          zIndex={2000}
          onClose={handleCancelLeave}
        >
          <div className="flex gap-3 justify-center">
            <GameButton buttonType="secondary" size="sm" onClick={handleCancelLeave}>
              Cancel
            </GameButton>
            <GameButton variant="error" size="sm" onClick={handleConfirmLeave}>
              Leave
            </GameButton>
          </div>
        </GameMenuModal>
      )}

      <GameMenuModal
        title="Game Lobby"
        subtitle={`${playerCount} player${playerCount !== 1 ? "s" : ""} joined`}
        onBack={() => void navigate("/")}
        visible={visible}
        onExited={onExited}
      >
        {/* Leave Button - positioned at top-left of modal content */}
        <GameButton
          buttonType="textonly"
          size="sm"
          onClick={openLeaveConfirm}
          className="absolute top-8 left-8 !p-2.5 hover:text-red-400"
        >
          <svg
            width="22"
            height="22"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="-scale-x-100"
          >
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
        </GameButton>

        {/* Settings Button - positioned at top-right of modal content (host only) */}
        {isHost && (
          <GameButton
            buttonType="textonly"
            size="sm"
            onClick={() => setShowLobbySettings(true)}
            className="absolute top-8 right-8 !p-2.5 hover:text-space-blue-300"
            aria-label="Lobby settings"
          >
            <svg
              width="22"
              height="22"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </GameButton>
        )}

        <style>{`
          @keyframes playerSlideIn {
            from { opacity: 0; transform: translateY(-8px) scale(0.95); max-height: 0; }
            to { opacity: 1; transform: translateY(0) scale(1); max-height: 60px; }
          }
          @keyframes playerSlideOut {
            from { opacity: 1; transform: translateY(0) scale(1); max-height: 60px; }
            to { opacity: 0; transform: translateY(-8px) scale(0.95); max-height: 0; padding: 0; margin: 0; }
          }
        `}</style>

        {/* Player List */}
        <div className="mb-6">
          <h3 className="text-white text-sm font-semibold mb-2 uppercase tracking-wide">Players</h3>
          <div className="flex flex-col gap-2">
            {(() => {
              const playerMap = new Map();
              if (hasCurrentPlayer && game.currentPlayer) {
                playerMap.set(game.currentPlayer.id, game.currentPlayer);
              }
              game.otherPlayers?.forEach((otherPlayer) => {
                playerMap.set(otherPlayer.id, otherPlayer);
              });

              const orderedPlayers = getOrderedPlayers(playerMap, game);

              const playerItems = orderedPlayers.map((player) => ({
                id: player.id,
                name: player.name,
                color: player.color || "",
                playerType: player.playerType as string,
                botStatus: (player.botStatus as string) || undefined,
                botDifficulty: (player.botDifficulty as string) || undefined,
                botSpeed: (player.botSpeed as string) || undefined,
                demoReady: (player as PlayerDto | OtherPlayerDto).demoReady || false,
                isLeaving: false,
              }));

              leavingPlayers.forEach((lp) => {
                if (!playerMap.has(lp.id)) {
                  playerItems.push({
                    id: lp.id,
                    name: lp.name,
                    color: "",
                    playerType: "human",
                    botStatus: undefined,
                    botDifficulty: undefined,
                    botSpeed: undefined,
                    demoReady: false,
                    isLeaving: true,
                  });
                }
              });

              return playerItems.map((player) => {
                let animClass = "";
                if (player.isLeaving) {
                  animClass = "animate-[playerSlideOut_0.3s_ease-out_forwards] overflow-hidden";
                } else if (newPlayerIds.has(player.id)) {
                  animClass = "animate-[playerSlideIn_0.3s_ease-out]";
                }

                const isCurrentPlayer = player.id === playerId;
                const canEditColor =
                  (isCurrentPlayer || (isHost && player.playerType === "bot")) && !player.isLeaving;
                const showingPicker = colorPickerForPlayer === player.id;

                return (
                  <div
                    key={player.id}
                    className={`relative flex justify-between items-center py-2 px-3 bg-black/40 rounded-lg border border-space-blue-600/50 ${animClass}`}
                    onAnimationEnd={
                      !player.isLeaving && newPlayerIds.has(player.id)
                        ? () => handleAnimationEnd(player.id)
                        : undefined
                    }
                  >
                    <div className="flex gap-2 items-center">
                      <div
                        className="relative flex items-center"
                        ref={showingPicker ? colorPickerRef : undefined}
                      >
                        <button
                          className={`w-4 h-4 rounded-full transition-all flex-shrink-0 p-0 leading-none ${
                            canEditColor ? "cursor-pointer hover:scale-125" : "cursor-default"
                          }`}
                          style={{ backgroundColor: player.color || "#555" }}
                          onClick={() =>
                            canEditColor &&
                            setColorPickerForPlayer((v) => (v === player.id ? null : player.id))
                          }
                          disabled={!canEditColor}
                        />
                        {showingPicker && game.settings.availablePlayerColors && (
                          <div className="absolute top-full left-0 mt-1 bg-black border border-white/20 rounded-lg p-3 z-20">
                            <ColorPicker
                              colors={game.settings.availablePlayerColors}
                              takenColors={takenColorsFor(player.id)}
                              currentColor={getPlayerColor(player.id)}
                              onSelect={handleColorSelect}
                            />
                          </div>
                        )}
                      </div>
                      <span className="text-white text-sm font-medium">{player.name}</span>
                    </div>
                    <div className="flex gap-1.5 items-center">
                      {player.id === playerId && (
                        <span className="bg-space-blue-800 text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                          You
                        </span>
                      )}
                      {game.hostPlayerId === player.id && (
                        <span className="bg-gradient-to-br from-[#ffa500] to-[#ff8c00] text-white py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                          Host
                        </span>
                      )}
                      {player.playerType === "bot" && (
                        <>
                          <BotDifficultyChip
                            difficulty={player.botDifficulty}
                            botStatus={player.botStatus}
                            showStatusIcon
                          />
                          <BotSpeedChip speed={player.botSpeed} />
                        </>
                      )}
                      {isDemoGame &&
                        player.playerType !== "bot" &&
                        !player.isLeaving &&
                        (player.demoReady ? (
                          <span className="bg-green-700/60 text-green-300 py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                            Ready
                          </span>
                        ) : (
                          <span className="bg-red-900/40 text-red-300/70 py-0.5 px-1.5 rounded text-[10px] font-bold uppercase">
                            Not Ready
                          </span>
                        ))}
                      {isHost && player.id !== playerId && !player.isLeaving && (
                        <button
                          onClick={() => void globalWebSocketManager.kickPlayer(player.id)}
                          className="ml-1 text-red-400 hover:text-red-300 transition-colors cursor-pointer"
                          title={`Kick ${player.name}`}
                        >
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          >
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                          </svg>
                        </button>
                      )}
                    </div>
                  </div>
                );
              });
            })()}
          </div>

          {/* Spectators Section */}
          {game.spectators && game.spectators.length > 0 && (
            <div className="mt-4">
              <h3 className="text-white/60 text-xs font-semibold mb-2 uppercase tracking-wide">
                Spectators
              </h3>
              <div className="flex flex-col gap-1.5">
                {game.spectators.map((spectator) => (
                  <div
                    key={spectator.id}
                    className="flex justify-between items-center py-1.5 px-3 bg-black/25 rounded-lg border border-white/10"
                  >
                    <span className="text-white/70 text-sm">{spectator.name}</span>
                    <div className="flex gap-1.5 items-center">
                      {isHost && (
                        <button
                          onClick={() => void globalWebSocketManager.kickSpectator(spectator.id)}
                          className="ml-1 text-red-400/60 hover:text-red-300 transition-colors cursor-pointer"
                        >
                          <svg
                            width="12"
                            height="12"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                          >
                            <line x1="18" y1="6" x2="6" y2="18" />
                            <line x1="6" y1="6" x2="18" y2="18" />
                          </svg>
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Join Link & Add Bot */}
          <div className="mt-4 flex justify-center gap-2">
            <CopyLinkButton
              textToCopy={joinUrl}
              defaultText="Link"
              copiedText="Copied!"
              icon={
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
              }
            />
            {isHost && game.settings.hasClaudeApiKey && (
              <div className="relative" ref={botDropdownRef}>
                <GameButton
                  buttonType="secondary"
                  size="md"
                  onClick={() => setShowBotDropdown((prev) => !prev)}
                >
                  <span className="inline-flex items-center gap-2">
                    Bot
                    <svg
                      width="14"
                      height="14"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    >
                      <line x1="12" y1="5" x2="12" y2="19" />
                      <line x1="5" y1="12" x2="19" y2="12" />
                    </svg>
                  </span>
                </GameButton>
                {showBotDropdown && (
                  <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 bg-black border border-white/20 rounded-lg overflow-hidden shadow-lg z-10">
                    <div className="grid grid-cols-[auto_60px_60px] text-center">
                      <div />
                      {["Fast", "Thinker"].map((s) => (
                        <div
                          key={s}
                          className="px-3 py-1.5 text-white/40 text-[10px] font-bold uppercase tracking-wide"
                        >
                          {s}
                        </div>
                      ))}
                      {[
                        { key: "normal", label: "Normal" },
                        { key: "hard", label: "Hard" },
                        { key: "extreme", label: "Actual Bot" },
                      ].map((diff) => (
                        <React.Fragment key={diff.key}>
                          <div className="px-3 py-2 text-white/60 text-[11px] font-semibold flex items-center whitespace-nowrap">
                            {diff.label}
                          </div>
                          {["fast", "thinker"].map((spd) => (
                            <button
                              key={spd}
                              onClick={() => handleAddBot(diff.key, spd)}
                              className="px-3 py-2 hover:bg-white/10 transition-colors cursor-pointer text-white text-lg"
                            >
                              +
                            </button>
                          ))}
                        </React.Fragment>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Demo Game: Configure button */}
        {isDemoGame && (
          <div className="text-center">
            <GameButton
              buttonType="secondary"
              size="md"
              onClick={() => setShowDemoSetup(true)}
              className="w-full"
            >
              DEMO CONFIG
            </GameButton>
          </div>
        )}

        {isDemoGame && <div className="h-2" />}

        {/* Start Game Button (Host only) */}
        {isHost && (
          <div className="text-center">
            <GameButton
              size="lg"
              onClick={handleStartGame}
              disabled={playerCount < 1 || !allBotsReady || !allDemoPlayersReady}
              className="w-full"
            >
              START GAME
            </GameButton>
          </div>
        )}

        {!isHost && (
          <p className="text-white/50 text-sm text-center">Waiting for host to start the game...</p>
        )}
      </GameMenuModal>

      {/* Lobby Settings Overlay (host only) */}
      {isHost && showLobbySettings && (
        <LobbySettingsOverlay
          game={game}
          playerId={playerId}
          isOpen={true}
          onClose={() => setShowLobbySettings(false)}
        />
      )}

      {/* Demo Setup Overlay */}
      {isDemoGame && showDemoSetup && game && playerId && (
        <DemoSetupOverlay
          game={game}
          playerId={playerId}
          isOpen={true}
          onClose={() => setShowDemoSetup(false)}
        />
      )}
    </>
  );
};

export default WaitingRoomOverlay;
