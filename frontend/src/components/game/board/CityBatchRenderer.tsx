import { memo, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import type { CityPlot } from "./cityLayout";
import { CityBatchStore } from "./cityBatch";
import { CITY_EMERGENCE_DURATION } from "./boardConstants";
import { getMaterials } from "./CityRenderer";
import { splitSnippet } from "./shaders";
import emergence from "./shaders/city-emergence.vert.glsl?raw";

function CityBatchRenderer({
  plots,
  newlyPlaced,
}: {
  plots: CityPlot[];
  newlyPlaced: Set<string>;
}) {
  const textures = useTextures();
  const birthTimes = useRef(new Map<string, number>());
  const time = useMemo(() => ({ value: performance.now() / 1000 }), []);
  useFrame(() => {
    time.value = performance.now() / 1000;
  });
  const materials = useMemo(
    () => createCityBatchMaterials(textures, time),
    [textures.concrete, textures.grass, textures.sand, textures.cityFacades, time],
  );
  const store = useMemo(() => new CityBatchStore(materials), [materials]);
  useLayoutEffect(() => {
    const keys = new Set(plots.map((p) => HexGrid2D.coordinateToKey(p.coordinate)));
    for (const key of birthTimes.current.keys()) {
      if (!keys.has(key)) {
        birthTimes.current.delete(key);
      }
    }
    for (const key of keys) {
      if (!birthTimes.current.has(key)) {
        birthTimes.current.set(key, newlyPlaced.has(key) ? performance.now() / 1000 : -1000);
      }
    }
    store.update(plots, birthTimes.current);
  }, [plots, store]);
  useLayoutEffect(() => () => store.dispose(), [store]);
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
  textures: Parameters<typeof getMaterials>[1],
  time: { value: number },
) {
  const result = new Map<string, THREE.MeshStandardMaterial>();
  const snippet = splitSnippet(emergence);
  for (const mode of ["normal", "bright"] as const) {
    for (const [index, source] of getMaterials(mode, textures).entries()) {
      if (mode === "bright" && index >= 3) {
        continue;
      }
      const material = source.clone();
      material.onBeforeCompile = (shader) => {
        shader.uniforms.uCityTime = time;
        shader.uniforms.uCityDuration = { value: CITY_EMERGENCE_DURATION };
        shader.vertexShader =
          snippet.header +
          "\n" +
          shader.vertexShader.replace("#include <begin_vertex>", snippet.body);
      };
      result.set(`${mode}:${index}`, material);
    }
  }
  return result;
}

export default memo(CityBatchRenderer);
