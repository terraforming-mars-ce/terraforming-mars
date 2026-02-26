import { Suspense, useEffect, useState, useRef, useCallback } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { PanControls } from "../controls/PanControls.tsx";
import { FreeCamera, CameraFrustumHelper } from "../controls/FreeCamera.tsx";
import MarsSphere from "../board/MarsSphere.tsx";
import { TileHighlightMode } from "../board/Tile.tsx";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import SkyboxLoader from "./SkyboxLoader.tsx";
import GameIcon from "../../ui/display/GameIcon.tsx";
import { GameDto } from "@/types/generated/api-types.ts";
import { MarsRotationProvider } from "../../../contexts/MarsRotationContext.tsx";
import { webSocketService } from "../../../services/webSocketService.ts";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext.tsx";
import GpuWarmup from "../board/GpuWarmup.tsx";
import PerformanceProbe from "../board/PerformanceProbe.tsx";

function SkyboxRotation() {
  const { scene } = useThree();

  useFrame((_, delta) => {
    const skybox = scene.children.find(
      (child) =>
        child instanceof THREE.Mesh &&
        child.geometry instanceof THREE.SphereGeometry &&
        (child.material as THREE.MeshBasicMaterial).side === THREE.BackSide,
    );
    if (skybox) {
      skybox.rotation.y += delta * 0.002;
    }
  });

  return null;
}

function FreeCameraFrustum({ fov }: { fov: number }) {
  const { size } = useThree();
  const { storedCameraState } = useWorld3DSettings();

  if (!storedCameraState) return null;

  return (
    <CameraFrustumHelper
      storedState={storedCameraState}
      fov={fov}
      aspect={size.width / size.height}
    />
  );
}

function DynamicSunLight() {
  const { settings } = useWorld3DSettings();
  const lightRef = useRef<THREE.DirectionalLight>(null);

  useFrame(() => {
    if (lightRef.current) {
      // Scale direction by distance to position the light
      const distance = 18;
      lightRef.current.position.set(
        settings.sunDirectionX * distance,
        settings.sunDirectionY * distance,
        settings.sunDirectionZ * distance + 5,
      );
      lightRef.current.intensity = 2.6 * settings.sunIntensity;
      lightRef.current.color.setRGB(settings.sunColor.r, settings.sunColor.g, settings.sunColor.b);
    }
  });

  return (
    <directionalLight
      ref={lightRef}
      position={[8, 6, 15]}
      intensity={2.6}
      color="#ffdcb8"
      castShadow
      shadow-mapSize-width={2048}
      shadow-mapSize-height={2048}
      shadow-camera-far={50}
      shadow-camera-left={-20}
      shadow-camera-right={-20}
      shadow-camera-top={20}
      shadow-camera-bottom={-20}
    />
  );
}

interface Game3DViewProps {
  gameState: GameDto;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  animateHexEntrance?: boolean;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
  showUI?: boolean;
  uiAnimationClass?: string;
}

