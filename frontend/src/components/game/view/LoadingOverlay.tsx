import { useState, useEffect, useRef } from "react";
import { Z_INDEX } from "@/constants/zIndex";

interface LoadingOverlayProps {
  message?: string;
  subtitle?: string;
  isLoaded: boolean;
  onTransitionEnd?: () => void;
}

export default function LoadingOverlay({
  message = "Loading",
  subtitle,
  isLoaded,
  onTransitionEnd,
}: LoadingOverlayProps) {
  const [hasMounted, setHasMounted] = useState(false);
  const [fadeComplete, setFadeComplete] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    requestAnimationFrame(() => {
      setHasMounted(true);
    });
  }, []);

  const showLoaded = hasMounted && isLoaded;

  useEffect(() => {
    if (!showLoaded || !ref.current) {
      return;
    }

    const animation = ref.current.animate([{ opacity: 1 }, { opacity: 0 }], {
      duration: 800,
      easing: "ease-out",
      fill: "forwards",
    });

    void animation.finished.then(() => {
      setFadeComplete(true);
      onTransitionEnd?.();
    });

    return () => {
      animation.cancel();
    };
  }, [showLoaded, onTransitionEnd]);

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
        opacity: fadeComplete ? 0 : 1,
        pointerEvents: showLoaded ? "none" : "auto",
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
