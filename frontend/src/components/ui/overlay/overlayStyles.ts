/**
 * Shared style constants for card selection overlays
 * Extracted to eliminate duplication across ProductionCardSelection, StartingCardSelection,
 * CardDrawSelection, and PendingCardSelection overlays
 */

import { getZIndex } from "@/constants/zIndex.ts";

export const OVERLAY_CONTAINER_STYLE = { zIndex: getZIndex("LOCAL", 1) };

export const OVERLAY_ROOT_SAFE_AREA_CLASS =
  "compact:pt-[var(--safe-top)] compact:pr-[var(--safe-right)] compact:pb-[var(--safe-bottom)] compact:pl-[var(--safe-left)]";

export const OVERLAY_CONTAINER_CLASS =
  "relative w-[90%] max-w-[1400px] max-h-[90dvh] flex flex-col game-panel game-panel-clipped game-window overflow-hidden compact:w-full compact:max-w-none compact:h-full compact:max-h-full compact:rounded-none compact:[--panel-cut:0px]";

export const OVERLAY_BACKDROP_BLUR_CLASS = "absolute inset-0 backdrop-blur-sm";
export const OVERLAY_BACKDROP_TINT_CLASS =
  "absolute inset-0 bg-black/60 animate-[fadeIn_0.3s_ease]";

export const OVERLAY_HEADER_CLASS =
  "py-6 px-8 bg-black/40 border-b border-white/15 compact:shrink-0 compact:min-h-11 compact:py-1 compact:px-4 compact:flex compact:flex-col compact:justify-center";

export const OVERLAY_TITLE_CLASS =
  "m-0 font-orbitron text-[28px] font-bold text-white text-shadow-glow tracking-wider compact:text-base compact:leading-5 compact:tracking-wide compact:truncate";

export const OVERLAY_DESCRIPTION_CLASS =
  "mt-2 mb-0 text-base text-white/80 compact:mt-0 compact:text-xs compact:leading-4 compact:truncate";

export const OVERLAY_CARDS_CONTAINER_CLASS =
  "flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 compact:min-h-0 compact:justify-center-safe compact:px-3 compact:py-0 compact:snap-x compact:snap-proximity compact:overscroll-x-contain compact:[&::-webkit-scrollbar]:h-1";

export const OVERLAY_CARDS_INNER_CLASS =
  "flex gap-6 mx-auto py-5 compact:mx-0 compact:py-2 compact:[&>*]:snap-center";

export const OVERLAY_FOOTER_CLASS =
  "py-6 px-8 bg-black/40 border-t border-white/15 flex justify-between items-center compact:shrink-0 compact:h-14 compact:py-0 compact:px-4 compact:gap-3";

export const OVERLAY_FOOTER_LEFT_CLASS = "flex gap-8 items-center compact:gap-4 compact:min-w-0";

export const OVERLAY_FOOTER_RIGHT_CLASS = "flex items-center gap-6 compact:gap-3 compact:min-w-0";

export const OVERLAY_ACTION_BUTTON_CLASS =
  "whitespace-nowrap compact:min-h-11 compact:py-2 compact:px-5";

export const RESOURCE_LABEL_CLASS =
  "text-sm text-white/60 uppercase tracking-[0.5px] compact:text-[11px] compact:whitespace-nowrap";

export const RESOURCE_DISPLAY_CLASS = "flex items-center gap-3 compact:gap-1.5";

export const COLONY_TRACK_SCROLL_CLASS = "compact:overflow-x-auto compact:overscroll-x-contain";

export const COLONY_TRACK_INNER_CLASS = "compact:min-w-[432px]";
