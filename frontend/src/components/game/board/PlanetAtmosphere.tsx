import { useEffect, useLayoutEffect, useMemo, useRef } from "react";
import * as THREE from "three";
import { useFrame } from "@react-three/fiber";
import { ATMOSPHERE_ENABLED, useAtmospheres } from "../../../contexts/PlanetAtmosphereContext";
import { createPlanetAtmosphereMaterial } from "./shaders";
import type { AtmosphereProfile } from "./solarSystemConfig";

const ignoreRaycast = () => {};
const HALO_SEGMENTS = [64, 32] as const;
const HULL_SEGMENTS = [48, 24] as const;

// A low-poly envelope of an irregular body: each direction of a coarse sphere is pushed out to the
// furthest extent of the mesh along it. The glow is soft, so this convex outline reads the same as
// the full mesh at a fraction of the triangles.
function radialHull(geometry: THREE.BufferGeometry) {
  const hull = new THREE.SphereGeometry(1, HULL_SEGMENTS[0], HULL_SEGMENTS[1]);
  const source = geometry.attributes.position;
  const target = hull.attributes.position;
  const direction = new THREE.Vector3();
  const step = Math.max(1, Math.ceil(source.count / 16000));
  for (let i = 0; i < target.count; i++) {
    direction.fromBufferAttribute(target, i).normalize();
    let reach = 0;
    for (let j = 0; j < source.count; j += step) {
      reach = Math.max(
        reach,
        source.getX(j) * direction.x + source.getY(j) * direction.y + source.getZ(j) * direction.z,
      );
    }
    target.setXYZ(i, direction.x * reach, direction.y * reach, direction.z * reach);
  }
  hull.computeVertexNormals();
  return hull;
}

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
  const shell = useRef<THREE.Mesh>(null);
  const bodies = useAtmospheres();
  // Profiles may be mutated in place (Mars's climate), so the shell follows the live thickness.
  useFrame(() => {
    shell.current?.scale.setScalar(radius * (1 + profile.thickness));
  });
  const hull = useMemo(() => (geometry ? radialHull(geometry) : null), [geometry]);
  useEffect(() => () => hull?.dispose(), [hull]);
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
    <group ref={group} userData={{ perfGroup: "atmosphere halos" }}>
      {profile.halo > 0 && !geometry && (
        <mesh
          ref={shell}
          scale={radius * (1 + profile.thickness)}
          material={material}
          renderOrder={-10}
          raycast={ignoreRaycast}
        >
          <sphereGeometry args={[1, ...HALO_SEGMENTS]} />
        </mesh>
      )}
      {profile.halo > 0 &&
        hull &&
        Array.from({ length: 16 }, (_, layer) => (
          <mesh
            key={layer}
            scale={radius * (1 + (profile.thickness * (layer + 1)) / 16)}
            geometry={hull}
            material={material}
            renderOrder={-10}
            raycast={ignoreRaycast}
          />
        ))}
    </group>
  );
}
