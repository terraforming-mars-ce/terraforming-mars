import { tileHighlightSource } from "../tileHighlightMaterials";
import collapseVertex from "./nuclear-collapse.vert.glsl?raw";
import * as THREE from "three";
import { MOHOLE_STENCIL_BIT, NUCLEAR_STENCIL_BIT, SPHERE_RADIUS } from "../boardConstants";
import {
  nuclearProfile,
  NUCLEAR_WAVE_SECONDS,
  NUCLEAR_WAVE_DUST_SECONDS,
  NUCLEAR_WAVE_RADIUS,
} from "../nuclearGeometry";
import type { AtmosphereProfile } from "../solarSystemConfig";
import planetHazeFragment from "./planet-haze.frag.glsl?raw";
import planetAtmosphereVertex from "./planet-atmosphere.vert.glsl?raw";
import planetAtmosphereFragment from "./planet-atmosphere.frag.glsl?raw";
import sunSurfaceVertex from "./sun-surface.vert.glsl?raw";
import sunSurfaceFragment from "./sun-surface.frag.glsl?raw";
import sunCoronaVertex from "./sun-corona.vert.glsl?raw";
import sunCoronaFragment from "./sun-corona.frag.glsl?raw";
import sunProminenceFragment from "./sun-prominence.frag.glsl?raw";
import sphereProjectionVertexRaw from "./sphere-projection.vert.glsl?raw";
import oceanBorderFragmentRaw from "./ocean-border.frag.glsl?raw";

import tileBorderVertexRaw from "./tile-border.vert.glsl?raw";
import tileBorderFragmentRaw from "./tile-border.frag.glsl?raw";
import volcanoVertexRaw from "./volcano.vert.glsl?raw";
import volcanoFragmentRaw from "./volcano.frag.glsl?raw";
import nuclearMaskFragment from "./nuclear-mask.frag.glsl?raw";
import nuclearWaveFragment from "./nuclear-wave.frag.glsl?raw";
import nuclearZoneVertexRaw from "./nuclear-zone.vert.glsl?raw";
import nuclearSurfaceVertex from "./nuclear-surface.vert.glsl?raw";
import nuclearZoneFragmentRaw from "./nuclear-zone.frag.glsl?raw";
import worldTreeVertexRaw from "./world-tree.vert.glsl?raw";
import worldTreeFragmentRaw from "./world-tree.frag.glsl?raw";
import moholeVertexRaw from "./mohole.vert.glsl?raw";
import moholeFragmentRaw from "./mohole.frag.glsl?raw";
import moholeMaskFragmentRaw from "./mohole-mask.frag.glsl?raw";
import marsClimateFragment from "./mars-climate.frag.glsl?raw";
import marsSurface from "./mars-surface.glsl?raw";

export { default as tileSurfaceVertexSnippet } from "./tile-surface.vert.glsl?raw";
export { default as greeneryGroundVertexSnippet } from "./greenery-ground.vert.glsl?raw";
export { default as greeneryGroundFragmentSnippet } from "./greenery-ground.frag.glsl?raw";

