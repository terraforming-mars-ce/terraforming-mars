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
  heightAtSpawn: number;
}

const MUSHROOM_PARTICLE_COUNT = 40;
const FIRE_PARTICLE_COUNT = 30;
const EMBER_PARTICLE_COUNT = 40;
const HAZE_PARTICLE_COUNT = 14;
const MUSHROOM_LIFETIME = 10.0;
const BLAST_FLASH_DURATION = 0.7;
const SHOCKWAVE_DURATION = 2.5;
const MUSHROOM_COLOR_BASE = new THREE.Color(1.0, 0.4, 0.05);
const MUSHROOM_COLOR_TOP = new THREE.Color(0.18, 0.12, 0.08);
const FIRE_COLOR_HOT = new THREE.Color(1.0, 0.8, 0.2);
const FIRE_COLOR_COOL = new THREE.Color(0.9, 0.25, 0.02);
const EMBER_COLOR = new THREE.Color(1.0, 0.55, 0.1);
const HAZE_COLOR = new THREE.Color(0.22, 0.16, 0.1);

export default function NuclearCloud({ isNewlyPlaced }: NuclearCloudProps) {
  const groupRef = useRef<THREE.Group>(null);
  const mushroomRef = useRef<CloudParticle[]>([]);
  const fireRef = useRef<CloudParticle[]>([]);
  const emberRef = useRef<CloudParticle[]>([]);
  const hazeRef = useRef<CloudParticle[]>([]);
  const flashRef = useRef<THREE.Sprite | null>(null);
  const secondFlashRef = useRef<THREE.Sprite | null>(null);
  const shockwaveRef = useRef<THREE.Mesh | null>(null);
  const blastStartRef = useRef<number | null>(null);
  const blastCompleteRef = useRef(!isNewlyPlaced);
  const { smoke: smokeTexture } = useTextures();

  const resetHazeParticle = (p: CloudParticle) => {
    const angle = Math.random() * Math.PI * 2;
    const offset = Math.random() * 0.02;
    p.sprite.position.set(Math.cos(angle) * offset, Math.sin(angle) * offset, 0);
    p.age = 0;
    p.maxAge = 8.0 + Math.random() * 6.0;
    p.initialScale = 0.04 + Math.random() * 0.03;
    p.sprite.scale.setScalar(p.initialScale);
    const spread = 0.001 + Math.random() * 0.002;
    const newAngle = Math.random() * Math.PI * 2;
    p.velocity.set(
      Math.cos(newAngle) * spread,
      Math.sin(newAngle) * spread,
      0.008 + Math.random() * 0.006,
    );
    p.rotationSpeed = (Math.random() - 0.5) * 0.3;
    (p.sprite.material as THREE.SpriteMaterial).opacity = 0;
    (p.sprite.material as THREE.SpriteMaterial).rotation = Math.random() * Math.PI * 2;
  };

  useEffect(() => {
    const group = groupRef.current;
    if (!group) return;

    if (isNewlyPlaced) {
      // Mushroom cloud particles — massive fiery column
      const mushrooms: CloudParticle[] = [];
      for (let i = 0; i < MUSHROOM_PARTICLE_COUNT; i++) {
        const isStem = i < 15;
        const colorT = isStem ? 0.1 + Math.random() * 0.25 : 0.5 + Math.random() * 0.5;
        const particleColor = MUSHROOM_COLOR_BASE.clone().lerp(MUSHROOM_COLOR_TOP, colorT);

        const material = new THREE.SpriteMaterial({
          map: smokeTexture,
          transparent: true,
          opacity: 0,
          depthWrite: false,
          blending: isStem ? THREE.AdditiveBlending : THREE.NormalBlending,
          color: particleColor,
        });

        const sprite = new THREE.Sprite(material);
        sprite.renderOrder = 100;
        group.add(sprite);

        const angle = Math.random() * Math.PI * 2;
        const spread = isStem ? 0.006 : 0.015 + Math.random() * 0.018;

        const p: CloudParticle = {
          sprite,
          velocity: new THREE.Vector3(
            Math.cos(angle) * spread,
            Math.sin(angle) * spread,
            isStem ? 0.05 + Math.random() * 0.035 : 0.035 + Math.random() * 0.025,
          ),
          rotationSpeed: (Math.random() - 0.5) * 0.7,
          age: 0,
          maxAge: MUSHROOM_LIFETIME + Math.random() * 3.0,
          initialScale: isStem ? 0.04 + Math.random() * 0.03 : 0.07 + Math.random() * 0.06,
          delay: isStem ? i * 0.03 : 0.15 + i * 0.04,
          isMushroom: true,
          heightAtSpawn: 0,
        };

        sprite.position.set(0, 0, 0);
        mushrooms.push(p);
      }
      mushroomRef.current = mushrooms;

      // Fire core particles — intense glowing inferno
      const fires: CloudParticle[] = [];
      for (let i = 0; i < FIRE_PARTICLE_COUNT; i++) {
        const colorT = Math.random();
        const fireColor = FIRE_COLOR_HOT.clone().lerp(FIRE_COLOR_COOL, colorT);

        const material = new THREE.SpriteMaterial({
          map: smokeTexture,
          transparent: true,
          opacity: 0,
          depthWrite: false,
          blending: THREE.AdditiveBlending,
          color: fireColor,
        });

        const sprite = new THREE.Sprite(material);
        sprite.renderOrder = 102;
        group.add(sprite);

        const angle = Math.random() * Math.PI * 2;
        const spread = 0.008 + Math.random() * 0.012;

        const p: CloudParticle = {
          sprite,
          velocity: new THREE.Vector3(
            Math.cos(angle) * spread,
            Math.sin(angle) * spread,
            0.05 + Math.random() * 0.04,
          ),
          rotationSpeed: (Math.random() - 0.5) * 1.0,
          age: 0,
          maxAge: 4.0 + Math.random() * 3.0,
          initialScale: 0.045 + Math.random() * 0.04,
          delay: i * 0.025,
          isMushroom: false,
          heightAtSpawn: 0,
        };

        sprite.position.set(0, 0, 0);
        fires.push(p);
      }
      fireRef.current = fires;

      // Ember/spark particles — debris flying everywhere
      const embers: CloudParticle[] = [];
      for (let i = 0; i < EMBER_PARTICLE_COUNT; i++) {
        const material = new THREE.SpriteMaterial({
          map: smokeTexture,
          transparent: true,
          opacity: 0,
          depthWrite: false,
          blending: THREE.AdditiveBlending,
          color: EMBER_COLOR,
        });

        const sprite = new THREE.Sprite(material);
        sprite.renderOrder = 103;
        group.add(sprite);

        const angle = Math.random() * Math.PI * 2;
        const speed = 0.04 + Math.random() * 0.08;

        const p: CloudParticle = {
          sprite,
          velocity: new THREE.Vector3(
            Math.cos(angle) * speed,
            Math.sin(angle) * speed,
            0.015 + Math.random() * 0.05,
          ),
          rotationSpeed: (Math.random() - 0.5) * 3.0,
          age: 0,
          maxAge: 2.0 + Math.random() * 2.5,
          initialScale: 0.012 + Math.random() * 0.018,
          delay: Math.random() * 0.3,
          isMushroom: false,
          heightAtSpawn: 0,
        };

        sprite.position.set(0, 0, 0);
        embers.push(p);
      }
      emberRef.current = embers;

      // Primary flash — massive blinding burst
      const flashMat = new THREE.SpriteMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
        color: new THREE.Color(1.0, 0.9, 0.6),
      });
      const flash = new THREE.Sprite(flashMat);
      flash.renderOrder = 104;
      flash.scale.setScalar(0.5);
      group.add(flash);
      flashRef.current = flash;

      // Secondary flash — rising fireball
      const secondFlashMat = new THREE.SpriteMaterial({
        map: smokeTexture,
        transparent: true,
        opacity: 0,
        depthWrite: false,
        blending: THREE.AdditiveBlending,
        color: new THREE.Color(1.0, 0.5, 0.08),
      });
      const secondFlash = new THREE.Sprite(secondFlashMat);
      secondFlash.renderOrder = 104;
      secondFlash.scale.setScalar(0.35);
      group.add(secondFlash);
      secondFlashRef.current = secondFlash;

      // Shockwave ring — massive fiery blast wave
      const ringGeo = new THREE.RingGeometry(0.001, 0.015, 64);
      const ringMat = new THREE.MeshBasicMaterial({
        color: new THREE.Color(1.0, 0.65, 0.15),
        transparent: true,
        opacity: 0.9,
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
        initialScale: 0.04,
        delay: isNewlyPlaced
          ? MUSHROOM_LIFETIME + (i / HAZE_PARTICLE_COUNT) * 4.0
          : (i / HAZE_PARTICLE_COUNT) * 10.0,
        isMushroom: false,
        heightAtSpawn: 0,
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
      for (const p of fireRef.current) {
        (p.sprite.material as THREE.Material).dispose();
      }
      for (const p of emberRef.current) {
        (p.sprite.material as THREE.Material).dispose();
      }
      for (const p of hazeRef.current) {
        (p.sprite.material as THREE.Material).dispose();
      }
      if (flashRef.current) {
        (flashRef.current.material as THREE.Material).dispose();
      }
      if (secondFlashRef.current) {
        (secondFlashRef.current.material as THREE.Material).dispose();
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

      // Primary flash — massive blinding detonation
      if (flashRef.current) {
        const flashMat = flashRef.current.material as THREE.SpriteMaterial;
        if (blastElapsed < BLAST_FLASH_DURATION) {
          const t = blastElapsed / BLAST_FLASH_DURATION;
          const flashCurve = t < 0.1 ? t / 0.1 : (1 - t) / 0.9;
          flashMat.opacity = flashCurve * 1.0;
          flashRef.current.scale.setScalar(0.5 + t * 0.4);
        } else {
          flashMat.opacity = 0;
        }
      }

      // Secondary flash — rising fireball pulse
      if (secondFlashRef.current) {
        const flashMat = secondFlashRef.current.material as THREE.SpriteMaterial;
        const secondDelay = 0.1;
        const secondDuration = 0.8;
        const secondT = (blastElapsed - secondDelay) / secondDuration;
        if (secondT > 0 && secondT < 1) {
          const curve = secondT < 0.15 ? secondT / 0.15 : (1 - secondT) / 0.85;
          flashMat.opacity = curve * 0.85;
          secondFlashRef.current.scale.setScalar(0.25 + secondT * 0.35);
          secondFlashRef.current.position.z = secondT * 0.08;
        } else {
          flashMat.opacity = 0;
        }
      }

      // Shockwave — massive expanding blast ring
      if (shockwaveRef.current) {
        const ringMat = shockwaveRef.current.material as THREE.MeshBasicMaterial;
        if (blastElapsed < SHOCKWAVE_DURATION) {
          const t = blastElapsed / SHOCKWAVE_DURATION;
          const radius = t * 0.6;
          const thickness = 0.025 * (1 - t * 0.4);
          shockwaveRef.current.geometry.dispose();
          shockwaveRef.current.geometry = new THREE.RingGeometry(
            Math.max(0.001, radius - thickness),
            radius,
            64,
          );
          ringMat.opacity = (1 - t * t) * 0.8;
        } else {
          ringMat.opacity = 0;
        }
      }
    }

    // Mushroom cloud particles — towering fiery column into dark smoke
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

      p.velocity.z *= 1 - 0.35 * delta;
      p.sprite.position.x += p.velocity.x * delta;
      p.sprite.position.y += p.velocity.y * delta;
      p.sprite.position.z += p.velocity.z * delta;

      const mat = p.sprite.material as THREE.SpriteMaterial;
      mat.rotation += p.rotationSpeed * delta;

      if (lifeRatio > 0.25) {
        mat.blending = THREE.NormalBlending;
      }

      const fadeIn = Math.min(lifeRatio * 6, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 1.1);
      mat.opacity = fadeIn * fadeOut * 0.85;

      const scale = p.initialScale * (1 + lifeRatio * 6.0);
      p.sprite.scale.setScalar(scale);

      p.velocity.x *= 1 - 0.12 * delta;
      p.velocity.y *= 1 - 0.12 * delta;
    }

    // Fire core particles — blazing inferno rising from ground zero
    for (const p of fireRef.current) {
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

      p.velocity.z *= 1 - 0.25 * delta;
      p.sprite.position.x += p.velocity.x * delta;
      p.sprite.position.y += p.velocity.y * delta;
      p.sprite.position.z += p.velocity.z * delta;

      const mat = p.sprite.material as THREE.SpriteMaterial;
      mat.rotation += p.rotationSpeed * delta;

      const fadeIn = Math.min(lifeRatio * 10, 1);
      const fadeOut = 1 - Math.pow(lifeRatio, 0.7);
      mat.opacity = fadeIn * fadeOut * 0.95;

      const scale = p.initialScale * (1 + lifeRatio * 4.0);
      p.sprite.scale.setScalar(scale);

      p.velocity.x *= 1 - 0.25 * delta;
      p.velocity.y *= 1 - 0.25 * delta;
    }

    // Ember particles — burning debris flung outward
    for (const p of emberRef.current) {
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

      p.sprite.position.x += p.velocity.x * delta;
      p.sprite.position.y += p.velocity.y * delta;
      p.sprite.position.z += p.velocity.z * delta;

      p.velocity.z -= 0.012 * delta;

      const mat = p.sprite.material as THREE.SpriteMaterial;
      mat.rotation += p.rotationSpeed * delta;

      const fadeIn = Math.min(lifeRatio * 12, 1);
      const fadeOut = 1 - lifeRatio * lifeRatio;
      mat.opacity = fadeIn * fadeOut * 0.9;

      p.sprite.scale.setScalar(p.initialScale * (1 - lifeRatio * 0.4));
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
      mat.opacity = fadeIn * fadeOut * 0.3;

      const scale = p.initialScale * (1 + lifeRatio * 3.5);
      p.sprite.scale.setScalar(scale);
    }
  });

  return <group ref={groupRef} position={[0, 0, 0.01]} />;
}
