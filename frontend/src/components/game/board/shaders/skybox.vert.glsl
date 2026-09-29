varying vec3 vDirection;

void main() {
  vDirection = position;
  // Ignore camera translation so the sky stays at infinity throughout travel.
  vec4 clipPosition = projectionMatrix * vec4(mat3(viewMatrix) * position, 1.0);
  gl_Position = clipPosition.xyww;
}
