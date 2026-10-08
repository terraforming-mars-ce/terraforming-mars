import React, { useEffect, useRef } from "react";
import { createPortal } from "react-dom";
import TileInfoContent, { type TileTooltipData } from "./TileInfoContent.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

const POINTER_OFFSET = 12;
const VIEWPORT_MARGIN = 8;

function clampToViewport(pointer: number, size: number, viewport: number): number {
  const after = pointer + POINTER_OFFSET;
  if (after + size + VIEWPORT_MARGIN <= viewport) {
    return after;
  }
  const before = pointer - POINTER_OFFSET - size;
  return Math.max(VIEWPORT_MARGIN, Math.min(before, viewport - size - VIEWPORT_MARGIN));
}

interface TileTooltipProps {
  data: TileTooltipData | null;
  positionRef: React.RefObject<{ x: number; y: number }>;
}

const TileTooltip: React.FC<TileTooltipProps> = ({ data, positionRef }) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!data || !containerRef.current) {
      return;
    }

    let rafId: number;
    let lastX = NaN;
    let lastY = NaN;
    let lastWidth = NaN;
    let lastHeight = NaN;
    let lastViewportWidth = NaN;
    let lastViewportHeight = NaN;
    const update = () => {
      const container = containerRef.current;
      if (container && positionRef.current) {
        const { x, y } = positionRef.current;
        const width = container.offsetWidth;
        const height = container.offsetHeight;
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;
        if (
          x !== lastX ||
          y !== lastY ||
          width !== lastWidth ||
          height !== lastHeight ||
          viewportWidth !== lastViewportWidth ||
          viewportHeight !== lastViewportHeight
        ) {
          const left = clampToViewport(x, width, viewportWidth);
          const top = clampToViewport(y, height, viewportHeight);
          container.style.transform = `translate3d(${left}px, ${top}px, 0)`;
          lastX = x;
          lastY = y;
          lastWidth = width;
          lastHeight = height;
          lastViewportWidth = viewportWidth;
          lastViewportHeight = viewportHeight;
        }
      }
      rafId = requestAnimationFrame(update);
    };
    rafId = requestAnimationFrame(update);
    return () => cancelAnimationFrame(rafId);
  }, [data, positionRef]);

  if (!data) {
    return null;
  }

  return createPortal(
    <div
      ref={containerRef}
      role="tooltip"
      className="fixed w-max max-w-52 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: 0, top: 0, zIndex: Z_INDEX.FLOATING_TOOLTIP }}
    >
      <div
        className="game-panel text-white/90 text-[11px] leading-tight px-3 py-2"
        style={{ "--panel-cut": "14px" } as React.CSSProperties}
      >
        <TileInfoContent data={data} />
      </div>
    </div>,
    document.body,
  );
};

export default TileTooltip;
