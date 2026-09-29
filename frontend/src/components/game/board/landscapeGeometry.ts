import * as THREE from "three";
import { mergeGeometries } from "three/examples/jsm/utils/BufferGeometryUtils.js";
import { HexGrid2D, type HexCoordinate } from "../../../utils/hex-grid-2d";
import type { LandscapePlan, connectOceanGround } from "./landscapeNetwork";
import { SPHERE_RADIUS } from "./boardConstants";

export interface Point2 {
  x: number;
  y: number;
}
export interface RoadPath {
  exit?: number;
  points: Point2[];
  width: number;
  endWidth?: number;
  kind: "road" | "path";
}
export const BOARD_SCALE = (0.4 * Math.PI) / 2;
export const TILE_RADIUS = 0.166;

export function hashSeed(value: string): number {
  let hash = 2166136261;
  for (const byte of new TextEncoder().encode(value)) {
    hash = Math.imul(hash ^ byte, 16777619);
  }
  return hash >>> 0;
}
export function boardCenter(c: HexCoordinate): Point2 {
  return { x: 0.3 * Math.sqrt(3) * (c.q + c.r / 2) * BOARD_SCALE, y: -0.45 * c.r * BOARD_SCALE };
}
export function projectBoardPoint(
  x: number,
  y: number,
  height = 0,
  target = new THREE.Vector3(),
): THREE.Vector3 {
  const distance = Math.hypot(x, y);
  const radius = SPHERE_RADIUS + height;
  if (distance < 1e-10) {
    return target.set(0, 0, radius);
  }
  const scale = (Math.sin(distance / SPHERE_RADIUS) * radius) / distance;
  return target.set(x * scale, y * scale, Math.cos(distance / SPHERE_RADIUS) * radius);
}
export function nearestOnRoad(point: Point2, roads: RoadPath[]) {
  let result = { x: point.x, y: point.y, halfWidth: 0, distance: Infinity };
  for (const road of roads) {
    for (let i = 1; i < road.points.length; i++) {
      const a = road.points[i - 1],
        b = road.points[i];
      const dx = b.x - a.x,
        dy = b.y - a.y;
      const length = dx * dx + dy * dy;
      const t = length
        ? Math.max(0, Math.min(1, ((point.x - a.x) * dx + (point.y - a.y) * dy) / length))
        : 0;
      const x = a.x + dx * t,
        y = a.y + dy * t;
      const distance = Math.hypot(point.x - x, point.y - y);
      if (distance - road.width / 2 < result.distance - result.halfWidth) {
        result = { x, y, halfWidth: road.width / 2, distance };
      }
    }
  }
  return result;
}
export function polar(radius: number, angle: number): Point2 {
  return { x: Math.cos(angle) * radius, y: Math.sin(angle) * radius };
}
export function circleRoad(
  radius: number,
  width = 0.007,
  offset: Point2 = { x: 0, y: 0 },
): RoadPath {
  return {
    kind: "road",
    width,
    points: Array.from({ length: 97 }, (_, i) => {
      const p = polar(radius, (i * Math.PI) / 48);
      return { x: p.x + offset.x, y: p.y + offset.y };
    }),
  };
}

// Get a biome value (0-1) for a tile based on large-scale noise
// This creates regions with different vegetation densities
export function getBiomeValue(q: number, r: number): number {
  // Use larger scale for regional variation (divide coords to get smoother regions)
  const scale = 0.3;
  const x = q * scale;
  const y = r * scale;

  // Layer multiple noise octaves for more interesting patterns
  const n1 = noise2D(x, y, 12345);
  const n2 = noise2D(x * 2, y * 2, 67890) * 0.5;
  const n3 = noise2D(x * 4, y * 4, 11111) * 0.25;

  return Math.min(1, Math.max(0, (n1 + n2 + n3) / 1.75));
}

