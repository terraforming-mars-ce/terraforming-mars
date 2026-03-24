import { useMemo, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useModels } from "../../../hooks/useModels";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import CelestialTileGrid from "./CelestialTileGrid";
import { GameDto } from "../../../types/generated/api-types";
import { PHOBOS_CONFIG, getMarsOrbitalPosition } from "./solarSystemConfig";

const PHOBOS_OFFSET = new THREE.Vector3(3.2, 0.3, 0.5);
const PHOBOS_SCALE = 0.027;

interface PhobosBodyProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}

export default function PhobosBody({ gameState, onHexClick }: PhobosBodyProps) {
  const { phobosScene } = useModels();
  const { activePlanet } = usePlanetFocus();
  const groupRef = useRef<THREE.Group>(null);
  const tileOpacity = useRef(0);
  const worldCenterRef = useRef(new THREE.Vector3());

  const clonedScene = useMemo(() => phobosScene.clone(), [phobosScene]);

  const _marsPos = useMemo(() => new THREE.Vector3(), []);
  const _toOrigin = useMemo(() => new THREE.Vector3(), []);
  const _quat = useMemo(() => new THREE.Quaternion(), []);
  const _zAxis = useMemo(() => new THREE.Vector3(0, 0, 1), []);
  const _rotatedOffset = useMemo(() => new THREE.Vector3(), []);

  useFrame((state) => {
    const target = activePlanet === "mars" ? 1 : 0;
    tileOpacity.current = THREE.MathUtils.lerp(tileOpacity.current, target, 0.05);

    if (groupRef.current) {
      const marsPos = getMarsOrbitalPosition(state.clock.elapsedTime);
      _marsPos.set(marsPos[0], marsPos[1], marsPos[2]);
      _toOrigin.copy(_marsPos).negate().normalize();
      _quat.setFromUnitVectors(_zAxis, _toOrigin);
      _rotatedOffset.copy(PHOBOS_OFFSET).applyQuaternion(_quat);
      groupRef.current.position.copy(_marsPos).add(_rotatedOffset);
      groupRef.current.lookAt(0, 0, 0);
      worldCenterRef.current.copy(groupRef.current.position);
    }
  });

  return (
    <group ref={groupRef}>
      <primitive object={clonedScene} scale={[PHOBOS_SCALE, PHOBOS_SCALE, PHOBOS_SCALE]} />
      <CelestialTileGrid
        gameState={gameState}
        onHexClick={onHexClick}
        tileOpacity={tileOpacity}
        location={PHOBOS_CONFIG.tileLocation!}
        radius={PHOBOS_CONFIG.radius}
        coordOffset={PHOBOS_CONFIG.coordOffset}
        worldCenter={worldCenterRef.current}
        activePlanetId="mars"
      />
    </group>
  );
}
