import { memo, useEffect, useLayoutEffect, useMemo } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { LandscapeSurface } from "./landscapeSurface";
import { boardCenter } from "./landscapeGeometry";
import type { LandscapeState } from "./landscapeTypes";
const noop = () => {};
function LandscapeRenderer({
  state,
  capacity = 128,
  onReady = noop,
  hoveredOceanHexKey = null,
  groupInverseMatrix,
}: {
  state: LandscapeState;
  capacity?: number;
  onReady?: (state: LandscapeState, transitionStart: number) => void;
  hoveredOceanHexKey?: string | null;
  groupInverseMatrix?: THREE.Matrix4;
}) {
  const textures = useTextures();
  const gl = useThree((s) => s.gl);
  const { settings } = useWorld3DSettings();
  const surface = useMemo(
    () => new LandscapeSurface(textures, capacity),
    [
      capacity,
      textures.mars,
      textures.grass,
      textures.sand,
      textures.rock,
      textures.concrete,
      textures.waterNormals,
      textures.leafyGrass,
      textures.leafyGrassDetail,
      textures.forestLitter,
      textures.forestLitterDetail,
      textures.wetSoil,
      textures.wetSoilDetail,
    ],
  );
  useLayoutEffect(() => {
    surface.enqueue(state);
  }, [surface, state]);
  useLayoutEffect(() => {
    gl.initTexture(surface.terrain);
    gl.initTexture(surface.materials);
  }, [gl, surface]);
  useEffect(() => {
    surface.retain();
    return () => surface.release();
  }, [surface]);
  useFrame(({ camera }) => {
    surface.tick(gl, performance.now(), onReady);
    const u = surface.water.uniforms;
    u.time.value = performance.now() / 1000;
    u.eye.value.copy(camera.position);
    if (groupInverseMatrix) {
      u.eye.value.applyMatrix4(groupInverseMatrix);
    }
    u.sunDirection.value
      .set(settings.sunDirectionX, settings.sunDirectionY, settings.sunDirectionZ)
      .normalize();
    u.sunColor.value.set(settings.sunColor.r, settings.sunColor.g, settings.sunColor.b);
    u.sunIntensity.value = settings.sunIntensity;
    u.waterColor.value.set(settings.waterColor.r, settings.waterColor.g, settings.waterColor.b);
    u.rf0.value = settings.reflectance;
    u.uHoverActive.value = hoveredOceanHexKey ? 1 : 0;
    if (hoveredOceanHexKey) {
      const [q, r, s] = hoveredOceanHexKey.split(",").map(Number);
      const p = boardCenter({ q, r, s });
      u.uHoverCenter.value.set(p.x, p.y);
    }
  });
  return <primitive object={surface.group} dispose={null} />;
}
export default memo(LandscapeRenderer);