export function noise2D(x: number, y: number, seed: number): number {
  const hash = (ix: number, iy: number) => {
    const n = ix * 374761393 + iy * 668265263 + seed;
    return ((n * (n * n * 15731 + 789221) + 1376312589) >>> 0) / 4294967296;
  };

  const floorX = Math.floor(x);
  const floorY = Math.floor(y);
  const fracX = x - floorX;
  const fracY = y - floorY;

  const v00 = hash(floorX, floorY);
  const v10 = hash(floorX + 1, floorY);
  const v01 = hash(floorX, floorY + 1);
  const v11 = hash(floorX + 1, floorY + 1);

  const sx = fracX * fracX * (3 - 2 * fracX);
  const sy = fracY * fracY * (3 - 2 * fracY);

  const nx0 = v00 * (1 - sx) + v10 * sx;
  const nx1 = v01 * (1 - sx) + v11 * sx;

  return nx0 * (1 - sy) + nx1 * sy;
}

/** Shared vertices along the strip keep spline bends and circular roads seamless. */
export function createRoadRibbon(
  roads: RoadPath[],
  extra: number,
  height: number,
  births: number[] = [],
  project = true,
): THREE.BufferGeometry {
  const positions: number[] = [],
    uvs: number[] = [],
    appearances: number[] = [];
  for (const [roadIndex, road] of roads.entries()) {
    const points: Point2[] = [road.points[0]];
    for (let i = 1; i < road.points.length; i++) {
      const a = road.points[i - 1],
        b = road.points[i];
      const span = Math.hypot(b.x - a.x, b.y - a.y);
      if (span < 1e-10) {
        continue;
      }
      const steps = Math.max(1, Math.ceil(span / 0.006));
      for (let step = 1; step <= steps; step++) {
        points.push({
          x: a.x + ((b.x - a.x) * step) / steps,
          y: a.y + ((b.y - a.y) * step) / steps,
        });
      }
    }
    if (points.length < 2) {
      continue;
    }
    const closed =
      Math.hypot(points[0].x - points.at(-1)!.x, points[0].y - points.at(-1)!.y) < 1e-8;
    const lengths = [0];
    for (let i = 1; i < points.length; i++) {
      lengths.push(
        lengths[i - 1] + Math.hypot(points[i].x - points[i - 1].x, points[i].y - points[i - 1].y),
      );
    }
    const edges = points.map((p, i) => {
      let before = points[Math.max(0, i - 1)],
        after = points[Math.min(points.length - 1, i + 1)];
      if (closed && (i === 0 || i === points.length - 1)) {
        before = points[points.length - 2];
        after = points[1];
      }
      let ax = p.x - before.x,
        ay = p.y - before.y,
        bx = after.x - p.x,
        by = after.y - p.y;
      const al = Math.hypot(ax, ay),
        bl = Math.hypot(bx, by);
      if (al > 1e-10) {
        ax /= al;
        ay /= al;
      } else {
        ax = bx / bl;
        ay = by / bl;
      }
      if (bl > 1e-10) {
        bx /= bl;
        by /= bl;
      } else {
        bx = ax;
        by = ay;
      }
      const span = Math.hypot(ax + bx, ay + by);
      const nx = -(ay + by) / Math.max(span, 1e-10),
        ny = (ax + bx) / Math.max(span, 1e-10);
      const progress = lengths[i] / Math.max(1e-10, lengths.at(-1)!);
      const width =
        (road.width + ((road.endWidth ?? road.width) - road.width) * progress) / 2 + extra;
      const miter = width / Math.max(0.65, nx * -by + ny * bx);
      return [
        { x: p.x - nx * miter, y: p.y - ny * miter },
        { x: p.x + nx * miter, y: p.y + ny * miter },
      ];
    });
    for (let i = 1; i < edges.length; i++) {
      const [a, b] = edges[i - 1],
        [c, d] = edges[i];
      for (const p of [a, c, d, a, d, b]) {
        const vertex = project
          ? projectBoardPoint(p.x, p.y, height)
          : new THREE.Vector3(p.x, p.y, height);
        positions.push(vertex.x, vertex.y, vertex.z);
        uvs.push(p.x * 35, p.y * 35);
        appearances.push(births[roadIndex] ?? -1000);
      }
    }
  }
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setAttribute("connectionBirth", new THREE.Float32BufferAttribute(appearances, 1));
  geometry.computeVertexNormals();
  return geometry;
}

const groundPatchCache = new Map<string, THREE.BufferGeometry>();

