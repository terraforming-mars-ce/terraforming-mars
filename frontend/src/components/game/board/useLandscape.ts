import { useEffect, useRef, useState } from "react";
import { buildLandscape, type LandscapePlan, type LandscapeTile } from "./landscapeNetwork";
import type { LandscapeSpace } from "./landscapeRoutes";
import type { LandscapeGroundData } from "./landscapeGeometry";
import type { HexCoordinate } from "../../../utils/hex-grid-2d";

export interface LandscapeRequest {
  id: number;
  tiles: LandscapeTile[];
  spaces: LandscapeSpace[];
  oceans: { coordinate: HexCoordinate }[];
}
export interface LandscapeResult {
  id: number;
  planKey: string;
  plan: LandscapePlan;
  ground: LandscapeGroundData;
}
const EMPTY_RESULT: LandscapeResult = {
  id: 0,
  planKey: "",
  plan: buildLandscape([]),
  ground: { attributes: {}, index: null },
};

export function useLandscape(input: Omit<LandscapeRequest, "id">, signature: string) {
  const worker = useRef<Worker | null>(null);
  const generation = useRef(0);
  const [result, setResult] = useState(EMPTY_RESULT);
  useEffect(() => {
    const instance = new Worker(new URL("./landscape.worker.ts", import.meta.url), {
      type: "module",
    });
    worker.current = instance;
    instance.onmessage = ({ data }: MessageEvent<LandscapeResult>) => {
      if (data.id === generation.current) {
        setResult((previous) =>
          previous.planKey === data.planKey ? { ...data, plan: previous.plan } : data,
        );
      }
    };
    return () => {
      instance.terminate();
      worker.current = null;
    };
  }, []);
  useEffect(() => {
    worker.current!.postMessage({ ...input, id: ++generation.current } satisfies LandscapeRequest);
  }, [signature]);
  return result;
}
