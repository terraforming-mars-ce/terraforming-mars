import { useRef, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../../hooks/useTextures";

interface DustEffectProps {
  position: THREE.Vector3;
  normal: THREE.Vector3;
  duration?: number;
  particleColor?: THREE.Color;
  onComplete?: () => void;
}

interface SmokeParticle {
  mesh: THREE.Mesh;
  velocity: THREE.Vector3;
  rotationSpeed: number;
  lifetime: number;
  maxLifetime: number;
  initialScale: number;
}

export default function DustEffect({
  position,
  normal,
  duration = 2600,
  particleColor,
  onComplete,
}: DustEffectProps) {
  const { scene } = useThree();
  const particlesRef = useRef<SmokeParticle[]>([]);
  const startTimeRef = useRef<number | null>(null);
  const completedRef = useRef(false);
  const groupRef = useRef<THREE.Group | null>(null);

  const { smoke: smokeTexture } = useTextures();

  useEffect(() => {
    const group = new THREE.Group();
    group.position.copy(position);
    scene.add(group);
    groupRef.current = group;

    const particles: SmokeParticle[] = [];
    const particleCount = 45;

    const tangent = new THREE.Vector3();
    if (Math.abs(normal.y) < 0.9) {
      tangent.crossVectors(normal, new THREE.Vector3(0, 1, 0)).normalize();
    } else {
      tangent.crossVectors(normal, new THREE.Vector3(1, 0, 0)).normalize();
    }
    const bitangent = new THREE.Vector3().crossVectors(normal, tangent).normalize();

    for (let i = 0; i < particleCount; i++) {
      const geometry = new THREE.PlaneGeometry(0.15, 0.15);
      const material = new THREE.MeshBasicMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        depthTest: false,
        blending: THREE.NormalBlending,
        side: THREE.DoubleSide,
        color: particleColor ?? new THREE.Color(0.72, 0.35, 0.18),
      });

      const mesh = new THREE.Mesh(geometry, material);

      const angle = Math.random() * Math.PI * 2;
      const radius = Math.random() * 0.12;
      const offsetX = Math.cos(angle) * radius;
      const offsetY = Math.sin(angle) * radius;
      mesh.position.copy(tangent.clone().multiplyScalar(offsetX));
      mesh.position.add(bitangent.clone().multiplyScalar(offsetY));

      mesh.rotation.z = Math.random() * Math.PI * 2;

      const initialScale = 0.3 + Math.random() * 0.4;
      mesh.scale.setScalar(initialScale);

      const spreadAngle = Math.random() * Math.PI * 2;
      const spreadRadius = 0.03 + Math.random() * 0.04;
      const velocity = new THREE.Vector3()
        .copy(normal)
        .multiplyScalar(0.02 + Math.random() * 0.015)
        .add(tangent.clone().multiplyScalar(Math.cos(spreadAngle) * spreadRadius))
        .add(bitangent.clone().multiplyScalar(Math.sin(spreadAngle) * spreadRadius));

      group.add(mesh);

      particles.push({
        mesh,
        velocity,
        rotationSpeed: (Math.random() - 0.5) * 1.4,
        lifetime: 0,
        maxLifetime: 1.3 + Math.random() * 1.3,
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
  }, [scene, position, normal, smokeTexture]);

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

        const fadeIn = Math.min(lifeRatio * 4, 1);
        const fadeOut = 1 - Math.pow(lifeRatio, 2);
        const opacity = fadeIn * fadeOut * 0.7;
        (particle.mesh.material as THREE.MeshBasicMaterial).opacity = opacity;

        const scale = particle.initialScale * (1 + lifeRatio * 1.5);
        particle.mesh.scale.setScalar(scale);

        particle.velocity.multiplyScalar(0.98);
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
