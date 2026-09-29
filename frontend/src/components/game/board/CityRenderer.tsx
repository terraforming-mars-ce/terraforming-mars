import { memo, useEffect, useMemo, useRef, useState } from "react";
import { useFrame } from "@react-three/fiber";
import DustEffect from "./effects/DustEffect";
import { Text } from "@react-three/drei";
import * as THREE from "three";
import { createGeometry } from "./cityGeometry";
import { assetUrl } from "@/assets";
import { useTextures } from "../../../hooks/useTextures";
import { addSphereProjectionWithSoftEdges } from "./GreeneryRenderer";

import { randomSequence, type CityStyle, type CityPlot } from "./cityLayout";

const ORIGIN = new THREE.Vector3();
const materialCache = new Map<CityStyle["lighting"], THREE.MeshStandardMaterial[]>();

export function getMaterials(
  lighting: CityStyle["lighting"],
  textures: Pick<ReturnType<typeof useTextures>, "concrete" | "grass" | "sand">,
) {
  const cached = materialCache.get(lighting);
  if (cached) {
    return cached;
  }
  const canvas = document.createElement("canvas");
  const emissiveCanvas = document.createElement("canvas");
  canvas.width = emissiveCanvas.width = 256;
  canvas.height = emissiveCanvas.height = 512;
  const ctx = canvas.getContext("2d")!;
  const glow = emissiveCanvas.getContext("2d")!;
  ctx.fillStyle = "#cbd0d0";
  ctx.fillRect(0, 0, 256, 512);
  glow.fillStyle = "#62696c";
  glow.fillRect(0, 0, 256, 512);
  const rng = randomSequence(475);
  for (let row = 0; row < 8; row++) {
    ctx.fillStyle = "#89969c";
    ctx.fillRect(0, row * 64 + 57, 256, 4);
    for (let column = 0; column < 4; column++) {
      const x = column * 64 + 9;
      const y = row * 64 + 10;
      const lit = rng() > (lighting === "bright" ? 0.15 : 0.58);
      ctx.fillStyle = "#77868d";
      ctx.fillRect(x - 3, y - 3, 48, 41);
      ctx.fillStyle = lit ? "#b9cbd0" : "#233e50";
      ctx.fillRect(x, y, 42, 35);
      ctx.fillStyle = lit ? "#e7d9a9" : "#416475";
      ctx.fillRect(x + 3, y + 3, 36, 7);
      ctx.fillStyle = "#89969c";
      ctx.fillRect(x + 20, y, 2, 35);
      if (lit) {
        glow.fillStyle = "#9caeb8";
        glow.fillRect(x, y, 42, 35);
        glow.fillStyle = "#f0cf89";
        glow.fillRect(x + 3, y + 3, 36, 7);
      }
    }
  }
  const map = new THREE.CanvasTexture(canvas);
  const emissiveMap = new THREE.CanvasTexture(emissiveCanvas);
  for (const texture of [map, emissiveMap]) {
    texture.wrapS = texture.wrapT = THREE.RepeatWrapping;
    texture.colorSpace = THREE.SRGBColorSpace;
    texture.anisotropy = 4;
  }
  const materials = ["#f2eee3", "#c3d3d8", "#a8b8c5"].map(
    (color) =>
      new THREE.MeshStandardMaterial({
        color,
        map,
        emissiveMap,
        emissive: "#ffffff",
        emissiveIntensity: lighting === "bright" ? 1.4 : 0.6,
        roughness: 0.65,
        metalness: 0.15,
      }),
  );
  for (const [color, roughness, metalness] of [
    ["#68777d", 0.85, 0.1],
    ["#b8b6a9", 0.9, 0],
    ["#344650", 0.7, 0.2],
    ["#b67c47", 0.5, 0.6],
    ["#496d38", 0.95, 0],
    ["#244b6d", 0.35, 0.5],
  ] as const) {
    materials.push(
      new THREE.MeshStandardMaterial({
        color,
        roughness,
        metalness,
        emissive: color,
        emissiveIntensity: 0.22,
      }),
    );
  }
  materials.push(
    new THREE.MeshStandardMaterial({
      color: "#79b1bb",
      roughness: 0.25,
      metalness: 0.45,
      emissive: "#264754",
      emissiveIntensity: 0.25,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      color: "#8dced9",
      transparent: true,
      opacity: 0.3,
      depthWrite: false,
      roughness: 0.15,
      metalness: 0.25,
      emissive: "#488ba6",
      emissiveIntensity: 0.2,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      color: "#ffd899",
      emissive: "#ffc56c",
      emissiveIntensity: 2,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      color: "#1b252b",
      roughness: 0.9,
      emissive: "#1b252b",
      emissiveIntensity: 0.15,
      side: THREE.DoubleSide,
    }),
  );
  materials[4].side = THREE.DoubleSide;
  for (const index of [3, 4, 5]) {
    materials[index].map = textures.concrete;
    materials[index].bumpMap = textures.concrete;
    materials[index].bumpScale = 0.0003;
    materials[index].emissiveMap = textures.concrete;
  }
  materials[4].color.set("#a8a399");
  materials[7].map = textures.grass;
  materials[7].emissiveMap = textures.grass;
  materials.push(
    new THREE.MeshStandardMaterial({
      map: textures.concrete,
      bumpMap: textures.concrete,
      bumpScale: 0.00025,
      color: "#697175",
      roughness: 0.98,
      emissive: "#697175",
      emissiveMap: textures.concrete,
      emissiveIntensity: 0.25,
    }),
  );
  const pool = document.createElement("canvas");
  pool.width = pool.height = 64;
  const poolContext = pool.getContext("2d")!;
  const gradient = poolContext.createRadialGradient(32, 32, 0, 32, 32, 32);
  gradient.addColorStop(0, "rgba(255,255,255,0.8)");
  gradient.addColorStop(0.35, "rgba(255,255,255,0.4)");
  gradient.addColorStop(1, "rgba(255,255,255,0)");
  poolContext.fillStyle = gradient;
  poolContext.fillRect(0, 0, 64, 64);
  const poolMap = new THREE.CanvasTexture(pool);
  materials.push(
    new THREE.MeshStandardMaterial({
      map: poolMap,
      color: "#000000",
      transparent: true,
      opacity: 0.55,
      depthWrite: false,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      map: poolMap,
      color: "#edaf59",
      emissive: "#edaf59",
      emissiveMap: poolMap,
      emissiveIntensity: 1.2,
      transparent: true,
      opacity: 0.3,
      depthWrite: false,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      map: textures.sand,
      bumpMap: textures.sand,
      bumpScale: 0.0005,
      color: "#88694f",
      roughness: 1,
      emissive: "#88694f",
      emissiveMap: textures.sand,
      emissiveIntensity: 0.2,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      color: "#99aab0",
      roughness: 0.38,
      metalness: 0.75,
    }),
  );
  materials.push(
    new THREE.MeshStandardMaterial({
      color: "#dce5e8",
      roughness: 0.3,
      metalness: 0.25,
      side: THREE.DoubleSide,
    }),
  );
  materials.forEach((material) => {
    material.vertexColors = true;
  });
  materialCache.set(lighting, materials);
  return materials;
}

