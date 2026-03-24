import { useMemo, useRef, useEffect, useState } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";
import { usePlanetFocus } from "../../../contexts/PlanetFocusContext";
import {
  ORBITAL_STATION_ORBIT_RADIUS,
  ORBITAL_STATION_ORBIT_SPEED,
  ORBITAL_STATION_TILT,
} from "../board/boardConstants";
import {
  getPlanetCenter,
  getPlanetCameraTargetOffset,
  getPlanetCameraDefaultSpherical,
  getPlanetOrbit,
  getMarsOrbitalPosition,
} from "../board/solarSystemConfig";
import { audioService } from "../../../services/audioService";

export const panState = { isPanning: false, hasDragged: false };

const ORBITAL_STATION_ORBIT_CONFIG = { minDistance: 0.3, maxDistance: 5, defaultRadius: 0.8 };

function getOrbitalStationPosition(elapsedTime: number): THREE.Vector3 {
  const angle = elapsedTime * ORBITAL_STATION_ORBIT_SPEED;
  const r = ORBITAL_STATION_ORBIT_RADIUS;
  const tiltY = Math.sin(ORBITAL_STATION_TILT) * r * 0.3;
  const marsPos = getMarsOrbitalPosition(elapsedTime);
  return new THREE.Vector3(
    marsPos[0] + Math.cos(angle) * r,
    marsPos[1] + Math.sin(angle) * tiltY,
    marsPos[2] + Math.sin(angle) * r,
  );
}

function computePlanetCenter(
  planet: string,
  elapsedTime: number,
  toSun: THREE.Vector3,
  quat: THREE.Quaternion,
  zAxis: THREE.Vector3,
  camOffset: THREE.Vector3,
): THREE.Vector3 {
  let center: THREE.Vector3;
  if (planet === "orbital-station") {
    center = getOrbitalStationPosition(elapsedTime);
  } else if (planet === "solar-system") {
    center = new THREE.Vector3(0, 0, 0);
  } else {
    center = getPlanetCenter(planet, elapsedTime);
  }

  const offset = getPlanetCameraTargetOffset(planet);
  if (offset && planet !== "solar-system") {
    toSun.copy(center).negate().normalize();
    quat.setFromUnitVectors(zAxis, toSun);
    camOffset.set(offset[0], offset[1], offset[2]);
    camOffset.applyQuaternion(quat);
    center.add(camOffset);
  }

  return center;
}

