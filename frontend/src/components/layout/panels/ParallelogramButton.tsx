import React, { useState } from "react";
import { useHoverSound } from "@/hooks/useHoverSound.ts";

export const ANGLE_INDENT = 20;
export const BUTTON_SPACING = 6;
const BORDER_COLOR = "rgba(60,60,70,0.7)";

export type EdgeStyle = "slope-left" | "slope-right" | "flat";

export interface ParallelogramButtonProps {
  width: number;
  height: number;
  color: string;
  children?: React.ReactNode;
  onClick?: () => void;
  buttonRef?: React.RefObject<HTMLButtonElement | null>;
  isActive?: boolean;
  leftEdge?: EdgeStyle;
  rightEdge?: EdgeStyle;
  className?: string;
  style?: React.CSSProperties;
}

const ParallelogramButton: React.FC<ParallelogramButtonProps> = ({
  width,
  height,
  color,
  children,
  onClick,
  buttonRef,
  isActive = false,
  leftEdge: left = "flat",
  rightEdge: right = "flat",
  className = "",
  style: extraStyle,
}) => {
  const [isHovered, setIsHovered] = useState(false);
  const hoverSound = useHoverSound();
  const active = isHovered || isActive;
  const isClickable = typeof onClick === "function";

  const ai = ANGLE_INDENT;
  const w = width;
  const h = height;

  // slope-left (\): top indented, bottom at edge
  // slope-right (/): top at edge, bottom indented
  const tl = left === "slope-left" ? ai : 0;
  const bl = left === "slope-right" ? ai : 0;
  const tr = right === "slope-left" ? w - ai : w;
  const br = right === "slope-right" ? w - ai : w;

  const fillPoints = `${tl},0 ${tr},0 ${br},${h} ${bl},${h}`;
  const contentOffsetX = (tl + tr + br + bl) / 4 - w / 2;

  const edges: Array<{ x1: number; y1: number; x2: number; y2: number }> = [];
  if (left === "slope-left") {
    edges.push({ x1: ai, y1: 0, x2: 0, y2: h });
  } else if (left === "slope-right") {
    edges.push({ x1: 0, y1: 0, x2: ai, y2: h });
  } else {
    edges.push({ x1: 0, y1: 0, x2: 0, y2: h });
  }
  if (right === "slope-left") {
    edges.push({ x1: w - ai, y1: 0, x2: w, y2: h });
  } else if (right === "slope-right") {
    edges.push({ x1: w, y1: 0, x2: w - ai, y2: h });
  } else {
    edges.push({ x1: w, y1: 0, x2: w, y2: h });
  }

  const handleClick = () => {
    if (!isClickable) {
      return;
    }
    hoverSound.onClick?.();
    onClick();
  };

  const handleMouseEnter = () => {
    setIsHovered(true);
    if (isClickable) {
      hoverSound.onMouseEnter?.();
    }
  };

  const cursorClass = isClickable ? "cursor-pointer" : "cursor-default";

  return (
    <button
      ref={buttonRef}
      onClick={handleClick}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={() => setIsHovered(false)}
      className={`relative pointer-events-auto ${cursorClass} outline-none ${className}`}
      style={{ width, height, ...extraStyle }}
    >
      <svg
        className="absolute inset-0 w-full h-full"
        viewBox={`0 0 ${w} ${h}`}
        preserveAspectRatio="none"
      >
        <polygon
          points={fillPoints}
          fill={active ? "rgba(20,20,25,0.95)" : "rgba(10,10,15,0.95)"}
        />
        <line
          x1={tl}
          y1={0}
          x2={tr}
          y2={0}
          stroke={active ? color : BORDER_COLOR}
          strokeWidth="3"
        />
        {edges.map((e, i) => (
          <line
            key={i}
            x1={e.x1}
            y1={e.y1}
            x2={e.x2}
            y2={e.y2}
            stroke={BORDER_COLOR}
            strokeWidth="2"
          />
        ))}
      </svg>
      <div
        className="relative z-10 h-full flex items-center justify-center px-4"
        style={{ transform: contentOffsetX ? `translateX(${contentOffsetX}px)` : undefined }}
      >
        <span
          className={`font-orbitron font-bold text-sm uppercase tracking-wider inline-flex items-center justify-center leading-none transition-colors duration-200 ${
            active ? "text-white" : "text-white/80"
          }`}
        >
          {children}
        </span>
      </div>
    </button>
  );
};

export default ParallelogramButton;
