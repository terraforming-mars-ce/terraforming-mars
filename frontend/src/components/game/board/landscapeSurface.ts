import * as THREE from "three";
import {
  FIELD_SIZE,
  FIELD_SAMPLES,
  FIELD_BORDER,
  PATCH_SIZE,
  WATER_LEVEL,
  BEACH_WIDTH_MIN,
  BEACH_WIDTH_MAX,
  LAKE_MASK_REACH,
  LAKE_GROUND_REACH,
  TRANSITION_MS,
} from "./landscapeFields";
import { EMPTY_LANDSCAPE, type LandscapeState, type LandscapePatch } from "./landscapeTypes";
import { splitSnippet } from "./shaders";
import { MOHOLE_STENCIL_BIT, LAKE_STENCIL_BIT } from "./boardConstants";
import field from "./shaders/landscape-field.glsl?raw";
import vertex from "./shaders/landscape-ground.vert.glsl?raw";
import fragment from "./shaders/landscape-ground.frag.glsl?raw";
import detail from "./shaders/landscape-detail.frag.glsl?raw";
import normal from "./shaders/landscape-normal.vert.glsl?raw";
import waterVertex from "./shaders/ocean-renderer.vert.glsl?raw";
import waterFragment from "./shaders/ocean-renderer.frag.glsl?raw";
import maskVertex from "./shaders/landscape-mask.vert.glsl?raw";
import maskFragment from "./shaders/landscape-mask.frag.glsl?raw";
import type { useTextures } from "../../../hooks/useTextures";
const DEFAULT_CAPACITY = 128;
const STRIDE = FIELD_SIZE * FIELD_SIZE * 4;
const noop = () => {};
export class LandscapeSurface {
  readonly group = new THREE.Group();
  readonly terrain: THREE.DataArrayTexture;
  readonly materials: THREE.DataArrayTexture;
  readonly uniforms;
  readonly ground: THREE.MeshStandardMaterial;
  readonly basin: THREE.MeshStandardMaterial;
  readonly mask: THREE.ShaderMaterial;
  readonly water: THREE.ShaderMaterial;
  private geometry = new THREE.PlaneGeometry(1, 1, 32, 32);
  private origins: THREE.InstancedBufferAttribute;
  private layers: THREE.InstancedBufferAttribute;
  private births: THREE.InstancedBufferAttribute;
  private meshes: THREE.InstancedMesh[];
  private slots = new Map<string, number>();
  private committed = EMPTY_LANDSCAPE;
  private pending: LandscapeState | null = null;
  private staging: LandscapeState | null = null;
  private queue: { key: string; patch?: LandscapePatch }[] = [];
  private touched: number[] = [];
  private transitionEnd = 0;
  private basinPresent = false;
  private nextBasinPresent = false;
  private users = 0;
  retain() {
    this.users++;
  }
  release() {
    this.users--;
    queueMicrotask(() => {
      if (this.users === 0) {
        this.dispose();
      }
    });
  }
  constructor(
    textures: ReturnType<typeof useTextures>,
    private readonly capacity = DEFAULT_CAPACITY,
  ) {
    this.terrain = new THREE.DataArrayTexture(
      new Uint16Array(STRIDE * capacity * 2),
      FIELD_SIZE,
      FIELD_SIZE,
      capacity * 2,
    );
    this.materials = new THREE.DataArrayTexture(
      new Uint8Array(STRIDE * capacity * 2),
      FIELD_SIZE,
      FIELD_SIZE,
      capacity * 2,
    );
    this.origins = new THREE.InstancedBufferAttribute(new Float32Array(capacity * 2), 2);
    this.layers = new THREE.InstancedBufferAttribute(new Float32Array(capacity), 1);
    this.births = new THREE.InstancedBufferAttribute(new Float32Array(capacity).fill(-1000), 1);
    this.terrain.type = THREE.HalfFloatType;
    for (const texture of [this.terrain, this.materials]) {
      texture.minFilter = texture.magFilter = THREE.LinearFilter;
      texture.wrapS = texture.wrapT = THREE.ClampToEdgeWrapping;
      texture.generateMipmaps = false;
      texture.needsUpdate = true;
    }
    for (let i = 0; i < this.capacity * 2; i++) {
      this.clearTerrain(i);
    }
    this.uniforms = {
      uTerrain: { value: this.terrain },
      uMaterials: { value: this.materials },
      uLandscapeTime: { value: 0 },
      uLayerCapacity: { value: this.capacity },
      uPatchSize: { value: PATCH_SIZE },
      uFieldSize: { value: FIELD_SIZE },
      uFieldSamples: { value: FIELD_SAMPLES },
      uFieldBorder: { value: FIELD_BORDER },
      uWaterLevel: { value: WATER_LEVEL },
      uBeachWidths: { value: new THREE.Vector2(BEACH_WIDTH_MIN, BEACH_WIDTH_MAX) },
      uBasinReach: { value: new THREE.Vector2(LAKE_MASK_REACH, LAKE_GROUND_REACH) },
    };
    const vs = splitSnippet(vertex),
      fs = splitSnippet(fragment);
    this.ground = new THREE.MeshStandardMaterial({
      transparent: true,
      depthWrite: false,
      alphaTest: 0.001,
      roughness: 0.9,
    });
    this.ground.onBeforeCompile = (shader) => {
      Object.assign(shader.uniforms, this.uniforms, {
        uMars: { value: textures.mars },
        uDryGrass: { value: textures.grass },
        uLushGrass: { value: textures.leafyGrass },
        uForestFloor: { value: textures.forestLitter },
        uWetSoil: { value: textures.wetSoil },
        uLushDetail: { value: textures.leafyGrassDetail },
        uLitterDetail: { value: textures.forestLitterDetail },
        uSoilDetail: { value: textures.wetSoilDetail },
        uSand: { value: textures.sand },
        uRock: { value: textures.rock },
        uPaving: { value: textures.concrete },
      });
      shader.vertexShader =
        field +
        "\n" +
        vs.header +
        "\n" +
        shader.vertexShader
          .replace("#include <beginnormal_vertex>", normal)
          .replace("#include <begin_vertex>", vs.body);
      shader.fragmentShader =
        field +
        "\n" +
        fs.header +
        "\n" +
        shader.fragmentShader
          .replace("#include <map_fragment>", fs.body)
          .replace("#include <normal_fragment_maps>", "#include <normal_fragment_maps>\n" + detail);
    };
    this.ground.customProgramCacheKey = () => field + vertex + fragment + normal + detail;
    this.basin = new THREE.MeshStandardMaterial({
      roughness: 0.9,
    });
    this.basin.defines = { STANDARD: "", LANDSCAPE_BASIN: 1 };
    this.basin.onBeforeCompile = this.ground.onBeforeCompile;
    this.basin.customProgramCacheKey = () => "basin" + field + vertex + fragment + normal + detail;
    this.mask = new THREE.ShaderMaterial({
      userData: { planetHaze: false },
      vertexShader: field + "\n" + maskVertex,
      fragmentShader: field + "\n" + maskFragment,
      uniforms: this.uniforms,
      colorWrite: false,
      depthWrite: false,
      stencilWrite: true,
      stencilRef: LAKE_STENCIL_BIT,
      stencilWriteMask: LAKE_STENCIL_BIT,
      stencilFunc: THREE.AlwaysStencilFunc,
      stencilZPass: THREE.ReplaceStencilOp,
    });
    this.water = new THREE.ShaderMaterial({
      userData: { planetHaze: true },
      vertexShader: field + "\n" + waterVertex,
      fragmentShader: field + "\n" + waterFragment,
      uniforms: {
        ...this.uniforms,
        normalSampler: { value: textures.waterNormals },
        time: { value: 0 },
        eye: { value: new THREE.Vector3() },
        sunDirection: { value: new THREE.Vector3(0.9, 0, 0.8) },
        sunColor: { value: new THREE.Vector3(1, 1, 1) },
        sunIntensity: { value: 1 },
        waterColor: { value: new THREE.Vector3(0.01, 0.03, 0.03) },
        rf0: { value: 0.1 },
        uHoverCenter: { value: new THREE.Vector2() },
        uHoverActive: { value: 0 },
      },
      transparent: true,
      depthWrite: false,
    });
    for (const material of [this.ground, this.basin, this.water]) {
      material.stencilWrite = true;
      material.stencilWriteMask = 0;
      material.stencilFunc = THREE.EqualStencilFunc;
      material.stencilRef = 0;
      material.stencilFuncMask = MOHOLE_STENCIL_BIT;
    }
    this.geometry.setAttribute("patchOrigin", this.origins);
    this.geometry.setAttribute("patchLayer", this.layers);
    this.geometry.setAttribute("patchBirth", this.births);
    const identity = new THREE.Matrix4();
    this.meshes = [this.mask, this.basin, this.ground, this.water].map((material, i) => {
      const mesh = new THREE.InstancedMesh(this.geometry, material, this.capacity);
      for (let slot = 0; slot < this.capacity; slot++) {
        mesh.setMatrixAt(slot, identity);
        this.layers.setX(slot, slot);
        this.origins.setXY(slot, 20, 20);
      }
      mesh.count = 0;
      mesh.frustumCulled = false;
      mesh.raycast = noop;
      mesh.renderOrder = [-2, 11, 11, 12][i];
      this.group.add(mesh);
      return mesh;
    });
  }
  private clearTerrain(layer: number) {
    const data = this.terrain.image.data as Uint16Array;
    const height = THREE.DataUtils.toHalfFloat(0.0003),
      shore = THREE.DataUtils.toHalfFloat(0.2);
    for (let i = layer * STRIDE; i < (layer + 1) * STRIDE; i += 4) {
      data[i] = height;
      data[i + 1] = shore;
      data[i + 2] = data[i + 3] = 0;
    }
  }
  enqueue(state: LandscapeState) {
    this.pending = state;
  }
  private updatePassVisibility() {
    for (const mesh of this.meshes) {
      mesh.visible = mesh.material === this.ground || this.basinPresent;
    }
  }
  tick(
    renderer: THREE.WebGLRenderer,
    now: number,
    onReady: (state: LandscapeState, transitionStart: number) => void,
  ) {
    this.uniforms.uLandscapeTime.value = now / 1000;
    if (now >= this.transitionEnd) {
      this.basinPresent = this.nextBasinPresent;
      this.updatePassVisibility();
      for (const [key, slot] of this.slots) {
        if (!this.committed.patches.has(key) && !this.staging) {
          this.slots.delete(key);
          this.origins.setXY(slot, 20, 20);
          this.origins.needsUpdate = true;
        }
      }
    }
    if (!this.staging && this.pending && now >= this.transitionEnd) {
      this.staging = this.pending;
      this.pending = null;
      const state = this.staging;
      this.queue = [...state.patches]
        .filter(([key, p]) => this.committed.patches.get(key) !== p)
        .map(([key, patch]) => ({ key, patch }));
      this.queue.push(
        ...[...this.committed.patches.keys()]
          .filter((key) => !state.patches.has(key))
          .map((key) => ({ key })),
      );
      this.touched = [];
    }
    if (!this.staging) {
      return;
    }
    const start = performance.now();
    while (this.queue.length && performance.now() - start < 2) {
      const { key, patch } = this.queue.shift()!;
      let slot = this.slots.get(key);
      if (slot === undefined) {
        const used = new Set(this.slots.values());
        slot = Array.from({ length: this.capacity }, (_, i) => i).find((i) => !used.has(i));
        if (slot === undefined) {
          throw new Error("Landscape patch capacity exceeded");
        }
        this.slots.set(key, slot);
      }
      const old = slot * STRIDE,
        next = (slot + this.capacity) * STRIDE;
      const terrain = this.terrain.image.data as Uint16Array,
        materials = this.materials.image.data as Uint8Array;
      terrain.copyWithin(old, next, next + STRIDE);
      materials.copyWithin(old, next, next + STRIDE);
      if (patch) {
        terrain.set(patch.terrain, next);
        materials.set(patch.materials, next);
        this.origins.setXY(slot, patch.x, patch.y);
      } else {
        this.clearTerrain(slot + this.capacity);
        materials.fill(0, next, next + STRIDE);
      }
      // Keep the old field visible until every affected patch has reached the GPU.
      this.births.setX(slot, 1e12);
      this.terrain.addLayerUpdate(slot);
      this.terrain.addLayerUpdate(slot + this.capacity);
      this.materials.addLayerUpdate(slot);
      this.materials.addLayerUpdate(slot + this.capacity);
      this.terrain.needsUpdate = this.materials.needsUpdate = true;
      renderer.initTexture(this.terrain);
      renderer.initTexture(this.materials);
      this.touched.push(slot);
    }
    this.origins.needsUpdate = this.births.needsUpdate = true;
    for (const mesh of this.meshes) {
      mesh.count = this.slots.size ? Math.max(...this.slots.values()) + 1 : 0;
    }
    if (!this.queue.length) {
      const first = this.committed.id === 0;
      for (const slot of this.touched) {
        this.births.setX(slot, first ? -1000 : now / 1000);
      }
      this.births.needsUpdate = true;
      this.nextBasinPresent = this.staging.sources.some((source) => source.kind === "ocean");
      this.basinPresent ||= this.nextBasinPresent;
      this.updatePassVisibility();
      this.committed = this.staging;
      this.staging = null;
      this.transitionEnd = first ? now : now + TRANSITION_MS;
      onReady(this.committed, first ? now - TRANSITION_MS : now);
    }
  }
  dispose() {
    this.terrain.dispose();
    this.materials.dispose();
    this.ground.dispose();
    this.basin.dispose();
    this.mask.dispose();
    this.water.dispose();
    this.geometry.dispose();
    for (const m of this.meshes) {
      m.dispose();
    }
    this.group.clear();
  }
}
