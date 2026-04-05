precision highp float;

uniform float uSphereRadius;
uniform vec3 uSphereCenter;
uniform float uZOffset;
uniform float uProjectionScale;

out vec2 vFlatPos;
out vec3 vLocalPos;
out vec3 vNormal;
out vec3 vLocalNormal;
out vec3 vTangent;
out vec3 vBitangent;

void main() {
  vFlatPos = position.xy;

  float x = position.x * uProjectionScale;
  float y = position.y * uProjectionScale;
  float r = sqrt(x * x + y * y);

  vec3 localSpherePos;
  if (r < 0.0001) {
    localSpherePos = vec3(0.0, 0.0, uSphereRadius + uZOffset);
  } else {
    float theta = atan(y, x);
    float phi = (r / uSphereRadius) * (3.14159265 / 2.0);
    vec3 sphereDir = vec3(
      sin(phi) * cos(theta),
      sin(phi) * sin(theta),
      cos(phi)
    );
    localSpherePos = sphereDir * (uSphereRadius + uZOffset);
  }

  // Transform to world space via modelMatrix (accounts for Mars orbit + rotation)
  vec4 worldPos = modelMatrix * vec4(localSpherePos, 1.0);
  vec3 projectedPos = worldPos.xyz;

  // Sphere normal in local space, then transform to world
  vec3 localNormal = normalize(localSpherePos);
  vec3 worldNormal = normalize(mat3(modelMatrix) * localNormal);

  vLocalPos = localSpherePos;
  vLocalNormal = localNormal;

  vec3 localUp = abs(localNormal.y) < 0.999 ? vec3(0.0, 1.0, 0.0) : vec3(1.0, 0.0, 0.0);
  vec3 localTangent = normalize(cross(localUp, localNormal));
  vec3 localBitangent = cross(localNormal, localTangent);

  // Pass local-space tangent frame — lighting is done in local space
  vNormal = localNormal;
  vTangent = localTangent;
  vBitangent = localBitangent;

  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
