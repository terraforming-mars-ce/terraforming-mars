import { useLayoutEffect, useMemo, type ReactNode, type RefObject } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import {
  ATMOSPHERE_ENABLED,
  AtmospheresContext,
  type AtmosphereBody,
} from "../../../contexts/PlanetAtmosphereContext";
import {
  addPlanetHaze,
  createPlanetHazeUniforms,
  receivesPlanetHaze,
  MAX_ATMOSPHERES,
} from "./shaders";

interface AtmosphereRendererProps {
  sunLight: RefObject<THREE.PointLight | null>;
  enabled?: boolean;
  children: ReactNode;
}

export default function AtmosphereRenderer(props: AtmosphereRendererProps) {
  if (!ATMOSPHERE_ENABLED) {
    return <>{props.children}</>;
  }
  return <EnabledAtmosphereRenderer {...props} />;
}

function EnabledAtmosphereRenderer({
  sunLight,
  enabled = true,
  children,
}: AtmosphereRendererProps) {
  const { size } = useThree();
  const bodies = useMemo(() => new Set<AtmosphereBody>(), []);
  const uniforms = useMemo(() => createPlanetHazeUniforms(), []);
  const materials = useMemo(() => new Map<THREE.Material, () => void>(), []);
  const scratch = useMemo(
    () => ({
      center: new THREE.Vector3(),
      viewCenter: new THREE.Vector3(),
      sun: new THREE.Vector3(),
      color: new THREE.Color(),
      profile: new THREE.Vector3(),
    }),
    [],
  );

  useLayoutEffect(
    () => () => {
      materials.forEach((restore) => restore());
      materials.clear();
    },
    [materials],
  );

  useFrame(({ gl, scene, camera }) => {
    scene.updateMatrixWorld();
    camera.updateMatrixWorld();
    const light = sunLight.current;
    let count = 0;
    if (light) {
      light.getWorldPosition(scratch.sun);
    }
    for (const body of bodies) {
      scratch.center.setFromMatrixPosition(body.object.matrixWorld);
      scratch.viewCenter.copy(scratch.center).applyMatrix4(camera.matrixWorldInverse);
      const projectedRadius =
        (body.radius * camera.projectionMatrix.elements[5] * size.height) /
        (2 * Math.max(-scratch.viewCenter.z, body.radius));
      const visibility = enabled ? THREE.MathUtils.smoothstep(projectedRadius, 3, 12) : 0;
      const intensity = body.profile.intensity * (light?.intensity ?? 0) * visibility;
      const halo = body.halo.uniforms;
      halo.uSunDirection.value.copy(scratch.sun).sub(scratch.center).normalize();
      if (light) {
        halo.uSunColor.value.copy(light.color);
      }
      halo.uIntensity.value = intensity * body.profile.halo;
      if (count === MAX_ATMOSPHERES || intensity === 0 || scratch.viewCenter.z > body.radius) {
        continue;
      }
      scratch.center.toArray(uniforms.uHazeBodies.value, count * 4);
      uniforms.uHazeBodies.value[count * 4 + 3] = body.radius;
      scratch.color.set(body.profile.color).toArray(uniforms.uHazeColors.value, count * 3);
      scratch.color
        .set(body.profile.shadowColor)
        .toArray(uniforms.uHazeShadowColors.value, count * 3);
      scratch.profile
        .set(body.profile.thickness, intensity, Number(body.followsMesh))
        .toArray(uniforms.uHazeProfiles.value, count * 3);
      count++;
    }
    uniforms.uHazeCount.value = count;
    uniforms.uHazeSunPosition.value.copy(scratch.sun);
    if (light) {
      uniforms.uHazeSunColor.value.copy(light.color);
    }
    uniforms.uHazeCameraWorld.value.copy(camera.matrixWorld);
    uniforms.uHazeProjectionInverse.value.copy(camera.projectionMatrixInverse);
    scene.traverseVisible((object) => {
      if (!(object instanceof THREE.Mesh)) {
        return;
      }
      const meshMaterials = Array.isArray(object.material) ? object.material : [object.material];
      for (const material of meshMaterials) {
        if (materials.has(material) || !receivesPlanetHaze(material)) {
          continue;
        }
        const restore = addPlanetHaze(material, uniforms);
        const release = () => {
          material.removeEventListener("dispose", release);
          materials.delete(material);
          restore();
        };
        material.addEventListener("dispose", release);
        materials.set(material, release);
      }
    });
    const autoUpdate = scene.matrixWorldAutoUpdate;
    scene.matrixWorldAutoUpdate = false;
    try {
      gl.render(scene, camera);
    } finally {
      scene.matrixWorldAutoUpdate = autoUpdate;
    }
  }, 1);

  return <AtmospheresContext.Provider value={bodies}>{children}</AtmospheresContext.Provider>;
}
