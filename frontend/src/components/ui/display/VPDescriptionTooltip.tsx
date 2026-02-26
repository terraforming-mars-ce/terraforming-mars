import React from "react";
import { createPortal } from "react-dom";
import { FormattedDescription } from "./FormattedDescription.tsx";

interface VPDescriptionTooltipProps {
  description: string | null;
  position: { x: number; y: number } | null;
}

const VPDescriptionTooltip: React.FC<VPDescriptionTooltipProps> = ({ description, position }) => {
  if (!description || !position) return null;

  return createPortal(
    <div
      className="fixed w-max max-w-40 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: position.x, top: position.y, zIndex: 99999 }}
    >
      <div
        className="relative bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-2 shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
        style={{
          clipPath:
            "polygon(0 0, calc(100% - 14px) 0, 100% 14px, 100% 100%, 14px 100%, 0 calc(100% - 14px))",
        }}
      >
        <FormattedDescription text={description} />
        <svg
          className="absolute top-0 right-0 w-[14px] h-[14px] pointer-events-none"
          viewBox="0 0 14 14"
        >
          <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
        <svg
          className="absolute bottom-0 left-0 w-[14px] h-[14px] pointer-events-none"
          viewBox="0 0 14 14"
        >
          <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
        </svg>
      </div>
    </div>,
    document.body,
  );
};

export default VPDescriptionTooltip;
