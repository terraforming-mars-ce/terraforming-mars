uniform sampler2D normalSampler;
uniform sampler2D uClimateNoise;
uniform sampler2DArray uIceAlbedo;
uniform sampler2DArray uIceDetail;
uniform float time;
uniform vec3 eye;
uniform vec3 sunDirection;
uniform vec3 sunColor;
uniform float sunIntensity;
uniform float rf0;
uniform float uIce;
uniform float uIceAge;
uniform float uIceReach;
uniform float uClarity;
uniform float uAlgae;
uniform float uWaves;
uniform float uFoam;
uniform float uShimmer;
varying vec2 vBoardPos;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
varying vec3 vLocalPos;
varying vec3 vNormal;
vec3 clarityRamp(vec3 murky,vec3 teal,vec3 blue) {
  return mix(mix(murky,teal,smoothstep(0.0,0.55,uClarity)),blue,smoothstep(0.45,1.0,uClarity));
}
vec4 iceLayer(sampler2DArray layers,vec2 uv) {
  vec4 solid=texture(layers,vec3(uv.x,-uv.y,0.0));
  vec4 thin=texture(layers,vec3(uv.x,-uv.y,1.0));
  return mix(solid,thin,uIceAge);
}
float shorelineFoam(float distanceToShore) {
  vec2 drift=vec2(time*0.006,-time*0.004);
  float variation=smoothstep(0.35,0.65,texture2D(uClimateNoise,vBoardPos*1.7).r);
  float breakup=texture2D(uClimateNoise,vBoardPos*12.0+drift).r;
  float bubbles=texture2D(uClimateNoise,vBoardPos*65.0-drift*1.3).r;
  // Spread arrival times continuously across the board without tile seams or small timing pockets.
  float arrivalOffset=dot(vBoardPos,vec2(2.1,2.9));
  float waveDistance=distanceToShore+(variation-0.5)*0.012+arrivalOffset*0.075;
  // Positive time moves constant-phase crests toward smaller shore distances.
  float phase=(waveDistance+time*0.005625)/0.075;
  float crestDistance=(fract(phase+0.5)-0.5)*0.075+(breakup-0.5)*0.003;
  float aa=max(fwidth(waveDistance),0.00025);
  float crest=1.0-smoothstep(0.0006,0.00125+aa,abs(crestDistance));
  float wake=smoothstep(0.0,0.0015+aa,crestDistance)*
    (1.0-smoothstep(0.002,0.006,crestDistance));
  float age=clamp(crestDistance/0.006,0.0,1.0);
  float porous=smoothstep(0.44+age*0.18,0.62+age*0.18,breakup*0.65+bubbles*0.35);
  float brokenCrest=smoothstep(0.46,0.62,breakup);
  float offshore=1.0-smoothstep(0.012,0.035,distanceToShore);
  float resolved=1.0-smoothstep(0.003,0.012,fwidth(waveDistance));
  float activity=mix(0.65,1.0,uWaves)*mix(0.65,1.0,uFoam);
  float thawed=smoothstep(0.0,0.05,1.0-uIce);
  return clamp((crest*brokenCrest+wake*porous*0.3)*offshore*resolved*
    activity*thawed*(0.65+variation*0.35)*0.35,0.0,0.3);
}
void main() {
  vec4 field=landscapeField(vFieldLocal,vFieldLayer,vFieldBirth);
  float aa=max(fwidth(field.y),0.00025);
  float water=1.0-smoothstep(-0.009-aa,0.001+aa,field.y);
  float depth=max(0.0,uWaterLevel-field.x);
  water*=smoothstep(0.0,0.0003,depth);
  if(water<0.001) {discard;}
  vec2 uv=vBoardPos*38.0;
  vec3 n1=texture2D(normalSampler,uv+vec2(time*0.025,time*0.017)).xyz*2.0-1.0;
  vec3 n2=texture2D(normalSampler,uv*0.63-vec2(time*0.016,time*0.023)).xyz*2.0-1.0;
  vec3 radial=normalize(vNormal);
  vec3 tangent=normalize(cross(vec3(0.0,1.0,0.0),radial));
  vec3 bitangent=cross(radial,tangent);
  float waveAmplitude=mix(0.025,0.085,uWaves);
  vec3 normal=normalize(radial+tangent*(n1.x+n2.x)*waveAmplitude+bitangent*(n1.y+n2.y)*waveAmplitude);
  vec3 view=normalize(eye-vLocalPos);
  float daylight=0.3+0.7*max(radial.z,0.0);
  float shallow=1.0-smoothstep(0.0003,0.014,depth);
  vec3 deep=clarityRamp(vec3(0.12,0.09,0.05),vec3(0.04,0.09,0.08),vec3(0.05,0.09,0.1))+vec3(0.025,0.06,0.07);
  vec3 shallowColor=clarityRamp(vec3(0.26,0.21,0.13),vec3(0.13,0.27,0.3),vec3(0.12,0.3,0.4));
  vec3 color=mix(deep,shallowColor,shallow*0.65)*sunColor*daylight;
  // Algae stays as sparse, faint patches in the shallows rather than a continuous rim.
  float bloom=smoothstep(0.72,0.9,texture2D(normalSampler,vBoardPos*2.5+time*0.002).x*0.5+0.5);
  color=mix(color,vec3(0.1,0.2,0.08)*daylight,uAlgae*shallow*bloom*0.3);
  float fres=rf0+(1.0-rf0)*pow(1.0-max(dot(normal,view),0.0),5.0);
  color=mix(color,vec3(0.12,0.2,0.26)*daylight,fres*0.65);
  float spec=pow(max(dot(reflect(-normalize(sunDirection),normal),view),0.0),100.0);
  color+=sunColor*spec*sunIntensity*0.35*(1.0+uShimmer*2.0);
  float foam=shorelineFoam(max(-field.y,0.0));
  vec3 foamColor=vec3(0.92,0.95,0.91)*sunColor*daylight;
  color=mix(color,foamColor,foam);
  // Ice melts from the shore inward as uIce falls.
  // The ice textures only add crack and plate detail; the colour comes from the climate stage.
  if(uIce>0.0) {
    vec2 iceUv=vBoardPos*12.0;
    vec2 iceUvB=mat2(0.6,-0.8,0.8,0.6)*vBoardPos*7.3+0.41;
    vec4 iceDetail=mix(iceLayer(uIceDetail,iceUv),iceLayer(uIceDetail,iceUvB),0.4);
    float inland=max(-field.y,0.0)+(iceDetail.a-0.5)*0.004;
    float edge=(1.0-uIce)*uIceReach;
    // Pockets enclosed by three lakes lie deeper than the melt edge reaches; the last floes there
    // dissolve with the remaining ice instead of popping when uIce reaches zero.
    float ice=smoothstep(edge,edge+0.003,inland)*smoothstep(0.0,0.15,uIce);
    // Shore ice is thin: it fades to translucent so the lakebed shows through at the waterline.
    float thickness=smoothstep(0.0,0.02,inland-edge);
    vec2 slope=(iceDetail.rg*2.0-1.0)*0.12;
    vec3 iceNormal=normalize(radial+tangent*slope.x+bitangent*slope.y);
    float plates=dot(mix(iceLayer(uIceAlbedo,iceUv),iceLayer(uIceAlbedo,iceUvB),0.4).rgb,vec3(0.3333));
    vec3 solidIce=vec3(0.8,0.87,0.92);
    vec3 thinIce=mix(deep,vec3(0.55,0.68,0.75),0.6);
    vec3 iceBase=mix(thinIce,mix(solidIce,thinIce,uIceAge*0.75),thickness)*(0.9+0.2*plates);
    vec3 sun=normalize(sunDirection);
    float diffuse=0.55+0.45*max(dot(iceNormal,sun),0.0);
    vec3 iceColor=iceBase*diffuse*sunColor*daylight;
    float iceFresnel=pow(1.0-max(dot(iceNormal,view),0.0),4.0);
    iceColor=mix(iceColor,vec3(0.7,0.8,0.9)*daylight,iceFresnel*0.4);
    float roughness=mix(0.25,0.6,iceDetail.b);
    float glint=pow(max(dot(reflect(-sun,iceNormal),view),0.0),mix(220.0,30.0,roughness));
    iceColor+=sunColor*glint*sunIntensity*0.4*(1.0-roughness);
    color=mix(color,iceColor,ice);
    water*=mix(1.0,mix(0.55,1.0,thickness),ice);
  }
  gl_FragColor=vec4(color,water);
}
