import React, { useEffect, useState, useRef } from "react";
import { createPortal } from "react-dom";
import { PlayerDto, OtherPlayerDto, TriggeredEffectDto } from "@/types/generated/api-types.ts";
import BehaviorSection from "./BehaviorSection";
import { useHoverSound } from "@/hooks/useHoverSound.ts";

interface EffectNotification {
  id: string;
  effect: TriggeredEffectDto;
  visible: boolean;
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
}) => {
  const hoverSound = useHoverSound(hasPendingTilePlacement);
  const isPassed = player.passed;
  const isDisconnected = !player.isConnected;
  const hasUnlimitedActions = player.availableActions === -1;
  const actionsRemaining = player.availableActions;

  // Determine button text - PASS for unlimited actions or 2 actions remaining, otherwise SKIP
  const buttonText = hasUnlimitedActions || actionsRemaining === 2 ? "PASS" : "SKIP";

  // Notification state for triggered effects
  const [notifications, setNotifications] = useState<EffectNotification[]>([]);

  // Track processed effect batches to avoid duplicates
  const processedBatchesRef = useRef<Set<string>>(new Set());

  // Filter effects for this player and manage notifications
  useEffect(() => {
    const playerEffects = triggeredEffects.filter((e) => e.playerId === player.id);
    if (playerEffects.length === 0) return;

    // Create a unique batch ID based on the effects
    const batchId = playerEffects
      .map((e) => `${e.cardName}-${e.outputs.map((o) => `${o.type}:${o.amount}`).join(",")}`)
      .join("|");

    // Skip if we've already processed this exact batch
    if (processedBatchesRef.current.has(batchId)) return;
    processedBatchesRef.current.add(batchId);

    // Add new notifications
    const newNotificationIds: string[] = [];
    const newNotifications = playerEffects.map((effect, i) => {
      const id = `${Date.now()}-${i}-${Math.random()}`;
      newNotificationIds.push(id);
      return {
        id,
        effect,
        visible: true,
      };
    });

    setNotifications((prev) => [...prev, ...newNotifications]);

    // Auto-dismiss after 3 seconds (not tied to effect cleanup)
    setTimeout(() => {
      setNotifications((prev) =>
        prev.map((n) => (newNotificationIds.includes(n.id) ? { ...n, visible: false } : n)),
      );
      // Remove from DOM after fade out
      setTimeout(() => {
        setNotifications((prev) => prev.filter((n) => !newNotificationIds.includes(n.id)));
        // Clean up processed batch after removal
        processedBatchesRef.current.delete(batchId);
      }, 300);
    }, 3000);
  }, [triggeredEffects, player.id]);

  // Ref for positioning notifications relative to the card
  const cardRef = useRef<HTMLDivElement>(null);
  const [cardRect, setCardRect] = useState<DOMRect | null>(null);

  // Update card position when notifications change
  useEffect(() => {
    if (notifications.length > 0 && cardRef.current) {
      setCardRect(cardRef.current.getBoundingClientRect());
    }
  }, [notifications.length]);

  return (
    <div
      ref={cardRef}
      className={`relative w-full h-[60px] overflow-visible pointer-events-auto ${isCurrentTurn ? "mb-1.5" : "mb-2"}`}
    >
      {/* Main player card with angled edge */}
      <div
        className={`relative h-full bg-[rgba(10,10,15,0.95)] border-l-[6px] border-t border-t-[rgba(60,60,70,0.7)] pl-2 pr-2 transition-all duration-300 flex items-center [clip-path:polygon(0_0,calc(100%-8px)_0,100%_100%,0_100%)] max-w-[270px] z-[2] shadow-[0_2px_8px_rgba(0,0,0,0.5),-2px_0_6px_var(--player-color)] ${isDisconnected ? "opacity-20" : ""} ${!isCurrentTurn ? "opacity-60" : ""} ${isCurrentTurn ? "border-l-8 shadow-[0_4px_16px_rgba(0,0,0,0.6),-4px_0_12px_var(--player-color)]" : ""}`}
        style={{ "--player-color": playerColor } as React.CSSProperties}
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
            onClick={() => {
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

      {/* Triggered effect notifications - rendered via portal to avoid clipping */}
      {notifications.length > 0 &&
        cardRect &&
        createPortal(
          <div
            className="fixed flex flex-row gap-1 z-[9999] pointer-events-none"
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
            .notification-content img {
              min-width: 32px !important;
              min-height: 32px !important;
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
            {/* Group notifications by card name */}
            {(() => {
              const grouped = new Map<
                string,
                {
                  ids: string[];
                  outputs: (typeof notifications)[0]["effect"]["outputs"];
                  visible: boolean;
                }
              >();

              for (const { id, effect, visible } of notifications) {
                const existing = grouped.get(effect.cardName);
                if (existing) {
                  existing.ids.push(id);
                  existing.outputs.push(...effect.outputs);
                  existing.visible = existing.visible && visible;
                } else {
                  grouped.set(effect.cardName, {
                    ids: [id],
                    outputs: [...effect.outputs],
                    visible,
                  });
                }
              }

              return Array.from(grouped.entries()).map(([cardName, { ids, outputs, visible }]) => (
                <div
                  key={ids.join("-")}
                  className="notification-content flex flex-col items-center gap-1 px-3 py-2 bg-[rgba(10,10,15,0.95)] border border-[rgba(60,60,70,0.7)] shadow-lg"
                  style={{
                    animation: visible
                      ? "notificationEnter 0.3s ease-out forwards"
                      : "notificationExit 0.3s ease-in forwards",
                  }}
                >
                  <span className="text-white text-xs font-bold font-orbitron [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
                    {cardName}
                  </span>
                  <BehaviorSection behaviors={[{ outputs }]} />
                </div>
              ));
            })()}
          </div>,
          document.body,
        )}
    </div>
  );
};

export default PlayerCard;
