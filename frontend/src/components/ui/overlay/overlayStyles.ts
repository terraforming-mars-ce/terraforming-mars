/**
 * Shared style constants for card selection overlays
 * Extracted to eliminate duplication across ProductionCardSelection, StartingCardSelection,
 * CardDrawSelection, and PendingCardSelection overlays
 */

export const OVERLAY_CONTAINER_CLASS =
  "relative z-[1] w-[90%] max-w-[1400px] max-h-[90vh] flex flex-col bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] overflow-hidden backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_60px_rgba(30,60,150,0.5)] max-[768px]:w-full max-[768px]:h-screen max-[768px]:max-h-screen max-[768px]:rounded-none";

export const OVERLAY_BACKDROP_BLUR_CLASS = "absolute inset-0 backdrop-blur-sm";
export const OVERLAY_BACKDROP_TINT_CLASS =
  "absolute inset-0 bg-black/60 animate-[fadeIn_0.3s_ease]";

export const OVERLAY_HEADER_CLASS =
  "py-6 px-8 bg-black/40 border-b border-space-blue-600 max-[768px]:p-5";

export const OVERLAY_TITLE_CLASS =
  "m-0 font-orbitron text-[28px] font-bold text-white text-shadow-glow tracking-wider max-[768px]:text-2xl";

export const OVERLAY_DESCRIPTION_CLASS = "mt-2 mb-0 text-base text-white/80 max-[768px]:text-sm";

export const OVERLAY_CARDS_CONTAINER_CLASS =
  "flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5";

export const OVERLAY_CARDS_INNER_CLASS = "flex gap-6 mx-auto py-5 max-[768px]:gap-4";

export const OVERLAY_FOOTER_CLASS =
  "py-6 px-8 bg-black/40 border-t border-space-blue-600 flex justify-between items-center max-[768px]:p-5 max-[768px]:flex-col max-[768px]:gap-5";

export const OVERLAY_FOOTER_LEFT_CLASS =
  "flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between";

export const OVERLAY_FOOTER_RIGHT_CLASS =
  "flex items-center gap-6 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3";

export const RESOURCE_LABEL_CLASS = "text-sm text-white/60 uppercase tracking-[0.5px]";

export const RESOURCE_DISPLAY_CLASS = "flex items-center gap-3";
