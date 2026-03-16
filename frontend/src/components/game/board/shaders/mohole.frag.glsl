#version 100
precision highp float;

uniform float uTime;
uniform vec3 uSunDirection;
uniform float uSunIntensity;
uniform vec3 uSunColor;
uniform float uHoleRadius;
uniform float uHoleDepth;
uniform float uSeed;
uniform float uEmergenceRadius;
uniform sampler2D uConcreteTexture;

varying vec2 vUv;
varying vec2 vCentered;
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
  float dist = vDistFromCenter;
  vec2 seedOff = vec2(seedParam(0.0), seedParam(1.0)) * 100.0;

  // Extend past hole edge to overlap with flat rim (avoids gap from sphere projection)
  float effectiveRadius = uHoleRadius * uEmergenceRadius;
  if (dist >= effectiveRadius + 0.1) {
    discard;
  }

  vec3 n = normalize(vWorldNormal);
  vec3 sunDir = normalize(uSunDirection);
  float maxDepth = uHoleDepth * vEmergence;

  // Visual depth based on distance from hole edge (for texturing, not geometry)
  float depthRatio = clamp(1.0 - dist / effectiveRadius, 0.0, 1.0);

  // === Hole walls — concrete building blocks ===
  float NdotL = dot(n, sunDir);
  float wrapDiffuse = NdotL * 0.5 + 0.5;
  wrapDiffuse *= wrapDiffuse;
  vec3 lighting = vec3(0.2) + uSunColor * wrapDiffuse * 0.7 * uSunIntensity;

  // Cylindrical unwrap: angle for horizontal, vertex height for vertical
  float angle = atan(vCentered.y, vCentered.x);
  float u = angle / 6.283 + 0.5;
  float v = clamp(-vHeight / (maxDepth + 0.0001), 0.0, 1.0);

  // Block grid — rows and columns of concrete panels
  float blockCols = 12.0;
  float blockRows = 3.0;
  float colCell = fract(u * blockCols);
  float rowCell = fract(v * blockRows);

  // Mortar lines (thin dark gaps between blocks)
  float mortarW = 0.05;
  float hMortar = smoothstep(0.0, mortarW, colCell) * smoothstep(0.0, mortarW, 1.0 - colCell);
  float vMortar = smoothstep(0.0, mortarW, rowCell) * smoothstep(0.0, mortarW, 1.0 - rowCell);
  float mortar = hMortar * vMortar;

  // Concrete texture per block
  vec2 wallUV = vec2(u * blockCols, v * blockRows);
  vec3 concreteSample = texture2D(uConcreteTexture, wallUV * 0.5 + seedOff * 0.01).rgb;

  // Concrete texture — match ground darkness
  vec3 blockColor = concreteSample * mix(vec3(0.6, 0.56, 0.52), vec3(0.45, 0.42, 0.39), v * 0.5);
  vec3 mortarColor = vec3(0.3, 0.28, 0.26);

  vec3 wallColor = mix(mortarColor, blockColor, mortar);
  wallColor *= lighting;

  // AO: fade to black deeper in the shaft
  wallColor *= mix(1.0, 0.0, pow(v, 1.5));

  // Subtle heat haze near top
  float shimmer = sin(uTime * 2.0 + snoise(vCentered * 4.0 + seedOff) * 3.0) * 0.5 + 0.5;
  float shimmerMask = (1.0 - depthRatio) * 0.3;
  float emissiveFade = smoothstep(0.3, 0.8, vEmergence);
  wallColor += vec3(0.3, 0.12, 0.03) * shimmer * shimmerMask * 0.06 * emissiveFade;

  // Faint orange glow from the center depths, randomly breathing
  float deepMask = smoothstep(0.6, 1.0, depthRatio);
  float breath1 = snoise(vec2(uTime * 0.6, uSeed * 3.0));
  float breath2 = snoise(vec2(uTime * 0.3, uSeed * 7.0 + 50.0));
  float breath = 0.5 + 0.3 * breath1 + 0.2 * breath2;
  float glowRadius = effectiveRadius * (0.25 + 0.25 * breath);
  float centerMask = 1.0 - smoothstep(0.0, glowRadius, dist);
  float glowIntensity = centerMask * deepMask * (0.6 + 0.4 * breath) * emissiveFade;
  wallColor += vec3(1.0, 0.4, 0.08) * glowIntensity;

  gl_FragColor = vec4(wallColor, 1.0);
}
