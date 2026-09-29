varying vec2 vUv;
varying vec3 vLocalPosition;
varying vec3 vNormal;
varying vec3 vViewDirection;
void main() {
  vUv = uv;
  vLocalPosition = normalize(position);
  vec4 viewPosition = modelViewMatrix * vec4(position, 1.0);
  vNormal = normalize(normalMatrix * normal);
  vViewDirection = -viewPosition.xyz;
  gl_Position = projectionMatrix * viewPosition;
}
