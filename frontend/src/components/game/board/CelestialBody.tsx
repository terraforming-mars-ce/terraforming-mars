import { sphereRaycast } from "../../../utils/sphereRaycast";
import { useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { usePlanetSurface, type PlanetSurfaceKey } from "../../../hooks/useTextures";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import CelestialTileGrid from "./CelestialTileGrid";
import PlanetAtmosphere from "./PlanetAtmosphere";
import { GameDto } from "../../../types/generated/api-types";
import type { PlanetConfig, MoonConfig } from "./solarSystemConfig";
import { getPlanetOrbitalPosition } from "./solarSystemConfig";
import { GRAPHICS } from "@/utils/graphicsQuality.ts";
import { isDragClick } from "../controls/PanControls";

function MoonSphere({
  moon,
  textureKey,
  parentPlanetId,
  gameState,
  onHexClick,
}: {
  moon: MoonConfig;
  textureKey: PlanetSurfaceKey;
  parentPlanetId: string;
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}) {
  const { activePlanet } = usePlanetFocus();
  const texture = usePlanetSurface(textureKey, activePlanet === parentPlanetId);
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

  const geometry = useMemo(
    () => new THREE.SphereGeometry(moon.radius, ...GRAPHICS.sphereSegments),
    [moon.radius],
  );

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: texture,
        roughness: 0.9,
        metalness: 0,
        fog: false,
      }),
    [texture],
  );

  return (
    <group ref={moonGroupRef} position={[moon.position[0], moon.position[1], moon.position[2]]}>
      <mesh geometry={geometry} raycast={sphereRaycast} material={material} />
      <PlanetAtmosphere radius={moon.radius} profile={moon.atmosphere} />
      {moon.tileLocation && (
        <CelestialTileGrid
          highlightRoot={moonGroupRef}
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

function PlanetClouds({
  textureKey,
  radius,
  closeUp,
}: {
  textureKey: PlanetSurfaceKey;
  radius: number;
  closeUp: boolean;
}) {
  const cloudRef = useRef<THREE.Mesh>(null);
  const texture = usePlanetSurface(textureKey, closeUp);

  useFrame((state) => {
    if (cloudRef.current) {
      cloudRef.current.rotation.y = state.clock.elapsedTime * 0.02;
    }
  });

  const geometry = useMemo(
    () => new THREE.SphereGeometry(radius * 1.005, ...GRAPHICS.sphereSegments),
    [radius],
  );

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: texture,
        transparent: true,
        opacity: 0.4,
        roughness: 1.0,
        metalness: 0.0,
        fog: false,
        depthWrite: false,
      }),
    [texture],
  );

  return <mesh ref={cloudRef} geometry={geometry} material={material} />;
}

interface CelestialBodyProps {
  config: PlanetConfig;
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}

export default function CelestialBody({ config, gameState, onHexClick }: CelestialBodyProps) {
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const { gl } = useThree();
  const tileOpacity = useRef(0);
  const groupRef = useRef<THREE.Group>(null);
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
  });

  const texture = usePlanetSurface(config.textureKey, isActive);

  const geometry = useMemo(
    () => new THREE.SphereGeometry(config.radius, ...GRAPHICS.sphereSegments),
    [config.radius],
  );

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({ map: texture, roughness: 0.8, metalness: 0, fog: false }),
    [texture],
  );

  const hasTiles =
    config.coordOffset.q !== 0 || config.coordOffset.r !== 0 || config.coordOffset.s !== 0;

  const hitGeometry = useMemo(
    () => new THREE.SphereGeometry(config.radius * 1.2, 32, 16),
    [config.radius],
  );

  return (
    <group ref={groupRef}>
      <mesh
        geometry={geometry}
        raycast={sphereRaycast}
        material={material}
        onPointerOver={(e) => e.stopPropagation()}
        onClick={(e) => e.stopPropagation()}
      />

      {!isActive && (
        <mesh
          geometry={hitGeometry}
          raycast={sphereRaycast}
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
            if (isDragClick()) {
              return;
            }
            if (e.intersections[0]?.object !== e.object) {
              return;
            }
            gl.domElement.style.cursor = "default";
            setActivePlanet(config.id as Parameters<typeof setActivePlanet>[0]);
          }}
        />
      )}

      {config.cloudTextureKey && (
        <PlanetClouds
          textureKey={config.cloudTextureKey}
          radius={config.radius}
          closeUp={isActive}
        />
      )}

      <PlanetAtmosphere radius={config.radius} profile={config.atmosphere} />

      {isActive && hasTiles && (
        <CelestialTileGrid
          highlightRoot={groupRef}
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

      {config.moons.map(
        (moon) =>
          moon.textureKey && (
            <MoonSphere
              key={moon.id}
              moon={moon}
              textureKey={moon.textureKey}
              parentPlanetId={config.id}
              gameState={gameState}
              onHexClick={onHexClick}
            />
          ),
      )}
    </group>
  );
}
