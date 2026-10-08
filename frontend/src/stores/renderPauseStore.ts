import { useEffect } from "react";
import { create } from "zustand";

interface RenderPauseState {
  reasons: string[];

  pause: (reason: string) => void;
  resume: (reason: string) => void;
}

export const useRenderPauseStore = create<RenderPauseState>((set) => ({
  reasons: [],

  pause: (reason) => set((state) => ({ reasons: [...state.reasons, reason] })),
  resume: (reason) =>
    set((state) => {
      const index = state.reasons.indexOf(reason);
      if (index === -1) {
        return state;
      }
      return { reasons: state.reasons.filter((_, i) => i !== index) };
    }),
}));

export const isRenderPaused = (state: RenderPauseState) => state.reasons.length > 0;

export function useRenderPause(reason: string, active: boolean) {
  useEffect(() => {
    if (!active) {
      return;
    }
    const { pause, resume } = useRenderPauseStore.getState();
    pause(reason);
    return () => resume(reason);
  }, [reason, active]);
}

export function usePauseWhileDocumentHidden(reason: string) {
  useEffect(() => {
    const { pause, resume } = useRenderPauseStore.getState();
    let paused = false;
    const update = () => {
      if (document.hidden && !paused) {
        paused = true;
        pause(reason);
      } else if (!document.hidden && paused) {
        paused = false;
        resume(reason);
      }
    };
    update();
    document.addEventListener("visibilitychange", update);
    return () => {
      document.removeEventListener("visibilitychange", update);
      if (paused) {
        resume(reason);
      }
    };
  }, [reason]);
}
