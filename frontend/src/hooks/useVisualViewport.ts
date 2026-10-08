import { useEffect, useState } from "react";

export interface VisualViewportSize {
  height: number;
  offsetTop: number;
}

function readViewport(): VisualViewportSize {
  const viewport = window.visualViewport;
  if (viewport) {
    return { height: viewport.height, offsetTop: viewport.offsetTop };
  }
  return { height: window.innerHeight, offsetTop: 0 };
}

export function useVisualViewport(enabled = true): VisualViewportSize {
  const [size, setSize] = useState<VisualViewportSize>(readViewport);

  useEffect(() => {
    if (!enabled) {
      return undefined;
    }
    const update = () => {
      const next = readViewport();
      setSize((prev) =>
        prev.height === next.height && prev.offsetTop === next.offsetTop ? prev : next,
      );
    };
    const viewport = window.visualViewport;
    viewport?.addEventListener("resize", update);
    viewport?.addEventListener("scroll", update);
    window.addEventListener("resize", update);
    update();
    return () => {
      viewport?.removeEventListener("resize", update);
      viewport?.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
    };
  }, [enabled]);

  return size;
}
