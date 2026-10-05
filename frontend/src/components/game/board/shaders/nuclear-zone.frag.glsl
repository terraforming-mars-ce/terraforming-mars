uniform float uSeed;
uniform float uScorch;
uniform sampler2D uMars;
uniform vec4 uGouges[5];
varying vec2 vNuclearUv;
varying vec2 vGroundPosition;
varying float vNuclearCoverage;
varying float vNuclearApron;
varying vec3 vNuclearPlanetDirection;
float hash(vec2 p) { return fract(sin(dot(p, vec2(127.1,311.7))) * 43758.5453); }
float noise(vec2 p) {
  vec2 i = floor(p), f = fract(p);
  f = f*f*(3.0-2.0*f);
  return mix(mix(hash(i),hash(i+vec2(1,0)),f.x),
    mix(hash(i+vec2(0,1)),hash(i+vec2(1,1)),f.x),f.y);
}
//#pragma body
{
  float r = vNuclearUv.y;
  vec3 planetDirection=normalize(vNuclearPlanetDirection);
  vec2 marsUv=vec2(atan(planetDirection.z,-planetDirection.x)/6.28318530718,
    1.0-acos(clamp(planetDirection.y,-1.0,1.0))/3.14159265359);
  vec3 mars=marsClimateSurface(texture2D(uMars,marsUv).rgb,marsUv);
  vec2 p = vGroundPosition / 0.1;
  vec2 seed = vec2(mod(uSeed,997.0)*0.037,mod(uSeed,113.0));
  vec2 warped = p + vec2(noise(p*3.1+seed),noise(p*3.7-seed))*0.12;
  float soil = noise(warped*4.0+seed)*0.6 + noise(warped*10.0-seed)*0.4;
  float grainVisibility=1.0-smoothstep(0.35,1.0,length(fwidth(p*48.0)));
  float grain = mix(0.5,noise(p*24.0+seed)*0.65 + noise(p*48.0-seed)*0.35,grainVisibility);
  float centerWeight=1.0-smoothstep(0.28,0.78,r+(soil-0.5)*0.08);
  float soot = smoothstep(0.36,0.72,noise(warped*3.5-seed))*(1.0-smoothstep(0.6,1.1,r));
  vec3 rock = mix(vec3(0.105,0.037,0.018),vec3(0.22,0.082,0.037),soil);
  rock *= mix(0.96,1.04,grain);
  rock = mix(rock,vec3(0.055,0.023,0.014),soot*0.35*(1.0-centerWeight*0.65));
  float grooves = 0.0;
  for (int i=0;i<5;i++) {
    vec2 a=uGouges[i].xy/0.1, b=uGouges[i].zw/0.1;
    vec2 ab=b-a;
    float t=clamp(dot(warped-a,ab)/dot(ab,ab),0.0,1.0);
    float distance=length(warped-a-ab*t);
    float width=0.025+0.014*float(i%3);
    float broken=smoothstep(0.32,0.65,noise(warped*17.0+seed+float(i)));
    grooves=max(grooves,(1.0-smoothstep(width*0.2,width,distance))*broken
      *smoothstep(0.0,0.2,t)*(1.0-smoothstep(0.72,1.0,t)));
  }
  rock *= 1.0-grooves*0.22;
  vec3 floorColor=vec3(0.065,0.025,0.013)*mix(0.97,1.03,grain);
  rock=mix(rock,floorColor,centerWeight*0.88);
  rock = mix(rock,mars*mix(0.92,1.0,soil),smoothstep(0.65,1.12,r));
  float alpha = 1.0;
  if (uScorch>0.5) {
    float t=vNuclearApron;
    float soilPatch=noise(warped*6.0+seed)*0.65+noise(warped*19.0-seed)*0.35;
    rock=mix(rock,mars,smoothstep(0.0,0.65,t));
    float raggedFade=t+(soilPatch-0.5)*0.48*smoothstep(0.0,0.25,t);
    alpha=(1.0-smoothstep(0.0,1.0,raggedFade))
      *mix(1.0,0.20+soilPatch*0.38,smoothstep(0.0,0.7,t))
      *(1.0-smoothstep(0.85,1.0,t))*vNuclearCoverage;
    if(alpha<0.002) { discard; }
  }
  float ao=mix(0.74,1.0,smoothstep(0.28,1.0,r));
  diffuseColor.rgb=rock*ao;
  diffuseColor.a*=alpha;
}
