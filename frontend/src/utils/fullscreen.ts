import { getLayoutMode } from "@/hooks/useLayoutMode.ts";

type LockableOrientation = ScreenOrientation & {
  lock?: (orientation: "landscape") => Promise<void>;
};

export function isFullscreenSupported(): boolean {
  return (
    typeof document !== "undefined" &&
    document.fullscreenEnabled === true &&
    typeof document.documentElement.requestFullscreen === "function"
  );
}

export function isFullscreen(): boolean {
  return !!document.fullscreenElement;
}

async function lockLandscape(): Promise<void> {
  const orientation = screen.orientation as LockableOrientation | undefined;
  if (!orientation?.lock) {
    return;
  }
  try {
    await orientation.lock("landscape");
  } catch {
    // Unsupported or not allowed on this browser; fullscreen still works without the lock.
  }
}

export async function toggleFullscreen(): Promise<void> {
  if (!isFullscreenSupported()) {
    return;
  }
  try {
    if (isFullscreen()) {
      await document.exitFullscreen();
      return;
    }
    await document.documentElement.requestFullscreen({ navigationUI: "hide" });
    if (getLayoutMode().isCompact) {
      await lockLandscape();
    }
  } catch {
    // The browser refused the fullscreen change (e.g. no user gesture); nothing to recover.
  }
}
