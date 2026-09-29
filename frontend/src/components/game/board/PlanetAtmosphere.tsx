import { useEffect, useLayoutEffect, useMemo, useRef } from "react";
import * as THREE from "three";
import { ATMOSPHERE_ENABLED, useAtmospheres } from "../../../contexts/PlanetAtmosphereContext";
import { createPlanetAtmosphereMaterial } from "./shaders";
import type { AtmosphereProfile } from "./solarSystemConfig";

const ignoreRaycast = () => {};

interface PlanetAtmosphereProps {
  radius: number;
  profile: AtmosphereProfile;
  geometry?: THREE.BufferGeometry;
}

export default function PlanetAtmosphere(props: PlanetAtmosphereProps) {
  if (!ATMOSPHERE_ENABLED) {
    return null;
  }
  return <EnabledPlanetAtmosphere {...props} />;
}

function EnabledPlanetAtmosphere({ radius, profile, geometry }: PlanetAtmosphereProps) {
  const group = useRef<THREE.Group>(null);
  const bodies = useAtmospheres();
  const material = useMemo(() => {
    const halo = createPlanetAtmosphereMaterial(profile);
    halo.uniforms.uFollowsMesh.value = Number(!!geometry);
    return halo;
  }, [profile, geometry]);
  useEffect(() => () => material.dispose(), [material]);
  useLayoutEffect(() => {
    if (!group.current) {
      return;
    }
    const body = {
      object: group.current,
      radius,
      profile,
      halo: material,
      followsMesh: !!geometry,
    };
    bodies.add(body);
    return () => {
      bodies.delete(body);
    };
  }, [bodies, radius, profile, material, geometry]);
  return (
    <group ref={group}>
      {profile.halo > 0 && !geometry && (
        <mesh
          scale={radius * (1 + profile.thickness)}
          material={material}
          renderOrder={-10}
          raycast={ignoreRaycast}
        >
          <sphereGeometry args={[1, 128, 64]} />
        </mesh>
      )}
      {profile.halo > 0 &&
        geometry &&
        Array.from({ length: 16 }, (_, layer) => (
          <mesh
            key={layer}
            scale={radius * (1 + (profile.thickness * (layer + 1)) / 16)}
            geometry={geometry}
            material={material}
            renderOrder={-10}
            raycast={ignoreRaycast}
          />
        ))}
    </group>
  );
}
