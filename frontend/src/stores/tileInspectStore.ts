import { create } from "zustand";
import type { TileTooltipData } from "@/components/ui/display/TileInfoContent.tsx";

export interface ScreenPoint {
  x: number;
  y: number;
}

interface TileInspectState {
  inspectedHex: string | null;
  info: TileTooltipData | null;
  position: ScreenPoint | null;
  inspect: (hex: string, info: TileTooltipData, position: ScreenPoint) => void;
  clear: () => void;
}

export const useTileInspectStore = create<TileInspectState>((set) => ({
  inspectedHex: null,
  info: null,
  position: null,
  inspect: (hex, info, position) => set({ inspectedHex: hex, info, position }),
  clear: () => set({ inspectedHex: null, info: null, position: null }),
}));

export function toggleInspectedHex(
  key: string | null,
  describe: (key: string) => TileTooltipData | null,
  position: ScreenPoint,
) {
  const { inspectedHex, inspect, clear } = useTileInspectStore.getState();
  if (!key || inspectedHex === key) {
    clear();
    return;
  }
  const info = describe(key);
  if (info) {
    inspect(key, info, position);
  } else {
    clear();
  }
}
