import {
  createContext,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  type ReactNode,
  type RefObject,
} from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { collapseParameters, collapseProgress, type NuclearTransition } from "./nuclearTransitions";
import { nuclearElapsed, NUCLEAR_FLIGHT_SECONDS } from "./nuclearGeometry";
import { createNuclearCollapseMaterial } from "./shaders";

interface CollapseContext {
  transition?: NuclearTransition;
  matrix: THREE.Matrix4;
}
const Context = createContext<CollapseContext | null>(null);
export const useNuclearCollapse = () => useContext(Context);
export const useNuclearImpacted = () => useNuclearCollapse()?.transition?.impacted ?? false;
const ZERO = new THREE.Vector3();
const UP = new THREE.Vector3(0, 0, 1);

/** Apply collapse after surface projection, which otherwise pins meshes back onto Mars. */
export function NuclearCollapse({
  children,
  transition,
  seed,
  origin = ZERO,
  normal = UP,
}: {
  children: ReactNode;
  transition?: NuclearTransition;
  seed: number;
  origin?: THREE.Vector3;
  normal?: THREE.Vector3;
}) {
  const root = useRef<THREE.Group>(null);
  const matrix = useMemo(() => new THREE.Matrix4(), []);
  const value = useMemo(() => ({ transition, matrix }), [transition, matrix]);
  return (
    <Context.Provider value={value}>
      <group ref={root}>{children}</group>
      {transition && !transition.collapsed && (
        <CollapseAnimator
          key={transition.id}
          root={root}
          value={value}
          seed={seed}
          origin={origin}
          normal={normal}
        />
      )}
    </Context.Provider>
  );
}
function CollapseAnimator({
  root,
  value,
  seed,
  origin,
  normal,
}: {
  root: RefObject<THREE.Group | null>;
  value: CollapseContext;
  seed: number;
  origin: THREE.Vector3;
  normal: THREE.Vector3;
}) {
  const transition = value.transition!;
  const clock = useThree((s) => s.clock);
  const scratch = useMemo(
    () => ({
      frame: new THREE.Matrix4(),
      inverse: new THREE.Matrix4(),
      local: new THREE.Matrix4(),
      rotation: new THREE.Quaternion(),
      axis: new THREE.Vector3(),
      position: new THREE.Vector3(),
      scale: new THREE.Vector3(1, 1, 1),
    }),
    [],
  );
  const entries = useRef<
    Array<{ mesh: THREE.Mesh; matrix: THREE.Matrix4; depth: number; seed: number }>
  >([]);
  const updateMatrix = (matrix: THREE.Matrix4, partSeed: number, depth: number) => {
    const age =
      nuclearElapsed(clock.elapsedTime - (transition?.start ?? Infinity)) - NUCLEAR_FLIGHT_SECONDS;
    const part = collapseParameters(partSeed);
    const progress = collapseProgress(age, part.delay);
    scratch.axis.set(Math.cos(part.direction), Math.sin(part.direction), 0);
    scratch.rotation.setFromAxisAngle(scratch.axis, part.tilt * progress);
    scratch.position.set(0, 0, -depth * progress);
    scratch.local.compose(scratch.position, scratch.rotation, scratch.scale);
    matrix.copy(scratch.frame).multiply(scratch.local).multiply(scratch.inverse);
  };
  useFrame(() => {
    if (!transition || !root.current) {
      return;
    }
    root.current.updateWorldMatrix(true, true);
    scratch.rotation.setFromUnitVectors(UP, normal);
    scratch.frame
      .compose(origin, scratch.rotation, scratch.scale)
      .premultiply(root.current.matrixWorld);
    scratch.inverse.copy(scratch.frame).invert();
    updateMatrix(value.matrix, seed, 0.25);
    for (const entry of entries.current) {
      updateMatrix(entry.matrix, entry.seed, entry.depth);
      if (
        !Array.isArray(entry.mesh.material) &&
        !entry.mesh.material.colorWrite &&
        transition.impacted
      ) {
        entry.mesh.visible = false;
      }
    }
  }, -50);
  useLayoutEffect(() => {
    if (!transition || !root.current) {
      return;
    }
    const restore: Array<() => void> = [];
    root.current.updateWorldMatrix(true, true);
    const localInverse = root.current.matrixWorld.clone().invert();
    const bounds = new THREE.Box3();
    const point = new THREE.Vector3();
    root.current.traverse((object) => {
      if (!(object instanceof THREE.Mesh)) {
        return;
      }
      const original = object.material;
      const sources = Array.isArray(original) ? original : [original];
      const transform = { value: new THREE.Matrix4() };
      const viewRotation = { value: new THREE.Matrix3() };
      const worldRotation = { value: new THREE.Matrix3() };
      const worldTransform = new THREE.Matrix4();
      const viewTransform = new THREE.Matrix4();
      let height = 0;
      bounds.setFromObject(object);
      for (let i = 0; i < 8; i++) {
        point.set(
          i & 1 ? bounds.max.x : bounds.min.x,
          i & 2 ? bounds.max.y : bounds.min.y,
          i & 4 ? bounds.max.z : bounds.min.z,
        );
        point.applyMatrix4(localInverse).sub(origin);
        height = Math.max(height, point.dot(normal) + point.length() * 0.11);
      }
      const materials = sources.map((source) => {
        if (source instanceof THREE.ShaderMaterial) {
          height = Math.max(
            height,
            (source.uniforms.uHeight?.value ?? source.uniforms.uTrunkHeight?.value ?? 0) * 1.5,
          );
        }
        return createNuclearCollapseMaterial(source, transform, viewRotation, worldRotation);
      });
      const before = object.onBeforeRender;
      const culled = object.frustumCulled;
      const visible = object.visible;
      object.frustumCulled = false;
      object.material = Array.isArray(original) ? materials : materials[0];
      object.onBeforeRender = function (renderer, scene, camera, geometry, material, group) {
        before.call(this, renderer, scene, camera, geometry, material, group);
        transform.value
          .copy(camera.projectionMatrix)
          .multiply(camera.matrixWorldInverse)
          .multiply(worldTransform)
          .multiply(camera.matrixWorld)
          .multiply(camera.projectionMatrixInverse);
        worldRotation.value.setFromMatrix4(worldTransform);
        viewRotation.value.setFromMatrix4(
          viewTransform
            .copy(camera.matrixWorldInverse)
            .multiply(worldTransform)
            .multiply(camera.matrixWorld),
        );
      };
      entries.current.push({
        mesh: object,
        matrix: worldTransform,
        depth: height + 0.065,
        seed: seed + entries.current.length * 97,
      });
      restore.push(() => {
        object.material = original;
        object.onBeforeRender = before;
        object.frustumCulled = culled;
        object.visible = visible;
        materials.forEach((material) => material.dispose());
      });
    });
    return () => {
      restore.forEach((fn) => fn());
      entries.current = [];
    };
  }, [transition?.id, seed, origin, normal]);
  return null;
}
