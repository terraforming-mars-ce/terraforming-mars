import { useEffect, useRef } from "react";
import { useThree } from "@react-three/fiber";
import type * as THREE from "three";

interface WebGLContextLossHandlerProps {
  onLost: () => void;
  onRestored: (gl: THREE.WebGLRenderer) => void;
}

/** Keeps a lost WebGL context restorable (preventDefault) and reports loss and restore. */
export default function WebGLContextLossHandler({
  onLost,
  onRestored,
}: WebGLContextLossHandlerProps) {
  const gl = useThree((state) => state.gl);
  const handlers = useRef({ onLost, onRestored });
  useEffect(() => {
    handlers.current = { onLost, onRestored };
  }, [onLost, onRestored]);

  useEffect(() => {
    const canvas = gl.domElement;
    const lost = (event: Event) => {
      event.preventDefault();
      handlers.current.onLost();
    };
    const restored = () => handlers.current.onRestored(gl);
    canvas.addEventListener("webglcontextlost", lost);
    canvas.addEventListener("webglcontextrestored", restored);
    return () => {
      canvas.removeEventListener("webglcontextlost", lost);
      canvas.removeEventListener("webglcontextrestored", restored);
    };
  }, [gl]);

  return null;
}
