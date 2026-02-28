#version 100
precision highp float;

uniform float uTime;
uniform vec3 uSunDirection;
uniform float uSunIntensity;
uniform vec3 uSunColor;
uniform float uCraterRadius;
uniform float uCraterDepth;
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
  float craterRad = uCraterRadius * (0.85 + seedParam(0.0) * 0.3);
  vec2 craterOff = (vec2(seedParam(4.0), seedParam(5.0)) - 0.5) * 0.03;
  float ellipX = 0.92 + seedParam(6.0) * 0.16;
  float ellipY = 0.92 + seedParam(7.0) * 0.16;
  vec2 seedOff = vec2(seedParam(8.0), seedParam(9.0)) * 100.0;

  // === Crater geometry ===
  vec2 cp = (centered - craterOff) * vec2(ellipX, ellipY);
  float craterDist = length(cp);
  float dist = length(centered);

  // === Surface noise for variation ===
  float rockN1 = snoise(centered * 10.0 + seedOff + 5.3) * 0.5 + 0.5;
  float rockN2 = snoise(centered * 20.0 + seedOff + 9.7) * 0.5 + 0.5;
  float rockNoise = rockN1 * 0.65 + rockN2 * 0.35;

  // === Three-zone color scheme ===
  vec3 charredGround = vec3(0.08, 0.06, 0.04);
  vec3 scorchedEarth = vec3(0.25, 0.10, 0.06);
  vec3 disturbedSoil = vec3(0.40, 0.22, 0.12);

  // Blend zones based on distance from crater center
  float innerZone = smoothstep(craterRad * 0.6, craterRad * 0.2, craterDist);
  float midZone = smoothstep(craterRad * 1.5, craterRad * 0.7, craterDist)
                * (1.0 - innerZone);
  float outerZone = 1.0 - innerZone - midZone;

  vec3 surfaceColor = charredGround * innerZone
                    + scorchedEarth * midZone
                    + disturbedSoil * outerZone;

  // Add noise variation
  surfaceColor *= 0.85 + rockNoise * 0.3;

  // === Slope-based darkening ===
  vec3 n = normalize(vWorldNormal);
  vec3 up = normalize(vWorldPosition);
  float slope = 1.0 - abs(dot(n, up));
  surfaceColor *= mix(0.8, 1.0, 1.0 - slope * 0.5);

  // === Radioactive glow — pulsing green-yellow emissive in crater center ===
  float glowMask = smoothstep(craterRad * 0.9, craterRad * 0.1, craterDist);
  glowMask *= glowMask;
  float glowPulse = 0.5 + 0.5 * sin(uTime * 1.8 + snoise(centered * 3.0 + seedOff) * 2.0);
  float glowDetail = snoise(centered * 5.0 + vec2(uTime * 0.15, -uTime * 0.1) + seedOff) * 0.3 + 0.7;
  vec3 radioactiveColor = mix(vec3(0.15, 0.8, 0.1), vec3(0.6, 0.9, 0.1), glowPulse * 0.5);
  float glowIntensity = glowMask * glowDetail * (0.3 + glowPulse * 0.4);

  // === Lighting: wrap diffuse + AO (same pattern as volcano) ===
  vec3 sunDir = normalize(uSunDirection);
  float NdotL = dot(n, sunDir);
  float wrapDiffuse = NdotL * 0.5 + 0.5;
  wrapDiffuse *= wrapDiffuse;
  vec3 lighting = vec3(0.3) + uSunColor * wrapDiffuse * 0.7 * uSunIntensity;

  // AO: darken crater interior
  float craterAO = mix(0.35, 1.0, smoothstep(0.0, craterRad * 1.2, craterDist));
  float slopeAO = 1.0 + slope * 0.12;
  float ao = craterAO * slopeAO;

  // === Combine ===
  vec3 color = surfaceColor * lighting * ao;

  // Radioactive emissive (ignores lighting, scales with emergence)
  float emissiveFade = smoothstep(0.3, 0.8, vEmergence);
  color += radioactiveColor * glowIntensity * 1.5 * emissiveFade;

  // Faint rim glow (green tint at crater edge)
  float rimGlow = smoothstep(craterRad * 1.4, craterRad, craterDist)
                * (1.0 - smoothstep(0.0, craterRad * 0.7, craterDist));
  color += vec3(0.1, 0.4, 0.05) * rimGlow * 0.2 * glowPulse * emissiveFade;

  // Edge fade for soft blending with ground
  float edgeAlpha = smoothstep(1.0, 0.85, rawDist);

  gl_FragColor = vec4(color, edgeAlpha);
}
