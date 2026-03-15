import { Suspense, useEffect, useState, useRef, useCallback } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { PanControls } from "../controls/PanControls.tsx";
import { FreeCamera, CameraFrustumHelper } from "../controls/FreeCamera.tsx";
import MarsSphere from "../board/MarsSphere.tsx";
import VenusSphere from "../board/VenusSphere.tsx";

import SkyboxLoader from "./SkyboxLoader.tsx";
import GameIcon from "../../ui/display/GameIcon.tsx";
import { GameDto } from "@/types/generated/api-types.ts";
import { MarsRotationProvider } from "../../../contexts/MarsRotationContext.tsx";
import { PlanetFocusProvider, usePlanetFocus } from "../../../contexts/PlanetFocusContext.tsx";
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

function DynamicSunLight({ startDark = false }: { startDark?: boolean }) {
  const { settings } = useWorld3DSettings();
  const lightRef = useRef<THREE.DirectionalLight>(null);
  const sunriseStartTime = useRef<number | null>(null);

  useFrame((state) => {
    if (!lightRef.current) {
      return;
    }

    const distance = 18;
    lightRef.current.position.set(
      settings.sunDirectionX * distance,
      settings.sunDirectionY * distance,
      settings.sunDirectionZ * distance + 5,
    );

    let intensityMultiplier = 1;
    if (startDark) {
      intensityMultiplier = 0;
      sunriseStartTime.current = null;
    } else if (sunriseStartTime.current === null) {
      sunriseStartTime.current = state.clock.elapsedTime;
      intensityMultiplier = 0;
    } else {
      const elapsed = state.clock.elapsedTime - sunriseStartTime.current;
      const t = Math.min(elapsed / 1.5, 1);
      intensityMultiplier = 1 - (1 - t) * (1 - t);
    }

    lightRef.current.intensity = 2.6 * settings.sunIntensity * intensityMultiplier;
    lightRef.current.color.setRGB(settings.sunColor.r, settings.sunColor.g, settings.sunColor.b);
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

function AutoNavigateForTileSelection({ gameState }: { gameState: GameDto }) {
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const pendingTileSelection = gameState.currentPlayer?.pendingTileSelection;

  useEffect(() => {
    if (!pendingTileSelection || !gameState.board?.tiles) return;

    const venusTileKeys = new Set(
      gameState.board.tiles
        .filter((t) => t.location === "venus")
        .map((t) => `${t.coordinates.q},${t.coordinates.r},${t.coordinates.s}`),
    );

    const hasVenusHex = pendingTileSelection.availableHexes.some((hex) => venusTileKeys.has(hex));
    const hasNonVenusHex = pendingTileSelection.availableHexes.some(
      (hex) => !venusTileKeys.has(hex),
    );

    if (hasVenusHex && !hasNonVenusHex && activePlanet !== "venus") {
      setActivePlanet("venus");
    } else if (!hasVenusHex && hasNonVenusHex && activePlanet !== "mars") {
      setActivePlanet("mars");
    }
  }, [pendingTileSelection, gameState.board?.tiles, activePlanet, setActivePlanet]);

  return null;
}

function ReturnToMarsButton() {
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const showButton = activePlanet === "venus";
  const [mounted, setMounted] = useState(false);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (showButton) {
      setMounted(true);
      requestAnimationFrame(() => setVisible(true));
      return;
    }
    setVisible(false);
    const timeout = setTimeout(() => setMounted(false), 300);
    return () => clearTimeout(timeout);
  }, [showButton]);

  if (!mounted) return null;
  return (
    <div
      className="absolute right-[15%] top-1/2 transform -translate-y-1/2 z-50"
      style={{
        opacity: visible ? 1 : 0,
        transition: "opacity 0.3s ease-out",
      }}
    >
      <button
        onClick={() => setActivePlanet("mars")}
        className="bg-space-black/90 backdrop-blur-space border border-space-blue-500
                   rounded-lg px-6 py-3 shadow-glow-lg font-orbitron text-sm text-white
                   tracking-wider-2xl hover:border-white/50 transition-colors cursor-pointer"
      >
        RETURN TO MARS
      </button>
    </div>
  );
}

interface Game3DViewProps {
  gameState: GameDto;
  animateHexEntrance?: boolean;
  startDark?: boolean;
  tilesHidden?: boolean;
  onSkyboxReady?: () => void;
  onGpuReady?: () => void;
  showUI?: boolean;
  uiAnimationClass?: string;
}

export default function Game3DView({
  gameState,
  animateHexEntrance = false,
  startDark = false,
  tilesHidden = false,
  onSkyboxReady,
  onGpuReady,
  showUI = true,
  uiAnimationClass = "",
}: Game3DViewProps) {
  const venusNextEnabled = gameState.settings?.venusNextEnabled ?? false;
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
    <PlanetFocusProvider>
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

        {venusNextEnabled && <AutoNavigateForTileSelection gameState={gameState} />}
        {venusNextEnabled && <ReturnToMarsButton />}

        <Canvas
          camera={{
            position: cameraConfig.position,
            fov: cameraConfig.fov,
          }}
          style={{
            background: "#000000",
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
              <SkyboxLoader onReady={onSkyboxReady} />
              <SkyboxRotation />

              <ambientLight intensity={0.4} color="#2a2a3e" />
              <DynamicSunLight startDark={startDark} />
              <directionalLight position={[-8, -3, -10]} intensity={0.35} color="#4488ff" />
              <fog attach="fog" args={["#0a0a1a", 8, 25]} />

              <MarsSphere
                gameState={gameState}
                onHexClick={handleHexClick}
                animateHexEntrance={animateHexEntrance}
                startHidden={tilesHidden}
              />

              {venusNextEnabled && (
                <VenusSphere gameState={gameState} onHexClick={handleHexClick} />
              )}

              <GpuWarmup onReady={onGpuReady} />
              <PerformanceProbe />

              <PanControls />
              <FreeCamera />
              <FreeCameraFrustum fov={cameraConfig.fov} />
            </Suspense>
          </MarsRotationProvider>
        </Canvas>
      </div>
    </PlanetFocusProvider>
  );
}
