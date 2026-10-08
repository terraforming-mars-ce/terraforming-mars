import * as THREE from "three";
import type { PlanetSurfaceKey } from "../../../hooks/useTextures";

export interface MoonConfig {
  id: string;
  name: string;
  radius: number;
  position: [number, number, number];
  renderType: "sphere" | "glb";
  atmosphere: AtmosphereProfile;
  textureKey?: PlanetSurfaceKey;
  modelKey?: string;
  tileLocation?: string;
  coordOffset: { q: number; r: number; s: number };
}

export interface AtmosphereProfile {
  color: string;
  shadowColor: string;
  thickness: number;
  intensity: number;
  halo: number;
}

export interface PlanetConfig {
  id: string;
  name: string;
  radius: number;
  textureKey: PlanetSurfaceKey;
  cloudTextureKey?: PlanetSurfaceKey;
  atmosphere: AtmosphereProfile;
  tileLocation: string;
  coordOffset: { q: number; r: number; s: number };
  moons: MoonConfig[];
  cameraTargetOffset?: [number, number, number];
  cameraDefaultSpherical?: { radius: number; phi: number; theta: number };
  orbit: { minDistance: number; maxDistance: number; defaultRadius: number };
  orbitRadius: number;
  orbitAngle: number;
  orbitPeriod: number;
}

export function orbitPos(radius: number, angleDeg: number): [number, number, number] {
  const rad = (angleDeg * Math.PI) / 180;
  return [Math.cos(rad) * radius, 0, Math.sin(rad) * radius];
}

export function getPlanetOrbitalAngle(
  config: { orbitAngle: number; orbitPeriod: number },
  elapsedTime: number,
): number {
  return config.orbitAngle + ((elapsedTime * orbitSpeedMultiplier) / config.orbitPeriod) * 360;
}

export function getPlanetOrbitalPosition(
  config: { orbitRadius: number; orbitAngle: number; orbitPeriod: number },
  elapsedTime: number,
): [number, number, number] {
  const angle = getPlanetOrbitalAngle(config, elapsedTime);
  return orbitPos(config.orbitRadius, angle);
}

const ROCKY_HAZE: AtmosphereProfile = {
  color: "#cec3ad",
  shadowColor: "#6583a5",
  thickness: 0.055,
  intensity: 1.1,
  halo: 0.9,
};