interface CityRendererProps {
  renderBuildings?: boolean;
  activeExits?: readonly number[];
  plot: CityPlot;
  sphereCenter?: THREE.Vector3;
  groupInverseMatrix?: THREE.Matrix4;
  showLabel?: boolean;
  connectedGround?: boolean;
  isNewlyPlaced?: boolean;
}

export function CityGroundPatch({
  x = 0,
  y = 0,
  elevation = 0.0003,
  renderOrder = 11,
  radius,
  surface,
  sphereCenter,
  groupInverseMatrix,
}: {
  x?: number;
  y?: number;
  elevation?: number;
  renderOrder?: number;
  radius: number;
  surface: "soil" | "grass" | "paving" | "asphalt";
  sphereCenter: THREE.Vector3;
  groupInverseMatrix?: THREE.Matrix4;
}) {
  const { sand, grass, concrete, noiseMid, noiseHigh } = useTextures();
  const geometry = useMemo(() => {
    const geo = new THREE.PlaneGeometry(radius * 3, radius * 3, 24, 24);
    const uv = geo.getAttribute("uv");
    for (let i = 0; i < uv.count; i++) {
      uv.setXY(i, uv.getX(i) * 3, uv.getY(i) * 3);
    }
    return geo;
  }, [radius]);
  const material = useMemo(() => {
    let map = sand;
    let color = "#9a6b4b";
    let opacity = 0.72;
    if (surface === "grass") {
      map = grass;
      color = "#75865c";
      opacity = 0.95;
    } else if (surface === "asphalt") {
      map = concrete;
      color = "#8b9294";
      opacity = 1;
    } else if (surface === "paving") {
      map = concrete;
      color = "#6c7779";
      opacity = 1;
    }
    const mat = new THREE.MeshStandardMaterial({
      map,
      color,
      opacity,
      bumpMap: map,
      bumpScale: 0.0005,
      roughness: 0.95,
      transparent: true,
      depthWrite: false,
      alphaTest: 0.005,
      side: THREE.DoubleSide,
    });
    const paved = surface === "paving";
    addSphereProjectionWithSoftEdges(
      mat,
      elevation,
      noiseMid,
      noiseHigh,
      radius,
      sphereCenter,
      groupInverseMatrix,
      {
        circular: true,
        overflow: paved ? 0 : 0.12,
        bandWidth: paved ? 0.015 : 0.4,
        warp: paved ? 0 : 0.22,
      },
    );
    return mat;
  }, [
    sand,
    grass,
    concrete,
    noiseMid,
    noiseHigh,
    surface,
    radius,
    elevation,
    sphereCenter,
    groupInverseMatrix,
  ]);
  useEffect(
    () => () => {
      geometry.dispose();
      material.dispose();
    },
    [geometry, material],
  );
  return (
    <mesh
      position={[x, y, 0]}
      geometry={geometry}
      material={material}
      renderOrder={renderOrder}
      raycast={() => {}}
      dispose={null}
    />
  );
}

