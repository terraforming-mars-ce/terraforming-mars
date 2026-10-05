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
float landscapeProgress(float birth) {
  float t=clamp((uLandscapeTime-birth)/0.6,0.0,1.0);
  return t*t*(3.0-2.0*t);
}
vec2 fieldUv(vec2 local) {
  return (local*(uFieldSamples-1.0)+uFieldBorder+0.5)/uFieldSize;
}
vec4 landscapeField(vec2 local,float layer,float birth) {
  float t=landscapeProgress(birth);
  vec2 uv=fieldUv(local);
  vec4 next=texture(uTerrain,vec3(uv,layer+uLayerCapacity));
  if(t>=1.0) {return next;}
  return mix(texture(uTerrain,vec3(uv,layer)),next,t);
}
vec4 landscapeMaterials(vec2 local,float layer,float birth) {
  float t=landscapeProgress(birth);
  vec2 uv=fieldUv(local);
  vec4 next=texture(uMaterials,vec3(uv,layer+uLayerCapacity));
  if(t>=1.0) {return next;}
  return mix(texture(uMaterials,vec3(uv,layer)),next,t);
}
vec4 landscapeDetail(vec2 local,float layer,float birth) {
  float t=landscapeProgress(birth);
  vec2 uv=fieldUv(local);
  vec4 next=texture(uDetailField,vec3(uv,layer+uLayerCapacity));
  if(t>=1.0) {return next;}
  return mix(texture(uDetailField,vec3(uv,layer)),next,t);
}
float landscapeBasinShore(vec2 local,float layer,float birth) {
  float t=landscapeProgress(birth);
  vec2 uv=fieldUv(local);
  float next=texture(uTerrain,vec3(uv,layer+uLayerCapacity)).y;
  if(t>=1.0) {return next;}
  float previous=texture(uTerrain,vec3(uv,layer)).y;
  if(t<=0.0) {return previous;}
  return min(previous,next);
}
vec3 landscapeProject(vec2 p,float height) {
  float d=length(p);
  if(d<0.000001) {return vec3(0.0,0.0,2.02+height);}
  return vec3(p*sin(d/2.02)/d,cos(d/2.02))*(2.02+height);
}
