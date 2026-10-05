import { useLayoutEffect, useMemo, useRef, type ReactNode, type RefObject } from "react";
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

import { NuclearEffectsContext, type NuclearEffect } from "../../../contexts/NuclearEffectsContext";
import { NuclearCompositor } from "./nuclearCompositor";
import { useReducedMotion } from "../../../hooks/useReducedMotion";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { performanceStore } from "../../../services/performanceStore";
import { sceneCensus } from "./perfCensus";
import { GpuTimer } from "./gpuTimer";
import { TileHighlightSystem } from "./tileHighlightMaterials";
import { TileHighlightSystemContext } from "./TileHighlightContext";
import { useCardDragStore } from "../../../stores/cardDragStore";
import { panState } from "../controls/PanControls";

interface AtmosphereRendererProps {
  sunLight: RefObject<THREE.PointLight | null>;
  enabled?: boolean;
  shakeEnabled?: boolean;
  reportGpuStats?: boolean;
  children: ReactNode;
}

export default function AtmosphereRenderer({
  sunLight,
  enabled = true,
  shakeEnabled = false,
  reportGpuStats = false,
  children,
}: AtmosphereRendererProps) {
  const { size, gl: renderer } = useThree();
  const reducedMotion = useReducedMotion();
  const { settings } = useWorld3DSettings();
  const effects = useMemo(() => new Set<NuclearEffect>(), []);
  const compositor = useMemo(() => new NuclearCompositor(), []);
  const bodies = useMemo(() => new Set<AtmosphereBody>(), []);
  const uniforms = useMemo(() => createPlanetHazeUniforms(), []);
  const materials = useMemo(() => new Map<THREE.Material, () => void>(), []);
  const highlights = useMemo(() => new TileHighlightSystem(), []);
  const gpuTimerRef = useRef<GpuTimer | null>(null);
  useLayoutEffect(() => {
    if (!reportGpuStats) {
      return;
    }
    const timer = new GpuTimer(renderer.getContext() as WebGL2RenderingContext);
    gpuTimerRef.current = timer;
    return () => {
      gpuTimerRef.current = null;
      timer.dispose();
    };
  }, [reportGpuStats, renderer]);
  const hidden = useMemo(() => [] as THREE.Object3D[], []);
  const scratch = useMemo(
    () => ({
      center: new THREE.Vector3(),
      viewCenter: new THREE.Vector3(),
      sun: new THREE.Vector3(),
      profile: new THREE.Vector3(),
    }),
    [],
  );

  useLayoutEffect(
    () => () => {
      highlights.dispose();
      materials.forEach((restore) => restore());
      materials.clear();
      compositor.dispose();
    },
    [materials, compositor, highlights],
  );

  // Runs over every visible mesh each frame, so it must not allocate (no closures or arrays per call).
  const hazeVisitor = useMemo(() => {
    const patch = (material: THREE.Material) => {
      if (materials.has(material) || !receivesPlanetHaze(material)) {
        return;
      }
      const restore = addPlanetHaze(material, uniforms);
      const release = () => {
        material.removeEventListener("dispose", release);
        materials.delete(material);
        restore();
      };
      material.addEventListener("dispose", release);
      materials.set(material, release);
    };
    return (object: THREE.Object3D) => {
      if (!(object instanceof THREE.Mesh)) {
        return;
      }
      if (Array.isArray(object.material)) {
        for (const material of object.material) {
          patch(material);
        }
      } else {
        patch(object.material);
      }
    };
  }, [materials, uniforms]);

  const materialVisitor = useMemo(
    () => (object: THREE.Object3D) => {
      if (ATMOSPHERE_ENABLED) {
        hazeVisitor(object);
      }
      highlights.visit(object);
    },
    [hazeVisitor, highlights],
  );

  useFrame(({ gl, scene, camera, clock }, delta) => {
    scene.updateMatrixWorld();
    camera.updateMatrixWorld();
    highlights.tick(
      delta,
      clock.elapsedTime,
      camera,
      !panState.isPanning && !useCardDragStore.getState().isDraggingCard,
    );
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
      const visibility =
        enabled && ATMOSPHERE_ENABLED ? THREE.MathUtils.smoothstep(projectedRadius, 3, 12) : 0;
      const intensity = body.profile.intensity * (light?.intensity ?? 0) * visibility;
      const halo = body.halo.uniforms;
      halo.uSunDirection.value.copy(scratch.sun).sub(scratch.center).normalize();
      if (light) {
        halo.uSunColor.value.copy(light.color);
      }
      halo.uIntensity.value = intensity * body.profile.halo;
      // Profiles can change at runtime (Mars's climate); parse colours only when they do.
      if (body.halo.userData.color !== body.profile.color) {
        body.halo.userData.color = body.profile.color;
        halo.uColor.value.set(body.profile.color);
      }
      if (body.halo.userData.shadowColor !== body.profile.shadowColor) {
        body.halo.userData.shadowColor = body.profile.shadowColor;
        halo.uShadowColor.value.set(body.profile.shadowColor);
      }
      halo.uThickness.value = body.profile.thickness;
      if (count === MAX_ATMOSPHERES || intensity === 0 || scratch.viewCenter.z > body.radius) {
        continue;
      }
      scratch.center.toArray(uniforms.uHazeBodies.value, count * 4);
      uniforms.uHazeBodies.value[count * 4 + 3] = body.radius;
      halo.uColor.value.toArray(uniforms.uHazeColors.value, count * 3);
      halo.uShadowColor.value.toArray(uniforms.uHazeShadowColors.value, count * 3);
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
    scene.traverseVisible(materialVisitor);
    const autoUpdate = scene.matrixWorldAutoUpdate;
    const infoAutoReset = gl.info.autoReset;
    gl.info.autoReset = false;
    gl.info.reset();
    scene.matrixWorldAutoUpdate = false;
    hideGroups(scene, hidden);
    const gpuTimer = gpuTimerRef.current;
    const timing = gpuTimer !== null && performanceStore.watching;
    try {
      if (timing) {
        gpuTimer.begin();
      }
      compositor.render(
        gl,
        scene,
        camera,
        effects,
        shakeEnabled && !reducedMotion && !settings.freeCameraEnabled,
        size.width,
        size.height,
      );
      if (timing) {
        gpuTimer.end();
        if (performanceStore.sectionDue("GPU")) {
          const { average, worst, count } = gpuTimer.stats();
          performanceStore.setSection(
            "GPU",
            gpuTimer.available && count > 0
              ? { "frame ms (avg)": average.toFixed(2), "frame ms (worst)": worst.toFixed(2) }
              : { "frame ms": "unavailable" },
          );
        }
      }
      if (reportGpuStats) {
        performanceStore.updateGpuStats({
          drawCalls: gl.info.render.calls,
          triangles: gl.info.render.triangles,
          textureCount: gl.info.memory.textures,
          geometryCount: gl.info.memory.geometries,
        });
        if (performanceStore.detailDue(performance.now())) {
          performanceStore.updateCensus(sceneCensus(scene));
        }
      }
    } finally {
      for (const object of hidden) {
        object.visible = true;
      }
      hidden.length = 0;
      scene.matrixWorldAutoUpdate = autoUpdate;
      gl.info.autoReset = infoAutoReset;
    }
  }, 1);

  return (
    <NuclearEffectsContext.Provider value={effects}>
      <AtmospheresContext.Provider value={bodies}>
        <TileHighlightSystemContext.Provider value={highlights}>
          {children}
        </TileHighlightSystemContext.Provider>
      </AtmospheresContext.Provider>
    </NuclearEffectsContext.Provider>
  );
}

// Hides the scene groups toggled off in the performance window for this render only.
function hideGroups(scene: THREE.Scene, hidden: THREE.Object3D[]) {
  if (performanceStore.hiddenGroups.size === 0) {
    return;
  }
  scene.traverseVisible((object) => {
    const group = object.userData.perfGroup as string | undefined;
    if (group && performanceStore.hiddenGroups.has(group)) {
      hidden.push(object);
    }
  });
  for (const object of hidden) {
    object.visible = false;
  }
}
