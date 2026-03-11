import React, { useEffect, useState, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import {
  PlayerDto,
  OtherPlayerDto,
  TriggeredEffectDto,
  CalculatedOutputDto,
  CardBehaviorDto,
  VPConditionDto,
} from "@/types/generated/api-types.ts";
import BehaviorSection from "./BehaviorSection";
import { useHoverSound } from "@/hooks/useHoverSound.ts";
import GameIcon from "@/components/ui/display/GameIcon.tsx";
import CardIcon from "./BehaviorSection/components/CardIcon.tsx";
import VictoryPointIcon from "@/components/ui/display/VictoryPointIcon.tsx";

const resourceTypeToIconType: Record<string, string> = {
  credit: "credit",
  steel: "steel",
  titanium: "titanium",
  plant: "plant",
  energy: "energy",
  heat: "heat",
  microbe: "microbe",
  animal: "animal",
  floater: "floater",
  science: "science",
  asteroid: "asteroid",
  fighter: "fighter",
  disease: "disease",
  "credit-production": "credit-production",
  "steel-production": "steel-production",
  "titanium-production": "titanium-production",
  "plant-production": "plant-production",
  "energy-production": "energy-production",
  "heat-production": "heat-production",
  tr: "tr",
  oxygen: "oxygen",
  temperature: "temperature",
  "ocean-placement": "ocean-placement",
  "greenery-placement": "greenery-placement",
  "city-placement": "city-placement",
  "card-draw": "card-draw",
  "colony-tile": "colony-tile",
};

const cardResourceTypes: Record<string, "peek" | "take" | "buy" | "discard" | "none"> = {
  "card-draw": "none",
  "card-peek": "peek",
  "card-take": "take",
  "card-buy": "buy",
  "card-discard": "discard",
};

const GainedDisplay: React.FC<{
  outputs: CalculatedOutputDto[];
  vpConditions?: VPConditionDto[];
}> = ({ outputs, vpConditions }) => {
  const nonZero = outputs.filter((o) => o.amount !== 0);
  const hasVP = vpConditions && vpConditions.length > 0;
  if (nonZero.length === 0 && !hasVP) return null;

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <span className="text-[9px] text-gray-400 uppercase tracking-wider font-orbitron">
        Gained:
      </span>
      {nonZero.map((output, index) => {
        const badgeType = cardResourceTypes[output.resourceType];
        if (badgeType !== undefined) {
          return (
            <CardIcon
              key={index}
              amount={Math.abs(output.amount)}
              badgeType={badgeType}
              totalCardTypes={1}
            />
          );
        }
        const iconType = resourceTypeToIconType[output.resourceType] || output.resourceType;
        return (
          <div key={index} className="flex items-center gap-0.5">
            <GameIcon iconType={iconType} amount={output.amount} size="small" />
          </div>
        );
      })}
      {hasVP && <VictoryPointIcon vpConditions={vpConditions} />}
    </div>
  );
};

interface NotificationItem {
  id: string;
  cardName: string;
  sourceType: string;
  calculatedOutputs: CalculatedOutputDto[];
  behaviors: CardBehaviorDto[];
  vpConditions: VPConditionDto[];
}

interface PlayerCardProps {
  player: PlayerDto | OtherPlayerDto;
  playerColor: string;
  isCurrentPlayer: boolean;
  isCurrentTurn: boolean;
  isActionPhase: boolean;
  isHost?: boolean;
  onSkipAction?: () => void;
  hasPendingTile?: boolean;
  triggeredEffects?: TriggeredEffectDto[];
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
  onKickPlayer?: (playerId: string) => void;
  onConvertToBot?: (playerId: string) => void;
  minNameWidth?: number;
  minCardWidth?: number;
}

