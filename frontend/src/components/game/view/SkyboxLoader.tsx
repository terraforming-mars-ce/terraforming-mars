import { useEffect, useLayoutEffect, useMemo, useState } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { skyboxCache } from "../../../services/SkyboxCache.ts";
import { useWorld3DSettings, SKYBOX_OPTIONS } from "../../../contexts/World3DSettingsContext.tsx";
import skyboxVertex from "../board/shaders/skybox.vert.glsl?raw";
import skyboxFragment from "../board/shaders/skybox.frag.glsl?raw";
import { QUICK_MODE } from "../../../utils/quickMode.ts";
import { GRAPHICS } from "../../../utils/graphicsQuality.ts";
import ProceduralStarfield from "../../3d/ProceduralStarfield.tsx";

interface SkyboxLoaderProps {
  onReady?: () => void;
}

export default function SkyboxLoader({ onReady }: SkyboxLoaderProps) {
  if (QUICK_MODE) {
    return <SkippedSkybox onReady={onReady} />;
  }
  if (GRAPHICS.skybox === "procedural") {
    return <ProceduralSkybox onReady={onReady} />;
  }
  return <LoadedSkybox onReady={onReady} />;
}

/** Quick mode: report ready straight away and leave the plain black background. */
function SkippedSkybox({ onReady }: SkyboxLoaderProps) {
  useEffect(() => {
    onReady?.();
  }, [onReady]);
  return null;
}

function ProceduralSkybox({ onReady }: SkyboxLoaderProps) {
  useEffect(() => {
    onReady?.();
  }, [onReady]);
  return <ProceduralStarfield />;
}

function LoadedSkybox({ onReady }: SkyboxLoaderProps) {
  const { scene, invalidate } = useThree();
  const { settings } = useWorld3DSettings();
  const [texture, setTexture] = useState<THREE.Texture | null>(null);
  const uniforms = useMemo(
    () => ({ uSky: { value: texture }, uBrightness: { value: 0.35 } }),
    [texture],
  );

  const skyboxPath =
    SKYBOX_OPTIONS.find((o) => o.id === settings.skyboxId)?.path ?? SKYBOX_OPTIONS[0].path;

  useEffect(() => {
    let cancelled = false;

    skyboxCache
      .loadSkybox(skyboxPath)
      .then((loadedTexture) => {
        if (cancelled) {
          return;
        }
        loadedTexture.mapping = THREE.EquirectangularReflectionMapping;
        loadedTexture.colorSpace = THREE.LinearSRGBColorSpace;
        setTexture(loadedTexture);
      })
      .catch((error) => {
        console.error("Failed to load skybox:", error);
      });

    return () => {
      cancelled = true;
    };
  }, [skyboxPath]);

  useEffect(() => {
    if (!texture) {
      return;
    }
    scene.environment = texture;
    onReady?.();
    invalidate();
    return () => {
      if (scene.environment === texture) {
        scene.environment = null;
      }
    };
  }, [scene, texture, onReady, invalidate]);

  useLayoutEffect(() => {
    uniforms.uBrightness.value = settings.skyboxBrightness;
    invalidate();
  }, [uniforms, settings.skyboxBrightness, invalidate]);

  if (!texture) {
    return null;
  }

  return (
    <mesh frustumCulled={false} renderOrder={-1000} raycast={() => {}}>
      <sphereGeometry args={[1, 32, 16]} />
      <shaderMaterial
        vertexShader={skyboxVertex}
        fragmentShader={skyboxFragment}
        uniforms={uniforms}
        side={THREE.BackSide}
        depthTest={false}
        depthWrite={false}
        fog={false}
      />
    </mesh>
  );
}
