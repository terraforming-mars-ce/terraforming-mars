import * as THREE from "three";
import {
  ORBITAL_STATION_ORBIT_RADIUS,
  ORBITAL_STATION_ORBIT_SPEED,
  ORBITAL_STATION_TILT,
  SPHERE_RADIUS,
} from "../board/boardConstants";
import {
  getPlanetCenter,
  getPlanetCameraTargetOffset,
  getPlanetOrbit,
  getMarsOrbitalPosition,
  getPlanetConfig,
} from "../board/solarSystemConfig";

export const ORBITAL_STATION_ORBIT_CONFIG = {
  minDistance: 0.3,
  maxDistance: 5,
  defaultRadius: 0.8,
};
export const MIN_POLAR_ANGLE = Math.PI / 36;
export const MAX_POLAR_ANGLE = Math.PI - MIN_POLAR_ANGLE;
const MIN_SURFACE_DISTANCE_RATIO = 0.1;
const Z_AXIS = new THREE.Vector3(0, 0, 1);

export function getOrbitLimits(planet: string) {
  return planet === "orbital-station" ? ORBITAL_STATION_ORBIT_CONFIG : getPlanetOrbit(planet);
}

export function getBodyRadius(planet: string): number {
  if (planet === "mars") {
    return SPHERE_RADIUS;
  }
  return getPlanetConfig(planet)?.radius ?? 0;
}

export function getSurfaceDistance(radius: number, bodyRadius: number): number {
  return Math.max(radius - bodyRadius, radius * MIN_SURFACE_DISTANCE_RATIO);
}

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

export function computePlanetCenter(
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

export function sphericalToWorldOffset(
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

export function worldToOrbitFrame(
  planet: string,
  elapsedTime: number,
  worldPoints: readonly THREE.Vector3[],
): THREE.Vector3[] {
  const toSun = new THREE.Vector3();
  const quat = new THREE.Quaternion();
  const center = computePlanetCenter(planet, elapsedTime, toSun, quat, Z_AXIS, new THREE.Vector3());
  const inverse = new THREE.Quaternion();
  if (planet !== "solar-system") {
    toSun.copy(center).negate().normalize();
    inverse.setFromUnitVectors(Z_AXIS, toSun).invert();
  }
  return worldPoints.map((point) => point.clone().sub(center).applyQuaternion(inverse));
}
