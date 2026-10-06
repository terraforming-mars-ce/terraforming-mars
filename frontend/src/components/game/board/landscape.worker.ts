import { LandscapeBuilder } from "./landscapeFields";
import type { LandscapeRequest, LandscapeDelta } from "./landscapeTypes";
const builder = new LandscapeBuilder();
let pending: LandscapeRequest | null = null;
let scheduled = false;
self.onmessage = ({ data }: MessageEvent<LandscapeRequest>) => {
  pending = data;
  if (scheduled) {
    return;
  }
  scheduled = true;
  setTimeout(() => {
    scheduled = false;
    const request = pending!;
    const start = performance.now();
    const result = builder.build(request);
    const patches = [...result.values()]
      .filter((p) => request.known[p.key] !== p.signature)
      .map((p) => ({
        ...p,
        terrain: p.terrain.slice(),
        materials: p.materials.slice(),
        detail: p.detail.slice(),
      }));
    const delta: LandscapeDelta = {
      id: request.id,
      seed: request.seed,
      relief: request.relief,
      sources: request.sources,
      patches,
      removed: Object.keys(request.known).filter((key) => !result.has(key)),
      preparationMs: performance.now() - start,
    };
    self.postMessage(delta, {
      transfer: patches.flatMap((p) => [p.terrain.buffer, p.materials.buffer, p.detail.buffer]),
    });
  }, 0);
};
