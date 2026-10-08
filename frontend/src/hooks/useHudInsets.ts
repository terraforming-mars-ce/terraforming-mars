import { useSyncExternalStore } from "react";

export interface HudInsets {
  top: number;
  bottom: number;
  left: number;
  right: number;
}

const ZERO_INSETS: HudInsets = { top: 0, bottom: 0, left: 0, right: 0 };

function readPx(style: CSSStyleDeclaration, name: string): number {
  const value = parseFloat(style.getPropertyValue(name));
  return Number.isFinite(value) ? value : 0;
}

export function readHudInsets(): HudInsets {
  if (typeof document === "undefined") {
    return ZERO_INSETS;
  }
  const style = getComputedStyle(document.documentElement);
  return {
    top: readPx(style, "--hud-top-h"),
    bottom: readPx(style, "--hud-dock-h"),
    left: readPx(style, "--hud-rail-left-w"),
    right: readPx(style, "--hud-rail-right-w"),
  };
}

let cachedInsets: HudInsets = ZERO_INSETS;

function getSnapshot(): HudInsets {
  const next = readHudInsets();
  if (
    next.top === cachedInsets.top &&
    next.bottom === cachedInsets.bottom &&
    next.left === cachedInsets.left &&
    next.right === cachedInsets.right
  ) {
    return cachedInsets;
  }
  cachedInsets = next;
  return next;
}

function subscribe(onChange: () => void) {
  const observer = new MutationObserver(onChange);
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ["data-layout"],
  });
  window.addEventListener("resize", onChange);
  return () => {
    observer.disconnect();
    window.removeEventListener("resize", onChange);
  };
}

export function useHudInsets(): HudInsets {
  return useSyncExternalStore(subscribe, getSnapshot);
}
