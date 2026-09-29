uniform sampler2D uSurface;
uniform float uTime;
varying vec2 vUv;
varying vec3 vLocalPosition;
varying vec3 vNormal;
varying vec3 vViewDirection;

float hash(vec3 p) {
  p = fract(p * 0.3183099 + vec3(0.11, 0.27, 0.43));
  p *= 17.0;
  return fract(p.x * p.y * p.z * (p.x + p.y + p.z));
}
float noise(vec3 p) {
  vec3 i = floor(p), f = fract(p);
  f = f * f * (3.0 - 2.0 * f);
  return mix(mix(mix(hash(i), hash(i + vec3(1,0,0)), f.x),
                 mix(hash(i + vec3(0,1,0)), hash(i + vec3(1,1,0)), f.x), f.y),
             mix(mix(hash(i + vec3(0,0,1)), hash(i + vec3(1,0,1)), f.x),
                 mix(hash(i + vec3(0,1,1)), hash(i + vec3(1,1,1)), f.x), f.y), f.z);
}
void main() {
  vec3 p = normalize(vLocalPosition);
  float t = uTime * 0.055;
  vec3 flow = vec3(noise(p * 7.0 + t), noise(p * 7.0 - t + 12.0), noise(p * 7.0 + 26.0));
  vec2 drift = (flow.xy - 0.5) * 0.009;
  vec3 surface = texture2D(uSurface, vec2(fract(vUv.x + uTime * 0.0006 + drift.x), clamp(vUv.y + drift.y, 0.002, 0.998))).rgb;
  float granules = noise(p * 120.0 + flow * 1.8 + t);
  float heat = clamp(dot(surface, vec3(0.5, 0.35, 0.15)) * 1.5, 0.0, 1.0);
  vec3 color = mix(vec3(0.9, 0.06, 0.003), vec3(2.4, 0.8, 0.12), heat);
  color *= 0.68 + 0.65 * granules;
  float facing = max(dot(normalize(vNormal), normalize(vViewDirection)), 0.0);
  color *= 0.55 + 0.45 * pow(facing, 0.35);
  gl_FragColor = vec4(color, 1.0);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
