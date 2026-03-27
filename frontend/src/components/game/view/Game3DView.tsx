import { Suspense, useEffect, useMemo, useState, useRef, useCallback } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { PanControls } from "../controls/PanControls.tsx";
import { FreeCamera, CameraFrustumHelper } from "../controls/FreeCamera.tsx";
import MarsSphere from "../board/MarsSphere.tsx";
import CelestialBody from "../board/CelestialBody.tsx";
import PhobosBody from "../board/PhobosBody.tsx";
import OrbitalStation from "../board/OrbitalStation.tsx";
import AsteroidImpact from "../board/AsteroidImpact.tsx";

import SkyboxLoader from "./SkyboxLoader.tsx";
import GameIcon from "../../ui/display/GameIcon.tsx";
import { GameDto } from "@/types/generated/api-types.ts";
import { MarsRotationProvider } from "../../../contexts/MarsRotationContext.tsx";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext.tsx";
import { webSocketService } from "../../../services/webSocketService.ts";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext.tsx";
import { useTextures } from "../../../hooks/useTextures.ts";
import sunCoronaVert from "../board/shaders/sun-corona.vert.glsl?raw";
import sunCoronaFrag from "../board/shaders/sun-corona.frag.glsl?raw";
import GpuWarmup from "../board/GpuWarmup.tsx";
import PerformanceProbe from "../board/PerformanceProbe.tsx";
import SolarSystemOverview from "../board/SolarSystemOverview.tsx";
import {
  PLANET_CONFIGS,
  LOCATION_TO_PLANET,
  getMarsOrbitalPosition,
  getPlanetOrbit,
} from "../board/solarSystemConfig.ts";

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

function CentralSunLight({ startDark = false }: { startDark?: boolean }) {
  const { settings } = useWorld3DSettings();
  const lightRef = useRef<THREE.PointLight>(null);
  const sunriseStartTime = useRef<number | null>(null);

  useFrame((state) => {
    if (!lightRef.current) {
      return;
    }

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

    lightRef.current.intensity = settings.sunIntensity * intensityMultiplier;
    lightRef.current.color.setRGB(settings.sunColor.r, settings.sunColor.g, settings.sunColor.b);
  });

  return <pointLight ref={lightRef} position={[0, 0, 0]} intensity={1} distance={0} decay={0} />;
}

function SunMesh() {
  const { sun: sunTexture } = useTextures();

  const SUN_RADIUS = 22;

  const geometry = useMemo(() => new THREE.SphereGeometry(SUN_RADIUS, 64, 32), []);
  const material = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        map: sunTexture,
        color: new THREE.Color(1, 1, 1),
        fog: false,
      }),
    [sunTexture],
  );

  const coronaGeometry = useMemo(() => new THREE.SphereGeometry(SUN_RADIUS, 48, 24), []);

  const outerCoronaMaterial = useMemo(
    () =>
      new THREE.ShaderMaterial({
        uniforms: {
          glowColor: { value: new THREE.Color(1.0, 0.5, 0.1) },
          glowPower: { value: 2.0 },
          glowStrength: { value: 1.2 },
          uTime: { value: 0 },
          noiseScale: { value: 3.0 },
          noiseStrength: { value: 0.7 },
        },
        vertexShader: sunCoronaVert,
        fragmentShader: sunCoronaFrag,
        side: THREE.FrontSide,
        blending: THREE.AdditiveBlending,
        transparent: true,
        depthWrite: false,
        fog: false,
      }),
    [],
  );

  const innerCoronaMaterial = useMemo(
    () =>
      new THREE.ShaderMaterial({
        uniforms: {
          glowColor: { value: new THREE.Color(1.0, 0.85, 0.5) },
          glowPower: { value: 3.5 },
          glowStrength: { value: 2.0 },
          uTime: { value: 0 },
          noiseScale: { value: 5.0 },
          noiseStrength: { value: 0.3 },
        },
        vertexShader: sunCoronaVert,
        fragmentShader: sunCoronaFrag,
        side: THREE.FrontSide,
        blending: THREE.AdditiveBlending,
        transparent: true,
        depthWrite: false,
        fog: false,
      }),
    [],
  );

  useFrame((state) => {
    const t = state.clock.elapsedTime;
    outerCoronaMaterial.uniforms.uTime.value = t;
    innerCoronaMaterial.uniforms.uTime.value = t;
  });

  return (
    <group>
      <mesh geometry={geometry} material={material} />
      <mesh geometry={coronaGeometry} material={outerCoronaMaterial} scale={[1.3, 1.3, 1.3]} />
      <mesh geometry={coronaGeometry} material={innerCoronaMaterial} scale={[1.15, 1.15, 1.15]} />
    </group>
  );
}

function DynamicFog() {
  const { scene } = useThree();
  const { activePlanet } = usePlanetFocus();

  useFrame(() => {
    if (activePlanet === "solar-system") {
      scene.fog = null;
    } else if (!scene.fog) {
      scene.fog = new THREE.Fog("#0a0a0a", 8, 25);
    }
  });

  return <fog attach="fog" args={["#0a0a0a", 8, 25]} />;
}

