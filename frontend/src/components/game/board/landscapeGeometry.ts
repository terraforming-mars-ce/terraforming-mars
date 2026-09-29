import * as THREE from "three";
import { type HexCoordinate } from "../../../utils/hex-grid-2d";
import { SPHERE_RADIUS } from "./boardConstants";

export interface Point2 {
  x: number;
  y: number;
}
export interface RoadPath {
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
  surfaceRadius = SPHERE_RADIUS,
): THREE.Vector3 {
  const distance = Math.hypot(x, y);
  const radius = surfaceRadius + height;
  if (distance < 1e-10) {
    return target.set(0, 0, radius);
  }
  const scale = (Math.sin(distance / surfaceRadius) * radius) / distance;
  return target.set(x * scale, y * scale, Math.cos(distance / surfaceRadius) * radius);
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
/** Shared vertices along the strip keep spline bends and circular roads seamless. */
export function createRoadRibbon(
  roads: RoadPath[],
  extra: number,
  height: number,
  project = true,
): THREE.BufferGeometry {
  const positions: number[] = [],
    uvs: number[] = [];
  for (const road of roads) {
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
      }
    }
  }
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.computeVertexNormals();
  return geometry;
}