export const PLANET_CONFIGS: PlanetConfig[] = [
  {
    id: "mercury",
    atmosphere: ROCKY_HAZE,
    name: "MERCURY",
    radius: 1.2,
    textureKey: "mercury",
    tileLocation: "mercury",
    coordOffset: { q: 400, r: 0, s: -400 },
    moons: [],
    orbit: { minDistance: 2.0, maxDistance: 15, defaultRadius: 6 },
    orbitRadius: 80,
    orbitAngle: 45,
    orbitPeriod: 340,
  },
  {
    id: "venus",
    atmosphere: {
      color: "#f4d7a0",
      shadowColor: "#857ba0",
      thickness: 0.06,
      intensity: 1.25,
      halo: 1,
    },
    name: "VENUS",
    radius: 2.02,
    textureKey: "venus",
    tileLocation: "venus",
    coordOffset: { q: 100, r: 0, s: -100 },
    moons: [],
    orbit: { minDistance: 3.0, maxDistance: 20, defaultRadius: 8 },
    orbitRadius: 140,
    orbitAngle: 120,
    orbitPeriod: 550,
  },
  {
    id: "earth",
    atmosphere: {
      color: "#b1d9ff",
      shadowColor: "#2870b3",
      thickness: 0.055,
      intensity: 1.1,
      halo: 1,
    },
    name: "EARTH",
    radius: 2.02,
    textureKey: "earth",
    cloudTextureKey: "earthClouds",
    tileLocation: "earth",
    coordOffset: { q: 0, r: 0, s: 0 },
    moons: [
      {
        id: "luna",
        name: "Moon",
        radius: 0.27,
        position: [3.5, 0.4, 0.8],
        renderType: "sphere",
        atmosphere: ROCKY_HAZE,
        textureKey: "moon",
        tileLocation: "luna",
        coordOffset: { q: 300, r: 0, s: -300 },
      },
    ],
    cameraTargetOffset: [2.5, 0, 0],
    cameraDefaultSpherical: { radius: 4.7, phi: 1.4184, theta: 0.3586 },
    orbit: { minDistance: 1.5, maxDistance: 10, defaultRadius: 3.5 },
    orbitRadius: 200,
    orbitAngle: 200,
    orbitPeriod: 700,
  },
  {
    id: "ceres",
    atmosphere: ROCKY_HAZE,
    name: "CERES",
    radius: 0.28,
    textureKey: "ceres",
    tileLocation: "ceres",
    coordOffset: { q: 0, r: 0, s: 0 },
    moons: [],
    orbit: { minDistance: 0.5, maxDistance: 5, defaultRadius: 1.5 },
    orbitRadius: 320,
    orbitAngle: 0,
    orbitPeriod: 1500,
  },
  {
    id: "jupiter",
    atmosphere: {
      color: "#d9c9aa",
      shadowColor: "#627d96",
      thickness: 0.065,
      intensity: 1.5,
      halo: 1,
    },
    name: "JUPITER",
    radius: 11.3,
    textureKey: "jupiter",
    tileLocation: "jupiter",
    coordOffset: { q: 0, r: 0, s: 0 },
    moons: [
      {
        id: "ganymede",
        name: "Ganymede",
        radius: 0.5,
        position: [16, 0.5, 1.0],
        renderType: "sphere",
        atmosphere: ROCKY_HAZE,
        textureKey: "ganymede",
        tileLocation: "ganymede",
        coordOffset: { q: 200, r: 0, s: -200 },
      },
    ],
    cameraTargetOffset: [12, 0, 0],
    cameraDefaultSpherical: { radius: 8.6, phi: 1.5714, theta: 0.689 },
    orbit: { minDistance: 5, maxDistance: 40, defaultRadius: 14 },
    orbitRadius: 400,
    orbitAngle: 330,
    orbitPeriod: 2400,
  },
  {
    id: "saturn",
    atmosphere: {
      color: "#e6d5ac",
      shadowColor: "#718795",
      thickness: 0.065,
      intensity: 1.5,
      halo: 1,
    },
    name: "SATURN",
    radius: 9.5,
    textureKey: "saturn",
    tileLocation: "saturn",
    coordOffset: { q: 0, r: 0, s: 0 },
    moons: [],
    orbit: { minDistance: 12, maxDistance: 70, defaultRadius: 30 },
    orbitRadius: 550,
    orbitAngle: 15,
    orbitPeriod: 3800,
  },
  {
    id: "uranus",
    atmosphere: {
      color: "#90d5dc",
      shadowColor: "#347894",
      thickness: 0.06,
      intensity: 1.35,
      halo: 1,
    },
    name: "URANUS",
    radius: 8.1,
    textureKey: "uranus",
    tileLocation: "uranus",
    coordOffset: { q: 0, r: 0, s: 0 },
    moons: [],
    orbit: { minDistance: 10, maxDistance: 60, defaultRadius: 25 },
    orbitRadius: 700,
    orbitAngle: 160,
    orbitPeriod: 6400,
  },
  {
    id: "neptune",
    atmosphere: {
      color: "#719ee8",
      shadowColor: "#234f9a",
      thickness: 0.065,
      intensity: 1.5,
      halo: 1,
    },
    name: "NEPTUNE",
    radius: 7.9,
    textureKey: "neptune",
    tileLocation: "neptune",
    coordOffset: { q: 0, r: 0, s: 0 },
    moons: [],
    orbit: { minDistance: 10, maxDistance: 60, defaultRadius: 25 },
    orbitRadius: 850,
    orbitAngle: 80,
    orbitPeriod: 9000,
  },
];

export const MARS_ORBIT_RADIUS = 250;
export const MARS_ORBIT_ANGLE = 270;
export const MARS_ORBIT_PERIOD = 960;

let orbitSpeedMultiplier = 1.0;
export function setOrbitSpeedMultiplier(value: number) {
  orbitSpeedMultiplier = value;
}

export function getMarsOrbitalAngle(elapsedTime: number): number {
  return MARS_ORBIT_ANGLE + ((elapsedTime * orbitSpeedMultiplier) / MARS_ORBIT_PERIOD) * 360;
}

