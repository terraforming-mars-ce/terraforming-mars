import { useFrame, useThree } from "@react-three/fiber";
import { useLayoutEffect } from "react";
import type { Texture } from "three";
import { coldStartTrace, performanceStore } from "@/services/performanceStore.ts";

function textureDetails(texture?: Texture | null) {
  const image = texture?.image as HTMLImageElement | undefined;
  return {
    texture: texture?.uuid,
    url: image?.currentSrc || image?.src,
    width: image?.width,
    height: image?.height,
  };
}

export default function PerformanceProbe() {
  const { gl } = useThree();

  useLayoutEffect(() => {
    if (!coldStartTrace.enabled) {
      return;
    }
    const render = gl.render;
    const draw = gl.renderBufferDirect;
    const initTexture = gl.initTexture;
    let captureFirstRender = true;
    gl.render = function (...args) {
      if (!coldStartTrace.active) {
        captureFirstRender = true;
        return render.apply(this, args);
      }
      const programs = gl.info.programs?.length ?? 0;
      const textures = gl.info.memory.textures;
      const geometries = gl.info.memory.geometries;
      const start = performance.now();
      try {
        return render.apply(this, args);
      } finally {
        const duration = performance.now() - start;
        if (
          captureFirstRender ||
          duration >= 8 ||
          programs !== gl.info.programs?.length ||
          textures !== gl.info.memory.textures ||
          geometries !== gl.info.memory.geometries
        ) {
          coldStartTrace.record("webgl:render-cpu", start, duration, {
            programsBefore: programs,
            programsAfter: gl.info.programs?.length,
            texturesBefore: textures,
            texturesAfter: gl.info.memory.textures,
            geometriesBefore: geometries,
            geometriesAfter: gl.info.memory.geometries,
            drawCalls: gl.info.render.calls,
            triangles: gl.info.render.triangles,
          });
        }
        captureFirstRender = false;
      }
    };
    gl.renderBufferDirect = function (...args) {
      if (!coldStartTrace.active) {
        return draw.apply(this, args);
      }
      const programs = gl.info.programs?.length ?? 0;
      const start = performance.now();
      try {
        return draw.apply(this, args);
      } finally {
        const duration = performance.now() - start;
        if (duration >= 2 || programs !== gl.info.programs?.length) {
          const material = args[3];
          const map = "map" in material ? (material.map as Texture | null) : null;
          coldStartTrace.record("webgl:draw-cpu", start, duration, {
            material: material.type,
            materialId: material.uuid,
            object: args[4].name || args[4].type,
            vertexColors: material.vertexColors,
            programsAdded: (gl.info.programs?.length ?? 0) - programs,
            ...textureDetails(map),
          });
        }
      }
    };
    gl.initTexture = function (texture) {
      if (!coldStartTrace.active) {
        return initTexture.call(this, texture);
      }
      const before = gl.info.memory.textures;
      const start = performance.now();
      try {
        return initTexture.call(this, texture);
      } finally {
        const duration = performance.now() - start;
        if (duration >= 2 || gl.info.memory.textures !== before) {
          coldStartTrace.record("webgl:init-texture-cpu", start, duration, textureDetails(texture));
        }
      }
    };
    coldStartTrace.begin("scene-warmup");
    return () => {
      gl.render = render;
      gl.renderBufferDirect = draw;
      gl.initTexture = initTexture;
    };
  }, [gl]);

  useFrame(() => {
    const info = gl.info;
    performanceStore.updateGpuStats({
      drawCalls: info.render.calls,
      triangles: info.render.triangles,
      textureCount: info.memory.textures,
      geometryCount: info.memory.geometries,
    });
  });

  return null;
}
