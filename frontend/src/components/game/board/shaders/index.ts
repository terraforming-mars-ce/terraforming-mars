import * as THREE from "three";
import oceanVertexRaw from "./ocean.vert.glsl?raw";
import oceanFragmentRaw from "./ocean.frag.glsl?raw";
import sphereProjectionVertexRaw from "./sphere-projection.vert.glsl?raw";
import oceanBorderFragmentRaw from "./ocean-border.frag.glsl?raw";
import hoverGlowFragmentRaw from "./hover-glow.frag.glsl?raw";
import availableGlowFragmentRaw from "./available-glow.frag.glsl?raw";
import endgameHighlightFragmentRaw from "./endgame-highlight.frag.glsl?raw";
import tileBorderVertexRaw from "./tile-border.vert.glsl?raw";
import tileBorderFragmentRaw from "./tile-border.frag.glsl?raw";
import volcanoVertexRaw from "./volcano.vert.glsl?raw";
import volcanoFragmentRaw from "./volcano.frag.glsl?raw";

export { default as tileSurfaceVertexSnippet } from "./tile-surface.vert.glsl?raw";
export { default as greeneryGroundVertexSnippet } from "./greenery-ground.vert.glsl?raw";
export { default as greeneryGroundFragmentSnippet } from "./greenery-ground.frag.glsl?raw";

// Strip #version directive — Three.js prepends its own #version 300 es at runtime
function stripVersion(raw: string): string {
  return raw.replace(/^#version\s+\d+(\s+es)?\s*\n/, "");
}

export const oceanVertexShader = stripVersion(oceanVertexRaw);
export const oceanFragmentShader = stripVersion(oceanFragmentRaw);
export const sphereProjectionVertex = stripVersion(sphereProjectionVertexRaw);
export const oceanBorderFragment = stripVersion(oceanBorderFragmentRaw);
export const hoverGlowFragment = stripVersion(hoverGlowFragmentRaw);
export const availableGlowFragment = stripVersion(availableGlowFragmentRaw);
export const endgameHighlightFragment = stripVersion(endgameHighlightFragmentRaw);
export const tileBorderVertex = stripVersion(tileBorderVertexRaw);
export const tileBorderFragment = stripVersion(tileBorderFragmentRaw);
export const volcanoVertex = stripVersion(volcanoVertexRaw);
export const volcanoFragment = stripVersion(volcanoFragmentRaw);

export function splitSnippet(raw: string): { header: string; body: string } {
  const marker = "//#pragma body\n";
  const idx = raw.indexOf(marker);
  if (idx === -1) return { header: "", body: raw };
  return {
    header: raw.slice(0, idx).trim(),
    body: raw.slice(idx + marker.length).trim(),
  };
}

export function createOceanMaterial(
  waterNormals: THREE.Texture,
  sandTexture: THREE.Texture,
  overrides?: Record<string, { value: unknown }>,
): THREE.ShaderMaterial {
  const defaults: Record<string, { value: unknown }> = {
    uSphereRadius: { value: 2.02 },
    time: { value: 0.0 },
    size: { value: 250.0 },
    alpha: { value: 1.0 },
    rf0: { value: 0.1 },
    sunIntensity: { value: 1.0 },
    normalSampler: { value: waterNormals },
    sunColor: { value: new THREE.Vector3(1.0, 1.0, 1.0) },
    sunDirection: { value: new THREE.Vector3(0.9, 0.0, 0.8).normalize() },
    eye: { value: new THREE.Vector3() },
    waterColor: { value: new THREE.Vector3(0.01, 0.03, 0.03) },
    uRadius: { value: 0.5 },
    uAspect: { value: 1.0 },
    uRotation: { value: 0 },
    uEdgeBand: { value: 0.08 },
    uEdgeStrength: { value: 0.11 },
    uEdgeScale: { value: 3.5 },
    uWarpScale: { value: 2.0 },
    uWarpAmount: { value: 0.07 },
    uSandWidth: { value: 0.8 },
    uGrainScale: { value: 18.0 },
    sandSampler: { value: sandTexture },
    uSandTexScale: { value: 3.0 },
    uShallowWidth: { value: 0.22 },
    uShallowStrength: { value: 0.55 },
    uEdgeSoftness: { value: 0.03 },
    uSeedOffset: { value: new THREE.Vector2(0, 0) },
    uFoamWidth: { value: 0.08 },
    uFoamStrength: { value: 0.7 },
    uFoamScale: { value: 8.0 },
    uFoamSpeed: { value: 0.08 },
    uFoamCutoff: { value: 0.52 },
    uFoamPulseSpeed: { value: 0.9 },
    uFoamPulseAmount: { value: 0.5 },
  };

  const merged = { ...defaults, ...overrides };

  return new THREE.ShaderMaterial({
    vertexShader: oceanVertexShader,
    fragmentShader: oceanFragmentShader,
    uniforms: THREE.UniformsUtils.merge([THREE.UniformsLib.lights, merged]),
    lights: true,
    transparent: true,
    premultipliedAlpha: true,
    side: THREE.DoubleSide,
    depthWrite: false,
  });
}

export function createVolcanoMaterial(
  grassTexture: THREE.Texture,
  flowTexture: THREE.Texture,
  seed: number,
): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    vertexShader: volcanoVertex,
    fragmentShader: volcanoFragment,
    uniforms: {
      uSphereRadius: { value: 2.02 },
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
