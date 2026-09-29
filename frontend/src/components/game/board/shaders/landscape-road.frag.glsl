uniform float uConnectionTime;
varying float vConnectionBirth;
//#pragma body
#include <alphamap_fragment>
diffuseColor.a*=smoothstep(0.0,0.6,uConnectionTime-vConnectionBirth);
