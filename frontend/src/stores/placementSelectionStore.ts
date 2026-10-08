import { create } from "zustand";
import { webSocketService } from "@/services/webSocketService.ts";

interface PlacementSelectionState {
  selectedHex: string | null;
  select: (hex: string) => void;
  clear: () => void;
}

export const usePlacementSelectionStore = create<PlacementSelectionState>((set) => ({
  selectedHex: null,
  select: (hex) => set({ selectedHex: hex }),
  clear: () => set({ selectedHex: null }),
}));

export function sendTileSelection(hexCoordinate: string) {
  const [q, r, s] = hexCoordinate.split(",").map(Number);
  usePlacementSelectionStore.getState().clear();
  try {
    webSocketService.selectTile({ q, r, s });
  } catch (error) {
    console.error("Failed to send tile selection:", error);
  }
}
