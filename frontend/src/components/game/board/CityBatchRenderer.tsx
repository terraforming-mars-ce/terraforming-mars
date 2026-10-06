import { memo, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { useClimate } from "../../../contexts/ClimateContext";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import type { CityPlot } from "./cityLayout";
import { CityBatchStore } from "./cityBatch";
import { CITY_EMERGENCE_DURATION } from "./boardConstants";
import { NUCLEAR_PLAYBACK_RATE } from "./nuclearGeometry";
import { getMaterials } from "./CityRenderer";
import { splitSnippet } from "./shaders";
import type { NuclearTransitions } from "./nuclearTransitions";
import collapseNormal from "./shaders/city-collapse-normal.vert.glsl?raw";
import emergence from "./shaders/city-emergence.vert.glsl?raw";
import frostFragment from "./shaders/city-frost.frag.glsl?raw";

function CityBatchRenderer({
  plots,
  newlyPlaced,
  nuclearTransitions = EMPTY_TRANSITIONS,
  visible = true,
  onReady,
}: {
  plots: CityPlot[];
  newlyPlaced: Set<string>;
  nuclearTransitions?: NuclearTransitions;
  visible?: boolean;
  onReady?: () => void;
}) {
  const clock = useThree((s) => s.clock);
  const textures = useTextures();
  const { runtime } = useClimate();
  const birthTimes = useRef(new Map<string, number>());
  const time = useMemo(() => ({ value: clock.elapsedTime }), []);
  const frost = useMemo(() => ({ value: runtime.current.current.frost }), [runtime]);
  const materials = useMemo(
    () => createCityBatchMaterials(textures, time, frost),
    [
      textures.concrete,
      textures.grass,
      textures.sand,
      textures.cityFacades,
      textures.noiseMid,
      time,
      frost,
    ],
  );
  const store = useMemo(() => new CityBatchStore(materials), [materials]);
  const latest = useRef({ plots, nuclearTransitions, onReady });
  useFrame(() => {
    time.value = clock.elapsedTime;
    frost.value = runtime.current.current.frost;
    store.group.visible = !store.ready || visible;
  });
  useLayoutEffect(() => {
    store.beginPreparation();
    let frame = 0;
    const prepare = () => {
      if (!store.prepareNext()) {
        frame = requestAnimationFrame(prepare);
        return;
      }
      for (const [key, birth] of birthTimes.current) {
        if (birth >= 0) {
          birthTimes.current.set(key, clock.elapsedTime);
        }
      }
      store.update(latest.current.plots, birthTimes.current);
      store.setCollapses(latest.current.nuclearTransitions);
      latest.current.onReady?.();
    };
    frame = requestAnimationFrame(prepare);
    return () => {
      cancelAnimationFrame(frame);
      store.dispose();
    };
  }, [store, clock]);
  useLayoutEffect(() => {
    latest.current = { plots, nuclearTransitions, onReady };
    const keys = new Set(plots.map((p) => HexGrid2D.coordinateToKey(p.coordinate)));
    for (const key of birthTimes.current.keys()) {
      if (!keys.has(key)) {
        birthTimes.current.delete(key);
      }
    }
    for (const key of keys) {
      if (!birthTimes.current.has(key)) {
        birthTimes.current.set(key, newlyPlaced.has(key) ? clock.elapsedTime : -1000);
      }
    }
    if (store.ready) {
      store.update(plots, birthTimes.current);
      store.setCollapses(nuclearTransitions);
    }
  }, [plots, store, nuclearTransitions, newlyPlaced, onReady, clock]);
  useEffect(
    () => () => {
      for (const material of materials.values()) {
        material.dispose();
      }
    },
    [materials],
  );
  return <primitive object={store.group} dispose={null} />;
}

export function createCityBatchMaterials(
  textures: Parameters<typeof getMaterials>[1] & Pick<ReturnType<typeof useTextures>, "noiseMid">,
  time: { value: number },
  frost: { value: number },
) {
  const result = new Map<string, THREE.MeshStandardMaterial>();
  const snippet = splitSnippet(emergence);
  const frostSnippet = splitSnippet(frostFragment);
  for (const mode of ["normal", "bright"] as const) {
    for (const [index, source] of getMaterials(mode, textures).entries()) {
      if (mode === "bright" && index >= 3) {
        continue;
      }
      const material = source.clone();
      const frostable = index !== 11 && index !== 14 && index !== 15;
      material.onBeforeCompile = (shader) => {
        shader.uniforms.uCityTime = time;
        shader.uniforms.uCityDuration = { value: CITY_EMERGENCE_DURATION };
        shader.uniforms.uNuclearPlaybackRate = { value: NUCLEAR_PLAYBACK_RATE };
        shader.vertexShader =
          snippet.header +
          "\n" +
          shader.vertexShader
            .replace("#include <begin_vertex>", snippet.body)
            .replace("#include <beginnormal_vertex>", collapseNormal);
        if (frostable) {
          shader.uniforms.uCityFrost = frost;
          shader.uniforms.uCityFrostNoise = { value: textures.noiseMid };
          shader.uniforms.uCityFrostGlass = { value: index === 10 ? 1 : 0 };
          shader.fragmentShader =
            frostSnippet.header +
            "\n" +
            shader.fragmentShader.replace(
              "#include <emissivemap_fragment>",
              "#include <emissivemap_fragment>\n" + frostSnippet.body,
            );
        }
      };
      material.customProgramCacheKey = () =>
        emergence + collapseNormal + (frostable ? frostFragment : "");
      result.set(`${mode}:${index}`, material);
    }
  }
  return result;
}

const EMPTY_TRANSITIONS: NuclearTransitions = new Map();
export default memo(CityBatchRenderer);
