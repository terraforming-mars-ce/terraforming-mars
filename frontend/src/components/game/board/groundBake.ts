import * as THREE from "three";
import { performanceStore } from "../../../services/performanceStore";

const WIDTH = 4096;
const HEIGHT = 2048;
const STRIPS = 4;
const GRID = 7;
// Extra board area baked around the view, so panning stays on the bake.
const MARGIN = 0.15;
// Rebake for sharpness once the view needs this much more texel density than the bake has.
const DENSITY_SLACK = 1.5;
const STABLE_FRAMES = 3;
const RADIUS = 2.02;

export interface GroundBakeUniforms {
  uGroundBakeAlbedo: { value: THREE.Texture };
  uGroundBakeSurface: { value: THREE.Texture };
  uGroundBakeBounds: { value: THREE.Vector4 };
  uGroundBaked: { value: number };
}

// A single-level runtime virtual texture of the landscape ground. The board region in view is
// rendered once, at screen density, into an albedo and a surface (slope, roughness, occlusion)
// target, so idle frames read two texels instead of recomposing every layer per pixel. Any
// change to the inputs (climate, fields, plant shade, a view outside the bake) falls back to the
// live shader until a new bake, spread over a few frames, completes.
export class GroundBake {
  readonly target = new THREE.WebGLRenderTarget(WIDTH, HEIGHT, {
    count: 2,
    depthBuffer: false,
    type: THREE.UnsignedByteType,
    minFilter: THREE.LinearMipmapLinearFilter,
    magFilter: THREE.LinearFilter,
    generateMipmaps: true,
  });
  readonly region = new THREE.Vector4(0, 0, 1, 1);
  private readonly scene = new THREE.Scene();
  private readonly camera = new THREE.OrthographicCamera();
  private readonly bakeRegion: { value: THREE.Vector4 };
  private readonly signature: number[] = [];
  // The board rectangle on screen, and that rectangle grown by the panning margin.
  private readonly view = new THREE.Vector4();
  private readonly want = new THREE.Vector4();
  private readonly previousWant = new THREE.Vector4();
  private readonly inverse = new THREE.Matrix4();
  private readonly origin = new THREE.Vector3();
  private readonly ray = new THREE.Vector3();
  private readonly hit = new THREE.Vector3();
  private readonly drawingBuffer = new THREE.Vector2();
  private readonly boards = new Float32Array(GRID * GRID * 2);
  private readonly hits = new Uint8Array(GRID * GRID);
  private readonly clearColor = new THREE.Color();
  private density = 0;
  private wantDensity = 0;
  private stable = 0;
  private strip = -1;
  private initialized = false;
  private bakes = 0;
  private lastBakeMs = 0;

  // `uniforms` are the live ground's bake inputs; `bakeRegion` drives the bake pass.
  constructor(
    private readonly meshes: { ground: THREE.InstancedMesh; basin: THREE.InstancedMesh },
    bakeRegion: { value: THREE.Vector4 },
    private readonly uniforms: GroundBakeUniforms,
  ) {
    this.bakeRegion = bakeRegion;
    const [albedo, surface] = this.target.textures;
    albedo.colorSpace = THREE.SRGBColorSpace;
    surface.colorSpace = THREE.NoColorSpace;
    // The board is often seen at a grazing angle; without anisotropy the soft edge smears outward.
    albedo.anisotropy = surface.anisotropy = 8;
    this.scene.add(meshes.ground, meshes.basin);
    uniforms.uGroundBakeAlbedo.value = albedo;
    uniforms.uGroundBakeSurface.value = surface;
    uniforms.uGroundBaked.value = 0;
  }

