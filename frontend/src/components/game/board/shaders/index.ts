import * as THREE from "three";
import { SPHERE_RADIUS } from "../boardConstants";
import oceanVertexRaw from "./ocean.vert.glsl?raw";
import oceanFragmentRaw from "./ocean.frag.glsl?raw";
import sphereProjectionVertexRaw from "./sphere-projection.vert.glsl?raw";
import oceanBorderFragmentRaw from "./ocean-border.frag.glsl?raw";
import hoverGlowFragmentRaw from "./hover-glow.frag.glsl?raw";
import availableGlowFragmentRaw from "./available-glow.frag.glsl?raw";
import vpHighlightFragmentRaw from "./vp-highlight.frag.glsl?raw";

import tileBorderVertexRaw from "./tile-border.vert.glsl?raw";
import tileBorderFragmentRaw from "./tile-border.frag.glsl?raw";
import volcanoVertexRaw from "./volcano.vert.glsl?raw";
import volcanoFragmentRaw from "./volcano.frag.glsl?raw";
import nuclearZoneVertexRaw from "./nuclear-zone.vert.glsl?raw";
import nuclearZoneFragmentRaw from "./nuclear-zone.frag.glsl?raw";
import worldTreeVertexRaw from "./world-tree.vert.glsl?raw";
import worldTreeFragmentRaw from "./world-tree.frag.glsl?raw";
import moholeVertexRaw from "./mohole.vert.glsl?raw";
import moholeFragmentRaw from "./mohole.frag.glsl?raw";
import moholeMaskFragmentRaw from "./mohole-mask.frag.glsl?raw";

export { default as tileSurfaceVertexSnippet } from "./tile-surface.vert.glsl?raw";
export { default as greeneryGroundVertexSnippet } from "./greenery-ground.vert.glsl?raw";
export { default as greeneryGroundFragmentSnippet } from "./greenery-ground.frag.glsl?raw";

