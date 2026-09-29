attribute float roadEdge;
attribute float connectionBirth;
varying float vRoadEdge;
varying float vConnectionBirth;
varying vec2 vShoulderBoard;
//#pragma body
#include <begin_vertex>
vRoadEdge=roadEdge;
vConnectionBirth=connectionBirth;
vShoulderBoard=uv/35.0;
