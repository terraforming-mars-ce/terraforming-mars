uniform vec3 glowColor;
uniform float glowPower;
uniform float glowStrength;
uniform float uTime;
uniform float noiseScale;
uniform float noiseStrength;

varying vec3 vNormal;
varying vec3 vViewPosition;
varying vec3 vWorldPosition;

float hash(vec3 p) {
  return fract(sin(dot(p, vec3(12.9898, 78.233, 23.112))) * 43758.5453);
}

float noise3d(vec3 p) {
  vec3 i = floor(p);
  vec3 f = fract(p);
  vec3 u = f * f * (3.0 - 2.0 * f);

  float a = hash(i);
  float b = hash(i + vec3(1.0, 0.0, 0.0));
  float c = hash(i + vec3(0.0, 1.0, 0.0));
  float d = hash(i + vec3(1.0, 1.0, 0.0));
  float e = hash(i + vec3(0.0, 0.0, 1.0));
  float f1 = hash(i + vec3(1.0, 0.0, 1.0));
  float g = hash(i + vec3(0.0, 1.0, 1.0));
  float h = hash(i + vec3(1.0, 1.0, 1.0));

  return mix(
    mix(mix(a, b, u.x), mix(c, d, u.x), u.y),
    mix(mix(e, f1, u.x), mix(g, h, u.x), u.y),
    u.z
  );
}

float fbm(vec3 p) {
  float v = 0.0;
  float amp = 0.5;
  float t = uTime * 0.08;

  for (int i = 0; i < 5; i++) {
    v += amp * noise3d(p + t);
    p *= 2.0;
    amp *= 0.6;
    t *= 1.3;
  }
  return v;
}

float domainWarpedFbm(vec3 p) {
  vec3 q = vec3(
    fbm(p + vec3(0.0, 0.0, 0.0)),
    fbm(p + vec3(5.2, 1.3, 3.1)),
    fbm(p + vec3(1.7, 9.2, 4.8))
  );
  return fbm(p + 3.0 * q);
}

void main() {
  float d = dot(normalize(-vViewPosition), normalize(vNormal));
  float fresnel = pow(max(d, 0.0), glowPower);

  float n = 0.0;
  if (noiseStrength > 0.0) {
    vec3 noisePos = normalize(vWorldPosition) * noiseScale;
    n = domainWarpedFbm(noisePos);
    fresnel *= mix(1.0, 0.4 + n * 1.2, noiseStrength);
  }

  vec3 color = glowColor * fresnel * glowStrength;
  gl_FragColor = vec4(color, fresnel * glowStrength);
}
