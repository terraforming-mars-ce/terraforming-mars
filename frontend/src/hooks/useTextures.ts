import { assetUrl } from "@/assets";
import { useMemo } from "react";
import { useTexture } from "@react-three/drei";
import { useLoader } from "@react-three/fiber";
import * as THREE from "three";
import { QUICK_MODE } from "@/utils/quickMode.ts";

type PlanetName =
  | "mars"
  | "venus"
  | "earth"
  | "earth-clouds"
  | "jupiter"
  | "mercury"
  | "saturn"
  | "neptune"
  | "uranus"
  | "ceres"
  | "moon"
  | "ganymede"
  | "sun";

/** Quick mode swaps the 8K planet surfaces for their 512px overview maps. */
function planetSurface(name: PlanetName) {
  return QUICK_MODE
    ? assetUrl(`textures/planets/${name}/overview`)
    : assetUrl(`textures/planets/${name}/surface`);
}

const TEXTURE_PATHS = {
  mars: planetSurface("mars"),
  venus: planetSurface("venus"),
  earth: planetSurface("earth"),
  earthClouds: planetSurface("earth-clouds"),
  jupiter: planetSurface("jupiter"),
  mercury: planetSurface("mercury"),
  saturn: planetSurface("saturn"),
  neptune: planetSurface("neptune"),
  uranus: planetSurface("uranus"),
  ceres: planetSurface("ceres"),
  moon: planetSurface("moon"),
  ganymede: planetSurface("ganymede"),
  sun: planetSurface("sun"),
  grass: assetUrl("textures/terrain/grass"),
  rock: assetUrl("textures/terrain/rock"),
  sand: assetUrl("textures/terrain/sand"),
  waterNormals: assetUrl("textures/terrain/waternormals"),
  noiseMid: assetUrl("textures/terrain/noise-mid"),
  noiseHigh: assetUrl("textures/terrain/noise-high"),
  smoke: assetUrl("textures/effects/smoke"),
  concrete: assetUrl("textures/terrain/concrete"),
  marsLod: assetUrl("textures/planets/mars/overview"),
  venusLod: assetUrl("textures/planets/venus/overview"),
  earthLod: assetUrl("textures/planets/earth/overview"),
  earthCloudsLod: assetUrl("textures/planets/earth-clouds/overview"),
  jupiterLod: assetUrl("textures/planets/jupiter/overview"),
  mercuryLod: assetUrl("textures/planets/mercury/overview"),
  saturnLod: assetUrl("textures/planets/saturn/overview"),
  neptuneLod: assetUrl("textures/planets/neptune/overview"),
  uranusLod: assetUrl("textures/planets/uranus/overview"),
  ceresLod: assetUrl("textures/planets/ceres/overview"),
  moonLod: assetUrl("textures/planets/moon/overview"),
  ganymedeLod: assetUrl("textures/planets/ganymede/overview"),
  sunLod: assetUrl("textures/planets/sun/overview"),
} as const;

const CITY_FACADE_PATHS = [
  [
    assetUrl("textures/terrain/concrete-tile-facade-color"),
    assetUrl("textures/terrain/concrete-tile-facade-normal"),
    assetUrl("textures/terrain/concrete-tile-facade-roughness"),
  ],
  [
    assetUrl("textures/terrain/rectangular-facade-tiles-color"),
    assetUrl("textures/terrain/rectangular-facade-tiles-normal"),
    assetUrl("textures/terrain/rectangular-facade-tiles-roughness"),
  ],
  [
    assetUrl("textures/terrain/ribbed-concrete-wall-color"),
    assetUrl("textures/terrain/ribbed-concrete-wall-normal"),
    assetUrl("textures/terrain/ribbed-concrete-wall-roughness"),
  ],
];
// The loader cache is keyed by the whole URL list, so each preload must pass exactly the
// array that useTextures later requests; per-URL preloads would never be reused.
const CITY_FACADE_URLS = CITY_FACADE_PATHS.flat();
useTexture.preload(CITY_FACADE_URLS);

// Layer order of the landscape's ground texture array; shaders index layers by position.
const GROUND_LAYER_NAMES = [
  "leafy-grass",
  "leaves-forest-ground",
  "brown-mud-leaves-01",
  "frost",
  "frost-patchy",
] as const;
const GROUND_LAYER_PATHS = {
  color: GROUND_LAYER_NAMES.map((name) => assetUrl(`textures/terrain/${name}-color` as const)),
  detail: GROUND_LAYER_NAMES.map((name) => assetUrl(`textures/terrain/${name}-detail` as const)),
};
const ICE_LAYER_NAMES = ["ice-solid", "ice-thin"] as const;
const ICE_LAYER_PATHS = {
  color: ICE_LAYER_NAMES.map((name) => assetUrl(`textures/terrain/${name}-color` as const)),
  detail: ICE_LAYER_NAMES.map((name) => assetUrl(`textures/terrain/${name}-detail` as const)),
};
for (const paths of [GROUND_LAYER_PATHS, ICE_LAYER_PATHS]) {
  useTexture.preload(paths.color);
  useTexture.preload(paths.detail);
}

