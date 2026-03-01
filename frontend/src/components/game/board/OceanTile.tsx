import { useMemo, useRef, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { panState } from "../controls/PanControls";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { useTextures } from "../../../hooks/useTextures";
import { sphereProjectionVertex, oceanBorderFragment, createOceanMaterial } from "./shaders";
import { SPHERE_RADIUS, CHROME_Z_BASE, easeOutCubic } from "./boardConstants";

const OCEAN_EMERGENCE_DURATION = 600;

interface OceanTileProps {
  overlayGeometry: THREE.BufferGeometry;
  isOceanSpace: boolean;
  tileType: string;
  onClick: () => void;
  onHoverChange: (hovered: boolean) => void;
  onPointerEnterCapture?: (event: THREE.Event & { nativeEvent: PointerEvent }) => void;
  onPointerMoveCapture?: (event: THREE.Event & { nativeEvent: PointerEvent }) => void;
  onPointerLeaveCapture?: () => void;
}

export default function OceanTile({
  overlayGeometry,
  isOceanSpace,
  tileType,
  onClick,
  onHoverChange,
  onPointerEnterCapture,
  onPointerMoveCapture,
  onPointerLeaveCapture,
}: OceanTileProps) {
  const { camera } = useThree();
  const { settings: world3DSettings } = useWorld3DSettings();

  const { waterNormals, sand: sandTexture } = useTextures();

  const emergenceStartRef = useRef<number | null>(null);
  const isEmergingRef = useRef(false);
  const prevTileTypeRef = useRef(tileType);

  // Self-detect ocean emergence: track tileType transition to "ocean"
  useEffect(() => {
    if (tileType === "ocean" && prevTileTypeRef.current !== "ocean") {
      isEmergingRef.current = true;
      emergenceStartRef.current = null;
    }
    prevTileTypeRef.current = tileType;
  }, [tileType]);

  const oceanBorderMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: sphereProjectionVertex,
      fragmentShader: oceanBorderFragment,
      uniforms: {
        time: { value: 0.0 },
        uSphereRadius: { value: SPHERE_RADIUS },
        uZOffset: { value: CHROME_Z_BASE + 0.004 },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
    });
  }, []);

  const targetRadiusRef = useRef(THREE.MathUtils.lerp(0.55, 0.68, Math.random()) / 1.3);

  const oceanWaterMaterial = useMemo(() => {
    return createOceanMaterial(waterNormals, sandTexture, {
      uRadius: { value: targetRadiusRef.current },
      uAspect: { value: THREE.MathUtils.lerp(0.95, 1.12, Math.random()) },
      uRotation: { value: Math.random() * Math.PI * 2 },
      uEdgeScale: { value: THREE.MathUtils.lerp(2.8, 4.2, Math.random()) },
      uSeedOffset: {
        value: new THREE.Vector2(Math.random() * 100, Math.random() * 100),
      },
    });
  }, [waterNormals, sandTexture]);

  const oceanWaterGeometry = useMemo(() => {
    const radius = 0.166 * 1.3;
    const radialSegments = 8;
    const angularSegments = 32;

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
      const ringStart = 1 + (r - 1) * angularSegments;
      const nextRingStart = 1 + r * angularSegments;
      for (let a = 0; a < angularSegments; a++) {
        const next = (a + 1) % angularSegments;
        indices.push(ringStart + a, nextRingStart + a, nextRingStart + next);
        indices.push(ringStart + a, nextRingStart + next, ringStart + next);
      }
    }

    geometry.setAttribute("position", new THREE.Float32BufferAttribute(vertices, 3));
    geometry.setAttribute("uv", new THREE.Float32BufferAttribute(uvs, 2));
    geometry.setIndex(indices);
    geometry.computeVertexNormals();

    return geometry;
  }, []);

  useFrame((state) => {
    if (oceanBorderMaterial.uniforms) {
      oceanBorderMaterial.uniforms.time.value = state.clock.elapsedTime;
    }

    if (oceanWaterMaterial.uniforms) {
      oceanWaterMaterial.uniforms.time.value = state.clock.elapsedTime * 0.3;
      oceanWaterMaterial.uniforms.eye.value.copy(camera.position);
      oceanWaterMaterial.uniforms.sunDirection.value
        .set(
          world3DSettings.sunDirectionX,
          world3DSettings.sunDirectionY,
          world3DSettings.sunDirectionZ,
        )
        .normalize();
      oceanWaterMaterial.uniforms.sunIntensity.value = world3DSettings.sunIntensity;
      oceanWaterMaterial.uniforms.rf0.value = world3DSettings.reflectance;
      oceanWaterMaterial.uniforms.waterColor.value.set(
        world3DSettings.waterColor.r,
        world3DSettings.waterColor.g,
        world3DSettings.waterColor.b,
      );

      // Emergence animation: grow uRadius and fade in alpha
      if (isEmergingRef.current) {
        if (emergenceStartRef.current === null) {
          emergenceStartRef.current = state.clock.elapsedTime * 1000;
          // Reset uniforms to 0 at animation start
          oceanWaterMaterial.uniforms.uRadius.value = 0;
          oceanWaterMaterial.uniforms.alpha.value = 0;
          oceanWaterMaterial.uniforms.uSandWidth.value = 0;
          oceanWaterMaterial.uniforms.uFoamStrength.value = 0;
          oceanWaterMaterial.uniforms.uFoamWidth.value = 0;
        }
        const elapsed = state.clock.elapsedTime * 1000 - emergenceStartRef.current;
        const t = Math.min(elapsed / OCEAN_EMERGENCE_DURATION, 1);
        const progress = easeOutCubic(t);

        oceanWaterMaterial.uniforms.uRadius.value = targetRadiusRef.current * progress;
        oceanWaterMaterial.uniforms.alpha.value = progress;
        oceanWaterMaterial.uniforms.uSandWidth.value = 0.8 * progress;
        oceanWaterMaterial.uniforms.uFoamStrength.value = 0.7 * progress;
        oceanWaterMaterial.uniforms.uFoamWidth.value = 0.08 * progress;

        if (t >= 1) {
          isEmergingRef.current = false;
        }
      }
    }
  });

  return (
    <>
      {/* Ocean space indicator - blue gradient fading to center */}
      {tileType === "empty" && isOceanSpace && (
        <mesh geometry={overlayGeometry} material={oceanBorderMaterial} renderOrder={21} />
      )}

      {/* Animated ocean water effect - vertices projected onto sphere in shader */}
      {tileType === "ocean" && (
        <mesh
          geometry={oceanWaterGeometry}
          material={oceanWaterMaterial}
          renderOrder={12}
          onPointerEnter={(event: any) => {
            if (!panState.isPanning) {
              onHoverChange(true);
              onPointerEnterCapture?.(event);
            }
          }}
          onPointerMove={(event: any) => {
            if (!panState.isPanning) {
              onPointerMoveCapture?.(event);
            }
          }}
          onPointerLeave={() => {
            onHoverChange(false);
            onPointerLeaveCapture?.();
          }}
          onClick={(event) => {
            if (panState.isPanning) return;
            event.stopPropagation();
            onClick();
          }}
        />
      )}
    </>
  );
}
