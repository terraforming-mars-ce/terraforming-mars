uniform mat4 uTileHighlightCameraWorld;
uniform mat4 uTileHighlightProjectionInverse;
varying vec3 vTileHighlightWorldPosition;
//#pragma body
void main() {
  tileHighlightSourceVertex();
  vec4 world=uTileHighlightCameraWorld*uTileHighlightProjectionInverse*gl_Position;
  vTileHighlightWorldPosition=world.xyz/world.w;
}
