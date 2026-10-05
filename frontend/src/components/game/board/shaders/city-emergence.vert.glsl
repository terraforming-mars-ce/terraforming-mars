attribute float cityBirth;
attribute float cityFrostExposure;
varying vec3 vCityFrostPosition;
varying vec3 vCityFrostNormal;
varying float vCityFrostExposure;
uniform float uCityTime;
uniform float uCityDuration;
uniform float uNuclearPlaybackRate;
attribute float cityCollapseStart;
attribute vec3 cityCollapsePivot;
attribute vec4 cityCollapseMotion;
float cityCollapseProgress() {
  float t = clamp(((uCityTime-cityCollapseStart)*uNuclearPlaybackRate-cityCollapseMotion.x)/0.75,0.0,1.0);
  return t*t;
}
mat3 cityCollapseRotation() {
  vec2 tilt = cityCollapseMotion.yz * cityCollapseProgress();
  float cx=cos(tilt.x), sx=sin(tilt.x), cy=cos(tilt.y), sy=sin(tilt.y);
  return mat3(cy,0.0,-sy, sy*sx,cx,cy*sx, sy*cx,-sx,cy*cx);
}
//#pragma body
#include <begin_vertex>
vCityFrostPosition=position;
vCityFrostNormal=normal;
vCityFrostExposure=cityFrostExposure;
float cityProgress=clamp((uCityTime-cityBirth)/uCityDuration,0.0,1.0);
transformed.z-=0.08*pow(1.0-cityProgress,3.0);
if (uCityTime >= cityCollapseStart) {
  transformed = cityCollapsePivot + cityCollapseRotation() * (transformed-cityCollapsePivot);
  transformed.z -= cityCollapseMotion.w * cityCollapseProgress();
}
