attribute vec2 patchOrigin;
attribute float patchLayer;
attribute float patchBirth;
varying vec2 vBoardPos;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
varying vec3 vLocalPos;
varying vec3 vNormal;
void main() {
  vec2 local=position.xy+0.5;
  vec2 board=patchOrigin+local*uPatchSize;
  vec3 p=landscapeProject(board,uWaterLevel+0.00003);
  vBoardPos=board;
  vFieldLocal=local;
  vFieldLayer=patchLayer;
  vFieldBirth=patchBirth;
  vLocalPos=p;
  vNormal=normalize(p);
  gl_Position=projectionMatrix*modelViewMatrix*vec4(p,1.0);
}
