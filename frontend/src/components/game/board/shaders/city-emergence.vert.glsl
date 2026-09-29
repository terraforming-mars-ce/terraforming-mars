attribute float cityBirth;
uniform float uCityTime;
//#pragma body
#include <begin_vertex>
float cityProgress=clamp((uCityTime-cityBirth)/0.8,0.0,1.0);
transformed.z-=0.08*pow(1.0-cityProgress,3.0);
