import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame } from "@react-three/fiber";
import { Text } from "@react-three/drei";
import * as THREE from "three";
import { HexTile2D } from "../../../utils/hex-grid-2d";
import { panState } from "../controls/PanControls";
import OceanTile from "./OceanTile";
import BuildingTile from "./BuildingTile";
import VolcanoTile from "./VolcanoTile";
import { useTextures } from "../../../hooks/useTextures";
import {
  sphereProjectionVertex,
  hoverGlowFragment,
  availableGlowFragment,
  endgameHighlightFragment,
  tileBorderVertex,
  tileBorderFragment,
  tileSurfaceVertexSnippet,
  splitSnippet,
} from "./shaders";
import { SPHERE_RADIUS, CHROME_Z_BASE, easeOutCubic } from "./boardConstants";

const HIGHLIGHT_COLOR_GREENERY = new THREE.Vector3(0.13, 0.77, 0.27);
const HIGHLIGHT_COLOR_CITY = new THREE.Vector3(0.58, 0.64, 0.7);
const HIGHLIGHT_COLOR_ADJACENT = new THREE.Vector3(1.0, 0.84, 0.0);
const BONUS_ICON_TINT = new THREE.Color(0.7, 0.7, 0.7);

function createSubdividedHexagonGeometry(radius: number, rings: number): THREE.BufferGeometry {
  const geometry = new THREE.BufferGeometry();
  const vertices: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  vertices.push(0, 0, 0);
  uvs.push(0.5, 0.5);

  for (let ring = 1; ring <= rings; ring++) {
    const ringRadius = (ring / rings) * radius;
    const verticesInRing = 6 * ring;

    for (let i = 0; i < verticesInRing; i++) {
      const edgeIndex = Math.floor(i / ring);
      const posOnEdge = i % ring;

      const angle1 = (edgeIndex * Math.PI) / 3;
      const angle2 = ((edgeIndex + 1) * Math.PI) / 3;

      const t = posOnEdge / ring;
      const x = ringRadius * (Math.cos(angle1) * (1 - t) + Math.cos(angle2) * t);
      const y = ringRadius * (Math.sin(angle1) * (1 - t) + Math.sin(angle2) * t);

      vertices.push(x, y, 0);
      uvs.push(0.5 + (x / radius) * 0.5, 0.5 + (y / radius) * 0.5);
    }
  }

  for (let i = 0; i < 6; i++) {
    const next = (i + 1) % 6;
    indices.push(0, 1 + i, 1 + next);
  }

  let prevRingStart = 1;
  for (let ring = 2; ring <= rings; ring++) {
    const currRingStart = prevRingStart + 6 * (ring - 1);
    const prevRingVerts = 6 * (ring - 1);
    const currRingVerts = 6 * ring;

    let prevIdx = 0;
    let currIdx = 0;

    for (let edge = 0; edge < 6; edge++) {
      for (let i = 0; i < ring; i++) {
        const curr0 = currRingStart + currIdx;
        const curr1 = currRingStart + ((currIdx + 1) % currRingVerts);

        if (i < ring - 1) {
          const prev0 = prevRingStart + prevIdx;
          const prev1 = prevRingStart + ((prevIdx + 1) % prevRingVerts);

          indices.push(prev0, curr0, curr1);
          indices.push(prev0, curr1, prev1);
          prevIdx++;
        } else {
          const prev0 = prevRingStart + (prevIdx % prevRingVerts);
          indices.push(prev0, curr0, curr1);
        }
        currIdx++;
      }
    }

    prevRingStart = currRingStart;
  }

  geometry.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();

  return geometry;
}

