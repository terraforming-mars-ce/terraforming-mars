precision highp float;
precision highp sampler2D;

in vec2 vFlatPos;
in vec4 vWorldPos;
in vec3 vNormal;
in vec3 vLocalNormal;
in vec3 vTangent;
in vec3 vBitangent;
in vec3 vViewNormal;

out vec4 fragColor;

// SDF data
uniform sampler2D uOceanData;
uniform int uPointCount;
uniform int uEdgeCount;
uniform float uCapsuleRadius;

// Visual
uniform float time;
uniform float oceanSize;
uniform float rf0;
uniform float sunIntensity;
uniform sampler2D normalSampler;
uniform vec3 sunColor;
uniform vec3 sunDirection;
uniform vec3 eye;
uniform vec3 waterColor;

// Edge warping
uniform float uEdgeBand;
uniform float uEdgeStrength;
uniform float uEdgeScale;
uniform float uWarpScale;
uniform float uWarpAmount;

// Sand
uniform float uSandWidth;
uniform float uGrainScale;
uniform sampler2D sandSampler;
uniform float uSandTexScale;

// Shallow water
uniform float uShallowWidth;
uniform float uShallowStrength;

// Edge softness
uniform float uEdgeSoftness;

// Foam
uniform float uFoamWidth;
uniform float uFoamStrength;
uniform float uFoamScale;
uniform float uFoamSpeed;
uniform float uFoamCutoff;
uniform float uFoamPulseSpeed;
uniform float uFoamPulseAmount;

// Hover
uniform vec2 uHoverCenter;
uniform float uHoverActive;

// --- Simplex noise ---

