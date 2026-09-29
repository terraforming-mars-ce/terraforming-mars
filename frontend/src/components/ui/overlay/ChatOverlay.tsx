import React, { useState, useRef, useEffect, useCallback, useLayoutEffect } from "react";
import { ChatMessageDto } from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex";

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
            return (
              <div key={i} className="text-sm leading-relaxed flex shrink-0 text-left min-w-0">
                <span className="shrink-0">
                  {timeStr && <span className="text-white/20 text-[10px] mr-1.5">{timeStr}</span>}
                  <span
                    className="font-semibold"
                    style={{
                      color: (msg.senderId && playerColorMap?.get(msg.senderId)) || msg.senderColor,
                    }}
                  >
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

        <div className="border-t border-white/15 shrink-0">
          <input
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Message…"
            aria-label="Chat message"
            spellCheck={false}
            autoComplete="off"
            className="w-full bg-transparent text-white text-sm px-1 py-2 outline-none placeholder:text-white/25 border-b border-white/15"
          />
        </div>
      </div>
    </>
  );
};

export default ChatOverlay;