// Strip #version directive — Three.js prepends its own #version 300 es at runtime
function stripVersion(raw: string): string {
  return raw.replace(/^#version\s+\d+(\s+es)?\s*\n/, "");
}

export const sphereProjectionVertex = stripVersion(sphereProjectionVertexRaw);
export const oceanBorderFragment = stripVersion(oceanBorderFragmentRaw);

export const tileBorderVertex = stripVersion(tileBorderVertexRaw);
export const tileBorderFragment = stripVersion(tileBorderFragmentRaw);
export const volcanoVertex = stripVersion(volcanoVertexRaw);
export const volcanoFragment = stripVersion(volcanoFragmentRaw);
export const nuclearZoneVertex = stripVersion(nuclearZoneVertexRaw);
export const worldTreeVertex = stripVersion(worldTreeVertexRaw);
export const worldTreeFragment = stripVersion(worldTreeFragmentRaw);
export const moholeVertex = stripVersion(moholeVertexRaw);
export const moholeFragment = stripVersion(moholeFragmentRaw);
export const moholeMaskFragment = stripVersion(moholeMaskFragmentRaw);

export const MAX_ATMOSPHERES = 16;

export function createSunSurfaceMaterial(texture: THREE.Texture) {
  return new THREE.ShaderMaterial({
    uniforms: { uSurface: { value: texture }, uTime: { value: 0 } },
    vertexShader: sunSurfaceVertex,
    fragmentShader: sunSurfaceFragment,
    fog: false,
    userData: { planetHaze: false },
  });
}

export function createSunCoronaMaterial() {
  return new THREE.ShaderMaterial({
    uniforms: { uTime: { value: 0 } },
    vertexShader: sunCoronaVertex,
    fragmentShader: sunCoronaFragment,
    side: THREE.BackSide,
    blending: THREE.AdditiveBlending,
    transparent: true,
    depthWrite: false,
    fog: false,
  });
}

export function createSunProminenceMaterial() {
  return new THREE.ShaderMaterial({
    uniforms: { uTime: { value: 0 } },
    vertexShader: sunSurfaceVertex,
    fragmentShader: sunProminenceFragment,
    blending: THREE.AdditiveBlending,
    transparent: true,
    depthWrite: false,
    fog: false,
  });
}

export function createPlanetAtmosphereMaterial(profile: AtmosphereProfile) {
  return new THREE.ShaderMaterial({
    vertexShader: planetAtmosphereVertex,
    fragmentShader: planetAtmosphereFragment,
    uniforms: {
      uColor: { value: new THREE.Color(profile.color) },
      uShadowColor: { value: new THREE.Color(profile.shadowColor) },
      uSunColor: { value: new THREE.Color() },
      uSunDirection: { value: new THREE.Vector3(0, 0, 1) },
      uIntensity: { value: 0 },
      uThickness: { value: profile.thickness },
      uFollowsMesh: { value: 0 },
    },
    side: THREE.BackSide,
    transparent: true,
    blending: THREE.AdditiveBlending,
    depthTest: true,
    depthWrite: false,
    fog: false,
  });
}

export function createMarsClimateUniforms(noise: THREE.Texture) {
  return {
    uClimateNoise: { value: noise },
    uChill: { value: 0 },
    uSweep: { value: 0 },
    uSweepFade: { value: 0 },
  };
}

export function addMarsClimate(
  material: THREE.MeshStandardMaterial,
  uniforms: ReturnType<typeof createMarsClimateUniforms>,
) {
  const snippet = splitSnippet(marsClimateFragment);
  material.onBeforeCompile = (shader) => {
    Object.assign(shader.uniforms, uniforms);
    shader.fragmentShader =
      marsSurface +
      "\n" +
      snippet.header +
      "\n" +
      shader.fragmentShader.replace(
        "#include <map_fragment>",
        "#include <map_fragment>\n" + snippet.body,
      );
  };
  material.customProgramCacheKey = () => marsSurface + marsClimateFragment;
}

export function createPlanetHazeUniforms() {
  return {
    uHazeCameraWorld: { value: new THREE.Matrix4() },
    uHazeProjectionInverse: { value: new THREE.Matrix4() },
    uHazeSunPosition: { value: new THREE.Vector3() },
    uHazeSunColor: { value: new THREE.Color() },
    uHazeCount: { value: 0 },
    // Prepack once per frame instead of flattening vector arrays for every material upload.
    uHazeBodies: { value: new Float32Array(MAX_ATMOSPHERES * 4) },
    uHazeColors: { value: new Float32Array(MAX_ATMOSPHERES * 3) },
    uHazeShadowColors: { value: new Float32Array(MAX_ATMOSPHERES * 3) },
    uHazeProfiles: { value: new Float32Array(MAX_ATMOSPHERES * 3) },
  };
}

export function receivesPlanetHaze(material: THREE.Material) {
  return (
    material.userData.planetHaze !== false &&
    (material instanceof THREE.MeshStandardMaterial ||
      (material instanceof THREE.MeshLambertMaterial && material.userData.planetHaze === true) ||
      (material instanceof THREE.ShaderMaterial &&
        !(material instanceof THREE.RawShaderMaterial) &&
        (material.depthWrite || material.userData.planetHaze === true)))
  );
}

export function addPlanetHaze(
  material: THREE.Material,
  uniforms: ReturnType<typeof createPlanetHazeUniforms>,
) {
  // Three reuses cached uniform objects even after needsUpdate. Rebinding after
  // hot reload must release those programs so they cannot retain an old camera.
  material.dispose();
  const compile = material.onBeforeCompile;
  const programKey = material.customProgramCacheKey;
  const baseKey = material.customProgramCacheKey();
  const linearColor =
    material instanceof THREE.MeshStandardMaterial || material instanceof THREE.MeshLambertMaterial;
  material.onBeforeCompile = (shader, renderer) => {
    compile.call(material, shader, renderer);
    Object.assign(shader.uniforms, uniforms);
    const position = linearColor ? "-vViewPosition" : "(uHazeProjectionInverse * gl_Position).xyz";
    shader.vertexShader =
      "varying vec3 vHazeViewPosition;\nuniform mat4 uHazeProjectionInverse;\n" +
      shader.vertexShader.replace(/void\s+main\s*\(\s*\)/, "void atmosphereSourceVertex()") +
      `\nvoid main() { atmosphereSourceVertex(); vHazeViewPosition = ${position}; }`;
    shader.fragmentShader =
      `#define MAX_ATMOSPHERES ${MAX_ATMOSPHERES}\n` +
      planetHazeFragment +
      "\n" +
      shader.fragmentShader;
    if (linearColor) {
      shader.fragmentShader = shader.fragmentShader.replace(
        "#include <opaque_fragment>",
        "outgoingLight = applyPlanetHaze(outgoingLight, true, inverseTransformDirection(normal, viewMatrix));\n#include <opaque_fragment>",
      );
    } else {
      // Custom board shaders already output display colors; preserve that pipeline.
      const output = shader.fragmentShader.match(/out\s+vec4\s+(\w+)\s*;/)?.[1] ?? "gl_FragColor";
      shader.fragmentShader = shader.fragmentShader.replace(
        /void\s+main\s*\(\s*\)/,
        "void atmosphereSourceMain()",
      );
      shader.fragmentShader += `\nvoid main() { atmosphereSourceMain(); ${output}.rgb = applyPlanetHaze(${output}.rgb, false, vec3(0.0)); }`;
    }
  };
  const hazeKey = baseKey + planetHazeFragment + linearColor;
  material.customProgramCacheKey = () => hazeKey;
  material.needsUpdate = true;
  return () => {
    material.onBeforeCompile = compile;
    material.customProgramCacheKey = programKey;
    material.needsUpdate = true;
  };
}

export function splitSnippet(raw: string): { header: string; body: string } {
  const marker = "//#pragma body\n";
  const idx = raw.indexOf(marker);
  if (idx === -1) return { header: "", body: raw };
  return {
    header: raw.slice(0, idx).trim(),
    body: raw.slice(idx + marker.length).trim(),
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
      uVegetation: { value: 1 },
      uFrost: { value: 0 },
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
    forceSinglePass: true,
  });
}