  // `inputs` lists every value the baked material depends on; `patches` is the board-space
  // extent (x, y, width, height) of the drawn landscape, or null when there is none.
  update(
    gl: THREE.WebGLRenderer,
    camera: THREE.Camera,
    landscape: THREE.Object3D,
    patches: THREE.Vector4 | null,
    basinPresent: boolean,
    settled: boolean,
    inputs: readonly number[],
  ) {
    const changed = this.readInputs(inputs);
    const visible = patches !== null && this.measureView(gl, camera, landscape, patches);
    if (!visible || !settled || changed) {
      this.invalidate();
      return;
    }
    if (this.uniforms.uGroundBaked.value === 1) {
      // Rebaking for sharpness only pays off when a new region would actually hold more texels;
      // otherwise a view needing more than the target can give would rebake forever.
      const achievable = Math.min(WIDTH / this.want.z, HEIGHT / this.want.w);
      const sharper =
        this.wantDensity > this.density * DENSITY_SLACK &&
        achievable > this.density * DENSITY_SLACK;
      if (this.covers() && !sharper) {
        return;
      }
      this.invalidate();
    }
    if (this.strip < 0) {
      // Wait for the view to settle, so a moving camera does not bake every few frames.
      const drift = Math.max(this.want.z, this.want.w) * 0.02;
      const steady =
        Math.abs(this.want.x - this.previousWant.x) < drift &&
        Math.abs(this.want.y - this.previousWant.y) < drift &&
        Math.abs(this.want.z - this.previousWant.z) < drift &&
        Math.abs(this.want.w - this.previousWant.w) < drift;
      this.previousWant.copy(this.want);
      this.stable = steady ? this.stable + 1 : 0;
      if (this.stable < STABLE_FRAMES) {
        return;
      }
      this.region.copy(this.want);
      this.density = Math.min(WIDTH / this.region.z, HEIGHT / this.region.w);
      this.strip = 0;
    }
    this.renderStrip(gl, basinPresent);
  }

  stats() {
    return {
      state: this.uniforms.uGroundBaked.value === 1 ? "baked" : this.strip >= 0 ? "baking" : "live",
      bakes: this.bakes,
      "last bake (CPU ms)": this.lastBakeMs.toFixed(2),
      "texels / unit": Math.round(this.density),
      "view needs": Math.round(this.wantDensity),
    };
  }

  dispose() {
    this.uniforms.uGroundBaked.value = 0;
    this.target.dispose();
  }

  private invalidate() {
    this.uniforms.uGroundBaked.value = 0;
    this.strip = -1;
    this.stable = 0;
  }

  private readInputs(inputs: readonly number[]) {
    let changed = inputs.length !== this.signature.length;
    for (let i = 0; i < inputs.length && !changed; i++) {
      changed = Math.abs(inputs[i] - this.signature[i]) > 1e-4;
    }
    if (changed) {
      this.signature.length = 0;
      for (const value of inputs) {
        this.signature.push(value);
      }
    }
    return changed;
  }

  private covers() {
    const r = this.region,
      v = this.view;
    return v.x >= r.x && v.y >= r.y && v.x + v.z <= r.x + r.z && v.y + v.w <= r.y + r.w;
  }