vec3 oceanMod289v3(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec2 oceanMod289v2(vec2 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec3 oceanPermute(vec3 x) { return oceanMod289v3(((x * 34.0) + 1.0) * x); }

float oceanSnoise(vec2 v) {
  const vec4 C = vec4(0.211324865405187, 0.366025403784439,
                      -0.577350269189626, 0.024390243902439);
  vec2 i = floor(v + dot(v, C.yy));
  vec2 x0 = v - i + dot(i, C.xx);
  vec2 i1 = (x0.x > x0.y) ? vec2(1.0, 0.0) : vec2(0.0, 1.0);
  vec4 x12 = x0.xyxy + C.xxzz;
  x12.xy -= i1;
  i = oceanMod289v2(i);
  vec3 p = oceanPermute(oceanPermute(i.y + vec3(0.0, i1.y, 1.0)) + i.x + vec3(0.0, i1.x, 1.0));
  vec3 m = max(0.5 - vec3(dot(x0, x0), dot(x12.xy, x12.xy), dot(x12.zw, x12.zw)), 0.0);
  m = m * m;
  m = m * m;
  vec3 x = 2.0 * fract(p * C.www) - 1.0;
  vec3 h = abs(x) - 0.5;
  vec3 ox = floor(x + 0.5);
  vec3 a0 = x - ox;
  m *= 1.79284291400159 - 0.85373472095314 * (a0 * a0 + h * h);
  vec3 g;
  g.x = a0.x * x0.x + h.x * x0.y;
  g.yz = a0.yz * x12.xz + h.yz * x12.yw;
  return 130.0 * dot(m, g);
}

// --- Water normal sampling ---

vec4 oceanGetNoise(vec2 uv) {
  float t1 = sin(time * 0.3) * 2.0;
  float t2 = cos(time * 0.25) * 2.0;
  float t3 = sin(time * 0.2 + 1.0) * 2.0;
  float t4 = cos(time * 0.22 + 0.5) * 2.0;

  vec2 uv0 = (uv / 103.0) + vec2(t1 / 17.0, t2 / 29.0);
  vec2 uv1 = uv / 107.0 - vec2(t2 / -19.0, t1 / 31.0);
  vec2 uv2 = uv / vec2(8907.0, 9803.0) + vec2(t3 / 101.0, t4 / 97.0);
  vec2 uv3 = uv / vec2(1091.0, 1027.0) - vec2(t4 / 109.0, t3 / -113.0);
  vec4 noise = texture(normalSampler, uv0) +
    texture(normalSampler, uv1) +
    texture(normalSampler, uv2) +
    texture(normalSampler, uv3);
  return noise * 0.5 - 1.0;
}

// --- Sun lighting ---

void oceanSunLight(const vec3 surfaceNormal, const vec3 eyeDirection, float shiny, float spec, float diffAmt, inout vec3 diffuseClr, inout vec3 specularClr) {
  vec3 reflection = normalize(reflect(-sunDirection, surfaceNormal));
  float direction = max(0.0, dot(eyeDirection, reflection));
  specularClr += pow(direction, shiny) * sunColor * spec;
  diffuseClr += max(dot(sunDirection, surfaceNormal), 0.0) * sunColor * diffAmt;
}

// --- Capsule SDF (distance to line segment) ---

float sdSegment(vec2 p, vec2 a, vec2 b) {
  vec2 pa = p - a;
  vec2 ba = b - a;
  float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
  return length(pa - ba * h);
}

void main() {
  vec2 p = vFlatPos;

  // --- Domain warp: distort p BEFORE SDF so the entire shape deforms organically ---
  float dw1 = oceanSnoise(p * uWarpScale);
  float dw2 = oceanSnoise((p + 31.7) * uWarpScale);
  vec2 warp = vec2(dw1, dw2) * uWarpAmount;
  vec2 warpedP = p + warp;

  // --- Compute merged capsule SDF using warped coordinates ---
  float minDist = 99999.0;

  for (int i = 0; i < uPointCount; i++) {
    vec4 data = texelFetch(uOceanData, ivec2(i, 0), 0);
    vec2 center = data.xy;
    float emergence = data.z;
    float dist = length(warpedP - center) - uCapsuleRadius * emergence;
    minDist = min(minDist, dist);
  }

  for (int i = 0; i < uEdgeCount; i++) {
    vec4 posData = texelFetch(uOceanData, ivec2(uPointCount + i * 2, 0), 0);
    vec4 idxData = texelFetch(uOceanData, ivec2(uPointCount + i * 2 + 1, 0), 0);
    vec2 a = posData.xy;
    vec2 b = posData.zw;

    float emA = texelFetch(uOceanData, ivec2(int(idxData.x), 0), 0).z;
    float emB = texelFetch(uOceanData, ivec2(int(idxData.y), 0), 0).z;
    float edgeEmergence = min(emA, emB);

    float dist = sdSegment(warpedP, a, b) - uCapsuleRadius * edgeEmergence;
    minDist = min(minDist, dist);
  }

  // Early discard for fragments far from any ocean
  if (minDist > uSandWidth + 0.05) {
    discard;
  }

  // --- Fine edge detail (small-scale roughness on top of the domain warp) ---
  float edgeNoise = oceanSnoise(warpedP * uEdgeScale);
  edgeNoise = edgeNoise / (1.0 + abs(edgeNoise));
  float edgeBand = 1.0 - smoothstep(0.0, uEdgeBand, abs(minDist));

  float shoreDist = minDist + edgeNoise * uEdgeStrength * edgeBand;
  float aa = fwidth(shoreDist) * 1.5;
  float waterMask = 1.0 - smoothstep(-uEdgeSoftness - aa, uEdgeSoftness + aa, shoreDist);

  // Discard fully transparent fragments
  if (waterMask < 0.001 && shoreDist > uSandWidth + 0.02) {
    discard;
  }

  // --- Sand rendering ---
  float sandJitter = oceanSnoise(p * (uEdgeScale * 1.2) + 17.3) * 0.5 + 0.5;
  sandJitter = (sandJitter - 0.5) * 0.03;
  float sandWidthVal = max(0.02, uSandWidth + sandJitter);
  float shoreT = smoothstep(0.0 - aa, sandWidthVal + aa, shoreDist);

  float sandAlpha = (1.0 - shoreT) * step(0.0, shoreDist);
  sandAlpha *= 0.95;
  sandAlpha = clamp(sandAlpha, 0.0, 1.0);

  // --- Shallow water ---
  float shallowMask = smoothstep(-uShallowWidth - aa, 0.0 + aa, shoreDist) * waterMask;

  // --- Foam ---
  float foamBand = smoothstep(-uFoamWidth - aa, 0.0 + aa, shoreDist) * waterMask;
  foamBand = pow(clamp(foamBand, 0.0, 1.0), 1.6);

  vec2 foamUv = (p + warp) * uFoamScale;
  foamUv += vec2(time * uFoamSpeed, time * (uFoamSpeed * 0.73));
  float fn1 = oceanSnoise(foamUv);
  float fn2 = oceanSnoise(foamUv * 2.1 + 11.3);
  float fn3 = oceanSnoise(foamUv * 4.0 - 27.1);
  float foamNoise = fn1 * 0.55 + fn2 * 0.30 + fn3 * 0.15;
  foamNoise = foamNoise * 0.5 + 0.5;
  float foamSpots = smoothstep(uFoamCutoff, 1.0, foamNoise);

  float pulseBase = sin(time * uFoamPulseSpeed);
  pulseBase = pulseBase * 0.5 + 0.5;
  float pulse = smoothstep(0.35, 0.95, pulseBase);
  pulse = mix(1.0, pulse, uFoamPulseAmount);
  float foamMask = clamp(foamBand * foamSpots * pulse, 0.0, 1.0);

  // --- Water normals ---
  // Use flat hex-grid position for normal map UVs — gives smooth spatial variation
  vec4 noise = oceanGetNoise(vFlatPos * oceanSize);

  vec3 tangentNormal = normalize(noise.xzy * vec3(1.5, 1.0, 1.5));
  vec3 surfaceNormal = normalize(
    vTangent * tangentNormal.x +
    vNormal * tangentNormal.y +
    vBitangent * tangentNormal.z
  );

  // --- Lighting ---
  vec3 oceanDiffuseLight = vec3(0.0);
  vec3 oceanSpecularLight = vec3(0.0);
  vec3 worldToEye = eye - vWorldPos.xyz;
  vec3 eyeDirection = normalize(worldToEye);

  oceanSunLight(surfaceNormal, eyeDirection, 100.0, 2.0 * sunIntensity, 0.5 * sunIntensity, oceanDiffuseLight, oceanSpecularLight);

  float theta = max(dot(eyeDirection, surfaceNormal), 0.0);
  float reflectance = rf0 + (1.0 - rf0) * pow((1.0 - theta), 5.0);
  vec3 scatter = max(0.0, dot(surfaceNormal, eyeDirection)) * waterColor;

  // Use local-space normal for sun lighting — Mars always faces the sun,
  // so the local sun direction is always (0,0,1) regardless of orbital position
  float dotNL = max(vLocalNormal.z, 0.0);
  vec3 sceneIrradiance = vec3(0.3) + sunColor * sunIntensity * dotNL * 0.7;

  float envBrightness = mix(0.1, 0.5, dotNL);

  vec3 skyColor = vec3(0.2, 0.3, 0.5) * envBrightness * sunColor;
  vec3 horizonColor = vec3(0.3, 0.35, 0.45) * envBrightness * sunColor;
  float skyGradient = max(0.0, surfaceNormal.y);
  vec3 reflectionSample = mix(horizonColor, skyColor, skyGradient);

  vec3 deepWaterColor = mix(
    (sunColor * oceanDiffuseLight * 0.3 + scatter),
    (vec3(0.05) + reflectionSample * 0.6 + reflectionSample * oceanSpecularLight * 0.5),
    reflectance
  );

  // --- Sand color ---
  vec2 sandUvCoord = (p + warp * 0.5) * uSandTexScale;
  vec3 sandCol = texture(sandSampler, sandUvCoord).rgb;
  float grain = oceanSnoise(p * uGrainScale) * 0.5 + 0.5;
  sandCol *= mix(0.95, 1.05, grain);
  sandCol *= sceneIrradiance;

  // --- Compositing ---
  vec3 shallowWaterColor = vec3(0.32, 0.58, 0.68) * sceneIrradiance;
  vec3 waterCol = deepWaterColor;
  waterCol = mix(waterCol, shallowWaterColor, shallowMask * uShallowStrength);

  float fres = pow(1.0 - clamp(vViewNormal.z, 0.0, 1.0), 3.0);
  waterCol += fres * 0.06;

  waterCol = mix(waterCol, vec3(envBrightness), foamMask * uFoamStrength);

  vec3 col = sandCol;
  col = mix(col, waterCol, waterMask);

  float wetMix = 0.10;
  col = mix(col, mix(sandCol, waterCol, 0.75), waterMask * sandAlpha * wetMix);

  float finalAlpha = clamp(waterMask + sandAlpha, 0.0, 1.0);

  // --- Hover glow ---
  if (uHoverActive > 0.5) {
    float hoverDist = length(p - uHoverCenter);
    float hoverGlow = 1.0 - smoothstep(0.0, 0.2, hoverDist);
    hoverGlow *= waterMask * 0.25;
    col += vec3(0.8, 0.9, 1.0) * hoverGlow;
  }

  fragColor = vec4(col, finalAlpha);
}
