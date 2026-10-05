// Replaces aomap_fragment: landscape occlusion dims indirect light, and a micro-shadow term
// (Naughty Dog, Uncharted 4) dims the sun too, since the board has no shadow maps.
float ambientOcclusion=landscapeAO;
reflectedLight.indirectDiffuse*=ambientOcclusion;
#if defined(USE_ENVMAP)&&defined(STANDARD)
  float occlusionNV=saturate(dot(geometryNormal,geometryViewDir));
  reflectedLight.indirectSpecular*=computeSpecularOcclusion(occlusionNV,ambientOcclusion,material.roughness);
#endif
vec3 sunView=normalize((viewMatrix*vec4(0.0,0.0,0.0,1.0)).xyz+vViewPosition);
float microShadow=saturate(dot(geometryNormal,sunView)+2.0*ambientOcclusion*ambientOcclusion-1.0);
reflectedLight.directDiffuse*=mix(1.0,microShadow,0.75);
reflectedLight.directSpecular*=mix(1.0,microShadow,0.75);
