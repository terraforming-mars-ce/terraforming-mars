import React, { useEffect, useRef } from "react";
import { createPortal } from "react-dom";
import GameIcon from "./GameIcon.tsx";
import { Z_INDEX } from "@/constants/zIndex.ts";

export interface TileTooltipData {
  tileType: string;
  displayName?: string;
  placedName?: string;
  ownerName?: string;
  ownerColor?: string;
  reservedByName?: string;
  isOceanSpace: boolean;
  isVolcanic: boolean;
  bonuses: { [key: string]: number };
}

const TILE_TYPE_LABELS: Record<string, string> = {
  empty: "Land Space",
  ocean: "Ocean",
  city: "City",
  greenery: "Greenery",
  volcano: "Volcano",
  "nuclear-zone": "Nuclear Zone",
  mining: "Mining Area",
  restricted: "Reserved Area",
  special: "Special",
  "ecological-zone": "Ecological Zone",
  "natural-preserve": "Natural Preserve",
};

interface TileTooltipProps {
  data: TileTooltipData | null;
  positionRef: React.RefObject<{ x: number; y: number }>;
}

const TileTooltip: React.FC<TileTooltipProps> = ({ data, positionRef }) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!data || !containerRef.current) return;

    let rafId: number;
    let lastX = NaN;
    let lastY = NaN;
    const update = () => {
      if (containerRef.current && positionRef.current) {
        const { x, y } = positionRef.current;
        if (x !== lastX || y !== lastY) {
          containerRef.current.style.transform = `translate3d(${x + 12}px, ${y + 12}px, 0)`;
          lastX = x;
          lastY = y;
        }
      }
      rafId = requestAnimationFrame(update);
    };
    rafId = requestAnimationFrame(update);
    return () => cancelAnimationFrame(rafId);
  }, [data, positionRef]);

  if (!data) return null;

  const isCity = data.tileType === "city";
  const label = isCity
    ? "City"
    : data.displayName || TILE_TYPE_LABELS[data.tileType] || data.tileType;
  const isEmptySpace = data.tileType === "empty";
  const spaceLabel = data.isOceanSpace ? "Ocean Space" : "Land Space";
  const bonusEntries = Object.entries(data.bonuses);

  return createPortal(
    <div
      ref={containerRef}
      role="tooltip"
      className="fixed w-max max-w-52 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: 0, top: 0, zIndex: Z_INDEX.LOADING_OVERLAY }}
    >
      <div
        className="game-panel text-white/90 text-[11px] leading-tight px-3 py-2"
        style={{ "--panel-cut": "14px" } as React.CSSProperties}
      >
        <div className="font-orbitron font-bold text-xs text-white mb-1">
          {isEmptySpace && !data.displayName ? spaceLabel : label}
        </div>

        {isCity && data.placedName && (
          <div className="font-orbitron text-[10px] text-white/60 mb-1 break-words">
            {data.placedName}
          </div>
        )}

        {data.isVolcanic && isEmptySpace && (
          <div className="text-[10px] text-orange-400 mb-1">Volcanic</div>
        )}

        {data.ownerName && (
          <div className="flex items-center gap-1.5 mb-1">
            <div
              className="w-2 h-2 rounded-full shrink-0"
              style={{ backgroundColor: data.ownerColor || "#888" }}
            />
            <span className="text-white/70">{data.ownerName}</span>
          </div>
        )}

        {data.reservedByName && !data.ownerName && (
          <div className="text-white/50 text-[10px] mb-1">Reserved by {data.reservedByName}</div>
        )}

        {bonusEntries.length > 0 && (
          <div className="flex items-center gap-1.5 mt-1">
            <span className="text-white/50 text-[10px]">Bonus:</span>
            {bonusEntries.map(([type, amount]) => (
              <GameIcon key={type} iconType={type} amount={amount} size="small" />
            ))}
          </div>
        )}
      </div>
    </div>,
    document.body,
  );
};

export default TileTooltip;
