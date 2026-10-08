import { useEffect, useRef, type RefObject } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { usePlanetFocus } from "@/contexts/PlanetFocusContext.tsx";
import { useLayoutMode } from "@/hooks/useLayoutMode.ts";
import { useCameraFocusStore, type CameraFocusRequest } from "@/stores/cameraFocusStore.ts";
import { useCameraFraming, type CameraFramingState } from "./CameraFraming";
import { getOrbitLimits, worldToOrbitFrame } from "./orbitFrame";

const FOCUS_MARGIN = 0.8;
const HEX_EXTENT = 0.17;

export function computeFocusRequest(
  planet: string,
  elapsedTime: number,
  worldPoints: readonly THREE.Vector3[],
  framing: CameraFramingState,
  viewportHeight: number,
): CameraFocusRequest | null {
  const points = worldToOrbitFrame(planet, elapsedTime, worldPoints);
  const direction = new THREE.Vector3();
  for (const point of points) {
    direction.add(point);
  }
  if (direction.lengthSq() < 1e-8) {
    return null;
  }
  direction.normalize();

  const halfFovTan = Math.tan(THREE.MathUtils.degToRad(framing.fov) / 2);
  const visibleShare = Math.min(framing.availableWidth, framing.availableHeight) / viewportHeight;
  const tanLimit = halfFovTan * visibleShare * FOCUS_MARGIN;
  if (tanLimit <= 0) {
    return null;
  }

  const lateral = new THREE.Vector3();
  let distance = 0;
  for (const point of points) {
    const depth = point.dot(direction);
    lateral.copy(direction).multiplyScalar(-depth).add(point);
    distance = Math.max(distance, depth + (lateral.length() + HEX_EXTENT) / tanLimit);
  }
  const limits = getOrbitLimits(planet);
  return {
    direction: [direction.x, direction.y, direction.z],
    distance: THREE.MathUtils.clamp(distance, limits.minDistance, limits.maxDistance),
  };
}

export function useAutoFramePlacement(
  localPoints: readonly THREE.Vector3[],
  root: RefObject<THREE.Object3D | null>,
  planetId: string,
) {
  const { activePlanet } = usePlanetFocus();
  const { isCompact } = useLayoutMode();
  const framing = useCameraFraming();
  const clock = useThree((s) => s.clock);
  const viewportHeight = useThree((s) => s.size.height);
  const framedRef = useRef(false);
  const latest = useRef({ localPoints, framing, viewportHeight });
  latest.current = { localPoints, framing, viewportHeight };
  const hasPoints = localPoints.length > 0;

  useEffect(() => {
    if (!hasPoints) {
      framedRef.current = false;
      return;
    }
    const body = root.current;
    if (framedRef.current || !isCompact || activePlanet !== planetId || !body) {
      return;
    }
    framedRef.current = true;
    body.updateMatrixWorld(true);
    const current = latest.current;
    const worldPoints = current.localPoints.map((point) =>
      point.clone().applyMatrix4(body.matrixWorld),
    );
    const request = computeFocusRequest(
      planetId,
      clock.elapsedTime,
      worldPoints,
      current.framing,
      current.viewportHeight,
    );
    if (request) {
      useCameraFocusStore.getState().requestFocus(request);
    }
  }, [hasPoints, isCompact, activePlanet, planetId, root, clock]);
}
