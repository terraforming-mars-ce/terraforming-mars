import { useRef, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";

interface BirdFlightProps {
  count?: number;
  seed?: number;
}

function mulberry32(seed: number): () => number {
  return function () {
    let t = (seed += 0x6d2b79f5);
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

interface BirdData {
  orbitRadius: number;
  orbitHeight: number;
  orbitSpeed: number;
  orbitPhase: number;
  orbitTilt: number;
  flapSpeed: number;
  bodyColor: THREE.Color;
}

export default function BirdFlight({ count = 4, seed = 12345 }: BirdFlightProps) {
  const groupRef = useRef<THREE.Group>(null);
  const birdRefs = useRef<(THREE.Group | null)[]>([]);
  const wingRefs = useRef<(THREE.Mesh | null)[]>([]);

  const birds = useMemo((): BirdData[] => {
    const rng = mulberry32(seed);
    return Array.from({ length: count }, () => ({
      orbitRadius: 0.05 + rng() * 0.07,
      orbitHeight: 0.06 + rng() * 0.04,
      orbitSpeed: 0.4 + rng() * 0.4,
      orbitPhase: rng() * Math.PI * 2,
      orbitTilt: (rng() - 0.5) * 0.3,
      flapSpeed: 3.0 + rng() * 2.0,
      bodyColor: new THREE.Color().setHSL(0.0, 0.0, 0.1 + rng() * 0.2),
    }));
  }, [count, seed]);

  const bodyGeometry = useMemo(() => new THREE.ConeGeometry(0.002, 0.008, 4), []);
  const wingGeometry = useMemo(() => new THREE.PlaneGeometry(0.01, 0.003), []);

  useFrame((state) => {
    const t = state.clock.elapsedTime;

    for (let i = 0; i < birds.length; i++) {
      const bird = birds[i];
      const birdGroup = birdRefs.current[i];
      if (!birdGroup) continue;

      const angle = t * bird.orbitSpeed + bird.orbitPhase;
      const x = Math.cos(angle) * bird.orbitRadius;
      const y = Math.sin(angle) * bird.orbitRadius * (0.6 + Math.sin(bird.orbitTilt) * 0.4);
      const z = bird.orbitHeight + Math.sin(angle * 0.5) * 0.01;

      birdGroup.position.set(x, y, z);
      birdGroup.rotation.z = angle + Math.PI / 2;

      const leftWing = wingRefs.current[i * 2];
      const rightWing = wingRefs.current[i * 2 + 1];
      if (leftWing && rightWing) {
        const flapAngle = Math.sin(t * bird.flapSpeed * Math.PI * 2) * 0.6;
        leftWing.rotation.x = flapAngle;
        rightWing.rotation.x = -flapAngle;
      }
    }
  });

  return (
    <group ref={groupRef}>
      {birds.map((bird, i) => (
        <group
          key={i}
          ref={(el) => {
            birdRefs.current[i] = el;
          }}
        >
          <mesh geometry={bodyGeometry} rotation={[Math.PI / 2, 0, 0]}>
            <meshStandardMaterial color={bird.bodyColor} roughness={0.8} />
          </mesh>
          <mesh
            ref={(el) => {
              wingRefs.current[i * 2] = el;
            }}
            geometry={wingGeometry}
            position={[0.003, 0, 0]}
          >
            <meshStandardMaterial color={bird.bodyColor} roughness={0.8} side={THREE.DoubleSide} />
          </mesh>
          <mesh
            ref={(el) => {
              wingRefs.current[i * 2 + 1] = el;
            }}
            geometry={wingGeometry}
            position={[-0.003, 0, 0]}
          >
            <meshStandardMaterial color={bird.bodyColor} roughness={0.8} side={THREE.DoubleSide} />
          </mesh>
        </group>
      ))}
    </group>
  );
}
