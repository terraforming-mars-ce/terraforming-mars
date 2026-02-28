import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import NuclearCloud from "./effects/NuclearCloud";
import DustEffect from "./effects/DustEffect";
import { useTextures } from "../../../hooks/useTextures";
import { createNuclearZoneMaterial } from "./shaders";
import { addSphereProjectionWithSoftEdges } from "./GreeneryRenderer";
import { easeOutCubic } from "./boardConstants";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";

interface NuclearZoneTileProps {
  isNewlyPlaced?: boolean;
  seed?: number;
  surfaceNormal?: THREE.Vector3;
  worldPosition?: THREE.Vector3;
}

const HEX_RADIUS = 0.166;

function createRadialDiscGeometry(
  radius: number,
  radialSegments: number,
  angularSegments: number,
): THREE.BufferGeometry {
  const geometry = new THREE.BufferGeometry();
  const vertices: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  vertices.push(0, 0, 0);
  uvs.push(0.5, 0.5);

  for (let r = 1; r <= radialSegments; r++) {
    const ringRadius = (r / radialSegments) * radius;
    for (let a = 0; a < angularSegments; a++) {
      const angle = (a / angularSegments) * Math.PI * 2;
      const x = Math.cos(angle) * ringRadius;
      const y = Math.sin(angle) * ringRadius;
      vertices.push(x, y, 0);
      uvs.push(0.5 + (x / radius) * 0.5, 0.5 + (y / radius) * 0.5);
    }
  }

  for (let a = 0; a < angularSegments; a++) {
    const next = (a + 1) % angularSegments;
    indices.push(0, 1 + a, 1 + next);
  }

  for (let r = 1; r < radialSegments; r++) {
    const innerStart = 1 + (r - 1) * angularSegments;
    const outerStart = 1 + r * angularSegments;
    for (let a = 0; a < angularSegments; a++) {
      const nextA = (a + 1) % angularSegments;
      indices.push(innerStart + a, outerStart + a, outerStart + nextA);
      indices.push(innerStart + a, outerStart + nextA, innerStart + nextA);
    }
  }

  geometry.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  return geometry;
}

export default function NuclearZoneTile({
  isNewlyPlaced = false,
  seed: seedProp,
  surfaceNormal,
  worldPosition,
}: NuclearZoneTileProps) {
  const groupRef = useRef<THREE.Group>(null);
  const emergenceStartRef = useRef<number | null>(null);
  const isEmergingRef = useRef(isNewlyPlaced);
  const [showCloud, setShowCloud] = useState(!isNewlyPlaced);
  const [showDust, setShowDust] = useState(isNewlyPlaced);
  const { settings: world3DSettings } = useWorld3DSettings();

  useEffect(() => {
    if (isNewlyPlaced) {
      isEmergingRef.current = true;
      setShowDust(true);
      setShowCloud(false);
      emergenceStartRef.current = null;
    }
  }, [isNewlyPlaced]);

  const seed = useMemo(() => seedProp ?? Math.random() * 10000, [seedProp]);

  const craterGeometry = useMemo(() => {
    return createRadialDiscGeometry(0.14, 28, 56);
  }, []);

  const groundGeometry = useMemo(() => {
    const geo = createRadialDiscGeometry(HEX_RADIUS * 1.8, 10, 36);
    geo.rotateZ(Math.PI / 2);
    return geo;
  }, []);

  const { noiseMid: noiseTexture, noiseHigh: noiseHighTexture } = useTextures();

  const nuclearZoneMaterial = useMemo(() => {
    const mat = createNuclearZoneMaterial(seed);
    mat.uniforms.uEmergence.value = isEmergingRef.current ? 0.0 : 1.0;
    return mat;
  }, [seed]);

  const groundMaterial = useMemo(() => {
    const mat = new THREE.MeshStandardMaterial({
      color: new THREE.Color(0.18, 0.12, 0.08),
      roughness: 0.95,
      metalness: 0.0,
      side: THREE.DoubleSide,
      transparent: true,
      alphaTest: 0.01,
      depthWrite: false,
    });
    addSphereProjectionWithSoftEdges(mat, 0.003, noiseTexture, noiseHighTexture, HEX_RADIUS);
    return mat;
  }, [noiseTexture, noiseHighTexture]);

  useFrame((state) => {
    nuclearZoneMaterial.uniforms.uTime.value = state.clock.elapsedTime;
    nuclearZoneMaterial.uniforms.uSunDirection.value
      .set(
        world3DSettings.sunDirectionX,
        world3DSettings.sunDirectionY,
        world3DSettings.sunDirectionZ,
      )
      .normalize();
    nuclearZoneMaterial.uniforms.uSunIntensity.value = world3DSettings.sunIntensity;
    nuclearZoneMaterial.uniforms.uSunColor.value.set(
      world3DSettings.sunColor.r,
      world3DSettings.sunColor.g,
      world3DSettings.sunColor.b,
    );

    if (!isEmergingRef.current) {
      nuclearZoneMaterial.uniforms.uEmergence.value = 1.0;
      return;
    }

    if (!groupRef.current) return;

    if (emergenceStartRef.current === null) {
      emergenceStartRef.current = state.clock.elapsedTime * 1000;
    }

    const elapsed = state.clock.elapsedTime * 1000 - emergenceStartRef.current;
    const duration = 1200;
    const t = Math.min(elapsed / duration, 1);
    const eased = easeOutCubic(t);

    nuclearZoneMaterial.uniforms.uEmergence.value = eased;

    const shakeIntensity = 0.015 * (1 - eased);
    const shakeX = (Math.random() - 0.5) * 2 * shakeIntensity;
    const shakeY = (Math.random() - 0.5) * 2 * shakeIntensity;
    groupRef.current.position.set(shakeX, shakeY, 0);

    if (t >= 1) {
      isEmergingRef.current = false;
      setShowCloud(true);
      groupRef.current.position.set(0, 0, 0);
    }
  });

  const groundQuaternion = useMemo(() => new THREE.Quaternion(), []);

  return (
    <>
      <group ref={groupRef}>
        <mesh
          geometry={groundGeometry}
          material={groundMaterial}
          quaternion={groundQuaternion}
          renderOrder={12}
        />

        <mesh
          geometry={craterGeometry}
          material={nuclearZoneMaterial}
          renderOrder={13}
          frustumCulled={false}
        />
      </group>

      {showCloud && <NuclearCloud isNewlyPlaced={isNewlyPlaced} />}

      {showDust && surfaceNormal && worldPosition && (
        <DustEffect
          position={worldPosition}
          normal={surfaceNormal}
          duration={3000}
          particleColor={new THREE.Color(0.35, 0.25, 0.15)}
          onComplete={() => setShowDust(false)}
        />
      )}
    </>
  );
}
