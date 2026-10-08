import { useEffect, useLayoutEffect, useRef, type RefObject } from "react";

const TILE_SELECTOR = "[data-hand-card-id]";

interface Anchor {
  candidates: string[];
  offset: number;
}

function tilesOf(row: HTMLElement): HTMLElement[] {
  return Array.from(row.querySelectorAll<HTMLElement>(TILE_SELECTOR));
}

function captureAnchor(row: HTMLElement): Anchor | null {
  const tiles = tilesOf(row);
  const index = tiles.findIndex((tile) => tile.offsetLeft + tile.offsetWidth > row.scrollLeft);
  if (index === -1) {
    return null;
  }
  const ids = tiles.map((tile) => tile.dataset.handCardId ?? "");
  const candidates = ids.slice(index);
  for (let before = index - 1; before >= 0; before -= 1) {
    candidates.push(ids[before]);
  }
  return {
    candidates,
    offset: tiles[index].offsetLeft - row.scrollLeft,
  };
}

function restoreAnchor(row: HTMLElement, anchor: Anchor) {
  const tiles = new Map(tilesOf(row).map((tile) => [tile.dataset.handCardId ?? "", tile]));
  for (const id of anchor.candidates) {
    const tile = tiles.get(id);
    if (tile) {
      row.scrollLeft = tile.offsetLeft - anchor.offset;
      return;
    }
  }
}

/** Keeps the first visible card (or its nearest surviving neighbour) in place when the row changes. */
export function useScrollAnchor(rowRef: RefObject<HTMLElement | null>, orderKey: string) {
  const anchorRef = useRef<Anchor | null>(null);

  useLayoutEffect(() => {
    const row = rowRef.current;
    if (!row) {
      return;
    }
    if (anchorRef.current) {
      restoreAnchor(row, anchorRef.current);
    }
    anchorRef.current = captureAnchor(row);
  }, [rowRef, orderKey]);

  useEffect(() => {
    const row = rowRef.current;
    if (!row) {
      return;
    }
    const capture = () => {
      anchorRef.current = captureAnchor(row);
    };
    row.addEventListener("scroll", capture, { passive: true });
    return () => row.removeEventListener("scroll", capture);
  }, [rowRef]);
}
