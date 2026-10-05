import { createContext, useContext, useLayoutEffect, useMemo, useRef, type RefObject } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import {
  TileHighlightBoard,
  type HighlightRecord,
  type HighlightState,
  type HighlightTile,
} from "./tileHighlights";
import type { TileHighlightSystem } from "./tileHighlightMaterials";

export const TileHighlightSystemContext = createContext<TileHighlightSystem | null>(null);
export const TileHighlightBoardContext = createContext<TileHighlightBoard | null>(null);
export function useTileHighlightBoard(
  root: RefObject<THREE.Group | null>,
  radius: number,
  ySign: number,
  gameId: string | undefined,
  offsetQ = 0,
  offsetR = 0,
) {
  const system = useContext(TileHighlightSystemContext);
  const board = useMemo(
    () => new TileHighlightBoard(radius, ySign, { q: offsetQ, r: offsetR }),
    [radius, ySign, offsetQ, offsetR, gameId],
  );
  useLayoutEffect(() => {
    system?.boards.add(board);
    return () => {
      if (system) {
        system.release(board);
      } else {
        board.dispose();
      }
    };
  }, [system, board]);
  useFrame(() => {
    board.root = root.current;
  });
  return board;
}
export function useTileHighlight(
  tile: HighlightTile,
  hovered: boolean,
  available: boolean,
  vpIntensity: number,
  vpColor: readonly number[],
) {
  const board = useContext(TileHighlightBoardContext);
  const record = useRef<HighlightRecord | null>(null);
  const state = useRef<HighlightState>({ hovered, available, vpIntensity, vpColor });
  useLayoutEffect(() => {
    if (!board) {
      return;
    }
    const registration = board.register(tile, state.current);
    record.current = registration;
    return () => {
      board.remove(registration);
      record.current = null;
    };
  }, [board, tile.coordinate.q, tile.coordinate.r, tile.coordinate.s]);
  useLayoutEffect(() => {
    if (board && record.current) {
      board.updateTile(record.current, tile);
    }
  }, [board, tile]);
  useLayoutEffect(() => {
    Object.assign(state.current, { hovered, available, vpIntensity, vpColor });
  }, [hovered, available, vpIntensity, vpColor]);
  return record;
}
