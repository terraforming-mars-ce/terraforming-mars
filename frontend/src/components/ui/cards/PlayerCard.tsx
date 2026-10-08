import GameButton from "@/components/ui/buttons/GameButton.tsx";
import React, { useEffect, useLayoutEffect, useState, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import {
  PlayerDto,
  OtherPlayerDto,
  TriggeredEffectDto,
  PlayerStatusTile,
  PlayerStatusSelection,
  PlayerStatusSelectingProductionCards,
} from "@/types/generated/api-types.ts";
import { Z_INDEX, getZIndex } from "@/constants/zIndex.ts";
import { PlayerChip } from "@/components/ui/display/BotChips.tsx";
import RevealTrigger from "@/components/ui/display/RevealTrigger.tsx";
import { useLongPress } from "@/hooks/useLongPress.ts";
import { usePlayerHostActions } from "@/hooks/usePlayerHostActions.ts";
import PlayerPresence from "./PlayerPresence.tsx";
import TriggeredEffectToast, {
  groupTriggeredEffects,
  type TriggeredEffectItem,
} from "./TriggeredEffectToast.tsx";

interface PlayerCardProps {
  player: PlayerDto | OtherPlayerDto;
  playerColor: string;
  isCurrentPlayer: boolean;
  isCurrentTurn: boolean;
  isActionPhase: boolean;
  isHost?: boolean;
  onSkipAction?: () => void;
  triggeredEffects?: TriggeredEffectDto[];
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
  onKickPlayer?: (playerId: string) => void;
  onConvertToBot?: (playerId: string) => void;
}

const PlayerCard: React.FC<PlayerCardProps> = ({
  player,
  playerColor,
  isCurrentPlayer,
  isCurrentTurn,
  isActionPhase,
  isHost = false,
  onSkipAction,
  triggeredEffects = [],
  onPlayerClick,
  onKickPlayer,
  onConvertToBot,
}) => {
  const hasPendingTile = player.status === PlayerStatusTile;
  const hasSelection = player.status === PlayerStatusSelection;
  const isInProduction = player.status === PlayerStatusSelectingProductionCards;
  const isBlocked = hasPendingTile || hasSelection;
  const isPassed = player.passed;
  const isDisconnected = player.playerType !== "bot" && !player.isConnected;
  const isExited = player.isExited;
  const hasUnlimitedActions = player.availableActions === -1;
  const actionsRemaining = player.availableActions;
  const hasNoAvailableActions =
    "hasAvailableActions" in player && player.hasAvailableActions === false;
  const showStuckIndicator =
    isCurrentPlayer && isCurrentTurn && isActionPhase && hasNoAvailableActions;

  const [shouldFlash, setShouldFlash] = useState(false);
  const prevIsCurrentTurnRef = useRef(isCurrentTurn);

  useEffect(() => {
    const wasCurrentTurn = prevIsCurrentTurnRef.current;
    prevIsCurrentTurnRef.current = isCurrentTurn;

    if (isCurrentTurn && !wasCurrentTurn && isCurrentPlayer) {
      setShouldFlash(true);
      const timer = setTimeout(() => setShouldFlash(false), 5000);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [isCurrentTurn, isCurrentPlayer]);

  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null);
  const contextMenuRef = useRef<HTMLDivElement>(null);

  const hostActions = usePlayerHostActions({
    player,
    isHost,
    isCurrentPlayer,
    onKickPlayer,
    onConvertToBot,
  });
  const hasContextMenu = hostActions.length > 0;

  const openContextMenu = useCallback(
    (point: { x: number; y: number }) => {
      if (hasContextMenu) {
        setContextMenu(point);
      }
    },
    [hasContextMenu],
  );
  const longPress = useLongPress<HTMLDivElement>(openContextMenu);

  const handleContextMenu = (e: React.MouseEvent<HTMLDivElement>) => {
    longPress.onContextMenu(e);
    if (!hasContextMenu || e.isPropagationStopped()) {
      return;
    }
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY });
  };

  useLayoutEffect(() => {
    const menu = contextMenuRef.current;
    if (!contextMenu || !menu) {
      return;
    }
    const rect = menu.getBoundingClientRect();
    menu.style.left = `${Math.max(8, Math.min(contextMenu.x, window.innerWidth - rect.width - 8))}px`;
    menu.style.top = `${Math.max(8, Math.min(contextMenu.y, window.innerHeight - rect.height - 8))}px`;
  }, [contextMenu]);

  useEffect(() => {
    if (!contextMenu) return;
    const handleDismiss = () => setContextMenu(null);
    const timeoutId = setTimeout(() => {
      document.addEventListener("click", handleDismiss);
      document.addEventListener("contextmenu", handleDismiss);
    }, 0);
    return () => {
      clearTimeout(timeoutId);
      document.removeEventListener("click", handleDismiss);
      document.removeEventListener("contextmenu", handleDismiss);
    };
  }, [contextMenu]);

  const buttonText = hasUnlimitedActions || actionsRemaining === 2 ? "PASS" : "SKIP";

  const [queue, setQueue] = useState<TriggeredEffectItem[]>([]);
  const [active, setActive] = useState<{ item: TriggeredEffectItem; visible: boolean } | null>(
    null,
  );
  const prevTriggeredEffectsRef = useRef<TriggeredEffectDto[]>([]);

  // Process incoming triggered effects into the notification queue
  useEffect(() => {
    if (triggeredEffects === prevTriggeredEffectsRef.current) return;
    prevTriggeredEffectsRef.current = triggeredEffects;

    const items = groupTriggeredEffects(triggeredEffects, player.id);
    if (items.length > 0) {
      setQueue((prev) => [...prev, ...items]);
    }
  }, [triggeredEffects, player.id]);

  // Show next item from queue when nothing is active
  const showNext = useCallback(() => {
    setQueue((prev) => {
      if (prev.length === 0) return prev;
      const [next, ...rest] = prev;
      setActive({ item: next, visible: true });
      return rest;
    });
  }, []);

  useEffect(() => {
    if (active === null && queue.length > 0) {
      showNext();
    }
  }, [active, queue.length, showNext]);

  // Auto-dismiss active notification
  useEffect(() => {
    if (!active) return;

    if (active.visible) {
      const timer = setTimeout(() => {
        setActive((prev) => (prev ? { ...prev, visible: false } : null));
      }, 3000);
      return () => clearTimeout(timer);
    }

    // Fade-out finished, clear active to trigger next
    const timer = setTimeout(() => setActive(null), 300);
    return () => clearTimeout(timer);
  }, [active]);

  const cardRef = useRef<HTMLDivElement>(null);
  const [cardRect, setCardRect] = useState<DOMRect | null>(null);

  useEffect(() => {
    if (active && cardRef.current) {
      setCardRect(cardRef.current.getBoundingClientRect());
    }
  }, [active]);

  return (
    <div
      ref={cardRef}
      data-player-id={player.id}
      className={`relative w-[260px] max-w-[calc(100vw-16px)] h-[66px] overflow-visible pointer-events-auto pointer-coarse:select-none [-webkit-touch-callout:none] ${isCurrentTurn ? "mb-1.5" : "mb-2"} ${onPlayerClick ? "cursor-pointer" : ""}`}
      {...longPress}
      onClick={() => onPlayerClick?.(player)}
      onContextMenu={handleContextMenu}
    >
      {/* Main player card with angled edge */}
      <div
        data-player-card-inner
        className={`relative h-full bg-[rgba(10,10,15,0.95)] border-l-[6px] border-t border-t-[rgba(60,60,70,0.7)] pl-2 transition-all duration-300 flex items-center [clip-path:polygon(0_0,calc(100%-8px)_0,100%_100%,0_100%)] w-full shadow-[0_2px_8px_rgba(0,0,0,0.5),-2px_0_6px_var(--player-color)] ${isExited || isDisconnected ? "opacity-20" : ""} ${!isCurrentTurn ? "opacity-60" : ""} ${isCurrentTurn ? "border-l-8 shadow-[0_4px_16px_rgba(0,0,0,0.6),-4px_0_12px_var(--player-color)]" : ""}`}
        style={
          {
            "--player-color": playerColor,
            borderLeftColor: playerColor,
            paddingRight: "24px",
            zIndex: getZIndex("LOCAL", 2),
          } as React.CSSProperties
        }
      >
        {shouldFlash && (
          <div
            className="absolute inset-0 pointer-events-none"
            style={{
              boxShadow: `0 0 80px ${playerColor}, 0 0 40px ${playerColor}, inset 0 0 40px ${playerColor}99`,
              animation: "turnFlash 5s ease-out forwards",
              clipPath: "polygon(0 0, calc(100% - 8px) 0, 100% 100%, 0 100%)",
            }}
          />
        )}
        <div className="flex min-w-0 flex-1 flex-col items-start justify-center gap-1">
          <div className="player-chip-group flex-nowrap gap-1">
            {isCurrentPlayer && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(60,100,150,0.8)] text-white border border-[rgba(80,130,180,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                YOU
              </PlayerChip>
            )}
            {isPassed && !isExited && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(80,80,90,0.6)] text-[rgb(140,140,150)] border border-[rgba(60,60,70,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                PASSED
              </PlayerChip>
            )}
            {player.playerType === "bot" && player.botStatus === "thinking" && !isInProduction && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(120,80,200,0.6)] text-[rgb(200,180,255)] border border-[rgba(140,100,220,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] flex items-center gap-1">
                THINKING
                <svg
                  className="animate-spin"
                  width="8"
                  height="8"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="3"
                >
                  <path d="M12 2a10 10 0 0 1 10 10" strokeLinecap="round" />
                </svg>
              </PlayerChip>
            )}
            {player.playerType === "bot" && player.botStatus !== "thinking" && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(120,80,200,0.4)] text-[rgb(180,160,230)] border border-[rgba(120,80,200,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                BOT
              </PlayerChip>
            )}
            {isExited && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border border-[rgba(180,60,60,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                EXITED
              </PlayerChip>
            )}
            {isDisconnected && !isExited && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border border-[rgba(180,60,60,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                DISCONNECTED
              </PlayerChip>
            )}
            {player.botStatus === "failed" && (
              <RevealTrigger
                className="inline-flex cursor-default"
                maxWidth={260}
                cornerSize={8}
                content={
                  <>
                    {player.botError && (
                      <span className="block font-sans text-xs font-normal leading-relaxed">
                        {player.botError}
                      </span>
                    )}
                    <span className="block mt-1 font-sans text-[10px] text-white/50">
                      Playing on autopilot until retried.
                    </span>
                  </>
                }
              >
                <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(200,50,50,0.6)] text-[rgb(255,140,140)] border border-[rgba(200,50,50,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                  ERROR
                </PlayerChip>
              </RevealTrigger>
            )}
            {showStuckIndicator && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(180,140,50,0.6)] text-[rgb(255,220,140)] border border-[rgba(180,140,50,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                No actions available
              </PlayerChip>
            )}
            {hasPendingTile && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(180,140,50,0.6)] text-[rgb(255,220,140)] border border-[rgba(180,140,50,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                TILE
              </PlayerChip>
            )}
            {hasSelection && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(60,140,180,0.6)] text-[rgb(140,220,255)] border border-[rgba(60,140,180,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                SELECTION
              </PlayerChip>
            )}
            {isInProduction && (
              <PlayerChip className="text-[8px] tracking-[0.5px] bg-[rgba(180,120,40,0.6)] text-[rgb(255,200,120)] border border-[rgba(180,120,40,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] flex items-center gap-1">
                PRODUCTION
                <svg
                  className="animate-spin"
                  width="8"
                  height="8"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="3"
                >
                  <path d="M12 2a10 10 0 0 1 10 10" strokeLinecap="round" />
                </svg>
              </PlayerChip>
            )}
          </div>
          <div className="flex max-w-full items-center gap-1.5">
            <span className="truncate text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] tracking-[0.3px]">
              {player.name}
            </span>
            {isCurrentTurn && isActionPhase && !isPassed && (
              <span className="shrink-0 text-[10px] font-bold font-orbitron text-[rgb(140,160,190)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                {hasUnlimitedActions ? "∞" : `${actionsRemaining}/${player.totalActions}`}
              </span>
            )}
          </div>
          {player.corporation && (
            <span className="-mt-1 max-w-full truncate text-[9px] font-orbitron uppercase tracking-[0.5px] text-white/50 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
              {player.corporation.name}
            </span>
          )}
        </div>
        {/* TR Display */}
        <div className="flex items-center bg-[rgba(30,50,80,0.9)] border border-[rgba(60,100,150,0.6)] px-2.5 py-1 shrink-0 ml-3">
          <span className="text-sm font-bold font-orbitron text-[rgb(180,210,255)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            {player.terraformRating}
          </span>
        </div>
        {/* PASS/SKIP button */}
        {isCurrentPlayer && isCurrentTurn && isActionPhase && (
          <GameButton
            emphasis="quiet"
            className={`pointer-fine:min-h-0 py-1.5 px-3 text-[9px] font-bold font-orbitron uppercase tracking-[0.5px] transition-all duration-200 shrink-0 ml-2 ${
              isBlocked
                ? "bg-[rgba(40,40,45,0.9)] text-[rgb(100,100,110)] border border-[rgba(60,60,70,0.5)] cursor-default"
                : "bg-[rgba(50,100,160,0.95)] text-white border border-[rgba(80,140,200,0.8)] cursor-pointer hover:bg-[rgba(60,120,180,1)] hover:border-[rgba(100,160,220,0.9)]"
            } ${
              showStuckIndicator && !isBlocked
                ? "ring-1 ring-[rgba(255,210,120,0.9)] shadow-[0_0_12px_rgba(255,200,100,0.7)] animate-pulse"
                : ""
            }`}
            onClick={(e) => {
              e.stopPropagation();
              if (isBlocked) return;
              onSkipAction?.();
            }}
            disabled={isBlocked}
          >
            {buttonText}
          </GameButton>
        )}
        {hasContextMenu && (
          <GameButton
            emphasis="quiet"
            clickSound={false}
            aria-label={`Host actions for ${player.name}`}
            aria-haspopup="menu"
            aria-expanded={contextMenu !== null}
            className="hidden pointer-coarse:inline-flex size-11 p-0 shrink-0 ml-1 text-white/70"
            onClick={(e) => {
              e.stopPropagation();
              const rect = e.currentTarget.getBoundingClientRect();
              setContextMenu((current) => (current ? null : { x: rect.left, y: rect.bottom }));
            }}
          >
            <svg width="18" height="4" viewBox="0 0 18 4" fill="currentColor" aria-hidden="true">
              <circle cx="2" cy="2" r="2" />
              <circle cx="9" cy="2" r="2" />
              <circle cx="16" cy="2" r="2" />
            </svg>
          </GameButton>
        )}
      </div>

      <PlayerPresence playerId={player.id} anchorRef={cardRef} />

      {/* Right-click context menu */}
      {contextMenu &&
        createPortal(
          <div
            ref={contextMenuRef}
            role="menu"
            className="fixed bg-[rgba(15,15,20,0.98)] border border-[rgba(60,60,70,0.7)] rounded-lg shadow-[0_8px_32px_rgba(0,0,0,0.7)] py-1 min-w-[180px]"
            style={{ left: contextMenu.x, top: contextMenu.y, zIndex: Z_INDEX.POPOVER }}
          >
            {hostActions.map((action, index) => (
              <React.Fragment key={action.id}>
                {index > 0 && <div className="border-t border-[#333]" />}
                <GameButton
                  emphasis="quiet"
                  role="menuitem"
                  className={`w-full flex items-center gap-3 px-4 py-3 text-left text-sm ${action.tone === "danger" ? "text-red-400" : "text-white"} hover:bg-white/10 transition-colors cursor-pointer`}
                  onClick={(e) => {
                    e.stopPropagation();
                    setContextMenu(null);
                    action.run();
                  }}
                >
                  {action.label}
                </GameButton>
              </React.Fragment>
            ))}
          </div>,
          document.body,
        )}

      {/* Triggered effect notification - rendered via portal to avoid clipping */}
      {active &&
        cardRect &&
        createPortal(
          <div
            className="fixed pointer-events-none"
            style={{
              left: `${cardRect.right + 10}px`,
              top: `${cardRect.top + cardRect.height / 2}px`,
              transform: "translateY(-50%)",
              zIndex: Z_INDEX.PLAYER_EFFECT_TOAST,
            }}
          >
            <style>{`
            @keyframes notificationEnter {
              0% {
                opacity: 0;
                transform: translateX(50px);
              }
              100% {
                opacity: 1;
                transform: translateX(0);
              }
            }
            @keyframes notificationExit {
              0% {
                opacity: 1;
                transform: translateX(0);
              }
              100% {
                opacity: 0;
                transform: translateX(-50px);
              }
            }
          `}</style>
            <TriggeredEffectToast
              key={active.item.id}
              item={active.item}
              style={{
                animation: active.visible
                  ? "notificationEnter 0.3s ease-out forwards"
                  : "notificationExit 0.3s ease-in forwards",
              }}
            />
          </div>,
          document.body,
        )}
    </div>
  );
};

export default PlayerCard;
