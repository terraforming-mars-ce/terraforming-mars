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
  onSkipAction?: () => void;
  hasPendingTilePlacement?: boolean;
  triggeredEffects?: TriggeredEffectDto[];
  onPlayerClick?: (player: PlayerDto | OtherPlayerDto) => void;
}

const PlayerCard: React.FC<PlayerCardProps> = ({
  player,
  playerColor,
  isCurrentPlayer,
  isCurrentTurn,
  isActionPhase,
  onSkipAction,
  hasPendingTilePlacement = false,
  triggeredEffects = [],
  onPlayerClick,
}) => {
  const hoverSound = useHoverSound(hasPendingTilePlacement);
  const isPassed = player.passed;
  const isDisconnected = !player.isConnected;
  const hasUnlimitedActions = player.availableActions === -1;
  const actionsRemaining = player.availableActions;

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
      className={`relative w-full h-[60px] overflow-visible pointer-events-auto ${isCurrentTurn ? "mb-1.5" : "mb-2"} ${onPlayerClick ? "cursor-pointer" : ""}`}
      onClick={() => onPlayerClick?.(player)}
    >
      {/* Main player card with angled edge */}
      <div
        className={`relative h-full bg-[rgba(10,10,15,0.95)] border-l-[6px] border-t border-t-[rgba(60,60,70,0.7)] pl-2 pr-2 transition-all duration-300 flex items-center [clip-path:polygon(0_0,calc(100%-8px)_0,100%_100%,0_100%)] max-w-[270px] z-[2] shadow-[0_2px_8px_rgba(0,0,0,0.5),-2px_0_6px_var(--player-color)] ${isDisconnected ? "opacity-20" : ""} ${!isCurrentTurn ? "opacity-60" : ""} ${isCurrentTurn ? "border-l-8 shadow-[0_4px_16px_rgba(0,0,0,0.6),-4px_0_12px_var(--player-color)]" : ""}`}
        style={
          { "--player-color": playerColor, borderLeftColor: playerColor } as React.CSSProperties
        }
      >
        <div className="flex flex-col items-start justify-center w-full gap-1">
          <div className="flex gap-1 flex-wrap justify-start items-center relative z-[2]">
            {isCurrentPlayer && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(60,100,150,0.8)] text-white border border-[rgba(80,130,180,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                YOU
              </span>
            )}
            {isPassed && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(80,80,90,0.6)] text-[rgb(140,140,150)] border border-[rgba(60,60,70,0.7)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                PASSED
              </span>
            )}
            {isDisconnected && (
              <span className="px-1.5 py-0.5 text-[8px] font-bold font-orbitron uppercase tracking-[0.5px] bg-[rgba(180,60,60,0.4)] text-[rgb(220,140,140)] border border-[rgba(180,60,60,0.5)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] relative z-[3]">
                DISCONNECTED
              </span>
            )}
          </div>
          <span className="text-sm font-bold font-orbitron text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] tracking-[0.3px] shrink-0">
            {player.name}
          </span>
        </div>
        {/* TR Display - score only */}
        <div className="absolute right-24 top-1/2 -translate-y-1/2 flex items-center bg-[rgba(30,50,80,0.9)] border border-[rgba(60,100,150,0.6)] px-2.5 py-1">
          <span className="text-sm font-bold font-orbitron text-[rgb(180,210,255)] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            {player.terraformRating}
          </span>
        </div>
        {isCurrentPlayer && isCurrentTurn && isActionPhase && (
          <button
            className={`absolute right-5 top-1/2 -translate-y-1/2 py-1.5 px-3 text-[9px] font-bold font-orbitron uppercase tracking-[0.5px] transition-all duration-200 ${
              hasPendingTilePlacement
                ? "bg-[rgba(40,40,45,0.9)] text-[rgb(100,100,110)] border border-[rgba(60,60,70,0.5)] cursor-not-allowed"
                : "bg-[rgba(50,100,160,0.95)] text-white border border-[rgba(80,140,200,0.8)] cursor-pointer hover:bg-[rgba(60,120,180,1)] hover:border-[rgba(100,160,220,0.9)]"
            }`}
            onClick={(e) => {
              e.stopPropagation();
              if (hasPendingTilePlacement) return;
              hoverSound.onClick?.();
              onSkipAction?.();
            }}
            onMouseEnter={hoverSound.onMouseEnter}
            disabled={hasPendingTilePlacement}
          >
            {buttonText}
          </button>
        )}
      </div>

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
