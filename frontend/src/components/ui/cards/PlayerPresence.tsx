import React, { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useBotPresenceStore, useBotThoughtsPreferenceStore } from "@/stores/botPresenceStore.ts";
import EmoteIcon from "@/components/ui/display/EmoteIcon.tsx";

const THOUGHT_DURATION_MS = 4000;
const EMOTE_DURATION_MS = 2500;

interface PlayerPresenceProps {
  playerId: string;
  anchorRef: React.RefObject<HTMLDivElement | null>;
}

const PlayerPresence: React.FC<PlayerPresenceProps> = ({ playerId, anchorRef }) => {
  const thought = useBotPresenceStore((s) => s.thoughts[playerId]);
  const emote = useBotPresenceStore((s) => s.emotes[playerId]);
  const showBotThoughts = useBotThoughtsPreferenceStore((s) => s.showBotThoughts);
  const [rect, setRect] = useState<DOMRect | null>(null);

  const visibleThought = showBotThoughts ? thought : undefined;

  useEffect(() => {
    if (!thought) {
      return undefined;
    }
    const timer = setTimeout(() => {
      useBotPresenceStore.getState().clearThought(playerId, thought.seq);
    }, THOUGHT_DURATION_MS);
    return () => clearTimeout(timer);
  }, [thought, playerId]);

  useEffect(() => {
    if (!emote) {
      return undefined;
    }
    const timer = setTimeout(() => {
      useBotPresenceStore.getState().clearEmote(playerId, emote.seq);
    }, EMOTE_DURATION_MS);
    return () => clearTimeout(timer);
  }, [emote, playerId]);

  useEffect(() => {
    if ((visibleThought || emote) && anchorRef.current) {
      setRect(anchorRef.current.getBoundingClientRect());
    }
  }, [visibleThought, emote, anchorRef]);

  if (!rect || (!visibleThought && !emote)) {
    return null;
  }

  return createPortal(
    <>
      {visibleThought && (
        <div
          key={visibleThought.seq}
          className="fixed pointer-events-none"
          style={{
            left: `${rect.left + rect.width / 2}px`,
            top: `${rect.top - 8}px`,
            zIndex: Z_INDEX.PLAYER_PRESENCE,
            animation: `botThoughtBubble ${THOUGHT_DURATION_MS}ms ease-out forwards`,
          }}
        >
          <div className="relative max-w-[240px] w-max bg-[rgba(10,10,15,0.95)] border border-[rgba(140,100,220,0.6)] rounded-lg px-3 py-1.5 shadow-[0_2px_10px_rgba(0,0,0,0.6)]">
            <span className="block text-[11px] italic leading-snug text-white/90">
              {visibleThought.text}
            </span>
            <span className="absolute left-1/2 -bottom-[5px] h-2.5 w-2.5 -translate-x-1/2 rotate-45 bg-[rgba(10,10,15,0.95)] border-r border-b border-[rgba(140,100,220,0.6)]" />
          </div>
        </div>
      )}
      {emote && (
        <div
          key={emote.seq}
          className="fixed pointer-events-none drop-shadow-[0_2px_6px_rgba(0,0,0,0.8)]"
          style={{
            left: `${rect.right - 36}px`,
            top: `${rect.top + rect.height / 2}px`,
            zIndex: Z_INDEX.PLAYER_PRESENCE,
            animation: `emotePop ${EMOTE_DURATION_MS}ms ease-out forwards`,
          }}
        >
          <EmoteIcon emote={emote.emote} size={40} />
        </div>
      )}
    </>,
    document.body,
  );
};

export default PlayerPresence;
