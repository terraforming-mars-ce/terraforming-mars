import type { HexCoordinate } from "../../../utils/hex-grid-2d";
import type { CityLayout } from "./cityLayout";

export type LandscapeKind =
  | "city"
  | "greenery"
  | "special-greenery"
  | "ocean"
  | "volcano"
  | "excluded";
export interface LandscapeSource {
  coordinate: HexCoordinate;
  kind: LandscapeKind;
  seed: number;
  layout?: CityLayout;
}
export interface LandscapeInput {
  seed: number;
  sources: LandscapeSource[];
}
export interface LandscapePlant {
  id: string;
  tileKey: string;
  kind: "tree" | "bush" | "clover" | "flower" | "rock";
  x: number;
  y: number;
  height: number;
  scale: number;
  seed: number;
}
export interface LandscapePatch {
  key: string;
  signature: string;
  x: number;
  y: number;
  terrain: Uint16Array;
  materials: Uint8Array;
  plants: LandscapePlant[];
  water: boolean;
}
export interface LandscapeRequest extends LandscapeInput {
  id: number;
  known: Record<string, string>;
}
export interface LandscapeDelta {
  id: number;
  seed: number;
  sources: LandscapeSource[];
  patches: LandscapePatch[];
  removed: string[];
  preparationMs: number;
}
export interface LandscapeState {
  id: number;
  seed: number;
  sources: LandscapeSource[];
  patches: Map<string, LandscapePatch>;
  plants: LandscapePlant[];
}
export const EMPTY_LANDSCAPE: LandscapeState = {
  id: 0,
  seed: 0,
  sources: [],
  patches: new Map(),
  plants: [],
};
