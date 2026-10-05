import { memo, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useModels } from "../../../hooks/useModels";
import { useTextures } from "../../../hooks/useTextures";
import { createNuclearDebrisMaterial, sharedRockGeometry } from "./rockGeometry";
import { NuclearDebrisBuffer, type NuclearSite } from "./nuclearDebris";
import type { NuclearTransitions } from "./nuclearTransitions";

const noop = () => {};
function NuclearDebrisRenderer({
  sites,
  transitions,
}: {
  sites: ReadonlyMap<string, NuclearSite>;
  transitions: NuclearTransitions;
}) {
  const { rockScene } = useModels();
  const { rock } = useTextures();
  const clock = useThree((state) => state.clock);
  const geometry = useMemo(() => sharedRockGeometry(rockScene), [rockScene]);
  const material = useMemo(() => createNuclearDebrisMaterial(rock), [rock]);
  const mesh = useRef<THREE.InstancedMesh>(null);
  const buffer = useMemo(() => new NuclearDebrisBuffer(sites), [sites]);
  const capacity = Math.max(1, THREE.MathUtils.ceilPowerOfTwo(buffer.count));
  useEffect(() => () => material.dispose(), [material]);
  useLayoutEffect(() => {
    const target = mesh.current!;
    return () => target.dispose();
  }, [capacity, geometry, material]);
  useLayoutEffect(() => {
    buffer.write(mesh.current!, transitions, clock.elapsedTime, true);
  }, [buffer, transitions, capacity]);
  useFrame(({ clock: frameClock }) => {
    buffer.write(mesh.current!, transitions, frameClock.elapsedTime);
  });
  return (
    <instancedMesh
      ref={mesh}
      args={[geometry, material, capacity]}
      dispose={null}
      frustumCulled={false}
      renderOrder={13}
      raycast={noop}
    />
  );
}
export default memo(NuclearDebrisRenderer);