  // Casts a grid of screen rays onto the planet to find the board region in view and the texel
  // density it needs. Returns false when no landscape is on screen.
  private measureView(
    gl: THREE.WebGLRenderer,
    camera: THREE.Camera,
    landscape: THREE.Object3D,
    patches: THREE.Vector4,
  ) {
    this.inverse.copy(landscape.matrixWorld).invert();
    this.origin.setFromMatrixPosition(camera.matrixWorld).applyMatrix4(this.inverse);
    let minX = Infinity,
      minY = Infinity,
      maxX = -Infinity,
      maxY = -Infinity;
    for (let j = 0; j < GRID; j++) {
      for (let i = 0; i < GRID; i++) {
        const index = j * GRID + i;
        this.hits[index] = 0;
        this.ray
          .set((i / (GRID - 1)) * 2 - 1, (j / (GRID - 1)) * 2 - 1, 0.5)
          .unproject(camera)
          .applyMatrix4(this.inverse)
          .sub(this.origin)
          .normalize();
        const b = this.origin.dot(this.ray);
        const disc = b * b - (this.origin.lengthSq() - RADIUS * RADIUS);
        if (disc < 0) {
          continue;
        }
        const t = -b - Math.sqrt(disc);
        if (t < 0) {
          continue;
        }
        this.hit.copy(this.origin).addScaledVector(this.ray, t);
        if (this.hit.z <= 0) {
          continue;
        }
        const arc = RADIUS * Math.acos(Math.min(1, this.hit.z / RADIUS));
        const planar = Math.hypot(this.hit.x, this.hit.y);
        const x = planar < 1e-9 ? 0 : (this.hit.x / planar) * arc;
        const y = planar < 1e-9 ? 0 : (this.hit.y / planar) * arc;
        this.hits[index] = 1;
        this.boards[index * 2] = x;
        this.boards[index * 2 + 1] = y;
        minX = Math.min(minX, x);
        minY = Math.min(minY, y);
        maxX = Math.max(maxX, x);
        maxY = Math.max(maxY, y);
      }
    }
    // Rays sample the view coarsely; widen by half a grid step so screen edges stay inside.
    const stepX = (maxX - minX) / (GRID - 1) / 2,
      stepY = (maxY - minY) / (GRID - 1) / 2;
    minX = Math.max(minX - stepX, patches.x);
    minY = Math.max(minY - stepY, patches.y);
    maxX = Math.min(maxX + stepX, patches.x + patches.z);
    maxY = Math.min(maxY + stepY, patches.y + patches.w);
    if (!(maxX > minX && maxY > minY)) {
      return false;
    }
    const size = gl.getDrawingBufferSize(this.drawingBuffer);
    const pixelStepX = size.x / (GRID - 1),
      pixelStepY = size.y / (GRID - 1);
    let density = 0;
    for (let j = 0; j < GRID; j++) {
      for (let i = 0; i < GRID; i++) {
        const index = j * GRID + i;
        if (!this.hits[index]) {
          continue;
        }
        if (i + 1 < GRID && this.hits[index + 1]) {
          const d = Math.hypot(
            this.boards[index * 2 + 2] - this.boards[index * 2],
            this.boards[index * 2 + 3] - this.boards[index * 2 + 1],
          );
          density = Math.max(density, pixelStepX / Math.max(d, 1e-6));
        }
        if (j + 1 < GRID && this.hits[index + GRID]) {
          const d = Math.hypot(
            this.boards[(index + GRID) * 2] - this.boards[index * 2],
            this.boards[(index + GRID) * 2 + 1] - this.boards[index * 2 + 1],
          );
          density = Math.max(density, pixelStepY / Math.max(d, 1e-6));
        }
      }
    }
    this.wantDensity = density;
    this.view.set(minX, minY, maxX - minX, maxY - minY);
    const marginX = (maxX - minX) * MARGIN,
      marginY = (maxY - minY) * MARGIN;
    const x0 = Math.max(minX - marginX, patches.x),
      y0 = Math.max(minY - marginY, patches.y);
    const x1 = Math.min(maxX + marginX, patches.x + patches.z),
      y1 = Math.min(maxY + marginY, patches.y + patches.w);
    this.want.set(x0, y0, x1 - x0, y1 - y0);
    return true;
  }

  private renderStrip(gl: THREE.WebGLRenderer, basinPresent: boolean) {
    const started = performance.now();
    if (this.strip === 0) {
      this.lastBakeMs = 0;
    }
    const [albedo, surface] = this.target.textures;
    if (!this.initialized) {
      gl.initRenderTarget(this.target);
      this.initialized = true;
    }
    const last = this.strip === STRIPS - 1;
    albedo.generateMipmaps = surface.generateMipmaps = last;
    this.bakeRegion.value.copy(this.region);
    this.meshes.basin.visible = basinPresent;
    const stripHeight = HEIGHT / STRIPS;
    this.target.scissor.set(0, this.strip * stripHeight, WIDTH, stripHeight);
    this.target.scissorTest = true;
    const previousTarget = gl.getRenderTarget();
    const previousAlpha = gl.getClearAlpha();
    gl.getClearColor(this.clearColor);
    gl.setRenderTarget(this.target);
    gl.setClearColor(0x000000, 0);
    gl.clear(true, false, false);
    gl.render(this.scene, this.camera);
    gl.setClearColor(this.clearColor, previousAlpha);
    gl.setRenderTarget(previousTarget);
    this.target.scissorTest = false;
    this.lastBakeMs += performance.now() - started;
    if (!last) {
      this.strip++;
      return;
    }
    this.strip = -1;
    this.bakes++;
    this.uniforms.uGroundBakeBounds.value.copy(this.region);
    this.uniforms.uGroundBaked.value = 1;
    if (performanceStore.watching) {
      performanceStore.setSection("Ground bake", this.stats());
    }
  }
}