function CityRenderer({
  plot,
  sphereCenter = ORIGIN,
  groupInverseMatrix,
  showLabel = false,
  connectedGround = false,
  isNewlyPlaced = false,
  activeExits,
  renderBuildings = true,
}: CityRendererProps) {
  const buildings = useRef<THREE.Group>(null);
  const started = useRef<number | null>(null);
  const emerging = useRef(isNewlyPlaced);
  const [dust, setDust] = useState(isNewlyPlaced);
  useEffect(() => {
    if (isNewlyPlaced) {
      started.current = null;
      emerging.current = true;
      setDust(true);
    }
  }, [isNewlyPlaced]);
  useFrame(({ clock }) => {
    if (!buildings.current || !emerging.current) {
      return;
    }
    started.current ??= clock.elapsedTime;
    const t = Math.min(1, (clock.elapsedTime - started.current) / 0.8);
    buildings.current.position.z = -0.08 * Math.pow(1 - t, 3);
    if (t === 1) {
      emerging.current = false;
    }
  });
  const geometries = useMemo(
    () => (renderBuildings ? createGeometry(plot.layout, plot, activeExits) : []),
    [plot, activeExits, renderBuildings],
  );
  useEffect(
    () => () => {
      for (const geometry of geometries) {
        geometry?.dispose();
      }
    },
    [geometries],
  );
  const textures = useTextures();
  const cityMaterials = getMaterials(plot.layout.style.lighting, textures);
  let groundSurface: "soil" | "grass" | "paving" | "asphalt" = "soil";
  if (plot.layout.style.cover === "dome") {
    groundSurface = "paving";
  } else if (plot.layout.style.density === "dense") {
    groundSurface = "asphalt";
  } else if (plot.layout.style.landscaping !== "sparse") {
    groundSurface = "grass";
  }
  const dirtPatches = useMemo(() => {
    if (plot.layout.style.exposure !== "mostly-buried") {
      return [];
    }
    const random = randomSequence(4271);
    return Array.from({ length: 13 }, (_, i) => {
      const angle = (i / 13) * Math.PI * 2 + random() * 0.3;
      const distance = 0.065 + random() * 0.055;
      return {
        x: Math.cos(angle) * distance,
        y: Math.sin(angle) * distance,
        radius: 0.014 + random() * 0.022,
      };
    });
  }, [plot.layout]);
  const quaternion = useMemo(
    () => new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 0, 1), plot.normal),
    [plot.normal],
  );

  return (
    <group position={plot.worldPosition} quaternion={quaternion}>
      {plot.layout.style.cover === "dome" && (
        <CityGroundPatch
          radius={0.15}
          surface="asphalt"
          elevation={0.0002}
          renderOrder={10}
          sphereCenter={sphereCenter}
          groupInverseMatrix={groupInverseMatrix}
        />
      )}
      {(!connectedGround || groundSurface !== "grass") && (
        <CityGroundPatch
          radius={plot.layout.style.cover === "dome" ? 0.113 : 0.155}
          surface={groundSurface}
          sphereCenter={sphereCenter}
          groupInverseMatrix={groupInverseMatrix}
        />
      )}
      {dirtPatches.map((patch, i) => (
        <CityGroundPatch
          key={`dirt-${i}`}
          {...patch}
          surface="soil"
          elevation={0.0005}
          sphereCenter={sphereCenter}
          groupInverseMatrix={groupInverseMatrix}
        />
      ))}
      {groundSurface !== "grass" &&
        plot.layout.parks.map((park, i) => (
          <CityGroundPatch
            key={i}
            x={park.x}
            y={park.y}
            elevation={0.0004}
            radius={park.radius}
            surface="grass"
            sphereCenter={sphereCenter}
            groupInverseMatrix={groupInverseMatrix}
          />
        ))}
      <group ref={buildings}>
        {geometries.map(
          (geometry, index) =>
            geometry && (
              <mesh
                key={index}
                geometry={geometry}
                material={cityMaterials[index]}
                renderOrder={15}
                raycast={() => {}}
                dispose={null}
              />
            ),
        )}
      </group>
      {dust && (
        <DustEffect
          position={plot.worldPosition}
          normal={plot.normal}
          onComplete={() => setDust(false)}
        />
      )}
      {showLabel && (
        <Text
          position={[0, -0.137, 0.005]}
          font={assetUrl("fonts/prototype")}
          fontSize={0.01}
          maxWidth={0.18}
          color="#fff2d5"
          anchorX="center"
          anchorY="middle"
          raycast={() => {}}
        >
          {plot.layout.name}
        </Text>
      )}
    </group>
  );
}

export default memo(CityRenderer);
