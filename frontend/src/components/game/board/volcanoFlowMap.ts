import * as THREE from "three";

const GRID_SIZE = 64;

// --- Simplex noise port (matches GLSL exactly) ---

function hash(n: number): number {
  return (((Math.sin(n) * 43758.5453123) % 1) + 1) % 1;
}

function seedParam(seed: number, idx: number): number {
  return hash(seed * 127.1 + idx * 311.7);
}

function mod289(x: number): number {
  return x - Math.floor(x * (1.0 / 289.0)) * 289.0;
}

function permute(x: number): number {
  return mod289((x * 34.0 + 1.0) * x);
}

function snoise(vx: number, vy: number): number {
  const C0 = 0.211324865405187;
  const C1 = 0.366025403784439;
  const C2 = -0.577350269189626;
  const C3 = 0.024390243902439;

  const s = (vx + vy) * C1;
  const ix = Math.floor(vx + s);
  const iy = Math.floor(vy + s);

  const t = (ix + iy) * C0;
  const x0x = vx - ix + t;
  const x0y = vy - iy + t;

  const i1x = x0x > x0y ? 1.0 : 0.0;
  const i1y = x0x > x0y ? 0.0 : 1.0;

  const x12_x = x0x + C0 - i1x;
  const x12_y = x0y + C0 - i1y;
  const x12_z = x0x + C2;
  const x12_w = x0y + C2;

  const mi = mod289(ix);
  const miy = mod289(iy);

  const p1 = permute(permute(miy) + mi);
  const p2 = permute(permute(miy + i1y) + mi + i1x);
  const p3 = permute(permute(miy + 1.0) + mi + 1.0);

  let m0 = 0.5 - (x0x * x0x + x0y * x0y);
  let m1 = 0.5 - (x12_x * x12_x + x12_y * x12_y);
  let m2 = 0.5 - (x12_z * x12_z + x12_w * x12_w);

  m0 = m0 > 0 ? m0 : 0;
  m1 = m1 > 0 ? m1 : 0;
  m2 = m2 > 0 ? m2 : 0;

  m0 *= m0;
  m0 *= m0;
  m1 *= m1;
  m1 *= m1;
  m2 *= m2;
  m2 *= m2;

  const x_0 = 2.0 * ((p1 * C3) % 1) - 1.0;
  const x_1 = 2.0 * ((p2 * C3) % 1) - 1.0;
  const x_2 = 2.0 * ((p3 * C3) % 1) - 1.0;

  const h0 = Math.abs(x_0) - 0.5;
  const h1 = Math.abs(x_1) - 0.5;
  const h2 = Math.abs(x_2) - 0.5;

  const ox0 = Math.floor(x_0 + 0.5);
  const ox1 = Math.floor(x_1 + 0.5);
  const ox2 = Math.floor(x_2 + 0.5);

  const a0_0 = x_0 - ox0;
  const a0_1 = x_1 - ox1;
  const a0_2 = x_2 - ox2;

  m0 *= 1.79284291400159 - 0.85373472095314 * (a0_0 * a0_0 + h0 * h0);
  m1 *= 1.79284291400159 - 0.85373472095314 * (a0_1 * a0_1 + h1 * h1);
  m2 *= 1.79284291400159 - 0.85373472095314 * (a0_2 * a0_2 + h2 * h2);

  const g0 = a0_0 * x0x + h0 * x0y;
  const g1 = a0_1 * x12_x + h1 * x12_y;
  const g2 = a0_2 * x12_z + h2 * x12_w;

  return 130.0 * (m0 * g0 + m1 * g1 + m2 * g2);
}

function smoothstep(edge0: number, edge1: number, x: number): number {
  const t = Math.max(0, Math.min(1, (x - edge0) / (edge1 - edge0)));
  return t * t * (3 - 2 * t);
}

// --- Port of getHeight from vertex shader (must match exactly) ---

