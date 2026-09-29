uniform sampler2D uNoiseMap;
uniform sampler2D uNoiseMapHigh;
uniform float uConnectionTime;
varying float vRoadEdge;
varying float vConnectionBirth;
varying vec2 vShoulderBoard;
//#pragma body
#include <alphamap_fragment>
float n=texture2D(uNoiseMap,vShoulderBoard*(1.5/0.166)).r*0.7;
n+=texture2D(uNoiseMapHigh,vShoulderBoard*(1.5/0.166)*2.5).r*0.3;
float fine=texture2D(uNoiseMapHigh,vShoulderBoard*(1.5/0.166)*11.0).r;
float disturbed=1.0-smoothstep(0.16,1.0,abs(vRoadEdge)+(n-0.5)*0.9);
disturbed*=mix(0.4,1.0,smoothstep(0.15,0.8,fine));
diffuseColor.rgb*=mix(0.65,1.15,n);
diffuseColor.a*=disturbed*smoothstep(0.0,0.6,uConnectionTime-vConnectionBirth);
