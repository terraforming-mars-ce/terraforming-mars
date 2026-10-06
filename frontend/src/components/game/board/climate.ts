export const TEMPERATURE_MIN = -30;
export const TEMPERATURE_MAX = 8;
export const OXYGEN_MAX = 14;
export const OCEANS_MAX = 9;

export interface ClimateParameters {
  temperature: number;
  oxygen: number;
  oceans: number;
  maxOceans?: number;
}

export const CLIMATE_KEYS = [
  "frost",
  "plantFrost",
  "ice",
  "iceAge",
  "moss",
  "grass",
  "bush",
  "tree",
  "treeScale",
  "greening",
  "meadow",
  "flower",
  "birds",
  "clarity",
  "algae",
  "waves",
  "foam",
  "sky",
] as const;

export type ClimateKey = (typeof CLIMATE_KEYS)[number];
export type ClimateState = Record<ClimateKey, number>;

export const BARREN_PARAMETERS: ClimateParameters = {
  temperature: TEMPERATURE_MIN,
  oxygen: 0,
  oceans: 0,
};

export const TERRAFORMED_PARAMETERS: ClimateParameters = {
  temperature: TEMPERATURE_MAX,
  oxygen: OXYGEN_MAX,
  oceans: OCEANS_MAX,
};

function smoothstep(edge0: number, edge1: number, x: number): number {
  const t = Math.min(1, Math.max(0, (x - edge0) / (edge1 - edge0)));
  return t * t * (3 - 2 * t);
}

// Plants follow the global requirements of the plant cards: Lichen -24 °C, Grass -16 °C,
// Bushes -10 °C, Trees -4 °C, Moss 3 oceans. Ice melts from the shore between -4 and 0 °C.
export function climateFromParameters(parameters: ClimateParameters): ClimateState {
  const t = parameters.temperature;
  const o = parameters.oxygen;
  const w = (parameters.oceans * OCEANS_MAX) / Math.max(1, parameters.maxOceans ?? OCEANS_MAX);
  const warmth = smoothstep(TEMPERATURE_MIN, TEMPERATURE_MAX, t);
  const air = Math.min(1, Math.max(0, o / OXYGEN_MAX));
  const ice = 1 - smoothstep(-4, 0, t);
  return {
    frost: 1 - smoothstep(TEMPERATURE_MIN, 0, t),
    // Plants shed their frost as the greenery ground thaws, well before the frozen soil does.
    plantFrost: 1 - smoothstep(-32, -20, t),
    ice,
    iceAge: smoothstep(TEMPERATURE_MIN, -10, t),
    moss: smoothstep(1, 3, w),
    grass: smoothstep(-17, -12, t),
    bush: smoothstep(-11, -6, t),
    // Broadleaf trees arrive at the -4 °C requirement; temperature decides how many, not how big.
    tree: smoothstep(-4, 2, t),
    treeScale: 0.6 + 0.4 * air,
    greening: smoothstep(-28, 6, t) * (0.75 + 0.25 * air),
    meadow: smoothstep(-6, 2, t) * smoothstep(1, 3, w),
    flower: smoothstep(9, 11, o) * smoothstep(-1, 1, t),
    birds: smoothstep(5.5, 10, o),
    clarity: air,
    algae: smoothstep(3, 6, o) * (1 - ice),
    waves: smoothstep(2, OCEANS_MAX, w),
    foam: smoothstep(6, 8, w),
    sky: warmth,
  };
}

// Frost and ice ease faster so crossing 0 °C reads as a single melt sweep.
const EASE_SECONDS: Partial<Record<ClimateKey, number>> = { frost: 0.6, ice: 0.6 };
const DEFAULT_EASE_SECONDS = 1;

export function easeClimate(current: ClimateState, target: ClimateState, dt: number): boolean {
  let moving = false;
  for (const key of CLIMATE_KEYS) {
    const delta = target[key] - current[key];
    if (Math.abs(delta) < 1e-4) {
      current[key] = target[key];
      continue;
    }
    current[key] += delta * (1 - Math.exp(-dt / (EASE_SECONDS[key] ?? DEFAULT_EASE_SECONDS)));
    moving = true;
  }
  return moving;
}

export type ClimateMilestone = "temperature" | "oceans" | "terraformed";

function reached(before: number, after: number, max: number) {
  return before < max && after >= max;
}

function terraformed(p: ClimateParameters) {
  return (
    p.temperature >= TEMPERATURE_MAX &&
    p.oxygen >= OXYGEN_MAX &&
    p.oceans >= (p.maxOceans ?? OCEANS_MAX)
  );
}

export function climateMilestones(
  previous: ClimateParameters,
  next: ClimateParameters,
): ClimateMilestone[] {
  const maxOceans = next.maxOceans ?? OCEANS_MAX;
  const milestones: ClimateMilestone[] = [];
  if (reached(previous.temperature, next.temperature, TEMPERATURE_MAX)) {
    milestones.push("temperature");
  }
  if (reached(previous.oceans, next.oceans, maxOceans)) {
    milestones.push("oceans");
  }
  if (!terraformed(previous) && terraformed(next)) {
    milestones.push("terraformed");
  }
  return milestones;
}

export interface PlantForm {
  width: number;
  height: number;
}

const FULL_SIZE: PlantForm = { width: 1, height: 1 };

// Plants keep a fixed size once they appear; only oxygen thickens trees and bushes a little.
export function plantForm(kind: string, climate: ClimateState): PlantForm {
  if (kind === "tree" || kind === "bush") {
    return { width: climate.treeScale, height: climate.treeScale };
  }
  if (kind === "pine") {
    const size = 0.85 + 0.15 * ((climate.treeScale - 0.6) / 0.4);
    return { width: size, height: size };
  }
  return FULL_SIZE;
}
