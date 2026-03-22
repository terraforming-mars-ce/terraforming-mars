import { Suspense, useRef, useMemo, useEffect, useState } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useGLTF } from "@react-three/drei";
import { SkeletonUtils } from "three-stdlib";
import { useTextures } from "../../../hooks/useTextures";
import { SPHERE_RADIUS, easeOutCubic } from "./boardConstants";
import { useAsteroidEventStore } from "../../../stores/asteroidEventStore";
import { useSoundEffects } from "../../../hooks/useSoundEffects";

const ASTEROID_MODEL_PATH = "/assets/models/asteroid.glb";
const ASTEROID_SCALE = 0.15;
const FLIGHT_DURATION = 1.8;
const IMPACT_DUST_DURATION = 3000;
const DUST_PARTICLE_COUNT = 60;

interface DustParticle {
  mesh: THREE.Mesh;
  velocity: THREE.Vector3;
  rotationSpeed: number;
  lifetime: number;
  maxLifetime: number;
  initialScale: number;
}

function ImpactDust({
  position,
  normal,
  onComplete,
  smokeTexture,
}: {
  position: THREE.Vector3;
  normal: THREE.Vector3;
  onComplete: () => void;
  smokeTexture: THREE.Texture;
}) {
  const { scene } = useThree();
  const particlesRef = useRef<DustParticle[]>([]);
  const startTimeRef = useRef<number | null>(null);
  const completedRef = useRef(false);
  const groupRef = useRef<THREE.Group | null>(null);

  useEffect(() => {
    const group = new THREE.Group();
    group.position.copy(position);
    scene.add(group);
    groupRef.current = group;

    const particles: DustParticle[] = [];

    const tangent = new THREE.Vector3();
    if (Math.abs(normal.y) < 0.9) {
      tangent.crossVectors(normal, new THREE.Vector3(0, 1, 0)).normalize();
    } else {
      tangent.crossVectors(normal, new THREE.Vector3(1, 0, 0)).normalize();
    }
    const bitangent = new THREE.Vector3().crossVectors(normal, tangent).normalize();

    for (let i = 0; i < DUST_PARTICLE_COUNT; i++) {
      const geometry = new THREE.PlaneGeometry(0.2, 0.2);
      const material = new THREE.MeshBasicMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        depthTest: false,
        blending: THREE.NormalBlending,
        side: THREE.DoubleSide,
        color: new THREE.Color(0.8, 0.4, 0.15),
      });

      const mesh = new THREE.Mesh(geometry, material);

      const angle = Math.random() * Math.PI * 2;
      const radius = Math.random() * 0.08;
      mesh.position
        .copy(tangent.clone().multiplyScalar(Math.cos(angle) * radius))
        .add(bitangent.clone().multiplyScalar(Math.sin(angle) * radius));

      mesh.renderOrder = 20;
      mesh.rotation.z = Math.random() * Math.PI * 2;
      const initialScale = 0.4 + Math.random() * 0.6;
      mesh.scale.setScalar(initialScale);

      const spreadAngle = Math.random() * Math.PI * 2;
      const spreadRadius = 0.04 + Math.random() * 0.06;
      const velocity = new THREE.Vector3()
        .copy(normal)
        .multiplyScalar(0.03 + Math.random() * 0.025)
        .add(tangent.clone().multiplyScalar(Math.cos(spreadAngle) * spreadRadius))
        .add(bitangent.clone().multiplyScalar(Math.sin(spreadAngle) * spreadRadius));

      group.add(mesh);

      particles.push({
        mesh,
        velocity,
        rotationSpeed: (Math.random() - 0.5) * 1.8,
        lifetime: 0,
        maxLifetime: 1.5 + Math.random() * 1.5,
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
  }, [scene, smokeTexture, position, normal]);

  useFrame((state, delta) => {
    if (startTimeRef.current === null) {
      startTimeRef.current = state.clock.elapsedTime * 1000;
    }

    const elapsed = state.clock.elapsedTime * 1000 - startTimeRef.current;

    for (const particle of particlesRef.current) {
      particle.lifetime += delta;
      const lifeRatio = particle.lifetime / particle.maxLifetime;

      if (lifeRatio < 1) {
        particle.mesh.position.add(particle.velocity.clone().multiplyScalar(delta));
        particle.mesh.rotation.z += particle.rotationSpeed * delta;

        const fadeIn = Math.min(lifeRatio * 5, 1);
        const fadeOut = 1 - Math.pow(lifeRatio, 2);
        (particle.mesh.material as THREE.MeshBasicMaterial).opacity = fadeIn * fadeOut * 0.85;

        particle.mesh.scale.setScalar(particle.initialScale * (1 + lifeRatio * 2));
        particle.velocity.multiplyScalar(0.97);
      } else {
        (particle.mesh.material as THREE.MeshBasicMaterial).opacity = 0;
      }
    }

    if (elapsed >= IMPACT_DUST_DURATION && !completedRef.current) {
      completedRef.current = true;
      onComplete();
    }
  });

  return null;
}

function AsteroidFlight({
  onImpact,
}: {
  onImpact: (position: THREE.Vector3, normal: THREE.Vector3) => void;
}) {
  const { scene: asteroidScene } = useGLTF(ASTEROID_MODEL_PATH);
  const { playAsteroidImpactSound } = useSoundEffects();

  const groupRef = useRef<THREE.Group>(null);
  const flightStartRef = useRef<number | null>(null);
  const soundPlayedRef = useRef(false);

  const asteroidModel = useMemo(() => {
    const clone = SkeletonUtils.clone(asteroidScene);
    const box = new THREE.Box3().setFromObject(clone);
    const size = box.getSize(new THREE.Vector3());
    const maxDim = Math.max(size.x, size.y, size.z);
    const scale = ASTEROID_SCALE / maxDim;
    clone.scale.setScalar(scale);

    clone.traverse((child: THREE.Object3D) => {
      if (child instanceof THREE.Mesh) {
        child.material = Array.isArray(child.material)
          ? child.material.map((m: THREE.Material) => m.clone())
          : child.material.clone();
      }
    });

    return clone;
  }, [asteroidScene]);

  const impactPoint = useMemo(() => {
    const theta = Math.random() * Math.PI * 0.6 - Math.PI * 0.3;
    const phi = Math.random() * Math.PI * 0.4;
    return new THREE.Vector3(
      SPHERE_RADIUS * Math.sin(phi) * Math.cos(theta),
      SPHERE_RADIUS * Math.sin(phi) * Math.sin(theta),
      SPHERE_RADIUS * Math.cos(phi),
    );
  }, []);

  const startPoint = useMemo(() => {
    return new THREE.Vector3(
      impactPoint.x + 3 + Math.random() * 2,
      impactPoint.y + 4 + Math.random() * 2,
      impactPoint.z + 6,
    );
  }, [impactPoint]);

  const rotationAxis = useMemo(
    () =>
      new THREE.Vector3(Math.random() - 0.5, Math.random() - 0.5, Math.random() - 0.5).normalize(),
    [],
  );

  useFrame((state) => {
    if (!groupRef.current) {
      return;
    }

    if (flightStartRef.current === null) {
      flightStartRef.current = state.clock.elapsedTime;
    }

    const elapsed = state.clock.elapsedTime - flightStartRef.current;
    const t = Math.min(elapsed / FLIGHT_DURATION, 1);
    const eased = easeOutCubic(t);

    const pos = new THREE.Vector3().lerpVectors(startPoint, impactPoint, eased);
    groupRef.current.position.copy(pos);

    const tumbleSpeed = 4 + (1 - t) * 8;
    groupRef.current.rotateOnAxis(rotationAxis, tumbleSpeed * 0.016);

    const scale = 1 - eased * 0.3;
    groupRef.current.scale.setScalar(scale);

    if (t >= 1 && !soundPlayedRef.current) {
      soundPlayedRef.current = true;
      void playAsteroidImpactSound();
      onImpact(impactPoint, impactPoint.clone().normalize());
    }
  });

  return (
    <group ref={groupRef}>
      <primitive object={asteroidModel} />
    </group>
  );
}

export default function AsteroidImpact() {
  const currentEvent = useAsteroidEventStore((s) => s.queue[0]);
  const dequeue = useAsteroidEventStore((s) => s.dequeue);
  const { smoke: smokeTexture } = useTextures();

  const [phase, setPhase] = useState<"idle" | "flying" | "dust">("idle");
  const impactDataRef = useRef<{ position: THREE.Vector3; normal: THREE.Vector3 } | null>(null);

  useEffect(() => {
    if (currentEvent && phase === "idle") {
      setPhase("flying");
    }
  }, [currentEvent, phase]);

  const handleImpact = (position: THREE.Vector3, normal: THREE.Vector3) => {
    impactDataRef.current = { position, normal };
    setPhase("dust");
  };

  const handleDustComplete = () => {
    impactDataRef.current = null;
    setPhase("idle");
    dequeue();
  };

  if (!currentEvent) {
    return null;
  }

  return (
    <>
      {phase === "flying" && (
        <Suspense fallback={null}>
          <AsteroidFlight onImpact={handleImpact} />
        </Suspense>
      )}
      {phase === "dust" && impactDataRef.current && (
        <ImpactDust
          position={impactDataRef.current.position}
          normal={impactDataRef.current.normal}
          onComplete={handleDustComplete}
          smokeTexture={smokeTexture}
        />
      )}
    </>
  );
}
