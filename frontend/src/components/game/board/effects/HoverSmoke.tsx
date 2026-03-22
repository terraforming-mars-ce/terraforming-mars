import { useRef, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../../hooks/useTextures";
import { usePrimitiveInstances } from "../PrimitiveManager";

interface HoverSmokeProps {
  position: THREE.Vector3;
  normal: THREE.Vector3;
  active: boolean;
}

interface SmokeParticle {
  position: THREE.Vector3;
  velocity: THREE.Vector3;
  rotation: number;
  rotationSpeed: number;
  age: number;
  maxAge: number;
  initialScale: number;
  alive: boolean;
}

const POOL_SIZE = 100;
const PARTICLES_PER_FRAME = 2;

function createHoverSmokeMaterial(smokeTexture: THREE.Texture | undefined): THREE.ShaderMaterial {
  return new THREE.ShaderMaterial({
    uniforms: {
      uTexture: { value: smokeTexture ?? null },
    },
    vertexShader: `
      varying vec2 vUv;
      varying vec3 vInstanceColor;

      void main() {
        vUv = uv;
        vInstanceColor = instanceColor;

        vec3 instancePos = vec3(
          instanceMatrix[3][0],
          instanceMatrix[3][1],
          instanceMatrix[3][2]
        );
        float scaleX = length(vec3(instanceMatrix[0][0], instanceMatrix[0][1], instanceMatrix[0][2]));

        vec4 mvInstancePos = modelViewMatrix * vec4(instancePos, 1.0);

        float rot = vInstanceColor.g * 6.283185;
        float cr = cos(rot);
        float sr = sin(rot);
        vec2 centered = position.xy;
        vec2 rotated = vec2(
          centered.x * cr - centered.y * sr,
          centered.x * sr + centered.y * cr
        );

        mvInstancePos.xy += rotated * scaleX;
        gl_Position = projectionMatrix * mvInstancePos;
      }
    `,
    fragmentShader: `
      uniform sampler2D uTexture;
      varying vec2 vUv;
      varying vec3 vInstanceColor;

      void main() {
        vec4 tex = texture2D(uTexture, vUv);
        float opacity = vInstanceColor.r;
        vec3 smokeColor = vec3(0.72, 0.55, 0.35);
        gl_FragColor = vec4(smokeColor * tex.rgb, tex.a * opacity);
      }
    `,
    transparent: true,
    depthWrite: false,
    side: THREE.DoubleSide,
  });
}

const _tmpMatrix = new THREE.Matrix4();
const _tmpWorldPos = new THREE.Vector3();

export default function HoverSmoke({ position, normal, active }: HoverSmokeProps) {
  const { smoke: smokeTexture } = useTextures();

  const planeGeometry = useMemo(() => new THREE.PlaneGeometry(1, 1), []);
  const material = useMemo(() => createHoverSmokeMaterial(smokeTexture), [smokeTexture]);

  const { setTransforms, setColors } = usePrimitiveInstances(
    "hover-smoke-" + useMemo(() => Math.random().toString(36).slice(2, 8), []),
    planeGeometry,
    material,
    102,
  );

  const matricesRef = useRef<THREE.Matrix4[]>([]);
  const colorsRef = useRef<THREE.Color[]>([]);
  const particlesRef = useRef<SmokeParticle[]>([]);
  const initializedRef = useRef(false);

  const posRef = useRef(position.clone());
  const normalRef = useRef(normal.clone());

  // Compute tangent frame once
  const tangentRef = useRef(new THREE.Vector3());
  const bitangentRef = useRef(new THREE.Vector3());

  if (!initializedRef.current) {
    initializedRef.current = true;

    const nrm = normalRef.current;
    if (Math.abs(nrm.y) < 0.9) {
      tangentRef.current.crossVectors(nrm, new THREE.Vector3(0, 1, 0)).normalize();
    } else {
      tangentRef.current.crossVectors(nrm, new THREE.Vector3(1, 0, 0)).normalize();
    }
    bitangentRef.current.crossVectors(nrm, tangentRef.current).normalize();

    const matrices: THREE.Matrix4[] = [];
    const colors: THREE.Color[] = [];
    const particles: SmokeParticle[] = [];

    for (let i = 0; i < POOL_SIZE; i++) {
      matrices.push(new THREE.Matrix4().makeScale(0, 0, 0));
      colors.push(new THREE.Color(0, 0, 0));
      particles.push({
        position: new THREE.Vector3(),
        velocity: new THREE.Vector3(),
        rotation: Math.random() * Math.PI * 2,
        rotationSpeed: (Math.random() - 0.5) * 0.8,
        age: 0,
        maxAge: 1,
        initialScale: 0,
        alive: false,
      });
    }

    matricesRef.current = matrices;
    colorsRef.current = colors;
    particlesRef.current = particles;
  }

  useFrame((_state, delta) => {
    const particles = particlesRef.current;
    const matrices = matricesRef.current;
    const colors = colorsRef.current;
    const tangent = tangentRef.current;
    const bitangent = bitangentRef.current;
    const pos = posRef.current;

    // Spawn new particles in a ring
    if (active) {
      let spawned = 0;
      for (let i = 0; i < POOL_SIZE && spawned < PARTICLES_PER_FRAME; i++) {
        const p = particles[i];
        if (!p.alive) {
          p.alive = true;

          // Spawn in a ring around the ship position
          const angle = Math.random() * Math.PI * 2;
          const radius = 0.04 + Math.random() * 0.04;
          p.position.copy(pos);
          p.position.addScaledVector(tangent, Math.cos(angle) * radius);
          p.position.addScaledVector(bitangent, Math.sin(angle) * radius);

          // Velocity: radially outward + slight upward
          const outSpeed = 0.015 + Math.random() * 0.015;
          p.velocity.set(0, 0, 0);
          p.velocity.addScaledVector(tangent, Math.cos(angle) * outSpeed);
          p.velocity.addScaledVector(bitangent, Math.sin(angle) * outSpeed);
          p.velocity.addScaledVector(normalRef.current, 0.003 + Math.random() * 0.003);

          p.age = 0;
          p.maxAge = 0.8 + Math.random() * 0.7;
          p.initialScale = 0.015 + Math.random() * 0.015;
          p.rotation = Math.random() * Math.PI * 2;
          p.rotationSpeed = (Math.random() - 0.5) * 0.8;
          spawned++;
        }
      }
    }

    // Update all particles
    for (let i = 0; i < POOL_SIZE; i++) {
      const p = particles[i];

      if (!p.alive) {
        matrices[i].makeScale(0, 0, 0);
        colors[i].setRGB(0, 0, 0);
        continue;
      }

      p.age += delta;

      if (p.age >= p.maxAge) {
        p.alive = false;
        matrices[i].makeScale(0, 0, 0);
        colors[i].setRGB(0, 0, 0);
        continue;
      }

      const lifeRatio = p.age / p.maxAge;

      p.position.x += p.velocity.x * delta;
      p.position.y += p.velocity.y * delta;
      p.position.z += p.velocity.z * delta;

      p.rotation += p.rotationSpeed * delta;
      p.velocity.multiplyScalar(1 - 0.5 * delta);

      const fadeIn = Math.min(lifeRatio * 5, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 1.3);
      const opacity = fadeIn * fadeOut * 0.5;

      const scale = p.initialScale * (1 + lifeRatio * 3.0);

      _tmpWorldPos.copy(p.position);
      _tmpMatrix.makeScale(scale, scale, scale);
      _tmpMatrix.setPosition(_tmpWorldPos);
      matrices[i].copy(_tmpMatrix);

      const normalizedRotation = ((p.rotation % (Math.PI * 2)) + Math.PI * 2) % (Math.PI * 2);
      colors[i].setRGB(opacity, normalizedRotation / (Math.PI * 2), 0);
    }

    setTransforms(matrices);
    setColors(colors);
  });

  return null;
}
