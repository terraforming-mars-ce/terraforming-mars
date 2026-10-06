import { useRef, useMemo, useLayoutEffect } from "react";
import type { TileDto } from "../types/generated/api-types";

export function changedOccupants(
  previous: Map<string, string> | null,
  current: Map<string, string>,
) {
  const placed = new Set<string>();
  if (previous) {
    for (const [key, occupant] of current) {
      if (occupant && previous.get(key) !== occupant) {
        placed.add(key);
      }
    }
  }
  return placed;
}

export function usePreviousTiles(tiles: TileDto[] | undefined, gameId: string | undefined) {
  const previous = useRef<{ gameId: string | undefined; tiles: Map<string, TileDto> } | null>(null);
  const current = useMemo(
    () =>
      new Map(
        (tiles ?? []).map((tile) => [
          `${tile.coordinates.q},${tile.coordinates.r},${tile.coordinates.s}`,
          tile,
        ]),
      ),
    [tiles],
  );
  const changes = useMemo(() => {
    const prior =
      previous.current && previous.current.gameId === gameId ? previous.current.tiles : null;
    const newlyPlaced = changedOccupants(
      prior ? new Map([...prior].map(([key, tile]) => [key, tile.occupiedBy?.type ?? ""])) : null,
      new Map([...current].map(([key, tile]) => [key, tile.occupiedBy?.type ?? ""])),
    );
    const replacedTiles = new Map<string, TileDto>();
    for (const key of newlyPlaced) {
      const tile = prior?.get(key);
      if (tile) {
        replacedTiles.set(key, tile);
      }
    }
    return { newlyPlaced, replacedTiles };
  }, [current, gameId]);
  useLayoutEffect(() => {
    if (tiles) {
      previous.current = { gameId, tiles: current };
    }
  }, [current, tiles, gameId]);
  return changes;
}
