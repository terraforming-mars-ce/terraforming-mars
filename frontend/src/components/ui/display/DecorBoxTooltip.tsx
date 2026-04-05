import React, { ReactNode } from "react";
import { createPortal } from "react-dom";
import { FormattedDescription } from "./FormattedDescription.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface DecorBoxTooltipProps {
  description?: string | null;
  children?: ReactNode;
  position: { x: number; y: number } | null;
  placement?: "below" | "above";
  cornerSize?: number;
}

const DecorBoxTooltip: React.FC<DecorBoxTooltipProps> = ({
  description,
  children,
  position,
  placement = "below",
  cornerSize = 14,
}) => {
  if ((!description && !children) || !position) return null;

  const paddingClass = placement === "below" ? "pt-1" : "pb-2";
  const translateY = placement === "below" ? "0" : "-100%";

  return createPortal(
    <div
      className={`fixed w-max max-w-40 ${paddingClass} pointer-events-none animate-[fadeIn_150ms_ease-in]`}
      style={{
        left: position.x,
        top: position.y,
        transform: `translate(-50%, ${translateY})`,
        zIndex: Z_INDEX.LOADING_OVERLAY,
      }}
    >
      <div
        className="relative bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-2 shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
        style={{
          clipPath: `polygon(0 0, calc(100% - ${cornerSize}px) 0, 100% ${cornerSize}px, 100% 100%, ${cornerSize}px 100%, 0 calc(100% - ${cornerSize}px))`,
        }}
      >
        {children || <FormattedDescription text={description!} />}
        <svg
          className={`absolute top-0 right-0 pointer-events-none`}
          style={{ width: cornerSize, height: cornerSize }}
          viewBox={`0 0 ${cornerSize} ${cornerSize}`}
        >
          <line
            x1="0"
            y1="0"
            x2={cornerSize}
            y2={cornerSize}
            stroke="rgba(60,60,70,0.7)"
            strokeWidth="1.5"
          />
        </svg>
        <svg
          className={`absolute bottom-0 left-0 pointer-events-none`}
          style={{ width: cornerSize, height: cornerSize }}
          viewBox={`0 0 ${cornerSize} ${cornerSize}`}
        >
          <line
            x1="0"
            y1="0"
            x2={cornerSize}
            y2={cornerSize}
            stroke="rgba(60,60,70,0.7)"
            strokeWidth="1.5"
          />
        </svg>
      </div>
    </div>,
    document.body,
  );
};

export default DecorBoxTooltip;
