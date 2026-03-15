import { create } from "zustand";

export type TransitionPhase =
  | "idle"
  | "lobby"
  | "loading"
  | "fadeOutLobby"
  | "marsRevealed"
  | "animateUI"
  | "complete";

interface TransitionState {
  transitionPhase: TransitionPhase;
  isSkyboxReady: boolean;
  isGpuReady: boolean;
  overlayVisible: boolean;
  marsRevealedReady: boolean;
  lobbyMounted: boolean;

  setTransitionPhase: (phase: TransitionPhase) => void;
  setSkyboxReady: (ready: boolean) => void;
  setGpuReady: (ready: boolean) => void;
  setOverlayVisible: (visible: boolean) => void;
  setMarsRevealedReady: (ready: boolean) => void;
  setLobbyMounted: (mounted: boolean) => void;
}

export const useTransitionStore = create<TransitionState>((set) => ({
  transitionPhase: "idle",
  isSkyboxReady: false,
  isGpuReady: false,
  overlayVisible: true,
  marsRevealedReady: false,
  lobbyMounted: false,

  setTransitionPhase: (transitionPhase) => set({ transitionPhase }),
  setSkyboxReady: (isSkyboxReady) => set({ isSkyboxReady }),
  setGpuReady: (isGpuReady) => set({ isGpuReady }),
  setOverlayVisible: (overlayVisible) => set({ overlayVisible }),
  setMarsRevealedReady: (marsRevealedReady) => set({ marsRevealedReady }),
  setLobbyMounted: (lobbyMounted) => set({ lobbyMounted }),
}));
