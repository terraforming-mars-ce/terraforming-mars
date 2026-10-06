precision highp sampler2DArray;
uniform sampler2DArray uTerrain;
uniform sampler2DArray uMaterials;
uniform sampler2DArray uDetailField;
uniform float uLandscapeTime;
uniform float uLayerCapacity;
uniform float uPatchSize;
uniform float uFieldSize;
uniform float uFieldSamples;
uniform float uFieldBorder;
uniform float uWaterLevel;
uniform float uMarsReliefDepth;
uniform vec2 uBasinReach;
// x: lake radius, y: shore distance of land no lake reaches.
uniform vec2 uLakeShape;
vec2 fieldUv(vec2 local) {
  return (local*(uFieldSamples-1.0)+uFieldBorder+0.5)/uFieldSize;
}
float landscapeElapsed(float birth) {
  return clamp((uLandscapeTime-birth)/0.6,0.0,1.0);
}
// A growing lake settles along one front that runs at constant speed from its deepest point out to
// the edge of its reach, each texel blending as the front passes. Water, sand, banks and frost all
// derive from the same blended fields, so they arrive together in one outward motion. Other
// changes blend in place. Returns the weight of the next field. Mirrors lakeBlendWeight in
// landscapeFields.ts.
float landscapeWeight(float previous,float next,float elapsed) {
  if(next>=previous) {return elapsed*elapsed*(3.0-2.0*elapsed);}
  float front=-uLakeShape.x+elapsed*(uLakeShape.y+uLakeShape.x);
  // Pockets between merging lakes lie deeper than the front's start; clamped, they settle first.
  float settle=max(next,-uLakeShape.x);
  // The last texels before the edge of the reach finish with the transition.
  return max(smoothstep(settle,settle+0.1,front),smoothstep(0.85,1.0,elapsed));
}
float landscapeWeightAt(vec2 uv,float layer,float elapsed) {
  float previous=texture(uTerrain,vec3(uv,layer)).y;
  float next=texture(uTerrain,vec3(uv,layer+uLayerCapacity)).y;
  return landscapeWeight(previous,next,elapsed);
}
vec4 landscapeField(vec2 local,float layer,float birth) {
  vec2 uv=fieldUv(local);
  float elapsed=landscapeElapsed(birth);
  vec4 next=texture(uTerrain,vec3(uv,layer+uLayerCapacity));
  if(elapsed>=1.0) {return next;}
  vec4 previous=texture(uTerrain,vec3(uv,layer));
  return mix(previous,next,landscapeWeight(previous.y,next.y,elapsed));
}
vec4 landscapeMaterials(vec2 local,float layer,float birth) {
  vec2 uv=fieldUv(local);
  float elapsed=landscapeElapsed(birth);
  vec4 next=texture(uMaterials,vec3(uv,layer+uLayerCapacity));
  if(elapsed>=1.0) {return next;}
  return mix(texture(uMaterials,vec3(uv,layer)),next,landscapeWeightAt(uv,layer,elapsed));
}
vec4 landscapeDetail(vec2 local,float layer,float birth) {
  vec2 uv=fieldUv(local);
  float elapsed=landscapeElapsed(birth);
  vec4 next=texture(uDetailField,vec3(uv,layer+uLayerCapacity));
  if(elapsed>=1.0) {return next;}
  return mix(texture(uDetailField,vec3(uv,layer)),next,landscapeWeightAt(uv,layer,elapsed));
}
float landscapeBasinShore(vec2 local,float layer,float birth) {
  vec2 uv=fieldUv(local);
  float elapsed=landscapeElapsed(birth);
  float next=texture(uTerrain,vec3(uv,layer+uLayerCapacity)).y;
  if(elapsed>=1.0) {return next;}
  float previous=texture(uTerrain,vec3(uv,layer)).y;
  if(elapsed<=0.0) {return previous;}
  // A growing lake claims its hole and basin texels only once they start settling, so unsettled
  // ground is never drawn opaque by the basin.
  if(next<previous&&landscapeWeight(previous,next,elapsed)<0.02) {return previous;}
  return min(previous,next);
}
vec3 landscapeProject(vec2 p,float height) {
  float d=length(p);
  if(d<0.000001) {return vec3(0.0,0.0,2.02+height);}
  return vec3(p*sin(d/2.02)/d,cos(d/2.02))*(2.02+height);
}
