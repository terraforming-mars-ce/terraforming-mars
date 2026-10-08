import { Profiler, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { coldStartTrace } from "@/services/performanceStore.ts";
import { useFrame, useThree } from "@react-three/fiber";
import { Html } from "@react-three/drei";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import {
  PLANET_CONFIGS,
  MARS_ORBIT_RADIUS,
  getMarsOrbitalPosition,
  getMarsOrbitalAngle,
  getPlanetOrbitalPosition,
  getPlanetOrbitalAngle,
  type PlanetConfig,
} from "./solarSystemConfig";
import { isDragClick } from "../controls/PanControls";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

const ORBIT_SEGMENTS = 1024;
const ORBIT_COLOR = new THREE.Color(0.25, 0.35, 0.5);
const MARS_RING_COLOR = new THREE.Color(0.8, 0.3, 0.2);
const TRAIL_LENGTH_RAD = Math.PI * 0.8;
const BASE_BRIGHTNESS = 0.08;
const DEG_TO_RAD = Math.PI / 180;
const MARS_LABEL_COLOR = "#ff6b4a";

function OrbitRing({
  radius,
  color,
  planetAngleRef,
}: {
  radius: number;
  color: THREE.Color;
  planetAngleRef: React.RefObject<number>;
}) {
  const { geometry, material } = useMemo(() => {
    const points: THREE.Vector3[] = [];
    const colors: number[] = [];
    for (let i = 0; i <= ORBIT_SEGMENTS; i++) {
      const angle = (i / ORBIT_SEGMENTS) * Math.PI * 2;
      points.push(new THREE.Vector3(Math.cos(angle) * radius, 0, Math.sin(angle) * radius));
      colors.push(color.r * BASE_BRIGHTNESS, color.g * BASE_BRIGHTNESS, color.b * BASE_BRIGHTNESS);
    }
    const geo = new THREE.BufferGeometry().setFromPoints(points);
    geo.setAttribute("color", new THREE.Float32BufferAttribute(colors, 3));
    const mat = new THREE.LineBasicMaterial({
      vertexColors: true,
      transparent: true,
      fog: false,
    });
    return { geometry: geo, material: mat };
  }, [radius, color]);

  useFrame(() => {
    const colorAttr = geometry.getAttribute("color") as THREE.BufferAttribute;
    const planetRad = planetAngleRef.current * DEG_TO_RAD;

    for (let i = 0; i <= ORBIT_SEGMENTS; i++) {
      const vertexAngle = (i / ORBIT_SEGMENTS) * Math.PI * 2;
      let delta = planetRad - vertexAngle;
      delta = ((delta % (Math.PI * 2)) + Math.PI * 2) % (Math.PI * 2);

      let brightness = BASE_BRIGHTNESS;
      if (delta > 0.005 && delta < TRAIL_LENGTH_RAD) {
        const t = 1 - (delta - 0.005) / (TRAIL_LENGTH_RAD - 0.005);
        brightness = BASE_BRIGHTNESS + (1.0 - BASE_BRIGHTNESS) * t * t;
      }

      colorAttr.setXYZ(
        i,
        color.r * brightness + (1 - color.r) * Math.max(0, brightness - BASE_BRIGHTNESS),
        color.g * brightness + (1 - color.g) * Math.max(0, brightness - BASE_BRIGHTNESS),
        color.b * brightness + (1 - color.b) * Math.max(0, brightness - BASE_BRIGHTNESS),
      );
    }
    colorAttr.needsUpdate = true;
  });

  const lineObj = useMemo(() => new THREE.Line(geometry, material), [geometry, material]);
  return <primitive object={lineObj} />;
}

interface MarkerProps {
  selected: boolean;
  compact: boolean;
  onSelect: () => void;
}

function MarkerLabel({
  name,
  color,
  hovered,
  selected,
  compact,
  onHoverChange,
  onSelect,
}: {
  name: string;
  color?: string;
  hovered: boolean;
  selected: boolean;
  compact: boolean;
  onHoverChange: (hovered: boolean) => void;
  onSelect: () => void;
}) {
  const { gl } = useThree();
  const emphasized = compact ? selected : hovered;
  let hint = "Click to travel";
  let frameClass = "";
  if (compact) {
    hint = "Tap again to travel";
    frameClass = selected
      ? "px-3 py-2 bg-black/60 outline outline-1 outline-white/50"
      : "px-3 py-2";
  }

  return (
    <Html center>
      <div
        onPointerEnter={() => {
          onHoverChange(true);
          gl.domElement.style.cursor = "pointer";
        }}
        onPointerLeave={() => {
          onHoverChange(false);
          gl.domElement.style.cursor = "grab";
        }}
        onClick={() => onSelect()}
        className={frameClass}
        style={{
          transform: compact ? "translateY(-24px)" : "translateY(-14px)",
          opacity: emphasized || compact ? 1 : 0.7,
          transition: "opacity 0.2s",
          cursor: "pointer",
          whiteSpace: "nowrap",
          textShadow: "0 0 8px rgba(0,0,0,0.9)",
        }}
      >
        <div
          className={`font-orbitron font-bold tracking-wider text-center ${compact ? "text-[13px]" : "text-[11px]"} ${color ? "" : "text-white"}`}
          style={{ color }}
        >
          {name}
        </div>
        <div
          className={`text-center mt-0.5 ${compact ? "text-[11px] text-white/80" : "text-[9px] text-white/50"}`}
          style={{ opacity: emphasized ? 1 : 0, transition: "opacity 0.15s" }}
        >
          {hint}
        </div>
      </div>
    </Html>
  );
}

function PlanetMarker({
  config,
  selected,
  compact,
  onSelect,
}: MarkerProps & { config: PlanetConfig }) {
  const textures = useTextures();
  const { gl } = useThree();
  const [hovered, setHovered] = useState(false);
  const groupRef = useRef<THREE.Group>(null);

  const texture = (textures as unknown as Record<string, THREE.Texture>)[config.textureKey + "Lod"];

  const hitSize = 30;

  const geometry = useMemo(() => new THREE.SphereGeometry(config.radius, 32, 16), [config.radius]);
  const hitGeometry = useMemo(() => new THREE.BoxGeometry(hitSize, hitSize, hitSize), []);

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: texture,
        roughness: 0.8,
        metalness: 0,
        fog: false,
      }),
    [texture],
  );

  useFrame((state) => {
    if (groupRef.current) {
      const pos = getPlanetOrbitalPosition(config, state.clock.elapsedTime);
      groupRef.current.position.set(pos[0], pos[1], pos[2]);
    }
  });

  return (
    <group ref={groupRef}>
      <mesh geometry={geometry} material={material} />
      <mesh
        geometry={hitGeometry}
        visible={false}
        onPointerEnter={() => {
          setHovered(true);
          gl.domElement.style.cursor = "pointer";
        }}
        onPointerLeave={() => {
          setHovered(false);
          gl.domElement.style.cursor = "grab";
        }}
        onClick={(e) => {
          e.stopPropagation();
          if (isDragClick()) {
            return;
          }
          gl.domElement.style.cursor = "grab";
          onSelect();
        }}
      />
      <MarkerLabel
        name={config.name}
        hovered={hovered}
        selected={selected}
        compact={compact}
        onHoverChange={setHovered}
        onSelect={onSelect}
      />
    </group>
  );
}

