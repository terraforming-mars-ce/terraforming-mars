import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import * as THREE from "three";
import DustEffect from "./effects/DustEffect";
import { useTextures } from "../../../hooks/useTextures";
import { addSphereProjectionWithSoftEdges } from "./GreeneryRenderer";
import { easeOutCubic } from "./boardConstants";

interface ReservedAreaTileProps {
  isNewlyPlaced?: boolean;
  ownerColor?: string;
  surfaceNormal?: THREE.Vector3;
  worldPosition?: THREE.Vector3;
}

const HEX_RADIUS = 0.166;
const FENCE_INSET = 0.145;
const POST_HEIGHT = 0.05;
const POST_RADIUS = 0.003;
const RAIL_RADIUS = 0.0015;

function createRadialDiscGeometry(
  radius: number,
  radialSegments: number,
  angularSegments: number,
): THREE.BufferGeometry {
  const geometry = new THREE.BufferGeometry();
  const vertices: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  vertices.push(0, 0, 0);
  uvs.push(0.5, 0.5);

  for (let r = 1; r <= radialSegments; r++) {
    const ringRadius = (r / radialSegments) * radius;
    for (let a = 0; a < angularSegments; a++) {
      const angle = (a / angularSegments) * Math.PI * 2;
      const x = Math.cos(angle) * ringRadius;
      const y = Math.sin(angle) * ringRadius;
      vertices.push(x, y, 0);
      uvs.push(0.5 + (x / radius) * 0.5, 0.5 + (y / radius) * 0.5);
    }
  }

  for (let a = 0; a < angularSegments; a++) {
    const next = (a + 1) % angularSegments;
    indices.push(0, 1 + a, 1 + next);
  }

  for (let r = 1; r < radialSegments; r++) {
    const innerStart = 1 + (r - 1) * angularSegments;
    const outerStart = 1 + r * angularSegments;
    for (let a = 0; a < angularSegments; a++) {
      const nextA = (a + 1) % angularSegments;
      indices.push(innerStart + a, outerStart + a, outerStart + nextA);
      indices.push(innerStart + a, outerStart + nextA, innerStart + nextA);
    }
  }

  geometry.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();
  return geometry;
}

function getHexVertex(index: number, radius: number): THREE.Vector3 {
  const angle = (Math.PI / 3) * index + Math.PI / 6;
  return new THREE.Vector3(Math.cos(angle) * radius, Math.sin(angle) * radius, 0);
}

export default function ReservedAreaTile({
  isNewlyPlaced = false,
  ownerColor,
  surfaceNormal,
  worldPosition,
}: ReservedAreaTileProps) {
  const groupRef = useRef<THREE.Group>(null);
  const emergenceStartRef = useRef<number | null>(null);
  const isEmergingRef = useRef(isNewlyPlaced);
  const [showDust, setShowDust] = useState(isNewlyPlaced);

  useEffect(() => {
    if (isNewlyPlaced) {
      isEmergingRef.current = true;
      setShowDust(true);
      emergenceStartRef.current = null;
    }
  }, [isNewlyPlaced]);

  const groundGeometry = useMemo(() => {
    const geo = createRadialDiscGeometry(HEX_RADIUS * 1.8, 10, 36);
    geo.rotateZ(Math.PI / 2);
    return geo;
  }, []);

  const { noiseMid: noiseTexture, noiseHigh: noiseHighTexture } = useTextures();

  const groundMaterial = useMemo(() => {
    const mat = new THREE.MeshStandardMaterial({
      color: new THREE.Color(0.4, 0.28, 0.18),
      roughness: 0.9,
      metalness: 0.0,
      side: THREE.DoubleSide,
      transparent: true,
      alphaTest: 0.01,
      depthWrite: false,
    });
    addSphereProjectionWithSoftEdges(mat, 0.003, noiseTexture, noiseHighTexture, HEX_RADIUS);
    return mat;
  }, [noiseTexture, noiseHighTexture]);

  const fenceColor = useMemo(() => new THREE.Color(ownerColor || "#888888"), [ownerColor]);

  const fenceMaterial = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        color: fenceColor,
        roughness: 0.4,
        metalness: 0.3,
      }),
    [fenceColor],
  );

  const postGeometry = useMemo(
    () => new THREE.CylinderGeometry(POST_RADIUS, POST_RADIUS, POST_HEIGHT, 6),
    [],
  );

  const railGeometry = useMemo(
    () => new THREE.CylinderGeometry(RAIL_RADIUS, RAIL_RADIUS, 1, 4),
    [],
  );

  const hexVertices = useMemo(() => {
    return Array.from({ length: 6 }, (_, i) => getHexVertex(i, FENCE_INSET));
  }, []);

  const postProgressRef = useRef(isNewlyPlaced ? 0 : 1);
  const railProgressRef = useRef(isNewlyPlaced ? 0 : 1);

  useFrame((state) => {
    if (!isEmergingRef.current) return;
    if (!groupRef.current) return;

    if (emergenceStartRef.current === null) {
      emergenceStartRef.current = state.clock.elapsedTime * 1000;
    }

    const elapsed = state.clock.elapsedTime * 1000 - emergenceStartRef.current;

    const postDuration = 600;
    const postT = Math.min(elapsed / postDuration, 1);
    postProgressRef.current = easeOutCubic(postT);

    const railDelay = 400;
    const railDuration = 400;
    const railElapsed = Math.max(0, elapsed - railDelay);
    const railT = Math.min(railElapsed / railDuration, 1);
    railProgressRef.current = easeOutCubic(railT);

    if (elapsed >= postDuration + railDuration) {
      isEmergingRef.current = false;
      postProgressRef.current = 1;
      railProgressRef.current = 1;
    }
  });

  const groundQuaternion = useMemo(() => new THREE.Quaternion(), []);

  return (
    <>
      <group ref={groupRef}>
        <mesh
          geometry={groundGeometry}
          material={groundMaterial}
          quaternion={groundQuaternion}
          renderOrder={12}
        />

        {hexVertices.map((vertex, i) => {
          const postHeight = POST_HEIGHT * postProgressRef.current;
          return (
            <group key={`post-${i}`}>
              <mesh
                geometry={postGeometry}
                material={fenceMaterial}
                position={[vertex.x, vertex.y, postHeight / 2]}
                scale={[1, postProgressRef.current, 1]}
                rotation={[Math.PI / 2, 0, 0]}
                renderOrder={14}
              />
            </group>
          );
        })}

        {hexVertices.map((vertex, i) => {
          const next = hexVertices[(i + 1) % 6];
          const midX = (vertex.x + next.x) / 2;
          const midY = (vertex.y + next.y) / 2;
          const dx = next.x - vertex.x;
          const dy = next.y - vertex.y;
          const length = Math.sqrt(dx * dx + dy * dy);
          const angle = Math.atan2(dy, dx);

          return [0.25, 0.5, 0.75].map((heightFraction, ri) => {
            const railZ = POST_HEIGHT * heightFraction * railProgressRef.current;
            return (
              <mesh
                key={`rail-${i}-${ri}`}
                geometry={railGeometry}
                material={fenceMaterial}
                position={[midX, midY, railZ]}
                rotation={[0, 0, angle - Math.PI / 2]}
                scale={[1, length, railProgressRef.current]}
                renderOrder={14}
              />
            );
          });
        })}
      </group>

      {showDust && surfaceNormal && worldPosition && (
        <DustEffect
          position={worldPosition}
          normal={surfaceNormal}
          duration={2000}
          particleColor={new THREE.Color(0.5, 0.35, 0.2)}
          onComplete={() => setShowDust(false)}
        />
      )}
    </>
  );
}
