uniform mat4 uNuclearClipTransform;
uniform mat3 uNuclearViewRotation;
uniform mat3 uNuclearWorldRotation;
void main() {
  nuclearOriginalMain();
  gl_Position = uNuclearClipTransform * gl_Position;
  #ifdef NUCLEAR_VIEW_NORMAL
    #ifndef FLAT_SHADED
      vNormal = normalize(uNuclearViewRotation * vNormal);
    #endif
  #endif
  #ifdef NUCLEAR_WORLD_NORMAL
    vWorldNormal = normalize(uNuclearWorldRotation * vWorldNormal);
  #endif
}
