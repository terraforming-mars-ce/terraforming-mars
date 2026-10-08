import type React from "react";

export function boardBottomOverlayStyle(maxWidthPx: number): React.CSSProperties {
  return {
    bottom: "calc(var(--hud-dock-h) + var(--safe-bottom) + 8px)",
    left: "calc((100vw + var(--hud-rail-left-w) - var(--hud-rail-right-w)) / 2)",
    maxWidth: `min(${maxWidthPx}px, calc(100vw - var(--hud-rail-left-w) - var(--hud-rail-right-w) - 16px))`,
  };
}

export function getTileIconType(tileType: string): string {
  switch (tileType) {
    case "city":
      return "city-tile";
    case "greenery":
      return "greenery-tile";
    case "ocean":
      return "ocean-tile";
    case "volcano":
      return "volcano-tile";
    default:
      return "tile-placement";
  }
}
