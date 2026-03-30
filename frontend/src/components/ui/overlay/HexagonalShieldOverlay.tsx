import React, { useState, useEffect } from "react";
import { PlayerCardDto } from "@/types/generated/api-types.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface HexagonalShieldOverlayProps {
  card: PlayerCardDto | null;
  reason: string | null;
  isVisible: boolean;
}

const HexagonalShieldOverlay: React.FC<HexagonalShieldOverlayProps> = ({
  card,
  reason,
  isVisible,
}) => {
  const [shouldRender, setShouldRender] = useState(false);
  const [isAnimatingOut, setIsAnimatingOut] = useState(false);
  const [isAnimatingIn, setIsAnimatingIn] = useState(false);
  const [lastValidCard, setLastValidCard] = useState<PlayerCardDto | null>(null);
  const [lastValidReason, setLastValidReason] = useState<string | null>(null);

  useEffect(() => {
    if (isVisible && card && reason) {
      setShouldRender(true);
      setLastValidCard(card);
      setLastValidReason(reason);

      requestAnimationFrame(() => {
        setIsAnimatingOut(false);
        setIsAnimatingIn(true);
      });
      return undefined;
    } else if (shouldRender) {
      setIsAnimatingOut(true);
      setIsAnimatingIn(false);

      const timer = setTimeout(() => {
        setShouldRender(false);
        setIsAnimatingOut(false);
        setLastValidCard(null);
        setLastValidReason(null);
      }, 300);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [isVisible, card, reason, shouldRender]);

  if (!shouldRender) {
    return null;
  }

  const overlayClass = isAnimatingOut ? "hidden" : isAnimatingIn ? "visible" : "hidden";

  const displayCard = lastValidCard;
  const displayReason = lastValidReason;

  if (!displayCard || !displayReason) {
    return null;
  }

  // Hex grid configuration
  const maxCols = 7;
  const totalRows = 5;
  const hexSize = 192;
  const hexWidth = hexSize * Math.sqrt(3);
  const hexHeight = hexSize * 2;
  const vertSpacing = hexHeight * 0.75;
  const hexRadius = hexSize;
  const padding = hexRadius;
  const gridWidth = (maxCols - 1) * hexWidth + hexWidth;
  const gridHeight = (totalRows - 1) * vertSpacing + hexHeight;
  const svgWidth = gridWidth + padding * 2;
  const svgHeight = gridHeight + padding * 2;

  const generateHexagonPattern = () => {
    const hexagons = [];
    const centerX = svgWidth / 2;
    const centerY = svgHeight / 2;
    const hexagonsPerRow = [4, 5, 6, 5, 4];
    const actualRows = hexagonsPerRow.length;

    for (let row = 0; row < actualRows; row++) {
      const colsInThisRow = hexagonsPerRow[row];

      for (let col = 0; col < colsInThisRow; col++) {
        let offsetX = hexWidth / 2;
        if (row === 1 || row === 3) {
          offsetX += hexWidth / 2;
        }
        if (row === 0 || row === 2 || row === 4) {
          offsetX += hexWidth / 2;
        }

        const maxRowWidth = Math.max(...hexagonsPerRow);
        const rowStartX = ((maxRowWidth - colsInThisRow) * hexWidth) / 2;
        const x = padding + rowStartX + col * hexWidth + offsetX;
        const y = padding + row * vertSpacing + hexSize;

        const distanceFromCenter = Math.sqrt(Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2));
        const maxRadius = Math.min(svgWidth, svgHeight) / 2;
        const normalizedDistance = distanceFromCenter / maxRadius;
        const opacity = Math.max(0, 1 - normalizedDistance * 0.8);

        hexagons.push(
          <polygon
            key={`hex-${row}-${col}`}
            points={generateHexagonPoints(x, y, hexSize)}
            fill="rgba(40, 20, 5, 0.85)"
            stroke="rgba(255, 152, 0, 0.8)"
            strokeWidth="2"
            opacity={opacity}
            className="transition-opacity duration-500 [filter:drop-shadow(0_0_8px_rgba(255,152,0,0.9))]"
          />,
        );
      }
    }
    return hexagons;
  };

  const generateHexagonPoints = (centerX: number, centerY: number, size: number) => {
    const points = [];
    for (let i = 0; i < 6; i++) {
      const angle = (i * 60 - 90) * (Math.PI / 180);
      const x = centerX + size * Math.cos(angle);
      const y = centerY + size * Math.sin(angle);
      points.push(`${x},${y}`);
    }
    return points.join(" ");
  };

  return (
    <div
      className={`hexagonal-shield-overlay fixed top-[calc(50%+10px)] left-1/2 -translate-x-1/2 -translate-y-1/2 w-[108vw] h-[96vh] pointer-events-none flex justify-center items-center transition-opacity duration-300 ${overlayClass === "visible" ? "opacity-100" : "opacity-0"}`}
      style={{ zIndex: Z_INDEX.STANDARD_MODAL }}
    >
      <div className="relative w-full h-full flex justify-center items-center max-w-[800px] max-h-[600px]">
        <svg
          className="absolute top-0 left-0 w-full h-full opacity-90 [filter:blur(0.5px)_drop-shadow(0_0_12px_rgba(255,152,0,0.8))]"
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
        >
          {generateHexagonPattern()}
        </svg>

        <div className="absolute top-0 left-0 w-full h-full flex items-center justify-center z-[3] pointer-events-none">
          <div
            className={`text-center max-w-[80%] transition-all duration-300 ${overlayClass === "visible" ? "opacity-100 scale-100" : "opacity-0 scale-90"}`}
          >
            <div className="flex items-center gap-3 bg-black/70 py-3 px-5 rounded-xl border-2 border-[rgba(255,152,0,0.6)] backdrop-blur-[8px] shadow-[0_0_24px_rgba(255,152,0,0.6)]">
              <span className="text-white text-lg font-medium [text-shadow:2px_2px_4px_rgba(0,0,0,0.9)] leading-[1.3]">
                {displayReason}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HexagonalShieldOverlay;
