#version 100
precision highp float;

uniform float uTime;
uniform vec3 uSunDirection;
uniform float uSunIntensity;
uniform vec3 uSunColor;
uniform float uTrunkHeight;
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

void main() {
  vec2 centered = (vUv - 0.5) * 2.0;
  float rawDist = vDistFromCenter;

  // === Per-tile parameters (must match vertex shader) ===
  float canopyH = uTrunkHeight * (0.85 + seedParam(0.0) * 0.3);
  float canopyRad = 0.65 + seedParam(1.0) * 0.15;
  vec2 centerOff = (vec2(seedParam(2.0), seedParam(3.0)) - 0.5) * 0.04;
  vec2 seedOff = vec2(seedParam(5.0), seedParam(6.0)) * 100.0;

  vec2 w = centered + centerOff;
  float dist = length(w);
  float normHeight = clamp(vHeight / canopyH, 0.0, 1.0);

  // === Canopy edge (must match vertex shader) ===
  float edgeNoise = snoise(w * 3.5 + seedOff + 21.0);
  float edgeWobble = edgeNoise * 0.06;
  float edgeDist = dist - (canopyRad + edgeWobble);
  float canopyMask = smoothstep(0.0, -0.15, edgeDist);

  // === Slope ===
  vec3 n = normalize(vWorldNormal);
  vec3 up = normalize(vWorldPosition);
  float slope = 1.0 - abs(dot(n, up));

  // === Foliage colors — rich greens with variation ===
  vec3 deepGreen = vec3(0.05, 0.20, 0.02);
  vec3 brightGreen = vec3(0.15, 0.45, 0.05);
  vec3 sunlitGreen = vec3(0.30, 0.55, 0.08);
  vec3 shadowGreen = vec3(0.03, 0.12, 0.02);

  // Leaf cluster noise at multiple scales
  float leafN1 = snoise(w * 8.0 + seedOff) * 0.5 + 0.5;
  float leafN2 = snoise(w * 16.0 + seedOff + 7.3) * 0.5 + 0.5;
  float leafN3 = snoise(w * 32.0 + seedOff + 14.1) * 0.5 + 0.5;
  float leafPattern = leafN1 * 0.5 + leafN2 * 0.3 + leafN3 * 0.2;

  // Blend greens based on leaf pattern and height
  vec3 foliageColor = mix(deepGreen, brightGreen, leafPattern);
  foliageColor = mix(foliageColor, sunlitGreen, normHeight * 0.6);
  foliageColor = mix(foliageColor, shadowGreen, (1.0 - leafPattern) * 0.3);

  // === Branch structure visible through canopy (subtle dark lines) ===
  float angle = atan(w.y, w.x);
  float branchFreq = 5.0 + seedParam(4.0) * 3.0;
  float branchLine = abs(sin(angle * branchFreq + seedParam(8.0) * 6.28));
  branchLine = pow(branchLine, 6.0);
  float branchRadial = smoothstep(0.05, 0.15, dist) * smoothstep(canopyRad * 0.9, 0.2, dist);
  float branchMask = branchLine * branchRadial;

  // Secondary smaller branches
  float subBranch = abs(sin(angle * branchFreq * 2.5 + seedParam(9.0) * 6.28 + dist * 8.0));
  subBranch = pow(subBranch, 8.0);
  float subBranchMask = subBranch * smoothstep(0.2, 0.35, dist) * smoothstep(canopyRad * 0.85, 0.3, dist);
  branchMask = max(branchMask * 0.35, subBranchMask * 0.15);

  // Darken foliage where branches are visible underneath
  vec3 branchColor = vec3(0.08, 0.14, 0.03);
  foliageColor = mix(foliageColor, branchColor, branchMask);

  // === Trunk peek — tiny brown center visible through canopy gap ===
  float trunkPeek = smoothstep(0.08, 0.02, dist);
  vec3 trunkColor = vec3(0.25, 0.15, 0.06);
  foliageColor = mix(foliageColor, trunkColor, trunkPeek * 0.4);

  // === Ground/moss around canopy edge ===
  vec3 mossColor = vec3(0.10, 0.25, 0.05);
  vec3 groundColor = vec3(0.15, 0.12, 0.07);
  float mossN = snoise(centered * 10.0 + seedOff + 15.7) * 0.5 + 0.5;
  vec3 outerColor = mix(groundColor, mossColor, mossN);
  float outerMask = smoothstep(-0.05, 0.1, edgeDist);

  // Final surface = canopy inside, ground outside
  vec3 surfaceColor = mix(foliageColor, outerColor, outerMask);

  // === Magical glow — energy pulsing through the canopy ===
  float veinN1 = snoise(w * 5.0 + seedOff * 0.1 + vec2(uTime * 0.06, -uTime * 0.04));
  float veinN2 = snoise(w * 10.0 + seedOff * 0.1 + vec2(-uTime * 0.05, uTime * 0.07) + 13.7);
  float vein = smoothstep(0.35, 0.65, veinN1 * 0.6 + veinN2 * 0.4);
  vein *= canopyMask;

  // Radial energy lines following branch structure
  float energyLine = pow(branchLine, 3.0) * branchRadial;
  float glowMask = max(vein * 0.3, energyLine * 0.5);

  // Crown glow at top
  float crownGlow = smoothstep(0.25, 0.03, dist) * normHeight;

  glowMask += crownGlow * 0.4;

  // Pulsing animation
  float pulse = 0.6 + 0.4 * sin(uTime * 0.7 + dist * 5.0);
  float pulse2 = 0.7 + 0.3 * sin(uTime * 1.1 + angle * 2.0 + 1.8);
  float combinedPulse = pulse * 0.6 + pulse2 * 0.4;

  // Glow color: golden-green energy
  vec3 glowInner = vec3(0.5, 1.0, 0.3);
  vec3 glowOuter = vec3(0.9, 0.8, 0.2);
  float glowBlend = smoothstep(0.1, 0.5, dist);
  vec3 glowColor = mix(glowInner, glowOuter, glowBlend);

  // === Lighting ===
  vec3 sunDir = normalize(uSunDirection);
  float NdotL = dot(n, sunDir);
  float wrapDiffuse = NdotL * 0.5 + 0.5;
  wrapDiffuse *= wrapDiffuse;
  vec3 lighting = vec3(0.35) + uSunColor * wrapDiffuse * 0.65 * uSunIntensity;

  // Subsurface scattering approximation for leaves
  float sss = max(0.0, -dot(n, sunDir)) * 0.15;
  lighting += vec3(0.1, 0.3, 0.05) * sss;

  // AO: darken canopy underside/edges
  float canopyAO = mix(0.6, 1.0, normHeight);
  float leafAO = 0.85 + leafPattern * 0.15;
  float ao = canopyAO * leafAO;

  // === Combine ===
  vec3 color = surfaceColor * lighting * ao;

  // Emissive glow
  float emissiveFade = smoothstep(0.3, 0.8, vEmergence);
  float glowIntensity = glowMask * combinedPulse * emissiveFade;
  color += glowColor * glowIntensity * 0.45;

  // Canopy rim light — bright green edge catch
  float rimAngle = 1.0 - max(0.0, dot(n, sunDir));
  float rimMask = pow(rimAngle, 3.0) * canopyMask * normHeight;
  color += vec3(0.1, 0.3, 0.05) * rimMask * 0.4 * emissiveFade;

  // Edge fade
  float edgeAlpha = smoothstep(1.0, 0.85, rawDist);

  gl_FragColor = vec4(color, edgeAlpha);
}
