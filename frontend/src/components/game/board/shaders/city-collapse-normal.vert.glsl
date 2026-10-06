#include <beginnormal_vertex>
if (uCityTime >= cityCollapseStart) {
objectNormal = cityCollapseRotation() * objectNormal;
#ifdef USE_TANGENT
  objectTangent = cityCollapseRotation() * objectTangent;
#endif

}
