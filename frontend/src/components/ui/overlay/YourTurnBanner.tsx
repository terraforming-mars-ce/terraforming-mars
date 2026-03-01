import { useEffect } from "react";
import { createPortal } from "react-dom";

interface YourTurnBannerProps {
  visible: boolean;
  onDismiss: () => void;
}

export default function YourTurnBanner({ visible, onDismiss }: YourTurnBannerProps) {
  useEffect(() => {
    if (!visible) return;
    const timer = setTimeout(onDismiss, 2500);
    return () => clearTimeout(timer);
  }, [visible, onDismiss]);

  if (!visible) return null;

  return createPortal(
    <div
      className="fixed inset-0 flex items-center justify-center pointer-events-none"
      style={{ zIndex: 2000 }}
    >
      <div className="font-orbitron text-5xl font-black text-white tracking-[0.2em] animate-[yourTurnIn_0.4s_ease-out_forwards,yourTurnOut_0.4s_ease-in_2.1s_forwards] [text-shadow:0_0_40px_rgba(100,200,255,0.8),0_0_80px_rgba(100,200,255,0.4),0_2px_4px_rgba(0,0,0,0.9)]">
        YOUR TURN
      </div>
    </div>,
    document.body,
  );
}
