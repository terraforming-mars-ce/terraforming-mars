import { useLayoutEffect, useState, type RefObject } from "react";
import { createPortal } from "react-dom";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { useBotPresenceStore } from "@/stores/botPresenceStore.ts";
import InlineEmote from "../../ui/display/InlineEmote.tsx";

const DOT_DELAYS = ["0ms", "150ms", "300ms"];

function TypingDots() {
  return (
    <span className="flex items-center gap-1 h-6 px-2 rounded-full bg-[rgba(10,10,15,0.92)] border border-white/15 shadow-[0_2px_6px_rgba(0,0,0,0.6)]">
      {DOT_DELAYS.map((delay) => (
        <span
          key={delay}
          className="w-1.5 h-1.5 rounded-full bg-white/80 motion-safe:animate-bounce"
          style={{ animationDelay: delay }}
        />
      ))}
    </span>
  );
}

interface MobilePlayerPresenceProps {
  playerId: string;
  anchorRef: RefObject<HTMLElement | null>;
}

export default function MobilePlayerPresence({ playerId, anchorRef }: MobilePlayerPresenceProps) {
  const hasEmote = useBotPresenceStore((s) => s.emotes[playerId] !== undefined);
  const isTyping = useBotPresenceStore((s) => s.typing[playerId] === true);
  const [rect, setRect] = useState<DOMRect | null>(null);
  const active = hasEmote || isTyping;

  useLayoutEffect(() => {
    const anchor = anchorRef.current;
    if (!active || !anchor) {
      return undefined;
    }
    const update = () => setRect(anchor.getBoundingClientRect());
    update();
    const observer = new ResizeObserver(update);
    observer.observe(anchor);
    const scroller = anchor.parentElement;
    scroller?.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update);
    return () => {
      observer.disconnect();
      scroller?.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
    };
  }, [active, anchorRef]);

  if (!active || !rect) {
    return null;
  }

  return createPortal(
    <div
      className="fixed flex items-center gap-1.5 pointer-events-none"
      aria-hidden="true"
      style={{
        zIndex: Z_INDEX.MOBILE_FEEDBACK,
        right: window.innerWidth - rect.left + 6,
        top: rect.top + rect.height / 2,
        transform: "translateY(-50%)",
      }}
    >
      {hasEmote && <InlineEmote playerId={playerId} />}
      {isTyping && <TypingDots />}
    </div>,
    document.body,
  );
}
