import { useRef, useState, useMemo, useEffect, memo, type RefObject } from "react";
import { useFrame } from "@react-three/fiber";
import { Text } from "@react-three/drei";
import * as THREE from "three";
import { HexTile2D } from "../../../utils/hex-grid-2d";
import OceanTile from "./OceanTile";
import BuildingTile from "./BuildingTile";
import VolcanoTile from "./VolcanoTile";
import NuclearZoneTile from "./NuclearZoneTile";
import MiningTile from "./MiningTile";
import ReservedAreaTile from "./ReservedAreaTile";
import WorldTreeTile from "./WorldTreeTile";
import { useTextures } from "../../../hooks/useTextures";
import {
  sphereProjectionVertex,
  hoverGlowFragment,
  availableGlowFragment,
  vpHighlightFragment,
  tileBorderVertex,
  tileBorderFragment,
  tileSurfaceVertexSnippet,
  splitSnippet,
} from "./shaders";
import { SPHERE_RADIUS, CHROME_Z_BASE, easeOutCubic } from "./boardConstants";

const BONUS_ICON_TINT = new THREE.Color(0.7, 0.7, 0.7);
const ORIGIN = new THREE.Vector3(0, 0, 0);
const _bbWorldPos = new THREE.Vector3();
const _bbNormal = new THREE.Vector3();
const _bbToCamera = new THREE.Vector3();

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
  const depthFixedRef = useRef(false);
  useFrame(({ camera }) => {
    const group = ref.current;
    if (!group) return;

    if (!depthFixedRef.current) {
      group.traverse((child) => {
        if (child instanceof THREE.Mesh && child.material) {
          const materials = Array.isArray(child.material) ? child.material : [child.material];
          for (const mat of materials) {
            mat.depthTest = false;
            mat.depthWrite = false;
            mat.transparent = true;
          }
        }
      });
      depthFixedRef.current = true;
    }

    group.getWorldPosition(_bbWorldPos);
    _bbNormal.copy(_bbWorldPos).normalize();
    _bbToCamera.copy(camera.position).sub(_bbWorldPos).normalize();
    const dot = _bbNormal.dot(_bbToCamera);

    const opacity = THREE.MathUtils.smoothstep(dot, -0.15, 0.15);
    group.visible = opacity > 0.001;
    if (!group.visible) return;

    group.traverse((child) => {
      if (child instanceof THREE.Mesh && child.material) {
        const mats = Array.isArray(child.material) ? child.material : [child.material];
        for (const mat of mats) {
          mat.opacity = opacity;
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

interface TileHoverInfo {
  position: { x: number; y: number };
  tileType: string;
  displayName?: string;
  ownerId: string | null;
  reservedById: string | null;
  isOceanSpace: boolean;
  isVolcanic: boolean;
  bonuses: { [key: string]: number };
}

interface TileProps {
  tileData: TileData3D;
  tileType:
    | "empty"
    | "ocean"
    | "greenery"
    | "city"
    | "special"
    | "volcano"
    | "nuclear-zone"
    | "mining"
    | "restricted"
    | "ecological-zone"
    | "natural-preserve"
    | "world-tree";
  ownerId?: string | null;
  ownerColor?: string;
  reservedById?: string | null;
  displayName?: string;
  isOceanSpace?: boolean;
  bonuses?: { [key: string]: number };
  onClick: () => void;
  isAvailableForPlacement?: boolean;
  animateEntrance?: boolean;
  entranceDelay?: number;
  isNewlyPlaced?: boolean;
  isVolcanic?: boolean;
  isHovered?: boolean;
  onHoverInfo?: (data: TileHoverInfo) => void;
  onHoverMove?: (position: { x: number; y: number }) => void;
  onHoverLeave?: () => void;
  sphereRadius?: number;
  sphereCenter?: THREE.Vector3;
  tileOpacity?: RefObject<number>;
  vpHighlightIntensity?: number;
  vpHighlightColor?: [number, number, number];
}

function Tile({
  tileData,
  tileType,
  ownerId,
  ownerColor,
  reservedById,
  displayName,
  isOceanSpace: _isOceanSpace = false,
  bonuses: _bonuses = {},
  onClick: _onClick,
  isAvailableForPlacement = false,
  animateEntrance = false,
  entranceDelay = 0,
  isNewlyPlaced = false,
  isVolcanic = false,
  isHovered: isHoveredProp = false,
  sphereRadius = SPHERE_RADIUS,
  sphereCenter = ORIGIN,
  tileOpacity,
  vpHighlightIntensity = 0,
  vpHighlightColor = [0.95, 0.95, 1.0],
}: TileProps) {
  const tileGroupRef = useRef<THREE.Group>(null);
  const meshRef = useRef<THREE.Mesh>(null);

  const extraMatsRef = useRef<THREE.Material[] | null>(null);
  const hovered = isHoveredProp;

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
        uSphereRadius: { value: sphereRadius },
        uZOffset: { value: CHROME_Z_BASE + 0.003 },
        uSphereCenter: { value: sphereCenter },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, [sphereRadius, sphereCenter]);

  const availableGlowMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: availableGlowFragment,
      uniforms: {
        time: { value: 0.0 },
        uSphereRadius: { value: sphereRadius },
        uZOffset: { value: CHROME_Z_BASE + 0.005 },
        uSphereCenter: { value: sphereCenter },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, [sphereRadius, sphereCenter]);

  const vpHighlightMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: vpHighlightFragment,
      uniforms: {
        uColor: { value: new THREE.Vector3(0.95, 0.95, 1.0) },
        opacity: { value: 0.0 },
        uSphereRadius: { value: sphereRadius },
        uZOffset: { value: CHROME_Z_BASE + 0.004 },
        uSphereCenter: { value: sphereCenter },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, [sphereRadius, sphereCenter]);

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
        uSphereRadius: { value: sphereRadius },
        uZOffset: { value: CHROME_Z_BASE + 0.004 },
        uSphereCenter: { value: sphereCenter },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
    });
  }, [sphereRadius, sphereCenter]);

  useFrame((state) => {
    // Update colors for hover state (avoids recreating materials)
    if (hovered) {
      hexTileMaterial.color.set("#ffff88");
      borderMaterial.uniforms.uColor.value.set("#ffffff");
    } else {
      hexTileMaterial.color.copy(baseTileColor);
      borderMaterial.uniforms.uColor.value.copy(baseBorderColor);
    }

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

    if (vpHighlightMaterial.uniforms) {
      vpHighlightMaterial.uniforms.uColor.value.set(
        vpHighlightColor[0],
        vpHighlightColor[1],
        vpHighlightColor[2],
      );
      const lerpSpeed =
        vpHighlightIntensity > vpHighlightMaterial.uniforms.opacity.value ? 0.08 : 0.04;
      vpHighlightMaterial.uniforms.opacity.value = THREE.MathUtils.lerp(
        vpHighlightMaterial.uniforms.opacity.value,
        vpHighlightIntensity,
        lerpSpeed,
      );
    }

    if (tileOpacity && tileGroupRef.current) {
      const o = tileOpacity.current;
      hexTileMaterial.opacity = baseHexOpacity * o;
      borderMaterial.uniforms.uOpacity.value = 0.9 * o;
      hoverGlowMaterial.uniforms.opacity.value *= o;

      if (!extraMatsRef.current) {
        const knownMats = new Set<THREE.Material>([
          hexTileMaterial,
          borderMaterial,
          hoverGlowMaterial,
          availableGlowMaterial,
          volcanicTintMaterial,
          vpHighlightMaterial,
        ]);
        const extras: THREE.Material[] = [];
        tileGroupRef.current.traverse((child) => {
          if (!(child instanceof THREE.Mesh)) return;
          const mats = Array.isArray(child.material) ? child.material : [child.material];
          for (const mat of mats) {
            if (knownMats.has(mat)) continue;
            if (!mat.transparent) {
              mat.transparent = true;
              mat.needsUpdate = true;
            }
            extras.push(mat);
          }
        });
        extraMatsRef.current = extras;
      }

      for (const mat of extraMatsRef.current) {
        mat.opacity = o;
      }
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

  const baseTileColor = useMemo(() => {
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
      case "nuclear-zone":
        return new THREE.Color("#2a1a0f");
      case "mining":
        return new THREE.Color("#5c3a1e");
      case "restricted":
        return new THREE.Color("#5a4030");
      case "ecological-zone":
      case "natural-preserve":
        return new THREE.Color("#2d6e2e");
      case "world-tree":
        return new THREE.Color("#1a3a12");
      default:
        return new THREE.Color("#6d4c41").multiplyScalar(0.8);
    }
  }, [tileType]);

  const baseBorderColor = useMemo(() => {
    if (ownerColor) {
      return new THREE.Color(ownerColor);
    }
    return baseTileColor.clone().multiplyScalar(0.25);
  }, [baseTileColor, ownerColor]);

  const hexTileMaterial = useMemo(() => {
    const isGreenery = tileType === "greenery";
    const material = new THREE.MeshStandardMaterial({
      color: baseTileColor,
      transparent: true,
      opacity:
        isGreenery ||
        tileType === "ocean" ||
        tileType === "city" ||
        tileType === "volcano" ||
        tileType === "nuclear-zone" ||
        tileType === "mining" ||
        tileType === "restricted" ||
        tileType === "ecological-zone" ||
        tileType === "natural-preserve" ||
        tileType === "world-tree"
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
      shader.uniforms.uSphereRadius = { value: sphereRadius };
      shader.uniforms.uZOffset = { value: CHROME_Z_BASE + 0.002 };
      shader.uniforms.uSphereCenter = { value: sphereCenter };
      shader.vertexShader =
        snippet.header +
        "\n" +
        shader.vertexShader.replace("#include <begin_vertex>", snippet.body);
    };

    return material;
  }, [baseTileColor, tileType, sphereRadius, sphereCenter]);

  const baseHexOpacity = useMemo(() => {
    if (
      tileType === "greenery" ||
      tileType === "city" ||
      tileType === "volcano" ||
      tileType === "nuclear-zone"
    )
      return 0;
    if (tileType === "empty") return 0.3;
    return 0.7;
  }, [tileType]);

  const borderMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: tileBorderVertex,
      fragmentShader: tileBorderFragment,
      uniforms: {
        uSphereRadius: { value: sphereRadius },
        uZOffset: { value: CHROME_Z_BASE + 0.0025 },
        uColor: { value: new THREE.Color(baseBorderColor.r, baseBorderColor.g, baseBorderColor.b) },
        uOpacity: { value: 0.9 },
        uNoiseTex: { value: borderNoiseTexture },
        uSphereCenter: { value: sphereCenter },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
    });
  }, [baseBorderColor, borderNoiseTexture, sphereRadius, sphereCenter]);

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
      ref={tileGroupRef}
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
          raycast={() => {}}
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

      {/* VP counting highlight */}
      <mesh geometry={overlayGeometry} material={vpHighlightMaterial} renderOrder={24} />

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

      {/* Nuclear Zone 3D tile */}
      {tileType === "nuclear-zone" && (
        <NuclearZoneTile
          isNewlyPlaced={isNewlyPlaced}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
      )}

      {/* Mining 3D tile */}
      {tileType === "mining" && (
        <MiningTile
          isNewlyPlaced={isNewlyPlaced}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
      )}

      {/* Reserved Area fence tile */}
      {tileType === "restricted" && (
        <ReservedAreaTile
          isNewlyPlaced={isNewlyPlaced}
          ownerColor={ownerColor}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
      )}

      {/* World Tree 3D tile */}
      {tileType === "world-tree" && (
        <WorldTreeTile
          isNewlyPlaced={isNewlyPlaced}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
      )}

      {/* Special tile label (rendered via displayName below) */}

      {/* Billboard display name and/or bonus icons */}
      {(displayName ||
        (tileType !== "greenery" &&
          tileType !== "ecological-zone" &&
          tileType !== "natural-preserve" &&
          tileType !== "world-tree" &&
          bonusIconGroups.length > 0)) && (
        <ClampedBillboard position={[0, 0, 0.02]} renderOrder={110}>
          {displayName && (
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
          )}
          {tileType !== "greenery" &&
            tileType !== "ecological-zone" &&
            tileType !== "natural-preserve" &&
            tileType !== "world-tree" &&
            bonusIconGroups.length > 0 && (
              <group position={[0, displayName ? -0.08 : 0, 0]}>
                {calculateIconPositions(bonusIconGroups).map((pos) => (
                  <BonusIcon
                    key={`${pos.group.type}-${pos.indexInGroup}`}
                    texture={pos.group.texture}
                    position={[pos.x, 0, 0]}
                    isCredits={pos.group.isCredits}
                    creditAmount={pos.group.isCredits ? pos.group.count : undefined}
                  />
                ))}
              </group>
            )}
        </ClampedBillboard>
      )}

      {/* Reserved tile marker (land claim) */}
      {/* Reserved tile marker (fallback for non-fence reserved tiles) */}
      {reservedById && !ownerId && tileType !== "restricted" && (
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
    </group>
  );
}

const MemoizedTile = memo(Tile);
export default MemoizedTile;

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
      <mesh renderOrder={110}>
        <planeGeometry args={dimensions} />
        <meshBasicMaterial
          transparent
          alphaTest={0.1}
          map={texture}
          toneMapped={false}
          color={BONUS_ICON_TINT}
          depthTest={false}
          depthWrite={false}
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
          renderOrder={111}
        >
          {creditAmount}
        </Text>
      )}
    </group>
  );
}
