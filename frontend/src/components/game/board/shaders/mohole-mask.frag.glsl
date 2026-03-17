#version 100
precision highp float;

uniform float uHoleRadius;
uniform float uEmergenceRadius;

varying float vDistFromCenter;

void main() {
  if (vDistFromCenter >= uHoleRadius * uEmergenceRadius) discard;
  gl_FragColor = vec4(0.0);
}
