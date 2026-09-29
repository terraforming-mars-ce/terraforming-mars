varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
void main() {
  if(landscapeBasinShore(vFieldLocal,vFieldLayer,vFieldBirth)>=uBasinReach.x) {discard;}
  gl_FragColor=vec4(0.0);
}
