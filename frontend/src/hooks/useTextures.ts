import { useMemo } from "react";
import { useTexture } from "@react-three/drei";
import { useLoader } from "@react-three/fiber";
import * as THREE from "three";

const TEXTURE_PATHS = {
  mars: "/assets/textures/mars_8k.jpg",
  venus: "/assets/textures/4k_venus_atmosphere.jpg",
  earth: "/assets/textures/earth_8k.jpg",
  earthClouds: "/assets/textures/earth_clouds_8k.jpg",
  jupiter: "/assets/textures/jupiter_8k.jpg",
  mercury: "/assets/textures/mercury_8k.jpg",
  saturn: "/assets/textures/saturn_8k.jpg",
  neptune: "/assets/textures/neptune_2k.jpg",
  uranus: "/assets/textures/uranus_2k.jpg",
  ceres: "/assets/textures/ceres_4k.jpg",
  moon: "/assets/textures/moon_8k.jpg",
  ganymede: "/assets/textures/ganymede_4k.png",
  sun: "/assets/textures/sun_8k.jpg",
  grass: "/assets/textures/grass.jpg",
  rock: "/assets/textures/rock.jpg",
  sand: "/assets/textures/sand.jpg",
  waterNormals: "/assets/textures/waternormals.jpg",
  noiseMid: "/assets/textures/noise_mid.png",
  noiseHigh: "/assets/textures/noise_high.png",
  smoke: "/assets/effects/smoke.png",
  concrete: "/assets/textures/concrete.jpg",
  marsLod: "/assets/textures/mars_8k_512.jpg",
  venusLod: "/assets/textures/4k_venus_atmosphere_512.jpg",
  earthLod: "/assets/textures/earth_8k_512.jpg",
  earthCloudsLod: "/assets/textures/earth_clouds_8k_512.jpg",
  jupiterLod: "/assets/textures/jupiter_8k_512.jpg",
  mercuryLod: "/assets/textures/mercury_8k_512.jpg",
  saturnLod: "/assets/textures/saturn_8k_512.jpg",
  neptuneLod: "/assets/textures/neptune_2k_512.jpg",
  uranusLod: "/assets/textures/uranus_2k_512.jpg",
  ceresLod: "/assets/textures/ceres_4k_512.jpg",
  moonLod: "/assets/textures/moon_8k_512.jpg",
  ganymedeLod: "/assets/textures/ganymede_4k_512.png",
  sunLod: "/assets/textures/sun_8k_512.jpg",
} as const;

const RESOURCE_ICON_PATHS = {
  steel: "/assets/resources/steel.png",
  titanium: "/assets/resources/titanium.png",
  plant: "/assets/resources/plant.png",
  megacredit: "/assets/resources/megacredit.png",
  card: "/assets/resources/card.png",
} as const;

type ResourceIconName = keyof typeof RESOURCE_ICON_PATHS;

const BONUS_TYPE_TO_ICON: Record<string, ResourceIconName> = {
  steel: "steel",
  titanium: "titanium",
  plants: "plant",
  plant: "plant",
  cards: "card",
  "card-draw": "card",
  credit: "megacredit",
};

// Module-level preloads
useTexture.preload(TEXTURE_PATHS.mars);
useTexture.preload(TEXTURE_PATHS.venus);
useTexture.preload(TEXTURE_PATHS.earth);
useTexture.preload(TEXTURE_PATHS.earthClouds);
useTexture.preload(TEXTURE_PATHS.jupiter);
useTexture.preload(TEXTURE_PATHS.mercury);
useTexture.preload(TEXTURE_PATHS.saturn);
useTexture.preload(TEXTURE_PATHS.neptune);
useTexture.preload(TEXTURE_PATHS.uranus);
useTexture.preload(TEXTURE_PATHS.ceres);
useTexture.preload(TEXTURE_PATHS.moon);
useTexture.preload(TEXTURE_PATHS.ganymede);
useTexture.preload(TEXTURE_PATHS.sun);
useTexture.preload(TEXTURE_PATHS.marsLod);
useTexture.preload(TEXTURE_PATHS.venusLod);
useTexture.preload(TEXTURE_PATHS.earthLod);
useTexture.preload(TEXTURE_PATHS.earthCloudsLod);
useTexture.preload(TEXTURE_PATHS.jupiterLod);
useTexture.preload(TEXTURE_PATHS.mercuryLod);
useTexture.preload(TEXTURE_PATHS.saturnLod);
useTexture.preload(TEXTURE_PATHS.neptuneLod);
useTexture.preload(TEXTURE_PATHS.uranusLod);
useTexture.preload(TEXTURE_PATHS.ceresLod);
useTexture.preload(TEXTURE_PATHS.moonLod);
useTexture.preload(TEXTURE_PATHS.ganymedeLod);
useTexture.preload(TEXTURE_PATHS.sunLod);
useTexture.preload(TEXTURE_PATHS.grass);
useTexture.preload(TEXTURE_PATHS.rock);
useTexture.preload(TEXTURE_PATHS.sand);
useTexture.preload(TEXTURE_PATHS.waterNormals);
useTexture.preload(TEXTURE_PATHS.noiseMid);
useTexture.preload(TEXTURE_PATHS.noiseHigh);
useLoader.preload(THREE.TextureLoader, TEXTURE_PATHS.smoke);
useTexture.preload(TEXTURE_PATHS.concrete);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.steel);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.titanium);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.plant);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.megacredit);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.card);

