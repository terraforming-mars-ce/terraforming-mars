attribute vec2 patchOrigin;
attribute float patchLayer;
attribute float patchBirth;
varying vec2 vBoardPos;
varying vec2 vFieldLocal;
varying float vFieldLayer;
varying float vFieldBirth;
//#pragma body
vec2 local=position.xy+0.5;
vec2 board=patchOrigin+local*uPatchSize;
float h=landscapeField(local,patchLayer,patchBirth).x;
vec3 transformed=landscapeProject(board,h);
vBoardPos=board;
vFieldLocal=local;
vFieldLayer=patchLayer;
vFieldBirth=patchBirth;
