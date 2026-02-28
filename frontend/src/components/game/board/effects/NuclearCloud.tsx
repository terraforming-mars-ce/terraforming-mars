import { useRef, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../../hooks/useTextures";

interface NuclearCloudProps {
  isNewlyPlaced: boolean;
}

interface CloudParticle {
  sprite: THREE.Sprite;
  velocity: THREE.Vector3;
  rotationSpeed: number;
  age: number;
  maxAge: number;
  initialScale: number;
  delay: number;
  isMushroom: boolean;
}

const MUSHROOM_PARTICLE_COUNT = 18;
const HAZE_PARTICLE_COUNT = 10;
const MUSHROOM_LIFETIME = 7.0;
const BLAST_FLASH_DURATION = 0.3;
const SHOCKWAVE_DURATION = 1.5;
const MUSHROOM_COLOR = new THREE.Color(0.12, 0.1, 0.08);
const HAZE_COLOR = new THREE.Color(0.15, 0.18, 0.1);

export default function NuclearCloud({ isNewlyPlaced }: NuclearCloudProps) {
  const groupRef = useRef<THREE.Group>(null);
  const mushroomRef = useRef<CloudParticle[]>([]);
  const hazeRef = useRef<CloudParticle[]>([]);
  const flashRef = useRef<THREE.Sprite | null>(null);
  const shockwaveRef = useRef<THREE.Mesh | null>(null);
  const blastStartRef = useRef<number | null>(null);
  const blastCompleteRef = useRef(!isNewlyPlaced);
  const { smoke: smokeTexture } = useTextures();

  const resetHazeParticle = (p: CloudParticle) => {
    const angle = Math.random() * Math.PI * 2;
    const offset = Math.random() * 0.015;
    p.sprite.position.set(Math.cos(angle) * offset, Math.sin(angle) * offset, 0);
    p.age = 0;
    p.maxAge = 8.0 + Math.random() * 6.0;
    p.initialScale = 0.03 + Math.random() * 0.02;
    p.sprite.scale.setScalar(p.initialScale);
    const spread = 0.001 + Math.random() * 0.001;
    const newAngle = Math.random() * Math.PI * 2;
    p.velocity.set(
      Math.cos(newAngle) * spread,
      Math.sin(newAngle) * spread,
      0.006 + Math.random() * 0.004,
    );
    p.rotationSpeed = (Math.random() - 0.5) * 0.3;
    (p.sprite.material as THREE.SpriteMaterial).opacity = 0;
    (p.sprite.material as THREE.SpriteMaterial).rotation = Math.random() * Math.PI * 2;
  };

  useEffect(() => {
    const group = groupRef.current;
    if (!group) return;

    // Mushroom cloud particles (blast only)
    if (isNewlyPlaced) {
      const mushrooms: CloudParticle[] = [];
      for (let i = 0; i < MUSHROOM_PARTICLE_COUNT; i++) {
        const material = new THREE.SpriteMaterial({
          map: smokeTexture,
          transparent: true,
          opacity: 0,
          depthWrite: false,
          blending: THREE.NormalBlending,
          color: MUSHROOM_COLOR,
        });

        const sprite = new THREE.Sprite(material);
        sprite.renderOrder = 100;
        group.add(sprite);

        const isStem = i < 6;
        const angle = Math.random() * Math.PI * 2;
        const spread = isStem ? 0.003 : 0.008 + Math.random() * 0.01;

        const p: CloudParticle = {
          sprite,
          velocity: new THREE.Vector3(
            Math.cos(angle) * spread,
            Math.sin(angle) * spread,
            isStem ? 0.025 + Math.random() * 0.015 : 0.018 + Math.random() * 0.012,
          ),
          rotationSpeed: (Math.random() - 0.5) * 0.5,
          age: 0,
          maxAge: MUSHROOM_LIFETIME + Math.random() * 2.0,
          initialScale: isStem ? 0.02 + Math.random() * 0.015 : 0.04 + Math.random() * 0.03,
          delay: isStem ? i * 0.05 : 0.3 + i * 0.08,
          isMushroom: true,
        };

        sprite.position.set(0, 0, 0);
        mushrooms.push(p);
      }
      mushroomRef.current = mushrooms;

      // Flash sprite
      const flashMat = new THREE.SpriteMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
        color: new THREE.Color(1.0, 1.0, 0.9),
      });
      const flash = new THREE.Sprite(flashMat);
      flash.renderOrder = 101;
      flash.scale.setScalar(0.15);
      group.add(flash);
      flashRef.current = flash;

      // Shockwave ring
      const ringGeo = new THREE.RingGeometry(0.001, 0.01, 32);
      const ringMat = new THREE.MeshBasicMaterial({
        color: new THREE.Color(0.8, 0.6, 0.3),
        transparent: true,
        opacity: 0.6,
        depthWrite: false,
        side: THREE.DoubleSide,
      });
      const ring = new THREE.Mesh(ringGeo, ringMat);
      ring.renderOrder = 99;
      ring.position.z = 0.01;
      group.add(ring);
      shockwaveRef.current = ring;

      blastStartRef.current = null;
      blastCompleteRef.current = false;
    }

    // Lingering haze particles (permanent)
    const hazes: CloudParticle[] = [];
    for (let i = 0; i < HAZE_PARTICLE_COUNT; i++) {
      const material = new THREE.SpriteMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        blending: THREE.NormalBlending,
        color: HAZE_COLOR,
      });

      const sprite = new THREE.Sprite(material);
      sprite.renderOrder = 100;
      group.add(sprite);

      const p: CloudParticle = {
        sprite,
        velocity: new THREE.Vector3(),
        rotationSpeed: 0,
        age: 0,
        maxAge: 3,
        initialScale: 0.03,
        delay: isNewlyPlaced
          ? MUSHROOM_LIFETIME + (i / HAZE_PARTICLE_COUNT) * 4.0
          : (i / HAZE_PARTICLE_COUNT) * 10.0,
        isMushroom: false,
      };

      resetHazeParticle(p);
      p.delay = isNewlyPlaced
        ? MUSHROOM_LIFETIME + (i / HAZE_PARTICLE_COUNT) * 4.0
        : (i / HAZE_PARTICLE_COUNT) * 10.0;
      hazes.push(p);
    }
    hazeRef.current = hazes;

    return () => {
      for (const p of mushroomRef.current) {
        (p.sprite.material as THREE.Material).dispose();
      }
      for (const p of hazeRef.current) {
        (p.sprite.material as THREE.Material).dispose();
      }
      if (flashRef.current) {
        (flashRef.current.material as THREE.Material).dispose();
      }
      if (shockwaveRef.current) {
        shockwaveRef.current.geometry.dispose();
        (shockwaveRef.current.material as THREE.Material).dispose();
      }
    };
  }, [smokeTexture, isNewlyPlaced]);

  useFrame((state, delta) => {
    // Blast effects timing
    if (isNewlyPlaced && !blastCompleteRef.current) {
      if (blastStartRef.current === null) {
        blastStartRef.current = state.clock.elapsedTime;
      }
      const blastElapsed = state.clock.elapsedTime - blastStartRef.current;

      // Flash
      if (flashRef.current) {
        const flashMat = flashRef.current.material as THREE.SpriteMaterial;
        if (blastElapsed < BLAST_FLASH_DURATION) {
          const t = blastElapsed / BLAST_FLASH_DURATION;
          flashMat.opacity = (1 - t) * 0.9;
          flashRef.current.scale.setScalar(0.15 + t * 0.1);
        } else {
          flashMat.opacity = 0;
        }
      }

      // Shockwave
      if (shockwaveRef.current) {
        const ringMat = shockwaveRef.current.material as THREE.MeshBasicMaterial;
        if (blastElapsed < SHOCKWAVE_DURATION) {
          const t = blastElapsed / SHOCKWAVE_DURATION;
          const radius = t * 0.25;
          shockwaveRef.current.geometry.dispose();
          shockwaveRef.current.geometry = new THREE.RingGeometry(
            Math.max(0.001, radius - 0.008),
            radius,
            32,
          );
          ringMat.opacity = (1 - t) * 0.5;
        } else {
          ringMat.opacity = 0;
        }
      }
    }

    // Mushroom cloud particles
    for (const p of mushroomRef.current) {
      if (p.delay > 0) {
        p.delay -= delta;
        continue;
      }

      p.age += delta;

      if (p.age >= p.maxAge) {
        (p.sprite.material as THREE.SpriteMaterial).opacity = 0;
        continue;
      }

      const lifeRatio = p.age / p.maxAge;

      p.velocity.z *= 1 - 0.5 * delta;
      p.sprite.position.x += p.velocity.x * delta;
      p.sprite.position.y += p.velocity.y * delta;
      p.sprite.position.z += p.velocity.z * delta;

      const mat = p.sprite.material as THREE.SpriteMaterial;
      mat.rotation += p.rotationSpeed * delta;

      const fadeIn = Math.min(lifeRatio * 4, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 1.2);
      mat.opacity = fadeIn * fadeOut * 0.7;

      const scale = p.initialScale * (1 + lifeRatio * 4.0);
      p.sprite.scale.setScalar(scale);

      p.velocity.x *= 1 - 0.2 * delta;
      p.velocity.y *= 1 - 0.2 * delta;
    }

    // Haze particles (continuous loop)
    for (const p of hazeRef.current) {
      if (p.delay > 0) {
        p.delay -= delta;
        continue;
      }

      p.age += delta;

      if (p.age >= p.maxAge) {
        resetHazeParticle(p);
        continue;
      }

      const lifeRatio = p.age / p.maxAge;

      p.sprite.position.x += p.velocity.x * delta;
      p.sprite.position.y += p.velocity.y * delta;
      p.sprite.position.z += p.velocity.z * delta;

      const mat = p.sprite.material as THREE.SpriteMaterial;
      mat.rotation += p.rotationSpeed * delta;

      const fadeIn = Math.min(lifeRatio * 5, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 1.5);
      mat.opacity = fadeIn * fadeOut * 0.2;

      const scale = p.initialScale * (1 + lifeRatio * 3.0);
      p.sprite.scale.setScalar(scale);
    }
  });

  return <group ref={groupRef} position={[0, 0, 0.01]} />;
}
