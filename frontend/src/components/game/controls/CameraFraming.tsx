import { useLayoutEffect, useMemo } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useHudInsets, type HudInsets } from "@/hooks/useHudInsets.ts";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";

export const DESKTOP_FOV = 50;
const MAX_FOV = 100;
// Mars at its default distance keeps the share of the shorter visible axis it has on a 16:9 desktop at 50°
const REQUIRED_HALF_FOV_TAN = Math.tan(THREE.MathUtils.degToRad(DESKTOP_FOV / 2));
// Phones show Mars larger so it fills most of the height left between the HUD bars
const COMPACT_MARS_SCALE = 1.22;

export interface CameraFramingState {
  fov: number;
  offsetX: number;
  offsetY: number;
  availableWidth: number;
  availableHeight: number;
}

export function computeCameraFraming(
  width: number,
  height: number,
  insets: HudInsets,
  isCompact: boolean,
): CameraFramingState {
  const availableWidth = Math.max(width - insets.left - insets.right, 1);
  const availableHeight = Math.max(height - insets.top - insets.bottom, 1);
  const requiredTan = isCompact
    ? REQUIRED_HALF_FOV_TAN / COMPACT_MARS_SCALE
    : REQUIRED_HALF_FOV_TAN;
  const tanForHeight = (requiredTan * height) / availableHeight;
  const tanForWidth = (requiredTan * height) / availableWidth;
  const halfFovTan = Math.max(tanForHeight, tanForWidth);
  const fov = Math.min(THREE.MathUtils.radToDeg(Math.atan(halfFovTan)) * 2, MAX_FOV);
  return {
    fov,
    offsetX: (insets.left - insets.right) / 2,
    offsetY: (insets.top - insets.bottom) / 2,
    availableWidth,
    availableHeight,
  };
}

export function useCameraFraming(): CameraFramingState {
  const width = useThree((s) => s.size.width);
  const height = useThree((s) => s.size.height);
  const insets = useHudInsets();
  const { isCompact } = useLayoutMode();
  return useMemo(
    () => computeCameraFraming(width, height, insets, isCompact),
    [width, height, insets, isCompact],
  );
}

export function CameraFraming() {
  const camera = useThree((s) => s.camera);
  const width = useThree((s) => s.size.width);
  const height = useThree((s) => s.size.height);
  const framing = useCameraFraming();

  useLayoutEffect(() => {
    if (!(camera instanceof THREE.PerspectiveCamera)) {
      return;
    }
    camera.fov = framing.fov;
    if (framing.offsetX === 0 && framing.offsetY === 0) {
      camera.clearViewOffset();
    } else {
      camera.setViewOffset(width, height, -framing.offsetX, -framing.offsetY, width, height);
    }
    camera.updateProjectionMatrix();
  }, [camera, framing, width, height]);

  return null;
}
