import { Component, Suspense, useEffect, useMemo, useRef, type ReactNode } from "react";
import { useFrame, useThree } from "@react-three/fiber";
import { useGLTF, useTexture } from "@react-three/drei";
import * as THREE from "three";
import { useAppPhaseStore } from "@/stores/appPhaseStore";
import SkyboxLoader from "../game/view/SkyboxLoader";

class DecorativeAsset extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false };
  static getDerivedStateFromError() {
    return { failed: true };
  }
  render() {
    return this.state.failed ? null : this.props.children;
  }
}

const atmosphereVertex = `
  varying vec3 vNormal;
  varying vec3 vPosition;
  void main() {
    vNormal = normalize(normalMatrix * normal);
    vec4 p = modelViewMatrix * vec4(position, 1.0);
    vPosition = p.xyz;
    gl_Position = projectionMatrix * p;
  }
`;
const atmosphereFragment = `
  varying vec3 vNormal;
  varying vec3 vPosition;
  void main() {
    vec3 n = normalize(vNormal);
    float rim = pow(1.0 - abs(dot(n, normalize(-vPosition))), 4.0);
    float light = smoothstep(-0.15, 0.7, dot(n, normalize(vec3(-0.9, 0.35, 0.15))));
    gl_FragColor = vec4(vec3(0.65, 0.35, 0.22), rim * light * 0.55);
  }
`;

function Planet({ reduced }: { reduced: boolean }) {
  const texture = useTexture("/assets/textures/mars_8k.jpg");
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
      <mesh>
        <sphereGeometry args={[3.045, 64, 48]} />
        <shaderMaterial
          vertexShader={atmosphereVertex}
          fragmentShader={atmosphereFragment}
          transparent
          depthWrite={false}
          blending={THREE.AdditiveBlending}
        />
      </mesh>
    </>
  );
}

function Phobos({ reduced }: { reduced: boolean }) {
  const { scene } = useGLTF("/assets/models/phobos.glb");
  const moon = useMemo(() => {
    const clone = scene.clone(true);
    const bounds = new THREE.Box3().setFromObject(clone);
    const size = bounds.getSize(new THREE.Vector3());
    const center = bounds.getCenter(new THREE.Vector3());
    clone.position.sub(center);
    const group = new THREE.Group();
    group.add(clone);
    group.scale.setScalar(0.32 / Math.max(size.x, size.y, size.z, 0.001));
    return group;
  }, [scene]);
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
      <primitive object={moon} />
    </group>
  );
}

export default function MenuMars({ reduced }: { reduced: boolean }) {
  const { camera, size, invalidate } = useThree();
  const phase = useAppPhaseStore((state) => state.phase);
  const planet = useRef<THREE.Group>(null);
  const pointer = useRef(new THREE.Vector2());
  const currentPointer = useRef(new THREE.Vector2());
  const target = useMemo(() => new THREE.Vector3(), []);
  const compact = size.width < 1024;
  const landing = phase.kind === "menu" && phase.route === "landing";
  const lobby = phase.kind !== "menu";
  const pose = useMemo(() => {
    if (compact) {
      return { x: 0.8, y: -2.8, z: -1.5 };
    }
    if (lobby) {
      return { x: 3.9, y: -3.7, z: 1.4 };
    }
    if (landing) {
      return { x: 3.9, y: 0, z: 0 };
    }
    return { x: 5.4, y: -0.8, z: -0.5 };
  }, [compact, landing, lobby]);
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
    <>
      <SkyboxLoader onReady={invalidate} />
      <ambientLight intensity={0.025} />
      <directionalLight position={[-7, 4, 1]} intensity={2.2} color="#fff0df" />
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
    </>
  );
}