export default function Game3DView({
  gameState,
  tileHighlightMode,
  vpIndicators = [],
  animateHexEntrance = false,
  onSkyboxReady,
  onGpuReady,
  showUI = true,
  uiAnimationClass = "",
}: Game3DViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [cameraConfig, setCameraConfig] = useState({
    position: [0, 0, 8] as [number, number, number],
    fov: 50,
  });

  const updateCameraConfig = useCallback(() => {
    const width = window.innerWidth;
    let fov = 50;
    let position: [number, number, number] = [0, 0, 8];

    if (width <= 768) {
      fov = 60;
      position = [0, 0, 10];
    } else if (width <= 1200) {
      fov = 55;
      position = [0, 0, 9];
    }

    setCameraConfig({ position, fov });
  }, []);

  useEffect(() => {
    updateCameraConfig();
    window.addEventListener("resize", updateCameraConfig);

    return () => window.removeEventListener("resize", updateCameraConfig);
  }, [updateCameraConfig]);

  const handleHexClick = useCallback(
    (hexCoordinate: string) => {
      // Parse hexCoordinate string (format: "q,r,s") back to coordinate object
      const [q, r, s] = hexCoordinate.split(",").map(Number);
      const coordinate = { q, r, s };

      // Check if current player has a pending tile selection (from cards OR standard projects)
      const currentPlayer = gameState.currentPlayer;
      if (!currentPlayer?.pendingTileSelection) {
        return;
      }

      const { pendingTileSelection } = currentPlayer;

      // Validate that the clicked hex is in the available positions provided by backend
      if (!pendingTileSelection.availableHexes.includes(hexCoordinate)) {
        return;
      }

      // Send tile selection to backend (works for both cards and standard projects)
      try {
        webSocketService.selectTile(coordinate);
      } catch (error) {
        console.error("❌ Failed to send tile selection:", error);
      }
    },
    [gameState.currentPlayer],
  );

  // Determine tile icon type from tileType string
  const getTileIconType = (tileType: string): string => {
    switch (tileType) {
      case "city":
        return "city-tile";
      case "greenery":
        return "greenery-tile";
      case "ocean":
        return "ocean-tile";
      case "volcano":
        return "volcano-tile";
      default:
        return "city-tile"; // fallback
    }
  };

  const pendingTileSelection = gameState.currentPlayer?.pendingTileSelection;

  return (
    <div
      ref={containerRef}
      style={{
        flex: 1,
        height: "100%",
        width: "100%",
        minHeight: 0,
        position: "relative",
      }}
    >
      {/* Tile Selection Banner */}
      {showUI && pendingTileSelection && (
        <div
          className={`absolute top-[66px] left-1/2 transform -translate-x-1/2 z-50
                     bg-space-black/90 backdrop-blur-space border border-space-blue-500
                     rounded-lg px-6 py-3 shadow-glow-lg ${uiAnimationClass}`}
        >
          <div className="flex items-center gap-2">
            <span className="font-orbitron text-lg text-white tracking-wider-2xl">Place</span>
            <GameIcon iconType={getTileIconType(pendingTileSelection.tileType)} size="medium" />
          </div>
        </div>
      )}

      <Canvas
        camera={{
          position: cameraConfig.position,
          fov: cameraConfig.fov,
        }}
        style={{
          background: "#000000", // Fallback background
          width: "100%",
          height: "100%",
          position: "relative",
          zIndex: 0,
        }}
        resize={{ scroll: false, debounce: { scroll: 50, resize: 0 } }}
        dpr={typeof window !== "undefined" ? window.devicePixelRatio : 1}
        shadows
      >
        <MarsRotationProvider>
          <Suspense fallback={null}>
            {/* EXR Skybox - now uses cached texture */}
            <SkyboxLoader onReady={onSkyboxReady} />
            <SkyboxRotation />

            {/* Realistic Lighting Setup */}
            {/* Ambient light - enough to keep shadow side of vegetation visible */}
            <ambientLight intensity={0.4} color="#2a2a3e" />

            {/* Main sun light - controlled by 3D World settings */}
            <DynamicSunLight />

            {/* Cool blue rim light for moody atmosphere */}
            <directionalLight position={[-8, -3, -10]} intensity={0.35} color="#4488ff" />

            {/* Atmospheric fog for depth and mood */}
            <fog attach="fog" args={["#0a0a1a", 8, 25]} />

            {/* Mars with hexagonal tiles */}
            <MarsSphere
              gameState={gameState}
              onHexClick={handleHexClick}
              tileHighlightMode={tileHighlightMode}
              vpIndicators={vpIndicators}
              animateHexEntrance={animateHexEntrance}
            />

            <GpuWarmup onReady={onGpuReady} />

            <PerformanceProbe />

            {/* Camera controls */}
            <PanControls />
            <FreeCamera />
            <FreeCameraFrustum fov={cameraConfig.fov} />
          </Suspense>
        </MarsRotationProvider>
      </Canvas>
    </div>
  );
}
