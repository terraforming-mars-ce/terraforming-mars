uniform float uTime;
varying vec2 vUv;
varying vec3 vNormal;
varying vec3 vViewDirection;
void main() {
  float flow = 0.7 + 0.3 * sin(vUv.x * 55.0 - uTime * 1.4 + sin(vUv.x * 19.0 + uTime * 0.4));
  float facing = abs(dot(normalize(vNormal), normalize(vViewDirection)));
  vec3 color = mix(vec3(0.45, 0.018, 0.002), vec3(1.2, 0.16, 0.012), facing);
  gl_FragColor = vec4(color * flow, pow(facing, 1.5) * 0.65);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
