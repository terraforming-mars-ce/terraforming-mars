import { useMemo } from "react";
import { useTexture } from "@react-three/drei";
import { useLoader } from "@react-three/fiber";
import * as THREE from "three";

const TEXTURE_PATHS = {
  mars: "/assets/textures/mars_8k.jpg",
  venus: "/assets/textures/4k_venus_atmosphere.jpg",
  grass: "/assets/textures/grass.jpg",
  rock: "/assets/textures/rock.jpg",
  sand: "/assets/textures/sand.jpg",
  waterNormals: "/assets/textures/waternormals.jpg",
  noiseMid: "/assets/textures/noise_mid.png",
  noiseHigh: "/assets/textures/noise_high.png",
  smoke: "/assets/effects/smoke.png",
  concrete: "/assets/textures/concrete.jpg",
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
  grass: THREE.Texture;
  rock: THREE.Texture;
  sand: THREE.Texture;
  waterNormals: THREE.Texture;
  noiseMid: THREE.Texture;
  noiseHigh: THREE.Texture;
  concrete: THREE.Texture;
  smoke: THREE.Texture;
  resourceIcons: Record<ResourceIconName, THREE.Texture>;
  getResourceIcon: (bonusType: string) => THREE.Texture;
}

export function useTextures(): TextureAssets {
  const mars = useTexture(TEXTURE_PATHS.mars);
  const venus = useTexture(TEXTURE_PATHS.venus);
  const grass = useTexture(TEXTURE_PATHS.grass);
  const rock = useTexture(TEXTURE_PATHS.rock);
  const sand = useTexture(TEXTURE_PATHS.sand);
  const waterNormals = useTexture(TEXTURE_PATHS.waterNormals);
  const noiseMid = useTexture(TEXTURE_PATHS.noiseMid);
  const noiseHigh = useTexture(TEXTURE_PATHS.noiseHigh);
  const concrete = useTexture(TEXTURE_PATHS.concrete);
  const smoke = useLoader(THREE.TextureLoader, TEXTURE_PATHS.smoke);

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

    grass.wrapS = grass.wrapT = THREE.RepeatWrapping;
    grass.repeat.set(6.9, 6.9);

    sand.wrapS = sand.wrapT = THREE.RepeatWrapping;
    sand.colorSpace = THREE.SRGBColorSpace;

    waterNormals.wrapS = waterNormals.wrapT = THREE.RepeatWrapping;

    noiseMid.wrapS = noiseMid.wrapT = THREE.RepeatWrapping;
    noiseHigh.wrapS = noiseHigh.wrapT = THREE.RepeatWrapping;

    concrete.wrapS = concrete.wrapT = THREE.RepeatWrapping;
    concrete.colorSpace = THREE.SRGBColorSpace;
  }, [mars, venus, grass, sand, waterNormals, noiseMid, noiseHigh, concrete]);

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
    grass,
    rock,
    sand,
    waterNormals,
    noiseMid,
    noiseHigh,
    concrete,
    smoke,
    resourceIcons,
    getResourceIcon,
  };
}
