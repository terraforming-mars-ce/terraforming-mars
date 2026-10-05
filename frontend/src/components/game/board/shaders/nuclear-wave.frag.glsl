uniform float uAge;
uniform float uDuration;
uniform float uDustDuration;
uniform float uRadius;
varying vec2 vGroundPosition;
float dustHash(vec2 p) {
  return fract(sin(dot(p,vec2(127.1,311.7)))*43758.5453);
}
float dustNoise(vec2 p) {
  vec2 cell = floor(p);
  vec2 f = fract(p);
  f = f*f*(3.0-2.0*f);
  return mix(mix(dustHash(cell),dustHash(cell+vec2(1.0,0.0)),f.x),
    mix(dustHash(cell+vec2(0.0,1.0)),dustHash(cell+vec2(1.0)),f.x),f.y);
}
void main() {
  float radius = length(vGroundPosition);
  float travel = clamp(uAge/uDuration,0.0,1.0);
  float front = uRadius*(1.0-pow(1.0-travel,1.35));
  float width = mix(0.013,0.025,travel);
  float ring = 1.0-smoothstep(width*0.3,width,abs(radius-front));
  float dust = exp(-pow((radius-front*0.94)/mix(0.045,0.085,travel),2.0));
  float waveOpacity = (ring*0.46+dust*0.15)*smoothstep(0.0,0.04,uAge)
    *(1.0-smoothstep(0.4,1.0,travel));

  // Invert the front's expansion curve so each patch fades after the wave passes it.
  float arrival = 1.0-pow(1.0-clamp(radius/uRadius,0.0,1.0),1.0/1.35);
  float dustAge = uAge-arrival*uDuration;
  float dustLife = clamp(dustAge/uDustDuration,0.0,1.0);
  vec2 drift = vGroundPosition/max(radius,0.001)*dustLife*0.025;
  vec2 p = vGroundPosition-drift;
  float clumps = dustNoise(p*25.0)*0.65+dustNoise(p*67.0+vec2(8.3,2.1))*0.35;
  float trailOpacity = 0.16*smoothstep(0.28,0.72,clumps)
    *smoothstep(0.0,0.09,dustLife)*(1.0-smoothstep(0.15,1.0,dustLife))
    *smoothstep(0.025,0.1,radius)*(1.0-smoothstep(0.55,1.0,arrival));
  float opacity = waveOpacity+trailOpacity;
  if (opacity < 0.001) {
    discard;
  }
  vec3 waveColor = mix(vec3(0.45,0.20,0.07),vec3(1.8,1.25,0.65),ring);
  vec3 dustColor = mix(vec3(0.22,0.065,0.025),vec3(0.48,0.19,0.075),clumps);
  gl_FragColor = vec4((waveColor*waveOpacity+dustColor*trailOpacity)/opacity,min(opacity,0.95));
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
