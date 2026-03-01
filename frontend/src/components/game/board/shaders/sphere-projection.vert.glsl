#version 100
precision highp float;

uniform float uSphereRadius;
uniform float uZOffset;
uniform vec3 uSphereCenter;
varying vec2 vUv;

void main() {
  vUv = uv;
  vec4 worldPos = modelMatrix * vec4(position, 1.0);
  vec3 offset = worldPos.xyz - uSphereCenter;
  vec3 sphereDir = normalize(offset);
  vec3 projectedPos = uSphereCenter + sphereDir * (uSphereRadius + uZOffset);
  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
