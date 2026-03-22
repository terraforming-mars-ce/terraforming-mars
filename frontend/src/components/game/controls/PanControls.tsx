import { useRef, useEffect, useState } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import {
  VENUS_POSITION,
  ORBITAL_STATION_ORBIT_RADIUS,
  ORBITAL_STATION_ORBIT_SPEED,
  ORBITAL_STATION_TILT,
} from "../board/boardConstants";

export const panState = { isPanning: false, hasDragged: false };

const MARS_CENTER = new THREE.Vector3(0, 0, 0);
const VENUS_CENTER = new THREE.Vector3(VENUS_POSITION[0], VENUS_POSITION[1], VENUS_POSITION[2]);

const MARS_ORBIT = { minDistance: 2.4, maxDistance: 20, defaultRadius: 8 };
const VENUS_ORBIT = { minDistance: 2.4, maxDistance: 20, defaultRadius: 6.5 };
const ORBITAL_STATION_ORBIT = { minDistance: 0.3, maxDistance: 5, defaultRadius: 0.8 };

function getOrbitalStationPosition(elapsedTime: number): THREE.Vector3 {
  const angle = elapsedTime * ORBITAL_STATION_ORBIT_SPEED;
  const r = ORBITAL_STATION_ORBIT_RADIUS;
  const tiltY = Math.sin(ORBITAL_STATION_TILT) * r * 0.3;
  return new THREE.Vector3(Math.cos(angle) * r, Math.sin(angle) * tiltY, Math.sin(angle) * r);
}

