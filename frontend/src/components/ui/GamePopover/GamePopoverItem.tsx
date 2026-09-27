import React from "react";
import { GamePopoverItemProps } from "./types";
import { useSoundEffects } from "@/hooks/useSoundEffects.ts";

const GamePopoverItem: React.FC<GamePopoverItemProps> = ({
  state,
  onClick,
  error,
  warning,
  info,
  statusBadge,
  animationDelay = 0,
  children,
  className = "",
  borderColor,
  style: externalStyle,
}) => {
  const { playButtonHoverSound, playButtonClickSound } = useSoundEffects();
  const isClickable = state === "available" && onClick;
  const hasError = error && state === "disabled";
  let contentLayout = "";
  if (!borderColor) {
    contentLayout = hasError ? "flex flex-col items-stretch gap-2" : "flex items-center gap-3";
  }
  let surfaceOpacity = 0.2;
  let borderOpacity = 0.3;
  if (state === "disabled") {
    surfaceOpacity = 0.1;
    borderOpacity = 0.15;
  } else if (state === "claimed") {
    surfaceOpacity = 0.3;
    borderOpacity = 1;
  }
  const itemBorder = borderColor
    ? borderColor + "60"
    : `rgba(var(--popover-accent-rgb),${borderOpacity})`;
  const itemSurface = borderColor
    ? "#050506"
    : `linear-gradient(rgba(var(--popover-accent-rgb),${surfaceOpacity}), rgba(var(--popover-accent-rgb),${surfaceOpacity})), #050506`;

  const handleClick = () => {
    if (isClickable) {
      void playButtonClickSound();
      onClick?.();
    }
  };

  const getStateClasses = () => {
    if (borderColor) {
      return `${state === "disabled" ? "opacity-50" : ""} ${
        isClickable ? "cursor-pointer hover:brightness-125" : "cursor-default"
      }`;
    }

    switch (state) {
      case "available":
        return `border-[rgba(var(--popover-accent-rgb),0.3)] bg-[rgba(var(--popover-accent-rgb),0.2)] ${
          isClickable ? "cursor-pointer hover:brightness-125" : "cursor-default"
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
      className={`relative ${contentLayout} py-2.5 px-[15px] rounded-none border transition-all duration-200 animate-[itemSlideIn_0.4s_ease-out_both] max-[768px]:py-2 max-[768px]:px-3 ${getStateClasses()} ${className}`}
      role={isClickable ? "button" : undefined}
      tabIndex={isClickable ? 0 : undefined}
      onKeyDown={(event) => {
        if (
          event.target === event.currentTarget &&
          (event.key === "Enter" || event.key === " ") &&
          isClickable
        ) {
          event.preventDefault();
          handleClick();
        }
      }}
      onClick={isClickable ? handleClick : undefined}
      onMouseEnter={isClickable ? () => void playButtonHoverSound() : undefined}
      style={
        {
          animationDelay: `${animationDelay}s`,
          ...(borderColor ? { borderColor: borderColor + "60" } : {}),
          "--item-border": externalStyle?.borderColor ?? itemBorder,
          "--item-surface": externalStyle?.background ?? itemSurface,
          ...externalStyle,
        } as React.CSSProperties
      }
    >
      {hasError && (
        <div className={`flex justify-end ${borderColor ? "mb-2" : ""}`}>
          <span className="popover-status popover-status-error relative">
            {error.message}
            {error.count && error.count > 1 && ` (+${error.count - 1})`}
          </span>
        </div>
      )}

      {warning && state === "available" && (
        <div className="absolute top-2 right-2 z-[4] bg-[linear-gradient(135deg,#f39c12,#e67e22)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(243,156,18,0.8)] shadow-[0_2px_8px_rgba(243,156,18,0.4)] flex items-center gap-1">
          <span>{warning.message}</span>
        </div>
      )}

      {info && state === "disabled" && !error && (
        <div className="absolute top-2 right-2 z-[4] bg-space-black-darker/90 text-white/60 text-[9px] font-bold px-2 py-1 rounded border border-white/20 flex items-center gap-1">
          <span>{info.message}</span>
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