// Strip #version directive — Three.js prepends its own #version 300 es at runtime
function stripVersion(raw: string): string {
  return raw.replace(/^#version\s+\d+(\s+es)?\s*\n/, "");
}

export const sphereProjectionVertex = stripVersion(sphereProjectionVertexRaw);
export const oceanBorderFragment = stripVersion(oceanBorderFragmentRaw);
export const hoverGlowFragment = stripVersion(hoverGlowFragmentRaw);
export const availableGlowFragment = stripVersion(availableGlowFragmentRaw);
export const vpHighlightFragment = stripVersion(vpHighlightFragmentRaw);

export const tileBorderVertex = stripVersion(tileBorderVertexRaw);
export const tileBorderFragment = stripVersion(tileBorderFragmentRaw);
export const volcanoVertex = stripVersion(volcanoVertexRaw);
export const volcanoFragment = stripVersion(volcanoFragmentRaw);
export const nuclearZoneVertex = stripVersion(nuclearZoneVertexRaw);
export const nuclearZoneFragment = stripVersion(nuclearZoneFragmentRaw);
export const worldTreeVertex = stripVersion(worldTreeVertexRaw);
export const worldTreeFragment = stripVersion(worldTreeFragmentRaw);
export const moholeVertex = stripVersion(moholeVertexRaw);
export const moholeFragment = stripVersion(moholeFragmentRaw);
export const moholeMaskFragment = stripVersion(moholeMaskFragmentRaw);

export function splitSnippet(raw: string): { header: string; body: string } {
  const marker = "//#pragma body\n";
  const idx = raw.indexOf(marker);
  if (idx === -1) return { header: "", body: raw };
  return {
    header: raw.slice(0, idx).trim(),
    body: raw.slice(idx + marker.length).trim(),
  };
}

export { default as oceanVertexSnippet } from "./ocean.vert.glsl?raw";
export { default as oceanFragmentSnippet } from "./ocean.frag.glsl?raw";

export interface OceanUniformOverrides {
  uSphereCenter?: THREE.Vector3;
  uRadius?: number;
  uAspect?: number;
  uRotation?: number;
  uEdgeScale?: number;
  uSeedOffset?: THREE.Vector2;
}

export function addOceanProjection(
  material: THREE.Material,
  waterNormals: THREE.Texture,
  sandTexture: THREE.Texture,
  sphereCenter: THREE.Vector3,
  zOffset: number,
  overrides?: OceanUniformOverrides,
): void {
  const vertSnippet = splitSnippet(oceanVertexRaw);
  const fragSnippet = splitSnippet(oceanFragmentRaw);

  material.onBeforeCompile = (shader) => {
    (material as any).__shader = shader;

    // Vertex uniforms
    shader.uniforms.uSphereRadius = { value: SPHERE_RADIUS };
    shader.uniforms.uZOffset = { value: zOffset };
    shader.uniforms.uSphereCenter = { value: sphereCenter };

    // Fragment uniforms
    shader.uniforms.time = { value: 0.0 };
    shader.uniforms.oceanSize = { value: 250.0 };
    shader.uniforms.oceanAlpha = { value: 1.0 };
    shader.uniforms.rf0 = { value: 0.1 };
    shader.uniforms.sunIntensity = { value: 1.0 };
    shader.uniforms.normalSampler = { value: waterNormals };
    shader.uniforms.sunColor = { value: new THREE.Vector3(1.0, 1.0, 1.0) };
    shader.uniforms.sunDirection = { value: new THREE.Vector3(0.9, 0.0, 0.8).normalize() };
    shader.uniforms.eye = { value: new THREE.Vector3() };
    shader.uniforms.waterColor = { value: new THREE.Vector3(0.01, 0.03, 0.03) };
    shader.uniforms.uRadius = { value: overrides?.uRadius ?? 0.5 };
    shader.uniforms.uAspect = { value: overrides?.uAspect ?? 1.0 };
    shader.uniforms.uRotation = { value: overrides?.uRotation ?? 0 };
    shader.uniforms.uEdgeBand = { value: 0.08 };
    shader.uniforms.uEdgeStrength = { value: 0.11 };
    shader.uniforms.uEdgeScale = { value: overrides?.uEdgeScale ?? 3.5 };
    shader.uniforms.uWarpScale = { value: 2.0 };
    shader.uniforms.uWarpAmount = { value: 0.07 };
    shader.uniforms.uSandWidth = { value: 0.8 };
    shader.uniforms.uGrainScale = { value: 18.0 };
    shader.uniforms.sandSampler = { value: sandTexture };
    shader.uniforms.uSandTexScale = { value: 3.0 };
    shader.uniforms.uShallowWidth = { value: 0.22 };
    shader.uniforms.uShallowStrength = { value: 0.55 };
    shader.uniforms.uEdgeSoftness = { value: 0.03 };
    shader.uniforms.uSeedOffset = { value: overrides?.uSeedOffset ?? new THREE.Vector2(0, 0) };
    shader.uniforms.uFoamWidth = { value: 0.08 };
    shader.uniforms.uFoamStrength = { value: 0.7 };
    shader.uniforms.uFoamScale = { value: 8.0 };
    shader.uniforms.uFoamSpeed = { value: 0.08 };
    shader.uniforms.uFoamCutoff = { value: 0.52 };
    shader.uniforms.uFoamPulseSpeed = { value: 0.9 };
    shader.uniforms.uFoamPulseAmount = { value: 0.5 };

    // Vertex: inject sphere projection
    shader.vertexShader =
      vertSnippet.header +
      "\n" +
      shader.vertexShader.replace("#include <begin_vertex>", vertSnippet.body).replace(
        "#include <project_vertex>",
        `vec4 mvPosition = viewMatrix * vec4(projectedPos, 1.0);
           gl_Position = projectionMatrix * mvPosition;`,
      );

    // Fragment: inject ocean color computation
    shader.fragmentShader =
      fragSnippet.header +
      "\n" +
      shader.fragmentShader.replace("#include <opaque_fragment>", fragSnippet.body);
  };
}

export function createVolcanoMaterial(
  grassTexture: THREE.Texture,
  flowTexture: THREE.Texture,
  seed: number,
  sphereCenter?: THREE.Vector3,
): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    vertexShader: volcanoVertex,
    fragmentShader: volcanoFragment,
    uniforms: {
      uSphereRadius: { value: 2.02 },
      uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
      uHeight: { value: 0.14 },
      uCraterRadius: { value: 0.22 },
      uCraterDepth: { value: 0.08 },
      uEmergence: { value: 1.0 },
      uTime: { value: 0.0 },
      uSeed: { value: seed },
      uSunDirection: { value: new THREE.Vector3(0.9, 0.0, 0.8).normalize() },
      uSunIntensity: { value: 1.0 },
      uSunColor: { value: new THREE.Vector3(1.0, 0.86, 0.72) },
      uGrassTexture: { value: grassTexture },
      uFlowTex: { value: flowTexture },
      uDebugMode: { value: 0 },
    },
    transparent: true,
    depthWrite: true,
    depthTest: true,
    side: THREE.DoubleSide,
  });
}

