import { useGLTF } from "@react-three/drei";
import * as THREE from "three";

const MODEL_PATHS = {
  trees: "/assets/models/trees.glb",
  rock: "/assets/models/rock.glb",
  city: "/assets/models/city.glb",
  flowers: "/assets/models/flowers.glb",
  bird: "/assets/models/bird.glb",
  fence: "/assets/models/fence.glb",
} as const;

useGLTF.preload(MODEL_PATHS.trees);
useGLTF.preload(MODEL_PATHS.rock);
useGLTF.preload(MODEL_PATHS.city);
useGLTF.preload(MODEL_PATHS.flowers);
useGLTF.preload(MODEL_PATHS.bird);
useGLTF.preload(MODEL_PATHS.fence);

interface Models {
  treesScene: THREE.Group;
  rockScene: THREE.Group;
  cityScene: THREE.Group;
  flowersScene: THREE.Group;
  birdScene: THREE.Group;
  birdAnimations: THREE.AnimationClip[];
  fenceScene: THREE.Group;
}

export function useModels(): Models {
  const { scene: treesScene } = useGLTF(MODEL_PATHS.trees);
  const { scene: rockScene } = useGLTF(MODEL_PATHS.rock);
  const { scene: cityScene } = useGLTF(MODEL_PATHS.city);
  const { scene: flowersScene } = useGLTF(MODEL_PATHS.flowers);
  const { scene: birdScene, animations: birdAnimations } = useGLTF(MODEL_PATHS.bird);
  const { scene: fenceScene } = useGLTF(MODEL_PATHS.fence);

  return { treesScene, rockScene, cityScene, flowersScene, birdScene, birdAnimations, fenceScene };
}
