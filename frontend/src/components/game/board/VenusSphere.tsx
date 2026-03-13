import { useMemo, useRef, useState, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import { Html } from "@react-three/drei";
import * as THREE from "three";
import { useTextures } from "../../../hooks/useTextures";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import { VENUS_RADIUS, VENUS_POSITION } from "./boardConstants";
import VenusTileGrid from "./VenusTileGrid";
import { GameDto } from "../../../types/generated/api-types";

const LABEL_BOX_W = 100;
const LABEL_BOX_H = 42;
const CORNER = 8;
const BORDER_COLOR = "rgba(60, 60, 70, 0.7)";

function VenusLabel({ visible }: { visible: boolean }) {
  const [mounted, setMounted] = useState(false);
  useEffect(() => {
    requestAnimationFrame(() => setMounted(true));
  }, []);

  return (
    <Html
      position={[-VENUS_RADIUS * 0.7, VENUS_RADIUS * 0.7, VENUS_RADIUS * 0.5]}
      style={{ pointerEvents: "none" }}
    >
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "flex-end",
          pointerEvents: "none",
          transform: "translate(-100%, -100%)",
          opacity: mounted && visible ? 1 : 0,
          transition: "opacity 0.3s ease-out",
        }}
      >
        <svg
          width={LABEL_BOX_W}
          height={LABEL_BOX_H}
          style={{ display: "block", overflow: "visible" }}
        >
          <polygon
            points={`0,0 ${LABEL_BOX_W - CORNER},0 ${LABEL_BOX_W},${CORNER} ${LABEL_BOX_W},${LABEL_BOX_H} ${CORNER},${LABEL_BOX_H} 0,${LABEL_BOX_H - CORNER}`}
            fill="rgba(10, 10, 15, 0.95)"
            stroke={BORDER_COLOR}
            strokeWidth="1"
          />
          <text
            x={LABEL_BOX_W / 2}
            y={16}
            textAnchor="middle"
            dominantBaseline="central"
            fill="white"
            fontFamily="Orbitron, sans-serif"
            fontSize="11"
            letterSpacing="0.15em"
          >
            VENUS
          </text>
          <text
            x={LABEL_BOX_W / 2}
            y={32}
            textAnchor="middle"
            dominantBaseline="central"
            fill="rgba(160, 160, 170, 0.8)"
            fontFamily="Orbitron, sans-serif"
            fontSize="8"
          >
            Click to travel
          </text>
        </svg>
        <svg width="30" height="30" style={{ display: "block" }}>
          <line x1="0" y1="0" x2="30" y2="30" stroke={BORDER_COLOR} strokeWidth="1" />
        </svg>
      </div>
    </Html>
  );
}

interface VenusSphereProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}

export default function VenusSphere({ gameState, onHexClick }: VenusSphereProps) {
  const [hovered, setHovered] = useState(false);
  const [labelMounted, setLabelMounted] = useState(false);
  const { activePlanet, setActivePlanet } = usePlanetFocus();
  const { venus: venusTexture } = useTextures();
  const { gl } = useThree();
  const tileOpacity = useRef(0);

  useFrame(() => {
    const target = activePlanet === "venus" ? 1 : 0;
    tileOpacity.current = THREE.MathUtils.lerp(tileOpacity.current, target, 0.05);
  });

  const showLabel = hovered && activePlanet === "mars";

  useEffect(() => {
    if (showLabel) {
      setLabelMounted(true);
      return;
    }
    const timeout = setTimeout(() => setLabelMounted(false), 300);
    return () => clearTimeout(timeout);
  }, [showLabel]);

  const geometry = useMemo(() => new THREE.SphereGeometry(VENUS_RADIUS, 64, 32), []);

  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        map: venusTexture,
        roughness: 0.8,
        metalness: 0.05,
        fog: false,
      }),
    [venusTexture],
  );

  return (
    <group position={[VENUS_POSITION[0], VENUS_POSITION[1], VENUS_POSITION[2]]}>
      <mesh
        geometry={geometry}
        material={material}
        onPointerEnter={(e) => {
          if (activePlanet !== "mars") return;
          if (e.intersections[0]?.object !== e.object) return;
          setHovered(true);
          gl.domElement.style.cursor = "pointer";
        }}
        onPointerLeave={() => {
          setHovered(false);
          gl.domElement.style.cursor = "grab";
        }}
        onClick={(e) => {
          e.stopPropagation();
          if (activePlanet !== "mars") return;
          if (e.intersections[0]?.object !== e.object) return;
          gl.domElement.style.cursor = "grab";
          setActivePlanet("venus");
        }}
      />
      <VenusTileGrid gameState={gameState} onHexClick={onHexClick} tileOpacity={tileOpacity} />
      {labelMounted && <VenusLabel visible={showLabel} />}
    </group>
  );
}
