import { useEffect, useMemo } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";

const STAR_COUNT = 4000;
const SEED = 0x5eed1234;

const vertexShader = /* glsl */ `
attribute float aSize;
attribute vec3 aColor;
uniform float uPixelRatio;
varying vec3 vColor;

void main() {
  vColor = aColor;
  // Ignore camera translation so the stars stay at infinity, like the EXR skybox.
  vec4 clipPosition = projectionMatrix * vec4(mat3(viewMatrix) * position, 1.0);
  gl_Position = clipPosition.xyww;
  gl_PointSize = aSize * uPixelRatio;
}
`;

const fragmentShader = /* glsl */ `
varying vec3 vColor;

void main() {
  float falloff = 1.0 - smoothstep(0.15, 0.5, length(gl_PointCoord - 0.5));
  if (falloff <= 0.0) {
    discard;
  }
  gl_FragColor = vec4(vColor * falloff, 1.0);
  #include <tonemapping_fragment>
  #include <colorspace_fragment>
}
`;

function mulberry32(seed: number) {
  let state = seed;
  return () => {
    state = (state + 0x6d2b79f5) | 0;
    let t = Math.imul(state ^ (state >>> 15), 1 | state);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function createStarGeometry() {
  const random = mulberry32(SEED);
  const positions = new Float32Array(STAR_COUNT * 3);
  const colors = new Float32Array(STAR_COUNT * 3);
  const sizes = new Float32Array(STAR_COUNT);
  const cool = new THREE.Color(0.72, 0.82, 1);
  const warm = new THREE.Color(1, 0.9, 0.72);
  const color = new THREE.Color();
  for (let i = 0; i < STAR_COUNT; i++) {
    const z = random() * 2 - 1;
    const angle = random() * Math.PI * 2;
    const ring = Math.sqrt(1 - z * z);
    positions[i * 3] = ring * Math.cos(angle);
    positions[i * 3 + 1] = z;
    positions[i * 3 + 2] = ring * Math.sin(angle);

    const magnitude = random();
    const brightness = 0.2 + 0.8 * magnitude ** 4;
    const tint = random();
    color.setRGB(1, 1, 1);
    if (tint < 0.3) {
      color.lerp(cool, 1 - tint / 0.3);
    } else if (tint > 0.75) {
      color.lerp(warm, (tint - 0.75) / 0.25);
    }
    colors[i * 3] = color.r * brightness;
    colors[i * 3 + 1] = color.g * brightness;
    colors[i * 3 + 2] = color.b * brightness;
    sizes[i] = 1.2 + 2.6 * magnitude ** 6;
  }
  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
  geometry.setAttribute("aColor", new THREE.BufferAttribute(colors, 3));
  geometry.setAttribute("aSize", new THREE.BufferAttribute(sizes, 1));
  return geometry;
}

/** Cheap stand-in for the EXR skybox on the low graphics tier. */
export default function ProceduralStarfield() {
  const dpr = useThree((state) => state.viewport.dpr);
  const geometry = useMemo(() => createStarGeometry(), []);
  const material = useMemo(
    () =>
      new THREE.ShaderMaterial({
        vertexShader,
        fragmentShader,
        uniforms: { uPixelRatio: { value: 1 } },
        depthTest: false,
        depthWrite: false,
        fog: false,
      }),
    [],
  );

  useEffect(() => {
    material.uniforms.uPixelRatio.value = dpr;
  }, [material, dpr]);

  useEffect(
    () => () => {
      geometry.dispose();
      material.dispose();
    },
    [geometry, material],
  );

  return (
    <points
      geometry={geometry}
      material={material}
      frustumCulled={false}
      renderOrder={-1000}
      raycast={() => {}}
    />
  );
}
