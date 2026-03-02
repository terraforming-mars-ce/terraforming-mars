uniform float uSphereRadius;
uniform float uZOffset;
uniform vec3 uSphereCenter;
//#pragma body
vec4 worldPos = modelMatrix * vec4(position, 1.0);
vec3 offset = worldPos.xyz - uSphereCenter;
vec3 sphereDir = normalize(offset);
vec3 projectedPos = uSphereCenter + sphereDir * (uSphereRadius + uZOffset);
vec3 transformed = (inverse(modelMatrix) * vec4(projectedPos, 1.0)).xyz;
