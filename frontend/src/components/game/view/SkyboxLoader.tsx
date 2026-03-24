import { useEffect } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { skyboxCache } from "../../../services/SkyboxCache.ts";
import { useWorld3DSettings, SKYBOX_OPTIONS } from "../../../contexts/World3DSettingsContext.tsx";

interface SkyboxLoaderProps {
  onReady?: () => void;
}

export default function SkyboxLoader({ onReady }: SkyboxLoaderProps) {
  const { scene } = useThree();
  const { settings } = useWorld3DSettings();

  const skyboxPath =
    SKYBOX_OPTIONS.find((o) => o.id === settings.skyboxId)?.path ?? SKYBOX_OPTIONS[0].path;

  useEffect(() => {
    let cancelled = false;

    skyboxCache
      .loadSkybox(skyboxPath)
      .then((texture) => {
        if (cancelled) {
          return;
        }
        texture.mapping = THREE.EquirectangularReflectionMapping;
        texture.colorSpace = THREE.LinearSRGBColorSpace;
        scene.background = texture;
        scene.environment = texture;
        onReady?.();
      })
      .catch((error) => {
        console.error("Failed to load skybox:", error);
      });

    return () => {
      cancelled = true;
      scene.background = null;
      scene.environment = null;
    };
  }, [scene, onReady, skyboxPath]);

  useEffect(() => {
    scene.backgroundIntensity = settings.skyboxBrightness;
  }, [scene, settings.skyboxBrightness]);

  return null;
}
