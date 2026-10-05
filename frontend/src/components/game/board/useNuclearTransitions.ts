import { useState } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import type { TileDto } from "../../../types/generated/api-types";
import {
  reconcileNuclearTransitions,
  advanceNuclearTransitions,
  type NuclearTransitionState,
} from "./nuclearTransitions";

export function useNuclearTransitions(
  tiles: TileDto[] | undefined,
  gameId: string | undefined,
  newlyPlaced: Set<string>,
  replacedTiles: Map<string, TileDto>,
) {
  const clock = useThree((s) => s.clock);
  const [committed, setState] = useState<NuclearTransitionState>({
    gameId,
    tiles: undefined,
    serial: 0,
    active: new Map(),
  });
  let state = committed;
  if (state.tiles !== tiles || state.gameId !== gameId) {
    state = reconcileNuclearTransitions(
      state,
      tiles,
      gameId,
      newlyPlaced,
      replacedTiles,
      clock.elapsedTime,
    );
    setState(state);
  }
  useFrame(({ clock: frameClock }) => {
    if (!state.active.size) {
      return;
    }
    const active = advanceNuclearTransitions(state.active, frameClock.elapsedTime);
    if (active !== state.active) {
      setState((prior) => ({
        ...prior,
        active: advanceNuclearTransitions(prior.active, frameClock.elapsedTime),
      }));
    }
  }, -100);
  return state.active;
}
