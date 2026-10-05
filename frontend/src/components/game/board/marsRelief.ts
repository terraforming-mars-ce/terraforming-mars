import * as THREE from "three";
import { SPHERE_RADIUS } from "./boardConstants";

export const MARS_RELIEF_DEPTH = 0.036;
const WIDTH = 512;
const HEIGHT = 256;
export interface MarsRelief {
  width: number;
  height: number;
  depths: Float32Array;
  revision: number;
}
const cache = new WeakMap<THREE.Texture, MarsRelief>();

export function createMarsRelief(
  rgba: Uint8ClampedArray,
  width: number,
  height: number,
): MarsRelief {
  const red = new Float32Array(width * height);
  for (let i = 0; i < red.length; i++) {
    const value = rgba[i * 4] / 255;
    red[i] = value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  }
  const depths = new Float32Array(red.length);
  const kernel = [1, 2, 1];
  let revision = 2166136261;
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      let smooth = 0;
      for (let dy = -1; dy <= 1; dy++) {
        const row = Math.max(0, Math.min(height - 1, y + dy));
        for (let dx = -1; dx <= 1; dx++) {
          const col = (x + dx + width) % width;
          smooth += (red[row * width + col] * kernel[dx + 1] * kernel[dy + 1]) / 16;
        }
      }
      const depth = MARS_RELIEF_DEPTH * (1 - Math.pow(Math.min(1, smooth / 0.65), 0.85));
      depths[y * width + x] = depth;
      revision = Math.imul(revision ^ Math.round(depth * 1e7), 16777619);
    }
  }
  // All longitudes meet at each pole.
  for (const row of [0, height - 1]) {
    let mean = 0;
    for (let x = 0; x < width; x++) {
      mean += depths[row * width + x] / width;
    }
    depths.fill(mean, row * width, (row + 1) * width);
  }
  return { width, height, depths, revision: revision >>> 0 };
}

export function marsReliefFromTexture(texture: THREE.Texture): MarsRelief {
  const cached = cache.get(texture);
  if (cached) {
    return cached;
  }
  const canvas = new OffscreenCanvas(WIDTH, HEIGHT);
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) {
    throw new Error("Cannot sample the Mars surface for terrain relief");
  }
  context.drawImage(texture.image as CanvasImageSource, 0, 0, WIDTH, HEIGHT);
  const relief = createMarsRelief(context.getImageData(0, 0, WIDTH, HEIGHT).data, WIDTH, HEIGHT);
  cache.set(texture, relief);
  return relief;
}

export function marsReliefAtUv(relief: MarsRelief, u: number, v: number) {
  const { width, height, depths } = relief;
  const x = (((u % 1) + 1) % 1) * width - 0.5;
  const y = Math.max(0, Math.min(height - 1, (1 - v) * height - 0.5));
  const ix = Math.floor(x),
    iy = Math.floor(y);
  const x0 = (ix + width) % width,
    x1 = (x0 + 1) % width;
  const y1 = Math.min(height - 1, iy + 1);
  const tx = x - ix,
    ty = y - iy;
  const a = depths[iy * width + x0] * (1 - tx) + depths[iy * width + x1] * tx;
  const b = depths[y1 * width + x0] * (1 - tx) + depths[y1 * width + x1] * tx;
  return -(a * (1 - ty) + b * ty);
}

export function marsReliefAtDirection(relief: MarsRelief, x: number, y: number, z: number) {
  const length = Math.hypot(x, y, z);
  const u = Math.atan2(z, -x) / (Math.PI * 2);
  const v = 1 - Math.acos(Math.max(-1, Math.min(1, y / length))) / Math.PI;
  return marsReliefAtUv(relief, u, v);
}

export function marsReliefAtBoard(relief: MarsRelief | undefined, x: number, y: number) {
  if (!relief) {
    return 0;
  }
  const d = Math.hypot(x, y);
  const scale = d > 1e-10 ? Math.sin(d / SPHERE_RADIUS) / d : 0;
  return marsReliefAtDirection(relief, x * scale, y * scale, Math.cos(d / SPHERE_RADIUS));
}

export function createMarsReliefGeometry(relief: MarsRelief) {
  const geometry = new THREE.SphereGeometry(SPHERE_RADIUS, WIDTH, HEIGHT);
  const positions = geometry.getAttribute("position");
  const point = new THREE.Vector3();
  for (let i = 0; i < positions.count; i++) {
    point.fromBufferAttribute(positions, i).normalize();
    const radius = SPHERE_RADIUS + marsReliefAtDirection(relief, point.x, point.y, point.z);
    positions.setXYZ(i, point.x * radius, point.y * radius, point.z * radius);
  }
  geometry.computeVertexNormals();
  // Join duplicated seam vertices so lighting is continuous at longitude zero.
  const normals = geometry.getAttribute("normal");
  for (let row = 0; row <= HEIGHT; row++) {
    const a = row * (WIDTH + 1),
      b = a + WIDTH;
    point
      .set(
        normals.getX(a) + normals.getX(b),
        normals.getY(a) + normals.getY(b),
        normals.getZ(a) + normals.getZ(b),
      )
      .normalize();
    normals.setXYZ(a, point.x, point.y, point.z);
    normals.setXYZ(b, point.x, point.y, point.z);
  }
  geometry.computeBoundingSphere();
  return geometry;
}

export function createMarsReliefRaycast(relief: MarsRelief) {
  const inverse = new THREE.Matrix4();
  const ray = new THREE.Ray();
  const shell = new THREE.Sphere(new THREE.Vector3(), SPHERE_RADIUS);
  const point = new THREE.Vector3();
  const end = new THREE.Vector3();
  const signedDistance = (distance: number) => {
    ray.at(distance, point);
    return (
      point.length() - SPHERE_RADIUS - marsReliefAtDirection(relief, point.x, point.y, point.z)
    );
  };
  return function (
    this: THREE.Mesh,
    raycaster: THREE.Raycaster,
    intersections: THREE.Intersection[],
  ) {
    inverse.copy(this.matrixWorld).invert();
    ray.copy(raycaster.ray).applyMatrix4(inverse);
    shell.radius = SPHERE_RADIUS;
    if (!ray.intersectSphere(shell, point)) {
      return;
    }
    let low = ray.origin.distanceTo(point);
    shell.radius = SPHERE_RADIUS - MARS_RELIEF_DEPTH;
    let high = ray.intersectSphere(shell, end)
      ? ray.origin.distanceTo(end)
      : -ray.origin.dot(ray.direction);
    if (high < low) {
      return;
    }
    // Bracket the first visible crossing; grazing rays may miss the inner shell entirely.
    const start = low,
      span = high - low;
    let found = false;
    for (let step = 0; step <= 32; step++) {
      high = start + (span * step) / 32;
      if (signedDistance(high) <= 1e-8) {
        found = true;
        break;
      }
      low = high;
    }
    if (!found) {
      return;
    }
    for (let step = 0; step < 20; step++) {
      const middle = (low + high) / 2;
      if (signedDistance(middle) > 0) {
        low = middle;
      } else {
        high = middle;
      }
    }
    ray.at((low + high) / 2, point).applyMatrix4(this.matrixWorld);
    const distance = raycaster.ray.origin.distanceTo(point);
    if (distance >= raycaster.near && distance <= raycaster.far) {
      intersections.push({ distance, point: point.clone(), object: this });
    }
  };
}
