import { useRef, useEffect, useMemo, useState, useCallback } from "react";
import { useFrame, useThree, ThreeEvent } from "@react-three/fiber";
import * as THREE from "three";
import { Html } from "@react-three/drei";
import {
  ORBITAL_STATION_ORBIT_RADIUS,
  ORBITAL_STATION_ORBIT_SPEED,
  ORBITAL_STATION_TILT,
  easeOutCubic,
} from "./boardConstants";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";

interface OrbitalStationProps {
  filledSeats: number;
  totalSeats: number;
  isCompleted: boolean;
  name: string;
}

const MODULE_COUNT = 6;
const ENTRANCE_SPEED = 1.5;
const HIT_SPHERE_RADIUS = 0.12;

function OrbitalStationTooltip({
  name,
  filledSeats,
  totalSeats,
  isCompleted,
}: {
  name: string;
  filledSeats: number;
  totalSeats: number;
  isCompleted: boolean;
}) {
  return (
    <Html center style={{ pointerEvents: "none" }}>
      <div
        className="w-max max-w-52 pointer-events-none animate-[fadeIn_150ms_ease-in]"
        style={{ transform: "translate(20px, -50%)" }}
      >
        <div
          className="relative bg-[rgba(10,10,15,0.98)] border border-[rgba(60,60,70,0.7)] text-white/90 text-[11px] leading-tight px-3 py-2 shadow-[0_2px_8px_rgba(0,0,0,0.5)]"
          style={{
            clipPath:
              "polygon(0 0, calc(100% - 14px) 0, 100% 14px, 100% 100%, 14px 100%, 0 calc(100% - 14px))",
          }}
        >
          <div className="font-orbitron font-bold text-xs text-white mb-1">{name}</div>
          <div className="text-white/60 text-[10px]">
            {isCompleted ? (
              <span className="text-emerald-400">Completed</span>
            ) : (
              <span>
                Seats: {filledSeats}/{totalSeats}
              </span>
            )}
          </div>
          <svg
            className="absolute top-0 right-0 w-[14px] h-[14px] pointer-events-none"
            viewBox="0 0 14 14"
          >
            <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
          </svg>
          <svg
            className="absolute bottom-0 left-0 w-[14px] h-[14px] pointer-events-none"
            viewBox="0 0 14 14"
          >
            <line x1="0" y1="0" x2="14" y2="14" stroke="rgba(60,60,70,0.7)" strokeWidth="1.5" />
          </svg>
        </div>
      </div>
    </Html>
  );
}

