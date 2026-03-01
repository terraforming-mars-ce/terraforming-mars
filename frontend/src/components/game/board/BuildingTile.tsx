import { useRef, useState, useEffect, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { SkeletonUtils } from "three-stdlib";
import DustEffect from "./effects/DustEffect";
import { useModels } from "../../../hooks/useModels";

interface BuildingTileProps {
  position: [number, number, number];
  isNewlyPlaced?: boolean;
  surfaceNormal?: THREE.Vector3;
  worldPosition?: THREE.Vector3;
  onEmergenceComplete?: () => void;
}

export default function BuildingTile({
  position,
  isNewlyPlaced = false,
  surfaceNormal,
  worldPosition,
  onEmergenceComplete,
}: BuildingTileProps) {
  const groupRef = useRef<THREE.Group>(null);
  const emergenceStartRef = useRef<number | null>(null);
  const [isEmerging, setIsEmerging] = useState(isNewlyPlaced);
  const [showParticles, setShowParticles] = useState(isNewlyPlaced);

  useEffect(() => {
    if (isNewlyPlaced) {
      setIsEmerging(true);
      setShowParticles(true);
      emergenceStartRef.current = null;
    }
  }, [isNewlyPlaced]);

  useFrame((state) => {
    if (!isEmerging || !groupRef.current) return;

    if (emergenceStartRef.current === null) {
      emergenceStartRef.current = state.clock.elapsedTime * 1000;
    }

    const elapsed = state.clock.elapsedTime * 1000 - emergenceStartRef.current;
    const riseDuration = 800;
    const riseDepth = 0.08;

    const riseProgress = Math.min(elapsed / riseDuration, 1);
    const easedProgress = 1 - Math.pow(1 - riseProgress, 3);

    const zOffset = -riseDepth * (1 - easedProgress);

    const shakeIntensity = 0.01 * (1 - easedProgress);
    const shakeX = (Math.random() - 0.5) * 2 * shakeIntensity;
    const shakeY = (Math.random() - 0.5) * 2 * shakeIntensity;

    groupRef.current.position.set(
      position[0] + shakeX,
      position[1] + shakeY,
      position[2] + zOffset,
    );

    if (riseProgress >= 1) {
      setIsEmerging(false);
      groupRef.current.position.set(position[0], position[1], position[2]);
      onEmergenceComplete?.();
    }
  });

  const handleParticleComplete = () => {
    setShowParticles(false);
  };

  const { cityScene: scene } = useModels();

  const configuredModel = useMemo(() => {
    const clonedScene = SkeletonUtils.clone(scene);

    const targetSize = 0.28;
    const box = new THREE.Box3().setFromObject(clonedScene);
    const size = box.getSize(new THREE.Vector3());
    const maxDimension = Math.max(size.x, size.y, size.z);
    const scaleFactor = targetSize / maxDimension;

    clonedScene.scale.setScalar(scaleFactor);

    const scaledBox = new THREE.Box3().setFromObject(clonedScene);
    const scaledCenter = scaledBox.getCenter(new THREE.Vector3());

    clonedScene.position.set(-scaledCenter.x, -scaledCenter.y, -scaledCenter.z);

    clonedScene.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        child.material = Array.isArray(child.material)
          ? child.material.map((m) => m.clone())
          : child.material.clone();
        child.castShadow = true;
        child.receiveShadow = true;
      }
    });

    return clonedScene;
  }, [scene]);

  const rotation: [number, number, number] = useMemo(() => {
    return [Math.PI / 2, 0, 0];
  }, []);

  return (
    <>
      <group ref={groupRef} position={position} rotation={rotation}>
        <primitive object={configuredModel} />
      </group>
      {showParticles && surfaceNormal && worldPosition && (
        <DustEffect
          position={worldPosition}
          normal={surfaceNormal}
          duration={2600}
          onComplete={handleParticleComplete}
        />
      )}
    </>
  );
}
