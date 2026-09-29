uniform sampler2D uNoiseMap;
uniform sampler2D uNoiseMapHigh;
varying vec2 vBoardPos;
varying vec4 vLandscapeField;
varying float vShoreDistance;
//#pragma body
#include <alphamap_fragment>
float n1=texture2D(uNoiseMap,vBoardPos*(1.5/0.166)).r;
float n2=texture2D(uNoiseMapHigh,vBoardPos*(1.5/0.166)*2.5).r;
float terrainNoise=n1*0.7+n2*0.3;
float edgeWarp=(terrainNoise*2.0-1.0)*0.012;
float grassAlpha=smoothstep(-0.007,0.014,vLandscapeField.x+edgeWarp);
float cityPatches=mix(0.28,0.94,smoothstep(0.22,0.61,terrainNoise));
float forestBlend=smoothstep(-0.025,0.055,vLandscapeField.y+edgeWarp);
grassAlpha*=mix(cityPatches,1.0,forestBlend);
// Moisture supplies broad patches independently of the direction of nearby cities.
vec2 moistureUv=vBoardPos*3.1;
vec2 moistureWarp=vec2(texture2D(uNoiseMap,moistureUv).r,texture2D(uNoiseMap,moistureUv+vec2(0.37,0.61)).r)-0.5;
float moisture=texture2D(uNoiseMap,moistureUv+moistureWarp*0.45).r;
float shoreReach=mix(0.035,0.15,smoothstep(0.25,0.75,moisture));
float shoreGrowth=1.0-smoothstep(shoreReach*0.25,shoreReach,vShoreDistance+(terrainNoise-0.5)*0.035);
float shorePatches=mix(0.3,0.95,smoothstep(0.2,0.7,terrainNoise+moisture*0.2));
grassAlpha=max(grassAlpha,shoreGrowth*shorePatches);
grassAlpha*=smoothstep(0.0,0.002,vLandscapeField.z);
grassAlpha*=smoothstep(0.0,0.025,vLandscapeField.w);
diffuseColor.a*=grassAlpha;
