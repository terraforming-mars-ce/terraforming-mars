import { memo, useMemo, useEffect, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { type LandscapePlan } from "./landscapeNetwork";
import {
  unpackLandscapeGround,
  createRoadRibbon,
  type LandscapeGroundData,
} from "./landscapeGeometry";
import { splitSnippet } from "./shaders";
import fragment from "./shaders/landscape-ground.frag.glsl?raw";
import roadVertex from "./shaders/landscape-road.vert.glsl?raw";
import roadFragment from "./shaders/landscape-road.frag.glsl?raw";
import shoulderVertex from "./shaders/road-shoulder.vert.glsl?raw";
import shoulderFragment from "./shaders/road-shoulder.frag.glsl?raw";
import vertex from "./shaders/landscape-ground.vert.glsl?raw";

function LandscapeRenderer({
  plan,
  groundData,
}: {
  plan: LandscapePlan;
  groundData: LandscapeGroundData;
}) {
  const textures = useTextures();
  const born = useRef<Map<string, number> | null>(null);
  const births = useMemo(() => {
    const first = born.current === null;
    born.current ??= new Map();
    const keys = new Set(plan.connections.map((c) => c.key));
    for (const key of born.current.keys()) {
      if (!keys.has(key)) {
        born.current.delete(key);
      }
    }
    return plan.connections.flatMap((c) => {
      if (!born.current!.has(c.key)) {
        born.current!.set(c.key, first ? -1000 : performance.now() / 1000);
      }
      return c.routes.map(() => born.current!.get(c.key)!);
    });
  }, [plan.connections]);
  const animationTime = useMemo(() => ({ value: performance.now() / 1000 }), []);
  useFrame(() => {
    animationTime.value = performance.now() / 1000;
  });

  const ground = useMemo(() => unpackLandscapeGround(groundData), [groundData]);
  const materials = useMemo(() => {
    const grass = new THREE.MeshStandardMaterial({
      map: textures.grass,
      color: "#819266",
      transparent: true,
      depthWrite: false,
      alphaTest: 0.005,
      roughness: 1,
      side: THREE.DoubleSide,
    });
    const vs = splitSnippet(vertex),
      fs = splitSnippet(fragment);
    grass.onBeforeCompile = (shader) => {
      Object.assign(shader.uniforms, {
        uNoiseMap: { value: textures.noiseMid },
        uNoiseMapHigh: { value: textures.noiseHigh },
      });
      shader.vertexShader =
        vs.header + "\n" + shader.vertexShader.replace("#include <begin_vertex>", vs.body);
      shader.fragmentShader =
        fs.header + "\n" + shader.fragmentShader.replace("#include <alphamap_fragment>", fs.body);
    };
    const curb = new THREE.MeshStandardMaterial({
      map: textures.concrete,
      color: "#a8a399",
      bumpMap: textures.concrete,
      bumpScale: 0.0003,
      emissive: "#b8b6a9",
      emissiveMap: textures.concrete,
      emissiveIntensity: 0.22,
      roughness: 0.9,
      side: THREE.DoubleSide,
    });
    const asphalt = new THREE.MeshStandardMaterial({
      map: textures.concrete,
      bumpMap: textures.concrete,
      bumpScale: 0.00025,
      color: "#697175",
      emissive: "#697175",
      emissiveMap: textures.concrete,
      emissiveIntensity: 0.25,
      roughness: 0.98,
      side: THREE.DoubleSide,
    });
    const path = new THREE.MeshStandardMaterial({
      map: textures.sand,
      color: "#a99b7c",
      roughness: 1,
      side: THREE.DoubleSide,
    });
    for (const material of [curb, asphalt, path]) {
      material.transparent = true;
      const vert = splitSnippet(roadVertex),
        frag = splitSnippet(roadFragment);
      material.onBeforeCompile = (shader) => {
        shader.uniforms.uConnectionTime = animationTime;
        shader.vertexShader =
          vert.header + "\n" + shader.vertexShader.replace("#include <begin_vertex>", vert.body);
        shader.fragmentShader =
          frag.header +
          "\n" +
          shader.fragmentShader.replace("#include <alphamap_fragment>", frag.body);
      };
    }
    const shoulder = new THREE.MeshStandardMaterial({
      color: "#74716b",
      transparent: true,
      opacity: 0.48,
      depthWrite: false,
      roughness: 1,
      side: THREE.DoubleSide,
    });
    const shoulderVs = splitSnippet(shoulderVertex),
      shoulderFs = splitSnippet(shoulderFragment);
    shoulder.onBeforeCompile = (shader) => {
      shader.uniforms.uConnectionTime = animationTime;
      shader.uniforms.uNoiseMap = { value: textures.noiseMid };
      shader.uniforms.uNoiseMapHigh = { value: textures.noiseHigh };
      shader.vertexShader =
        shoulderVs.header +
        "\n" +
        shader.vertexShader.replace("#include <begin_vertex>", shoulderVs.body);
      shader.fragmentShader =
        shoulderFs.header +
        "\n" +
        shader.fragmentShader.replace("#include <alphamap_fragment>", shoulderFs.body);
    };
    return [grass, shoulder, curb, asphalt, path];
  }, [
    textures.grass,
    textures.noiseMid,
    textures.noiseHigh,
    textures.concrete,
    textures.sand,
    animationTime,
  ]);
  const roads = useMemo(() => {
    const shoulder = createRoadRibbon(plan.roads, 0.013, 0.00045, births);
    const edges = new Float32Array(shoulder.getAttribute("position").count);
    for (let i = 0; i < edges.length; i++) {
      edges[i] = [-1, -1, 1, -1, 1, 1][i % 6];
    }
    shoulder.setAttribute("roadEdge", new THREE.BufferAttribute(edges, 1));
    return [
      shoulder,
      createRoadRibbon(
        plan.roads.filter((r) => r.kind === "road"),
        0.0012,
        0.0006,
        births.filter((_, i) => plan.roads[i].kind === "road"),
      ),
      createRoadRibbon(
        plan.roads.filter((r) => r.kind === "road"),
        0,
        0.001,
        births.filter((_, i) => plan.roads[i].kind === "road"),
      ),
      createRoadRibbon(
        plan.roads.filter((r) => r.kind === "path"),
        0,
        0.001,
        births.filter((_, i) => plan.roads[i].kind === "path"),
      ),
    ];
  }, [plan.roads, births]);
  useEffect(() => () => ground.dispose(), [ground]);
  useEffect(() => () => roads.forEach((g) => g.dispose()), [roads]);
  useEffect(() => () => materials.forEach((m) => m.dispose()), [materials]);
  return (
    <group dispose={null}>
      {Boolean(groundData.index) && (
        <mesh geometry={ground} material={materials[0]} renderOrder={11} raycast={() => {}} />
      )}
      {roads.map((geometry, i) => (
        <mesh
          key={i}
          geometry={geometry}
          material={materials[i + 1]}
          renderOrder={i === 0 ? 12 : 13}
          raycast={() => {}}
        />
      ))}
    </group>
  );
}

export default memo(LandscapeRenderer);
