import React from "react";
import { GamePopoverItemProps } from "./types";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

const GamePopoverItem: React.FC<GamePopoverItemProps> = ({
  state,
  onClick,
  error,
  warning,
  statusBadge,
  hoverEffect = "background",
  animationDelay = 0,
  children,
  className = "",
}) => {
  const { playButtonHoverSound, playButtonClickSound } = useSoundEffects();
  const isClickable = state === "available" && onClick;

  const handleClick = () => {
    if (isClickable) {
      void playButtonClickSound();
      onClick?.();
    }
  };

  const getStateClasses = () => {
    switch (state) {
      case "available":
        return `border-[rgba(var(--popover-accent-rgb),0.3)] bg-[rgba(var(--popover-accent-rgb),0.2)] ${
          isClickable
            ? `cursor-pointer ${
                hoverEffect === "translate-x"
                  ? "hover:translate-x-1 hover:shadow-[0_4px_15px_rgba(var(--popover-accent-rgb),0.4)]"
                  : hoverEffect === "glow"
                    ? "hover:shadow-[0_4px_15px_rgba(var(--popover-accent-rgb),0.4)]"
                    : ""
              } hover:border-[var(--popover-accent)] hover:bg-[rgba(var(--popover-accent-rgb),0.3)]`
            : "cursor-default"
        }`;
      case "disabled":
        return "border-[rgba(var(--popover-accent-rgb),0.15)] bg-[rgba(var(--popover-accent-rgb),0.1)] opacity-60 cursor-default";
      case "claimed":
        return "border-[var(--popover-accent)] bg-[rgba(var(--popover-accent-rgb),0.3)] cursor-default";
      default:
        return "";
    }
  };

  return (
    <div
      className={`relative flex items-center gap-3 py-2.5 px-[15px] rounded-lg border transition-all duration-200 animate-[itemSlideIn_0.4s_ease-out_both] max-[768px]:py-2 max-[768px]:px-3 ${getStateClasses()} ${className}`}
      onClick={isClickable ? handleClick : undefined}
      onMouseEnter={isClickable ? () => void playButtonHoverSound() : undefined}
      style={{ animationDelay: `${animationDelay}s` }}
    >
      {error && state === "disabled" && (
        <div className="absolute top-2 right-2 z-[4] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(231,76,60,0.8)] shadow-[0_2px_8px_rgba(231,76,60,0.4)] flex items-center gap-1">
          <span>⚠</span>
          <span>
            {error.message}
            {error.count && error.count > 1 && ` (+${error.count - 1})`}
          </span>
        </div>
      )}

      {warning && state === "available" && (
        <div className="absolute top-2 right-2 z-[4] bg-[linear-gradient(135deg,#f39c12,#e67e22)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(243,156,18,0.8)] shadow-[0_2px_8px_rgba(243,156,18,0.4)] flex items-center gap-1">
          <span>⚠</span>
          <span>{warning.message}</span>
        </div>
      )}

      {statusBadge && (
        <span className="absolute top-2 right-2 text-[10px] text-[var(--popover-accent)] bg-[rgba(var(--popover-accent-rgb),0.3)] px-1.5 py-0.5 rounded border border-[rgba(var(--popover-accent-rgb),0.5)]">
          {statusBadge}
        </span>
      )}

      {children}
    </div>
  );
};

export default GamePopoverItem;
