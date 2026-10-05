attribute float shadeStrength;
attribute float shadeTaper;
varying vec2 vUv;
varying float vStrength;
varying float vTaper;
void main() {
  vUv=uv;
  vStrength=shadeStrength;
  vTaper=shadeTaper;
  gl_Position=projectionMatrix*viewMatrix*instanceMatrix*vec4(position,1.0);
}