export function getMarsOrbitalPosition(elapsedTime: number): [number, number, number] {
  return orbitPos(MARS_ORBIT_RADIUS, getMarsOrbitalAngle(elapsedTime));
}

export const PHOBOS_CONFIG: MoonConfig = {
  id: "phobos",
  name: "Phobos",
  radius: 0.35,
  position: [3.2, 0.3, 0.5],
  renderType: "glb",
  atmosphere: { ...ROCKY_HAZE, thickness: 0.075, intensity: 1.1, halo: 1.2 },
  modelKey: "phobos",
  tileLocation: "phobos",
  coordOffset: { q: 500, r: 0, s: -500 },
};

export const MARS_ATMOSPHERE: AtmosphereProfile = {
  color: "#efd2a3",
  shadowColor: "#4b8cb5",
  thickness: 0.055,
  intensity: 1.1,
  halo: 1,
};

// Mars's sky as it warms: a thin cold blue limb, a salmon haze, then a thick warm red-orange.
export const MARS_CLIMATE_ATMOSPHERES: AtmosphereProfile[] = [
  { color: "#c3cddb", shadowColor: "#56708f", thickness: 0.045, intensity: 0.85, halo: 0.9 },
  { color: "#f0b9a2", shadowColor: "#5b7d9c", thickness: 0.055, intensity: 1.1, halo: 1 },
  { color: "#ec9468", shadowColor: "#7a4636", thickness: 0.07, intensity: 1.25, halo: 1.1 },
];

export const PLANET_FILL_LIGHT = {
  keyColor: "#fff1d4",
  keyIntensityRatio: 3.2,
  skyColor: "#62758b",
  groundColor: "#0a2442",
  intensityRatio: 0.35,
};

export const MARS_ORBIT = { minDistance: 2.4, maxDistance: 20, defaultRadius: 7 };

export const SOLAR_SYSTEM_ORBIT = {
  minDistance: 300,
  maxDistance: 2500,
  defaultRadius: 1588,
};

export const LOCATION_TO_PLANET: Record<string, string> = {
  mars: "mars",
  venus: "venus",
  jupiter: "jupiter",
  ganymede: "jupiter",
  earth: "earth",
  luna: "earth",
  mercury: "mercury",
  phobos: "mars",
  saturn: "saturn",
  neptune: "neptune",
  uranus: "uranus",
  ceres: "ceres",
};

export function getOrbitalAngleRad(planetId: string, elapsedTime: number): number {
  if (planetId === "mars" || planetId === "orbital-station") {
    return (getMarsOrbitalAngle(elapsedTime) * Math.PI) / 180;
  }
  const config = getPlanetConfig(planetId);
  if (config) {
    return (getPlanetOrbitalAngle(config, elapsedTime) * Math.PI) / 180;
  }
  return 0;
}

export function getPlanetConfig(planetId: string): PlanetConfig | undefined {
  return PLANET_CONFIGS.find((p) => p.id === planetId);
}

export function getPlanetCenter(planetId: string, elapsedTime: number): THREE.Vector3 {
  if (planetId === "solar-system") {
    return new THREE.Vector3(0, 0, 0);
  }
  if (planetId === "mars") {
    const pos = getMarsOrbitalPosition(elapsedTime);
    return new THREE.Vector3(...pos);
  }
  const config = getPlanetConfig(planetId);
  if (config) {
    const pos = getPlanetOrbitalPosition(config, elapsedTime);
    return new THREE.Vector3(...pos);
  }
  return new THREE.Vector3(0, 0, 0);
}

export function getPlanetCameraTargetOffset(
  planetId: string,
): [number, number, number] | undefined {
  const config = getPlanetConfig(planetId);
  return config?.cameraTargetOffset;
}

export function getPlanetCameraDefaultSpherical(
  planetId: string,
): { radius: number; phi: number; theta: number } | undefined {
  const config = getPlanetConfig(planetId);
  return config?.cameraDefaultSpherical;
}

export function getPlanetOrbit(planetId: string): {
  minDistance: number;
  maxDistance: number;
  defaultRadius: number;
} {
  if (planetId === "mars") {
    return MARS_ORBIT;
  }
  if (planetId === "solar-system") {
    return SOLAR_SYSTEM_ORBIT;
  }
  const config = getPlanetConfig(planetId);
  if (config) {
    return config.orbit;
  }
  return MARS_ORBIT;
}
