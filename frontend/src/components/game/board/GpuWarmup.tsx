import { useMemo, useRef, useLayoutEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { SkeletonUtils } from "three-stdlib";
import {
  addOceanProjection,
  createVolcanoMaterial,
  createNuclearZoneMaterial,
  createWorldTreeMaterial,
} from "./shaders";
import { computeFlowMap } from "./volcanoFlowMap";
import {
  variantCache,
  createVariantsFromScene,
  TREE_NAMES,
  BUSH_NAMES,
  CLOVER_NAMES,
  FLOWER_NAMES,
  type TreeVariant,
} from "./GreeneryRenderer";
import { useModels } from "../../../hooks/useModels";
import { useTextures } from "../../../hooks/useTextures";
import sunCoronaVert from "./shaders/sun-corona.vert.glsl?raw";
import sunCoronaFrag from "./shaders/sun-corona.frag.glsl?raw";

const WARMUP_SCALE = 0.001;
const WARMUP_FRAMES = 3;

interface GpuWarmupProps {
  onReady?: () => void;
}

export default function GpuWarmup({ onReady }: GpuWarmupProps) {
  const { treesScene, rockScene, cityScene, flowersScene } = useModels();
  const textures = useTextures();
  const {
    rock: rockTexture,
    sand: sandTexture,
    waterNormals,
    smoke: smokeTexture,
    grass: grassTexture,
  } = textures;

  const frameCount = useRef(0);
  const readyFired = useRef(false);

  useFrame(() => {
    if (readyFired.current) return;
    frameCount.current++;
    if (frameCount.current >= WARMUP_FRAMES) {
      readyFired.current = true;
      onReady?.();
    }
  });

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
    if (variantCache.rock) return variantCache.rock;

    let geo: THREE.BufferGeometry = new THREE.DodecahedronGeometry(0.015, 1);
    rockScene.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        const name = child.name.toLowerCase();
        if (name.includes("plane") || name.includes("ground")) return;
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

  const warmupCity = useMemo(() => {
    const cloned = SkeletonUtils.clone(cityScene);
    cloned.scale.setScalar(WARMUP_SCALE);
    return cloned;
  }, [cityScene]);

  const oceanWarmupMaterial = useMemo(() => {
    waterNormals.wrapS = waterNormals.wrapT = THREE.RepeatWrapping;
    const mat = new THREE.MeshStandardMaterial({
      transparent: true,
      side: THREE.DoubleSide,
      premultipliedAlpha: true,
    });
    addOceanProjection(mat, waterNormals, sandTexture, new THREE.Vector3(), 0.008, {
      uSeedOffset: new THREE.Vector2(50, 50),
    });
    mat.needsUpdate = true;
    return mat;
  }, [waterNormals, sandTexture]);

  const oceanGeometry = useMemo(() => new THREE.CircleGeometry(0.3, 32), []);

  const smokeWarmupMaterial = useMemo(() => {
    return new THREE.MeshBasicMaterial({
      map: smokeTexture,
      transparent: true,
      opacity: 0,
      depthWrite: false,
      depthTest: false,
      blending: THREE.NormalBlending,
      side: THREE.DoubleSide,
    });
  }, [smokeTexture]);

  const smokeGeometry = useMemo(() => new THREE.PlaneGeometry(0.3, 0.3), []);

  const volcanoFlowTexture = useMemo(() => computeFlowMap(42), []);
  const volcanoWarmupMaterial = useMemo(() => {
    return createVolcanoMaterial(grassTexture, volcanoFlowTexture, 42);
  }, [grassTexture, volcanoFlowTexture]);

  const volcanoGeometry = useMemo(() => new THREE.CircleGeometry(0.15, 32), []);

  const nuclearZoneWarmupMaterial = useMemo(() => {
    return createNuclearZoneMaterial(42);
  }, []);

  const nuclearZoneGeometry = useMemo(() => new THREE.CircleGeometry(0.14, 32), []);

  const worldTreeWarmupMaterial = useMemo(() => {
    return createWorldTreeMaterial(42);
  }, []);

  const worldTreeGeometry = useMemo(() => new THREE.CircleGeometry(0.14, 32), []);

  const volcanoSmokeWarmupMaterial = useMemo(() => {
    return new THREE.SpriteMaterial({
      map: smokeTexture,
      transparent: true,
      opacity: 0,
      depthWrite: false,
      blending: THREE.NormalBlending,
      color: new THREE.Color(0.03, 0.025, 0.02),
    });
  }, [smokeTexture]);

  // --- Solar system warmup materials ---
  const solarSphereGeometry = useMemo(() => new THREE.SphereGeometry(WARMUP_SCALE, 8, 4), []);

  const sunCoronaMaterial = useMemo(
    () =>
      new THREE.ShaderMaterial({
        uniforms: {
          viewVector: { value: new THREE.Vector3(0, 0, 1) },
          glowColor: { value: new THREE.Color(1.0, 0.6, 0.2) },
          glowPower: { value: 3.0 },
        },
        vertexShader: sunCoronaVert,
        fragmentShader: sunCoronaFrag,
        side: THREE.BackSide,
        blending: THREE.AdditiveBlending,
        transparent: true,
        depthWrite: false,
        fog: false,
      }),
    [],
  );

  const sunBasicMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        map: textures.sun,
        color: new THREE.Color(1, 1, 1),
        fog: false,
      }),
    [textures.sun],
  );

  const planetTextureKeys = [
    "venus",
    "earth",
    "jupiter",
    "mercury",
    "saturn",
    "neptune",
    "uranus",
    "ceres",
    "moon",
    "ganymede",
    "earthClouds",
  ] as const;

  const planetMaterials = useMemo(
    () =>
      planetTextureKeys.map(
        (key) =>
          new THREE.MeshStandardMaterial({
            map: (textures as unknown as Record<string, THREE.Texture>)[key],
            roughness: 0.8,
            metalness: 0.05,
            fog: false,
          }),
      ),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    planetTextureKeys.map((key) => (textures as unknown as Record<string, THREE.Texture>)[key]),
  );

  const orbitLineMaterial = useMemo(
    () => new THREE.LineBasicMaterial({ color: 0x405080, transparent: true, fog: false }),
    [],
  );

  const orbitLineObj = useMemo(() => {
    const points = [new THREE.Vector3(0, 0, 0), new THREE.Vector3(WARMUP_SCALE, 0, 0)];
    const geo = new THREE.BufferGeometry().setFromPoints(points);
    return new THREE.Line(geo, orbitLineMaterial);
  }, [orbitLineMaterial]);

  const treeRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const bushRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const cloverRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const flowerRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const rockRef = useRef<THREE.InstancedMesh>(null);

  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of treeRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
      mesh.setColorAt(0, new THREE.Color(1, 1, 1));
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
    }
  }, [treeVariants]);

  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of bushRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
      mesh.setColorAt(0, new THREE.Color(1, 1, 1));
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
    }
  }, [bushVariants]);

  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of cloverRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
    }
  }, [cloverVariants]);

  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of flowerRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
    }
  }, [flowerVariants]);

  useLayoutEffect(() => {
    if (!rockRef.current) return;
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    rockRef.current.setMatrixAt(0, matrix);
    rockRef.current.instanceMatrix.needsUpdate = true;
  }, [rockGeometry]);

  const renderVariants = (
    variants: TreeVariant[],
    prefix: string,
    refs: React.RefObject<Map<string, THREE.InstancedMesh | null>>,
    materialOverrides?: THREE.Material[][],
  ) =>
    variants.map((variant, vIdx) =>
      variant.primitives.map((prim, pIdx) => {
        const key = `${prefix}-${vIdx}-${pIdx}`;
        const mat = materialOverrides ? materialOverrides[vIdx][pIdx] : prim.material;
        return (
          <instancedMesh
            key={key}
            ref={(el) => {
              refs.current.set(key, el);
            }}
            args={[prim.geometry, mat, 1]}
            frustumCulled={false}
          />
        );
      }),
    );

  return (
    <group>
      {renderVariants(treeVariants, "warmup-tree", treeRefs)}
      {renderVariants(bushVariants, "warmup-bush", bushRefs)}
      {renderVariants(cloverVariants, "warmup-clover", cloverRefs)}
      {renderVariants(flowerVariants, "warmup-flower", flowerRefs)}
      <instancedMesh ref={rockRef} args={[rockGeometry, rockMaterial, 1]} frustumCulled={false} />
      <primitive object={warmupCity} />
      <mesh geometry={oceanGeometry} material={oceanWarmupMaterial} frustumCulled={false} />
      <mesh geometry={smokeGeometry} material={smokeWarmupMaterial} frustumCulled={false} />
      <mesh geometry={volcanoGeometry} material={volcanoWarmupMaterial} frustumCulled={false} />
      <mesh
        geometry={nuclearZoneGeometry}
        material={nuclearZoneWarmupMaterial}
        frustumCulled={false}
      />
      <mesh geometry={worldTreeGeometry} material={worldTreeWarmupMaterial} frustumCulled={false} />
      <sprite
        material={volcanoSmokeWarmupMaterial}
        scale={[WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE]}
      />

      {/* Solar system warmup */}
      <mesh geometry={solarSphereGeometry} material={sunBasicMaterial} frustumCulled={false} />
      <mesh geometry={solarSphereGeometry} material={sunCoronaMaterial} frustumCulled={false} />
      {planetMaterials.map((mat, i) => (
        <mesh
          key={`planet-warmup-${i}`}
          geometry={solarSphereGeometry}
          material={mat}
          frustumCulled={false}
        />
      ))}
      <primitive object={orbitLineObj} />
    </group>
  );
}
