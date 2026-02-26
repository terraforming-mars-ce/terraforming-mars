#version 100
precision highp float;

uniform float uSphereRadius;
uniform float uHeight;
uniform float uCraterRadius;
uniform float uCraterDepth;
uniform float uEmergence;
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
  // === Per-tile parameters derived from seed ===
  float coneHeight = uHeight * (0.85 + seedParam(0.0) * 0.3);
  float craterRad = uCraterRadius * (0.85 + seedParam(1.0) * 0.3);
  float craterDep = uCraterDepth * (0.8 + seedParam(2.0) * 0.4);
  float rimHeight = 0.012 + seedParam(3.0) * 0.012;
  float rimWidth = 0.06 + seedParam(4.0) * 0.04;
  vec2 craterOff = (vec2(seedParam(5.0), seedParam(6.0)) - 0.5) * 0.05;
  float ellipX = 0.9 + seedParam(7.0) * 0.2;
  float ellipY = 0.9 + seedParam(8.0) * 0.2;
  vec2 coneOff = (vec2(seedParam(9.0), seedParam(10.0)) - 0.5) * 0.08;
  float gullyFreq = 5.0 + seedParam(11.0) * 4.0;
  vec2 seedOff = vec2(seedParam(12.0), seedParam(13.0)) * 100.0;

  // === Domain warp (faded near center to prevent twist) ===
  float wN1 = snoise(centered * 1.8 + seedOff);
  float wN2 = snoise(centered * 1.8 + seedOff + vec2(50.0, 50.0));
  float rawDist = length(centered);
  float warpMask = smoothstep(0.12, 0.35, rawDist);
  vec2 w = centered + coneOff + vec2(wN1, wN2) * (0.12 * warpMask);

  // === Shape field to warp distance (breaks circular contours) ===
  float s1 = snoise(w * 0.7 + seedOff * 0.05);
  float s2 = snoise(w * 0.35 + seedOff * 0.05 + 19.1);
  float shape = s1 * 0.7 + s2 * 0.3;
  float dist = length(w) * (1.0 + shape * 0.18);

  // === 1. Main cone (concave stratovolcano profile) ===
  float coneT = max(0.0, 1.0 - dist);
  float h = pow(coneT, 1.4) * coneHeight;

  // === 2. Lumpy non-radial apron ===
  float apronBase = exp(-dist * dist * 5.0) * coneHeight * 0.12;
  float apronNoise = snoise(centered * 2.0 + seedOff + vec2(200.0, 200.0)) * 0.5 + 0.5;
  float apronLump = 1.0 + 0.4 * (apronNoise - 0.5);
  h += apronBase * apronLump;

  // === 3. Crater rim (raised irregular ring) ===
  vec2 cp = (w - craterOff) * vec2(ellipX, ellipY);
  float cd = length(cp);
  float rimT = (cd - craterRad) / rimWidth;
  float rim = exp(-rimT * rimT) * rimHeight;
  float rimNoise = snoise(cp * 3.0 + seedOff * 0.05) * 0.4 + 0.8;
  rim *= rimNoise * smoothstep(0.8, 0.4, dist);
  h += rim;

  // === 4. Crater bowl (elliptical, off-center) ===
  float bowlT = cd / max(craterRad, 0.001);
  float bowl = 1.0 - smoothstep(0.0, 1.0, bowlT);
  bowl *= bowl;
  h -= bowl * craterDep;

  // === 5. Erosion gullies (seam-free, no angle) ===
  float g1 = snoise(w * gullyFreq + seedOff * 0.05);
  float g2 = snoise(w * (gullyFreq * 2.0) + seedOff * 0.05 + 31.7);
  float ridge = 1.0 - abs(g1 * 0.8 + g2 * 0.2);
  ridge = pow(ridge, 2.0);
  float gs = smoothstep(0.18, 0.32, dist) * smoothstep(0.88, 0.58, dist);
  h -= ridge * gs * 0.03;

  // === 6. Rocky surface noise ===
  float n1 = snoise(centered * 8.0 + seedOff + 3.7);
  float n2 = snoise(centered * 16.0 + seedOff + 11.3);
  h += (n1 * 0.65 + n2 * 0.35) * 0.004 * max(0.0, 1.0 - dist * 0.7);

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
