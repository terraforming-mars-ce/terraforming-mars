import React, { useRef, useState, useEffect } from "react";
import { createPortal } from "react-dom";
import { GamePopoverProps } from "./types";
import { getThemeStyles } from "./themes";
import { usePopover } from "./usePopover";
import { Z_INDEX } from "@/constants/zIndex";

const GamePopover: React.FC<GamePopoverProps> = ({
  isVisible,
  onClose,
  position,
  theme,
  header,
  arrow,
  width = 320,
  maxHeight = 400,
  zIndex = Z_INDEX.POPOVER,
  animation = "slideUp",
  children,
  className = "",
  excludeRef,
  contentRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [computedPosition, setComputedPosition] = useState<{
    top?: number;
    left?: number;
    right?: number;
    bottom?: number;
  }>({});

  const anchorRef = position.type === "anchor" ? position.anchorRef : excludeRef;

  usePopover({
    isVisible,
    onClose,
    popoverRef,
    anchorRef,
  });

  useEffect(() => {
    if (!isVisible) return;

    if (position.type === "anchor" && position.anchorRef.current) {
      const rect = position.anchorRef.current.getBoundingClientRect();
      const padding = 30;
      const popoverWidth = typeof width === "number" ? width : 320;

      if (position.placement === "above") {
        const bottom = window.innerHeight - rect.top + 15;
        const right = Math.max(padding, window.innerWidth - rect.right);
        setComputedPosition({ bottom, right });
      } else {
        const top = rect.bottom + 15;
        const wouldOverflowRight = rect.left + popoverWidth > window.innerWidth - padding;

        if (wouldOverflowRight) {
          const right = Math.max(padding, window.innerWidth - rect.right);
          setComputedPosition({ top, right });
        } else {
          const left = Math.max(padding, rect.left);
          setComputedPosition({ top, left });
        }
      }
    } else if (position.type === "fixed") {
      setComputedPosition({
        top: position.top,
        left: position.left,
        right: position.right,
        bottom: position.bottom,
      });
    }
  }, [isVisible, position, width]);

  if (!isVisible) return null;

  const themeStyles = getThemeStyles(theme);
  const animationClass =
    animation === "slideUp"
      ? "animate-[popoverSlideUp_0.3s_ease-out]"
      : "animate-[popoverSlideDown_0.3s_ease-out]";

  const widthStyle = typeof width === "number" ? `${width}px` : width;
  const maxHeightStyle = typeof maxHeight === "number" ? `${maxHeight}px` : maxHeight;

  const getArrowPosition = () => {
    if (!arrow?.enabled) return "";
    const offset = arrow.offset ?? 30;
    switch (arrow.position) {
      case "left":
        return `left-[${offset}px]`;
      case "center":
        return "left-1/2 -translate-x-1/2";
      case "right":
      default:
        return `right-[${offset}px]`;
    }
  };

  return createPortal(
    <div
      ref={popoverRef}
      className={`fixed bg-space-black-darker/95 border-2 border-[var(--popover-accent)] rounded-xl shadow-[0_10px_40px_rgba(0,0,0,0.8),0_0_20px_rgba(var(--popover-accent-rgb),0.5)] backdrop-blur-space ${animationClass} flex flex-col overflow-hidden isolate pointer-events-auto max-[768px]:w-[280px] ${className}`}
      style={{
        ...themeStyles,
        ...computedPosition,
        width: widthStyle,
        maxHeight: maxHeightStyle,
        zIndex,
      }}
    >
      {arrow?.enabled && (
        <div
          className={`absolute -bottom-2 ${getArrowPosition()} w-0 h-0 border-l-[8px] border-l-transparent border-r-[8px] border-r-transparent border-t-[8px] border-t-[var(--popover-accent)]`}
          style={arrow.offset !== undefined ? { right: `${arrow.offset}px` } : undefined}
        />
      )}

      {header && (
        <div className="flex items-center justify-between py-[15px] px-5 bg-black/40 border-b border-b-[var(--popover-accent)]/60">
          <div className="flex items-center gap-2.5">
            <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
              {header.title}
            </h3>
            {header.badge && (
              <div className="text-white/80 text-xs bg-[rgba(var(--popover-accent-rgb),0.2)] py-1 px-2 rounded-md border border-[rgba(var(--popover-accent-rgb),0.3)]">
                {header.badge}
              </div>
            )}
          </div>
          <div className="flex items-center gap-2">
            {header.rightContent}
            {header.showCloseButton && (
              <button
                className="text-white/70 hover:text-white text-xl leading-none transition-colors"
                onClick={onClose}
              >
                ×
              </button>
            )}
          </div>
        </div>
      )}

      <div
        ref={contentRef}
        className="flex-1 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:var(--popover-accent)_rgba(30,60,150,0.3)] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-[rgba(30,60,150,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[var(--popover-accent)]/70 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[var(--popover-accent)]"
      >
        {children}
      </div>
    </div>,
    document.body,
  );
};

export default GamePopover;
