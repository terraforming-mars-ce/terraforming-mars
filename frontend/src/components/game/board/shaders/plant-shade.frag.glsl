varying vec2 vUv;
varying float vStrength;
varying float vTaper;
void main() {
  float r=length(vUv*2.0-1.0);
  float falloff=1.0-smoothstep(0.0,1.0,r);
  // Cast shadows are darkest at the trunk and fade towards the tip (uv.x runs away from the plant).
  float fade=1.0-vTaper*smoothstep(0.15,1.0,vUv.x);
  gl_FragColor=vec4(vec3(vStrength*falloff*falloff*fade),1.0);
}
