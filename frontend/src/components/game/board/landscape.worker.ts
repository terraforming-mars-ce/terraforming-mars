import { buildLandscape, connectOceanGround } from "./landscapeNetwork";
import { createLandscapeGround, packLandscapeGround } from "./landscapeGeometry";
import type { LandscapeRequest, LandscapeResult } from "./useLandscape";

let previousKey = "";
let previousPlan = buildLandscape([]);

self.onmessage = ({ data }: MessageEvent<LandscapeRequest>) => {
  const planKey = JSON.stringify([data.tiles, data.spaces]);
  if (planKey !== previousKey) {
    previousPlan = buildLandscape(data.tiles, data.spaces);
    previousKey = planKey;
  }
  const plan = previousPlan;
  const geometry = createLandscapeGround(plan, connectOceanGround(plan, data.oceans));
  const ground = packLandscapeGround(geometry);
  const transfers: Transferable[] = Object.values(ground.attributes).map((a) => a.array.buffer);
  if (ground.index) {
    transfers.push(ground.index.buffer);
  }
  const result: LandscapeResult = { id: data.id, planKey, plan, ground };
  self.postMessage(result, { transfer: transfers });
  geometry.dispose();
};