/** Static coverage is baked once; only the shared noise maps are sampled each frame. */
export function createLandscapeGround(
  plan: LandscapePlan,
  coast: ReturnType<typeof connectOceanGround>,
) {
  const occupied = new Set(coast.tiles.map((t) => HexGrid2D.coordinateToKey(t.coordinate)));
  const boundaries: { a: Point2; b: Point2 }[] = [];
  for (const tile of coast.tiles) {
    const center = boardCenter(tile.coordinate);
    for (const neighbor of HexGrid2D.getNeighbors(tile.coordinate)) {
      if (occupied.has(HexGrid2D.coordinateToKey(neighbor))) {
        continue;
      }
      const n = boardCenter(neighbor),
        dx = n.x - center.x,
        dy = n.y - center.y;
      const span = Math.hypot(dx, dy),
        mx = (center.x + n.x) / 2,
        my = (center.y + n.y) / 2;
      const tx = (-dy / span) * 0.0942477796,
        ty = (dx / span) * 0.0942477796;
      boundaries.push({ a: { x: mx - tx, y: my - ty }, b: { x: mx + tx, y: my + ty } });
    }
  }
  const holes = plan.tiles
    .filter((t) => t.layout?.style.cover === "dome" || t.layout?.style.ground === "recessed")
    .map((t) => ({
      ...boardCenter(t.coordinate),
      radius: t.layout!.style.cover === "dome" ? 0.118 : 0.07,
    }));
  const patches: THREE.BufferGeometry[] = [];
  const projected = new THREE.Vector3();
  const segmentDistance = (p: Point2, a: Point2, b: Point2) => {
    const dx = b.x - a.x,
      dy = b.y - a.y;
    const t = Math.max(
      0,
      Math.min(1, ((p.x - a.x) * dx + (p.y - a.y) * dy) / Math.max(dx * dx + dy * dy, 1e-7)),
    );
    const x = p.x - a.x - dx * t,
      y = p.y - a.y - dy * t;
    return Math.sqrt(x * x + y * y);
  };
  for (const tile of coast.tiles) {
    const c = boardCenter(tile.coordinate);
    const nearby = (a: Point2, b: Point2, margin: number) =>
      c.x >= Math.min(a.x, b.x) - margin &&
      c.x <= Math.max(a.x, b.x) + margin &&
      c.y >= Math.min(a.y, b.y) - margin &&
      c.y <= Math.max(a.y, b.y) + margin;
    const circles = plan.circles.filter(
      (circle) => Math.hypot(c.x - circle.x, c.y - circle.y) < circle.radius + 0.28,
    );
    const links = coast.links.filter((link) =>
      nearby(link.points[0], link.points[1], link.width / 2 + 0.28),
    );
    const localHoles = holes.filter((h) => Math.hypot(h.x - c.x, h.y - c.y) < h.radius + 0.28);
    const edges = boundaries.filter((edge) => nearby(edge.a, edge.b, 0.28));
    const water = coast.water.filter((p) => Math.hypot(c.x - p.x, c.y - p.y) < 0.5);
    const waterLinks = coast.waterLinks.filter((link) => nearby(link.a, link.b, 0.5));
    const key = JSON.stringify([
      tile.coordinate,
      circles,
      links,
      localHoles,
      edges,
      water,
      waterLinks,
    ]);
    const cached = groundPatchCache.get(key);
    if (cached) {
      groundPatchCache.delete(key);
      groundPatchCache.set(key, cached);
      patches.push(cached);
      continue;
    }
    const positions: number[] = [],
      normals: number[] = [],
      uv: number[] = [],
      board: number[] = [],
      fields: number[] = [],
      shore: number[] = [],
      indices: number[] = [];
    const radius = (0.3 * Math.PI * 0.4) / 2,
      steps = 12;
    for (let edge = 0; edge < 6; edge++) {
      const angle = Math.PI / 6 + (edge * Math.PI) / 3;
      const b = { x: Math.cos(angle) * radius, y: Math.sin(angle) * radius };
      const d = {
        x: Math.cos(angle + Math.PI / 3) * radius,
        y: Math.sin(angle + Math.PI / 3) * radius,
      };
      const base = positions.length / 3;
      const index = (row: number, col: number) => base + (row * (row + 1)) / 2 + col;
      for (let row = 0; row <= steps; row++) {
        for (let col = 0; col <= row; col++) {
          const p = {
            x: c.x + (b.x * (row - col)) / steps + (d.x * col) / steps,
            y: c.y + (b.y * (row - col)) / steps + (d.y * col) / steps,
          };
          projectBoardPoint(p.x, p.y, 0.0003, projected);
          positions.push(projected.x, projected.y, projected.z);
          projected.normalize();
          normals.push(projected.x, projected.y, projected.z);
          uv.push(p.x * 22, p.y * 22);
          board.push(p.x, p.y);
          let coverage = -1,
            forest = -1,
            hole = 1,
            boundary = 1;
          for (const circle of circles) {
            const distance =
              circle.radius - Math.sqrt((p.x - circle.x) ** 2 + (p.y - circle.y) ** 2);
            coverage = Math.max(coverage, distance);
            if (circle.forest) {
              forest = Math.max(forest, distance);
            }
          }
          for (const link of links) {
            const distance = link.width / 2 - segmentDistance(p, link.points[0], link.points[1]);
            coverage = Math.max(coverage, distance);
          }
          for (const h of localHoles) {
            hole = Math.min(hole, Math.hypot(p.x - h.x, p.y - h.y) - h.radius);
          }
          for (const edge of edges) {
            boundary = Math.min(boundary, segmentDistance(p, edge.a, edge.b));
          }
          fields.push(coverage, forest, hole, boundary);
          let shoreDistance = 1;
          for (const w of water) {
            shoreDistance = Math.min(
              shoreDistance,
              Math.hypot(p.x - w.x, p.y - w.y) - 0.17 * BOARD_SCALE,
            );
          }
          for (const link of waterLinks) {
            shoreDistance = Math.min(
              shoreDistance,
              segmentDistance(p, link.a, link.b) - 0.17 * BOARD_SCALE,
            );
          }
          shore.push(shoreDistance);
          if (row < steps) {
            indices.push(index(row, col), index(row + 1, col), index(row + 1, col + 1));
            if (col < row) {
              indices.push(index(row, col), index(row + 1, col + 1), index(row, col + 1));
            }
          }
        }
      }
    }
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
    geometry.setAttribute("normal", new THREE.Float32BufferAttribute(normals, 3));
    geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uv, 2));
    geometry.setAttribute("boardPosition", new THREE.Float32BufferAttribute(board, 2));
    geometry.setAttribute("landscapeField", new THREE.Float32BufferAttribute(fields, 4));
    geometry.setAttribute("shoreDistance", new THREE.Float32BufferAttribute(shore, 1));
    geometry.setIndex(indices);
    groundPatchCache.set(key, geometry);
    patches.push(geometry);
    if (groundPatchCache.size > 128) {
      const oldest = groundPatchCache.keys().next().value!;
      groundPatchCache.get(oldest)!.dispose();
      groundPatchCache.delete(oldest);
    }
  }
  return patches.length ? mergeGeometries(patches)! : new THREE.BufferGeometry();
}

export interface LandscapeGroundData {
  attributes: Record<string, { array: Float32Array; itemSize: number }>;
  index: Uint16Array | Uint32Array | null;
}

export function packLandscapeGround(geometry: THREE.BufferGeometry): LandscapeGroundData {
  return {
    attributes: Object.fromEntries(
      Object.entries(geometry.attributes).map(([key, attr]) => [
        key,
        {
          array: attr.array as Float32Array,
          itemSize: attr.itemSize,
        },
      ]),
    ),
    index: (geometry.index?.array as Uint16Array | Uint32Array | null) ?? null,
  };
}

export function unpackLandscapeGround(data: LandscapeGroundData) {
  const geometry = new THREE.BufferGeometry();
  for (const [key, attr] of Object.entries(data.attributes)) {
    geometry.setAttribute(key, new THREE.BufferAttribute(attr.array, attr.itemSize));
  }
  if (data.index) {
    geometry.setIndex(new THREE.BufferAttribute(data.index, 1));
  }
  return geometry;
}
