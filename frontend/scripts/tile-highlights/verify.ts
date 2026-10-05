import assert from "node:assert/strict";
import { dirname, resolve } from "node:path";
import { plugin } from "bun";
import * as THREE from "three";
import {
  TileHighlightBoard,
  highlightCoordinate,
} from "../../src/components/game/board/tileHighlights";
import { boardCenter, projectBoardPoint } from "../../src/components/game/board/landscapeGeometry";

await plugin({
  name: "highlight-raw",
  setup(build) {
    build.onResolve({ filter: /\.glsl\?raw$/ }, (args) => ({
      path: resolve(dirname(args.importer), args.path.slice(0, -4)),
      namespace: "landscape-glsl",
    }));
    build.onLoad({ filter: /.*/, namespace: "landscape-glsl" }, async (args) => ({
      contents: `export default ${JSON.stringify(await Bun.file(args.path).text())}`,
      loader: "js",
    }));
  },
});
const { TileHighlightSystem, addTileHighlight, receivesTileHighlight } =
  await import("../../src/components/game/board/tileHighlightMaterials");
const shaders = await import("../../src/components/game/board/shaders");
const { LandscapeSurface } = await import("../../src/components/game/board/landscapeSurface");
const camera = new THREE.PerspectiveCamera(45, 1, 0.1, 5000);
camera.position.set(0, 0, 5);
camera.updateMatrixWorld();
const board = new TileHighlightBoard(2.02, -1, { q: 0, r: 0 });
board.root = new THREE.Group();
const normal = new THREE.Vector3(0, 0, 1);
const tile = {
  coordinate: { q: 0, r: 0, s: 0 },
  spherePosition: normal.clone().multiplyScalar(2.02),
  normal,
};
const state = { hovered: true, available: false, vpIntensity: 0, vpColor: [0.95, 0.95, 1] };
const record = board.register(tile, state);
for (let frame = 0; frame < 120; frame++) {
  board.tick(1 / 60, frame / 60, camera);
}
assert.equal(record.hover, 0.3);
assert.equal(board.uniforms.uTileHighlightActive.value, 1);
let version = board.texture.version;

for (let frame = 0; frame < 120; frame++) {
  board.tick(1 / 60, frame / 60, camera);
}
assert.equal(board.texture.version, version, "Settled hovering does not upload textures");
board.updateTile(record, {
  ...tile,
  spherePosition: tile.spherePosition.clone(),
  normal: normal.clone(),
});
board.tick(1 / 60, 4, camera);
assert.equal(
  board.texture.version,
  version,
  "Equivalent board updates preserve the fade and atlas",
);
state.available = true;
state.vpIntensity = 0.5;
for (let frame = 0; frame < 240; frame++) {
  board.tick(1 / 60, frame / 60, camera);
}
assert.equal(record.vp, 0.5);
version = board.texture.version;
board.tick(1 / 60, 9, camera);
assert.equal(board.texture.version, version, "Placement pulses use time, not texture uploads");
record.visibility = 0;
board.tick(1 / 60, 10, camera);
assert.equal(board.uniforms.uTileHighlightActive.value, 0);
record.visibility = 1;
board.tick(1 / 60, 10, camera, false);
assert.equal(
  state.hovered,
  false,
  "Dragging/panning clears hover rather than reviving stale state",
);
state.available = false;
state.vpIntensity = 0;
for (let frame = 0; frame < 360; frame++) {
  board.tick(1 / 60, frame / 60, camera);
}
assert.equal(board.uniforms.uTileHighlightActive.value, 0);
assert.equal(record.hover, 0);
assert.equal(record.vp, 0);
for (const fps of [30, 60, 120]) {
  const b = new TileHighlightBoard(2.02, -1, { q: 0, r: 0 });
  const entry = b.register(tile, { ...state, hovered: true });
  for (let i = 0; i < fps / 10; i++) {
    b.tick(1 / fps, i / fps, camera);
  }
  assert.ok(Math.abs(entry.hover - 0.3 * (1 - Math.exp(-0.975))) < 1e-8);
  b.dispose();
}
for (let q = -4; q <= 4; q++) {
  for (let r = -4; r <= 4; r++) {
    if (Math.abs(q + r) > 4) {
      continue;
    }
    for (const ySign of [-1, 1]) {
      const center = boardCenter({ q, r, s: -q - r });
      const axis = projectBoardPoint(center.x, center.y * -ySign).normalize();
      const basis = new THREE.Quaternion().setFromUnitVectors(normal, axis);
      for (let i = 0; i < 24; i++) {
        const angle = (i * Math.PI) / 12;
        const point = new THREE.Vector3(Math.cos(angle) * 0.14, Math.sin(angle) * 0.14, 0)
          .applyQuaternion(basis)
          .add(axis.clone().multiplyScalar(2.03))
          .normalize();
        assert.deepEqual(
          highlightCoordinate(point, 2.02, ySign),
          { q, r },
          "Real projected tile interiors select their own atlas cell",
        );
      }
    }
  }
}
const moon = new TileHighlightBoard(0.4, 1, { q: 100, r: 100 });
const moonRecord = moon.register(
  {
    ...tile,
    coordinate: { q: 100, r: 100, s: -200 },
    spherePosition: normal.clone().multiplyScalar(0.4),
  },
  { ...state, available: true },
);
moon.tick(1 / 60, 0, camera);
assert.equal(moonRecord.slot, 0);
assert.deepEqual(moon.uniforms.uTileHighlightGrid.value.toArray(), [0, 0, 1, 1]);
assert.notEqual(moon.texture, board.texture);
assert.equal((moon.texture.image.data as Float32Array)[1], 1);
assert.equal((board.texture.image.data as Float32Array)[1], 0);

