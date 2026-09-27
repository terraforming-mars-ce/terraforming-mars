import { assetUrl } from "@/assets";
import { useGLTF } from "@react-three/drei";
import * as THREE from "three";

const MODEL_PATHS = {
  trees: assetUrl("models/trees"),
  rock: assetUrl("models/rock"),
  city: assetUrl("models/city"),
  flowers: assetUrl("models/flowers"),
  bird: assetUrl("models/bird"),
  fence: assetUrl("models/fence"),
  spaceship: assetUrl("models/spaceship"),
  satellite: assetUrl("models/satellite"),
  phobos: assetUrl("models/phobos"),
} as const;

useGLTF.preload(MODEL_PATHS.trees);
useGLTF.preload(MODEL_PATHS.rock);
useGLTF.preload(MODEL_PATHS.city);
useGLTF.preload(MODEL_PATHS.flowers);
useGLTF.preload(MODEL_PATHS.bird);
useGLTF.preload(MODEL_PATHS.fence);
useGLTF.preload(MODEL_PATHS.spaceship);
useGLTF.preload(MODEL_PATHS.satellite);
useGLTF.preload(MODEL_PATHS.phobos);

interface Models {
  treesScene: THREE.Group;
  rockScene: THREE.Group;
  cityScene: THREE.Group;
  flowersScene: THREE.Group;
  birdScene: THREE.Group;
  birdAnimations: THREE.AnimationClip[];
  fenceScene: THREE.Group;
  spaceshipScene: THREE.Group;
  satelliteScene: THREE.Group;
  phobosScene: THREE.Group;
}

export function useModels(): Models {
  const { scene: treesScene } = useGLTF(MODEL_PATHS.trees);
  const { scene: rockScene } = useGLTF(MODEL_PATHS.rock);
  const { scene: cityScene } = useGLTF(MODEL_PATHS.city);
  const { scene: flowersScene } = useGLTF(MODEL_PATHS.flowers);
  const { scene: birdScene, animations: birdAnimations } = useGLTF(MODEL_PATHS.bird);
  const { scene: fenceScene } = useGLTF(MODEL_PATHS.fence);
  const { scene: spaceshipScene } = useGLTF(MODEL_PATHS.spaceship);
  const { scene: satelliteScene } = useGLTF(MODEL_PATHS.satellite);
  const { scene: phobosScene } = useGLTF(MODEL_PATHS.phobos);

  return {
    treesScene,
    rockScene,
    cityScene,
    flowersScene,
    birdScene,
    birdAnimations,
    fenceScene,
    spaceshipScene,
    satelliteScene,
    phobosScene,
  };
}
