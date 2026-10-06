vec3 dx=terrainDx,dy=terrainDy;
vec2 ux=terrainUvDx,uy=terrainUvDy;
vec3 px=cross(dy,normal),py=cross(normal,dx);
vec3 tangent=px*ux.x+py*uy.x;
vec3 bitangent=px*ux.y+py*uy.y;
float basisScale=inversesqrt(max(max(dot(tangent,tangent),dot(bitangent,bitangent)),1e-20));
// Stronger relief up close, eased back where texels shrink below a pixel to avoid shimmer.
vec2 slope=surfaceSlope*mix(0.55,0.3,smoothstep(0.15,0.6,detailFootprint));
normal=normalize(normal*sqrt(max(0.0,1.0-dot(slope,slope)))+(tangent*slope.x+bitangent*slope.y)*basisScale);
roughnessFactor=surfaceRoughness;
vec3 pavingX=cross(dy,normal),pavingY=cross(normal,dx);
float pavingDet=dot(dx,pavingX);
vec3 pavingSlope=(pavingGradient.x*pavingX+pavingGradient.y*pavingY)*sign(pavingDet)/max(abs(pavingDet),1e-12);
normal=normalize(normal-pavingSlope);
