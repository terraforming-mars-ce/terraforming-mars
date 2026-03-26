#version 100
precision highp float;

uniform float uSphereRadius;
uniform vec3 uSphereCenter;
uniform float uCraterDepth;
uniform float uCraterRadius;
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
  float craterRad = uCraterRadius * (0.85 + seedParam(0.0) * 0.3);
  float craterDep = uCraterDepth * (0.8 + seedParam(1.0) * 0.4);
  float rimHeight = 0.008 + seedParam(2.0) * 0.008;
  float rimWidth = 0.08 + seedParam(3.0) * 0.04;
  vec2 craterOff = (vec2(seedParam(4.0), seedParam(5.0)) - 0.5) * 0.03;
  float ellipX = 0.92 + seedParam(6.0) * 0.16;
  float ellipY = 0.92 + seedParam(7.0) * 0.16;
  vec2 seedOff = vec2(seedParam(8.0), seedParam(9.0)) * 100.0;

  float dist = length(centered);

  // === 1. Crater bowl — smooth concave depression ===
  vec2 cp = (centered - craterOff) * vec2(ellipX, ellipY);
  float cd = length(cp);
  float bowlT = cd / max(craterRad, 0.001);
  float bowl = 1.0 - smoothstep(0.0, 1.0, bowlT);
  bowl *= bowl;
  float h = -bowl * craterDep;

  // === 2. Raised rim — subtle ring at crater edge ===
  float rimT = (cd - craterRad) / rimWidth;
  float rim = exp(-rimT * rimT) * rimHeight;
  float rimNoise = snoise(cp * 3.0 + seedOff * 0.05) * 0.3 + 0.85;
  rim *= rimNoise;
  h += rim;

  // === 3. Rubble noise inside crater ===
  float rubbleN = snoise(centered * 12.0 + seedOff + 3.7);
  float rubbleN2 = snoise(centered * 24.0 + seedOff + 11.3);
  float rubble = (rubbleN * 0.6 + rubbleN2 * 0.4) * 0.003;
  float rubbleMask = smoothstep(craterRad * 1.1, craterRad * 0.3, cd);
  h += rubble * rubbleMask;

  // === 4. Radial cracks in polar coordinates ===
  float angle = atan(centered.y, centered.x);
  float crackFreq = 6.0 + seedParam(10.0) * 4.0;
  float crack = abs(sin(angle * crackFreq + seedParam(11.0) * 6.28));
  crack = pow(crack, 8.0);
  float crackMask = smoothstep(craterRad * 0.5, craterRad * 1.2, cd)
                  * smoothstep(craterRad * 2.0, craterRad * 1.3, cd);
  h -= crack * crackMask * 0.004;

  // === 5. Surface noise outside crater ===
  float surfNoise = snoise(centered * 6.0 + seedOff + 7.7);
  float surfMask = smoothstep(craterRad * 0.8, craterRad * 1.5, cd);
  h += surfNoise * 0.002 * surfMask;

  return h;
}

void main() {
  vUv = uv;

  vec4 worldPos = modelMatrix * vec4(position, 1.0);
  vec3 sphereDir = normalize(worldPos.xyz - uSphereCenter);

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

  vec3 projectedPos = uSphereCenter + sphereDir * (uSphereRadius + 0.003 + h);
  vWorldPosition = projectedPos;

  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
