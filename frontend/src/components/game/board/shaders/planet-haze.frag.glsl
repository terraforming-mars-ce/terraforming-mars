uniform mat4 uHazeCameraWorld;
uniform mat4 uHazeProjectionInverse;
uniform vec2 uHazeViewport;
uniform vec3 uHazeSunPosition;
uniform vec3 uHazeSunColor;
uniform int uHazeCount;
uniform vec4 uHazeBodies[MAX_ATMOSPHERES];
uniform vec3 uHazeColors[MAX_ATMOSPHERES];
uniform vec3 uHazeShadowColors[MAX_ATMOSPHERES];
uniform vec3 uHazeProfiles[MAX_ATMOSPHERES];

vec3 applyPlanetHaze(vec3 color, bool linearColor, vec3 surfaceNormal) {
  // Fragment depth includes displaced terrain and instanced vegetation.
  vec2 uv = gl_FragCoord.xy / uHazeViewport;
  vec4 viewPosition = uHazeProjectionInverse * vec4(uv * 2.0 - 1.0, gl_FragCoord.z * 2.0 - 1.0, 1.0);
  vec3 worldPosition = (uHazeCameraWorld * vec4(viewPosition.xyz / viewPosition.w, 1.0)).xyz;
  vec3 viewDirection = normalize(uHazeCameraWorld[3].xyz - worldPosition);
  for (int i = 0; i < MAX_ATMOSPHERES; i++) {
    if (i >= uHazeCount) { break; }
    vec3 offset = worldPosition - uHazeBodies[i].xyz;
    float height = length(offset) / uHazeBodies[i].w - 1.0;
    float thickness = uHazeProfiles[i].x;
    if (height > thickness || height < -0.5) { continue; }
    vec3 normal = normalize(offset);
    if (uHazeProfiles[i].z > 0.5 && linearColor) {
      normal = normalize(surfaceNormal);
      height = 0.0;
    }
    vec3 sunDirection = normalize(uHazeSunPosition - uHazeBodies[i].xyz);
    vec3 keyDirection = normalize(sunDirection * 0.45 + vec3(0.0, 0.9, 0.0));
    float warmth = smoothstep(-0.35, 0.65, dot(normal, keyDirection));
    float daylight = (0.28 + 0.72 * smoothstep(-0.25, 0.65, dot(normal, sunDirection)));
    float facing = clamp(dot(normal, viewDirection), 0.0, 1.0);
    float path = 0.65 * pow(1.0 - facing, 2.2) + 0.01 * warmth * warmth;
    float altitude = max(height, 0.0) / thickness;
    float altitudeFade = exp(-2.0 * altitude) * (1.0 - smoothstep(0.6, 1.0, altitude));
    float transmission = exp(-path * daylight * uHazeProfiles[i].y * altitudeFade);
    vec3 radiance = mix(uHazeShadowColors[i] * 0.3, uHazeColors[i] * uHazeSunColor * 1.2, warmth);
    if (!linearColor) {
      #ifdef TONE_MAPPING
        radiance = toneMapping(radiance);
      #endif
      radiance = linearToOutputTexel(vec4(radiance, 1.0)).rgb;
    }
    color = color * transmission + radiance * (1.0 - transmission);
  }
  return color;
}
