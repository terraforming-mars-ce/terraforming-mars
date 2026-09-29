#version 100
precision highp float;

uniform vec3 uColor;
uniform float uOpacity;
uniform sampler2D uNoiseTex;
varying vec2 vUv;
varying vec2 vWorldUv;

void main() {
  float noise = texture2D(uNoiseTex, vWorldUv).r;
  float alpha = uOpacity * mix(0.92, 1.0, noise);
  gl_FragColor = vec4(uColor, alpha);
}
