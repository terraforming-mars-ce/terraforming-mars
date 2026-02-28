import { useRef, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import BirdFlight from "./effects/BirdFlight";

interface LivingGreeneryTileProps {
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

const HEX_RADIUS = 0.166;

function isInsideHex(x: number, y: number, radius: number, margin: number = 0): boolean {
  const effectiveRadius = radius - margin;
  const absX = Math.abs(y);
  const absY = Math.abs(x);
  const q2 = effectiveRadius;
  const q1 = (effectiveRadius * Math.sqrt(3)) / 2;
  return absY <= q1 && q1 * absX + 0.5 * effectiveRadius * absY <= q1 * q2;
}

const FLOWER_COLORS = [
  new THREE.Color(0.95, 0.4, 0.6),
  new THREE.Color(0.95, 0.85, 0.2),
  new THREE.Color(0.9, 0.2, 0.2),
  new THREE.Color(0.95, 0.95, 0.9),
  new THREE.Color(0.7, 0.4, 0.9),
];

interface FlowerData {
  position: THREE.Vector3;
  color: THREE.Color;
  scale: number;
}

interface PondData {
  position: THREE.Vector3;
  radius: number;
  shimmerPhase: number;
}

export default function LivingGreeneryTile({ seed: seedProp }: LivingGreeneryTileProps) {
  const seed = useMemo(() => seedProp ?? Math.random() * 10000, [seedProp]);
  const pondMeshRefs = useRef<(THREE.Mesh | null)[]>([]);

  const flowers = useMemo((): FlowerData[] => {
    const rng = mulberry32(seed + 100);
    const count = 5 + Math.floor(rng() * 4);
    const result: FlowerData[] = [];

    for (let i = 0; i < count; i++) {
      for (let attempt = 0; attempt < 30; attempt++) {
        const x = (rng() - 0.5) * HEX_RADIUS * 1.5;
        const y = (rng() - 0.5) * HEX_RADIUS * 1.5;
        if (!isInsideHex(x, y, HEX_RADIUS, 0.02)) continue;

        result.push({
          position: new THREE.Vector3(x, y, 0.003),
          color: FLOWER_COLORS[Math.floor(rng() * FLOWER_COLORS.length)],
          scale: 0.004 + rng() * 0.003,
        });
        break;
      }
    }
    return result;
  }, [seed]);

  const ponds = useMemo((): PondData[] => {
    const rng = mulberry32(seed + 200);
    const count = 1 + (rng() > 0.5 ? 1 : 0);
    const result: PondData[] = [];

    for (let i = 0; i < count; i++) {
      for (let attempt = 0; attempt < 30; attempt++) {
        const x = (rng() - 0.5) * HEX_RADIUS * 1.0;
        const y = (rng() - 0.5) * HEX_RADIUS * 1.0;
        if (!isInsideHex(x, y, HEX_RADIUS, 0.04)) continue;

        let tooClose = false;
        for (const existing of result) {
          const dx = x - existing.position.x;
          const dy = y - existing.position.y;
          if (Math.sqrt(dx * dx + dy * dy) < 0.05) {
            tooClose = true;
            break;
          }
        }
        if (tooClose) continue;

        result.push({
          position: new THREE.Vector3(x, y, 0.002),
          radius: 0.012 + rng() * 0.01,
          shimmerPhase: rng() * Math.PI * 2,
        });
        break;
      }
    }
    return result;
  }, [seed]);

  const flowerGeometry = useMemo(() => new THREE.CircleGeometry(1, 8), []);
  const pondGeometry = useMemo(() => new THREE.CircleGeometry(1, 16), []);

  useFrame((state) => {
    const t = state.clock.elapsedTime;
    for (let i = 0; i < ponds.length; i++) {
      const mesh = pondMeshRefs.current[i];
      if (!mesh) continue;
      const mat = mesh.material as THREE.MeshStandardMaterial;
      mat.roughness = 0.1 + Math.sin(t * 1.5 + ponds[i].shimmerPhase) * 0.05;
    }
  });

  return (
    <group>
      {flowers.map((flower, i) => (
        <mesh
          key={`flower-${i}`}
          geometry={flowerGeometry}
          position={flower.position}
          scale={[flower.scale, flower.scale, 1]}
          renderOrder={15}
        >
          <meshStandardMaterial
            color={flower.color}
            roughness={0.7}
            metalness={0.0}
            side={THREE.DoubleSide}
          />
        </mesh>
      ))}

      {ponds.map((pond, i) => (
        <mesh
          key={`pond-${i}`}
          ref={(el) => {
            pondMeshRefs.current[i] = el;
          }}
          geometry={pondGeometry}
          position={pond.position}
          scale={[pond.radius, pond.radius, 1]}
          renderOrder={15}
        >
          <meshStandardMaterial
            color={new THREE.Color(0.15, 0.4, 0.7)}
            roughness={0.1}
            metalness={0.5}
            side={THREE.DoubleSide}
          />
        </mesh>
      ))}

      <BirdFlight count={3 + Math.floor(seed % 3)} seed={seed} />
    </group>
  );
}
