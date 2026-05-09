import * as THREE from "three";
import { HexGrid2D, type HexCoordinate } from "../../../utils/hex-grid-2d";

export interface OceanPoint {
  x: number;
  y: number;
  emergence: number;
}

export interface OceanDataResult {
  texture: THREE.DataTexture;
  data: Float32Array;
  pointCount: number;
  edgeCount: number;
  points: OceanPoint[];
}

const HEX_SIZE = 0.3;

export function hexToPixel(coord: HexCoordinate): { x: number; y: number } {
  const x = HEX_SIZE * Math.sqrt(3) * (coord.q + coord.r / 2);
  const y = -((HEX_SIZE * 3) / 2) * coord.r;
  return { x, y };
}

export function buildOceanDataTexture(oceanCoordinates: HexCoordinate[]): OceanDataResult {
  const oceanKeySet = new Set<string>();
  const points: OceanPoint[] = [];
  const keyToIndex = new Map<string, number>();

  for (const coord of oceanCoordinates) {
    const key = HexGrid2D.coordinateToKey(coord);
    if (oceanKeySet.has(key)) {
      continue;
    }
    oceanKeySet.add(key);
    const pos = hexToPixel(coord);
    keyToIndex.set(key, points.length);
    points.push({ x: pos.x, y: pos.y, emergence: 1.0 });
  }

  const edges: Array<{ aIdx: number; bIdx: number }> = [];
  const edgeSet = new Set<string>();

  for (const coord of oceanCoordinates) {
    const key = HexGrid2D.coordinateToKey(coord);
    const aIdx = keyToIndex.get(key);
    if (aIdx === undefined) {
      continue;
    }

    const neighbors = HexGrid2D.getNeighbors(coord);
    for (const neighbor of neighbors) {
      const neighborKey = HexGrid2D.coordinateToKey(neighbor);
      const bIdx = keyToIndex.get(neighborKey);
      if (bIdx === undefined) {
        continue;
      }

      const edgeKey = aIdx < bIdx ? `${aIdx}:${bIdx}` : `${bIdx}:${aIdx}`;
      if (edgeSet.has(edgeKey)) {
        continue;
      }
      edgeSet.add(edgeKey);
      edges.push({ aIdx, bIdx });
    }
  }

  const pointCount = points.length;
  const edgeCount = edges.length;
  const texWidth = Math.max(pointCount + edgeCount * 2, 1);
  const data = new Float32Array(texWidth * 4);

  for (let i = 0; i < pointCount; i++) {
    const offset = i * 4;
    data[offset] = points[i].x;
    data[offset + 1] = points[i].y;
    data[offset + 2] = points[i].emergence;
    data[offset + 3] = 0;
  }

  for (let i = 0; i < edgeCount; i++) {
    const edge = edges[i];
    const posOffset = (pointCount + i * 2) * 4;
    data[posOffset] = points[edge.aIdx].x;
    data[posOffset + 1] = points[edge.aIdx].y;
    data[posOffset + 2] = points[edge.bIdx].x;
    data[posOffset + 3] = points[edge.bIdx].y;
    const idxOffset = (pointCount + i * 2 + 1) * 4;
    data[idxOffset] = edge.aIdx;
    data[idxOffset + 1] = edge.bIdx;
  }

  const texture = new THREE.DataTexture(data, texWidth, 1, THREE.RGBAFormat, THREE.FloatType);
  texture.minFilter = THREE.NearestFilter;
  texture.magFilter = THREE.NearestFilter;
  texture.wrapS = THREE.ClampToEdgeWrapping;
  texture.wrapT = THREE.ClampToEdgeWrapping;
  texture.needsUpdate = true;

  return { texture, data, pointCount, edgeCount, points };
}

export function updateEmergenceInTexture(
  data: Float32Array,
  pointIndex: number,
  emergence: number,
): void {
  data[pointIndex * 4 + 2] = emergence;
}
