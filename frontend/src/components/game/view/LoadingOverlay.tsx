import { useState, useEffect } from "react";
import { Z_INDEX } from "@/constants/zIndex";

interface LoadingOverlayProps {
  message?: string;
  isLoaded: boolean;
  onTransitionEnd?: () => void;
}

export default function LoadingOverlay({
  message = "Loading...",
  isLoaded,
  onTransitionEnd,
}: LoadingOverlayProps) {
  const [hasMounted, setHasMounted] = useState(false);

  useEffect(() => {
    requestAnimationFrame(() => {
      setHasMounted(true);
    });
  }, []);

  const showLoaded = hasMounted && isLoaded;

  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        width: "100vw",
        height: "100vh",
        backgroundColor: "#000000",
        zIndex: Z_INDEX.LOADING_OVERLAY,
        opacity: showLoaded ? 0 : 1,
        transition: "opacity 0.8s ease-out",
        pointerEvents: showLoaded ? "none" : "auto",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        color: "white",
        fontSize: "18px",
      }}
      onTransitionEnd={(e) => {
        if (e.propertyName === "opacity" && showLoaded) {
          onTransitionEnd?.();
        }
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
    </div>
  );
}
