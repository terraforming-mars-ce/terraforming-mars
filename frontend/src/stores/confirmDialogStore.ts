import { create } from "zustand";

export type ConfirmDialogKind = "kick" | "convertToBot";

export interface ConfirmDialogRequest {
  kind: ConfirmDialogKind;
  playerId: string;
  playerName: string;
}

interface ConfirmDialogState {
  pending: ConfirmDialogRequest | null;

  request: (request: ConfirmDialogRequest) => void;
  dismiss: () => void;
}

export const useConfirmDialogStore = create<ConfirmDialogState>((set) => ({
  pending: null,

  request: (request) => set({ pending: request }),
  dismiss: () => set({ pending: null }),
}));