function createSubdividedHexRingGeometry(
  innerRadius: number,
  outerRadius: number,
  segmentsPerEdge: number,
): THREE.BufferGeometry {
  const geometry = new THREE.BufferGeometry();
  const vertices: number[] = [];
  const uvs: number[] = [];
  const indices: number[] = [];

  const totalSegments = 6 * segmentsPerEdge;

  for (let ring = 0; ring <= 1; ring++) {
    const radius = ring === 0 ? innerRadius : outerRadius;

    for (let edge = 0; edge < 6; edge++) {
      const angle1 = (edge * Math.PI) / 3;
      const angle2 = ((edge + 1) * Math.PI) / 3;

      const corner1 = { x: Math.cos(angle1) * radius, y: Math.sin(angle1) * radius };
      const corner2 = { x: Math.cos(angle2) * radius, y: Math.sin(angle2) * radius };

      for (let seg = 0; seg < segmentsPerEdge; seg++) {
        const t = seg / segmentsPerEdge;
        const x = corner1.x * (1 - t) + corner2.x * t;
        const y = corner1.y * (1 - t) + corner2.y * t;

        vertices.push(x, y, 0);
        uvs.push(0.5 + (x / outerRadius) * 0.5, 0.5 + (y / outerRadius) * 0.5);
      }
    }
  }

  const innerStart = 0;
  const outerStart = totalSegments;

  for (let i = 0; i < totalSegments; i++) {
    const next = (i + 1) % totalSegments;

    const inner0 = innerStart + i;
    const inner1 = innerStart + next;
    const outer0 = outerStart + i;
    const outer1 = outerStart + next;

    indices.push(inner0, outer0, outer1);
    indices.push(inner0, outer1, inner1);
  }

  geometry.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
  geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
  geometry.setIndex(indices);
  geometry.computeVertexNormals();

  return geometry;
}

interface TileData3D extends HexTile2D {
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
}

function ClampedBillboard({
  children,
  damping = 0.175,
  ...groupProps
}: {
  children: React.ReactNode;
  damping?: number;
} & React.JSX.IntrinsicElements["group"]) {
  const ref = useRef<THREE.Group>(null);
  const billboardQuat = useMemo(() => new THREE.Quaternion(), []);
  const flatQuat = useMemo(() => new THREE.Quaternion(), []);
  useFrame(({ camera }) => {
    const group = ref.current;
    if (!group) return;

    group.traverse((child) => {
      if (child instanceof THREE.Mesh && child.material) {
        const materials = Array.isArray(child.material) ? child.material : [child.material];
        for (const mat of materials) {
          mat.depthTest = false;
          mat.depthWrite = false;
        }
      }
    });

    group.lookAt(camera.position);
    billboardQuat.copy(group.quaternion);

    const angle = flatQuat.angleTo(billboardQuat);
    if (angle > 0.001) {
      const dampedAngle = damping * Math.atan(angle / damping);
      const t = dampedAngle / angle;
      group.quaternion.copy(flatQuat).slerp(billboardQuat, t);
    }
  });

  return (
    <group ref={ref} {...groupProps}>
      {children}
    </group>
  );
}

export type TileHighlightMode = "greenery" | "city" | "adjacent" | null;

interface TileProps {
  tileData: TileData3D;
  tileType: "empty" | "ocean" | "greenery" | "city" | "special" | "volcano";
  ownerId?: string | null;
  reservedById?: string | null;
  displayName?: string;
  onClick: () => void;
  isAvailableForPlacement?: boolean;
  highlightMode?: TileHighlightMode;
  vpAmount?: number;
  vpAnimating?: boolean;
  animateEntrance?: boolean;
  entranceDelay?: number;
  isNewlyPlaced?: boolean;
  isVolcanic?: boolean;
}

