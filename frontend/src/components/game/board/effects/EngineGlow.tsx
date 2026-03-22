import { useRef, useMemo, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../../hooks/useTextures";
import { usePrimitiveInstances } from "../PrimitiveManager";

interface EngineGlowProps {
  getShipPosition: () => THREE.Vector3 | null;
  getShipQuat: () => THREE.Quaternion | null;
  active: boolean;
  fadeInDuration?: number;
}

interface FlameParticle {
  position: THREE.Vector3;
  velocity: THREE.Vector3;
  rotation: number;
  rotationSpeed: number;
  age: number;
  maxAge: number;
  initialScale: number;
  alive: boolean;
}

const POOL_SIZE = 60;
const PARTICLES_PER_FRAME = 4;
const GLOW_OFFSET = 0.035;

function createFlameMaterial(smokeTexture: THREE.Texture | undefined): THREE.ShaderMaterial {
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
        float lifeRatio = vInstanceColor.b;

        vec3 coreColor = vec3(0.7, 0.85, 1.0);
        vec3 yellowColor = vec3(1.0, 0.95, 0.4);
        vec3 orangeColor = vec3(1.0, 0.6, 0.1);
        vec3 redColor = vec3(0.8, 0.15, 0.0);

        vec3 color;
        if (lifeRatio < 0.15) {
          color = mix(coreColor, yellowColor, lifeRatio / 0.15);
        } else if (lifeRatio < 0.4) {
          color = mix(yellowColor, orangeColor, (lifeRatio - 0.15) / 0.25);
        } else {
          color = mix(orangeColor, redColor, (lifeRatio - 0.4) / 0.6);
        }

        color += vec3(0.3) * (1.0 - lifeRatio);

        gl_FragColor = vec4(color * tex.rgb, tex.a * opacity);
      }
    `,
    transparent: true,
    depthWrite: false,
    side: THREE.DoubleSide,
    blending: THREE.AdditiveBlending,
  });
}

const _tmpMatrix = new THREE.Matrix4();
const _tmpWorldPos = new THREE.Vector3();
const _tmpRearDir = new THREE.Vector3();
const _tmpPos = new THREE.Vector3();
const _tmpPerp1 = new THREE.Vector3();
const _tmpPerp2 = new THREE.Vector3();

export default function EngineGlow({
  getShipPosition,
  getShipQuat,
  active,
  fadeInDuration = 500,
}: EngineGlowProps) {
  const { scene } = useThree();
  const { smoke: smokeTexture } = useTextures();
  const elapsedRef = useRef(0);

  const planeGeometry = useMemo(() => new THREE.PlaneGeometry(1, 1), []);
  const material = useMemo(() => createFlameMaterial(smokeTexture), [smokeTexture]);

  const { setTransforms, setColors } = usePrimitiveInstances(
    "engine-glow-" + useMemo(() => Math.random().toString(36).slice(2, 8), []),
    planeGeometry,
    material,
    103,
  );

  const matricesRef = useRef<THREE.Matrix4[]>([]);
  const colorsRef = useRef<THREE.Color[]>([]);
  const particlesRef = useRef<FlameParticle[]>([]);
  const initializedRef = useRef(false);

  // Core sprite for bright center
  const coreSpriteRef = useRef<THREE.Sprite | null>(null);

  useEffect(() => {
    const coreMat = new THREE.SpriteMaterial({
      map: smokeTexture,
      transparent: true,
      opacity: 0,
      depthWrite: false,
      blending: THREE.AdditiveBlending,
      color: new THREE.Color(0.7, 0.85, 1.0),
    });
    const sprite = new THREE.Sprite(coreMat);
    sprite.renderOrder = 104;
    sprite.scale.set(0.025, 0.025, 1);
    scene.add(sprite);
    coreSpriteRef.current = sprite;

    return () => {
      coreMat.dispose();
      scene.remove(sprite);
    };
  }, [scene, smokeTexture]);

  if (!initializedRef.current) {
    initializedRef.current = true;

    const matrices: THREE.Matrix4[] = [];
    const colors: THREE.Color[] = [];
    const particles: FlameParticle[] = [];

    for (let i = 0; i < POOL_SIZE; i++) {
      matrices.push(new THREE.Matrix4().makeScale(0, 0, 0));
      colors.push(new THREE.Color(0, 0, 0));
      particles.push({
        position: new THREE.Vector3(),
        velocity: new THREE.Vector3(),
        rotation: Math.random() * Math.PI * 2,
        rotationSpeed: (Math.random() - 0.5) * 1.0,
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

  useFrame((state, delta) => {
    const particles = particlesRef.current;
    const matrices = matricesRef.current;
    const colors = colorsRef.current;
    const coreSprite = coreSpriteRef.current;

    if (!active) {
      for (let i = 0; i < POOL_SIZE; i++) {
        matrices[i].makeScale(0, 0, 0);
        colors[i].setRGB(0, 0, 0);
      }
      if (coreSprite) {
        (coreSprite.material as THREE.SpriteMaterial).opacity = 0;
      }
      elapsedRef.current = 0;
      setTransforms(matrices);
      setColors(colors);
      return;
    }

    elapsedRef.current += delta * 1000;
    const time = state.clock.elapsedTime;

    const shipPos = getShipPosition();
    const shipQuat = getShipQuat();
    if (!shipPos || !shipQuat) {
      return;
    }

    const fadeT = Math.min(elapsedRef.current / fadeInDuration, 1.0);
    const intensity = fadeT * fadeT;

    // Compute rear direction and perpendicular frame
    _tmpRearDir.set(0, 0, 1).applyQuaternion(shipQuat).normalize();

    if (Math.abs(_tmpRearDir.y) < 0.9) {
      _tmpPerp1.crossVectors(_tmpRearDir, new THREE.Vector3(0, 1, 0)).normalize();
    } else {
      _tmpPerp1.crossVectors(_tmpRearDir, new THREE.Vector3(1, 0, 0)).normalize();
    }
    _tmpPerp2.crossVectors(_tmpRearDir, _tmpPerp1).normalize();

    // Nozzle position
    _tmpPos.copy(shipPos).addScaledVector(_tmpRearDir, GLOW_OFFSET);

    // Update core sprite
    if (coreSprite) {
      coreSprite.position.copy(_tmpPos);
      const flicker = 0.85 + 0.15 * Math.sin(time * 20);
      (coreSprite.material as THREE.SpriteMaterial).opacity = 0.9 * intensity * flicker;
      const coreScale = 0.025 * (0.9 + 0.1 * Math.sin(time * 15));
      coreSprite.scale.set(coreScale, coreScale, 1);
    }

    // Spawn new particles
    const spawnCount = Math.max(1, Math.round(PARTICLES_PER_FRAME * intensity));
    let spawned = 0;
    for (let i = 0; i < POOL_SIZE && spawned < spawnCount; i++) {
      const p = particles[i];
      if (!p.alive) {
        p.alive = true;

        const angle = Math.random() * Math.PI * 2;
        const radialOffset = Math.random() * 0.008;
        p.position.copy(_tmpPos);
        p.position.addScaledVector(_tmpPerp1, Math.cos(angle) * radialOffset);
        p.position.addScaledVector(_tmpPerp2, Math.sin(angle) * radialOffset);

        const speed = (0.15 + Math.random() * 0.15) * (0.5 + intensity * 0.5);
        p.velocity.copy(_tmpRearDir).multiplyScalar(speed);
        p.velocity.addScaledVector(_tmpPerp1, (Math.random() - 0.5) * 0.04);
        p.velocity.addScaledVector(_tmpPerp2, (Math.random() - 0.5) * 0.04);

        p.age = 0;
        p.maxAge = 0.08 + Math.random() * 0.12;
        p.initialScale = 0.012 + Math.random() * 0.008;
        p.rotation = Math.random() * Math.PI * 2;
        p.rotationSpeed = (Math.random() - 0.5) * 1.0;
        spawned++;
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

      const fadeIn = Math.min(lifeRatio * 10, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 1.5);
      const jitter = 0.85 + Math.random() * 0.15;
      const opacity = fadeIn * fadeOut * jitter * 0.8;

      const growShrink = (1 + lifeRatio * 0.3) * (1 - lifeRatio * 0.3);
      const scale = p.initialScale * growShrink;

      _tmpWorldPos.copy(p.position);
      _tmpMatrix.makeScale(scale, scale, scale);
      _tmpMatrix.setPosition(_tmpWorldPos);
      matrices[i].copy(_tmpMatrix);

      const normalizedRotation = ((p.rotation % (Math.PI * 2)) + Math.PI * 2) % (Math.PI * 2);
      colors[i].setRGB(opacity, normalizedRotation / (Math.PI * 2), lifeRatio);
    }

    setTransforms(matrices);
    setColors(colors);
  });

  return null;
}
