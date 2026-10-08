import { useEffect, useState } from "react";

const HIGHLIGHT_MS = 600;
const EMPTY = new Set<string>();

/** Ids that joined the hand since the previous change, cleared after a short display time. */
export function useNewCardHighlight(cardIds: string[]): ReadonlySet<string> {
  const key = cardIds.join("|");
  const [state, setState] = useState(() => ({ key, ids: cardIds, fresh: EMPTY }));

  let current = state;
  if (state.key !== key) {
    const previous = new Set(state.ids);
    const added = cardIds.filter((id) => !previous.has(id));
    current = { key, ids: cardIds, fresh: added.length > 0 ? new Set(added) : EMPTY };
    setState(current);
  }

  const fresh = current.fresh;
  useEffect(() => {
    if (fresh.size === 0) {
      return;
    }
    const timer = setTimeout(() => {
      setState((prev) => (prev.fresh === fresh ? { ...prev, fresh: EMPTY } : prev));
    }, HIGHLIGHT_MS);
    return () => clearTimeout(timer);
  }, [fresh]);

  return fresh;
}
