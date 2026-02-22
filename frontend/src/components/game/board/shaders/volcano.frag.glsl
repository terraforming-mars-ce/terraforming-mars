#version 100
precision highp float;

uniform float uTime;
uniform vec3 uSunDirection;
uniform float uSunIntensity;
uniform vec3 uSunColor;
uniform float uHeight;
uniform float uCraterRadius;
uniform float uCraterDepth;
uniform sampler2D uGrassTexture;
uniform sampler2D uFlowTex;
uniform float uSeed;
uniform int uDebugMode;  // 0=normal, 1=height, 2=slope, 3=gully, 4=flow

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

  // === Same per-tile parameters as vertex shader ===
  float coneHeight = uHeight * (0.85 + seedParam(0.0) * 0.3);
  float craterRad = uCraterRadius * (0.85 + seedParam(1.0) * 0.3);
  vec2 craterOff = (vec2(seedParam(5.0), seedParam(6.0)) - 0.5) * 0.05;
  float ellipX = 0.9 + seedParam(7.0) * 0.2;
  float ellipY = 0.9 + seedParam(8.0) * 0.2;
  vec2 coneOff = (vec2(seedParam(9.0), seedParam(10.0)) - 0.5) * 0.08;
  float gullyFreq = 5.0 + seedParam(11.0) * 4.0;
  vec2 seedOff = vec2(seedParam(12.0), seedParam(13.0)) * 100.0;

  // === Domain warp (must match vertex shader) ===
  float wN1 = snoise(centered * 1.8 + seedOff);
  float wN2 = snoise(centered * 1.8 + seedOff + vec2(50.0, 50.0));
  float rawDistF = length(centered);
  float warpMask = smoothstep(0.12, 0.35, rawDistF);
  vec2 w = centered + coneOff + vec2(wN1, wN2) * (0.12 * warpMask);

  // === Shape field (must match vertex shader) ===
  float s1 = snoise(w * 0.7 + seedOff * 0.05);
  float s2 = snoise(w * 0.35 + seedOff * 0.05 + 19.1);
  float shape = s1 * 0.7 + s2 * 0.3;
  float dist = length(w) * (1.0 + shape * 0.18);

  // === Slope and height for material blending ===
  vec3 n = normalize(vWorldNormal);
  vec3 up = normalize(vWorldPosition);
  float slope = 1.0 - abs(dot(n, up));
  float normHeight = clamp(vHeight / coneHeight, 0.0, 1.0);

  // === Gully pattern (must match vertex shader — XY space, no angle) ===
  float g1 = snoise(w * gullyFreq + seedOff * 0.05);
  float g2 = snoise(w * (gullyFreq * 2.0) + seedOff * 0.05 + 31.7);
  float gridge = 1.0 - abs(g1 * 0.8 + g2 * 0.2);
  gridge = pow(gridge, 2.0);
  float gs = smoothstep(0.18, 0.32, dist) * smoothstep(0.88, 0.58, dist);
  float gullyField = gridge * gs;

  // === Rock material: 3 layers by height and slope ===
  vec3 basalt = vec3(0.22, 0.08, 0.05);
  vec3 ashRock = vec3(0.40, 0.16, 0.10);
  vec3 dustRock = vec3(0.52, 0.25, 0.14);

  float rockN1 = snoise(centered * 10.0 + seedOff + 5.3) * 0.5 + 0.5;
  float rockN2 = snoise(centered * 20.0 + seedOff + 9.7) * 0.5 + 0.5;
  float rockNoise = rockN1 * 0.65 + rockN2 * 0.35;

  vec3 rockColor = mix(basalt, ashRock, smoothstep(0.0, 0.5, normHeight));
  rockColor = mix(rockColor, dustRock, smoothstep(0.5, 0.9, normHeight));
  rockColor *= mix(0.75, 1.0, 1.0 - slope * 0.5);
  rockColor *= 0.85 + rockNoise * 0.3;
  rockColor *= 1.0 - gullyField * 0.3;

  // === Crater geometry ===
  vec2 cp = (w - craterOff) * vec2(ellipX, ellipY);
  float craterDist = length(cp);

  // === Flow map: valley-seeking lava from CPU computation ===
  vec2 flowUV = vUv;
  float flowVal = texture2D(uFlowTex, flowUV).r;

  // Threshold flow into lava mask (hot core + crust edge)
  float lavaMask = smoothstep(0.45, 0.80, flowVal);
  float crustMask = smoothstep(0.30, 0.50, flowVal) * (1.0 - lavaMask);

  // Lava only outside crater bowl
  float outsideCrater = smoothstep(craterRad * 0.7, craterRad * 1.1, craterDist);
  lavaMask *= outsideCrater;
  crustMask *= outsideCrater;

  // Subtle turbulence inside lava (no zebra stripes)
  float lavaTurb = snoise(centered * 8.0 + vec2(uTime * 0.2, -uTime * 0.15) + seedOff) * 0.15;
  lavaMask = clamp(lavaMask + lavaTurb * lavaMask, 0.0, 1.0);

  // Lava colors
  float lavaPulse = 0.85 + 0.15 * sin(uTime * 1.5 + dist * 3.0);
  vec3 hotLava = mix(vec3(1.0, 0.3, 0.0), vec3(1.0, 0.7, 0.1), lavaPulse * 0.5) * 2.0;
  vec3 crustColor = vec3(0.15, 0.08, 0.05);

  // === Crater magma ===
  float craterMask = 1.0 - smoothstep(0.0, craterRad * 0.8, craterDist);
  craterMask *= craterMask;
  float magmaPulse = 0.6 + 0.4 * sin(uTime * 1.2);
  float magmaDetail = snoise(centered * 6.0 + vec2(uTime * 0.3, -uTime * 0.2) + seedOff) * 0.5 + 0.5;
  vec3 magmaColor = mix(vec3(1.0, 0.25, 0.0), vec3(1.0, 0.8, 0.15), magmaDetail * magmaPulse);

  // === Grass at base/apron only ===
  vec3 grassColor = texture2D(uGrassTexture, vUv * 3.0).rgb * vec3(0.55, 0.62, 0.48);
  float grassEdge = snoise(centered * 6.0 + seedOff + 7.7) * 0.08;
  float grassMask = smoothstep(0.55, 0.85 + grassEdge, rawDist);

  // === Lighting: wrap diffuse + AO ===
  vec3 sunDir = normalize(uSunDirection);
  float NdotL = dot(n, sunDir);
  float wrapDiffuse = NdotL * 0.5 + 0.5;
  wrapDiffuse *= wrapDiffuse;
  vec3 lighting = vec3(0.3) + uSunColor * wrapDiffuse * 0.7 * uSunIntensity;

  // AO: darken crater interior, gully concavities, and ridge crests brightened
  float craterAO = mix(0.4, 1.0, smoothstep(0.0, craterRad * 1.2, craterDist));
  float gullyAO = 1.0 - gullyField * 0.25;
  float slopeAO = 1.0 + slope * 0.15;  // slightly brighten steep ridges
  float ao = craterAO * gullyAO * slopeAO;

  // === Combine ===
  vec3 surfaceColor = mix(rockColor, grassColor, grassMask);
  vec3 color = surfaceColor * lighting * ao;

  // Scale emissive effects by emergence (lava fades in as volcano rises)
  float emissiveFade = smoothstep(0.3, 0.8, vEmergence);

  // Lava crust (dark cooled edges)
  color = mix(color, crustColor, crustMask * 0.8 * emissiveFade);

  // Hot lava (emissive, ignores lighting)
  color = mix(color, hotLava, lavaMask * 0.9 * emissiveFade);

  // Crater magma (strongest emissive)
  color = mix(color, magmaColor * 2.5, craterMask * emissiveFade);

  // Rim glow
  float rimGlow = smoothstep(craterRad * 1.6, craterRad, craterDist)
                * (1.0 - smoothstep(0.0, craterRad * 0.8, craterDist));
  color += vec3(1.0, 0.3, 0.0) * rimGlow * 0.3 * magmaPulse * emissiveFade;

  // Edge fade for soft blending with ground
  float edgeAlpha = smoothstep(1.0, 0.85, rawDist);

  // === Debug views ===
  if (uDebugMode == 1) {
    // Height (grayscale)
    float hNorm = clamp(vHeight / (uHeight * 1.2), -0.2, 1.0) * 0.5 + 0.5;
    gl_FragColor = vec4(vec3(hNorm), edgeAlpha);
    return;
  }
  if (uDebugMode == 2) {
    // Slope
    gl_FragColor = vec4(vec3(slope), edgeAlpha);
    return;
  }
  if (uDebugMode == 3) {
    // Gully field
    gl_FragColor = vec4(vec3(gullyField), edgeAlpha);
    return;
  }
  if (uDebugMode == 4) {
    // Flow map
    gl_FragColor = vec4(vec3(flowVal), edgeAlpha);
    return;
  }

  gl_FragColor = vec4(color, edgeAlpha);
}