function getHeight(
  cx: number,
  cy: number,
  seed: number,
  uHeight: number,
  uCraterRadius: number,
  uCraterDepth: number,
): number {
  const coneHeight = uHeight * (0.85 + seedParam(seed, 0.0) * 0.3);
  const craterRad = uCraterRadius * (0.85 + seedParam(seed, 1.0) * 0.3);
  const craterDep = uCraterDepth * (0.8 + seedParam(seed, 2.0) * 0.4);
  const rimHeight = 0.012 + seedParam(seed, 3.0) * 0.012;
  const rimWidth = 0.06 + seedParam(seed, 4.0) * 0.04;
  const craterOffX = (seedParam(seed, 5.0) - 0.5) * 0.05;
  const craterOffY = (seedParam(seed, 6.0) - 0.5) * 0.05;
  const ellipX = 0.9 + seedParam(seed, 7.0) * 0.2;
  const ellipY = 0.9 + seedParam(seed, 8.0) * 0.2;
  const coneOffX = (seedParam(seed, 9.0) - 0.5) * 0.08;
  const coneOffY = (seedParam(seed, 10.0) - 0.5) * 0.08;
  const gullyFreq = 5.0 + seedParam(seed, 11.0) * 4.0;
  const seedOffX = seedParam(seed, 12.0) * 100.0;
  const seedOffY = seedParam(seed, 13.0) * 100.0;

  // Domain warp (faded near center to prevent twist)
  const wN1 = snoise(cx * 1.8 + seedOffX, cy * 1.8 + seedOffY);
  const wN2 = snoise(cx * 1.8 + seedOffX + 50.0, cy * 1.8 + seedOffY + 50.0);
  const rawDist = Math.sqrt(cx * cx + cy * cy);
  const warpMask = smoothstep(0.12, 0.35, rawDist);
  const wx = cx + coneOffX + wN1 * (0.12 * warpMask);
  const wy = cy + coneOffY + wN2 * (0.12 * warpMask);

  // Shape field to warp distance (breaks circular contours)
  const s1 = snoise(wx * 0.7 + seedOffX * 0.05, wy * 0.7 + seedOffY * 0.05);
  const s2 = snoise(wx * 0.35 + seedOffX * 0.05 + 19.1, wy * 0.35 + seedOffY * 0.05 + 19.1);
  const shape = s1 * 0.7 + s2 * 0.3;
  const dist = Math.sqrt(wx * wx + wy * wy) * (1.0 + shape * 0.18);

  // 1. Main cone
  const coneT = Math.max(0.0, 1.0 - dist);
  let h = Math.pow(coneT, 1.4) * coneHeight;

  // 2. Lumpy apron
  const apronBase = Math.exp(-dist * dist * 5.0) * coneHeight * 0.12;
  const apronNoise = snoise(cx * 2.0 + seedOffX + 200.0, cy * 2.0 + seedOffY + 200.0) * 0.5 + 0.5;
  const apronLump = 1.0 + 0.4 * (apronNoise - 0.5);
  h += apronBase * apronLump;

  // 3. Crater rim (cp-based noise, no angle)
  const cpx = (wx - craterOffX) * ellipX;
  const cpy = (wy - craterOffY) * ellipY;
  const cd = Math.sqrt(cpx * cpx + cpy * cpy);
  const rimT = (cd - craterRad) / rimWidth;
  let rim = Math.exp(-rimT * rimT) * rimHeight;
  const rimNoise = snoise(cpx * 3.0 + seedOffX * 0.05, cpy * 3.0 + seedOffY * 0.05) * 0.4 + 0.8;
  rim *= rimNoise * smoothstep(0.8, 0.4, dist);
  h += rim;

  // 4. Crater bowl
  const bowlT = cd / Math.max(craterRad, 0.001);
  let bowl = 1.0 - smoothstep(0.0, 1.0, bowlT);
  bowl *= bowl;
  h -= bowl * craterDep;

  // 5. Gullies (seam-free, no angle)
  const g1 = snoise(wx * gullyFreq + seedOffX * 0.05, wy * gullyFreq + seedOffY * 0.05);
  const g2 = snoise(
    wx * (gullyFreq * 2.0) + seedOffX * 0.05 + 31.7,
    wy * (gullyFreq * 2.0) + seedOffY * 0.05 + 31.7,
  );
  let ridge = 1.0 - Math.abs(g1 * 0.8 + g2 * 0.2);
  ridge = Math.pow(ridge, 2.0);
  const gs = smoothstep(0.18, 0.32, dist) * smoothstep(0.88, 0.58, dist);
  h -= ridge * gs * 0.03;

  // 6. Rocky noise
  const n1 = snoise(cx * 8.0 + seedOffX + 3.7, cy * 8.0 + seedOffY + 3.7);
  const n2 = snoise(cx * 16.0 + seedOffX + 11.3, cy * 16.0 + seedOffY + 11.3);
  h += (n1 * 0.65 + n2 * 0.35) * 0.004 * Math.max(0.0, 1.0 - dist * 0.7);

  return h;
}

// --- Flow accumulation ---

