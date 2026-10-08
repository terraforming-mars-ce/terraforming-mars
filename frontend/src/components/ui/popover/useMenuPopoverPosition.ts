import type { RefObject } from "react";
import type { PopoverPosition } from "../GamePopover/types";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

export const MENU_POPOVER_WIDTH = 360;
const PORTRAIT_TOP = 120;
const LANDSCAPE_TOP = 12;

export function useMenuPopoverPosition(anchorRef: RefObject<HTMLElement | null>): PopoverPosition {
  const { isSmallViewport, isPortrait } = useLayoutMode();
  if (!isSmallViewport) {
    return { type: "anchor", anchorRef, placement: "below" };
  }
  return {
    type: "fixed",
    top: isPortrait ? PORTRAIT_TOP : LANDSCAPE_TOP,
    left: Math.max(0, (window.innerWidth - MENU_POPOVER_WIDTH) / 2),
  };
}
