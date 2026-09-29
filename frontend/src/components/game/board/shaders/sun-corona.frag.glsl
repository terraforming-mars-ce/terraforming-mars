uniform float uTime;
varying vec3 vWorldNormal;
varying vec3 vWorldPosition;
void main() {
  vec3 normal = normalize(vWorldNormal);
  vec3 view = normalize(cameraPosition - vWorldPosition);
  float facing = dot(normal, view);
  float altitude = max(1.5 * sqrt(max(1.0 - facing * facing, 0.0)) - 1.0, 0.0);
  vec3 tangent = normalize(normal - facing * view);
  float t = uTime * 0.12;
  float filaments = sin(dot(tangent, vec3(83.0, 127.0, 61.0)) + sin(dot(tangent, vec3(19.0, 31.0, 47.0)) + t) * 2.5);
  float streamers = pow(0.5 + 0.5 * filaments, 4.0);
  float lengthScale = 9.0 + 8.0 * (0.5 + 0.5 * sin(dot(tangent, vec3(7.0, 11.0, 5.0)) - t));
  float density = exp(-altitude * lengthScale) * (0.14 + 0.11 * streamers);
  density += 0.4 * exp(-altitude * 65.0);
  density *= 1.0 - smoothstep(0.28, 0.5, altitude);
  vec3 color = mix(vec3(1.0, 0.5, 0.14), vec3(1.0, 0.13, 0.018), smoothstep(0.0, 0.2, altitude));
  gl_FragColor = vec4(color * density, 1.0);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
