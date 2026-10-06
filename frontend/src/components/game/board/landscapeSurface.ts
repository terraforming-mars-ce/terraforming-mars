import { MARS_RELIEF_DEPTH, marsReliefAtBoard, type MarsRelief } from "./marsRelief";
import * as THREE from "three";
import {
  FIELD_SIZE,
  FIELD_SAMPLES,
  FIELD_BORDER,
  FIELD_STEP,
  PATCH_SIZE,
  WATER_LEVEL,
  BEACH_WIDTH_MIN,
  BEACH_WIDTH_MAX,
  LAKE_MASK_REACH,
  LAKE_GROUND_REACH,
  TRANSITION_MS,
  WATER_RADIUS,
  EMPTY_SHORE,
} from "./landscapeFields";
import { EMPTY_LANDSCAPE, type LandscapeState, type LandscapePatch } from "./landscapeTypes";
import { splitSnippet } from "./shaders";
import { MOHOLE_STENCIL_BIT, LAKE_STENCIL_BIT, NUCLEAR_STENCIL_BIT } from "./boardConstants";
import field from "./shaders/landscape-field.glsl?raw";
import vertex from "./shaders/landscape-ground.vert.glsl?raw";
import fragment from "./shaders/landscape-ground.frag.glsl?raw";
import detail from "./shaders/landscape-detail.frag.glsl?raw";
import occlusion from "./shaders/landscape-ao.frag.glsl?raw";
import normal from "./shaders/landscape-normal.vert.glsl?raw";
import waterVertex from "./shaders/ocean-renderer.vert.glsl?raw";
import waterFragment from "./shaders/ocean-renderer.frag.glsl?raw";
import maskVertex from "./shaders/landscape-mask.vert.glsl?raw";
import maskFragment from "./shaders/landscape-mask.frag.glsl?raw";
import bakeVertex from "./shaders/landscape-bake.vert.glsl?raw";
import { GroundBake } from "./groundBake";
import marsSurface from "./shaders/mars-surface.glsl?raw";
import type { useTextures } from "../../../hooks/useTextures";
const DEFAULT_CAPACITY = 128;
const STRIDE = FIELD_SIZE * FIELD_SIZE * 4;
const noop = () => {};
const NO_PLANT_SHADE = new THREE.DataTexture(new Uint8Array([0, 0, 0, 255]), 1, 1);
NO_PLANT_SHADE.needsUpdate = true;
export interface LandscapeTextureArrays {
  groundAlbedo: THREE.DataArrayTexture;
  groundDetail: THREE.DataArrayTexture;
  iceAlbedo: THREE.DataArrayTexture;
  iceDetail: THREE.DataArrayTexture;
}
export class LandscapeSurface {
  readonly group = Object.assign(new THREE.Group(), { userData: { perfGroup: "landscape" } });
  readonly terrain: THREE.DataArrayTexture;
  readonly materials: THREE.DataArrayTexture;
  readonly detail: THREE.DataArrayTexture;
  readonly uniforms;
  readonly ground: THREE.MeshStandardMaterial;
  readonly groundEdge: THREE.MeshStandardMaterial;
  readonly basin: THREE.MeshStandardMaterial;
  readonly mask: THREE.ShaderMaterial;
  readonly water: THREE.ShaderMaterial;
  private geometry = new THREE.PlaneGeometry(1, 1, 32, 32);
  private origins: THREE.InstancedBufferAttribute;
  private layers: THREE.InstancedBufferAttribute;
  private births: THREE.InstancedBufferAttribute;
  private meshes: THREE.InstancedMesh[];
  private bakeMeshes: {
    ground: THREE.InstancedMesh<THREE.PlaneGeometry, THREE.ShaderMaterial>;
    basin: THREE.InstancedMesh<THREE.PlaneGeometry, THREE.ShaderMaterial>;
  };
  private bakeRegion = { value: new THREE.Vector4(0, 0, 1, 1) };
  private groundBake: GroundBake | null = null;
  private patchBounds = new THREE.Vector4();
  private boundsRevision = -1;
  private readonly bakeInputs: number[] = [];
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
    arrays: LandscapeTextureArrays,
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
    this.detail = new THREE.DataArrayTexture(
      new Uint8Array(STRIDE * capacity * 2),
      FIELD_SIZE,
      FIELD_SIZE,
      capacity * 2,
    );
    this.origins = new THREE.InstancedBufferAttribute(new Float32Array(capacity * 2), 2);
    this.layers = new THREE.InstancedBufferAttribute(new Float32Array(capacity), 1);
    this.births = new THREE.InstancedBufferAttribute(new Float32Array(capacity).fill(-1000), 1);
    this.terrain.type = THREE.HalfFloatType;
    for (const texture of [this.terrain, this.materials, this.detail]) {
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
      uDetailField: { value: this.detail },
      uLandscapeTime: { value: 0 },
      uLayerCapacity: { value: this.capacity },
      uPatchSize: { value: PATCH_SIZE },
      uFieldSize: { value: FIELD_SIZE },
      uFieldSamples: { value: FIELD_SAMPLES },
      uFieldBorder: { value: FIELD_BORDER },
      uWaterLevel: { value: WATER_LEVEL },
      uMarsReliefDepth: { value: MARS_RELIEF_DEPTH },
      uBeachWidths: { value: new THREE.Vector2(BEACH_WIDTH_MIN, BEACH_WIDTH_MAX) },
      uBasinReach: { value: new THREE.Vector2(LAKE_MASK_REACH, LAKE_GROUND_REACH) },
      uLakeShape: { value: new THREE.Vector2(WATER_RADIUS, EMPTY_SHORE) },
      uFrost: { value: 0 },
      uChill: { value: 0 },
      uGreening: { value: 1 },
      uMeadow: { value: 1 },
      uPlantShade: { value: NO_PLANT_SHADE as THREE.Texture },
      uPlantShadeBounds: { value: new THREE.Vector4(0, 0, 1, 1) },
      uClimateNoise: { value: textures.noiseMid },
      uIce: { value: 0 },
      uIceAge: { value: 1 },
      uIceReach: { value: WATER_RADIUS },
      uClarity: { value: 1 },
      uAlgae: { value: 0 },
      uWaves: { value: 1 },
      uFoam: { value: 1 },
      uShimmer: { value: 0 },
      uGroundBakeAlbedo: { value: NO_PLANT_SHADE as THREE.Texture },
      uGroundBakeSurface: { value: NO_PLANT_SHADE as THREE.Texture },
      uGroundBakeBounds: { value: new THREE.Vector4(0, 0, 1, 1) },
      uGroundBaked: { value: 0 },
    };
    const vs = splitSnippet(vertex),
      fs = splitSnippet(fragment);
    this.ground = new THREE.MeshStandardMaterial({
      roughness: 0.9,
    });
    const groundTextures = {
      uMars: { value: textures.mars },
      uDryGrass: { value: textures.grass },
      uGroundAlbedo: { value: arrays.groundAlbedo },
      uGroundDetail: { value: arrays.groundDetail },
      uSand: { value: textures.sand },
      uRock: { value: textures.rock },
      uPaving: { value: textures.concrete },
    };
    this.ground.onBeforeCompile = (shader) => {
      Object.assign(shader.uniforms, this.uniforms, groundTextures);
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
        marsSurface +
        "\n" +
        fs.header +
        "\n" +
        shader.fragmentShader
          .replace("#include <map_fragment>", fs.body)
          .replace("#include <normal_fragment_maps>", "#include <normal_fragment_maps>\n" + detail)
          .replace("#include <aomap_fragment>", occlusion);
    };
    this.ground.customProgramCacheKey = () =>
      field + marsSurface + vertex + fragment + normal + detail + occlusion;
    this.groundEdge = new THREE.MeshStandardMaterial({
      transparent: true,
      depthWrite: false,
      alphaTest: 0.001,
      roughness: 0.9,
    });
    this.groundEdge.defines = { STANDARD: "", LANDSCAPE_EDGE: 1 };
    this.groundEdge.onBeforeCompile = this.ground.onBeforeCompile;
    this.groundEdge.customProgramCacheKey = () =>
      "edge" + field + marsSurface + vertex + fragment + normal + detail + occlusion;
    this.basin = new THREE.MeshStandardMaterial({
      roughness: 0.9,
    });
    this.basin.defines = { STANDARD: "", LANDSCAPE_BASIN: 1 };
    this.basin.onBeforeCompile = this.ground.onBeforeCompile;
    this.basin.customProgramCacheKey = () =>
      "basin" + field + marsSurface + vertex + fragment + normal + detail + occlusion;
    // The ground bake renders the same material stage flat into board space, writing albedo and
    // surface terms instead of lighting them.
    const bakeMaterial = (basin: boolean) =>
      new THREE.ShaderMaterial({
        glslVersion: THREE.GLSL3,
        userData: { planetHaze: false },
        defines: basin ? { GROUND_BAKE: 1, LANDSCAPE_BASIN: 1 } : { GROUND_BAKE: 1 },
        uniforms: { ...this.uniforms, ...groundTextures, uBakeRegion: this.bakeRegion },
        vertexShader: field + "\n" + bakeVertex,
        fragmentShader: [
          "#include <common>",
          field,
          marsSurface,
          fs.header,
          "varying vec3 vViewPosition;",
          "layout(location=0) out vec4 bakeAlbedo;",
          "layout(location=1) out vec4 bakeSurface;",
          "void main() {",
          "vec4 diffuseColor=vec4(1.0);",
          fs.body,
          // Premultiplied, so mip levels and filtering at the ground's soft edge stay unbiased.
          "bakeAlbedo=vec4(diffuseColor.rgb*diffuseColor.a,diffuseColor.a);",
          "bakeSurface=vec4(clamp(surfaceSlope*0.5+0.5,0.0,1.0),surfaceRoughness,landscapeAO)*diffuseColor.a;",
          "}",
        ].join("\n"),
        depthTest: false,
        depthWrite: false,
      });
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
      userData: { planetHaze: true, tileHighlights: true },
      vertexShader: field + "\n" + waterVertex,
      fragmentShader: field + "\n" + waterFragment,
      uniforms: {
        ...this.uniforms,
        normalSampler: { value: textures.waterNormals },
        uIceAlbedo: { value: arrays.iceAlbedo },
        uIceDetail: { value: arrays.iceDetail },
        time: { value: 0 },
        eye: { value: new THREE.Vector3() },
        sunDirection: { value: new THREE.Vector3(0.9, 0, 0.8) },
        sunColor: { value: new THREE.Vector3(1, 1, 1) },
        sunIntensity: { value: 1 },
        rf0: { value: 0.1 },
      },
      transparent: true,
      depthWrite: false,
    });
    for (const material of [this.ground, this.groundEdge, this.basin, this.water]) {
      material.stencilWrite = true;
      material.stencilWriteMask = 0;
      material.stencilFunc = THREE.EqualStencilFunc;
      material.stencilRef = 0;
      material.stencilFuncMask = MOHOLE_STENCIL_BIT | NUCLEAR_STENCIL_BIT;
    }
    this.geometry.setAttribute("patchOrigin", this.origins);
    this.geometry.setAttribute("patchLayer", this.layers);
    this.geometry.setAttribute("patchBirth", this.births);
    const identity = new THREE.Matrix4();
    this.bakeMeshes = {
      ground: new THREE.InstancedMesh(this.geometry, bakeMaterial(false), this.capacity),
      basin: new THREE.InstancedMesh(this.geometry, bakeMaterial(true), this.capacity),
    };
    this.meshes = [this.mask, this.basin, this.ground, this.water, this.groundEdge].map(
      (material, i) => {
        const mesh = new THREE.InstancedMesh(this.geometry, material, this.capacity);
        for (let slot = 0; slot < this.capacity; slot++) {
          mesh.setMatrixAt(slot, identity);
          this.layers.setX(slot, slot);
          this.origins.setXY(slot, 20, 20);
        }
        mesh.count = 0;
        mesh.frustumCulled = false;
        mesh.raycast = noop;
        mesh.renderOrder = [-2, 11, 11, 12, 11][i];
        this.group.add(mesh);
        return mesh;
      },
    );
  }
  private clearTerrain(layer: number, x = 0, y = 0, relief?: MarsRelief) {
    const data = this.terrain.image.data as Uint16Array;
    const detailData = this.detail.image.data as Uint8Array;
    const shore = THREE.DataUtils.toHalfFloat(EMPTY_SHORE);
    for (let pixel = 0; pixel < FIELD_SIZE * FIELD_SIZE; pixel++) {
      const i = layer * STRIDE + pixel * 4;
      const height = marsReliefAtBoard(
        relief,
        x + ((pixel % FIELD_SIZE) - FIELD_BORDER) * FIELD_STEP,
        y + (Math.floor(pixel / FIELD_SIZE) - FIELD_BORDER) * FIELD_STEP,
      );
      data[i] = THREE.DataUtils.toHalfFloat(height + 0.0003);
      data[i + 1] = shore;
      data[i + 2] = data[i + 3] = 0;
      detailData[i + 3] = Math.round((-height / MARS_RELIEF_DEPTH) * 255);
    }
  }
  enqueue(state: LandscapeState) {
    this.pending = state;
  }
  // Draw only up to the highest live slot; released patches must not keep being drawn off-screen.
  private updateInstanceCount() {
    let count = 0;
    for (const slot of this.slots.values()) {
      count = Math.max(count, slot + 1);
    }
    for (const mesh of this.meshes) {
      mesh.count = count;
    }
    this.bakeMeshes.ground.count = this.bakeMeshes.basin.count = count;
  }
  private updatePassVisibility() {
    for (const mesh of this.meshes) {
      mesh.visible =
        mesh.material === this.ground || mesh.material === this.groundEdge || this.basinPresent;
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
      let released = false;
      for (const [key, slot] of this.slots) {
        if (!this.committed.patches.has(key) && !this.staging) {
          this.slots.delete(key);
          this.origins.setXY(slot, 20, 20);
          this.origins.needsUpdate = true;
          released = true;
        }
      }
      if (released) {
        this.updateInstanceCount();
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
      const newSlot = slot === undefined;
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
        materials = this.materials.image.data as Uint8Array,
        detailData = this.detail.image.data as Uint8Array;
      terrain.copyWithin(old, next, next + STRIDE);
      materials.copyWithin(old, next, next + STRIDE);
      detailData.copyWithin(old, next, next + STRIDE);
      if (patch) {
        if (newSlot) {
          this.clearTerrain(slot, patch.x, patch.y, this.staging.relief);
        }
        terrain.set(patch.terrain, next);
        materials.set(patch.materials, next);
        detailData.set(patch.detail, next);
        this.origins.setXY(slot, patch.x, patch.y);
      } else {
        materials.fill(0, next, next + STRIDE);
        detailData.fill(0, next, next + STRIDE);
        this.clearTerrain(
          slot + this.capacity,
          this.origins.getX(slot),
          this.origins.getY(slot),
          this.staging.relief,
        );
      }
      // Keep the old field visible until every affected patch has reached the GPU.
      this.births.setX(slot, 1e12);
      this.terrain.addLayerUpdate(slot);
      this.terrain.addLayerUpdate(slot + this.capacity);
      this.materials.addLayerUpdate(slot);
      this.materials.addLayerUpdate(slot + this.capacity);
      this.detail.addLayerUpdate(slot);
      this.detail.addLayerUpdate(slot + this.capacity);
      this.terrain.needsUpdate = this.materials.needsUpdate = this.detail.needsUpdate = true;
      renderer.initTexture(this.terrain);
      renderer.initTexture(this.materials);
      renderer.initTexture(this.detail);
      this.touched.push(slot);
    }
    this.origins.needsUpdate = this.births.needsUpdate = true;
    this.updateInstanceCount();
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
  // Bakes the ground in view once the landscape and its inputs hold still; see GroundBake.
  updateBake(
    gl: THREE.WebGLRenderer,
    camera: THREE.Camera,
    now: number,
    plantShadeRevision: number,
  ) {
    this.groundBake ??= new GroundBake(this.bakeMeshes, this.bakeRegion, this.uniforms);
    const u = this.uniforms;
    const bounds = u.uPlantShadeBounds.value;
    const inputs = this.bakeInputs;
    inputs.length = 0;
    inputs.push(
      u.uFrost.value,
      u.uChill.value,
      u.uGreening.value,
      u.uMeadow.value,
      u.uIce.value,
      plantShadeRevision,
      bounds.x,
      bounds.y,
      bounds.z,
      bounds.w,
      this.committed.id,
      this.committed.relief?.revision ?? 0,
    );
    const settled =
      !this.staging && !this.pending && !this.queue.length && now >= this.transitionEnd;
    this.groundBake.update(
      gl,
      camera,
      this.group,
      this.updatePatchBounds(),
      this.basinPresent,
      settled,
      inputs,
    );
  }
  // Draws both bake programs once, so the first bake does not stall. Compiling alone is not enough:
  // drivers finish a program on its first draw. Three keys programs on the bound target's tone
  // mapping and color space, so the draw goes into a render target like the bake's.
  warmBake(gl: THREE.WebGLRenderer) {
    const scene = new THREE.Scene();
    scene.add(this.bakeMeshes.ground, this.bakeMeshes.basin);
    this.bakeMeshes.basin.visible = true;
    const bounds = this.updatePatchBounds();
    if (bounds) {
      this.bakeRegion.value.copy(bounds);
    }
    const target = new THREE.WebGLRenderTarget(1, 1, { count: 2, depthBuffer: false });
    const previousTarget = gl.getRenderTarget();
    gl.setRenderTarget(target);
    gl.render(scene, new THREE.OrthographicCamera());
    gl.setRenderTarget(previousTarget);
    target.dispose();
    scene.clear();
  }
  // Allocates the bake target up front; its first allocation is large enough to stall a frame.
  prepareBake(gl: THREE.WebGLRenderer) {
    this.groundBake ??= new GroundBake(this.bakeMeshes, this.bakeRegion, this.uniforms);
    this.groundBake.prepare(gl);
  }
  bakeStats() {
    return this.groundBake?.stats() ?? { state: "off" };
  }
  private updatePatchBounds() {
    if (this.boundsRevision !== this.committed.id) {
      this.boundsRevision = this.committed.id;
      let minX = Infinity,
        minY = Infinity,
        maxX = -Infinity,
        maxY = -Infinity;
      for (const patch of this.committed.patches.values()) {
        minX = Math.min(minX, patch.x);
        minY = Math.min(minY, patch.y);
        maxX = Math.max(maxX, patch.x + PATCH_SIZE);
        maxY = Math.max(maxY, patch.y + PATCH_SIZE);
      }
      this.patchBounds.set(minX, minY, maxX - minX, maxY - minY);
    }
    return this.committed.patches.size ? this.patchBounds : null;
  }
  stats() {
    return {
      "patches shown": this.committed.patches.size,
      "slots used": `${this.slots.size} / ${this.capacity}`,
      "uploads queued": this.queue.length,
      transitioning: this.staging ? "yes" : "no",
    };
  }
  dispose() {
    this.terrain.dispose();
    this.materials.dispose();
    this.detail.dispose();
    this.ground.dispose();
    this.groundEdge.dispose();
    this.basin.dispose();
    this.mask.dispose();
    this.water.dispose();
    this.bakeMeshes.ground.material.dispose();
    this.bakeMeshes.basin.material.dispose();
    this.groundBake?.dispose();
    this.geometry.dispose();
    for (const m of this.meshes) {
      m.dispose();
    }
    this.group.clear();
  }
}
