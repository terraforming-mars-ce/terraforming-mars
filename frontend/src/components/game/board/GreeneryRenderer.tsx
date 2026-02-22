import { useMemo, useRef, useEffect, useLayoutEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { mergeGeometries } from "three/examples/jsm/utils/BufferGeometryUtils.js";
import { useModels } from "../../../hooks/useModels";
import { useTextures } from "../../../hooks/useTextures";
import {
  greeneryGroundVertexSnippet,
  greeneryGroundFragmentSnippet,
  splitSnippet,
} from "./shaders";
import { SPHERE_RADIUS, easeOutCubic } from "./boardConstants";

// Module-level temp objects for animation (avoid per-frame allocation)
const _tmpMatrix = new THREE.Matrix4();
const _tmpQuat = new THREE.Quaternion();
const _tmpScale = new THREE.Vector3();

// Animation timing config (ms)
const ANIM_DELAYS = { ground: 0, rocks: 100, bushes: 150, trees: 200, clover: 250 };
const ANIM_DURATIONS = { ground: 500, rocks: 400, bushes: 400, trees: 500, clover: 350 };

export const TREE_NAMES = ["Tree-01-1", "Tree-01-2", "Tree-01-3", "Tree-01-4"];
export const BUSH_NAMES = ["Bush-01", "Bush-02", "Bush-03", "Bush-04", "Bush-05"];
export const CLOVER_NAMES = ["Clover-01", "Clover-02", "Clover-03", "Clover-04", "Clover-05"];

export interface TreePrimitive {
  geometry: THREE.BufferGeometry;
  material: THREE.Material;
}

export interface TreeVariant {
  primitives: TreePrimitive[];
}

// Module-level cache for shared geometry/materials
export const variantCache: {
  trees: TreeVariant[] | null;
  bushes: TreeVariant[] | null;
  clover: TreeVariant[] | null;
  rock: { geometry: THREE.BufferGeometry; material: THREE.Material } | null;
} = {
  trees: null,
  bushes: null,
  clover: null,
  rock: null,
};

function mulberry32(seed: number): () => number {
  return function () {
    let t = (seed += 0x6d2b79f5);
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function getTileSeed(q: number, r: number, s: number): number {
  let h = (q + 128) * 73856093;
  h = Math.imul(h ^ (h >>> 16), 0x45d9f3b);
  h = h + (r + 128) * 19349663;
  h = Math.imul(h ^ (h >>> 16), 0x45d9f3b);
  h = h + (s + 128) * 83492791;
  h = Math.imul(h ^ (h >>> 16), 0x45d9f3b);
  return (h >>> 0) % 2147483647;
}

// Get a biome value (0-1) for a tile based on large-scale noise
// This creates regions with different vegetation densities
function getBiomeValue(q: number, r: number): number {
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

function noise2D(x: number, y: number, seed: number): number {
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

function isInsideHex(x: number, y: number, radius: number, margin: number = 0): boolean {
  const effectiveRadius = radius - margin;
  // Swap x/y to match pointy-top hex orientation (rotated π/2)
  const absX = Math.abs(y);
  const absY = Math.abs(x);
  const q2 = effectiveRadius;
  const q1 = (effectiveRadius * Math.sqrt(3)) / 2;
  return absY <= q1 && q1 * absX + 0.5 * effectiveRadius * absY <= q1 * q2;
}

function randomHexPosition(
  rng: () => number,
  radius: number,
  margin: number,
  existingPositions: { x: number; y: number }[],
  minSpacing: number,
  maxAttempts: number = 50,
): { x: number; y: number } | null {
  for (let i = 0; i < maxAttempts; i++) {
    const x = (rng() - 0.5) * 2 * radius;
    const y = (rng() - 0.5) * 2 * radius;

    if (!isInsideHex(x, y, radius, margin)) continue;

    let tooClose = false;
    for (const pos of existingPositions) {
      const dx = x - pos.x;
      const dy = y - pos.y;
      if (dx * dx + dy * dy < minSpacing * minSpacing) {
        tooClose = true;
        break;
      }
    }

    if (!tooClose) {
      return { x, y };
    }
  }
  return null;
}

function createNoisyHexGeometry(
  radius: number,
  rings: number,
  noiseScale: number,
  noiseAmplitude: number,
  seed: number,
): THREE.BufferGeometry {
  const geometry = new THREE.BufferGeometry();
  const vertices: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  vertices.push(0, 0, noise2D(0, 0, seed) * noiseAmplitude);
  uvs.push(0.5, 0.5);

  for (let ring = 1; ring <= rings; ring++) {
    const ringRadius = (ring / rings) * radius;
    const verticesInRing = 6 * ring;

    for (let i = 0; i < verticesInRing; i++) {
      const edgeIndex = Math.floor(i / ring);
      const posOnEdge = i % ring;

      const angle1 = (edgeIndex * Math.PI) / 3;
      const angle2 = ((edgeIndex + 1) * Math.PI) / 3;

      const t = posOnEdge / ring;
      const x = ringRadius * (Math.cos(angle1) * (1 - t) + Math.cos(angle2) * t);
      const y = ringRadius * (Math.sin(angle1) * (1 - t) + Math.sin(angle2) * t);

      const z = noise2D(x * noiseScale, y * noiseScale, seed) * noiseAmplitude;

      vertices.push(x, y, z);
      uvs.push(0.5 + (x / radius) * 0.5, 0.5 + (y / radius) * 0.5);
    }
  }

  for (let i = 0; i < 6; i++) {
    const next = (i + 1) % 6;
    indices.push(0, 1 + i, 1 + next);
  }

  let prevRingStart = 1;
  for (let ring = 2; ring <= rings; ring++) {
    const currRingStart = prevRingStart + 6 * (ring - 1);
    const prevRingVerts = 6 * (ring - 1);
    const currRingVerts = 6 * ring;

    let prevIdx = 0;
    let currIdx = 0;

    for (let edge = 0; edge < 6; edge++) {
      for (let i = 0; i < ring; i++) {
        const curr0 = currRingStart + currIdx;
        const curr1 = currRingStart + ((currIdx + 1) % currRingVerts);

        if (i < ring - 1) {
          const prev0 = prevRingStart + prevIdx;
          const prev1 = prevRingStart + ((prevIdx + 1) % prevRingVerts);

          indices.push(prev0, curr0, curr1);
          indices.push(prev0, curr1, prev1);
          prevIdx++;
        } else {
          const prev0 = prevRingStart + (prevIdx % prevRingVerts);
          indices.push(prev0, curr0, curr1);
        }
        currIdx++;
      }
    }

    prevRingStart = currRingStart;
  }

  geometry.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();

  return geometry;
}

export function addSphereProjectionWithSoftEdges(
  material: THREE.Material,
  zOffset: number,
  noiseMap: THREE.Texture,
  noiseMapHigh: THREE.Texture,
  hexRadius: number,
): void {
  const grassOverflow = hexRadius * 0.25;
  const bandWidth = hexRadius * 0.35;
  const warpAmount = hexRadius * 0.2;
  const noiseScale = 1.5 / hexRadius;

  const vertSnippet = splitSnippet(greeneryGroundVertexSnippet);
  const fragSnippet = splitSnippet(greeneryGroundFragmentSnippet);

  material.onBeforeCompile = (shader) => {
    (material as any).__shader = shader;
    shader.uniforms.uSphereRadius = { value: SPHERE_RADIUS };
    shader.uniforms.uZOffset = { value: zOffset };
    shader.uniforms.uNoiseMap = { value: noiseMap };
    shader.uniforms.uNoiseMapHigh = { value: noiseMapHigh };
    shader.uniforms.uHexRadius = { value: hexRadius };
    shader.uniforms.uGrassOverflow = { value: grassOverflow };
    shader.uniforms.uBandWidth = { value: bandWidth };
    shader.uniforms.uWarpAmount = { value: warpAmount };
    shader.uniforms.uNoiseScale = { value: noiseScale };
    shader.uniforms.uFadeProgress = { value: 1.0 };

    shader.vertexShader =
      vertSnippet.header +
      "\n" +
      shader.vertexShader.replace("#include <begin_vertex>", vertSnippet.body);

    shader.fragmentShader =
      fragSnippet.header +
      "\n" +
      shader.fragmentShader.replace("#include <alphamap_fragment>", fragSnippet.body);
  };
}

export function createVariantsFromScene(
  scene: THREE.Object3D,
  names: string[],
  targetHeight: number,
): TreeVariant[] {
  const variants: TreeVariant[] = [];

  for (const name of names) {
    const obj = scene.getObjectByName(name);
    if (!obj) continue;

    const primitives: TreePrimitive[] = [];
    const allGeometries: THREE.BufferGeometry[] = [];

    obj.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        child.updateWorldMatrix(true, false);
        const baseGeo = child.geometry.clone();
        baseGeo.applyMatrix4(child.matrixWorld);

        if (Array.isArray(child.material)) {
          const groups = child.geometry.groups;
          child.material.forEach((mat, idx) => {
            if (groups && groups[idx]) {
              const group = groups[idx];
              const geo = baseGeo.clone();
              const indices = baseGeo.index!.array.slice(group.start, group.start + group.count);
              geo.setIndex(Array.from(indices));
              primitives.push({ geometry: geo, material: mat.clone() });
              allGeometries.push(geo.clone());
            }
          });
        } else {
          primitives.push({
            geometry: baseGeo,
            material: (child.material as THREE.Material).clone(),
          });
          allGeometries.push(baseGeo.clone());
        }
      }
    });

    if (primitives.length === 0) continue;

    const mergedForBounds = mergeGeometries(allGeometries);
    if (!mergedForBounds) continue;

    const rotationMatrix = new THREE.Matrix4().makeRotationX(Math.PI / 2);
    mergedForBounds.applyMatrix4(rotationMatrix);

    const box = new THREE.Box3().setFromBufferAttribute(
      mergedForBounds.getAttribute("position") as THREE.BufferAttribute,
    );
    const size = box.getSize(new THREE.Vector3());
    const center = box.getCenter(new THREE.Vector3());

    const maxDim = Math.max(size.x, size.y, size.z);
    const scale = targetHeight / maxDim;

    const transform = new THREE.Matrix4()
      .makeScale(scale, scale, scale)
      .multiply(new THREE.Matrix4().makeTranslation(-center.x, -center.y, -box.min.z))
      .multiply(rotationMatrix);

    primitives.forEach((p) => {
      p.geometry.applyMatrix4(transform);
      p.geometry.computeVertexNormals();
      if (p.material instanceof THREE.Material) {
        p.material.depthWrite = true;
        p.material.depthTest = true;
        const mat = p.material as THREE.MeshStandardMaterial;
        mat.alphaTest = 0.15;
        mat.polygonOffset = true;
        mat.polygonOffsetFactor = 1;
        mat.polygonOffsetUnits = 1;
        mat.emissive = new THREE.Color(0x4a6a35);
        mat.emissiveIntensity = 0.35;
        if (mat.map) mat.emissiveMap = mat.map;
        if (mat.map) {
          mat.map.minFilter = THREE.LinearMipmapLinearFilter;
          mat.map.magFilter = THREE.LinearFilter;
          mat.map.anisotropy = 4;
        }
      }
    });

    variants.push({ primitives });
  }

  return variants;
}

interface GreeneryTileData {
  coordinate: { q: number; r: number; s: number };
  worldPosition: THREE.Vector3;
  normal: THREE.Vector3;
}

interface GreeneryRendererProps {
  tiles: GreeneryTileData[];
  volcanoTiles?: GreeneryTileData[];
  newTileKeys: Set<string>;
  hexRadius?: number;
}

export default function GreeneryRenderer({
  tiles,
  volcanoTiles = [],
  newTileKeys,
  hexRadius = 0.166,
}: GreeneryRendererProps) {
  const treeInstanceRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const bushInstanceRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const cloverInstanceRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const rockInstanceRef = useRef<THREE.InstancedMesh>(null);
  const groundMeshRefs = useRef<THREE.Mesh[]>([]);

  const { treesScene, rockScene } = useModels();
  const {
    grass: grassTexture,
    rock: rockTexture,
    noiseMid: noiseTexture,
    noiseHigh: noiseHighTexture,
  } = useTextures();

  // Create variants once (cached)
  const treeVariants = useMemo(() => {
    if (!variantCache.trees) {
      variantCache.trees = createVariantsFromScene(treesScene, TREE_NAMES, 0.08);
    }
    return variantCache.trees;
  }, [treesScene]);

  const bushVariants = useMemo(() => {
    if (!variantCache.bushes) {
      variantCache.bushes = createVariantsFromScene(treesScene, BUSH_NAMES, 0.035);
    }
    return variantCache.bushes;
  }, [treesScene]);

  const cloverVariants = useMemo(() => {
    if (!variantCache.clover) {
      variantCache.clover = createVariantsFromScene(treesScene, CLOVER_NAMES, 0.012);
    }
    return variantCache.clover;
  }, [treesScene]);

  const { geometry: rockGeometry, material: rockMaterial } = useMemo(() => {
    if (variantCache.rock) {
      return variantCache.rock;
    }

    let geo: THREE.BufferGeometry = new THREE.DodecahedronGeometry(0.015, 1);

    rockScene.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        const name = child.name.toLowerCase();
        if (name.includes("plane") || name.includes("ground")) {
          return;
        }
        geo = child.geometry.clone();
        child.updateWorldMatrix(true, false);
        geo.applyMatrix4(child.matrixWorld);
      }
    });

    const box = new THREE.Box3().setFromBufferAttribute(
      geo.getAttribute("position") as THREE.BufferAttribute,
    );
    const size = box.getSize(new THREE.Vector3());

    const targetSize = 0.04;
    const maxDim = Math.max(size.x, size.y, size.z);
    const scale = targetSize / maxDim;

    const rotationMatrix = new THREE.Matrix4().makeRotationX(Math.PI / 2);
    geo.applyMatrix4(rotationMatrix);

    const boxRotated = new THREE.Box3().setFromBufferAttribute(
      geo.getAttribute("position") as THREE.BufferAttribute,
    );
    const centerRotated = boxRotated.getCenter(new THREE.Vector3());

    const transform = new THREE.Matrix4()
      .makeScale(scale, scale, scale)
      .multiply(
        new THREE.Matrix4().makeTranslation(-centerRotated.x, -centerRotated.y, -boxRotated.min.z),
      );

    geo.applyMatrix4(transform);
    geo.computeVertexNormals();

    const mat = new THREE.MeshStandardMaterial({
      map: rockTexture,
      color: 0xffffff,
      roughness: 0.9,
      metalness: 0.0,
    });

    variantCache.rock = { geometry: geo, material: mat };
    return variantCache.rock;
  }, [rockScene, rockTexture]);

  // Generate all instance data for all tiles
  const { treeInstances, bushInstances, cloverInstances, rockInstances, groundData } =
    useMemo(() => {
      const treeInstances: {
        position: THREE.Vector3;
        rotation: number;
        scale: number;
        variantIdx: number;
        tint: number;
        tileKey: string;
      }[] = [];
      const bushInstances: {
        position: THREE.Vector3;
        rotation: number;
        scale: number;
        variantIdx: number;
        tint: number;
        tileKey: string;
      }[] = [];
      const cloverInstances: {
        position: THREE.Vector3;
        rotation: number;
        scale: number;
        variantIdx: number;
        tileKey: string;
      }[] = [];
      const rockInstances: {
        position: THREE.Vector3;
        rotation: THREE.Euler;
        scale: number;
        tileKey: string;
      }[] = [];
      const groundData: {
        geometry: THREE.BufferGeometry;
        position: THREE.Vector3;
        normal: THREE.Vector3;
        tileKey: string;
      }[] = [];

      for (const tile of tiles) {
        const tileKey = `${tile.coordinate.q},${tile.coordinate.r},${tile.coordinate.s}`;
        const seed = getTileSeed(tile.coordinate.q, tile.coordinate.r, tile.coordinate.s);

        // Get biome value for this tile (0-1)
        // High = forested (more trees), Low = scrubland (more bushes/rocks)
        const biome = getBiomeValue(tile.coordinate.q, tile.coordinate.r);

        // Pre-roll rock tier (separate RNG, doesn't affect tree sequence)
        const rockPreRoll = mulberry32(seed + 12345)();
        const hasBigRock = rockPreRoll >= 0.6;
        const hasMountain = rockPreRoll >= 0.85;

        // Ground geometry per tile (unique due to noise)
        const groundGeo = createNoisyHexGeometry(hexRadius * 2.0, 8, 6, 0.008, seed);
        groundGeo.rotateZ(Math.PI / 2);
        groundData.push({
          geometry: groundGeo,
          position: tile.worldPosition,
          normal: tile.normal,
          tileKey,
        });

        // Generate rocks FIRST so vegetation can avoid them
        const rockRng = mulberry32(seed + 12345);
        const rockExclusions: { x: number; y: number; r: number }[] = [];
        const rockSpacing: { x: number; y: number }[] = [];
        const rockRoll = rockRng();
        const rockBaseSize = 0.04;

        let rockScales: number[] = [];
        if (rockRoll < 0.1) {
          // 10%: no rocks
        } else if (rockRoll < 0.55) {
          // 45%: 2-6 small rocks
          const count = 2 + Math.floor(rockRng() * 5);
          for (let i = 0; i < count; i++) {
            rockScales.push(0.3 + rockRng() * 0.4);
          }
        } else if (rockRoll < 0.85) {
          // 25%: 1 large rock + 1-3 small companions
          rockScales.push(1.2 + rockRng() * 0.6);
          const companions = 1 + Math.floor(rockRng() * 3);
          for (let i = 0; i < companions; i++) {
            rockScales.push(0.2 + rockRng() * 0.4);
          }
        } else {
          // 15%: 1 mountain rock + 2-4 small companions
          rockScales.push(2.4 + rockRng() * 0.3);
          const companions = 2 + Math.floor(rockRng() * 3);
          for (let i = 0; i < companions; i++) {
            rockScales.push(0.3 + rockRng() * 0.5);
          }
        }

        for (let i = 0; i < rockScales.length; i++) {
          const pos = randomHexPosition(rockRng, hexRadius * 0.75, 0.03, rockSpacing, 0.04);
          if (pos) {
            rockSpacing.push(pos);
            rockExclusions.push({ x: pos.x, y: pos.y, r: rockBaseSize * rockScales[i] * 0.6 });
            const localPos = new THREE.Vector3(pos.x, pos.y, 0.0);
            const worldPos = localPos
              .clone()
              .applyMatrix4(
                new THREE.Matrix4().compose(
                  tile.worldPosition,
                  new THREE.Quaternion().setFromUnitVectors(
                    new THREE.Vector3(0, 0, 1),
                    tile.normal,
                  ),
                  new THREE.Vector3(1, 1, 1),
                ),
              );
            rockInstances.push({
              position: worldPos,
              rotation: new THREE.Euler(
                rockRng() * Math.PI * 2,
                rockRng() * Math.PI * 2,
                rockRng() * Math.PI * 2,
              ),
              scale: rockScales[i],
              tileKey,
            });
          }
        }

        const isInsideRock = (x: number, y: number): boolean => {
          for (const rock of rockExclusions) {
            const dx = x - rock.x;
            const dy = y - rock.y;
            if (dx * dx + dy * dy < rock.r * rock.r) return true;
          }
          return false;
        };

        // Generate trees - more in high biome areas, boosted near mountains
        const treeRng = mulberry32(seed);
        const treeMult = hasMountain ? 1.8 : hasBigRock ? 1.4 : 1.0;
        const baseTreeCount = Math.floor(treeRng() * 10) + 9;
        const treeCount = Math.floor(baseTreeCount * (0.3 + biome * 1.4) * treeMult);
        const treePositions: { x: number; y: number }[] = [];

        for (let i = 0; i < treeCount; i++) {
          const pos = randomHexPosition(treeRng, hexRadius * 0.9, 0.012, treePositions, 0.025);
          if (pos) {
            if (isInsideRock(pos.x, pos.y)) continue;
            treePositions.push(pos);
            const localPos = new THREE.Vector3(pos.x, pos.y, 0.003);
            const worldPos = localPos
              .clone()
              .applyMatrix4(
                new THREE.Matrix4().compose(
                  tile.worldPosition,
                  new THREE.Quaternion().setFromUnitVectors(
                    new THREE.Vector3(0, 0, 1),
                    tile.normal,
                  ),
                  new THREE.Vector3(1, 1, 1),
                ),
              );
            const tint = noise2D(worldPos.x * 8, worldPos.y * 8, 77777);
            let rockProximityBoost = 0;
            for (const rock of rockExclusions) {
              const dx = pos.x - rock.x;
              const dy = pos.y - rock.y;
              const dist = Math.sqrt(dx * dx + dy * dy);
              const influenceRadius = rock.r * 3;
              if (dist < influenceRadius) {
                const proximity = 1 - dist / influenceRadius;
                rockProximityBoost = Math.max(rockProximityBoost, proximity * 0.5);
              }
            }
            const baseTreeScale = hasMountain ? 1.3 + treeRng() * 0.3 : 1.1 + treeRng() * 0.3;
            const treeScale = baseTreeScale + rockProximityBoost;
            treeInstances.push({
              position: worldPos,
              rotation: treeRng() * Math.PI * 2,
              scale: treeScale,
              variantIdx: Math.floor(treeRng() * 4),
              tint,
              tileKey,
            });
          }
        }

        // Generate bushes - more in low biome areas (inverse of trees)
        const bushRng = mulberry32(seed + 54321);
        const baseBushCount = Math.floor(bushRng() * 100) + 160;
        const bushCount = Math.floor(baseBushCount * (0.5 + (1 - biome) * 1.2));
        const bushPositions: { x: number; y: number }[] = [];

        for (let i = 0; i < bushCount; i++) {
          const pos = randomHexPosition(bushRng, hexRadius * 0.92, 0.003, bushPositions, 0.006);
          if (pos) {
            bushPositions.push(pos);
            const localPos = new THREE.Vector3(pos.x, pos.y, 0.001);
            const worldPos = localPos
              .clone()
              .applyMatrix4(
                new THREE.Matrix4().compose(
                  tile.worldPosition,
                  new THREE.Quaternion().setFromUnitVectors(
                    new THREE.Vector3(0, 0, 1),
                    tile.normal,
                  ),
                  new THREE.Vector3(1, 1, 1),
                ),
              );
            const bushTint = noise2D(worldPos.x * 8, worldPos.y * 8, 77777);
            bushInstances.push({
              position: worldPos,
              rotation: bushRng() * Math.PI * 2,
              scale: 0.7 + bushRng() * 0.6,
              variantIdx: Math.floor(bushRng() * 5),
              tint: bushTint,
              tileKey,
            });
          }
        }

        // Generate clover for this tile
        const cloverRng = mulberry32(seed + 99999);
        const cloverCount = Math.floor(cloverRng() * 20) + 40;
        const cloverPositions: { x: number; y: number }[] = [];

        for (let i = 0; i < cloverCount; i++) {
          const pos = randomHexPosition(cloverRng, hexRadius * 0.95, 0.005, cloverPositions, 0.01);
          if (pos) {
            if (isInsideRock(pos.x, pos.y)) continue;
            cloverPositions.push(pos);
            const localPos = new THREE.Vector3(pos.x, pos.y, 0.0005);
            const worldPos = localPos
              .clone()
              .applyMatrix4(
                new THREE.Matrix4().compose(
                  tile.worldPosition,
                  new THREE.Quaternion().setFromUnitVectors(
                    new THREE.Vector3(0, 0, 1),
                    tile.normal,
                  ),
                  new THREE.Vector3(1, 1, 1),
                ),
              );
            cloverInstances.push({
              position: worldPos,
              rotation: cloverRng() * Math.PI * 2,
              scale: 0.5 + cloverRng() * 1.0,
              variantIdx: Math.floor(cloverRng() * 5),
              tileKey,
            });
          }
        }
      }

      // === Volcano tiles: trees + bushes OUTSIDE the dark volcano area ===
      const volcanoExclusionRadius = 0.105;
      for (const tile of volcanoTiles) {
        const tileKey = `${tile.coordinate.q},${tile.coordinate.r},${tile.coordinate.s}`;
        const seed = getTileSeed(tile.coordinate.q, tile.coordinate.r, tile.coordinate.s);

        const isInsideVolcano = (x: number, y: number): boolean => {
          return x * x + y * y < volcanoExclusionRadius * volcanoExclusionRadius;
        };

        const tileMatrix = new THREE.Matrix4().compose(
          tile.worldPosition,
          new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 0, 1), tile.normal),
          new THREE.Vector3(1, 1, 1),
        );

        // Trees around volcano (80% scale, placed outside exclusion zone)
        const vTreeRng = mulberry32(seed + 77777);
        const vTreeCount = 8 + Math.floor(vTreeRng() * 6);
        const vTreePositions: { x: number; y: number }[] = [];

        for (let i = 0; i < vTreeCount; i++) {
          const pos = randomHexPosition(vTreeRng, hexRadius * 0.9, 0.012, vTreePositions, 0.025);
          if (pos) {
            if (isInsideVolcano(pos.x, pos.y)) continue;
            vTreePositions.push(pos);
            const worldPos = new THREE.Vector3(pos.x, pos.y, 0.003).applyMatrix4(tileMatrix);
            const tint = noise2D(worldPos.x * 8, worldPos.y * 8, 77777);
            treeInstances.push({
              position: worldPos,
              rotation: vTreeRng() * Math.PI * 2,
              scale: (0.9 + vTreeRng() * 0.2) * 0.8,
              variantIdx: Math.floor(vTreeRng() * 4),
              tint,
              tileKey,
            });
          }
        }

        // Bushes around volcano (placed outside exclusion zone)
        const vBushRng = mulberry32(seed + 88888);
        const vBushPositions: { x: number; y: number }[] = [];

        for (let i = 0; i < 200; i++) {
          const pos = randomHexPosition(vBushRng, hexRadius * 0.92, 0.003, vBushPositions, 0.006);
          if (pos) {
            if (isInsideVolcano(pos.x, pos.y)) continue;
            vBushPositions.push(pos);
            const worldPos = new THREE.Vector3(pos.x, pos.y, 0.001).applyMatrix4(tileMatrix);
            const bushTint = noise2D(worldPos.x * 8, worldPos.y * 8, 77777);
            bushInstances.push({
              position: worldPos,
              rotation: vBushRng() * Math.PI * 2,
              scale: 0.6 + vBushRng() * 0.5,
              variantIdx: Math.floor(vBushRng() * 5),
              tint: bushTint,
              tileKey,
            });
          }
        }
      }

      return { treeInstances, bushInstances, cloverInstances, rockInstances, groundData };
    }, [tiles, volcanoTiles, hexRadius]);

  // Group instances by variant
  const treesByVariant = useMemo(() => {
    const groups: {
      position: THREE.Vector3;
      rotation: number;
      scale: number;
      tint: number;
      tileKey: string;
    }[][] = [[], [], [], []];
    for (const inst of treeInstances) {
      groups[inst.variantIdx].push({
        position: inst.position,
        rotation: inst.rotation,
        scale: inst.scale,
        tint: inst.tint,
        tileKey: inst.tileKey,
      });
    }
    return groups;
  }, [treeInstances]);

  const bushesByVariant = useMemo(() => {
    const groups: {
      position: THREE.Vector3;
      rotation: number;
      scale: number;
      tint: number;
      tileKey: string;
    }[][] = [[], [], [], [], []];
    for (const inst of bushInstances) {
      groups[inst.variantIdx].push({
        position: inst.position,
        rotation: inst.rotation,
        scale: inst.scale,
        tint: inst.tint,
        tileKey: inst.tileKey,
      });
    }
    return groups;
  }, [bushInstances]);

  const cloverByVariant = useMemo(() => {
    const groups: {
      position: THREE.Vector3;
      rotation: number;
      scale: number;
      tileKey: string;
    }[][] = [[], [], [], [], []];
    for (const inst of cloverInstances) {
      groups[inst.variantIdx].push({
        position: inst.position,
        rotation: inst.rotation,
        scale: inst.scale,
        tileKey: inst.tileKey,
      });
    }
    return groups;
  }, [cloverInstances]);

  // Ground material (shared)
  const groundMaterial = useMemo(() => {
    const mat = new THREE.MeshStandardMaterial({
      map: grassTexture,
      color: new THREE.Color(0.4, 0.45, 0.35),
      roughness: 0.9,
      metalness: 0.0,
      side: THREE.DoubleSide,
      transparent: true,
      alphaTest: 0.01,
      depthWrite: false,
    });
    addSphereProjectionWithSoftEdges(mat, 0.003, noiseTexture, noiseHighTexture, hexRadius);
    return mat;
  }, [grassTexture, noiseTexture, noiseHighTexture, hexRadius]);

  // --- Animation tracking ---
  const pendingAnimTilesRef = useRef<Set<string>>(new Set());
  const warmupTilesRef = useRef<Set<string>>(new Set());
  const tileAnimStartRef = useRef<Map<string, number>>(new Map());
  const groundMaterialClonesRef = useRef<Map<string, THREE.MeshStandardMaterial>>(new Map());

  useEffect(() => {
    for (const key of newTileKeys) {
      pendingAnimTilesRef.current.add(key);
    }
  }, [newTileKeys]);

  useFrame(({ clock }) => {
    const elapsed = clock.elapsedTime;

    // Step 1: Tiles that warmed up last frame now get their animation start time.
    // The warmup frame forced shader compilation, so the clock starts fresh here.
    for (const key of warmupTilesRef.current) {
      if (!tileAnimStartRef.current.has(key)) {
        tileAnimStartRef.current.set(key, elapsed);
      }
    }
    warmupTilesRef.current.clear();

    // Step 2: New pending tiles get one warmup frame (render at scale 0, compile shaders)
    for (const key of pendingAnimTilesRef.current) {
      warmupTilesRef.current.add(key);
    }
    pendingAnimTilesRef.current.clear();

    // Early return if nothing is animating
    if (tileAnimStartRef.current.size === 0) return;

    const completedKeys: string[] = [];

    for (const [tileKey, startTime] of tileAnimStartRef.current) {
      const timeSinceStart = (elapsed - startTime) * 1000; // ms

      // Check if all elements are done (max total = 700ms for trees)
      const maxTotal = ANIM_DELAYS.trees + ANIM_DURATIONS.trees;
      if (timeSinceStart >= maxTotal) {
        completedKeys.push(tileKey);
        continue;
      }

      // Ground fade
      const cloneMat = groundMaterialClonesRef.current.get(tileKey);
      if (cloneMat) {
        const shader = (cloneMat as any).__shader;
        if (shader) {
          const groundT = Math.max(
            0,
            Math.min(1, (timeSinceStart - ANIM_DELAYS.ground) / ANIM_DURATIONS.ground),
          );
          shader.uniforms.uFadeProgress.value = easeOutCubic(groundT);
        }
      }

      // Vegetation grow - rocks
      const rockDelay = ANIM_DELAYS.rocks;
      const rockDur = ANIM_DURATIONS.rocks;
      for (let i = 0; i < rockInstances.length; i++) {
        if (rockInstances[i].tileKey !== tileKey) continue;
        const rawT = Math.max(0, Math.min(1, (timeSinceStart - rockDelay) / rockDur));
        const animScale = rockInstances[i].scale * easeOutCubic(rawT);
        const data = rockInstances[i];
        _tmpQuat.setFromEuler(data.rotation);
        _tmpScale.setScalar(animScale);
        _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
        if (rockInstanceRef.current) {
          rockInstanceRef.current.setMatrixAt(i, _tmpMatrix);
          rockInstanceRef.current.instanceMatrix.needsUpdate = true;
        }
      }

      // Vegetation grow - trees
      const treeDelay = ANIM_DELAYS.trees;
      const treeDur = ANIM_DURATIONS.trees;
      treesByVariant.forEach((transforms, variantIdx) => {
        const variant = treeVariants[variantIdx];
        if (!variant) return;
        variant.primitives.forEach((_, primIdx) => {
          const key = `tree-${variantIdx}-${primIdx}`;
          const mesh = treeInstanceRefs.current.get(key);
          if (!mesh) return;
          transforms.forEach((data, i) => {
            if (data.tileKey !== tileKey) return;
            const rawT = Math.max(0, Math.min(1, (timeSinceStart - treeDelay) / treeDur));
            const animScale = data.scale * easeOutCubic(rawT);
            _tmpQuat.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
            _tmpScale.setScalar(animScale);
            _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
            mesh.setMatrixAt(i, _tmpMatrix);
            mesh.instanceMatrix.needsUpdate = true;
          });
        });
      });

      // Vegetation grow - bushes
      const bushDelay = ANIM_DELAYS.bushes;
      const bushDur = ANIM_DURATIONS.bushes;
      bushesByVariant.forEach((transforms, variantIdx) => {
        const variant = bushVariants[variantIdx];
        if (!variant) return;
        variant.primitives.forEach((_, primIdx) => {
          const key = `bush-${variantIdx}-${primIdx}`;
          const mesh = bushInstanceRefs.current.get(key);
          if (!mesh) return;
          transforms.forEach((data, i) => {
            if (data.tileKey !== tileKey) return;
            const rawT = Math.max(0, Math.min(1, (timeSinceStart - bushDelay) / bushDur));
            const animScale = data.scale * easeOutCubic(rawT);
            _tmpQuat.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
            _tmpScale.set(animScale, animScale, animScale);
            _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
            mesh.setMatrixAt(i, _tmpMatrix);
            mesh.instanceMatrix.needsUpdate = true;
          });
        });
      });

      // Vegetation grow - clover
      const cloverDelay = ANIM_DELAYS.clover;
      const cloverDur = ANIM_DURATIONS.clover;
      cloverByVariant.forEach((transforms, variantIdx) => {
        const variant = cloverVariants[variantIdx];
        if (!variant) return;
        variant.primitives.forEach((_, primIdx) => {
          const key = `clover-${variantIdx}-${primIdx}`;
          const mesh = cloverInstanceRefs.current.get(key);
          if (!mesh) return;
          transforms.forEach((data, i) => {
            if (data.tileKey !== tileKey) return;
            const rawT = Math.max(0, Math.min(1, (timeSinceStart - cloverDelay) / cloverDur));
            const animScale = data.scale * easeOutCubic(rawT);
            _tmpQuat.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
            _tmpScale.set(animScale, animScale, animScale);
            _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
            mesh.setMatrixAt(i, _tmpMatrix);
            mesh.instanceMatrix.needsUpdate = true;
          });
        });
      });
    }

    // Cleanup completed animations — write final full-scale matrices for ALL element types
    for (const tileKey of completedKeys) {
      tileAnimStartRef.current.delete(tileKey);

      // Final matrices for rocks
      for (let i = 0; i < rockInstances.length; i++) {
        if (rockInstances[i].tileKey !== tileKey) continue;
        const data = rockInstances[i];
        _tmpQuat.setFromEuler(data.rotation);
        _tmpScale.setScalar(data.scale);
        _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
        if (rockInstanceRef.current) {
          rockInstanceRef.current.setMatrixAt(i, _tmpMatrix);
          rockInstanceRef.current.instanceMatrix.needsUpdate = true;
        }
      }

      // Final matrices for trees
      treesByVariant.forEach((transforms, variantIdx) => {
        const variant = treeVariants[variantIdx];
        if (!variant) return;
        variant.primitives.forEach((_, primIdx) => {
          const key = `tree-${variantIdx}-${primIdx}`;
          const mesh = treeInstanceRefs.current.get(key);
          if (!mesh) return;
          transforms.forEach((data, i) => {
            if (data.tileKey !== tileKey) return;
            _tmpQuat.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
            _tmpScale.setScalar(data.scale);
            _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
            mesh.setMatrixAt(i, _tmpMatrix);
            mesh.instanceMatrix.needsUpdate = true;
          });
        });
      });

      // Final matrices for bushes
      bushesByVariant.forEach((transforms, variantIdx) => {
        const variant = bushVariants[variantIdx];
        if (!variant) return;
        variant.primitives.forEach((_, primIdx) => {
          const key = `bush-${variantIdx}-${primIdx}`;
          const mesh = bushInstanceRefs.current.get(key);
          if (!mesh) return;
          transforms.forEach((data, i) => {
            if (data.tileKey !== tileKey) return;
            _tmpQuat.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
            _tmpScale.set(data.scale, data.scale, data.scale);
            _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
            mesh.setMatrixAt(i, _tmpMatrix);
            mesh.instanceMatrix.needsUpdate = true;
          });
        });
      });

      // Final matrices for clover
      cloverByVariant.forEach((transforms, variantIdx) => {
        const variant = cloverVariants[variantIdx];
        if (!variant) return;
        variant.primitives.forEach((_, primIdx) => {
          const key = `clover-${variantIdx}-${primIdx}`;
          const mesh = cloverInstanceRefs.current.get(key);
          if (!mesh) return;
          transforms.forEach((data, i) => {
            if (data.tileKey !== tileKey) return;
            _tmpQuat.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
            _tmpScale.set(data.scale, data.scale, data.scale);
            _tmpMatrix.compose(data.position, _tmpQuat, _tmpScale);
            mesh.setMatrixAt(i, _tmpMatrix);
            mesh.instanceMatrix.needsUpdate = true;
          });
        });
      });

      // Clean up ground material clone
      const cloneMat = groundMaterialClonesRef.current.get(tileKey);
      if (cloneMat) {
        const groundIdx = groundData.findIndex((g) => g.tileKey === tileKey);
        if (groundIdx >= 0 && groundMeshRefs.current[groundIdx]) {
          groundMeshRefs.current[groundIdx].material = groundMaterial;
        }
        cloneMat.dispose();
        groundMaterialClonesRef.current.delete(tileKey);
      }
    }
  });

  // Update tree instance matrices and colors
  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4();
    const quaternion = new THREE.Quaternion();
    const scaleVec = new THREE.Vector3();
    const color = new THREE.Color();

    treesByVariant.forEach((transforms, variantIdx) => {
      if (transforms.length === 0) return;

      const variant = treeVariants[variantIdx];
      if (!variant) return;

      variant.primitives.forEach((_, primIdx) => {
        const key = `tree-${variantIdx}-${primIdx}`;
        const mesh = treeInstanceRefs.current.get(key);
        if (!mesh) return;

        transforms.forEach((data, i) => {
          const isAnimating = newTileKeys.has(data.tileKey);
          quaternion.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
          scaleVec.setScalar(isAnimating ? 0 : data.scale);
          matrix.compose(data.position, quaternion, scaleVec);
          mesh.setMatrixAt(i, matrix);

          const t = data.tint;
          const r = 0.3 + t * 0.55;
          const g = 0.4 + t * 0.6;
          const b = 0.15 + t * 0.25;
          color.setRGB(r, g, b);
          mesh.setColorAt(i, color);
        });

        mesh.instanceMatrix.needsUpdate = true;
        if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
      });
    });
  }, [treesByVariant, treeVariants, newTileKeys]);

  // Update bush instance matrices and colors
  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4();
    const quaternion = new THREE.Quaternion();
    const color = new THREE.Color();

    bushesByVariant.forEach((transforms, variantIdx) => {
      if (transforms.length === 0) return;

      const variant = bushVariants[variantIdx];
      if (!variant) return;

      variant.primitives.forEach((_, primIdx) => {
        const key = `bush-${variantIdx}-${primIdx}`;
        const mesh = bushInstanceRefs.current.get(key);
        if (!mesh) return;

        transforms.forEach((data, i) => {
          const isAnimating = newTileKeys.has(data.tileKey);
          const s = isAnimating ? 0 : data.scale;
          quaternion.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
          matrix.compose(data.position, quaternion, new THREE.Vector3(s, s, s));
          mesh.setMatrixAt(i, matrix);

          const t = data.tint;
          const r = 0.4 + t * 0.4;
          const g = 0.5 + t * 0.5;
          const b = 0.25 + t * 0.2;
          color.setRGB(r, g, b);
          mesh.setColorAt(i, color);
        });

        mesh.instanceMatrix.needsUpdate = true;
        if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
      });
    });
  }, [bushesByVariant, bushVariants, newTileKeys]);

  // Update clover instance matrices
  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4();
    const quaternion = new THREE.Quaternion();

    cloverByVariant.forEach((transforms, variantIdx) => {
      if (transforms.length === 0) return;

      const variant = cloverVariants[variantIdx];
      if (!variant) return;

      variant.primitives.forEach((_, primIdx) => {
        const key = `clover-${variantIdx}-${primIdx}`;
        const mesh = cloverInstanceRefs.current.get(key);
        if (!mesh) return;

        transforms.forEach((data, i) => {
          const isAnimating = newTileKeys.has(data.tileKey);
          const s = isAnimating ? 0 : data.scale;
          quaternion.setFromAxisAngle(new THREE.Vector3(0, 0, 1), data.rotation);
          matrix.compose(data.position, quaternion, new THREE.Vector3(s, s, s));
          mesh.setMatrixAt(i, matrix);
        });

        mesh.instanceMatrix.needsUpdate = true;
      });
    });
  }, [cloverByVariant, cloverVariants, newTileKeys]);

  // Update rock instance matrices
  useLayoutEffect(() => {
    if (!rockInstanceRef.current || rockInstances.length === 0) return;

    const mesh = rockInstanceRef.current;
    const matrix = new THREE.Matrix4();
    const quaternion = new THREE.Quaternion();

    rockInstances.forEach((data, i) => {
      const isAnimating = newTileKeys.has(data.tileKey);
      const s = isAnimating ? 0 : data.scale;
      quaternion.setFromEuler(data.rotation);
      matrix.compose(data.position, quaternion, new THREE.Vector3(s, s, s));
      mesh.setMatrixAt(i, matrix);
    });

    mesh.instanceMatrix.needsUpdate = true;
  }, [rockInstances, newTileKeys]);

  // Warm-up geometry: tiny plane to force shader compilation on mount
  const warmupGeometry = useMemo(() => new THREE.PlaneGeometry(0.001, 0.001), []);

  return (
    <group>
      {/* Warm-up mesh forces ground material shader compilation before first tile */}
      {groundData.length === 0 && (
        <mesh geometry={warmupGeometry} material={groundMaterial} renderOrder={-1} />
      )}

      {/* Ground meshes - one per tile */}
      {groundData.map((ground, idx) => {
        const quaternion = new THREE.Quaternion().setFromUnitVectors(
          new THREE.Vector3(0, 0, 1),
          ground.normal,
        );
        const isAnimating = newTileKeys.has(ground.tileKey);
        let mat = groundMaterial;
        if (isAnimating) {
          let clone = groundMaterialClonesRef.current.get(ground.tileKey);
          if (!clone) {
            clone = groundMaterial.clone();
            addSphereProjectionWithSoftEdges(
              clone,
              0.003,
              noiseTexture,
              noiseHighTexture,
              hexRadius,
            );
            clone.needsUpdate = true;
            groundMaterialClonesRef.current.set(ground.tileKey, clone);
          }
          mat = clone;
        }
        return (
          <mesh
            key={`ground-${idx}`}
            ref={(el) => {
              if (el) groundMeshRefs.current[idx] = el;
            }}
            geometry={ground.geometry}
            material={mat}
            position={ground.position}
            quaternion={quaternion}
            renderOrder={12}
            receiveShadow
          />
        );
      })}

      {/* Trees - single InstancedMesh per variant */}
      {treeVariants.map((variant, variantIdx) => {
        const transforms = treesByVariant[variantIdx];
        if (!transforms || transforms.length === 0) return null;

        return variant.primitives.map((prim, primIdx) => (
          <instancedMesh
            key={`tree-${variantIdx}-${primIdx}`}
            ref={(el) => {
              treeInstanceRefs.current.set(`tree-${variantIdx}-${primIdx}`, el);
            }}
            args={[prim.geometry, prim.material, transforms.length]}
            frustumCulled={false}
            renderOrder={15}
          />
        ));
      })}

      {/* Bushes - single InstancedMesh per variant */}
      {bushVariants.map((variant, variantIdx) => {
        const transforms = bushesByVariant[variantIdx];
        if (!transforms || transforms.length === 0) return null;

        return variant.primitives.map((prim, primIdx) => (
          <instancedMesh
            key={`bush-${variantIdx}-${primIdx}`}
            ref={(el) => {
              bushInstanceRefs.current.set(`bush-${variantIdx}-${primIdx}`, el);
            }}
            args={[prim.geometry, prim.material, transforms.length]}
            frustumCulled={false}
            renderOrder={15}
          />
        ));
      })}

      {/* Clover - single InstancedMesh per variant */}
      {cloverVariants.map((variant, variantIdx) => {
        const transforms = cloverByVariant[variantIdx];
        if (!transforms || transforms.length === 0) return null;

        return variant.primitives.map((prim, primIdx) => (
          <instancedMesh
            key={`clover-${variantIdx}-${primIdx}`}
            ref={(el) => {
              cloverInstanceRefs.current.set(`clover-${variantIdx}-${primIdx}`, el);
            }}
            args={[prim.geometry, prim.material, transforms.length]}
            frustumCulled={false}
            renderOrder={15}
          />
        ));
      })}

      {/* Rocks - single InstancedMesh for all */}
      {rockInstances.length > 0 && (
        <instancedMesh
          ref={rockInstanceRef}
          args={[rockGeometry, rockMaterial, rockInstances.length]}
          frustumCulled={false}
          renderOrder={15}
        />
      )}
    </group>
  );
}
