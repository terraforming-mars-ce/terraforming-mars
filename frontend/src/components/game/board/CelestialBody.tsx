import { useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import CelestialTileGrid from "./CelestialTileGrid";
import { GameDto } from "../../../types/generated/api-types";
import type { PlanetConfig, MoonConfig } from "./solarSystemConfig";
import { getPlanetOrbitalPosition } from "./solarSystemConfig";

function MoonSphere({
  moon,
  parentPlanetId,
  gameState,
  onHexClick,
}: {
  moon: MoonConfig;
  parentPlanetId: string;
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}) {
  const { activePlanet } = usePlanetFocus();
  const textures = useTextures();
  const tileOpacity = useRef(0);
  const moonGroupRef = useRef<THREE.Group>(null);
  const groupInverseMatrixRef = useRef(new THREE.Matrix4());
  const worldCenterRef = useRef(new THREE.Vector3());

  useFrame(() => {
    const target = activePlanet === parentPlanetId ? 1 : 0;
    tileOpacity.current = THREE.MathUtils.lerp(tileOpacity.current, target, 0.05);
    if (moonGroupRef.current) {
      moonGroupRef.current.updateMatrixWorld(true);
      moonGroupRef.current.getWorldPosition(worldCenterRef.current);
      groupInverseMatrixRef.current.copy(moonGroupRef.current.matrixWorld).invert();
    }
  });

  const texture = moon.textureKey
    ? (textures as unknown as Record<string, THREE.Texture>)[moon.textureKey]
    : undefined;

  const geometry = useMemo(() => new THREE.SphereGeometry(moon.radius, 32, 16), [moon.radius]);

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: texture,
        roughness: 0.9,
        metalness: 0.05,
        fog: false,
      }),
    [texture],
  );

  return (
    <group ref={moonGroupRef} position={[moon.position[0], moon.position[1], moon.position[2]]}>
      <mesh geometry={geometry} material={material} />
      {moon.tileLocation && (
        <CelestialTileGrid
          gameState={gameState}
          onHexClick={onHexClick}
          tileOpacity={tileOpacity}
          location={moon.tileLocation}
          radius={moon.radius}
          coordOffset={moon.coordOffset}
          worldCenter={worldCenterRef.current}
          activePlanetId={parentPlanetId}
          groupInverseMatrix={groupInverseMatrixRef.current}
        />
      )}
    </group>
  );
}

interface CelestialBodyProps {
  config: PlanetConfig;
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}

export default function CelestialBody({ config, gameState, onHexClick }: CelestialBodyProps) {
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const { gl } = useThree();
  const textures = useTextures();
  const tileOpacity = useRef(0);
  const groupRef = useRef<THREE.Group>(null);
  const cloudRef = useRef<THREE.Mesh>(null);
  const worldCenterRef = useRef(new THREE.Vector3());
  const groupInverseMatrixRef = useRef(new THREE.Matrix4());

  const isActive = activePlanet === config.id;

  useFrame((state) => {
    const target = isActive ? 1 : 0;
    tileOpacity.current = THREE.MathUtils.lerp(tileOpacity.current, target, 0.05);

    if (groupRef.current) {
      const pos = getPlanetOrbitalPosition(config, state.clock.elapsedTime);
      groupRef.current.position.set(pos[0], pos[1], pos[2]);
      groupRef.current.lookAt(0, 0, 0);
      groupRef.current.updateMatrixWorld(true);
      worldCenterRef.current.set(pos[0], pos[1], pos[2]);
      groupInverseMatrixRef.current.copy(groupRef.current.matrixWorld).invert();
    }
    if (cloudRef.current) {
      cloudRef.current.rotation.y = state.clock.elapsedTime * 0.02;
    }
  });

  const texture = (textures as unknown as Record<string, THREE.Texture>)[config.textureKey];
  const cloudTexture = config.cloudTextureKey
    ? (textures as unknown as Record<string, THREE.Texture>)[config.cloudTextureKey]
    : undefined;

  const geometry = useMemo(() => new THREE.SphereGeometry(config.radius, 64, 32), [config.radius]);

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: texture,
        roughness: 0.8,
        metalness: 0.05,
        fog: false,
      }),
    [texture],
  );

  const cloudGeometry = useMemo(
    () => (cloudTexture ? new THREE.SphereGeometry(config.radius * 1.005, 64, 32) : undefined),
    [cloudTexture, config.radius],
  );

  const cloudMaterial = useMemo(
    () =>
      cloudTexture
        ? new THREE.MeshStandardMaterial({
            map: cloudTexture,
            transparent: true,
            opacity: 0.4,
            roughness: 1.0,
            metalness: 0.0,
            fog: false,
            depthWrite: false,
          })
        : undefined,
    [cloudTexture],
  );

  const hasTiles =
    config.coordOffset.q !== 0 || config.coordOffset.r !== 0 || config.coordOffset.s !== 0;

  const hitSize = config.radius * 4;
  const hitGeometry = useMemo(() => new THREE.BoxGeometry(hitSize, hitSize, hitSize), [hitSize]);

  return (
    <group ref={groupRef}>
      <mesh
        geometry={geometry}
        material={material}
        onPointerOver={(e) => e.stopPropagation()}
        onClick={(e) => e.stopPropagation()}
      />

      {!isActive && (
        <mesh
          geometry={hitGeometry}
          visible={false}
          onPointerEnter={(e) => {
            if (e.intersections[0]?.object !== e.object) {
              return;
            }
            gl.domElement.style.cursor = "pointer";
          }}
          onPointerLeave={() => {
            gl.domElement.style.cursor = "default";
          }}
          onClick={(e) => {
            e.stopPropagation();
            if (e.intersections[0]?.object !== e.object) {
              return;
            }
            gl.domElement.style.cursor = "default";
            setActivePlanet(config.id as Parameters<typeof setActivePlanet>[0]);
          }}
        />
      )}

      {cloudGeometry && cloudMaterial && (
        <mesh ref={cloudRef} geometry={cloudGeometry} material={cloudMaterial} />
      )}

      {isActive && hasTiles && (
        <CelestialTileGrid
          gameState={gameState}
          onHexClick={onHexClick}
          tileOpacity={tileOpacity}
          location={config.tileLocation}
          radius={config.radius}
          coordOffset={config.coordOffset}
          worldCenter={worldCenterRef.current}
          activePlanetId={config.id}
          groupInverseMatrix={groupInverseMatrixRef.current}
        />
      )}

      {config.moons.map((moon) => (
        <MoonSphere
          key={moon.id}
          moon={moon}
          parentPlanetId={config.id}
          gameState={gameState}
          onHexClick={onHexClick}
        />
      ))}
    </group>
  );
}
