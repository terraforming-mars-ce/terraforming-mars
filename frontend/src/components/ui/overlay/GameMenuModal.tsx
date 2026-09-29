import CloseButton from "@/components/ui/buttons/CloseButton.tsx";
import React, { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import BackButton from "../buttons/BackButton.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface GameMenuModalProps {
  layout?: "dialog" | "lobby";
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
  layout = "dialog",
  title,
  subtitle,
  children,
  onBack,
  visible,
  onExited,
  showBackdrop = false,
  zIndex = Z_INDEX.MENU_DROPDOWN,
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

  const content = (
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
          data-overlay-layer
          className={`fixed inset-0 bg-black/60 backdrop-blur-sm ${
            animState === "exiting"
              ? "animate-[backdropFadeOut_0.2s_ease-out_forwards]"
              : "animate-[backdropFadeIn_0.3s_ease-out]"
          }`}
          style={{ zIndex: zIndex - 1 }}
          onClick={onClose}
        />
      )}

      {onBack && (
        <div className="fixed top-[30px] left-[30px]" style={{ zIndex: Z_INDEX.POPOVER }}>
          <BackButton onClick={onBack} />
        </div>
      )}
      <div
        data-overlay-layer={showBackdrop || undefined}
        className={`fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 ${layout === "lobby" ? "w-[1100px]" : "w-[450px]"} max-w-[calc(100vw-32px)] ${animationClass}`}
        style={{ zIndex }}
        onAnimationEnd={handleAnimationEnd}
      >
        <div className="relative game-panel p-6 sm:p-8 max-h-[calc(100dvh-100px)] overflow-y-auto">
          {showCloseButton && onClose && (
            <div className="absolute top-3 right-3">
              <CloseButton onClick={onClose} />
            </div>
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

  return showBackdrop ? createPortal(content, document.body) : content;
};

export default GameMenuModal;
