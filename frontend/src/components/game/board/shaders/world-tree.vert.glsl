#version 100
precision highp float;

uniform float uSphereRadius;
uniform float uTrunkHeight;
uniform float uEmergence;
uniform float uTime;
uniform float uSeed;

varying vec2 vUv;
varying float vDistFromCenter;
varying vec3 vWorldNormal;
varying vec3 vWorldPosition;
varying float vHeight;
varying float vEmergence;

float hash(float n) { return fract(sin(n) * 43758.5453123); }
float seedParam(float idx) { return hash(uSeed * 127.1 + idx * 311.7); }

vec3 mod289_v3(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec2 mod289_v2(vec2 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec3 permute_v(vec3 x) { return mod289_v3(((x * 34.0) + 1.0) * x); }

float snoise(vec2 v) {
  const vec4 C = vec4(0.211324865405187, 0.366025403784439,
                      -0.577350269189626, 0.024390243902439);
  vec2 i = floor(v + dot(v, C.yy));
  vec2 x0 = v - i + dot(i, C.xx);
  vec2 i1 = (x0.x > x0.y) ? vec2(1.0, 0.0) : vec2(0.0, 1.0);
  vec4 x12 = x0.xyxy + C.xxzz;
  x12.xy -= i1;
  i = mod289_v2(i);
  vec3 p = permute_v(permute_v(i.y + vec3(0.0, i1.y, 1.0)) + i.x + vec3(0.0, i1.x, 1.0));
  vec3 m = max(0.5 - vec3(dot(x0, x0), dot(x12.xy, x12.xy), dot(x12.zw, x12.zw)), 0.0);
  m = m * m;
  m = m * m;
  vec3 x_ = 2.0 * fract(p * C.www) - 1.0;
  vec3 h = abs(x_) - 0.5;
  vec3 ox = floor(x_ + 0.5);
  vec3 a0 = x_ - ox;
  m *= 1.79284291400159 - 0.85373472095314 * (a0 * a0 + h * h);
  vec3 g;
  g.x = a0.x * x0.x + h.x * x0.y;
  g.yz = a0.yz * x12.xz + h.yz * x12.yw;
  return 130.0 * dot(m, g);
}

float getHeight(vec2 centered) {
  float canopyH = uTrunkHeight * (0.85 + seedParam(0.0) * 0.3);
  float canopyRad = 0.65 + seedParam(1.0) * 0.15;
  vec2 centerOff = (vec2(seedParam(2.0), seedParam(3.0)) - 0.5) * 0.04;
  vec2 seedOff = vec2(seedParam(5.0), seedParam(6.0)) * 100.0;

  vec2 w = centered + centerOff;
  float dist = length(w);

  // === 1. Canopy dome — broad smooth hemisphere ===
  float canopyT = dist / canopyRad;
  float dome = max(0.0, 1.0 - canopyT * canopyT);
  dome = sqrt(dome);
  float h = dome * canopyH;

  // === 2. Leaf cluster bumps — organic canopy surface ===
  float leafN1 = snoise(w * 8.0 + seedOff);
  float leafN2 = snoise(w * 16.0 + seedOff + 7.3);
  float leafN3 = snoise(w * 28.0 + seedOff + 14.1);
  float leafBumps = leafN1 * 0.5 + leafN2 * 0.3 + leafN3 * 0.2;
  float canopyMask = smoothstep(canopyRad, canopyRad * 0.2, dist);
  h += leafBumps * canopyMask * 0.012;

  // === 3. Canopy edge irregularity — not a perfect circle ===
  float edgeNoise = snoise(w * 3.5 + seedOff + 21.0);
  float edgeWobble = edgeNoise * 0.06;
  float edgeDist = dist - (canopyRad + edgeWobble);
  float edgeFalloff = smoothstep(0.0, -0.15, edgeDist);
  h *= edgeFalloff;

  // === 4. Central peak — trunk crown pushing up ===
  float peakRad = 0.12 + seedParam(7.0) * 0.05;
  float peak = exp(-dist * dist / (peakRad * peakRad)) * canopyH * 0.25;
  h += peak * canopyMask;

  // === 5. Ground moss around canopy base ===
  float groundN = snoise(centered * 4.0 + seedOff + 9.1);
  float groundMask = smoothstep(canopyRad * 0.7, canopyRad * 1.2, dist)
                   * smoothstep(1.0, canopyRad * 0.9, dist);
  h += groundN * groundMask * 0.002;

  return h;
}

void main() {
  vUv = uv;

  vec4 worldPos = modelMatrix * vec4(position, 1.0);
  vec3 sphereDir = normalize(worldPos.xyz);

  vec2 centered = (uv - 0.5) * 2.0;
  vDistFromCenter = length(centered);

  float h = getHeight(centered) * uEmergence;
  vHeight = h;
  vEmergence = uEmergence;

  // Normal via finite differences
  float eps = 0.008;
  float hL = getHeight(centered + vec2(-eps, 0.0)) * uEmergence;
  float hR = getHeight(centered + vec2(eps, 0.0)) * uEmergence;
  float hD = getHeight(centered + vec2(0.0, -eps)) * uEmergence;
  float hU = getHeight(centered + vec2(0.0, eps)) * uEmergence;

  vec3 up = sphereDir;
  vec3 right;
  if (abs(up.y) < 0.99) {
    right = normalize(cross(vec3(0.0, 1.0, 0.0), up));
  } else {
    right = normalize(cross(vec3(1.0, 0.0, 0.0), up));
  }
  vec3 fwd = cross(up, right);

  float dhdx = (hR - hL) / (2.0 * eps);
  float dhdy = (hU - hD) / (2.0 * eps);

  float uvScale = 0.15;
  vWorldNormal = normalize(up * uvScale - right * dhdx - fwd * dhdy);

  vec3 projectedPos = sphereDir * (uSphereRadius + 0.003 + h);
  vWorldPosition = projectedPos;

  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
