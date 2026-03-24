import * as THREE from "three";

export interface MoonConfig {
  id: string;
  name: string;
  radius: number;
  position: [number, number, number];
  renderType: "sphere" | "glb";
  textureKey?: string;
  modelKey?: string;
  tileLocation?: string;
  coordOffset: { q: number; r: number; s: number };
}

export interface PlanetConfig {
  id: string;
  name: string;
  radius: number;
  textureKey: string;
  cloudTextureKey?: string;
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
  return config.orbitAngle + (elapsedTime / config.orbitPeriod) * 360;
}

export function getPlanetOrbitalPosition(
  config: { orbitRadius: number; orbitAngle: number; orbitPeriod: number },
  elapsedTime: number,
): [number, number, number] {
  const angle = getPlanetOrbitalAngle(config, elapsedTime);
  return orbitPos(config.orbitRadius, angle);
}

export const PLANET_CONFIGS: PlanetConfig[] = [
  {
    id: "mercury",
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

export function getMarsOrbitalAngle(elapsedTime: number): number {
  return MARS_ORBIT_ANGLE + (elapsedTime / MARS_ORBIT_PERIOD) * 360;
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
  modelKey: "phobos",
  tileLocation: "phobos",
  coordOffset: { q: 500, r: 0, s: -500 },
};

export const MARS_ORBIT = { minDistance: 2.4, maxDistance: 20, defaultRadius: 8 };

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
