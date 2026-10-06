import * as THREE from "three";
import type { SceneGroupStats } from "../../../services/performanceStore";

// Estimates what each part of the scene submits to the GPU: visible meshes, their draw calls and
// triangles (instances included). Objects are grouped by the nearest `userData.perfGroup` tag.
// This is an upper bound, since frustum culling is not applied.
export function sceneCensus(scene: THREE.Scene): SceneGroupStats[] {
  const groups = new Map<string, SceneGroupStats>();
  scene.traverseVisible((object) => {
    const mesh = object as THREE.Mesh;
    if (!mesh.isMesh && !(object as THREE.Points).isPoints && !(object as THREE.Line).isLine) {
      return;
    }
    const geometry = mesh.geometry as THREE.BufferGeometry | undefined;
    if (!geometry) {
      return;
    }
    const instances = (mesh as THREE.InstancedMesh).isInstancedMesh
      ? (mesh as THREE.InstancedMesh).count
      : 1;
    if (instances === 0) {
      return;
    }
    const elements = geometry.index
      ? geometry.index.count
      : (geometry.attributes.position?.count ?? 0);
    const range = Math.min(elements, geometry.drawRange.count);
    const materials = Array.isArray(mesh.material) ? mesh.material.length : 1;
    let label: string | undefined;
    for (let node: THREE.Object3D | null = object; node; node = node.parent) {
      label = node.userData.perfGroup as string | undefined;
      if (label) {
        break;
      }
    }
    label ??= `other: ${object.name || geometry.type}`;
    const entry = groups.get(label) ?? { group: label, objects: 0, drawCalls: 0, triangles: 0 };
    entry.objects++;
    entry.drawCalls += materials;
    entry.triangles += mesh.isMesh ? Math.floor(range / 3) * instances : 0;
    groups.set(label, entry);
  });
  return [...groups.values()].sort((a, b) => b.triangles - a.triangles);
}
