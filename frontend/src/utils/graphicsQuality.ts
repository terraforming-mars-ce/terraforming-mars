import { QUICK_MODE } from "@/utils/quickMode.ts";

export type GraphicsTier = "high" | "low";

export interface GraphicsSettings {
  tier: GraphicsTier;
  planetTexturePixels: (name: string) => number;
  maxDpr: number;
  antialias: boolean;
  skybox: "exr" | "procedural";
  warmupAllPlanets: boolean;
  foliageDensity: number;
  birds: boolean;
  particleScale: number;
  sphereSegments: readonly [number, number];
}

const GFX_PARAM = "gfx";
const MIN_HIGH_TIER_TEXTURE_SIZE = 8192;

function isPhone(): boolean {
  if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
    return false;
  }
  return (
    window.matchMedia("(pointer: coarse)").matches &&
    Math.min(window.screen.width, window.screen.height) < 600
  );
}

function maxTextureSize(): number {
  try {
    const canvas = document.createElement("canvas");
    const gl = canvas.getContext("webgl2") ?? canvas.getContext("webgl");
    if (!gl) {
      return 0;
    }
    const size = gl.getParameter(gl.MAX_TEXTURE_SIZE) as number;
    gl.getExtension("WEBGL_lose_context")?.loseContext();
    return size;
  } catch {
    return 0;
  }
}

/**
 * Detection order: a `?gfx=high|low` URL param wins; otherwise quick mode means low;
 * otherwise a phone (coarse pointer and a screen shorter side under 600px) means low;
 * otherwise a GPU whose MAX_TEXTURE_SIZE is below 8192 means low; everything else is high.
 */
function detectTier(): GraphicsTier {
  const param = new URLSearchParams(window.location.search).get(GFX_PARAM);
  if (param === "high" || param === "low") {
    return param;
  }
  if (QUICK_MODE) {
    return "low";
  }
  if (isPhone()) {
    return "low";
  }
  if (maxTextureSize() < MIN_HIGH_TIER_TEXTURE_SIZE) {
    return "low";
  }
  return "high";
}

const HIGH: GraphicsSettings = {
  tier: "high",
  planetTexturePixels: () => Infinity,
  maxDpr: 1.5,
  antialias: true,
  skybox: "exr",
  warmupAllPlanets: true,
  foliageDensity: 1,
  birds: true,
  particleScale: 1,
  sphereSegments: [96, 64],
};

const LOW: GraphicsSettings = {
  tier: "low",
  planetTexturePixels: (name) => (name === "mars" ? 2048 : 1024),
  maxDpr: 1.25,
  antialias: false,
  skybox: "procedural",
  warmupAllPlanets: false,
  foliageDensity: 0.35,
  birds: false,
  particleScale: 0.3,
  sphereSegments: [48, 32],
};

export const GRAPHICS: Readonly<GraphicsSettings> = Object.freeze(
  detectTier() === "low" ? LOW : HIGH,
);

/** Scales a particle or sprite count by the tier, rounding up and keeping at least one. */
export function scaledParticleCount(count: number): number {
  return Math.max(1, Math.ceil(count * GRAPHICS.particleScale));
}
