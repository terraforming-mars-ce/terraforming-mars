uniform float uCityFrost;
uniform float uCityFrostGlass;
uniform sampler2D uCityFrostNoise;
varying vec3 vCityFrostPosition;
varying vec3 vCityFrostNormal;
varying float vCityFrostExposure;
//#pragma body
if(uCityFrost>0.001) {
  vec3 frostNormal=normalize(vCityFrostNormal);
  vec3 frostWeights=pow(abs(frostNormal),vec3(4.0));
  frostWeights/=max(dot(frostWeights,vec3(1.0)),0.0001);
  vec3 frostUv=vCityFrostPosition*45.0;
  float frostGrain=
    texture2D(uCityFrostNoise,frostUv.yz).r*frostWeights.x+
    texture2D(uCityFrostNoise,frostUv.xz).r*frostWeights.y+
    texture2D(uCityFrostNoise,frostUv.xy).r*frostWeights.z;
  float frostCrystals=texture2D(uCityFrostNoise,
    (vCityFrostPosition.xy+vCityFrostPosition.z*vec2(0.63,0.81))*220.0).r;
  float frostUp=smoothstep(-0.1,0.8,frostNormal.z);
  float frostRim=1.0-smoothstep(0.02,0.045,vCityFrostPosition.z);
  float frostExposure=mix(0.3+0.7*frostUp,0.6+0.25*frostUp+0.15*frostRim,uCityFrostGlass);
  float frostPattern=0.2+0.8*smoothstep(0.35,0.65,frostGrain);
  float cityFrost=uCityFrost*vCityFrostExposure*frostExposure*frostPattern*0.6;
  vec3 frostColor=mix(vec3(0.55,0.66,0.73),vec3(0.8,0.86,0.9),smoothstep(0.4,0.62,frostCrystals));
  diffuseColor.rgb=mix(diffuseColor.rgb,frostColor,cityFrost);
  roughnessFactor=mix(roughnessFactor,0.9,cityFrost);
  metalnessFactor*=1.0-cityFrost*0.85;
  totalEmissiveRadiance*=1.0-cityFrost*0.35;
  diffuseColor.a=mix(diffuseColor.a,0.65,cityFrost*uCityFrostGlass);
}
