uniform float uSweep;
uniform float uSweepFade;
//#pragma body
diffuseColor.rgb=marsClimateSurface(diffuseColor.rgb,vMapUv);
if(uSweep>0.0) {
  float band=exp(-pow((vMapUv.y-(uSweep*1.3-0.15))/0.05,2.0));
  float glow=mix(band,sin(uSweep*PI)*0.4,uSweepFade);
  diffuseColor.rgb+=vec3(0.35,0.5,0.3)*glow;
}
