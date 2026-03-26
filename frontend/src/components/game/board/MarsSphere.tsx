import { useMemo, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import TileGrid from "./TileGrid.tsx";

import { GameDto } from "../../../types/generated/api-types.ts";
import { useMarsRotation } from "../../../contexts/MarsRotationContext.tsx";
import { useTextures } from "../../../hooks/useTextures.ts";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext.tsx";
import { SPHERE_RADIUS } from "./boardConstants.ts";
import { getMarsOrbitalPosition } from "./solarSystemConfig.ts";

interface MarsSphereProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
  animateHexEntrance?: boolean;
  startHidden?: boolean;
}

export default function MarsSphere({
  gameState,
  onHexClick,
  animateHexEntrance = false,
  startHidden = false,
}: MarsSphereProps) {
  const { marsGroupRef } = useMarsRotation();
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const worldCenterRef = useRef(new THREE.Vector3());
  const groupInverseMatrixRef = useRef(new THREE.Matrix4());

  const { mars: diffuseMap } = useTextures();

  useFrame((state) => {
    if (marsGroupRef.current) {
      const pos = getMarsOrbitalPosition(state.clock.elapsedTime);
      marsGroupRef.current.position.set(pos[0], pos[1], pos[2]);
      marsGroupRef.current.lookAt(0, 0, 0);
      marsGroupRef.current.updateMatrixWorld(true);
      worldCenterRef.current.set(pos[0], pos[1], pos[2]);
      groupInverseMatrixRef.current.copy(marsGroupRef.current.matrixWorld).invert();
    }
  });

  const sphereGeometry = useMemo(() => {
    return new THREE.SphereGeometry(SPHERE_RADIUS, 128, 64);
  }, []);

  const marsMaterial = useMemo(() => {
    const mat = new THREE.MeshStandardMaterial({
      map: diffuseMap,
      roughness: 0.8,
      metalness: 0.05,
      fog: false,
    });

    mat.stencilWrite = true;
    mat.stencilFunc = THREE.NotEqualStencilFunc;
    mat.stencilRef = 1;
    mat.stencilFuncMask = 0xff;
    mat.stencilFail = THREE.KeepStencilOp;
    mat.stencilZFail = THREE.KeepStencilOp;
    mat.stencilZPass = THREE.KeepStencilOp;

    return mat;
  }, [diffuseMap]);

  return (
    <group ref={marsGroupRef}>
      <mesh
        geometry={sphereGeometry}
        material={marsMaterial}
        onClick={(e) => {
          if (activePlanet !== "mars") {
            e.stopPropagation();
            setActivePlanet("mars");
          }
        }}
      />

      <TileGrid
        gameState={gameState}
        onHexClick={onHexClick}
        animateHexEntrance={animateHexEntrance}
        startHidden={startHidden}
        sphereCenter={worldCenterRef.current}
        groupInverseMatrix={groupInverseMatrixRef.current}
      />
    </group>
  );
}
