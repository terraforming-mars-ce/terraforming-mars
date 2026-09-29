uniform vec3 uColor;
uniform vec3 uShadowColor;
uniform vec3 uSunColor;
uniform vec3 uSunDirection;
uniform float uIntensity;
uniform float uThickness;
uniform float uFollowsMesh;
varying vec3 vWorldNormal;
varying vec3 vWorldPosition;

void main() {
  vec3 normal = normalize(vWorldNormal);
  vec3 viewDirection = normalize(cameraPosition - vWorldPosition);
  float facing = dot(normal, viewDirection);
  float rayRadius = (1.0 + uThickness) * sqrt(max(1.0 - facing * facing, 0.0));
  float altitude = max((rayRadius - 1.0) / uThickness, 0.0);
  float density = 0.3 * exp(-4.0 * altitude) * (1.0 - smoothstep(0.6, 1.0, altitude));
  // Nested copies of an irregular silhouette integrate a soft edge without a spherical bubble.
  if (uFollowsMesh > 0.5) {
    density = 0.3;
  }
  vec3 tangentNormal = normal - facing * viewDirection;
  normal = tangentNormal / max(length(tangentNormal), 0.0001);
  vec3 keyDirection = normalize(uSunDirection * 0.45 + vec3(0.0, 0.9, 0.0));
  float warmth = smoothstep(-0.35, 0.65, dot(normal, keyDirection));
  float daylight = (0.28 + 0.72 * smoothstep(-0.25, 0.65, dot(normal, uSunDirection)));
  vec3 scatteringColor = mix(uShadowColor * 0.9, uColor * uSunColor * 1.5, warmth);
  float opacity = mix(1.0, 1.0 / 16.0, uFollowsMesh);
  gl_FragColor = vec4(scatteringColor * density * daylight * uIntensity, opacity);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