export function PanControls() {
  const { camera, gl, size } = useThree();
  const {
    settings,
    storedCameraState,
    setStoredCameraState,
    cameraStateRef,
    pendingCameraTransformRef,
  } = useWorld3DSettings();
  const { activePlanet, isTransitioning, setIsTransitioning } = usePlanetFocus();
  const isPointerDown = useRef(false);
  const previousPointer = useRef({ x: 0, y: 0 });
  const pointerDownOrigin = useRef({ x: 0, y: 0 });
  const [shouldRecenter, setShouldRecenter] = useState(false);
  const previousSize = useRef({ width: size.width, height: size.height });

  const [spherical] = useState(() => {
    const s = new THREE.Spherical();
    s.setFromVector3(camera.position);
    if (s.radius < 3) s.radius = 8;
    return s;
  });

  const targetSpherical = useRef(
    new THREE.Spherical(spherical.radius, spherical.phi, spherical.theta),
  );

  const orbitCenter = useRef(new THREE.Vector3(0, 0, 0));
  const targetOrbitCenter = useRef(new THREE.Vector3(0, 0, 0));
  const previousPlanet = useRef(activePlanet);
  const activePlanetRef = useRef(activePlanet);
  const isTransitioningRef = useRef(isTransitioning);
  activePlanetRef.current = activePlanet;
  isTransitioningRef.current = isTransitioning;

  const wasFreeCameraEnabled = useRef(settings.freeCameraEnabled);

  useFrame((state) => {
    // Handle free camera toggle
    if (settings.freeCameraEnabled && !wasFreeCameraEnabled.current) {
      setStoredCameraState({
        position: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
        spherical: { radius: spherical.radius, phi: spherical.phi, theta: spherical.theta },
      });
      wasFreeCameraEnabled.current = true;
    } else if (!settings.freeCameraEnabled && wasFreeCameraEnabled.current) {
      if (storedCameraState) {
        spherical.radius = storedCameraState.spherical.radius;
        spherical.phi = storedCameraState.spherical.phi;
        spherical.theta = storedCameraState.spherical.theta;
        targetSpherical.current.radius = storedCameraState.spherical.radius;
        targetSpherical.current.phi = storedCameraState.spherical.phi;
        targetSpherical.current.theta = storedCameraState.spherical.theta;
      }
      wasFreeCameraEnabled.current = false;
    }

    if (settings.freeCameraEnabled) {
      cameraStateRef.current = {
        position: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
        rotation: { x: camera.rotation.x, y: camera.rotation.y, z: camera.rotation.z },
      };
      return;
    }

    // Detect planet change and start transition
    if (activePlanet !== previousPlanet.current) {
      previousPlanet.current = activePlanet;
      setIsTransitioning(true);

      let orbit = MARS_ORBIT;
      if (activePlanet === "venus") {
        orbit = VENUS_ORBIT;
        targetOrbitCenter.current.copy(VENUS_CENTER);
      } else if (activePlanet === "orbital-station") {
        orbit = ORBITAL_STATION_ORBIT;
      } else {
        targetOrbitCenter.current.copy(MARS_CENTER);
      }
      targetSpherical.current.radius = orbit.defaultRadius;
      targetSpherical.current.phi = Math.PI / 2;
      targetSpherical.current.theta = 0;
    }

    // Continuously track orbital station position since it moves
    if (activePlanet === "orbital-station") {
      targetOrbitCenter.current.copy(getOrbitalStationPosition(state.clock.elapsedTime));
    }

    if (size.width !== previousSize.current.width || size.height !== previousSize.current.height) {
      setShouldRecenter(true);
      previousSize.current = { width: size.width, height: size.height };
    }

    if (shouldRecenter) {
      targetSpherical.current.theta = 0;
      targetSpherical.current.phi = Math.PI / 2;
      setShouldRecenter(false);
    }

    const isOrbitalStation = activePlanet === "orbital-station";
    const travelLerp = isTransitioning ? 0.03 : isOrbitalStation ? 0.15 : 0.1;
    const panLerp = 0.1;

    orbitCenter.current.lerp(targetOrbitCenter.current, travelLerp);

    spherical.theta += (targetSpherical.current.theta - spherical.theta) * panLerp;
    spherical.phi += (targetSpherical.current.phi - spherical.phi) * panLerp;
    spherical.radius += (targetSpherical.current.radius - spherical.radius) * panLerp;

    const pending = pendingCameraTransformRef.current;
    if (pending) {
      if (pending.position) {
        camera.position.set(pending.position.x, pending.position.y, pending.position.z);
        const newSpherical = new THREE.Spherical().setFromVector3(
          camera.position.clone().sub(orbitCenter.current),
        );
        spherical.radius = newSpherical.radius;
        spherical.phi = newSpherical.phi;
        spherical.theta = newSpherical.theta;
        targetSpherical.current.radius = newSpherical.radius;
        targetSpherical.current.phi = newSpherical.phi;
        targetSpherical.current.theta = newSpherical.theta;
      }
      pendingCameraTransformRef.current = null;
    }

    const offset = new THREE.Vector3().setFromSpherical(spherical);
    camera.position.copy(orbitCenter.current).add(offset);
    camera.lookAt(orbitCenter.current);

    cameraStateRef.current = {
      position: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
      rotation: { x: camera.rotation.x, y: camera.rotation.y, z: camera.rotation.z },
    };

    if (isTransitioning) {
      const centerDist = orbitCenter.current.distanceTo(targetOrbitCenter.current);
      if (centerDist < 0.05) {
        setIsTransitioning(false);
      }
    }
  });

  useEffect(() => {
    if (!settings.freeCameraEnabled) {
      const offset = new THREE.Vector3().setFromSpherical(spherical);
      camera.position.copy(orbitCenter.current).add(offset);
      camera.lookAt(orbitCenter.current);
    }

    const handleWindowResize = () => {
      setShouldRecenter(true);
    };

    const handlePointerDown = (event: PointerEvent) => {
      if (settings.freeCameraEnabled) return;
      isPointerDown.current = true;
      panState.isPanning = true;
      panState.hasDragged = false;
      previousPointer.current = { x: event.clientX, y: event.clientY };
      pointerDownOrigin.current = { x: event.clientX, y: event.clientY };
      gl.domElement.style.cursor = "grabbing";

      document.addEventListener("pointermove", handlePointerMove);
      document.addEventListener("pointerup", handlePointerUp);
    };

    const handlePointerMove = (event: PointerEvent) => {
      if (!isPointerDown.current) return;

      if (!panState.hasDragged) {
        const dx = event.clientX - pointerDownOrigin.current.x;
        const dy = event.clientY - pointerDownOrigin.current.y;
        if (dx * dx + dy * dy > 9) {
          panState.hasDragged = true;
        }
      }

      const deltaX = event.clientX - previousPointer.current.x;
      const deltaY = event.clientY - previousPointer.current.y;

      const orbitSpeed = 0.0003;
      const maxAngle = Math.PI / 4;
      const equator = Math.PI / 2;

      targetSpherical.current.theta -= deltaX * orbitSpeed;
      targetSpherical.current.phi -= deltaY * orbitSpeed;

      targetSpherical.current.theta = Math.max(
        -maxAngle,
        Math.min(maxAngle, targetSpherical.current.theta),
      );
      targetSpherical.current.phi = Math.max(
        equator - maxAngle,
        Math.min(equator + maxAngle, targetSpherical.current.phi),
      );

      previousPointer.current = { x: event.clientX, y: event.clientY };
    };

    const handlePointerUp = () => {
      isPointerDown.current = false;
      panState.isPanning = false;
      gl.domElement.style.cursor = "grab";

      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
    };

    const handleWheel = (event: WheelEvent) => {
      if (settings.freeCameraEnabled) return;
      event.preventDefault();
      const planet = activePlanetRef.current;
      const orbit =
        planet === "venus"
          ? VENUS_ORBIT
          : planet === "orbital-station"
            ? ORBITAL_STATION_ORBIT
            : MARS_ORBIT;
      const zoomSpeed = 0.5;
      const zoomDelta = event.deltaY * zoomSpeed * 0.01;

      targetSpherical.current.radius += zoomDelta;
      targetSpherical.current.radius = Math.max(
        orbit.minDistance,
        Math.min(orbit.maxDistance, targetSpherical.current.radius),
      );
    };

    const domElement = gl.domElement;

    if (!settings.freeCameraEnabled) {
      domElement.style.cursor = "grab";
    }

    domElement.addEventListener("pointerdown", handlePointerDown);
    domElement.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("resize", handleWindowResize);

    return () => {
      domElement.removeEventListener("pointerdown", handlePointerDown);
      domElement.removeEventListener("wheel", handleWheel);

      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("resize", handleWindowResize);
    };
  }, [camera, gl, spherical, settings.freeCameraEnabled]);

  return null;
}
