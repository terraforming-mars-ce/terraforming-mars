import { useRef, useMemo, useState, useCallback, useEffect } from "react";
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
import { useModels } from "../../../hooks/useModels";
import { getMarsOrbitalPosition } from "./solarSystemConfig";

interface OrbitalStationProps {
  filledSeats: number;
  totalSeats: number;
  isCompleted: boolean;
  name: string;
}

const HIT_SPHERE_RADIUS = 0.12;
const SPAWN_SPEED = 1.2;

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
  const modelRef = useRef<THREE.Group>(null);
  const spawnProgress = useRef(isCompleted ? 1 : 0);
  const wasCompleted = useRef(isCompleted);
  const [hovered, setHovered] = useState(false);
  const { gl } = useThree();
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const { satelliteScene } = useModels();

  const satelliteClone = useMemo(() => {
    const clone = satelliteScene.clone(true);
    const toRemove: THREE.Object3D[] = [];
    clone.traverse((child) => {
      if (child.name === "Text" || child.name === "Earth") {
        toRemove.push(child);
      }
    });
    toRemove.forEach((child) => child.removeFromParent());
    const sat = clone.getObjectByName("Satlite");
    if (sat) {
      sat.position.set(0, 0, 0);
    }
    return clone;
  }, [satelliteScene]);

  useEffect(() => {
    if (isCompleted && !wasCompleted.current) {
      spawnProgress.current = 0;
    }
    wasCompleted.current = isCompleted;
  }, [isCompleted]);

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

  const hitSphereGeometry = useMemo(() => new THREE.SphereGeometry(HIT_SPHERE_RADIUS, 8, 8), []);
  const hitSphereMaterial = useMemo(() => new THREE.MeshBasicMaterial({ visible: false }), []);

  const lookTarget = useMemo(() => new THREE.Vector3(), []);

  useFrame((state, delta) => {
    if (!groupRef.current) {
      return;
    }

    const t = state.clock.elapsedTime;
    const angle = t * ORBITAL_STATION_ORBIT_SPEED;
    const r = ORBITAL_STATION_ORBIT_RADIUS;
    const tiltY = Math.sin(ORBITAL_STATION_TILT) * r * 0.3;

    const marsPos = getMarsOrbitalPosition(t);
    groupRef.current.position.set(
      marsPos[0] + Math.cos(angle) * r,
      marsPos[1] + Math.sin(angle) * tiltY,
      marsPos[2] + Math.sin(angle) * r,
    );

    const nextAngle = angle + 0.1;
    lookTarget.set(
      marsPos[0] + Math.cos(nextAngle) * r,
      marsPos[1] + Math.sin(nextAngle) * tiltY,
      marsPos[2] + Math.sin(nextAngle) * r,
    );
    groupRef.current.lookAt(lookTarget);

    if (isCompleted && spawnProgress.current < 1) {
      spawnProgress.current = Math.min(spawnProgress.current + delta * SPAWN_SPEED, 1);
    }

    if (modelRef.current) {
      const scale = easeOutCubic(spawnProgress.current) * 0.006;
      modelRef.current.scale.setScalar(scale);
      modelRef.current.visible = isCompleted;
      modelRef.current.rotation.z = t * 0.3;
    }
  });

  return (
    <group ref={groupRef} visible={isCompleted}>
      <primitive ref={modelRef} object={satelliteClone} />

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
