import { useEffect, useRef, useState } from "react";
import { useProgress } from "@react-three/drei";
import { Z_INDEX } from "@/constants/zIndex";

/** Real asset progress fills this much of the bar; the rest covers scene build and GPU work. */
const ASSET_SHARE = 90;
/** The time-based estimate approaches ESTIMATE_CEILING with this time constant. */
const ESTIMATE_TIME_CONSTANT_MS = 6000;
const ESTIMATE_CEILING = 95;
const TICK_MS = 100;

/**
 * Percentage for the loading screen: the larger of three.js loader progress and a
 * time-based estimate, eased so it keeps creeping forward and never goes back.
 * Holds below 100 until the load is done.
 */
function useEstimatedProgress(done: boolean) {
  const [shown, setShown] = useState(0);

  useEffect(() => {
    if (done) {
      setShown(100);
      return;
    }
    const startedAt = performance.now();
    const id = window.setInterval(() => {
      const { loaded, total } = useProgress.getState();
      const assets = total > 0 ? (loaded / total) * ASSET_SHARE : 0;
      const elapsed = performance.now() - startedAt;
      const estimate = ESTIMATE_CEILING * (1 - Math.exp(-elapsed / ESTIMATE_TIME_CONSTANT_MS));
      const target = Math.min(99, Math.max(assets, estimate));
      setShown((previous) => Math.max(previous, previous + (target - previous) * 0.25));
    }, TICK_MS);
    return () => window.clearInterval(id);
  }, [done]);

  return shown;
}

function ProgressReadout({ done }: { done: boolean }) {
  const progress = useEstimatedProgress(done);
  return (
    <div style={{ width: "220px", marginBottom: "16px", textAlign: "center" }}>
      <div style={{ fontSize: "32px", fontWeight: 700, marginBottom: "10px" }}>
        {Math.floor(progress)}%
      </div>
      <div
        style={{
          height: "3px",
          backgroundColor: "rgba(255, 255, 255, 0.1)",
          overflow: "hidden",
        }}
      >
        <div
          style={{
            height: "100%",
            width: `${progress}%`,
            backgroundColor: "white",
            transition: `width ${TICK_MS}ms linear`,
          }}
        />
      </div>
    </div>
  );
}

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
  fadeDurationMs?: number;
  /**
   * Show an estimated percentage and bar instead of the spinner, with a single detail line
   * below it: the subtitle, or the message when there is no subtitle.
   */
  showProgress?: boolean;
}

type Phase = "waiting" | "showing" | "fading" | "done";

export default function LoadingOverlay({
  message = "Loading",
  subtitle,
  isLoaded,
  onTransitionEnd,
  showDelayMs = 500,
  minDurationMs = 200,
  fadeDurationMs = 800,
  showProgress = false,
}: LoadingOverlayProps) {
  const [phase, setPhase] = useState<Phase>(() => (showDelayMs === 0 ? "showing" : "waiting"));
  const shownAtRef = useRef<number | null>(showDelayMs === 0 ? Date.now() : null);
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
      duration: fadeDurationMs,
      easing: "ease-out",
      fill: "forwards",
    });
    void animation.finished.then(
      () => {
        setPhase("done");
        onTransitionEnd?.();
      },
      () => {},
    );
    return () => animation.cancel();
  }, [phase, onTransitionEnd, fadeDurationMs]);

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
      {showProgress ? (
        <ProgressReadout done={isLoaded} />
      ) : (
        <>
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
        </>
      )}
      {!showProgress && message}
      {(subtitle || (showProgress && message)) && (
        <div
          style={{
            marginTop: "8px",
            fontSize: "13px",
            color: "rgba(255, 255, 255, 0.4)",
            fontFamily: "Orbitron, sans-serif",
          }}
        >
          {subtitle || message}
        </div>
      )}
    </div>
  );
}