function sphericalToWorldOffset(
  sph: THREE.Spherical,
  planet: string,
  center: THREE.Vector3,
  toSun: THREE.Vector3,
  quat: THREE.Quaternion,
  zAxis: THREE.Vector3,
  out: THREE.Vector3,
): THREE.Vector3 {
  out.setFromSpherical(sph);
  if (planet !== "solar-system") {
    toSun.copy(center).negate().normalize();
    quat.setFromUnitVectors(zAxis, toSun);
    out.applyQuaternion(quat);
  }
  return out;
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
  const { activePlanet } = usePlanetFocus();
  const isPointerDown = useRef(false);
  const previousPointer = useRef({ x: 0, y: 0 });
  const pointerDownOrigin = useRef({ x: 0, y: 0 });
  const [shouldRecenter, setShouldRecenter] = useState(false);
  const previousSize = useRef({ width: size.width, height: size.height });

  const marsInitPos = getMarsOrbitalPosition(0);
  const marsOrbit = getPlanetOrbit("mars");

  const [spherical] = useState(() => {
    return new THREE.Spherical(marsOrbit.defaultRadius, Math.PI / 2, 0);
  });
  const targetSpherical = useRef(new THREE.Spherical(marsOrbit.defaultRadius, Math.PI / 2, 0));

  const orbitCenter = useRef(new THREE.Vector3(marsInitPos[0], marsInitPos[1], marsInitPos[2]));
  const lastPlanetRef = useRef(activePlanet);
  const activePlanetRef = useRef(activePlanet);
  activePlanetRef.current = activePlanet;

  const wasFreeCameraEnabled = useRef(settings.freeCameraEnabled);

  // Reusable temp vectors
  const _toSun = useMemo(() => new THREE.Vector3(), []);
  const _quat = useMemo(() => new THREE.Quaternion(), []);
  const _zAxis = useMemo(() => new THREE.Vector3(0, 0, 1), []);
  const _offset = useMemo(() => new THREE.Vector3(), []);
  const _camOffset = useMemo(() => new THREE.Vector3(), []);

  useFrame((state) => {
    // --- Free camera toggle ---
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

    // --- Detect planet change → instant teleport ---
    if (activePlanet !== lastPlanetRef.current) {
      void audioService.playTravelSound();
      lastPlanetRef.current = activePlanet;

      const orbit =
        activePlanet === "orbital-station"
          ? ORBITAL_STATION_ORBIT_CONFIG
          : getPlanetOrbit(activePlanet);
      const defaultSpherical = getPlanetCameraDefaultSpherical(activePlanet);
      targetSpherical.current.radius = defaultSpherical?.radius ?? orbit.defaultRadius;
      targetSpherical.current.phi =
        activePlanet === "solar-system" ? 1.281 : (defaultSpherical?.phi ?? Math.PI / 2);
      targetSpherical.current.theta = defaultSpherical?.theta ?? 0;

      // Snap spherical immediately
      spherical.radius = targetSpherical.current.radius;
      spherical.phi = targetSpherical.current.phi;
      spherical.theta = targetSpherical.current.theta;

      // Snap orbit center
      const center = computePlanetCenter(
        activePlanet,
        state.clock.elapsedTime,
        _toSun,
        _quat,
        _zAxis,
        _camOffset,
      );
      orbitCenter.current.copy(center);

      // Snap camera position
      sphericalToWorldOffset(spherical, activePlanet, center, _toSun, _quat, _zAxis, _offset);
      camera.position.copy(center).add(_offset);
      camera.lookAt(center);
    }

    // --- Steady state: follow orbiting planet + apply pan/zoom ---
    const targetCenter = computePlanetCenter(
      activePlanet,
      state.clock.elapsedTime,
      _toSun,
      _quat,
      _zAxis,
      _camOffset,
    );
    orbitCenter.current.copy(targetCenter);

    if (size.width !== previousSize.current.width || size.height !== previousSize.current.height) {
      setShouldRecenter(true);
      previousSize.current = { width: size.width, height: size.height };
    }

    if (shouldRecenter) {
      targetSpherical.current.theta = 0;
      targetSpherical.current.phi = Math.PI / 2;
      setShouldRecenter(false);
    }

    const radiusDelta = Math.abs(targetSpherical.current.radius - spherical.radius);
    const panLerp = radiusDelta > 50 ? 0.02 : 0.1;

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

    sphericalToWorldOffset(
      spherical,
      activePlanet,
      orbitCenter.current,
      _toSun,
      _quat,
      _zAxis,
      _offset,
    );
    camera.position.copy(orbitCenter.current).add(_offset);
    camera.lookAt(orbitCenter.current);

    cameraStateRef.current = {
      position: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
      rotation: { x: camera.rotation.x, y: camera.rotation.y, z: camera.rotation.z },
    };
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
      if (settings.freeCameraEnabled) {
        return;
      }
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
      if (!isPointerDown.current) {
        return;
      }

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

      targetSpherical.current.theta -= deltaX * orbitSpeed;
      targetSpherical.current.phi -= deltaY * orbitSpeed;

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
      if (settings.freeCameraEnabled) {
        return;
      }
      event.preventDefault();
      const planet = activePlanetRef.current;
      if (planet === "solar-system") {
        return;
      }
      const orbit =
        planet === "orbital-station" ? ORBITAL_STATION_ORBIT_CONFIG : getPlanetOrbit(planet);
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
