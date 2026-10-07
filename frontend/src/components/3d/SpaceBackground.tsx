import { useEffect, useState } from "react";
import { Canvas } from "@react-three/fiber";
import CanvasClock from "./CanvasClock.tsx";
import MenuMars from "./MenuMars.tsx";
import { useReducedMotion } from "@/hooks/useReducedMotion.ts";
import { Z_INDEX } from "@/constants/zIndex.ts";
import { QUICK_MODE } from "@/utils/quickMode.ts";

export default function SpaceBackground({ active = true }: { active?: boolean }) {
  const reduced = useReducedMotion();
  const [loadScene, setLoadScene] = useState(active && !QUICK_MODE);
  const [visible, setVisible] = useState(() => !document.hidden);
  useEffect(() => {
    if (active && !QUICK_MODE) {
      setLoadScene(true);
    }
  }, [active]);
  useEffect(() => {
    const update = () => setVisible(!document.hidden);
    document.addEventListener("visibilitychange", update);
    return () => document.removeEventListener("visibilitychange", update);
  }, []);
  const running = active && visible;
  let frameloop: "always" | "demand" | "never" = "never";
  if (running) {
    frameloop = reduced ? "demand" : "always";
  }
  return (
    <div
      aria-hidden="true"
      className="fixed inset-0 pointer-events-none bg-black"
      style={{ zIndex: Z_INDEX.GAME_BOARD_BACKGROUND }}
    >
      {loadScene && (
        <Canvas
          camera={{ position: [0, 0, 10], fov: 42 }}
          frameloop={frameloop}
          dpr={[1, 1.5]}
          fallback={<div className="size-full bg-black" />}
        >
          <CanvasClock />
          <MenuMars reduced={reduced} />
        </Canvas>
      )}
      <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(0,0,0,0.5),transparent_70%)]" />
    </div>
  );
}
