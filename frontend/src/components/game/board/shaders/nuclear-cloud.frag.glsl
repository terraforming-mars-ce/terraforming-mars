uniform sampler2D uDepth;
uniform mat4 uInverseProjection;
uniform mat4 uCameraWorld;
uniform mat4 uWorldToLocal;
uniform float uAge;
uniform float uSeed;
varying vec2 vUv;

float hash(vec3 p) {
  p = fract(p * 0.1031);
  p += dot(p, p.yzx + 33.33);
  return fract((p.x+p.y)*p.z);
}
float noise(vec3 p) {
  vec3 i=floor(p), f=fract(p);
  f=f*f*(3.0-2.0*f);
  return mix(mix(mix(hash(i),hash(i+vec3(1,0,0)),f.x),
                 mix(hash(i+vec3(0,1,0)),hash(i+vec3(1,1,0)),f.x),f.y),
             mix(mix(hash(i+vec3(0,0,1)),hash(i+vec3(1,0,1)),f.x),
                 mix(hash(i+vec3(0,1,1)),hash(i+vec3(1,1,1)),f.x),f.y),f.z);
}
float turbulence(vec3 p) {
  return noise(p)*0.57+noise(p*2.07+4.1)*0.28+noise(p*4.13-7.3)*0.15;
}
float ellipsoid(vec3 p, vec3 center, vec3 radii) {
  return (length((p-center)/radii)-1.0)*min(radii.x,min(radii.y,radii.z));
}
vec3 field(vec3 p) {
  float t = max(uAge,0.0);
  float rise = smoothstep(0.12,2.1,t);
  float spread = smoothstep(0.18,1.35,t);
  float height = mix(0.055,0.21,rise);
  float width = mix(0.045,0.15,spread);
  float capDepth = mix(0.030,0.063,spread);
  float angle = atan(p.y,p.x);
  float radial = length(p.xy);
  vec3 flow = p;
  float roll = atan(p.z-height,radial-width*0.72);
  flow.xy *= 1.0+0.08*sin(roll-t*1.6);
  flow.z -= t*0.035;
  vec3 offset = vec3(mod(uSeed,71.0),mod(uSeed,127.0),mod(uSeed,31.0));
  float n = turbulence(flow*36.0+offset);
  float folds = sin(angle*7.0+t*0.5)*0.006+sin(angle*11.0-t*0.3)*0.004;

  // Broad billows sit above the flat underside and rolled, overhanging lip.
  float crownLift = (sin(p.x*68.0+sin(p.y*43.0)-t*0.4)*0.008
    +cos(p.y*75.0-t*0.3)*0.006)*spread;
  float crown = ellipsoid(p,vec3(0,0,height+crownLift),vec3(width+folds,width,capDepth));
  crown = max(crown,height-0.014-p.z);
  vec2 rimSection = vec2((radial-width*0.76)/(width*0.27),
    (p.z-height+0.006)/mix(0.014,0.023,spread));
  float rim = (length(rimSection)-1.0)*0.023;
  float cap = min(crown,rim)+(n-0.48)*0.026;

  float foot = 0.013*(1.0-smoothstep(0.0,0.045,p.z));
  float collar = 0.008*smoothstep(height-0.045,height-0.010,p.z);
  float billows = (sin(p.z*91.0-t*1.8+sin(angle*3.0))*0.008
    +sin(p.z*143.0-t*1.1+angle*5.0)*0.004)*rise;
  float stemWidth = mix(0.025,0.035,smoothstep(0.12,0.6,t))+foot+collar+billows;
  vec2 stemOffset = vec2(sin(p.z*39.0-t*1.3),cos(p.z*53.0-t*0.9))*0.012*rise;
  float stem = max(length(p.xy-stemOffset)-stemWidth, max(-0.012-p.z,p.z-height));
  stem += (n-0.48)*0.024;
  float fireball = length(p-vec3(0,0,0.045))-mix(0.022,0.075,smoothstep(0.0,0.22,t));
  fireball += smoothstep(0.25,0.65,t)*0.3;
  float shape = min(min(cap,stem),fireball+(n-0.48)*0.016);
  float density = (1.0-smoothstep(-0.010,0.012,shape));
  density *= smoothstep(-0.018,-0.004,p.z);
  density *= 1.0-smoothstep(3.6,6.2,t);
  float breakup = smoothstep(4.0,6.2,t);
  density *= 1.0-smoothstep(0.75-breakup*0.6,0.9-breakup*0.6,1.0-n);
  float hotLife = 1.0-smoothstep(2.5,4.7,t);
  float heat = hotLife*smoothstep(0.38,0.76,n);
  heat *= mix(1.0,0.45,smoothstep(height-0.005,height+0.04,p.z));
  float underside = exp(-pow((p.z-height+0.012)/0.020,2.0))
    * (1.0-smoothstep(width*0.65,width,radial));
  heat = max(heat,underside*0.65*hotLife);
  // Hold a fully incandescent cloud before revealing the turbulent pockets of fire.
  float incandescent = 1.0-smoothstep(0.65,2.0,t);
  heat = mix(heat,0.94+0.06*n,incandescent);

  // A filled, low dust blanket follows the curved ground rather than leaving a hollow ring.
  float groundZ = p.z+0.01+radial*radial/4.04;
  float groundRadius = mix(0.035,0.235,smoothstep(0.0,1.6,t));
  float bed = (length(vec3(p.xy/groundRadius,(groundZ-0.016)/0.024))-1.0)*0.024;
  bed += (n-0.48)*0.014;
  float ground = (1.0-smoothstep(-0.010,0.012,bed))*smoothstep(-0.008,0.002,groundZ);
  ground *= 1.0-smoothstep(3.0,6.2,t);
  float groundShare = ground/max(density+ground,0.001);
  heat *= 1.0-groundShare*0.75*smoothstep(0.25,0.8,t);
  density = max(density,ground*0.85);

  float haloRadius = 0.07+t*0.11;
  float halo = exp(-pow((radial-haloRadius)/0.018,2.0)-pow((p.z-height*0.8)/0.007,2.0));
  float haloLife = smoothstep(0.3,0.55,t)*(1.0-smoothstep(0.7,1.3,t));
  density = max(density,halo*haloLife*0.3);
  return vec3(density,heat,halo*haloLife);
}
void main() {
  if (uAge<0.0 || uAge>=6.2) { discard; }
  vec4 viewFar = uInverseProjection*vec4(vUv*2.0-1.0,1.0,1.0);
  vec3 rayWorld = normalize((uCameraWorld*vec4(viewFar.xyz/viewFar.w,0.0)).xyz);
  vec3 originWorld = uCameraWorld[3].xyz;
  vec3 ro = (uWorldToLocal*vec4(originWorld,1.0)).xyz;
  vec3 rd = normalize((uWorldToLocal*vec4(rayWorld,0.0)).xyz);
  vec3 inv = 1.0 / (rd+vec3(0.0000001));
  vec3 a = (vec3(-0.36,-0.36,-0.06)-ro)*inv;
  vec3 b = (vec3(0.36,0.36,0.50)-ro)*inv;
  vec3 lo = min(a,b), hi=max(a,b);
  float entry = max(0.0,max(lo.x,max(lo.y,lo.z)));
  float leave = min(hi.x,min(hi.y,hi.z));
  float depth = texture2D(uDepth,vUv).r;
  vec4 sceneView = uInverseProjection*vec4(vUv*2.0-1.0,depth*2.0-1.0,1.0);
  float sceneDistance = length(sceneView.xyz/sceneView.w);
  leave = min(leave,sceneDistance);
  if (leave<=entry) { discard; }
  float stepSize = (leave-entry)/48.0;
  float distance = entry+stepSize*0.5;
  float whiteHeat = 1.0-smoothstep(0.65,1.75,uAge);
  vec4 result=vec4(0.0);
  for (int i=0;i<48;i++) {
    vec3 p = ro+rd*distance;
    vec3 f=field(p);
    if (f.x>0.005) {
      float alpha = 1.0-exp(-f.x*stepSize*60.0);
      float upper = smoothstep(0.0,0.42,p.z);
      vec3 smoke = mix(vec3(0.050,0.009,0.003),vec3(0.115,0.029,0.011),upper);
      float light = clamp(0.4+(f.x-field(p+vec3(-0.016,0.012,0.024)).x)*1.7,0.18,1.0);
      vec3 fire = mix(vec3(2.8,0.22,0.005),vec3(5.0,1.7,0.22),smoothstep(0.25,0.95,f.y));
      fire = mix(fire,vec3(8.0,6.6,4.3),whiteHeat*smoothstep(0.65,0.92,f.y));
      vec3 color = smoke*light + fire*f.y*f.y;
      color = mix(color,vec3(0.75,0.70,0.60),clamp(f.z*2.0,0.0,1.0));
      result.rgb += (1.0-result.a)*color*alpha;
      result.a += (1.0-result.a)*alpha;
      if (result.a>0.985) { break; }
    }
    distance+=stepSize;
  }
  gl_FragColor=result;
}
