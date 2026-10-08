import * as THREE from "three";
import vertexShader from "./shaders/plant-shade.vert.glsl?raw";
import fragmentShader from "./shaders/plant-shade.frag.glsl?raw";
import { performanceStore } from "../../../services/performanceStore";

const RESOLUTION = 2048;
const CAPACITY = 65536;
const MARGIN = 0.05;

export interface PlantShadeBlob {
  x: number;
  y: number;
  radius: number;
  strength: number;
  // How far the cast shadow reaches from the base, in board units.
  length: number;
}

const liveShadeMaps = new Set<PlantShadeMap>();

// A restored context has an empty shade target; owners rebake when `contentLost` is set.
export function invalidatePlantShadesAfterContextRestore() {
  for (const shade of liveShadeMaps) {
    shade.contentLost = true;
  }
}

// A top-down shade map of every visible plant, baked on the GPU whenever the plant set, sizes or
// sun change. Each plant gets a soft contact patch plus a cast shadow stretched away from the
// fixed sun. The ground samples it, so shade follows terrain relief and overlapping shadows
// saturate (max blending) instead of stacking.
export class PlantShadeMap {
  readonly target = new THREE.WebGLRenderTarget(RESOLUTION, RESOLUTION, {
    format: THREE.RedFormat,
    type: THREE.UnsignedByteType,
    depthBuffer: false,
    generateMipmaps: false,
    minFilter: THREE.LinearFilter,
    magFilter: THREE.LinearFilter,
  });
  readonly bounds = new THREE.Vector4(0, 0, 1, 1);
  private readonly scene = new THREE.Scene();
  private readonly camera = new THREE.OrthographicCamera(0, 1, 1, 0, 0.1, 10);
  private readonly geometry = new THREE.PlaneGeometry(1, 1);
  private readonly strengths = new THREE.InstancedBufferAttribute(new Float32Array(CAPACITY), 1);
  private readonly tapers = new THREE.InstancedBufferAttribute(new Float32Array(CAPACITY), 1);
  private readonly position = new THREE.Vector3();
  private readonly rotation = new THREE.Quaternion();
  private readonly scale = new THREE.Vector3();
  private readonly up = new THREE.Vector3(0, 0, 1);
  private readonly material = new THREE.ShaderMaterial({
    vertexShader,
    fragmentShader,
    blending: THREE.CustomBlending,
    blendEquation: THREE.MaxEquation,
    blendSrc: THREE.OneFactor,
    blendDst: THREE.OneFactor,
    depthTest: false,
    depthWrite: false,
  });
  private readonly mesh: THREE.InstancedMesh;
  private readonly matrix = new THREE.Matrix4();
  private readonly clearColor = new THREE.Color();
  private bakes = 0;
  contentLost = false;

  // Changes on every bake, so consumers can tell when the map content changed.
  get revision() {
    return this.bakes;
  }

  constructor() {
    this.geometry.setAttribute("shadeStrength", this.strengths);
    this.geometry.setAttribute("shadeTaper", this.tapers);
    this.mesh = new THREE.InstancedMesh(this.geometry, this.material, CAPACITY);
    this.mesh.frustumCulled = false;
    this.mesh.count = 0;
    this.scene.add(this.mesh);
    this.camera.position.set(0, 0, 1);
    this.camera.lookAt(0, 0, 0);
    liveShadeMaps.add(this);
  }

  // `shadowX`/`shadowY` is the unit board-plane direction shadows fall in.
  bake(gl: THREE.WebGLRenderer, blobs: PlantShadeBlob[], shadowX: number, shadowY: number) {
    const started = performance.now();
    let count = 0;
    let minX = Infinity,
      minY = Infinity,
      maxX = -Infinity,
      maxY = -Infinity;
    const angle = Math.atan2(shadowY, shadowX);
    for (const { x, y, radius, strength, length } of blobs) {
      if (count + 2 > CAPACITY) {
        break;
      }
      this.place(count++, x, y, 0, radius * 2, radius * 2, strength, 0);
      const reach = length + radius;
      const cx = x + shadowX * reach * 0.5,
        cy = y + shadowY * reach * 0.5;
      this.place(count++, cx, cy, angle, reach + radius, radius * 1.7, strength * 0.85, 0.75);
      const extent = reach * 0.5 + radius;
      minX = Math.min(minX, x - radius, cx - extent);
      minY = Math.min(minY, y - radius, cy - extent);
      maxX = Math.max(maxX, x + radius, cx + extent);
      maxY = Math.max(maxY, y + radius, cy + extent);
    }
    this.mesh.count = count;
    this.mesh.instanceMatrix.needsUpdate = true;
    this.strengths.needsUpdate = true;
    this.tapers.needsUpdate = true;
    if (count) {
      this.bounds.set(
        minX - MARGIN,
        minY - MARGIN,
        maxX - minX + MARGIN * 2,
        maxY - minY + MARGIN * 2,
      );
    }
    const size = Math.max(this.bounds.z, this.bounds.w);
    this.bounds.z = this.bounds.w = size;
    this.camera.left = this.bounds.x;
    this.camera.right = this.bounds.x + size;
    this.camera.bottom = this.bounds.y;
    this.camera.top = this.bounds.y + size;
    this.camera.updateProjectionMatrix();
    const previousTarget = gl.getRenderTarget();
    const previousAlpha = gl.getClearAlpha();
    gl.getClearColor(this.clearColor);
    gl.setRenderTarget(this.target);
    gl.setClearColor(0x000000, 1);
    gl.clear(true, false, false);
    gl.render(this.scene, this.camera);
    gl.setClearColor(this.clearColor, previousAlpha);
    gl.setRenderTarget(previousTarget);
    this.bakes++;
    this.contentLost = false;
    performanceStore.setSection("Plant shade", {
      bakes: this.bakes,
      "last bake (CPU ms)": Number((performance.now() - started).toFixed(2)),
      "shapes drawn": `${count} / ${CAPACITY}`,
    });
  }

  private place(
    index: number,
    x: number,
    y: number,
    angle: number,
    length: number,
    width: number,
    strength: number,
    taper: number,
  ) {
    this.position.set(x, y, 0);
    this.rotation.setFromAxisAngle(this.up, angle);
    this.scale.set(length, width, 1);
    this.matrix.compose(this.position, this.rotation, this.scale);
    this.mesh.setMatrixAt(index, this.matrix);
    this.strengths.setX(index, strength);
    this.tapers.setX(index, taper);
  }

  dispose() {
    liveShadeMaps.delete(this);
    this.target.dispose();
    this.geometry.dispose();
    this.material.dispose();
    this.mesh.dispose();
  }
}