interface TextureAssets {
  mars: THREE.Texture;
  venus: THREE.Texture;
  earth: THREE.Texture;
  earthClouds: THREE.Texture;
  jupiter: THREE.Texture;
  mercury: THREE.Texture;
  saturn: THREE.Texture;
  neptune: THREE.Texture;
  uranus: THREE.Texture;
  ceres: THREE.Texture;
  moon: THREE.Texture;
  ganymede: THREE.Texture;
  sun: THREE.Texture;
  grass: THREE.Texture;
  rock: THREE.Texture;
  sand: THREE.Texture;
  waterNormals: THREE.Texture;
  noiseMid: THREE.Texture;
  noiseHigh: THREE.Texture;
  concrete: THREE.Texture;
  smoke: THREE.Texture;
  marsLod: THREE.Texture;
  venusLod: THREE.Texture;
  earthLod: THREE.Texture;
  earthCloudsLod: THREE.Texture;
  jupiterLod: THREE.Texture;
  mercuryLod: THREE.Texture;
  saturnLod: THREE.Texture;
  neptuneLod: THREE.Texture;
  uranusLod: THREE.Texture;
  ceresLod: THREE.Texture;
  moonLod: THREE.Texture;
  ganymedeLod: THREE.Texture;
  sunLod: THREE.Texture;
  resourceIcons: Record<ResourceIconName, THREE.Texture>;
  getResourceIcon: (bonusType: string) => THREE.Texture;
}

