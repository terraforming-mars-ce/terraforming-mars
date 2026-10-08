import { useCallback } from "react";
import { useInstallAppStore } from "@/stores/installAppStore.ts";

export interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
}

export type InstallPlatform = "prompt" | "ios" | "firefox-android";
type PromptOutcome = "accepted" | "dismissed";
export type InstallOutcome = PromptOutcome | "how-to";

const STANDALONE_QUERIES = [
  "(display-mode: standalone)",
  "(display-mode: fullscreen)",
  "(display-mode: minimal-ui)",
];

export function isStandalone(): boolean {
  if (STANDALONE_QUERIES.some((query) => window.matchMedia(query).matches)) {
    return true;
  }
  return (navigator as { standalone?: boolean }).standalone === true;
}

function isIos(): boolean {
  const ua = navigator.userAgent;
  if (/iPhone|iPad|iPod/.test(ua)) {
    return true;
  }
  return /Macintosh/.test(ua) && navigator.maxTouchPoints > 1;
}

function isFirefoxAndroid(): boolean {
  const ua = navigator.userAgent;
  return /Android/.test(ua) && /Firefox\//.test(ua);
}

function platformFor(
  deferredPrompt: BeforeInstallPromptEvent | null,
  installed: boolean,
): InstallPlatform | null {
  if (installed || isStandalone()) {
    return null;
  }
  if (deferredPrompt) {
    return "prompt";
  }
  if (isIos()) {
    return "ios";
  }
  if (isFirefoxAndroid()) {
    return "firefox-android";
  }
  return null;
}

export function installPlatform(): InstallPlatform | null {
  const { deferredPrompt, installed } = useInstallAppStore.getState();
  return platformFor(deferredPrompt, installed);
}

export async function promptInstall(): Promise<PromptOutcome | null> {
  const { deferredPrompt, setDeferredPrompt } = useInstallAppStore.getState();
  if (!deferredPrompt) {
    return null;
  }
  setDeferredPrompt(null);
  try {
    await deferredPrompt.prompt();
    const choice = await deferredPrompt.userChoice;
    return choice.outcome;
  } catch {
    return "dismissed";
  }
}

export function initInstallDetection() {
  window.addEventListener("beforeinstallprompt", (event) => {
    event.preventDefault();
    useInstallAppStore.getState().setDeferredPrompt(event as BeforeInstallPromptEvent);
  });
  window.addEventListener("appinstalled", () => {
    useInstallAppStore.getState().markInstalled();
  });
}

export function useInstallOffer() {
  const deferredPrompt = useInstallAppStore((s) => s.deferredPrompt);
  const installed = useInstallAppStore((s) => s.installed);
  const platform = platformFor(deferredPrompt, installed);

  const install = useCallback(async (): Promise<InstallOutcome> => {
    const outcome = await promptInstall();
    if (outcome) {
      return outcome;
    }
    useInstallAppStore.getState().setHowToOpen(true);
    return "how-to";
  }, []);

  return { platform, install };
}