function MarsMarker({ selected, compact, onSelect }: MarkerProps) {
  const { marsLod: texture } = useTextures();
  const { gl } = useThree();
  const [hovered, setHovered] = useState(false);
  const groupRef = useRef<THREE.Group>(null);

  const marsRadius = 2.02;
  const hitSize = 30;

  const geometry = useMemo(() => new THREE.SphereGeometry(marsRadius, 32, 16), []);
  const hitGeometry = useMemo(() => new THREE.BoxGeometry(hitSize, hitSize, hitSize), []);
  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: texture,
        roughness: 0.8,
        metalness: 0,
        fog: false,
      }),
    [texture],
  );

  useFrame((state) => {
    if (groupRef.current) {
      const pos = getMarsOrbitalPosition(state.clock.elapsedTime);
      groupRef.current.position.set(pos[0], pos[1], pos[2]);
    }
  });

  return (
    <group ref={groupRef}>
      <mesh geometry={geometry} material={material} />
      <mesh
        geometry={hitGeometry}
        visible={false}
        onPointerEnter={() => {
          setHovered(true);
          gl.domElement.style.cursor = "pointer";
        }}
        onPointerLeave={() => {
          setHovered(false);
          gl.domElement.style.cursor = "grab";
        }}
        onClick={(e) => {
          e.stopPropagation();
          if (isDragClick()) {
            return;
          }
          gl.domElement.style.cursor = "grab";
          onSelect();
        }}
      />
      <MarkerLabel
        name="MARS"
        color={MARS_LABEL_COLOR}
        hovered={hovered}
        selected={selected}
        compact={compact}
        onHoverChange={setHovered}
        onSelect={onSelect}
      />
    </group>
  );
}

export default function SolarSystemOverview() {
  if (coldStartTrace.enabled) {
    return (
      <Profiler id="solar-system" onRender={coldStartTrace.reactRender}>
        <SolarSystemOverviewContent />
      </Profiler>
    );
  }
  return <SolarSystemOverviewContent />;
}

function SolarSystemOverviewContent() {
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const { isCompact } = useLayoutMode();
  const [selectedBody, setSelectedBody] = useState<string | null>(null);

  useEffect(() => {
    setSelectedBody(null);
  }, [activePlanet, isCompact]);

  const handleBodyTap = (bodyId: string) => {
    if (isCompact && selectedBody !== bodyId) {
      setSelectedBody(bodyId);
      return;
    }
    setActivePlanet(bodyId as Parameters<typeof setActivePlanet>[0]);
  };

  useLayoutEffect(() => {
    coldStartTrace.mark("solar-system:commit", { activePlanet });
  }, [activePlanet]);

  const planetAngleRefs = useRef<Record<string, { current: number }>>(
    Object.fromEntries(PLANET_CONFIGS.map((c) => [c.id, { current: c.orbitAngle }])),
  );
  const marsAngleRef = useRef(270);

  const groupRef = useRef<THREE.Group>(null);

  useFrame((state) => {
    const t = state.clock.elapsedTime;
    for (const config of PLANET_CONFIGS) {
      planetAngleRefs.current[config.id].current = getPlanetOrbitalAngle(config, t);
    }
    marsAngleRef.current = getMarsOrbitalAngle(t);
  });

  if (activePlanet !== "solar-system") {
    return null;
  }

  return (
    <group ref={groupRef}>
      <OrbitRing radius={MARS_ORBIT_RADIUS} color={MARS_RING_COLOR} planetAngleRef={marsAngleRef} />
      {PLANET_CONFIGS.map((config) => (
        <OrbitRing
          key={config.id}
          radius={config.orbitRadius}
          color={ORBIT_COLOR}
          planetAngleRef={planetAngleRefs.current[config.id]}
        />
      ))}

      <MarsMarker
        selected={selectedBody === "mars"}
        compact={isCompact}
        onSelect={() => handleBodyTap("mars")}
      />
      {PLANET_CONFIGS.map((config) => (
        <PlanetMarker
          key={config.id}
          config={config}
          selected={selectedBody === config.id}
          compact={isCompact}
          onSelect={() => handleBodyTap(config.id)}
        />
      ))}
    </group>
  );
}
