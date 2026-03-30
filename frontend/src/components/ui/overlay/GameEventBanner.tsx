import { useEffect } from "react";
import { createPortal } from "react-dom";
import type { GameEvent } from "../../../hooks/useGameEvent.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

function renderColoredTitle(title: string, playerName: string, playerColor: string) {
  const idx = title.indexOf(playerName);
  if (idx === -1) {
    return title;
  }
  const before = title.slice(0, idx);
  const after = title.slice(idx + playerName.length);
  return (
    <>
      {before}
      <span style={{ color: playerColor }}>{playerName}</span>
      {after}
    </>
  );
}

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
  const isLarge = event.size === "large";
  const outDelay = Math.max(0, (event.duration - 400) / 1000);
  const animateStyle = {
    animation: `yourTurnIn 0.4s ease-out forwards, yourTurnOut 0.4s ease-in ${outDelay}s forwards`,
  };

  const titleParts =
    isLarge && event.playerName && event.playerColor
      ? renderColoredTitle(event.title, event.playerName, event.playerColor)
      : event.title;

  return createPortal(
    <div
      className="fixed inset-0 flex items-center justify-center pointer-events-none"
      style={{ zIndex: Z_INDEX.STANDARD_MODAL }}
    >
      <div key={event.id} className="flex flex-col items-center gap-2" style={animateStyle}>
        <div
          className={`font-orbitron font-black text-white tracking-[0.2em] ${isLarge ? "text-6xl [text-shadow:0_2px_4px_rgba(0,0,0,0.9)]" : "text-4xl [text-shadow:0_0_40px_rgba(100,200,255,0.8),0_0_80px_rgba(100,200,255,0.4),0_2px_4px_rgba(0,0,0,0.9)]"}`}
        >
          {titleParts}
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