export function computeFlowMap(
  seed: number,
  uHeight = 0.14,
  uCraterRadius = 0.22,
  uCraterDepth = 0.08,
): THREE.DataTexture {
  const size = GRID_SIZE;
  const heights = new Float32Array(size * size);
  const flow = new Float32Array(size * size);

  // Sample height field on grid (centered UV mapped to [-1, 1])
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const cx = ((x + 0.5) / size - 0.5) * 2.0;
      const cy = ((y + 0.5) / size - 0.5) * 2.0;
      heights[y * size + x] = getHeight(cx, cy, seed, uHeight, uCraterRadius, uCraterDepth);
    }
  }

  // Find crater rim cells as flow sources
  const craterRad = uCraterRadius * (0.85 + seedParam(seed, 1.0) * 0.3);
  const sources: number[] = [];

  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const cx = ((x + 0.5) / size - 0.5) * 2.0;
      const cy = ((y + 0.5) / size - 0.5) * 2.0;
      const dist = Math.sqrt(cx * cx + cy * cy);
      if (Math.abs(dist - craterRad) < 0.08) {
        sources.push(y * size + x);
      }
    }
  }

  // Sort all cells by height descending for flow accumulation
  const indices = new Uint32Array(size * size);
  for (let i = 0; i < size * size; i++) indices[i] = i;
  indices.sort((a, b) => heights[b] - heights[a]);

  // Initialize flow at source cells
  for (const src of sources) {
    flow[src] += 1.0;
  }

  // 8-neighbor offsets
  const dx = [-1, 0, 1, -1, 1, -1, 0, 1];
  const dy = [-1, -1, -1, 0, 0, 1, 1, 1];

  // Distribute flow downhill (steepest descent)
  for (let i = 0; i < indices.length; i++) {
    const idx = indices[i];
    if (flow[idx] <= 0) continue;

    const ix = idx % size;
    const iy = (idx - ix) / size;
    const h = heights[idx];

    let bestSlope = 0;
    let bestIdx = -1;

    for (let d = 0; d < 8; d++) {
      const nx = ix + dx[d];
      const ny = iy + dy[d];
      if (nx < 0 || nx >= size || ny < 0 || ny >= size) continue;
      const nIdx = ny * size + nx;
      const dh = h - heights[nIdx];
      const neighborDist = Math.abs(dx[d]) + Math.abs(dy[d]) === 2 ? 1.414 : 1.0;
      const slope = dh / neighborDist;
      if (slope > bestSlope) {
        bestSlope = slope;
        bestIdx = nIdx;
      }
    }

    if (bestIdx >= 0) {
      flow[bestIdx] += flow[idx];
    }
  }

  // Normalize flow
  let maxFlow = 0;
  for (let i = 0; i < flow.length; i++) {
    if (flow[i] > maxFlow) maxFlow = flow[i];
  }
  if (maxFlow > 0) {
    const invMax = 1.0 / maxFlow;
    for (let i = 0; i < flow.length; i++) {
      flow[i] *= invMax;
    }
  }

  // Gaussian blur (3x3, 2 passes)
  const tmp = new Float32Array(size * size);
  for (let pass = 0; pass < 2; pass++) {
    const src = pass === 0 ? flow : tmp;
    const dst = pass === 0 ? tmp : flow;

    for (let y = 0; y < size; y++) {
      for (let x = 0; x < size; x++) {
        let sum = 0;
        let weight = 0;
        for (let dy2 = -1; dy2 <= 1; dy2++) {
          for (let dx2 = -1; dx2 <= 1; dx2++) {
            const nx = x + dx2;
            const ny = y + dy2;
            if (nx < 0 || nx >= size || ny < 0 || ny >= size) continue;
            const w = dx2 === 0 && dy2 === 0 ? 4 : Math.abs(dx2) + Math.abs(dy2) === 1 ? 2 : 1;
            sum += src[ny * size + nx] * w;
            weight += w;
          }
        }
        dst[y * size + x] = sum / weight;
      }
    }
  }

  // Re-normalize after blur
  maxFlow = 0;
  for (let i = 0; i < flow.length; i++) {
    if (flow[i] > maxFlow) maxFlow = flow[i];
  }
  if (maxFlow > 0) {
    const invMax = 1.0 / maxFlow;
    for (let i = 0; i < flow.length; i++) {
      flow[i] *= invMax;
    }
  }

  // Create DataTexture (RGBA)
  const data = new Uint8Array(size * size * 4);
  for (let i = 0; i < size * size; i++) {
    const v = Math.min(255, Math.floor(flow[i] * 255));
    data[i * 4] = v;
    data[i * 4 + 1] = v;
    data[i * 4 + 2] = v;
    data[i * 4 + 3] = 255;
  }

  const texture = new THREE.DataTexture(data, size, size, THREE.RGBAFormat);
  texture.wrapS = THREE.ClampToEdgeWrapping;
  texture.wrapT = THREE.ClampToEdgeWrapping;
  texture.magFilter = THREE.LinearFilter;
  texture.minFilter = THREE.LinearFilter;
  texture.needsUpdate = true;

  return texture;
}
