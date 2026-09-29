import { HexGrid2D, type HexCoordinate } from "../../../utils/hex-grid-2d";
import type { CityLayout } from "./cityLayout";
import {
  boardCenter,
  getBiomeValue,
  hashSeed,
  nearestOnRoad,
  type Point2,
  type RoadPath,
} from "./landscapeGeometry";
import { insideBuilding, onEntrancePath } from "./cityLayout";
import { routeLandscape, type LandscapeSpace } from "./landscapeRoutes";

export interface LandscapeTile {
  coordinate: HexCoordinate;
  kind: "city" | "greenery" | "special-greenery";
  seed: number;
  layout?: CityLayout;
}
export interface GroundCircle extends Point2 {
  radius: number;
  forest: boolean;
}
export interface LandscapeConnection {
  key: string;
  a: string;
  b: string;
  routes: RoadPath[];
}
export interface LandscapePlant extends Point2 {
  id: string;
  tileKey: string;
  kind: "tree" | "bush";
  scale: number;
  seed: number;
}
export interface LandscapePlan {
  tiles: LandscapeTile[];
  connections: LandscapeConnection[];
  roads: RoadPath[];
  reserved: RoadPath[];
  circles: GroundCircle[];
  greenLinks: RoadPath[];
  plants: LandscapePlant[];
  reserveCells: Map<string, RoadPath[]>;
  activeExits: Map<string, number[]>;
  buildingCells: Map<string, { tile: LandscapeTile; center: Point2 }[]>;
}
const keyOf = HexGrid2D.coordinateToKey;
const shift = (road: RoadPath, center: Point2): RoadPath => ({
  ...road,
  points: road.points.map((p) => ({ x: p.x + center.x, y: p.y + center.y })),
});

