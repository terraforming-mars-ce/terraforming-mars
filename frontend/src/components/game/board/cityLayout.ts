import type * as THREE from "three";
import type { CityStyleRequestDto } from "../../../types/generated/api-types";
import { nearestOnRoad, type RoadPath, type Point2 } from "./landscapeGeometry";
export interface CityPlant {
  x: number;
  y: number;
  z: number;
  scale: number;
  kind: "tree" | "bush";
}

export interface CityStyle {
  plan: "clustered" | "radial" | "grid" | "courtyard";
  density: "sparse" | "balanced" | "dense";
  heights: "low" | "mixed" | "tall";
  ground: "level" | "recessed";
  exposure: "surface" | "mostly-buried";
  perimeter: "open" | "low-wall" | "high-wall";
  cover: "none" | "dome" | "flat-glass";
  landmark: "none" | "hall" | "observatory";
  connections: "paths" | "skybridges" | "enclosed";
  landscaping: "sparse" | "parks" | "lush";
  lighting: "normal" | "bright";
  perimeterRoad: "none" | "ring";
  entranceMin: number;
  entranceMax: number;
  details: ("antennas" | "dishes" | "vents" | "shafts" | "pipes")[];
}

export interface CityBuilding {
  x: number;
  y: number;
  z: number;
  width: number;
  depth: number;
  height: number;
  stepped: boolean;
  facade: number;
  tint: string;
  landmark: boolean;
  round: boolean;
  habitat: boolean;
  rotation: number;
  entranceRoad?: { x: number; y: number; halfWidth: number };
}

export interface CityLayout {
  name: string;
  seed: number;
  roads: RoadPath[];
  entrances: number[];
  style: CityStyle;
  buildings: CityBuilding[];
  plants: CityPlant[];
  parks: { x: number; y: number; radius: number; z: number }[];
}

export interface CityPlot {
  coordinate: { q: number; r: number; s: number };
  worldPosition: THREE.Vector3;
  normal: THREE.Vector3;
  layout: CityLayout;
}

