import { create } from "zustand";

interface DebugHexHighlightState {
  hexes: ReadonlySet<string>;
  setHexes: (hexes: readonly string[]) => void;
  clear: () => void;
}

const EMPTY: ReadonlySet<string> = new Set();

export const useDebugHexHighlightStore = create<DebugHexHighlightState>((set) => ({
  hexes: EMPTY,
  setHexes: (hexes) => set({ hexes: hexes.length > 0 ? new Set(hexes) : EMPTY }),
  clear: () => set({ hexes: EMPTY }),
}));
