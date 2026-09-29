attribute vec2 boardPosition;
attribute vec4 landscapeField;
attribute float shoreDistance;
varying float vShoreDistance;
varying vec2 vBoardPos;
varying vec4 vLandscapeField;
//#pragma body
#include <begin_vertex>
vBoardPos=boardPosition;
vLandscapeField=landscapeField;
vShoreDistance=shoreDistance;
