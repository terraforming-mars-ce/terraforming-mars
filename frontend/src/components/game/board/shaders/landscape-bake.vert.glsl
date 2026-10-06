attribute vec2 patchOrigin;
attribute float patchLayer;
attribute float patchBirth;
uniform vec4 uBakeRegion;
varying vec2 vBoardPos;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
varying vec3 vViewPosition;
// Lays the patches flat over the bake target: board space maps straight to the region.
void main() {
  vec2 local=position.xy+0.5;
  vec2 board=patchOrigin+local*uPatchSize;
  vBoardPos=board;
  vFieldLocal=local;
  vFieldLayer=patchLayer;
  vFieldBirth=patchBirth;
  vViewPosition=vec3(board,0.0);
  gl_Position=vec4((board-uBakeRegion.xy)/uBakeRegion.zw*2.0-1.0,0.0,1.0);
}
