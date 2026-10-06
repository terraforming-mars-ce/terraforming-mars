import { useEffect, useRef, useState } from "react";
import { SHOWCASE_TILE_HOLD_MS } from "@/constants/gameConstants.ts";

/**
 * True for a moment after a showcase player finishes their tile placements or other
 * choices, so the result stays in view before the showcase returns. Consecutive tiles from
 * the same card keep `pending` true throughout, so the hold only starts after the last one.
 */
export function useShowcaseTileHold(pending: boolean): boolean {
  const [holding, setHolding] = useState(false);
  const wasPending = useRef(pending);

  useEffect(() => {
    const resolved = wasPending.current && !pending;
    wasPending.current = pending;
    if (!resolved) {
      setHolding(false);
      return;
    }
    setHolding(true);
    const timer = window.setTimeout(() => setHolding(false), SHOWCASE_TILE_HOLD_MS);
    return () => window.clearTimeout(timer);
  }, [pending]);

  return holding;
}