export function useTextures(): TextureAssets {
  const mars = useTexture(TEXTURE_PATHS.mars);
  const venus = useTexture(TEXTURE_PATHS.venus);
  const earth = useTexture(TEXTURE_PATHS.earth);
  const earthClouds = useTexture(TEXTURE_PATHS.earthClouds);
  const jupiter = useTexture(TEXTURE_PATHS.jupiter);
  const mercury = useTexture(TEXTURE_PATHS.mercury);
  const saturn = useTexture(TEXTURE_PATHS.saturn);
  const neptune = useTexture(TEXTURE_PATHS.neptune);
  const uranus = useTexture(TEXTURE_PATHS.uranus);
  const ceres = useTexture(TEXTURE_PATHS.ceres);
  const moonTex = useTexture(TEXTURE_PATHS.moon);
  const ganymede = useTexture(TEXTURE_PATHS.ganymede);
  const sun = useTexture(TEXTURE_PATHS.sun);
  const grass = useTexture(TEXTURE_PATHS.grass);
  const rock = useTexture(TEXTURE_PATHS.rock);
  const sand = useTexture(TEXTURE_PATHS.sand);
  const waterNormals = useTexture(TEXTURE_PATHS.waterNormals);
  const noiseMid = useTexture(TEXTURE_PATHS.noiseMid);
  const noiseHigh = useTexture(TEXTURE_PATHS.noiseHigh);
  const concrete = useTexture(TEXTURE_PATHS.concrete);
  const smoke = useLoader(THREE.TextureLoader, TEXTURE_PATHS.smoke);

  const marsLod = useTexture(TEXTURE_PATHS.marsLod);
  const venusLod = useTexture(TEXTURE_PATHS.venusLod);
  const earthLod = useTexture(TEXTURE_PATHS.earthLod);
  const earthCloudsLod = useTexture(TEXTURE_PATHS.earthCloudsLod);
  const jupiterLod = useTexture(TEXTURE_PATHS.jupiterLod);
  const mercuryLod = useTexture(TEXTURE_PATHS.mercuryLod);
  const saturnLod = useTexture(TEXTURE_PATHS.saturnLod);
  const neptuneLod = useTexture(TEXTURE_PATHS.neptuneLod);
  const uranusLod = useTexture(TEXTURE_PATHS.uranusLod);
  const ceresLod = useTexture(TEXTURE_PATHS.ceresLod);
  const moonLod = useTexture(TEXTURE_PATHS.moonLod);
  const ganymedeLod = useTexture(TEXTURE_PATHS.ganymedeLod);
  const sunLod = useTexture(TEXTURE_PATHS.sunLod);

  const steelIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.steel);
  const titaniumIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.titanium);
  const plantIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.plant);
  const megacreditIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.megacredit);
  const cardIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.card);

  // Configure textures (idempotent — drei caches texture objects by path)
  useMemo(() => {
    mars.colorSpace = THREE.SRGBColorSpace;
    mars.wrapS = mars.wrapT = THREE.ClampToEdgeWrapping;

    venus.colorSpace = THREE.SRGBColorSpace;
    venus.wrapS = venus.wrapT = THREE.ClampToEdgeWrapping;

    for (const tex of [
      earth,
      earthClouds,
      jupiter,
      mercury,
      saturn,
      neptune,
      uranus,
      ceres,
      moonTex,
      ganymede,
      marsLod,
      venusLod,
      earthLod,
      earthCloudsLod,
      jupiterLod,
      mercuryLod,
      saturnLod,
      neptuneLod,
      uranusLod,
      ceresLod,
      moonLod,
      ganymedeLod,
      sun,
      sunLod,
    ]) {
      tex.colorSpace = THREE.SRGBColorSpace;
      tex.wrapS = tex.wrapT = THREE.ClampToEdgeWrapping;
    }

    grass.wrapS = grass.wrapT = THREE.RepeatWrapping;
    grass.repeat.set(6.9, 6.9);

    sand.wrapS = sand.wrapT = THREE.RepeatWrapping;
    sand.colorSpace = THREE.SRGBColorSpace;

    waterNormals.wrapS = waterNormals.wrapT = THREE.RepeatWrapping;

    noiseMid.wrapS = noiseMid.wrapT = THREE.RepeatWrapping;
    noiseHigh.wrapS = noiseHigh.wrapT = THREE.RepeatWrapping;

    concrete.wrapS = concrete.wrapT = THREE.RepeatWrapping;
    concrete.colorSpace = THREE.SRGBColorSpace;
  }, [
    mars,
    venus,
    earth,
    earthClouds,
    jupiter,
    mercury,
    saturn,
    neptune,
    uranus,
    ceres,
    moonTex,
    ganymede,
    marsLod,
    venusLod,
    earthLod,
    earthCloudsLod,
    jupiterLod,
    mercuryLod,
    saturnLod,
    neptuneLod,
    uranusLod,
    ceresLod,
    moonLod,
    ganymedeLod,
    sun,
    sunLod,
    grass,
    sand,
    waterNormals,
    noiseMid,
    noiseHigh,
    concrete,
  ]);

  const resourceIcons = useMemo(
    () => ({
      steel: steelIcon,
      titanium: titaniumIcon,
      plant: plantIcon,
      megacredit: megacreditIcon,
      card: cardIcon,
    }),
    [steelIcon, titaniumIcon, plantIcon, megacreditIcon, cardIcon],
  );

  const getResourceIcon = useMemo(
    () =>
      (bonusType: string): THREE.Texture => {
        const iconName = BONUS_TYPE_TO_ICON[bonusType];
        return iconName ? resourceIcons[iconName] : resourceIcons.megacredit;
      },
    [resourceIcons],
  );

  return {
    mars,
    venus,
    earth,
    earthClouds,
    jupiter,
    mercury,
    saturn,
    neptune,
    uranus,
    ceres,
    moon: moonTex,
    ganymede,
    sun,
    grass,
    rock,
    sand,
    waterNormals,
    noiseMid,
    noiseHigh,
    concrete,
    smoke,
    marsLod,
    venusLod,
    earthLod,
    earthCloudsLod,
    jupiterLod,
    mercuryLod,
    saturnLod,
    neptuneLod,
    uranusLod,
    ceresLod,
    moonLod: moonLod,
    ganymedeLod,
    sunLod,
    resourceIcons,
    getResourceIcon,
  };
}
