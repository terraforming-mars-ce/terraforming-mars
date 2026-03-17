#version 100
precision highp float;

uniform float uSphereRadius;
uniform float uHoleRadius;
uniform float uHoleDepth;
uniform float uEmergence;
uniform float uEmergenceRadius;
uniform float uSeed;

varying vec2 vUv;
varying vec2 vCentered;
varying float vDistFromCenter;
varying vec3 vWorldNormal;
varying vec3 vWorldPosition;
varying float vHeight;
varying float vEmergence;

float hash(float n) { return fract(sin(n) * 43758.5453123); }
float seedParam(float idx) { return hash(uSeed * 127.1 + idx * 311.7); }

float getHeight(vec2 centered) {
  float dist = length(centered);
  float h = 0.0;

  // Cylinder: uniform depth inside, sharp edge at rim
  float effectiveRadius = uHoleRadius * uEmergenceRadius;
  float insideHole = smoothstep(effectiveRadius + 0.02, effectiveRadius, dist);
  h -= insideHole * uHoleDepth;

  // Metal bar grid — protrude along mortar lines inside the hole
  if (dist < effectiveRadius) {
    float angle = atan(centered.y, centered.x);
    float u = angle / 6.283 + 0.5;
    float v = clamp(insideHole * uHoleDepth / (uHoleDepth + 0.0001), 0.0, 1.0);

    float blockCols = 12.0;
    float blockRows = 3.0;
    float mortarW = 0.05;

    float colCell = fract(u * blockCols);
    float rowCell = fract(v * blockRows);

    float hBar = 1.0 - smoothstep(0.0, mortarW, colCell) * smoothstep(0.0, mortarW, 1.0 - colCell);
    float vBar = 1.0 - smoothstep(0.0, mortarW, rowCell) * smoothstep(0.0, mortarW, 1.0 - rowCell);
    float barMask = max(hBar, vBar);

    h += barMask * 0.003 * insideHole;
  }

  return h * uEmergence;
}

void main() {
  vUv = uv;
  vec2 centered = (uv - 0.5) * 2.0;
  vCentered = centered;
  vDistFromCenter = length(centered);
  vEmergence = uEmergence;

  vec4 worldPos = modelMatrix * vec4(position, 1.0);
  vec3 sphereDir = normalize(worldPos.xyz);

  float h = getHeight(centered);
  vHeight = h;

  // Normal via finite differences
  float eps = 0.008;
  float hL = getHeight(centered + vec2(-eps, 0.0));
  float hR = getHeight(centered + vec2(eps, 0.0));
  float hD = getHeight(centered + vec2(0.0, -eps));
  float hU = getHeight(centered + vec2(0.0, eps));

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

  vec3 projectedPos = sphereDir * (uSphereRadius + 0.005 + h);
  vWorldPosition = projectedPos;

  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