export function createNuclearMaskMaterial(sphereCenter = new THREE.Vector3()) {
  return new THREE.ShaderMaterial({
    vertexShader: nuclearZoneVertex,
    fragmentShader: nuclearMaskFragment,
    uniforms: { uSphereRadius: { value: SPHERE_RADIUS }, uSphereCenter: { value: sphereCenter } },
    colorWrite: false,
    depthWrite: false,
    side: THREE.FrontSide,
    stencilWrite: true,
    stencilRef: NUCLEAR_STENCIL_BIT,
    stencilWriteMask: NUCLEAR_STENCIL_BIT,
    stencilFunc: THREE.AlwaysStencilFunc,
    stencilZPass: THREE.ReplaceStencilOp,
    userData: { planetHaze: false },
  });
}

export function createNuclearWaveMaterial(sphereCenter = new THREE.Vector3()) {
  return new THREE.ShaderMaterial({
    vertexShader: nuclearZoneVertex,
    fragmentShader: nuclearWaveFragment,
    uniforms: {
      uAge: { value: 0 },
      uDuration: { value: NUCLEAR_WAVE_SECONDS },
      uDustDuration: { value: NUCLEAR_WAVE_DUST_SECONDS },
      uRadius: { value: NUCLEAR_WAVE_RADIUS },
      uSphereRadius: { value: SPHERE_RADIUS },
      uSphereCenter: { value: sphereCenter },
    },
    transparent: true,
    depthWrite: false,
    side: THREE.DoubleSide,
    forceSinglePass: true,
    userData: { planetHaze: false },
  });
}