export default function OrbitalStation({
  filledSeats,
  totalSeats,
  isCompleted,
  name,
}: OrbitalStationProps) {
  const groupRef = useRef<THREE.Group>(null);
  const prevSeatsRef = useRef(filledSeats);
  const moduleScalesRef = useRef<number[]>(
    Array.from({ length: MODULE_COUNT }, (_, i) => (i < filledSeats ? 1 : 0)),
  );
  const animatingRef = useRef<Set<number>>(new Set());
  const [hovered, setHovered] = useState(false);
  const { gl } = useThree();
  const { activePlanet, setActivePlanet } = usePlanetFocus();

  useEffect(() => {
    const prev = prevSeatsRef.current;
    if (filledSeats > prev) {
      for (let i = prev; i < filledSeats; i++) {
        animatingRef.current.add(i);
      }
    }
    prevSeatsRef.current = filledSeats;
  }, [filledSeats]);

  const handlePointerEnter = useCallback(
    (e: ThreeEvent<PointerEvent>) => {
      e.stopPropagation();
      setHovered(true);
      gl.domElement.style.cursor = "pointer";
    },
    [gl],
  );

  const handlePointerLeave = useCallback(() => {
    setHovered(false);
    gl.domElement.style.cursor = "grab";
  }, [gl]);

  const handleClick = useCallback(
    (e: ThreeEvent<MouseEvent>) => {
      e.stopPropagation();
      if (activePlanet !== "orbital-station") {
        setActivePlanet("orbital-station");
      }
    },
    [activePlanet, setActivePlanet],
  );

  const hubGeometry = useMemo(() => new THREE.CylinderGeometry(0.02, 0.02, 0.05, 8), []);
  const panelGeometry = useMemo(() => new THREE.BoxGeometry(0.07, 0.002, 0.025), []);
  const strutGeometry = useMemo(() => new THREE.CylinderGeometry(0.002, 0.002, 0.05, 4), []);
  const ringGeometry = useMemo(() => new THREE.TorusGeometry(0.04, 0.006, 8, 24), []);
  const dockGeometry = useMemo(() => new THREE.CylinderGeometry(0.008, 0.008, 0.03, 6), []);
  const antennaGeometry = useMemo(() => new THREE.CylinderGeometry(0.002, 0.002, 0.06, 4), []);
  const dishGeometry = useMemo(
    () => new THREE.SphereGeometry(0.012, 8, 6, 0, Math.PI * 2, 0, Math.PI / 2),
    [],
  );
  const hitSphereGeometry = useMemo(() => new THREE.SphereGeometry(HIT_SPHERE_RADIUS, 8, 8), []);

  const hubMaterial = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        color: "#4A90D9",
        emissive: "#1a3a6a",
        emissiveIntensity: 0.3,
        metalness: 0.8,
        roughness: 0.3,
      }),
    [],
  );

  const panelMaterial = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        color: "#1a2744",
        metalness: 0.9,
        roughness: 0.2,
      }),
    [],
  );

  const ringMaterial = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        color: "#6ab0e8",
        metalness: 0.6,
        roughness: 0.4,
      }),
    [],
  );

  const ghostMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        color: "#4A90D9",
        wireframe: true,
        transparent: true,
        opacity: 0.15,
      }),
    [],
  );

  const glowMaterial = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        color: "#4A90D9",
        emissive: "#4A90D9",
        emissiveIntensity: 1.2,
        metalness: 0.8,
        roughness: 0.2,
        transparent: true,
        opacity: 0.9,
      }),
    [],
  );

  const hitSphereMaterial = useMemo(
    () =>
      new THREE.MeshBasicMaterial({
        visible: false,
      }),
    [],
  );

  const lookTarget = useMemo(() => new THREE.Vector3(), []);

  useFrame((state, delta) => {
    if (!groupRef.current) {
      return;
    }

    const t = state.clock.elapsedTime;
    const angle = t * ORBITAL_STATION_ORBIT_SPEED;
    const r = ORBITAL_STATION_ORBIT_RADIUS;
    const tiltY = Math.sin(ORBITAL_STATION_TILT) * r * 0.3;

    groupRef.current.position.set(
      Math.cos(angle) * r,
      Math.sin(angle) * tiltY,
      Math.sin(angle) * r,
    );

    const nextAngle = angle + 0.1;
    lookTarget.set(Math.cos(nextAngle) * r, Math.sin(nextAngle) * tiltY, Math.sin(nextAngle) * r);
    groupRef.current.lookAt(lookTarget);

    for (const idx of animatingRef.current) {
      moduleScalesRef.current[idx] = Math.min(
        moduleScalesRef.current[idx] + delta * ENTRANCE_SPEED,
        1,
      );
      if (moduleScalesRef.current[idx] >= 1) {
        animatingRef.current.delete(idx);
      }
    }

    const children = groupRef.current.children;
    // Skip first child (hub) and last child (hit sphere)
    for (let i = 1; i < children.length - 1; i++) {
      const moduleIdx = i - 1;
      if (moduleIdx < MODULE_COUNT) {
        const raw = moduleScalesRef.current[moduleIdx];
        const scale = moduleIdx < filledSeats ? easeOutCubic(raw) : 0;
        children[i].scale.setScalar(scale);
        children[i].visible = scale > 0;
      }
    }
  });

  const completed = filledSeats >= MODULE_COUNT;

  return (
    <group ref={groupRef}>
      {/* Core hub - always visible */}
      <mesh
        geometry={hubGeometry}
        material={filledSeats > 0 ? (completed ? glowMaterial : hubMaterial) : ghostMaterial}
        rotation={[Math.PI / 2, 0, 0]}
      />

      {/* Module 0 (seat 1): Solar panel arm A */}
      <group>
        <mesh geometry={panelGeometry} material={panelMaterial} position={[0.055, 0, 0]} />
        <mesh
          geometry={strutGeometry}
          material={hubMaterial}
          position={[0.03, 0, 0]}
          rotation={[0, 0, Math.PI / 2]}
        />
      </group>

      {/* Module 1 (seat 2): Solar panel arm B */}
      <group>
        <mesh geometry={panelGeometry} material={panelMaterial} position={[-0.055, 0, 0]} />
        <mesh
          geometry={strutGeometry}
          material={hubMaterial}
          position={[-0.03, 0, 0]}
          rotation={[0, 0, Math.PI / 2]}
        />
      </group>

      {/* Module 2 (seat 3): Habitat ring (partial) */}
      <group>
        <mesh geometry={ringGeometry} rotation={[Math.PI / 2, 0, 0]}>
          <meshStandardMaterial
            color="#6ab0e8"
            metalness={0.6}
            roughness={0.4}
            transparent
            opacity={0.5}
          />
        </mesh>
      </group>

      {/* Module 3 (seat 4): Habitat ring full + docking port */}
      <group>
        <mesh geometry={ringGeometry} material={ringMaterial} rotation={[Math.PI / 2, 0, 0]} />
        <mesh
          geometry={dockGeometry}
          material={hubMaterial}
          position={[0, 0, 0.04]}
          rotation={[Math.PI / 2, 0, 0]}
        />
      </group>

      {/* Module 4 (seat 5): Antenna + comm dish */}
      <group>
        <mesh geometry={antennaGeometry} material={hubMaterial} position={[0, 0.05, 0]} />
        <mesh
          geometry={dishGeometry}
          material={panelMaterial}
          position={[0, 0.08, 0]}
          rotation={[Math.PI, 0, 0]}
        />
        <mesh
          geometry={dockGeometry}
          material={hubMaterial}
          position={[0, 0, -0.04]}
          rotation={[Math.PI / 2, 0, 0]}
        />
      </group>

      {/* Module 5 (seat 6): Completion glow ring */}
      <group>
        <mesh
          geometry={ringGeometry}
          material={glowMaterial}
          rotation={[0, 0, Math.PI / 2]}
          scale={1.3}
        />
      </group>

      {/* Invisible hit sphere for hover detection */}
      <mesh
        geometry={hitSphereGeometry}
        material={hitSphereMaterial}
        onPointerEnter={handlePointerEnter}
        onPointerLeave={handlePointerLeave}
        onClick={handleClick}
      />

      {hovered && (
        <OrbitalStationTooltip
          name={name}
          filledSeats={filledSeats}
          totalSeats={totalSeats}
          isCompleted={isCompleted}
        />
      )}
    </group>
  );
}
