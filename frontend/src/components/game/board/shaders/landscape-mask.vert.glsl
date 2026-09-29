attribute vec2 patchOrigin;
attribute float patchLayer;
attribute float patchBirth;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
void main() {
  vFieldLocal=position.xy+0.5;
  vFieldLayer=patchLayer;
  vFieldBirth=patchBirth;
  vec2 board=patchOrigin+vFieldLocal*uPatchSize;
  gl_Position=projectionMatrix*modelViewMatrix*vec4(landscapeProject(board,0.0),1.0);
}
