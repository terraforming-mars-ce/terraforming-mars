import React, { useState, useEffect } from "react";
import BackButton from "../buttons/BackButton.tsx";

interface GameMenuModalProps {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  onBack?: () => void;
  visible?: boolean;
  onExited?: () => void;
  showBackdrop?: boolean;
  zIndex?: number;
  onClose?: () => void;
  showCloseButton?: boolean;
}

const GameMenuModal: React.FC<GameMenuModalProps> = ({
  title,
  subtitle,
  children,
  onBack,
  visible,
  onExited,
  showBackdrop = false,
  zIndex = 1000,
  onClose,
  showCloseButton = false,
}) => {
  const [animState, setAnimState] = useState<"entering" | "visible" | "exiting">("entering");

  useEffect(() => {
    if (visible === false) {
      setAnimState("exiting");
    } else {
      setAnimState("entering");
    }
  }, [visible]);

  const animationClass =
    animState === "exiting"
      ? "animate-[lobbyExit_0.2s_ease-out_forwards]"
      : animState === "entering"
        ? "animate-[modalFadeIn_0.3s_ease-out]"
        : "";

  const handleAnimationEnd = (e: React.AnimationEvent) => {
    if (e.target !== e.currentTarget) return;
    if (animState === "entering") setAnimState("visible");
    if (animState === "exiting") onExited?.();
  };

  return (
    <>
      <style>{`
        @keyframes modalFadeIn {
          0% { opacity: 0; }
          100% { opacity: 1; }
        }
        @keyframes lobbyExit {
          from {
            transform: scale(1);
            opacity: 1;
          }
          to {
            transform: scale(0.95);
            opacity: 0;
          }
        }
        @keyframes backdropFadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        @keyframes backdropFadeOut {
          from { opacity: 1; }
          to { opacity: 0; }
        }
      `}</style>

      {showBackdrop && (
        <div
          className={`fixed inset-0 bg-black/60 backdrop-blur-sm ${
            animState === "exiting"
              ? "animate-[backdropFadeOut_0.2s_ease-out_forwards]"
              : "animate-[backdropFadeIn_0.3s_ease-out]"
          }`}
          style={{ zIndex: zIndex - 1 }}
          onClick={onClose}
        />
      )}

      {onBack && <BackButton onClick={onBack} className="fixed top-[30px] left-[30px] z-[10000]" />}
      <div
        className={`fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[450px] max-w-[90vw] ${animationClass}`}
        style={{ zIndex }}
        onAnimationEnd={handleAnimationEnd}
      >
        <div className="relative bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] p-8 backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)] max-h-[90vh] overflow-y-auto">
          {showCloseButton && onClose && (
            <button
              className="absolute top-4 right-4 text-white/70 hover:text-white text-xl leading-none transition-colors"
              onClick={onClose}
            >
              ×
            </button>
          )}
          <div className="text-center mb-6">
            <h2 className="font-orbitron text-white text-[24px] m-0 mb-2 text-shadow-glow font-bold tracking-wider">
              {title}
            </h2>
            {subtitle && <p className="text-white/60 text-sm m-0">{subtitle}</p>}
          </div>
          {children}
        </div>
      </div>
    </>
  );
};

export default GameMenuModal;
