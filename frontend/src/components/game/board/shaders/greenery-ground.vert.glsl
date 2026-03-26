uniform float uSphereRadius;
uniform float uZOffset;
uniform vec3 uSphereCenter;
uniform mat4 uGroupInverseMatrix;
varying vec2 vLocalPos;
varying vec2 vTileOffset;
//#pragma body
vLocalPos = position.xy;
vec3 tileCenter = (uGroupInverseMatrix * modelMatrix * vec4(0.0, 0.0, 0.0, 1.0)).xyz;
vTileOffset = tileCenter.xy * 37.0;
vec4 worldPos = modelMatrix * vec4(position, 1.0);
vec3 sphereDir = normalize(worldPos.xyz - uSphereCenter);
vec3 projectedPos = uSphereCenter + sphereDir * (uSphereRadius + uZOffset);
vec3 transformed = position;