export function buildLandscape(
  tiles: LandscapeTile[],
  spaces: LandscapeSpace[] = [],
): LandscapePlan {
  const sorted = [...tiles].sort((a, b) => keyOf(a.coordinate).localeCompare(keyOf(b.coordinate)));
  const byKey = new Map(sorted.map((t) => [keyOf(t.coordinate), t]));
  const result: LandscapePlan = {
    tiles: sorted,
    connections: [],
    roads: [],
    reserved: [],
    circles: [],
    greenLinks: [],
    plants: [],
    reserveCells: new Map(),
    activeExits: new Map(),
    buildingCells: new Map(),
  };
  const green = (tile: LandscapeTile) =>
    tile.kind !== "city" ||
    (tile.layout?.style.landscaping !== "sparse" && tile.layout?.style.cover !== "dome");
  for (const tile of sorted) {
    const center = boardCenter(tile.coordinate);
    if (green(tile)) {
      result.circles.push({
        ...center,
        radius: tile.kind === "city" ? 0.14 : 0.184,
        forest: tile.kind !== "city",
      });
    }
    for (const neighbor of HexGrid2D.getNeighbors(tile.coordinate)) {
      const other = byKey.get(keyOf(neighbor));
      if (!other || keyOf(tile.coordinate) >= keyOf(neighbor)) {
        continue;
      }
      if (green(tile) && green(other)) {
        result.greenLinks.push({
          points: [center, boardCenter(neighbor)],
          width: tile.kind === "city" || other.kind === "city" ? 0.2 : 0.3,
          kind: "path",
        });
      }
    }
    if (tile.layout) {
      for (
        let x = Math.floor((center.x - 0.19) / 0.2);
        x <= Math.floor((center.x + 0.19) / 0.2);
        x++
      ) {
        for (
          let y = Math.floor((center.y - 0.19) / 0.2);
          y <= Math.floor((center.y + 0.19) / 0.2);
          y++
        ) {
          const key = `${x}:${y}`,
            bucket = result.buildingCells.get(key) ?? [];
          bucket.push({ tile, center });
          result.buildingCells.set(key, bucket);
        }
      }
      result.reserved.push(...tile.layout.roads.map((r) => shift(r, center)));
    }
  }
  result.connections = routeLandscape(sorted, spaces);
  result.roads = result.connections.flatMap((connection) => connection.routes);
  result.reserved.push(...result.roads);
  for (const tile of sorted) {
    if (!tile.layout) {
      continue;
    }
    const center = boardCenter(tile.coordinate);
    const active = tile.layout.roads
      .filter((r) => r.exit !== undefined)
      .filter((r) => {
        const end = r.points.at(-1)!;
        return result.roads.some((road) =>
          [road.points[0], road.points.at(-1)!].some(
            (p) => Math.hypot(p.x - center.x - end.x, p.y - center.y - end.y) < 1e-6,
          ),
        );
      })
      .map((r) => r.exit!);
    result.activeExits.set(keyOf(tile.coordinate), active);
  }
  for (const road of result.reserved) {
    for (let i = 1; i < road.points.length; i++) {
      const a = road.points[i - 1],
        b = road.points[i],
        margin = road.width / 2 + 0.07;
      const segment = { ...road, points: [a, b] };
      for (
        let x = Math.floor((Math.min(a.x, b.x) - margin) / 0.04);
        x <= Math.floor((Math.max(a.x, b.x) + margin) / 0.04);
        x++
      ) {
        for (
          let y = Math.floor((Math.min(a.y, b.y) - margin) / 0.04);
          y <= Math.floor((Math.max(a.y, b.y) + margin) / 0.04);
          y++
        ) {
          const key = `${x}:${y}`,
            bucket = result.reserveCells.get(key) ?? [];
          bucket.push(segment);
          result.reserveCells.set(key, bucket);
        }
      }
    }
  }
  result.plants = scatterLandscape(result);
  return result;
}
export function isReserved(plan: LandscapePlan, x: number, y: number, clearance: number): boolean {
  const road = nearestOnRoad(
    { x, y },
    plan.reserveCells.get(`${Math.floor(x / 0.04)}:${Math.floor(y / 0.04)}`) ?? [],
  );
  if (road.distance < road.halfWidth + clearance) {
    return true;
  }
  for (const { tile, center } of plan.buildingCells.get(
    `${Math.floor(x / 0.2)}:${Math.floor(y / 0.2)}`,
  ) ?? []) {
    if (!tile.layout) {
      continue;
    }
    const lx = x - center.x,
      ly = y - center.y;
    if (Math.hypot(lx, ly) > 0.17) {
      continue;
    }
    const style = tile.layout.style;
    if (style.cover === "dome" && Math.hypot(lx, ly) > 0.103 && Math.hypot(lx, ly) < 0.13) {
      return true;
    }
    if (style.perimeter !== "open" && Math.abs(Math.hypot(lx, ly) - 0.121) < clearance + 0.005) {
      return true;
    }
    if (style.ground === "recessed" && Math.hypot(lx, ly) < 0.09) {
      return true;
    }
    if (
      tile.layout.buildings.some(
        (b) => insideBuilding(b, lx, ly, clearance) || onEntrancePath(b, lx, ly, clearance),
      )
    ) {
      return true;
    }
  }
  return false;
}
const plantCache = new Map<string, LandscapePlant[]>();
function scatterLandscape(plan: LandscapePlan): LandscapePlant[] {
  const roadBounds = plan.reserved.map((road) => {
    const margin = road.width / 2 + 0.009 + 0.184;
    return {
      road,
      minX: Math.min(...road.points.map((p) => p.x)) - margin,
      maxX: Math.max(...road.points.map((p) => p.x)) + margin,
      minY: Math.min(...road.points.map((p) => p.y)) - margin,
      maxY: Math.max(...road.points.map((p) => p.y)) + margin,
    };
  });
  const cities = plan.tiles
    .filter((tile) => tile.layout)
    .map((tile) => ({
      tile,
      center: boardCenter(tile.coordinate),
    }));
  const plants: LandscapePlant[] = [];
  for (const tile of plan.tiles) {
    const center = boardCenter(tile.coordinate);
    const roads = roadBounds
      .filter(
        (r) => center.x >= r.minX && center.x <= r.maxX && center.y >= r.minY && center.y <= r.maxY,
      )
      .map((r) => r.road);
    const neighbors = cities
      .filter((city) => Math.hypot(center.x - city.center.x, center.y - city.center.y) <= 0.355)
      .map(({ tile }) => [tile.coordinate, tile.layout!.style, tile.layout!.buildings]);
    const key = JSON.stringify([
      tile.coordinate,
      tile.seed,
      tile.kind,
      tile.layout?.style.landscaping,
      roads,
      neighbors,
    ]);
    let local = plantCache.get(key);
    if (local) {
      plantCache.delete(key);
    } else {
      local = scatterTile(plan, tile);
    }
    plantCache.set(key, local);
    if (plantCache.size > 128) {
      plantCache.delete(plantCache.keys().next().value!);
    }
    plants.push(...local);
  }
  return plants;
}

