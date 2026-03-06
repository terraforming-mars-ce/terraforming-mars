import { useEffect, useRef } from "react";
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
  const skyboxRef = useRef<THREE.Mesh | null>(null);

  const skyboxPath =
    SKYBOX_OPTIONS.find((o) => o.id === settings.skyboxId)?.path ?? SKYBOX_OPTIONS[0].path;

  useEffect(() => {
    let cancelled = false;

    function setupSkybox(texture: THREE.Texture) {
      if (cancelled) return;
      try {
        if (skyboxRef.current) {
          scene.remove(skyboxRef.current);
          skyboxRef.current.geometry.dispose();
          if (skyboxRef.current.material instanceof THREE.Material) {
            skyboxRef.current.material.dispose();
          }
        }

        texture.mapping = THREE.EquirectangularReflectionMapping;
        texture.colorSpace = THREE.LinearSRGBColorSpace;

        const b = settings.skyboxBrightness;
        const geometry = new THREE.SphereGeometry(500, 32, 16);
        const material = new THREE.MeshBasicMaterial({
          map: texture,
          side: THREE.BackSide,
          fog: false,
          color: new THREE.Color(b, b, b),
        });

        const skyboxMesh = new THREE.Mesh(geometry, material);
        skyboxRef.current = skyboxMesh;

        scene.add(skyboxMesh);
        onReady?.();

        scene.environment = texture;
      } catch (error) {
        console.error("Failed to setup skybox:", error);
      }
    }

    skyboxCache
      .loadSkybox(skyboxPath)
      .then((texture) => {
        setupSkybox(texture);
      })
      .catch((error) => {
        console.error("Failed to load skybox:", error);
      });

    return () => {
      cancelled = true;
      if (skyboxRef.current) {
        scene.remove(skyboxRef.current);
        skyboxRef.current.geometry.dispose();
        if (skyboxRef.current.material instanceof THREE.Material) {
          skyboxRef.current.material.dispose();
        }
        skyboxRef.current = null;
      }

      if (scene.environment) {
        scene.environment = null;
      }
    };
  }, [scene, onReady, skyboxPath]);

  useEffect(() => {
    if (!skyboxRef.current) return;
    const mat = skyboxRef.current.material;
    if (mat instanceof THREE.MeshBasicMaterial) {
      const b = settings.skyboxBrightness;
      mat.color.setRGB(b, b, b);
    }
  }, [settings.skyboxBrightness]);

  return null;
}
