varying vec4 vOceanWorldPosition;
varying vec3 vOceanNormal;
varying vec3 vWorldSphereNormal;
varying vec3 vOceanTangent;
varying vec3 vOceanBitangent;
varying vec2 vOceanUv;
varying vec3 vOceanViewNormal;

uniform float time;
uniform float oceanSize;
uniform float oceanAlpha;
uniform float rf0;
uniform float sunIntensity;
uniform sampler2D normalSampler;
uniform vec3 sunColor;
uniform vec3 sunDirection;
uniform vec3 eye;
uniform vec3 waterColor;

uniform float uRadius;
uniform float uAspect;
uniform float uRotation;

uniform float uEdgeBand;
uniform float uEdgeStrength;
uniform float uEdgeScale;
uniform float uWarpScale;
uniform float uWarpAmount;

uniform float uSandWidth;
uniform float uGrainScale;
uniform sampler2D sandSampler;
uniform float uSandTexScale;

uniform float uShallowWidth;
uniform float uShallowStrength;

uniform float uEdgeSoftness;

uniform vec2 uSeedOffset;

uniform float uFoamWidth;
uniform float uFoamStrength;
uniform float uFoamScale;
uniform float uFoamSpeed;
uniform float uFoamCutoff;
uniform float uFoamPulseSpeed;
uniform float uFoamPulseAmount;

mat2 oceanRot2(float a) {
  float c = cos(a), s = sin(a);
  return mat2(c, -s, s, c);
}

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

vec4 oceanGetNoise(vec2 uv) {
  float t1 = sin(time * 0.3) * 2.0;
  float t2 = cos(time * 0.25) * 2.0;
  float t3 = sin(time * 0.2 + 1.0) * 2.0;
  float t4 = cos(time * 0.22 + 0.5) * 2.0;

  vec2 uv0 = (uv / 103.0) + vec2(t1 / 17.0, t2 / 29.0);
  vec2 uv1 = uv / 107.0 - vec2(t2 / -19.0, t1 / 31.0);
  vec2 uv2 = uv / vec2(8907.0, 9803.0) + vec2(t3 / 101.0, t4 / 97.0);
  vec2 uv3 = uv / vec2(1091.0, 1027.0) - vec2(t4 / 109.0, t3 / -113.0);
  vec4 noise = texture2D(normalSampler, uv0) +
    texture2D(normalSampler, uv1) +
    texture2D(normalSampler, uv2) +
    texture2D(normalSampler, uv3);
  return noise * 0.5 - 1.0;
}

