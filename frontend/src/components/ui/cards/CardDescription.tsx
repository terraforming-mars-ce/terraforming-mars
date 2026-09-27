import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import { FormattedDescription } from "../display/FormattedDescription.tsx";
import GameButton from "../buttons/GameButton.tsx";

interface CardDescriptionProps {
  text: string;
  maxHeight?: number;
  autoPan: boolean;
}

export default function CardDescription({ text, maxHeight, autoPan }: CardDescriptionProps) {
  const viewportRef = useRef<HTMLDivElement>(null);
  const textRef = useRef<HTMLDivElement>(null);
  const reducedMotion = useReducedMotion();
  const [overflow, setOverflow] = useState(false);
  const [status, setStatus] = useState<"waiting" | "running" | "paused" | "finished">("waiting");

  useLayoutEffect(() => {
    const viewport = viewportRef.current;
    const content = textRef.current;
    if (!viewport || !content) {
      return;
    }
    const measure = () => {
      const end = Math.max(0, viewport.scrollHeight - viewport.clientHeight);
      viewport.scrollTop = Math.min(viewport.scrollTop, end);
      setOverflow(end > 1);
    };
    const observer = new ResizeObserver(measure);
    observer.observe(viewport);
    observer.observe(content);
    measure();
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const viewport = viewportRef.current;
    if (
      !viewport ||
      !overflow ||
      !autoPan ||
      reducedMotion ||
      status === "paused" ||
      status === "finished"
    ) {
      return;
    }
    let frame = 0;
    let previous: number | null = null;
    let delay = status === "waiting" ? 1000 : 0;
    let position = viewport.scrollTop;
    const resetClock = () => {
      previous = null;
    };
    const tick = (now: number) => {
      if (document.hidden) {
        previous = null;
      } else {
        const elapsed = previous === null ? 0 : Math.min(now - previous, 100);
        previous = now;
        if (delay > 0) {
          delay -= elapsed;
        } else {
          const end = Math.max(0, viewport.scrollHeight - viewport.clientHeight);
          position = Math.min(end, position + elapsed * 0.012);
          viewport.scrollTop = position;
          if (position >= end) {
            setStatus("finished");
            return;
          }
        }
      }
      frame = requestAnimationFrame(tick);
    };
    document.addEventListener("visibilitychange", resetClock);
    frame = requestAnimationFrame(tick);
    return () => {
      cancelAnimationFrame(frame);
      document.removeEventListener("visibilitychange", resetClock);
    };
  }, [autoPan, overflow, reducedMotion, status]);

  const stop = () => setStatus("paused");
  const running = status === "waiting" || status === "running";
  return (
    <div>
      {overflow && autoPan && !reducedMotion && status !== "finished" && (
        <div className="flex justify-end pb-1">
          <GameButton
            emphasis="quiet"
            size="xs"
            onClick={() => setStatus(running ? "paused" : "running")}
            aria-label={running ? "Pause description scrolling" : "Resume description scrolling"}
          >
            {running ? "Pause" : "Resume"}
          </GameButton>
        </div>
      )}
      <div
        ref={viewportRef}
        role="region"
        aria-label="Card description"
        tabIndex={overflow ? 0 : undefined}
        className={
          autoPan
            ? "touch-none overflow-y-auto overscroll-contain [scrollbar-width:thin] focus-visible:outline focus-visible:outline-blue-400"
            : "touch-none"
        }
        style={{ maxHeight: autoPan ? maxHeight : undefined }}
        onWheel={stop}
        onPointerDown={stop}
        onTouchStart={stop}
        onFocus={stop}
        onKeyDown={(event) => {
          if (
            ["ArrowUp", "ArrowDown", "PageUp", "PageDown", "Home", "End", " "].includes(event.key)
          ) {
            stop();
          }
        }}
      >
        <div ref={textRef}>
          <FormattedDescription text={text} />
        </div>
      </div>
    </div>
  );
}
