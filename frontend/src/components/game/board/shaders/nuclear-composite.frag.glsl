uniform sampler2D uCloud;
uniform sampler2D uDepth;
uniform vec2 uCloudTexel;
uniform mat4 uInverseProjection;
varying vec2 vUv;
float distanceAt(vec2 uv) {
  vec4 p=uInverseProjection*vec4(uv*2.0-1.0,texture2D(uDepth,uv).r*2.0-1.0,1.0);
  return length(p.xyz/p.w);
}
void main() {
  float center=distanceAt(vUv);
  vec4 cloud=vec4(0.0);
  float total=0.0;
  for (int x=0;x<2;x++) {
    for (int y=0;y<2;y++) {
      vec2 sampleUv=vUv+(vec2(float(x),float(y))-0.5)*uCloudTexel;
      float weight=exp(-abs(distanceAt(sampleUv)-center)*100.0)+0.0001;
      cloud+=texture2D(uCloud,sampleUv)*weight;
      total+=weight;
    }
  }
  cloud/=total;
  if (cloud.a < 0.001) { discard; }
  gl_FragColor=vec4(cloud.rgb/cloud.a,cloud.a);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
  gl_FragColor.rgb *= gl_FragColor.a;
}
