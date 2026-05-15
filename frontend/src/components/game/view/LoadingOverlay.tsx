import { useEffect, useRef, useState } from "react";
import { Z_INDEX } from "@/constants/zIndex";

interface LoadingOverlayProps {
  message?: string;
  subtitle?: string;
  isLoaded: boolean;
  onTransitionEnd?: () => void;
  /**
   * Wait this many milliseconds before showing the overlay. If the load
   * finishes before this elapses, the overlay is skipped entirely so fast
   * transitions don't briefly flash a black panel.
   */
  showDelayMs?: number;
  /**
   * Once shown, keep the overlay visible at least this long even if the
   * load finishes quickly. Prevents a jarring pop-out on medium-speed loads.
   */
  minDurationMs?: number;
}

type Phase = "waiting" | "showing" | "fading" | "done";

export default function LoadingOverlay({
  message = "Loading",
  subtitle,
  isLoaded,
  onTransitionEnd,
  showDelayMs = 500,
  minDurationMs = 200,
}: LoadingOverlayProps) {
  const [phase, setPhase] = useState<Phase>("waiting");
  const shownAtRef = useRef<number | null>(null);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (phase !== "waiting") {
      return;
    }
    const id = window.setTimeout(() => {
      shownAtRef.current = Date.now();
      setPhase("showing");
    }, showDelayMs);
    return () => window.clearTimeout(id);
  }, [phase, showDelayMs]);

  useEffect(() => {
    if (!isLoaded) {
      return;
    }
    if (phase === "waiting") {
      setPhase("done");
      onTransitionEnd?.();
      return;
    }
    if (phase !== "showing") {
      return;
    }
    const shownAt = shownAtRef.current ?? Date.now();
    const elapsed = Date.now() - shownAt;
    const remaining = Math.max(0, minDurationMs - elapsed);
    if (remaining === 0) {
      setPhase("fading");
      return;
    }
    const id = window.setTimeout(() => setPhase("fading"), remaining);
    return () => window.clearTimeout(id);
  }, [isLoaded, phase, minDurationMs, onTransitionEnd]);

  useEffect(() => {
    if (phase !== "fading" || !ref.current) {
      return;
    }
    const animation = ref.current.animate([{ opacity: 1 }, { opacity: 0 }], {
      duration: 800,
      easing: "ease-out",
      fill: "forwards",
    });
    void animation.finished.then(() => {
      setPhase("done");
      onTransitionEnd?.();
    });
    return () => animation.cancel();
  }, [phase, onTransitionEnd]);

  useEffect(() => {
    if (!isLoaded && phase === "done") {
      shownAtRef.current = null;
      setPhase("waiting");
    }
  }, [isLoaded, phase]);

  if (phase === "waiting" || phase === "done") {
    return null;
  }

  return (
    <div
      ref={ref}
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        width: "100vw",
        height: "100vh",
        backgroundColor: "#000000",
        zIndex: Z_INDEX.LOADING_OVERLAY,
        opacity: 1,
        pointerEvents: phase === "fading" ? "none" : "auto",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        color: "white",
        fontSize: "18px",
        fontFamily: "Orbitron, sans-serif",
      }}
    >
      <div
        style={{
          width: "40px",
          height: "40px",
          border: "4px solid rgba(255, 255, 255, 0.1)",
          borderTop: "4px solid white",
          borderRadius: "50%",
          animation: "spin 1s linear infinite",
          marginBottom: "16px",
        }}
      />
      <style>
        {`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `}
      </style>
      {message}
      {subtitle && (
        <div
          style={{
            marginTop: "8px",
            fontSize: "13px",
            color: "rgba(255, 255, 255, 0.4)",
            fontFamily: "Orbitron, sans-serif",
          }}
        >
          {subtitle}
        </div>
      )}
    </div>
  );
}