export interface TextureLayers {
  color: THREE.Texture[];
  detail: THREE.Texture[];
}

const RESOURCE_ICON_PATHS = {
  steel: assetUrl("icons/resources/steel", 256),
  titanium: assetUrl("icons/resources/titanium", 256),
  plant: assetUrl("icons/resources/plant", 256),
  megacredit: assetUrl("icons/resources/megacredit", 256),
  card: assetUrl("icons/resources/card", 256),
  heat: assetUrl("icons/resources/heat", 256),
  power: assetUrl("icons/resources/power", 256),
  microbe: assetUrl("icons/resources/microbe", 256),
  animal: assetUrl("icons/resources/animal", 256),
  science: assetUrl("icons/resources/science", 256),
  data: assetUrl("icons/resources/data", 256),
  director: assetUrl("icons/resources/director", 256),
  temperature: assetUrl("icons/parameters/temperature", 256),
  ocean: assetUrl("icons/placements/ocean", 256),
  wild: assetUrl("icons/resources/wild", 256),
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
  heat: "heat",
  energy: "power",
  "energy-production": "power",
  microbe: "microbe",
  animal: "animal",
  science: "science",
  data: "data",
  delegate: "director",
  temperature: "temperature",
  ocean: "ocean",
  "ocean-placement": "ocean",
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
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.heat);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.power);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.microbe);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.animal);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.science);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.data);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.director);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.temperature);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.ocean);
useLoader.preload(THREE.TextureLoader, RESOURCE_ICON_PATHS.wild);

interface TextureAssets {
  cityFacades: { color: THREE.Texture; normal: THREE.Texture; roughness: THREE.Texture }[];
  groundLayers: TextureLayers;
  iceLayers: TextureLayers;
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
  const facadeTextures = useTexture(CITY_FACADE_URLS);
  const cityFacades = useMemo(
    () =>
      CITY_FACADE_PATHS.map((_, i) => {
        const [color, normal, roughness] = facadeTextures.slice(i * 3, i * 3 + 3);
        color.colorSpace = THREE.SRGBColorSpace;
        normal.colorSpace = roughness.colorSpace = THREE.NoColorSpace;
        for (const texture of [color, normal, roughness]) {
          texture.wrapS = texture.wrapT = THREE.RepeatWrapping;
          texture.anisotropy = 4;
        }
        return { color, normal, roughness };
      }),
    [facadeTextures],
  );
  const groundColors = useTexture(GROUND_LAYER_PATHS.color);
  const groundDetails = useTexture(GROUND_LAYER_PATHS.detail);
  const groundLayers = useMemo(
    () => ({ color: groundColors, detail: groundDetails }),
    [groundColors, groundDetails],
  );
  const iceColors = useTexture(ICE_LAYER_PATHS.color);
  const iceDetails = useTexture(ICE_LAYER_PATHS.detail);
  const iceLayers = useMemo(
    () => ({ color: iceColors, detail: iceDetails }),
    [iceColors, iceDetails],
  );
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
  const heatIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.heat);
  const powerIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.power);
  const microbeIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.microbe);
  const animalIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.animal);
  const scienceIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.science);
  const dataIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.data);
  const directorIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.director);
  const temperatureIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.temperature);
  const oceanIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.ocean);
  const wildIcon = useLoader(THREE.TextureLoader, RESOURCE_ICON_PATHS.wild);

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
      heat: heatIcon,
      power: powerIcon,
      microbe: microbeIcon,
      animal: animalIcon,
      science: scienceIcon,
      data: dataIcon,
      director: directorIcon,
      temperature: temperatureIcon,
      ocean: oceanIcon,
      wild: wildIcon,
    }),
    [
      steelIcon,
      titaniumIcon,
      plantIcon,
      megacreditIcon,
      cardIcon,
      heatIcon,
      powerIcon,
      microbeIcon,
      animalIcon,
      scienceIcon,
      dataIcon,
      directorIcon,
      temperatureIcon,
      oceanIcon,
      wildIcon,
    ],
  );

  const getResourceIcon = useMemo(
    () =>
      (bonusType: string): THREE.Texture => {
        const iconName = BONUS_TYPE_TO_ICON[bonusType];
        return iconName ? resourceIcons[iconName] : resourceIcons.wild;
      },
    [resourceIcons],
  );

  return {
    cityFacades,
    groundLayers,
    iceLayers,
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