const PlayerCard: React.FC<PlayerCardProps> = ({
  player,
  playerColor,
  isCurrentPlayer,
  isCurrentTurn,
  isActionPhase,
  isHost = false,
  onSkipAction,
  hasPendingTile = false,
  triggeredEffects = [],
  onPlayerClick,
  onKickPlayer,
  onConvertToBot,
  minNameWidth,
  minCardWidth,
}) => {
  const hoverSound = useHoverSound(hasPendingTile);
  const isPassed = player.passed;
  const isDisconnected = !player.isConnected;
  const isExited = player.isExited;
  const isInProduction =
    player.productionPhase != null && !player.productionPhase.selectionComplete;
  const hasUnlimitedActions = player.availableActions === -1;
  const actionsRemaining = player.availableActions;

  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null);
  const contextMenuRef = useRef<HTMLDivElement>(null);

  const canKick = isHost && !isCurrentPlayer && !isExited && onKickPlayer;
  const canConvertToBot =
    isHost && !isCurrentPlayer && !isExited && player.playerType !== "bot" && onConvertToBot;
  const hasContextMenu = canKick || canConvertToBot;

  const handleContextMenu = useCallback(
    (e: React.MouseEvent) => {
      if (!hasContextMenu) return;
      e.preventDefault();
      setContextMenu({ x: e.clientX, y: e.clientY });
    },
    [hasContextMenu],
  );

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

  const [queue, setQueue] = useState<NotificationItem[]>([]);
  const [active, setActive] = useState<{ item: NotificationItem; visible: boolean } | null>(null);
  const prevTriggeredEffectsRef = useRef<TriggeredEffectDto[]>([]);

  // Process incoming triggered effects into the notification queue
  useEffect(() => {
    if (triggeredEffects === prevTriggeredEffectsRef.current) return;
    prevTriggeredEffectsRef.current = triggeredEffects;

    const playerEffects = triggeredEffects.filter((e) => e.playerId === player.id);
    if (playerEffects.length === 0) return;

    // Group by cardName + sourceType so different notification types stay separate
    const grouped = new Map<string, NotificationItem>();
    for (const effect of playerEffects) {
      const key = `${effect.cardName}::${effect.sourceType}`;
      const existing = grouped.get(key);
      if (existing) {
        if (effect.calculatedOutputs) existing.calculatedOutputs.push(...effect.calculatedOutputs);
        if (effect.behaviors) existing.behaviors.push(...effect.behaviors);
        if (effect.vpConditions) existing.vpConditions.push(...effect.vpConditions);
      } else {
        grouped.set(key, {
          id: `${Date.now()}-${Math.random()}`,
          cardName: effect.cardName,
          sourceType: effect.sourceType || "",
          calculatedOutputs: effect.calculatedOutputs ? [...effect.calculatedOutputs] : [],
          behaviors: effect.behaviors ? [...effect.behaviors] : [],
          vpConditions: effect.vpConditions ? [...effect.vpConditions] : [],
        });
      }
    }

    // Filter out items with no content to display
    const items = Array.from(grouped.values()).filter((item) => {
      const hasOutputs = item.calculatedOutputs.some((o) => o.amount !== 0);
      const hasVP = item.vpConditions.length > 0;
      const hasBehaviors = item.behaviors.length > 0;
      return hasOutputs || hasVP || hasBehaviors;
    });

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
      className={`relative h-[66px] overflow-visible pointer-events-auto ${isCurrentTurn ? "mb-1.5" : "mb-2"} ${onPlayerClick ? "cursor-pointer" : ""}`}
      onClick={() => onPlayerClick?.(player)}
      onContextMenu={handleContextMenu}
    >
      {/* Main player card with angled edge */}
      <div
        data-player-card-inner
        className={`relative h-full bg-[rgba(10,10,15,0.95)] border-l-[6px] border-t border-t-[rgba(60,60,70,0.7)] pl-2 transition-all duration-300 flex items-center [clip-path:polygon(0_0,calc(100%-8px)_0,100%_100%,0_100%)] w-fit z-[2] shadow-[0_2px_8px_rgba(0,0,0,0.5),-2px_0_6px_var(--player-color)] ${isExited || isDisconnected ? "opacity-20" : ""} ${!isCurrentTurn ? "opacity-60" : ""} ${isCurrentTurn ? "border-l-8 shadow-[0_4px_16px_rgba(0,0,0,0.6),-4px_0_12px_var(--player-color)]" : ""}`}
        style={
          {
            "--player-color": playerColor,
            borderLeftColor: playerColor,
            minWidth: minCardWidth ? `${minCardWidth}px` : undefined,
            paddingRight: "24px",
          } as React.CSSProperties
        }
      >
        <div className="flex flex-col items-start justify-center gap-1">
          <div className="flex gap-1 flex-wrap justify-start items-center">
            {isCurrentPlayer && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(60,100,150,0.8)] text-white border border-[rgba(80,130,180,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                YOU
              </span>
            )}
            {isPassed && !isExited && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(80,80,90,0.6)] text-[rgb(140,140,150)] border border-[rgba(60,60,70,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                PASSED
              </span>
            )}
            {player.playerType === "bot" && player.botStatus === "thinking" && !isInProduction && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(120,80,200,0.6)] text-[rgb(200,180,255)] border border-[rgba(140,100,220,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] flex items-center gap-1">
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
              </span>
            )}
            {player.playerType === "bot" && player.botStatus !== "thinking" && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(120,80,200,0.4)] text-[rgb(180,160,230)] border border-[rgba(120,80,200,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                BOT
              </span>
            )}
            {isExited && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border border-[rgba(180,60,60,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                EXITED
              </span>
            )}
            {isDisconnected && !isExited && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border border-[rgba(180,60,60,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                DISCONNECTED
              </span>
            )}
            {player.botStatus === "failed" && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(200,50,50,0.6)] text-[rgb(255,140,140)] border border-[rgba(200,50,50,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                ERROR
              </span>
            )}
            {hasPendingTile && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(180,140,50,0.6)] text-[rgb(255,220,140)] border border-[rgba(180,140,50,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                TILE
              </span>
            )}
            {isInProduction && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(180,120,40,0.6)] text-[rgb(255,200,120)] border border-[rgba(180,120,40,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] flex items-center gap-1">
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
              </span>
            )}
          </div>
          <div
            className="flex items-center gap-1.5"
            style={minNameWidth ? { minWidth: `${minNameWidth}px` } : undefined}
          >
            <span className="text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] tracking-[0.3px] shrink-0">
              {player.name}
            </span>
            {isCurrentTurn && isActionPhase && !isPassed && (
              <span className="text-[10px] font-bold font-orbitron text-[rgb(140,160,190)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                {hasUnlimitedActions ? "∞" : `${actionsRemaining}/${player.totalActions}`}
              </span>
            )}
          </div>
        </div>
        {/* TR Display */}
        <div className="flex items-center bg-[rgba(30,50,80,0.9)] border border-[rgba(60,100,150,0.6)] px-2.5 py-1 shrink-0 ml-auto">
          <span className="text-sm font-bold font-orbitron text-[rgb(180,210,255)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            {player.terraformRating}
          </span>
        </div>
        {/* PASS/SKIP button */}
        {isCurrentPlayer && isCurrentTurn && isActionPhase ? (
          <button
            className={`py-1.5 px-3 text-[9px] font-bold font-orbitron uppercase tracking-[0.5px] transition-all duration-200 shrink-0 ml-2 ${
              hasPendingTile
                ? "bg-[rgba(40,40,45,0.9)] text-[rgb(100,100,110)] border border-[rgba(60,60,70,0.5)] cursor-default"
                : "bg-[rgba(50,100,160,0.95)] text-white border border-[rgba(80,140,200,0.8)] cursor-pointer hover:bg-[rgba(60,120,180,1)] hover:border-[rgba(100,160,220,0.9)]"
            }`}
            onClick={(e) => {
              e.stopPropagation();
              if (hasPendingTile) return;
              hoverSound.onClick?.();
              onSkipAction?.();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
            disabled={hasPendingTile}
          >
            {buttonText}
          </button>
        ) : (
          <div className="py-1.5 px-3 text-[9px] font-orbitron shrink-0 ml-2 invisible">PASS</div>
        )}
      </div>

      {/* Right-click context menu */}
      {contextMenu &&
        createPortal(
          <div
            ref={contextMenuRef}
            className="fixed z-[10000] bg-[rgba(15,15,20,0.98)] border border-[rgba(60,60,70,0.7)] rounded-lg shadow-[0_8px_32px_rgba(0,0,0,0.7)] py-1 min-w-[180px]"
            style={{ left: contextMenu.x, top: contextMenu.y }}
          >
            {canConvertToBot && (
              <button
                className="w-full flex items-center gap-3 px-4 py-3 text-left text-sm text-red-400 hover:bg-white/10 transition-colors cursor-pointer"
                onClick={(e) => {
                  e.stopPropagation();
                  setContextMenu(null);
                  onConvertToBot?.(player.id);
                }}
              >
                Convert to bot
              </button>
            )}
            {canConvertToBot && canKick && <div className="border-t border-[#333]" />}
            {canKick && (
              <button
                className="w-full flex items-center gap-3 px-4 py-3 text-left text-sm text-red-400 hover:bg-white/10 transition-colors cursor-pointer"
                onClick={(e) => {
                  e.stopPropagation();
                  setContextMenu(null);
                  onKickPlayer?.(player.id);
                }}
              >
                Kick player
              </button>
            )}
          </div>,
          document.body,
        )}

      {/* Triggered effect notification - rendered via portal to avoid clipping */}
      {active &&
        cardRect &&
        createPortal(
          <div
            className="fixed z-[9999] pointer-events-none"
            style={{
              left: `${cardRect.right + 10}px`,
              top: `${cardRect.top + cardRect.height / 2}px`,
              transform: "translateY(-50%)",
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
            .notification-content div {
              flex-wrap: nowrap !important;
            }
            .notification-content > div > div {
              background: none !important;
              border-color: transparent !important;
              box-shadow: none !important;
            }
          `}</style>
            {(() => {
              const { item, visible } = active;
              const isActionAdded = item.sourceType === "action_added";
              const isEffectAdded = item.sourceType === "effect_added";
              const hasBehaviors = item.behaviors.length > 0;

              return (
                <div
                  key={item.id}
                  className="notification-content relative flex flex-col items-center gap-1 px-3 py-2 bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
                  style={{
                    clipPath:
                      "polygon(0 0, calc(100% - 10px) 0, 100% 10px, 100% 100%, 10px 100%, 0 calc(100% - 10px))",
                    animation: visible
                      ? "notificationEnter 0.3s ease-out forwards"
                      : "notificationExit 0.3s ease-in forwards",
                  }}
                >
                  <svg
                    className="absolute top-0 right-0 w-[10px] h-[10px] pointer-events-none"
                    viewBox="0 0 10 10"
                  >
                    <line
                      x1="0"
                      y1="0"
                      x2="10"
                      y2="10"
                      stroke="rgba(60,60,70,0.7)"
                      strokeWidth="1.5"
                    />
                  </svg>
                  <svg
                    className="absolute bottom-0 left-0 w-[10px] h-[10px] pointer-events-none"
                    viewBox="0 0 10 10"
                  >
                    <line
                      x1="0"
                      y1="0"
                      x2="10"
                      y2="10"
                      stroke="rgba(60,60,70,0.7)"
                      strokeWidth="1.5"
                    />
                  </svg>
                  <span className="text-white text-xs font-bold font-orbitron [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                    {item.cardName}
                  </span>
                  {isActionAdded && hasBehaviors ? (
                    <div className="flex flex-col items-center gap-0.5">
                      <span className="text-[9px] text-gray-400 uppercase tracking-wider font-orbitron">
                        New Action:
                      </span>
                      <BehaviorSection behaviors={item.behaviors} />
                    </div>
                  ) : isEffectAdded && hasBehaviors ? (
                    <div className="flex flex-col items-center gap-0.5">
                      <span className="text-[9px] text-gray-400 uppercase tracking-wider font-orbitron">
                        New Effect:
                      </span>
                      <BehaviorSection behaviors={item.behaviors} />
                    </div>
                  ) : (
                    <GainedDisplay
                      outputs={item.calculatedOutputs}
                      vpConditions={item.vpConditions.length > 0 ? item.vpConditions : undefined}
                    />
                  )}
                </div>
              );
            })()}
          </div>,
          document.body,
        )}
    </div>
  );
};

export default PlayerCard;
