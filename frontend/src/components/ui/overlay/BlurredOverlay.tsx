import React from "react";

interface BlurredOverlayProps {
  visible: boolean;
  onClose: () => void;
  zIndex?: number;
  children: React.ReactNode;
}

export default function BlurredOverlay({
  visible,
  onClose,
  zIndex = 1099,
  children,
}: BlurredOverlayProps) {
  return (
    <>
      <div
        className="fixed inset-0 transition-opacity duration-[245ms] ease-in-out"
        style={{
          background: "rgba(0, 0, 0, 0.4)",
          backdropFilter: "blur(6px)",
          zIndex,
          opacity: visible ? 1 : 0,
          pointerEvents: visible ? "auto" : "none",
          cursor: "default",
        }}
        onClick={onClose}
      />
      <div
        className="fixed inset-0"
        style={{
          zIndex: zIndex + 1,
          pointerEvents: "none",
          opacity: visible ? 1 : 0,
          transition: "opacity 245ms ease",
        }}
      >
        {children}
      </div>
    </>
  );
}
