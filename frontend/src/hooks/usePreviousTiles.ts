import { useRef, useMemo } from "react";
import { TileDto } from "../types/generated/api-types";

/**
 * Hook to detect newly placed tiles by comparing current tiles with previous state.
 * Returns a Set of coordinate keys for tiles that were just occupied.
 */
export function usePreviousTiles(tiles: TileDto[] | undefined): Set<string> {
  const previousTilesRef = useRef<Map<string, TileDto>>(new Map());
  const isInitializedRef = useRef(false);

  const newlyPlacedTiles = useMemo(() => {
    const placed = new Set<string>();

    if (!tiles) {
      return placed;
    }

    const currentTilesMap = new Map<string, TileDto>();

    for (const tile of tiles) {
      const key = `${tile.coordinates.q},${tile.coordinates.r},${tile.coordinates.s}`;
      currentTilesMap.set(key, tile);

      // Skip detection on first render to avoid triggering for existing tiles
      if (isInitializedRef.current) {
        const previousTile = previousTilesRef.current.get(key);
        const isOccupiedNow = tile.occupiedBy != null;
        const wasOccupiedBefore = previousTile?.occupiedBy != null;

        if (
          isOccupiedNow &&
          (!wasOccupiedBefore || tile.occupiedBy!.type !== previousTile!.occupiedBy!.type)
        ) {
          placed.add(key);
        }
      }
    }

    previousTilesRef.current = currentTilesMap;
    isInitializedRef.current = true;

    return placed;
  }, [tiles]);

  return newlyPlacedTiles;
}
