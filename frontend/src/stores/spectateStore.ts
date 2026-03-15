import { create } from "zustand";

interface SpectateState {
  spectatePlayerId: string | null;
  replaySpectatePlayerId: string | null;

  setSpectatePlayerId: (id: string | null) => void;
  toggleSpectatePlayer: (id: string, currentPlayerId?: string) => void;
  setReplaySpectatePlayerId: (id: string | null) => void;
  clearSpectating: () => void;
}

export const useSpectateStore = create<SpectateState>((set) => ({
  spectatePlayerId: null,
  replaySpectatePlayerId: null,

  setSpectatePlayerId: (spectatePlayerId) => set({ spectatePlayerId }),
  toggleSpectatePlayer: (id, currentPlayerId) => {
    if (id === currentPlayerId) {
      set({ spectatePlayerId: null });
      return;
    }
    set((state) => ({
      spectatePlayerId: state.spectatePlayerId === id ? null : id,
    }));
  },
  setReplaySpectatePlayerId: (replaySpectatePlayerId) => set({ replaySpectatePlayerId }),
  clearSpectating: () => set({ spectatePlayerId: null }),
}));
