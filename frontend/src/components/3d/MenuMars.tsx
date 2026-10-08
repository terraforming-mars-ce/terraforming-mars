import { assetUrl } from "@/assets";
import { Component, Suspense, useEffect, useMemo, useRef, type ReactNode } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import { useGLTF, useTexture } from "@react-three/drei";
import * as THREE from "three";
import { useAppPhaseStore } from "@/stores/appPhaseStore";
import { GRAPHICS } from "@/utils/graphicsQuality.ts";
import SkyboxLoader from "../game/view/SkyboxLoader";
import AtmosphereRenderer from "../game/board/AtmosphereRenderer";
import PlanetAtmosphere from "../game/board/PlanetAtmosphere";
import { MARS_ATMOSPHERE, PLANET_FILL_LIGHT } from "../game/board/solarSystemConfig";

const MENU_SUN_INTENSITY = 0.65;

class DecorativeAsset extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false };
  static getDerivedStateFromError() {
    return { failed: true };
  }
  render() {
    return this.state.failed ? null : this.props.children;
  }
}

function Planet({ reduced }: { reduced: boolean }) {
  const texture = useTexture(
    assetUrl("textures/planets/mars/surface", GRAPHICS.planetTexturePixels("mars")),
  );
  const sphere = useRef<THREE.Mesh>(null);
  useEffect(() => {
    texture.colorSpace = THREE.SRGBColorSpace;
  }, [texture]);
  useFrame((state) => {
    if (sphere.current && !reduced) {
      sphere.current.rotation.y = (state.clock.elapsedTime * Math.PI * 2) / 300;
    }
  });
  return (
    <>
      <mesh ref={sphere} rotation={[0.2, 0.6, 0.1]}>
        <sphereGeometry args={[3, 96, 64]} />
        <meshStandardMaterial map={texture} roughness={1} metalness={0} envMapIntensity={0} />
      </mesh>
      <PlanetAtmosphere radius={3} profile={MARS_ATMOSPHERE} />
    </>
  );
}

function Phobos({ reduced }: { reduced: boolean }) {
  const { scene } = useGLTF(assetUrl("models/phobos"));
  const moon = useMemo(() => {
    const clone = scene.clone(true);
    const materials: THREE.Material[] = [];
    clone.traverse((object) => {
      if (!(object instanceof THREE.Mesh)) {
        return;
      }
      const cloneMaterial = (source: THREE.Material) => {
        const material = source.clone();
        materials.push(material);
        return material;
      };
      object.material = Array.isArray(object.material)
        ? object.material.map(cloneMaterial)
        : cloneMaterial(object.material);
    });
    const bounds = new THREE.Box3().setFromObject(clone);
    const size = bounds.getSize(new THREE.Vector3());
    const center = bounds.getCenter(new THREE.Vector3());
    clone.position.sub(center);
    const group = new THREE.Group();
    group.add(clone);
    group.scale.setScalar(0.32 / Math.max(size.x, size.y, size.z, 0.001));
    return { group, materials };
  }, [scene]);
  useEffect(() => () => moon.materials.forEach((material) => material.dispose()), [moon]);
  const ref = useRef<THREE.Group>(null);
  useFrame((state) => {
    if (!ref.current) {
      return;
    }
    const angle = 2.5 + (reduced ? 0 : (state.clock.elapsedTime * Math.PI * 2) / 240);
    ref.current.position.set(Math.cos(angle) * 4.2, Math.sin(angle) * 0.9, Math.sin(angle) * 4.2);
    ref.current.rotation.set(0.2, -angle, 0.4);
  });
  return (
    <group ref={ref}>
      <primitive object={moon.group} />
    </group>
  );
}

