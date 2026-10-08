import { create } from "zustand";
import type { BeforeInstallPromptEvent } from "@/utils/installApp.ts";

interface InstallAppState {
  deferredPrompt: BeforeInstallPromptEvent | null;
  installed: boolean;
  howToOpen: boolean;

  setDeferredPrompt: (deferredPrompt: BeforeInstallPromptEvent | null) => void;
  markInstalled: () => void;
  setHowToOpen: (howToOpen: boolean) => void;
}

export const useInstallAppStore = create<InstallAppState>((set) => ({
  deferredPrompt: null,
  installed: false,
  howToOpen: false,

  setDeferredPrompt: (deferredPrompt) => set({ deferredPrompt }),
  markInstalled: () => set({ installed: true, deferredPrompt: null, howToOpen: false }),
  setHowToOpen: (howToOpen) => set({ howToOpen }),
}));
