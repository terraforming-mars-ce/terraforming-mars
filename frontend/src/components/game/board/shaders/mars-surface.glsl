uniform sampler2D uClimateNoise;
uniform float uChill;
// Shared by the Mars sphere and the landscape's bare ground so both match across the seam.
vec3 marsClimateSurface(vec3 color,vec2 uv) {
  // A barely-there cold cast that lifts as the planet warms.
  return mix(color,color*vec3(0.9,0.96,1.08),uChill*0.6);
}
