import { useEffect, useRef, useState } from "react";
import GameCard from "../cards/GameCard.tsx";
import type { PlayedCardNotification as PlayedCardNotificationType } from "@/hooks/usePlayedCardNotification.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

type AnimationPhase = "entering" | "visible" | "exiting";

const DISPLAY_DURATION = 2000;
const ANIMATION_DURATION = 400;

interface PlayedCardNotificationProps {
  notification: PlayedCardNotificationType;
  isPinned: boolean;
  onTogglePin: () => void;
  onAdvance: () => void;
}

function getAnimationStyle(phase: AnimationPhase): React.CSSProperties {
  if (phase === "entering") {
    return { animation: `playedCardSlideIn ${ANIMATION_DURATION}ms ease-out forwards` };
  }
  if (phase === "exiting") {
    return { animation: `playedCardSlideOut ${ANIMATION_DURATION}ms ease-in forwards` };
  }
  return { opacity: 1 };
}

export default function PlayedCardNotificationOverlay({
  notification,
  isPinned,
  onTogglePin,
  onAdvance,
}: PlayedCardNotificationProps) {
  const [phase, setPhase] = useState<AnimationPhase>("entering");
  const prevNotificationId = useRef(notification.id);
  const wasPinnedRef = useRef(isPinned);

  if (notification.id !== prevNotificationId.current) {
    prevNotificationId.current = notification.id;
    setPhase("entering");
  }

  useEffect(() => {
    if (phase !== "entering") {
      return;
    }
    const timer = setTimeout(() => {
      setPhase("visible");
    }, ANIMATION_DURATION);
    return () => clearTimeout(timer);
  }, [phase]);

  useEffect(() => {
    if (phase !== "visible" || isPinned) {
      return;
    }
    const timer = setTimeout(() => {
      setPhase("exiting");
    }, DISPLAY_DURATION);
    return () => clearTimeout(timer);
  }, [phase, isPinned]);

  useEffect(() => {
    if (wasPinnedRef.current && !isPinned && phase === "visible") {
      setPhase("exiting");
    }
    wasPinnedRef.current = isPinned;
  }, [isPinned, phase]);

  useEffect(() => {
    if (phase !== "exiting") {
      return;
    }
    const timer = setTimeout(() => {
      onAdvance();
    }, ANIMATION_DURATION);
    return () => clearTimeout(timer);
  }, [phase, onAdvance]);

  return (
    <div
      className="fixed top-1/2 pointer-events-auto"
      style={{
        left: "75vw",
        zIndex: Z_INDEX.PLAYER_OVERLAY,
        transform: "translateY(-50%)",
        ...getAnimationStyle(phase),
      }}
    >
      <div className="text-center mb-4 font-orbitron text-xs tracking-wider">
        <span style={{ color: notification.playerColor }}>{notification.playerName}</span>
        <span className="text-white/60"> played</span>
      </div>
      <div style={{ width: 200 }}>
        <GameCard card={notification.card} isSelected={isPinned} onSelect={() => onTogglePin()} />
      </div>
    </div>
  );
}
