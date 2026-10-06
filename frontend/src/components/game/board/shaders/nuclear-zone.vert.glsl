uniform vec3 uSphereCenter;
uniform float uSphereRadius;
attribute float nuclearCoverage;
attribute float nuclearApron;
attribute vec3 nuclearPlanetDirection;
varying float vNuclearCoverage;
varying float vNuclearApron;
varying vec3 vNuclearPlanetDirection;
varying vec2 vUv;
varying vec2 vGroundPosition;
varying vec3 vWorldPosition;
varying vec3 vUp;
void main() {
  vUv = uv;
  vNuclearCoverage = nuclearCoverage;
  vNuclearApron = nuclearApron;
  vNuclearPlanetDirection = nuclearPlanetDirection;
  vGroundPosition = position.xy;
  vec3 base = (modelMatrix * vec4(position.xy, 0.0, 1.0)).xyz;
  vUp = normalize(base - uSphereCenter);
  vWorldPosition = uSphereCenter + vUp * (uSphereRadius + position.z);
  gl_Position = projectionMatrix * viewMatrix * vec4(vWorldPosition, 1.0);
}
