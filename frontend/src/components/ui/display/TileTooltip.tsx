import React from "react";
import { createPortal } from "react-dom";
import GameIcon from "./GameIcon.tsx";

export interface TileTooltipData {
  position: { x: number; y: number };
  tileType: string;
  displayName?: string;
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

const TileTooltip: React.FC<{ data: TileTooltipData | null }> = ({ data }) => {
  if (!data) return null;

  const label = data.displayName || TILE_TYPE_LABELS[data.tileType] || data.tileType;
  const isEmptySpace = data.tileType === "empty";
  const spaceLabel = data.isOceanSpace ? "Ocean Space" : "Land Space";
  const bonusEntries = Object.entries(data.bonuses);

  return createPortal(
    <div
      className="fixed w-max max-w-52 pt-1 pointer-events-none animate-[fadeIn_150ms_ease-in]"
      style={{ left: data.position.x + 12, top: data.position.y + 12, zIndex: 99999 }}
    >
      <div
        className="relative bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-2 shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
        style={{
          clipPath:
            "polygon(0 0, calc(100% - 14px) 0, 100% 14px, 100% 100%, 14px 100%, 0 calc(100% - 14px))",
        }}
      >
        <div className="font-orbitron font-bold text-xs text-white mb-1">
          {isEmptySpace ? spaceLabel : label}
        </div>

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

export default TileTooltip;
