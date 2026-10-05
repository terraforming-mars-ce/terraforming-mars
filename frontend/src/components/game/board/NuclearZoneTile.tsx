import type { NuclearSite } from "./nuclearDebris";
import { useState, useMemo, useEffect } from "react";
import * as THREE from "three";
import NuclearCloud from "./effects/NuclearCloud";
import { createNuclearZoneMaterial, createNuclearMaskMaterial } from "./shaders";
import { createNuclearGeometry } from "./nuclearGeometry";
import { useTextures } from "../../../hooks/useTextures";
import { useClimate } from "../../../contexts/ClimateContext";

interface NuclearZoneTileProps {
  startTime?: number;
  isNewlyPlaced?: boolean;
  site: NuclearSite;
  sphereCenter: THREE.Vector3;
}

export default function NuclearZoneTile({
  isNewlyPlaced = false,
  site,
  startTime,
  sphereCenter,
}: NuclearZoneTileProps) {
  const { seed } = site;
  const [active, setActive] = useState(isNewlyPlaced);
  const [impacted, setImpacted] = useState(!isNewlyPlaced);
  const { mars } = useTextures();
  const { runtime } = useClimate();
  const geometry = useMemo(() => createNuclearGeometry(seed, site.ground, site), [seed, site]);
  const bowl = useMemo(
    () => createNuclearZoneMaterial(seed, mars, sphereCenter),
    [seed, mars, sphereCenter],
  );
  const scorch = useMemo(
    () => createNuclearZoneMaterial(seed, mars, sphereCenter, true),
    [seed, mars, sphereCenter],
  );
  const mask = useMemo(() => createNuclearMaskMaterial(sphereCenter), [sphereCenter]);
  useEffect(
    () => () => {
      Object.values(geometry).forEach((item) => item.dispose());
    },
    [geometry],
  );
  useEffect(
    () => () => {
      bowl.material.dispose();
      scorch.material.dispose();
      mask.dispose();
    },
    [bowl, scorch, mask],
  );
  const updateClimate = () => {
    const chill = 1 - runtime.current.current.sky;
    bowl.uniforms.uChill.value = chill;
    scorch.uniforms.uChill.value = chill;
  };
  return (
    <>
      <group visible={impacted}>
        <mesh
          geometry={geometry.mask}
          material={mask}
          renderOrder={-1}
          frustumCulled={false}
          raycast={() => {}}
        />
        <mesh
          geometry={geometry.bowl}
          material={bowl.material}
          renderOrder={11}
          frustumCulled={false}
          onBeforeRender={updateClimate}
          raycast={() => {}}
        />
        <mesh
          geometry={geometry.scorch}
          material={scorch.material}
          onBeforeRender={updateClimate}
          renderOrder={12}
          frustumCulled={false}
          raycast={() => {}}
        />
      </group>
      {active && (
        <NuclearCloud
          seed={seed}
          startTime={startTime}
          sphereCenter={sphereCenter}
          onImpact={() => setImpacted(true)}
          onComplete={() => setActive(false)}
        />
      )}
    </>
  );
}
