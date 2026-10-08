import { useRef, type ReactNode } from "react";
import { useElementSize } from "@/hooks/useElementSize.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { CardFitHeightContext } from "../cards/CardFitHeightContext.ts";
import { OVERLAY_CARDS_CONTAINER_CLASS, OVERLAY_CARDS_INNER_CLASS } from "./overlayStyles.ts";

const COMPACT_ROW_PADDING = 16;

/** Horizontal row of selectable cards; in compact mode every card is fitted to the row height. */
export default function CardSelectionRow({ children }: { children: ReactNode }) {
  const { isCompact } = useLayoutMode();
  const containerRef = useRef<HTMLDivElement>(null);
  const { height } = useElementSize(containerRef);
  let fitHeight: number | null = null;
  if (isCompact && height > 0) {
    fitHeight = Math.max(0, height - COMPACT_ROW_PADDING);
  }
  return (
    <div ref={containerRef} className={OVERLAY_CARDS_CONTAINER_CLASS}>
      <div className={OVERLAY_CARDS_INNER_CLASS}>
        <CardFitHeightContext.Provider value={fitHeight}>{children}</CardFitHeightContext.Provider>
      </div>
    </div>
  );
}
