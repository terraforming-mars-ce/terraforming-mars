import React, { useState, useRef, useEffect, useCallback, useLayoutEffect, useMemo } from "react";
import { ChatMessageDto } from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex";
import { useGameStore } from "@/stores/gameStore.ts";
import { useBotPresenceStore } from "@/stores/botPresenceStore.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import EmoteIcon, { EMOTES } from "@/components/ui/display/EmoteIcon.tsx";

const BAR_HEIGHT = 90;
const SNAP_THRESHOLD = 40;

interface ChatOverlayProps {
  messages: ChatMessageDto[];
  onSendMessage: (message: string) => void;
  embedded?: boolean;
  isEndgame?: boolean;
  playerColorMap?: Map<string, string>;
  onBoundsChange?: (bounds: DOMRectReadOnly | null) => void;
}

const CHAT_WIDTH = 416;
const MIN_VISIBLE = 80;

const ChatOverlay: React.FC<ChatOverlayProps> = ({
  messages,
  onSendMessage,
  embedded,
  isEndgame,
  playerColorMap,
  onBoundsChange,
}) => {
  const [inputValue, setInputValue] = useState("");
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [isSnapped, setIsSnapped] = useState(true);
  const [snapX, setSnapX] = useState(0);
  const dragOffset = useRef({ x: 0, y: 0 });
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [showEmotePicker, setShowEmotePicker] = useState(false);
  const emotePickerRef = useRef<HTMLDivElement>(null);
  const isSpectator = useGameStore((s) => s.isSpectator);
  const game = useGameStore((s) => s.game);
  const typing = useBotPresenceStore((s) => s.typing);

  const typingNames = useMemo(() => {
    const names: string[] = [];
    const players = [
      ...(game?.currentPlayer?.id ? [game.currentPlayer] : []),
      ...(game?.otherPlayers ?? []),
    ];
    for (const player of players) {
      if (typing[player.id]) {
        names.push(player.name);
      }
    }
    return names;
  }, [game?.currentPlayer, game?.otherPlayers, typing]);

  useEffect(() => {
    if (!showEmotePicker) {
      return undefined;
    }
    const handleClick = (e: MouseEvent) => {
      if (emotePickerRef.current && !emotePickerRef.current.contains(e.target as Node)) {
        setShowEmotePicker(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [showEmotePicker]);

  const snapBottom = BAR_HEIGHT;

  useLayoutEffect(() => {
    const element = containerRef.current;
    if (!element || embedded || !onBoundsChange) {
      return;
    }
    const measure = () => onBoundsChange(element.getBoundingClientRect());
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    window.addEventListener("resize", measure);
    measure();
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", measure);
    };
  }, [embedded, onBoundsChange, position, snapX, isSnapped, isEndgame]);

  useEffect(() => () => onBoundsChange?.(null), [onBoundsChange]);

  useEffect(() => {
    messagesEndRef.current?.parentElement?.scrollTo({
      top: messagesEndRef.current.parentElement.scrollHeight,
    });
  }, [messages]);

  // Reset to default snapped position on window resize
  useEffect(() => {
    const handleResize = () => {
      setIsSnapped(true);
      setSnapX(0);
    };
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  const clampPosition = (x: number, y: number) => {
    const w = containerRef.current?.offsetWidth ?? CHAT_WIDTH;
    const h = containerRef.current?.offsetHeight ?? 0;
    const clampedX = Math.max(MIN_VISIBLE - w, Math.min(window.innerWidth - MIN_VISIBLE, x));
    const clampedY = Math.max(0, Math.min(window.innerHeight - h, y));
    return { x: clampedX, y: clampedY };
  };

  const handleMouseDown = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (embedded || (e.target as HTMLElement).tagName === "INPUT") return;
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return;
      dragOffset.current = { x: e.clientX - rect.left, y: e.clientY - rect.top };
      setIsDragging(true);
    },
    [embedded],
  );

  useEffect(() => {
    if (!isDragging) return;

    const handleMouseMove = (e: MouseEvent) => {
      const rawX = e.clientX - dragOffset.current.x;
      const rawY = e.clientY - dragOffset.current.y;
      const containerHeight = containerRef.current?.offsetHeight ?? 0;
      const bottomEdge = rawY + containerHeight;
      const snapY = window.innerHeight - snapBottom;

      if (Math.abs(bottomEdge - snapY) < SNAP_THRESHOLD) {
        const clamped = clampPosition(rawX, 0);
        setIsSnapped(true);
        setSnapX(clamped.x);
      } else {
        const clamped = clampPosition(rawX, rawY);
        setIsSnapped(false);
        setPosition(clamped);
      }
    };

    const handleMouseUp = () => setIsDragging(false);

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isDragging, snapBottom]);

  const handleSend = () => {
    const trimmed = inputValue.trim();
    if (!trimmed) return;
    onSendMessage(trimmed);
    setInputValue("");
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSend();
    }
  };

  const style: React.CSSProperties = isSnapped
    ? isEndgame
      ? {
          position: "fixed",
          right: 16,
          top: "50%",
          transform: "translateY(-50%)",
          zIndex: Z_INDEX.UI_BASE,
        }
      : {
          position: "fixed",
          bottom: snapBottom,
          right: snapX > 0 ? undefined : 80,
          left: snapX > 0 ? snapX : undefined,
          zIndex: Z_INDEX.UI_BASE,
        }
    : {
        position: "fixed",
        left: position.x,
        top: position.y,
        zIndex: Z_INDEX.UI_BASE,
      };

  return (
    <>
      {isDragging && (
        <div
          className="fixed inset-0"
          style={{ zIndex: Z_INDEX.LOADING_OVERLAY, cursor: "default" }}
        />
      )}
      <div
        ref={containerRef}
        onMouseDown={handleMouseDown}
        className={
          embedded
            ? "w-full min-w-0 min-h-[240px] max-h-[320px] flex-1 flex flex-col border-t border-white/10 pt-3"
            : "w-[416px] select-none bg-white/5"
        }
        style={embedded ? undefined : { ...style, cursor: "default" }}
      >
        <div
          className={`${embedded ? "h-0 min-h-0 flex-1" : "h-[200px]"} overflow-y-auto overflow-x-hidden px-2 py-1 flex flex-col gap-1.5`}
        >
          {messages.map((msg, i) => {
            const time = msg.timestamp ? new Date(msg.timestamp) : null;
            const timeStr = time
              ? `${String(time.getHours()).padStart(2, "0")}:${String(time.getMinutes()).padStart(2, "0")}`
              : "";
            const senderColor =
              (msg.senderId && playerColorMap?.get(msg.senderId)) || msg.senderColor;
            if (msg.kind === "system") {
              return (
                <div key={i} className="shrink-0 text-left text-xs italic text-white/40 min-w-0">
                  {timeStr && (
                    <span className="not-italic text-white/20 text-[10px] mr-1.5">{timeStr}</span>
                  )}
                  {msg.message}
                </div>
              );
            }
            if (msg.kind === "recap") {
              return (
                <div
                  key={i}
                  className="shrink-0 text-left min-w-0 border-l-2 bg-white/5 px-2 py-1.5 my-0.5"
                  style={{ borderLeftColor: senderColor }}
                >
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="font-orbitron text-[9px] font-bold tracking-[0.15em] text-white/50">
                      RECAP
                    </span>
                    <span className="text-sm font-semibold" style={{ color: senderColor }}>
                      {msg.senderName}
                    </span>
                  </div>
                  <p className="text-sm text-white/75 leading-relaxed break-words">{msg.message}</p>
                </div>
              );
            }
            return (
              <div key={i} className="text-sm leading-relaxed flex shrink-0 text-left min-w-0">
                <span className="shrink-0">
                  {timeStr && <span className="text-white/20 text-[10px] mr-1.5">{timeStr}</span>}
                  <span className="font-semibold" style={{ color: senderColor }}>
                    {msg.senderName}
                  </span>
                  {msg.isSpectator && (
                    <span className="text-white/25 text-[10px] ml-0.5">(spectator)</span>
                  )}
                  <span className="text-white/40 mr-2"> </span>
                </span>
                <span className="text-white/70 break-words min-w-0">{msg.message}</span>
              </div>
            );
          })}
          <div ref={messagesEndRef} />
        </div>

        {typingNames.length > 0 && (
          <div className="shrink-0 px-2 pb-1 text-[11px] italic text-white/40 text-left">
            {`${typingNames.join(", ")} ${typingNames.length > 1 ? "are" : "is"} typing…`}
          </div>
        )}

        <div className="border-t border-white/15 shrink-0 flex items-center gap-1">
          {!isSpectator && (
            <div
              ref={emotePickerRef}
              className="relative shrink-0"
              onMouseDown={(e) => e.stopPropagation()}
            >
              <button
                type="button"
                aria-label="Send an emote"
                aria-expanded={showEmotePicker}
                onClick={() => setShowEmotePicker((v) => !v)}
                className={`flex items-center justify-center w-7 h-7 cursor-pointer transition-opacity ${showEmotePicker ? "opacity-100" : "opacity-50 hover:opacity-100"}`}
              >
                <EmoteIcon emote="laugh" size={18} />
              </button>
              {showEmotePicker && (
                <div
                  className="absolute bottom-full left-0 mb-1 w-max grid grid-cols-[repeat(4,2.25rem)] gap-1 p-1.5 bg-[rgba(10,10,15,0.97)] border border-white/20 rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.6)]"
                  style={{ zIndex: Z_INDEX.DROPDOWNS }}
                >
                  {EMOTES.map(({ kind, label }) => (
                    <button
                      key={kind}
                      type="button"
                      aria-label={label}
                      onClick={() => {
                        setShowEmotePicker(false);
                        void globalWebSocketManager.sendEmote(kind);
                      }}
                      className="flex items-center justify-center w-9 h-9 rounded cursor-pointer hover:bg-white/10 hover:scale-110 transition-transform"
                    >
                      <EmoteIcon emote={kind} size={24} />
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Message…"
            aria-label="Chat message"
            spellCheck={false}
            autoComplete="off"
            className="flex-1 min-w-0 bg-transparent text-white text-sm px-1 py-2 outline-none placeholder:text-white/25 border-b border-white/15"
          />
        </div>
      </div>
    </>
  );
};

export default ChatOverlay;
