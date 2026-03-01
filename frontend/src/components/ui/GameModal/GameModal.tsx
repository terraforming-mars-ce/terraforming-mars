import React from "react";
import { GameModalProps, ModalSize } from "./types";
import { getThemeStyles } from "./themes";
import { useModal } from "./useModal";

const sizeClasses: Record<ModalSize, string> = {
  small: "max-w-[600px]",
  medium: "max-w-[900px]",
  large: "max-w-[1200px]",
  full: "max-w-[1400px]",
};

const GameModal: React.FC<GameModalProps> = ({
  isVisible,
  onClose,
  theme,
  size = "large",
  animation = "slideIn",
  zIndex = 3000,
  closeOnBackdrop = true,
  closeOnEscape = true,
  lockScroll = true,
  preventClose = false,
  onPreventedClose,
  glow = true,
  outerContent,
  children,
  className = "",
}) => {
  useModal({
    isVisible,
    onClose,
    closeOnEscape,
    lockScroll,
    preventClose,
    onPreventedClose,
  });

  if (!isVisible) return null;

  const themeStyles = getThemeStyles(theme);

  const handleBackdropClick = () => {
    if (preventClose) {
      onPreventedClose?.();
    } else if (closeOnBackdrop) {
      onClose();
    }
  };

  const animationClass =
    animation === "slideIn"
      ? "animate-[modalSlideIn_0.25s_ease-out]"
      : animation === "fadeIn"
        ? "animate-[modalFadeIn_0.3s_ease-out]"
        : "";

  const modalBox = (
    <div
      className={`relative w-full ${sizeClasses[size]} max-h-[90vh] bg-space-black-darker/95 border-2 border-[var(--modal-accent)] rounded-[20px] overflow-hidden ${glow ? "shadow-[0_10px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(var(--modal-accent-rgb),0.5)]" : "shadow-[0_10px_40px_rgba(0,0,0,0.8)]"} backdrop-blur-space ${animationClass} flex flex-col ${className}`}
      style={themeStyles}
    >
      {children}
    </div>
  );

  return (
    <div
      className="fixed top-0 left-0 right-0 bottom-0 flex items-center justify-center p-5"
      style={{ zIndex }}
    >
      <div
        className="absolute top-0 left-0 right-0 bottom-0 bg-black/60 cursor-default animate-[modalFadeIn_0.3s_ease-out]"
        onClick={handleBackdropClick}
      />

      {outerContent ? (
        <div className="relative flex items-center w-fit max-w-full">
          {modalBox}
          {outerContent}
        </div>
      ) : (
        modalBox
      )}
    </div>
  );
};

export default GameModal;