function scatterTile(plan: LandscapePlan, tile: LandscapeTile): LandscapePlant[] {
  const plants: LandscapePlant[] = [];
  // World-cell candidates have fixed positions and ownership, independent of occupied neighbors.
  const center = boardCenter(tile.coordinate),
    tileKey = keyOf(tile.coordinate);
  if (tile.kind === "city" && tile.layout?.style.landscaping === "sparse") {
    return plants;
  }
  const neighbors = HexGrid2D.getNeighbors(tile.coordinate).map((coordinate) => ({
    center: boardCenter(coordinate),
    key: keyOf(coordinate),
  }));
  const biome = getBiomeValue(tile.coordinate.q, tile.coordinate.r);
  for (const [kind, spacing] of [
    ["tree", 0.029],
    ["bush", 0.01],
  ] as const) {
    for (
      let ix = Math.floor((center.x - 0.19) / spacing);
      ix <= Math.ceil((center.x + 0.19) / spacing);
      ix++
    ) {
      for (
        let iy = Math.floor((center.y - 0.19) / spacing);
        iy <= Math.ceil((center.y + 0.19) / spacing);
        iy++
      ) {
        const cellSeed = hashSeed(`${kind}:${ix}:${iy}`);
        const x = (ix + 0.3 + (cellSeed % 401) / 1000) * spacing,
          y = (iy + 0.3 + ((cellSeed >>> 10) % 401) / 1000) * spacing;
        const distance = Math.hypot(x - center.x, y - center.y);
        if (distance > 0.184) {
          continue;
        }
        const edgeDistance = Math.max(
          Math.abs(x - center.x),
          Math.abs(0.5 * (x - center.x) + 0.866025404 * (y - center.y)),
          Math.abs(-0.5 * (x - center.x) + 0.866025404 * (y - center.y)),
        );
        if (kind === "tree" && tile.kind !== "city" && edgeDistance > 0.125) {
          continue;
        }
        const belongs = neighbors.every(({ center: p, key }) => {
          const d = Math.hypot(x - p.x, y - p.y);
          return distance < d || (distance === d && tileKey < key);
        });
        if (!belongs) {
          continue;
        }
        const seed = hashSeed(`${tile.seed}:${cellSeed}`);
        const city = tile.kind === "city";
        if (city && distance > 0.131) {
          continue;
        }
        let threshold = 0.35;
        if (kind === "tree") {
          threshold = 0.2 + biome * 0.3;
          if (city) {
            threshold = 0.55;
          }
        }
        if (
          (seed % 1000) / 1000 > threshold ||
          isReserved(plan, x, y, kind === "tree" ? 0.009 : 0.003)
        ) {
          continue;
        }
        plants.push({
          id: `${kind}:${ix}:${iy}`,
          tileKey,
          x,
          y,
          seed,
          kind,
          scale: city ? 0.28 + (seed % 100) / 600 : 0.9 + (seed % 100) / 500,
        });
      }
    }
  }
  return plants;
}

export function connectOceanGround(plan: LandscapePlan, oceans: { coordinate: HexCoordinate }[]) {
  const tiles = new Map(
    plan.tiles.map((tile) => [keyOf(tile.coordinate), { coordinate: tile.coordinate }]),
  );
  const water = new Map(
    oceans.map((tile) => [keyOf(tile.coordinate), boardCenter(tile.coordinate)]),
  );
  const waterLinks: { a: Point2; b: Point2 }[] = [];
  for (const ocean of oceans) {
    const key = keyOf(ocean.coordinate);
    tiles.set(key, ocean);
    for (const neighbor of HexGrid2D.getNeighbors(ocean.coordinate)) {
      const otherKey = keyOf(neighbor);
      tiles.set(otherKey, { coordinate: neighbor });
      const other = water.get(otherKey);
      if (other && key < otherKey) {
        waterLinks.push({ a: water.get(key)!, b: other });
      }
    }
  }
  return {
    tiles: [...tiles.values()],
    links: plan.greenLinks,
    water: [...water.values()],
    waterLinks,
  };
}
