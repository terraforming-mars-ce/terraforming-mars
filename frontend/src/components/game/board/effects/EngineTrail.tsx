import { useRef, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../../hooks/useTextures";
import { usePrimitiveInstances } from "../PrimitiveManager";

interface EngineTrailProps {
  getShipPosition: () => THREE.Vector3 | null;
  getShipForward?: () => THREE.Vector3 | null;
  active: boolean;
  trailDuration?: number;
  startDelay?: number;
}

interface TrailParticle {
  position: THREE.Vector3;
  velocity: THREE.Vector3;
  rotation: number;
  rotationSpeed: number;
  age: number;
  maxAge: number;
  initialScale: number;
  alive: boolean;
}

const POOL_SIZE = 600;
// Distance between emission points — uniform spacing regardless of speed
const EMIT_SPACING = 0.03;
// Particles spawned per emission point (more during backward exhaust)
const PARTICLES_PER_POINT = 4;
const PARTICLES_PER_POINT_EXHAUST = 1;

function createTrailMaterial(smokeTexture: THREE.Texture | undefined): THREE.ShaderMaterial {
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
        vec3 smokeColor = vec3(0.95, 0.93, 0.9);
        // Brighten: mix texture toward white so dark smoke texture doesn't dominate
        vec3 brightened = mix(smokeColor, vec3(1.0), 0.4) * tex.rgb + smokeColor * 0.2;
        gl_FragColor = vec4(brightened, tex.a * opacity);
      }
    `,
    transparent: true,
    depthWrite: false,
    side: THREE.DoubleSide,
  });
}

const _tmpMatrix = new THREE.Matrix4();
const _tmpWorldPos = new THREE.Vector3();
const _tmpLerpPos = new THREE.Vector3();

const BACKWARD_EXHAUST_MS = 1200;

export default function EngineTrail({
  getShipPosition,
  getShipForward,
  active,
  trailDuration = 3500,
  startDelay = 0,
}: EngineTrailProps) {
  const { smoke: smokeTexture } = useTextures();
  const emitTimeRef = useRef(0);
  // Distance-based emission state
  const prevPosRef = useRef<THREE.Vector3 | null>(null);
  const distanceAccRef = useRef(0);

  const planeGeometry = useMemo(() => new THREE.PlaneGeometry(1, 1), []);
  const trailMaterial = useMemo(() => createTrailMaterial(smokeTexture), [smokeTexture]);

  const { setTransforms, setColors } = usePrimitiveInstances(
    "engine-trail-" + useMemo(() => Math.random().toString(36).slice(2, 8), []),
    planeGeometry,
    trailMaterial,
    101,
  );

  const matricesRef = useRef<THREE.Matrix4[]>([]);
  const colorsRef = useRef<THREE.Color[]>([]);
  const particlesRef = useRef<TrailParticle[]>([]);
  const initializedRef = useRef(false);
  const nextFreeRef = useRef(0);

  if (!initializedRef.current) {
    initializedRef.current = true;
    const matrices: THREE.Matrix4[] = [];
    const colors: THREE.Color[] = [];
    const particles: TrailParticle[] = [];

    for (let i = 0; i < POOL_SIZE; i++) {
      matrices.push(new THREE.Matrix4().makeScale(0, 0, 0));
      colors.push(new THREE.Color(0, 0, 0));
      particles.push({
        position: new THREE.Vector3(),
        velocity: new THREE.Vector3(),
        rotation: Math.random() * Math.PI * 2,
        rotationSpeed: (Math.random() - 0.5) * 0.6,
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

  // Find next free particle slot (ring buffer)
  const acquireParticle = (): TrailParticle | null => {
    const particles = particlesRef.current;
    for (let attempt = 0; attempt < POOL_SIZE; attempt++) {
      const idx = nextFreeRef.current % POOL_SIZE;
      nextFreeRef.current++;
      if (!particles[idx].alive) {
        return particles[idx];
      }
    }
    // Pool exhausted — recycle oldest (ring buffer wraps)
    const idx = nextFreeRef.current % POOL_SIZE;
    nextFreeRef.current++;
    return particles[idx];
  };

  const spawnAt = (pos: THREE.Vector3, emitElapsed: number) => {
    const useBackward = emitElapsed < BACKWARD_EXHAUST_MS && getShipForward;
    const forward = useBackward ? getShipForward!() : null;
    const count = forward ? PARTICLES_PER_POINT_EXHAUST : PARTICLES_PER_POINT;

    for (let j = 0; j < count; j++) {
      const p = acquireParticle();
      if (!p) {
        return;
      }
      p.alive = true;
      p.position.copy(pos);

      if (forward) {
        // Backward exhaust: particles shoot opposite to ship forward direction
        // Fade the backward force over the exhaust duration
        const exhaustT = 1.0 - emitElapsed / BACKWARD_EXHAUST_MS;
        const backwardSpeed = (0.15 + Math.random() * 0.08) * exhaustT;
        p.velocity.copy(forward).multiplyScalar(-backwardSpeed);
        p.velocity.x += (Math.random() - 0.5) * 0.01;
        p.velocity.y += (Math.random() - 0.5) * 0.01;
        p.velocity.z += (Math.random() - 0.5) * 0.01;
      } else {
        p.velocity.set(
          (Math.random() - 0.5) * 0.006,
          (Math.random() - 0.5) * 0.006,
          (Math.random() - 0.5) * 0.006,
        );
      }

      p.age = 0;
      p.maxAge = 0.6 + Math.random() * 0.6;
      p.initialScale = forward ? 0.025 + Math.random() * 0.025 : 0.02 + Math.random() * 0.02;
      p.rotation = Math.random() * Math.PI * 2;
      p.rotationSpeed = (Math.random() - 0.5) * 0.6;
    }
  };

  useFrame((_state, delta) => {
    const particles = particlesRef.current;
    const matrices = matricesRef.current;
    const colors = colorsRef.current;

    emitTimeRef.current += delta * 1000;
    const emitting =
      active &&
      emitTimeRef.current > startDelay &&
      emitTimeRef.current < startDelay + trailDuration;

    // Emission logic
    if (emitting) {
      const shipPos = getShipPosition();
      const emitElapsed = emitTimeRef.current - startDelay;

      if (shipPos) {
        if (!prevPosRef.current) {
          prevPosRef.current = shipPos.clone();
          // Immediately emit on first frame (don't skip)
          spawnAt(shipPos, emitElapsed);
        }

        const distanceTraveled = prevPosRef.current.distanceTo(shipPos);

        // During backward exhaust phase: emit every frame at ship position
        // (ship is barely moving, distance-based would be too slow)
        if (emitElapsed < BACKWARD_EXHAUST_MS) {
          spawnAt(shipPos, emitElapsed);
        }

        // Distance-based interpolated emission for the trail line
        const totalDistance = distanceAccRef.current + distanceTraveled;
        const pointsToSpawn = Math.floor(totalDistance / EMIT_SPACING);
        distanceAccRef.current = totalDistance - pointsToSpawn * EMIT_SPACING;

        if (pointsToSpawn > 0 && distanceTraveled > 0) {
          for (let i = 0; i < pointsToSpawn; i++) {
            const distAlongPath =
              (i + 1) * EMIT_SPACING - (totalDistance - distanceTraveled - distanceAccRef.current);
            const t = Math.max(0, Math.min(1, distAlongPath / distanceTraveled));
            _tmpLerpPos.lerpVectors(prevPosRef.current, shipPos, t);
            spawnAt(_tmpLerpPos, emitElapsed);
          }
        }

        prevPosRef.current.copy(shipPos);
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

      const fadeIn = Math.min(lifeRatio * 6, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 1.2);
      const opacity = fadeIn * fadeOut * 0.25;

      const scale = p.initialScale * (1 + lifeRatio * 4.0);

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
