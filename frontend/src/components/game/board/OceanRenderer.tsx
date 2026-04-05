import { useMemo, useRef, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { useTextures } from "../../../hooks/useTextures";
import { HexGrid2D, type HexCoordinate } from "../../../utils/hex-grid-2d";
import { createOceanRendererMaterial } from "./shaders";
import { easeOutCubic } from "./boardConstants";
import {
  buildOceanDataTexture,
  updateEmergenceInTexture,
  hexToPixel,
  type OceanDataResult,
} from "./oceanDataTexture";

const OCEAN_EMERGENCE_DURATION = 600;

interface OceanTileData {
  coordinate: HexCoordinate;
}

interface OceanRendererProps {
  oceanTiles: OceanTileData[];
  newOceanKeys: Set<string>;
  hoveredOceanHexKey: string | null;
  sphereCenter: THREE.Vector3;
}

export default function OceanRenderer({
  oceanTiles,
  newOceanKeys,
  hoveredOceanHexKey,
  sphereCenter,
}: OceanRendererProps) {
  const { camera } = useThree();
  const { settings: world3DSettings } = useWorld3DSettings();
  const { waterNormals, sand: sandTexture } = useTextures();

  const emergenceMapRef = useRef<Map<string, number | null>>(new Map());
  const oceanDataRef = useRef<OceanDataResult | null>(null);
  const prevOceanKeysRef = useRef<string>("");

  const geometry = useMemo(() => {
    const radius = 3.5;
    const radialSegments = 32;
    const angularSegments = 128;

    const geo = new THREE.BufferGeometry();
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
      const ringStart = 1 + (r - 1) * angularSegments;
      const nextRingStart = 1 + r * angularSegments;
      for (let a = 0; a < angularSegments; a++) {
        const next = (a + 1) % angularSegments;
        indices.push(ringStart + a, nextRingStart + a, nextRingStart + next);
        indices.push(ringStart + a, nextRingStart + next, ringStart + next);
      }
    }

    geo.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
    geo.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
    geo.setIndex(indices);
    geo.computeVertexNormals();

    return geo;
  }, []);

  const emptyTexture = useMemo(() => {
    const data = new Float32Array(4);
    const tex = new THREE.DataTexture(data, 1, 1, THREE.RGBAFormat, THREE.FloatType);
    tex.minFilter = THREE.NearestFilter;
    tex.magFilter = THREE.NearestFilter;
    tex.needsUpdate = true;
    return tex;
  }, []);

  const material = useMemo(() => {
    return createOceanRendererMaterial(waterNormals, sandTexture, sphereCenter, emptyTexture);
  }, [waterNormals, sandTexture, sphereCenter, emptyTexture]);

  useEffect(() => {
    const currentKeys = oceanTiles
      .map((t) => HexGrid2D.coordinateToKey(t.coordinate))
      .sort()
      .join("|");

    if (currentKeys === prevOceanKeysRef.current) {
      return;
    }
    prevOceanKeysRef.current = currentKeys;

    const coordinates = oceanTiles.map((t) => t.coordinate);
    const result = buildOceanDataTexture(coordinates);

    const prevEmergence = emergenceMapRef.current;
    for (let i = 0; i < result.points.length; i++) {
      const key = HexGrid2D.coordinateToKey(oceanTiles[i].coordinate);
      if (newOceanKeys.has(key) && !prevEmergence.has(key)) {
        result.points[i].emergence = 0;
        updateEmergenceInTexture(result.data, i, 0);
        emergenceMapRef.current.set(key, null);
      } else if (!prevEmergence.has(key)) {
        result.points[i].emergence = 1.0;
        emergenceMapRef.current.set(key, -1);
      }
    }

    result.texture.needsUpdate = true;
    oceanDataRef.current = result;

    material.uniforms.uOceanData.value = result.texture;
    material.uniforms.uPointCount.value = result.pointCount;
    material.uniforms.uEdgeCount.value = result.edgeCount;
  }, [oceanTiles, newOceanKeys, material]);

  useFrame((state) => {
    const elapsedMs = state.clock.elapsedTime * 1000;
    const oceanData = oceanDataRef.current;

    material.uniforms.time.value = state.clock.elapsedTime * 0.3;
    material.uniforms.eye.value.copy(camera.position);
    material.uniforms.sunDirection.value
      .set(
        world3DSettings.sunDirectionX,
        world3DSettings.sunDirectionY,
        world3DSettings.sunDirectionZ,
      )
      .normalize();
    material.uniforms.sunIntensity.value = world3DSettings.sunIntensity;
    material.uniforms.sunColor.value.set(
      world3DSettings.sunColor.r,
      world3DSettings.sunColor.g,
      world3DSettings.sunColor.b,
    );
    material.uniforms.rf0.value = world3DSettings.reflectance;
    material.uniforms.waterColor.value.set(
      world3DSettings.waterColor.r,
      world3DSettings.waterColor.g,
      world3DSettings.waterColor.b,
    );
    material.uniforms.uSphereCenter.value.copy(sphereCenter);

    // Hover
    if (hoveredOceanHexKey) {
      const parts = hoveredOceanHexKey.split(",").map(Number);
      const pos = hexToPixel({ q: parts[0], r: parts[1], s: parts[2] });
      material.uniforms.uHoverCenter.value.set(pos.x, pos.y);
      material.uniforms.uHoverActive.value = 1.0;
    } else {
      material.uniforms.uHoverActive.value = 0.0;
    }

    // Emergence animation
    if (!oceanData) {
      return;
    }

    let needsTextureUpdate = false;
    for (let i = 0; i < oceanData.points.length; i++) {
      const key = HexGrid2D.coordinateToKey(oceanTiles[i].coordinate);
      const startTime = emergenceMapRef.current.get(key);

      if (startTime === undefined || startTime === -1) {
        continue;
      }

      if (startTime === null) {
        emergenceMapRef.current.set(key, elapsedMs);
        updateEmergenceInTexture(oceanData.data, i, 0);
        needsTextureUpdate = true;
        continue;
      }

      const elapsed = elapsedMs - startTime;
      const t = Math.min(elapsed / OCEAN_EMERGENCE_DURATION, 1);
      const progress = easeOutCubic(t);

      oceanData.points[i].emergence = progress;
      updateEmergenceInTexture(oceanData.data, i, progress);
      needsTextureUpdate = true;

      if (t >= 1) {
        emergenceMapRef.current.set(key, -1);
      }
    }

    if (needsTextureUpdate) {
      oceanData.texture.needsUpdate = true;
    }
  });

  if (oceanTiles.length === 0) {
    return null;
  }

  return (
    <mesh
      geometry={geometry}
      material={material}
      renderOrder={12}
      raycast={() => {}}
      frustumCulled={false}
    />
  );
}