export default function Tile({
  tileData,
  tileType,
  ownerId,
  reservedById,
  displayName,
  onClick,
  isAvailableForPlacement = false,
  highlightMode = null,
  vpAmount,
  vpAnimating = false,
  animateEntrance = false,
  entranceDelay = 0,
  isNewlyPlaced = false,
  isVolcanic = false,
}: TileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const vpTextRef = useRef<THREE.Group>(null);
  const [hovered, setHovered] = useState(false);
  const animationStartTimeRef = useRef<number | null>(null);

  const [entranceScale, setEntranceScale] = useState(animateEntrance ? 0 : 1);
  const entranceStartRef = useRef<number | null>(null);
  const entranceDoneRef = useRef(!animateEntrance);

  const { noiseMid: borderNoiseTexture, getResourceIcon } = useTextures();

  useEffect(() => {
    if (animateEntrance && entranceDoneRef.current) {
      setEntranceScale(0);
      entranceStartRef.current = null;
      entranceDoneRef.current = false;
    }
  }, [animateEntrance]);

  const hexGeometry = useMemo(() => {
    const radius = 0.166;
    const geometry = createSubdividedHexagonGeometry(radius, 6);
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  const borderGeometry = useMemo(() => {
    const geometry = createSubdividedHexRingGeometry(0.156, 0.166, 4);
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  const overlayGeometry = useMemo(() => {
    const geometry = createSubdividedHexagonGeometry(0.166, 6);
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  const hoverGlowMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: hoverGlowFragment,
      uniforms: {
        time: { value: 0.0 },
        opacity: { value: 0.0 },
        uSphereRadius: { value: SPHERE_RADIUS },
        uZOffset: { value: CHROME_Z_BASE + 0.003 },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, []);

  const availableGlowMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: availableGlowFragment,
      uniforms: {
        time: { value: 0.0 },
        uSphereRadius: { value: SPHERE_RADIUS },
        uZOffset: { value: CHROME_Z_BASE + 0.005 },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, []);

  const endGameHighlightMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: endgameHighlightFragment,
      uniforms: {
        time: { value: 0.0 },
        highlightColor: { value: new THREE.Vector3(0.13, 0.77, 0.27) },
        opacity: { value: 0.0 },
        uSphereRadius: { value: SPHERE_RADIUS },
        uZOffset: { value: CHROME_Z_BASE + 0.006 },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, []);

  const volcanicTintMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: `
        precision highp float;
        varying vec2 vUv;
        void main() {
          vec2 center = vUv - 0.5;
          float distFromCenter = length(center);
          float gradient = smoothstep(0.2, 0.45, distFromCenter);
          vec3 color = vec3(0.75, 0.12, 0.05);
          float alpha = gradient * 0.4;
          gl_FragColor = vec4(color, alpha);
        }
      `,
      uniforms: {
        uSphereRadius: { value: SPHERE_RADIUS },
        uZOffset: { value: CHROME_Z_BASE + 0.004 },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
    });
  }, []);

  useFrame((state) => {
    if (animateEntrance && !entranceDoneRef.current) {
      if (entranceStartRef.current === null) {
        entranceStartRef.current = state.clock.elapsedTime;
      }
      const elapsed = (state.clock.elapsedTime - entranceStartRef.current) * 1000;
      if (elapsed >= entranceDelay) {
        const animDuration = 400;
        const t = Math.min((elapsed - entranceDelay) / animDuration, 1);
        const eased = easeOutCubic(t);
        setEntranceScale(eased);
        if (t >= 1) {
          entranceDoneRef.current = true;
        }
      }
    }

    if (hoverGlowMaterial.uniforms) {
      hoverGlowMaterial.uniforms.time.value = state.clock.elapsedTime;

      const targetOpacity = hovered ? 0.3 : 0.0;
      hoverGlowMaterial.uniforms.opacity.value = THREE.MathUtils.lerp(
        hoverGlowMaterial.uniforms.opacity.value,
        targetOpacity,
        0.15,
      );
    }

    if (availableGlowMaterial.uniforms) {
      availableGlowMaterial.uniforms.time.value = state.clock.elapsedTime;
    }

    if (endGameHighlightMaterial.uniforms) {
      endGameHighlightMaterial.uniforms.time.value = state.clock.elapsedTime;

      const targetOpacity = highlightMode ? 1.0 : 0.0;
      endGameHighlightMaterial.uniforms.opacity.value = THREE.MathUtils.lerp(
        endGameHighlightMaterial.uniforms.opacity.value,
        targetOpacity,
        0.1,
      );

      if (highlightMode) {
        let color: THREE.Vector3;
        switch (highlightMode) {
          case "greenery":
            color = HIGHLIGHT_COLOR_GREENERY;
            break;
          case "city":
            color = HIGHLIGHT_COLOR_CITY;
            break;
          case "adjacent":
            color = HIGHLIGHT_COLOR_ADJACENT;
            break;
          default:
            color = HIGHLIGHT_COLOR_GREENERY;
        }
        endGameHighlightMaterial.uniforms.highlightColor.value = color;
      }
    }

    if (vpTextRef.current && vpAmount !== undefined && vpAnimating) {
      if (animationStartTimeRef.current === null) {
        animationStartTimeRef.current = state.clock.elapsedTime;
      }

      const elapsed = state.clock.elapsedTime - animationStartTimeRef.current;
      const duration = 2.0;

      if (elapsed < 0.3) {
        const progress = elapsed / 0.3;
        vpTextRef.current.scale.setScalar(progress);
        vpTextRef.current.position.z = 0.02 + progress * 0.05;
      } else if (elapsed < 1.8) {
        vpTextRef.current.scale.setScalar(1);
        vpTextRef.current.position.z = 0.07 + Math.sin(elapsed * 2) * 0.01;
      } else if (elapsed < duration) {
        const progress = 1 - (elapsed - 1.8) / 0.2;
        vpTextRef.current.scale.setScalar(Math.max(0, progress));
      } else {
        vpTextRef.current.scale.setScalar(0);
      }
    } else if (vpTextRef.current && !vpAnimating) {
      animationStartTimeRef.current = null;
      vpTextRef.current.scale.setScalar(vpAmount !== undefined ? 1 : 0);
      vpTextRef.current.position.z = 0.07;
    }
  });

  const surfaceQuaternion = useMemo(() => {
    const up = new THREE.Vector3(0, 0, 1);
    const quaternion = new THREE.Quaternion();
    quaternion.setFromUnitVectors(up, tileData.normal);
    return quaternion;
  }, [tileData.normal]);

  const adjustedPosition = useMemo(() => {
    return tileData.spherePosition.clone().add(tileData.normal.clone().multiplyScalar(0.01));
  }, [tileData.spherePosition, tileData.normal]);

  const tileColor = useMemo(() => {
    if (hovered) return new THREE.Color("#ffff88");

    switch (tileType) {
      case "ocean":
        return new THREE.Color("#1e88e5");
      case "greenery":
        return new THREE.Color("#43a047");
      case "city":
        return new THREE.Color("#ff6f00");
      case "special":
        return new THREE.Color("#8e24aa");
      case "volcano":
        return new THREE.Color("#4a3728");
      default:
        return tileData.isOceanSpace
          ? new THREE.Color("#6d4c41").multiplyScalar(0.8)
          : new THREE.Color("#6d4c41").multiplyScalar(0.8);
    }
  }, [tileType, tileData.isOceanSpace, hovered]);

  const borderColor = useMemo(() => {
    if (hovered) return new THREE.Color("#ffffff");
    return tileColor.clone().multiplyScalar(0.25);
  }, [tileColor, hovered]);

  const hexTileMaterial = useMemo(() => {
    const isGreenery = tileType === "greenery";
    const material = new THREE.MeshStandardMaterial({
      color: tileColor,
      transparent: true,
      opacity:
        isGreenery || tileType === "city" || tileType === "volcano"
          ? 0
          : tileType === "empty"
            ? 0.3
            : 0.7,
      depthWrite: false,
      roughness: 0.7,
      metalness: 0.1,
      side: THREE.DoubleSide,
    });

    const snippet = splitSnippet(tileSurfaceVertexSnippet);
    material.onBeforeCompile = (shader) => {
      shader.uniforms.uSphereRadius = { value: SPHERE_RADIUS };
      shader.uniforms.uZOffset = { value: CHROME_Z_BASE + 0.002 };
      shader.vertexShader =
        snippet.header +
        "\n" +
        shader.vertexShader.replace("#include <begin_vertex>", snippet.body);
    };

    return material;
  }, [tileColor, tileType]);

  const borderMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: tileBorderVertex,
      fragmentShader: tileBorderFragment,
      uniforms: {
        uSphereRadius: { value: SPHERE_RADIUS },
        uZOffset: { value: CHROME_Z_BASE + 0.0025 },
        uColor: { value: new THREE.Color(borderColor.r, borderColor.g, borderColor.b) },
        uOpacity: { value: 0.9 },
        uNoiseTex: { value: borderNoiseTexture },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
    });
  }, [borderColor, borderNoiseTexture]);

  interface BonusIconGroup {
    type: string;
    texture: THREE.Texture;
    count: number;
    isCredits: boolean;
  }

  const bonusIconGroups = useMemo((): BonusIconGroup[] => {
    const entries = Object.entries(tileData.bonuses);
    if (entries.length === 0) return [];

    return entries.map(([key, value]) => ({
      type: key,
      texture: getResourceIcon(key),
      count: value,
      isCredits: key === "credit",
    }));
  }, [tileData.bonuses, getResourceIcon]);

  const calculateIconPositions = (groups: BonusIconGroup[]) => {
    const ICON_GAP = 0.005;
    const GROUP_GAP = 0.01;
    const ICON_SIZE = 0.05;

    const positions: { x: number; group: BonusIconGroup; indexInGroup: number }[] = [];

    let totalWidth = 0;
    groups.forEach((group, groupIndex) => {
      if (groupIndex > 0) totalWidth += GROUP_GAP;
      const iconCount = group.isCredits ? 1 : group.count;
      totalWidth += iconCount * ICON_SIZE + Math.max(0, iconCount - 1) * ICON_GAP;
    });

    let currentX = -totalWidth / 2;
    groups.forEach((group, groupIndex) => {
      if (groupIndex > 0) currentX += GROUP_GAP;

      const iconCount = group.isCredits ? 1 : group.count;
      for (let i = 0; i < iconCount; i++) {
        if (i > 0) currentX += ICON_GAP;
        positions.push({ x: currentX + ICON_SIZE / 2, group, indexInGroup: i });
        currentX += ICON_SIZE;
      }
    });

    return positions;
  };

  return (
    <group
      position={adjustedPosition}
      quaternion={surfaceQuaternion}
      scale={[entranceScale, entranceScale, entranceScale]}
    >
      {/* Main hex tile - hidden for ocean (water mesh handles rendering) */}
      {tileType !== "ocean" && (
        <mesh
          ref={meshRef}
          geometry={hexGeometry}
          material={hexTileMaterial}
          renderOrder={10}
          onPointerEnter={() => {
            if (!panState.isPanning) setHovered(true);
          }}
          onPointerLeave={() => {
            setHovered(false);
          }}
          onClick={(event) => {
            if (panState.isPanning) return;
            event.stopPropagation();
            onClick();
          }}
        />
      )}

      {/* Hex border - hidden for ocean tiles */}
      {tileType !== "ocean" && (
        <mesh geometry={borderGeometry} material={borderMaterial} renderOrder={20} />
      )}

      {/* Ocean rendering (water shader, ocean space indicator) */}
      <OceanTile
        overlayGeometry={overlayGeometry}
        isOceanSpace={tileData.isOceanSpace}
        tileType={tileType}
        onClick={onClick}
        onHoverChange={setHovered}
      />

      {/* Volcanic space indicator - red tint for unoccupied volcanic tiles */}
      {tileType === "empty" && isVolcanic && (
        <mesh geometry={overlayGeometry} material={volcanicTintMaterial} renderOrder={21} />
      )}

      {/* Hover glow effect - hidden for ocean tiles */}
      {tileType !== "ocean" && (
        <mesh geometry={overlayGeometry} material={hoverGlowMaterial} renderOrder={22} />
      )}

      {/* Available placement glow */}
      {isAvailableForPlacement && (
        <mesh geometry={overlayGeometry} material={availableGlowMaterial} renderOrder={23} />
      )}

      {/* End game VP counting highlight - hidden for ocean tiles */}
      {tileType !== "ocean" && (
        <mesh geometry={overlayGeometry} material={endGameHighlightMaterial} renderOrder={24} />
      )}

      {/* Building (city) 3D model */}
      {tileType === "city" && (
        <BuildingTile
          position={[0, 0, 0.03]}
          isNewlyPlaced={isNewlyPlaced}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
      )}

      {/* Volcano 3D tile */}
      {tileType === "volcano" && (
        <VolcanoTile
          isNewlyPlaced={isNewlyPlaced}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
      )}

      {/* Special tile fallback */}
      {tileType === "special" && (
        <Text
          position={[0, 0, 0.01]}
          fontSize={0.08}
          color="white"
          anchorX="center"
          anchorY="middle"
        >
          ⭐
        </Text>
      )}

      {/* Billboard display name for named tiles */}
      {displayName && (
        <ClampedBillboard position={[0, 0, 0.02]} renderOrder={110}>
          <Text
            fontSize={0.045}
            font="/assets/Prototype.ttf"
            color="white"
            outlineWidth={0.004}
            outlineColor="black"
            anchorX="center"
            anchorY="middle"
            textAlign="center"
            maxWidth={0.18}
            renderOrder={110}
          >
            {displayName}
          </Text>
        </ClampedBillboard>
      )}

      {/* Bonus icons for non-greenery, non-volcano tiles */}
      {tileType !== "greenery" && tileType !== "volcano" && bonusIconGroups.length > 0 && (
        <>
          {(() => {
            const yOffset = displayName ? -0.03 : 0;
            const positions = calculateIconPositions(bonusIconGroups);
            return positions.map((pos) => (
              <BonusIcon
                key={`${pos.group.type}-${pos.indexInGroup}`}
                texture={pos.group.texture}
                position={[pos.x, yOffset, 0.01]}
                isCredits={pos.group.isCredits}
                creditAmount={pos.group.isCredits ? pos.group.count : undefined}
              />
            ));
          })()}
        </>
      )}

      {/* Bonus icons for volcano tiles (kept flat on tile) */}
      {tileType === "volcano" && bonusIconGroups.length > 0 && (
        <>
          {(() => {
            const positions = calculateIconPositions(bonusIconGroups);
            return positions.map((pos) => (
              <BonusIcon
                key={`${pos.group.type}-${pos.indexInGroup}`}
                texture={pos.group.texture}
                position={[pos.x, 0, 0.01]}
                isCredits={pos.group.isCredits}
                creditAmount={pos.group.isCredits ? pos.group.count : undefined}
              />
            ));
          })()}
        </>
      )}

      {/* Owner indicator */}
      {ownerId && (
        <mesh position={[0.1, 0.1, 0.01]}>
          <circleGeometry args={[0.02, 16]} />
          <meshBasicMaterial color={`hsl(${(ownerId.charCodeAt(0) * 137.5) % 360}, 70%, 50%)`} />
        </mesh>
      )}

      {/* Reserved tile marker (land claim) */}
      {reservedById && !ownerId && (
        <group position={[0, 0, 0.01]}>
          <mesh position={[0.08, 0.05, 0]}>
            <boxGeometry args={[0.004, 0.06, 0.004]} />
            <meshBasicMaterial color="#333333" />
          </mesh>
          <mesh position={[0.1, 0.07, 0]}>
            <circleGeometry args={[0.025, 3]} />
            <meshBasicMaterial
              color={`hsl(${(reservedById.charCodeAt(0) * 137.5) % 360}, 70%, 50%)`}
            />
          </mesh>
        </group>
      )}

      {/* Floating VP indicator text */}
      {vpAmount !== undefined && (
        <group ref={vpTextRef} position={[0, 0, 0.07]}>
          <Text
            fontSize={0.08}
            color="#FFD700"
            anchorX="center"
            anchorY="middle"
            fontWeight="bold"
            outlineWidth={0.005}
            outlineColor="#000000"
          >
            +{vpAmount}
          </Text>
        </group>
      )}
    </group>
  );
}

interface BonusIconProps {
  texture: THREE.Texture;
  position: [number, number, number];
  isCredits?: boolean;
  creditAmount?: number;
}

function BonusIcon({ texture, position, isCredits, creditAmount }: BonusIconProps) {
  const dimensions = useMemo((): [number, number] => {
    if (!texture.image) return [0.05, 0.05];

    const aspect = texture.image.width / texture.image.height;
    const maxSize = 0.05;

    if (aspect > 1) {
      return [maxSize, maxSize / aspect];
    } else {
      return [maxSize * aspect, maxSize];
    }
  }, [texture]);

  return (
    <group position={position}>
      <mesh>
        <planeGeometry args={dimensions} />
        <meshBasicMaterial
          transparent
          alphaTest={0.1}
          map={texture}
          toneMapped={false}
          color={BONUS_ICON_TINT}
        />
      </mesh>

      {isCredits && creditAmount !== undefined && (
        <Text
          position={[0, 0, 0.002]}
          fontSize={0.025}
          font="/assets/Prototype.ttf"
          color="black"
          anchorX="center"
          anchorY="middle"
        >
          {creditAmount}
        </Text>
      )}
    </group>
  );
}
