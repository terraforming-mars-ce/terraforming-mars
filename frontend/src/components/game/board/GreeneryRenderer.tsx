import { memo, useMemo, useRef, useLayoutEffect } from "react";
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
import type { LandscapePlant, LandscapeState } from "./landscapeTypes";
import { landscapeHeightAt, TRANSITION_MS } from "./landscapeFields";

import { projectBoardPoint } from "./landscapeGeometry";
export const TREE_NAMES = ["Tree-01-1", "Tree-01-2", "Tree-01-3", "Tree-01-4"];
export const BUSH_NAMES = ["Bush-01", "Bush-02", "Bush-03", "Bush-04", "Bush-05"];
export const CLOVER_NAMES = ["Clover-01", "Clover-02", "Clover-03", "Clover-04", "Clover-05"];
export const FLOWER_NAMES = ["Circle003", "Circle008", "Circle015"];

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
  flowers: TreeVariant[] | null;
  rock: { geometry: THREE.BufferGeometry; material: THREE.Material } | null;
} = {
  trees: null,
  bushes: null,
  clover: null,
  flowers: null,
  rock: null,
};

export function addSphereProjectionWithSoftEdges(
  material: THREE.Material,
  zOffset: number,
  noiseMap: THREE.Texture,
  noiseMapHigh: THREE.Texture,
  hexRadius: number,
  sphereCenter?: THREE.Vector3,
  groupInverseMatrix?: THREE.Matrix4,
  edges: {
    circular?: boolean;
    overflow?: number;
    bandWidth?: number;
    warp?: number;
    sphereRadius?: number;
  } = {},
): void {
  const grassOverflow = hexRadius * (edges.overflow ?? 0.25);
  const bandWidth = hexRadius * (edges.bandWidth ?? 0.35);
  const warpAmount = hexRadius * (edges.warp ?? 0.2);
  const noiseScale = 1.5 / hexRadius;
  const centerVec = sphereCenter || new THREE.Vector3(0, 0, 0);
  const invMatrix = groupInverseMatrix || new THREE.Matrix4();

  const vertSnippet = splitSnippet(greeneryGroundVertexSnippet);
  const fragSnippet = splitSnippet(greeneryGroundFragmentSnippet);

  material.onBeforeCompile = (shader) => {
    (material as any).__shader = shader;
    shader.uniforms.uSphereRadius = { value: edges.sphereRadius ?? SPHERE_RADIUS };
    shader.uniforms.uZOffset = { value: zOffset };
    shader.uniforms.uSphereCenter = { value: centerVec };
    shader.uniforms.uGroupInverseMatrix = { value: invMatrix };
    shader.uniforms.uNoiseMap = { value: noiseMap };
    shader.uniforms.uNoiseMapHigh = { value: noiseMapHigh };
    shader.uniforms.uHexRadius = { value: hexRadius };
    shader.uniforms.uCircularEdge = { value: edges.circular ?? false };
    shader.uniforms.uGrassOverflow = { value: grassOverflow };
    shader.uniforms.uBandWidth = { value: bandWidth };
    shader.uniforms.uWarpAmount = { value: warpAmount };
    shader.uniforms.uNoiseScale = { value: noiseScale };
    shader.uniforms.uFadeProgress = { value: 1.0 };

    shader.vertexShader =
      vertSnippet.header +
      "\n" +
      shader.vertexShader.replace("#include <begin_vertex>", vertSnippet.body).replace(
        "#include <project_vertex>",
        `vec4 mvPosition = viewMatrix * vec4(projectedPos, 1.0);
           gl_Position = projectionMatrix * mvPosition;`,
      );

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

function GreeneryRenderer({
  landscape,
  previousLandscape,
  transitionStart,
}: {
  landscape: LandscapeState;
  previousLandscape: LandscapeState;
  transitionStart: number;
}) {
  const { treesScene, rockScene, flowersScene } = useModels();
  const { rock: rockTexture } = useTextures();
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

  const flowerVariants = useMemo(() => {
    if (!variantCache.flowers) {
      variantCache.flowers = createVariantsFromScene(flowersScene, FLOWER_NAMES, 0.025);
    }
    return variantCache.flowers;
  }, [flowersScene]);

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

  const groups = useMemo(() => {
    const result = new Map<string, LandscapePlant[]>();
    for (const p of landscape.plants) {
      const counts = {
        tree: TREE_NAMES.length,
        bush: BUSH_NAMES.length,
        clover: CLOVER_NAMES.length,
        flower: FLOWER_NAMES.length,
        rock: 1,
      };
      const key = `${p.kind}:${p.seed % counts[p.kind]}`;
      const list = result.get(key) ?? [];
      list.push(p);
      result.set(key, list);
    }
    return result;
  }, [landscape]);
  const variants = [
    ...treeVariants.map((v, i) => ({ kind: "tree", i, v })),
    ...bushVariants.map((v, i) => ({ kind: "bush", i, v })),
    ...cloverVariants.map((v, i) => ({ kind: "clover", i, v })),
    ...flowerVariants.map((v, i) => ({ kind: "flower", i, v })),
    { kind: "rock", i: 0, v: { primitives: [{ geometry: rockGeometry, material: rockMaterial }] } },
  ];
  return (
    <group>
      {variants.flatMap(({ kind, i, v }) =>
        v.primitives.map((p, j) => (
          <PlantBatch
            key={`${kind}:${i}:${j}`}
            geometry={p.geometry}
            material={p.material}
            plants={groups.get(`${kind}:${i}`) ?? EMPTY_PLANTS}
            revision={landscape.id}
            transitionStart={transitionStart}
            previousLandscape={previousLandscape}
          />
        )),
      )}
    </group>
  );
}
const EMPTY_PLANTS: LandscapePlant[] = [];
const UP = new THREE.Vector3(0, 0, 1);
function PlantBatch({
  geometry,
  material,
  plants,
  revision,
  transitionStart,
  previousLandscape,
}: {
  geometry: THREE.BufferGeometry;
  material: THREE.Material;
  plants: LandscapePlant[];
  revision: number;
  transitionStart: number;
  previousLandscape: LandscapeState;
}) {
  const mesh = useRef<THREE.InstancedMesh>(null);
  const slots = useRef(new Map<string, { slot: number; plant: LandscapePlant }>());
  const animations = useRef(
    new Map<number, { plant: LandscapePlant; fromHeight: number; birth: number; grow: boolean }>(),
  );
  const seen = useRef(false);
  const scratch = useMemo(
    () => ({
      matrix: new THREE.Matrix4(),
      q: new THREE.Quaternion(),
      turn: new THREE.Quaternion(),
      scale: new THREE.Vector3(),
      position: new THREE.Vector3(),
      normal: new THREE.Vector3(),
      color: new THREE.Color(),
    }),
    [],
  );
  const write = (slot: number, plant: LandscapePlant, growth: number, height: number) => {
    projectBoardPoint(plant.x, plant.y, height, scratch.position);
    scratch.normal.copy(scratch.position).normalize();
    scratch.q.setFromUnitVectors(UP, scratch.normal);
    scratch.turn.setFromAxisAngle(UP, (plant.seed % 6283) / 1000);
    scratch.q.multiply(scratch.turn);
    scratch.scale.setScalar(plant.scale * growth);
    scratch.matrix.compose(scratch.position, scratch.q, scratch.scale);
    mesh.current!.setMatrixAt(slot, scratch.matrix);
    mesh.current!.instanceMatrix.addUpdateRange(slot * 16, 16);
  };
  useLayoutEffect(() => {
    const target = mesh.current!;
    let changed = false;
    const ids = new Set(plants.map((p) => p.id));
    for (const [id, entry] of slots.current) {
      if (!ids.has(id)) {
        changed = true;
        scratch.matrix.makeScale(0, 0, 0);
        target.setMatrixAt(entry.slot, scratch.matrix);
        target.instanceMatrix.addUpdateRange(entry.slot * 16, 16);
        animations.current.delete(entry.slot);
        slots.current.delete(id);
      }
    }
    const used = new Set([...slots.current.values()].map((e) => e.slot));
    let next = 0;
    for (const p of plants) {
      const prior = slots.current.get(p.id);
      if (prior?.plant === p) {
        continue;
      }
      changed = true;
      while (used.has(next)) {
        next++;
      }
      const slot = prior?.slot ?? next;
      if (slot >= 4096) {
        throw new Error("Vegetation batch capacity exceeded");
      }
      used.add(slot);
      const grow = seen.current && !prior;
      const previousHeight = previousLandscape.id
        ? landscapeHeightAt(previousLandscape.patches, p.x, p.y)
        : p.height;
      write(slot, p, grow ? 0 : 1, previousHeight);
      if (grow || previousHeight !== p.height) {
        animations.current.set(slot, {
          plant: p,
          fromHeight: previousHeight,
          birth: transitionStart,
          grow,
        });
      }
      const tint = 0.65 + (p.seed % 100) / 400;
      scratch.color.setRGB(tint, tint, tint);
      target.setColorAt(slot, scratch.color);
      target.instanceColor!.addUpdateRange(slot * 3, 3);
      slots.current.set(p.id, { slot, plant: p });
    }
    target.count = used.size ? Math.max(...used) + 1 : 0;
    if (changed) {
      target.instanceMatrix.needsUpdate = true;
      if (target.instanceColor) {
        target.instanceColor.needsUpdate = true;
      }
    }
    if (revision > 0) {
      seen.current = true;
    }
  }, [plants, revision, transitionStart, previousLandscape]);
  useFrame(() => {
    if (!animations.current.size) {
      return;
    }
    const now = performance.now();
    for (const [slot, a] of animations.current) {
      const t = Math.max(0, Math.min(1, (now - a.birth) / TRANSITION_MS));
      const progress = easeOutCubic(t);
      const heightProgress = t * t * (3 - 2 * t);
      write(
        slot,
        a.plant,
        a.grow ? progress : 1,
        a.fromHeight + (a.plant.height - a.fromHeight) * heightProgress,
      );
      if (progress === 1) {
        animations.current.delete(slot);
      }
    }
    mesh.current!.instanceMatrix.needsUpdate = true;
  });
  return (
    <instancedMesh
      ref={mesh}
      args={[geometry, material, 4096]}
      frustumCulled={false}
      raycast={() => {}}
      renderOrder={15}
    />
  );
}
export default memo(GreeneryRenderer);