export function createNuclearZoneMaterial(
  seed: number,
  sphereCenter?: THREE.Vector3,
): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    vertexShader: nuclearZoneVertex,
    fragmentShader: nuclearZoneFragment,
    uniforms: {
      uSphereRadius: { value: 2.02 },
      uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
      uCraterDepth: { value: 0.04 },
      uCraterRadius: { value: 0.35 },
      uEmergence: { value: 1.0 },
      uTime: { value: 0.0 },
      uSeed: { value: seed },
      uSunDirection: { value: new THREE.Vector3(0.9, 0.0, 0.8).normalize() },
      uSunIntensity: { value: 1.0 },
      uSunColor: { value: new THREE.Vector3(1.0, 0.86, 0.72) },
    },
    transparent: true,
    depthWrite: true,
    depthTest: true,
    side: THREE.DoubleSide,
  });
}

export function createWorldTreeMaterial(
  seed: number,
  sphereCenter?: THREE.Vector3,
): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    vertexShader: worldTreeVertex,
    fragmentShader: worldTreeFragment,
    uniforms: {
      uSphereRadius: { value: 2.02 },
      uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
      uTrunkHeight: { value: 0.1 },
      uEmergence: { value: 1.0 },
      uTime: { value: 0.0 },
      uSeed: { value: seed },
      uSunDirection: { value: new THREE.Vector3(0.9, 0.0, 0.8).normalize() },
      uSunIntensity: { value: 1.0 },
      uSunColor: { value: new THREE.Vector3(1.0, 0.86, 0.72) },
    },
    transparent: true,
    depthWrite: true,
    depthTest: true,
    side: THREE.DoubleSide,
  });
}

export function createMoholeMaterial(
  seed: number,
  concreteTexture?: THREE.Texture,
  sphereCenter?: THREE.Vector3,
): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    vertexShader: moholeVertex,
    fragmentShader: moholeFragment,
    uniforms: {
      uSphereRadius: { value: 2.02 },
      uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
      uHoleRadius: { value: 0.4 },
      uHoleDepth: { value: 0.06 },
      uEmergence: { value: 1.0 },
      uTime: { value: 0.0 },
      uSeed: { value: seed },
      uSunDirection: { value: new THREE.Vector3(0.9, 0.0, 0.8).normalize() },
      uSunIntensity: { value: 1.0 },
      uSunColor: { value: new THREE.Vector3(1.0, 0.86, 0.72) },
      uEmergenceRadius: { value: 1.0 },
      uConcreteTexture: { value: concreteTexture ?? null },
    },
    transparent: true,
    depthWrite: true,
    depthTest: true,
    side: THREE.DoubleSide,
  });
}

export function createMoholeMaskMaterial(
  seed: number,
  sphereCenter?: THREE.Vector3,
): THREE.ShaderMaterial {
  const mat = new THREE.ShaderMaterial({
    vertexShader: moholeVertex,
    fragmentShader: moholeMaskFragment,
    uniforms: {
      uSphereRadius: { value: 2.02 },
      uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
      uHoleRadius: { value: 0.4 },
      uHoleDepth: { value: 0.0 },
      uEmergence: { value: 1.0 },
      uEmergenceRadius: { value: 1.0 },
      uSeed: { value: seed },
    },
    transparent: false,
    colorWrite: false,
    depthWrite: false,
    side: THREE.DoubleSide,
  });

  mat.stencilWrite = true;
  mat.stencilRef = 1;
  mat.stencilFunc = THREE.AlwaysStencilFunc;
  mat.stencilZPass = THREE.ReplaceStencilOp;
  mat.stencilFail = THREE.KeepStencilOp;
  mat.stencilZFail = THREE.KeepStencilOp;

  return mat;
}
