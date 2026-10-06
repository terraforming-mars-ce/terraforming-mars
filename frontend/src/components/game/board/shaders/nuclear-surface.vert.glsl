uniform vec3 uSphereCenter;
uniform float uSphereRadius;
attribute float nuclearCoverage;
attribute float nuclearApron;
attribute vec3 nuclearPlanetDirection;
varying float vNuclearCoverage;
varying float vNuclearApron;
varying vec3 vNuclearPlanetDirection;
varying vec2 vNuclearUv;
varying vec2 vGroundPosition;
//#pragma body
vNuclearUv = uv;
vNuclearCoverage = nuclearCoverage;
vNuclearApron = nuclearApron;
vNuclearPlanetDirection = nuclearPlanetDirection;
vGroundPosition = position.xy;
vec3 nuclearBase = (modelMatrix * vec4(position.xy, 0.0, 1.0)).xyz;
vec3 nuclearUp = normalize(nuclearBase - uSphereCenter);
vec3 nuclearWorld = uSphereCenter + nuclearUp * (uSphereRadius + position.z);
// Keep the standard view/world-position pipeline on the displaced surface.
vec3 transformed = (inverse(modelMatrix) * vec4(nuclearWorld, 1.0)).xyz;
