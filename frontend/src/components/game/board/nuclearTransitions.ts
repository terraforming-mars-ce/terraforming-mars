import type { TileDto } from "../../../types/generated/api-types";
import {
  nuclearElapsed,
  NUCLEAR_FLIGHT_SECONDS,
  NUCLEAR_DURATION_SECONDS,
} from "./nuclearGeometry";

export const NUCLEAR_COLLAPSE_SECONDS = 0.9;
export interface NuclearTransition {
  id: number;
  start: number;
  outgoing?: TileDto;
  impacted: boolean;
  collapsed: boolean;
}
export type NuclearTransitions = ReadonlyMap<string, NuclearTransition>;
export interface NuclearTransitionState {
  gameId?: string;
  tiles: TileDto[] | undefined;
  serial: number;
  active: Map<string, NuclearTransition>;
}
export function reconcileNuclearTransitions(
  state: NuclearTransitionState,
  tiles: TileDto[] | undefined,
  gameId: string | undefined,
  newlyPlaced: Set<string>,
  replacedTiles: Map<string, TileDto>,
  now: number,
): NuclearTransitionState {
  const active = new Map(state.gameId === gameId ? state.active : []);
  let serial = state.serial;
  const keys = new Set<string>();
  for (const tile of tiles ?? []) {
    const key = `${tile.coordinates.q},${tile.coordinates.r},${tile.coordinates.s}`;
    if (tile.location !== "mars" || tile.occupiedBy?.type !== "nuclear-zone-tile") {
      continue;
    }
    keys.add(key);
    if (newlyPlaced.has(key) && !active.has(key)) {
      active.set(key, {
        id: ++serial,
        start: now,
        outgoing: replacedTiles.get(key),
        impacted: false,
        collapsed: false,
      });
    }
  }
  for (const key of active.keys()) {
    if (!keys.has(key)) {
      active.delete(key);
    }
  }
  return { gameId, tiles, serial, active };
}
export function advanceNuclearTransitions(active: Map<string, NuclearTransition>, now: number) {
  let next = active;
  for (const [key, transition] of active) {
    const elapsed = nuclearElapsed(now - transition.start);
    const age = elapsed - NUCLEAR_FLIGHT_SECONDS;
    const impacted = age >= 0;
    const collapsed = age >= NUCLEAR_COLLAPSE_SECONDS;
    if (elapsed >= NUCLEAR_DURATION_SECONDS) {
      if (next === active) {
        next = new Map(active);
      }
      next.delete(key);
    } else if (impacted !== transition.impacted || collapsed !== transition.collapsed) {
      if (next === active) {
        next = new Map(active);
      }
      next.set(key, {
        ...transition,
        impacted,
        collapsed,
        outgoing: collapsed ? undefined : transition.outgoing,
      });
    }
  }
  return next;
}
export function collapseParameters(seed: number) {
  const random = (salt: number) => {
    const value = Math.sin(seed * 12.9898 + salt * 78.233) * 43758.5453;
    return value - Math.floor(value);
  };
  return {
    delay: random(1) * 0.12,
    tilt: ((2 + random(2) * 4) * Math.PI) / 180,
    direction: random(3) * Math.PI * 2,
  };
}
export function collapseProgress(age: number, delay: number) {
  const t = Math.max(0, Math.min(1, (age - delay) / 0.75));
  return t * t;
}
