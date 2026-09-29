#include <common>

uniform sampler2D uSky;
uniform float uBrightness;
varying vec3 vDirection;

void main() {
  vec3 direction = normalize(vDirection);
  vec3 color = max(texture2D(uSky, equirectUv(direction)).rgb, vec3(0.0));
  float luminance = dot(color, vec3(0.2126, 0.7152, 0.0722));
  // Lift faint galactic structure without raising black or multiplying bright stars.
  float lift = 1.0 + 1.2 / (1.0 + luminance / 0.008);
  gl_FragColor = vec4(color * lift * uBrightness, 1.0);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
