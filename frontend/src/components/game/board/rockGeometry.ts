import * as THREE from "three";

const cache = new WeakMap<THREE.Group, THREE.BufferGeometry>();

export function sharedRockGeometry(scene: THREE.Group) {
  const cached = cache.get(scene);
  if (cached) {
    return cached;
  }
  let source: THREE.Mesh | undefined;
  scene.traverse((child) => {
    if (child instanceof THREE.Mesh && !/plane|ground/i.test(child.name)) {
      source = child;
    }
  });
  if (!source) {
    throw new Error("Rock model contains no rock mesh");
  }
  source.updateWorldMatrix(true, false);
  const geometry = source.geometry.clone().applyMatrix4(source.matrixWorld);
  geometry.computeBoundingBox();
  const scale = 0.04 / Math.max(...geometry.boundingBox!.getSize(new THREE.Vector3()).toArray());
  geometry.rotateX(Math.PI / 2);
  geometry.computeBoundingBox();
  const bounds = geometry.boundingBox!;
  const center = bounds.getCenter(new THREE.Vector3());
  geometry.translate(-center.x, -center.y, -bounds.min.z);
  geometry.scale(scale, scale, scale);
  geometry.computeVertexNormals();
  geometry.computeBoundingBox();
  geometry.computeBoundingSphere();
  cache.set(scene, geometry);
  return geometry;
}

export function createNuclearDebrisMaterial(map: THREE.Texture) {
  const material = new THREE.MeshLambertMaterial({
    map,
    emissive: new THREE.Color(0.13, 0.07, 0.035),
    emissiveIntensity: 0.15,
    emissiveMap: map,
  });
  material.userData.planetHaze = true;
  return material;
}
