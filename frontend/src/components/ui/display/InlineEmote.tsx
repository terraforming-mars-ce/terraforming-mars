import React, { useEffect } from "react";
import { useBotPresenceStore } from "@/stores/botPresenceStore.ts";
import EmoteIcon from "./EmoteIcon.tsx";

const EMOTE_DURATION_MS = 2500;

interface InlineEmoteProps {
  playerId: string;
}

const InlineEmote: React.FC<InlineEmoteProps> = ({ playerId }) => {
  const emote = useBotPresenceStore((s) => s.emotes[playerId]);

  useEffect(() => {
    if (!emote) {
      return undefined;
    }
    const timer = setTimeout(() => {
      useBotPresenceStore.getState().clearEmote(playerId, emote.seq);
    }, EMOTE_DURATION_MS);
    return () => clearTimeout(timer);
  }, [emote, playerId]);

  return (
    <span className="relative inline-block w-6 h-6 shrink-0 align-middle">
      {emote && (
        <span
          key={emote.seq}
          className="absolute left-1/2 top-1/2 pointer-events-none drop-shadow-[0_2px_6px_rgba(0,0,0,0.8)]"
          style={{ animation: `emotePop ${EMOTE_DURATION_MS}ms ease-out forwards` }}
        >
          <EmoteIcon emote={emote.emote} size={26} />
        </span>
      )}
    </span>
  );
};

export default InlineEmote;
