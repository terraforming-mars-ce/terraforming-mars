import { useEffect } from "react";
import { createPortal } from "react-dom";
import type { GameEvent } from "../../../hooks/useGameEvent.ts";

interface GameEventBannerProps {
  event: GameEvent | null;
  onDismiss: () => void;
}

export default function GameEventBanner({ event, onDismiss }: GameEventBannerProps) {
  useEffect(() => {
    if (!event) {
      return;
    }
    const timer = setTimeout(onDismiss, event.duration);
    return () => clearTimeout(timer);
  }, [event, onDismiss]);

  if (!event) {
    return null;
  }

  const hasSubtitle = event.achievementName && event.playerName;
  const animateClass =
    event.duration <= 2500
      ? "animate-[yourTurnIn_0.4s_ease-out_forwards,yourTurnOut_0.4s_ease-in_2.1s_forwards]"
      : "animate-[yourTurnIn_0.4s_ease-out_forwards,yourTurnOut_0.4s_ease-in_3.6s_forwards]";

  return createPortal(
    <div
      className="fixed inset-0 flex items-center justify-center pointer-events-none"
      style={{ zIndex: 2000 }}
    >
      <div key={event.id} className={`flex flex-col items-center gap-2 ${animateClass}`}>
        <div className="font-orbitron text-4xl font-black text-white tracking-[0.2em] [text-shadow:0_0_40px_rgba(100,200,255,0.8),0_0_80px_rgba(100,200,255,0.4),0_2px_4px_rgba(0,0,0,0.9)]">
          {event.title}
        </div>
        {hasSubtitle && (
          <div className="text-lg tracking-wide [text-shadow:0_2px_4px_rgba(0,0,0,0.9)]">
            <span className="font-orbitron font-bold text-white">{event.achievementName}</span>
            <span className="text-gray-300"> by </span>
            <span className="font-semibold" style={{ color: event.playerColor }}>
              {event.playerName}
            </span>
          </div>
        )}
      </div>
    </div>,
    document.body,
  );
}