function AutoNavigateForTileSelection({ gameState }: { gameState: GameDto }) {
  const { setActivePlanet } = usePlanetFocus();
  const pendingTileSelection = gameState.currentPlayer?.pendingTileSelection;

  useEffect(() => {
    if (!pendingTileSelection || !gameState.board?.tiles) {
      return;
    }

    const tileKeyToLocation = new Map<string, string>();
    for (const tile of gameState.board.tiles) {
      const key = `${tile.coordinates.q},${tile.coordinates.r},${tile.coordinates.s}`;
      tileKeyToLocation.set(key, tile.location);
    }

    const targetPlanets = new Set<string>();
    for (const hex of pendingTileSelection.availableHexes) {
      const location = tileKeyToLocation.get(hex);
      if (location) {
        const planet = LOCATION_TO_PLANET[location] || "mars";
        targetPlanets.add(planet);
      }
    }

    if (targetPlanets.size === 1) {
      const target = [...targetPlanets][0];
      setActivePlanet(target as Parameters<typeof setActivePlanet>[0]);
    }
    // Only navigate once when pendingTileSelection changes — don't re-trigger on manual navigation
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pendingTileSelection, gameState.board?.tiles]);

  return null;
}

function TravelFade() {
  const { activePlanet } = usePlanetFocus();
  const divRef = useRef<HTMLDivElement>(null);
  const prevPlanetRef = useRef(activePlanet);

  useEffect(() => {
    if (activePlanet === prevPlanetRef.current) {
      return;
    }
    prevPlanetRef.current = activePlanet;
    const el = divRef.current;
    if (!el) {
      return;
    }

    // Instantly go black (no transition) to hide the teleport
    el.style.transition = "none";
    el.style.opacity = "1";
    // Force reflow so the browser applies opacity:1 immediately
    void el.offsetHeight;
    // Then fade back to transparent
    el.style.transition = "opacity 400ms ease-in-out";
    el.style.opacity = "0";
  }, [activePlanet]);

  return (
    <div
      ref={divRef}
      style={{
        position: "absolute",
        inset: 0,
        backgroundColor: "black",
        opacity: 0,
        pointerEvents: "none",
        zIndex: 10,
      }}
    />
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
  const orbitalProject = gameState.projectFunding?.find((p) => p.id === "pf_orbital_station");
  const orbitalStationSeats = orbitalProject ? orbitalProject.seatOwners.length : 0;
  const containerRef = useRef<HTMLDivElement>(null);
  const initialCameraPos = useMemo((): [number, number, number] => {
    const mp = getMarsOrbitalPosition(0);
    const center = new THREE.Vector3(mp[0], mp[1], mp[2]);
    const radius = getPlanetOrbit("mars").defaultRadius;
    const offset = new THREE.Vector3().setFromSpherical(
      new THREE.Spherical(radius, Math.PI / 2, 0),
    );
    const toSun = center.clone().negate().normalize();
    const quat = new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 0, 1), toSun);
    offset.applyQuaternion(quat);
    return [center.x + offset.x, center.y + offset.y, center.z + offset.z];
  }, []);

  const [cameraConfig, setCameraConfig] = useState({
    position: initialCameraPos,
    fov: 50,
  });

  const updateCameraConfig = useCallback(() => {
    const width = window.innerWidth;
    let fov = 50;

    if (width <= 768) {
      fov = 60;
    } else if (width <= 1200) {
      fov = 55;
    }

    setCameraConfig({ position: initialCameraPos, fov });
  }, [initialCameraPos]);

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
        return "tile-placement";
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

      <AutoNavigateForTileSelection gameState={gameState} />
      <TravelFade />

      <Canvas
        camera={{
          position: cameraConfig.position,
          fov: cameraConfig.fov,
          near: 0.1,
          far: 5000,
        }}
        style={{
          background: "#000000",
          width: "100%",
          height: "100%",
          position: "relative",
          zIndex: 0,
        }}
        resize={{ scroll: false, debounce: { scroll: 50, resize: 0 } }}
        gl={{ stencil: true }}
        dpr={typeof window !== "undefined" ? window.devicePixelRatio : 1}
        shadows={{ type: THREE.PCFSoftShadowMap }}
      >
        <MarsRotationProvider>
          <Suspense fallback={null}>
            <SkyboxLoader onReady={onSkyboxReady} />

            <ambientLight intensity={0.4} color="#2a2a2a" />
            <CentralSunLight startDark={startDark} />
            <SunMesh />
            <DynamicFog />

            <MarsSphere
              gameState={gameState}
              onHexClick={handleHexClick}
              animateHexEntrance={animateHexEntrance}
              startHidden={tilesHidden}
            />

            <PhobosBody gameState={gameState} onHexClick={handleHexClick} />

            {PLANET_CONFIGS.map((config) => (
              <CelestialBody
                key={config.id}
                config={config}
                gameState={gameState}
                onHexClick={handleHexClick}
              />
            ))}

            <SolarSystemOverview />

            {orbitalProject && (
              <OrbitalStation
                filledSeats={orbitalStationSeats}
                totalSeats={orbitalProject.seats.length}
                isCompleted={orbitalProject.isCompleted}
                name={orbitalProject.name}
              />
            )}

            <AsteroidImpact />

            <GpuWarmup onReady={onGpuReady} />
            <PerformanceProbe />

            <PanControls />
            <FreeCamera />
            <FreeCameraFrustum fov={cameraConfig.fov} />
          </Suspense>
        </MarsRotationProvider>
      </Canvas>
    </div>
  );
}