export function createNuclearZoneMaterial(
  seed: number,
  mars: THREE.Texture,
  sphereCenter = new THREE.Vector3(),
  scorch = false,
) {
  const uniforms = {
    uSphereRadius: { value: SPHERE_RADIUS },
    uSphereCenter: { value: sphereCenter },
    uSeed: { value: seed },
    uMars: { value: mars },
    uChill: { value: 0 },
    uGouges: { value: nuclearProfile(seed).gouges },
    uScorch: { value: Number(scorch) },
  };
  const material = new THREE.MeshStandardMaterial({
    roughness: 0.8,
    metalness: 0,
    emissive: 0,
    fog: false,
    flatShading: true,
    userData: { tileHighlights: true },
    transparent: scorch,
    depthWrite: !scorch,
    side: THREE.DoubleSide,
    forceSinglePass: true,
  });
  const vs = splitSnippet(nuclearSurfaceVertex);
  const fs = splitSnippet(nuclearZoneFragmentRaw);
  material.onBeforeCompile = (shader) => {
    Object.assign(shader.uniforms, uniforms);
    shader.vertexShader =
      vs.header + "\n" + shader.vertexShader.replace("#include <begin_vertex>", vs.body);
    shader.fragmentShader =
      marsSurface +
      "\n" +
      fs.header +
      "\n" +
      shader.fragmentShader.replace("#include <map_fragment>", fs.body);
  };
  material.customProgramCacheKey = () =>
    nuclearSurfaceVertex + marsSurface + nuclearZoneFragmentRaw;
  return { material, uniforms };
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
    forceSinglePass: true,
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
    forceSinglePass: true,
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
  mat.stencilRef = MOHOLE_STENCIL_BIT;
  mat.stencilWriteMask = MOHOLE_STENCIL_BIT;
  mat.stencilFunc = THREE.AlwaysStencilFunc;
  mat.stencilZPass = THREE.ReplaceStencilOp;
  mat.stencilFail = THREE.KeepStencilOp;
  mat.stencilZFail = THREE.KeepStencilOp;

  return mat;
}

export function createNuclearCollapseMaterial(
  source: THREE.Material,
  transform: { value: THREE.Matrix4 },
  viewRotation: { value: THREE.Matrix3 },
  worldRotation: { value: THREE.Matrix3 },
) {
  const material = source.clone();
  // The source compiler already carries any planet-haze wrapper. Do not apply it twice.
  material.userData.planetHaze = false;
  if (source instanceof THREE.ShaderMaterial && material instanceof THREE.ShaderMaterial) {
    material.uniforms = { ...source.uniforms };
  }
  const highlightSource = tileHighlightSource(source);
  const compile = highlightSource.compile;
  material.onBeforeCompile = (
    shader: THREE.WebGLProgramParametersWithUniforms,
    renderer: THREE.WebGLRenderer,
  ) => {
    compile.call(material, shader, renderer);
    shader.uniforms.uNuclearClipTransform = transform;
    shader.uniforms.uNuclearViewRotation = viewRotation;
    shader.uniforms.uNuclearWorldRotation = worldRotation;
    let defines = "";
    if (shader.vertexShader.includes("#include <normal_pars_vertex>")) {
      defines += "#define NUCLEAR_VIEW_NORMAL\n";
    }
    if (/varying\s+vec3\s+vWorldNormal/.test(shader.vertexShader)) {
      defines += "#define NUCLEAR_WORLD_NORMAL\n";
    }
    shader.vertexShader =
      shader.vertexShader.replace(/void\s+main\s*\(\s*\)/, "void nuclearOriginalMain()") +
      "\n" +
      defines +
      collapseVertex;
  };
  material.customProgramCacheKey = () => `${highlightSource.key.call(source)}:nuclear-collapse`;
  return material;
}
