import { buildAssets } from "./pipeline.ts";
const result = await buildAssets();
console.log(`Assets: ${result.built} generated, ${result.total - result.built} unchanged.`);
