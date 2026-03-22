import { useRef, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../../hooks/useTextures";

interface ExhaustEffectProps {
  position: THREE.Vector3;
  normal: THREE.Vector3;
  duration?: number;
  particleCount?: number;
  particleScale?: number;
  onComplete?: () => void;
}

interface ExhaustParticle {
  mesh: THREE.Mesh;
  velocity: THREE.Vector3;
  rotationSpeed: number;
  lifetime: number;
  maxLifetime: number;
  initialScale: number;
}

export default function ExhaustEffect({
  position,
  normal,
  duration = 1500,
  particleCount = 20,
  particleScale = 1.0,
  onComplete,
}: ExhaustEffectProps) {
  const { scene } = useThree();
  const particlesRef = useRef<ExhaustParticle[]>([]);
  const startTimeRef = useRef<number | null>(null);
  const completedRef = useRef(false);
  const groupRef = useRef<THREE.Group | null>(null);

  const { smoke: smokeTexture } = useTextures();

  const initialPositionRef = useRef(position.clone());
  const initialNormalRef = useRef(normal.clone());

  useEffect(() => {
    const pos = initialPositionRef.current;
    const nrm = initialNormalRef.current;

    const group = new THREE.Group();
    group.position.copy(pos);
    scene.add(group);
    groupRef.current = group;

    const particles: ExhaustParticle[] = [];

    const tangent = new THREE.Vector3();
    if (Math.abs(nrm.y) < 0.9) {
      tangent.crossVectors(nrm, new THREE.Vector3(0, 1, 0)).normalize();
    } else {
      tangent.crossVectors(nrm, new THREE.Vector3(1, 0, 0)).normalize();
    }
    const bitangent = new THREE.Vector3().crossVectors(nrm, tangent).normalize();

    // Exhaust direction is opposite to surface normal (downward from ship)
    const exhaustDir = nrm.clone().negate();

    for (let i = 0; i < particleCount; i++) {
      const pSize = 0.1 * particleScale;
      const geometry = new THREE.PlaneGeometry(pSize, pSize);
      const material = new THREE.MeshBasicMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        depthTest: false,
        blending: THREE.NormalBlending,
        side: THREE.DoubleSide,
        color: new THREE.Color(0.9, 0.85, 0.7),
      });

      const mesh = new THREE.Mesh(geometry, material);

      const angle = Math.random() * Math.PI * 2;
      // Spawn in a ring around the ship, not through it
      const minRadius = 0.04 * particleScale;
      const maxRadius = 0.1 * particleScale;
      const radius = minRadius + Math.random() * (maxRadius - minRadius);
      mesh.position.copy(tangent.clone().multiplyScalar(Math.cos(angle) * radius));
      mesh.position.add(bitangent.clone().multiplyScalar(Math.sin(angle) * radius));

      mesh.renderOrder = 20;
      mesh.rotation.z = Math.random() * Math.PI * 2;

      const initialScale = 0.2 + Math.random() * 0.3;
      mesh.scale.setScalar(initialScale);

      // Velocity: spread outward radially from ship center, not through it
      const spreadAngle = Math.random() * Math.PI * 2;
      const outwardSpeed = 0.02 + Math.random() * 0.02;
      const velocity = new THREE.Vector3()
        .add(tangent.clone().multiplyScalar(Math.cos(spreadAngle) * outwardSpeed))
        .add(bitangent.clone().multiplyScalar(Math.sin(spreadAngle) * outwardSpeed))
        .addScaledVector(exhaustDir, 0.005 + Math.random() * 0.005);

      group.add(mesh);

      particles.push({
        mesh,
        velocity,
        rotationSpeed: (Math.random() - 0.5) * 1.4,
        lifetime: 0,
        maxLifetime: 0.5 + Math.random() * 0.8,
        initialScale,
      });
    }

    particlesRef.current = particles;

    return () => {
      particles.forEach((p) => {
        p.mesh.geometry.dispose();
        (p.mesh.material as THREE.Material).dispose();
      });
      scene.remove(group);
    };
  }, [scene, smokeTexture]);

  useFrame((state, delta) => {
    if (startTimeRef.current === null) {
      startTimeRef.current = state.clock.elapsedTime * 1000;
    }

    const elapsed = state.clock.elapsedTime * 1000 - startTimeRef.current;
    const particles = particlesRef.current;

    for (const particle of particles) {
      particle.lifetime += delta;
      const lifeRatio = particle.lifetime / particle.maxLifetime;

      if (lifeRatio < 1) {
        particle.mesh.position.add(particle.velocity.clone().multiplyScalar(delta));
        particle.mesh.rotation.z += particle.rotationSpeed * delta;

        const fadeIn = Math.min(lifeRatio * 5, 1);
        const fadeOut = 1 - Math.pow(lifeRatio, 2);
        const opacity = fadeIn * fadeOut * 0.6;
        (particle.mesh.material as THREE.MeshBasicMaterial).opacity = opacity;

        const scale = particle.initialScale * (1 + lifeRatio * 2.0);
        particle.mesh.scale.setScalar(scale);

        particle.velocity.multiplyScalar(0.97);
      } else {
        (particle.mesh.material as THREE.MeshBasicMaterial).opacity = 0;
      }
    }

    if (elapsed >= duration && !completedRef.current) {
      completedRef.current = true;
      onComplete?.();
    }
  });

  return null;
}
