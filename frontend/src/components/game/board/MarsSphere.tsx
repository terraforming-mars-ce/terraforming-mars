import { useMemo } from "react";
import * as THREE from "three";
import TileGrid from "./TileGrid.tsx";

import { GameDto } from "../../../types/generated/api-types.ts";
import { useMarsRotation } from "../../../contexts/MarsRotationContext.tsx";
import { useTextures } from "../../../hooks/useTextures.ts";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext.tsx";
import { SPHERE_RADIUS } from "./boardConstants.ts";

interface MarsSphereProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
  animateHexEntrance?: boolean;
}

export default function MarsSphere({
  gameState,
  onHexClick,
  animateHexEntrance = false,
}: MarsSphereProps) {
  const { marsGroupRef } = useMarsRotation();
  const { activePlanet, setActivePlanet } = usePlanetFocus();

  const { mars: diffuseMap } = useTextures();

  // Get Mars color based on terraforming progress for tinting
  const marsColorTint = useMemo(() => {
    const temp = gameState?.globalParameters?.temperature || -30;
    const oxygen = gameState?.globalParameters?.oxygen || 0;

    const tempProgress = Math.max(0, (temp + 30) / 38);
    const oxygenProgress = oxygen / 14;

    const red = 1 - tempProgress * 0.3;
    const green = tempProgress * 0.2 + oxygenProgress * 0.3;
    const blue = oxygenProgress * 0.2;

    return new THREE.Color(red, green, blue);
  }, [gameState?.globalParameters]);

  const sphereGeometry = useMemo(() => {
    return new THREE.SphereGeometry(SPHERE_RADIUS, 128, 64);
  }, []);

  // Create material with terraforming tint and desaturated texture
  const marsMaterial = useMemo(() => {
    const baseMarsColor = new THREE.Color(1, 1, 1);
    const tintedColor = baseMarsColor.lerp(marsColorTint, 0.3);

    const mat = new THREE.MeshStandardMaterial({
      map: diffuseMap,
      color: tintedColor,
      roughness: 0.85,
      metalness: 0.05,
    });

    mat.onBeforeCompile = (shader) => {
      shader.uniforms.uSaturation = { value: 0.3 };
      shader.fragmentShader = shader.fragmentShader.replace(
        "#include <map_fragment>",
        `
        #ifdef USE_MAP
          vec4 sampledDiffuseColor = texture2D( map, vMapUv );
          float grey = dot(sampledDiffuseColor.rgb, vec3(0.299, 0.587, 0.114));
          sampledDiffuseColor.r = mix(grey, sampledDiffuseColor.r, 0.7);
          sampledDiffuseColor.g = mix(grey, sampledDiffuseColor.g, 0.85);
          sampledDiffuseColor.b = mix(grey, sampledDiffuseColor.b, 0.85);
          diffuseColor *= sampledDiffuseColor;
        #endif
        `,
      );
    };

    return mat;
  }, [diffuseMap, marsColorTint]);

  return (
    <group ref={marsGroupRef}>
      <mesh
        geometry={sphereGeometry}
        material={marsMaterial}
        rotation={[0, (65 * Math.PI) / 180, 0]}
        castShadow
        receiveShadow
        onClick={(e) => {
          if (activePlanet === "venus") {
            e.stopPropagation();
            setActivePlanet("mars");
          }
        }}
      />

      {/* Projected hexagonal grid "wrapped" around Mars sphere */}
      <TileGrid
        gameState={gameState}
        onHexClick={onHexClick}
        animateHexEntrance={animateHexEntrance}
      />
    </group>
  );
}