void oceanSunLight(const vec3 surfaceNormal, const vec3 eyeDirection, float shiny, float spec, float diffAmt, inout vec3 diffuseClr, inout vec3 specularClr) {
  vec3 reflection = normalize(reflect(-sunDirection, surfaceNormal));
  float direction = max(0.0, dot(eyeDirection, reflection));
  specularClr += pow(direction, shiny) * sunColor * spec;
  diffuseClr += max(dot(sunDirection, surfaceNormal), 0.0) * sunColor * diffAmt;
}
//#pragma body
{
  vec2 p = (vOceanUv - 0.5) * 2.0;
  p = oceanRot2(uRotation) * p;
  vec2 q = p / vec2(uRadius * uAspect, uRadius);

  float baseSdf = length(q) - 1.0;
  float edgeBand = 1.0 - smoothstep(0.0, uEdgeBand, abs(baseSdf));

  vec2 seedOff = uSeedOffset;
  float w1 = oceanSnoise(p * uWarpScale + seedOff);
  float w2 = oceanSnoise((p + 31.7) * uWarpScale + seedOff);
  vec2 warp = vec2(w1, w2) * uWarpAmount;

  float edgeNoise = oceanSnoise((p + warp) * uEdgeScale + seedOff);
  edgeNoise = edgeNoise / (1.0 + abs(edgeNoise));

  float displacedSdf = baseSdf + edgeNoise * uEdgeStrength * edgeBand;
  float aa = fwidth(displacedSdf) * 1.5;
  float waterMask = 1.0 - smoothstep(-uEdgeSoftness - aa, uEdgeSoftness + aa, displacedSdf);

  float shoreDist = displacedSdf;
  float sandJitter = oceanSnoise(p * (uEdgeScale * 1.2) + seedOff + 17.3) * 0.5 + 0.5;
  sandJitter = (sandJitter - 0.5) * 0.03;
  float sandWidthVal = max(0.02, uSandWidth + sandJitter);
  float shoreT = smoothstep(0.0 - aa, sandWidthVal + aa, shoreDist);

  float sandAlpha = (1.0 - shoreT) * step(0.0, shoreDist);
  sandAlpha *= 0.95;
  float discEdge = length((vOceanUv - 0.5) * 2.0);
  sandAlpha *= 1.0 - smoothstep(0.75, 1.0, discEdge);
  sandAlpha = clamp(sandAlpha, 0.0, 1.0);

  float shallowMask = smoothstep(-uShallowWidth - aa, 0.0 + aa, shoreDist) * waterMask;

  float foamBand = smoothstep(-uFoamWidth - aa, 0.0 + aa, shoreDist) * waterMask;
  foamBand = pow(clamp(foamBand, 0.0, 1.0), 1.6);

  vec2 foamUv = (p + warp) * uFoamScale + seedOff * 0.7;
  foamUv += vec2(time * uFoamSpeed, time * (uFoamSpeed * 0.73));
  float fn1 = oceanSnoise(foamUv);
  float fn2 = oceanSnoise(foamUv * 2.1 + 11.3);
  float fn3 = oceanSnoise(foamUv * 4.0 - 27.1);
  float foamNoise = fn1 * 0.55 + fn2 * 0.30 + fn3 * 0.15;
  foamNoise = foamNoise * 0.5 + 0.5;
  float foamSpots = smoothstep(uFoamCutoff, 1.0, foamNoise);

  float pulseBase = sin(time * uFoamPulseSpeed + dot(seedOff, vec2(0.12, 0.37)));
  pulseBase = pulseBase * 0.5 + 0.5;
  float pulse = smoothstep(0.35, 0.95, pulseBase);
  pulse = mix(1.0, pulse, uFoamPulseAmount);
  float foamMask = clamp(foamBand * foamSpots * pulse, 0.0, 1.0);

  vec2 oceanProjPos = vec2(
    dot(vOceanWorldPosition.xyz, vOceanTangent),
    dot(vOceanWorldPosition.xyz, vOceanBitangent)
  );
  vec4 noise = oceanGetNoise(oceanProjPos * oceanSize);

  vec3 tangentNormal = normalize(noise.xzy * vec3(1.5, 1.0, 1.5));
  vec3 surfaceNormal = normalize(
    vOceanTangent * tangentNormal.x +
    vOceanNormal * tangentNormal.y +
    vOceanBitangent * tangentNormal.z
  );

  vec3 oceanDiffuseLight = vec3(0.0);
  vec3 oceanSpecularLight = vec3(0.0);
  vec3 worldToEye = eye - vOceanWorldPosition.xyz;
  vec3 eyeDirection = normalize(worldToEye);

  oceanSunLight(surfaceNormal, eyeDirection, 100.0, 2.0 * sunIntensity, 0.5 * sunIntensity, oceanDiffuseLight, oceanSpecularLight);

  float theta = max(dot(eyeDirection, surfaceNormal), 0.0);
  float reflectance = rf0 + (1.0 - rf0) * pow((1.0 - theta), 5.0);
  vec3 scatter = max(0.0, dot(surfaceNormal, eyeDirection)) * waterColor;

  vec3 toSun = normalize(-vOceanWorldPosition.xyz);
  float dotNL = max(dot(vWorldSphereNormal, toSun), 0.0);
  vec3 sceneIrradiance = vec3(0.1) + sunColor * sunIntensity * dotNL * 0.5;

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

  vec2 sandUvCoord = (p + warp * 0.5) * uSandTexScale + seedOff * 0.3;
  vec3 sandCol = texture2D(sandSampler, sandUvCoord).rgb;
  float grain = oceanSnoise(p * uGrainScale + seedOff * 2.0) * 0.5 + 0.5;
  sandCol *= mix(0.95, 1.05, grain);
  sandCol *= sceneIrradiance * RECIPROCAL_PI;

  vec3 shallowWaterColor = vec3(0.32, 0.58, 0.68) * sceneIrradiance * RECIPROCAL_PI;
  vec3 waterCol = deepWaterColor;
  waterCol = mix(waterCol, shallowWaterColor, shallowMask * uShallowStrength);

  float fres = pow(1.0 - clamp(vOceanViewNormal.z, 0.0, 1.0), 3.0);
  waterCol += fres * 0.06;

  waterCol = mix(waterCol, vec3(envBrightness), foamMask * uFoamStrength);

  vec3 col = sandCol;
  col = mix(col, waterCol, waterMask);

  float wetMix = 0.10;
  col = mix(col, mix(sandCol, waterCol, 0.75), waterMask * sandAlpha * wetMix);

  float finalAlpha = clamp(waterMask + sandAlpha, 0.0, 1.0);
  float outA = oceanAlpha * finalAlpha;

  gl_FragColor = vec4(col * oceanAlpha, outA);
}
