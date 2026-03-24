import { useMemo, useRef, useEffect } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { useTextures } from "../../../hooks/useTextures";
import { sphereProjectionVertex, oceanBorderFragment, addOceanProjection } from "./shaders";
import { SPHERE_RADIUS, CHROME_Z_BASE, easeOutCubic } from "./boardConstants";

const OCEAN_EMERGENCE_DURATION = 600;

interface OceanTileProps {
  overlayGeometry: THREE.BufferGeometry;
  isOceanSpace: boolean;
  tileType: string;
  sphereCenter?: THREE.Vector3;
}

export default function OceanTile({
  overlayGeometry,
  isOceanSpace,
  tileType,
  sphereCenter,
}: OceanTileProps) {
  const { camera } = useThree();
  const { settings: world3DSettings } = useWorld3DSettings();

  const { waterNormals, sand: sandTexture } = useTextures();

  const emergenceStartRef = useRef<number | null>(null);
  const isEmergingRef = useRef(false);
  const prevTileTypeRef = useRef(tileType);

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
        uSphereCenter: { value: sphereCenter || new THREE.Vector3() },
      },
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
    });
  }, []);

  const targetRadiusRef = useRef(THREE.MathUtils.lerp(0.55, 0.68, Math.random()) / 1.3);

  const oceanWaterMaterial = useMemo(() => {
    const mat = new THREE.MeshStandardMaterial({
      transparent: true,
      side: THREE.DoubleSide,
      premultipliedAlpha: true,
    });
    addOceanProjection(mat, waterNormals, sandTexture, sphereCenter || new THREE.Vector3(), 0.008, {
      uRadius: targetRadiusRef.current,
      uAspect: THREE.MathUtils.lerp(0.95, 1.12, Math.random()),
      uRotation: Math.random() * Math.PI * 2,
      uEdgeScale: THREE.MathUtils.lerp(2.8, 4.2, Math.random()),
      uSeedOffset: new THREE.Vector2(Math.random() * 100, Math.random() * 100),
    });
    mat.needsUpdate = true;
    return mat;
  }, [waterNormals, sandTexture]);

  const oceanWaterGeometry = useMemo(() => {
    const radius = 0.166;
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
    if (sphereCenter) {
      oceanBorderMaterial.uniforms.uSphereCenter.value.copy(sphereCenter);
    }

    if (oceanBorderMaterial.uniforms) {
      oceanBorderMaterial.uniforms.time.value = state.clock.elapsedTime;
    }

    const shader = (oceanWaterMaterial as any).__shader;
    if (shader) {
      shader.uniforms.time.value = state.clock.elapsedTime * 0.3;
      shader.uniforms.eye.value.copy(camera.position);
      shader.uniforms.sunDirection.value
        .set(
          world3DSettings.sunDirectionX,
          world3DSettings.sunDirectionY,
          world3DSettings.sunDirectionZ,
        )
        .normalize();
      shader.uniforms.sunIntensity.value = world3DSettings.sunIntensity;
      shader.uniforms.sunColor.value.set(
        world3DSettings.sunColor.r,
        world3DSettings.sunColor.g,
        world3DSettings.sunColor.b,
      );
      shader.uniforms.rf0.value = world3DSettings.reflectance;
      shader.uniforms.waterColor.value.set(
        world3DSettings.waterColor.r,
        world3DSettings.waterColor.g,
        world3DSettings.waterColor.b,
      );

      if (isEmergingRef.current) {
        if (emergenceStartRef.current === null) {
          emergenceStartRef.current = state.clock.elapsedTime * 1000;
          shader.uniforms.uRadius.value = 0;
          shader.uniforms.oceanAlpha.value = 0;
          shader.uniforms.uSandWidth.value = 0;
          shader.uniforms.uFoamStrength.value = 0;
          shader.uniforms.uFoamWidth.value = 0;
        }
        const elapsed = state.clock.elapsedTime * 1000 - emergenceStartRef.current;
        const t = Math.min(elapsed / OCEAN_EMERGENCE_DURATION, 1);
        const progress = easeOutCubic(t);

        shader.uniforms.uRadius.value = targetRadiusRef.current * progress;
        shader.uniforms.oceanAlpha.value = progress;
        shader.uniforms.uSandWidth.value = 0.8 * progress;
        shader.uniforms.uFoamStrength.value = 0.7 * progress;
        shader.uniforms.uFoamWidth.value = 0.08 * progress;

        if (t >= 1) {
          isEmergingRef.current = false;
        }
      }
    }
  });

  return (
    <>
      {tileType === "empty" && isOceanSpace && (
        <mesh geometry={overlayGeometry} material={oceanBorderMaterial} renderOrder={21} />
      )}

      {tileType === "ocean" && (
        <mesh
          geometry={oceanWaterGeometry}
          material={oceanWaterMaterial}
          renderOrder={12}
          raycast={() => {}}
        />
      )}
    </>
  );
}