const system = new TileHighlightSystem();
const scene = new THREE.Scene();
const rootA = new THREE.Group(),
  rootB = new THREE.Group();
board.root = rootA;
moon.root = rootB;
scene.add(rootA, rootB);
system.boards.add(board);
system.boards.add(moon);
const source = new THREE.MeshStandardMaterial();
const geometry = new THREE.BoxGeometry();
const meshA = new THREE.Mesh(geometry, source),
  meshB = new THREE.Mesh(geometry, source);
rootA.add(meshA);
rootB.add(meshB);
system.tick(1 / 60, 0, camera, true);
scene.traverseVisible(system.visit);
assert.equal(meshA.material, source);
assert.notEqual(meshB.material, source, "Shared source materials get independent board bindings");
assert.equal(
  source.customProgramCacheKey(),
  meshB.material.customProgramCacheKey(),
  "Board bindings share the shader program",
);
const fixtures: Array<{
  name: string;
  vertex: string;
  fragment: string;
  defines: Record<string, unknown>;
}> = [];
function compile(
  name: string,
  material: THREE.Material,
  library = "standard",
  defines: Record<string, unknown> = {},
) {
  const lib = THREE.ShaderLib[library];
  const shader = {
    vertexShader:
      material instanceof THREE.ShaderMaterial ? material.vertexShader : lib.vertexShader,
    fragmentShader:
      material instanceof THREE.ShaderMaterial ? material.fragmentShader : lib.fragmentShader,
    uniforms: {},
  } as THREE.WebGLProgramParametersWithUniforms;
  material.onBeforeCompile(shader, {} as THREE.WebGLRenderer);
  assert.equal((shader.vertexShader.match(/void main\s*\(/g) ?? []).length, 1);
  assert.equal((shader.fragmentShader.match(/void main\s*\(/g) ?? []).length, 1);
  assert.equal((shader.fragmentShader.match(/vec3 applyTileHighlight\(/g) ?? []).length, 1);
  fixtures.push({
    name,
    vertex: shader.vertexShader,
    fragment: shader.fragmentShader,
    defines: {
      ...((material as THREE.ShaderMaterial).defines ?? {}),
      ...((material as THREE.MeshStandardMaterial).flatShading ? { FLAT_SHADED: "" } : {}),
      ...(material.side === THREE.DoubleSide ? { DOUBLE_SIDED: "" } : {}),
      ...defines,
    },
  });
  return shader;
}
assert.equal(compile("standard", source).uniforms.uTileHighlightData.value, board.texture);
assert.equal(
  compile("moon-standard", meshB.material).uniforms.uTileHighlightData.value,
  moon.texture,
);
const stableMaterialVersion = source.version;
for (let frame = 0; frame < 20; frame++) {
  scene.traverseVisible(system.visit);
}
assert.equal(source.version, stableMaterialVersion, "Hover does not recompile materials");
assert.ok(stableMaterialVersion > 0);
const texture = new THREE.Texture();
const haze = shaders.createPlanetHazeUniforms();
const customMaterials = [
  ["crater", shaders.createNuclearZoneMaterial(42, texture).material],
  ["apron", shaders.createNuclearZoneMaterial(42, texture, new THREE.Vector3(), true).material],
  ["volcano", shaders.createVolcanoMaterial(texture, texture, 42)],
  ["world-tree", shaders.createWorldTreeMaterial(42)],
  ["mohole", shaders.createMoholeMaterial(42, texture)],
] as const;
for (const [name, material] of customMaterials) {
  assert.equal(receivesTileHighlight(material), true);
  if (shaders.receivesPlanetHaze(material)) {
    shaders.addPlanetHaze(material, haze);
  }
  addTileHighlight(material, board);
  compile(name, material);
  const collapse = shaders.createNuclearCollapseMaterial(
    material,
    { value: new THREE.Matrix4() },
    { value: new THREE.Matrix3() },
    { value: new THREE.Matrix3() },
  );
  addTileHighlight(collapse, board);
  const collapsed = compile(name + "-collapse", collapse);
  assert.ok(
    collapsed.vertexShader.indexOf("uNuclearClipTransform * gl_Position") <
      collapsed.vertexShader.lastIndexOf("world=uTileHighlightCameraWorld"),
    "Highlight position is captured after collapse",
  );
  collapse.dispose();
  material.dispose();
}
const array = new THREE.DataArrayTexture(new Uint8Array(4), 1, 1, 1);
const textures = new Proxy({}, { get: () => texture }) as ConstructorParameters<
  typeof LandscapeSurface
>[0];
const landscape = new LandscapeSurface(
  textures,
  { groundAlbedo: array, groundDetail: array, iceAlbedo: array, iceDetail: array },
  1,
);
for (const [name, material] of [
  ["ground", landscape.ground],
  ["ground-edge", landscape.groundEdge],
  ["basin", landscape.basin],
  ["water", landscape.water],
] as const) {
  shaders.addPlanetHaze(material, haze);
  addTileHighlight(material, board);
  compile(name, material, "standard", { USE_INSTANCING: "" });
}
const lambert = new THREE.MeshLambertMaterial({ map: texture, alphaTest: 0.5 });
lambert.onBeforeCompile = (shader) => {
  shader.vertexShader = "invariant gl_Position;\n" + shader.vertexShader;
};
addTileHighlight(lambert, board);
const foliage = compile("foliage", lambert, "lambert", {
  USE_INSTANCING: "",
  USE_MAP: "",
  USE_ALPHATEST: "",
  MAP_UV: "uv",
});
assert.ok(foliage.vertexShader.includes("invariant gl_Position;"));
assert.ok(foliage.fragmentShader.includes("#include <alphatest_fragment>"));
const depth = new THREE.MeshBasicMaterial({ colorWrite: false });
assert.equal(receivesTileHighlight(depth), false);
assert.equal(receivesTileHighlight(shaders.createNuclearMaskMaterial()), false);
assert.equal(receivesTileHighlight(shaders.createNuclearWaveMaterial()), false);
for (const mesh of [landscape.bakeMeshes.ground, landscape.bakeMeshes.basin]) {
  assert.equal(
    receivesTileHighlight(mesh.material),
    false,
    "Ground baking excludes interaction tint",
  );
  fixtures.push({
    name: mesh.material.defines.LANDSCAPE_BASIN ? "basin-bake" : "ground-bake",
    vertex: mesh.material.vertexShader,
    fragment: mesh.material.fragmentShader,
    defines: { ...mesh.material.defines, USE_INSTANCING: "" },
  });
}
fixtures.push({
  name: "landscape-mask",
  vertex: landscape.mask.vertexShader,
  fragment: landscape.mask.fragmentShader,
  defines: { ...landscape.mask.defines, USE_INSTANCING: "" },
});
system.release(board);
assert.equal(meshB.material === source, false);
assert.equal(
  compile("moon-after-mars-release", meshB.material).uniforms.uTileHighlightData.value,
  moon.texture,
);
system.release(moon);
assert.equal(meshB.material, source);
assert.equal(source.onBeforeCompile, THREE.Material.prototype.onBeforeCompile);
system.dispose();
geometry.dispose();
source.dispose();
lambert.dispose();
depth.dispose();
texture.dispose();
array.dispose();
landscape.dispose();
if (process.argv[2]) {
  await Bun.write(process.argv[2], JSON.stringify({ fixtures, chunks: THREE.ShaderChunk }));
}
console.log(
  "Tile lookup, fades, atlas uploads, board isolation, shader composition and teardown verified.",
);