export default function MenuMars({ reduced }: { reduced: boolean }) {
  const { camera, size, invalidate } = useThree();
  const phase = useAppPhaseStore((state) => state.phase);
  const planet = useRef<THREE.Group>(null);
  const sunLight = useRef<THREE.PointLight>(null);
  const pointer = useRef(new THREE.Vector2());
  const currentPointer = useRef(new THREE.Vector2());
  const target = useMemo(() => new THREE.Vector3(), []);
  const compact = size.width < 1024;
  const landscape = size.width > size.height;
  const landing = phase.kind === "menu" && phase.route === "landing";
  const lobby = phase.kind !== "menu";
  const pose = useMemo(() => {
    if (compact && landscape) {
      return { x: 4.6, y: -1.4, z: -1.5 };
    }
    if (compact) {
      return { x: 2.6, y: -3.2, z: -1.5 };
    }
    if (lobby) {
      return { x: 3.9, y: -3.7, z: 1.4 };
    }
    if (landing) {
      return { x: 3.9, y: 0, z: 0 };
    }
    return { x: 5.4, y: -0.8, z: -0.5 };
  }, [compact, landscape, landing, lobby]);
  const initialized = useRef(false);
  useEffect(() => {
    const move = (event: PointerEvent) => {
      const focused = document.activeElement;
      const typing =
        focused instanceof HTMLInputElement ||
        focused instanceof HTMLTextAreaElement ||
        focused instanceof HTMLSelectElement;
      const modal = document.querySelector('[role="dialog"]');
      if (reduced || typing || modal || event.pointerType !== "mouse") {
        pointer.current.set(0, 0);
        return;
      }
      pointer.current.set(
        (event.clientX / window.innerWidth - 0.5) * 0.12,
        (event.clientY / window.innerHeight - 0.5) * 0.08,
      );
    };
    const reset = () => pointer.current.set(0, 0);
    window.addEventListener("pointermove", move);
    window.addEventListener("blur", reset);
    window.addEventListener("focusin", reset);
    return () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("blur", reset);
      window.removeEventListener("focusin", reset);
    };
  }, [reduced]);
  useEffect(() => {
    invalidate();
  }, [pose, reduced, invalidate]);
  useFrame((state, delta) => {
    const blend = reduced || !initialized.current ? 1 : 1 - Math.exp(-Math.min(delta, 0.05) * 6);
    if (planet.current) {
      planet.current.position.lerp(target.set(pose.x, pose.y, pose.z), blend);
    }
    currentPointer.current.lerp(reduced ? new THREE.Vector2() : pointer.current, blend);
    const time = reduced ? 0 : state.clock.elapsedTime;
    camera.position.set(
      currentPointer.current.x + Math.sin(time / 80) * 0.16,
      -currentPointer.current.y + Math.sin(time / 110) * 0.08,
      10,
    );
    camera.lookAt(0, 0, 0);
    initialized.current = true;
  });
  return (
    <AtmosphereRenderer sunLight={sunLight}>
      <SkyboxLoader onReady={invalidate} />
      <ambientLight intensity={0.025} />
      <pointLight
        ref={sunLight}
        position={[-7, 4, 10]}
        intensity={MENU_SUN_INTENSITY}
        color={[1, 0.93, 0.85]}
        decay={0}
      />
      <directionalLight
        position={[0, 1, 0]}
        intensity={MENU_SUN_INTENSITY * PLANET_FILL_LIGHT.keyIntensityRatio}
        color={PLANET_FILL_LIGHT.keyColor}
      />
      <hemisphereLight
        args={[
          PLANET_FILL_LIGHT.skyColor,
          PLANET_FILL_LIGHT.groundColor,
          MENU_SUN_INTENSITY * PLANET_FILL_LIGHT.intensityRatio,
        ]}
      />
      <group ref={planet}>
        <DecorativeAsset>
          <Suspense fallback={null}>
            <Planet reduced={reduced} />
          </Suspense>
        </DecorativeAsset>
        <DecorativeAsset>
          <Suspense fallback={null}>
            <Phobos reduced={reduced} />
          </Suspense>
        </DecorativeAsset>
      </group>
    </AtmosphereRenderer>
  );
}
