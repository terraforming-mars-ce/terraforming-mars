import * as THREE from "three";

// Pick the spherical surface without walking its render triangles.
export function sphereRaycast(
  this: THREE.Mesh<THREE.SphereGeometry>,
  raycaster: THREE.Raycaster,
  intersections: THREE.Intersection[],
) {
  inverse.copy(this.matrixWorld).invert();
  ray.copy(raycaster.ray).applyMatrix4(inverse);
  sphere.radius = this.geometry.parameters.radius;
  if (!ray.intersectSphere(sphere, point)) {
    return;
  }
  point.applyMatrix4(this.matrixWorld);
  const distance = raycaster.ray.origin.distanceTo(point);
  if (distance < raycaster.near || distance > raycaster.far) {
    return;
  }
  intersections.push({ distance, point: point.clone(), object: this });
}

const inverse = new THREE.Matrix4();
const ray = new THREE.Ray();
const sphere = new THREE.Sphere();
const point = new THREE.Vector3();
