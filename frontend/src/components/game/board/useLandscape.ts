import { useEffect, useRef, useState } from "react";
import {
  EMPTY_LANDSCAPE,
  type LandscapeInput,
  type LandscapeState,
  type LandscapeRequest,
  type LandscapeDelta,
} from "./landscapeTypes";
export function useLandscape(input: LandscapeInput, signature: string) {
  const worker = useRef<Worker | null>(null);
  const generation = useRef(0);
  const accepted = useRef(EMPTY_LANDSCAPE);
  const [result, setResult] = useState<LandscapeState>(EMPTY_LANDSCAPE);
  useEffect(() => {
    const instance = new Worker(new URL("./landscape.worker.ts", import.meta.url), {
      type: "module",
    });
    worker.current = instance;
    accepted.current = EMPTY_LANDSCAPE;
    instance.onmessage = ({ data }: MessageEvent<LandscapeDelta>) => {
      if (data.id !== generation.current) {
        return;
      }
      const patches = new Map(accepted.current.patches);
      for (const key of data.removed) {
        patches.delete(key);
      }
      for (const patch of data.patches) {
        patches.set(patch.key, patch);
      }
      const state = {
        id: data.id,
        seed: data.seed,
        relief: data.relief,
        sources: data.sources,
        patches,
        plants: [...patches.values()].flatMap((p) => p.plants),
      };
      accepted.current = state;
      setResult(state);
    };
    instance.onerror = (event) => {
      console.error("Landscape worker failed", event.message);
    };
    return () => {
      instance.terminate();
      worker.current = null;
    };
  }, []);
  useEffect(() => {
    worker.current!.postMessage({
      ...input,
      id: ++generation.current,
      known: Object.fromEntries(
        [...accepted.current.patches].map(([key, p]) => [key, p.signature]),
      ),
    } satisfies LandscapeRequest);
  }, [signature]);
  return result;
}
