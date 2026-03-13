#version 100
precision highp float;

uniform vec3 uColor;
uniform float opacity;
varying vec2 vUv;

void main() {
  vec2 center = vUv - 0.5;
  float distFromCenter = length(center);
  float gradient = smoothstep(0.1, 0.45, distFromCenter);
  float alpha = gradient * opacity;
  gl_FragColor = vec4(uColor, alpha);
}