export function randomSequence(seed: number) {
  return () => {
    seed |= 0;
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t ^= t + Math.imul(t ^ (t >>> 7), 61 | t);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const BASE_STYLE: CityStyle = {
  plan: "radial",
  density: "balanced",
  heights: "mixed",
  ground: "level",
  exposure: "surface",
  perimeter: "open",
  cover: "none",
  landmark: "none",
  connections: "paths",
  landscaping: "parks",
  lighting: "normal",
  perimeterRoad: "none",
  entranceMin: 2,
  entranceMax: 3,
  details: [],
};

const CITY_RECIPES: { name: string; seed: number; style: Partial<CityStyle> }[] = [
  {
    name: "Research Outpost",
    seed: 271,
    style: {
      plan: "clustered",
      density: "sparse",
      heights: "low",
      landmark: "observatory",
      connections: "enclosed",
      perimeter: "high-wall",
      landscaping: "parks",
      details: ["antennas", "dishes"],
    },
  },
  {
    name: "Capital",
    seed: 839,
    style: {
      landmark: "hall",
      plan: "radial",
      landscaping: "lush",
      entranceMin: 3,
      entranceMax: 3,
    },
  },
  {
    name: "Noctis City",
    seed: 1423,
    style: {
      plan: "clustered",
      connections: "skybridges",
      landscaping: "lush",
      details: ["antennas"],
    },
  },
  {
    name: "Cupola City",
    seed: 2069,
    style: {
      cover: "dome",
      heights: "low",
      perimeter: "high-wall",
      connections: "enclosed",
      landscaping: "sparse",
    },
  },
  {
    name: "Urbanized Area",
    seed: 3181,
    style: {
      plan: "grid",
      density: "dense",
      heights: "tall",
      lighting: "bright",
      perimeterRoad: "ring",
      connections: "skybridges",
      landscaping: "sparse",
    },
  },
  {
    name: "Underground City",
    seed: 4271,
    style: {
      exposure: "mostly-buried",
      entranceMin: 2,
      entranceMax: 2,
      ground: "recessed",
      cover: "flat-glass",
      density: "sparse",
      heights: "low",
      landscaping: "sparse",
      details: ["vents", "shafts", "pipes"],
    },
  },
  {
    name: "Open City",
    seed: 5393,
    style: {
      plan: "courtyard",
      entranceMin: 3,
      entranceMax: 4,
      heights: "low",
      landscaping: "lush",
    },
  },
];

export function insideBuilding(building: CityBuilding, x: number, y: number, padding: number) {
  const dx = x - building.x;
  const dy = y - building.y;
  const localX = dx * Math.cos(building.rotation) + dy * Math.sin(building.rotation);
  const localY = -dx * Math.sin(building.rotation) + dy * Math.cos(building.rotation);
  return (
    Math.abs(localX) < building.width / 2 + padding &&
    Math.abs(localY) < building.depth / 2 + padding
  );
}

export function onEntrancePath(building: CityBuilding, x: number, y: number, clearance: number) {
  if (!building.entranceRoad) {
    return false;
  }
  const target = building.entranceRoad;
  const dx = target.x - building.x;
  const dy = target.y - building.y;
  const lengthSquared = dx * dx + dy * dy;
  if (lengthSquared < 0.000001) {
    return false;
  }
  const t = Math.max(
    0,
    Math.min(1, ((x - building.x) * dx + (y - building.y) * dy) / lengthSquared),
  );
  return Math.hypot(x - building.x - dx * t, y - building.y - dy * t) < clearance;
}

function neighborhoodGrid(seed: number, rotation: number) {
  const rng = randomSequence(seed ^ 0x3489);
  const ys = [-0.14, -0.103, -0.069, -0.033, 0.004, 0.043, 0.077, 0.111, 0.14];
  const blocks: { x: number; y: number; width: number; depth: number }[] = [];
  const edgeRows = new Map<
    string,
    { fixed: number; horizontal: boolean; spans: [number, number][] }
  >();
  const edge = (a: Point2, b: Point2) => {
    const horizontal = a.y === b.y,
      fixed = horizontal ? a.y : a.x,
      key = `${horizontal}:${fixed}`;
    const row = edgeRows.get(key) ?? { fixed, horizontal, spans: [] };
    row.spans.push(horizontal ? [a.x, b.x] : [a.y, b.y]);
    edgeRows.set(key, row);
  };
  for (let row = 0; row < ys.length - 1; row++) {
    const bottom = ys[row],
      top = ys[row + 1];
    let left = -0.15,
      right = 0.15;
    for (let side = 0; side < 6; side++) {
      const angle = (side * Math.PI) / 3 - rotation;
      const nx = Math.cos(angle),
        ny = Math.sin(angle);
      for (const y of [bottom, top]) {
        if (Math.abs(nx) > 1e-6) {
          const x = (0.135 - ny * y) / nx;
          if (nx > 0) {
            right = Math.min(right, x);
          } else {
            left = Math.max(left, x);
          }
        }
      }
    }
    left += rng() * 0.005;
    right -= rng() * 0.005;
    if (right - left < 0.024) {
      continue;
    }
    const cells = Math.max(1, Math.round((right - left) / 0.034)),
      xs = [left];
    for (let col = 1; col < cells; col++) {
      xs.push(left + ((right - left) * col) / cells + (rng() - 0.5) * 0.01);
    }
    xs.push(right);
    for (let col = 0; col < cells; col++) {
      const x1 = xs[col];
      let x2 = xs[col + 1];
      if (col < cells - 1 && rng() < 0.15) {
        x2 = xs[++col + 1];
      }
      blocks.push({
        x: (x1 + x2) / 2,
        y: (bottom + top) / 2,
        width: x2 - x1 - 0.008,
        depth: top - bottom - 0.008,
      });
      edge({ x: x1, y: bottom }, { x: x2, y: bottom });
      edge({ x: x1, y: top }, { x: x2, y: top });
      edge({ x: x1, y: bottom }, { x: x1, y: top });
      edge({ x: x2, y: bottom }, { x: x2, y: top });
    }
  }
  const roads: RoadPath[] = [];
  for (const { fixed, horizontal, spans } of edgeRows.values()) {
    spans.sort((a, b) => a[0] - b[0]);
    const merged: [number, number][] = [];
    for (const span of spans) {
      const last = merged.at(-1);
      if (last && span[0] <= last[1] + 1e-8) {
        last[1] = Math.max(last[1], span[1]);
      } else {
        merged.push([...span]);
      }
    }
    for (const [from, to] of merged) {
      roads.push({
        kind: "road",
        width: 0.005,
        points: horizontal
          ? [
              { x: from, y: fixed },
              { x: to, y: fixed },
            ]
          : [
              { x: fixed, y: from },
              { x: fixed, y: to },
            ],
      });
    }
  }
  return { blocks, roads };
}

function createLayout(recipe: (typeof CITY_RECIPES)[number], rotation: number): CityLayout {
  const style: CityStyle = { ...BASE_STYLE, ...recipe.style };
  const rng = randomSequence(recipe.seed);
  const buildings: CityBuilding[] = [];
  const parks: CityLayout["parks"] = [];
  const plants: CityPlant[] = [];
  const addBuilding = (x: number, y: number, width = 0.016, depth = 0.017) => {
    let height = 0.033 + rng() * 0.041;
    if (style.heights === "low") {
      height = 0.025 + rng() * 0.016;
    } else if (style.heights === "tall") {
      height = 0.045 + rng() * 0.062;
    }
    if (style.exposure === "mostly-buried") {
      height = 0.009 + rng() * 0.01;
    }
    buildings.push({
      x,
      y,
      z: 0.007,
      width: width * (0.92 + rng() * 0.12),
      depth: depth * (0.92 + rng() * 0.12),
      height: Math.max(0.011, Math.round(height / 0.011) * 0.011),
      stepped: rng() > 0.35,
      facade: Math.floor(rng() * 3),
      tint: ["#ffffff", "#8595a1", "#515961", "#ac9785", "#687c79", "#d7d3c9"][
        buildings.length % 6
      ],
      landmark: false,
      round: rng() < 0.22,
      habitat: false,
      rotation: 0,
    });
  };

  if (style.exposure === "mostly-buried") {
    for (let i = 0; i < 5; i++) {
      const angle = (i * Math.PI) / 4;
      addBuilding(Math.cos(angle) * 0.09, Math.sin(angle) * 0.09, 0.026, 0.023);
    }
  } else if (style.plan === "grid" && style.perimeterRoad === "none") {
    for (const block of neighborhoodGrid(recipe.seed, rotation).blocks) {
      addBuilding(block.x, block.y, block.width, block.depth);
      const building = buildings.at(-1)!;
      building.width = block.width;
      building.depth = block.depth;
    }
  } else if (style.plan === "grid") {
    for (let row = 0; row < 6; row++) {
      for (let column = 0; column < 6; column++) {
        const x = (column - 2.5) * 0.033;
        const y = (row - 2.5) * 0.033;
        if (Math.hypot(x, y) < 0.115) {
          addBuilding(x, y, 0.013 + rng() * 0.007, 0.013 + rng() * 0.008);
        }
      }
    }
  } else {
    const rings = [
      { radius: 0.104, count: style.density === "sparse" ? 16 : 24 },
      { radius: 0.055, count: style.density === "sparse" ? 10 : 14 },
      { radius: 0.023, count: style.density === "sparse" ? 0 : 6 },
    ];
    for (const ring of rings) {
      for (let i = 0; i < ring.count; i++) {
        const angle = (i / ring.count) * Math.PI * 2 + (rng() - 0.5) * 0.12;
        const x = Math.cos(angle) * ring.radius;
        const y = Math.sin(angle) * ring.radius;
        if (style.landmark !== "none" && Math.abs(x) < 0.035 && y > -0.025 && y < 0.06) {
          continue;
        }
        if (style.cover === "dome" && ring.radius > 0.09) {
          addBuilding(x * 0.95, y * 0.95, 0.013, 0.013);
          buildings[buildings.length - 1].height = 0.022;
        } else {
          const size = ring.radius < 0.03 ? 0.018 : 0.022;
          addBuilding(x, y, size + rng() * 0.002, size + rng() * 0.002);
        }
      }
    }
  }

  if (style.landmark !== "none") {
    const hall = style.landmark === "hall";
    buildings.push({
      x: 0,
      y: 0.018,
      z: 0.007,
      width: hall ? 0.054 : 0.039,
      depth: hall ? 0.055 : 0.039,
      height: hall ? 0.077 : 0.03,
      stepped: hall,
      facade: 0,
      tint: "#ffffff",
      landmark: true,
      round: !hall,
      habitat: false,
      rotation: 0,
    });
  }
  return {
    name: recipe.name,
    seed: recipe.seed,
    roads: [],
    entrances: [],
    style,
    buildings,
    plants,
    parks,
  };
}

const layoutCache = new Map<string, CityLayout>();

export function generateCityLayout(
  seed: number,
  request: CityStyleRequestDto = {},
  name = "City",
): CityLayout {
  const key = JSON.stringify([seed, request, name]);
  const cached = layoutCache.get(key);
  if (cached) {
    return cached;
  }
  const choices = randomSequence(seed ^ 0x196c31);
  const style: CityStyle = {
    ...BASE_STYLE,
    plan: "grid",
    ...request,
    details: (request.details ?? ["pipes"]) as CityStyle["details"],
  };
  if (style.cover === "dome") {
    style.plan = "radial";
    style.entranceMin = style.entranceMax = 2;
  }
  const min = Math.max(1, Math.min(4, style.entranceMin));
  const max = Math.max(min, Math.min(4, style.entranceMax));
  const count = min + Math.floor(choices() * (max - min + 1));
  const rotation = choices() * Math.PI * 2;
  const layout = createLayout({ name, seed: seed ^ 0x5742, style }, rotation);
  layout.name = name;
  layout.seed = seed;
  layout.entrances = Array.from({ length: count }, (_, i) => rotation + (i * Math.PI * 2) / count);
  const roads: RoadPath[] = [];
  const line = (a: Point2, b: Point2, width = 0.007) =>
    roads.push({ points: [a, b], width, kind: "road" });
  const circle = (radius: number, offset = { x: 0, y: 0 }) =>
    roads.push({
      kind: "road",
      width: 0.007,
      points: Array.from({ length: 97 }, (_, i) => ({
        x: Math.cos((i * Math.PI) / 48) * radius + offset.x,
        y: Math.sin((i * Math.PI) / 48) * radius + offset.y,
      })),
    });
  if (style.exposure === "mostly-buried") {
    circle(0.077);
  } else if (style.plan === "grid" && style.perimeterRoad === "none") {
    roads.push(...neighborhoodGrid(seed ^ 0x5742, rotation).roads);
  } else if (style.plan === "grid") {
    if (style.perimeterRoad === "ring") {
      circle(0.108);
    }
    const extent = style.perimeterRoad === "ring" ? 0.108 : 0.098;
    for (const offset of [-0.066, -0.033, 0, 0.033, 0.066]) {
      const length = Math.sqrt(extent ** 2 - offset ** 2);
      line({ x: offset, y: -length }, { x: offset, y: length }, 0.0065);
      line({ x: -length, y: offset }, { x: length, y: offset }, 0.0065);
    }
  } else if (style.plan === "courtyard") {
    roads.push({
      kind: "road",
      width: 0.007,
      points: [
        { x: -0.056, y: -0.056 },
        { x: 0.056, y: -0.056 },
        { x: 0.056, y: 0.056 },
        { x: -0.056, y: 0.056 },
        { x: -0.056, y: -0.056 },
      ],
    });
  } else {
    circle(0.079, style.plan === "clustered" ? { x: 0.007, y: -0.004 } : undefined);
  }
  if (style.perimeterRoad === "ring" && style.plan !== "grid") {
    circle(0.108);
  }
  for (let i = 0; i < count; i++) {
    const angle = (i * Math.PI * 2) / count;
    const gateRadius = style.plan === "grid" && style.perimeterRoad === "none" ? 0.151 : 0.124;
    const gate = { x: Math.cos(angle) * gateRadius, y: Math.sin(angle) * gateRadius };
    const nearest = nearestOnRoad(gate, roads);
    line({ x: nearest.x, y: nearest.y }, gate);
    roads[roads.length - 1].exit = i;
  }
  const rotate = (point: Point2) => ({
    x: point.x * Math.cos(rotation) - point.y * Math.sin(rotation),
    y: point.x * Math.sin(rotation) + point.y * Math.cos(rotation),
  });
  layout.roads = roads.map((r) => ({ ...r, points: r.points.map(rotate) }));
  if (style.exposure === "mostly-buried") {
    for (const b of layout.buildings) {
      b.x *= 1.15;
      b.y *= 1.15;
    }
  }
  layout.buildings = layout.buildings.filter((b) => {
    if (b.landmark) {
      return true;
    }
    const nearest = nearestOnRoad(b, roads);
    const half = Math.abs(nearest.x - b.x) > Math.abs(nearest.y - b.y) ? b.width / 2 : b.depth / 2;
    return nearest.distance > half + nearest.halfWidth + 0.001;
  });
  for (const b of layout.buildings) {
    const p = rotate(b);
    b.x = p.x;
    b.y = p.y;
    b.rotation += rotation;
    {
      const target = nearestOnRoad(
        b,
        layout.roads.filter((r) => r.exit === undefined),
      );
      if (style.plan === "grid") {
        const dx = target.x - b.x,
          dy = target.y - b.y;
        const localX = dx * Math.cos(rotation) + dy * Math.sin(rotation),
          localY = -dx * Math.sin(rotation) + dy * Math.cos(rotation);
        if (Math.abs(localX) > Math.abs(localY)) {
          [b.width, b.depth] = [b.depth, b.width];
        }
      }
      b.entranceRoad = target;
      b.rotation = Math.atan2(target.y - b.y, target.x - b.x) + Math.PI / 2;
    }
  }
  if (style.exposure !== "mostly-buried") {
    const architecture = randomSequence(seed ^ 0x78219);
    const cluster = { x: Math.cos(rotation + 0.7) * 0.048, y: Math.sin(rotation + 0.7) * 0.048 };
    const candidates = layout.buildings
      .filter((b) => !b.landmark)
      .sort(
        (a, b) =>
          Math.hypot(a.x - cluster.x, a.y - cluster.y) -
          Math.hypot(b.x - cluster.x, b.y - cluster.y),
      );
    const towers = style.heights === "mixed" ? 4 : 0;
    for (const [i, b] of candidates.entries()) {
      if (style.heights === "mixed") {
        b.height =
          i < towers
            ? 0.05 + architecture() * 0.024
            : 0.012 + Math.floor(architecture() * 3) * 0.009;
      }
      if (
        i >= towers &&
        i % 4 === 0 &&
        style.heights !== "tall" &&
        Math.max(b.width, b.depth) / Math.min(b.width, b.depth) < 1.4
      ) {
        b.habitat = true;
        b.round = true;
        b.stepped = false;
        b.height = Math.min(b.width, b.depth) * 0.65;
      }
    }
  }
  // Independent planting stream: changing architecture details cannot resample vegetation.
  const planting = randomSequence(seed ^ 0x87cd);
  layout.plants = [];
  layout.parks = [];
  const target = style.landscaping === "lush" ? 140 : style.landscaping === "parks" ? 110 : 8;
  for (let i = 0; i < 4000 && layout.plants.length < target; i++) {
    const x = (planting() - 0.5) * 0.254,
      y = (planting() - 0.5) * 0.254;
    const size = 0.25 + planting() * 0.17;
    if (
      Math.hypot(x, y) > (style.perimeter === "open" ? 0.127 : 0.104) ||
      nearestOnRoad({ x, y }, layout.roads).distance < 0.013 ||
      (style.ground === "recessed" && Math.hypot(x, y) < 0.085)
    ) {
      continue;
    }
    if (
      layout.buildings.some(
        (b) => insideBuilding(b, x, y, 0.006) || onEntrancePath(b, x, y, 0.006),
      ) ||
      layout.plants.some((p) => Math.hypot(p.x - x, p.y - y) < 0.006)
    ) {
      continue;
    }
    layout.plants.push({ x, y, z: 0.0085, scale: size, kind: planting() > 0.5 ? "tree" : "bush" });
  }
  layoutCache.set(key, layout);
  if (layoutCache.size > 256) {
    layoutCache.delete(layoutCache.keys().next().value!);
  }
  return layout;
}

export const createCityShowcase = () =>
  CITY_RECIPES.map((recipe, i) =>
    generateCityLayout(recipe.seed, recipe.style, `${i + 1} ${recipe.name}`),
  );

export function visibleCityRoads(layout: CityLayout, activeExits: readonly number[]): RoadPath[] {
  return layout.roads.filter((road) => road.exit === undefined || activeExits.includes(road.exit));
}
