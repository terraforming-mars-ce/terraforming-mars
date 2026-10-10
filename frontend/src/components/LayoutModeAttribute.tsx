import { useEffect, useLayoutEffect } from "react";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

function isEditable(target: EventTarget | null): boolean {
  return (
    target instanceof HTMLElement &&
    (target.isContentEditable || target.closest("input, textarea, select") !== null)
  );
}

function blockTouchContextMenu(event: MouseEvent) {
  const pointerType = (event as PointerEvent).pointerType;
  if (!pointerType || pointerType === "mouse" || isEditable(event.target)) {
    return;
  }
  event.preventDefault();
}

export default function LayoutModeAttribute() {
  const { mode, isCompact } = useLayoutMode();

  useLayoutEffect(() => {
    document.documentElement.dataset.layout = mode;
  }, [mode]);

  useEffect(() => {
    if (!isCompact) {
      return;
    }
    document.addEventListener("contextmenu", blockTouchContextMenu);
    return () => document.removeEventListener("contextmenu", blockTouchContextMenu);
  }, [isCompact]);

  return null;
}
